# API-Key

## 1 创建API-Key

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 创建API-Key | |
| 端点 | /api-keys | |
| 版本 | v1 | |
| method | POST | 创建资源 |
| Content-Type | application/json | - |

### 输入参数

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| description | string | API-Key描述 | Y | - |
| expired_time | int64 | 过期时间 | N | -1表示永不过期；其他为Unix时间戳（秒） |
| enabled | bool | 是否启用 | N | 默认true |
| unlimited_quota | bool | 是否无限配额 | N | 默认false |
| models | []string | 允许访问的模型白名单 | N | 包含"*"表示不限制，默认值为不限制 |
| subnet | []string | 允许的客户端子网 | N | 包含"*"表示不限制，默认值为不限制 |
| quota_plan | object | 配额计划 | N | 同Quota Plan结构（不含balance），若未设置则使用默认值 |
| rate_limit_policy | object | 限流策略 | N | 同Rate Limit Policy结构，若未设置则使用默认值（enabled=false） |
| route_rules | object | 路由规则 | N | 同Route Rules结构，若未设置则使用默认值（enabled=false, rules为空） |
| entity_id | string | 挂载的Entity ID | N | 为空表示不挂载 |

**约束**
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若 `entity_id` 不为空，该Entity必须存在

#### 请求示例
```shell
curl -X POST "http://api-server:port/open-api/v1/api-keys" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "description": "BFE项目测试Key",
    "expired_time": -1,
    "unlimited_quota": false,
    "models": ["*"],
    "subnet": ["*"],
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    },
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {"name": "1分钟窗口", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1}
            ],
            "rpm": [
                {"name": "1分钟请求", "model": "*", "window_minutes": 1, "max_requests": 100}
            ],
            "max_concurrency": 50
        }
    },
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "apikey-default",
                "Cond": "default_t()",
                "targets": [
                    {
                        "ClusterName": "cluster_apikey",
                        "Model": "",
                        "Weight": 100
                    }
                ],
                "fallbacks": []
            }
        ]
    },
    "entity_id": "ent-zhangsan-001"
}
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | API-Key唯一标识 | 系统生成 |
| key | string | API-Key值 | 系统生成，用于请求头鉴权 |
| description | string | 描述 | - |
| enabled | bool | 是否启用 | - |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |
| update_time | int64 | 更新时间 | Unix时间戳（秒） |
| expired_time | int64 | 过期时间 | -1表示永不过期 |
| unlimited_quota | bool | 是否无限配额 | - |
| models | []string | 允许访问的模型白名单 | - |
| subnet | []string | 允许的客户端子网 | - |
| quota_plan | object | 配额计划 | 不含balance字段 |
| rate_limit_policy | object | 限流策略 | - |
| route_rules | object | 路由规则 | - |
| entity_id | string | 挂载的Entity ID | - |
| entity | object | 挂载的Entity摘要 | 包含id、name、type |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "apikey-001",
        "key": "ak-2v8x9k3m7p",
        "description": "BFE项目测试Key",
        "enabled": true,
        "create_time": 1716883200,
        "update_time": 1716883200,
        "expired_time": -1,
        "unlimited_quota": false,
        "models": ["*"],
        "subnet": ["*"],
        "quota_plan": {
            "unlimited": false,
            "pass_when_no_enough_quota": false,
            "quota": 100000000,
            "unit": "total_token",
            "reset_period": "monthly"
        },
        "rate_limit_policy": {
            "enabled": true,
            "rules": {
                "tpm": [{"name": "1分钟窗口", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1}],
                "rpm": [{"name": "1分钟请求", "model": "*", "window_minutes": 1, "max_requests": 100}],
                "max_concurrency": 50
            }
        },
        "route_rules": {
            "enabled": true,
            "rules": [
                {
                    "name": "apikey-default",
                    "Cond": "default_t()",
                    "targets": [
                        {
                            "ClusterName": "cluster_apikey",
                            "Model": "",
                            "Weight": 100
                        }
                    ],
                    "fallbacks": []
                }
            ]
        },
        "entity_id": "ent-zhangsan-001",
        "entity": {
            "id": "ent-zhangsan-001",
            "name": "zhangsan",
            "type": "person"
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 2 查询API-Key列表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询API-Key列表 | |
| 端点 | /api-keys | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### Query 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| enabled | bool | 是否启用过滤 | N | - |
| entity_id | string | 按挂载的Entity ID过滤 | N | - |
| unlimited_quota | bool | 是否无限配额过滤 | N | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/api-keys?page=1&page_size=20&enabled=true" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| list | []object | API-Key列表 | 元素字段同创建接口返回，quota_plan中不含balance字段 |
| pagination | object | 分页信息 | - |
| pagination.page | int | 当前页码 | - |
| pagination.page_size | int | 每页条数 | - |
| pagination.total | int | 总条数 | - |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "list": [
            {
                "id": "apikey-001",
                "key": "ak-2v8x9k3m7p",
                "description": "BFE项目测试Key",
                "enabled": true,
                "create_time": 1716883200,
                "update_time": 1716883200,
                "expired_time": -1,
                "unlimited_quota": false,
                "models": ["*"],
                "subnet": ["*"],
                "quota_plan": {
                    "unlimited": false,
                    "pass_when_no_enough_quota": false,
                    "quota": 100000000,
                    "unit": "total_token",
                    "reset_period": "monthly"
                },
                "rate_limit_policy": {
                    "enabled": true,
                    "rules": {
                        "tpm": [{"name": "1分钟窗口", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1}],
                        "rpm": [{"name": "1分钟请求", "model": "*", "window_minutes": 1, "max_requests": 100}],
                        "max_concurrency": 50
                    }
                },
                "route_rules": {
                    "enabled": false,
                    "rules": []
                },
                "entity_id": "ent-zhangsan-001",
                "entity": {
                    "id": "ent-zhangsan-001",
                    "name": "zhangsan",
                    "type": "person"
                }
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 1
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 3 查询单个API-Key

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询单个API-Key | |
| 端点 | /api-keys/{id} | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/api-keys/apikey-001" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
字段同创建接口返回，quota_plan中不含balance字段。

#### 成功返回参数示例
同创建接口返回示例。

---

## 4 全量更新API-Key

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 全量更新API-Key | |
| 端点 | /api-keys/{id} | |
| 版本 | v1 | |
| method | PUT | 全量更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### Body 参数
同创建API-Key的Body参数。

**约束**
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan

#### 请求示例
```shell
curl -X PUT "http://api-server:port/open-api/v1/api-keys/apikey-001" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 同创建接口示例。

