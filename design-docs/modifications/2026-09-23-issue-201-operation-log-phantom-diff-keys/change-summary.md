# Issue #201：操作日志 diff_keys 幻影字段（apiKeyParamToMap nil 物化）修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/201](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/201)（发现 TC：SC2101-TC046，断言 `assert-update-logs`，P1）

> 对已存在的 API Key 发起部分更新（`PATCH /open-api/v1/api-keys/{id}`，仅提交 `description`）成功时，审计记录 `change_summary.diff_keys` 实际为 `[description enabled id key]`，预期为 `[description]`——请求中**未提交**的 `id`/`enabled`/`key` 被当作变更记入 diff_keys（幻影 diff）。这使 `diff_keys` 失去「仅含受控变更字段」的语义，干扰基于 diff_keys 的审计/合规对账。

关联定位：

- 与 #147/#151（省略覆盖 storager 家族，写库 nil-skip 已修）**同族**：本卡正是该族「审计快照未修」一侧的向量；
- 与 #185（失败审计 `error_msg` 回显裸 Key，已修复）**同 TC 不同根因**：#185 是凭证泄漏（P0），本卡是审计语义失真（P1），不断言 `error_msg`，#185 的掩码修复不覆盖本路径；
- 产品条款依据：`model/ioperlog/change_summary.go:30-32` 设计注释（「partial updates omit unchanged fields, including them produces false-positive diffs」）+ `operation-logs.md` 数据模型中 `change_summary`/`diff_keys` 语义。

## 2. 根因（源码定向定位，行号为当前主干）

**nil 物化**：`model/api_key/api_key_operation_log.go:72-88` `apiKeyParamToMap` 走 `json.Marshal(param)` → `json.Unmarshal` 成 map。`APIKeyParam` 的 `ID *string json:"id"`（`api_key.go:34`）、`Enable *bool json:"enabled"`（`:35`）、`Key *string json:"key"`（`:42`）三个字段**无 `omitempty`**——nil 指针序列化为 `"null"`，unmarshal 后这些键的值是 nil 但**键始终存在**。

**幻影注入**：`UpdateAPIKey` 成功路径（`api_key.go:635`）`after = apiKeyParamToMap(param)`——param 是请求体，仅 `description` 非 nil，其余 nil 指针全部物化为 null；`before = apiKeyParamToMap(oldAPIKey)`（`:625` 失败路径同构）——oldAPIKey 来自 DB，`id`/`enabled`/`key` 均有值。

**diff 误记**：`model/ioperlog/change_summary.go:57-76` `computeDiffKeys` 遍历 **after** 的键，`before[key]` 不存在或 `!reflect.DeepEqual` 即记入。`after` 中 `id/enabled/key=null` vs `before` 中有值 → `null != 有值` → 三个幻影键入 diff_keys（排序后即 `[description enabled id key]`）。

两点补充：

- **before 侧无此问题**：`computeDiffKeys` 只遍历 after 的键，before 多出键被设计性忽略（`change_summary.go:30-32`）；缺陷单向来自 after 构造；
- **「省略」与「显式置零」在本 Param 层可区分**：`enabled=false` 是非 nil 指针（`json:"enabled"` 输出 `false`），与 nil 指针（省略/`null`）在 map 构造层可分辨——这使「丢弃 nil 值键」的修复语义精确：省略不进 diff，显式置零/置 false 仍进 diff。

## 3. 目标

1. `PATCH {description}` 成功审计 `diff_keys == ["description"]`，`after` 不含未提交字段（`id`/`enabled`/`key`）；
2. 保留「显式置零」语义：`PATCH {"description": "x", "enabled": false}` → `diff_keys == ["description","enabled"]`；
3. 失败审计路径（`api_key.go:625`）与 PUT 全量更新同规则受益，无需特例；
4. 同族收敛：全 model 域排查显示仅 provider/cluster 与 api_key 同型（marshal 往返），一并在本变更内修复；
5. 回归验证通过：单测 + model 覆盖率门禁（≥70%）+ SC2101-TC046 requeue real-verification（`assert-update-logs` 转绿、#185 断言不回退）。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `model/ioperlog/change_summary.go`（新增共享助手 `ParamToMap`）、`model/api_key/api_key_operation_log.go`、`model/iprovider/provider_operation_log.go`、`model/icluster_conf/cluster_operation_log.go`、对应 `_test.go`、`test/integration/tests/operation_log/list/`（#201 回归集成用例，见 §5.4） |
| 文档同步 | `operation-logs.md` `change_summary` 行补「省略字段不进 after/diff_keys」注记；`design-docs/sys-design/details/操作日志模块.md` 补构造规则 |
| 接口契约 | 不变。审计落库内容（`after`/`diff_keys`）修复为条款语义；Open API 请求/响应、错误码均不动 |
| 数据迁移 | 无（历史行可选清洗，见 §8） |

