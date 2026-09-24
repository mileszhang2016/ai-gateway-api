# Issue #161：配额计划/限流策略审计日志缺失与归属修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/161](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/161)

> 按 design-docs/api-define/OpenAPI接口定义/operation-logs.md §3 契约，`/quota-plans` 的 POST / PUT / DELETE / reset 均应触发操作日志，且 `resource_parent_id` 应填资源父级业务 ID（对配额计划即 owner Entity/API Key 的业务 ID）。当前实现存在两层缺陷：
>
> - **缺陷①（主）**：通过 Entity / API Key 生命周期嵌套创建、更新、删除配额计划时，不产生任何 `resource_type=quota_plan` 的操作日志，配额计划全生命周期不可追溯。
> - **缺陷②（次）**：手动 reset（唯一走审计层的路径）产生的日志 `resource_parent_id` 恒为空串，无法反查所属 Entity/API Key。
>
> 同族隐患：`rate_limit_policy` 与 `quota_plan` 逐行同构，嵌套路径同样直连 storager，manager 日志调用 `parentID=""`，且当前无 TC 把守。

影响：SC2101-TC047 real-verification FAILED_PRODUCT（8 条归属断言全挂），catalog 已登记。

## 2. 根因核实（本地 v0.0.9 / 5fcd567 代码逐点验证，issue 基于 99d53df 的行号已漂移）

### 2.1 嵌套路径绕过审计层（缺陷①）

`EntityManager` / `APIKeyManager` 的嵌套配额 CRUD 全部直连 storager（处于 owner 资源的事务内）：

| 操作 | 位置 | 调用形态 |
|------|------|---------|
| Entity 创建 | `model/entity/entity_manager.go:108-113` | `m.quotaPlanStorager.CreateQuotaPlan`（`txn.AtomExecute` 内） |
| Entity 更新 | `model/entity/entity_manager.go:287-298` | 有 plan 则 `UpdateQuotaPlan`，否则 `CreateQuotaPlan`（事务内） |
| Entity 删除 | `model/entity/entity_manager.go:416-420` | `DeleteQuotaPlan`（事务内） |
| API Key 创建 | `model/api_key/api_key.go:652-657` | `CreateQuotaPlan` |
| API Key 更新 | `model/api_key/api_key.go:518-528` | `UpdateQuotaPlan` / `CreateQuotaPlan` |
| API Key 删除 | `model/api_key/api_key.go:442-446` | `DeleteQuotaPlan` |

而带审计的 `QuotaPlanManager.CreateQuotaPlan/UpdateQuotaPlan/DeleteQuotaPlan`（`model/quota/quota_plan_manager.go:59/81/111`）在全仓**无任何调用方（死代码）**，全仓 grep `QuotaPlanManager\.(Create|Update|Delete)QuotaPlan` 零命中（已验证）。

`rate_limit_policy` 同构：`entity_manager.go:116/302-316/398-423` 直连 `rateLimitPolicyStorager`；`RateLimitPolicyManager`（`model/rate_limit_policy/rate_limit_policy_manager.go:61-120`）7 处 `recordRateLimitPolicyOperation(..., "", ...)` 同样无 CRUD 调用方。

### 2.2 归属管道缺失（缺陷②）

`recordQuotaPlanOperation`（`model/quota/operation_log.go:25`）的 `parentID` 参数在全部 10 个调用点硬编码传 `""`（`quota_plan_manager.go:62/66/88/98/106/118/127/135/203/216`）；`CreateQuotaPlan/UpdateQuotaPlan/DeleteQuotaPlan/ResetBalance` 签名均不接收 owner 标识。

唯一走 manager 的 reset 路径：`endpoints/openapi_v1/entity/reset_quota.go:110` 与 `api_key/reset_quota.go:101` 调 `QuotaPlanManager.ResetBalance(ctx, planID, quota, false)` —— URI 中的 owner ID（`resetReq.EntityID` / api key ID）在手但未透传。

### 2.3 正确先例（产品中已有）

`model/api_key/api_key_operation_log.go:40-43`：`api_key` 类型日志的 `resource_parent_id` 从库中 `apiKey.EntityID` 取，同窗日志 id=29192 已正确填充——证明归属实现模式可行，只需把 owner 身份从请求上下文传入。

## 3. 修复方案（决策摘要）

**推荐方案：manager 单一审计入口 + 写/审分层（issue 建议 1 的修正版）**，`rate_limit_policy` 同构一并修复（issue 建议 3，同卡同 PR）。详细设计见 [design-changes.md](design-changes.md)。

| 决策点 | 结论 | 理由 |
|--------|------|------|
| 嵌套写路径是否改为经 manager | 写操作保持 storager 直连（事务内不变），审计统一走 manager 暴露的审计方法 | 已验证 manager CRUD **不自开事务**（`CreateQuotaPlan` 直接调 storager，`:59-70`），但事务内随写记日志会在后续步骤回滚时残留"成功日志"（entity 自身日志在事务外记录，记 failed），两种语义必须对齐 |
| 死代码处理 | manager CRUD 扩展 owner 入参后成为嵌套路径的审计载体与唯一审计入口，不再删除 | 保留未来独立 `/quota-plans` 端点的落点 |
| 归属来源 | reset：URI owner ID 直接透传；嵌套 create/update：请求侧解析（`param.EntityID` / api key ID，entity ID 已序列表预分配 #132）；delete：库中快照取 | 失败场景也必须可归属 |
| API 契约 / DB | 无变化 | 属 operation-logs.md §3 契约兑现 |
| 同族修复 | `rate_limit_policy` 相同形态同 PR | 无 TC 把守的同类缺口 |

### 3.1 改动范围

- `model/shared/`：新增 `ResourceOwner` 归属类型与 `QuotaPlanAuditor` / `RateLimitPolicyAuditor` 窄接口（依赖下沉，避免环）
- `model/quota/`：4 个公开方法增加 owner 入参；core/audit 拆分
- `model/rate_limit_policy/`：同构改造
- `model/entity/`、`model/api_key/`：嵌套调用点接入审计，构造器新增 auditor 注入
- `endpoints/openapi_v1/entity/reset_quota.go`、`api_key/reset_quota.go`：透传 URI owner ID
- manager 装配点（container/stateful）与 `mocks_test.go` 同步

### 3.2 验证计划

1. 单测：`model/entity`、`model/api_key`、`model/quota`、`model/rate_limit_policy` 断言 `resource_type`/`resource_parent_id`（成功 + 失败分支）；`go test ./...` 与 `make test-model-cover-gate`（≥70%）。
2. 集成回归：requeue SC2101-TC047 real-verification（断言 `resource_parent_id == owner.id`）。
3. 手工核验（issue 现场复现两步）：建带配额的 Entity → 窗口查询应出现 `quota_plan/create` 且 parent=新 Entity ID；reset 后日志 parent=该 Entity ID。

### 3.3 可选附带项（不阻塞本卡）

entity/api-key 日志的 `change_summary` 不含嵌套 `quota_plan` 变更（`entityParamToMap` 无该字段），归因时可评估是否补充。
