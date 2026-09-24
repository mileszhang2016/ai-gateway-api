# Issue #178：Entity `type` 在全量/部分更新中可被改写修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/178](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/178)

> SC1203-TC010（Entity 参数校验与失败原子性）`immutable-type` 负例：Entity 的 `type` 字段在全量更新（PUT §2.4）与部分更新（PATCH §2.5）接口中均应**不可修改**，但实际接口返回 `200` 并把 `type` 持久化改写为请求体中的新值。分类 `FAILED_PRODUCT`（Oracle-backed，不可放宽）。

接口文档 `design-docs/api-define/OpenAPI接口定义/entities.md` 的 Oracle 条款：

- §2.4 全量更新 Entity（line 391）约束：`type` 不可修改（创建后固定）。
- §2.5 部分更新 Entity（line 443）约束：`type` 不可修改。

复现路径（E2E SC1203-TC010）：

1. 创建两个 Entity-Type：root（level 1）、child（level 2）；
2. POST /entities 创建 Entity，`type=child`；
3. PUT /entities/{id} 将 `type` 改为 root → 返回 `ErrNum=200 success`；
4. GET 回读确认 `type` 已被持久化改写为 root——一个 level-2 child 被静默"提升"为 level-1 root，绕过 `parent_id` 的 level 校验，并使既有 parent_id 祖先链失效。

预期：PUT/PATCH 修改 `type` 应被拒绝（4xx），`type` 保持创建时的值。

## 2. 根因

`type` 从请求体到持久化全链路无任何不可变守卫：

1. **接口层**：`endpoints/openapi_v1/entity/full_update.go` 与 `update.go` 只做字段格式校验（`validateEntityParam(param, false)`），未将请求体 `type` 与库中现有值比对；
2. **模型层**：`model/entity/entity_manager.go` 的 `UpdateEntity`（:255）无 `type` 不可变检查。更严重的是，当请求同时携带 `type` + `parent_id` 时，`checkEntityLevel` 用**请求体的新 type** 做 level 校验（:273-278）——child 提升为 root 后新 type 是 level 1、parent 为 null 时根本不触发校验，守卫逻辑反而为改写背书；
3. **存储层**：`storage/rdb/entity/entity.go` 的 `entityDataToParamForUpdate` → `entityBaseDataToParam`（:185-195）无条件透传 `Type: param.Type`；DAO `internal.Update` 的 nil-skip 仅跳过 nil 列，非 nil 的新 `type` 被正常写入。同文件在 issue #152 修复了 `allow_models`/`block_models` 的"省略覆盖"，但 `type` 可写问题未覆盖。

对照：`Entity.Type` 决定层级语义（Entity-Type 的 `level`）与祖先链约束，属于创建时固定的不变量；同模块 `entity_type` 的更新路径不存在同类缺陷。

## 3. 目标

1. PUT / PATCH 请求体携带的 `type` 与库中现有值**不同**时，拒绝请求（4xx），且库中 `type` 保持不变；
2. 请求体携带的 `type` 与现有值**相同**时，放行（GET→修改→PUT 的回环场景必须可用，§2.4 约束是"不可修改"而非"不可携带"）；
3. PATCH 省略 `type` 时行为不变（保持原值）；
4. 回归验证：重跑 E2E SC1203-TC010 预期 PASSED，解除 issue #174 的 deployment gate 阻塞。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `endpoints/openapi_v1/entity/full_update.go`、`endpoints/openapi_v1/entity/update.go`、`model/entity/entity_manager.go` |
| 接口契约 | 请求/响应结构不变；仅将"PUT/PATCH 携带不同 `type`"从"静默改写"修正为"拒绝"。契约条款（§2.4/§2.5）文档中已存在，**无需改动 api-define** |
| 不在范围 | Create 路径；`parent_id` level 校验逻辑本身；已被改写的历史数据不回溯修复 |
| 数据迁移 | 无 |
| 不受影响面 | `model/imods` 导出/导入不调用 `UpdateEntity`，不受影响；Dashboard（ai-gateway-web）PUT 载荷回传 GET 得到的原 `type`，值相等自然放行 |

错误码选择：**422（PARAM，"Param Illegal"）**，与既有参数类违规一致（如 Create 时 `entity type not found` 走 `xerror.WrapParamError` → 422）。拒绝语义是"请求参数违反不可变契约"，而非与其他资源的冲突（409）。

