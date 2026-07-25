# 部分更新API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 部分更新API-Key |
| 方法 | PATCH |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 部分更新指定 API-Key 的字段，仅传需修改的字段 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | 否 | 描述，最大511字符 |
| enabled | bool | 否 | 是否启用 |
| key | string | 否 | API-Key 值，必须以 productName- 开头 |
| unlimited_quota | bool | 否 | 是否无限配额 |
| expired_time | int64 | 否 | 过期时间（-1=永不过期） |
| models | []string | 否 | 允许访问的模型白名单 |
| subnet | []string | 否 | 允许的客户端子网（CIDR格式） |
| entity_id | string | 否 | 挂载的 Entity ID |
| quota_plan | object | 否 | 配额计划 |
| rate_limit_policy | object | 否 | 限流策略 |

### 返回数据字段

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
| quota_plan | object | 配额计划（含 balance） |
| rate_limit_policy | object | 限流策略 |
| entity_id | string | 挂载的 Entity ID（可选） |
| remaining_quota | int64 | 剩余配额 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-5-001 | 仅修改 description | 正常参数 | 验证部分更新 description |
| AK-5-002 | 禁用 API-Key | 正常参数 | 验证 enabled=false |
| AK-5-003 | 启用 API-Key | 正常参数 | 验证 enabled=true |
| AK-5-004 | 更新不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-5-005 | 更新 expired_time（永不过期） | 正常参数 | expired_time=-1 |
| AK-5-006 | 更新 expired_time（未来时间） | 正常参数 | 设置未来过期时间 |
| AK-5-007 | 更新 quota_plan | 正常参数 | 更新配额计划 |
| AK-5-008 | 更新 rate_limit_policy | 正常参数 | 更新限流策略 |
| AK-5-009 | 更新 models | 正常参数 | 更新允许的模型列表 |
| AK-5-010 | 更新 subnet | 正常参数 | 更新允许的子网 |
| AK-5-011 | 更新 entity_id | 正常参数 | 挂载到指定 Entity |
| AK-5-012 | description 超长（512字符） | 边界值 | 验证 ErrNum=422 |
| AK-5-013 | expired_time=-2（非法值） | 边界值 | 验证 ErrNum=422 |
| AK-5-014 | expired_time=过去时间 | 边界值 | 验证 ErrNum=422 |
| AK-5-015 | subnet 格式错误 | 异常参数 | 验证 ErrNum=422 |
| AK-5-016 | rate_limit_policy.enabled=true 但无规则 | 异常参数 | 验证 ErrNum=422 |
| AK-5-017 | quota_plan.quota<0 | 异常参数 | 验证 ErrNum=422 |
| AK-5-018 | entity_id 指向不存在的 Entity | 异常参数 | 验证 ErrNum=422 |
| AK-5-019 | id 超长（>256字符） | 异常参数 | 验证 ErrNum=422 |
| AK-5-020 | 验证返回结构完整性 | 返回数据校验 | 验证所有字段存在 |

---

## 四、测试场景详细设计

---

### AK-5-001：仅修改 description

#### 设计思路

验证 PATCH 仅传 description 时，其他字段保持不变。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PATCH 请求，仅修改 description
3. 验证 description 更新，其他字段不变

#### 请求参数

