# Issue #183 详细设计：unlimited reset 错误语义修复（500 → 422）

## 1. 错误语义框架现状

`lib/xerror` 采用「错误类型前缀 → HTTP 语义」的映射（`resolve.go:50-101`）：

| 前缀（`et*` 常量） | ErrNo | Type |
|------|------|------|
| `PARAM` | 422 | Param Illegal |
| `Model.NullData` | 404 | Record Not Exist |
| `Model.Conflict` | 409 | Conflict |
| `Model` / `DAO` / … | 500 | Biz / Database Exception |
| 无前缀（default） | 500 | **Unknown Exception** |

包装由 `lib/xerror/wrap.go` 提供：`WrapParamErrorWithMsg(tip, args...)` = `errors.Wrap(fmt.Errorf(tip, args...), "PARAM")`（`wrap.go:50-64`，幂等：已带 `PARAM` 前缀则原样返回）。

envelope 组装（`lib/xreq/result.go:135-147`）：

- `ErrNum = rr.ErrNo`；HTTP status 取 `ErrNum`，400-499 原样透传（`result.go:188`）；
- `ErrMsg = rr.Type + ": " + rr.Msg`，其中 `Msg = errors.Cause(err)` 的文本（**不含类型前缀**）。

因此 `WrapParamErrorWithMsg("cannot reset balance for unlimited quota")` 的完整表现为：

```
HTTP 422
{"ErrNum":422, "ErrMsg":"Param Illegal: cannot reset balance for unlimited quota"}
```

与 issue 预期表逐字一致，并与同 handler「API-Key has no quota plan」(`api_key/reset_quota.go:72`) 同族同码。

## 2. 修复点（模型层单点）

### 2.1 代码变更

`model/quota/quota_plan_manager.go`（ResetBalance 事务闭包内）：

```go
// 2. 如果是无限配额，返回错误
if plan.Unlimited != nil && *plan.Unlimited {
-   return fmt.Errorf("cannot reset balance for unlimited quota")
+   return xerror.WrapParamErrorWithMsg("cannot reset balance for unlimited quota")
}
```

import 增加 `"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"`。`fmt` 保留（`:158` 的 `quota_plan not found` 仍在用）。

两个端点（`api_key/reset_quota.go`、`entity/reset_quota.go`）**不改**：错误透传即可被 `Resolve` 正确映射。

### 2.2 为什么必须在模型层包装（关键决策）

TC047 断言 3 要求 unlimited reset 失败产生**审计日志**：`quota_plan/reset`、`status=2`、非空错误原因、owner/resource 关联、配额不变。

reset 的审计日志**只有一个发射点**：`ResetBalance` 的错误分支（`quota_plan_manager.go:194-208`）：

```go
err := m.txn.AtomExecute(ctx, func(ctx context.Context) error { ... })
if err != nil {
    ...
    m.recordQuotaPlanOperation(ctx, string(ioperlog.ActionReset), quotaPlanIDString(planID), owner.ID, ..., err)  // status=2 仅在此产生
    return err
}
```

对比方案：

| 方案 | HTTP 语义 | 审计 status=2 | 评估 |
|------|----------|--------------|------|
| A. 端点预检 `plan.Unlimited` 后直接返回 422 | ✔ | ✘ 绕过 `ResetBalance`，无 reset 日志 | 挂断言 3，否决 |
| B. 端点对 `ResetBalance` 的一切错误统一 `WrapParamError` | ✔ 但 DB 故障也成 422 | ✔ | 内部故障误标 4xx，破坏 5xx 告警语义，否决 |
| C. 模型层 sentinel error + 端点 `errors.Is` 选择性包装 | ✔ | ✔ | 可行但多一层间接，无收益 |
| **D. 模型层该分支精确包装 PARAM（选定）** | ✔ | ✔ | 单点改动，语义校验错误在产生处定型 |

选 D 的额外依据：模型层语义校验错误用 `xerror.WrapParamErrorWithMsg` 是既有惯例（`model/imodel_price/validate.go:143-292` 等 265 处调用），且 storager/DB 错误路径不受影响——它们仍无前缀 → 500，保持「内部故障 5xx、参数语义 4xx」的区分。

### 2.3 事务与配额不变性

unlimited 分支在事务闭包内第 2 步即返回错误，`AtomExecute` 回滚；本就不产生任何写（步骤 3/4 未执行）。「不改变配额」在修复前后均成立，修复不触碰该性质。

## 3. 审计日志影响

`recordQuotaPlanOperation`（`model/quota/operation_log.go:26-52`）对失败取 `errorMsg = ioperlog.TruncateErrorMessageDefault(err)`，即 `err.Error()` 截断 1024 字符。

修复后 unlimited reset 失败的审计条目：