## 5. 修复方案

分层防御：接口层 fail-fast + 模型层权威守卫（带审计），镜像既有 `checkEntityLevel` 失败分支的处理模式。

### 5.1 接口层（fail-fast）

`full_update.go` 与 `update.go` 中，`existing` 已在绑参前取出；在 `validateEntityParam` 之后、`UpdateEntity` 之前追加比对：

```go
if param.Type != nil && existing.Type != nil && *param.Type != *existing.Type {
    return nil, xerror.WrapParamErrorWithMsg("type is immutable, cannot be modified after creation")
}
```

- 提前拒绝，不进入模型层事务；
- 与接口层其余参数校验（`validateEntityParam`）同位置、同错误类别（PARAM → 422）；
- 接口层参数校验失败不记操作日志，与既有 endpoint 参数失败行为一致。

### 5.2 模型层（权威守卫 + 审计）

`model/entity/entity_manager.go` `UpdateEntity` 的**事务前置校验区**（入口只读查询 `oldEntity` 之后、`checkEntityLevel` 之前）追加：

```go
if param.Type != nil && oldEntity != nil && oldEntity.Type != nil && *param.Type != *oldEntity.Type {
    err := xerror.WrapParamErrorWithMsg("type is immutable, cannot be modified from %s to %s", *oldEntity.Type, *param.Type)
    entityID, entityName, parentID := resolveEntityIdentifiers(filter, param, oldEntity)
    m.recordEntityOperation(ctx, string(ioperlog.ActionUpdate), entityID, entityName, parentID, entityParamToMap(oldEntity), entityParamToMap(param), err)
    return 0, err
}
```

审计调用沿用同函数内 `checkEntityLevel` 前置失败分支的既有写法（`resolveEntityIdentifiers` 身份解析 + before/after 快照 + 错误记录）。

> 落位说明：相较于"事务内 `one := list[0]` 之后"的方案表述，最终放在**事务前置区、level 校验之前**——否则 `type`+`parent_id` 同传时会先用**新 type** 跑 level 校验，可能抛出误导性的 level 错误而非不可变错误；前置区是 `checkEntityLevel` 同模式的既有守卫位置，且仍在任何写入之前，失败原子性与审计语义不变。

模型层守卫是契约完整性的最终防线：

- 保护所有调用方（含未来新增调用路径），不只保护两个 OpenAPI handler；
- 拒绝发生在事务内任何写入之前，天然保证失败原子性（SC1203-TC010 同 TC 关注的属性）；
- 拒绝被记录操作日志，与 `checkEntityLevel` 失败同待遇。

### 5.3 存储层

不改。DAO nil-skip 语义不变；模型层保证到达存储层时 `param.Type` 要么为 nil、要么与库中值相等。

### 5.4 语义边界

| 场景 | 行为 |
|------|------|
| PUT/PATCH 携带 `type` ≠ 现有值 | 422 拒绝，库中值不变（接口层+模型层双重拦截） |
| PUT/PATCH 携带 `type` = 现有值 | 放行，幂等无变化 |
| PATCH 省略 `type` | 放行，保持原值（nil 跳过，DAO nil-skip 既有语义） |
| PUT 省略 `type` | 保持现状：不报错、不改写（§2.4 未将 `type` 列为更新必填；`name` 才是必填） |
| 存量数据 `type` 为空、请求设置非空 `type` | 值不等 → 拒绝（历史空值视为"创建时未固定"，不允许事后补写） |
| 请求同携 `type`（新值）+ `parent_id` | 先被不可变守卫拒绝，不会进入 `checkEntityLevel`；`type` 不可变后，:273-278 用新 type 做 level 校验的分支在"改 type"场景下成为死路，parent-only 修改场景行为不变 |

## 6. 单元测试

1. `model/entity/entity_manager_test.go`（mock storager 模式）：
   - `TestUpdateEntity_TypeChangeRejected`：`param.Type` 与库中不同 → 返回 PARAM 错误，storager 的 `UpdateEntity` 未被调用（断言 mock 调用次数为 0）；
   - `TestUpdateEntity_SameTypeAllowed`：相同 `type` → 正常更新；
   - `TestUpdateEntity_NilTypeAllowed`：PATCH 省略 `type` → 正常更新；
   - 失败分支操作日志记录符合既有断言模式。
