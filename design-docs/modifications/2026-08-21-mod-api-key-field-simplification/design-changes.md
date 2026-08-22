# ai-gateway-api mod-api-key 字段精简与 TagLevel 扩展

## 1. 背景

BFE 的 `token_rule.data` 配置已完成字段精简：

- **QuotaPlan**：删除 `CreateTime`、`ResetMode`。
- **Token**：删除 `status`、`update_time`；`enabled` 类型从 `int` 调整为 `bool`。
- **ApikeyTag**：新增 `TagLevel` 字段，取值为 1~5 的整数。

`ai-gateway-api` 作为 `/configs/mod-api-key` 接口的实现方，需要同步调整导出数据结构和生成逻辑，确保响应与 BFE 配置文档保持一致。

---

## 2. 目标

1. `/configs/mod-api-key` 接口返回的 Token 配置中删除 `status`、`update_time`。
2. `/configs/mod-api-key` 接口返回的 Token 配置中 `enabled` 字段类型从 `int` 改为 `bool`。
3. `/configs/mod-api-key` 接口返回的 QuotaPlan 配置中删除 `CreateTime`、`ResetMode`。
4. `/configs/mod-api-key` 接口返回的 Tags 中新增 `TagLevel` 字段。
5. 保持现有 API-Key 启用/禁用、过期判断逻辑不变。
6. 单元测试、集成测试与设计文档同步更新。

---

## 3. 方案概述

### 3.1 TokenFile 结构变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前结构：

```go
type TokenFile struct {
    Key            string  `json:"key"`
    KeyID          string  `json:"key_id"`
    Enabled        int     `json:"enabled"`
    Status         int     `json:"status"`
    UpdateTime     int64   `json:"update_time"`
    ExpiredTime    int64   `json:"expired_time"`
    UnlimitedQuota bool    `json:"unlimited_quota"`
    Models         *string `json:"allow_models"`
    BlockModels    *string `json:"block_models"`
    Subnet         *string `json:"subnet"`
    Tags           []ApikeyTag
    QuotaPlans     []string `json:"quota_plans"`
    models         []string
    blockModels    []string
    subnet         []*net.IPNet
}
```

改造后结构：

```go
type TokenFile struct {
    Key            string  `json:"key"`
    KeyID          string  `json:"key_id"`
    Enabled        bool    `json:"enabled"`
    ExpiredTime    int64   `json:"expired_time"`
    UnlimitedQuota bool    `json:"unlimited_quota"`
    Models         *string `json:"allow_models"`
    BlockModels    *string `json:"block_models"`
    Subnet         *string `json:"subnet"`
    Tags           []ApikeyTag
    QuotaPlans     []string `json:"quota_plans"`
    models         []string
    blockModels    []string
    subnet         []*net.IPNet
}
```

### 3.2 Token 生成逻辑变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前 `APIKeyRuleGenerator` 中构造 `TokenFile` 时同时计算 `Enabled` 和 `Status`：

```go
status := TokenStatusDisabled
var remainQuota float64
if one.Enable != nil && *one.Enable {
    isUnlimited := one.UnlimitedQuota != nil && *one.UnlimitedQuota
    if !isUnlimited {
        status = TokenStatusEnabled
        remainingQuota, err := api_key.GetRemainingQuota(ctx, rlm.quotaCache, one)
        ...
        if remainingQuota != nil {
            remainQuota = *remainingQuota
            if expiredTime != int64(UnlimitedQuota) && time.Now().Local().Unix() >= expiredTime {
                status = TokenStatusExpired
            } else if remainQuota <= 0 {
                status = TokenStatusExhausted
            }
        }
    } else {
        status = TokenStatusEnabled
    }
}

tokenFile := &TokenFile{
    Key:            *one.Key,
    KeyID:          *one.ID,
    Enabled:        2,
    Status:         status,
    ExpiredTime:    expiredTime,
    UpdateTime:     one.KeyCreateAt.Unix(),
    UnlimitedQuota: one.UnlimitedQuota != nil && *one.UnlimitedQuota,
}
if one.Enable != nil && *one.Enable {
    tokenFile.Enabled = 1
}
```

改造后逻辑：

```go
enabled := one.Enable != nil && *one.Enable

// 当 allow_models 交集为空时，强制禁用
if len(finalAllowModels) == 0 {
    modelsStr := ""
    tokenFile.Models = &modelsStr
    enabled = false
}

tokenFile := &TokenFile{
    Key:            *one.Key,
    KeyID:          *one.ID,
    Enabled:        enabled,
    ExpiredTime:    expiredTime,
    UnlimitedQuota: one.UnlimitedQuota != nil && *one.UnlimitedQuota,
}
```

