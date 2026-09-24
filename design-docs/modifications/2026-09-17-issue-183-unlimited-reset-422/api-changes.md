# Issue #183 API 变更说明：quota-plan/reset 错误契约修正

## 1. 变更接口

| 端点 | Method | 变更类型 |
|------|--------|---------|
| `/open-api/v1/api-keys/{id}/quota-plan/reset` | POST | 错误语义修正（单分支） |
| `/open-api/v1/entities/{id}/quota-plan/reset` | POST | 错误语义修正（单分支，同根因同修复） |

请求 / 成功响应字段：无变化。

## 2. 错误契约变化

仅「目标配额计划 `unlimited=true`」这一确定性语义错误分支变化：

| 项目 | 修复前 | 修复后 |
|------|--------|--------|
| HTTP status | 500 | 422 |
| `ErrNum` | 500 | 422 |
| `ErrMsg` | `Unknown Exception: cannot reset balance for unlimited quota` | `Param Illegal: cannot reset balance for unlimited quota` |
| 操作日志 | `quota_plan/reset` `status=2`，error_msg 为裸消息 | `quota_plan/reset` `status=2`，error_msg=`PARAM: cannot reset balance for unlimited quota`（非空、稳定），其余字段不变 |

其余失败分支语义不变：

| 场景 | 语义 | 现状 |
|------|------|------|
| API-Key / Entity 不存在 | 404 Record Not Exist | 不变 |
| 资源无 quota plan | 422 Param Illegal | 不变（本次修复即与之对齐） |
| `quota` 参数非法（负数 / 精度越界等） | 422 Param Illegal | 不变 |
| 配额计划行缺失（引用完整性破坏） | 500 Unknown Exception | 不变（本卡不修，见 change-summary §3 兄弟缺陷） |
| DB / Redis 内部故障 | 500 | 不变 |

规范依据：SC2101-TC047 契约卡（`coordinator_approved` / normative）——管理写入遵循 2xx/4xx/422 语义，失败审计保留稳定 status/error 语义。

## 3. api-define 文档同步（Step 3）

### 3.1 `design-docs/api-define/OpenAPI接口定义/api-keys.md` §2.8

「执行逻辑」第 1 步现状文字：

> 找到该API-Key的quota_plan（如果不存在或unlimited=true，返回404）

与实现及 normative 契约均不符（「无 quota plan」实现为 422，unlimited 修复后亦为 422；404 仅对应路径上的 API-Key 本身不存在）。修改为：

> 1. 若 API-Key 不存在，返回 404；找到该 API-Key 的 quota_plan：若未关联 quota plan 或 `unlimited=true`，返回 422（Param Illegal，不产生配额变更，并记录 `status=2` 的操作日志）；若传入 quota，更新 quota_plan.quota 为新的值

并在 §2.8 末尾补充**失败响应说明**小节：

| HTTP status | ErrNum | ErrMsg | 触发条件 |
|------|------|------|------|
| 404 | 404 | `API-Key Record Not Exist` | 路径中的 API-Key 不存在 |
| 422 | 422 | `Param Illegal: <原因>` | 未关联 quota plan；unlimited 配额计划；quota 参数非法 |
| 500 | 500 | `Unknown Exception: <原因>` | 内部故障 |

### 3.2 `design-docs/api-define/OpenAPI接口定义/entities.md` reset 节

同构修改（「找到该Entity的quota_plan（如果不存在或unlimited=true，返回404）」→ 与 3.1 一致的错误语义描述 + 失败响应说明，404 文案为 `Entity Record Not Exist`）。

### 3.3 sys-design 同步（Step 4）

- `design-docs/sys-design/模型层设计文档.md` ResetBalance 伪代码（`:1820-1851`）：unlimited 分支 `return fmt.Errorf(...)` 同步为 `return xerror.WrapParamErrorWithMsg("cannot reset balance for unlimited quota")`，并加一行注释说明 PARAM → 422 映射。（该段未含 #161 引入的 `owner` 参数，属 #161 遗留文档债，本卡不展开。）
- `design-docs/sys-design/details/配额余额同步机制.md` §6.4「手动重置接口」末尾补一段「错误语义」：

> 手动重置为管理写入，失败遵循 2xx/4xx/422 语义：目标计划 `unlimited=true` 时返回 422 `Param Illegal: cannot reset balance for unlimited quota`，不产生配额变更，并在事务返回后记录 `quota_plan/reset`、`status=2`、error_msg 非空、resource_parent_id 为 owner 业务 ID 的操作日志（SC2101-TC047 断言 3）。

## 4. 兼容性结论

- envelope 结构、字段名、成功响应：零变化。
- 仅 unlimited reset 分支 500→422；对按 `ErrNum != 200` 判错的调用方透明，对按 5xx 告警的监控是**修正性**变化（消除确定性语义错误的故障误报）。
- 无版本协商、无迁移步骤。
