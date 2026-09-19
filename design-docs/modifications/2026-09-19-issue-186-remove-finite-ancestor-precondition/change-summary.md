# 变更摘要：删除 API-Key 挂载前置条件中的"finite 祖先 Quota Plan"要求（Issue #186）

## 背景

GitHub Issue [rainway-ai-gateway/ai-gateway-api#186](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/186) 指出设计文档内部存在自相矛盾：

- `api-define/OpenAPI接口定义/api-keys.md` §2.4（PUT，line 347）与 §2.5（PATCH，line 399）规定：若将 `entity_id` 修改为非空，且 `unlimited_quota=false` 且 `quota_plan.unlimited=false`，则要求新 Entity 或其祖先链上至少存在一个有效的（finite）Quota Plan，否则应被拒绝。
- 但 `workflows.md` §5（运行时配额扣减流程）step 4a 规定：API-Key 自身的 finite `quota_plan` **无条件**加入扣减列表并扣减其自身 QuotaBalance，与祖先链是否有 finite Plan 无关。
- `object-relations.md` line 174/183/186 进一步确认：API-Key 必须自带 1 个 QuotaPlan，QuotaPlan 与 QuotaBalance 一一对应，运行时生效 Plan = Key 自身 + Entity + 祖先（去重）。

因此 §2.4/§2.5 的"须有 finite 祖先"前置条件在运行时是死条款：finite Key 挂到全 unlimited 的 Entity 下，仍受自身 QuotaPlan/QuotaBalance 约束，不会逃逸为无限；自身 remaining 耗尽时由 §5 step 6/8 自然拒绝（429002），与挂载目标无关。

同时 §2.2.1 create（line 132）只要求"Entity 必须存在"，与 §2.4/§2.5 对同一语义状态给出相反裁决。

## 目标

以 `workflows.md` §5 为准，删除 §2.4/§2.5 的 finite 祖先前置，将 create / PUT / PATCH 三处的挂载校验统一为"**该 Entity 必须存在**"。

## 关键决策

- **采纳 Issue 中的更正方向 A（对齐 §5）**：删除前置条件。方向 B（改 §5 step 4a 不再扣 Key 自身 plan，改为从 entity 链取最近 finite pool）会推翻 object-relations 与整个现有扣减模型，不可取。
- **纯文档修复，代码无需改动**：现有实现（`endpoints/openapi_v1/api_key/update.go:82`、`model/api_key/api_key.go:657-664`）本来就只校验 entity 存在性，未实现 finite 祖先前置。修复后代码行为即合规。
- **§2.2.1 create（line 132）保持不变**，其"Entity 必须存在"即为统一后的正确基线。

## 影响范围

| 对象 | 影响 |
| - | - |
| `api-define/OpenAPI接口定义/api-keys.md` | §2.4 line 347、§2.5 line 399 约束条款改写（见 api-changes.md） |
| `api-define/OpenAPI接口定义/workflows.md` | 无改动（作为裁决基准） |
| `api-define/OpenAPI接口定义/object-relations.md` | 无改动（作为裁决基准） |
| 产品代码 | 无改动（已全仓核实，见下节"代码与测试核实结论"） |
| ai-gateway-api 单元测试 | 无改动（已核实，见下节） |
| integration-test 仓库 | 无改动（已核实，见下节） |
| SC1301-TC028 等 E2E 用例及其设计稿 | 设计依据被撤销：原断言"finite Key 挂载到全 unlimited Entity 应被拒绝"失去依据，需按"设计修订"路径处理（reclassify FAILED_PRODUCT → FAILED_TEST，删除/放宽 put-invalid-leaf、patch-invalid-leaf 等断言后重跑）。**该 E2E 资产不在本工作区**（见下节），需在其所在仓库处理。 |

## 代码与测试核实结论

对"是否无需改代码和测试"做了全面核实（而非仅依据 Issue 转述）：

**产品代码——确认无需改动。** 全仓检索 `ancestor` / `祖先` / `有效的Quota Plan` / `finite`，无任何代码实现该前置条件。挂载校验仅查 Entity 存在性：
- create：`endpoints/openapi_v1/api_key/create.go:71-77`
- PUT / PATCH：`endpoints/openapi_v1/api_key/update.go:82`、`model/api_key/api_key.go:657-664`
- `checker.go` 仅校验 `quota_plan` 字段自身合法性，与 Entity 链无关。

**ai-gateway-api 单元测试——确认无需改动。** `api_key_test.go:481-499` 仅覆盖 "entity not found" 场景；无任何用例断言"finite Key 挂载到全 unlimited Entity 应被拒绝"。

**integration-test 仓库——确认无需改动。** 该仓库无 SC1301 编号体系（场景编号至 SC29），`put-invalid-leaf` / `patch-invalid-leaf` / `TC028`（SC1301 语境）全仓无匹配；SC14（QuotaPlan 层级继承）场景的 4 个用例亦无挂载拒绝类断言。

**E2E 资产缺失——需外部跟进。** Issue 提及的 `docs/test-design/SC1301/SC1301-TC028.md`（line 49 "不得放宽为成功"特别裁决）在 `ai-gateway-api`、`integration-test` 及整个工作区均不存在（ai-gateway-api 无 `docs/test-design/` 目录）。该 E2E 套件应维护于本工作区之外的仓库或测试管理平台，需在其所在处按第四节指引修订。

## 文档变更明细

详见同目录 [api-changes.md](./api-changes.md)（本次变更属于 api-define 契约修订，无数据模型/流程变化，故无需 design-changes.md）。

## 后续事项

1. ~~按六步法 Step 3 执行 api-keys.md 修改并 review~~ **已完成（2026-09-19）**：§2.4 line 347、§2.5 line 400 已按 api-changes.md 改写，全文件复核无残留旧条款，create（line 132）/ PUT / PATCH 三处口径已统一；
2. SC1301-TC028 的 TC 设计稿（line 49 "不得放宽为成功"特别裁决）随本次更正同步调整——经全工作区核实，该文件不在 `ai-gateway-api`、`integration-test` 或本工作区任何位置，确认其所在仓库后同步修订；
3. 集成测试中 patch-invalid-leaf / put-invalid-leaf（及未达的 POST-create-reject）断言按修订后的设计删除或放宽，重新执行用例（同样在 E2E 资产所在仓库执行）。
