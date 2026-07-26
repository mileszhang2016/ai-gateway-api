# InnerAPI 接口定义

**版本号**：v0.3.0  

---

## 一、概述

### 1.1 文档目的

本文档描述 AI 网关内部 API（Inner-API）的接口设计，用于 BFE 节点从控制面拉取配置数据。这些接口是 BFE 配置同步机制的核心，BFE 节点定期调用这些接口获取最新的配置信息。

### 1.2 接口特点

- **只读接口**：所有 Inner-API 均为 GET 请求，用于导出配置数据
- **版本控制**：支持基于版本的增量同步，当配置未变化时返回空响应
- **鉴权机制**：使用 McUserProbe 中间件进行身份验证
- **统一格式**：遵循现有 Inner-API 的设计模式和数据结构

### 1.3 与 OpenAPI 的关系

| 维度 | OpenAPI (v0.3.0) | Inner-API |
|------|------------------|-----------|
| **用途** | 管理面 CRUD 操作 | 数据面配置同步 |
| **调用方** | Web 管理控制台 | BFE 节点 |
| **接口类型** | RESTful (GET/POST/PUT/PATCH/DELETE) | 只读导出 (GET) |
| **数据粒度** | 单个资源操作 | 批量配置导出 |
| **路径前缀** | `/open-api/v1` | `/inner-api/v1` |

### 1.4 依赖的相关文档及版本号

|                  |        |      |
|----------|--------|------|
| OpenAPI 接口定义 | v0.3.0 |  |

---

## 二、通用说明

### 2.1 基本信息

| 项目 | 值 | 说明 |
|------|------|------|
| URL 格式 | `http://api_server:port/inner-api/v1/{endpoint}?{arg=value}` | 例：`http://127.0.0.1:8086/inner-api/v1/configs/mod-api-key` |
| 鉴权方式 | McUserProbe | HTTP Header 中间件鉴权 |
| 版本控制 | 基于 version 参数 | 配置未变化时返回空响应 |

### 2.2 返回值格式

所有 Inner-API 的返回值格式：

```json
{
    "ErrNum": 200,
    "Data": {},
    "ErrMsg": "success",
    "WorkMode": "current mode"
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ErrNum | int | 返回码，200 表示成功 |
| Data | object | 返回的数据结构，配置未变化时返回 null |
| ErrMsg | string | 文本消息，成功时为 "success"，失败时为错误信息 |
| WorkMode | string | 控制台工作模式 |

### 2.3 版本控制机制

所有导出接口支持 `version` 查询参数：

- **首次拉取**：不传 `version` 或传空字符串，返回完整配置
- **增量拉取**：传入上次返回的 `version`，若配置未变化返回 `Data: null`
- **配置变化**：返回新的配置数据和新的 `version`

---

## 三、现有接口清单

### 3.1 接口总览

| 序号 | 接口路径 | 功能描述 | 状态 |
|------|----------|----------|------|
| 1 | `/configs/tls_conf/server_data_conf` | 导出 TLS/Server 配置 | 现有 |
| 2 | `/configs/gslb_data/gslb` | 导出 GSLB 配置 | 现有 |
| 3 | `/configs/gslb_data/cluster_table` | 导出集群表配置 | 现有 |
| 4 | `/configs/protocol/server_cert_conf` | 导出证书配置 | 现有 |
| 5 | `/configs/extra_files/{filename}` | 导出额外文件 | 现有 |
| 6 | `/configs/mod-api-key` | 导出 API-Key 配置 | **需扩充** |
| 7 | `/configs/rate-limit-policy` | 导出限流策略配置 | **新增** |
| 8 | `/configs/ai-route` | 导出 AI 路由配置 | **新增** |

---

## 四、mod-api-key 接口

### 4.1 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 API-Key 及配额配置 | 供 BFE 进行 Token 鉴权和配额检查 |
| 端点 | `/configs/mod-api-key` | - |
| Method | GET | - |
| 鉴权 | `FeatureAPIKey + ActionExport` | - |

### 4.2 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | 否 | 上次返回的版本号，用于增量同步 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/mod-api-key?version=v1.0.1" \
  -H "Authorization:Token TOKEN_STRING"
```