**全模型域排查结论（`*ParamToMap` 逐一核对，本卡新增证据）**：

| 构造模式 | 域 / 函数 | 是否有幻影缺陷 |
|----------|-----------|----------------|
| `json.Marshal`→`Unmarshal` 往返 | `api_key.apiKeyParamToMap`（`api_key_operation_log.go:72`）；`icluster_conf.clusterParamToMap`/`clusterToMap`（`cluster_operation_log.go:53`/`:71`）；`iprovider.providerParamToMap`/`providerToMap`（`provider_operation_log.go:52`/`:70`） | **有**（本卡修复 + 同族迁移） |
| 手写 nil-guard（`if ptr != nil { m[k] = *ptr }`） | entity（`operation_log.go:89`/`:108`）、route_rules 家族（`operation_log.go:54-140`）、quota（`operation_log.go:105`）、iauth（`operation_log.go:103`）、iprotocol（`operation_log.go:52`）、iroute_conf（`operation_log.go:114`）、rate_limit_policy（`operation_log.go:127`） | 无（本来就对，不改） |

**调用点爆炸半径已核对**：全部 13 个 `*ParamToMap` 函数的调用点均为各域 `record*Operation` 审计写入路径（如 `api_key.go:625/635`、`provider.go:167-372`），无任何 API 响应 / Inner 导出 / 配置序列化复用——修改 map 构造语义不影响接口契约。

**行为变更声明**：provider/cluster 的部分 PATCH 成功审计，修复后 `after`/`diff_keys` 将同样不再含未提交字段（幻影键消失）。这是修复意图本身；现存单测/integration 断言中无对幻影键的依赖（`diff_keys` 断言仅见于 entity 域与 ioperlog 自测，均为手写 map 输入，不受影响）。

## 5. 最终方案

**在审计管线归属地（`model/ioperlog`）新增共享助手「marshal 往返 + 丢弃 nil 值键」，4 个 marshal 模式函数改为委托。** 修复点与缺陷产生点同层（map 构造），不碰 `APIKeyParam` json tag、不做 DB 重查、无时间戳副作用。

### 5.1 共享助手（`model/ioperlog/change_summary.go` 或新文件 `param_to_map.go`）

```go
// ParamToMap marshals a param struct into a map for change_summary
// before/after snapshots. Keys whose value is nil are dropped, so that
// fields omitted from a partial update do not materialize as null entries:
// partial updates omit unchanged fields, and including them produces
// false-positive diffs (see BuildChangeSummary). Explicit zero values
// (e.g. enabled=false) are non-nil and are therefore retained.
func ParamToMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	for k, val := range m {
		if val == nil {
			delete(m, k)
		}
	}
	return m
}
```

- **放 `ioperlog` 的理由**：「省略不物化」是审计摘要的构造约定，与 `BuildChangeSummary`/`MaskSensitiveFields`/`TruncateErrorMessageDefault` 同属审计管线咽喉语义；全部域的 `operation_log.go` 已依赖 `model/ioperlog`，无新依赖方向；
- **nil 值键 vs 显式零值**：`nil`（省略/`null`）丢弃；`false`/`0`/`""`（显式提交）保留——精确区分「未提交」与「置零」，后者仍按契约记入 diff_keys；
- 与 `MaskSensitiveFields` 无冲突：掩码在 map 构造之后按键名清单执行，丢键只减少掩码输入。

### 5.2 迁移 4 个 marshal 模式函数（每个改为单行委托）

