# Issue #205：cluster 审计快照裸 Go 字段名泄漏修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/205](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/205)（发现 TC：SC2101-TC046，断言 `assert-update-logs` 中 cluster 目标的 `DiffKeys` 断言，P1）

> 对既有 Cluster 发起部分更新（`PATCH /open-api/v1/clusters/{name}`，body 仅 `{"description"}`）成功时，operation-log 审计记录 `change_summary` 的 `before`/`after` 快照键名全部为**首字母大写的 Go 结构体字段名**（`Description`/`BalanceMode`/`LLMConfig`/`ID` 及全嵌套），`diff_keys=["Description"]`，而非 API 小写词汇（`description` 等）。失真不止大小写两层：(a) 值表示——`sticky_sessions.hash_strategy` 记内部枚举整数 `1`，API 返回字符串 `"CLIENT_IP_ONLY"`；(b) 内部字段泄漏——`ID`/`Ready`/`ProductID`/`Scheduler`（含 `GSLB_BLACKHOLE` 键）/`SubClusters[].ClusterID` 等为 `GET /clusters` 未暴露的内部记账字段。cluster 快照键名与 api_key/provider/entity（均为小写 API 词汇）全面脱节，消费方无法按 API 字段名解析。

关联定位：

- 属 issue #201（[修复方案](../2026-09-23-issue-201-operation-log-phantom-diff-keys/change-summary.md)）**明确预警的「同族排查 provider/cluster/entity 的 `*ParamToMap`」中的 cluster 侧变体**：#201 修的是 nil 物化为 null 的幻影 diff（`717378e`），并把 cluster 的 `clusterParamToMap`/`clusterToMap` 收敛为委托 `ioperlog.ParamToMap`——但 cluster 结构体**无 json tag**，marshal 往返天然产出大写 Go 字段名；#201 新增的守卫测试 `cluster_operation_log_test.go:31` 反而把 `{"Description": "d1"}` 固化成了断言；
- 与 #185（`error_msg` 回显裸 Key，P0 凭证泄漏）**不同根因不同严重级**：本卡是审计语义失真（P1），不涉及敏感信息回显；
- 产品条款依据：`operation-logs.md#1` 数据模型（`change_summary` 为变更前后存储值快照，**字段键名与资源 API 字段名一致**，示例均为小写 JSON 字段名）+ `clusters.md`（Cluster 字段词汇与部分更新语义）。

## 2. 根因（源码定向定位，行号为当前主干）

**键名层**：`model/icluster_conf/cluster.go` 的 `ClusterParam`（`:127-158`）与 `Cluster`（`:280-301`）及全部嵌套结构体（`ClusterBasicParam` 族 `:70-100`、`ClusterStickySessionsParam` `:102-106`、`ClusterPassiveHealthCheckParam` `:108-115`、`ClusterBasic` 族 `:160-190`、`ClusterStickySessions` `:192-196`、`ClusterPassiveHealthCheck` `:251-258`）**均无 json tag**。有 tag 的只有 `LLMConfig` 子树（`:198-249`，因 `llm_config` 是 clusters 表唯一 JSON 列而设 tag，`storage/rdb/cluster_conf/table_clusters.go:60`）与 `Instance`（`pool.go:76-82`，因 `pools.instance_detail` JSON 列而设 tag）——这就是 issue 证据中「同快照大小写混杂」的来源。

**构造层**：`model/icluster_conf/cluster_operation_log.go:52-58` `clusterParamToMap`/`clusterToMap` 为单行委托 `ioperlog.ParamToMap`（`model/ioperlog/param_to_map.go:27-48`，marshal 往返 + 丢 nil 键）。无 tag 字段经 `json.Marshal` 后键名 = Go 字段名原样大写。

