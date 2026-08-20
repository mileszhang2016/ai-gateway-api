# 配额支持人民币（RMB）与模型定价管理：系统设计变更说明

## 1. 概述

### 1.1 变更背景

当前 `ai-gateway-api` 的配额计划 `QuotaPlan` 仅支持以 **Token** 为单位（`unit = "total_token"`）。随着业务需要按人民币成本进行精细化管理，现有设计存在以下瓶颈：

- 无法直接按金额设定配额上限与查看余额；
- 缺乏统一的模型定价数据源，BFE 无法将实际 Token 消耗换算为人民币成本；
- 模型价格分散维护，运营成本高，且难以与上游厂商报价同步。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-11 |
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `ai-gateway-api/model/iquota_plan`、`ai-gateway-api/model/icluster_conf`、`ai-gateway-api/endpoints/openapi_v1/product_cluster`、`ai-gateway-api/endpoints/openapi_v1/product_model_price` |
| 变更类型 | 数据模型扩展 + 新增管理接口 + InnerAPI 导出改造 + 接口文档更新 |

本次变更围绕 **配额支持人民币（RMB）单位** 与 **模型定价表管理** 展开：

1. 新增 `model_prices` 表与 `/v1/model-prices` 管理接口，统一维护模型定价；
2. `QuotaPlan` 支持 `unit = "RMB"`，`quota` / `used` / `remaining` 由 `int64` 扩展为 `number`；
3. `/api-keys` 与 `/entities` 的 `quota_plan` 同步支持 RMB 单位；
4. `/clusters` 的 `llm_config` 新增 `provider` 字段；`model_table` 仅通过 InnerAPI 下发到 BFE 的 `ClusterConf.AIConf`；
5. v0.4 仅支持人民币（RMB），不扩展美元、欧元等多币种。

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **单一数据源** | 模型定价统一采用 `model-list.yaml` 格式，解析后持久化到 `model_prices` 表，作为 OpenAPI 与 InnerAPI 的唯一数据源。 |
| **OpenAPI 最小暴露** | `model_table` 不在 OpenAPI `/clusters` 中展示，仅暴露可配置的 `provider`；定价表通过 InnerAPI 自动填充下发，避免用户直接修改。 |
| **Fail-open 运行时** | BFE 未命中模型定价时默认放行并按 0 成本处理，不阻塞请求；通过 Warn 日志与监控指标暴露配置缺失风险。 |
| **精度可控** | RMB 余额内部按 8 位小数（1e-8 元）定点数存储，对外统一按 4 位小数展示；Redis Lua 脚本使用整数运算避免浮点误差。 |
| **兼容存量** | `unit` 默认仍为 `total_token`，存量 Token 配额数据与调用方式不变。 |

---

## 2. 数据模型设计

### 2.1 ai-gateway-api 侧：`ModelPrice`

新增 `ModelPrice` 实体，对应数据库 `model_prices` 表：

```go
type ModelPrice struct {
    ID                   int64           `json:"id"`
    Provider             string          `json:"provider"`              // 对应 model-list.yaml 的 provider
    Model                string          `json:"model"`                 // 对应 model-list.yaml 的 model
    BaseModel            string          `json:"base_model"`            // 归一化模型名
    Mode                 string          `json:"mode"`                  // 模型模式
    Capabilities         []string        `json:"capabilities"`          // 能力列表
    SupportedParameters  []string        `json:"supported_parameters"`  // 支持的请求参数列表
    Limits               map[string]any  `json:"limits"`                // 限制对象
    Prices               map[string]any  `json:"prices"`                // 价格对象
    PriceCurrency        string          `json:"price_currency"`        // 当前固定 RMB
    Metadata             map[string]any  `json:"metadata"`              // 元数据
    CreatedAt            time.Time       `json:"created_at"`
    UpdatedAt            time.Time       `json:"updated_at"`
}
```

**唯一索引**：`(provider, model, mode)`

### 2.2 ai-gateway-api 侧：`QuotaPlan`

`QuotaPlan` 相关字段由 `int64` 扩展为 `number`（浮点数）：

```go
type QuotaPlan struct {
    Unlimited               *bool          `json:"unlimited"`
    PassWhenNoEnoughQuota   *bool          `json:"pass_when_no_enough_quota"`
    Quota                   *float64       `json:"quota"`        // 由 *int64 改为 *float64
    Unit                    *string        `json:"unit"`         // total_token / RMB
    ResetPeriod             *string        `json:"reset_period"`
    Balance                 *BalanceSummary `json:"balance"`
}

type BalanceSummary struct {
    Used      *float64 `json:"used"`       // 由 *int64 改为 *float64
    Remaining *float64 `json:"remaining"`  // 由 *int64 改为 *float64
}
```

