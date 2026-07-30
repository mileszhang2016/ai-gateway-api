# rate-limit-policy 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出限流策略配置 | 供 BFE 执行 TPM/RPM/并发限制检查 |
| 端点 | `/configs/rate-limit-policy` | - |
| Method | GET | - |
| 鉴权 | `FeatureRateLimitPolicy + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/rate-limit-policy?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

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
        "Version": "00010101000000"
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

### 3.2 Config 结构（路由规则）

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

### 3.3 RateLimitPolicies 结构（限流策略）

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

### 3.4 ApikeyRateLimitPolicyBindings 结构（绑定关系）

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

## 4. 成功返回示例

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
        "Version": "00010101000000"
    },
    "WorkMode": "ModeNormal"
}
```

## 5. 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

