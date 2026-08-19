# ai-gateway-api mod-api-key key_id 字段设计变更

## 1. 背景

BFE 的 `token_rule.data` 配置已调整为：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | Y | API-Key 值 |
| `key_id` | string | Y | API-Key 标识 ID，用于唯一标识该 API-Key |

不再包含 `name` 字段。

`ai-gateway-api/design-docs/api-define/InnerAPI接口定义/mod-api-key.md` 需要与 BFE 保持一致：
- 新增 `key_id` 字段
- 删除 `name` 字段

`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/api-keys.md` 中 API-Key 的数据模型已有 `id` 字段，该字段将直接映射为 BFE 的 `key_id`。

## 2. 目标

1. `/configs/mod-api-key` 接口返回的 Token 配置中新增 `key_id` 字段。
2. `/configs/mod-api-key` 接口返回的 Token 配置中不再包含 `name` 字段。
3. 保持现有配额计划（QuotaPlan）的 ID 和 RedisKey 生成逻辑不变，继续使用 API-Key 的 `key` 值。
4. 数据库 schema 不变。
5. 单元测试和集成测试同步更新。

## 3. 方案概述

### 3.1 TokenFile 结构变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

在 `TokenFile` 结构体中新增 `KeyID` 字段，删除 `Name` 字段：

```go
type TokenFile struct {
    Key            string  `json:"key"`
    KeyID          string  `json:"key_id"`      // 新增
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
    // ... unexported fields
}
```

### 3.2 Token 生成逻辑变更

**文件：** `ai-gateway-api/model/imods/exporter.go`

在 `APIKeyRuleGenerator` 中构造 `TokenFile` 时：

```go
tokenFile := &TokenFile{
    Key:            *one.Key,
    KeyID:          *one.ID,                       // 新增
    Enabled:        2,
    Status:         status,
    ExpiredTime:    expiredTime,
    UpdateTime:     one.KeyCreateAt.Unix(),
    UnlimitedQuota: one.UnlimitedQuota != nil && *one.UnlimitedQuota,
}
```

### 3.3 关于 QuotaPlan ID 和 RedisKey 的说明

当前代码中，API-Key 自身关联的 QuotaPlan 使用 `*apiKey.Key` 作为 `Id` 和 `RedisKey` 的标识：

```go
qp := convertQuotaPlanToExport(quotaPlan, *apiKey.Key, *apiKey.Key)
```

文档示例中显示 `quota_plans: ["ak-test-key-001", ...]`，看起来与 `key_id` 相同。但这只是示例约定，BFE 实际消费时只要求 `quota_plans` 中的 ID 能匹配到顶层 `QuotaPlans` 列表中的某个 `Id` 即可。

**本方案保持 QuotaPlan ID 和 RedisKey 继续使用 `key` 值不变**，原因：
- 避免同时改动配额缓存键（`model/api_key/api_key.go` 中 `quotaCache.GetRemaining/SetRemaining` 使用 `*param.Key`）。
- 避免影响已部署环境的 Redis 数据。
- 文档示例可后续统一调整为使用 `key` 值，或在需要时再切换为 `key_id`。

如果未来决定让 QuotaPlan ID/RedisKey 与 `key_id` 对齐，需要额外修改：
- `model/imods/exporter.go`：`convertQuotaPlanToExport(quotaPlan, *apiKey.ID, *apiKey.ID)`
- `model/api_key/api_key.go`：`quotaCache.GetRemaining/SetRemaining` 的参数从 `*param.Key` 改为 `*param.ID`

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `ai-gateway-api/model/imods/exporter.go` | `TokenFile` 新增 `KeyID`、删除 `Name`；`APIKeyRuleGenerator` 中填充 `KeyID` |
| `ai-gateway-api/model/imods/exporter_test.go` | 单元测试断言 `KeyID`，删除 `Name` 断言 |
| `ai-gateway-api/model/imods/mocks_test.go` | mock/fake 数据中补充 `ID` |
| `ai-gateway-api/endpoints/innerapi_v1/mod_api_key/export_test.go` | endpoint 层测试断言 `key_id` |
| `ai-gateway-api/test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go` | 集成测试断言返回 token 包含 `key_id`，且不包含 `name` |
| `ai-gateway-api/design-docs/api-define/InnerAPI接口定义/mod-api-key.md` | 删除 Token 字段说明中的 `name`，新增 `key_id`；更新示例 JSON |

## 5. 测试计划

### 5.1 单元测试

**文件：** `ai-gateway-api/model/imods/exporter_test.go`

- 在现有的 Token 导出测试中，断言 `TokenFile.KeyID` 等于 API-Key 的 `ID`。
- 删除所有对 `TokenFile.Name` 的断言。

### 5.2 Endpoint 测试

**文件：** `ai-gateway-api/endpoints/innerapi_v1/mod_api_key/export_test.go`

- 在反序列化后的响应中检查 `key_id` 字段存在且非空。
- 检查响应中不再包含 `name` 字段。

### 5.3 集成测试

**文件：** `ai-gateway-api/test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go`

- 调用 `/inner-api/v1/configs/mod-api-key` 后，断言返回的 token 包含 `key_id`。
- 断言返回的 token 不包含 `name` 字段。

### 5.4 验证命令

```bash
cd ai-gateway-api

go test ./model/imods/...
go test ./endpoints/innerapi_v1/mod_api_key/...
go test ./test/integration/tests/innerapi/mod_api_key/...
```

## 6. 兼容性说明

- `/configs/mod-api-key` 响应新增 `key_id` 字段，属于新增字段，BFE 侧已支持解析。
- `/configs/mod-api-key` 响应删除 `name` 字段，BFE 侧已删除对 `name` 的依赖。
- 数据库 schema 不变。

## 7. 后续可选扩展

如果未来需要让 QuotaPlan ID/RedisKey 与 `key_id` 完全对齐：

1. 将 `model/imods/exporter.go` 中 API-Key 自身 QuotaPlan 的 ID/RedisKey 从 `*apiKey.Key` 改为 `*apiKey.ID`。
2. 同步修改 `model/api_key/api_key.go` 中配额缓存的读写键为 `*param.ID`。
3. 更新 `db_ddl.sql`/`db_ddl_sqlite.sql` 中的索引/约束（如有依赖 `key` 的配额相关索引）。
4. 更新相关设计文档中的 Redis Key 约定。

本阶段不做上述扩展，保持改动范围清晰可控。