```json
{
    "description": "patched-description"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.description**："patched-description"  
**Data.id**：不变  
**Data.enabled**：不变

---

### AK-5-002：禁用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=false 禁用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=true 的 API-Key

#### 请求参数

```json
{
    "enabled": false
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.enabled**：false

---

### AK-5-003：启用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=true 启用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=false 的 API-Key

#### 请求参数

```json
{
    "enabled": true
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.enabled**：true

---

### AK-5-004：更新不存在的 API-Key

#### 设计思路

验证 PATCH 不存在的 ID 时返回 404。

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 "API-Key" 和 "not exist"

---

### AK-5-005：更新 expired_time（永不过期）

#### 设计思路

验证 PATCH 设置 expired_time=-1（永不过期）。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": -1
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.expired_time**：-1

---

### AK-5-006：更新 expired_time（未来时间）

#### 设计思路

验证 PATCH 设置未来过期时间。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key
2. 获取当前时间 + 3600 秒作为过期时间
3. 发送 PATCH 请求

#### 请求参数

```json
{
    "expired_time": <当前时间+3600秒>
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.expired_time**：等于请求中的值

---

### AK-5-007：更新 quota_plan

#### 设计思路

验证 PATCH 更新配额计划。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "quota_plan": {
        "quota": 10000,
        "unit": "token",
        "reset_period": "daily"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.quota_plan.quota**：10000

---

### AK-5-008：更新 rate_limit_policy

#### 设计思路

验证 PATCH 更新限流策略。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "window_minutes": 1,
                    "step_minutes": 1,
                    "max_tokens": 1000
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.rate_limit_policy.enabled**：true

---

### AK-5-009：更新 models

#### 设计思路

验证 PATCH 更新允许访问的模型列表。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "models": ["gpt-3.5-turbo", "gpt-4"]
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.models**：包含 ["gpt-3.5-turbo", "gpt-4"]

---

### AK-5-010：更新 subnet

#### 设计思路

验证 PATCH 更新允许的客户端子网。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "subnet": ["192.168.1.0/24"]
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.subnet**：包含 ["192.168.1.0/24"]

---

### AK-5-011：更新 entity_id

#### 设计思路

验证 PATCH 将 API-Key 挂载到指定 Entity。

#### 前提数据准备

- 创建 Entity-Type（type_name=test_type_patch, level=1）
- 创建 Entity（type=test_type_patch）
- 创建 API-Key（不挂载 Entity）

#### 请求参数

```json
{
    "entity_id": "<entity_id>"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.entity_id**：等于请求中的 entity_id

---

### AK-5-012：description 超长（512字符）

#### 设计思路

验证 description 超过最大长度 511 字符时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "description": "<512个字符>"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description must be less than 512 characters"

---

### AK-5-013：expired_time=-2（非法值）

#### 设计思路

验证 expired_time 小于 -1 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": -2
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Invalid expired_time"

---

### AK-5-014：expired_time=过去时间

#### 设计思路

验证 expired_time 设置为过去时间时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": <当前时间-3600秒>
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "expired_time must be >= current time"

---

### AK-5-015：subnet 格式错误

#### 设计思路

验证 subnet 格式不正确时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "subnet": ["invalid-subnet"]
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "invalid subnet format"

---

### AK-5-016：rate_limit_policy.enabled=true 但无规则

#### 设计思路

验证 rate_limit_policy.enabled=true 但未设置 rules 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "rate_limit_policy": {
        "enabled": true
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "when rate_limit_policy.enabled is true, rules must be set"

---

### AK-5-017：quota_plan.quota<0

#### 设计思路

验证 quota_plan.quota 为负数时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "quota_plan": {
        "quota": -100
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "quota must be >= 0"

---

### AK-5-018：entity_id 指向不存在的 Entity

#### 设计思路

验证 entity_id 指向不存在的 Entity 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "entity_id": "nonexistent-entity-id"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Entity not found"

---

### AK-5-019：id 超长（>256字符）

#### 设计思路

验证 URI 参数 id 超长时返回参数错误。

#### 请求参数

URI：`/open-api/v1/api-keys/<257个字符>`

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "id" 和 "invalid"

---

### AK-5-020：验证返回结构完整性

#### 设计思路

验证部分更新成功后返回的结构完整性。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key
2. 发送 PATCH 请求修改 description
3. 验证返回结构包含所有必填字段

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 必填字段校验**：

| 键名 | 预期类型 |
|------|---------|
| id | string（非空） |
| key | string（非空） |
| description | string |
| enabled | bool |
| create_time | number（>0） |
| update_time | number（>0） |
| expired_time | number |
| unlimited_quota | bool |
| models | array |
| subnet | array |
| quota_plan | object（非空） |
| rate_limit_policy | object（非空） |