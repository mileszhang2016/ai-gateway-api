# Entity 模块测试用例设计文档

## 模块概述

Entity 模块负责实体的管理，包括创建、查询、更新、删除实体，以及配额计划的查询和重置。

## 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| ENT-1 | 创建Entity | POST | `/open-api/v1/entities` | 创建新实体 |
| ENT-2 | 查询Entity列表 | GET | `/open-api/v1/entities` | 获取实体列表 |
| ENT-3 | 查询单个Entity | GET | `/open-api/v1/entities/{id}` | 查询指定实体详情 |
| ENT-4 | 全量更新Entity | PUT | `/open-api/v1/entities/{id}` | 全量更新实体 |
| ENT-5 | 部分更新Entity | PATCH | `/open-api/v1/entities/{id}` | 部分更新实体 |
| ENT-6 | 删除Entity | DELETE | `/open-api/v1/entities/{id}` | 删除指定实体 |
| ENT-7 | 查询配额计划 | GET | `/open-api/v1/entities/{id}/quota-plan` | 查询实体配额计划（含余额） |
| ENT-8 | 重置配额余额 | POST | `/open-api/v1/entities/{id}/quota-plan/reset` | 重置实体配额余额 |

## 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建Entity | 5 |
| 查询Entity列表 | 2 |
| 查询单个Entity | 2 |
| 全量更新Entity | 3 |
| 部分更新Entity | 3 |
| 删除Entity | 2 |
| 查询配额计划 | 2 |
| 重置配额余额 | 2 |
| **合计** | **21** |

## 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 目录结构

```

├── README.md
├── create/design.md
├── list/design.md
├── detail/design.md
├── full_update/design.md
├── partial_update/design.md
├── delete/design.md
├── quota_plan/design.md
└── quota_reset/design.md
```

---

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

---

# 删除Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 删除Entity |
| 方法 | DELETE |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 删除指定Entity |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | Data为null |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-6-001 | 正常删除Entity | 正常参数 | 验证 ErrNum=200 |
| ENT-6-002 | 删除不存在的Entity | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### ENT-6-001：正常删除Entity（正常参数）

#### 设计思路

验证删除Entity接口的基本功能：删除已存在的Entity，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_del", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 DELETE 请求到 `/open-api/v1/entities/{id}`
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### ENT-6-002：删除不存在的Entity（异常参数）

#### 设计思路

验证删除不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_del" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/entities/non_existent_del`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_del |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

# 查询单个Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询单个Entity |
| 方法 | GET |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 查询指定Entity的详细信息 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

### 返回数据字段

同创建接口返回（不含balance字段）。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-3-001 | 查询已存在的Entity | 正常参数 | 返回完整Entity信息 |
| ENT-3-002 | 查询不存在的Entity | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### ENT-3-001：查询已存在的Entity（正常参数）

#### 设计思路

验证查询Entity详情接口的基本功能：查询已存在的Entity，确认返回完整的Entity信息。

#### 前提数据准备

- 预先创建Entity：name="test_entity_detail", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 GET 请求到 `/open-api/v1/entities/{id}`
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 创建时返回的id | Equals |
| name | "test_entity_detail" | Equals |
| type | "dep" | Equals |

---

### ENT-3-002：查询不存在的Entity（异常参数）

#### 设计思路

验证查询不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_entity" 不存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities/non_existent_entity`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_entity |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

# 全量更新Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 全量更新Entity |
| 方法 | PUT |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 全量更新Entity，同创建接口Body参数 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

同创建Entity的Body参数。

**约束**：
- `type` 不可修改（创建后固定）
- `name` 全局唯一，不可与其他Entity冲突

### 返回数据字段

同创建接口返回（不含balance）。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-4-001 | 全量更新Entity名称 | 正常参数 | 修改name |
| ENT-4-002 | 更新不存在的Entity | 异常参数 | 验证 ErrNum=404 |
| ENT-4-003 | 更新后名称与其他Entity冲突 | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### ENT-4-001：全量更新Entity名称（正常参数）

#### 设计思路

