# Provider 全量名称查询接口变更摘要

## 1. 背景

Issue #87：`/open-api/v1/providers` 为分页接口，调用方无法一次性获取全量 provider 名称列表。

## 2. 目标

新增专用接口 `GET /providers/actions/get-provider-names`，返回所有 Provider 的 `name` 列表，按字典序升序排列。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及文档 | `design-docs/api-define/OpenAPI接口定义/providers.md`、`design-docs/sys-design/接口层设计文档.md` |
| 涉及模块 | `endpoints/openapi_v1/providers`、`model/iprovider`、`storage/rdb/provider` |
| 数据迁移 | 无 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 独立 action 接口 | 不改造现有分页列表接口，避免破坏其语义与调用方。 |
| 仅返回名称 | 最小化响应体，避免加载 provider JSON 大字段。 |
| 字典序升序 | 与 `/model-prices/actions/get-providers` 保持一致，便于前端展示。 |
| 无查询参数 | 首期不提供过滤，直接返回全量。 |

## 5. 改动点

| 仓库 | 文件/模块 | 修改内容 |
|------|-----------|----------|
| `ai-gateway-api` | `design-docs/api-define/OpenAPI接口定义/providers.md` | 新增 2.7 节接口定义 |
| `ai-gateway-api` | `design-docs/sys-design/接口层设计文档.md` | 接口清单增加新端点 |
| `ai-gateway-api` | `endpoints/openapi_v1/provider/` | 新增 `get_provider_names` handler |
| `ai-gateway-api` | `endpoints/openapi_v1/endpoints.go` | 注册 `/providers/actions/get-provider-names` 路由 |
| `ai-gateway-api` | `model/iprovider/provider.go` | `ProviderManager` 新增 `ListProviderNames` 方法 |
| `ai-gateway-api` | `model/iprovider/provider.go` | `ProviderStorager` 新增 `FetchProviderNames` 接口方法 |
| `ai-gateway-api` | `storage/rdb/provider/` | DAO 实现 `FetchProviderNames`，仅查询 `name` 字段 |
