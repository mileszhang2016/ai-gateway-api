# Infinity AI Gateway - OpenAPI接口定义

**版本号**：v0.3.0

---

- [1. 通用说明](#1-通用说明)
- [2. /api-keys](#2-api-keys)
- [3. /entity-types](#3-entity-types)
- [4. /entities](#4-entities)
- [5. /global-route-rules](#5-global-route-rules)
- [6. /route-tables](#6-route-tables)
- [7. /alb-pool](#7-alb-pool)
- [8. /auth](#8-auth)
- [9. /certificates](#9-certificates)
- [10. /clusters](#10-clusters)
- [11. /model-provider-types](#11-model-provider-types)
- [12. /tools](#12-tools)
- [13. 关键业务流程](#13-关键业务流程)
- [14. 对象关系图](#14-对象关系图)
- [15. 版本修改记录](#15-版本修改记录)

---

## 1. 通用说明

### 基本信息

| 项目 | 值 | 说明 |
| - | - | - |
| URL格式 | http://api_server:port/open-api/{ver}/{endpoint}?{arg=value} | 例：http://127.1:8086/open-api/v1/api-keys |
| 版本 | v1 | - |
| 鉴权方式 | Token | HTTP Authorization Header |

### 返回值格式

所有API的返回值格式：

```json
{
    "ErrNum": 200,
    "Data": json_object,
    "ErrMsg": "string message"
}
```

- **ErrNum**: 返回码
  - 200：调用成功
  - 401：鉴权失败
  - 402：没有调用权限造成的失败
  - 404：查询/修改/删除不存在的对象时
  - 409：资源依赖冲突时
  - 422：参数不合法造成的失败
  - 500：其他业务逻辑错误，一律返回500
  - 510：集群/分流规则创建时实例池未ready
  - 555：创建重复对象时
  - 556：数据重复时
- **Data**: 返回的数据结构，调用成功时返回json格式数据，失败时返回null
- **ErrMsg**: 文本消息，成功时为"success"或空串，失败时为错误信息

### Method约定

| Method | 含义 |
| - | - |
| GET | 获取一条或多条 |
| POST | 创建 |
| PUT | 全量更新 |
| PATCH | 部分更新 |
| DELETE | 删除 |

### 通用Query参数（列表接口）

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| sort_by | string | 排序字段 | N | - |
| sort_order | string | 排序方向 | N | asc/desc，默认desc |


---

## 2. /api-keys

### 2.1 数据模型

```json
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
    "reset_period": "monthly",
    "balance": {
      "used": 50000000,
      "remaining": 50000000
    }
  },
  "rate_limit_policy": {
    "enabled": false,
    "rules": {
      "tpm": [
        {"name": "1分钟窗口", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1}
      ],
      "rpm": [
        {"name": "1分钟请求", "model": "*", "window_minutes": 1, "max_requests": 100}
      ],
      "max_concurrency": -1
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
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `id` | string | API-Key唯一标识（内部使用） | 系统生成，如`apikey-001` |
| `key` | string | API-Key值（用于请求头鉴权） | 系统生成，如`ak-2v8x9k3m7p` |
| `description` | string | 描述 | 自定义 |
| `enabled` | bool | 是否启用 | `true`（启用）、`false`（禁用） |
| `create_time` | int64 | 创建时间 | Unix时间戳（秒） |
| `update_time` | int64 | 更新时间 | Unix时间戳（秒） |
| `expired_time` | int64 | 过期时间 | `-1`表示永不过期；其他为Unix时间戳（秒） |
| `unlimited_quota` | bool | 是否无限配额 | `true`：不执行配额检查；`false`：执行配额控制；**默认值为false** |
| `models` | []string | 允许访问的模型白名单 | 包含"*"表示不限制，**默认值为不限制** |
| `subnet` | []string | 允许的客户端子网 | 包含"*"表示不限制，**默认值为不限制** |
| `quota_plan` | object | 配额计划 | 见下方结构，**不会为空** |
| `rate_limit_policy` | object | 限流策略 | 见下方结构，**不会为空** |
| `route_rules` | object | 路由规则 | 见下方结构，**不会为空** |
| `entity_id` | string | 挂载到的Entity ID | 为空表示未挂载到任何Entity |
| `entity` | object | 挂载的Entity摘要（只读） | 包含id、name、type |

**quota_plan结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `unlimited` | bool | 是否无限配额，**默认true** |
| `pass_when_no_enough_quota` | bool | 配额不足时是否放行，**默认false** |
| `quota` | int64 | 配额总量（初始配额） |
| `unit` | string | 配额单位，默认`total_token`，暂时只支持`total_token` |
| `reset_period` | string | 配额重置周期，取值：`never`（永不自动重置）、`weekly`（按周重置）、`monthly`（按月重置），**默认`never`**；重置均基于日历周期（如自然周/自然月） |
| `balance` | object | 余额状态（只读），包含`used`和`remaining` |

**rate_limit_policy结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `enabled` | bool | 是否启用，**默认false** |
| `rules` | object | 限流规则，见下方 |

**rules结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `tpm` | []TPMConfig | Token每分钟限制配置，最多3个；为空不做tpm限制 |
| `rpm` | []RPMConfig | 请求每分钟限制配置，最多3个；为空不做rpm限制 |
| `max_concurrency` | int | 最大并发数，最小值1；**默认值为-1（不限制）** |

**TPMConfig结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 规则名称，同一policy中不能重复 |
| `model` | string | 适用的模型，选填，默认值为"*"（表示适用于所有模型） |
| `window_minutes` | int | 统计时间窗口（分钟），取值范围1-360 |
| `max_tokens` | int | 最大Token数 |
| `step_minutes` | int | 滑动步长（分钟），取值范围1-360，默认1，必须`<=window_minutes` |

**RPMConfig结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 规则名称，同一policy中不能重复 |
| `model` | string | 适用的模型，选填，默认值为"*"（表示适用于所有模型） |
| `window_minutes` | int | 统计时间窗口（分钟），取值范围1-360 |
| `max_requests` | int | 最大请求数 |

**route_rules结构**

| 字段 | 类型 | 说明 |
|------|------|------|
| `enabled` | bool | 是否启用该路由规则，**默认false** |
| `rules` | array | 规则列表；为空表示未配置任何规则 |

- `rules` 元素结构、`targets` 元素结构、`fallbacks` 元素结构同 [5.1 数据模型](#51-数据模型)。

**约束**

- 若 `rate_limit_policy` 的 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一

---

### 2.2 接口清单

#### 2.2.1 创建API-Key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建API-Key | - |
| 端点 | /api-keys | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| key | string | API-Key值 | N | 可选。若传入则使用该值作为API-Key；若不传则由后台生成。用于从其他系统导入API-Key |
| description | string | API-Key描述 | Y | - |
| expired_time | int64 | 过期时间 | N | -1表示永不过期；其他为Unix时间戳（秒） |
| enabled | bool | 是否启用 | N | 默认true |
| unlimited_quota | bool | 是否无限配额 | N | 默认false |
| models | []string | 允许访问的模型白名单 | N | 包含"*"表示不限制，**默认值为不限制** |
| subnet | []string | 允许的客户端子网 | N | 包含"*"表示不限制，**默认值为不限制** |
| quota_plan | object | 配额计划 | N | 同2.1中quota_plan结构（不含balance），若未设置则使用默认值 |
| rate_limit_policy | object | 限流策略 | N | 同2.1中rate_limit_policy结构，若未设置则使用默认值（enabled=false） |
| route_rules | object | 路由规则 | N | 同2.1中route_rules结构，若未设置则使用默认值（enabled=false, rules为空） |
| entity_id | string | 挂载的Entity ID | N | 为空表示不挂载 |

**约束**

- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若 `entity_id` 不为空，该Entity必须存在
- 若传入 `key`，其值需在系统中全局唯一；若与已有API-Key的`key`重复，返回422

**HTTP BODY参数示例**

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

**执行逻辑**

1. 校验参数合法性
2. 若未传入 `quota_plan`，使用默认值（unlimited=true, pass_when_no_enough_quota=false, quota=0, unit=total_token, reset_period=never）
3. 若未传入 `rate_limit_policy`，使用默认值（enabled=false, rules为空）
4. 若未传入 `route_rules`，使用默认值（enabled=false, rules为空）
5. 若传入 `quota_plan`，创建对应的QuotaBalance（remaining = quota，used = 0）
6. 若传入 `rate_limit_policy`，创建RateLimitPolicy
7. 若传入 `route_rules`，创建路由规则配置
8. 若传入 `key`，使用输入的 `key` 作为API-Key；否则在后台生成新的API-Key，并绑定上述资源
9. 返回结果，含完整的嵌套结构（不含balance）

**返回数据（Data内容）**

字段同2.1数据模型（不含balance）。

**成功返回示例**

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
    }
}
```

---

#### 2.2.2 查询API-Key列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询API-Key列表 | - |
| 端点 | /api-keys | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| enabled | bool | 是否启用过滤 | N | - |
| entity_id | string | 按挂载的Entity ID过滤 | N | - |
| unlimited_quota | bool | 是否无限配额过滤 | N | - |


**返回数据（Data内容）**

Data为数组，元素字段同2.1数据模型，`quota_plan`中包含`balance`字段（含`used`和`remaining`）。

---

#### 2.2.3 查询单个API-Key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询单个API-Key | - |
| 端点 | /api-keys/{id} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

**返回数据（Data内容）**

字段同2.1数据模型，`quota_plan`中包含`balance`字段（含`used`和`remaining`）。

---

#### 2.2.4 全量更新API-Key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新API-Key | - |
| 端点 | /api-keys/{id} | - |
| 版本 | v1 | - |
| method | PUT | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

**输入参数（Body）**

同2.2.1创建API-Key的Body参数。

注：`key` 字段仅在创建时生效，更新时会被忽略。

**约束**

- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan

**执行逻辑**

1. 若传入新的 `quota_plan`：
   - 若 `quota_plan.unlimited` 为 `false`，且 `quota` 的值发生了改变，则对 balance 进行重置（remaining = 新的quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `true` 改为 `false`，则创建新的 balance（remaining = quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `false` 改为 `true`，则删除对应的 balance
2. 若传入新的 `rate_limit_policy` 且 `enabled` 为 `true`：
   - 对于 `tpm` 和 `rpm` 规则：
     - 若规则的 `name` 与之前的规则相同，但其他参数（如 `window_minutes`、`max_tokens`、`max_requests`、`step_minutes`）发生变化，则重置该规则对应的计数器
     - 若规则的 `name` 是新的（之前不存在），则创建新的计数器
     - 若之前存在的规则 `name` 在新传入的规则中消失了，则删除该规则对应的计数器
   - 若 `enabled` 由 `true` 改为 `false`，则保留计数器数据但不执行限流检查（下次启用时继续使用）
   - 若 `enabled` 由 `false` 改为 `true`，则根据当前规则创建/恢复计数器
3. 若传入新的 `route_rules`，替换对应的路由规则配置
4. 更新API-Key其他字段

**返回数据（Data内容）**

同2.1数据模型（不含balance）。

---

#### 2.2.5 部分更新API-Key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 部分更新API-Key | - |
| 端点 | /api-keys/{id} | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

**输入参数（Body）**

同2.2.1创建API-Key的Body参数，仅传需修改字段。

注：`key` 字段仅在创建时生效，更新时会被忽略。

**约束**

- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若将 `entity_id` 修改为非空（挂载到新Entity），且 `unlimited_quota` 为 `false` 且 `quota_plan.unlimited` 为 `false`，则要求新Entity或其祖先链上至少存在一个有效的Quota Plan
- 若修改 `quota_plan.quota`，视为重置配额（同步更新balance.remaining和used）
- 若修改 `route_rules`，视为全量替换该路由规则配置

**返回数据（Data内容）**

同2.1数据模型（不含balance）。

---

#### 2.2.6 删除API-Key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除API-Key | - |
| 端点 | /api-keys/{id} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

**返回数据（Data内容）**

Data为null。

**说明**

- 删除API-Key时，级联删除其专属的quota_plan、rate_limit_policy、route_rules及底层资源（如果这些资源不被其他API-Key或Entity引用）
- 删除API-Key可能会影响正在处理中的请求

---

#### 2.2.7 查询配额计划（含Balance）

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询API-Key的配额计划（含实时余额） | - |
| 端点 | /api-keys/{id}/quota-plan | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| key | string | API-Key标识 | Y | - |

**返回数据（Data内容）**

完整的quota_plan对象，包含`balance`字段（含`used`和`remaining`）。

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| unlimited | bool | 是否无限配额 | - |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 | - |
| quota | int64 | 配额总量 | - |
| unit | string | 配额单位 | - |
| reset_period | string | 配额重置周期 | - |
| balance | object | 余额状态（只读） | 包含used和remaining |

**balance结构**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| used | int64 | 已用量 | - |
| remaining | int64 | 剩余量 | - |

**成功返回示例**

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
    }
}
```

---

#### 2.2.8 重置配额余额（QuotaBalance Reset）

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 重置API-Key的配额余额 | - |
| 端点 | /api-keys/{id}/quota-plan/reset | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | API-Key标识 | Y | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| quota | int64 | 重置后的配额总量 | N | 若传入则更新quota并同步重置balance；若不传则按当前quota重置 |
| reason | string | 重置原因 | N | 用于审计 |

**执行逻辑**

1. 找到该API-Key的quota_plan（如果不存在或unlimited=true，返回404）
2. 若传入quota，更新quota_plan.quota为新的值
3. 触发balance的reset：
   - balance.remaining = 当前quota（或新的quota）
   - balance.used = 0

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | API-Key标识 | - |
| previous_quota | int64 | 重置前配额 | - |
| new_quota | int64 | 重置后配额 | - |
| balance | object | 余额变更详情 | 见下方 |

**balance结构**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| previous_remaining | int64 | 重置前剩余量 | - |
| new_remaining | int64 | 重置后剩余量 | - |
| used | int64 | 当前已用量 | 重置后为0 |

**成功返回示例**

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
    }
}
```

---

## 3. /entity-types

### 3.1 数据模型

```json
{
  "type_name": "dep",
  "description": "一级部门，用于组织架构中的部门层级",
  "level": 1,
  "create_time": 1716883200
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `type_name` | string | 类型标识 | 全局唯一，1-32字符，仅含小写字母、数字、下划线、连字符 |
| `description` | string | 类型描述 | 自定义 |
| `level` | int | 层级级别 | 取值范围1-5，数字越小级别越高 |
| `create_time` | int64 | 创建时间 | Unix时间戳（秒） |

---

### 3.2 接口清单

#### 3.2.1 创建Entity-Type

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建Entity类型定义 | - |
| 端点 | /entity-types | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | 全局唯一，1-32字符，仅含小写字母、数字、下划线、连字符 |
| description | string | 类型描述 | N | - |
| level | int | 层级级别 | Y | 取值范围1-5，数字越小级别越高 |

**HTTP BODY参数示例**

```json
{
    "type_name": "dep",
    "description": "一级部门",
    "level": 1
}
```

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| type_name | string | 类型名 | - |
| description | string | 描述 | - |
| level | int | 层级级别 | - |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |

---

#### 3.2.2 查询Entity-Type列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询Entity-Type列表 | - |
| 端点 | /entity-types | - |
| 版本 | v1 | - |
| method | GET | - |

**返回数据（Data内容）**

Data为数组，元素字段同3.2.1创建Entity-Type返回数据。

---

#### 3.2.3 查询单个Entity-Type

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询单个Entity-Type | - |
| 端点 | /entity-types/{type_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

**返回数据（Data内容）**

同3.2.1创建Entity-Type返回数据。

---

#### 3.2.4 更新Entity-Type

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 更新Entity-Type描述 | - |
| 端点 | /entity-types/{type_name} | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| description | string | 类型描述 | N | - |

**返回数据（Data内容）**

同3.2.1创建Entity-Type返回数据。

---

#### 3.2.5 删除Entity-Type

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除Entity-Type | - |
| 端点 | /entity-types/{type_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

**返回数据（Data内容）**

Data为null。

**约束**：若该Entity-Type已被任何Entity引用，返回ErrNum=409。

---

## 4. /entities

### 4.1 数据模型

```json
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
    "enabled": false,
    "rules": {
      "tpm": [
        {"name": "1分钟窗口", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1}
      ],
      "rpm": [
        {"name": "1分钟请求", "model": "*", "window_minutes": 1, "max_requests": 100}
      ],
      "max_concurrency": -1
    }
  },
  "route_rules": {
    "enabled": false,
    "rules": []
  },
  "create_time": 1716883200,
  "update_time": 1716883200
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `id` | string | Entity唯一标识 | 系统生成，如`ent-001` |
| `name` | string | Entity名称 | 在全局范围内唯一 |
| `type` | string | Entity类型 | 必须引用已定义的Entity-Type |
| `parent_id` | string | 父Entity ID | 为空表示根节点 |
| `allow_models` | []string | 允许访问的模型白名单 | 包含"*"表示允许访问所有模型，**默认值为允许访问所有模型** |
| `block_models` | []string | 禁止访问的模型黑名单 | 包含"*"表示禁止访问所有模型，**默认值为空数组**；若某模型同时出现在`allow_models`和`block_models`中，以`block_models`为准 |
| `quota_plan` | object | 对该Entity设置的配额计划 | 同2.1中quota_plan结构，**不会为空** |
| `rate_limit_policy` | object | 对该Entity设置的限流策略 | 同2.1中rate_limit_policy结构，**不会为空** |
| `route_rules` | object | 对该Entity设置的路由规则 | 同2.1中route_rules结构，**不会为空** |
| `create_time` | int64 | 创建时间 | Unix时间戳（秒） |
| `update_time` | int64 | 更新时间 | Unix时间戳（秒） |

---

### 4.2 接口清单

#### 4.2.1 创建Entity

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建Entity | - |
| 端点 | /entities | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| name | string | Entity名称 | Y | 全局唯一 |
| type | string | Entity类型 | Y | 必须引用已定义的Entity-Type |
| parent_id | string | 父Entity ID | N | 为空表示根节点 |
| allow_models | []string | 允许访问的模型白名单 | N | 包含"*"表示允许访问所有模型，**默认值为允许访问所有模型** |
| block_models | []string | 禁止访问的模型黑名单 | N | 包含"*"表示禁止访问所有模型，**默认值为空数组**；若某模型同时出现在`allow_models`和`block_models`中，以`block_models`为准 |
| quota_plan | object | 配额计划 | N | 同2.1中quota_plan结构（不含balance），若未设置则使用默认值 |
| rate_limit_policy | object | 限流策略 | N | 同2.1中rate_limit_policy结构，若未设置则使用默认值（enabled=false） |
| route_rules | object | 路由规则 | N | 同2.1中route_rules结构，若未设置则使用默认值（enabled=false, rules为空） |

**约束**

- `type` 必须引用系统中已存在的Entity-Type
- `parent_id` 若不为空，该父Entity对应的Entity-Type的 `level` 必须**小于**本Entity对应的Entity-Type的 `level`（数字越小级别越高，父节点级别必须更高）
- `name` 全局唯一
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一

**HTTP BODY参数示例**

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
    },
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "entity-default",
                "Cond": "default_t()",
                "targets": [
                    {
                        "ClusterName": "cluster_entity",
                        "Model": "",
                        "Weight": 100
                    }
                ],
                "fallbacks": []
            }
        ]
    }
}
```

**执行逻辑**

1. 校验参数合法性
2. 若未传入 `quota_plan`，使用默认值（unlimited=true, pass_when_no_enough_quota=false, quota=0, unit=total_token, reset_period=never）
3. 若未传入 `rate_limit_policy`，使用默认值（enabled=false, rules为空）
4. 若未传入 `route_rules`，使用默认值（enabled=false, rules为空）
5. 若传入 `quota_plan`，创建对应的QuotaBalance（remaining = quota，used = 0）
6. 若传入 `rate_limit_policy`，创建RateLimitPolicy
7. 若传入 `route_rules`，创建路由规则配置
8. 写入entity
9. 返回结果，含完整的嵌套结构（不含balance）

**返回数据（Data内容）**

字段同4.1数据模型，包含系统生成的id、create_time、update_time（不含balance）。

**成功返回示例**

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
        "route_rules": {
            "enabled": true,
            "rules": [
                {
                    "name": "entity-default",
                    "Cond": "default_t()",
                    "targets": [
                        {
                            "ClusterName": "cluster_entity",
                            "Model": "",
                            "Weight": 100
                        }
                    ],
                    "fallbacks": []
                }
            ]
        },
        "create_time": 1716883200,
        "update_time": 1716883200
    }
}
```

---

#### 4.2.2 查询Entity列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询Entity列表 | - |
| 端点 | /entities | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| type | string | 按类型过滤 | N | - |
| parent_id | string | 按父Entity过滤 | N | - |


**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| list | []object | Entity列表 | 包含完整Entity字段 |
| pagination | object | 分页信息 | - |
| pagination.page | int | 当前页码 | - |
| pagination.page_size | int | 每页条数 | - |
| pagination.total | int | 总条数 | - |

**list 对象字段说明**

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

**成功返回示例**

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
    }
}
```

---

#### 4.2.3 查询单个Entity

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询单个Entity | - |
| 端点 | /entities/{id} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**返回数据（Data内容）**

字段同4.1数据模型，但`quota_plan`中的`balance`字段不返回（仅调用独立quota-plan查询接口时返回）。

---

#### 4.2.4 全量更新Entity

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新Entity | - |
| 端点 | /entities/{id} | - |
| 版本 | v1 | - |
| method | PUT | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**输入参数（Body）**

同4.2.1创建Entity的Body参数。

**约束**

- `type` 不可修改（创建后固定）
- 若修改 `parent_id`，新父Entity对应的Entity-Type的 `level` 必须**小于**本Entity对应的Entity-Type的 `level`
- `name` 全局唯一，不可与其他Entity冲突
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一

**执行逻辑**

1. 若传入新的 `quota_plan`：
   - 若 `quota_plan.unlimited` 为 `false`，且 `quota` 的值发生了改变，则对 balance 进行重置（remaining = 新的quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `true` 改为 `false`，则创建新的 balance（remaining = quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `false` 改为 `true`，则删除对应的 balance
2. 若传入新的 `rate_limit_policy` 且 `enabled` 为 `true`：
   - 对于 `tpm` 和 `rpm` 规则：
     - 若规则的 `name` 与之前的规则相同，但其他参数发生变化，则重置该规则对应的计数器
     - 若规则的 `name` 是新的（之前不存在），则创建新的计数器
     - 若之前存在的规则 `name` 在新传入的规则中消失了，则删除该规则对应的计数器
   - 若 `enabled` 由 `true` 改为 `false`，则保留计数器数据但不执行限流检查
   - 若 `enabled` 由 `false` 改为 `true`，则根据当前规则创建/恢复计数器
3. 若传入新的 `route_rules`，替换对应的路由规则配置
4. 更新Entity其他字段

**返回数据（Data内容）**

同4.1数据模型（不含balance）。

---

#### 4.2.5 部分更新Entity

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 部分更新Entity | - |
| 端点 | /entities/{id} | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**输入参数（Body）**

同4.2.1创建Entity的Body参数，仅传需修改字段。

**约束**

- `type` 不可修改
- 若修改 `parent_id`，新父Entity对应的Entity-Type的 `level` 必须**小于**本Entity对应的Entity-Type的 `level`
- `name` 全局唯一，不可与其他Entity冲突
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若修改 `quota_plan.quota`，视为重置配额（同步更新balance.remaining和used）
- 若修改 `route_rules`，视为全量替换该路由规则配置

**执行逻辑**

1. 若传入新的 `quota_plan`：
   - 若 `quota_plan.unlimited` 为 `false`，且 `quota` 的值发生了改变，则对 balance 进行重置（remaining = 新的quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `true` 改为 `false`，则创建新的 balance（remaining = quota，used = 0）
   - 若 `quota_plan.unlimited` 由 `false` 改为 `true`，则删除对应的 balance
2. 若传入新的 `rate_limit_policy` 且 `enabled` 为 `true`：
   - 对于 `tpm` 和 `rpm` 规则：
     - 若规则的 `name` 与之前的规则相同，但其他参数发生变化，则重置该规则对应的计数器
     - 若规则的 `name` 是新的（之前不存在），则创建新的计数器
     - 若之前存在的规则 `name` 在新传入的规则中消失了，则删除该规则对应的计数器
   - 若 `enabled` 由 `true` 改为 `false`，则保留计数器数据但不执行限流检查
   - 若 `enabled` 由 `false` 改为 `true`，则根据当前规则创建/恢复计数器
3. 若传入新的 `route_rules`，替换对应的路由规则配置
4. 更新Entity其他字段

**返回数据（Data内容）**

同4.1数据模型（不含balance）。

---

#### 4.2.6 删除Entity

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除Entity | - |
| 端点 | /entities/{id} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**返回数据（Data内容）**

Data为null。

**约束**：
- 若该Entity存在子Entity（有其他Entity的parent_id指向它），返回ErrNum=409
- 若该Entity已被任何API-Key挂载，返回ErrNum=409

**说明**

- 删除Entity时，级联删除其专属的quota_plan、rate_limit_policy、route_rules及底层资源（如果这些资源不被其他API-Key或Entity引用）

---

#### 4.2.7 查询配额计划（含Balance）

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询Entity的配额计划（含实时余额） | - |
| 端点 | /entities/{id}/quota-plan | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**返回数据（Data内容）**

完整的quota_plan对象，包含`balance`字段（含`used`和`remaining`）。

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| unlimited | bool | 是否无限配额 | - |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 | - |
| quota | int64 | 配额总量 | - |
| unit | string | 配额单位 | - |
| reset_period | string | 配额重置周期 | - |
| balance | object | 余额状态（只读） | 包含used和remaining |

**balance结构**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| used | int64 | 已用量 | - |
| remaining | int64 | 剩余量 | - |

---

#### 4.2.8 重置配额余额（QuotaBalance Reset）

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 重置Entity的配额余额 | - |
| 端点 | /entities/{id}/quota-plan/reset | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| id | string | Entity标识 | Y | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| quota | int64 | 重置后的配额总量 | N | 若传入则更新quota并同步重置balance；若不传则按当前quota重置 |
| reason | string | 重置原因 | N | 用于审计 |

**执行逻辑**

1. 找到该Entity的quota_plan（如果不存在或unlimited=true，返回404）
2. 若传入quota，更新quota_plan.quota为新的值
3. 触发balance的reset：
   - balance.remaining = 当前quota（或新的quota）
   - balance.used = 0

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| id | string | Entity标识 | - |
| previous_quota | int64 | 重置前配额 | - |
| new_quota | int64 | 重置后配额 | - |
| balance | object | 余额变更详情 | 见下方 |

**balance结构**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| previous_remaining | int64 | 重置前剩余量 | - |
| new_remaining | int64 | 重置后剩余量 | - |
| used | int64 | 当前已用量 | 重置后为0 |

---
## 5. /global-route-rules

### 5.1 数据模型

```json
{
  "enabled": true,
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
      "fallbacks": [
        {
          "ClusterName": "cluster_global_fallback",
          "Model": ""
        }
      ]
    }
  ]
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `enabled` | bool | 是否启用该Global路由表 | `true`（启用）、`false`（禁用），**默认值为`true`** |
| `rules` | array | 规则列表，按顺序匹配 | - |

**rules元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 规则名称，用于日志和监控 | 同一`global`路由表内唯一 |
| `Cond` | string | BFE条件表达式 | 命中则使用该规则 |
| `targets` | array | 转发目标列表 | 至少1个元素 |
| `fallbacks` | array | 降级目标列表 | 允许为空 |

**targets元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `ClusterName` | string | 后端集群名称 | - |
| `Model` | string | 模型名称，空字符串表示透传原始模型 | - |
| `Weight` | int | 权重 | 单个target时为100，多个target时总和为100 |

**fallbacks元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `ClusterName` | string | 后端集群名称 | - |
| `Model` | string | 模型名称，空字符串表示透传原始模型 | - |

**约束**

- `rules` 中规则的 `name` 在同一 `global` 路由表内必须唯一
- 每个规则的 `targets` 中所有 `Weight` 之和必须等于100
- 每个规则的 `fallbacks` 中的元素 `ClusterName` 不能为空

---

### 5.2 接口清单

#### 5.2.1 更新Global路由表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新Global路由表 | - |
| 端点 | /global-route-rules | - |
| 版本 | v1 | - |
| method | PUT | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| enabled | bool | 是否启用该Global路由表 | N | **默认值为`true`** |
| rules | array | 规则列表 | Y | 同5.1数据模型中rules结构 |

**HTTP BODY参数示例**

```json
{
    "enabled": true,
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
            "fallbacks": [
                {
                    "ClusterName": "cluster_global_fallback",
                    "Model": ""
                }
            ]
        }
    ]
}
```

**执行逻辑**

1. 校验参数合法性
2. 若未传入 `enabled`，使用默认值 `true`
3. 校验 `rules` 中规则名称是否重复
4. 校验每个规则的 `targets` 权重总和是否为100
5. 校验每个规则的 `Cond` 表达式是否合法
6. 在持久化存储中，将 `type` 固定设置为 `global`，`owner` 固定设置为 `global`
7. 写入或替换 `global_default` 路由表
8. 返回结果

**返回数据（Data内容）**

字段同5.1数据模型。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "enabled": true,
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
                "fallbacks": [
                    {
                        "ClusterName": "cluster_global_fallback",
                        "Model": ""
                    }
                ]
            }
        ]
    }
}
```

---

#### 5.2.2 查询Global路由表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询Global路由表 | - |
| 端点 | /global-route-rules | - |
| 版本 | v1 | - |
| method | GET | - |

**返回数据（Data内容）**

字段同5.1数据模型。若Global路由表不存在，返回Data为null。

---

## 6. /route-tables

### 6.1 数据模型

```json
{
  "type": "global",
  "owner": "global",
  "enabled": true
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `type` | string | 路由表类型 | `global`（全局路由表）、`entity`（Entity路由表）、`api_key`（API-Key路由表） |
| `owner` | string | 所有者标识 | `global` 类型固定为 `global`；`entity` 类型为 `entity_id`；`api_key` 类型为 `apikey_id` |
| `enabled` | bool | 是否启用该路由表 | `true`（启用）、`false`（停用） |

---

### 6.2 接口清单

#### 6.2.1 查询路由表列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询路由表列表 | - |
| 端点 | /route-tables | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| sort_by | string | 排序字段 | N | - |
| sort_order | string | 排序方向 | N | asc/desc，默认desc |
| type | string | 按路由表类型过滤 | N | 可选值：`global`、`entity`、`api_key` |
| owner | string | 按所有者标识过滤 | N | 支持精确匹配 |
| enabled | bool | 按启用状态过滤 | N | - |

**返回数据（Data内容）**

Data为数组，元素字段同6.1数据模型。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "list": [
            {
                "type": "global",
                "owner": "global",
                "enabled": true
            },
            {
                "type": "entity",
                "owner": "ent-001",
                "enabled": false
            },
            {
                "type": "api_key",
                "owner": "apikey-001",
                "enabled": true
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 3
        }
    }
}
```

---


---

## 7. /alb-pool

### 7.1 数据模型

```json
{
  "name": "BFE.aipool",
  "instances": [
    {
      "hostname": "127.0.0.1",
      "ip": "127.0.0.1",
      "weight": 1,
      "ports": {
        "Default": 8080
      },
      "tags": {
        "key": "value"
      }
    }
  ]
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 实例池完整名称 | 如 `BFE.aipool` |
| `instances` | []Instance | 实例列表 | - |

**Instance 结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `hostname` | string | 实例所在主机名 | 无 DNS 时可填写 IP 地址 |
| `ip` | string | 实例 IP 地址 | - |
| `weight` | int | 实例权重 | 范围 [0,100] |
| `ports` | map[string]int | 实例端口 | 至少包含 `Default` 端口 |

**约束**

- AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置。
- 实例池名称由配置项 `DefaultAIInstancePoolName` 提供，请求中无需传入 `name`。

### 7.2 接口清单

#### 7.2.1 获取默认 AI 网关实例池详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取默认 AI 网关实例池的详情 | - |
| 端点 | /alb-pool | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**执行逻辑**

1. 从配置文件 `RunTime` 中读取 `DefaultAIInstancePoolName`（默认值：`BFE.aipool`）。
2. 查询实例池，若不存在则返回错误。
3. 返回实例池详情（包含实例列表）。

**返回数据（Data内容）**

字段同 [7.1 数据模型](#71-数据模型)。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "name": "BFE.aipool",
        "instances": [
            {
                "hostname": "127.0.0.1",
                "ip": "127.0.0.1",
                "weight": 1,
                "ports": {
                    "Default": 8080
                },
                "tags": {
                    "key": "value"
                }
            }
        ]
    }
}
```

#### 7.2.2 更新默认 AI 网关实例池

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新默认 AI 网关实例池的实例列表 | 该更新是全量更新，不支持仅添加部分数据 |
| 端点 | /alb-pool | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| instances | []Instance | 实例列表 | Y | 全量替换当前实例池中的实例 |
| instances[].hostname | string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instances[].ip | string | 实例 IP 地址 | Y | - |
| instances[].weight | int | 实例权重 | Y | 范围 [0,100] |
| instances[].ports | map[string]int | 实例端口 | Y | 至少包含 `Default` 端口 |
| instances[].tags | map[string]string | 实例标签 | N | - |

**约束**

- 实例池名称由配置项 `DefaultAIInstancePoolName` 提供，请求中无需传入 `name`。
- AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置。

**HTTP BODY参数示例**

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "key": "value"
            }
        }
    ]
}
```