### 4.3 返回数据结构

#### 4.3.1 顶层结构

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "version": "v1.0.2",
        "config": {
            "product_name": [/* API-Key 路由规则 */]
        },
        "QuotaPlans": {
            "product_name": [/* 配额计划定义 */]
        },
        "tokens": {
            "product_name": {
                "api_key_value": {/* Token 配置 */}
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| version | string | 配置版本号 |
| config | object | 按产品线分组的 API-Key 路由规则 |
| QuotaPlans | object | 按产品线分组的配额计划定义，Token 通过 `quota_plans` 数组引用 |
| tokens | object | 按产品线分组的 Token 配置 |

#### 4.3.2 config 结构（路由规则）

```json
{
    "config": {
        "ai_product": [
            {
                "Cond": "req_host_in(\"api.example.com\")",
                "action": {
                    "cmd": "CHECK_TOKEN"
                }
            }
        ]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Cond | string | 路由匹配条件表达式 |
| action | object | 匹配后执行的动作 |
| action.cmd | string | 动作命令，固定为 `CHECK_TOKEN` |

#### 4.3.3 tokens 结构（Token 配置）

**实际导出的 TokenFile 结构**

```json
{
    "tokens": {
        "ai_product": {
            "AI_product-abcdef123456": {
                "key": "AI_product-abcdef123456",
                "enabled": 1,
                "status": 1,
                "name": "ak-test-key-001",
                "update_time": 1716883200,
                "expired_time": -1,
                "unlimited_quota": false,
                "allow_models": "gpt-4,gpt-3.5-turbo",
                "block_models": "gpt-4-32k",
                "subnet": "192.168.0.0/16,10.0.0.0/8",
                "quota_plans": ["ak-test-key-001", "dept-engineering"],
                "Tags": [
                    {"TagName": "department", "TagValue": "dept-engineering"}
                ]
            }
        }
    }
}
```

**Token 字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| key | string | API-Key 值 | 系统生成的 Key 字符串 |
| enabled | int | 是否启用 | `1`: 启用, `2`: 禁用 |
| status | int | Key 状态 | `1`: 启用, `2`: 禁用, `3`: 已过期, `4`: 配额耗尽 |
| name | string | API-Key ID | 对应 API-Key 的唯一标识 |
| update_time | int64 | 创建时间 | Unix 时间戳（秒） |
| expired_time | int64 | 过期时间 | `-1`: 永不过期；其他为 Unix 时间戳（秒） |
| unlimited_quota | bool | 是否无限配额 | `true`: 不检查配额；`false`: 检查配额 |
| allow_models | string | 允许访问的模型 | 逗号分隔，空或 `""` 表示不限制 |
| block_models | string | 禁止访问的模型 | 逗号分隔，空或 `""` 表示无不允许模型 |
| subnet | string | 允许的客户端子网 | 逗号分隔的 CIDR 列表，空或 `""` 表示不限制 |
| quota_plans | []string | 关联的配额计划 ID 列表 | 引用顶层 `QuotaPlans` 中的定义，包含 API-Key 自身和 Entity 层级的所有配额计划 |
| Tags | []ApikeyTag | Entity 层级标签列表 | 包含 `TagName` 和 `TagValue` 字段 |

**ApikeyTag 结构（Entity 层级标签，按字段名导出）**

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| TagName | string | Entity 类型 | `department`, `team`, `project` |
| TagValue | string | Entity 名称 | `dept-engineering`, `team-core` |

#### 4.3.4 QuotaPlans 结构（配额计划定义）

配额计划定义在顶层 `QuotaPlans` 中按产品线分组，Token 通过 `quota_plans` 数组引用这些计划的 ID。

```json
{
    "QuotaPlans": {
        "ai_product": [
            {
                "Id": "ak-test-key-001",
                "Unlimited": false,
                "PassNoQuota": false,
                "RedisKey": "QUOTA_ak-test-key-001",
                "CreateTime": 1716883200,
                "ExpiredTime": -1,
                "Quota": 100000000,
                "ResetMode": 1
            }
        ]
    }
}
```

**QuotaPlan 字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| Id | string | 配额计划 ID | 通常为 API-Key ID 或 Entity ID |
| Unlimited | bool | 是否无限配额 | `true`/`false` |
| PassNoQuota | bool | 配额不足时是否放行 | `true`: 放行；`false`: 拒绝 |
| RedisKey | string | Redis 中存储配额余额的 Key | 格式 `QUOTA_{id}` |
| CreateTime | int64 | 创建时间 | Unix 时间戳（秒） |
| ExpiredTime | int64 | 过期时间 | `-1`: 永不过期 |
| Quota | int64 | 配额总量 | 初始配额值 |
| ResetMode | int | 重置模式 | `0`: 非周期性；`1`: 周期性（weekly/monthly） |

### 4.4 状态码说明

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 1 | 启用 | API-Key 正常可用 |
| 2 | 禁用 | API-Key 已被禁用 |
| 3 | 已过期 | API-Key 已过期 |
| 4 | 配额耗尽 | API-Key 配额已用完 |

同时，`enabled` 字段独立表示是否启用：
- `1` = 启用
- `2` = 禁用

### 4.5 成功返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "version": "v1.0.2",
        "config": {
            "ai_product": [
                {
                    "Cond": "req_host_in(\"api.example.com\")",
                    "Action": {
                        "Cmd": "CHECK_TOKEN"
                    }
                },
                {
                    "Cond": "default_t()",
                    "Action": {
                        "Cmd": "CHECK_TOKEN"
                    }
                }
            ]
        },
        "QuotaPlans": {
            "ai_product": [
                {
                    "Id": "ak-test-key-001",
                    "Unlimited": false,
                    "PassNoQuota": false,
                    "RedisKey": "QUOTA_ak-test-key-001",
                    "CreateTime": 1716883200,
                    "ExpiredTime": -1,
                    "Quota": 100000000,
                    "ResetMode": 1
                },
                {
                    "Id": "dept-engineering",
                    "Unlimited": false,
                    "PassNoQuota": false,
                    "RedisKey": "QUOTA_dept-engineering",
                    "CreateTime": 1716000000,
                    "ExpiredTime": -1,
                    "Quota": 500000000,
                    "ResetMode": 0
                }
            ]
        },
        "tokens": {
            "ai_product": {
                "AI_product-abcdef123456": {
                    "key": "AI_product-abcdef123456",
                    "enabled": 1,
                    "status": 1,
                    "name": "ak-test-key-001",
                    "update_time": 1716883200,
                    "expired_time": -1,
                    "unlimited_quota": false,
                    "allow_models": "gpt-4,gpt-3.5-turbo",
                    "block_models": "gpt-4-32k",
                    "subnet": "192.168.0.0/16,10.0.0.0/8",
                    "quota_plans": ["ak-test-key-001", "dept-engineering"],
                    "Tags": [
                        {"TagName": "department", "TagValue": "dept-engineering"}
                    ]
                },
                "AI_product-ghijkl789012": {
                    "key": "AI_product-ghijkl789012",
                    "enabled": 1,
                    "status": 1,
                    "name": "ak-prod-key-002",
                    "update_time": 1716883200,
                    "expired_time": -1,
                    "unlimited_quota": true,
                    "allow_models": "",
                    "block_models": "",
                    "subnet": "",
                    "quota_plans": [],
                    "Tags": []
                }
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

### 4.6 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

## 五、rate-limit-policy 接口

### 5.1 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出限流策略配置 | 供 BFE 执行 TPM/RPM/并发限制检查 |
| 端点 | `/configs/rate-limit-policy` | - |
| Method | GET | - |
| 鉴权 | `FeatureRateLimit + ActionExport` | **新增权限** |

### 5.2 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | 否 | 上次返回的版本号，用于增量同步 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/rate-limit-policy?version=v1.0.1" \
  -H "Authorization:Token TOKEN_STRING"
```

### 5.3 返回数据结构

#### 5.3.1 顶层结构

与 BFE 动态配置文件 `ai_rate_limit.data` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Config": {
            "AI_product": [/* 路由规则 */]
        },
        "RateLimitPolicies": {
            "rlp-0001": {/* 限流策略 */}
        },
        "ApikeyRateLimitPolicyBindings": {
            "ak-2v8x9k3m7p": ["rlp-0001", "rlp-0002"]
        },
        "Version": "v1.0.2"
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Config | object | 按产品线组织的路由规则，key 为产品线名称 |
| RateLimitPolicies | object | 限流策略定义，key 为策略 ID（如 `rlp-0001`） |
| ApikeyRateLimitPolicyBindings | object | API-Key 到策略 ID 列表的绑定关系 |
| Version | string | 配置版本号 |

#### 5.3.2 Config 结构（路由规则）

```json
{
    "Config": {
        "AI_product": [
            {
                "cond": "default_t()",
                "hit_action": {
                    "cmd": "FINISH",
                    "params": []
                }
            }
        ]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| cond | string | 路由匹配条件表达式，通常填 `default_t()` |
| hit_action | object | 命中条件后的动作 |
| hit_action.cmd | string | 动作命令，支持 `PASS`、`RETURN`、`REDIRECT`、`FINISH`、`CLOSE` |
| hit_action.params | []string | 动作参数 |

**说明**：命中 cond 后，模块自动执行限流检查流程（提取 API-Key → 查找策略 → 执行限流），`hit_action.cmd` 用于指定被限流规则拒绝后的行为。

#### 5.3.3 RateLimitPolicies 结构（限流策略）

```json
{
    "RateLimitPolicies": {
        "rlp-0001": {
            "name": "rlp-0001",
            "enabled": true,
            "rules": {
                "tpm": [
                    {
                        "name": "win2min",
                        "models": ["*"],
                        "window_minutes": 1,
                        "max_tokens": 10000,
                        "step_minutes": 1
                    },
                    {
                        "name": "win10min",
                        "models": ["gpt-4"],
                        "window_minutes": 10,
                        "max_tokens": 50000,
                        "step_minutes": 1
                    }
                ],
                "rpm": [
                    {
                        "name": "win2min",
                        "models": ["gpt-4"],
                        "window_minutes": 1,
                        "max_requests": 100,
                        "burst": 1
                    }
                ],
                "max_concurrency": 50
            }
        }
    }
}
```

**RateLimitPolicy 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 策略名称 |
| enabled | bool | 是否启用限流，false 时跳过该策略 |
| rules | object | 限流规则集合 |

**rules 结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| tpm | []TPMConfig | Token Per Minute 限制配置，最多 3 个；为空则不做 TPM 限制 |
| rpm | []RPMConfig | Request Per Minute 限制配置，最多 3 个；为空则不做 RPM 限制 |
| max_concurrency | int | 最大并发数，>=1 表示限制；0 表示封禁；<0 表示不限制 |

**TPMConfig 结构**

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| name | string | 规则名称 | 同一策略内不能重复 |
| models | []string | 目标模型列表 | 为空或 `["*"]` 表示不限制模型；非空时仅对匹配模型执行限流，多个 model 共用同一限流器 |
| window_minutes | int | 统计时间窗口（分钟） | 取值范围 1-360 |
| max_tokens | int | 最大 Token 数 | >0: 有限制；0: 封禁；<0: 不限制 |
| step_minutes | int | 滑动步长（分钟） | 取值范围 1-360，必须 <= window_minutes |

**RPMConfig 结构**

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| name | string | 规则名称 | 同一策略内不能重复 |
| models | []string | 目标模型列表 | 为空或 `["*"]` 表示不限制模型；非空时仅对匹配模型执行限流，多个 model 共用同一限流器 |
| window_minutes | int | 统计时间窗口（分钟） | 取值范围 1-360，默认 1 |
| max_requests | int | 最大请求数 | >=1: 有限制；0: 封禁；<0: 不限制 |
| burst | int | 突发请求数 | 最小值 1，默认 1 |

#### 5.3.4 ApikeyRateLimitPolicyBindings 结构（绑定关系）

```json
{
    "ApikeyRateLimitPolicyBindings": {
        "ak-2v8x9k3m7p": ["rlp-0001", "rlp-0002"],
        "ak-3w9y0k4n8q": ["rlp-0002"]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | API-Key 字符串 |
| value | []string | 绑定的策略 ID 列表，绑定多个策略时必须全部满足才通过 |

### 5.4 成功返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Config": {
            "AI_product": [
                {
                    "cond": "default_t()",
                    "hit_action": {
                        "cmd": "FINISH",
                        "params": []
                    }
                }
            ]
        },
        "RateLimitPolicies": {
            "rlp-0001": {
                "name": "rlp-0001",
                "enabled": true,
                "rules": {
                    "tpm": [
                        {
                            "name": "global-tpm",
                            "models": ["*"],
                            "window_minutes": 1,
                            "max_tokens": 1000000,
                            "step_minutes": 1
                        },
                        {
                            "name": "gpt4-tpm",
                            "models": ["gpt-4"],
                            "window_minutes": 1,
                            "max_tokens": 500000,
                            "step_minutes": 1
                        }
                    ],
                    "rpm": [
                        {
                            "name": "global-rpm",
                            "models": ["*"],
                            "window_minutes": 1,
                            "max_requests": 60,
                            "burst": 1
                        }
                    ],
                    "max_concurrency": 50
                }
            },
            "rlp-0002": {
                "name": "rlp-0002",
                "enabled": true,
                "rules": {
                    "tpm": [
                        {
                            "name": "gpt4-tpm-limit",
                            "models": ["gpt-4"],
                            "window_minutes": 1,
                            "max_tokens": 5000,
                            "step_minutes": 1
                        }
                    ],
                    "max_concurrency": 10
                }
            },
            "rlp-0003": {
                "name": "rlp-0003",
                "enabled": false,
                "rules": {
                    "tpm": [],
                    "rpm": [],
                    "max_concurrency": -1
                }
            }
        },
        "ApikeyRateLimitPolicyBindings": {
            "ak-2v8x9k3m7p": ["rlp-0001", "rlp-0002"],
            "ak-3w9y0k4n8q": ["rlp-0002"],
            "ak-9z8y7x6w5v": ["rlp-0001"]
        },
        "Version": "v1.0.2"
    },
    "WorkMode": "ModeNormal"
}
```

### 5.5 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

## 六、ai-route 接口

### 6.1 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 AI 网关路由配置 | 供 BFE 的 `mod_ai_route` 模块执行 apikey → entity → global 三级路由查找 |
| 端点 | `/configs/ai-route` | - |
| Method | GET | - |
| 鉴权 | `FeatureRoute + ActionExport` | 复用或新增路由导出权限 |

### 6.2 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | 否 | 上次返回的版本号，用于增量同步 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/ai-route?version=20260718131505" \
  -H "Authorization:Token TOKEN_STRING"
```

