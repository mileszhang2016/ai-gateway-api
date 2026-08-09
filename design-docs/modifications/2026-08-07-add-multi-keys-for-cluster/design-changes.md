# Cluster 多 API-Key 支持：系统设计变更说明

## 1. 概述

### 1.1 变更背景

当前 `ai-gateway-api` 的 `/clusters` 接口中，每个集群的 `llm_config` 仅支持配置单个 API-Key（`key` 字段）。BFE 侧 `AIConf` 同样只支持单个 `Key`。该设计在以下场景存在瓶颈：

- 单一 Key 达到 provider 速率上限后，无法继续扩容；
- 无法做 Key 级别的负载均衡、故障转移、灰度切换；
- 不同 Key 绑定不同模型/额度时，无法精细化管理。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-07 |
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `ai-gateway-api/model/icluster_conf`、`ai-gateway-api/endpoints/openapi_v1/product_cluster` |
| 变更类型 | 数据模型扩展 + 转发逻辑改造 + 接口文档更新 |

为 `/clusters` 引入多 API-Key 支持：

1. 一个 cluster 可配置多个 API-Key；
2. 支持按权重在多个 Key 之间分配流量；
3. 单个 Key 异常时可在单次请求内自动切换；
4. 直接以 `keys` 数组替换原有单 `key` 字段，不保留旧字段；
5. `keys` 可为空数组，表示该 cluster 不配置 API-Key。

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| ** fail fast** | 配置不一致（如 `keys` 为空但 header 含 `${API_KEY}`）在校验阶段拒绝，而非运行时才暴露。 |
| **最小侵入** | `clusters.llm_config` 仍为 JSON 字符串，无需新增数据库字段；BFE `AIConf` 扩展字段同时兼容旧单 Key 逻辑移除。 |
| **运行时状态不持久化** | Key 级失败状态仅作用于单次请求，不引入持久化熔断/降级机制，降低系统复杂度。 |
| **控制面与数据面解耦** | ai-gateway-api 负责配置校验与导出；数据面（BFE）负责运行时 Key 选择、轮换与退避。 |

---

## 2. 数据模型设计

### 2.1 ai-gateway-api 侧：`LLMConfig`

**当前结构：**

```go
type LLMConfig struct {
    ModelEndpoint  *ModelEndpoint   `json:"model_endpoint"`
    Models         []string         `json:"models"`
    ModelMappings  []ModelMapping   `json:"model_mappings"`
    Key            *string          `json:"key"`          // 将被删除
    ProviderType   *string          `json:"provider_type"`
}
```

**改造后结构：**

```go
type APIKey struct {
    Name   *string `json:"name"`   // 必填；长度 1-128；同一 keys 数组内唯一
    Key    *string `json:"key"`    // 必填；非空；长度 1-512
    Weight *int    `json:"weight"` // 必填；[0,100]
}

type KeyPolicy struct {
    Strategy            *string `json:"strategy"`              // 默认 weighted_random
    MaxRetries          *int    `json:"max_retries"`           // 默认 0
    RetryBackoffInitial *int    `json:"retry_backoff_initial"` // 默认 500
    RetryBackoffMax     *int    `json:"retry_backoff_max"`     // 默认 5000
}

type LLMConfig struct {
    ModelEndpoint  *ModelEndpoint   `json:"model_endpoint"`
    Models         []string         `json:"models"`
    ModelMappings  []ModelMapping   `json:"model_mappings"`
    Keys           []APIKey         `json:"keys"`          // 新增；默认空数组
    KeyPolicy      *KeyPolicy       `json:"key_policy"`    // 新增
    ProviderType   *string          `json:"provider_type"`
}
```

### 2.2 InnerAPI 导出侧：`AIConf`

`ai-gateway-api` 通过 InnerAPI `/configs/tls_conf/server_data_conf` 向 BFE 导出 cluster 配置，其中 AI 相关配置映射为 BFE 的 `AIConf` 结构。为支持多 API-Key，`AIConf` 需同步扩展：

