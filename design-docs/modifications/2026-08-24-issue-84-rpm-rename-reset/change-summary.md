# Issue #84：RPM 规则改名导致计数器重置修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/84](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/84)

> 对绑定 RPM 限流策略（`max_requests=1, window_minutes=1`）的 API Key：
> - 新建 Key 后同窗口第 2 个请求返回 429（限流生效，符合预期）；
> - 通过管理面 PATCH 仅修改 RPM 规则名（其余字段不变）后，同窗口内请求被再次放行（返回 200）——计数器从零重新开始。
>
> 即：**修改 RPM 规则名称 = 计数器重置**。

## 2. 根因分析

BFE 的 RPM 计数器 Redis key 由「策略 ID + 规则名」构成：

```go
// source/bfe/bfe_modules/mod_ai_rate_limit/policy_limiter.go

func buildRpmInstId(rule *RPMRuleConf) string {
    if rule.Name != "" {
        return rule.Name // 规则名非空时直接作实例 ID
    }
    ...
}

redisKey := buildRedisKey(policyId, fmt.Sprintf("rpm_%s", rpmInstId))
// => default_bfe_<policyId>_rpm_<ruleName>
```

规则名变更 → 新的 Redis key → 计数器从零。

而控制面（`ai-gateway-api`）把 per-rule name 原样透传给 BFE：

```go
// source/ai-gateway-api/model/rate_limit_policy/rate_limit_policy_manager.go:187-188

exportPolicy.Rules.RPM = append(..., ExportRPMConfig{Name: rpm.Name, ...})
```

> **注意**：用 `rule.Model` 替代 `rule.Name` 同样不稳定。`model` 是用户可编辑字段，修改 model 也会改变 Redis key；且通配符 `*`、多个规则指向同一 model 等场景都会带来 key 冲突或语义混乱。因此计数器 key 不能依赖任何用户可编辑的业务字段（name / model）。

## 2.1 参考：quota_plan 的 Redis key 生成方式

`ai-gateway-api` 在导出 quota_plan 时，由控制面生成稳定的 `RedisKey` 并下发给 BFE：

```go
// source/ai-gateway-api/model/imods/exporter.go:533-536

func convertQuotaPlanToExport(qp *quota.QuotaPlanParam, id string, redisKeyID string) *QuotaPlan {
    result := &QuotaPlan{
        Id:          id,
        RedisKey:    fmt.Sprintf("QUOTA_%s", redisKeyID),
        ...
    }
    ...
}
```

其中 `redisKeyID` 取的是**配额归属方**的稳定标识：
- API-Key 级配额：`redisKeyID = apiKey.Key`
- Entity 级配额：`redisKeyID = entity.EntityID`

BFE 侧只消费现成的 `QuotaPlan.RedisKey`，不再根据计划名称/ID 自己拼装 key。这保证了：修改 quota_plan 名称、修改 model 等都不会重置 Redis 计数器。

RPM/TPM 限流应参照同一原则：**由控制面生成稳定的 Redis key 并下发，BFE 直接使用，key 不依赖用户可编辑字段**。

## 3. 设计文档依据（漂移点）

设计文档 `source/ai-gateway-api/design-docs/sys-design/details/限流策略与导出.md`：

- §3.2 数据模型 `RPMConfig{Model, Limit}` —— 无 `name` 字段；
- §3.3 BFE 侧 `BfeRateLimitPolicy{RpmLimits map[string]int}` —— 按 `model` 为 key；
- §7 BFE 行为：按 apikey 绑定找策略 → 按 model 匹配 → 超限 429，全程不涉及规则名；
- §6.5 导出策略名固定 `rlp-<policy_id>`（改名不改变 `policy_id`）。

**结论**：设计文档未定义「规则名 / 改名重置计数器」语义，其模型（按 `policy + model` 计数）暗示改名不应重置。实际实现引入 per-rule name 维度并以其作计数 key，属设计文档与实现漂移，语义需产品侧明确。

## 4. 产品决策建议（推荐方案）

### 4.1 决策：规则名 / model **都不应**作为计数器实例标识

按现有设计文档语义，RPM 限流是**策略级**收敛；规则名/model 仅用于管理面展示与请求匹配，改名或改 model 都不应导致计数器重置。

### 4.2 若采用此决策，当前实现为缺陷

风险：持有管理面权限者可通过反复改名/改 model 绕过 RPM 计数（仅限管理面，数据面调用方无法直接利用，但语义必须明确）。

### 4.3 备选决策（不推荐）

若产品侧坚持「改名 = 新规则、新计数器」：
- 当前实现符合该语义，但需在 `ai-gateway-api` 补写设计文档语义，并在 OpenAPI / 用户文档中明确说明「改名会重置限流计数」。

## 5. 修复方案（基于推荐决策）

### 5.1 总体思路

参照 quota_plan 做法：
1. 由 `ai-gateway-api` 在导出时为每条 RPM/TPM 规则生成**稳定的 Redis key**；
2. key 基于不会随用户编辑而变的标识：`(policy_id, 规则在数组中的下标)`；
3. BFE 直接使用导出的 `RedisKey`，不再根据 `rule.Name` / `rule.Model` 自己拼装 key。

这样：
- 改名不改 key → 计数器不重置；
- 改 model 不改 key → 计数器不重置；
- 删除/新增规则会改变后续规则下标（计数器会重置），这属于规则集合结构性变更，与单条规则改名不同，语义上可接受。

### 5.2 改动点