**注入层**：7 个审计写入点全部直接消费这两个函数——create `cluster.go:581`/`:585`，update（校验失败 `:755`、事务失败 `:826`、成功 `:830`，`before = clusterToMap(oldData)`、`after = clusterParamToMap(param)`），delete `:1141`/`:1145`。已核对这两个函数**无任何其他调用方**（grep 全仓），修改 map 构造语义不影响任何 API 响应 / Inner 导出 / DB 序列化。

**值表示层**：`Cluster.StickySessions.HashStrategy int32`（`cluster.go:194`）/`ClusterStickySessionsParam.HashStrategy *int32`（`:104`）在快照中直接落整数；API 读路径的 int32→string 翻译在 endpoints 层（`endpoints/openapi_v1/product_cluster/one.go:147-151` 内联 map，词汇 `CLIENT_ID_ONLY`/`CLIENT_IP_ONLY`/`CLIENT_ID_PREFERED`，常量定于 `create.go:139-141`，写路径 S→I 为 `create.go:459-469`），模型层无法复用（不可反向依赖 endpoints）。

**字段范围层**：`Cluster.SubClusters []*SubCluster`（`sub_cluster.go:38-55`，无 tag，嵌套 `Pool`/`ibasic.Product` 均无 tag）与 `Cluster.Scheduler`（`cluster.go:290`）被一并 marshal 进 before 快照；`ID`/`Ready`/`ProductID` 同为 API 未暴露字段。产品条款（`clusters.md:270`/`:532`）：`sub_clusters` 与 `scheduler` 为系统内部自动生成，不对外暴露；`clusters.md:441`：返回数据不再含 `instance_pool`。

## 3. 目标

1. `PATCH {description}` 成功审计 `diff_keys == ["description"]`（小写 API 字段名），`after == {"description": "updated cluster"}`，`before` 为**与 `GET /clusters` 返回同形态**的小写 API 词汇快照；
2. 快照**全部键名**（顶层 + 全嵌套）统一为 API 小写词汇：`name`/`description`/`basic`/`sticky_sessions`/`passive_health_check`/`llm_config`/`balance_mode`/`epp_config`（词汇基准：`endpoints/openapi_v1/product_cluster/one.go:98-112` `ClusterData` 与 `clusters.md#1`）；
3. 快照**值表示**与 API 对齐：`sticky_sessions.hash_strategy` 落字符串枚举（`"CLIENT_IP_ONLY"` 而非 `1`）；`sticky_sessions.enabled`（内部字段名 `SessionSticky`）；`epp_config` 落解码后的 JSON 对象（与 API `json.RawMessage` 回显一致），而非转义字符串；
4. **内部字段按规范裁剪**：`id`/`ready`/`product_id`/`scheduler`/`sub_clusters`（整树，含 `ClusterID`/`InstancePool`/`Product` 嵌套泄漏）/`instance_pool` 不进快照；嵌套裁剪 `passive_health_check.schema`（API 未暴露）、`basic.retries.max_retry_cross_subcluster`（API 仅暴露 `max_retry_in_cluster`，见 `one.go:65-68`）、`basic.buffers.req_flush_interval`/`res_flush_interval`（API 仅暴露 `req_write_buffer_size`，见 `one.go:70-73`）；`balance_mode` 按存储值规范化（空 → `"WRR"`，同 `one.go:162-165` 与 `cluster.go:312-319` `getBalanceMode()`）；
5. 保留 #201 语义不回退：部分更新省略字段不物化（nil-guard）、显式置零仍计入 diff；`llm_config` 子树（本就小写）保持不动；api_key/provider/entity 路径不动；
6. 回归验证通过：单测 + model 覆盖率门禁（≥70%）+ SC2101-TC046 requeue real-verification（cluster 目标 `DiffKeys` 断言转绿、#201/#185 断言不回退）。

**目标快照形态**（`PATCH {description}` 成功，create 失败路径同词汇）：