| 字段 | 值 |
|------|-----|
| action | `reset` |
| resource_type | `quota_plan` |
| resource_id | `<planID>` |
| resource_parent_id | owner ID（api-key/entity 业务 ID，#161 已修复透传） |
| status | 2（`StatusFailed`） |
| error_msg | `PARAM: cannot reset balance for unlimited quota` |

说明：error_msg 带 `PARAM: ` 类型前缀。满足断言 3「非空错误原因」且内容确定稳定。**备选**（`recordQuotaPlanOperation` 内改取 `errors.Cause(err)` 剥前缀）被否决：会一并改变其他 action 既有日志的 error_msg 内容（如嵌套路径传入的包装错误），超出本卡最小改动原则。

HTTP 层 `ErrMsg` 不受影响——`Resolve` 的 `Msg` 本就取 `errors.Cause`（`resolve.go:57`），不含前缀。

## 4. 影响面核实

- `ResetBalance` 生产调用点仅 2 个 reset 端点（已核实）；周期调度器 `QuotaResetScheduler` 走 `BalanceSyncer.ResetExpiredBalances`（`scheduler.go:142`），只扫 `unlimited=0`，不经过本分支。
- Entity 侧 `entity/reset_quota.go:118-121` 与 API-Key 侧同构透传，同一行修改同时闭合（issue「待确认 Entity 侧是否同病」→ **确认同病**）。
- 有限配额 reset 成功路径（status=1 审计、Redis 重置、响应字段）零变化。
- `quota_plan not found`（`quota_plan_manager.go:158`）同为裸 error → 500，但仅在引用完整性已破坏时可达（端点前置已取到 `apiKey.QuotaPlanID != nil`），本卡不修，建议另开 issue（候选：`xerror.WrapRecordNotExist("QuotaPlan")` → 404）。

## 5. 兼容性

- **API 契约**：仅 unlimited reset 单分支的 status/error 语义变化（500→422），无请求/响应字段增删，无 DB schema 变更。
- **调用方**：管理面消费者（Dashboard/脚本）按 envelope 错误分支处理，422 与 500 同属非 2xx，仅重试/告警语义被纠正（issue 核心诉求）。
- **数据平面**：控制面端点，BFE/Conf Agent 无涉。

## 6. 测试设计

### 6.1 单测

1. `model/quota/quota_plan_manager_test.go` —「unlimited plan cannot reset」用例强化：
   - 保留 `assert.Contains(err.Error(), "cannot reset balance for unlimited quota")`（包装后仍通过）；
   - 新增 `rr := xerror.Resolve(err); assert.Equal(422, rr.ErrNo); assert.Equal("Param Illegal", rr.Type)`；
   - 新增 `assert.Equal("cannot reset balance for unlimited quota", fmt.Sprintf("%v", xerror.Cause(err)))`。
2. `model/quota/operation_log_test.go` — unlimited reset 失败用例断言：记录 1 条 `ActionReset`、`Status=StatusFailed(2)`、`ErrorMsg` 非空且含原消息、`ResourceParentID=owner.ID`、`BuildChangeSummary` 前后快照不含 quota 变更。
3. 门槛：`make test-model-cover-gate`（model ≥70%）。

### 6.2 集成测试（新增）

`test/integration/tests/operation_log/` 下新增 `unlimited_reset/unlimited_reset_test.go`（仿 `nested_audit` 组织）：

- **OL-UR-001**：unlimited API-Key reset → HTTP 422、`ErrNum=422`、`ErrMsg` 前缀 `Param Illegal:`；操作日志 `quota_plan/reset` `status=2`、`error_msg` 非空、`resource_parent_id` = api-key id；随后 GET api-key 余额不变（unlimited sentinel）。
- **OL-UR-002**：unlimited Entity reset → 同上（entity 路径，验证同病同愈）。

### 6.3 回归与闭卡

1. `make test` 全绿。
2. QA requeue **SC2101-TC047** real-verification：`sc2101AuditRequireFailure`（`sc2101_operation_log_helpers.go:290`，`[400,500)`）由本修复满足，进入断言 3（status=2 + 非空错误原因），TC047 闭卡，#161 与 #183 同时关闭。
3. 手工核验 issue 复现步骤（建 unlimited key → reset → 422；窗口查日志 status=2；DELETE 清理）。

## 7. 文档同步（六步法 Step 3/4）

详见 [api-changes.md](api-changes.md)。要点：

- `api-keys.md` §2.8 / `entities.md` reset 节：修正「执行逻辑」第 1 步的错误语义描述并补错误码表（现存「返回404」文字与实现及 normative 契约卡均不符，属历史文档漂移）。
- `sys-design/模型层设计文档.md` ResetBalance 伪代码段：unlimited 分支同步为 PARAM 包装（该段签名未含 #161 的 owner 参数，属 #161 遗留文档债，不在本卡范围，仅备注）。
- `sys-design/details/配额余额同步机制.md` §6.4：补手动重置错误语义一小节。
