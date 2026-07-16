# Entity

## 1 创建Entity

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 创建Entity | |
| 端点 | /entities | |
| 版本 | v1 | |
| method | POST | 创建资源 |
| Content-Type | application/json | - |

### 输入参数

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| name | string | Entity名称 | Y | 全局唯一 |
| type | string | Entity类型 | Y | 必须引用已定义的Entity-Type |
| parent_id | string | 父Entity ID | N | 为空表示根节点 |
| allow_models | []string | 允许访问的模型白名单 | N | 包含"*"表示允许访问所有模型，默认值为允许访问所有模型 |
| block_models | []string | 禁止访问的模型黑名单 | N | 包含"*"表示不禁止任何模型，默认值为空数组；若某模型同时出现在allow_models和block_models中，以block_models为准 |
| quota_plan | object | 配额计划 | N | 同Quota Plan结构（不含balance），若未设置则使用默认值 |
| rate_limit_policy | object | 限流策略 | N | 同Rate Limit Policy结构，若未设置则使用默认值（enabled=false） |

**约束**
- `type` 必须引用系统中已存在的Entity-Type
- `parent_id` 若不为空，该父Entity对应的Entity-Type的 `level` 必须小于本Entity对应的Entity-Type的 `level`（数字越小级别越高，父节点级别必须更高）
- `name` 全局唯一
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一

#### 请求示例
```shell
curl -X POST "http://api-server:port/open-api/v1/entities" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "name": "op",
    "type": "dep",
    "parent_id": null,
    "allow_models": ["*"],
    "block_models": [],
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
    }
}
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | Entity唯一标识 | 系统生成 |
| name | string | Entity名称 | - |
| type | string | Entity类型 | - |
| parent_id | string | 父Entity ID | - |
| allow_models | []string | 允许访问的模型白名单 | - |
| block_models | []string | 禁止访问的模型黑名单 | - |
| quota_plan | object | 配额计划 | 不含balance字段 |
| rate_limit_policy | object | 限流策略 | - |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |
| update_time | int64 | 更新时间 | Unix时间戳（秒） |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "ent-001",
        "name": "op",
        "type": "dep",
        "parent_id": null,
        "allow_models": ["*"],
        "block_models": [],
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
        "create_time": 1716883200,
        "update_time": 1716883200
    },
    "WorkMode": "ModeNormal"
}
```

---

## 2 查询Entity列表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询Entity列表 | |
| 端点 | /entities | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### Query 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| type | string | 按类型过滤 | N | - |
| parent_id | string | 按父Entity过滤 | N | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/entities?page=1&page_size=20&type=dep" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| list | []object | Entity列表 | 包含完整Entity字段 |
| pagination | object | 分页信息 | - |
| pagination.page | int | 当前页码 | - |
| pagination.page_size | int | 每页条数 | - |
| pagination.total | int | 总条数 | - |

