# OpenAPI 路由规则字段命名规范化：接口变更说明

## 1. 变更概述

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-15 |
| 关联 Issue | <https://github.com/rainway-ai-gateway/ai-gateway-api/issues/51> |
| 影响模块 | `/entities`、`/api-keys`、`/global-route-rules`、`/route-tables` |
| 目标文档 | `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/00-common.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/api-keys.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/entities.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/global-route-rules.md` |

### 1.1 问题描述

OpenAPI 中大部分字段采用小写或下划线命名（如 `inner_id`、`allow_models`、`quota_plan`、`rate_limit_policy`、`route_rules` 等），但 `route_rules.rules[*]` 内部的字段使用了大写开头或驼峰命名，风格不一致：

- `Cond` → 应改为 `cond`
- `ClusterName` → 应改为 `cluster_name`
- `Model` → 应改为 `model`
- `Weight` → 应改为 `weight`

### 1.2 变更目标

统一 `route_rules.rules[*]` 及其子结构的 JSON 字段命名风格，使其与 OpenAPI 其他字段保持一致（小写/下划线）。

### 1.3 影响接口

| 接口 | Method | 影响范围 |
|------|--------|----------|
| `/entities` | GET / POST / PUT / PATCH | 响应中 `route_rules.rules[*]` 内的字段 |
| `/api-keys` | GET / POST / PUT / PATCH | 响应中 `route_rules.rules[*]` 内的字段 |
| `/global-route-rules` | GET / PUT | 响应中 `rules[*]` 内的字段 |
| `/route-tables` | GET / LIST | 列表中各路由表的 `rules` 内容（只读展示） |

> 说明：`/route-tables` 本身仅用于查询路由表元信息，不直接返回完整规则内容；若后续扩展返回规则详情，字段风格需保持一致。

---

## 2. 变更内容

### 2.1 公共类型 `RouteRule`（定义于 `00-common.md`）

#### 变更前

```json
{
  "name": "rule_name",
  "Cond": "default_t()",
  "targets": [
    {
      "ClusterName": "cluster_default",
      "Model": "",
      "Weight": 100
    }
  ],
  "fallbacks": [
    {
      "ClusterName": "cluster_fallback",
      "Model": ""
    }
  ]
}
```

#### 变更后

```json
{
  "name": "rule_name",
  "cond": "default_t()",
  "targets": [
    {
      "cluster_name": "cluster_default",
      "model": "",
      "weight": 100
    }
  ],
  "fallbacks": [
    {
      "cluster_name": "cluster_fallback",
      "model": ""
    }
  ]
}
```

#### 字段映射表

| 原字段名 | 新字段名 | 说明 |
|----------|----------|------|
| `Cond` | `cond` | 路由条件表达式，如 `default_t()` |
| `ClusterName` | `cluster_name` | 目标集群名称 |
| `Model` | `model` | 目标模型名称，可为空 |
| `Weight` | `weight` | 目标权重，范围 `[0, 100]`，同一规则下 targets 权重之和须为 100 |

### 2.2 `/entities` 接口

`route_rules` 字段内的 `rules[*]` 子结构字段按 [2.1 公共类型 `RouteRule`](#21-公共类型-routerule) 变更。

涉及位置：

- `POST /entities` 响应示例
- `GET /entities/{id}` 响应示例
- `PUT /entities/{id}` 响应示例
- `PATCH /entities/{id}` 响应示例

### 2.3 `/api-keys` 接口

`route_rules` 字段内的 `rules[*]` 子结构字段按 [2.1 公共类型 `RouteRule`](#21-公共类型-routerule) 变更。

涉及位置：

- `POST /api-keys` 响应示例
- `GET /api-keys` 列表/详情响应示例
- `PUT /api-keys/{id}` 响应示例
- `PATCH /api-keys/{id}` 响应示例

### 2.4 `/global-route-rules` 接口

`rules[*]` 子结构字段按 [2.1 公共类型 `RouteRule`](#21-公共类型-routerule) 变更。

涉及位置：

- `GET /global-route-rules` 响应示例
- `PUT /global-route-rules` 请求/响应示例

### 2.5 公共命名规范

在 `00-common.md` 中新增「字段命名规范」章节，明确：

- OpenAPI JSON 字段统一使用小写 + 下划线（snake_case）。
- 已全局固化的响应包装字段（`ErrNum`、`ErrMsg`、`Data`）维持不变。
- 因外部系统强制要求而必须使用非 snake_case 命名的字段，须显式标注为例外。
- 新增字段须先核对本规范，防止再次出现类似 Issue #51 的不一致问题。

---

## 3. 兼容性说明

### 3.1 请求侧兼容

OpenAPI 服务端在反序列化时同时接受新旧两种 key（即 `Cond` 与 `cond`、`ClusterName` 与 `cluster_name` 等均可被识别），以降低旧客户端迁移成本。新写入/更新时统一按新字段名处理。

### 3.2 响应侧变化

服务端响应**仅输出新字段名**（小写/下划线），不再输出旧的大写/驼峰字段名。调用方需在读取响应时适配新字段名。

### 3.3 存量数据

`route_rules` 在数据库中以 JSON 字符串形式存储。由于服务端反序列化同时兼容新旧 key，存量数据（旧字段名）仍可正常读取；数据发生更新后，会以新字段名重新写入。

### 3.4 不兼容范围

- InnerAPI 导出给 BFE 的 `AiRouteDataExport` 保持原有字段名（`Cond`、`ClusterName`、`Model`、`Weight`）不变，避免影响 BFE 消费侧。
- `alb-pool.md` 中 `ports` 的 `Default` 键、clusters 的 `Authorization` 值等不属于本变更范围。

---

## 4. 验证建议

1. 通过 `GET /global-route-rules` 检查响应中 `rules[*].cond`、`rules[*].targets[*].cluster_name` 等字段已变为小写。
2. 通过 `PUT /global-route-rules` 使用新字段名提交，确认可正常创建/更新。
3. 通过 `GET /entities/{id}`、`GET /api-keys/{id}` 检查关联的 `route_rules` 字段风格一致。
4. 使用旧字段名（`Cond`、`ClusterName` 等）提交请求，确认服务端仍可解析（兼容期内）。
