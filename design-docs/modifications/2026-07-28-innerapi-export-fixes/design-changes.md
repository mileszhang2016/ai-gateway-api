# InnerAPI 配置导出问题修复设计说明（2026-07-28）

## 1. 概述

### 1.1 变更背景

在 OpenAPI v0.3.0 收敛后，数据面配置（`/inner-api/v1/configs/*`）的导出逻辑暴露出若干与 Entity 层级、无限配额、禁用策略相关的边界问题。这些问题会导致：

- API-Key 挂载到中间层级 Entity 时，**仅能命中直接挂载 Entity 的路由规则**，无法继承父级 Entity 的路由规则；
- API-Key 的 `QuotaPlan` 在导出阶段未被加载，导致 `GetRemainingQuota` 无法正确判断 token 是否已耗尽；
- **无限配额计划**仍被写入 `mod_api_key` 的 `QuotaPlans` 与 `tokens[].quota_plan_ids`，下发到 BFE 后产生无意义的配额校验；
- **禁用的限流策略**仍被导出并绑定到 API-Key，导致 BFE 加载已失效的策略；
- AI 路由规则的类型常量 `api_key` 与 BFE 期望的 `apikey` 不一致。

本次修复全部属于 **InnerAPI 导出层（`model/imods`、`model/quota`、`model/shared`）** 的行为修正，不修改 OpenAPI 接口定义，也不修改数据库存储结构。

### 1.2 目标版本

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-28 |
| 目标版本 | 在 OpenAPI v0.3.0 基础上的热修复 |
| 基线 Commit | `a32c131 refactor(openapi): converge OpenAPI interfaces to v0.3.0` |
| 涉及文件 | `model/icluster_conf/api_key.go`<br>`model/imods/ai_route_exporter.go`<br>`model/imods/exporter.go`<br>`model/quota/rate_limit_policy_manager.go`<br>`model/shared/route_rules.go` |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **语义一致** | 导出行为与 Entity 继承、配额计划、限流策略的启用/禁用语义保持一致。 |
| **最小侵入** | 仅修改导出/转换逻辑，不动 OpenAPI 模型与数据库存储。 |
| **向下兼容** | 数据库 schema、OpenAPI 请求/响应结构均不变；仅 InnerAPI 导出的配置内容更精确。 |
| **以代码为准** | 本次修复直接源于代码实现缺陷，文档同步以实际 diff 为准。 |

---

## 2. 模块/子系统设计变更

### 2.1 API-Key 状态计算（`model/icluster_conf/api_key.go`）

#### 2.1.1 问题描述

`GetRemainingQuota(param *APIKeyParam)` 原本只检查 `param.UnlimitedQuota`（API-Key 旧字段），没有检查关联 `QuotaPlan.Unlimited`。当 API-Key 绑定了一个**无限配额计划**时，函数仍会尝试读取 Redis 并返回一个剩余量，导致调用方可能将无限配额误判为有限配额。

#### 2.1.2 修复设计

在函数入口增加对 `param.QuotaPlan.Unlimited` 的判断：

```go
// If the associated quota plan is unlimited, there is no remaining quota to track.
if param.QuotaPlan != nil && param.QuotaPlan.Unlimited != nil && *param.QuotaPlan.Unlimited {
    return nil, nil
}
```

#### 2.1.3 影响范围

- 管理面 OpenAPI 的余额展示逻辑无变化（`populateAssociatedData` 已处理无限配额）；
- 主要影响 **InnerAPI 导出阶段**的 token 状态计算（`enabled/disabled/expired/exhausted`）。

---

### 2.2 AI 路由规则导出（`model/imods/ai_route_exporter.go`）

#### 2.2.1 问题描述

原实现只为 API-Key 绑定**直接挂载 Entity** 的路由表：

```go
if apiKey.EntityID != nil && *apiKey.EntityID != "" {
    entity, exists := entityMap[*apiKey.EntityID]
    if exists && entity.RouteRulesID != nil {
        // 仅处理直接挂载 Entity 的路由规则
    }
}
```

这与配额计划、限流策略的“层级向上继承”语义不一致。若管理员将 API-Key 挂载到“项目 X”，而“部门 A”或“公司根”配置了全局兜底路由，则请求不会命中这些父级路由。

#### 2.2.2 修复设计

改为沿 `parent_id` 自底向上遍历 Entity 层级，依次收集每一级 Entity 的路由规则：

