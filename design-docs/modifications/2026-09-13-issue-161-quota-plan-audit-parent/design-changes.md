# Issue #161 详细设计：配额计划/限流策略审计归属修复

## 1. 归属模型

### 1.1 新增归属类型

`model/shared/types.go`（**不能放 model/quota**，原因见 §3.4）：

```go
// ResourceOwner 标识嵌套资源的归属（owner 资源的业务身份）。
// Type 取 ioperlog.ResourceTypeEntity / ResourceTypeAPIKey，
// ID 为 owner 的业务 ID（Entity ID / API Key ID），写入日志 resource_parent_id。
type ResourceOwner struct {
    Type string
    ID   string
}
```

- 日志条目的 `resource_type` 仍为 `quota_plan` / `rate_limit_policy`（`ioperlog/types.go:72-73` 已有常量），`resource_parent_id` 填 `ResourceOwner.ID`。
- owner 未知场景（未来独立 `/quota-plans` 端点若出现）允许 `ResourceOwner{}`（空 ID），行为与现状一致，不报错。

### 1.2 方法签名变化

`QuotaPlanManager`（`model/quota/quota_plan_manager.go`）：

| 方法 | 现状 | 修改后 |
|------|------|--------|
| `CreateQuotaPlan` | `(ctx, param)` | `(ctx, param, owner shared.ResourceOwner)` |
| `UpdateQuotaPlan` | `(ctx, filter, param)` | `(ctx, filter, param, owner shared.ResourceOwner)` |
| `DeleteQuotaPlan` | `(ctx, filter)` | `(ctx, filter, owner shared.ResourceOwner)` |
| `ResetBalance` | `(ctx, planID, newQuota, updateLastResetAt)` | `(ctx, planID, newQuota, updateLastResetAt, owner shared.ResourceOwner)` |

`RateLimitPolicyManager` 三个 CRUD 方法同构增加 `owner shared.ResourceOwner`。

调用方影响（已全仓核实）：

- CRUD 三方法现状**零调用方**，签名变化无破坏面。
- `ResetBalance` 仅 2 个调用点：`endpoints/openapi_v1/entity/reset_quota.go:110`、`endpoints/openapi_v1/api_key/reset_quota.go:101`，本次一并修改。

## 2. 写/审分层设计（关键决策）

### 2.1 为什么不直接在事务内调带审计的 manager 方法

已验证 `QuotaPlanManager` CRUD 不自开事务（`quota_plan_manager.go:59-70` 直接调 storager），在 `EntityManager` 的 `txn.AtomExecute` 内调用**技术上可行**。但：

- entity/api_key 自身的操作日志在**事务外**记录（成功在提交后、失败在返回时，统一一条 failed）；
- 若嵌套写在内层随写记日志，后续步骤失败回滚后，将出现 `quota_plan/create=success` 与 `entity=failed` 并存的残留不一致日志。

因此审计必须与 owner 资源的日志同生命周期。

### 2.2 分层结构

每个 manager 方法拆为两层：

```
CreateQuotaPlan(ctx, param, owner)            // 公开方法：createCore + 同步审计（供事务外/未来独立端点）
├── createCore(ctx, param) (int64, error)     // 纯 storager 写（未导出或包内可见）
└── recordQuotaPlanOperation(ctx, action, planID, owner.ID, before, after, err)
```

嵌套路径（entity/api_key）：

```
EntityManager.CreateEntity
├── txn.AtomExecute {
│       quotaPlanStorager.CreateQuotaPlan(...)   // 写路径保持现状（事务内）
│       ...                                      // 其余嵌套资源同
│   }
└── 事务外（与 recordEntityOperation 同点）：
        quotaPlanAuditor.AuditQuotaPlanCreate(ctx, param.QuotaPlan, quotaPlanID, owner, err)
        // quotaPlanAuditor 为 model/shared 接口字段，装配时注入 *QuotaPlanManager
```

即：**写入口单一在 storager（保持原子性），审计入口单一在 QuotaPlanManager**。嵌套路径"经 manager"体现为 manager 暴露的审计方法，manager CRUD 公开方法不再是死代码（承载事务外审计与独立端点两个用途），也不存在回滚残留。