> 说明：
> - `Expired` 状态由 BFE 根据 `expired_time` 实时判断，无需在导出时计算 `status`。
> - `Exhausted` 状态由 BFE 认证阶段根据 Redis 配额余额实时判断，无需在导出时计算 `status`。
> - `UpdateTime` 不再导出。

### 3.3 QuotaPlan 结构变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前结构：

```go
type QuotaPlan struct {
    Id          string
    Unlimited   bool
    PassNoQuota bool
    RedisKey    string
    CreateTime  int64
    ExpiredTime int64
    Quota       int64
    ResetMode   int
    Unit        string
}
```

改造后结构：

```go
type QuotaPlan struct {
    Id          string
    Unlimited   bool
    PassNoQuota bool
    RedisKey    string
    ExpiredTime int64
    Quota       int64
    Unit        string
}
```

`convertQuotaPlanToExport` 函数同步删除 `ResetMode` 与 `CreateTime` 赋值逻辑：

```go
func convertQuotaPlanToExport(qp *quota.QuotaPlanParam, id string, redisKeyID string) *QuotaPlan {
    result := &QuotaPlan{
        Id:          id,
        RedisKey:    fmt.Sprintf("QUOTA_%s", redisKeyID),
        Unlimited:   qp.Unlimited != nil && *qp.Unlimited,
        PassNoQuota: qp.PassWhenNoEnoughQuota != nil && *qp.PassWhenNoEnoughQuota,
        ExpiredTime: -1,
    }
    if qp.Quota != nil {
        result.Quota = golibquota.PtrToRedisValue(qp.Quota, qp.Unit)
    }
    if qp.Unit != nil {
        result.Unit = *qp.Unit
    }
    return result
}
```

### 3.4 ApikeyTag 结构变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前结构：

```go
type ApikeyTag struct {
    TagName  string
    TagValue string
}
```

改造后结构：

```go
type ApikeyTag struct {
    TagName  string
    TagValue string
    TagLevel int
}
```

#### 3.4.1 当前 `ApikeyTag` 的填充方式

`ApikeyTag` 在 `APIKeyRuleGenerator` 中通过 `fetchQuotaPlansWithEntityHierarchy` → `fetchEntityQuotaPlanHierarchy` 递归填充，当前仅使用 Entity 的 `Type` 和 `Name`：

```go
func (rlm *APIKeyRuleManager) fetchEntityQuotaPlanHierarchy(ctx context.Context, entity *entity.EntityParam, productName string, collectedQuotaPlans map[string][]*QuotaPlan) ([]string, []ApikeyTag, error) {
    quotaPlanIDs := make([]string, 0)
    tags := make([]ApikeyTag, 0)

    if entity.EntityID != nil && entity.Type != nil && entity.Name != nil {
        tags = append(tags, ApikeyTag{
            TagName:  *entity.Type,
            TagValue: *entity.Name,
        })
    }

    // ... 处理 quota plan 和父级 entity

    return quotaPlanIDs, tags, nil
}
```

即：`Tags` 的数据来源是 API-Key 关联的 Entity 及其父级 Entity 链上的 `Type` + `Name`。

#### 3.4.2 `TagLevel` 的数据来源

`TagLevel` 应取值为 Entity 所属 **EntityType 的 `Level`**。`ai-gateway-api` 中已有 `EntityType` 模型，包含 `Level` 字段：

```go
// ai-gateway-api/model/entity/entity_type.go
type EntityTypeParam struct {
    TypeName    *string `json:"type_name"`
    Description *string `json:"description"`
    Level       *int    `json:"level"`
    CreateTime  *int64  `json:"create_time,omitempty"`
    UpdateTime  *int64  `json:"update_time,omitempty"`
}
```

因此，填充 `ApikeyTag.TagLevel` 时，应根据 `entity.Type` 查询对应的 `EntityType`，再取其 `Level`。

#### 3.4.3 `TagLevel` 填充实现

**步骤 1：为 `APIKeyRuleManager` 引入 `EntityTypeStorager`**

`ai-gateway-api/model/imods/mod_api_key_rule.go`：