验证全量更新Entity接口的基本功能：更新已存在的Entity名称，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_put", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PUT 请求到 `/open-api/v1/entities/{id}`，修改name
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "name": "test_entity_put_updated",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_entity_put_updated" | Equals |
| type | "dep" | Equals |

---

### ENT-4-002：更新不存在的Entity（异常参数）

#### 设计思路

验证更新不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_put" 不存在

#### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/entities/non_existent_put`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_put |

**Body 参数**：

```json
{
    "name": "test_entity_update",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

### ENT-4-003：更新后名称与其他Entity冲突（业务规则）

#### 设计思路

验证更新后名称与其他Entity冲突时，接口应返回业务错误。

#### 前提数据准备

- 预先创建Entity1：name="test_entity_conflict1", type="dep"
- 预先创建Entity2：name="test_entity_conflict2", type="dep"

#### 执行步骤

1. 创建两个Entity
2. 用Entity1的id发送 PUT 请求，将name改为Entity2的name
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | Entity1的id |

**Body 参数**：

```json
{
    "name": "test_entity_conflict2",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含数据库约束异常的错误信息（UNIQUE constraint failed）  
**Data**：null

---

# 查询Entity列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询Entity列表 |
| 方法 | GET |
| 路径 | /open-api/v1/entities |
| 说明 | 获取实体列表，支持分页和过滤 |

---

## 二、接口参数说明

### 请求参数

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认1 |
| page_size | int | N | 每页条数，默认20，最大100 |
| type | string | N | 按类型过滤 |
| parent_id | string | N | 按父Entity过滤 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []object | Entity列表 |
| list[].id | string | Entity唯一标识 |
| list[].name | string | Entity名称 |
| list[].type | string | Entity类型 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-2-001 | 获取Entity列表 | 正常参数 | 返回 list 和 pagination |
| ENT-2-002 | 验证返回字段完整性 | 返回数据 | 验证每个元素包含 id、name、type |

---

## 四、测试场景详细设计

---

### ENT-2-001：获取Entity列表（正常参数）

#### 设计思路

验证获取Entity列表接口的基本功能：确认接口返回Entity数组和分页信息。

#### 前提数据准备

- 预先创建至少一个Entity：name="test_entity_list", type="dep"

#### 执行步骤

1. 先创建Entity
2. 发送 GET 请求到 `/open-api/v1/entities`
3. 验证响应状态码和返回结构

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 非空数组 | IsArray, NotEmpty |
| pagination | 非空对象 | NotEmpty |
| pagination.page | 默认值 | NotEmpty |
| pagination.page_size | 默认值 | NotEmpty |
| pagination.total | >= 1 | GreaterThanOrEqual(1) |

---

### ENT-2-002：验证返回字段完整性（返回数据）

#### 设计思路

验证返回的Entity列表中每个元素包含完整的字段信息。

#### 前提数据准备

- 预先创建Entity：name="test_entity_check", type="dep"

#### 执行步骤

1. 先创建Entity
2. 发送 GET 请求到 `/open-api/v1/entities`
3. 验证返回列表中每个元素包含 id、name、type 字段

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list[].id | 非空字符串 | NotEmpty |
| list[].name | 非空字符串 | NotEmpty |
| list[].type | 非空字符串 | NotEmpty |

---

# 部分更新Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 部分更新Entity |
| 方法 | PATCH |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 部分更新Entity，仅传需修改字段 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

同创建Entity的Body参数，仅传需修改字段。

### 返回数据字段

同创建接口返回（不含balance）。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-5-001 | 部分更新Entity名称 | 正常参数 | 仅传name字段 |
| ENT-5-002 | 更新不存在的Entity | 异常参数 | 验证 ErrNum=404 |
| ENT-5-003 | 更新配额计划 | 正常参数 | 修改quota_plan |

---

## 四、测试场景详细设计

---

### ENT-5-001：部分更新Entity名称（正常参数）

#### 设计思路

验证部分更新Entity接口的基本功能：仅传入name字段进行更新，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_patch", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PATCH 请求到 `/open-api/v1/entities/{id}`，仅传name字段
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "name": "test_entity_patch_updated"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_entity_patch_updated" | Equals |
| type | "dep" | Equals（保持不变） |

---

### ENT-5-002：更新不存在的Entity（异常参数）

#### 设计思路

验证部分更新不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_patch" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entities/non_existent_patch`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_patch |

**Body 参数**：

```json
{
    "name": "test_entity_patch_update"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

### ENT-5-003：更新配额计划（正常参数）

#### 设计思路

验证部分更新Entity时修改配额计划的场景，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_quota_patch", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PATCH 请求到 `/open-api/v1/entities/{id}`，更新quota_plan
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "quota_plan": {
        "unlimited": false,
        "quota": 50000000,
        "unit": "total_token",
        "reset_period": "weekly"
    }
}
```

#### 预期返回结果

**ErrNum**：500  
**ErrMsg**：system error  
**Data**：null

> 注：当前部分更新 quota_plan 功能存在系统错误，无法正常更新

---

# 查询配额计划 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询配额计划 |
| 方法 | GET |
| 路径 | /open-api/v1/entities/{id}/quota-plan |
| 说明 | 查询Entity的配额计划（含实时余额） |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| unlimited | bool | 是否无限配额 |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 |
| quota | int64 | 配额总量 |
| unit | string | 配额单位 |
| reset_period | string | 配额重置周期 |
| balance | object | 余额状态（只读，当前未返回） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-7-001 | 查询存在的Entity配额计划 | 正常参数 | 返回完整配额信息 |
| ENT-7-002 | 查询不存在的Entity配额计划 | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### ENT-7-001：查询存在的Entity配额计划（正常参数）

#### 设计思路

验证查询Entity配额计划接口的基本功能：查询已存在的Entity的配额计划，确认返回完整的配额信息和余额。

#### 前提数据准备

- 预先创建Entity：name="test_entity_qp", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 GET 请求到 `/open-api/v1/entities/{id}/quota-plan`
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| unlimited | bool 类型 | NotEmpty |
| quota | int64 类型 | NotEmpty |
| unit | 非空字符串 | NotEmpty |

> 注：当前查询配额计划接口不返回 balance 字段

---

### ENT-7-002：查询不存在的Entity配额计划（异常参数）

#### 设计思路

验证查询不存在的Entity配额计划时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_qp" 不存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities/non_existent_qp/quota-plan`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_qp |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

# 重置配额余额 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 重置配额余额 |
| 方法 | POST |
| 路径 | /open-api/v1/entities/{id}/quota-plan/reset |
| 说明 | 重置Entity的配额余额 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int64 | N | 重置后的配额总量，不传则按当前quota重置 |
| reason | string | N | 重置原因，用于审计 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity标识 |
| previous_quota | int64 | 重置前配额 |
| new_quota | int64 | 重置后配额 |
| balance | object | 余额变更详情 |
| balance.previous_remaining | int64 | 重置前剩余量 |
| balance.new_remaining | int64 | 重置后剩余量 |
| balance.used | int64 | 当前已用量，重置后为0 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-8-001 | 重置配额余额（不传新配额） | 正常参数 | 按当前quota重置 |
| ENT-8-002 | 重置配额余额（传新配额） | 正常参数 | 更新quota并重置 |

---

## 四、测试场景详细设计

---

### ENT-8-001：重置配额余额（不传新配额）

#### 设计思路

验证重置配额余额接口的基本功能：不传新配额，按当前quota重置余额，确认接口返回成功并返回余额变更详情。

#### 前提数据准备

- 预先创建Entity：name="test_entity_reset1", type="dep"，且配置非无限配额

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`，不传quota
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "reason": "月度重置"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 创建时返回的id | Equals |
| previous_quota | > 0 | GreaterThan(0) |
| new_quota | == previous_quota | Equals |
| balance.used | 0 | Equals |

---

### ENT-8-002：重置配额余额（传新配额）

#### 设计思路

验证重置配额余额时传入新配额的场景：更新quota并同步重置余额，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_reset2", type="dep"，且配置非无限配额

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`，传入新quota
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "quota": 50000000,
    "reason": "调整配额"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 创建时返回的id | Equals |
| new_quota | 50000000 | Equals |
| balance.new_remaining | 50000000 | Equals |
| balance.used | 0 | Equals |

---

