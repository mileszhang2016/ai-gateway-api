# Claude 协议转发支持

## 1. 背景

BFE 数据面在 AI Gateway 路径上已支持 OpenAI 兼容协议的转发。为了支持 Anthropic Claude Messages API 的原样转发，BFE 需要识别 Claude 协议风格（`anthropic`），并在认证头、版本头、Usage 解析等环节做差异化处理。

在 cluster/provider 分离后的控制面中，Provider 通过 `model_protocols` 字段声明支持的协议（如 `openai`、`anthropic`）。BFE 需要依据该字段判断请求协议风格是否被目标集群支持，并在不匹配时直接拒绝。因此，`ai-gateway-api` 需要把 provider 的 `model_protocols` 透传到导出的 BFE `AIConf` 中。

## 2. 目标

- 让 BFE 在导出配置中拿到每个 cluster 对应 provider 支持的协议列表（`AIConf.ModelProtocols`）。
- 让 `/providers/tools/discover-models` 在请求参数 `model_protocol=anthropic` 时能正确解析 Anthropic `/v1/models` 响应。
- 清理已废弃的 `ai-gateway-api/conf/ai` 目录。
- 保持 `/providers`、`/clusters` 等 OpenAPI 接口契约不变。

## 3. 关键设计决策

| 决策 | 说明 |
|------|------|
| `model_protocols` 已覆盖 `anthropic` | `/providers` 数据模型已包含该枚举，本期直接启用，无需新增接口。 |
| `AIConf` 新增 `ModelProtocols` | 阶段一最小改动：把 provider 的 `model_protocols` 原样透传给 BFE。 |
| 解析器下沉到代码 | 原 `conf/ai/model_definition.json` 中的 anthropic 解析规则，迁移到代码内的 `modelProtocolParsers`。 |
| 删除 `conf/ai` 目录 | `/model-provider-types` 与 `/tools/get-models-from-provider` 已移除，该目录全部废弃。 |
| 不依赖 provider 名称推断协议 | `provider` 为用户自定义名称，协议能力必须以 `model_protocols` 为准。 |

## 4. 控制面数据流

```text
OpenAPI /providers
    └── model_protocols: ["anthropic"]
            │
            ▼
OpenAPI /clusters
    └── llm_config.provider = "my-anthropic"
            │
            ▼
model/icluster_conf/cluster.go
    └── 查询 provider
    └── newAIConf(..., provider.ModelProtocols)
            │
            ▼
InnerAPI /configs/tls_conf/server_data_conf
    └── ClusterConf.Config.<cluster>.AIConf.ModelProtocols = ["anthropic"]
            │
            ▼
BFE doSingleAIForward
    └── DetectAuthStyle(req) == "anthropic"
    └── clusterSupportsAuthStyle(AIConf.ModelProtocols, "anthropic")
    └── 匹配通过 → 注入 x-api-key 与 anthropic-version
```

## 5. 涉及模块

| 模块 | 职责 |
|------|------|
| `model/iprovider` | Provider CRUD、模型发现；`model_protocols` 枚举与解析器定义。 |
| `model/icluster_conf` | Cluster 导出；在 `newAIConf` 中将 provider `model_protocols` 透传为 `AIConf.ModelProtocols`。 |
| `model/imodel_price` | 提供 `AIConf.ModelTable`，不限制 `provider` 取值，支持 `anthropic` 定价记录。 |
| `endpoints/innerapi_v1/server_data` | 导出 `server_data_conf`，响应中 `AIConf` 包含 `ModelProtocols`。 |

## 6. 与 BFE 数据面的协作

BFE 侧需要完成以下配合：

1. `bfe_basic/request_ai_basic.go`：识别请求协议风格（`openai` / `anthropic`），写入 `AiBasicInfo.AuthStyle`。
2. `bfe_modules/mod_ai_token_auth/mod_ai_token_auth.go`：按 `AuthStyle` 注入 `Authorization: Bearer` 或 `x-api-key`。
3. `bfe_server/reverseproxy.go`：`doSingleAIForward` 中根据 `AuthStyle` 注入 `anthropic-version`，并执行协议风格匹配校验。
4. `bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`：`AIConf` 新增 `ModelProtocols []string`。
5. `bfe_modules/mod_access_pb3/request_log.go`：访问日志增加 `ai_protocol` 字段。

详见上游分析报告：`document-ai-gateway/迭代系统设计/v0.5/claude协议支持/claude协议支持分析报告.md`。

## 7. 相关文档

- 变更摘要：`design-docs/modifications/2026-08-23-claude-protocol-support/change-summary.md`
- 详细设计：`design-docs/modifications/2026-08-23-claude-protocol-support/design-changes.md`
- 接口变更：`design-docs/modifications/2026-08-23-claude-protocol-support/api-changes.md`
- InnerAPI 接口定义：`design-docs/api-define/InnerAPI接口定义/server-data-conf.md`