### 6.3 返回数据结构

#### 6.3.1 顶层结构

与 BFE 动态配置文件 `ai_route.data` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "20260718131505",
        "RouteRules": {
            "apikey_ak_user_a": {/* API-Key 路由表 */},
            "entity_dept_ai": {/* Entity 路由表 */},
            "global_default": {/* Global 路由表 */}
        },
        "ApikeyRouteTableBindings": {
            "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"]
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号，建议使用时间戳格式（如 `YYYYMMDDHHMMSS`） |
| RouteRules | object | 所有路由表的集合，key 为 `<type>_<owner>`，保证全局唯一 |
| ApikeyRouteTableBindings | object | API-Key 到路由表查找顺序的映射 |

**说明**：仅导出 `route_rules.enabled = true` 的路由表。若 Global、API-Key 或 Entity 路由表的 `enabled = false`，则不生成该路由表，也不会加入 `ApikeyRouteTableBindings`。

#### 6.3.2 RouteRules 结构（路由表集合）

```json
{
    "RouteRules": {
        "apikey_ak_user_a": {
            "type": "apikey",
            "owner": "ak_user_a",
            "rules": [
                {
                    "name": "user_a-deepseek",
                    "Cond": "req_host_in(\"api.example.org\")",
                    "targets": [
                        {
                            "ClusterName": "cluster_deepseek_a",
                            "Model": "deepseek-v4-pro",
                            "Weight": 70
                        },
                        {
                            "ClusterName": "cluster_deepseek_b",
                            "Model": "deepseek-v4-pro",
                            "Weight": 30
                        }
                    ],
                    "fallbacks": [
                        {
                            "ClusterName": "cluster_deepseek_c",
                            "Model": "deepseek-v3.2"
                        }
                    ]
                }
            ]
        },
        "entity_dept_ai": {
            "type": "entity",
            "owner": "dept_ai",
            "rules": [
                {
                    "name": "dept_ai-default",
                    "Cond": "default_t()",
                    "targets": [
                        {
                            "ClusterName": "cluster_dept_ai",
                            "Model": "",
                            "Weight": 100
                        }
                    ],
                    "fallbacks": []
                }
            ]
        },
        "global_default": {
            "type": "global",
            "owner": "global",
            "rules": [
                {
                    "name": "global-default",
                    "Cond": "default_t()",
                    "targets": [
                        {
                            "ClusterName": "cluster_global",
                            "Model": "",
                            "Weight": 100
                        }
                    ],
                    "fallbacks": []
                }
            ]
        }
    }
}
```

