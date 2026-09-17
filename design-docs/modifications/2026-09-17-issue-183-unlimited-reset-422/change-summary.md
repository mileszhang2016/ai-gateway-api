# Issue #183：unlimited 配额计划 reset 返回 500 的修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/183](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/183)

> 对 `unlimited` 配额计划的 API Key 调用 `POST /open-api/v1/api-keys/{id}/quota-plan/reset` 时，网关返回 `HTTP 500`（envelope `ErrNum=500`、`ErrMsg="Unknown Exception: cannot reset balance for unlimited quota"`）。该条件是确定性、可预测的语义错误（unlimited plan 没有 balance 可 reset），违反已批准产品设计对管理写入「2xx/4xx/422 语义」与「失败审计须保留稳定 status/error 语义」的契约。

性质：#161（quota_plan 审计两层缺陷）修复 `26a4ed9` 部署后的**残留缺陷**——#161 闭合了审计日志层（嵌套 CRUD 零日志、reset 日志 parentID 硬编码空串，6/7 步通过），但 unlimited reset 分支的 error-wrapping 未被同次修复覆盖，导致 SC2101-TC047 无法闭卡（`sc2101AuditRequireFailure` 要求失败响应落在 `[400,500)`，500 先挂，进不到 `status=2` 日志断言）。

Oracle（normative，`coordinator_approved`）：`docs/test-design/SC2101/SC2101-TC047.md` 契约卡 + 断言 3（unlimited reset 必须失败：`status=2` + 非空错误原因 + 配额不变）；底层条款 `design-docs/api-define/OpenAPI接口定义/operation-logs.md` §3。

## 2. 根因核实（本地代码逐点验证）

错误链路三段：

1. **模型层裸错误**：`model/quota/quota_plan_manager.go:164` unlimited 分支 `return fmt.Errorf("cannot reset balance for unlimited quota")`——无任何错误类型前缀。
2. **端点原样透传**：`endpoints/openapi_v1/api_key/reset_quota.go:110-112` 与 `endpoints/openapi_v1/entity/reset_quota.go:119-121` 对 `ResetBalance` 返回的 error 不加包装直接 `return nil, err`。
3. **框架兜底 500**：`lib/xerror/resolve.go:96-98`——错误前缀不匹配任何已知类型（`PARAM`/`Model`/`DAO`/…）时落入 `default` 分支，`ErrNo=500, Type="Unknown Exception"`；envelope 组装见 `lib/xreq/result.go:143-145`（`ErrMsg = Type + ": " + Msg`，HTTP status 取 `ErrNum`，`result.go:188` 400-499 原样透传）。

对照组：同 handler「API-Key has no quota plan」走 `xerror.WrapParamErrorWithMsg`（`api_key/reset_quota.go:72`），`resolve.go:64-66` 映射为 `422 / Param Illegal`——即 issue 要求的「同族同码」。

调用面核实：`ResetBalance` 全仓仅 2 个生产调用点（上述两个 reset 端点）；周期调度器走 `BalanceSyncer.ResetExpiredBalances`（`model/quota/scheduler.go:142`），且只处理 `unlimited=0` 的计划，不受本次改动影响。

## 3. 修复方案（决策摘要）

**模型层一行修复**：`quota_plan_manager.go:164` 改为 `return xerror.WrapParamErrorWithMsg("cannot reset balance for unlimited quota")`（新增 `lib/xerror` import）。端点零改动。详见 [design-changes.md](design-changes.md)。

