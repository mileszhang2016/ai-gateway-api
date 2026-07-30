# mod-api-key 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 API-Key 及配额配置 | 供 BFE 进行 Token 鉴权和配额检查 |
| 端点 | `/configs/mod-api-key` | - |
| Method | GET | - |
| 鉴权 | `FeatureAPIKey + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/mod-api-key?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "version": "00010101000000",
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

### 3.2 config 结构（路由规则）

```json
{
    "config": {
        "ai_product": [
            {
                "Cond": "req_host_in(\"api.example.com\")",
                "Action": {
                    "Cmd": "CHECK_TOKEN"
                }
            }
        ]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Cond | string | 路由匹配条件表达式 |
| Action | object | 匹配后执行的动作 |
| Action.Cmd | string | 动作命令，固定为 `CHECK_TOKEN` |

### 3.3 tokens 结构（Token 配置）

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

### 3.4 QuotaPlans 结构（配额计划定义）

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

## 4. 状态码说明

| 状态码 | 含义 | 说明 |
|--------|------|------|
| 1 | 启用 | API-Key 正常可用 |
| 2 | 禁用 | API-Key 已被禁用 |
| 3 | 已过期 | API-Key 已过期 |
| 4 | 配额耗尽 | API-Key 配额已用完 |

同时，`enabled` 字段独立表示是否启用：
- `1` = 启用
- `2` = 禁用

## 5. 成功返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "version": "00010101000000",
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

## 6. 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

