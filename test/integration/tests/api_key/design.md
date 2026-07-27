# API-Key 测试用例设计文档

## 1. 模块概述

API-Key 模块负责 API-Key 的管理，包括创建、查询、更新、删除 API-Key；管理 API-Key 的配额计划（QuotaPlan）、限流策略（RateLimitPolicy）、路由规则（RouteRules）及 Entity 挂载关系。v0.3.0 起支持导入外部 `key`，且列表/详情接口的 `quota_plan` 包含实时 `balance`；`PUT/PATCH` 更新时 `key` 字段忽略。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| AK-1 | 创建 API-Key | POST | `/open-api/v1/api-keys` | 支持传入外部 key |
| AK-2 | 查询 API-Key 列表 | GET | `/open-api/v1/api-keys` | 分页列表，quota_plan 含 balance |
| AK-3 | 查询单个 API-Key | GET | `/open-api/v1/api-keys/{id}` | 详情，quota_plan 含 balance |
| AK-4 | 全量更新 API-Key | PUT | `/open-api/v1/api-keys/{id}` | 全量替换，key 字段忽略 |
| AK-5 | 部分更新 API-Key | PATCH | `/open-api/v1/api-keys/{id}` | 部分更新，key 字段忽略 |
| AK-6 | 删除 API-Key | DELETE | `/open-api/v1/api-keys/{id}` | 级联删除专属资源 |
| AK-7 | 查询配额计划 | GET | `/open-api/v1/api-keys/{id}/quota-plan` | 返回完整 quota_plan 含 balance |
| AK-8 | 重置配额余额 | POST | `/open-api/v1/api-keys/{id}/quota-plan/reset` | 重置余额，可指定新 quota |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建 API-Key | 5 |
| 查询 API-Key 列表 | 3 |
| 查询单个 API-Key | 2 |
| 全量更新 API-Key | 3 |
| 部分更新 API-Key | 3 |
| 删除 API-Key | 2 |
| 查询配额计划 | 2 |
| 重置配额余额 | 2 |
| **合计** | **22** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
api_key/
├── design.md
├── create/
│   └── create_test.go
├── list/
│   └── list_test.go
├── detail/
│   └── detail_test.go
├── full_update/
│   └── full_update_test.go
├── partial_update/
│   └── partial_update_test.go
├── delete/
│   └── delete_test.go
├── quota_query/
│   └── quota_query_test.go
└── quota_reset/
    └── quota_reset_test.go
```

## 6. 创建 API-Key

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 创建 API-Key |
| 方法 | POST |
| 路径 | `/open-api/v1/api-keys` |
| 说明 | 创建新 API-Key，支持传入外部 key |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | N | API-Key 值，传入则使用该值 |
| description | string | Y | API-Key 描述 |
| expired_time | int64 | N | 过期时间，-1 表示永不过期 |
| enabled | bool | N | 默认 true |
| unlimited_quota | bool | N | 默认 false |
| models | []string | N | 默认 ["*"] |
| subnet | []string | N | 默认 ["*"] |
| quota_plan | object | N | 配额计划 |
| rate_limit_policy | object | N | 限流策略 |
| route_rules | object | N | 路由规则 |
| entity_id | string | N | 挂载的 Entity ID |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 唯一标识 |
| key | string | API-Key 值 |
| description | string | 描述 |
| enabled | bool | 是否启用 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |
| expired_time | int64 | 过期时间 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 允许访问的模型白名单 |
| subnet | []string | 允许的客户端子网 |
| quota_plan | object | 配额计划（不含 balance） |
| rate_limit_policy | object | 限流策略 |
| route_rules | object | 路由规则 |
| entity_id | string | 挂载的 Entity ID |
| entity | object | 挂载的 Entity 摘要 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-1-001 | 最小参数创建 | 正常参数 | 验证默认值 |
| AK-1-002 | 完整参数创建 | 正常参数 | 验证嵌套结构 |
| AK-1-003 | 导入外部 key | 正常参数 | v0.3.0 新增能力 |
| AK-1-004 | 导入重复 key | 异常参数 | v0.3.0 全局唯一校验 |
| AK-1-005 | 缺少 description | 必填校验 | 验证 ErrNum=422 |

### 6.4 测试场景详细设计

#### 6.4.1 AK-1-001：最小参数创建（正常参数）

##### 设计思路

验证仅传 `description` 时，默认值正确填充。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/api-keys`。
2. 验证返回结构和默认值。

##### 请求参数