**执行逻辑**

1. 校验请求参数合法性。
2. 使用配置项 `DefaultAIInstancePoolName` 定位默认 AI 网关实例池。
3. 全量替换实例池中的实例列表。
4. 返回更新后的实例池详情。

**返回数据（Data内容）**

同 [7.2.1 获取默认 AI 网关实例池详情](#721-获取默认-ai-网关实例池详情)。

**成功返回示例**

同 [7.2.1 获取默认 AI 网关实例池详情](#721-获取默认-ai-网关实例池详情)。

---

## 8. /auth

### 8.1 数据模型

**用户（User）**

```json
{
  "user_name": "user_demo",
  "is_admin": true
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `user_name` | string | 用户名 | - |
| `is_admin` | bool | 是否为系统管理员 | `true`：System 权限；`false` 暂不支持 |

**Token**

```json
{
  "name": "token_demo",
  "token": "Xim4h3tR_Gp7o4h",
  "scope": "System"
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | Token 名称 | 全局唯一 |
| `token` | string | Token 值 | 用于请求头鉴权 |
| `scope` | string | 权限范围 | `System` / `Support` |

**认证方式**

API 请求时需在 Header 的 `Authorization` 中携带凭证：
- Session Key：`Authorization: Session {session_key}`
- Token：`Authorization: Token {token}`

**Scope 划分**

| Scope | 说明 |
|------|------|
| `System` | 全部权限，包括全局配置、产品线资源和导出资源 |
| `Support` | 仅导出类资源，供 BFE 数据面模块导出配置 |

**约束**

- 普通用户和 Token 都会设定可访问资源的 Scope，只能访问 Scope 内资源。

### 8.2 接口清单

#### 8.2.1 创建用户

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建用户 | - |
| 端点 | /auth/users | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| user_name | string | 用户名 | Y | - |
| password | string | 用户密码 | Y | - |
| is_admin | bool | 是否为系统管理员 | Y | 固定为 `true`，暂不支持 `false` |

**HTTP BODY参数示例**

```json
{
    "user_name": "user_demo",
    "password": "password@baidu.com",
    "is_admin": true
}
```

**执行逻辑**

1. 校验参数合法性。
2. 创建用户记录，`is_admin` 固定为 `true`（System 权限），暂不支持其他 Scope。
3. 返回成功状态（Data 为 null）。

**返回数据（Data内容）**

无

#### 8.2.2 删除用户

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除用户 | - |
| 端点 | /auth/users/{user_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| user_name | string | 待删除的用户名 | Y | - |

**执行逻辑**

1. 根据 `user_name` 查找用户。
2. 删除用户记录。
3. 返回成功状态（Data 为 null）。

**返回数据（Data内容）**

无


#### 8.2.3 重置用户密码

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 重置用户密码 | - |
| 端点 | /auth/users/{user_name}/passwd | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| user_name | string | 待修改密码的用户名 | Y | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| old_password | string | 旧的用户密码 | N | 当被修改的用户为当前登录用户，需要填入旧密码 |
| password | string | 用户新密码 | Y | - |

**HTTP BODY参数示例**

```json
{
    "old_password": "manager2123@$",
    "password": "manager2123@$"
}
```

**执行逻辑**

1. 校验参数合法性
2. 校验用户存在性
3. 若修改当前登录用户密码，校验 `old_password` 正确性
4. 更新用户密码
5. 返回成功状态（Data 为 null）

**返回数据（Data内容）**

无

#### 8.2.4 获取用户列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查看用户列表 | - |
| 端点 | /auth/users | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无

**返回数据（Data内容）**

Data为数组，每个元素为一个用户。

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| user_name | string | 用户名 | - |
| is_admin | bool | 是否为系统管理员 | 固定返回 `true`，暂不支持 `false` |

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        {
            "user_name": "user_demo",
            "is_admin": true
        }
    ]
}
```

#### 8.2.5 设置用户是否具有管理员权限

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 设置用户是否有管理员权限 | - |
| 端点 | /auth/users/{user_name}/is_admin | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| user_name | string | 待修改权限的用户的用户名 | Y | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| is_admin | bool | 是否为系统管理员 | Y | 固定为 `true`，暂不支持设置为 `false` |

**HTTP BODY参数示例**

```json
{
    "is_admin": true
}
```

**执行逻辑**

1. 校验参数合法性
2. 根据 `user_name` 查找用户
3. 更新用户 `is_admin` 字段，固定为 `true`（System 权限），暂不支持其他 Scope
4. 返回成功状态（Data 为 null）

**返回数据（Data内容）**

无

#### 8.2.6 使用账号名密码创建session key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 使用账号密码得到session key（可用来登录） | - |
| 端点 | /auth/session-keys | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| user_name | string | 用户名 | Y | - |
| password | string | 用户密码 | Y | - |

**HTTP BODY参数示例**

```json
{
    "user_name": "manager2",
    "password": "manager2123@$"
}
```

**执行逻辑**

1. 校验用户名与密码正确性
2. 生成 session key
3. 返回 session key、用户名及管理员标识

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| session_key | string | 会话密钥 | 在后续请求中需要在Header中带上该值，格式为 "Authorization: Session iMQW0z5ZwK_6FnPPT7Xj" |
| user_name | string | 用户名 | - |
| is_admin | bool | 是否是系统管理员 | 固定返回 `true`（暂不支持非管理员用户） |

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "user_name": "user_demo",
        "session_key": "iMQW0z5ZwK_6FnPPT7Xj",
        "is_admin": true
    }
}
```

#### 8.2.7 删除 session key

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除 session key | - |
| 端点 | /auth/session-keys/{session_key} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| session_key | string | 待删除的session key | Y | - |

**执行逻辑**

1. 根据 `session_key` 查找会话
2. 删除 session key 记录
3. 返回成功状态（Data 为 null）

**返回数据（Data内容）**

无

#### 8.2.8 创建Token

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建Token | - |
| 端点 | /auth/tokens | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| name | string | token名字 | Y | name必须全局唯一 |
| scope | string | scope | Y | 只能指定一个scope；取值 `System` / `Support` |

**HTTP BODY参数示例**

```json
{
    "name": "token_demo",
    "scope": "System"
}
```

**执行逻辑**

1. 校验参数合法性
2. 校验 `name` 全局唯一性
3. 生成 Token 值
4. 返回 Token 值

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| token | string | Token值 | 在后续请求中需要在Header中带上该值，格式为 "Authorization: Token Px2szn6R1HQo-WRSIJyt" |

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "token": "Px2szn6R1HQo-WRSIJyt"
    }
}
```

#### 8.2.9 删除Token

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除token | - |
| 端点 | /auth/tokens/{token_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| token_name | string | 待删除的token name | Y | - |

**执行逻辑**

1. 根据 `token_name` 查找 Token
2. 删除 Token 记录
3. 返回成功状态（Data 为 null）

**返回数据（Data内容）**

无

#### 8.2.10 查看Token详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查看Token详情 | - |
| 端点 | /auth/tokens/{token_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| token_name | string | token name | Y | - |

**返回数据（Data内容）**

| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| name | string | token名字 | - |
| token | string | token的值 | - |
| scope | string | scope | - |

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "name": "token_demo",
        "token": "Xim4h3tR_Gp7o4h",
        "scope": "System"
    }
}
```