| 文件 | 函数 | 改法 |
|------|------|------|
| `model/api_key/api_key_operation_log.go:72` | `apiKeyParamToMap` | `return ioperlog.ParamToMap(param)`（**本卡断言路径，必做**） |
| `model/icluster_conf/cluster_operation_log.go:53` | `clusterParamToMap`、`clusterToMap` | 同上（同族；`clusterToMap` 为 before/删除快照，同型 nil 物化一并收敛） |
| `model/iprovider/provider_operation_log.go:52` | `providerParamToMap` | 同上（同族） |
| `model/iprovider/provider_operation_log.go:70` | `providerToMap` | 同上（同型，`Provider` 结构同样含无 omitempty 指针字段） |

手写 nil-guard 的 7 个域**不改**（语义已正确）；若后续为统一风格迁移到 `ParamToMap`，属纯重构，另立变更。

### 5.3 单元测试

1. `model/ioperlog` 新增 `param_to_map_test.go`：
   - nil 入参 → nil；
   - 混合 struct（nil 指针 + 非 nil 指针 + 显式 `false`/`0`/`""`）→ nil 键被丢弃、显式零值保留；
   - 不可序列化字段（如 `chan`）→ marshal 失败返回 nil；
2. `model/api_key/api_key_operation_log_test.go` 新增（仿既有 fake auditor 捕获 `OperationLogEntry` 模式，参考 `:43`/`:104`）：
   - `apiKeyParamToMap` 单测：仅 `description` 非 nil → map 仅含 `description`；
   - `UpdateAPIKey` 成功路径 PATCH `{description}` → 捕获 entry `ChangeSummary["diff_keys"] == []string{"description"}`，且 `after` map 无 `id`/`enabled`/`key` 键；
   - 显式 `enabled=false` + `description` → `diff_keys == ["description","enabled"]`（置零语义不回退）；
   - 失败路径（`:625`）同规则（after 同样经 `ParamToMap`）。
3. `model/iprovider/operation_log_test.go`、`model/icluster_conf` 各补一个 delegate 断言：含 nil 指针字段的 param → 落库 map 无对应键。

### 5.4 集成测试（`test/integration`）

**必做，作为本卡回归锚点**。现有集成覆盖存在同类盲区：全仓唯一对 update 审计 `diff_keys` 的集成断言是 `operation_log/list/list_test.go:344`（entity 域），且只用 `assert.Contains(diffKeys, "description")`——Contains 断言对幻影键不敏感（多出的键不导致失败），`api_key/partial_update` 集成用例只验资源行为、不断言审计。若不加精确断言的回归用例，本缺陷修复后无本仓 CI 级护栏。

新增用例（`test/integration/tests/operation_log/list/` 包内，与既有 diff_keys 用例同风格；testutil 辅助 `CreateAPIKey`/`WaitForOperationLog`/`OperationLogEntry.ChangeSummary` 均已就位）：

1. **幻影 diff 回归（核心）**：创建 entity + api key → `PATCH /open-api/v1/api-keys/{id}` 仅 `{description}` → `WaitForOperationLog(resource_type=api_key, action=update, resource_id={id}, status=1)` →
   - `assert.ElementsMatch(diffKeys, ["description"])`——**精确匹配，禁止用 Contains**；
   - `change_summary.after` 断言**不含** `id`/`enabled`/`key` 键（比 diff_keys 多验一层落库快照语义）；
2. **显式置零语义**：同 key 再 `PATCH {enabled:false}` → `diff_keys` 精确等于 `["enabled"]`（置零是真实变更，不回退为省略）；
3. **加固既有用例**：`TestOperationLog_UpdateEntityTypeHasDiffKeys` 的 `Contains` 升级为精确匹配，消除同类盲区；
4. provider/cluster 的迁移由 §5.3 模型层单测覆盖，集成层不再加（避免慢测试膨胀）。

说明：SC2101-TC046 外部 E2E（integration-test 仓）是 Oracle 断言的归属地；本仓集成用例是随 CI 常驻的快速回归锚，二者互补不互替。

### 5.5 文档同步

- `design-docs/api-define/OpenAPI接口定义/operation-logs.md:48` `change_summary` 行注记扩展为：「包含 `before` / `after` / `diff_keys`，敏感字段已脱敏；部分更新省略的字段不出现在 `after` / `diff_keys`（partial updates omit unchanged fields）」——把 `change_summary.go:30-32` 的代码注释意图提升为文档条款，防回归；
- `design-docs/sys-design/details/操作日志模块.md` 变更摘要构造节补一行：`before`/`after` 由各域 `*ParamToMap` 构造，marshal 模式统一经 `ioperlog.ParamToMap`，nil 值键（省略字段）不物化。

