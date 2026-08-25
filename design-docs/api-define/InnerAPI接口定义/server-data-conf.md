# server_data_conf 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 BFE 路由/域名/集群综合配置 | 供 BFE `tls_conf/server_data_conf` 使用，包含 HostTable、RouteTable、ClusterConf 三部分 |
| 端点 | `/configs/tls_conf/server_data_conf` | - |
| Method | GET | - |
| 鉴权 | `FeatureRoute + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/tls_conf/server_data_conf?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

与 BFE 动态配置文件 `server_data_conf` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "HostTable": {
            "Version": "00010101000000",
            "DefaultProduct": "AI_product",
            "Hosts": {
                "host-tag-1": ["api.example.com"]
            },
            "HostTags": {
                "AI_product": ["host-tag-1"]
            }
        },
        "RouteTable": {
            "Version": "00010101000000",
            "BasicRule": {},
            "ProductRule": {
                "AI_product": [
                    {
                        "Cond": "req_host_in(\"api.example.com\")",
                        "ClusterName": "my-cluster"
                    }
                ]
            }
        },
        "ClusterConf": {
            "Version": "00010101000000",
            "Config": {
                "my-cluster": {
                    "BackendConf": {
                        "Protocol": "http",
                        "TimeoutConnSrv": 50000,
                        "TimeoutResponseHeader": 50000,
                        "MaxIdleConnsPerHost": 0,
                        "MaxConnsPerHost": 0,
                        "CancelOnClientClose": false
                    },
                    "CheckConf": {
                        "Schem": "http",
                        "Uri": "/",
                        "Host": "",
                        "HostType": "",
                        "StatusCode": 0,
                        "FailNum": 3,
                        "CheckInterval": 1000
                    },
                    "GslbBasic": {
                        "CrossRetry": 0,
                        "RetryMax": 2,
                        "HashConf": {
                            "HashStrategy": 0,
                            "HashHeader": "Cookie:USERID",
                            "SessionSticky": false
                        },
                        "BalanceMode": "WRR",
                        "EPPAddr": null
                    },
                    "ClusterBasic": {
                        "TimeoutReadClient": 30000,
                        "TimeoutWriteClient": 60000,
                        "TimeoutReadClientAgain": 30000,
                        "ReqWriteBufferSize": 512,
                        "ReqFlushInterval": 0,
                        "ResFlushInterval": 0,
                        "CancelOnClientClose": false,
                        "DisableHostHeader": false,
                        "DisableHealthCheck": false
                    },
                    "HTTPSConf": null
                }
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号 |
| HostTable | object | BFE 域名路由表，映射 host-tag、product、hostname 关系 |
| RouteTable | object | BFE 路由规则表，包含 BasicRule 与 ProductRule |
| ClusterConf | object | BFE 集群配置，key 为集群名称 |

### 3.2 ClusterConf 中 AI 相关字段

当 OpenAPI 的集群配置了 `llm_config` 时，导出后的集群配置（`ClusterConf.Config.<cluster_name>`）中会包含 `AIConf`：

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
                    },
                    "MatchPrefix": "deepseek/",
                    "StripPrefix": true,
                    "ModelProtocols": ["openai"],
                    "ModelTable": {
                        "Currency": "RMB",
                        "TimeZone": "Asia/Shanghai",
                        "Tiers": [
                            {
                                "Name": "peak",
                                "TimeRanges": [
                                    { "Weekdays": [1, 2, 3, 4, 5], "Start": "09:00", "End": "12:00" },
                                    { "Weekdays": [1, 2, 3, 4, 5], "Start": "14:00", "End": "18:00" }
                                ]
                            }
                        ],
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
                                    "output_cost_per_token": 0.000008,
                                    "cache_read_input_token_cost": 0.0000005
                                },
                                "TierPrices": {
                                    "peak": {
                                        "input_cost_per_token": 0.000004,
                                        "output_cost_per_token": 0.000016,
                                        "cache_read_input_token_cost": 0.000001
                                    }
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

| 字段 | 类型 | 说明 |
|------|------|------|
| Type | int | 固定为 0，保留字段 |
| ModelMapping | object | 模型名称映射，key 为请求模型名，value 为后端实际模型名 |
| Provider | string | 对应 OpenAPI `llm_config.provider`；默认空字符串 |
| Keys | array | API-Key 列表；为空数组时表示该 cluster 不配置 API-Key |
| Keys[].Name | string | Key 名称/标识（必填） |
| Keys[].Key | string | API-Key 值 |
| Keys[].Weight | int | 权重，范围 `[0,100]` |
| KeyPolicy | object | Key 路由策略 |
| KeyPolicy.Strategy | string | 本版仅支持 `weighted_random` |
| KeyPolicy.MaxRetries | int | 请求内总额外重试次数 |
| KeyPolicy.RetryBackoffInitial | int | 初始退避时间，单位毫秒 |
| KeyPolicy.RetryBackoffMax | int | 最大退避时间，单位毫秒 |
| MatchPrefix | string | 需要匹配的 provider/model 前缀；对应 OpenAPI `llm_config.match_prefix` |
| StripPrefix | bool | 是否裁剪 `MatchPrefix` 前缀；对应 OpenAPI `llm_config.strip_prefix` |
| ModelProtocols | []string | 该集群所属 provider 支持的模型访问协议；来源为 OpenAPI `/providers` 的 `model_protocols`。枚举值如 `openai`、`anthropic`；为空数组时 BFE 兜底为仅支持 `openai` |
| ModelTable | object | 该 cluster 的成本定价表 |
| ModelTable.Currency | string | 价格货币；固定为 `"RMB"` |
| ModelTable.TimeZone | string | 计算时段所使用的时区；默认 `"Asia/Shanghai"` |
| ModelTable.Tiers | array | 时段 tier 定义列表；为空时按固定价格处理 |
| ModelTable.Tiers[].Name | string | Tier 名称；**初期只支持 `"peak"`** |
| ModelTable.Tiers[].TimeRanges | array | 时段范围列表；命中任意一个即属于该 tier |
| ModelTable.Tiers[].TimeRanges[].Weekdays | []int | 星期几；0=周日，1=周一，...，6=周六；为空表示每天 |
| ModelTable.Tiers[].TimeRanges[].Start | string | 开始时间，格式 `HH:MM` |
| ModelTable.Tiers[].TimeRanges[].End | string | 结束时间，格式 `HH:MM`；采用左闭右开语义 |
| ModelTable.Models | array | 模型定价条目列表 |
| ModelTable.Models[].Provider | string | Provider 名 |
| ModelTable.Models[].Model | string | 模型名，用于匹配请求中的 target_model |
| ModelTable.Models[].BaseModel | string | 归一化模型名 |
| ModelTable.Models[].Mode | string | 请求模式，默认 `"chat"`；枚举值同 OpenAPI `model_prices.mode` |
| ModelTable.Models[].Capabilities | []string | 能力列表；枚举值同 OpenAPI `model_prices.capabilities` |
| ModelTable.Models[].SupportedParameters | []string | 支持的请求参数列表；枚举值同 OpenAPI `model_prices.supported_parameters` |
| ModelTable.Models[].Limits | object | 限制对象；键名枚举值同 OpenAPI `model_prices.limits` |
| ModelTable.Models[].Prices | object | 默认价格对象；键名枚举值同 OpenAPI `model_prices.prices`；未命中 tier 时作为 fallback |
| ModelTable.Models[].TierPrices | object | 分时段价格对象；key 为 tier name，value 为价格对象；键名枚举值同 OpenAPI `model_prices.prices` |

> **说明**：`ModelTable` 由 InnerAPI 根据 `cluster.llm_config.provider` 查询 `/providers` 的 `time_zone` / `tiers` 与 `/model-prices` 的 `prices` / `tier_prices` 拼接后自动填充，不在 OpenAPI `/clusters` 端点中展示。`Currency` 固定为 `RMB`。

## 4. 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---
