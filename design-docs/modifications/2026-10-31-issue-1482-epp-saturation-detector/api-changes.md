# 并发饱和检测配置（issue #1482）：ai-gateway-api 接口契约变更

> 日期：2026-10-31　　issue：[bfenetworks/bfe#1482](https://github.com/bfenetworks/bfe/issues/1482)
> 上游：`change-summary.md`（背景与决策一览）；逐文件实现见 `design-changes.md`
> 影响接口：OpenAPI `POST/PATCH /open-api/v1/clusters` 的 `epp_config`；InnerAPI `GET /inner-api/v1/configs/epp-data`（只增内容，结构不变）

## 1. OpenAPI：`epp_config` 新增 `saturation_detector`

### 1.1 新表：`epp_config.saturation_detector`

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 | 合法性条件 |
| - | - | - | - | - | - |
| mode | string | **饱和判定模式**：决定 EPP 用哪个检测器判断"池已饱和、需要排队/拒绝" | N | `utilization`：按 KV cache 利用率与后端排队深度判（**默认，等价改动前行为**）；`concurrency`：按**在飞并发**判（含已派发到后端的请求，池容量 = 端点数 × `max_concurrency`） | 非必填；默认 `utilization`；枚举 `utilization`、`concurrency` |
| max_concurrency | int | `mode=concurrency` 时**每端点的理想并发上限** | 条件可选 | 池容量 = 端点数 × 该值（`headroom` 允许突发）；建议取后端推理引擎的实际并发上限（vLLM `--max-num-seqs`） | 仅 `mode=concurrency` 允许；取值 `> 0`；默认 `100` |
| max_token_concurrency | int | `concurrency_mode=tokens` 时每端点的 token 并发上限 | N | 仅在按 token 计数时生效 | 仅 `mode=concurrency` 允许；取值 `> 0`；默认 `1000000` |
| headroom | number | 允许的突发余量（相对 `max_concurrency` 的比例） | N | `0` = 不允许突发（默认）；`0.2` = 允许 20% 突发（仅放宽**调度过滤**上限，池级背压仍按严格容量） | 仅 `mode=concurrency` 允许；取值 `[0, 1]`；默认 `0` |
| concurrency_mode | string | 并发计数模式 | N | `requests`：按请求数（默认）；`tokens`：按 token 数 | 仅 `mode=concurrency` 允许；枚举 `requests`、`tokens`；默认 `requests` |

> **`mode=utilization` 时不得出现** `max_concurrency` / `max_token_concurrency` / `headroom` / `concurrency_mode`——出现即 `422`（错误信息明确提示 `require mode=concurrency`）。理由：静默忽略会让运维误以为生效，正是问题6 投诉的形态重演。

### 1.2 `epp_config` 主表变更

在既有 `**表：EPP调度配置（epp_config）**` 末尾（`flow_control` 之后）新增一行：

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 | 合法性条件 |
| - | - | - | - | - | - |
| saturation_detector | object | 饱和判定模式与并发阈值（问题6）；未配置 = `mode=utilization`（改动前行为） | N | 见下方 表：epp_config.saturation_detector | 非必填；若传入须满足子字段约束 |

### 1.3 `flow_control.enable_eviction` 文案调整

| 字段 | 变更前 | 变更后 |
| - | - | - |
| `flow_control.enable_eviction` | `"epp_config.flow_control.enable_eviction is not supported, keep false"` | 同语义，文案补充指向：`"... is not supported yet (see the EPP overload-protection plan)"`；**行为不变**（`true` 仍拒绝） |

### 1.4 与既有字段的关系（文档必须写清）

| 字段 | `mode=utilization`（默认） | `mode=concurrency` |
| - | - | - |
| `kv_cache_utilization_max` | 参与饱和判定（检测器阈值）+ 候选过滤（`utilization-filter`） | **仅**作为候选过滤阈值；不参与"是否排队"判定 |
| `waiting_queue_max` | 同上（检测器 `queueDepthThreshold` 镜像该值，未启用时默认 5） | **仅**作为候选过滤阈值（`waiting_queue_max > 0` 时） |
| `running_requests_max` | 仅候选过滤 | 仅候选过滤 |
| `metrics_staleness_threshold_ms` | 坏后端剔除阈值（两种模式都生效） | 同左 |
| `flow_control.max_requests` / `queue_ttl` / `no_endpoint_queue_ttl` | 队列容量与排队预算（两模式都生效，但只有"被判饱和"的请求才会进队列） | 同左 |

### 1.5 请求示例

```json
{
  "name": "cluster_epp",
  "basic": { "protocol": "http" },
  "balance_mode": "EPP",
  "epp_config": {
    "load_profile": "balanced",
    "kv_cache_utilization_max": 0.9,
    "saturation_detector": {
      "mode": "concurrency",
      "max_concurrency": 256,
      "concurrency_mode": "requests",
      "headroom": 0
    },
    "flow_control": {
      "max_requests": 2,
      "queue_ttl": 1,
      "no_endpoint_queue_ttl": 1
    }
  }
}
```

`PATCH` 语义同 `POST`：`saturation_detector` 为**整体替换**（与 `epp_config` 其它子对象一致），省略即删除该段（回到 `mode=utilization`）。

## 2. InnerAPI：`epp-data` 的编译产物变化

`GET /inner-api/v1/configs/epp-data` 的结构**不变**（仍是 `EndpointPickerConfig`），只有 `plugins[]` 内容随配置变化：

| 配置 | `plugins[]` 中的检测器 | 其它 |
| - | - | - |
| 未配置 `saturation_detector` 或 `mode=utilization` | `{ "name": "saturation-detector", "type": "utilization-detector", "parameters": { kvCacheUtilThreshold, queueDepthThreshold, stalenessPolicy, headroom:0, metricsStalenessThreshold } }` | 与改动前逐字段一致 |
| `mode=concurrency` | `{ "name": "saturation-detector", "type": "concurrency-detector", "parameters": { maxConcurrency, maxTokenConcurrency, headroom, concurrencyMode } }` | `flowControl.saturationDetector.pluginRef` 仍为 `saturation-detector`；不新增/删除插件实例 |

示例（`mode=concurrency`、`max_concurrency=256`）：

```json
{
  "plugins": [
    { "name": "ep-discover", "type": "cluster-table-discovery", "parameters": {} },
    { "name": "saturation-detector", "type": "concurrency-detector",
      "parameters": { "maxConcurrency": 256, "maxTokenConcurrency": 1000000,
                      "headroom": 0, "concurrencyMode": "requests" } }
  ],
  "flowControl": {
    "saturationDetector": { "pluginRef": "saturation-detector" },
    "maxRequests": "2",
    "defaultRequestTTL": "1s",
    "noEndpointRequestTTL": "1s",
    "priorityBands": [{ "priority": 0, "maxRequests": "2", "maxBytes": "5Gi" }]
  }
}
```

> `inFlightLoadProducerName` 不下发：EPP 侧 `inflightload` 已注册为默认 producer，空值走默认数据源。
> 参数名沿用 llm-d `apiConfig` 的 JSON 字段（`maxConcurrency` 等 PascalCase），与既有 `utilization-detector` 的参数风格一致。

## 3. 错误码与校验

| 场景 | 返回 |
| - | - |
| `mode` 非 `utilization`/`concurrency` | `422`：`epp_config.saturation_detector.mode must be "utilization" or "concurrency", got "..."` |
| `mode=utilization` 但填了并发阈值 | `422`：`epp_config.saturation_detector.max_concurrency/max_token_concurrency/headroom/concurrency_mode require mode=concurrency` |
| `max_concurrency` / `max_token_concurrency` `<= 0` | `422`：`... must be > 0` |
| `headroom` 不在 `[0,1]` | `422`：`... must be in [0, 1]` |
| `concurrency_mode` 非 `requests`/`tokens` | `422`：`... must be "requests" or "tokens"` |
| `balance_mode=EPP` 且 `epp_config` 只含空对象 | `422`：沿用既有 `epp_config is required when balance_mode is "EPP"`（`IsEmpty` 已把新段计入非空） |
| 未知子字段（如 `saturation_detector.foo`） | `422`：`ParseEppConfig` 的 `DisallowUnknownFields` |

## 4. 兼容性

- **API 兼容**：新增可选字段，旧客户端不传即旧行为；`PATCH` 未带该段时整体替换语义会**删除**既有 `saturation_detector`（与 `epp_config` 其它子对象一致，控制台回显后提交不受影响）。
- **EPP 版本**：`concurrency-detector` 需 EPP 部署版本含 `llm-d-router ≥ v0.11.0-rc.2`；若下发到旧 EPP，引擎编译失败并保留上一代配置（服务不中断，但该集群新配置不生效）。**发布顺序：先升 EPP，再开放控制台该项**。
- **报表/bfe**：无影响；bfe 只关心 EPP 是否下发 `ImmediateResponse`（已在 `e2cd18a3` 支持）。