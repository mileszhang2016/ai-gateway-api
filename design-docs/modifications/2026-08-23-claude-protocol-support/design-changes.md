# Claude 协议转发支持：ai-gateway-api 控制面设计变更说明

> 本文档只描述 `ai-gateway-api` 控制面需要配合 BFE 数据面做的修改。BFE 数据面修改（`AuthStyle` 识别、API Key 注入、`anthropic-version` 注入、Usage 解析、访问日志等）见上游分析报告 `document-ai-gateway/迭代系统设计/v0.5/claude协议支持/claude协议支持分析报告.md`。

## 1. 概述

### 1.1 变更背景

BFE 数据面即将支持 Anthropic Claude Messages API 的原样转发。该协议与 OpenAI 兼容协议存在以下关键差异：

| 项目 | OpenAI / 兼容协议 | Anthropic Claude Messages API |
|------|-------------------|-------------------------------|
| 对话端点 | `POST /v1/chat/completions` | `POST /v1/messages` |
| 认证头 | `Authorization: Bearer <api_key>` | `x-api-key: <api_key>` |
| 版本头 | 无 | 必须 `anthropic-version: 2023-06-01` |
| 响应 usage | `prompt_tokens` / `completion_tokens` / `total_tokens` | `input_tokens` / `output_tokens` |

为了让 BFE 能在转发前判断“请求协议风格是否与目标集群 provider 支持的协议一致”，控制面必须在导出的 BFE 配置中携带 provider 支持的协议列表。

### 1.2 变更目标

1. 在 InnerAPI 导出的 BFE `AIConf` 中新增 `ModelProtocols []string`，内容来自 cluster 所引用 provider 的 `model_protocols`。
2. 让 `/providers/{name}/discover-models` 能按 `model_protocols=anthropic` 正确解析上游 `/v1/models` 响应。
3. 删除已废弃的 `ai-gateway-api/conf/ai` 目录。
4. 同步更新集成测试 schema，确保 `provider_type`、`model_endpoint` 等旧字段不再出现。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/icluster_conf`、`model/iprovider`、集成测试 schema、废弃配置目录 `conf/ai` |
| 变更类型 | 配置导出改造 + 模型发现解析器下沉 + 废弃目录清理 |

---

## 2. 数据模型与配置导出

### 2.1 Provider 的 `model_protocols` 已支持 `anthropic`

根据 `design-docs/api-define/OpenAPI接口定义/providers.md`，`/providers` 数据模型已经包含 `model_protocols` 字段。本期只启用其中两个枚举值：

- `openai`
- `anthropic`

> `gemini` 暂不在本期支持范围内。

用户创建 Claude 集群时，配置顺序为：

1. 先创建 `model_protocols` 包含 `anthropic` 的 provider；
2. 再创建 `llm_config.provider` 引用该 provider 的 cluster。

### 2.2 集群 `LLMConfig` 当前结构

在 cluster/provider 分离设计中，`LLMConfig` 定义在 `ai-gateway-api/model/icluster_conf/cluster.go`：

```go
type LLMConfig struct {
    Models        []string
    ModelMappings []*Mapping
    Keys          []ClusterKeyRef // 只保留 name + weight，引用 provider.keys
    KeyPolicy     *KeyPolicy
    Provider      *string  // 必填，引用 provider name
    MatchPrefix   *string
    StripPrefix   *bool
}
```

该结构本期保持不变。

### 2.3 BFE `AIConf` 新增 `ModelProtocols`

**BFE 侧结构**（`bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`）：

```go
type AIConf struct {
    Type         int
    ModelMapping *map[string]string
    Provider     string
    Keys         []AIKey
    KeyPolicy    *AIKeyPolicy
    ModelTable   *ModelTable
    MatchPrefix  string
    StripPrefix  bool

    // ModelProtocols 表示该集群所属 provider 支持的模型访问协议。
    // 来源：ai-gateway-api 的 provider.model_protocols。
    // 用于 BFE doSingleAIForward 中判断请求协议风格是否被当前集群支持。
    ModelProtocols []string
}
```

> 该字段在 BFE 侧定义，但由 `ai-gateway-api` 在导出时填充。

### 2.4 `newAIConf` 改造

在 `ai-gateway-api/model/icluster_conf/cluster.go` 中，把 cluster 所引用 provider 的 `model_protocols` 透传给 BFE：

```go
func newAIConf(llmConfig *LLMConfig, modelTable *cluster_conf.ModelTable,
               providerKeys []iprovider.ProviderKey,
               providerModelProtocols []string) *cluster_conf.AIConf {
    aiConf := &cluster_conf.AIConf{
        Type:           0,
        ModelMapping:   convertToBFEModelMapping(llmConfig.ModelMappings),
        Keys:           []cluster_conf.AIKey{},
        ModelProtocols: providerModelProtocols, // 新增
    }
    // ... 其余字段不变 ...
    return aiConf
}
```

调用方（`NewBfeClusterConf` 内部）需要先把 provider 查出来，再传入 `newAIConf`：

```go
provider, err := cm.providerStorager.FetchProvider(ctx,
    &iprovider.ProviderFilter{Name: *cluster.LLMConfig.Provider})