```json
// after
{"description": "updated cluster"}
// before（与 GET /open-api/v1/clusters/{name} Data 同形态）
{
  "name": "my-cluster",
  "description": "old desc",
  "basic": {"protocol": "http", "connection": {"max_idle_conn_per_rs": 0, "cancel_on_client_close": false},
            "retries": {"max_retry_in_cluster": 2}, "buffers": {"req_write_buffer_size": 512},
            "timeouts": {"timeout_conn_serv": 50000, "...": "..."}},
  "sticky_sessions": {"enabled": false, "hash_strategy": "CLIENT_IP_ONLY", "hash_header": ""},
  "passive_health_check": {"interval": 1000, "failnum": 3, "host": "", "uri": "/", "statuscode": 0},
  "llm_config": {"models": ["..."], "provider": "deepseek"},
  "balance_mode": "WRR"
}
// diff_keys
["description"]
```

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `model/icluster_conf/cluster_operation_log.go`（重写两个 ToMap 为手写 API 词汇映射 + 嵌套 helper）、`model/icluster_conf/cluster_operation_log_test.go`（翻转 #201 守卫断言 + 新增词汇/值/裁剪断言）、`model/icluster_conf/cluster_test.go`（UpdateCluster 审计路径断言） |
| 明确不动 | `model/icluster_conf/cluster.go` 结构体定义与 json tag（见 §7 备选 1）、7 个 `recordClusterOperation` 调用点（签名不变、自动受益）、api_key/provider/entity 域、`ioperlog.ParamToMap`/`BuildChangeSummary`/`MaskSensitiveFields` |
| 集成测试 | `test/integration/tests/operation_log/list/` 新增 cluster PATCH 审计词汇回归用例（§5.4） |
| 文档同步 | `operation-logs.md` `change_summary` 行注记补「键名与 API 字段名一致、值表示对齐、内部字段裁剪」；`design-docs/sys-design/details/操作日志模块.md:162` 分组描述改写（cluster 移至手写映射组） |
| 接口契约 | 不变。Open API 请求/响应、错误码、DB schema 均不动；仅审计落库内容修复为条款语义 |
| 数据迁移 | 无（历史行可选清洗，见 §8） |

**行为变更声明**：cluster 的 create/update/delete 审计快照，修复后键名全面小写化、`hash_strategy` 变为字符串枚举、`epp_config` 变为 JSON 对象、内部字段与未暴露嵌套字段消失——**这是修复意图本身**。已核对本仓现存断言无依赖旧形态：集成用例（`test/integration/tests/operation_log/list/list_test.go:116`/`:284-297`）对 cluster 只断身份字段（action/resource_type/resource_name/status/error_msg），不断言快照键；唯一固化旧形态的是模型层守卫测试 `cluster_operation_log_test.go:31`（本方案翻转）。

## 5. 最终方案

**`clusterParamToMap`/`clusterToMap` 由「单行委托 `ParamToMap`」改为「手写 API 词汇映射」，收敛在 `cluster_operation_log.go` 单文件内**——与同仓 entity 先例（`model/entity/operation_log.go:89-137`）同模式：nil-guard 逐字段、`m["api_key"] = *ptr` 直赋、嵌套子结构手写子 map。修复点与缺陷产生点同层（审计 map 构造），不碰模型结构体 tag、不碰 DB、不碰导出。

选手写派而非补 tag 的决定性理由：本资源的**模型词汇与 API 词汇在多处不是大小写关系而是结构关系**——`SessionSticky`→`enabled`、`MaxRetryInSubcluster`→`max_retry_in_cluster`（且 `MaxRetryCrossSubcluster` API 不存在）、`ClusterPassiveHealthCheck.Schema` API 不存在、`balance_mode` 需空值规范化、`epp_config` 需字符串→JSON 对象解码、`SubClusters`/`Scheduler`/`ID` 等整树裁剪。这些用 json tag 表达不了或表达错位，用 tag 对齐等于把模型结构体的序列化面绑死在 API 词汇上（§7 备选 1）。

