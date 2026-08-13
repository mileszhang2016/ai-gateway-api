# 配额支持人民币（RMB）与模型定价管理：OpenAPI 接口变更说明

## 1. 变更概述

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-11 |
| 影响模块 | `/model-prices`、`/api-keys`、`/entities`、`/clusters` |
| 目标文档 | `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/00-common.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/api-keys.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/entities.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/clusters.md`<br>`ai-gateway-api/design-docs/api-define/OpenAPI接口定义/model-prices.md`（新增）<br>`ai-gateway-api/design-docs/api-define/InnerAPI接口定义/server-data-conf.md` |

本次变更围绕 **配额支持人民币（RMB）单位** 与 **模型定价表管理** 展开：

1. 新增 `/v1/model-prices` 系列接口，用于维护 `model_prices` 表；
2. `QuotaPlan` 公共类型中 `quota`、`balance.used`、`balance.remaining` 由 `int64` 扩展为 `number`，`unit` 增加 `"RMB"`；
3. `/api-keys` 与 `/entities` 的 `quota_plan` 字段同步引用新的 `QuotaPlan` 定义；
4. `/clusters` 的 `llm_config` 新增 `provider` 字段；`model_table` 不在 OpenAPI 中展示，仅通过 InnerAPI 下发到 BFE 的 `ClusterConf.AIConf`。

> 说明：v0.4 仅支持人民币（RMB），不扩展美元、欧元等多币种。

---

## 2. 新增接口：`/v1/model-prices`

### 2.1 接口列表

| 接口 | 方法 | 说明 |
|------|------|------|
| `/v1/model-prices/import` | POST | 整表导入 `model-list.yaml`，支持 `replace` / `merge` 两种模式 |
| `/v1/model-prices` | POST | 新增单条模型定价记录 |
| `/v1/model-prices` | GET | 分页列表查询，支持按 `provider`、`mode` 过滤 |
| `/v1/model-prices/{id}` | GET | 按 `id` 查询单条记录 |
| `/v1/model-prices` | GET（带查询参数） | 按 `(provider, model, mode)` 组合查询单条记录 |
| `/v1/model-prices/{id}` | PUT | 按 `id` 修改单条记录 |
| `/v1/model-prices` | PUT（带查询参数） | 按 `(provider, model, mode)` 组合修改单条记录 |
| `/v1/model-prices/{id}` | DELETE | 按 `id` 删除单条记录 |
| `/v1/model-prices` | DELETE（带查询参数） | 按 `(provider, model, mode)` 组合删除单条记录 |

### 2.2 整表导入：`POST /v1/model-prices/import`

一次性全量或增量更新 `model_prices` 表。

```http
POST /v1/model-prices/import
Content-Type: multipart/form-data
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | file | Y | `model-list.yaml` 文件 |
| `mode` | string | N | 导入模式：`replace`（默认，全量替换）/ `merge`（增量合并） |

**处理逻辑：**

1. 解析 YAML，校验 `version`、`default_currency`（必须为 `RMB`）；
2. 校验每条记录的 `(provider, model, mode)` 唯一性；
3. 校验必填字段：`provider`、`model`、`base_model`、`mode`、`prices`；
4. 校验 `prices` 中至少包含一个价格字段；
5. `replace` 模式：先清空 `model_prices` 表，再写入新数据；
6. `merge` 模式：对已有 `(provider, model, mode)` 记录更新，新增记录插入。

**权限：** 仅允许管理员调用。

**响应：**

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "imported_count": 7,
    "skipped_count": 0,
    "errors": []
  }
}
```

### 2.3 单条记录结构

#### 2.3.1 新增：`POST /v1/model-prices`

```http
POST /v1/model-prices
Content-Type: application/json
```

**请求体：**

```json
{
  "provider": "deepseek",
  "model": "deepseek-v3",
  "base_model": "deepseek-v3",
  "mode": "chat",
  "capabilities": ["chat", "reasoning", "tools"],
  "supported_parameters": ["temperature", "max_tokens"],
  "limits": {
    "context_window": 128000,
    "max_input_tokens": 128000,
    "max_output_tokens": 8192
  },
  "prices": {
    "input_cost_per_token": 0.000002,
    "output_cost_per_token": 0.000008
  },
  "metadata": {
    "source": "https://platform.deepseek.com/pricing",
    "notes": "DeepSeek V3"
  }
}
```