#### 8.2.11 查看Token列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查看Token列表 | - |
| 端点 | /auth/tokens | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无

**返回数据（Data内容）**

Data为数组，每个元素为Token（详见“查看Token详情”）。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        {
            "name": "token_demo",
            "token": "Xim4h3tR_Gp7o4h",
            "scope": "System"
        }
    ]
}
```

## 9. /certificates

### 9.1 数据模型

```json
{
  "cert_name": "cert_demo",
  "description": "abc",
  "is_default": true,
  "cert_file_name": "demo_cert_file_name",
  "cert_file_content": "-----BEGIN ...-----END CERTIFICATE-----",
  "key_file_name": "demo_key_file_name",
  "key_file_content": "-----BEGIN RSA PRIVATE KEY-----...-----END RSA PRIVATE KEY-----",
  "expired_date": "2021-08-23 16:02:31"
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `cert_name` | string | 证书名 | 必须唯一 |
| `description` | string | 证书描述 | - |
| `is_default` | bool | 是否是默认证书 | 全局必须有且只有一个默认证书 |
| `cert_file_name` | string | 主证书文件名 | - |
| `cert_file_content` | string | 主证书文件内容 | 创建/更新时必填；返回时不包含 |
| `key_file_name` | string | 主证书密钥名 | - |
| `key_file_content` | string | 主证书密钥文件内容 | 创建/更新时必填；返回时不包含 |
| `expired_date` | string | 主证书过期时间 | 如 `2021-08-23 16:02:31` |

**约束**

- 全局必须有且只有一个默认证书。
- 默认证书不能被删除。
- 更新为默认证书时，旧的默认证书自动变为非默认证书。

### 9.2 接口清单

#### 9.2.1 创建证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建证书 | - |
| 端点 | /certificates | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名 | Y | 必须唯一 |
| description | string | 证书描述 | Y | - |
| is_default | bool | 是否是默认证书 | Y | 全局必须有且只有一个默认证书 |
| cert_file_name | string | 主证书文件名 | Y | - |
| cert_file_content | string | 主证书文件内容 | Y | - |
| key_file_name | string | 主证书密钥名 | Y | - |
| key_file_content | string | 主证书密钥文件内容 | Y | - |
| expired_date | string | 主证书过期时间 | Y | - |

**约束**

- 全局必须有且只有一个默认证书。
- 若系统中已存在默认证书，新证书设置为默认时，旧默认证书自动变为非默认。

**HTTP BODY参数示例**

```json
{
    "cert_name": "cert_demo",
    "description": "abc",
    "is_default": true,
    "cert_file_name": "demo_cert_file_name",
    "cert_file_content": "-----BEGIN ...-----END CERTIFICATE-----",
    "key_file_name": "demo_key_file_name",
    "key_file_content": "-----BEGIN RSA PRIVATE KEY-----...-----END RSA PRIVATE KEY-----",
    "expired_date": "2021-08-23 16:02:31"
}
```

**执行逻辑**

1. 校验参数合法性。
2. 校验证书文件与密钥文件匹配。
3. 若新证书设置为默认，将旧默认证书更新为非默认。
4. 保存证书元数据及文件内容。
5. 返回证书元数据（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

| 参数名 | 类型 | 说明 |
|------|------|------|
| `cert_name` | string | 证书名 |
| `description` | string | 证书描述 |
| `is_default` | bool | 是否是默认证书 |
| `cert_file_name` | string | 主证书文件名 |
| `key_file_name` | string | 主证书密钥名 |
| `expired_date` | string | 主证书过期时间 |

> `cert_file_content`、`key_file_content` 不返回。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "cert_name": "cert_demo",
        "description": "abc",
        "is_default": true,
        "cert_file_name": "demo_cert_file_name",
        "key_file_name": "demo_key_file_name",
        "expired_date": "2021-08-23 16:02:31"
    }
}
```

#### 9.2.2 证书列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取全体证书信息列表 | - |
| 端点 | /certificates | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**执行逻辑**

1. 查询所有证书元数据。
2. 返回证书列表（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

Data 为数组，元素字段同 9.2.1 创建证书返回数据。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        {
            "cert_name": "cert_demo",
            "description": "abc",
            "is_default": true,
            "cert_file_name": "demo_cert_file_name",
            "key_file_name": "demo_key_file_name",
            "expired_date": "2021-08-23 16:02:31"
        }
    ]
}
```