```go
type APIKeyRuleManager struct {
    txn                   itxn.TxnStorager
    versionControlManager *iversion_control.VersionControlManager
    apiKeyStorager        api_key.APIKeyStorager
    aiRouteStorager       iai_route.AIRouteRuleStorager
    quotaPlanStorager     quota.QuotaPlanStorager
    entityStorager        entity.EntityStorager
    entityTypeStorager    entity.EntityTypeStorager // 新增
    quotaCache            quotacache.QuotaCache
}

func NewAPIKeyRuleManager(txn itxn.TxnStorager,
    versionControlManager *iversion_control.VersionControlManager,
    apiKeyStorager api_key.APIKeyStorager,
    aiRouteStorager iai_route.AIRouteRuleStorager,
    quotaPlanStorager quota.QuotaPlanStorager,
    entityStorager entity.EntityStorager,
    entityTypeStorager entity.EntityTypeStorager, // 新增
    quotaCache quotacache.QuotaCache) *APIKeyRuleManager {
    return &APIKeyRuleManager{
        // ...
        entityTypeStorager: entityTypeStorager,
        // ...
    }
}
```

**步骤 2：容器初始化注入 `EntityTypeStorager`**

`ai-gateway-api/stateful/container/rdb/components.go` 在构造 `APIKeyRuleManager` 时传入 `container.EntityTypeStorager`。

**步骤 3：在 `fetchEntityQuotaPlanHierarchy` 中查询并填充 `TagLevel`**

```go
func (rlm *APIKeyRuleManager) fetchEntityQuotaPlanHierarchy(ctx context.Context, entity *entity.EntityParam, productName string, collectedQuotaPlans map[string][]*QuotaPlan) ([]string, []ApikeyTag, error) {
    quotaPlanIDs := make([]string, 0)
    tags := make([]ApikeyTag, 0)

    if entity.EntityID != nil && entity.Type != nil && entity.Name != nil {
        tag := ApikeyTag{
            TagName:  *entity.Type,
            TagValue: *entity.Name,
        }

        // 查询 entity type 对应的 level
        // 按业务约束，每个 Entity 的 Type 必然对应一个存在且 Level 有效的 EntityType
        entityType, err := rlm.entityTypeStorager.FetchEntityType(ctx, &entity.EntityTypeFilter{TypeName: entity.Type})
        if err != nil {
            return nil, nil, err
        }
        if entityType == nil || entityType.Level == nil {
            return nil, nil, fmt.Errorf("entity type %s not found or level invalid", *entity.Type)
        }
        tag.TagLevel = *entityType.Level

        tags = append(tags, tag)
    }

    // ... 处理 quota plan 和父级 entity

    return quotaPlanIDs, tags, nil
}
```

> 说明：
> - `TagLevel` 取值为 1~5 的整数，对应 `entity_types` 表中 `level` 字段。
> - 按业务约束，每个 Entity 的 Type 必然对应存在且 `Level` 有效的 EntityType；若查询不到或 Level 为空，按异常处理并返回错误。
> - 父级 Entity 的 TagLevel 同样按其自身 Type 对应的 EntityType.Level 填充。

### 3.5 状态常量清理

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前定义：

```go
const (
    TokenStatusEnabled   = 1
    TokenStatusDisabled  = 2
    TokenStatusExpired   = 3
    TokenStatusExhausted = 4
)
```

由于 `status` 字段删除，这些常量应同步删除。若其他模块引用了这些常量，需一并清理。

---

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `ai-gateway-api/model/imods/exporter.go` | `TokenFile` 删除 `Status`、`UpdateTime`，`Enabled` 改为 `bool`；`QuotaPlan` 删除 `CreateTime`、`ResetMode`；`ApikeyTag` 新增 `TagLevel`；删除 `TokenStatus*` 常量；调整生成逻辑 |
| `ai-gateway-api/model/imods/mod_api_key_rule.go` | `APIKeyRuleManager` 新增 `entityTypeStorager`，`NewAPIKeyRuleManager` 新增参数 |
| `ai-gateway-api/model/imods/exporter_test.go` | 删除 `Status`、`UpdateTime`、`CreateTime`、`ResetMode` 相关断言；`Enabled` 改为 `bool` 断言；补充 `TagLevel` 断言；mock `EntityTypeStorager` |
| `ai-gateway-api/model/imods/mod_api_key_rule_test.go` | 删除 `QuotaPlan.ResetMode`/`CreateTime` 相关断言；适配 `NewAPIKeyRuleManager` 新签名 |
| `ai-gateway-api/stateful/container/rdb/components.go` | 构造 `APIKeyRuleManager` 时传入 `EntityTypeStorager` |
| `ai-gateway-api/endpoints/innerapi_v1/mod_api_key/export_test.go` | endpoint 层测试断言 `enabled` 为 `bool`，不再断言 `status`/`update_time` |
| `ai-gateway-api/test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go` | 集成测试同步调整字段断言 |
| `ai-gateway-api/design-docs/api-define/InnerAPI接口定义/mod-api-key.md` | 已同步更新字段说明与示例 JSON |
| `ai-gateway-api/design-docs/modifications/2026-08-21-mod-api-key-field-simplification/design-changes.md` | 本设计变更文档（即本文档） |