**RouteTable 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 路由表类型，枚举值：`apikey`、`entity`、`global` |
| owner | string | 路由表属主。`apikey` 类型为 API-Key 标识；`entity` 类型为 Entity 名称；`global` 类型固定为 `global` |
| rules | array | 该路由表下的规则列表，按顺序匹配 |

**RouteRule 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 规则名称，用于日志、监控和问题定位 |
| Cond | string | BFE 条件表达式，命中则使用该规则 |
| targets | array | 转发目标列表 |
| fallbacks | array | 降级目标列表，允许为空 |

**Target 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传请求原始模型 |
| Weight | int | 权重，单个 target 时为 100，多个 target 时总和必须为 100 |

**Fallback 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传请求原始模型 |

#### 6.3.3 ApikeyRouteTableBindings 结构（绑定关系）

```json
{
    "ApikeyRouteTableBindings": {
        "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"],
        "ak_user_b": ["apikey_ak_user_b", "entity_dept_ai", "global_default"]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | API-Key 字符串 |
| value | []string | 该 API-Key 对应的路由表查找顺序，典型顺序为 `apikey → entity → global` |

### 6.4 成功返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "20260718131505",
        "RouteRules": {
            "apikey_ak_user_a": {
                "type": "apikey",
                "owner": "ak_user_a",
                "rules": [
                    {
                        "name": "user_a-deepseek",
                        "Cond": "req_host_in(\"api.example.org\")",
                        "targets": [
                            {
                                "ClusterName": "cluster_deepseek_a",
                                "Model": "deepseek-v4-pro",
                                "Weight": 70
                            },
                            {
                                "ClusterName": "cluster_deepseek_b",
                                "Model": "deepseek-v4-pro",
                                "Weight": 30
                            }
                        ],
                        "fallbacks": [
                            {
                                "ClusterName": "cluster_deepseek_c",
                                "Model": "deepseek-v3.2"
                            }
                        ]
                    },
                    {
                        "name": "user_a-default",
                        "Cond": "default_t()",
                        "targets": [
                            {
                                "ClusterName": "cluster_default",
                                "Model": "",
                                "Weight": 100
                            }
                        ],
                        "fallbacks": []
                    }
                ]
            },
            "entity_dept_ai": {
                "type": "entity",
                "owner": "dept_ai",
                "rules": [
                    {
                        "name": "dept_ai-default",
                        "Cond": "default_t()",
                        "targets": [
                            {
                                "ClusterName": "cluster_dept_ai",
                                "Model": "",
                                "Weight": 100
                            }
                        ],
                        "fallbacks": []
                    }
                ]
            },
            "global_default": {
                "type": "global",
                "owner": "global",
                "rules": [
                    {
                        "name": "global-default",
                        "Cond": "default_t()",
                        "targets": [
                            {
                                "ClusterName": "cluster_global",
                                "Model": "",
                                "Weight": 100
                            }
                        ],
                        "fallbacks": []
                    }
                ]
            }
        },
        "ApikeyRouteTableBindings": {
            "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"],
            "ak_user_b": ["apikey_ak_user_b", "entity_dept_ai", "global_default"]
        }
    },
    "WorkMode": "ModeNormal"
}
```