### 5.1 映射总表（模型字段 → 快照键）

**顶层**（Param 列为 nil-guard 条件，Cluster 列为取值来源）：

| 快照键 | `ClusterParam`（after） | `Cluster`（before/delete） | 备注 |
|--------|--------------------------|----------------------------|------|
| `name` | `Name *string` | `Name` | PATCH body 禁带 name（`update_basic.go:56-72`），update after 通常无此键 |
| `description` | `Description *string` | `Description` | 本卡断言路径 |
| `basic` | `Basic *ClusterBasicParam` → §5.2 | `Basic *ClusterBasic` → §5.2 | nil 则整键省略 |
| `sticky_sessions` | `StickySessions *ClusterStickySessionsParam` → §5.3 | `StickySessions *ClusterStickySessions` → §5.3 | 枚举翻译 |
| `passive_health_check` | `PassiveHealthCheck *ClusterPassiveHealthCheckParam` → §5.4 | `PassiveHealthCheck *ClusterPassiveHealthCheck` → §5.4 | 裁剪 `schema` |
| `llm_config` | `LLMConfig *LLMConfig` | `LLMConfig *LLMConfig` | 值经 `ioperlog.ParamToMap` 转换（§5.5） |
| `balance_mode` | `BalanceMode *string`（原样） | `getBalanceMode()`（空→`WRR`） | 与 API 读路径一致 |
| `epp_config` | `EppConfig *string` → 解码 | `EppConfig string` → 解码 | 空串则整键省略（§5.6） |
| —（裁剪） | `ID`/`ProductID`/`SubClusters`/`Scheduler`/`InstancePool` | `ID`/`Ready`/`ProductID`/`SubClusters`/`Scheduler` | 内部记账字段，不进快照 |

### 5.2 `basic` 子映射（两路共用键集，保证 DeepEqual 可比）

- `connection`：`max_idle_conn_per_rs` / `cancel_on_client_close`；
- `retries`：**仅** `max_retry_in_cluster`（Param 取 `MaxRetryInSubcluster`；Cluster 取 `MaxRetryInSubcluster`；`MaxRetryCrossSubcluster` 裁剪）；
- `buffers`：**仅** `req_write_buffer_size`（`ReqFlushInterval`/`ResFlushInterval` 裁剪）；
- `timeouts`：`timeout_conn_serv` / `timeout_response_header` / `timeout_readbody_client` / `timeout_read_client_again` / `timeout_write_client`（内部与 API 同名 snake，逐一映射）；
- `protocol`：`Protocol *string` / `*string`（nil 省略）。

### 5.3 `sticky_sessions` 子映射与枚举翻译

- 键：`enabled`（Param `SessionSticky *bool` / Cluster `SessionSticky bool`）、`hash_strategy`、`hash_header`；
- 枚举翻译：在 `cluster_operation_log.go` 内定义 `var clusterHashStrategyI2S = map[int32]string{ClusterHashStrategyClientIDOnlyI: "CLIENT_ID_ONLY", ClusterHashStrategyClientIPOnlyI: "CLIENT_IP_ONLY", ClusterHashStrategyClientIDPreferedI: "CLIENT_ID_PREFERED"}`（词汇与 `create.go:139-141` 逐字一致）；未识别值的兜底与 API 读路径同构（map 查缺得 `""`，实际不可达：写路径 `hashStrategyConvert` 已校验枚举，`cluster.go:1185-1200` `normalizeBFEHashStrategy` 兜底）；
- 依赖方向：模型层不得 import endpoints（会成环），故翻译表落在 icluster_conf 内；与 endpoints 两处既有表（`one.go:147-151`、`create.go:459-469`）的词汇去重为可选后续优化，本变更不做（保持外科手术式改动面）。

### 5.4 `passive_health_check` 子映射

