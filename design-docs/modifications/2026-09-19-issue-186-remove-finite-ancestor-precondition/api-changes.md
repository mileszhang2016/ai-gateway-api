# API 变更说明：API-Key 挂载校验统一为"Entity 必须存在"（Issue #186）

对应 Issue：[rainway-ai-gateway/ai-gateway-api#186](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/186)

变更文件：`api-define/OpenAPI接口定义/api-keys.md`（仅此一处；接口路径、方法、参数、错误码均无变化）

## 变更 1：§2.4 全量更新API-Key（PUT /api-keys/{id}）约束条款

位置：`api-keys.md` line 347（"约束"小节第二条）

**修改前**

> - 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan。

**修改后**

> - 若将 `entity_id` 修改为非空（挂载到新Entity），该Entity必须存在。
>
> 注：配额扣减遵循 [workflows.md](./workflows.md) §5——当 `unlimited_quota=false` 且 `quota_plan.unlimited=false` 时，Key 自身的 QuotaPlan 已计入扣减列表并扣减其自身 QuotaBalance，因此不要求新Entity或其祖先链上存在 finite 的 Quota Plan。若 Key 自身余额不足，请求按 §5 step 6/8 拒绝（429002），与挂载目标的配额类型无关。

## 变更 2：§2.5 部分更新API-Key（PATCH /api-keys/{id}）约束条款

位置：`api-keys.md` line 399（"约束"小节第二条）

**修改前**

> - 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan。

**修改后**

> - 若将 `entity_id` 修改为非空（挂载到新Entity），该Entity必须存在。不要求新Entity或其祖先链上存在 finite 的 Quota Plan（依据 [workflows.md](./workflows.md) §5）。

## 不变更项

- **§2.2.1 create（line 132）**："若 `entity_id` 不为空，该Entity必须存在。" 已是统一后的正确基线，保持不变。
- **错误码**：无新增/变更。Entity 不存在时仍按现有实现返回参数错误（422）；配额耗尽仍由运行时按 workflows.md §5 返回 429002，与本接口无关。
- **请求/响应字段**：无变化。
- **`workflows.md` §5、`object-relations.md`**：作为本次更正的裁决基准，不改动。

## 修复后一致性

create / PUT / PATCH 三处挂载校验统一为只校验 Entity 存在性；finite 祖先前置统一删除，与运行时扣减模型（Key 自身 Plan 无条件入列扣减）自洽。

## Review 检查单

- [x] 接口路径、方法、参数与变更说明一致（无变化）；
- [x] 字段命名符合现有规范（无新增字段）；
- [x] 错误码覆盖无新增失败场景（finite 祖先拒绝场景被有意移除，属设计修订而非遗漏）；
- [x] 不影响已有接口兼容性：该前置条件在代码中从未实现（`endpoints/openapi_v1/api_key/update.go:82`、`model/api_key/api_key.go:657-664` 仅校验 entity 存在性），因此本次修订是文档向实现对齐，不存在行为变更。

> 执行记录（2026-09-19）：§2.4 与 §2.5 约束条款已按本文"变更 1/变更 2"落地到 `api-keys.md`（PUT 位于 line 347，PATCH 位于 line 400，因 §2.4 新增注记行号整体后移）。复核确认旧条款在 `api-keys.md` 中已无残留，§2.2.1（line 132）保持原文未动。
