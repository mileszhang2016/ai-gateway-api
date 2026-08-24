# Claude 协议转发支持（ai-gateway-api 控制面配合）变更摘要

## 1. 背景

BFE 数据面正在扩展对 Anthropic Claude Messages API 的转发能力。与 OpenAI 兼容协议相比，Claude 协议在认证头（`x-api-key`）、版本头（`anthropic-version`）和 Usage 字段（`input_tokens` / `output_tokens`）上存在差异。

在 cluster/provider 分离后的新模型中，provider 通过 `model_protocols` 字段声明支持的协议（如 `openai`、`anthropic`）。BFE 需要依据该字段判断请求协议风格是否被目标集群支持，并在不匹配时直接拒绝。因此，`ai-gateway-api` 控制面需要把 provider 的 `model_protocols` 透传到导出的 BFE `AIConf` 中，并清理已废弃的 `conf/ai` 配置目录。

## 2. 目标

- 让 BFE 在导出配置中拿到每个 cluster 对应 provider 支持的协议列表（`AIConf.ModelProtocols`）。
- 让 `/providers/{name}/discover-models` 在 `model_protocols=anthropic` 时能正确解析 Anthropic `/v1/models` 响应。
- 清理已废弃的 `ai-gateway-api/conf/ai` 目录，避免遗留配置与代码产生歧义。
- 保持 `/providers`、`/clusters` 等 OpenAPI 接口契约不变（`model_protocols` 字段已提前支持 `anthropic`）。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `model/icluster_conf/cluster.go`、`model/iprovider/provider.go`、测试 schema、废弃配置目录 `conf/ai` |
| 接口契约 | `/providers`、`/clusters` 请求/响应字段不变；仅内部导出逻辑与模型发现解析器调整 |
| 数据迁移 | 无；`model_protocols` 字段已在 provider 模型中存在 |
| BFE 影响 | 导出的 `AIConf` 新增 `ModelProtocols` 字段，供 BFE `doSingleAIForward` 做协议匹配校验 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| `model_protocols` 已覆盖 `anthropic` | `/providers` 数据模型已包含该枚举，本期直接启用，无需新增接口。 |
| `AIConf` 新增 `ModelProtocols` | 阶段一的最小改动：把 provider 的 `model_protocols` 原样透传给 BFE。 |
| 解析器下沉到代码 | 原 `conf/ai/model_definition.json` 中的 anthropic 解析规则，迁移到代码内的 `modelProtocolParsers`。 |
| 删除 `conf/ai` 目录 | `/model-provider-types` 与 `/tools/get-models-from-provider` 已移除，该目录全部废弃。 |
| 不依赖 provider 名称推断协议 | `provider` 为用户自定义名称，协议能力必须以 `model_protocols` 为准。 |

## 5. 关联文档

- 详细设计：`design-changes.md`
- 接口变更：`api-changes.md`
- 上游分析报告：`document-ai-gateway/迭代系统设计/v0.5/claude协议支持/claude协议支持分析报告.md`
- 相关接口定义：`design-docs/api-define/OpenAPI接口定义/providers.md`