#### 9.2.3 证书详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取单个证书信息 | - |
| 端点 | /certificates/{cert_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 返回证书元数据（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

同 9.2.1 创建证书返回数据。

**成功返回示例**

同 9.2.1 创建证书成功返回示例。

#### 9.2.4 更新证书为默认证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 更新证书为默认证书 | - |
| 端点 | /certificates/{cert_name}/default | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**约束**

- 更新为默认证书时，旧的默认证书自动变为非默认证书。

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 将当前默认证书更新为非默认。
3. 将目标证书更新为默认证书。
4. 返回目标证书元数据。

**返回数据（Data内容）**

同 9.2.1 创建证书返回数据。

**成功返回示例**

同 9.2.1 创建证书成功返回示例。

#### 9.2.5 删除证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除证书 | - |
| 端点 | /certificates/{cert_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**约束**

- 默认证书不能被删除。
- 全局必须始终保留一个默认证书。

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 校验该证书不是默认证书。
3. 删除证书记录及证书文件内容。
4. 返回成功状态（Data 为 null）。

**返回数据（Data内容）**

无。

---

## 10. /clusters

### 10.1 数据模型

```json
{
    "name": "my-cluster",
    "description": "示例集群",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {"Default": 8080}
        }
    ],
    "basic": {
        "protocol": "http",
        "connection": {
            "max_idle_conn_per_rs": 0,
            "cancel_on_client_close": false
        },
        "retries": {
            "max_retry_in_cluster": 2
        },
        "buffers": {
            "req_write_buffer_size": 512
        },
        "timeouts": {
            "timeout_conn_serv": 50000,
            "timeout_response_header": 50000,
            "timeout_readbody_client": 30000,
            "timeout_read_client_again": 30000,
            "timeout_write_client": 60000
        }
    },
    "sticky_sessions": {
        "enabled": false,
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 10,
        "host": "health.example.com",
        "uri": "/health",
        "statuscode": 200
    },
    "llm_config": {
        "model_endpoint": {
            "schema": "https",
            "uri": "/v1/models",
            "headers": {
                "Authorization": "Bearer ${API_KEY}"
            }
        },
        "models": ["deepseek-chat", "deepseek-coder"],
        "model_mappings": [
            {"key": "gpt-4", "value": "deepseek-chat"}
        ],
        "key": "sk-xxxxxxxxxxxx",
        "provider_type": "deepseek"
    }
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 集群名 | 全局唯一 |
| `description` | string | 集群描述信息 | - |
| `instance_pool` | []Instance | 实例列表 | 系统自动据此创建实例池和子集群 |
| `basic` | object | 基本参数 | 见下方 表：连接设置、表：重试设置、表：超时设置 |
| `sticky_sessions` | object | 会话保持 | 见下方 表：会话保持 |
| `passive_health_check` | object | 被动健康检查 | 见下方 表：被动健康检查 |
| `llm_config` | object | AI LLM 服务配置 | 见下方 表：LLM配置 |

**Instance 结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `hostname` | string | 实例所在主机名 | 无 DNS 时可填写 IP 地址 |
| `ip` | string | 实例 IP 地址 | - |
| `weight` | int | 实例权重 | 范围 [0,100] |
| `ports` | map[string]int | 实例端口 | 至少包含 `Default` 端口 |

**表：连接设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_idle_conn_per_rs| int | 连接池| N | 每个BFE实例，为集群中每个RS维持的空闲长连接数。一般情况下，无需特别维持，设置为0 。<br/>设置为非0时，可以提升转发性能 |
| cancel_on_client_close| bool |  连接是否级联关闭 | N | 设置为true时，当客户端关闭连接后，BFE同时关闭对应RS的连接 <br/>设置为false时，当客户端关闭连接后，BFE按默认策略关闭对应RS的连接 |

**表：重试设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_retry_in_cluster| int |  同一个集群内重试次数 | N | 底层对应 `max_retry_in_subcluster`，**默认值为2** |

**表：会话保持**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| enabled| bool |  是否开启会话保持 | N | **默认false**。为true时开启会话保持；为false时关闭 |
| hash_strategy| string |  会话保持策略  | N | CLIENT_IP_ONLY，根据client ip做会话保持 <br/>	CLIENT_ID_ONLY，根据请求中header做会话保持(默认值) <br>	CLIENT_ID_PERFERED，优先基于特定header，如果请求中没有对应header，则使用client ip|
| hash_header| string |  指定CLIENT_ID使用的header | N | 当使用cookie作为会话保持的哈希key时，数据格式为Cookie:${key} |

**表：超时设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| timeout_conn_serv| int |  连接后端超时(ms)| N |  |
| timeout_response_header| int |  读后端响应头部超时(ms)| N |  |
| timeout_readbody_client| int |  读请求body超时(ms)| N |  |
| timeout_read_client_again| int |  与用户的长连接超时(ms) | N |  |
| timeout_write_client| int |  写响应超时(ms)| N | - |

**表：被动健康检查**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| failnum| int |  进入健康检查的失败次数阈值 | N | 连续转发失败多次后，BFE进入健康检查状态，对下游RS发起探活；**默认值为3** |
| interval| int |  连续健康检查的时间间隔 | N | 单位ms；**默认值为1000** |
| host| string |  健康检查请求的域名| N | 域名后的部分；为空时使用 `instance_pool` 中第一个实例的 `hostname` |
| uri| string |  健康检查请求的URI  | N | **默认值为 `/`** |
| statuscode| int |  期望的健康检查返回码 | N | 如果需要忽略返回码，此处可以填0；**默认值为0** |

**表：LLM配置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| model_endpoint| object |  模型列表端点配置 | N | 用于调用第三方AI模型提供商的模型列表接口，具体字段见下方 表：Endpoint；未设置时使用默认值 |
| models| []string |  支持的模型名称列表 | Y | 指定该集群支持的AI模型名称 |
| model_mappings| []object |  模型名称映射 | N | 用于将用户请求的模型名映射为后端实际使用的模型名，具体字段见下方 表：模型映射 |
| key| string |  服务认证密钥 | N | 用于后端AI服务的认证 |
| provider_type| string |  AI模型提供商类型 | N | 取值如：deepseek、openai、qwen 等。数据来自 `/model-provider-types` |

> **注意：** `enable` 字段已移除，设置 `llm_config` 时默认开启 AI 网关能力。

**表：Endpoint**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| schema| string |  请求协议 | N |  取值为 http、https；**默认值为 https** |
| uri| string |  请求URI | N |  **默认值为 `/v1/models`** |
| headers| map[string]string |  请求头参数 | N | 自定义请求头 |

**表：模型映射**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| key| string |  用户请求的模型名 | Y |  |
| value| string |  映射后的实际模型名 | Y |  |

**约束**

- `sub_clusters` 与 `scheduler` 为系统内部自动生成数据，不再对外暴露；每个集群只包含一个子集群，调度设置固定为 `GSLB_BLACKHOLE=0`。
- `llm_config.enable` 字段已移除，设置 `llm_config` 时默认开启 AI 网关能力。
- 删除集群时自动级联清理关联的实例池和子集群。

### 10.2 接口清单

#### 10.2.1 创建集群

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建集群（一键创建实例池 + 子集群 + 绑定） | - |
| 端点 | /clusters | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| name| string |  集群名 | Y | 集群名必须全局唯一 |
| description| string |  集群描述信息| N |  |
| instance_pool| []Instance |  实例列表 | Y | 系统自动据此创建实例池和子集群 |
| instance_pool[].hostname| string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instance_pool[].ip| string | 实例 IP 地址 | Y | |
| instance_pool[].weight| int | 实例权重，范围 [0,100] | Y | |
| instance_pool[].ports| map[string]int | 实例端口 | Y | 至少包含 Default 端口 |
| basic| object |  基本参数| N | AI网关场景默认推荐值：connection.max_idle_conn_per_rs=0、cancel_on_client_close=false；retries.max_retry_in_cluster=2；timeouts.timeout_conn_serv=50000、timeout_response_header=50000、timeout_readbody_client=30000、timeout_read_client_again=30000、timeout_write_client=60000 |
| sticky_sessions| object |  会话保持| N | AI网关场景默认推荐值：enabled=false；若开启，hash_strategy=CLIENT_ID_ONLY、hash_header为空 |
| passive_health_check| object |  被动健康检查| N | AI网关场景默认推荐值：failnum=3、interval=1000ms、host为空（使用instance_pool首个实例hostname）、uri="/"、statuscode=0 |
| llm_config| object |  AI LLM服务配置| Y | 见 10.1 数据模型中表：LLM配置 |

**HTTP BODY参数示例**

```json
{
    "name": "my-cluster",
    "description": "示例集群",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {
                "Default": 8080
            }
        },
        {
            "hostname": "backend-2",
            "ip": "10.0.0.2",
            "weight": 50,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "basic": {
        "protocol": "http",
        "connection": {
            "max_idle_conn_per_rs": 0,
            "cancel_on_client_close": false
        },
        "retries": {
            "max_retry_in_cluster": 2
        },
        "buffers": {
            "req_write_buffer_size": 512
        },
        "timeouts": {
            "timeout_conn_serv": 50000,
            "timeout_response_header": 50000,
            "timeout_readbody_client": 30000,
            "timeout_read_client_again": 30000,
            "timeout_write_client": 60000
        }
    },
    "sticky_sessions": {
        "enabled": false,
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 10,
        "host": "health.example.com",
        "uri": "/health",
        "statuscode": 200
    },
    "llm_config": {
        "model_endpoint": {
            "schema": "https",
            "uri": "/v1/models",
            "headers": {
                "Authorization": "Bearer ${API_KEY}"
            }
        },
        "models": [
            "deepseek-chat",
            "deepseek-coder"
        ],
        "model_mappings": [
            {
                "key": "gpt-4",
                "value": "deepseek-chat"
            }
        ],
        "key": "sk-xxxxxxxxxxxx",
        "provider_type": "deepseek"
    }
}
```

**执行逻辑**

创建集群时，系统自动执行以下步骤：

1. 校验请求参数（`name`、`instance_pool`、`llm_config` 等必填项；`basic`、`sticky_sessions`、`passive_health_check` 若未传则使用 AI 网关场景默认推荐值）
2. 若 `llm_config` 不为 nil，内部自动设置 `enable = true`
3. 创建集群
4. 自动创建实例池（名称格式：`{product_name}.{cluster_name}`）
5. 自动创建子集群（名称：`{cluster_name}`，绑定实例池）
6. 自动绑定子集群到集群

**返回数据（Data内容）**

| 参数名 | 类型 |参数含义 |
| - | -  | - |
| name | string | 集群名 |
| description | string | 集群描述信息 |
| instance_pool | []Instance | 实例列表 |
| llm_config | object | LLM 配置 |
| basic | object | 基本参数 |
| sticky_sessions | object | 会话保持 |
| passive_health_check | object | 被动健康检查 |

**成功返回示例**
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "name": "my-cluster",
        "description": "示例集群",
        "instance_pool": [
            {
                "hostname": "backend-1",
                "ip": "10.0.0.1",
                "weight": 50,
                "ports": {"Default": 8080}
            }
        ],
        "llm_config": { "...": "..." },
        "basic": { "...": "..." },
        "sticky_sessions": { "...": "..." },
        "passive_health_check": { "...": "..." }
    }
}
```

#### 10.2.2 集群列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 所有集群列表 | - |
| 端点 | /clusters | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**返回数据（Data内容）**

数组，单元素同创建接口。

#### 10.2.3 集群详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 单个集群详情 | - |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**返回数据（Data内容）**

同创建接口。

#### 10.2.4 更新集群基本配置

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 更新集群基本信息 | 可编辑描述信息、Basic配置段、sticky_sessions配置段、healthcheck配置段、instance_pool、llm_config |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**输入参数（Body）**
可修改字段含义同创建接口。若传入 `instance_pool` 字段，系统会自动同步更新对应的实例池。

> **注意：** `sub_clusters` 与 `scheduler` 为系统内部自动生成，更新时不支持手动修改，请通过 `instance_pool` 调整实例。

**HTTP BODY参数示例**

```json
{
    "name": "my-cluster",
    "description": "更新后的集群描述",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "basic": {
        "protocol": "http",
        "connection": {
            "max_idle_conn_per_rs": 0,
            "cancel_on_client_close": false
        },
        "retries": {
            "max_retry_in_cluster": 2
        },
        "buffers": {
            "req_write_buffer_size": 512
        },
        "timeouts": {
            "timeout_conn_serv": 50000,
            "timeout_response_header": 50000,
            "timeout_readbody_client": 30000,
            "timeout_read_client_again": 30000,
            "timeout_write_client": 60000
        }
    },
    "sticky_sessions": {
        "enabled": false,
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 10,
        "host": "health.example.com",
        "uri": "/health",
        "statuscode": 200
    },
    "llm_config": {
        "model_endpoint": {
            "schema": "https",
            "uri": "/v1/models"
        },
        "models": ["deepseek-chat"],
        "provider_type": "deepseek"
    }
}
```

**返回数据（Data内容）**

同创建接口。

#### 10.2.5 删除集群

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除集群（自动级联清理关联的实例池和子集群） | - |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**执行逻辑**
删除集群时，系统自动执行以下级联清理：
1. 解绑集群关联的子集群
2. 删除子集群
3. 删除子集群关联的实例池
4. 删除集群

**返回数据（Data内容）**

同创建接口。

## 11. /model-provider-types

### 11.1 获取AI模型提供商类型列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取 AI 模型提供商类型列表 | - |
| 端点 | /model-provider-types | - |
| 版本 | v1 | - |
| method | GET | - |
| Content-Type | application/x-www-form-urlencoded | - |

**输入参数（Query）**
无

**HTTP BODY参数示例**
无

**返回数据（Data内容）**
字符串数组，元素为模型提供商类型标识。

**成功返回示例**
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        "deepseek",
        "qwen",
        "openai"
    ]
}
```

## 12. /tools

### 12.1 从指定提供商获取AI模型列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 从指定提供商获取AI模型列表 | - |
| 端点 | /tools/get-models-from-provider | - |
| 版本 | v1 | - |
| method | POST | - |
| Content-Type | application/json | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| schema | string | http、https | Y | - |
| uri | string | 请求的uri | N | 路径前面可以有/，也可以无/。例如：/models 或者 models |
| hosts | []string | 请求的 ip、port 组合或者域名 | Y | 支持 ipv4、ipv6。ipv4："1.1.1.1:8080" ipv6："[2001:db8::1]:8080" |
| headers | map[string]string | 请求的 header 参数列表 | N | - |
| provider_type | string | AI 模型提供商类型 | N | 取值为：deepseek，openai，qwen |

**HTTP BODY参数示例**

```json
{
    "schema":"http",
    "uri":"/models",
    "hosts":["1.1.1.1:8080", "[2001:db8::1]:8080","www.a.com","www.b.com:8080"],
    "headers":{
        "Content-type":"application/json"
    },
    "provider_type":"deepseek"
}
```

**返回数据（Data内容）**

状态码 200 为成功。返回数据为列表结构。

| 参数名 | 类型 | 参数含义 |
| - | - | - |
| id | string | 模型ID |
| name | string | 名称 |

**成功返回示例**
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        {
            "id": "model1",
            "name": "Model 1"
        }
    ]
}
```