| 仓库 | 文件 | 修改内容 |
|------|------|----------|
| `ai-gateway-api` | `model/rate_limit_policy/rate_limit_policy.go` | 在 `ExportRPMConfig` / `ExportTPMConfig` 中新增 `RedisKey string` 字段。 |
| `ai-gateway-api` | `model/rate_limit_policy/rate_limit_policy_manager.go` | 导出规则时按 `(policyID, 规则下标)` 生成 `RedisKey`，例如 `fmt.Sprintf("RL_RPM_rlp-%d_%d", policyID, idx)`。 |
| `bfe` | `bfe_modules/mod_ai_rate_limit/data_load.go` | `RPMRuleConf` / `TPMRuleConf` 新增 `RedisKey string` 字段；`Convert()` 从配置文件中读取。 |
| `bfe` | `bfe_modules/mod_ai_rate_limit/policy_limiter.go` | `newRpmLimiterItem` / `newTpmLimiterItem` 直接使用 `rule.RedisKey` 构建 Redis key，废弃 `buildRpmInstId` / `buildTpmInstId` 中基于 `Name` 的逻辑。 |
| `ai-gateway-api` | `design-docs/sys-design/details/限流策略与导出.md` | 补充明确语义：限流计数器 key 由控制面按 `(policy_id, rule_index)` 稳定生成，不依赖规则名/model。 |
| `ai-gateway-api` / 测试 | E2E 用例 `SC1302-TC004` | 若原用例把「改名重置计数器」作为预期行为，需同步修正为「改名后计数器继续生效」。 |

### 5.3 关键代码示例

#### ai-gateway-api 侧

```go
// model/rate_limit_policy/rate_limit_policy_manager.go

for idx, rpm := range policy.RpmConfigs {
    models := []string{"*"}
    if rpm.Model != "" && rpm.Model != "*" {
        models = []string{rpm.Model}
    }
    exportPolicy.Rules.RPM = append(exportPolicy.Rules.RPM, ExportRPMConfig{
        Name:          rpm.Name,            // 保留展示名，但不参与 key 生成
        Models:        models,
        WindowMinutes: rpm.WindowMinutes,
        MaxRequests:   rpm.MaxRequests,
        Burst:         1,
        RedisKey:      fmt.Sprintf("RL_RPM_rlp-%d_%d", policyID, idx),
    })
}
```

#### BFE 侧

```go
// bfe_modules/mod_ai_rate_limit/data_load.go

type RPMRuleConfFile struct {
    Name          string   `json:"name"`
    WindowMinutes int      `json:"window_minutes"`
    MaxRequests   int64    `json:"max_requests"`
    Burst         int64    `json:"burst"`
    Models        []string `json:"models"`
    RedisKey      string   `json:"redis_key"` // 新增
}

type RPMRuleConf struct {
    Name        string
    TimeWindow  int
    MaxRequests int64
    Burst       int64
    Models      []string
    RedisKey    string // 新增
}

func (f *RPMRuleConfFile) Convert() *RPMRuleConf {
    return &RPMRuleConf{
        Name:        f.Name,
        TimeWindow:  f.WindowMinutes * 60,
        MaxRequests: f.MaxRequests,
        Burst:       f.Burst,
        Models:      f.Models,
        RedisKey:    f.RedisKey,
    }
}
```

```go
// bfe_modules/mod_ai_rate_limit/policy_limiter.go

func buildRpmRedisKey(policyId string, rule *RPMRuleConf) string {
    if rule.RedisKey != "" {
        // 优先使用控制面下发的稳定 key
        return buildRedisKey(policyId, rule.RedisKey)
    }
    // 兼容旧配置（无 redis_key 时按原 name 逻辑兜底，避免线上升级时计数器全丢）
    return buildRedisKey(policyId, fmt.Sprintf("rpm_%s", buildRpmInstId(rule)))
}

// newPolicyLimiterSet 中
for _, rule := range policy.Rules.RPM {
    redisKey := buildRpmRedisKey(policyId, rule)
    limiter := limit_rate.NewQPMLimiter(
        redisKey,
        rule.Burst,
        int64(rule.TimeWindow),
        rule.MaxRequests,
    )
    ...
}
```

## 6. 影响面

| 项目 | 说明 |
|------|------|
| 限流语义 | 明确 RPM/TPM 计数器 key 由 `(policy_id, rule_index)` 决定；改名、改 model 不重置计数器。 |
| 兼容性 | 导出格式新增 `redis_key` 字段；BFE 对旧配置保留 name 兜底，升级平滑。 |
| 数据面 | 新 key 格式为 `default_bfe_<policyId>_<RL_RPM_rlp-X_Y>`；历史基于 name 的 key 会自然过期。 |
| 测试 | E2E `SC1302-TC004` 需按新语义调整预期；需补充「改 model 不重置」的回归用例。 |
| 管理面 | 规则名/model 仍可作为展示/匹配字段保留，但不参与计数器 key 生成。 |

## 7. 上线检查清单

- [ ] 产品侧确认「改名/改 model 不重置计数器」语义；
- [ ] `ai-gateway-api` 侧为 RPM/TPM 规则生成并导出 `redis_key`；
- [ ] `bfe` 侧读取 `redis_key` 并用于构建 Redis key；
- [ ] 更新 `design-docs/sys-design/details/限流策略与导出.md` 语义说明；
- [ ] 修正 E2E 用例 `SC1302-TC004` 预期；
- [ ] 回归测试：新建 Key → 触发限流 429 → PATCH 改名 → 仍返回 429 → PATCH 改 model → 仍返回 429。

## 8. 参考文档

- [Issue #84](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/84)
- `ai-gateway-api/design-docs/sys-design/details/限流策略与导出.md`
- `ai-gateway-api/model/rate_limit_policy/rate_limit_policy_manager.go`
- `ai-gateway-api/model/imods/exporter.go`（quota_plan RedisKey 生成参考）
- `bfe/bfe_modules/mod_ai_rate_limit/policy_limiter.go`
- `bfe/bfe_modules/mod_ai_rate_limit/data_load.go`