**响应：** 返回完整记录，含生成的 `id`、`created_at`、`updated_at`。

#### 2.3.2 字段说明

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| `provider` | string | Y | Provider / Cluster 标识 | 非空；长度 1-255 |
| `model` | string | Y | 模型名 | 非空；长度 1-255 |
| `base_model` | string | Y | 归一化模型名 | 非空；长度 1-255 |
| `mode` | string | Y | 请求模式 | 枚举：`chat`、`completion`、`responses`、`image_generation`、`image_edit`、`embedding`、`rerank`、`audio_speech`、`audio_transcription`、`video_generation`、`ocr`、`search`、`realtime` |
| `capabilities` | []string | N | 模型支持的能力列表 | 默认空数组；取值应为枚举：`chat`、`vision`、`audio_input`、`video_input`、`reasoning`、`tools`、`structured_outputs`、`function_calling`、`prompt_caching`、`computer_use`、`web_search`、`serverless`、`image_generation`、`embedding`、`rerank`、`audio_speech`、`audio_transcription`、`video_generation`、`ocr`、`search`、`realtime` |
| `supported_parameters` | []string | N | 支持的请求参数列表 | 默认空数组；取值应为枚举：`temperature`、`top_p`、`max_tokens`、`tools`、`tool_choice`、`response_format`、`reasoning`、`image_input`、`video_input`、`audio_input`、`voice`、`speed`、`size`、`quality`、`style` |
| `limits` | object | N | 限制对象 | 默认空对象；键名应为枚举：`context_window`、`max_input_tokens`、`max_output_tokens`、`max_tokens` |
| `prices` | object | Y | 价格对象 | 至少包含一个价格字段；键名应为枚举：`input_cost_per_token`、`output_cost_per_token`、`cache_read_input_token_cost`、`cache_creation_input_token_cost`、`input_cost_per_token_above_200k_tokens`、`output_cost_per_token_above_200k_tokens`、`output_cost_per_image`、`output_cost_per_pixel`、`output_cost_per_image_low_quality`、`output_cost_per_image_high_quality`、`input_cost_per_audio_per_second`、`input_cost_per_video_per_second`、`output_cost_per_second`、`input_cost_per_query`、`search_context_cost_per_query`、`ocr_cost_per_page`、`output_cost_per_character`、`output_cost_per_image_hd`、`output_cost_per_video`、`output_cost_per_video_per_second`；所有价格字段必须为非负数 |
| `metadata` | object | N | 元数据 | 默认空对象；键名应为枚举：`source`、`notes` |

> `price_currency` 由系统根据 `model-list.yaml` 的 `default_currency` 固定填充为 `RMB`，请求体中无需传入。

#### 2.3.3 查询/修改/删除

**按 `id` 操作：**

```http
GET    /v1/model-prices/{id}
PUT    /v1/model-prices/{id}
DELETE /v1/model-prices/{id}
```

**按组合键操作：**

```http
GET    /v1/model-prices?provider={provider}&model={model}&mode={mode}
PUT    /v1/model-prices?provider={provider}&model={model}&mode={mode}
DELETE /v1/model-prices?provider={provider}&model={model}&mode={mode}
```

- `PUT` 请求体支持部分字段更新，未传入字段保持原值；
- `DELETE` 成功后返回标准成功响应；
- 查询参数 `provider`、`model`、`mode` 必须同时出现（或改用 `id` 路径参数）。

#### 2.3.4 分页列表查询

```http
GET /v1/model-prices?provider={provider}&mode={mode}&page=1&page_size=50
```

- `provider`、`mode` 为可选过滤条件；
- 返回列表中每条记录包含完整模型定价信息，包括 `id`。

### 2.4 校验规则