键：`interval` / `failnum` / `statuscode` / `host` / `uri`（API `PassiveHealthCheck` 即此五键，`one.go:89-95`）；**裁剪 `Schema`**（内部探活协议记账字段，API 未暴露）。

### 5.5 `llm_config` 子树：复用 `ParamToMap`，不手写

`LLMConfig` 及其子结构自带正确小写 tag（`cluster.go:224-249`），直接 `m["llm_config"] = ioperlog.ParamToMap(x)`（nil 入参返回 nil → 整键省略）。收益有二：

1. 子树内继续享受 #201 的 nil-drop 语义（部分更新的 llm_config 省略子字段不物化为 null）；
2. **维持脱敏递归不变式**：`MaskSensitiveFields`（`model/ioperlog/mask.go:41-73`）只递归 `map[string]interface{}` / `[]interface{}`；若快照 map 内嵌 Go struct 或 `json.RawMessage`，脱敏会静默跳过该子树。手写映射的全部值必须保持纯 JSON 类型（手搭 `map[string]interface{}`、经 `ParamToMap` 转换的 llm_config、解码后的 epp_config 对象、基础类型），这是本方案的硬约束。

### 5.6 `epp_config` 值表示

存储态为 raw JSON 字符串（`Cluster.EppConfig string` / `ClusterParam.EppConfig *string`，后者经 `validateClusterBalanceConfig` 校验改写）。快照值解码为 JSON 对象与 API `json.RawMessage` 回显对齐：

```go
func eppConfigSnapshotValue(raw string) interface{} {
    if raw == "" {
        return nil // 未设置 → 整键省略
    }
    var v interface{}
    if err := json.Unmarshal([]byte(raw), &v); err != nil {
        return raw // 防御：解码失败退回原始字符串（写路径已校验，实际不可达）
    }
    return v
}
```

### 5.7 实现骨架（`model/icluster_conf/cluster_operation_log.go`）

```go
func clusterParamToMap(param *ClusterParam) map[string]interface{} {
    if param == nil {
        return nil
    }
    m := map[string]interface{}{}
    if param.Name != nil { m["name"] = *param.Name }
    if param.Description != nil { m["description"] = *param.Description }
    if v := clusterBasicParamToSnapshot(param.Basic); v != nil { m["basic"] = v }
    if v := clusterStickySessionsParamToSnapshot(param.StickySessions); v != nil { m["sticky_sessions"] = v }
    if v := clusterPHCParamToSnapshot(param.PassiveHealthCheck); v != nil { m["passive_health_check"] = v }
    if v := ioperlog.ParamToMap(param.LLMConfig); v != nil { m["llm_config"] = v }
    if param.BalanceMode != nil { m["balance_mode"] = *param.BalanceMode }
    if v := eppConfigSnapshotValue(deref(param.EppConfig)); v != nil { m["epp_config"] = v }
    // ID / ProductID / SubClusters / Scheduler / InstancePool 为内部记账字段，按规范裁剪
    return m
}

func clusterToMap(cluster *Cluster) map[string]interface{} {
    if cluster == nil {
        return nil
    }
    m := map[string]interface{}{
        "name": cluster.Name, "description": cluster.Description,
        "balance_mode": cluster.getBalanceMode(),
    }
    // basic / sticky_sessions / passive_health_check / llm_config / epp_config 同 §5.2-5.6 词汇
    // ID / Ready / ProductID / SubClusters / Scheduler 为内部记账字段，按规范裁剪
    return m
}
```

（嵌套 helper 各约 10-30 行；`recordClusterOperation` 与 7 个调用点零改动。）

### 5.8 单元测试