### 返回数据(Data内容)
同创建接口返回（不含balance）。

#### 成功返回参数示例
同创建接口返回示例。

---

## 5 部分更新API-Key

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 部分更新API-Key | |
| 端点 | /api-keys/{id} | |
| 版本 | v1 | |
| method | PATCH | 部分更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### Body 参数
同创建API-Key的Body参数，仅传需修改字段。

**约束**
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan
- 若修改 `quota_plan.quota`，视为重置配额（同步更新balance.remaining和used）
- 若修改 `route_rules`，视为全量替换该路由规则配置

#### 请求示例
```shell
curl -X PATCH "http://api-server:port/open-api/v1/api-keys/apikey-001" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下（仅修改description）：
```json
{
    "description": "更新后的描述"
}
```

### 返回数据(Data内容)
同创建接口返回（不含balance）。

#### 成功返回参数示例
同创建接口返回示例。

---

## 6 删除API-Key

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 删除API-Key | |
| 端点 | /api-keys/{id} | |
| 版本 | v1 | |
| method | DELETE | 删除资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### 请求示例
```shell
curl -X DELETE "http://api-server:port/open-api/v1/api-keys/apikey-001" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
Data为null。

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

**说明**
- 删除API-Key时，级联删除其专属的quota_plan、rate_limit_policy、route_rules及底层资源（如果这些资源不被其他API-Key或Entity引用）
- 删除API-Key可能会影响正在处理中的请求

---

