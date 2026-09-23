# Issue #199：PUT 全量更新 API-Key 漏校验 entity_id 存在性 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api#199](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/199)（P0，SC1301-TC028 A-03，失败指纹 `428f6fc1`，verified_commit `0d362b03`）：

> 对已存在的 API Key 执行 `PUT /open-api/v1/api-keys/{key_id}`（全量更新），请求体中 `entity_id` 指向**不存在的 Entity** 时，接口返回 `HTTP 200 / ErrNum=200` 并接受挂载；同一 run 内同一 Key 用同一不存在 `entity_id` 执行 `PATCH` 则被正确拒绝（受控 4xx）。PUT/PATCH 校验路径不对称，PUT 漏校验 `entity_id` 存在性。

违反的产品条款（均为 released / normative，**修复的规范性依据，无需改设计文档**）：

- `api-keys.md` §2.4（line 347，PUT）："若将 `entity_id` 修改为非空（挂载到新Entity），该Entity必须存在。"
- `api-keys.md` §2.5（line 400，PATCH）：同条款（PATCH 已合规，作为 PUT 的对照基线）。
- `api-keys.md` §2.2.1（line 132，create）："若 `entity_id` 不为空，该Entity必须存在。"
- `object-relations.md#关系说明`：非空 `entity_id` 只要求 Entity 存在；SC1301-TC028 "特别裁决"：只有不存在 Entity 被接受才属 `FAILED_PRODUCT`。

## 2. 根因（源码定向定位）

**PATCH 有校验**（`endpoints/openapi_v1/api_key/update.go:75-84`，`APIKeyUpdateProcess` 内，位于 FetchAPIKey 之后、`UpdateAPIKey` 写库之前）：

```go
// 检查 entity_id 是否存在（如果传入的话）
if param.EntityID != nil && *param.EntityID != "" {
	entity, err := container.EntityManager.FetchEntity(ctx, &entity.EntityFilter{EntityID: param.EntityID})
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, xerror.WrapParamErrorWithMsg("%s", fmt.Sprintf("Entity not found: %s", *param.EntityID))
	}
}
```

**PUT 没有该校验**（`endpoints/openapi_v1/api_key/full_update.go:57-93`，`APIKeyFullUpdateProcess`）：流程为 `checkFullUpdateAPIKey`（`checker.go:115-124`，仅字段级合法性，最终复用 `checkUpdateAPIKey`，**不含任何存在性检查**）→ FetchAPIKey 确认 Key 存在 → 直接 `UpdateAPIKey` 写库。不存在的 `entity_id` 由此落库，形成悬空挂载。

**引入过程**（git 考古）：`full_update.go` 自 de13506 引入后即无该校验；f2c73d8 "update test caese and fix some bugs" 给 create.go 与 update.go 同时补上了 "Entity not found" 校验，**唯独漏了同目录的 full_update.go**——不对称自该提交起存在。

**model 层现状**：`CreateAPIKey` 有 entity 存在性检查（`model/api_key/api_key.go:657-666`），`UpdateAPIKey` 没有；`UpdateAPIKey` 的现网调用方仅 openapi 的 PATCH、PUT 两处（已全仓核实，无 innerapi 调用方）。

## 3. 同类排查结论（Issue "待确认"项的核实）

Issue 要求确认 #147 省略覆盖家族（API Key#151、Entity#152、model-price#154）是否存在同型 PUT 漏校验。已对全仓 `MethodPut` 路由逐一核实：

| PUT 路由 | 位置 | 结论 |
| - | - | - |
| PUT /api-keys/{id} | `api_key/full_update.go:34` | **存在本缺陷**（create/PATCH 校验、PUT 不校验） |
| PUT /entities/{id} | `entity/full_update.go:30` | 无不对称：PUT/PATCH 同走 `validateEntityParam`，model 层 `UpdateEntity` 校验 parent 存在性与层级（`model/entity/entity_manager.go:284-303`） |
| PUT /epp-assignments/{cluster} | `epp_assignments/put.go:38` | 无不对称（该域仅有 PUT）：cluster 存在性由 model 层 `OverrideAssignment` 校验（`model/epp_pool/assignment.go:241-266`，不存在返回 404 语义错误） |
| PUT /global-route-rules | `global_route_rules/endpoints.go:42` | 自包含配置（RouteRules 为独立配置对象），无跨对象引用存在性语义 |
| PUT /model-prices（两条） | `model_price/update.go:31,39` | 无不对称（该域仅有 PUT）：provider/model/mode 为记录自身标识三元组，create 同样不做存在性校验，无 PUT 特有缺口 |
| PUT /providers/{name}/pricing-tiers | `provider/pricing_tiers.go:35` | provider 经 `FetchProvider` 校验（`:85`） |