1. **翻转 #201 守卫断言**：`cluster_operation_log_test.go:31` `{"Description": "d1"}` → `{"description": "d1"}`——该断言是缺陷固化，必须连同注释一并修正；
2. **词汇断言（Param 路）**：满配 `ClusterParam` → 键集合精确等于 `{"name","description","basic","sticky_sessions","passive_health_check","llm_config","balance_mode","epp_config"}`；`basic` 嵌套键集合精确等于 API 形态（`retries` 仅 `max_retry_in_cluster`、`buffers` 仅 `req_write_buffer_size`）；
3. **词汇断言（Cluster 路）**：满配 `Cluster`（含 `SubClusters`/`Scheduler`/`Ready`/`ID` 填充）→ 快照无 `id`/`ready`/`product_id`/`scheduler`/`sub_clusters` 键；`passive_health_check` 无 `schema`；`balance_mode` 空存储值输出 `"WRR"`；
4. **值表示断言**：`hash_strategy` 输出 `"CLIENT_IP_ONLY"` 字符串（两路、三个枚举值）；`epp_config` 输出解码对象而非转义字符串；未设置时整键省略；
5. **#201 语义不回退**：仅 `Description` 非 nil → map 仅 `{"description"}`；nil 入参 → nil；
6. **UpdateCluster 审计路径**（复用 `cluster_test.go:1433` `fakeOperationLogRecorder` 模式）：PATCH 场景（仅 Description）成功 → 捕获 entry `diff_keys == ["description"]`、`after` 无大写键、`before` 为 §3 目标形态；事务失败路径（`:826`）同规则。

### 5.9 集成测试（`test/integration/tests/operation_log/list/`）

新增用例（仿 #201 的 api_key 幻影 diff 用例风格；`testutil.CreateCluster`/`client.Patch`/`WaitForOperationLog` 均已在位）：

1. 创建 run-scoped cluster（`testutil.CreateCluster`）→ `PATCH /open-api/v1/clusters/{name}` 仅 `{"description": "updated cluster"}` → `WaitForOperationLog(resource_type=cluster, action=update, resource_name={name}, status=1)` →
   - `assert.ElementsMatch(diffKeys, ["description"])`——**精确匹配，禁止 Contains**（#201 用例的盲区教训）；
   - `change_summary.after` 键集合精确等于 `{"description"}`（无大写键、无 `id`/`ready`/`scheduler`/`sub_clusters`）；
   - `change_summary.before` 键全为小写 API 词汇（遍历断言无大写起头键），且含 `balance_mode`；
2. **枚举值表示**：再 `PATCH {"sticky_sessions": {"enabled": true, "hash_strategy": "CLIENT_ID_ONLY", "hash_header": "x-uid"}}` → `diff_keys` 精确 `["sticky_sessions"]`，`after.sticky_sessions.hash_strategy == "CLIENT_ID_ONLY"`（字符串，非数值）；
3. 说明：SC2101-TC046 外部 E2E（integration-test 仓）是 Oracle 断言归属地，本仓用例是 CI 常驻快速回归锚，二者互补。

### 5.10 文档同步

- `operation-logs.md:48` `change_summary` 行注记扩展为：「……快照**键名与资源 API 字段名一致**（小写 JSON 词汇），值表示与 API 对齐（枚举以字符串表示），API 未暴露的内部记账字段不入快照；部分更新省略的字段不出现在 `after`/`diff_keys`……」——把本方案的三层修复意图提升为文档条款；
- `design-docs/sys-design/details/操作日志模块.md:162` 分组描述改写：cluster 从「marshal 模式（`ParamToMap`）」组移至「手写 nil-guard / API 词汇映射」组（与 entity 同组），并注明 cluster 手写映射承载三项审计语义：API 词汇对齐、枚举值翻译、内部字段裁剪；`:176` 处文件标注不变（构造仍在 `cluster_operation_log.go`）。