### 6.5 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

## 七、数据模型定义

### 7.1 ModAPIKeyRuleConf（实际实现）

```go
// ModAPIKeyRuleConf 定义 API-Key 规则配置结构
type ModAPIKeyRuleConf struct {
    Version    *string                          `json:"version"`
    Config     map[string][]*TokenRuleFile      `json:"config"`
    QuotaPlans map[string][]*QuotaPlan           `json:"QuotaPlans"`
    Tokens     map[string]map[string]*TokenFile `json:"tokens"`
}
```

### 7.2 TokenFile（实际实现）

```go
// TokenFile 定义导出到 BFE 的 API-Key 信息结构
type TokenFile struct {
    Key            string      `json:"key"`
    Enabled        int         `json:"enabled"`
    Status         int         `json:"status"`
    Name           string      `json:"name"`
    UpdateTime     int64       `json:"update_time"`
    ExpiredTime    int64       `json:"expired_time"`
    UnlimitedQuota bool        `json:"unlimited_quota"`
    Models         *string     `json:"allow_models"`
    BlockModels    *string     `json:"block_models"`
    Subnet         *string     `json:"subnet"`
    Tags           []ApikeyTag
    QuotaPlans     []string    `json:"quota_plans"`
}

type ApikeyTag struct {
    TagName  string  // Entity 类型，如 department, team
    TagValue string  // Entity 名称，如 dept-engineering
}
```