## 13. 关键业务流程

**运行时执行顺序**：网关先执行模型访问控制检查流程，如果通过，再执行限流检查流程，如果通过，再执行配额扣减流程。

### 13.1 创建API-Key的完整流程
```text
1. 调用 POST /api-keys
   └─ 校验quota_plan合法性（若传入）
   └─ 校验rate_limit_policy合法性（若传入且enabled=true，rules至少配置一项）
   └─ 校验entity_id合法性（若不为空，entity存在）
   └─ 校验models/subnet格式
2. 若未传入quota_plan，使用默认值（unlimited=true, pass_when_no_enough_quota=false, quota=0, unit=total_token, reset_period=never）
3. 若未传入rate_limit_policy，使用默认值（enabled=false, rules为空）
4. 若传入quota_plan且unlimited=false：
   └─ 创建QuotaBalance（remaining = quota，used = 0）
5. 若传入rate_limit_policy：
   └─ 创建RateLimitPolicy
6. 生成key，写入持久化存储，绑定上述资源
7. 返回结果，含完整的嵌套结构（不含balance）
```

### 13.2 创建Entity的完整流程
```text
1. 调用 POST /entities
   └─ 校验type合法性（必须已定义）
   └─ 校验parent_id合法性（若不为空，父Entity的Entity-Type的level必须小于本Entity的Entity-Type的level）
   └─ 校验name全局唯一性
   └─ 校验quota_plan合法性（若传入）
   └─ 校验rate_limit_policy合法性（若传入且enabled=true，rules至少配置一项）
2. 若未传入quota_plan，使用默认值
3. 若未传入rate_limit_policy，使用默认值（enabled=false）
4. 若传入quota_plan且unlimited=false：
   └─ 创建QuotaPackage及对应QuotaBalance
5. 若传入rate_limit_policy：
   └─ 创建RateLimitPolicy
6. 写入entity
7. 返回结果，含完整的嵌套结构（不含balance）
```