**结论：api_key PUT 是本类缺陷在全仓的唯一实例**，无需扩散修改。附带说明：api_key 的 `quota_plan`/`rate_limit_policy`/`route_rules` 为 Key 自包含的内嵌配置（由 model 层为每个 Key 独立建记录），不属于跨对象引用，不存在同类存在性校验问题，不在本次范围。

## 4. 目标

1. `PUT /api-keys/{id}` 携带非空且不存在的 `entity_id` 时，返回**与 PATCH 完全一致**的受控 4xx（HTTP 422 / etParam，msg `Entity not found: <id>`），且**不执行任何写库**，原 `entity_id` 绑定保持不变；
2. `entity_id` 为空字符串（解绑）或不传时行为不变（不触发存在性检查）；
3. 修复具备单测覆盖（model 层，可进 `make test` 门禁），集成用例补齐 PUT 缺失场景；
4. SC1301-TC028 A-03（put-missing-entity）requeue 转绿，SC1301-TC027 回归通过。

## 5. 范围

| 范围 | 说明 |
| - | - |
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `endpoints/openapi_v1/api_key/full_update.go`（主修复）、`model/api_key/api_key.go`（model 层兜底）、`model/api_key/api_key_test.go`、`test/integration/tests/api_key/full_update/full_update_test.go` |
| 设计文档 | api-define 契约（api-keys.md §2.2.1/§2.4/§2.5）已明确且现行，**无需修改**，本修复是使代码兑现契约；sys-design 仅 `details/API-Key与Entity关联及模型继承.md` §4.1 追加一句两层守卫说明（仿 §3.4 issue #178 体例），api-changes.md / design-changes.md 无需新增 |
| 接口契约 | 无变化：422/etParam 响应码型 PATCH 已在用，PUT 对齐之；成功路径响应结构不变 |
| 不在范围 | SC1301-TC028 等 E2E 资产修订（该资产不在本工作区，见 §7）；端点层 PATCH 冗余检查的收敛重构（见 §8 备选） |
| 数据迁移 | 无（存量悬空绑带的排查建议见 §6.4） |

## 6. 最终方案

两层修复：**端点层补齐（主修复，直接消除 PUT/PATCH 不对称）+ model 层兜底（防御纵深，与 Create 对称）**。

### 6.1 端点层主修复（`endpoints/openapi_v1/api_key/full_update.go`）

在 `APIKeyFullUpdateProcess` 中，FetchAPIKey 确认 Key 存在（`full_update.go:69-71`）之后、`container.APIKeyManager.UpdateAPIKey` 调用之前，插入与 `update.go:75-84` **逐字相同**的校验块：

```go
	// 检查 entity_id 是否存在（如果传入的话）
	if param.EntityID != nil && *param.EntityID != "" {
		entity, err := container.EntityManager.FetchEntity(ctx, &entity.EntityFilter{EntityID: param.EntityID})
		if err != nil {
			return nil, err
		}
		if entity == nil {
			return nil, xerror.WrapParamErrorWithMsg("%s", fmt.Sprintf("Entity not found: %s", *param.EntityID))
		}
	}
```

- **错误语义与 PATCH 逐字一致**：同一 `WrapParamErrorWithMsg` → 同一 HTTP 422/etParam、同一 msg 形态，E2E 以 PATCH 为对照基线即可直接比对；
- 校验位于所有写操作（`UpdateAPIKey`、`ApplyQuotaPlanChange`）之前，拒绝时**原绑定与配额余额均不变**，满足"不得改变原绑定"条款；
- `entity_id=""`（解绑）与 nil 不触发检查，与 create/PATCH 语义一致；
- 代价为每次带 `entity_id` 的 PUT 多一次 Entity 主键查询——PATCH 本已支付同等代价，可忽略。