### 7.3 QuotaPlan（实际实现）

```go
// QuotaPlan 定义配额计划结构
type QuotaPlan struct {
    Id          string  // 配额计划 ID
    Unlimited   bool    // 是否无限配额
    PassNoQuota bool    // 配额不足时是否放行
    RedisKey    string  // Redis Key，格式 QUOTA_{id}
    CreateTime  int64   // 创建时间戳
    ExpiredTime int64   // 过期时间戳，-1 表示永不过期
    Quota       int64   // 配额总量
    ResetMode   int     // 0: 非周期性; 1: 周期性
}
```

### 7.4 RateLimitPolicyExport

```go
// RateLimitPolicyExport 定义限流策略导出结构（与 BFE 动态配置格式一致）
type RateLimitPolicyExport struct {
    Config                        map[string][]*ExportRouteRule              `json:"Config"`
    RateLimitPolicies             map[string]*RateLimitPolicyConfig          `json:"RateLimitPolicies"`
    ApikeyRateLimitPolicyBindings map[string][]string                        `json:"ApikeyRateLimitPolicyBindings"`
    Version                       string                                     `json:"Version"`
}

// ExportRouteRule 定义路由规则
type ExportRouteRule struct {
    Cond       string            `json:"cond"`
    HitAction  *ExportHitAction  `json:"hit_action"`
}

// ExportHitAction 定义命中动作
type ExportHitAction struct {
    Cmd    string   `json:"cmd"`
    Params []string `json:"params"`
}

// RateLimitPolicyConfig 定义单个限流策略配置
type RateLimitPolicyConfig struct {
    Name    string           `json:"name"`
    Enabled bool             `json:"enabled"`
    Rules   *RateLimitRules  `json:"rules"`
}

// RateLimitRules 定义限流规则集合
type RateLimitRules struct {
    TPM             []*TPMConfigExport  `json:"tpm"`
    RPM             []*RPMConfigExport  `json:"rpm"`
    MaxConcurrency  int                 `json:"max_concurrency"`
}

// TPMConfigExport 定义 TPM 限制配置
type TPMConfigExport struct {
    Name            string   `json:"name"`
    Models          []string `json:"models"`
    WindowMinutes   int      `json:"window_minutes"`
    MaxTokens       int      `json:"max_tokens"`
    StepMinutes     int      `json:"step_minutes"`
}

// RPMConfigExport 定义 RPM 限制配置
type RPMConfigExport struct {
    Name            string   `json:"name"`
    Models          []string `json:"models"`
    WindowMinutes   int      `json:"window_minutes"`
    MaxRequests     int      `json:"max_requests"`
    Burst           int      `json:"burst"`
}
```

