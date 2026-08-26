# 同 cluster 多 API-Key 会话级亲和性——接口变更

## 1. 变更范围

影响以下 OpenAPI 与 InnerAPI 接口：

- `POST /clusters`
- `PUT /clusters/{name}`
- `GET /clusters`
- `GET /clusters/{name}`
- `GET /configs/tls_conf/server_data_conf`（InnerAPI）

## 2. OpenAPI `/clusters` 变更

### 2.1 请求/响应数据模型扩展

在 `llm_config` 中新增 `key_affinity` 字段：

```json
{
    "llm_config": {
        "models": ["deepseek-chat"],
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
        "key_affinity": {
            "enabled": true,
            "ttl": 600,
            "redis_prefix": "bfe:ai:key_affinity",
            "penalty_enable": true
        },
        "provider": "deepseek"
    }
}
```

### 2.2 新增字段说明

**表：`llm_config.key_affinity`**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 | 合法性条件 |
| - | - | - | - | - | - |
| `enabled` | bool | 是否开启会话级 Key 亲和性 | N | 默认 `false`；为 `true` 时开启 | 非必填；必须为 bool |
| `ttl` | int | 绑定空闲超时时间 | N | 单位秒，默认 `600`；命中绑定后 BFE 会刷新 TTL，持续请求则绑定保持 | 非必填；若传入，须为 >0 整数 |
| `redis_prefix` | string | Redis key 前缀 | N | 默认 `"bfe:ai:key_affinity"` | 非必填；若传入，必须非空 |
| `penalty_enable` | bool | 是否开启 Key 惩罚 | N | 默认 `true`；为 `true` 时，近期返回 429/401/403 的 Key 会被跳过 | 非必填；必须为 bool |

### 2.3 约束更新

在 `clusters.md` 接口文档的“约束”章节中追加：

- `llm_config.key_affinity` 非必填；若传入：
  - `enabled` 必须为 bool；
  - `ttl` 须为 >0 的整数；
  - `redis_prefix` 若传入须非空；
  - `penalty_enable` 必须为 bool。

## 3. InnerAPI `/configs/tls_conf/server_data_conf` 变更

### 3.1 导出字段扩展

InnerAPI 返回的 `ClusterConf.Config.<cluster>.AIConf.KeyPolicy` 中新增以下字段：

```json
{
    "KeyPolicy": {
        "Strategy": "weighted_random",
        "MaxRetries": 3,
        "RetryBackoffInitial": 500,
        "RetryBackoffMax": 5000,
        "SessionAffinity": true,
        "SessionAffinityTTL": 600,
        "SessionAffinityRedisPrefix": "bfe:ai:key_affinity",
        "SessionAffinityPenaltyEnable": true
    }
}
```

### 3.2 映射规则

| InnerAPI 返回字段 | 来源 |
| - | - |
| `SessionAffinity` | `llm_config.key_affinity.enabled`，默认 `false` |
| `SessionAffinityTTL` | `llm_config.key_affinity.ttl`，默认 `600` |
| `SessionAffinityRedisPrefix` | `llm_config.key_affinity.redis_prefix`，默认 `"bfe:ai:key_affinity"` |
| `SessionAffinityPenaltyEnable` | `llm_config.key_affinity.penalty_enable`，默认 `true` |

当用户未配置 `key_affinity` 时，上述字段使用默认值下发，保持 BFE 侧向后兼容。

## 4. 测试影响

- OpenAPI 创建/更新 cluster 的集成测试需要补充 `key_affinity` 用例。
- InnerAPI 导出测试需要校验 `AIConf.KeyPolicy` 中的 `SessionAffinity*` 字段。
- 向后兼容测试：未传 `key_affinity` 时，导出结果中包含默认值字段。