### 6.2 model 层兜底（`model/api_key/api_key.go` `UpdateAPIKey`）

在 `UpdateAPIKey` 的 `AtomExecute` 事务内、任何写操作（quota_plan / rate_limit_policy / route_rules 的 Create 与 api_keys 表 Update）之前，加入与 `CreateAPIKey:657-666` 同构的检查：

```go
	// Check if entity_id exists
	if param.EntityID != nil && *param.EntityID != "" && rppm.entityStorager != nil {
		entity, err := rppm.entityStorager.FetchEntity(ctx, &shared.EntityFilter{EntityID: param.EntityID})
		if err != nil {
			return err
		}
		if entity == nil {
			return xerror.WrapParamErrorWithMsg("%s", fmt.Sprintf("Entity not found: %s", *param.EntityID))
		}
	}
```

- 依赖已就绪：`NewAPIKeyManager` 已注入 `EntityStorager`（`stateful/container/rdb/components.go:342`），`nil` 守卫与 Create 路径一致（单测构造的可选依赖缺省场景不 panic）；
- 使 Create/Update 在 model 层获得对等的存在性防护，未来新增 `UpdateAPIKey` 调用方（inner API、新端点）无需各自记得校验；
- 与端点层检查并存后，PATCH 的端点检查成为冗余，但两者错误类型/消息逐字相同、行为零差异，**保留以最小化 P0 修复的爆破半径**（收敛见 §8 备选）。

### 6.3 测试

1. **model 单测**（`model/api_key/api_key_test.go`，进 `make test` / 覆盖率门禁）：
   - 为 `UpdateAPIKey` 新增 "entity not found" 用例：fake entityStorager 返回 nil，断言返回 err 且消息含 `Entity not found: <id>`，并断言 storager 的 Update 未被调用（拒绝即无写）——仿既有 create 用例（`:481-499`）的 fake 构造；
   - 新增反向用例：`entity_id` 为 nil、为 `""` 时不触发 entity 查询、更新正常落库；
   - 既有 UpdateAPIKey 用例（`:283-427` 等）构造的 param 大多不带 `entity_id`，model 兜底对其无影响；带非空 `entity_id` 的既有用例需补 fake entityStorager 返回存在实体（逐项跑红后再补，不允许删除既有断言）。
2. **端点集成测试**（`test/integration/tests/api_key/full_update/full_update_test.go`，需 live 环境）：
   - 新增用例：对已有 Key 执行 PUT，body 携带 run-scoped 唯一的不存在 `entity_id`（命名风格对齐 `sc1301-tc028-missing-entity-<runID>`），断言 HTTP 422（与 PATCH 同码），随后 GET `/api-keys/{id}` 读回 `entity_id` 等于原绑定；
   - 对照用例：PUT 携带 `entity_id=""` 解绑成功（防修复误伤解绑路径）。
3. **E2E**：SC1301-TC028 requeue（A-03 put-missing-entity 断言转绿），并回归 SC1301-TC027；E2E 资产（`docs/test-design/SC1301/`、`test-cases/SC1301/`）不在本工作区，在其所在仓库执行，无需改动。

### 6.4 存量数据处置（建议，可选）

修复只阻断新增悬空绑定。缺陷存续期间经 PUT 写入的 `api_keys.entity_id` 悬空记录可按如下口径排查（MySQL，表名以 `db_ddl.sql` 为准，`api_keys:298`、`entities:389`）：

```sql
SELECT k.id, k.entity_id, k.updated_time
FROM api_keys k
LEFT JOIN entities e ON k.entity_id = e.entity_id
WHERE k.entity_id IS NOT NULL AND k.entity_id <> '' AND e.entity_id IS NULL;
```

由运维定夺逐条处置（解绑置空 / 补建对应 Entity）；悬空绑定的下游影响面（mod-api-key 导出、`api-keys?entity_id=` 过滤、配额扣减找不到 QuotaBalance）随处置自然消除。

## 7. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`go test -cover ./model/...`（≥70% 门禁）全部通过；
2. 单测：§6.3-1 新增用例通过，既有 api_key 用例零回退；
3. 手动对照（本地起服务）：
   - 同一不存在 `entity_id` 分别 PUT / PATCH 同一 Key → 均 HTTP 422、msg 均为 `Entity not found: <id>`；
   - 拒绝后 GET 读回 `entity_id` 为原值；PUT `entity_id=""` 解绑成功；PUT 已存在 entity 200；
