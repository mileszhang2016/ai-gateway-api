# Provider 全量名称查询接口：设计变更说明

## 1. 概述

### 1.1 背景

`GET /providers` 为分页列表接口，返回完整 Provider 对象。Issue #87 需要一种轻量方式获取全量 provider 名称，用于下拉选择、自动补全等场景。

### 1.2 目标

新增无状态、轻量的 `GET /providers/actions/get-provider-names` 接口，仅返回 `name` 列表。

### 1.3 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `endpoints/openapi_v1/providers`、`model/iprovider`、`storage/rdb/provider` |
| 涉及文档 | `design-docs/api-define/OpenAPI接口定义/providers.md` |
| 数据迁移 | 无 |

---

## 2. 接口设计

### 2.1 端点与 Method

| 项目 | 值 |
|------|-----|
| 端点 | `/providers/actions/get-provider-names` |
| method | GET |
| 版本 | v1 |

### 2.2 请求参数

无。

### 2.3 响应参数

Data：

| 参数名 | 类型 | 参数含义 |
|--------|------|----------|
| `names` | []string | Provider 名称列表，字典序升序 |

---

## 3. 关键设计决策

| 决策 | 说明 |
|------|------|
| action 风格端点 | 与 `/model-prices/actions/get-providers` 保持一致。 |
| 仅查询 name 字段 | 避免加载 `instance_pool`、`keys` 等大 JSON 字段，降低开销。 |
| 不暴露 id | 调用方通常只需要可读名称用于引用。 |
| 排序由服务层完成 | 数据库查询后由 Go 排序，跨 MySQL/SQLite 行为一致。 |

---

## 4. 实现要点

### 4.1 Handler 层

- 新增 `endpoints/openapi_v1/provider/list_names.go`（或合并到现有 `list.go`）。
- 绑定路由 `/providers/actions/get-provider-names`。
- 响应结构：

  ```go
  type getProviderNamesResponse struct {
      Names []string `json:"names"`
  }
  ```

### 4.2 Model 层

- `ProviderStorager` 新增：

  ```go
  FetchProviderNames(ctx context.Context) ([]string, error)
  ```

- `ProviderManager.ListProviderNames`：
  - 调用 `storager.FetchProviderNames`。
  - 对结果排序（若 DAO 未排序）。
  - 返回名称列表。

### 4.3 Storage 层

- `storage/rdb/provider/` DAO 实现 `FetchProviderNames`。
- SQL 仅查询 `name` 字段，如：

  ```sql
  SELECT name FROM providers ORDER BY name ASC
  ```

  或在 Go 中统一排序以兼容 SQLite/MySQL 差异。

### 4.4 路由注册

在 `endpoints/openapi_v1/endpoints.go` 的 `Endpoints` 切片中加入新端点。

---

## 5. 待确认事项

| 事项 | 建议 |
|------|------|
| 排序规则 | 采用字典序升序，与 model-prices 保持一致。 |
| 字段名 | 响应字段使用 `names`，与端点 `get-provider-names` 对应。 |
| 权限 | 复用 `FeatureProvider + ActionRead`。 |
