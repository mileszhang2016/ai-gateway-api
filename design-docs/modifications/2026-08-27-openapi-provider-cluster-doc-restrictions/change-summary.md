# OpenAPI Provider/Cluster 文档与代码对齐（2026-08-27）

## 1. 背景

- Issue #103：`PATCH /open-api/v1/providers/{provider_name}` 在请求体未传 `name` 时返回 422，与接口文档示例不一致。
- `clusters.md` 2.4 节未明确说明更新接口不可修改 `name`。
- `clusters.md` 2.5 节未说明删除集群前会检查 AI 路由规则引用。

本次变更同步更新 OpenAPI 接口定义文档与对应代码实现，使文档描述与实际行为一致。

## 2. 目标

明确并落地以下行为，避免调用方误用：

1. `PATCH /providers/{provider_name}` 与 `PATCH /clusters/{cluster_name}` 的请求体**不包含 `name`**，即不能通过更新接口修改资源名称；若请求体中仍包含 `name`，返回 422。
2. `DELETE /clusters/{cluster_name}` 在级联删除前，会先检查该集群是否被 AI 路由规则引用；若被引用则删除失败。
3. 文档聚焦 AI 网关常用场景，仅向调用方描述 AI 路由规则引用检查；代码层保留产品级路由规则检查作为防御性校验。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及文档 | `design-docs/api-define/OpenAPI接口定义/providers.md`<br>`design-docs/api-define/OpenAPI接口定义/clusters.md`<br>`test/integration/tests/provider/design.md`<br>`test/integration/tests/clusters/design.md` |
| 涉及代码 | `endpoints/openapi_v1/provider/update.go`<br>`endpoints/openapi_v1/product_cluster/update_basic.go`<br>`test/integration/tests/provider/update/update_test.go`<br>`test/integration/tests/clusters/update/update_test.go` |
| 数据迁移 | 无 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 名称由 URI 决定 | Provider/Cluster 的 `name` 由路径参数指定；Handler 在绑定 JSON 前检查请求体是否包含 `name`，若包含则返回 422；否则使用 URI 中的名称填充 `param.Name`。 |
| 删除前检查 AI 路由规则 | 集群删除逻辑保留 `RouteRulesManager.ClusterDeleteChecker` 检查；文档补齐对应说明。 |
| 保留产品级路由规则检查 | `ClusterManager` 仍保留 `RouteRuleManager.ClusterDeleteChecker` 作为防御性校验；因 AI 网关场景不使用产品级路由规则，文档不再向调用方展开说明。 |

## 5. 改动点

| 仓库 | 文件 | 修改内容 |
|------|------|----------|
| `ai-gateway-api` | `design-docs/api-define/OpenAPI接口定义/providers.md` 2.4 节 | 增加限制：输入参数不包括 `name`，不能修改 provider 的 name；若请求体含 `name` 返回 422。 |
| `ai-gateway-api` | `endpoints/openapi_v1/provider/update.go` | `UpdateAction` 在绑定 JSON 前检查请求体是否包含 `name`，若包含则返回 422；否则将 `param.Name` 设置为 URI 中的 `provider_name`。 |
| `ai-gateway-api` | `test/integration/tests/provider/update/update_test.go` | 更新测试请求体不再携带 `name`；新增 `PV-4-005` 验证请求体不包含 `name` 时更新成功；新增 `PV-4-006` 验证请求体包含 `name` 时返回 422。 |
| `ai-gateway-api` | `test/integration/tests/provider/design.md` | 更新请求参数说明：`provider_name` 由 URI 指定，请求体无需传 `name`（含 `name` 返回 422）；增加 `PV-4-005`/`PV-4-006` 用例。 |
| `ai-gateway-api` | `design-docs/api-define/OpenAPI接口定义/clusters.md` 2.4 节 | 增加限制：输入参数不包括 `name`，不能修改 cluster 的 name；若请求体含 `name` 返回 422；同步删除 BODY 示例中的 `name` 字段。 |
| `ai-gateway-api` | `endpoints/openapi_v1/product_cluster/update_basic.go` | `newUpdateParam4Update` 在绑定参数前检查请求体是否包含 `name`，若包含则返回 422；否则将 `param.Name` 设置为 URI 中的 `cluster_name`。 |
| `ai-gateway-api` | `test/integration/tests/clusters/update/update_test.go` | 新增 `CL-4-012` 验证请求体不包含 `name` 时更新成功；新增 `CL-4-013` 验证请求体包含 `name` 时返回 422。 |
| `ai-gateway-api` | `test/integration/tests/clusters/design.md` 9.2.1/9.3 节 | 更新 Body 参数说明：`cluster_name` 由 URI 指定，请求体中无需传 `name`（含 `name` 返回 422）；增加 `CL-4-012`/`CL-4-013` 用例。 |
| `ai-gateway-api` | `design-docs/api-define/OpenAPI接口定义/clusters.md` 2.5 节 | 增加限制：删除前检查是否被 AI 路由规则（global / entity / api-key 级别）引用，若被引用则删除失败。 |
