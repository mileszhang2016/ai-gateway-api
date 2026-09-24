# /epp-pool 统一组规模校验 设计变更说明

变更记录：`change-summary.md`。涉及分层：模型层（`model/epp_pool`）、基础设施装配（`stateful/`）。

## 1. 模型层（`model/epp_pool`）

### 1.1 删除 ValidationMode 概念

- `epp_pool.go`：删除常量 `ValidationModeProduction` / `ValidationModeTest`；删除 `ManagerOptions.ValidationMode` 字段及 `withDefaults` 中的缺省逻辑；删除 `EppPoolManager.validationMode` 字段。
- `assignment.go`：删除 `groupMeetsValidationMode`；`findValidAssignment` / `repairOne` / `allocateCluster` / `candidateGroups` 移除 `validationMode` 参数，所有 `m.validationMode` 传参点删除。

### 1.2 组规模校验（`validateGroupSize`）

**原：** 按模式分流——test 模式 `>= 1`；production 模式（及缺省）`== 2`。

**新：** 统一为 `len(group.Instances)` 必须为 1 或 2，空组与 ≥3 实例组拒绝，错误信息统一为 `"epp pool group %q requires 1 or 2 instances, got %d"`。

### 1.3 分配器候选组（`candidateGroups`）

**原：** 过滤出满足部署形态（production=2 / test≥1）的组，按组名排序。

**新：** 池中所有组即候选（池中组必然为 1~2 实例——PATCH 校验拒绝空组与 ≥3 实例组；`buildPool` 按实例分组天然不产生空组），仍按组名排序保证确定性。

### 1.4 悬空修复语义变化（`repairOne`）

| 场景 | 原行为（production 模式） | 新行为 |
|------|--------------------------|--------|
| 组仍在，primary 被移除 | 同组重选 primary（不换组） | 同组重选 primary（不换组）——该场景不经过组规模校验，行为本来就与模式无关 |
| 组缩容到 1 实例，primary 仍在 | **跨组重分配**（1 实例组规模不满足） | 分配保留（1 实例组合法） |
| 组缩容到 1 实例，primary 被移除 | 同组重选 primary | 不变（同组剩余实例重选） |
| 组已不存在 | 跨组重分配；无候选则清除 | 不变 |

### 1.5 配置装配

- `stateful/config.go`：`RunTimeConfig` 删除 `EPPValidationMode` 字段。
- `stateful/container/rdb/components.go`：`ManagerOptions` 删除 `ValidationMode` 注入。

## 2. 测试调整

| 测试文件 | 调整 |
|----------|------|
| `model/epp_pool/epp_pool_manager_test.go` | `testManager` 去掉 `ValidationMode`；用例表删除 `"production requires exactly 2"`（单实例组已合法），新增 `"three instances rejected"`（3 实例组 422）、保留空组（`instances: []`）422 用例；删除 `"test mode allows single instance group"` 子用例（单实例组已是默认合法行为，由 `TestPatchPool_FullReplace` 等覆盖）；`TestManagerOptionsDefaults` 删除 validationMode 断言；`TestRepairDangling` 中 `"clear when pool has no candidate groups"` 改为空池场景（原 `// undersized` 单实例池现在是合法候选，须改为不播种实例 + 预置悬空 assignment） |
| `model/epp_pool/assignment_test.go` | 所有 `allocateCluster`/`repairOne` 调用去掉 mode 参数；`TestAllocateCluster_CandidateFilter` 改为验证"池中全部组均为候选（含单实例组）"，无候选场景改用空池（`makePool()`）；`TestRepairOne` 中 `"cross group when group undersized"` 语义反转——组缩容到 1 实例且 primary 仍在 → `repairKeep`（分配保留），跨组重分配仅保留"组已不存在"场景；`newManager` 去掉 `ValidationMode` |
| `model/epp_pool/epp_data_test.go` | 去掉 `ManagerOptions{ValidationMode: ...}` |
| `endpoints/openapi_v1/epp_pool/endpoints_test.go` | `setupManager` 去掉 mode 参数；`TestPatchAction_GroupSizeProduction` 改为验证单实例组 PATCH 成功（原 422 场景已合法），保留空组 422 校验 |
| `endpoints/openapi_v1/epp_assignments/endpoints_test.go` | 去掉 `ManagerOptions{ValidationMode: ...}` |
| `endpoints/innerapi_v1/epp_data/export_test.go` | 同上 |
| `model/icluster_conf/epp_test.go` | 同上 |

**集成测试（`test/integration/`，本次新增）**

| 测试文件 | 调整 |
|----------|------|
| `tests/epp_pool/update/update_test.go` | EP-2-013 改名（去掉"test 模式"措辞）；新增 `TestEppPool_Update_GroupSize`：EP-2-014（3 实例组 422 且全量替换不生效、池内容回读不变）、EP-2-015（双实例组主+备通过） |
| `tests/epp_assignments/get/get_test.go` | 新增 `TestEppAssignments_GroupShrinkRepair`：EA-1-005（组缩容到单实例且保留 primary → 分配不换组不换主、standby 消失）、EA-1-006（primary 被移除 → 同组重选、不换组） |

## 3. 文档同步

- `api-define/OpenAPI接口定义/epp-pool.md`：见 `api-changes.md`。
- `sys-design/模型层设计文档.md`、`sys-design/接口层设计文档.md`、`sys-design/数据库设计文档.md`、`sys-design/details/EPP调度对接.md`：将"生产恰 2 实例 / 测试允许单实例组 / 部署形态配置项"相关描述统一改为"每组 1~2 个实例（1=仅主，2=主+备），拒绝空组与 3 个及以上实例的组"。

## 4. 非目标

- 不改 `epp_instances` / `epp_assignments` 表结构。
- 不改 `epp_data` 导出格式与 `EPPAddr` 生成逻辑（standby 解析、primary-only 导出均保持现状）。
- 不改 BFE / EPP 侧任何代码。