## 6. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`make test-model-cover-gate`（≥70%）全部通过；
2. 单元级：§5.8 全部新用例通过；#201 守卫测试翻转后通过；api_key/provider/entity 的审计单测（`api_key_operation_log_test.go` 等）不回退（这些路径零改动）；`ioperlog` 自测不回退；
3. 集成级：§5.9 新用例通过；既有 `operation_log/list` 用例不回退（cluster 相关用例只断身份字段，已核对）；
4. 部署门禁：SC2101-TC046 requeue real-verification——
   - cluster 目标 `assert-update-logs`：`diff_keys=[description]` 转绿（本卡）；
   - api_key 目标 `assert-update-logs`（#201）、`sc2101AuditAssertSecretAbsent`（#185）保持绿；
5. 冒烟对照：`PATCH /clusters/{name}` 后 `GET /open-api/v1/clusters/{name}` 响应与审计 `before` 快照（下一轮的 before）字段词汇逐项一致；
6. 上线后抽样：`GET /open-api/v1/operation-logs?resource_type=cluster` 近 N 日记录，抽查快照键名无大写、无内部字段。

## 7. 备选方案（不采纳）

| 备选 | 不采纳理由 |
|------|-----------|
| 给 `ClusterParam`/`Cluster` 及嵌套族补 json tag + 内部字段 `json:"-"` | ①解决不了值表示层：`hash_strategy` 仍是 int32、`epp_config` 仍是转义字符串，仍需自定义 `MarshalJSON` 或后处理，复杂度不低于手写；②模型词汇 ≠ API 词汇非大小写关系：`SessionSticky`→`enabled`、`MaxRetryInSubcluster`→`max_retry_in_cluster`、`Schema`/flush 间隔/`MaxRetryCrossSubcluster` 需裁剪——tag 会把模型结构体的序列化面（被 DB 映射、export、GSLB 数据等多处复用）绑死在 API 词汇上；③`SubClusters` 整树裁剪用 `json:"-"` 虽可行，但 tag 方案总改动面（17+ 结构体 × 多消费方风险）远大于单文件手写映射；④tag 服务的是序列化契约，审计快照服务的是 API 词汇契约，两者在本资源上天然分叉 |
| 给 `ClusterStickySessions(Param)` 加自定义 `MarshalJSON` 做枚举翻译 | 只解决枚举一层；`MarshalJSON` 是**全局序列化语义**变更（所有 marshal 消费者继承），为审计单消费者改动过大；键名/裁剪问题仍需 tag 或后处理 |
| `ParamToMap` 输出后做键名改写后处理（大写→小写映射表） | 脆弱：多处映射非纯大小写变换（`MaxRetryInSubcluster`→`max_retry_in_cluster`）；嵌套 `SubClusters`/`InstancePool`/`Product` 整树需逐层特判裁剪；键名改写远离字段定义，新增字段必漏——正是 #201 总结的「缺陷产生点远离修复点易再犯」反模式 |
| 只修 `description` 一个键 | issue 明确要求「全部键名（顶层+全嵌套）统一、值表示对齐、内部字段裁剪，而非仅修 description」；且 `diff_keys` 词跟随 after 键，修单键不解决消费方解析问题 |
| create/delete 路径维持旧形态只修 update | 同两个 ToMap 函数、同 7 个调用点，分裂形态只会制造第二处不一致；create/delete 快照同样是条款语义下的审计产物 |

## 8. 历史数据处置（建议，可选）

代码修复只保证新增审计行语义正确；存量 `operation_logs` 中 `resource_type=cluster` 的 create/update/delete 行快照仍为大写键、枚举整数、含内部字段。本卡为 P1 语义失真（非凭证泄漏），处置从宽：

- 由运维/审计消费方定夺：① 一次性清洗（按 `resource_type=cluster` 定位，键名表改写 + 枚举整数→字符串 + 剔除内部字段，成本高于 #201 的幻影键清洗，需逐资源评估）；② 消费方侧仅信任修复后新增记录，历史记录按不可解析处理；③ 留存历史仅修复新增——若选此项，需在 issue #205 中显式记录该决策；
- 执行清洗前备份该表；清洗动作本身不再产生新审计噪音（审计只记录 Open API 写操作）。