#### list 对象字段说明
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | Entity唯一标识 | - |
| name | string | Entity名称 | - |
| type | string | Entity类型 | - |
| parent_id | string | 父Entity ID | - |
| allow_models | []string | 允许访问的模型白名单 | - |
| block_models | []string | 禁止访问的模型黑名单 | - |
| quota_plan | object | 配额计划 | 不含balance字段 |
| rate_limit_policy | object | 限流策略 | - |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |
| update_time | int64 | 更新时间 | Unix时间戳（秒） |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "list": [
            {
                "id": "ent-001",
                "name": "op",
                "type": "dep",
                "parent_id": null,
                "allow_models": ["*"],
                "block_models": [],
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
                "create_time": 1716883200,
                "update_time": 1716883200
            },
            {
                "id": "ent-bfe-001",
                "name": "bfe",
                "type": "team",
                "parent_id": "ent-001",
                "allow_models": ["*"],
                "block_models": [],
                "quota_plan": {
                    "unlimited": true
                },
                "rate_limit_policy": {
                    "enabled": false
                },
                "create_time": 1716883200,
                "update_time": 1716883200
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 2
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 3 查询单个Entity

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询单个Entity | |
| 端点 | /entities/{id} | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/entities/ent-001" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
同创建接口返回，但不含balance字段。

#### 成功返回参数示例
同创建接口返回示例。

---

## 4 全量更新Entity

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 全量更新Entity | |
| 端点 | /entities/{id} | |
| 版本 | v1 | |
| method | PUT | 全量更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### Body 参数
同创建Entity的Body参数。

**约束**
- `type` 不可修改（创建后固定）
- 若修改 `parent_id`，新父Entity对应的Entity-Type的 `level` 必须小于本Entity对应的Entity-Type的 `level`
- `name` 全局唯一，不可与其他Entity冲突
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一

#### 请求示例
```shell
curl -X PUT "http://api-server:port/open-api/v1/entities/ent-001" \
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

## 5 部分更新Entity

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 部分更新Entity | |
| 端点 | /entities/{id} | |
| 版本 | v1 | |
| method | PATCH | 部分更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### Body 参数
同创建Entity的Body参数，仅传需修改字段。

**约束**
- `type` 不可修改
- 若修改 `parent_id`，新父Entity对应的Entity-Type的 `level` 必须小于本Entity对应的Entity-Type的 `level`
- `name` 全局唯一，不可与其他Entity冲突
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若修改 `quota_plan.quota`，视为重置配额（同步更新balance.remaining和used）

#### 请求示例
```shell
curl -X PATCH "http://api-server:port/open-api/v1/entities/ent-001" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下（仅修改name）：
```json
{
    "name": "updated-name"
}
```

### 返回数据(Data内容)
同创建接口返回（不含balance）。

#### 成功返回参数示例
同创建接口返回示例。

---

## 6 删除Entity

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 删除Entity | |
| 端点 | /entities/{id} | |
| 版本 | v1 | |
| method | DELETE | 删除资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### 请求示例
```shell
curl -X DELETE "http://api-server:port/open-api/v1/entities/ent-001" \
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

**约束**
- 若该Entity存在子Entity（有其他Entity的parent_id指向它），返回ErrNum=409
- 若该Entity已被任何API-Key挂载，返回ErrNum=409

**说明**
- 删除Entity时，级联删除其专属的quota_plan、rate_limit_policy及底层资源（如果这些资源不被其他API-Key或Entity引用）

---

## 7 查询配额计划（含Balance）

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询Entity的配额计划（含实时余额） | |
| 端点 | /entities/{id}/quota-plan | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/entities/ent-001/quota-plan" \
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
| 含义 | 重置Entity的配额余额 | |
| 端点 | /entities/{id}/quota-plan/reset | |
| 版本 | v1 | |
| method | POST | 创建资源（执行重置操作） |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| quota | int64 | 重置后的配额总量 | N | 若传入则更新quota并同步重置balance；若不传则按当前quota重置 |
| reason | string | 重置原因 | N | 用于审计 |

#### 请求示例
```shell
curl -X POST "http://api-server:port/open-api/v1/entities/ent-001/quota-plan/reset" \
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
| id | string | Entity标识 | - |
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
        "id": "ent-001",
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
| reset_period | string | 配额重置周期，取值：never、weekly、monthly，默认never |
| balance | object | 余额状态（只读），包含used和remaining |

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
| model | string | 适用的模型，默认值为"*"（表示适用于所有模型） |
| window_minutes | int | 统计时间窗口（分钟），取值范围1-360 |
| max_tokens | int | 最大Token数 |
| step_minutes | int | 滑动步长（分钟），取值范围1-360，默认1，必须<=window_minutes |

### RPMConfig 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| name | string | 规则名称，同一policy中不能重复 |
| model | string | 适用的模型，默认值为"*"（表示适用于所有模型） |
| window_minutes | int | 统计时间窗口（分钟），取值范围1-360 |
| max_requests | int | 最大请求数 |
