# /epp-pool 统一组规模校验（删除 EPPValidationMode）变更摘要

## 1. 背景

`/epp-pool` 的组规模校验当前由部署形态配置项 `RunTime.EPPValidationMode` 控制：

- `"production"`（默认）：每组恰 2 实例（主备）；
- `"test"`：允许单实例组（仅主、无备）。

该配置在 `2026-09-08-epp-scheduling-integration` 中引入。实际使用中发现：

1. "每组 2 实例（主备）"是**部署形态建议**而非 API 层必须强制的契约——测试环境、PoC 环境、单实例部署同样需要合法工作；
2. 配置项增加了部署与运维复杂度（同一套软件因环境不同行为不同，且校验失败信息依赖配置才能解释）；
3. EPP 侧/BFE 侧对单实例组已有完整支持（standby 缺省时导出 `EPPAddr` 仅含 primary），控制面无必要再拦截。

因此决定**删除该配置项**，组规模校验统一为"每组 1 个（仅主）或 2 个（主+备）实例"。

## 2. 目标

1. 删除配置项 `RunTime.EPPValidationMode`：`conf/ai_gateway_api.toml`、`test/integration/conf/ai_gateway_api.toml`、`stateful/config.go`。
2. 删除 `model/epp_pool` 的 ValidationMode 概念：`ValidationModeProduction`/`ValidationModeTest` 常量、`ManagerOptions.ValidationMode`、`EppPoolManager.validationMode`。
3. `/epp-pool` PATCH 校验统一为：**每组 1~2 个实例**（1=仅主，2=主+备；拒绝空组与 3 个及以上实例的组），不再按部署形态区分校验强度。
4. 分配器（allocator/repair）不再按组规模过滤候选组：池中所有组均为合法候选。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 配置 | `conf/ai_gateway_api.toml`（删 `EPPValidationMode` 及注释）、`test/integration/conf/ai_gateway_api.toml`（同上） |
| 代码 | `stateful/config.go`（删字段）、`stateful/container/rdb/components.go`（删装配）、`model/epp_pool/epp_pool.go`（删常量/选项/校验分支）、`model/epp_pool/assignment.go`（删 `groupMeetsValidationMode` 及相关参数） |
| 测试 | `model/epp_pool/epp_pool_manager_test.go`、`model/epp_pool/assignment_test.go`、`model/epp_pool/epp_data_test.go`、`endpoints/openapi_v1/epp_pool/endpoints_test.go`、`endpoints/openapi_v1/epp_assignments/endpoints_test.go`、`endpoints/innerapi_v1/epp_data/export_test.go`、`model/icluster_conf/epp_test.go` |
| 文档 | `api-define/OpenAPI接口定义/epp-pool.md`、`sys-design/模型层设计文档.md`、`sys-design/接口层设计文档.md`、`sys-design/数据库设计文档.md`、`sys-design/details/EPP调度对接.md` |
| 不涉及 | `epp_instances`/`epp_assignments` 表结构、API 请求/响应字段、epp_data 导出格式、BFE/EPP 侧 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 校验规则统一为每组 1~2 个实例 | 新规则：`len(group.instances)` 为 1 或 2 合法，空组与 ≥3 实例组拒绝，错误信息统一为 `"epp pool group %q requires 1 or 2 instances, got %d"`。相当于原 production 的"恰 2"放宽为"1 或 2"，同时原 test 的"≥1 无上限"收紧掉 ≥3 的场景。 |
| 主备语义保留但不强制 | 组内 1 个实例 = 仅主（primary-only 导出）；2 个实例 = 主+备（第 2 实例作为 standby 导出，`EPPAddr` = `[primary, standby]`、`epp-assignments` 视图、`epp_data` 导出格式均不变）。BFE 侧对单元素 `EPPAddr` 的支持保持不变。 |
| 分配器候选组 = 池中全部组 | 池中组必然为 1~2 实例（PATCH 校验拒绝空组与 ≥3 实例组），故 `candidateGroups` 不再做规模过滤，仅按组名排序保证确定性。`findValidAssignment`/`repairOne` 仅判断"组是否存在 + primary 是否仍在组内"。 |
| 行为变化点（对既有部署的影响） | ① 放宽：原 production 模式下 PATCH 单实例组返回 422 → 现在允许；② 行为修正：原 production 模式下"组缩容到 1 实例"会被视为规模不满足、触发**跨组重分配** → 现在 1 实例组是合法组，既有分配保留（仅当 primary 实例被移除时才同组重选 primary）；③ 收紧：原 test 模式下 PATCH 3 个及以上实例的组返回 200 → 现在返回 422。 |
| 配置兼容性 | 配置经 `lib.LoadConfAuto` → `BurntSushi/toml.Unmarshal` 加载，**非严格解码**：旧配置文件若残留 `EPPValidationMode` 键可正常启动（静默忽略），无迁移风险，建议部署时顺手清理。 |
| 历史变更记录处理 | `2026-09-08-epp-scheduling-integration` 记录中的 ValidationMode 相关描述保留为历史快照，不再回改；现行约束以本次更新后的 `api-define/` 与 `sys-design/` 为准。 |