### 13.3 运行时网关模型访问控制检查流程（数据面）
```text
1. 网关收到请求，解析API-Key和请求模型（如gpt-4）
2. 查询API-Key元数据（缓存），获取models字段
3. 检查API-Key的models字段：
   a. 若models包含"*"，通过API-Key自身检查
   b. 若models不包含"*"，检查请求模型是否在models列表中：
      - 若不在列表中，拒绝请求（403）
      - 若在列表中，通过
4. 若API-Key的entity_id不为空，检查Entity层级模型访问控制：
   a. 从该Entity开始，向上递归遍历所有祖先Entity（包含自身），构建检查链
   b. 对每个Entity依次检查：
      - block_models检查（黑名单优先）：
        - 若block_models包含"*"，拒绝请求（403）
        - 若请求模型在block_models中，拒绝请求（403）
      - allow_models检查（白名单）：
        - 若allow_models包含"*"，通过该Entity检查
        - 若请求模型不在allow_models中，拒绝请求（403）
        - 若请求模型在allow_models中，通过该Entity检查
   c. 检查链中任一Entity触发拒绝，请求立即拒绝（403）
5. 模型访问控制检查通过后，进入限流检查流程（7.4）
```

**说明**：
- API-Key的`models`为白名单机制，仅控制该API-Key自身可访问的模型范围
- Entity的`block_models`为黑名单机制，优先级高于`allow_models`；任一Entity层级（含祖先）的`block_models`命中，即拒绝请求
- Entity的`allow_models`为白名单机制，检查链中所有Entity的`allow_models`都必须包含请求模型（取交集逻辑）
- 若API-Key未挂载到任何Entity（entity_id为空），仅执行API-Key自身的models检查