1. `provider`、`model`、`base_model`、`mode` 必填；
2. `(provider, model, mode)` 不能重复；
3. `prices` 必填，至少包含一个价格字段；
4. 所有价格字段必须为非负数；
5. `price_currency` 当前只支持 `RMB`；
6. `mode` 必须是预定义枚举值；
7. `capabilities`、`supported_parameters` 若传入，其元素应取自对应枚举值（非枚举值可接收但建议告警或记录，便于后续收敛）；
8. `limits`、`prices`、`metadata` 的键名应取自对应枚举值（非枚举键可接收但建议告警或记录，便于后续收敛）；
9. `file` 上传接口仅接受 YAML 文件，且 `default_currency` 必须为 `RMB`。

---

## 3. `QuotaPlan` 公共类型变更（`00-common.md`）

### 3.1 字段变更

| 字段 | 变更类型 | 说明 |
|------|----------|------|
| `quota` | **类型扩展** | 由 `int64` 改为 `number` |
| `balance.used` | **类型扩展** | 由 `int64` 改为 `number` |
| `balance.remaining` | **类型扩展** | 由 `int64` 改为 `number` |
| `unit` | **枚举扩展** | 可选值增加 `"RMB"` |

### 3.2 变更后 `QuotaPlan` 字段说明

| 字段 | 类型 | 必填 | 说明 | 合法性条件 |
|------|------|------|------|------------|
| `unlimited` | bool | N | 是否无限配额 | 默认 `true` |
| `pass_when_no_enough_quota` | bool | N | 配额不足时是否放行 | 默认 `false` |
| `quota` | number | N | 配额总量 | 非负数；`unit=total_token` 时必须为整数；`unit=RMB` 时内部最多保留 8 位小数，对外统一按 4 位小数展示 |
| `unit` | string | N | 配额单位 | 默认 `total_token`；可选值：`total_token`、`RMB` |
| `reset_period` | string | N | 配额重置周期 | 默认 `never`；可选值：`never`、`weekly`、`monthly` |
| `balance` | object | N | 余额状态（只读），包含 `used` 和 `remaining` | 作为输入时无需传入 |

### 3.3 `balance` 结构

| 字段 | 类型 | 说明 | 合法性条件 |
|------|------|------|------------|
| `used` | number | 已用量 | 内部最多 8 位小数，对外统一按 4 位小数展示 |
| `remaining` | number | 剩余量 | 内部最多 8 位小数，对外统一按 4 位小数展示 |

### 3.4 RMB 配额示例

```json
{
  "quota_plan": {
    "unlimited": false,
    "pass_when_no_enough_quota": false,
    "quota": 10000.00,
    "unit": "RMB",
    "reset_period": "monthly",
    "balance": {
      "used": 1234.56,
      "remaining": 8765.44
    }
  }
}
```

---

## 4. `/api-keys` 与 `/entities` 变更

涉及 `quota_plan` 的字段类型说明统一改为引用新的 `QuotaPlan` 定义：

- `quota` 类型由 `int64` 改为 `number`；
- `unit` 可选值增加 `"RMB"`；
- 重置配额接口 `/api-keys/{id}/quota-plan/reset` 与 `/entities/{id}/quota-plan/reset` 的 `quota`、`previous_quota`、`new_quota`、`previous_remaining`、`new_remaining` 统一改为 `number`。

### 4.1 校验规则补充

1. `unit` 仅允许 `total_token` 或 `RMB`；
2. 当 `unit = "total_token"` 时，`quota`、`used`、`remaining` 必须为大于等于 0 的整数；
3. 当 `unit = "RMB"` 时，`quota`、`used`、`remaining` 必须大于等于 0，且小数位不超过 8 位（即精确到 1e-8 元）；
4. `unlimited = true` 时，`quota` 仍可为 `0`，`balance` 不返回或返回 `0`。

### 4.2 影响接口

| 接口 | 变更说明 |
|------|----------|
| `POST /api-keys` | 请求体 `quota_plan` 支持 `unit = "RMB"` 与 `number` 类型配额 |
| `GET /api-keys` | 返回数据中 `quota_plan` 按 `number` 输出 |
| `GET /api-keys/{id}` | 同上 |
| `PATCH /api-keys/{id}` | 同上；`quota_plan` 整体替换 |
| `POST /api-keys/{id}/quota-plan/reset` | 重置结果中 `quota` / `remaining` 字段改为 `number` |
| `POST /entities` | 同 `POST /api-keys` |
| `GET /entities` | 同 `GET /api-keys` |
| `GET /entities/{id}` | 同上 |
| `PATCH /entities/{id}` | 同上 |
| `POST /entities/{id}/quota-plan/reset` | 同 `/api-keys/{id}/quota-plan/reset` |

