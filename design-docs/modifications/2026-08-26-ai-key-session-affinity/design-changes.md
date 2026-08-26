# 同 cluster 多 API-Key 会话级亲和性——设计变更说明

## 1. 概念定义

| 概念 | 定义 |
|------|------|
| **ClientKeyId** | 用于标识一次会话的 key。BFE 侧从请求头 `X-Client-Key-Id` 中读取；如未提供，则回退到加权随机选择，不保证亲和性。 |
| **Session Affinity** | 会话级 API-Key 亲和性：同一个 `ClientKeyId` 在同一个 cluster 下固定使用同一个 API-Key，直到绑定过期或该 Key 被惩罚。 |
| **Penalty** | 当某个 Key 近期返回 429/401/403 时，BFE 会跳过该 Key 并重新选择，避免持续命中失效 Key。 |
| **Redis 绑定** | BFE 使用 Redis 存储 `client_key_id` 到 `api_key_name` 的映射，key 格式为 `<prefix>:<cluster_name>:<client_key_id>`，默认空闲超时 600 秒；命中绑定后会刷新 TTL。 |

## 2. 配置归属与下发链路

### 2.1 控制面 → 数据面转换

```
/cluster my-cluster
    └── llm_config
        ├── keys: [key-primary, key-secondary]
        ├── key_policy: { strategy, max_retries, ... }
        └── key_affinity
            ├── enabled: true
            ├── ttl: 600
            ├── redis_prefix: "bfe:ai:key_affinity"
            └── penalty_enable: true

BFE cluster_conf.data
    └── AIConf.KeyPolicy
        ├── Strategy
        ├── MaxRetries
        ├── SessionAffinity              // 来自 key_affinity.enabled
        ├── SessionAffinityTTL           // 来自 key_affinity.ttl
        ├── SessionAffinityRedisPrefix   // 来自 key_affinity.redis_prefix
        └── SessionAffinityPenaltyEnable // 来自 key_affinity.penalty_enable
```

### 2.2 单 Key 优化

BFE 数据面已实现：当 cluster 只绑定一个 API-Key 时，不访问 Redis，直接使用该 Key。控制面不需要额外处理，只需保证 `keys` 数组长度为 1 时语义正确。

## 3. 数据模型改动

### 3.1 `ai-gateway-api` cluster 数据模型

在 `LLMConfig` 结构体中新增 `KeyAffinity *KeyAffinity` 字段：

```go
type KeyAffinity struct {
    Enabled       *bool   `json:"enabled"`        // 默认 false
    TTL           *int    `json:"ttl"`            // 绑定空闲超时时间，默认 600，单位秒
    RedisPrefix   *string `json:"redis_prefix"`   // Redis key 前缀，默认 "bfe:ai:key_affinity"
    PenaltyEnable *bool   `json:"penalty_enable"` // 是否开启 Key 惩罚，默认 true
}

type LLMConfig struct {
    Models        []string        `json:"models"`
    ModelMappings []*Mapping      `json:"model_mappings"`
    Keys          []ClusterKeyRef `json:"keys"`
    KeyPolicy     *KeyPolicy      `json:"key_policy"`
    KeyAffinity   *KeyAffinity    `json:"key_affinity"` // 新增
    Provider      *string         `json:"provider"`
    MatchPrefix   *string         `json:"match_prefix"`
    StripPrefix   *bool           `json:"strip_prefix"`
}
```

### 3.2 数据库模型

无需修改数据库表结构。`clusters.llm_config` 字段为 JSON 文本，直接扩展即可。

### 3.3 BFE 侧对应结构

BFE `AIKeyPolicy` 已扩展为：

```go
type AIKeyPolicy struct {
    Strategy                     string
    MaxRetries                   int
    RetryBackoffInitial          int
    RetryBackoffMax              int
    SessionAffinity              bool
    SessionAffinityTTL           int
    SessionAffinityRedisPrefix   string
    SessionAffinityPenaltyEnable bool
}
```

## 4. 配置导出逻辑改动

在 `model/icluster_conf/cluster.go` 的 `newAIConf` 函数中，将 `llmConfig.KeyAffinity` 映射到 `cluster_conf.AIKeyPolicy`：

```go
aiConf.KeyPolicy = &cluster_conf.AIKeyPolicy{
    Strategy:                     derefString(llmConfig.KeyPolicy.Strategy, "weighted_random"),
    MaxRetries:                   derefInt(llmConfig.KeyPolicy.MaxRetries, 0),
    RetryBackoffInitial:          derefInt(llmConfig.KeyPolicy.RetryBackoffInitial, 500),
    RetryBackoffMax:              derefInt(llmConfig.KeyPolicy.RetryBackoffMax, 5000),
    SessionAffinity:              derefBool(llmConfig.KeyAffinity.Enabled, false),
    SessionAffinityTTL:           derefInt(llmConfig.KeyAffinity.TTL, 600),
    SessionAffinityRedisPrefix:   derefString(llmConfig.KeyAffinity.RedisPrefix, "bfe:ai:key_affinity"),
    SessionAffinityPenaltyEnable: derefBool(llmConfig.KeyAffinity.PenaltyEnable, true),
}
```

注意：当 `llmConfig.KeyAffinity == nil` 时，所有亲和性字段使用 BFE 默认值，保持向后兼容。

## 5. 校验规则

新增 `KeyAffinity` 校验：

| 字段 | 校验规则 |
|------|----------|
| `enabled` | 可选；若传入，必须为 bool |
| `ttl` | 可选；若传入，必须为 >0 整数 |
| `redis_prefix` | 可选；若传入，必须非空字符串 |
| `penalty_enable` | 可选；若传入，必须为 bool |

当 `enabled=false` 时，`ttl`、`redis_prefix`、`penalty_enable` 仍可传入，但 BFE 不会使用。

## 6. 边界情况

- **未传 `key_affinity`**：BFE 使用默认值，亲和性关闭。
- **单 Key 场景**：BFE 自动跳过 Redis，控制面无需特殊处理。
- **BFE 未配置 Redis**：若 `enabled=true` 但 BFE 无可用 Redis，亲和性无法生效；该问题由 BFE 数据面处理（降级为加权随机）。
- **Key 被惩罚后重新绑定**：运行时为纯数据面行为，控制面不感知。
