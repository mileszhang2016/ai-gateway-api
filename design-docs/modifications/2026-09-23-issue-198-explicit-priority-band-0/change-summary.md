# 变更摘要：CompileEppConfig 显式下发 priority band 0（Issue #198）

## 背景

GitHub Issue [rainway-ai-gateway/ai-gateway-api#198](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/198)（open，2026-09-22，严重级别 High）：

`CompileEppConfig` 编译 `flow_control` 段时**不产出任何 `priorityBands`**。ai-gateway-epp 没有 InferenceObjective reconciler（全仓无任何 `InferenceObjective` 引用，`pkg/` 下无任何 `Priority` 逻辑），每个请求的 priority **恒为 0**，因此 band 0 是唯一会被使用的优先级带。band 0 未显式配置时落到 llm-d 的**隐藏默认** `maxRequests=5000 / maxBytes=1GB`，造成两个容量语义错误：

1. **静默截断**：全局 `flow_control.max_requests > 5000` 时被 band 默认 5000 截断，用户配置不生效且无感知；
2. **字节上限先于请求上限触发**：长 prompt / 大 body 场景下 band 的 1GB 字节容量可能先于 `max_requests` 成为瓶颈。

## 根因定位

- llm-d 侧 band 默认值：`llm-d-router/pkg/epp/flowcontrol/registry/config.go:41-49` —— `defaultPriorityBandMaxRequests = 5000`、`defaultPriorityBandMaxBytes = 1_000_000_000`；未声明的优先级经 `DefaultPriorityBand` 模板动态建带（`config.go:87-89`，nil 时由 `NewConfig` 填系统默认）。
- 容量检查语义：`llm-d-router/pkg/epp/flowcontrol/controller/internal/processor.go:370-389` `hasCapacity()` —— **全局与 band 限额各自独立判定、全部通过才放行**，等效容量 = min(全局, band)。（注：issue 原文引用路径 `controller/internal/processor.go:377-397` 为简写，实际完整路径如上。）
- apiix 契约：`llm-d-router/apix/config/v1alpha1/endpointpickerconfig_types.go:476,551-580` —— `priorityBands []PriorityBandConfig`；`PriorityBandConfig{Priority int (json:"priority"，无 omitempty), MaxBytes *resource.Quantity, MaxRequests *resource.Quantity, ...}`，Quantity 规范 JSON 形式为**字符串**（`"2000"`、`"4Gi"`）；`"0"` 视为未设置回落系统默认，**带级上限恒存在**，"不限"只能显式设大值。
- 本侧现状：`model/epp_pool/compiler.go:305-325` `compileFlowControl()` 只展开 `maxRequests/defaultRequestTTL/noEndpointRequestTTL/enableEviction`，产物无 `priorityBands`。
- 设计债务：`design-docs/modifications/2026-09-08-epp-scheduling-integration/api-changes.md:222` 明确"各优先级带另有 band 级限额，**本期不暴露、用系统默认**"——本 issue 即补上了这个暂缓项。因只有 band 0 在用，**band 0 的配置就等于整个流控容量与策略的定义**，必须显式下发、可审计。

## 目标

编译产物**始终显式下发 priority 0 的 band**：`maxRequests` 与全局配置联动（修静默截断）、`maxBytes` 给出显式默认值，使容量语义确定、可审计；用户侧 OpenAPI（`epp_config.flow_control`）不变、DB 无变更。

## 修复方案

### 1. 代码修复（`model/epp_pool/compiler.go`）

**新增 band 类型**（ mirroring apix `PriorityBandConfig`，只下发需要的子集）：

```go
// PriorityBandConfig mirrors apix PriorityBandConfig (only the subset we emit).
// MaxRequests/MaxBytes are strings because apix types them as resource.Quantity,
// whose canonical JSON form is a string ("2000", "4Gi").
type PriorityBandConfig struct {
	Priority    int    `json:"priority"`              // 必填（apix 无 omitempty）
	MaxRequests string `json:"maxRequests,omitempty"` // resource.Quantity 字符串
	MaxBytes    string `json:"maxBytes,omitempty"`    // resource.Quantity 字符串
}
```

**`FlowControlConfig` 增加字段**：

```go
PriorityBands []PriorityBandConfig `json:"priorityBands,omitempty"`
```

**`compileFlowControl` 始终产出 priority 0 band**：

```go
// 全局 max_requests 不限（-1 / 未设置）时 band 内使用的默认最大值；
// 取 10000（> llm-d 隐藏默认 5000，容量语义从"隐藏的 5000"变为"显式的 10000"）。
const defaultPriorityBandMaxRequests = "10000"
// band 字节上限默认值（> llm-d 隐藏默认 1GB，覆盖长 prompt / 多模态场景）。
const defaultPriorityBandMaxBytes = "5Gi"

func compileFlowControl(fc *FlowControlSimplified) *FlowControlConfig {
	compiled := &FlowControlConfig{}
	band := PriorityBandConfig{Priority: 0, MaxBytes: defaultPriorityBandMaxBytes}

	if fc.MaxRequests != nil && *fc.MaxRequests > 0 {
		q := strconv.Itoa(*fc.MaxRequests)
		compiled.MaxRequests = q
		band.MaxRequests = q // band0 == global，修"全局 > 5000 被静默截断"
	} else {
		// max_requests == -1（FlowControlUnlimited）或未设置：全局不限，
		// 但带级上限恒存在（apix 语义），band 内给显式默认值。
		band.MaxRequests = defaultPriorityBandMaxRequests
	}

	if fc.QueueTTL != nil {
		compiled.DefaultRequestTTL = secondsToDuration(*fc.QueueTTL)
	}
	if fc.NoEndpointQueueTTL != nil {
		compiled.NoEndpointRequestTTL = secondsToDuration(*fc.NoEndpointQueueTTL)
	}
	if fc.EnableEviction != nil {
		compiled.EnableEviction = *fc.EnableEviction
	}

	compiled.PriorityBands = []PriorityBandConfig{band}
	return compiled
}
```

`CompileEppConfig` 调用点不变（仍 `compiled.FlowControl = compileFlowControl(...)`，`compiler.go:277-280`）。

**取值依据**（沿用 issue §5 估算）：`band0.maxRequests` 推荐 500～2000 起调（峰值 QPS × 可容忍排队时延）；`band0.maxBytes` ≈ `maxRequests × 平均请求字节 × 1.5~2`（短 prompt 给 256Mi~1Gi，长文/多模态 ~1MB 给 4Gi），编译器默认取 **5Gi** 兼容两类形态。

**语义变化说明（评审关注点）**：全局"不限"（缺省/`-1`）时，修复前等效容量是隐藏的 5000，修复后是显式的 10000——上限实际放宽，这是有意为之（band 上限无法表达真正的"不限"，只能显式取大值；10000 配合默认 60s TTL，仅在停发速率 >166 req/s 时才先于 TTL 生效）。

### 2. 回归与新增测试（`model/epp_pool/compiler_test.go`）

| 用例 | 动作 |
| - | - |
| `TestCompileEppConfig_FlowControlJSONShape`（`compiler_test.go:227`） | **修复前即红**：`JSONEq({"maxRequests":"1000","defaultRequestTTL":"30s"})` 因新增 `priorityBands` 字段失败——构成本次修复的 red→green 回归证据。期望值更新为含 `"priorityBands":[{"priority":0,"maxRequests":"1000","maxBytes":"5Gi"}]` |
| `TestCompileEppConfig_FlowControl`（`:188`） | 增加断言：`PriorityBands` 恰含 1 个元素，`Priority==0`、`MaxRequests=="1000"`（与全局一致）、`MaxBytes=="5Gi"` |
| `TestCompileEppConfig_FlowControlMaxRequestsOmitted`（`:206`） | 未设置与显式 `-1` 两种形态：全局仍为空，但 band `MaxRequests=="10000"`，产物含 `priorityBands` |
| 新增 `TestCompileEppConfig_PriorityBand0` | 三种形态矩阵：全局设置 → band0.maxRequests==全局；缺省 → `"10000"`；`-1` → `"10000"`；三者 `maxBytes` 均为 `"5Gi"`、`priority` 均为 0 |
| `TestCompileEppConfig_DeterministicJSON`（`:278`） | 两次编译 marshal 结果一致、可解码回最小 struct 集（新增字段后 roundtrip 必须通过） |
| `epp_data_test.go` `TestCompileIntegration_FromStoredJSON`（`:196`） | 现有断言为按 key 取值/NotContains，新增 `priorityBands` 不破坏；补一条 `flowControl.priorityBands[0].priority == 0` 断言（`max_requests=-1` 形态下 `maxRequests=="10000"`），固定端到端链路 |

### 3. 集成测试（`test/integration/tests/innerapi/epp_data/epp_data_test.go`，扩展现有 IN-EPP-003）

现有 `IN-EPP-003`（`epp_data_test.go:205-330`）已覆盖 #198 的完整链路：OpenAPI 创建带 `flow_control`（`max_requests=200`）的 EPP cluster → DB 存储 → `/inner-api/v1/configs/epp_data/config` 导出编译产物，并逐 key 断言 `flowControl` 段（`:272-276`）。在其 `flowControl` 断言块末尾追加：

```go
bands := flowControl["priorityBands"].([]interface{})
require.Len(t, bands, 1, "band 0 must be explicitly emitted")
band := bands[0].(map[string]interface{})
assert.Equal(t, float64(0), band["priority"])
assert.Equal(t, "200", band["maxRequests"], "band0 mirrors global, no silent truncation")
assert.Equal(t, "5Gi", band["maxBytes"])
```

该断言**修复前即红**（`flowControl["priorityBands"]` 不存在，类型断言失败）、修复后绿，与单测的 JSONShape red→green 相互独立，构成两条回归证据链。**不为三形态矩阵（设置/缺省/-1）新建集成用例**——矩阵是编译器行为，由 §2 单测覆盖；集成层只守"导出链路确实把 band 0 带出去"这一 wiring 契约（与 #191 变更摘要的测试金字塔分工一致）。该用例位于离线 SQLite 必跑集（`test/integration/README.md`，无外部依赖），本地 `run_all_tests.sh` 全量与 CI 持续执行。

### 4. 设计文档同步

| 文档 | 修改 |
| - | - |
| `design-docs/modifications/2026-09-08-epp-scheduling-integration/api-changes.md` §3.2.1（`:222`、`:249`） | 编译规则表 `flow_control` 行补充：始终生成 `priorityBands: [{priority:0, maxRequests:<全局值或"10000">, maxBytes:"5Gi"}]`；`:222` "本期不暴露、用系统默认" 改为"band 0 由编译器显式下发（fixes #198）"。该文件被 `clusters.md:231`、`epp-data.md:78` 作为编译规则权威引用，需与代码同步 |
| `design-docs/api-define/InnerAPI接口定义/epp-data.md`（`:57`） | `flowControl` 示例补充 `priorityBands`，与真实编译产物一致 |
| `ai-gateway-epp/docs/zh_cn/configuration/EPP配置定义说明-epp_config.md`（`:186`） | `priorityBands` 行"用户侧不配置，用系统默认"改为"ai-gateway-api 编译器显式下发 priority 0 band（maxRequests 联动全局、maxBytes 默认 5Gi）" |
| `test/integration/tests/innerapi/design.md` | **既有缺口顺带补齐**：§2 接口列表与 §3 用例统计（合计 18 条）从未收录 epp_data 导出接口（IN-EPP-001~005 无 design 记录，EPP 功能上线时遗留）。本次补：接口列表加 epp_data 行、统计表加 5 条（合计 18→23）、目录结构补 `epp_data/`、新增章节简述五用例（IN-EPP-001 空池降级导出、IN-EPP-002 自动分配恢复 EPP、IN-EPP-003 导出含编译 epp_config 与 assignment（含本次 priorityBands 断言）、IN-EPP-004 同版本增量返回 Data null、IN-EPP-005 手工覆写版本推进） |

用户侧 OpenAPI（`design-docs/api-define/OpenAPI接口定义/clusters.md` `epp_config.flow_control`）**不变**。

### 5. 兼容性与发布说明

- **EPP 侧无需改代码**：ai-gateway-epp 当前代码已支持 `priorityBands` 解析（SC06 流控用例 `ai-gateway-epp/test/implementation/scenario-SC06-flow-control/sc06_test.go:56,141`、`test/integration/endtoend_test.go:163` 均以下发 `priorityBands` 配置运行）。
- **发布顺序**：先升级部署中的 epp 二进制到支持 `priorityBands` 的版本，再升级 ai-gateway-api——避免旧 EPP 收到未知字段（若旧版 strict decode 拒绝未知字段会导致该 cluster 编译失败兜底沿用旧引擎，风险可控但不必要）。
- 导出走版本控制框架：内容变化自动产生新版本，EPP 热加载；存量 `clusters.epp_config` 存储 JSON 不动；WRR 模式不编译、不受影响。

## 影响范围

| 对象 | 影响 |
| - | - |
| `model/epp_pool/compiler.go` | 新增 `PriorityBandConfig` 类型；`FlowControlConfig` 增 `PriorityBands`；`compileFlowControl` 始终产出 band 0（见修复方案 §1） |
| `model/epp_pool/compiler_test.go` | JSONShape 期望值更新（red→green 证据）；FlowControl / MaxRequestsOmitted 增断言；新增 `TestCompileEppConfig_PriorityBand0` |
| `model/epp_pool/epp_data_test.go` | FromStoredJSON 补 priorityBands 断言 |
| `test/integration/tests/innerapi/epp_data/epp_data_test.go` | IN-EPP-003 增加 priorityBands 断言（red→green 证据，见修复方案 §3） |
| `test/integration/tests/innerapi/design.md` | 补齐 epp_data 接口/用例记录（既有缺口，见修复方案 §4） |
| `2026-09-08-epp-scheduling-integration/api-changes.md` §3.2.1 | 编译规则表与"band 级限额"说明更新（见 §3） |
| `design-docs/api-define/InnerAPI接口定义/epp-data.md` | `flowControl` 示例更新 |
| `ai-gateway-epp` 配置定义文档 §7 | `priorityBands` 说明更新 |
| OpenAPI `clusters.md`、`epp_config` 用户侧字段 | **无变更**（存储/校验/回读不变，仅编译产物变化） |
| DB DDL / 存储层 | 无变更 |
| ai-gateway-epp 代码 | 无变更 |
| BFE / conf-agent | 无变更（epp_data 经既有版本控制链路下发） |

## 验证

1. `make test-model` 通过；重点：`go test ./model/epp_pool/...` 中 `TestCompileEppConfig_FlowControlJSONShape` 在修复前代码上应失败（red）、修复后通过（green）。
2. `make test-model-cover-gate`（model ≥70% 覆盖门槛）通过。
3. 产物目检：修复后导出 `epp_data` 的 `flowControl` 段应形如：
   ```json
   "flowControl": {
     "maxRequests": "5",
     "defaultRequestTTL": "10s",
     "noEndpointRequestTTL": "10m0s",
     "priorityBands": [{"priority": 0, "maxRequests": "5", "maxBytes": "5Gi"}]
   }
   ```
4. 升级后观察 EPP 日志/指标：band 0 容量上报值与下发值一致（`RecordFlowControlCapacityUtilization*` 指标，`processor.go:398-418`），不再出现 5000/1GB 隐藏默认值。
5. 集成测试（离线 SQLite 必跑集）：`cd test/integration && go test -v -count=1 -timeout 120s ./tests/innerapi/epp_data/` 通过；其中 IN-EPP-003 在修复前代码上应因 `priorityBands` 缺失而失败（red）、修复后通过（green）。

## 后续事项

1. 按修复方案提交代码与测试，PR 关联并关闭 Issue #198，在 Issue 中回复根因与修复摘要；
2. issue 中提及的 `enable_eviction` 关闭事项（"另见 issue"）不在本次范围，另行跟进；
3. 未来若 ai-gateway-epp 引入 InferenceObjective reconciler（多优先级带），`compileFlowControl` 的 band 生成需扩展为按优先级带下发，本方案的 band 0 显式下发是该扩展的最小基线。