### 7.5 AiRouteDataExport

```go
// AiRouteDataExport 定义 AI 路由配置导出结构（与 BFE ai_route.data 格式一致）
type AiRouteDataExport struct {
    Version                  string                `json:"Version"`
    RouteRules               map[string]*RouteTableExport `json:"RouteRules"`
    ApikeyRouteTableBindings map[string][]string   `json:"ApikeyRouteTableBindings"`
}

// RouteTableExport 定义单张路由表
type RouteTableExport struct {
    Type   string            `json:"type"`
    Owner  string            `json:"owner"`
    Rules  []*RouteRuleExport `json:"rules"`
}

// RouteRuleExport 定义路由规则
type RouteRuleExport struct {
    Name      string                `json:"name"`
    Cond      string                `json:"Cond"`
    Targets   []*AiRouteTargetExport `json:"targets"`
    Fallbacks []*AiRouteFallbackExport `json:"fallbacks"`
}

// AiRouteTargetExport 定义转发目标
type AiRouteTargetExport struct {
    ClusterName string `json:"ClusterName"`
    Model       string `json:"Model"`
    Weight      int    `json:"Weight"`
}

// AiRouteFallbackExport 定义降级目标
type AiRouteFallbackExport struct {
    ClusterName string `json:"ClusterName"`
    Model       string `json:"Model"`
}
```