---

## 5. 测试计划

### 5.1 单元测试

**文件：** `ai-gateway-api/model/imods/exporter_test.go`

1. 删除所有对 `TokenFile.Status` 的断言。
2. 删除所有对 `TokenFile.UpdateTime` 的断言。
3. 将 `Enabled` 断言从 `assert.Equal(t, 1, token.Enabled)` 改为 `assert.True(t, token.Enabled)`。
4. 将禁用场景断言从 `assert.Equal(t, 2, token.Enabled)` 改为 `assert.False(t, token.Enabled)`。
5. 删除 `QuotaPlan.ResetMode` 与 `CreateTime` 断言。
6. 补充 `ApikeyTag.TagLevel` 断言。

### 5.2 Endpoint 测试

**文件：** `ai-gateway-api/endpoints/innerapi_v1/mod_api_key/export_test.go`

1. 断言反序列化后的 token `enabled` 为 `bool` 类型。
2. 断言响应中不再包含 `status`、`update_time`。
3. 断言 QuotaPlan 中不再包含 `CreateTime`、`ResetMode`。
4. 断言 Tags 中包含 `TagLevel`。

### 5.3 集成测试

**文件：** `ai-gateway-api/test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go`

1. 调用 `/inner-api/v1/configs/mod-api-key` 后，断言 token 字段与文档一致。
2. 断言 QuotaPlan 字段与文档一致。

### 5.4 验证命令

```bash
cd ai-gateway-api

go test ./model/imods/...
go test ./endpoints/innerapi_v1/mod_api_key/...
go test ./test/integration/tests/innerapi/mod_api_key/...
```

---

## 6. 兼容性说明

- `/configs/mod-api-key` 响应删除 `status`、`update_time`、`CreateTime`、`ResetMode`，BFE 侧已删除对这些字段的依赖。
- `/configs/mod-api-key` 响应中 `enabled` 类型从 `int` 改为 `bool`，BFE 侧 `TokenFile.Enabled` 需同步改为 `bool`。
- `/configs/mod-api-key` 响应中 Tags 新增 `TagLevel`，BFE 侧 `ApikeyTag` 需同步新增该字段。
- 数据库 schema 不变；`ResetMode`/`CreateTime`/`status`/`update_time` 相关数据库字段可继续保留，仅不再导出。

---

## 7. 风险与缓解

| 风险 | 缓解措施 |
|------|---------|
| `enabled` 从 `int` 改为 `bool` 可能导致 BFE 旧版本解析失败 | 确保 BFE 与 ai-gateway-api 同步升级；旧配置中 JSON 的 `true`/`false` 对 Go `bool` 可正确解析，但 BFE 旧代码若仍按 `int` 解析会报错 |
| `status` 删除后，BFE 侧状态判断逻辑未迁移完成 | 参考 BFE 修改方案 `bfe/docs/zh_cn/modifications/2026-08-21-token-rule-data-field-simplification/design-changes.md` 完成 `enabled`/`expired_time`/实时配额检查迁移 |
| `TagLevel` 需引入 EntityType 查询 | 为 `APIKeyRuleManager` 注入 `EntityTypeStorager`，在导出时按需查询；按业务约束 EntityType 及其 Level 必然存在，异常时返回错误 |
| 外部系统仍依赖 `status`/`update_time` | 全局搜索 `/configs/mod-api-key` 消费者，同步通知改造 |

---

## 8. 关键代码索引

| 文件 | 行号范围 | 说明 |
|---|---|---|
| `ai-gateway-api/model/imods/exporter.go` | 37-42 | `TokenStatus*` 常量定义 |
| `ai-gateway-api/model/imods/exporter.go` | 57-60 | `ApikeyTag` 结构定义 |
| `ai-gateway-api/model/imods/exporter.go` | 66-76 | `QuotaPlan` 结构定义 |
| `ai-gateway-api/model/imods/exporter.go` | 78-94 | `TokenFile` 结构定义 |
| `ai-gateway-api/model/imods/exporter.go` | 235-273 | Token `status`/`enabled` 生成逻辑 |
| `ai-gateway-api/model/imods/exporter.go` | 576-603 | `convertQuotaPlanToExport` 函数 |