4. 集成测试 §6.3-2 通过；部署门禁 SC1301-TC028 requeue A-03 转绿 + SC1301-TC027 回归通过。

## 8. 备选方案（不采纳）

| 备选 | 不采纳理由 |
| - | - |
| 仅端点层修复，不加 model 兜底 | 满足本 issue，但 `UpdateAPIKey` 对未来调用方仍无防护，且 Create 双层 / Update 单层的防护不对称依旧存在 |
| 校验收敛至 model 单层，删除端点层 PATCH/PUT 检查 | 行为等价（错误消息逐字相同）但 P0 修复不宜改动现网 PATCH 路径的实现层级；收敛留给后续重构单独评审 |
| PUT 返回 404 而非对齐 PATCH 的 422 | PUT/PATCH 同条款同语义，响应必须一致；§2.4 未规定具体错误码，受控 4xx 即合规；PATCH 改 404 会破坏既有客户端与 E2E 基线 |
| 仅在 `checkFullUpdateAPIKey`/`checkUpdateAPIKey` 内做校验 | checker 为无 ctx 的纯字段级校验函数，无法访问存储；存在性校验属编排逻辑，留在 process/manager 层符合本仓库分层惯例 |

## 9. 实施记录（2026-09-23，已完成）

按 §6 方案完成实施，四处改动：

| 文件 | 改动 |
| - | - |
| `endpoints/openapi_v1/api_key/full_update.go` | §6.1：FetchAPIKey 之后、`UpdateAPIKey` 之前插入与 PATCH（`update.go:75-84`）逐字相同的 entity 存在性校验；新增 `fmt`、`model/entity` 导入 |
| `model/api_key/api_key.go` | §6.2：`UpdateAPIKey` 事务内、`param.Key = nil` 之后、任何写操作之前，加入与 `CreateAPIKey` 同构的 entity 存在性检查（`rppm.entityStorager` 守卫同 Create） |
| `model/api_key/api_key_test.go` | §6.3-1：`TestAPIKeyManager_UpdateAPIKey` 新增 2 个子用例——`entity not found`（断言报 `Entity not found: e1` 且 storager Update 未被调用）、`empty entity id skips existence check`（断言空串不触发 FetchEntity）；nil 场景由既有 `success` 用例覆盖 |
| `test/integration/tests/api_key/full_update/full_update_test.go` | §6.3-2：新增 `AK-4-006`（PUT 不存在 entity_id → 422，GET 读回绑定不变）与 `AK-4-007`（PUT `entity_id=""` 解绑成功，防误伤解绑路径） |
| `design-docs/sys-design/details/API-Key与Entity关联及模型继承.md` | §4.1 代码片段后追加一句守卫分层说明（接口层 create/PUT/PATCH fail-fast + 模型层 `CreateAPIKey`/`UpdateAPIKey` 权威校验，issue #199；空串解绑不触发），仿 §3.4 issue #178 体例；sys-design 其余文件经核实无失实内容，未动 |

验证结果（本节替代 §7 计划清单，逐项已执行）：

1. `go build ./...`、`go vet ./...` 通过；
2. 主模块 `go test ./...` 全量通过（含新增 model 单测）；
3. model 覆盖率门禁：82.1% ≥ 70% 通过；
4. 集成测试：`tests/api_key/...` 全模块（create/delete/detail/full_update/list/partial_update/quota_query/quota_reset/quota_update）与 `tests/innerapi/mod_api_key/` 全部通过，新增 AK-4-006/AK-4-007 通过——修复前 AK-4-006 必失败（PUT 返回 200），可证其回归效力；
5. 说明：实施过程中观察到 `model/ioperlog` 的 `TestOperationLogManager_RecordAsync` 一次时序抖动失败（异步 flush 50ms 超时 + 150ms sleep 的时敏用例，本 diff 未触及该包），隔离复跑 5 次均通过，判定为既有 flaky，与本修复无关。

待办：E2E SC1301-TC028 requeue（A-03 转绿）与 SC1301-TC027 回归在其所在仓库执行；存量悬空绑定排查按 §6.4 由运维定夺。