| 决策点 | 结论 | 理由 |
|--------|------|------|
| 修复落点：模型层 vs 端点层 | **模型层**（unlimited 分支精确包装 PARAM 类型） | 断言 3 要求 unlimited reset 失败产生 `quota_plan/reset status=2` 审计日志，该日志只在 `ResetBalance` 错误分支记录（`quota_plan_manager.go:194-208`）；端点预检会绕过审计，挂断言 3。端点若对 `ResetBalance` 全部错误统一包 PARAM，会把 DB 故障（应 5xx）误标为 4xx |
| 错误码取值 | **422 / Param Illegal**（`etParam`） | TC047 契约卡 normative 要求，与同 handler「no quota plan」同族同码；修复后 `ErrMsg="Param Illegal: cannot reset balance for unlimited quota"`，与 issue 预期逐字一致 |
| 模型层依赖 xerror 是否合规 | **合规** | 先例：`model/imodel_price/validate.go:143` 起模型层已大量使用 `xerror.WrapParamErrorWithMsg`（全仓 xerror 调用 265 处），语义校验错误在模型层定型是既有惯例 |
| Entity 侧是否同病 | **确认同病，同修复覆盖** | `entity/reset_quota.go:118-121` 同样透传裸错误，同一行模型层修改同时闭合两侧 |
| 审计 error_msg 形态 | 保持 `recordQuotaPlanOperation` 不变 | 落库 error_msg 为 `"PARAM: cannot reset balance for unlimited quota"`（`err.Error()` 经 `TruncateErrorMessageDefault`），非空、确定、满足断言 3；备选「用 `errors.Cause` 剥掉类型前缀」会改变全部既有审计日志内容，收益不抵风险，不做 |
| 兄弟缺陷 `quota_plan not found`（`quota_plan_manager.go:158` 同样裸 error → 500） | **本卡不修，另开 issue** | 仅在引用完整性被破坏（api_key/quota_plan 行缺失）时可达，与 issue 定界（「仅触及 unlimited reset 单分支」）一致；修复候选 `xerror.WrapRecordNotExist("QuotaPlan")` → 404，留给后续卡 |
| API 契约 / DB | 仅错误语义变化，无字段、无 schema 变更 | 属契约卡「稳定 status/error 语义」的兑现 |

### 3.1 改动范围

- `model/quota/quota_plan_manager.go`：unlimited 分支错误包装（+import），1 行核心改动。
- 单测：`model/quota/quota_plan_manager_test.go`（unlimited 分支断言强化）、`model/quota/operation_log_test.go`（失败审计 status=2 / error_msg / 归属断言）。
- 集成测试：`test/integration/tests/operation_log/` 新增 unlimited reset 失败用例（api-key + entity 两路径）。
- 文档同步：`api-keys.md` §2.8、`entities.md` reset 节、`sys-design/模型层设计文档.md` ResetBalance 伪代码、`sys-design/details/配额余额同步机制.md` §6.4（见 [api-changes.md](api-changes.md)）。

### 3.2 修复后行为

| 项目 | 修复前 | 修复后 |
|------|--------|--------|
| HTTP status | 500 | 422 |
| envelope ErrNum | 500 | 422 |
| envelope ErrMsg | `Unknown Exception: cannot reset balance for unlimited quota` | `Param Illegal: cannot reset balance for unlimited quota` |
| 失败审计 | 有（status=2，error_msg 为裸消息） | 有（status=2，error_msg=`PARAM: cannot reset balance for unlimited quota`，owner/resource 关联不变） |
| 配额是否改变 | 不改变（事务内提前返回 + 回滚） | 不改变（同左） |

### 3.3 验证计划

1. 单测：`go test ./model/quota/...`，断言 `xerror.Resolve(err)` → `ErrNo=422, Type="Param Illegal"`；`make test-model-cover-gate`（≥70%）。
2. 集成测试：新增用例断言 HTTP 422 + `ErrNum=422` + `ErrMsg` 前缀 `Param Illegal:` + 审计 `status=2` / 非空 `error_msg` / `resource_parent_id=owner id` + 余额不变。
3. 全量：`make test`。
4. QA 回归：requeue **SC2101-TC047** real-verification（闭 #161 卡 + 本 issue）；`sc2101AuditRequireFailure`（`[400,500)`）与本修复天然兼容，无需放宽。
5. 手工核验（issue 现场复现）：建 unlimited API Key → reset → 422 `Param Illegal: ...`；窗口查询 `quota_plan/reset` 日志 `status=2` 且 parent=api-key id；`DELETE` 清理无残留。

### 3.4 风险与兼容性

- 仅 unlimited reset 单分支的响应变化（500→422）；有限配额 reset、周期调度、其他端点行为不变。
- 监控/告警侧：该分支从 5xx 计数移出，正是 issue 要求修复的误报源。
- Dashboard 等管理面消费方按 envelope 错误处理，422 仍走错误分支，仅展示文案变化；数据平面（BFE）不涉及（控制面端点）。
