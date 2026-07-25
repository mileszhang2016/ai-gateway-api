# 创建Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 创建Entity |
| 方法 | POST |
| 路径 | /open-api/v1/entities |
| 说明 | 创建新实体，需要提供名称和类型 |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | Y | Entity名称，全局唯一 |
| type | string | Y | Entity类型，必须引用已定义的Entity-Type |
| parent_id | string | N | 父Entity ID，为空表示根节点 |
| allow_models | []string | N | 允许访问的模型白名单，默认["*"] |
| block_models | []string | N | 禁止访问的模型黑名单，默认[] |
| quota_plan | object | N | 配额计划，未设置则使用默认值 |
| rate_limit_policy | object | N | 限流策略，未设置则使用默认值(enabled=false) |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity唯一标识，系统生成 |
| name | string | Entity名称 |
| type | string | Entity类型 |
| parent_id | string | 父Entity ID |
| allow_models | []string | 允许访问的模型白名单 |
| block_models | []string | 禁止访问的模型黑名单 |
| quota_plan | object | 配额计划，不含balance字段 |
| rate_limit_policy | object | 限流策略 |
| create_time | int64 | 创建时间，Unix时间戳 |
| update_time | int64 | 更新时间，Unix时间戳 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-1-001 | 创建Entity（仅必填字段） | 正常参数 | name+type |
| ENT-1-002 | 创建Entity（含配额计划） | 正常参数 | 包含quota_plan |
| ENT-1-003 | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| ENT-1-004 | 缺少 type | 必填校验 | 验证 ErrNum=422 |
| ENT-1-005 | 重复创建同名Entity | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### ENT-1-001：创建Entity（仅必填字段）

#### 设计思路

验证创建Entity接口的基本功能：传入必填字段name和type，确认接口返回成功并返回完整的Entity信息。

#### 前提数据准备

- 确保Entity-Type "dep" 已存在

#### 执行步骤

1. 构造请求 Body：包含 name 和 type
2. 发送 POST 请求到 `/open-api/v1/entities`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "name": "test_entity_001",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串，格式如 ent-xxx | NotEmpty |
| name | "test_entity_001" | Equals |
| type | "dep" | Equals |
| parent_id | null | IsNull |

---

### ENT-1-002：创建Entity（含配额计划）

#### 设计思路

验证创建Entity时传入配额计划的场景，确认接口返回成功并正确保存配额计划信息。

#### 前提数据准备

- 确保Entity-Type "dep" 已存在

#### 执行步骤

1. 构造请求 Body：包含 name、type 和 quota_plan
2. 发送 POST 请求到 `/open-api/v1/entities`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "name": "test_entity_quota",
    "type": "dep",
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_entity_quota" | Equals |
| type | "dep" | Equals |

> 注：创建接口返回数据中不包含 quota_plan 字段，配额计划需通过查询配额计划接口验证

---

### ENT-1-003：缺少 name（必填校验）

#### 设计思路

验证 `name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 name 字段
2. 发送 POST 请求到 `/open-api/v1/entities`
3. 验证返回错误码

#### 请求参数

```json
{
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "name" 的错误信息  
**Data**：null

---

### ENT-1-004：缺少 type（必填校验）

#### 设计思路

验证 `type` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 type 字段
2. 发送 POST 请求到 `/open-api/v1/entities`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_entity_notype"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "type" 的错误信息  
**Data**：null

---

### ENT-1-005：重复创建同名Entity（业务规则）

#### 设计思路

验证 name 必须全局唯一，当尝试创建已存在的 name 时，接口应返回业务错误。

#### 前提数据准备

- 预先创建Entity：name="test_entity_dup", type="dep"

#### 执行步骤

1. 先创建Entity
2. 使用相同 name 再次发送 POST 请求到 `/open-api/v1/entities`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_entity_dup",
    "type": "team"
}
```

#### 预期返回结果

**ErrNum**：555  
**ErrMsg**：包含重复名称的错误信息  
**Data**：null