```go
currentEntityID := apiKey.EntityID
for currentEntityID != nil && *currentEntityID != "" {
    entity, exists := entityMap[*currentEntityID]
    if !exists {
        break
    }
    if entity.RouteRulesID != nil {
        // 加载并绑定该 Entity 的路由表
    }
    currentEntityID = entity.ParentID
}
```

绑定顺序仍保持：

1. `apikey_<key>`（API-Key 级）
2. `entity_<name>`（直接 Entity → 父 Entity → … → 根 Entity）
3. `global_default`（Global 级）

BFE 按列表顺序匹配，优先级从高到低。

#### 2.2.3 影响范围

- 仅影响 `/inner-api/v1/configs/ai-route` 导出的 `ApikeyRouteTableBindings`；
- OpenAPI 的 Entity / API-Key 管理接口不变。

---

### 2.3 mod-api-key 导出（`model/imods/exporter.go`）

#### 2.3.1 问题描述 1：导出时 QuotaPlan 未加载

`APIKeyRuleGenerator` 通过原始 storager 读取 `api_keys` 列表，该 storager 不会自动填充关联对象。后续调用 `GetRemainingQuota` 时，由于 `param.QuotaPlan` 为空或缺失 `Quota.Quota`，无法正确判断 token 是否耗尽，可能导致所有有限配额 token 被错误标记为 `exhausted`。

#### 2.3.2 修复设计 1：预加载 QuotaPlan

在生成 token 之前，为每个 API-Key 预加载其关联的 `QuotaPlan`：

```go
for _, one := range apiKeyList {
    if one.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
        quotaPlan, err := rlm.quotaPlanStorager.FetchQuotaPlan(ctx, &quota.QuotaPlanFilter{ID: one.QuotaPlanID})
        if err != nil {
            return nil, err
        }
        if quotaPlan != nil {
            one.QuotaPlan = &shared.QuotaPlanParam{
                Unlimited:             quotaPlan.Unlimited,
                PassWhenNoEnoughQuota: quotaPlan.PassWhenNoEnoughQuota,
                Quota:                 quotaPlan.Quota,
                Unit:                  quotaPlan.Unit,
                ResetPeriod:           quotaPlan.ResetPeriod,
            }
        }
    }
}
```

#### 2.3.3 问题描述 2：无限配额计划仍被导出

`fetchQuotaPlansWithEntityHierarchy` 与 `fetchEntityQuotaPlanHierarchy` 在收集 API-Key 自身及 Entity 层级的配额计划时，未过滤 `unlimited=true` 的计划。这些计划会被写入 `QuotaPlans` 映射，并追加到 `tokens[].quota_plan_ids`，但 BFE 侧无需对无限配额做计数。

#### 2.3.4 修复设计 2：跳过无限配额计划

在将配额计划加入导出缓存和 `quotaPlanIDs` 之前，判断 `quotaPlan.Unlimited`：

```go
if quotaPlan.Unlimited == nil || !*quotaPlan.Unlimited {
    qp := convertQuotaPlanToExport(quotaPlan, ownerKey, redisKey)
    if !rlm.isQuotaPlanCached(productName, qp.Id) {
        // 加入 QuotaPlans 缓存
    }
    quotaPlanIDs = append(quotaPlanIDs, qp.Id)
}
```

该逻辑同时作用于：

- API-Key 自身配额计划（`fetchQuotaPlansWithEntityHierarchy` 前半段）；
- Entity 层级递归配额计划（`fetchEntityQuotaPlanHierarchy`）。

#### 2.3.5 影响范围

- `/inner-api/v1/configs/mod-api-key` 导出的 `QuotaPlans` 与 `tokens[].quota_plan_ids` 更精简；
- token 状态计算更准确；
- BFE 侧无需再处理无意义的无限配额计划。

---

### 2.4 限流策略导出（`model/quota/rate_limit_policy_manager.go`）

#### 2.4.1 问题描述

`RateLimitPolicyGenerator` 在导出时未过滤 `enabled=false` 的策略。原代码虽然将 `Enabled` 字段按原值写入，但禁用的策略仍会被加入 `RateLimitPolicies` 映射，并生成与 API-Key 的绑定关系，导致 BFE 加载已失效策略。

#### 2.4.2 修复设计

在生成策略与绑定关系时，跳过禁用策略：

