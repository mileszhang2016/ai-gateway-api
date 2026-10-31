# 并发饱和检测配置（issue #1482）：ai-gateway-api 逐文件改动与验证计划

> 日期：2026-10-31　　issue：[bfenetworks/bfe#1482](https://github.com/bfenetworks/bfe/issues/1482)
> 上游：`change-summary.md`（背景/决策一览）、`api-changes.md`（OpenAPI 与 InnerAPI 契约）
> 代码基线：ai-gateway-api `v0.1.0-dev`
> 状态：**方案待实施**

## 1. 语义决策与论证

### D1 为什么是独立子对象 `saturation_detector`，而不是塞进 `flow_control`

`flowControl` 在 EPP 侧由两件事组成：**队列容量/TTL**（`maxRequests`/`defaultRequestTTL`/`noEndpointRequestTTL`/`priorityBands`）与**饱和检测器引用**（`saturationDetector.pluginRef`）。检测器本身是 `plugins[]` 里的独立插件实例（type + parameters）。把"选哪个检测器 + 阈值"放进 `flow_control` 会让用户以为它是队列参数；独立子对象与 EPP 的模型同构，也便于后续扩展（如 D4 的新参数）。

### D2 默认值必须"逐字节不变"

改动的验收底线是"未配置该段时，`epp-data` 导出与改动前完全一致"。因此 `compiler` 分两条分支，`utilization` 分支的代码**原样保留**（含 `stalenessPolicy=ignore`、`headroom=0`、`queueDepthThreshold` 镜像 `waiting_queue_max` 或默认 5），并新增一个 golden 单测锁住两者一致。

### D3 utilization 模式下填并发阈值 = 报错

用户最常见的误用是"以为改了 `max_concurrency` 就生效"（问题6 的投诉本质就是"配了没生效"）。静默忽略会重演同样的困惑，因此显式拒绝，错误信息直接指出前置条件。

### D4 并发参数与 EPP 插件默认值同源

`llm-d-router` 的 `concurrency-detector` 的 `apiConfig` 默认值：`maxConcurrency=100`、`headroom=0.0`、`concurrencyMode=requests`、`maxTokenConcurrency=1000000`。api 侧 `Effective*` 访问器逐一镜像，保证"api 不填"与"EPP 收不到"等价，避免双份默认值漂移。

### D5 concurrency 模式下 `kv_cache_utilization_max` / `waiting_queue_max` 仍生效，但只作用于"候选过滤"

`compileUtilFilterPlugin`（`compiler.go:349-377`）用这两个值构造 `utilization-filter` 的 conditions；饱和检测器用它们判"是否暂停派发"。切到 `concurrency` 后只解除后者，候选过滤照旧——这是有意的：一个端点的 KV 已满仍不该被调度过去。文档必须写清，否则运维会以为"改了没用"。

### D6 与 bfe 解耦

bfe 只关心 EPP 是否下发 ext-proc `ImmediateResponse`（`e2cd18a3` 已支持 429/503），不感知检测器类型。因此本仓改动可在 bfe 之后独立发布。

### D7 `enable_eviction` 仅改文案

`enable_eviction`（在飞负优先级请求驱逐）需要：ai-gateway-epp 注册 eviction 插件、bfe 处理 response 阶段 `ImmediateResponse`。属独立卡，本卡保持 `true → 422`，只把文案指向"另开卡"，避免用户误以为配置生效。

## 2. 逐文件改动

### 2.1 `model/epp_pool/epp_config.go`

```go
// 饱和判定模式（issue #1482）。默认 utilization = 改动前行为。
const (
	SaturationDetectorUtilization = "utilization"
	SaturationDetectorConcurrency = "concurrency"
)

// 并发计数模式，与 llm-d concurrency-detector 的 concurrencyMode 同名同值。
const (
	ConcurrencyModeRequests = "requests"
	ConcurrencyModeTokens   = "tokens"
)

// 与 llm-d concurrency-detector 的 apiConfig 默认值同源。
const (
	DefaultSaturationDetectorMode = SaturationDetectorUtilization
	DefaultMaxConcurrency         = 100
	DefaultMaxTokenConcurrency    = 1_000_000
	DefaultConcurrencyHeadroom    = 0.0
	DefaultConcurrencyMode        = ConcurrencyModeRequests
)

// SaturationDetectorSimplified 是 epp_config.saturation_detector 的用户形态。
type SaturationDetectorSimplified struct {
	Mode                *string  `json:"mode,omitempty"`
	MaxConcurrency      *int     `json:"max_concurrency,omitempty"`
	MaxTokenConcurrency *int     `json:"max_token_concurrency,omitempty"`
	Headroom            *float64 `json:"headroom,omitempty"`
	ConcurrencyMode     *string  `json:"concurrency_mode,omitempty"`
}
```

`EppConfigSimplified` 增加字段（放在 `FlowControl` 之后，作为第 ④ 段）：

```go
	// ── ④ 饱和判定（问题6 / issue #1482） ──
	SaturationDetector *SaturationDetectorSimplified `json:"saturation_detector,omitempty"`
```

改动点：

| 函数 | 改动 |
| - | - |
| `IsEmpty()`（`:108-122`） | 追加 `c.SaturationDetector == nil`——否则"只配了 saturation_detector"的集群会被判为 `epp_config` 为空，`balance_mode=EPP` 直接 422 |
| `Validate()`（`:124-181`） | 增加 `if c.SaturationDetector != nil { if err := c.SaturationDetector.Validate(); err != nil { return err } }` |
| `(*SaturationDetectorSimplified) Validate()`（新增） | 枚举/范围校验，规则见 `api-changes.md` §3；`utilization` 模式出现并发字段 → 报错 |
| `EffectiveSaturationDetectorMode()` 等五个访问器（新增） | 与 `EffectiveKVCacheUtilizationMax()`（`:240`）/`EffectiveWaitingQueueMax()`（`:248`）同款 |
| `(*FlowControlSimplified) Validate()`（`:206-210`） | `enable_eviction=true` 的文案补充"see the EPP overload-protection plan"，行为不变 |

### 2.2 `model/epp_pool/compiler.go`

| 位置 | 改动 |
| - | - |
| 常量块（`:146-166`） | `pluginTypeSaturationDetector` 拆成 `pluginTypeUtilizationDetector = "utilization-detector"` 与 `pluginTypeConcurrencyDetector = "concurrency-detector"`；`pluginNameSaturationDetector`（`:149`）不变（`flowControl.saturationDetector.pluginRef` 仍指向它） |
| `compileSaturationDetectorPlugin`（`:383-407`） | 按 `EffectiveSaturationDetectorMode()` 分支：`utilization` 分支保持原参数不动；`concurrency` 分支输出 `{maxConcurrency, maxTokenConcurrency, headroom, concurrencyMode}`（**不**下发 `inFlightLoadProducerName`，走 EPP 默认在飞 producer） |
| `compileFlowControl`（`:410-442`） | 不改（`saturationDetector.pluginRef` 与 band 0 逻辑与检测器类型无关） |
| `compileUtilFilterPlugin`（`:349-377`） | 不改（D5：候选过滤两种模式都用同一套阈值） |

### 2.3 InnerAPI `epp-data`（`model/epp_pool/epp_data.go`）

**不改代码**：`PluginConfig` 的 `Type`/`Parameters` 已是通用结构，`EndpointPickerConfig` 直接序列化。契约变化记录在 `api-changes.md` §2，并由单测锁定。

### 2.4 文档

| 文件 | 改动 |
| - | - |
| `design-docs/api-define/OpenAPI接口定义/clusters.md` | `epp_config` 主表新增 `saturation_detector` 一行 + 新增"表：epp_config.saturation_detector" + `enable_eviction` 文案 + 约束段补"`mode=utilization` 下不得填并发阈值" |
| `design-docs/api-define/InnerAPI接口定义/epp-data.md` | 说明检测器 `type` 可为 `utilization-detector`（默认）或 `concurrency-detector`，参数清单与默认值 |

### 2.5 相邻仓（不在本仓 PR 内，需协同）

| 仓 | 改动 |
| - | - |
| ai-gateway-web | `Clusters/components/GatewayConfig.vue`：新增"饱和判定模式"单选（利用率/并发）与条件显示的 `max_concurrency`/`max_token_concurrency`/`headroom`/`concurrency_mode`；`enable_eviction` 置灰 + tooltip；`Review.vue` 回显；i18n |
| ai-gateway-epp | 无需改代码；发布说明要求版本 ≥ 含 `llm-d-router v0.11.0-rc.2` |
| bfe | 已交付（`e2cd18a3`） |

## 3. 测试计划

### 3.1 单元测试

| 文件 | 用例 |
| - | - |
| `model/epp_pool/epp_config_test.go` | `TestSaturationDetectorValidate`（表驱动）：`mode` 非法 / `utilization` + `max_concurrency` / `headroom=-0.1` / `headroom=1.5` / `concurrency_mode=foo` / `max_concurrency=0` 各报错；`concurrency` 全量合法；`concurrency` 只填 `mode` 合法（其余取默认） |
| 同上 | `TestEppConfigIsEmpty`：只含 `saturation_detector` 时 `IsEmpty()==false`（保证 `balance_mode=EPP` 不被误判为空） |
| `model/epp_pool/compiler_test.go` | `TestCompileSaturationDetectorUtilizationUnchanged`：未配置 / `mode=utilization` 两种输入 → 与既有基线**逐字段一致**（golden，防回归） |
| 同上 | `TestCompileSaturationDetectorConcurrency`：全量参数 → `type=concurrency-detector`、四个参数值正确、`flowControl.saturationDetector.pluginRef=saturation-detector`、`plugins[]` 中仍只有一个 `saturation-detector` 实例 |
| 同上 | `TestCompileSaturationDetectorConcurrencyDefaults`：只设 `mode` → `100/1000000/0/requests` |
| 同上 | `TestCompileFeatureGateUnaffected`：`saturation_detector` 不改变 `featureGates`（仍只由 `flow_control` 是否存在决定） |
| `model/epp_pool/epp_data_test.go` | `epp-data` 样例快照：`concurrency` 集群的导出 JSON 包含 `concurrency-detector` 与四个参数；`utilization` 集群与既有样例逐字节一致 |
| `model/icluster_conf/epp_test.go`（或等价位置） | `validateClusterBalanceConfig`：`balance_mode=EPP` + 仅 `saturation_detector` → 通过；非法子字段 → 422 |

### 3.2 集成测试（本仓 `test/integration`）

- schema 测试（`tests/schema/openapi/cluster.go`）把 `saturation_detector` 及子字段加入字段白名单；
- 新增/扩展用例：创建 `balance_mode=EPP` 集群（`mode=concurrency`）→ `GET /clusters/{name}` 回显 → InnerAPI `epp-data` 导出断言（`plugins[]` 与 `flowControl`）；
- PATCH 用例：`mode=concurrency → utilization` 整体替换、以及"只 PATCH 其它字段导致该段被删除"的行为（与 `epp_config` 其它子对象一致）；
- 负向：`mode=utilization` + `max_concurrency` → 422。

### 3.3 端到端（integration-test 仓）

- 饱和后端 + 8 并发（问题6 原测试方法）：期望出现 `EPP_OVERLOAD`(429)；bfe 出口侧已固化（SC14 TC-09~TC-12）；
- 该端到端场景依赖真实 EPP + vLLM/sim，实施时与本仓 PR 分开跟踪。

## 4. 兼容性与灰度

- **API 兼容**：新增可选字段；不传即旧行为；`PATCH` 的"整体替换"语义与 `epp_config` 其它子对象一致（控制台回显后提交不会丢）。
- **导出兼容**：`utilization` 分支输出与改动前逐字节一致（golden 单测锁定）；`epp-data` 结构不变。
- **EPP 版本**：`concurrency-detector` 需 EPP 支持；旧 EPP 编译失败会保留上一代配置（不中断服务，但该集群新配置不生效）——发布说明写明"先升 EPP 再开放配置"。
- **灰度建议**：先在一个预发集群开 `mode=concurrency`，观察 bfe `EPP_OVERLOAD` 比例与 `epp_fallback_local_total`（应不增长）；确认阈值合理后再推控制台可见。

## 5. 风险与开放问题

1. **阈值语义易误解**：`max_concurrency` 是"每端点理想并发"，池容量 = 端点数 × 该值；建议控制台默认填后端引擎的 `--max-num-seqs`，并在帮助文案里给公式。
2. **在飞计数刷新窗口**：`inflightload` producer 的采集窗口决定过冲幅度，端到端验收时要留观察项（派发后请求数应在 1 个采集周期内反映）。
3. **utilization 与 concurrency 互斥**：单引用结构决定同一集群只能选一个；"两者取更严"需上游支持，列为后续增强。
4. **`PATCH` 删除语义**：客户端若只 PATCH 别的字段，`saturation_detector` 会被整体替换为空（与其它子对象一致），需在 api-define 文档明示，避免"配置莫名回到 utilization"。
5. **`RunningRequestsMax` 与 `MaxConcurrency` 概念相近**：前者是候选过滤（超过就不派给该端点），后者是池级背压判定；文档要并列解释，否则运维会用错一个。
6. **数字边界**：`max_concurrency` 与 `max_token_concurrency` 用 Go `int`（64 位平台）；JSON 侧 `>0` 校验即可，不做上限约束（超大值等价"不保护"，属配置错误，可在控制台加软提示）。

## 6. 实施顺序与产出

1. `epp_config.go`（结构/校验/访问器）+ 单测；
2. `compiler.go`（分支 + 常量）+ golden 与参数单测；
3. `epp_data_test.go` 快照 + schema 白名单 + 集成用例；
4. api-define 两份文档；
5. 控制台（web）与端到端（integration-test）分别跟踪。

产物：本目录三份文档 + 上述代码/测试/文档改动。

## 7. 实施记录（2026-10-31）

### 7.1 改动文件

| 文件 | 改动 |
| - | - |
| `model/epp_pool/epp_config.go` | 新增模式/计数模式常量与五个默认值（与 llm-d 插件同源）；新增 `SaturationDetectorSimplified` 与 `EppConfigSimplified.SaturationDetector`；`IsEmpty()` 计入新段；`Validate()` 委派 + `SaturationDetectorSimplified.Validate()`（`utilization` 下出现并发阈值报错、范围校验）；新增 `EffectiveMode/EffectiveMaxConcurrency/EffectiveMaxTokenConcurrency/EffectiveHeadroom/EffectiveConcurrencyMode`；`enable_eviction` 文案改为 "not supported yet" |
| `model/epp_pool/compiler.go` | `pluginTypeSaturationDetector` 拆为 `pluginTypeUtilizationDetector` / `pluginTypeConcurrencyDetector`；`compileSaturationDetectorPlugin` 按 mode 分支（`utilization` 分支与改动前逐字段一致；`concurrency` 分支输出四个参数、不下发 `inFlightLoadProducerName`） |
| `model/epp_pool/epp_config_test.go` | 新增 `TestSaturationDetectorSimplified_Validate`（9 个子用例，含"utilization 拒并发阈值"表驱动与严格解析未知字段）、`TestSaturationDetectorSimplified_EffectiveDefaults`；`TestEppConfigSimplified_IsEmpty` 增加"仅 saturation_detector"用例 |
| `model/epp_pool/compiler_test.go` | 既有断言改用 `pluginTypeUtilizationDetector`；新增 `TestCompileEppConfig_SaturationDetectorUtilizationUnchanged`（未配置 vs 显式 utilization 输出相等 + featureGates 不变）、`..._Concurrency`（type/四参数/pluginRef/单实例/util-filter 不受影响）、`..._ConcurrencyDefaults` |
| `model/epp_pool/epp_data_test.go` | 新增 `TestCompileIntegration_ConcurrencyDetector`：存储 JSON → parse → validate → compile → marshal 全链路，锁定 InnerAPI `epp-data` 契约（type/参数/pluginRef/单实例/filter 条件） |
| `test/integration/tests/schema/openapi/cluster.go` | 新增 `EppConfigSaturationDetectorSchema` 并挂到 `EppConfigSchema`（openapi schema 白名单） |
| `test/integration/tests/clusters/epp/epp_test.go` | 新增 CL-EPP-016（仅配 saturation_detector 视为非空、创建成功并回读）、CL-EPP-017（合法全量配置原样回读）、CL-EPP-018（9 类非法取值 422）、CL-EPP-019（PATCH 整体替换：concurrency → utilization 且并发阈值随段删除） |
| `design-docs/api-define/OpenAPI接口定义/clusters.md` | `epp_config` 主表新增 `saturation_detector` 行 + 新增"表：epp_config.saturation_detector"（含与 `kv_cache_utilization_max`/`waiting_queue_max` 的关系说明）；`enable_eviction` 文案；约束段校验清单；请求/响应示例补该段 |
| `design-docs/api-define/InnerAPI接口定义/epp-data.md` | 说明检测器 type 随 mode 变化、两套参数清单与默认值、`inFlightLoadProducerName` 不下发、EPP 版本要求 |

### 7.2 验证结果

```
go build ./...                                   → 通过
go vet ./model/... ./lib/... ./endpoints/...      → 通过
go test ./model/... -count=1                     → 全部通过（含 epp_pool 新增 6 个用例/子组）
集成（test/integration）：
  tests/schema/openapi   ok(6.4s)   tests/schema/innerapi ok(6.1s)
  tests/clusters/epp     ok(4.0s)   tests/innerapi/epp_data ok(4.1s)
```

> 集成harness 使用仓库根目录的**预构建** `ai-gateway-api.exe`（`testutil.StartServer` 只做复制，不编译）。首跑 CL-EPP-016~019 因二进制陈旧报 `json: unknown field "saturation_detector"`，执行 `go build -o ai-gateway-api.exe .` 后全部通过——这是本仓集成测试的既有前提（"run 'make build' first"），已在本次记录，后续改 model 后需先重编译。

### 7.3 与方案稿的偏差

1. **`enable_eviction` 文案**：方案稿写"补充指向问题6/方案 C"，实际落为更中性的 "not supported yet, keep false"（避免把内部方案编号写进对外错误信息）。
2. **`compileSaturationDetectorPlugin` 的 nil 安全**：`sd := conf.SaturationDetector` 可能为 nil，分支判断与默认值全部走 `(*SaturationDetectorSimplified).Effective*` 的 nil 接收者，未额外加空指针分支（与既有 `Effective*` 风格一致）。
3. **`inFlightLoadProducerName`**：方案稿已定"不下发"，实现中用单测显式断言该键不存在，防止后人"补全"成一个空串键（空串与缺省对 EPP 语义相同但会让导出噪音化）。
4. **集成用例编号**：方案稿只写"新增/扩展用例"，实施时沿用该文件既有序号，落为 CL-EPP-016~019。
5. **schema 白名单**：方案稿只说"把字段加入白名单"，实施时补了子对象 schema（`EppConfigSaturationDetectorSchema`）并带枚举，使未知子字段/枚举越界在 schema 层即可被发现。

### 7.4 未做（按方案列为独立卡）

- `enable_eviction`（在飞负优先级驱逐）：需 ai-gateway-epp 注册 eviction 插件 + bfe 处理 response 阶段 `ImmediateResponse`；
- 控制台（ai-gateway-web）表单项与置灰：相邻仓，随本卡发布跟进；
- 端到端（integration-test 仓）饱和后端场景：依赖真实 EPP + vLLM/sim，另行跟踪；
- 两种检测器"取更严"的组合：需上游支持。