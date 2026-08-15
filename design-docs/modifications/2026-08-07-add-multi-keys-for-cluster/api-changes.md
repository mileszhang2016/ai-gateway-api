# Cluster 多 API-Key 支持：OpenAPI 接口变更说明

## 1. 变更概述

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-07 |
| 影响模块 | `/clusters` |
| 目标文档 | `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/clusters.md` |

本次变更为 `/clusters` 接口的 `llm_config` 数据模型引入**多 API-Key 支持**：

- 删除原有的单 `key` 字段；
- 新增 `keys` 数组，支持为一个 cluster 配置多个 API-Key；
- 新增 `key_policy` 对象，用于声明 Key 选择策略、重试与退避参数；
- 运行时按权重在多个 Key 之间分配流量，并在单个 Key 异常时自动切换。

> 说明：`keys` 允许为空数组，表示该 cluster 不配置 API-Key（例如认证由其他机制处理）。

---

## 2. 数据模型变更

### 2.1 `llm_config` 字段变更

| 字段 | 变更类型 | 说明 |
|------|----------|------|
| `key` | **删除** | 原单 API-Key 字段，由 `keys` 数组替代 |
| `keys` | **新增** | API-Key 列表，详见 [2.2](#22-keys-元素结构) |
| `key_policy` | **新增** | Key 路由策略，详见 [2.3](#23-key_policy-结构) |

**变更前：**

```json
{
  "llm_config": {
    "model_endpoint": { ... },
    "models": ["deepseek-chat"],
    "model_mappings": [ ... ],
    "key": "sk-xxxxxxxxxxxx",
    "provider_type": "deepseek"
  }
}
```

**变更后：**

```json
{
  "llm_config": {
    "model_endpoint": { ... },
    "models": ["deepseek-chat", "deepseek-coder"],
    "model_mappings": [ ... ],
    "keys": [
      {
        "name": "key-primary",
        "key": "sk-xxxxxxxxxxxx",
        "weight": 70
      },
      {
        "name": "key-secondary",
        "key": "sk-yyyyyyyyyyyy",
        "weight": 30
      }
    ],
    "key_policy": {
      "strategy": "weighted_random",
      "max_retries": 3,
      "retry_backoff_initial": 500,
      "retry_backoff_max": 5000
    },
    "provider_type": "deepseek"
  }
}
```

### 2.2 `keys` 元素结构

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| `name` | string | Y | Key 名称/标识，用于日志、监控、运维识别 | 必填；长度 1-128 字符；同一 `keys` 数组内唯一 |
| `key` | string | Y | 实际用于后端认证的密钥 | 必填；非空；长度 1-512 字符 |
| `weight` | int | Y | 权重，用于加权随机选择 | 必填；取值范围 `[0,100]`；`0` 表示该 Key 不接收流量（等效于禁用） |

### 2.3 `key_policy` 结构

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| `strategy` | string | N | Key 选择策略 | 非必填；默认 `weighted_random`；本版仅支持 `weighted_random` |
| `max_retries` | int | N | 当前请求内的总额外重试次数（不是单个 Key 的重试次数） | 非必填；默认 `0`；须为 `>=0` 的整数 |
| `retry_backoff_initial` | int (ms) | N | 首次重试的退避时间，单位毫秒 | 非必填；默认 `500`；须为 `>=0` 的整数 |
| `retry_backoff_max` | int (ms) | N | 退避时间上限，单位毫秒 | 非必填；默认 `5000`；须为 `>=0` 的整数，且 `>= retry_backoff_initial` |

---

## 3. 校验规则

在原有 `llm_config` 校验基础上，增加/调整如下规则：

1. `keys` 非必填，默认值为空数组 `[]`；
2. 若 `keys` 非空：
   - 每个元素 `key` 必填且非空；
   - 每个元素 `weight` ∈ `[0,100]`；
   - 所有 Key 的 `weight` 之和必须等于 `100`；
   - 所有 `name` 在同一 `keys` 数组内唯一；
3. `key_policy` 若传入：
   - `strategy` 仅允许 `weighted_random`；
   - `max_retries` 须为 `>=0` 的整数；
   - `retry_backoff_initial`、`retry_backoff_max` 须为 `>=0` 的整数，且 `retry_backoff_max >= retry_backoff_initial`；
4. `model_endpoint.headers` 中的 `${API_KEY}` 占位符：
   - 当 `keys` 非空时，由当前选中的 Key 替换；
   - 当 `keys` 为空时，**不允许出现 `${API_KEY}` 占位符**，否则在校验阶段返回 `422`；
   - 运行时可按“不发送包含该占位符的 header”作为防御性兜底。

---

## 4. 接口级变更

| 接口 | 方法 | 变更说明 |
|------|------|----------|
| `/clusters` | POST | 请求体 `llm_config` 支持 `keys` 数组和 `key_policy`；删除旧字段 `key`；返回数据包含新结构 |
| `/clusters` | GET | 返回数据中 `llm_config` 以 `keys` 形式展示 |
| `/clusters/{cluster_name}` | GET | 同上 |
| `/clusters/{cluster_name}` | PATCH | 支持部分更新 `llm_config.keys`；`keys` 数组为**全量替换**（与 `instance_pool` 行为一致） |
| `/clusters/{cluster_name}` | DELETE | 无影响 |

> **关于 PATCH 的说明**：`llm_config.keys` 作为数组，按**全量替换**处理，即调用方需传入完整的最新 Key 列表。这与当前 `instance_pool`、`model_mappings` 的 PATCH 语义保持一致，避免数组元素增量合并带来的歧义。

---

## 5. InnerAPI 接口变更

### 5.1 影响文档

| 文档 | 变更说明 |
|------|----------|
| `design-docs/api-define/InnerAPI接口定义/server-data-conf.md` | 更新 `AIConf` 字段示例，删除旧 `Key` 字段，新增 `Keys` 数组与 `KeyPolicy` 对象 |

### 5.2 `AIConf` 结构变更

InnerAPI 导出的 `server_data_conf` 中，`AIConf` 由单 Key 改为多 Key：

**变更前：**

```json
{
    "AIConf": {
        "Type": 0,
        "ModelMapping": {
            "gpt-4": "deepseek-chat"
        },
        "Key": "sk-xxxxxxxxxxxx"
    }
}
```

**变更后：**

```json
{
    "AIConf": {
        "Type": 0,
        "ModelMapping": {
            "gpt-4": "deepseek-chat"
        },
        "Keys": [
            {
                "Name": "key-primary",
                "Key": "sk-aaaaaaaaaaaa",
                "Weight": 70
            },
            {
                "Name": "key-secondary",
                "Key": "sk-bbbbbbbbbbbb",
                "Weight": 30
            }
        ],
        "KeyPolicy": {
            "Strategy": "weighted_random",
            "MaxRetries": 3,
            "RetryBackoffInitial": 500,
            "RetryBackoffMax": 5000
        }
    }
}
```

### 5.3 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `Keys` | array | API-Key 列表；为空数组时表示该 cluster 不配置 API-Key |
| `Keys[].Name` | string | Key 名称/标识（必填） |
| `Keys[].Key` | string | API-Key 值 |
| `Keys[].Weight` | int | 权重，范围 `[0,100]` |
| `KeyPolicy` | object | Key 路由策略 |
| `KeyPolicy.Strategy` | string | 本版仅支持 `weighted_random` |
| `KeyPolicy.MaxRetries` | int | 请求内总额外重试次数 |
| `KeyPolicy.RetryBackoffInitial` | int | 初始退避时间，单位毫秒 |
| `KeyPolicy.RetryBackoffMax` | int | 最大退避时间，单位毫秒 |

---

## 6. 运行时行为补充

### 6.1 整体策略

Key 失败后的降级/恢复策略参考 [Bifrost](https://github.com/maximhq/bifrost) 开源版的 **per-request key rotation**：

- 状态跟踪仅在**单次请求内部**生效，不持久化到 cluster 级别；
- 429 等临时错误标记的 Key 在本次请求内可恢复；
- 401/402/403 等认证错误标记的 Key 在本次请求内永久失效；
- 本版**暂不引入**持久化的熔断/降级机制（如连续失败 N 次后冷却 M 秒）。

### 6.2 单次请求内处理流程

1. 若 `keys` 为空，跳过 Key 选择，`model_endpoint.headers` 中的 `${API_KEY}` 不替换；
2. 若 `keys` 非空：
   - **候选 Key 过滤**：过滤出 `weight > 0` 的 Key（`weight = 0` 等效于禁用，不参与路由）；
   - **Key 选择**：按 `weight` 做加权随机选择，将选中的 `key` 填充到 `model_endpoint.headers` 的 `${API_KEY}` 占位符。

### 6.3 Key 失败后降级/恢复

为当前请求维护两个集合：

| 集合 | 触发条件 | 作用范围 | 恢复方式 |
|------|----------|----------|----------|
| `used_set` | 429 Too Many Requests、5xx/网络错误 | 单次请求 | 当所有候选 Key 都进入 `used_set` 时，清空 `used_set` 开始新一轮加权选择 |
| `dead_set` | 401 Unauthorized / 403 Forbidden / 402 Payment Required | 单次请求 | 不恢复，仅作用于当前请求 |

在 `max_retries` 总重试预算内循环处理：

| 失败类型 | 状态码 | 处理方式 |
|----------|--------|----------|
| **速率限制** | 429 | 将该 Key 加入 `used_set`，轮换到下一个候选 Key，应用指数退避 |
| **认证/计费失败** | 401 / 403 / 402 | 将该 Key 加入 `dead_set`，立即轮换到下一个候选 Key，**不应用退避** |
| **服务端/网络错误** | 5xx / 网络错误 / DNS / 连接失败 | 视为 transient server failure，复用同一个 Key，应用指数退避后重试 |

**Key 池耗尽处理**：

- 若所有候选 Key 都进入 `dead_set`，停止重试，返回 `502 upstream_credentials_exhausted`，表明上游凭证全部失效；
- 若所有候选 Key 都进入 `used_set`（仅 429，无永久失效），清空 `used_set` 开始新一轮加权选择，继续尝试。

### 6.4 退避参数与 cluster 级回退

- **退避参数**：默认 `max_retries=0`、`retry_backoff_initial=500ms`、`retry_backoff_max=5000ms`；
- **cluster 级别重试/回退**：当该 cluster 的 Key 池全部耗尽或达到最大重试次数后，再执行 cluster 级别的重试/回退逻辑（如切换至 fallback cluster）。

---

## 7. 完整示例

### 7.1 创建集群（多 Key 模式）

```json
{
  "name": "deepseek-cluster",
  "description": "多 Key 示例集群",
  "instance_pool": [
    {"name": "backend-1", "addr": "10.0.0.1", "weight": 50, "port": 8080},
    {"name": "backend-2", "addr": "10.0.0.2", "weight": 50, "port": 8080}
  ],
  "llm_config": {
    "model_endpoint": {
      "schema": "https",
      "uri": "/v1/models",
      "headers": {
        "Authorization": "Bearer ${API_KEY}"
      }
    },
    "models": ["deepseek-chat", "deepseek-coder"],
    "model_mappings": [
      {"source_model": "gpt-4", "target_model": "deepseek-chat"}
    ],
    "keys": [
      {
        "name": "key-prod-01",
        "key": "sk-aaaaaaaaaaaa",
        "weight": 70
      },
      {
        "name": "key-prod-02",
        "key": "sk-bbbbbbbbbbbb",
        "weight": 30
      }
    ],
    "key_policy": {
      "strategy": "weighted_random",
      "max_retries": 3,
      "retry_backoff_initial": 500,
      "retry_backoff_max": 5000
    },
    "provider_type": "deepseek"
  }
}
```

### 7.2 更新集群（替换 Key 列表）

```json
{
  "llm_config": {
    "models": ["deepseek-chat", "deepseek-coder"],
    "keys": [
      {
        "name": "key-prod-01",
        "key": "sk-aaaaaaaaaaaa",
        "weight": 50
      },
      {
        "name": "key-prod-02",
        "key": "sk-bbbbbbbbbbbb",
        "weight": 50
      }
    ],
    "key_policy": {
      "strategy": "weighted_random",
      "max_retries": 3,
      "retry_backoff_initial": 500,
      "retry_backoff_max": 5000
    }
  }
}
```

### 7.3 创建集群（不配置 API-Key）

```json
{
  "name": "internal-cluster",
  "description": "内部集群，无需 API-Key",
  "instance_pool": [
    {"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}
  ],
  "llm_config": {
    "model_endpoint": {
      "schema": "http",
      "uri": "/v1/models"
    },
    "models": ["deepseek-chat"],
    "model_mappings": [],
    "keys": [],
    "key_policy": {
      "strategy": "weighted_random",
      "max_retries": 0,
      "retry_backoff_initial": 500,
      "retry_backoff_max": 5000
    },
    "provider_type": "deepseek"
  }
}
```

---

## 8. 对下游/实现的影响

| 下游/实现 | 影响说明 |
|-----------|----------|
| **OpenAPI 文档** | 需更新 `clusters.md` 中 `llm_config` 的字段定义、示例与合法性条件 |
| **模型层** | `icluster_conf.LLMConfig` 中删除 `Key *string`，新增 `Keys []APIKey`、`KeyPolicy *KeyPolicy` |
| **校验层** | 扩展 `validate.LLMConfig`：允许 `keys` 为空数组；非空时校验元素合法性、`weight` 总和为 100 |
| **控制层** | 更新 `normalizeLLMConfig`：不再处理旧 `key` 字段，未传 `keys` 时默认空数组 |
| **数据库** | 当前 `clusters.llm_config` 为 JSON 字符串字段，新增 `keys` 可直接序列化存储，**无需新增表字段** |
| **BFE 转发层** | 需从 cluster 配置读取 `keys`，实现加权随机选择、请求级 Key 轮换与退避；`model_endpoint.headers` 中 `${API_KEY}` 替换逻辑需支持运行时动态 Key |
| **conf-agent** | 若 `clusters.llm_config` 以 JSON 字符串形式下发，需确保 `keys` 字段能被正确解析；关注敏感字段 `keys[].key` 的加密/脱敏存储与传输 |
| **集成测试** | 需补充多 Key 创建、更新、查询、权重校验、`keys` 为空等场景 |

---

## 9. 已确认事项

1. **Key 失败后的降级/恢复策略**：参考 Bifrost 开源版的 per-request key rotation（请求内轮换，不退避/不持久化），暂时不引入持久化的熔断/降级机制。
2. **`keys` 为空时的 `${API_KEY}` 占位符处理**：在 OpenAPI 校验阶段即拒绝该配置（`keys` 为空但 header 值含 `${API_KEY}` 返回 `422`），运行时不做替换或直接移除包含占位符的 header 作为兜底。

---

*文档生成日期：2026-08-07*