```go
policyKey := fmt.Sprintf("rlp-%d", policyID)
// Skip disabled policies: only effective policies should be exported
// and bound to API keys.
if policy.Enabled == nil || !*policy.Enabled {
    continue
}
if _, exists := rateLimitPolicies[policyKey]; !exists {
    exportPolicy := &ExportRateLimitPolicy{
        Name:    policyKey,
        Enabled: true,
        // ...
    }
    // ...
}
```

由于只有启用的策略才会进入导出流程，因此导出结构中的 `Enabled` 固定为 `true`。

#### 2.4.3 影响范围

- `/inner-api/v1/configs/rate-limit-policy` 导出的 `RateLimitPolicies` 与 `ApikeyRateLimitPolicyBindings` 不再包含禁用策略；
- OpenAPI 的限流策略 CRUD 接口不变。

---

### 2.5 路由规则类型常量（`model/shared/route_rules.go`）

#### 2.5.1 问题描述

`RouteRulesTypeAPIKey` 常量值为 `"api_key"`，但 BFE 侧 AI 路由模块期望的 key 前缀/类型为 `"apikey"`（下划线差异）。该常量用于 `route_rules.type` 的语义标识以及导出 key 的构造，不一致会导致 BFE 无法正确索引 API-Key 级路由表。

#### 2.5.2 修复设计

将常量从 `"api_key"` 改为 `"apikey"`：

```go
const (
    RouteRulesTypeAPIKey = "apikey"
    RouteRulesTypeEntity = "entity"
    RouteRulesTypeGlobal = "global"
)
```

#### 2.5.3 影响范围

- `route_rules` 表已有记录的 `type` 字段语义不变（数据库仍兼容 `"api_key"` / `"apikey"` 两种历史值）；
- 新创建或更新后的 API-Key 路由表将写入 `"apikey"`；
- AI 路由导出 key 统一为 `apikey_<key>`，与 BFE 侧约定一致。

> 注：若历史数据中存在 `type="api_key"` 的记录，建议通过一次性数据迁移脚本统一改为 `"apikey"`，或保持导出逻辑同时兼容两种值。本次修复仅修改常量与新增记录的行为。

---

## 3. 数据模型与存储设计变更

本次修复**不涉及**数据库表结构变更，所有调整均发生在内存中的导出模型与转换逻辑。

### 3.1 不变更的表结构

| 表 | 说明 |
|------|------|
| `api_keys` | 无变更。 |
| `entities` | 无变更。 |
| `quota_plans` / `quota_balances` | 无变更。 |
| `rate_limit_policies` | 无变更。 |
| `route_rules` | 无结构变更；新记录 `type` 值由 `"api_key"` 变为 `"apikey"`。 |

### 3.2 导出数据结构变化

#### 3.2.1 `/configs/mod-api-key`

- `QuotaPlans` 中不再包含 `unlimited=true` 的计划；
- `tokens[].quota_plan_ids` 中不再包含无限配额计划 ID；
- token 状态（`Enabled`）计算更准确，依赖预加载的 `QuotaPlan.Quota`。

#### 3.2.2 `/configs/ai-route`

- `ApikeyRouteTableBindings` 中 API-Key 的绑定列表会包含所有祖先 Entity 的路由表 key；
- `RouteRules` 映射中新增祖先 Entity 对应的路由表。

#### 3.2.3 `/configs/rate-limit-policy`

- `RateLimitPolicies` 中仅包含 `enabled=true` 的策略；
- `ApikeyRateLimitPolicyBindings` 中不再包含禁用策略的绑定。

---

## 4. 接口层设计变更

### 4.1 InnerAPI 接口行为变化

| 接口 | 变化项 | 说明 |
|------|--------|------|
| `GET /inner-api/v1/configs/mod-api-key` | 导出内容精简 | 移除无限配额计划，修复 token 耗尽状态误判。 |
| `GET /inner-api/v1/configs/ai-route` | 导出内容增加 | API-Key 绑定列表包含祖先 Entity 的路由表。 |
| `GET /inner-api/v1/configs/rate-limit-policy` | 导出内容精简 | 仅导出启用状态的限流策略及绑定。 |

### 4.2 OpenAPI 接口变化

无。

---

## 5. 关键业务逻辑/流程变更

### 5.1 Entity 层级继承一致性

修复后，API-Key 在导出时对所有层级继承的资源采取统一策略：