if err != nil {
    return err
}

providerKeys := providerKeyTable[provider.Name]
clusterConf.AIConf = newAIConf(
    cluster.LLMConfig,
    modelTable,
    providerKeys,
    provider.ModelProtocols, // 新增
)
```

如果 `NewBfeClusterConf` 一次性处理全部 cluster，也可以预先构建 `providerProtocolTable map[string][]string`，再按 provider name 查找传入 `newAIConf`。

### 2.5 为什么不能用 `AIConf.Provider` 名称推断协议

cluster/provider 分离后，`provider` 是用户自定义名称（如 `my-anthropic`），不再限定为系统内置枚举。因此 BFE 不能通过 `AIConf.Provider == "anthropic"` 这种硬编码判断协议能力，必须以显式的 `ModelProtocols` 为准。

---

## 3. 模型发现解析器下沉

### 3.1 原配置文件中的 anthropic 解析器

原 `ai-gateway-api/conf/ai/model_definition.json` 中的 anthropic 解析器定义如下：

```json
"anthropic": {
    "list_path": "models",
    "id_field": "model_id",
    "name_field": "display_name",
    "created_field": "created_at",
    "type": "array",
    "description": "Anthropic Claude API格式"
}
```

由于 `ai-gateway-api/conf/ai` 目录整体废弃，该解析器不应继续以 JSON 配置文件形式存在，而应把定义下沉到代码中。

### 3.2 代码内解析器定义

在 `model/iprovider/provider.go`（或 discover-models 实现所在文件）中定义：

```go
// ModelParser 供 /providers/{name}/discover-models 使用，
// 用于从上游 /v1/models 响应中提取模型名列表。
type ModelParser struct {
    ListPath   string
    IDField    string
    NameField  string
    Type       string
}