---

## 5. `/clusters` 变更：`llm_config.provider`

在 OpenAPI 的 `/clusters` 接口中，`llm_config` 仅新增 `provider` 字段，用于显式指定该 cluster 在 `model_prices` 中对应的 `provider`。`model_table` **不会**在 OpenAPI `/clusters` 端点中展示，而是由 InnerAPI 根据 `provider` 自动填充后下发给 BFE。

### 5.1 `llm_config` 字段变更

| 字段 | 变更类型 | 说明 |
|------|----------|------|
| `provider` | **新增** | 显式指定该 cluster 在 `model_prices` 中对应的 `provider`；未填写时默认为空字符串；OpenAPI 可读写 |
| `model_table` | **不展示** | 该 cluster 的成本定价表，仅通过 InnerAPI 下发，不在 OpenAPI `/clusters` 中展示 |

### 5.2 变更后 `llm_config` 示例（OpenAPI）

```json
{
  "llm_config": {
    "model_endpoint": {
      "schema": "https",
      "uri": "/v1/chat/completions",
      "headers": {
        "Authorization": "Bearer ${API_KEY}"
      }
    },
    "models": ["deepseek-v3", "deepseek-chat"],
    "model_mappings": [
      {"source_model": "gpt-4", "target_model": "deepseek-v3"}
    ],
    "keys": [
      {"name": "default", "key": "sk-xxxxxxxxxxxx", "weight": 100, "models": [], "enabled": true}
    ],
    "key_selection_policy": "weighted_random",
    "provider_type": "deepseek",
    "provider": "deepseek"
  }
}
```

> 说明：OpenAPI 的 `/clusters` 请求/响应中均不包含 `model_table`。该字段由 InnerAPI 根据 `provider` 自动填充后下发给 BFE。

### 5.3 字段说明

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
|--------|------|----------|------|----------|
| `provider` | string | 该 cluster 在 `model_prices` 中对应的 `provider` | N | 默认空字符串；为空时 InnerAPI 下发的 `AIConf.ModelTable` 为空列表 |

### 5.4 影响接口

| 接口 | 变更说明 |
|------|----------|
| `POST /clusters` | 请求体 `llm_config` 允许传入 `provider`；响应中不返回 `model_table` |
| `GET /clusters` | 返回数据中 `llm_config` 仅包含 `provider`，不包含 `model_table` |
| `GET /clusters/{cluster_name}` | 同上 |
| `PATCH /clusters/{cluster_name}` | 支持部分更新 `llm_config.provider`；调用方传入 `model_table` 时忽略或报错 |

---

## 6. InnerAPI 变更：`ClusterConf.AIConf` 新增 `ModelTable`

`model_table` 不出现在 OpenAPI `/clusters` 端点中，而是在 InnerAPI 导出的 `ClusterConf.Config.<cluster_name>.AIConf` 中自动填充并下发给 BFE。

### 6.1 影响文档

| 文档 | 变更说明 |
|------|----------|
| `design-docs/api-define/InnerAPI接口定义/server-data-conf.md` | 更新 `AIConf` 字段说明，新增 `Provider` 与 `ModelTable` |

### 6.2 `AIConf` 结构变更

InnerAPI 导出的 `server_data_conf` 中，`ClusterConf.Config.<cluster_name>.AIConf` 新增 `Provider` 与 `ModelTable`：

**变更前：**

```json
{
    "ClusterConf": {
        "Version": "00010101000000",
        "Config": {
            "my-cluster": {
                "BackendConf": { ... },
                "CheckConf": { ... },
                "GslbBasic": { ... },
                "ClusterBasic": { ... },
                "AIConf": {
                    "Type": 0,
                    "ModelMapping": {
                        "gpt-4": "deepseek-chat"
                    },
                    "Keys": [ ... ],
                    "KeyPolicy": { ... }
                }
            }
        }
    }
}
```

**变更后：**

