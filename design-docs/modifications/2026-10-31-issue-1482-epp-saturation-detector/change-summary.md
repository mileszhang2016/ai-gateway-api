# EPP 过载保护：并发饱和检测可配（issue #1482）——ai-gateway-api 侧变更摘要

> 日期：2026-10-31　　issue：[bfenetworks/bfe#1482](https://github.com/bfenetworks/bfe/issues/1482)
> 上游背景：v0.7.0.1 测试发现问题6"EPP 流控不起过载保护作用"——后端实例都已满负荷（各 256 在处理 + 10~16 排队）时，`queue_ttl=1s` / `max_requests=2` 四组配置下 8 个请求全部 200、耗时 14~31 秒，没有任何拒绝或超时。
> 本目录产出：`change-summary.md`（本文）、`api-changes.md`（OpenAPI/InnerAPI 契约）、`design-changes.md`（逐文件改动 / 语义决策 / 兼容性 / 测试计划）
> 配套（同一问题拆两半）：
> - **bfe**（出口侧，已交付）：`v1.9.0-dev` commit `e2cd18a3`——识别 ext_proc `ImmediateResponse`（429/503）、不重试、不回退本地均衡、不计熔断，映射为 `EPP_OVERLOAD`(429) / `EPP_NO_CAPACITY`(503) 并透传 `Retry-After`；修改说明见 `bfe/docs/zh_cn/modifications/2026-10-31-issue-1482-epp-reject-short-circuit/`
> - **ai-gateway-epp**（零代码）：`concurrency-detector` 与 `inflightload` 默认 producer 已注册（`cmd/epp/plugins.go:96,97,101`）；仅要求部署版本 ≥ 含 `llm-d-router v0.11.0-rc.2`

## 1. 背景

问题6 是两侧共同造成的：

| 侧 | 缺口 | 归属 |
|---|---|---|
| 判定 | EPP 的饱和检测只按 KV cache 利用率 / 后端排队深度判；FC 的 occupancy 只统计 EPP 内排队，不看已派发在飞 → 后端 HTTP 饱和时请求被**立即 admit** | 本仓（配置能力）+ ai-gateway-epp（插件已就绪） |
| 出口 | EPP 的准入拒绝（`ImmediateResponse` 429/503）被 bfe 丢弃并回退本地均衡 | bfe（已修复 `e2cd18a3`） |

只修出口，`max_requests`/`queue_ttl` 仍没有机会触发；只修判定，客户端仍看不到明确错误码。两侧同批交付才闭环。

本仓的核心问题：`compileSaturationDetectorPlugin`（`model/epp_pool/compiler.go:383`）把检测器类型**写死**为 `utilization-detector`，`flow_control` 只暴露 `max_requests`/`queue_ttl`/`no_endpoint_queue_ttl`/`enable_eviction`（`model/epp_pool/epp_config.go:83-86`），**没有任何"按在飞并发判饱和"的开关**。

## 2. 目标与范围

| 目标 | 说明 | 归属 |
|---|---|---|
| G1 让"后端在飞饱和"能触发排队/拒绝 | 新增 `epp_config.saturation_detector`，`mode=concurrency` 时编译为 llm-d 内建的 `concurrency-detector`，按「端点数 × max_concurrency」判定池饱和 | **本仓主体** |
| G2 存量行为不变 | 未配置该段时编译输出与现状**逐字节一致**（仍 `utilization-detector`） | 本仓 |
| G3 控制台可配 | `ai-gateway-web` 增加模式选择与阈值表单，`enable_eviction` 置灰说明 | ai-gateway-web |
| G4 端到端可验证 | bfe 出口侧已固化（SC14 TC-09~TC-12）；判定侧需 api 导出正确 + 饱和后端场景联调 | 本仓 + integration-test |

**不做（本卡）**：

- `enable_eviction`（在飞负优先级请求驱逐）：需 ai-gateway-epp 注册 eviction 插件、bfe 处理 response 阶段 IR，另开卡；
- 两种检测器"取更严"的组合逻辑：llm-d 的 `flowControl.saturationDetector` 是单引用，组合需上游支持；
- 检测器阈值按模型自动推导：`max_concurrency` 由运维填写（建议 = vLLM `--max-num-seqs`）。

## 3. 决策一览（详细论证见 `design-changes.md` §1）

| # | 决策 | 取法 |
|---|---|---|
| D1 | 字段形态 | `epp_config.saturation_detector` 子对象（`mode` + 四个并发参数），与既有"简化用户形态"一致；不用散落在 `flow_control` 里 |
| D2 | 缺省 | `mode` 未设 = `utilization`（现状）；编译输出与改动前完全一致 |
| D3 | `utilization` 模式下填并发阈值 | **报错**（"require mode=concurrency"），不静默忽略——避免"配了没生效"的二次投诉 |
| D4 | 并发参数默认值 | 与 EPP 插件同源：`max_concurrency=100`、`max_token_concurrency=1000000`、`headroom=0`、`concurrency_mode=requests` |
| D5 | `kv_cache_utilization_max` / `waiting_queue_max` 在 concurrency 模式下的作用 | 不再参与"是否排队"判定，但仍是 `util-scheduler-filter` 的候选过滤条件（文档写清） |
| D6 | 与 bfe 的耦合 | 无：bfe 不感知检测器类型，只看 EPP 是否下发 IR |
| D7 | `enable_eviction` | 保持拒绝，只改错误文案（指向"另开卡"），不做行为开放 |

## 4. 跨仓交付顺序

1. **ai-gateway-epp 版本**：先确认部署版本含 `concurrency-detector`（`llm-d-router ≥ v0.11.0-rc.2`）；旧 EPP 收到该插件类型会导致引擎编译失败并保留上一代配置（不中断服务，但集群配置不生效）。
2. **本仓**：配置字段 + 校验 + 编译 + 单测 + api-define 文档。
3. **ai-gateway-web**：表单与置灰。
4. **联调验收**：饱和后端 + 8 并发 → 期望出现 `EPP_OVERLOAD`(429)（bfe 侧已就绪）。
5. **bfe**：无需再改（`e2cd18a3` 已包含出口侧）。

## 5. 验收标准（复现原测试方法）

| 步骤 | 期望 |
|---|---|
| 两 vLLM 各 272 长请求，`num_requests_waiting ≥ 10` | 池饱和 |
| 集群 `epp_config` 配 `saturation_detector.mode=concurrency`、`max_concurrency=256`、`headroom=0`，`flow_control.max_requests=2`、`queue_ttl=1` | 导出 `epp-data` 中 `plugins[]` 出现 `type=concurrency-detector` 且参数正确 |
| 8 个并发非流式请求（`max_tokens=32`，超时 60s） | **部分请求秒级返回 429 `EPP_OVERLOAD`**（或 503 `EPP_NO_CAPACITY`），不再出现"全部 200、14~31 秒" |
| 对照组：`max_requests=-1` | 排队但受 `queue_ttl` 约束，超时被拒 |
| 回归：未配置 `saturation_detector` 的集群 | 导出与改动前一致，行为不变 |

## 6. 风险摘要

1. 阈值设得过大等于没保护、过小会误拒——控制台必须给默认值与帮助文案（建议 = vLLM `--max-num-seqs`）；
2. 在飞计数依赖 EPP 的 `inflightload` producer 刷新窗口，派发瞬间可能有过冲；
3. `utilization` 与 `concurrency` 互斥选择，运维若同时"想保留原阈值"需文档说明二者分工（D5）。