---

## 八、附录

### 8.1 与 OpenAPI 数据对应关系

| Inner-API 字段 | OpenAPI 字段 | 数据来源 |
|----------------|--------------|----------|
| `tokens.{key}.quota_plan.unlimited` | `quota_plan.unlimited` | `quota_plans` 表 |
| `tokens.{key}.quota_plan.pass_when_no_enough_quota` | `quota_plan.pass_when_no_enough_quota` | `quota_plans` 表 |
| `tokens.{key}.quota_plan.quota` | `quota_plan.quota` | `quota_plans` 表 |
| `tokens.{key}.quota_plan.unit` | `quota_plan.unit` | `quota_plans` 表 |
| `tokens.{key}.quota_plan.reset_period` | `quota_plan.reset_period` | `quota_plans` 表 |
| `api_key_policies.{id}.enabled` | `rate_limit_policy.enabled` | `rate_limit_policies` 表 |
| `api_key_policies.{id}.rules.tpm` | `rate_limit_policy.rules.tpm` | `rate_limit_policies` 表 |
| `api_key_policies.{id}.rules.rpm` | `rate_limit_policy.rules.rpm` | `rate_limit_policies` 表 |
| `api_key_policies.{id}.rules.max_concurrency` | `rate_limit_policy.rules.max_concurrency` | `rate_limit_policies` 表 |
| `RouteRules.{key}.rules` | `api_keys.route_rules` / `entities.route_rules` / `global-route-rules` | `route_rules` 表 |
| `ApikeyRouteTableBindings.{api_key}` | API-Key 与 Entity 的挂载关系 | `api_keys.entity_id`、`entities.parent_id` |