```json
{
    "ClusterConf": {
        "Version": "00010101000000",
        "Config": {
            "deepseek-cluster": {
                "BackendConf": { ... },
                "CheckConf": { ... },
                "GslbBasic": { ... },
                "ClusterBasic": { ... },
                "AIConf": {
                    "Type": 0,
                    "ModelMapping": {
                        "gpt-4": "deepseek-v3"
                    },
                    "Provider": "deepseek",
                    "Keys": [ ... ],
                    "KeyPolicy": { ... },
                    "ModelTable": {
                        "Models": [
                            {
                                "Provider": "deepseek",
                                "Model": "deepseek-v3",
                                "BaseModel": "deepseek-v3",
                                "Mode": "chat",
                                "Capabilities": ["chat", "reasoning", "tools"],
                                "SupportedParameters": ["temperature", "max_tokens"],
                                "Limits": {
                                    "context_window": 128000,
                                    "max_input_tokens": 128000,
                                    "max_output_tokens": 8192
                                },
                                "Prices": {
                                    "input_cost_per_token": 0.000002,
                                    "output_cost_per_token": 0.000008
                                }
                            }
                        ]
                    }
                }
            }
        }
    }
}
```

### 6.3 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `Provider` | string | 对应 OpenAPI `llm_config.provider`；默认空字符串 |
| `ModelTable` | object | 该 cluster 的成本定价表 |
| `ModelTable.Models` | array | 模型定价条目列表 |
| `ModelTable.Models[].Provider` | string | Provider 名 |
| `ModelTable.Models[].Model` | string | 模型名，用于匹配请求中的 target_model |
| `ModelTable.Models[].BaseModel` | string | 归一化模型名 |
| `ModelTable.Models[].Mode` | string | 请求模式，默认 `"chat"` |
| `ModelTable.Models[].Capabilities` | []string | 能力列表，同 `model_prices.capabilities` 枚举 |
| `ModelTable.Models[].SupportedParameters` | []string | 支持的请求参数列表，同 `model_prices.supported_parameters` 枚举 |
| `ModelTable.Models[].Limits` | object | 限制对象，同 `model_prices.limits` 枚举键 |
| `ModelTable.Models[].Prices` | object | 价格对象，同 `model_prices.prices` 枚举键 |

### 6.4 下发规则

1. 每个 cluster 的 `ModelTable` 来源为 `model_prices` 中 `provider = AIConf.Provider` 的记录；仅当 `Provider` 非空时才会查询，若为空则 `ModelTable.Models` 为空列表；
2. 若某 cluster 对应的 `provider` 在 `model_prices` 中无记录，`ModelTable.Models` 为空列表；
3. `model_prices` 变更后**不会同步触发** InnerAPI 下发；BFE / conf-agent 按自身周期拉取配置，热加载到本地，无需重启。

---

## 7. 完整示例

### 7.1 导入 `model-list.yaml`

```bash
curl -X POST https://api.example.com/v1/model-prices/import \
  -F "file=@model-list.yaml" \
  -F "mode=replace"
```

### 7.2 创建 RMB 配额的 API-Key

```json
{
  "description": "RMB 配额测试 Key",
  "expired_time": -1,
  "unlimited_quota": false,
  "models": ["*"],
  "subnet": ["*"],
  "quota_plan": {
    "unlimited": false,
    "pass_when_no_enough_quota": false,
    "quota": 5000.00,
    "unit": "RMB",
    "reset_period": "monthly"
  },
  "rate_limit_policy": {
    "enabled": false
  },
  "route_rules": {
    "enabled": true,
    "rules": [
      {
        "name": "apikey-default",
        "Cond": "default_t()",
        "targets": [
          {"ClusterName": "deepseek-cluster", "Model": "", "Weight": 100}
        ],
        "fallbacks": []
      }
    ]
  }
}
```

### 7.3 创建 Cluster 并通过 InnerAPI 下发 `ModelTable`

OpenAPI 创建 Cluster（`llm_config` 中仅包含 `provider`，不含 `model_table`）：