## 6. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`make test-model-cover-gate`（≥70%）全部通过；
2. 单元级：§5.3 全部新用例通过；既有 `change_summary_test.go`（手写 map 输入）、entity `operation_log_test.go:102/165` 的 diff_keys 断言不回退；#185 的掩码用例（`api_key_operation_log_test.go:43/104`）不回退（掩码发生在丢键之后，无交互）；
3. 集成级：§5.4 新用例通过（`PATCH {description}` → `diff_keys` 精确 `[description]`、`after` 无幻影键；`PATCH {enabled:false}` → `diff_keys` 精确 `[enabled]`）；既有 `operation_log/list` 用例（含加固后的 entity diff_keys 精确断言）不回退；
4. 部署门禁：SC2101-TC046 requeue real-verification——
   - `assert-update-logs`：`diff_keys=[description]` 转绿；
   - `assert-failed-writes` / `sc2101AuditAssertSecretAbsent`（#185）保持绿；
   - `mutation_readback` 不回退；
5. provider/cluster 相关集成用例（SC1601 系列等）回归：部分 PATCH 的 `diff_keys` 幻影键消失属预期修复效果，若有用例断言旧幻影形态则按新语义更新；
6. 上线后抽样：`GET /open-api/v1/operation-logs?resource_type=api_key&action=update&status=1` 近 N 日记录，抽查 `diff_keys` 与提交字段集一致。

## 7. 备选方案（不采纳）

| 备选 | 不采纳理由 |
|------|-----------|
| 给 `APIKeyParam` 的 `id`/`enabled`/`key` 补 `omitempty` | `APIKeyParam` 被 Open API 响应直接复用：`endpoints/openapi_v1/api_key/one.go:62` `newResponse(ctx, []*api_key.APIKeyParam{one})` 序列化为 GET 响应体——补 tag 会把响应中的 `"id": null` 等字段整个省略，**改变对外接口契约**；issue #201 也明确建议不动 tag。修复应收敛在审计 map 构造层 |
| 在 `computeDiffKeys` 跳过 nil after 值 | 治标：落库 `after` 仍含幻影 null，违反「after 仅解释请求意图」的条款语义；过滤逻辑远离缺陷产生点，新增域/新增字段时易再犯；且把「构造层该做的省略物化抑制」推给了 diff 算法 |
| success 路径 `after` 改 DB 重读（issue A1） | `api_keys.updated_at ... ON UPDATE CURRENT_TIMESTAMP`（`db_ddl.sql`）引入 `update_time` 伪 diff，需额外过滤 + 一次 DB 重查 + 掩码路径差异；不如构造层修复外科 |
| 只修 `apiKeyParamToMap`，provider/cluster 不动 | 同型代码留雷（`ProviderParam.Name *string json:"name"` 等同样无 omitempty，`provider.go:102`）；迁移成本是每函数一行委托，调用点已全部核对为审计路径 |
| 各域继续手写 nil-guard | 无共享约定，新增域极易再犯——`apiKeyParamToMap` 正是手写 marshal 的产物；`ParamToMap` 把「省略不物化」收敛为单点咽喉语义，与 #162/#185「脱敏收敛在咽喉点」的既定架构原则一致 |

## 8. 历史数据处置（建议，可选）

代码修复只保证新增审计行语义正确；存量 `operation_logs` 中 api_key/provider/cluster 的成功 update 行 `diff_keys` 仍含幻影字段、`after` 仍含 null 值。本卡为 P1 语义失真（非凭证泄漏），处置从宽：

- 由运维/审计消费方定夺：① 一次性清洗（按 `resource_type`+`action=update`+`status=1` 定位，比对 `before` 重算 `diff_keys` 并剔除幻影键，`after` 中 null 值键删除）；② 消费方侧按 `before`/`after` 重算真实变更面，忽略历史 `diff_keys`；③ 留存历史仅修复新增——若选此项，需在 issue #201 中显式记录该决策；
- 执行清洗前备份该表；清洗动作本身不再产生新审计噪音（审计只记录 Open API 写操作）。