**当前结构：**

```go
// bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go
type AIConf struct {
    Type         int
    ModelMapping *map[string]string
    Key          *string
}
```

**改造后结构：**

```go
type AIKey struct {
    Name   string
    Key    string
    Weight int
}

type AIKeyPolicy struct {
    Strategy            string
    MaxRetries          int
    RetryBackoffInitial int
    RetryBackoffMax     int
}

type AIConf struct {
    Type         int
    ModelMapping *map[string]string
    Keys         []AIKey       // 新增
    KeyPolicy    *AIKeyPolicy  // 新增
}
```

> 说明：本节描述的是 BFE 侧目标结构，用于明确 `ai-gateway-api` 在 InnerAPI 导出时应生成的 JSON 格式。

### 2.3 字段映射关系

| ai-gateway-api | InnerAPI 导出（BFE `AIConf`） | 说明 |
|----------------|------------------------------|------|
| `LLMConfig.Keys[].Name` | `AIKey.Name` | Key 名称/标识 |
| `LLMConfig.Keys[].Key` | `AIKey.Key` | API-Key 值 |
| `LLMConfig.Keys[].Weight` | `AIKey.Weight` | 权重 |
| `LLMConfig.KeyPolicy.Strategy` | `AIKeyPolicy.Strategy` | 选择策略 |
| `LLMConfig.KeyPolicy.MaxRetries` | `AIKeyPolicy.MaxRetries` | 总重试次数 |
| `LLMConfig.KeyPolicy.RetryBackoffInitial` | `AIKeyPolicy.RetryBackoffInitial` | 初始退避（ms） |
| `LLMConfig.KeyPolicy.RetryBackoffMax` | `AIKeyPolicy.RetryBackoffMax` | 最大退避（ms） |

---

## 3. 关键系统设计

### 3.1 校验层设计

在 `validate.LLMConfig` 中新增校验：

1. `keys` 非必填，默认空数组 `[]`；
2. 若 `keys` 非空：
   - 每个元素 `key` 必填且非空，长度 1-512；
   - 每个元素 `name` 必填，长度 1-128，同一数组内唯一；
   - 每个元素 `weight` ∈ `[0,100]`；
   - 所有 `weight` 之和必须等于 `100`；
3. `key_policy` 若传入：
   - `strategy` 仅允许 `weighted_random`；
   - `max_retries` ≥ 0；
   - `retry_backoff_initial`、`retry_backoff_max` ≥ 0，且 `retry_backoff_max >= retry_backoff_initial`；
4. `model_endpoint.headers` 中若包含 `${API_KEY}` 占位符，则 `keys` 不能为空，否则返回 `422`。

### 3.2 配置导出设计

`model/icluster_conf/cluster.go` 中 `NewBfeClusterConf` 改造：

```go
if cluster.LLMConfig != nil {
    aiConf := &cluster_conf.AIConf{
        Type:         0,
        ModelMapping: convertToBFEModelMapping(cluster.LLMConfig.ModelMappings),
    }

    for _, k := range cluster.LLMConfig.Keys {
        aiConf.Keys = append(aiConf.Keys, cluster_conf.AIKey{
            Name:   derefString(k.Name),
            Key:    derefString(k.Key),
            Weight: derefInt(k.Weight),
        })
    }

    if cluster.LLMConfig.KeyPolicy != nil {
        aiConf.KeyPolicy = &cluster_conf.AIKeyPolicy{
            Strategy:            derefString(cluster.LLMConfig.KeyPolicy.Strategy, "weighted_random"),
            MaxRetries:          derefInt(cluster.LLMConfig.KeyPolicy.MaxRetries, 0),
            RetryBackoffInitial: derefInt(cluster.LLMConfig.KeyPolicy.RetryBackoffInitial, 500),
            RetryBackoffMax:     derefInt(cluster.LLMConfig.KeyPolicy.RetryBackoffMax, 5000),
        }
    }

    clusterConf.AIConf = aiConf
}
```