```json
{
  "name": "deepseek-cluster",
  "description": "DeepSeek 集群（RMB 定价）",
  "instance_pool": [
    {"name": "backend-1", "addr": "10.0.0.1", "weight": 50, "port": 8080},
    {"name": "backend-2", "addr": "10.0.0.2", "weight": 50, "port": 8080}
  ],
  "llm_config": {
    "model_endpoint": {
      "schema": "https",
      "uri": "/v1/chat/completions",
      "headers": {
        "Authorization": "Bearer ${API_KEY}"
      }
    },
    "models": ["deepseek-v3"],
    "keys": [
      {"name": "default", "key": "sk-xxxxxxxxxxxx", "weight": 100, "models": [], "enabled": true}
    ],
    "key_selection_policy": "weighted_random",
    "provider_type": "deepseek",
    "provider": "deepseek"
  }
}
```

InnerAPI 导出 `server_data_conf` 时，在 `ClusterConf.Config.deepseek-cluster.AIConf` 中自动附加 `ModelTable`：

```json
{
    "ClusterConf": {
        "Version": "00010101000000",
        "Config": {
            "deepseek-cluster": {
                "AIConf": {
                    "Type": 0,
                    "ModelMapping": {},
                    "Provider": "deepseek",
                    "Keys": [ ... ],
                    "KeyPolicy": { ... },
                    "ModelTable": {
                        "Models": [
                            {
                                "Provider": "deepseek",
                                "Model": "deepseek-v3",
                                "BaseModel": "deepseek-v3",
                                "Mode": "chat",
                                "Prices": {
                                    "input_cost_per_token": 0.000002,
                                    "output_cost_per_token": 0.000008
                                }
                            }
                        ]
                    }
                }
            }
        }
    }
}
```

---

## 8. 对下游/实现的影响

| 下游/实现 | 影响说明 |
|-----------|----------|
| **OpenAPI 文档** | 新增 `model-prices.md`；更新 `00-common.md`、`api-keys.md`、`entities.md`、`clusters.md` |
| **数据库** | 新增 `model_prices` 表；`quota_plans.quota`、`quota_balances.used` / `remaining` 建议改为 `DECIMAL(18,8)` |
| **模型层** | 新增 `ModelPrice` 实体及 CRUD 逻辑；`QuotaPlan` 相关字段由 `int64` 改为 `number`（内部可用定点整数存储） |
| **校验层** | 扩展 `QuotaPlan` 校验：区分 `total_token` 整数约束与 `RMB` 8 位小数约束；新增 `model_prices` 校验 |
| **YAML 解析层** | 实现 `model-list.yaml` 解析、校验、入库 |
| **InnerAPI 层** | 按 cluster / provider 拆分并导出 `model_table` |
| **BFE 转发层** | 加载 cluster `model_table`；在请求完成后按 `input_cost_per_token` / `output_cost_per_token` 计算 RMB 成本并扣减；RMB 余额建议使用定点数存储，避免 Lua 浮点误差 |
| **conf-agent** | 透传 `AIConf.ModelTable`；按自身周期拉取配置 |
| **集成测试** | 补充 RMB 配额创建、更新、余额查询、重置、请求扣减；`model-prices` 导入/CRUD；定价缺失 fail-open 等场景 |

---

## 9. 已确认事项

1. **货币范围**：v0.4 只支持人民币（RMB），不扩展美元、欧元等多币种。
2. **余额精度**：内部按 8 位小数（1e-8 元）定点数存储；对外展示统一按 4 位小数输出。
3. **未命中定价时的行为**：采用 Bifrost 式 fail-open：找不到定价时放行并按 0 成本计算，RMB 配额不扣减，记 Warn 日志；v0.4 暂不实现严格模式开关。
4. **请求前预检**：只做“余额 > 0”的粗略预检；不实现 `max_tokens` 最坏估算，不实现“预扣除 + 结算回滚”。
5. **缓存 / 分层定价**：`model-list.yaml` 可保留相关字段，但 BFE 扣减逻辑 v0.4 暂只使用 `input_cost_per_token` / `output_cost_per_token`。
6. **整表导入权限**：`/v1/model-prices/import` 只允许管理员调用；v0.4 暂时不需要审计日志。
7. **配置下发实时性**：BFE 定期调用 InnerAPI 拉取最新配置；`model_prices` 变更后不会同步触发下发。

---

*文档生成日期：2026-08-11*