---

### 13.4 运行时网关限流检查流程（数据面）
```text
1. 网关收到请求，解析API-Key和请求模型（如gpt-4）
2. 查询API-Key元数据（缓存），获取rate_limit_policy和entity_id
3. 构建Rate-Limit Policy检查列表（使用Set去重，相同policy_id只检查一次）：
   a. 若API-Key的rate_limit_policy.enabled=true，加入列表
   b. 若API-Key的entity_id不为空：
      - 查询该entity的rate_limit_policy，若enabled=true则加入
      - 递归查询该entity的所有祖先entity，将每个祖先的rate_limit_policy（若enabled=true）加入
4. 若检查列表为空，直接放行（不限流）
5. 按列表遍历每个policy：
   a. 获取Policy的rules，依次检查（仅检查与本次请求模型匹配的规则，model="*"或model等于请求模型时视为匹配）：
      - tpm: 采用滑动窗口机制，步长step_minutes，统计当前窗口内已用token数
        - 若已用token + 本次请求token > max_tokens，限流失败（429）
      - rpm: 采用固定窗口（计数器）机制，以window_minutes为固定周期
        - 检查当前窗口内已用请求数：
          - 若已用请求 + 1 > max_requests，限流失败（429）
          - 若已用请求 + 1 <= max_requests，计数器+1，限流通过
      - max_concurrency: 检查当前并发请求数
        - 若max_concurrency=-1，跳过并发检查
        - 若当前并发 >= max_concurrency，限流失败（429）
   b. 若当前Policy任一规则触发限流，标记失败，立即拒绝请求
6. 所有适用的Rate-Limit Policy必须全部通过，请求才进入配额扣减阶段（7.5）
   └─ 任一Policy触发限流，请求拒绝（429），不再执行配额扣减
```