```json
{
    "description": "test-key-min"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| key | 非空字符串 | NotEmpty |
| description | "test-key-min" | Equals |
| enabled | true | Equals |
| models | ["*"] | Equals |
| subnet | ["*"] | Equals |
| quota_plan | 非空对象 | IsObject |
| quota_plan.balance | 不存在 | NotExists |

---

#### 6.4.2 AK-1-002：完整参数创建（正常参数）

##### 设计思路

验证传入完整嵌套结构时，返回结构与输入一致，且 `entity` 摘要非空。

##### 前提数据准备

已创建有效 Entity 供挂载。

##### 执行步骤

1. 发送 POST 请求，传入完整参数。
2. 验证返回结构、字段及 `entity` 摘要。

##### 请求参数

```json
{
    "description": "test-key-full",
    "expired_time": -1,
    "enabled": true,
    "unlimited_quota": false,
    "models": ["gpt-4"],
    "subnet": ["10.0.0.0/8"],
    "quota_plan": {
        "unlimited": false,
        "quota": 1000000,
        "unit": "total_token",
        "reset_period": "monthly"
    },
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "1m",
                    "model": "*",
                    "window_minutes": 1,
                    "max_tokens": 10000,
                    "step_minutes": 1
                }
            ],
            "max_concurrency": 50
        }
    },
    "entity_id": "<entity_id>"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| description | "test-key-full" | Equals |
| models | ["gpt-4"] | Equals |
| subnet | ["10.0.0.0/8"] | Equals |
| quota_plan.quota | 1000000 | Equals |
| rate_limit_policy.enabled | true | Equals |
| entity.id | "<entity_id>" | Equals |
| entity.name | 非空字符串 | NotEmpty |
| quota_plan.balance | 不存在 | NotExists |

---

#### 6.4.3 AK-1-003：导入外部 key（正常参数）

##### 设计思路

验证传入外部 `key` 时，返回的 `key` 与传入值一致。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`key=my-imported-key-001`。
2. 验证返回 `Data.key` 等于传入值。

##### 请求参数

```json
{
    "description": "test-key-import",
    "key": "my-imported-key-001"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| key | "my-imported-key-001" | Equals |

---

#### 6.4.4 AK-1-004：导入重复 key（异常参数）

##### 设计思路

验证传入已存在的外部 `key` 时返回 422。

##### 前提数据准备

已用同一 key 创建 API-Key。

##### 执行步骤

1. 发送 POST 请求，使用重复的 `key`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "description": "test-key-dup",
    "key": "my-imported-key-001"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：key 已存在的错误信息  
**Data**：null

---

#### 6.4.5 AK-1-005：缺少 description（必填校验）

##### 设计思路

验证 `description` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 为空对象。
2. 验证返回错误码。

##### 请求参数

```json
{}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description" 的错误信息  
**Data**：null

---

## 7. 查询 API-Key 列表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询 API-Key 列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/api-keys` |
| 说明 | 分页列表，quota_plan 含 balance |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认 1 |
| page_size | int | N | 每页条数，默认 20，最大 100 |
| enabled | bool | N | 是否启用过滤 |
| entity_id | string | N | 按挂载的 Entity ID 过滤 |
| unlimited_quota | bool | N | 是否无限配额过滤 |

#### 7.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []APIKey | API-Key 列表 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

> 列表元素 `quota_plan` 包含 `balance` 字段。

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-2-001 | 列表返回包含 balance | 正常参数 | balance.used + balance.remaining = quota |
| AK-2-002 | 列表分页参数 | 边界值 | page=1&page_size=1 |
| AK-2-003 | 按 enabled 过滤 | 正常参数 | 仅返回指定状态 |

### 7.4 测试场景详细设计

#### 7.4.1 AK-2-001：列表返回包含 balance（正常参数）

##### 设计思路

验证列表接口返回中 `quota_plan` 包含 `balance`，且 `used + remaining = quota`。

##### 前提数据准备

已创建非无限配额 API-Key。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys`。
2. 验证返回结构和 `quota_plan.balance`。

##### 请求参数

```
page=1&page_size=20
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[0].quota_plan | 非空对象 | IsObject |
| list[0].quota_plan.balance | 非空对象 | IsObject |
| list[0].quota_plan.balance.used + list[0].quota_plan.balance.remaining | 等于 quota | SumEquals |

---

#### 7.4.2 AK-2-002：列表分页参数（边界值）

##### 设计思路

验证分页参数边界，`page_size=1` 时返回单条记录。

##### 前提数据准备

已创建至少 2 条 API-Key。

##### 执行步骤