### 2.3 审计方法形态（嵌套路径专用）

在 `model/quota/operation_log.go` 增加（`rate_limit_policy` 同构）：

```go
// AuditQuotaPlanCreate 记录嵌套创建配额计划的审计（写已由调用方在事务内完成）。
// err 为写路径返回的错误；成功/失败均记录，ResourceParentID 取 owner.ID。
func (m *QuotaPlanManager) AuditQuotaPlanCreate(ctx context.Context, param *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error)
// AuditQuotaPlanUpdate(ctx, oldPlan, param, planID, owner, err)
// AuditQuotaPlanDelete(ctx, oldPlan, planID, owner, err)
```

- before/after 快照规则与现有 `UpdateQuotaPlan`/`DeleteQuotaPlan` 内记录逻辑完全一致（`quota_plan_manager.go:88-106/118-135`），复用 `quotaPlanParamToMap`。
- update 的 before 快照：`EntityManager` 更新流程当前不读旧 plan（`:287-298`），需在校验阶段随 `oldEntity` 一并 `FetchQuotaPlan`（一次额外读，仅当 `param.QuotaPlan != nil` 且实体已有 plan）。

## 3. 嵌套调用点改造明细

### 3.1 Entity（owner.Type=entity，owner.ID=实体业务 ID）

| 场景 | 写路径（不变） | 审计接入点（事务外） | owner.ID 来源 |
|------|---------------|---------------------|---------------|
| 创建 | `entity_manager.go:108-113` | `recordEntityOperation` 成功/失败两处同点 | `param.EntityID`（序列表预分配，#132，事务前已就绪） |
| 更新 | `:287-298` | `UpdateEntity` 成功/失败记录点（`:336-363` 一带） | 请求侧 `filter`/`param` 实体 ID（`resolveEntityIdentifiers` 既有） |
| 删除 | `:416-420` | delete 记录点 | 库中实体快照 `one.EntityID` |

### 3.2 API Key（owner.Type=api_key，owner.ID=api key 业务 ID）

| 场景 | 写路径（不变） | 审计接入点 | owner.ID 来源 |
|------|---------------|-----------|---------------|
| 创建 | `api_key.go:652-657` | api key 自身日志记录点 | `param.ID`（api key 业务 ID） |
| 更新 | `:518-528` | 同上 | `one.ID`（库中快照） |
| 删除 | `:442-446` | 同上 | `one.ID` |

依据 contract：`resource_parent_id` 对配额计划即 owner 的业务 ID——Entity 嵌套填 Entity ID，API Key 嵌套填 API Key ID（api_key 类型日志的 parent 才填 Entity ID，见 `api_key_operation_log.go:40-43`，两者层级不同，勿混用）。

### 3.3 Reset（缺陷②）

- `endpoints/openapi_v1/entity/reset_quota.go:110`：`ResetBalance(..., shared.ResourceOwner{Type: entity, ID: resetReq.EntityID})`，`resetReq.EntityID` 即 URI `{id}`，已在手（`:136` 用于响应）。
- `endpoints/openapi_v1/api_key/reset_quota.go:101`：`shared.ResourceOwner{Type: api_key, ID: *apiKey.ID}`（库中取，URI 为 api key ID 亦可，取库中快照与 §3.2 一致）。

### 3.4 依赖与装配（含依赖方向修正）

**依赖事实**：`model/quota` 已 import `model/entity` / `model/api_key`（`adapters.go:18`、`balance_sync.go:9-10`），即依赖方向为 `quota → entity/api_key`。因此 `entity/api_key → quota` 会成环，**归属类型与 auditor 接口必须下沉到 `model/shared`**（entity/api_key/quota 三方均已有 shared 依赖）。

- `EntityManager`/`APIKeyManager` 构造器的 `quotaPlanStorager shared.QuotaPlanStorager` 参数保留（写路径仍用），**新增** auditor 接口字段：

