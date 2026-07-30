# 接口参数合法性条件 —— 代码与测试落地（2026-07-29 追加）

## 1. 本次补充说明的范围

在 `api-changes.md` 与 `design-changes.md` 已完成**接口定义文档增强**的基础上，本次继续完成：

1. **代码实现侧**：在 `endpoints/openapi_v1` 各端点中补齐/修复参数校验逻辑，使实现与文档中的 `合法性条件` 保持一致。
2. **集成测试侧**：补充 `create` / `update` 等端点的合法性校验用例，覆盖文档中新增的参数约束。
3. **单元测试侧**：为 `endpoints/openapi_v1` 中可独立测试的校验函数/方法增加 Go 单元测试。
4. **缺陷修复同步**：合并上游 PR #26，修复 `route-tables` 中 `apikey` 类型 `owner` 字段误用 API Key 字符串的问题，并强化对应集成测试。

---

## 2. 代码实现变更

### 2.1 新增通用校验函数

| 文件 | 变更 |
|------|------|
| `lib/validate/validate.go` | 新增 `EntityName(s string) error`，校验实体名：长度 1-64、不含控制字符、首尾无空白。 |

### 2.2 Entity 端点校验补齐

| 文件 | 变更 |
|------|------|
| `endpoints/openapi_v1/entity/validator.go` | 新增 `validateEntityParam`，集中校验 `name`、`type`、`rate_limit_policy`、`quota_plan`、`route_rules`；`update` 时若提供 `name` 也进行 `EntityName` 校验。 |
| `endpoints/openapi_v1/entity/create.go` | 创建时调用 `validateEntityParam(param, true)`。 |
| `endpoints/openapi_v1/entity/update.go` / `full_update.go` | 更新时调用 `validateEntityParam(param, false)`。 |

### 2.3 API-Key 端点校验补齐

| 文件 | 变更 |
|------|------|
| `endpoints/openapi_v1/api_key/checker.go` | 在 `checkCreateAPIKey` / `checkUpdateAPIKey` / `checkFullUpdateAPIKey` 中增加 `validate.RouteRules(param.RouteRules)` 调用。 |

### 2.4 其他端点校验函数/方法

以下端点中原本内联或分散的校验逻辑已被抽出为可测试的 `Validate()` 方法或独立函数：

- `endpoints/openapi_v1/ai_route/list.go`：`ProductRouteRuleParam.Validate()`
- `endpoints/openapi_v1/auth/token_create.go`：`TokenCreateParam.Validate()`
- `endpoints/openapi_v1/auth/user_create.go`：`UserCreateParam.Validate()`
- `endpoints/openapi_v1/auth/user_update_is_admin.go`：`UserUpdateIsAdminParam.Validate()`
- `endpoints/openapi_v1/auth/user_update_password.go`：`UserUpdatePasswordParam.Validate()`
- `endpoints/openapi_v1/bfe_cluster/create.go`：`BFEClusterCreateParam.Validate()`
- `endpoints/openapi_v1/domain/create.go`：`CreateParam.Validate()`
- `endpoints/openapi_v1/product_cluster/create.go`：`UpsertParam.Validate()`、`checkInstancePool()`
- `endpoints/openapi_v1/product_cluster/checker.go`：`checkLLMConfig()`、`CheckEPPServer()`
- `endpoints/openapi_v1/product_cluster/update_subcluster.go`：`BindSubCluster.Validate()`
- `endpoints/openapi_v1/product_pool/create.go`：`UpsertParam.Validate()`
- `endpoints/openapi_v1/subcluster/create.go`：`CreateParam.Validate()`
- `endpoints/openapi_v1/subcluster/update.go`：`UpdateParam.Validate()`

### 2.5 上游缺陷修复（PR #26）

| 文件 | 变更 |
|------|------|
| `model/icluster_conf/api_key.go` | `CreateAPIKey` 时若 `ID` 为空则自动生成 UUID；创建/更新 API Key 的 `RouteRules` 时，owner 关联键由 `param.Key` 改为 `param.ID`，使 `route-tables` 返回的 `owner` 为 `api_key.id`。 |

---

## 3. 集成测试变更

### 3.1 设计文档更新

`test/integration/tests/<module>/design.md` 全部刷新，新增/细化 `合法性条件` 列，并补充 `create` / `update` 等端点的校验场景。

### 3.2 新增/强化的测试用例

| 模块 | 端点 | 用例 ID | 说明 |
|------|------|---------|------|
| entity | create | E-1-XXX 系列 | 缺失 name/type、非法 name、非法 type、route_rules 重复名等 |
| entity | full_update | E-4-005 | name 首尾空白 |
| entity | partial_update | E-5-003 | route_rules 规则名重复 |
| api_key | create | AK-1-XXX 系列 | quota_plan.unit 非法、rate_limit_policy window=0 等 |
| api_key | full_update | AK-4-004 | quota_plan.unit 非法 |
| api_key | partial_update | AK-5-004 | rate_limit_policy window_minutes=0 |
| clusters | update | CL-4-005 | instance_pool IP 非法 |
| entity_type | update | ET-4-004 | description 超长 |
| route_tables | list | RT-1-005 | **强化**：按 `owner=apiKeyID` 过滤后必须返回 ≥1 条记录且 `owner` 均等于 `apiKeyID` |
| route_tables | list | RT-1-011 | 新增：按不存在的 `owner` 过滤应返回空列表 |

---

## 4. 单元测试变更

为 `endpoints/openapi_v1` 下各端点的校验函数/方法新增 `_test.go`：

| 包 | 测试文件 | 覆盖函数/方法 |
|----|----------|---------------|
| `entity` | `validator_test.go` | `validateEntityParam` |
| `api_key` | `checker_test.go` | `checkCreateAPIKey` / `checkUpdateAPIKey` / `checkFullUpdateAPIKey` |
| `product_cluster` | `checker_test.go` | `checkLLMConfig` / `CheckEPPServer` / `checkInstancePool` |
| `product_cluster` | `create_test.go` | `UpsertParam.Validate` |
| `product_cluster` | `update_subcluster_test.go` | `BindSubCluster.Validate` |
| `ai_route` | `list_test.go` | `ProductRouteRuleParam.Validate` |
| `auth` | `validator_test.go` | `TokenCreateParam.Validate` / `UserCreateParam.Validate` / `UserUpdateIsAdminParam.Validate` / `UserUpdatePasswordParam.Validate` |
| `bfe_cluster` | `create_test.go` | `BFEClusterCreateParam.Validate` |
| `domain` | `create_test.go` | `CreateParam.Validate` |
| `product_pool` | `create_test.go` | `UpsertParam.Validate` |
| `subcluster` | `create_test.go` | `CreateParam.Validate` |
| `subcluster` | `update_test.go` | `UpdateParam.Validate` |

---

## 5. 验证结果

### 5.1 单元测试

```bash
cd ai-gateway-api
go test -count=1 ./endpoints/openapi_v1/...
```

所有带测试的包均 `ok`。

### 5.2 集成测试

```bash
cd ai-gateway-api/test/integration
go test ./tests/...
```

全部模块通过（含强化后的 `route_tables/list`）。

---

## 6. 未覆盖说明

以下包目前仍以 handler 流程编排为主，校验直接调用 `lib/validate` 或依赖 `container.*Manager` / Redis / DB，尚未补充单元测试：

- `bfe_pool`、`certificate`、`entity_type`、`general`、`global_route_rules`
- `model_provider_type`、`product`、`route`、`route_tables`、`tool`、`traffic`

如需继续覆盖，建议通过 **提取纯校验函数** 或 **引入 mock/依赖注入** 两种方式推进。