1. 发送 GET 请求，`page=1&page_size=1`。
2. 验证 `list` 长度为 1，`pagination.total ≥ 2`。

##### 请求参数

```
page=1&page_size=1
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 长度为 1 | Len=1 |
| pagination.page | 1 | Equals |
| pagination.page_size | 1 | Equals |
| pagination.total | ≥ 2 | Gte |

---

#### 7.4.3 AK-2-003：按 enabled 过滤（正常参数）

##### 设计思路

验证按 `enabled` 过滤后，列表中仅返回指定状态的 API-Key。

##### 前提数据准备

已创建 enabled=true 和 enabled=false 的 API-Key。

##### 执行步骤

1. 发送 GET 请求，`enabled=false`。
2. 验证列表中所有元素的 `enabled` 均为 `false`。

##### 请求参数

```
enabled=false
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].enabled | 全部为 false | Equals |

---

## 8. 查询单个 API-Key

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询单个 API-Key |
| 方法 | GET |
| 路径 | `/open-api/v1/api-keys/{id}` |
| 说明 | 详情，quota_plan 含 balance |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

#### 8.2.2 返回数据字段

同 6.2.2，`quota_plan` 包含 `balance`。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-3-001 | 详情返回包含 balance | 正常参数 | quota_plan.balance 存在 |
| AK-3-002 | 查询不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |

### 8.4 测试场景详细设计

#### 8.4.1 AK-3-001：详情返回包含 balance（正常参数）

##### 设计思路

验证详情接口返回中 `quota_plan.balance` 存在且为对象。

##### 前提数据准备

已创建非无限配额 API-Key。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/{id}`。
2. 验证返回 `quota_plan.balance`。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| quota_plan | 非空对象 | IsObject |
| quota_plan.balance | 非空对象 | IsObject |
| quota_plan.balance.used | 大于等于 0 | Gte(0) |
| quota_plan.balance.remaining | 大于等于 0 | Gte(0) |

---

#### 8.4.2 AK-3-002：查询不存在的 API-Key（异常参数）

##### 设计思路

验证查询不存在的 ID 时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/non-existent-id`。
2. 验证返回错误码。

##### 请求参数

URI：`non-existent-id`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：API-Key 不存在的错误信息  
**Data**：null

---

## 9. 全量更新 API-Key

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 全量更新 API-Key |
| 方法 | PUT |
| 路径 | `/open-api/v1/api-keys/{id}` |
| 说明 | 全量替换，`key` 字段忽略 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

##### Body 参数

同创建接口。

#### 9.2.2 返回数据字段

同创建接口。

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-4-001 | 全量更新 quota_plan 触发余额重置 | 正常参数 | 验证级联更新 |
| AK-4-002 | 全量更新传入 key 被忽略 | 业务规则 | key 保持不变 |
| AK-4-003 | 全量更新后查询一致性 | 返回数据 | PUT 后立即 GET，验证数据一致 |

### 9.4 测试场景详细设计

#### 9.4.1 AK-4-001：全量更新 quota_plan 触发余额重置（正常参数）

##### 设计思路

