# Provider 与 Cluster 概念分离——API 接口变更说明

## 1. 新增 `/providers` 资源

### 1.1 接口清单

| 方法 | 端点 | 含义 |
|------|------|------|
| POST | `/providers` | 创建 provider |
| GET | `/providers` | 分页/过滤查询 provider 列表 |
| GET | `/providers/{provider_name}` | 查询单个 provider |
| PATCH | `/providers/{provider_name}` | 部分更新 provider |
| DELETE | `/providers/{provider_name}` | 删除 provider |
| POST | `/providers/{provider_name}/discover-models` | 触发模型发现 |

### 1.2 数据模型

```json
{
  "name": "deepseek",
  "description": "DeepSeek 官方 API",
  "model_endpoint": {
    "schema": "https",
    "uri": "/v1/models"
  },
  "models": ["deepseek-chat", "deepseek-coder"],
  "keys": [
    {"name": "key-primary", "key": "sk-aaaaaaaaaaaa"},
    {"name": "key-secondary", "key": "sk-bbbbbbbbbbbb"}
  ],
  "instance_pool": [
    {
      "name": "backend-1",
      "addr": "api.deepseek.com",
      "weight": 100,
      "port": 443
    }
  ],
  "model_protocols": ["openai"],
  "create_time": 1716883200,
  "update_time": 1716883200
}
```

### 1.3 关键字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | Provider 唯一标识，全局唯一 |
| `description` | string | 否 | 描述，长度 0-256 |
| `model_endpoint` | object | 否 | 模型发现端点，默认 `{schema: "https", uri: "/v1/models"}` |
| `models` | []string | 否 | 支持的模型列表 |
| `keys` | []ProviderKey | 否 | API Key 明文列表 |
| `instance_pool` | []Instance | 是 | 后端实例池，至少 1 个元素 |
| `model_protocols` | []string | 是 | 支持的模型访问协议，至少 1 个元素 |

**`model_protocols` 首期枚举**：`openai`、`anthropic`、`gemini`。

### 1.4 删除约束

- 若存在 `/clusters` 引用该 provider，返回 `409 Conflict`。
- 若存在 `/model-prices` 记录引用该 provider，返回 `409 Conflict`。

### 1.5 模型发现 `/providers/{provider_name}/discover-models`

- 调用 `model_endpoint`，按 `model_protocols` 选择认证头风格。
- 取 `keys` 中第一个 key 作为认证凭证。
- 根据协议选择响应解析器，提取模型名列表并回填到 `models`。
- 模型发现为辅助能力，失败不影响 provider 本身。

## 2. `/clusters` 改造

### 2.1 接口 URL 与 HTTP Method 保持不变

- `POST /clusters`
- `GET /clusters`
- `GET /clusters/{cluster_name}`
- `PATCH /clusters/{cluster_name}`
- `DELETE /clusters/{cluster_name}`

### 2.2 请求/响应体变化

**移除字段**：

- 顶层 `instance_pool`
- `llm_config.model_endpoint`
- `llm_config.provider_type`

**改造字段**：

- `llm_config.provider`：变为必填，必须引用已存在的 provider。
- `llm_config.keys`：元素结构变为 `{name, weight}`，只保留 key 名称和权重，不再返回明文。
- `llm_config.models`：含义变为“cluster 允许转发给该 provider 的模型”，须为 provider `models` 的子集。

**保留字段**：

- `model_mappings`、`key_policy`、`match_prefix`、`strip_prefix` 行为不变。
- `basic`、`sticky_sessions`、`passive_health_check` 等 BFE 集群参数不变。

### 2.3 新 `llm_config` 示例

```json
{
  "llm_config": {
    "models": ["deepseek-chat", "deepseek-coder"],
    "model_mappings": [
      {"source_model": "gpt-4", "target_model": "deepseek-chat"}
    ],
    "keys": [
      {"name": "key-primary", "weight": 70},
      {"name": "key-secondary", "weight": 30}
    ],
    "key_policy": {
      "strategy": "weighted_random",
      "max_retries": 3,
      "retry_backoff_initial": 500,
      "retry_backoff_max": 5000
    },
    "provider": "deepseek",
    "match_prefix": "deepseek/",
    "strip_prefix": true
  }
}
```

### 2.4 关键校验规则

1. `llm_config.provider` 必填且引用的 provider 必须存在。
2. `llm_config.models` 中每个模型必须存在于 provider 的 `models` 中。
3. `llm_config.keys` 中每个 `name` 必须存在于 provider 的 `keys` 中。
4. 若 `llm_config.keys` 非空，所有 `weight` 之和必须等于 100。

## 3. `/model-prices` 改造

### 3.1 变更点

- `provider` 字段含义收紧：从“Provider / Cluster 标识”变为“必须引用 `/providers` 中已存在的 provider”。
- 新增/导入/更新 `model-prices` 时，须校验 `provider` 在 provider 表中存在。
- 删除 provider 前，须校验无 `model-prices` 记录引用。

### 3.2 字段说明（变更部分）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `provider` | string | 是 | 必须引用 `/providers` 中已存在的 provider |

### 3.3 对 `model-list.yaml` 的影响

- 导入时 `provider` 字段同样必须对应已存在的 provider。
- 不存在的 provider 应作为 error 返回，并跳过该条记录。

## 4. 配置顺序

```
/providers → /model-prices → /clusters → 路由规则
```

| 步骤 | 资源 | 原因 |
|------|------|------|
| 1 | `/providers` | 定义下游接入能力 |
| 2 | `/model-prices` | 价格记录依赖 provider |
| 3 | `/clusters` | cluster 依赖 provider 的模型和 key |
| 4 | 路由规则 | 路由规则依赖 cluster |