## 7 查询配额计划（含Balance）

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询API-Key的配额计划（含实时余额） | |
| 端点 | /api-keys/{id}/quota-plan | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/api-keys/apikey-001/quota-plan" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| unlimited | bool | 是否无限配额 | - |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 | - |
| quota | int64 | 配额总量 | - |
| unit | string | 配额单位 | - |
| reset_period | string | 配额重置周期 | - |
| balance | object | 余额状态（只读） | 包含used和remaining |
| balance.used | int64 | 已用量 | - |
| balance.remaining | int64 | 剩余量 | - |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly",
        "balance": {
            "used": 12345679,
            "remaining": 87654321
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 8 重置配额余额（QuotaBalance Reset）

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 重置API-Key的配额余额 | |
| 端点 | /api-keys/{id}/quota-plan/reset | |
| 版本 | v1 | |
| method | POST | 创建资源（执行重置操作） |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| quota | int64 | 重置后的配额总量 | N | 若传入则更新quota并同步重置balance；若不传则按当前quota重置 |
| reason | string | 重置原因 | N | 用于审计 |

#### 请求示例
```shell
curl -X POST "http://api-server:port/open-api/v1/api-keys/apikey-001/quota-plan/reset" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "quota": 100000000,
    "reason": "月度重置"
}
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | API-Key标识 | - |
| previous_quota | int64 | 重置前配额 | - |
| new_quota | int64 | 重置后配额 | - |
| balance | object | 余额变更详情 | - |
| balance.previous_remaining | int64 | 重置前剩余量 | - |
| balance.new_remaining | int64 | 重置后剩余量 | - |
| balance.used | int64 | 当前已用量 | 重置后为0 |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "apikey-001",
        "previous_quota": 100000000,
        "new_quota": 100000000,
        "balance": {
            "previous_remaining": 87654321,
            "new_remaining": 100000000,
            "used": 0
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 数据模型

### Quota Plan 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| unlimited | bool | 是否无限配额，默认true |
| pass_when_no_enough_quota | bool | 配额不足时是否放行，默认false |
| quota | int64 | 配额总量（初始配额） |
| unit | string | 配额单位，默认total_token，暂时只支持total_token |
| reset_period | string | 配额重置周期，取值：never、weekly、monthly，默认never；重置均基于日历周期（如自然周/自然月） |
| balance | object | 余额状态（只读），包含used和remaining；仅在独立quota-plan查询接口返回 |

### Rate Limit Policy 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| enabled | bool | 是否启用，默认false |
| rules | object | 限流规则 |

### Rules 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| tpm | []TPMConfig | Token每分钟限制配置，最多3个；为空不做tpm限制 |
| rpm | []RPMConfig | 请求每分钟限制配置，最多3个；为空不做rpm限制 |
| max_concurrency | int | 最大并发数，最小值1；默认值为-1（不限制） |

### TPMConfig 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| name | string | 规则名称，同一policy中不能重复 |
| model | string | 适用的模型，选填，默认值为"*"（表示适用于所有模型） |
| window_minutes | int | 统计时间窗口（分钟），取值范围1-360 |
| max_tokens | int | 最大Token数 |
| step_minutes | int | 滑动步长（分钟），取值范围1-360，默认1，必须<=window_minutes |

### RPMConfig 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| name | string | 规则名称，同一policy中不能重复 |
| model | string | 适用的模型，选填，默认值为"*"（表示适用于所有模型） |
| window_minutes | int | 统计时间窗口（分钟），取值范围1-360 |
| max_requests | int | 最大请求数 |

### Route Rules 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| enabled | bool | 是否启用该路由规则，默认false |
| rules | []AiRouteRule | 规则列表；为空表示未配置任何规则 |

### AiRouteRule 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| name | string | 规则名称，同一路由表内唯一 |
| Cond | string | BFE条件表达式，命中则使用该规则 |
| targets | []AiRouteTarget | 转发目标列表，至少1个元素 |
| fallbacks | []AiRouteFallback | 降级目标列表，允许为空 |

### AiRouteTarget 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传原始模型 |
| Weight | int | 权重，单个target时为100，多个target时总和为100 |

### AiRouteFallback 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传原始模型 |

### 约束
- 若 `rate_limit_policy` 的 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 每个 `AiRouteRule` 的 `targets` 中所有 `Weight` 之和必须等于100
- 每个 `AiRouteRule` 的 `fallbacks` 中的元素 `ClusterName` 不能为空