2. `endpoints/openapi_v1/entity/validator_test.go`：若比对逻辑抽为独立函数（如 `validateTypeImmutable(param.Type, existing.Type)`），补表格用例（不同值/相同值/nil/空存量值）。
3. 既有测试回归：`go test ./model/entity/... ./endpoints/openapi_v1/entity/...` 全绿。

## 7. 集成 / E2E 回归

`test/integration/tests/entity/`（真实二进制子进程）：

- `full_update_test.go` 新增：PUT 携带不同 `type` → 断言 4xx（ErrNum=422）且 GET 回读 `type` 不变；PUT 携带相同 `type` → 200；
- `partial_update_test.go` 新增：PATCH 携带不同 `type` → 422 且 GET 回读不变；PATCH 省略 `type` → 200 且不变；
- 本地门禁：`go build ./...`、`go vet ./...`、`go test ./...`、`go test -cover ./model/...`（≥70%）全部通过；
- 部署后重跑 E2E SC1203-TC010 `immutable-type` 负例，预期 PASSED；SC1203-TC010 闭环后，issue #174 的 deployment gate 解除 pending。

## 8. 实施记录

- 代码实现（工作区已改，未提交）：
  - 接口层 fail-fast：`endpoints/openapi_v1/entity/validator.go`（`validateEntityParam` 内新增 type 不可变比对），`full_update.go` / `update.go` 在绑参后调用；
  - 模型层权威守卫：`model/entity/entity_manager.go` `UpdateEntity` 事务内、任何写入之前比对 `param.Type` 与库中值，拒绝时按 `checkEntityLevel` 失败分支的既有模式记录操作日志；
  - 存储层未改动（方案 §5.3）。
- 测试：`endpoints/openapi_v1/entity/validator_test.go`、`model/entity/entity_manager_test.go` 新增不可变负例/同值/省略用例；`test/integration/tests/entity/full_update`、`partial_update` 新增拒绝与放行用例；全量 `go test` 通过。
- sys-design 同步（本方案认定的两处轻量补充，见下节说明）：
  - `details/API-Key与Entity关联及模型继承.md` 新增 §3.4「不可变字段」；
  - `details/部分更新语义与DAO-nil-skip约定.md` 新增 §6，区分"省略保留"与"不可变拒绝"两种机制。
- 待办：提交代码与文档；部署后重跑 E2E SC1203-TC010 并解除 issue #174 deployment gate。

## 8. 实施记录

- 实现与本文档一致；唯一落位调整已在 §5.2"落位说明"标注：模型层守卫从事务内移到事务前置校验区（`checkEntityLevel` 之前）；
- 代码变更：
  - `endpoints/openapi_v1/entity/validator.go`：新增 `validateTypeImmutable(paramType, existingType *string) error`；
  - `endpoints/openapi_v1/entity/full_update.go`、`update.go`：`validateEntityParam` 之后调用 `validateTypeImmutable`（fail-fast）；
  - `model/entity/entity_manager.go`：`UpdateEntity` 前置校验区新增 type 不可变守卫，`resolveEntityIdentifiers` + `recordEntityOperation` 记录拒绝审计；
- 单测：`validator_test.go` 新增 `TestValidateTypeImmutable`（6 例）；`entity_manager_test.go` 在 `TestEntityManager_UpdateEntity` 新增 `type change rejected`（断言 storager 未被调用）/ `same type allowed` / `nil type allowed` 三个子用例；
- 集成测试：`full_update_test.go` 新增 E-4-007（PUT 改 type → 422 + 回读不变）、E-4-008（PUT 同 type → 200）；`partial_update_test.go` 新增 E-5-008（PATCH 改 type → 422 + 回读不变）、E-5-009（PATCH 省略 type → 不变）；
- 验证结果（2026-09-17，本地）：`go build ./...`、`go vet ./...`、`go test ./...` 全部通过；`go test -cover ./model/...` 总覆盖率 81.7%（≥70% 门禁）；集成测试 `tests/entity/full_update`、`tests/entity/partial_update` 全部 PASS（含新增 4 个回归用例，既有用例无回归）；
- 待办：部署到测试网关后重跑 E2E SC1203-TC010，翻转为 PASSED 后解除 issue #174 的 deployment gate。