### 13.5 运行时网关配额扣减流程（数据面）
```text
1. 限流检查（7.4）通过后，进入配额扣减阶段
2. 查询API-Key元数据（缓存），获取quota_plan、entity_id、unlimited_quota
3. 若unlimited_quota=true，直接放行
4. 构建Quota Plan扣减列表（使用Set去重，相同plan_id只扣减一次）：
   a. 若API-Key的quota_plan.unlimited=false，加入列表
   b. 若API-Key的entity_id不为空：
      - 查询该entity的quota_plan，若unlimited=false则加入
      - 递归查询该entity的所有祖先entity，将每个祖先的quota_plan（若unlimited=false）加入
5. 若扣减列表为空，扣减失败，请求拒绝（429002）
6. 按扣减列表遍历每个quota_plan：
   a. 对quota执行扣减：
      - 查询该quota对应的quota-balance
      - 检查quota-balance的remaining：
        - 若remaining == 0：扣减失败（无可用余额）
        - 若remaining >= needed：
          - 执行DECRBY balance.remaining by needed
          - 扣减成功
        - 若remaining < needed：
          - 若pass_when_no_enough_quota为true：
            - 执行DECRBY balance.remaining by remaining（扣减至0）
            - 扣减成功（已扣减部分保留）
          - 若pass_when_no_enough_quota为false：扣减失败（余额不足）
   b. 若当前Plan扣减失败，标记失败
7. 遍历完成后，若所有quota_plan的unlimited均为true（没有任何Plan被实际执行扣减），则扣减失败，请求拒绝（429002）
8. 所有实际执行扣减的Plan必须全部扣减成功，请求才通过
   └─ 任一实际执行扣减的Plan失败，触发所有已扣减balance的原子回滚（包含其他成功Plan的扣减），请求拒绝（429002）
```

### 13.6 配置变更的级联与隔离
| 变更操作 | 级联影响 | 隔离机制 |
|----------|----------|----------|
| 修改API-Key的quota_plan | 实时生效，影响后续请求；更新底层资源 | 旧资源若不被其他API-Key/Entity引用则级联删除 |
| 修改API-Key的rate_limit_policy | 实时生效，影响后续请求 | 旧资源若不被引用则级联删除 |
| 修改API-Key的route_rules | 实时生效，影响后续请求的路由选择 | 更新底层资源，旧资源若不被引用则级联删除 |
| 修改API-Key的models | 实时生效，影响后续请求的模型访问控制 | 更新API-Key元数据 |
| 修改API-Key的entity_id（挂载/解绑） | 实时生效，影响后续请求的层级限流、配额扣减、模型访问控制和路由选择 | 旧挂载关系立即解除，无残留 |
| 修改Entity的allow_models | 实时生效，影响所有挂载到该Entity及其后代的API-Key | 更新Entity元数据 |
| 修改Entity的block_models | 实时生效，影响所有挂载到该Entity及其后代的API-Key | 更新Entity元数据 |
| 修改Entity的parent_id | 实时生效，影响该Entity及其所有后代Entity挂载的API-Key | 禁止level违反层级关系的修改 |
| 修改Entity的quota_plan | 实时生效，影响所有挂载到该Entity及其后代的API-Key | 更新底层资源，旧资源若不被引用则级联删除 |
| 修改Entity的rate_limit_policy | 实时生效，影响所有挂载到该Entity及其后代的API-Key | 更新底层资源，旧资源若不被引用则级联删除 |
| 修改Entity的route_rules | 实时生效，影响所有挂载到该Entity及其后代的API-Key | 更新底层资源，旧资源若不被引用则级联删除 |
| 删除Entity | 必须先解绑所有API-Key且无子Entity | 有子Entity或挂载API-Key时禁止删除；级联删除其专属资源 |
| 删除API-Key | 级联删除其quota_plan、rate_limit_policy及底层资源（若不被其他引用） | 引用计数管理 |
| 修改quota_plan.quota | 同步更新balance.remaining和used | 视为重置配额 |

---
## 14. 对象关系图

```mermaid
classDiagram
    class APIKey {
        +string key
        +string description
        +bool enabled
        +int64 create_time
        +int64 update_time
        +int64 expired_time
        +bool unlimited_quota
        +[]string models
        +[]string subnet
        +QuotaPlan quota_plan
        +RateLimitPolicy rate_limit_policy
        +RouteRules route_rules
        +string entity_id
    }

    class Entity {
        +string id
        +string name
        +string type
        +string parent_id
        +[]string allow_models
        +[]string block_models
        +QuotaPlan quota_plan
        +RateLimitPolicy rate_limit_policy
        +RouteRules route_rules
        +int64 create_time
        +int64 update_time
    }

    class EntityType {
        +string type_name
        +string description
        +int level
        +int64 create_time
    }

    class QuotaPlan {
        +bool unlimited
        +bool pass_when_no_enough_quota
        +int64 quota
        +string unit
        +string reset_period
        +QuotaBalance balance
    }

    class QuotaBalance {
        +int64 used
        +int64 remaining
    }

    class RateLimitPolicy {
        +bool enabled
        +Rules rules
    }

    class Rules {
        +[]TPMConfig tpm
        +[]RPMConfig rpm
        +int max_concurrency
    }

    class TPMConfig {
        +string name
        +string model
        +int window_minutes
        +int max_tokens
        +int step_minutes
    }

    class RPMConfig {
        +string name
        +string model
        +int window_minutes
        +int max_requests
    }

    class RouteRules {
        +bool enabled
        +[]AiRouteRule rules
    }

    class GlobalRouteTable {
        +RouteRules route_rules
    }

    class AiRouteRule {
        +string name
        +string Cond
        +[]AiRouteTarget targets
        +[]AiRouteFallback fallbacks
    }

    class AiRouteTarget {
        +string ClusterName
        +string Model
        +int Weight
    }

    class AiRouteFallback {
        +string ClusterName
        +string Model
    }

    GlobalRouteTable "1" --> "1" RouteRules : route_rules
    RouteRules "1" --> "*" AiRouteRule : rules
    AiRouteRule "1" --> "*" AiRouteTarget : targets
    AiRouteRule "1" --> "*" AiRouteFallback : fallbacks

    APIKey "0..1" --> "0..1" Entity : entity_id挂载
    APIKey "1" --> "1" QuotaPlan : quota_plan
    APIKey "1" --> "1" RateLimitPolicy : rate_limit_policy
    APIKey "1" --> "1" RouteRules : route_rules
    Entity "0..1" --> "0..1" Entity : parent_id
    Entity "*" --> "1" EntityType : type
    Entity "1" --> "1" QuotaPlan : quota_plan
    Entity "1" --> "1" RateLimitPolicy : rate_limit_policy
    Entity "1" --> "1" RouteRules : route_rules
    QuotaPlan "1" --> "1" QuotaBalance : balance
    RateLimitPolicy "1" --> "1" Rules : rules
    Rules "1" --> "0..3" TPMConfig : tpm
    Rules "1" --> "0..3" RPMConfig : rpm
```

**关系说明**

- **API-Key → Entity**：一个API-Key可以挂载到0或1个Entity（通过entity_id）。一个Entity可以被多个API-Key挂载。
- **API-Key → QuotaPlan**：一个API-Key必须拥有1个QuotaPlan（通过嵌套的quota_plan）。若创建时未设置，使用默认值。QuotaPlan的生命周期与API-Key绑定。
- **API-Key → RateLimitPolicy**：一个API-Key必须拥有1个RateLimitPolicy（通过嵌套的rate_limit_policy）。若创建时未设置，使用默认值（enabled=false）。生命周期与API-Key绑定。
- **API-Key → RouteRules**：一个API-Key必须拥有1个RouteRules（通过嵌套的route_rules）。若创建时未设置，使用默认值（enabled=false, rules为空）。生命周期与API-Key绑定。
- **Entity → Entity（parent）**：一个Entity可以有0或1个父Entity，通过parent_id维护。一个Entity可以有多个子Entity。形成树形层级结构。父Entity的Entity-Type的 `level` 必须小于子Entity的Entity-Type的 `level`（数字越小级别越高）。
- **Entity → EntityType**：一个Entity必须属于一个EntityType。一个EntityType可以对应多个Entity。
- **Entity → QuotaPlan**：一个Entity必须拥有1个QuotaPlan（嵌套）。若创建时未设置，使用默认值。生命周期与Entity绑定。
- **Entity → RateLimitPolicy**：一个Entity必须拥有1个RateLimitPolicy（嵌套）。若创建时未设置，使用默认值（enabled=false）。生命周期与Entity绑定。
- **Entity → RouteRules**：一个Entity必须拥有1个RouteRules（嵌套）。若创建时未设置，使用默认值（enabled=false, rules为空）。生命周期与Entity绑定。
- **Entity模型访问控制**：`allow_models`和`block_models`共同决定该Entity及其挂载的API-Key可访问的模型范围。`block_models`优先级高于`allow_models`。
- **QuotaPlan → QuotaBalance**：一个QuotaPlan对应唯一的QuotaBalance（一一对应）。QuotaPlan为静态配置（unlimited、quota、unit、reset_period等），QuotaBalance为动态运行态（used、remaining）。运行时网关扣减操作直接作用于QuotaBalance的remaining。
- **QuotaPlan的balance字段**：为只读字段，实时反映对应QuotaBalance的used和remaining。仅在独立查询接口（/api-keys/{key}/quota-plan、/entities/{id}/quota-plan）中返回。
- **RateLimitPolicy → Rules**：一个RateLimitPolicy包含一组Rules，Rules中可配置tpm（最多3个）、rpm（最多3个）和max_concurrency（默认-1，表示不限制）。
- **运行时层级生效逻辑**：API-Key挂载到Entity后，运行时生效的QuotaPlan和RateLimitPolicy为该API-Key自身直接拥有的 + 该Entity直接拥有的 + 该Entity所有祖先Entity直接拥有的（去重）。
- **GlobalRouteTable → RouteRules → AiRouteRule → AiRouteTarget/AiRouteFallback**：Global路由表通过嵌套的 RouteRules 管理规则，每条规则包含转发目标（targets）和降级目标（fallbacks）。