验证全量更新 `quota_plan.quota` 为新值时，余额同步重置。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/api-keys/{id}`，传入完整 Body，`quota_plan.quota` 改为新值。
2. 验证返回 `quota_plan.quota` 为新值，查询余额 `remaining` 同步更新。

##### 请求参数

```json
{
    "description": "test-key-updated",
    "quota_plan": {
        "unlimited": false,
        "quota": 500000,
        "unit": "total_token",
        "reset_period": "monthly"
    },
    "rate_limit_policy": {
        "enabled": false,
        "rules": {}
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| description | "test-key-updated" | Equals |
| quota_plan.quota | 500000 | Equals |
| quota_plan.balance | 不存在 | NotExists |

---

#### 9.4.2 AK-4-002：全量更新传入 key 被忽略（业务规则）

##### 设计思路

验证更新接口传入 `key` 不生效，应保持原值不变。

##### 前提数据准备

已创建 API-Key，记录原 key 值。

##### 执行步骤

1. 发送 PUT 请求，完整 Body 含 `key`: "new-key"。
2. 验证返回 `Data.key` 保持原值不变。

##### 请求参数

```json
{
    "key": "new-key",
    "description": "test-key-ignore-key",
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false,
        "rules": {}
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| key | 与原 key 一致 | Equals |
| description | "test-key-ignore-key" | Equals |

---

#### 9.4.3 AK-4-003：全量更新后查询一致性（返回数据）

##### 设计思路

验证 PUT 更新成功后，立即通过 GET 查询，返回数据与更新请求一致。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 PUT 请求更新 API-Key。
2. 发送 GET 请求查询该 API-Key。
3. 对比两次返回的数据是否一致（忽略 balance）。

##### 请求参数

```json
{
    "description": "test-key-consistency",
    "enabled": false,
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false,
        "rules": {}
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| description | "test-key-consistency" | Equals |
| enabled | false | Equals |

---

## 10. 部分更新 API-Key

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 部分更新 API-Key |
| 方法 | PATCH |
| 路径 | `/open-api/v1/api-keys/{id}` |
| 说明 | 部分更新，`key` 字段忽略 |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

##### Body 参数

同创建接口，仅传需修改字段。

#### 10.2.2 返回数据字段

同创建接口。

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-5-001 | 部分更新 enabled | 正常参数 | enabled 更新，其余不变 |
| AK-5-002 | 部分更新 route_rules | 正常参数 | route_rules 更新 |
| AK-5-003 | 部分更新后查询一致性 | 返回数据 | PATCH 后立即 GET，验证数据一致 |

### 10.4 测试场景详细设计

#### 10.4.1 AK-5-001：部分更新 enabled（正常参数）

##### 设计思路

验证部分更新 `enabled` 成功，其余字段保持不变。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/api-keys/{id}`。
2. 验证返回 `Data.enabled=false`。

##### 请求参数

```json
{
    "enabled": false
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | false | Equals |
| description | 与原值一致 | Equals |

---

#### 10.4.2 AK-5-002：部分更新 route_rules（正常参数）

##### 设计思路

验证部分更新 `route_rules` 成功。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 PATCH 请求，传入新的 `route_rules`。
2. 验证返回 `route_rules` 与输入一致。

##### 请求参数

```json
{
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "default",
                "Cond": "default_t()",
                "targets": [
                    {
                        "ClusterName": "cluster1",
                        "Weight": 100
                    }
                ]
            }
        ]
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| route_rules.enabled | true | Equals |
| route_rules.rules[0].name | "default" | Equals |
| route_rules.rules[0].targets[0].ClusterName | "cluster1" | Equals |

---

#### 10.4.3 AK-5-003：部分更新后查询一致性（返回数据）

##### 设计思路

验证 PATCH 更新成功后，立即通过 GET 查询，返回数据与更新请求一致。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 PATCH 请求更新 `models`。
2. 发送 GET 请求查询该 API-Key。
3. 对比两次返回的 `models`。

##### 请求参数

```json
{
    "models": ["gpt-3.5-turbo"]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| models | ["gpt-3.5-turbo"] | Equals |

---

## 11. 删除 API-Key

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 删除 API-Key |
| 方法 | DELETE |
| 路径 | `/open-api/v1/api-keys/{id}` |
| 说明 | 级联删除专属资源 |

### 11.2 接口参数说明

#### 11.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

#### 11.2.2 返回数据字段

Data 为 null。

### 11.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-6-001 | 删除 API-Key | 正常参数 | 级联清理，再次查询返回 404 |
| AK-6-002 | 删除不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |

### 11.4 测试场景详细设计

#### 11.4.1 AK-6-001：删除 API-Key（正常参数）

##### 设计思路

验证删除 API-Key 成功，并级联清理专属资源。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/api-keys/{id}`。
2. 验证返回成功。
3. 再次查询，验证返回 404。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 11.4.2 AK-6-002：删除不存在的 API-Key（异常参数）

##### 设计思路

验证删除不存在的 API-Key 时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/api-keys/non-existent-id`。
2. 验证返回错误码。

##### 请求参数

URI：`non-existent-id`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：API-Key 不存在的错误信息  
**Data**：null

---

## 12. 查询配额计划

### 12.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询配额计划 |
| 方法 | GET |
| 路径 | `/open-api/v1/api-keys/{id}/quota-plan` |
| 说明 | 返回完整 quota_plan 含 balance |

### 12.2 接口参数说明

#### 12.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

#### 12.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| unlimited | bool | 是否无限配额 |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 |
| quota | int64 | 配额总量 |
| unit | string | 配额单位 |
| reset_period | string | 配额重置周期 |
| balance | object | 余额状态，包含 used 和 remaining |

### 12.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-7-001 | 查询配额计划含 balance | 正常参数 | 返回完整 quota_plan |
| AK-7-002 | 查询无限配额 API-Key 的 quota-plan | 边界场景 | 验证行为 |

### 12.4 测试场景详细设计

#### 12.4.1 AK-7-001：查询配额计划含 balance（正常参数）

##### 设计思路

验证独立 quota-plan 接口返回完整配额计划与余额。

##### 前提数据准备

已创建非无限配额 API-Key。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/{id}/quota-plan`。
2. 验证返回包含 `balance`。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| unlimited | false | Equals |
| quota | 与创建时一致 | Equals |
| balance | 非空对象 | IsObject |
| balance.used | 大于等于 0 | Gte(0) |
| balance.remaining | 大于等于 0 | Gte(0) |

---

#### 12.4.2 AK-7-002：查询无限配额 API-Key 的 quota-plan（边界场景）

##### 设计思路

验证查询无限配额 API-Key 的 quota-plan 行为（返回 404 或 unlimited=true 无 balance，以接口定义为准）。

##### 前提数据准备

已创建 `unlimited_quota=true` 或 `quota_plan.unlimited=true` 的 API-Key。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/{id}/quota-plan`。
2. 验证返回行为。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：404 或 200（以接口实现为准）  
**ErrMsg**：对应错误信息或 success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| unlimited | true（若为 200） | Equals |
| balance | 不存在（若为 200） | NotExists |

---

## 13. 重置配额余额

### 13.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 重置配额余额 |
| 方法 | POST |
| 路径 | `/open-api/v1/api-keys/{id}/quota-plan/reset` |
| 说明 | 重置 API-Key 的配额余额 |

### 13.2 接口参数说明

#### 13.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | API-Key 标识 |

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int64 | N | 重置后的配额总量 |
| reason | string | N | 重置原因 |

#### 13.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 标识 |
| previous_quota | int64 | 重置前配额 |
| new_quota | int64 | 重置后配额 |
| balance | object | 余额变更详情 |
| balance.previous_remaining | int64 | 重置前剩余量 |
| balance.new_remaining | int64 | 重置后剩余量 |
| balance.used | int64 | 当前已用量 |

### 13.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-8-001 | 重置配额余额 | 正常参数 | used=0，new_remaining=previous_quota |
| AK-8-002 | 重置并修改 quota | 正常参数 | new_quota 和 new_remaining 同步更新 |

### 13.4 测试场景详细设计

#### 13.4.1 AK-8-001：重置配额余额（正常参数）

##### 设计思路

验证不传 quota 时按当前 quota 重置余额。

##### 前提数据准备

已创建非无限配额 API-Key。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/api-keys/{id}/quota-plan/reset`。
2. 验证返回的 `balance.used=0`，`new_remaining=previous_quota`。

##### 请求参数

```json
{
    "reason": "test reset"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 与 URI 一致 | Equals |
| previous_quota | 与当前 quota 一致 | Equals |
| new_quota | 与 previous_quota 一致 | Equals |
| balance.used | 0 | Equals |
| balance.new_remaining | 与 new_quota 一致 | Equals |

---

#### 13.4.2 AK-8-002：重置并修改 quota（正常参数）

##### 设计思路

验证传入 quota 时同步更新 quota 并重置余额。

##### 前提数据准备

已创建非无限配额 API-Key。

##### 执行步骤

1. 发送 POST 请求，传入新的 `quota`。
2. 验证返回的 `new_quota` 和 `new_remaining` 均为新值，`used=0`。

##### 请求参数

```json
{
    "quota": 500000,
    "reason": "adjust"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| new_quota | 500000 | Equals |
| balance.new_remaining | 500000 | Equals |
| balance.used | 0 | Equals |

---

## 14. 依赖与数据准备

1. **Entity 数据**：AK 挂载用例需要预先创建 Entity-Type 与 Entity。
2. **证书/集群**：如测试 route_rules 中 `ClusterName` 有效性，可预先创建 Cluster。
3. **Redis Mock**：配额余额相关用例依赖测试环境内存 Redis Mock。
4. **外部 key 唯一性**：导入 key 用例需保证 key 在测试数据库中全局唯一。

## 15. 注意事项

1. v0.3.0 已删除 `/api-keys/actions/generate`，测试方案不再覆盖该接口。
2. 创建/更新接口返回的 `quota_plan` 不含 `balance`；仅列表、详情及独立 quota-plan 接口含 `balance`。
3. 更新接口传入 `key` 不生效，应在用例中显式断言 `key` 未变化。
4. 测试环境 `SkipTokenValidate=true`，无需构造真实 Token。
5. 同一模块内测试共享数据库，注意每个用例清理自身产生的 API-Key，避免名称/key 冲突。
