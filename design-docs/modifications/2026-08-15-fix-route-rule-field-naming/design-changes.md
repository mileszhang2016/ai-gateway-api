# OpenAPI 路由规则字段命名规范化：设计变更说明

## 1. 概述

### 1.1 变更背景

Issue #51 指出：OpenAPI 中 `route_rules.rules[*]` 及其子结构存在字段命名风格不一致。大部分字段采用小写/下划线（如 `inner_id`、`allow_models`、`quota_plan`），但路由规则内部字段使用了大写开头或驼峰命名（`Cond`、`ClusterName`、`Model`、`Weight`）。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-15 |
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `ai-gateway-api/model/shared`、`ai-gateway-api/endpoints/openapi_v1` |
| 变更类型 | 接口字段命名规范化 + 兼容性处理 |

将以下字段统一为小写/下划线风格，并保持 OpenAPI 与 InnerAPI 的边界清晰：

| 原字段名 | 新字段名 |
|----------|----------|
| `Cond` | `cond` |
| `ClusterName` | `cluster_name` |
| `Model` | `model` |
| `Weight` | `weight` |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **OpenAPI 风格一致** | OpenAPI 所有响应/请求字段统一采用小写/下划线。 |
| **InnerAPI 不变** | 导出给 BFE 的 InnerAPI 配置保持原字段名，避免影响下游消费方。 |
| **向后兼容** | 服务端反序列化同时接受新旧 key，降低旧客户端和存量数据迁移成本。 |
| **最小侵入** | 仅在 `model/shared/types.go` 中调整 JSON tag 和反序列化行为，不引入新的 DTO。 |
| **规范先行** | 在 `00-common.md` 中增加字段命名规范，防止后续新增字段再次出现风格不一致。 |

---

## 2. 数据模型变更

### 2.1 `ai-gateway-api` 侧：共享结构体

变更位置：`ai-gateway-api/model/shared/types.go`

#### 变更前

```go
type AiRouteTargetParam struct {
    ClusterName *string `json:"ClusterName"`
    Model       *string `json:"Model"`
    Weight      *int    `json:"Weight"`
}

type AiRouteFallbackParam struct {
    ClusterName *string `json:"ClusterName"`
    Model       *string `json:"Model"`
}

type AiRouteRuleParam struct {
    Name      *string                 `json:"name"`
    Cond      *string                 `json:"Cond"`
    Targets   []*AiRouteTargetParam   `json:"targets"`
    Fallbacks []*AiRouteFallbackParam `json:"fallbacks"`
}
```

#### 变更后

```go
type AiRouteTargetParam struct {
    ClusterName *string `json:"cluster_name"`
    Model       *string `json:"model"`
    Weight      *int    `json:"weight"`
}

type AiRouteFallbackParam struct {
    ClusterName *string `json:"cluster_name"`
    Model       *string `json:"model"`
}

type AiRouteRuleParam struct {
    Name      *string                 `json:"name"`
    Cond      *string                 `json:"cond"`
    Targets   []*AiRouteTargetParam   `json:"targets"`
    Fallbacks []*AiRouteFallbackParam `json:"fallbacks"`
}
```

### 2.2 兼容性反序列化

由于 `shared.AiRouteRuleParam` 及其子结构同时被以下两处使用：

1. **OpenAPI 端点**：请求体绑定、响应序列化。
2. **存储层**：`storage/rdb/quota/route_rules.go` 通过 `json.Marshal/Unmarshal` 将规则数组持久化到 `route_rules.rules` 字段。

若仅修改 JSON tag，新写入会使用新字段名，但存量数据（旧字段名）将无法正确反序列化。因此为以下三个结构体增加自定义 `UnmarshalJSON`：

- `AiRouteTargetParam`
- `AiRouteFallbackParam`
- `AiRouteRuleParam`

自定义反序列化逻辑：

- 优先读取新字段名（`cond`、`cluster_name`、`model`、`weight`）。
- 当新字段名为空且旧字段名（`Cond`、`ClusterName`、`Model`、`Weight`）存在时，使用旧字段名赋值。
- `MarshalJSON` 不自定义，继续使用标准库，仅输出新字段名。

实现模式示例：

```go
func (t *AiRouteTargetParam) UnmarshalJSON(data []byte) error {
    type Alias AiRouteTargetParam
    aux := &struct {
        *Alias
        OldClusterName *string `json:"ClusterName"`
        OldModel       *string `json:"Model"`
        OldWeight      *int    `json:"Weight"`
    }{
        Alias: (*Alias)(t),
    }
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    if t.ClusterName == nil && aux.OldClusterName != nil {
        t.ClusterName = aux.OldClusterName
    }
    if t.Model == nil && aux.OldModel != nil {
        t.Model = aux.OldModel
    }
    if t.Weight == nil && aux.OldWeight != nil {
        t.Weight = aux.OldWeight
    }
    return nil
}
```