var modelProtocolParsers = map[string]ModelParser{
    "openai": {
        ListPath:   "data",
        IDField:    "id",
        NameField:  "object",
        Type:       "array",
    },
    "anthropic": {
        ListPath:   "models",
        IDField:    "model_id",
        NameField:  "display_name",
        Type:       "array",
    },
}
```

### 3.3 discover-models 逻辑

`/providers/{name}/discover-models` 调用上游 `/v1/models` 后：

1. 读取 provider 的 `model_protocols`；
2. 按顺序选择第一个支持的协议（优先 `openai`，其次 `anthropic`），从 `modelProtocolParsers` 中取解析器；
3. 按 `ListPath`、`IDField` 从响应 JSON 中提取模型名列表；
4. 将提取结果回填到 provider 的 `models` 字段。

> 说明：`created_field` 仅用于展示，模型发现只关心模型名列表，因此代码内解析器可不保留该字段。

---

## 4. 废弃 `ai-gateway-api/conf/ai` 目录

### 4.1 目录现状

随着 `/model-provider-types` 与 `/tools/get-models-from-provider` 被移除、`/providers` 成为独立的资源配置，`ai-gateway-api/conf/ai` 目录下的配置文件全部变为废弃：

| 文件 | 原用途 | 现状 |
|------|--------|------|
| `conf/ai/models.json` | 维护 provider 类型枚举，供 `/model-provider-types` 返回 | 已不需要；`model_protocols` 枚举在 `model/iprovider/provider.go` 的 `ValidModelProtocols` 中维护 |
| `conf/ai/model_definition.json` | 维护不同 provider 的模型列表解析器，供 `/tools/get-models-from-provider` 使用 | 已不需要；模型发现由 `/providers/{name}/discover-models` 按 `model_protocols` 选择解析器 |
| `conf/ai/model_list.json` | 维护内置模型列表 | 已不需要；模型通过 provider 的 `models` 字段或模型发现接口维护 |

### 4.2 清理动作

本次改造应同步删除整个 `ai-gateway-api/conf/ai` 目录，包括：

- `conf/ai/models.json`
- `conf/ai/model_definition.json`
- `conf/ai/model_list.json`

同时检查代码中是否还有读取这些文件的逻辑，一并移除。

---

## 5. 校验与 Schema

### 5.1 已有校验规则

`model/icluster_conf/cluster.go` 的 `validateClusterLLMConfigAgainstProvider` 已经需要校验：

- `llm_config.provider` 必须引用已存在的 provider；
- `llm_config.models` 必须是 provider `models` 的子集；
- `llm_config.keys[].name` 必须存在于 provider `keys` 中。

本期新增 Claude 支持无需额外校验规则。

### 5.2 校验层

`ai-gateway-api/lib/validate/validate.go` 的 `LLMConfig()` 中：

- `provider_type` 字段已不存在，无需校验；
- `provider` 必填且引用已存在 provider 的校验已在模型层完成。

### 5.3 集成测试 Schema

#### OpenAPI schema

`ai-gateway-api/test/integration/tests/schema/openapi/cluster.go` 的 `LLMConfigSchema` 需要：

- 移除 `provider_type`、`model_endpoint`；
- 确保 `provider` 为字符串；
- 确保 `keys` 为 `{name, weight}` 结构。

> 阶段一新增的 `ModelProtocols` 仅影响 BFE `AIConf`，不影响 OpenAPI schema。

#### InnerAPI 导出 schema（受 `model/icluster_conf` 改动影响）

`model/icluster_conf/cluster.go` 中 `newAIConf` 新增 `ModelProtocols` 字段后，InnerAPI 导出的 BFE 配置结构随之变化。集成测试中需要同步更新对应 schema 或断言：

- 在 InnerAPI `/configs/tls_conf/server_data_conf` 的响应 schema 中，允许 cluster 的 `AIConf` 包含 `model_protocols` 字段；
- 增加断言：当 cluster 引用的 provider `model_protocols=["anthropic"]` 时，导出配置中对应 `AIConf.ModelProtocols` 必须等于 `["anthropic"]`；
- 当 provider `model_protocols` 为空或未设置时，断言 `AIConf.ModelProtocols` 为空数组（BFE 侧会据此兜底为仅支持 `openai`）。

### 5.4 Model-Price 校验

`ai-gateway-api/model/imodel_price/validate.go` 的 `ValidateModelPrice` 只要求 `Provider` 非空，不限制取值，因此 `provider=anthropic` 的定价记录可直接导入。

---

## 6. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `ai-gateway-api/model/icluster_conf/cluster.go` | 阶段一：`newAIConf` 将 provider 的 `model_protocols` 写入 BFE `AIConf.ModelProtocols`；调用方需先查询 provider 再传入。 |
| `ai-gateway-api/model/iprovider/provider.go` | `model_protocols` 枚举已包含 `anthropic`，无需修改；若 discover-models 实现位于此处，需新增 `modelProtocolParsers` 与解析逻辑。 |
| `ai-gateway-api/conf/ai/models.json` | 删除 |
| `ai-gateway-api/conf/ai/model_definition.json` | 删除 |
| `ai-gateway-api/conf/ai/model_list.json` | 删除 |
| `ai-gateway-api/test/integration/tests/schema/openapi/cluster.go` | 移除 `provider_type`/`model_endpoint`，确认 `provider` 引用结构。 |
| `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go` | 在 BFE `AIConf` 中新增 `ModelProtocols []string`（BFE 侧修改，此处列作依赖）。 |

---

## 7. 测试计划

### 7.1 单元测试

1. **`newAIConf` 导出测试**
   - 构造 provider `model_protocols=["anthropic"]` 的 cluster；
   - 断言导出的 `AIConf.ModelProtocols` 等于 `["anthropic"]`。

2. **默认兼容测试**
   - 构造 provider `model_protocols=["openai"]` 的 cluster；
   - 断言 `AIConf.ModelProtocols` 等于 `["openai"]`。

3. **多协议 provider 测试**
   - 构造 provider `model_protocols=["openai","anthropic"]` 的 cluster；
   - 断言 `AIConf.ModelProtocols` 原样透传。

### 7.2 集成测试

1. **创建 Claude provider**
   - `POST /providers`，`model_protocols: ["anthropic"]`，成功。

2. **创建引用 Claude provider 的 cluster**
   - `POST /clusters`，`llm_config.provider` 引用上述 provider；
   - `llm_config.keys` 通过 `name` 引用 provider 中的 key；
   - 断言成功。

3. **模型发现**
   - `POST /providers/{name}/discover-models`，使用 Anthropic `/v1/models` 响应 mock；
   - 断言返回的 `models` 列表正确提取自 `models[].model_id` 或 `models[].display_name`。

4. **InnerAPI 导出校验**
   - 调用 InnerAPI `/configs/tls_conf/server_data_conf`；
   - 断言对应 cluster 的 `AIConf.ModelProtocols` 包含 `anthropic`。

5. **Schema 回归**
   - 运行 `test/integration/tests/schema/openapi/cluster.go` 相关测试，确认旧字段不再校验。

---

## 8. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| BFE 侧未同步升级 | 若 BFE 还未支持 `AIConf.ModelProtocols`，导出配置会被忽略 | 本改造与 BFE 改造同步上线；BFE 侧应忽略未知 JSON 字段或做兼容处理 |
| `model_protocols` 为空 | 旧 provider 可能没有 `model_protocols` | BFE 侧兜底：当 `ModelProtocols` 为空时默认只支持 `openai` |
| 解析器字段遗漏 | 代码内解析器与 JSON 配置不完全一致 | 单元测试覆盖 openai 与 anthropic 两种解析路径 |
| 删除 `conf/ai` 影响启动 | 若启动时仍读取该目录会失败 | 同步移除相关加载逻辑并验证启动 |
| provider 名称自定义 | 不能通过 `provider` 名字判断协议 | 协议能力完全由 `model_protocols` 表达，并在导出时透传 |
