# Claude 协议转发支持：API 接口变更说明

## 1. 变更范围

| 接口类型 | 变更内容 |
|----------|----------|
| OpenAPI | 无变更。`/providers`、`/clusters` 等接口的字段契约保持不变，`model_protocols` 字段已提前支持 `anthropic` 枚举。 |
| InnerAPI | `/configs/tls_conf/server_data_conf` 响应中，cluster 的 `AIConf` 新增 `model_protocols` 字段。 |

---

## 2. InnerAPI 变更：`/configs/tls_conf/server_data_conf`

### 2.1 变更原因

BFE 数据面需要在转发前判断“请求协议风格（`openai` / `anthropic`）是否与目标集群 provider 支持的协议一致”。该能力依赖 `ai-gateway-api` 在导出 BFE 配置时，把 provider 的 `model_protocols` 透传到 cluster 的 `AIConf` 中。

### 2.2 响应结构变化

cluster 配置中的 `AIConf` 对象新增字段：

```json
{
  "Clusters": {
    "<cluster_name>": {
      "AIConf": {
        "Type": 0,
        "ModelMapping": { ... },
        "Provider": "my-anthropic",
        "Keys": [ ... ],
        "KeyPolicy": { ... },
        "ModelTable": { ... },
        "MatchPrefix": "",
        "StripPrefix": false,
        "ModelProtocols": ["anthropic"]
      }
    }
  }
}
```

### 2.3 新增字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `AIConf.ModelProtocols` | `[]string` | 否 | 该集群所属 provider 支持的模型访问协议。来源为 `/providers` 的 `model_protocols`。枚举值如 `openai`、`anthropic`。 |

### 2.4 取值规则

| provider `model_protocols` | 导出 `AIConf.ModelProtocols` |
|----------------------------|------------------------------|
| `["openai"]` | `["openai"]` |
| `["anthropic"]` | `["anthropic"]` |
| `["openai", "anthropic"]` | `["openai", "anthropic"]` |
| 未设置 / 为空 | `[]`（BFE 侧兜底为仅支持 `openai`） |

### 2.5 向后兼容

- 该字段为新增字段，BFE 侧应能安全忽略（若 BFE 版本较旧）。
- 新 BFE 版本读取到空数组时，按仅支持 `openai` 处理，保持对旧 provider 的兼容。

---

## 3. OpenAPI 无变更说明

### 3.1 `/providers`

`/providers` 数据模型中的 `model_protocols` 字段已存在，枚举已包含 `anthropic`，本期无需修改接口契约。

### 3.2 `/clusters`

`/clusters` 的 `llm_config` 结构不变，仍通过 `llm_config.provider` 引用 provider。协议能力由被引用的 provider `model_protocols` 表达。

---

## 4. 依赖的 BFE 侧变更

BFE 侧需在 `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go` 的 `AIConf` 结构中新增 `ModelProtocols []string` 字段，并在 `doSingleAIForward` 中根据该字段做协议匹配校验。详见上游分析报告 `document-ai-gateway/迭代系统设计/v0.5/claude协议支持/claude协议支持分析报告.md`。