> 说明：内部存储可使用 `DECIMAL(18,8)` 或定点整数；API 层统一按 `number` 输出。

### 2.3 InnerAPI 导出侧：`AIConf.ModelTable`

`ai-gateway-api` 通过 InnerAPI `/configs/tls_conf/server_data_conf` 向 BFE 导出 cluster 配置。`AIConf` 新增 `Provider` 与 `ModelTable`：

```go
// bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go（目标结构）
type ModelPriceEntry struct {
    Provider            string
    Model               string
    BaseModel           string
    Mode                string
    Capabilities        []string
    SupportedParameters []string
    Limits              map[string]any
    Prices              map[string]any
}

type ModelTable struct {
    Models []ModelPriceEntry
}

type AIConf struct {
    Type         int
    ModelMapping *map[string]string
    Provider     string          // 新增
    Keys         []AIKey
    KeyPolicy    *AIKeyPolicy
    ModelTable   *ModelTable     // 新增
}
```

> 说明：本节描述的是 BFE 侧目标结构，用于明确 `ai-gateway-api` 在 InnerAPI 导出时应生成的 JSON 格式。

### 2.4 字段映射关系

| ai-gateway-api | InnerAPI 导出（BFE `AIConf`） | 说明 |
|----------------|------------------------------|------|
| `LLMConfig.Provider` | `AIConf.Provider` | cluster 在 `model_prices` 中对应的 provider |
| `ModelPrice.Provider` | `ModelPriceEntry.Provider` | Provider 名 |
| `ModelPrice.Model` | `ModelPriceEntry.Model` | 模型名 |
| `ModelPrice.BaseModel` | `ModelPriceEntry.BaseModel` | 归一化模型名 |
| `ModelPrice.Mode` | `ModelPriceEntry.Mode` | 请求模式 |
| `ModelPrice.Capabilities` | `ModelPriceEntry.Capabilities` | 能力列表 |
| `ModelPrice.SupportedParameters` | `ModelPriceEntry.SupportedParameters` | 支持的请求参数列表 |
| `ModelPrice.Limits` | `ModelPriceEntry.Limits` | 限制对象 |
| `ModelPrice.Prices` | `ModelPriceEntry.Prices` | 价格对象 |

---

## 3. 关键系统设计

### 3.1 YAML 解析与导入设计

`POST /v1/model-prices/import` 处理逻辑：

1. 解析 YAML，校验 `version`、`default_currency`（必须为 `RMB`）；
2. 校验每条记录的 `(provider, model, mode)` 唯一性；
3. 校验必填字段：`provider`、`model`、`base_model`、`mode`、`prices`；
4. 校验 `prices` 中至少包含一个价格字段，所有价格字段为非负数；
5. `replace` 模式：清空 `model_prices` 表后写入新数据；
6. `merge` 模式：对已有记录更新，新增记录插入。

### 3.2 校验层设计

在 `validate.ModelPrice` 与 `validate.QuotaPlan` 中新增校验：

**`ModelPrice` 校验：**

1. `provider`、`model`、`base_model`、`mode` 必填；
2. `(provider, model, mode)` 不能重复；
3. `prices` 必填，至少包含一个价格字段；
4. 所有价格字段必须为非负数；
5. `price_currency` 当前只支持 `RMB`；
6. `mode` 必须是预定义枚举值；
7. `capabilities`、`supported_parameters` 若传入，元素应取自对应枚举值；
8. `limits`、`prices`、`metadata` 的键名应取自对应枚举值。

**`QuotaPlan` 校验：**

1. `unit` 仅允许 `total_token` 或 `RMB`；
2. `unit = "total_token"` 时，`quota`、`used`、`remaining` 必须为大于等于 0 的整数；
3. `unit = "RMB"` 时，`quota`、`used`、`remaining` 必须大于等于 0，且小数位不超过 8 位；
4. `unlimited = true` 时，`quota` 仍可为 `0`，`balance` 不返回或返回 `0`。

### 3.3 InnerAPI 导出设计

`model/icluster_conf/cluster.go` 中 `NewBfeClusterConf` 改造：

```go
if cluster.LLMConfig != nil && cluster.LLMConfig.Provider != nil && *cluster.LLMConfig.Provider != "" {
    provider := *cluster.LLMConfig.Provider
    modelPrices, err := modelPriceRepo.ListByProvider(provider)
    if err != nil {
        // 记录日志，不影响导出
    }

    modelTable := &cluster_conf.ModelTable{
        Models: make([]cluster_conf.ModelPriceEntry, 0, len(modelPrices)),
    }
    for _, mp := range modelPrices {
        modelTable.Models = append(modelTable.Models, cluster_conf.ModelPriceEntry{
            Provider:            mp.Provider,
            Model:               mp.Model,
            BaseModel:           mp.BaseModel,
            Mode:                mp.Mode,
            Capabilities:        mp.Capabilities,
            SupportedParameters: mp.SupportedParameters,
            Limits:              mp.Limits,
            Prices:              mp.Prices,
        })
    }

    aiConf.Provider = provider
    aiConf.ModelTable = modelTable
}
```