> `derefString` / `derefInt` 为取指针值的辅助函数，带默认值。

### 3.3 数据库设计

`clusters.llm_config` 为 JSON 字符串字段，新增 `keys` 与 `key_policy` 直接序列化存储，**无需新增表字段**。

### 3.4 conf-agent 设计

conf-agent 无需特殊改造：

- `clusters.llm_config` 以 JSON 字符串形式透传，新增的 `keys` / `key_policy` 自然随配置下发；
- 需关注敏感字段 `keys[].key` 在存储与传输过程中的加密/脱敏。

---

## 4. 接口影响

| 接口 | 方法 | 影响 |
|------|------|------|
| `/clusters` | POST | 请求体支持 `llm_config.keys` / `key_policy`；删除旧字段 `key` |
| `/clusters` | GET | 返回数据以 `keys` 形式展示 |
| `/clusters/{cluster_name}` | GET | 同上 |
| `/clusters/{cluster_name}` | PATCH | `keys` 数组全量替换 |
| InnerAPI `/configs/tls_conf/server_data_conf` | GET | `AIConf` 删除 `Key`，新增 `Keys` / `KeyPolicy` |

---

## 5. 实现步骤

1. **ai-gateway-api 模型层**
   - 修改 `model/icluster_conf/cluster.go` 中 `LLMConfig` 结构体；
   - 新增 `APIKey`、`KeyPolicy` 类型；
   - 更新 `NewBfeClusterConf` 导出逻辑。

2. **ai-gateway-api 校验层**
   - 扩展 `validate.LLMConfig`，新增 `keys` / `key_policy` / `${API_KEY}` 占位符校验。

3. **ai-gateway-api 控制层**
   - 更新 `normalizeLLMConfig`，未传 `keys` 时默认空数组，不再处理旧 `key` 字段。

4. **文档更新**
   - 更新 `design-docs/api-define/OpenAPI接口定义/clusters.md`；
   - 更新 `design-docs/api-define/InnerAPI接口定义/server-data-conf.md`。

5. **测试**
   - 多 Key 创建、更新、查询、权重校验；
   - `keys` 为空场景；
   - `key_policy` 默认值与合法性校验；
   - `${API_KEY}` 占位符与 `keys` 为空冲突校验；
   - InnerAPI 导出 `AIConf.Keys` / `AIConf.KeyPolicy` 格式正确性。

---

## 6. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 旧字段兼容 | 删除 `llm_config.key` 后，存量数据/调用方可能仍传入该字段 | 升级前清理存量数据；OpenAPI 层对旧字段返回 `422` 明确拒绝 |
| Key 明文安全 | InnerAPI 导出的 `server_data_conf` JSON 中 `Keys[].Key` 为明文 | 确保传输通道 TLS 加密；落盘时加密存储；必要时返回时脱敏 |
| 权重总和校验 | 多 Key 场景下 `weight` 之和必须为 100，配置错误概率增加 | 校验层强校验并返回清晰错误信息 |
| `name` 唯一性 | 同一 `keys` 数组内 `name` 必填且唯一 | 校验层强校验； InnerAPI 导出使用数组下标作为稳定标识 |
| 配置下发一致性 | ai-gateway-api 导出格式需与 BFE `AIConf` 目标结构保持一致 | 单元测试覆盖 InnerAPI 导出 JSON 结构；文档同步更新 |

---

## 7. 已确认设计决策

1. **Key 失败后降级/恢复策略**：参考 Bifrost 开源版 per-request key rotation，状态仅作用于单次请求，不引入持久化熔断/降级机制。
2. **`keys` 为空时的 `${API_KEY}` 占位符**：在 OpenAPI 校验阶段即拒绝该配置，运行时不做替换或直接移除包含占位符的 header 作为兜底。
3. **`keys[].name` 为必填字段**：长度 1-128，同一 `keys` 数组内唯一。

---

*文档生成日期：2026-08-07*