```go
// model/shared 定义窄接口（QuotaPlanParam 等类型 shared 已有）
type QuotaPlanAuditor interface {
    AuditQuotaPlanCreate(ctx context.Context, param *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
    AuditQuotaPlanUpdate(ctx context.Context, oldPlan, param *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
    AuditQuotaPlanDelete(ctx context.Context, oldPlan *QuotaPlanParam, planID int64, owner ResourceOwner, err error)
}
type RateLimitPolicyAuditor interface { /* 同构 */ }
```

- `QuotaPlanManager` / `RateLimitPolicyManager` 实现上述接口；装配点注入；`mocks_test.go` 增加 fake auditor，现有表驱动测试沿用。
- 依赖图：`entity → shared ← quota → entity`（quota→entity 为既有依赖，auditor 接口在 shared 处反向解耦，无新增环）。

## 4. rate_limit_policy 同族修复

`model/rate_limit_policy/` 按 §1–§3 完全同构执行：

- `rate_limit_policy_manager.go:61-120` 七处 `parentID=""` 调用点接入 owner；CRUD 签名加 `owner shared.ResourceOwner`。
- 嵌套写路径（`entity_manager.go:116/302-316/398-423`）审计接入点同上表。
- **该族无 TC 把守**：修复后除单测外需手工核验一次嵌套创建限流策略产生 `rate_limit_policy/create` 日志且 parent 为实体 ID（或在 SC2101 系补一条审计断言，建议随本卡补充）。

## 5. 兼容性与契约

- **API 契约**：无变化。本修复是 operation-logs.md §3 已声明行为的兑现（`/quota-plans` 的 POST/PUT/DELETE/reset 触发日志且 parent 为 owner 业务 ID）。
- **DB  schema**：无变化（操作日志表已有 `resource_parent_id` 列）。
- **对现有日志的影响**：仅新增日志条目与填充 parent，不回填历史数据（历史缺失日志无法补，issue 现场数据保持原样）。

## 6. 测试计划

1. **单测**（表驱动，fake auditor 捕获条目）：
   - entity 创建/更新/删除 × （成功/失败） → 断言 `resource_type=quota_plan`、`resource_parent_id=实体ID`、action/create/update/delete、failed 分支 status/error_msg；
   - api_key 同构（parent=api key ID）；
   - reset 两个端点 → parent=URI/snapshot ID；
   - `model/quota`、`model/rate_limit_policy` 公开方法 owner 传参后 parent 填充正确；
   - 失败分支 owner 从请求侧解析（update 前置校验失败仍带身份）。
2. **门槛**：`go test ./...` 全绿；`make test-model-cover-gate`（model ≥70%，实测 80.8%）。
3. **集成测试（已实现）**：`test/integration/tests/operation_log/nested_audit/nested_audit_test.go`，9 条用例（OL-N-001~009）覆盖 entity/api-key 嵌套 create/update/reset/delete 的 `resource_parent_id` 归属及 rate_limit_policy 同族断言；`testutil.OperationLogEntry` 补充 `resource_parent_id` 字段。
4. **集成回归**：SC2101-TC047 requeue real-verification（8 条归属断言：create/update/delete×2 + reset×2）。
5. **手工核验**（issue 现场复现）：
   - `POST /open-api/v1/entities` 带嵌套 `quota_plan` → ±10 分钟窗口出现 `quota_plan/create`，parent=新 Entity ID；
   - `POST /entities/{id}/quota-plan/reset` → 日志 parent=`{id}`。

## 7. 实施步骤（建议提交拆分）

1. `model/shared`：`ResourceOwner` 与 `QuotaPlanAuditor`/`RateLimitPolicyAuditor` 接口；`model/quota`：core/audit 拆分 + 4 方法签名（ResetBalance 2 调用点同步）。
2. `model/rate_limit_policy`：同构改造。
3. `model/entity`、`model/api_key`：审计接入 + 装配 + mocks。
4. 单测补齐（含 rate_limit_policy 审计断言）。
5. SC2101-TC047 requeue + 手工核验，关闭 issue。