| 资源类型 | 收集范围 | 过滤条件 |
|----------|----------|----------|
| `allow_models` | API-Key + 所有祖先 Entity | 取非空非 `*` 的交集 |
| `block_models` | 所有祖先 Entity | 取非空非 `*` 的并集 |
| `quota_plan` | API-Key + 所有祖先 Entity | 跳过 `unlimited=true` |
| `rate_limit_policy` | API-Key + 所有祖先 Entity | 跳过 `enabled=false` |
| `route_rules` | API-Key + 所有祖先 Entity + Global | 跳过 `enabled=false` |

### 5.2 配置级联与隔离

本次修复不改变 API-Key / Entity 的 CRUD 级联逻辑，仅改变导出时的收集范围与过滤条件。删除 API-Key 或 Entity 时，原有级联删除逻辑保持不变。

---

## 6. 兼容性/影响分析

### 6.1 对数据面/控制面的影响

| 影响点 | 说明 |
|--------|------|
| BFE `mod_api_key` | 不再收到无限配额计划，token 状态更准确。 |
| BFE `mod_ai_route` | API-Key 绑定列表可能变长（包含祖先 Entity 路由表），BFE 按顺序匹配即可。 |
| BFE 限流模块 | 仅收到启用策略，减少无效配置。 |
| Conf Agent | 通过版本控制机制正常拉取增量配置。 |

### 6.2 对集成测试的影响

| 测试文件 | 调整说明 |
|----------|----------|
| `test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go` | 创建 API-Key 时传入有限 `quota_plan`，确保 `QuotaPlans` 导出非空。 |
| `test/integration/tests/route_tables/list/list_test.go` | 将路由表类型断言由 `"api_key"` 更新为 `"apikey"`。 |
| `test/integration/tests/route_tables/design.md` | 同步更新类型值说明。 |

### 6.3 向后兼容性说明

| 维度 | 兼容性说明 |
|------|------------|
| 数据库 schema | **完全兼容**。无表结构变更。 |
| OpenAPI 接口 | **完全兼容**。无接口变化。 |
| InnerAPI 接口 | **配置内容变化**。对 BFE 是语义修正，建议触发一次全量拉取。 |
| 历史 `route_rules.type="api_key"` | 数据库兼容，但建议统一迁移为 `"apikey"`。 |

### 6.4 风险提示与建议

| 风险点 | 说明 | 建议 |
|--------|------|------|
| 历史 `route_rules.type` 值不一致 | 旧记录可能仍为 `"api_key"` | 提供一次性迁移脚本，或导出逻辑兼容两种值 |
| AI 路由绑定列表变长 | 祖先 Entity 越多，绑定列表越长 | 监控 BFE 配置大小，必要时增加去重/合并 |
| 限流策略禁用后绑定消失 | BFE 侧不再限流，符合预期 | 在 Dashboard 中明确提示“禁用策略不会下发” |
| 无限配额计划不再下发 | BFE 侧不做配额计数 | 确保 BFE 对无配额计划的 token 默认放行 |

---

## 7. 版本修改记录摘要

| 类别 | 变更点 |
|------|--------|
| **API-Key 状态** | `GetRemainingQuota` 增加对 `QuotaPlan.Unlimited` 的判断 |
| **AI 路由导出** | Entity 路由规则由“仅直接挂载”改为“沿 parent_id 向上遍历” |
| **mod-api-key 导出** | 预加载 `QuotaPlan`；跳过无限配额计划 |
| **限流策略导出** | 跳过 `enabled=false` 的策略及绑定 |
| **路由规则常量** | `RouteRulesTypeAPIKey` 由 `"api_key"` 改为 `"apikey"` |

---

## 8. 相关文档索引

| 文档 | 路径 |
|------|------|
| API 变更说明 | `design-docs/modifications/2026-07-28-innerapi-export-fixes/api-changes.md` |
| OpenAPI v0.3.0 设计变更 | `design-docs/modifications/2026-07-27-openapi-optimize/design-changes.md` |
| InnerAPI 配置导出与版本控制 | `design-docs/sys-design/details/InnerAPI配置导出与版本控制.md` |
| 限流策略与导出 | `design-docs/sys-design/details/限流策略与导出.md` |
| 路由规则管理 | `design-docs/sys-design/details/路由规则管理.md` |
| API-Key 与 Entity 关联及模型继承 | `design-docs/sys-design/details/API-Key与Entity关联及模型继承.md` |
| 配额余额同步机制 | `design-docs/sys-design/details/配额余额同步机制.md` |

---

*文档生成日期：2026-07-28*
*目标版本：OpenAPI v0.3.0 热修复*