> 若 `llm_config.provider` 为空，`AIConf.ModelTable.Models` 为空列表。

### 3.4 数据库设计

#### 3.4.1 新增表 `model_prices`

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint / uuid | 主键 |
| `provider` | varchar(255) | Provider / Cluster 标识 |
| `model` | varchar(255) | 模型名 |
| `base_model` | varchar(255) | 归一化模型名 |
| `mode` | varchar(50) | 模型模式 |
| `capabilities` | json | 能力列表 |
| `supported_parameters` | json | 支持参数列表 |
| `limits` | json | 限制对象 |
| `prices` | json | 价格对象 |
| `price_currency` | varchar(10) | 当前固定 `RMB` |
| `metadata` | json | 元数据 |
| `created_at` / `updated_at` | datetime | 时间戳 |

唯一索引：`(provider, model, mode)`

#### 3.4.2 配额相关表字段类型调整

| 表 | 字段 | 建议变更 |
|----|------|----------|
| `quota_plans` | `quota` | `BIGINT` → `DECIMAL(18,8)` |
| `quota_balances` | `used` | `BIGINT` → `DECIMAL(18,8)` |
| `quota_balances` | `remaining` | `BIGINT` → `DECIMAL(18,8)` |

#### 3.4.3 `clusters` 表

当前 `clusters.llm_config` 为 JSON 字符串字段，新增 `provider` 直接序列化存储，**无需新增表字段**。`model_table` 由 `ai-gateway-api` 根据 `model_prices` 自动生成，通过 InnerAPI 下发到 `ClusterConf.AIConf`。

---

## 4. 接口影响

### 4.1 OpenAPI 接口

| 接口 | 方法 | 影响 |
|------|------|------|
| `/v1/model-prices/import` | POST | 新增接口，整表导入 `model-list.yaml` |
| `/v1/model-prices` | POST | 新增接口，新增单条模型定价记录 |
| `/v1/model-prices` | GET | 新增接口，分页列表查询 |
| `/v1/model-prices/{id}` | GET | 新增接口，按 `id` 查询 |
| `/v1/model-prices` | GET（带查询参数） | 新增接口，按 `(provider, model, mode)` 查询 |
| `/v1/model-prices/{id}` | PUT | 新增接口，按 `id` 修改 |
| `/v1/model-prices` | PUT（带查询参数） | 新增接口，按 `(provider, model, mode)` 修改 |
| `/v1/model-prices/{id}` | DELETE | 新增接口，按 `id` 删除 |
| `/v1/model-prices` | DELETE（带查询参数） | 新增接口，按 `(provider, model, mode)` 删除 |
| `/api-keys` | POST/GET/PATCH | `quota_plan` 支持 `unit = "RMB"` 与 `number` 类型配额 |
| `/api-keys/{id}/quota-plan/reset` | POST | 重置结果中 `quota` / `remaining` 改为 `number` |
| `/entities` | POST/GET/PATCH | 同 `/api-keys` |
| `/entities/{id}/quota-plan/reset` | POST | 同 `/api-keys/{id}/quota-plan/reset` |
| `/clusters` | POST/GET | `llm_config` 新增 `provider`；响应中不返回 `model_table` |
| `/clusters/{cluster_name}` | GET/PATCH | 同上；`PATCH` 传入 `model_table` 时忽略或报错 |

### 4.2 InnerAPI 接口

| 接口 | 方法 | 影响 |
|------|------|------|
| `/configs/tls_conf/server_data_conf` | GET | `ClusterConf.Config.<cluster>.AIConf` 新增 `Provider` 与 `ModelTable` |

---

## 5. 实现步骤

1. **存储层**
   - 新增 `model_prices` 表；
   - 迁移 `quota_plans.quota`、`quota_balances.used` / `remaining` 为 `DECIMAL(18,8)`。

2. **ai-gateway-api 模型层**
   - 新增 `ModelPrice` 实体及 CRUD 逻辑；
   - 将 `QuotaPlanParam.Quota`、`BalanceSummary.Used` / `Remaining` 从 `*int64` 改为 `*float64`（或内部使用 decimal 类型）。

3. **ai-gateway-api YAML 解析层**
   - 实现 `model-list.yaml` 解析、校验、入库；
   - 支持 `replace` / `merge` 两种导入模式。

4. **ai-gateway-api 校验层**
   - 扩展 `QuotaPlan` 校验：区分 `total_token` 整数约束与 `RMB` 8 位小数约束；
   - 新增 `ModelPrice` 校验：枚举值检查、价格非负、唯一键检查。

