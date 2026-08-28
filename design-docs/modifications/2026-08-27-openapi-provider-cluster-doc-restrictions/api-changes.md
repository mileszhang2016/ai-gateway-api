# OpenAPI Provider/Cluster 文档与代码对齐：API 变更说明

## 1. 变更范围

### 1.1 文档

`design-docs/api-define/OpenAPI接口定义/` 目录下的两份接口定义文档：

- `providers.md` 2.4 节（更新 Provider）
- `clusters.md` 2.4 节（更新集群基本配置）
- `clusters.md` 2.5 节（删除集群）

### 1.2 代码

- `endpoints/openapi_v1/provider/update.go`
- `endpoints/openapi_v1/product_cluster/update_basic.go`

无接口路径、Method、请求/响应字段语义变更。

## 2. `providers.md` 2.4 节与代码变更

### 2.1 文档新增限制

在 **输入参数（Body）** 说明中增加：

> 可修改字段含义同创建接口，但**输入参数不包括 `name`，即不能修改 provider 的 name**（名称由 URI 中的 `provider_name` 指定）。若请求体中仍包含 `name`，返回 422。

### 2.2 代码变更

`endpoints/openapi_v1/provider/update.go` 的 `UpdateAction`：

1. 在绑定 JSON 前，通过 `rejectBodyField(req, "name")` 检查请求体是否包含 `name`；若包含，返回 422。
2. 绑定 JSON 后，将 `param.Name` 设置为 URI 中的 `provider_name`，确保校验与存储均使用 URI 名称。

```go
if err := rejectBodyField(req, "name"); err != nil {
    return nil, err
}
// ... BindJSON ...
param.Name = &name
```

### 2.3 新增/更新集成测试

- `test/integration/tests/provider/update/update_test.go`：
  - 原有子测试请求体不再携带 `name`；
  - 新增 `PV-4-005` 子测试，请求体不传 `name`，验证更新成功且返回的 `name` 与 URI 一致；
  - 新增 `PV-4-006` 子测试，请求体包含 `name`，验证返回 422。
- `test/integration/tests/provider/design.md`：测试用例表中增加 `PV-4-005`/`PV-4-006`，并说明含 `name` 返回 422。

### 2.4 影响

调用方在 `PATCH /open-api/v1/providers/{provider_name}` 时，不应在请求体中携带 `name`；若携带，接口返回 422。

## 3. `clusters.md` 2.4 节与代码变更

### 3.1 文档新增限制

在 **输入参数（Body）** 说明中增加：

> 可修改字段含义同创建接口，但**输入参数不包括 `name`，即不能修改 cluster 的 name**（名称由 URI 中的 `cluster_name` 指定）。若请求体中仍包含 `name`，返回 422。

### 3.2 示例修正

删除该节 HTTP BODY 参数示例中的 `"name": "my-cluster"`，避免与新增限制矛盾。

### 3.3 代码变更

`endpoints/openapi_v1/product_cluster/update_basic.go` 的 `newUpdateParam4Update`：

1. 在绑定参数前，通过 `rejectBodyField(req, "name")` 检查请求体是否包含 `name`；若包含，返回 422。
2. 绑定参数后，将 `param.Name` 设置为 URI 中的 `cluster_name`，确保名称由 URI 决定。

```go
if err := rejectBodyField(req, "name"); err != nil {
    return nil, err
}
// ... Bind ...
name := mux.Vars(req)["cluster_name"]
param.Name = &name
```

### 3.4 配套测试文档

`test/integration/tests/clusters/design.md` 9.2.1 节同步更新：

> 可修改字段含义同创建接口，但**输入参数不包括 `name`，即不能修改 cluster 的 name**（名称由 URI 中的 `cluster_name` 指定；若包含 `name` 返回 422）。

### 3.5 新增/更新集成测试

- `test/integration/tests/clusters/update/update_test.go`：
  - 新增 `CL-4-012` 子测试，请求体不传 `name`，验证更新成功且返回的 `name` 与 URI 一致；
  - 新增 `CL-4-013` 子测试，请求体包含 `name`，验证返回 422。
- `test/integration/tests/clusters/design.md`：测试用例表中增加 `CL-4-012`/`CL-4-013`，并在 9.2.1 节说明含 `name` 返回 422。

### 3.6 影响

调用方在 `PATCH /open-api/v1/clusters/{cluster_name}` 时，不应在请求体中携带 `name`；若携带，接口返回 422。

## 4. `clusters.md` 2.5 节与代码变更

### 4.1 文档新增限制

在 **执行逻辑** 中补充引用检查步骤：

> 删除集群时，系统先检查该集群是否被 **AI 路由规则**（global / entity / api-key 级别）引用；若被引用，删除失败。
>
> 通过引用检查后，系统自动执行以下级联清理：
> 1. 解绑集群关联的子集群
> 2. 删除子集群
> 3. 删除子集群关联的实例池
> 4. 删除集群

### 4.2 代码现状

`stateful/container/rdb/components.go` 中 `ClusterManager` 的 `deleteCheckers` 仍同时注册产品级与 AI 路由规则检查：

```go
map[string]func(context.Context, *ibasic.Product, *icluster_conf.Cluster) error{
    "rules":       container.RouteRuleManager.ClusterDeleteChecker,
    "route_rules": container.RouteRulesManager.ClusterDeleteChecker,
},
```

本次未修改该注册逻辑。因 AI 网关场景不使用产品级路由规则，文档仅向调用方描述 AI 路由规则（`route_rules`）引用检查。

### 4.3 背景说明

对应 AI 路由规则检查实现位于：

- `model/route_rules/route_rules.go:184`：`RouteRulesManager.ClusterDeleteChecker` 检查集群名是否出现在 AI 路由规则的 `targets` 或 `fallbacks` 中。

## 5. 影响面

| 影响项 | 说明 |
|--------|------|
| OpenAPI 文档 | 补充约束说明，明确调用方应如何使用更新/删除接口。 |
| 代码实现 | Provider/Cluster 更新 Handler 显式使用 URI 名称；Cluster 删除检查器保持产品级与 AI 路由规则双重检查，文档侧仅描述 AI 路由规则。 |
| 调用方 | 更新接口调用方可确认无需传 `name`；删除接口调用方可了解引用冲突来源。 |
| 数据迁移 | 无 |