> 说明：`AiRouteRuleParam` 和 `AiRouteFallbackParam` 采用同样模式处理 `Cond` / `ClusterName` / `Model`。

---

## 3. 存储层影响

### 3.1 写入行为

`storage/rdb/quota/route_rules.go` 中的 `marshalRouteRules` 使用 `json.Marshal` 序列化 `[]*shared.AiRouteRuleParam`。JSON tag 变更后，新写入数据库的 `rules` 字段将使用小写/下划线 key。

### 3.2 读取行为

`routeRulesDataToParam` 使用 `json.Unmarshal` 反序列化。由于自定义 `UnmarshalJSON` 兼容旧 key，存量数据仍可正常读取。

### 3.3 数据迁移

无需单独数据迁移脚本。存量数据在首次被读取时自动兼容；发生更新时自动按新字段名重写。

---

## 4. 接口层影响

### 4.1 OpenAPI 端点

以下端点通过 `shared.RouteRulesParam` 绑定请求/序列化响应，自动继承新字段名：

- `/open-api/v1/entities`（GET/POST/PUT/PATCH）
- `/open-api/v1/api-keys`（GET/POST/PUT/PATCH）
- `/open-api/v1/global-route-rules`（GET/PUT）

变更后：

- 响应中 `route_rules.rules[*]` 内字段统一为小写/下划线。
- 请求体同时接受新旧字段名（兼容期内）。

### 4.2 InnerAPI 端点

InnerAPI 导出结构位于 `model/imods/ai_route_exporter.go`，与 `shared.AiRouteRuleParam` 是独立的结构体。`convertToRouteTableExport` 使用 Go 结构体字段名（而非 JSON tag）进行赋值，因此 `shared` 结构体的 JSON tag 变更不会影响 InnerAPI 导出。

InnerAPI 继续向 BFE 输出大写/驼峰字段名：

```json
{
  "Cond": "default_t()",
  "ClusterName": "cluster_default",
  "Model": "",
  "Weight": 100
}
```

---

## 5. 测试更新

### 5.1 单元测试

- `ai-gateway-api/endpoints/openapi_v1/global_route_rules/endpoints_test.go`
  - 将 `PUT` 请求体中的 `Cond`、`ClusterName`、`Model`、`Weight` 改为小写。

### 5.2 集成测试

以下集成测试用例中构建请求或断言响应时使用了旧字段名，需要同步替换：

- `test/integration/tests/global_route_rules/get/get_test.go`
- `test/integration/tests/global_route_rules/update/update_test.go`
- `test/integration/tests/api_key/create/create_test.go`
- `test/integration/tests/api_key/partial_update/partial_update_test.go`
- `test/integration/tests/entity/partial_update/partial_update_test.go`
- `test/integration/tests/route_tables/list/list_test.go`
- `test/integration/tests/clusters/update/update_test.go`
- `test/integration/tests/clusters/delete/delete_test.go`

---

## 6. 文档更新

需要同步更新以下 OpenAPI 接口定义文档：

- `design-docs/api-define/OpenAPI接口定义/00-common.md`
  - 新增「字段命名规范」章节。
  - 将 `RouteRule`、`RouteRules` 示例与字段说明中的 `Cond`、`ClusterName`、`Model`、`Weight` 改为小写/下划线。
- `design-docs/api-define/OpenAPI接口定义/api-keys.md`
- `design-docs/api-define/OpenAPI接口定义/entities.md`
- `design-docs/api-define/OpenAPI接口定义/global-route-rules.md`
  - 同步更新示例中的 route rules 字段名。

---

## 7. 风险与回滚

### 7.1 主要风险

- 若跳过自定义 `UnmarshalJSON`，存量 DB 数据在读取时会丢失 `cond`、`cluster_name`、`model`、`weight` 字段内容，导致路由规则失效。
- 若 InnerAPI 导出结构误改，会导致 BFE 无法正确解析路由配置。

### 7.2 规避措施

- 实现兼容反序列化，确保存量数据可读。
- 不改 `model/imods/ai_route_exporter.go` 中的 JSON tag，保持 BFE 侧稳定。
- 变更后优先运行集成测试验证 `/global-route-rules`、`/entities`、`/api-keys` 的读写一致性。

### 7.3 回滚方案

若需要回滚，仅需将 `model/shared/types.go` 中的 JSON tag 恢复为大写/驼峰，并移除自定义 `UnmarshalJSON`。新写入的小写 key 数据在回滚后将无法读取，因此回滚前需评估数据迁移。