5. **ai-gateway-api OpenAPI 控制层**
   - 新增 `/v1/model-prices` 系列接口；
   - 更新 `/clusters` 控制层，支持 `llm_config.provider`，拒绝/忽略 `model_table`；
   - 更新 `/api-keys`、`/entities` 控制层，支持 `unit = "RMB"`。

6. **ai-gateway-api InnerAPI 层**
   - 更新 `NewBfeClusterConf`，按 `llm_config.provider` 查询 `model_prices` 并填充 `AIConf.ModelTable`。

7. **BFE 转发层**
   - 加载 cluster `AIConf.ModelTable`；
   - 在 `CalcReqUsedQuota` / `Deduct` 流程中区分 `total_token` 与 `RMB`；
   - 使用定点数存储 RMB 余额，避免 Lua 浮点误差；
   - 未命中定价时按 0 成本放行，记 Warn 日志。

8. **文档更新**
   - 更新 `design-docs/api-define/OpenAPI接口定义/00-common.md`；
   - 更新 `design-docs/api-define/OpenAPI接口定义/api-keys.md`；
   - 更新 `design-docs/api-define/OpenAPI接口定义/entities.md`；
   - 更新 `design-docs/api-define/OpenAPI接口定义/clusters.md`；
   - 新增 `design-docs/api-define/OpenAPI接口定义/model-prices.md`；
   - 更新 `design-docs/api-define/InnerAPI接口定义/server-data-conf.md`。

9. **测试**
   - `model-prices` 导入/CRUD/枚举校验；
   - RMB 配额创建、更新、余额查询、重置；
   - InnerAPI 导出 `AIConf.Provider` / `AIConf.ModelTable` 格式正确性；
   - BFE 请求扣减 RMB 与 Token 配额；
   - 定价缺失 fail-open 场景。

---

## 6. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 浮点精度误差 | RMB 余额若用浮点数直接扣减，会产生累计误差 | 内部使用 8 位小数定点整数或 `DECIMAL(18,8)`；Redis Lua 使用整数运算 |
| 定价缺失 | BFE 运行时找不到对应模型定价 | Fail-open 按 0 成本放行，记 Warn 日志；建议增加 `model_pricing_missing_total` 监控指标 |
| 配额预检不足 | 请求前无法知道最终输出 Token 数 | 仅做粗略预检（余额 > 0），精确扣减在响应完成后进行；v0.4 不做 `max_tokens` 最坏估算 |
| 币种扩展 | v0.4 仅支持 RMB，未来扩展多币种时需改造 `price_currency` 与汇率逻辑 | 当前 schema 预留币种字段，实现时按 RMB 硬编码处理 |
| OpenAPI 暴露 `model_table` 风险 | `model_table` 若通过 OpenAPI 暴露，可能被用户误改 | OpenAPI `/clusters` 仅暴露 `provider`；`model_table` 仅通过 InnerAPI 下发 |
| 存量 Token 配额兼容 | `quota` 字段类型由 `int64` 改为 `number` | 数据库字段扩展为 `DECIMAL(18,8)`，存量整数值不变；API 对 `total_token` 仍输出整数 |
| 敏感 Key 安全 | InnerAPI 导出的 `AIConf.Keys[].Key` 为明文 | 确保传输通道 TLS 加密；落盘时加密存储；必要时返回时脱敏 |

---

## 7. 已确认设计决策

1. **货币范围**：v0.4 只支持人民币（RMB），不扩展美元、欧元等多币种。
2. **余额精度**：内部按 8 位小数（1e-8 元）定点数存储；对外展示统一按 4 位小数输出。
3. **未命中定价时的行为**：采用 Bifrost 式 fail-open：找不到定价时放行并按 0 成本计算，RMB 配额不扣减，记 Warn 日志；v0.4 暂不实现严格模式开关。
4. **请求前预检**：只做“余额 > 0”的粗略预检；不实现 `max_tokens` 最坏估算，不实现“预扣除 + 结算回滚”。
5. **缓存 / 分层定价**：`model-list.yaml` 可保留相关字段，但 BFE 扣减逻辑 v0.4 暂只使用 `input_cost_per_token` / `output_cost_per_token`。
6. **整表导入权限**：`/v1/model-prices/import` 只允许管理员调用；v0.4 暂时不需要审计日志。
7. **配置下发实时性**：BFE 定期调用 InnerAPI 拉取最新配置；`model_prices` 变更后不会同步触发下发。
8. **`model_table` 展示位置**：`model_table` 不在 OpenAPI `/clusters` 中展示，仅通过 InnerAPI 下发到 `ClusterConf.AIConf`。

---

*文档生成日期：2026-08-11*
