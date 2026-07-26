# Entity-Type 模块测试用例设计文档

## 一、模块概述

Entity-Type 模块负责管理 Entity 的类型定义，包括类型的创建、查询、更新和删除操作。Entity-Type 定义了 Entity 的层级结构和属性约束，是创建 Entity 的前提条件。

## 二、接口列表

| 序号 | 接口名称 | 方法 | 路径 | 说明 | 用例数 |
|------|---------|------|------|------|--------|
| 1 | 创建Entity-Type | POST | /open-api/v1/entity-types | 创建新的实体类型定义 | 5 |
| 2 | 查询Entity-Type列表 | GET | /open-api/v1/entity-types | 获取实体类型列表 | 2 |
| 3 | 查询单个Entity-Type | GET | /open-api/v1/entity-types/{type_name} | 获取单个实体类型详情 | 2 |
| 4 | 更新Entity-Type | PATCH | /open-api/v1/entity-types/{type_name} | 更新实体类型描述 | 3 |
| 5 | 删除Entity-Type | DELETE | /open-api/v1/entity-types/{type_name} | 删除实体类型 | 3 |

## 三、测试用例统计

| 测试类型 | 用例数 | 说明 |
|---------|--------|------|
| 正常参数 | 5 | 正常场景下的接口调用 |
| 必填校验 | 2 | 验证必填字段缺失时的处理 |
| 边界值 | 2 | 参数边界值测试 |
| 异常参数 | 5 | 异常场景测试（如不存在的资源） |
| 业务规则 | 1 | 业务逻辑约束测试（删除时存在Entity） |
| **合计** | **15** | - |

## 四、认证方式

所有接口均使用 Token 认证，通过 `Authorization: Token TOKEN_STRING` 请求头传递。

## 五、目录结构

```

├── README.md              # 模块概述和接口列表
├── create/
│   └── design.md          # 创建Entity-Type接口测试用例设计
├── list/
│   └── design.md          # 查询Entity-Type列表接口测试用例设计
├── detail/
│   └── design.md          # 查询单个Entity-Type接口测试用例设计
├── update/
│   └── design.md          # 更新Entity-Type接口测试用例设计
└── delete/
    └── design.md          # 删除Entity-Type接口测试用例设计
```

## 六、测试用例编号规则

测试用例编号格式：`ET-{接口序号}-{用例序号}`

示例：
- ET-1-001：创建Entity-Type接口的第1个测试用例
- ET-2-002：查询Entity-Type列表接口的第2个测试用例

---

# 创建Entity-Type - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 创建Entity-Type |
| 方法 | POST |
| 路径 | /open-api/v1/entity-types |
| 说明 | 创建新的实体类型定义 |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名，全局唯一，1-32字符，仅含小写字母、数字、下划线、连字符 |
| description | string | N | 类型描述 |
| level | int | Y | 层级级别，取值范围1-5，数字越小级别越高 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| type_name | string | 类型名 |
| description | string | 描述 |
| level | int | 层级级别 |
| create_time | int64 | 创建时间，Unix时间戳（秒） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-1-001 | 创建Entity-Type（完整参数） | 正常参数 | type_name+level+description |
| ET-1-002 | 创建Entity-Type（仅必填字段） | 正常参数 | type_name+level |
| ET-1-003 | 缺少 type_name | 必填校验 | 验证 ErrNum=422 |
| ET-1-004 | 缺少 level | 必填校验 | 验证 ErrNum=422 |
| ET-1-005 | 重复创建同名Entity-Type | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### ET-1-001：创建Entity-Type（完整参数）

#### 设计思路

验证创建Entity-Type接口的基本功能：传入完整参数（type_name、level、description），确认接口返回成功并返回完整的Entity-Type信息。

#### 前提数据准备

- 确保Entity-Type "test_type_full" 不存在

#### 执行步骤

1. 构造请求 Body：包含 type_name、level、description
2. 发送 POST 请求到 `/open-api/v1/entity-types`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "type_name": "test_type_full",
    "description": "测试类型完整参数",
    "level": 1
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_full" | Equals |
| description | "测试类型完整参数" | Equals |
| level | 1 | Equals |
| create_time | 非空int64 | NotEmpty |

---

### ET-1-002：创建Entity-Type（仅必填字段）

#### 设计思路

验证创建Entity-Type接口仅传入必填字段时的功能，确认接口返回成功并返回完整的Entity-Type信息。

#### 前提数据准备

- 确保Entity-Type "test_type_min" 不存在

#### 执行步骤

1. 构造请求 Body：仅包含 type_name、level
2. 发送 POST 请求到 `/open-api/v1/entity-types`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "type_name": "test_type_min",
    "level": 2
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_min" | Equals |
| description | null | IsNull |
| level | 2 | Equals |
| create_time | 非空int64 | NotEmpty |

---

### ET-1-003：缺少 type_name（必填校验）

#### 设计思路

验证 `type_name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 type_name 字段
2. 发送 POST 请求到 `/open-api/v1/entity-types`
3. 验证返回错误码

#### 请求参数

```json
{
    "level": 1
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "type_name" 的错误信息  
**Data**：null

---

### ET-1-004：缺少 level（必填校验）

#### 设计思路

验证 `level` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 level 字段
2. 发送 POST 请求到 `/open-api/v1/entity-types`
3. 验证返回错误码

#### 请求参数

```json
{
    "type_name": "test_type_nolevel"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "level" 的错误信息  
**Data**：null

---

### ET-1-005：重复创建同名Entity-Type（业务规则）

#### 设计思路

验证 type_name 必须全局唯一，当尝试创建已存在的 type_name 时，接口应返回业务错误。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_dup", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 使用相同 type_name 再次发送 POST 请求到 `/open-api/v1/entity-types`
3. 验证返回错误码

#### 请求参数

```json
{
    "type_name": "test_type_dup",
    "level": 2
}
```

#### 预期返回结果

**ErrNum**：556  
**ErrMsg**：Duplicate Data: entity type Duplicate Data  
**Data**：null

---

# 删除Entity-Type - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 删除Entity-Type |
| 方法 | DELETE |
| 路径 | /open-api/v1/entity-types/{type_name} |
| 说明 | 删除实体类型 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| Data | null | 删除成功后返回null |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-5-001 | 正常删除Entity-Type | 正常参数 | 删除成功 |
| ET-5-002 | 删除不存在的Entity-Type | 异常参数 | 验证 ErrNum=404 |
| ET-5-003 | 删除存在Entity的Entity-Type | 业务规则 | 验证 ErrNum=409 |

---

## 四、测试场景详细设计

---

### ET-5-001：正常删除Entity-Type（正常参数）

#### 设计思路

验证删除Entity-Type接口的基本功能：删除已存在的Entity-Type，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_delete", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 DELETE 请求到 `/open-api/v1/entity-types/{type_name}`
4. 验证响应状态码
5. 查询该Entity-Type确认已删除

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### ET-5-002：删除不存在的Entity-Type（异常参数）

#### 设计思路

验证删除不存在的Entity-Type时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity-Type "non_existent_delete" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/entity-types/non_existent_delete`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | non_existent_delete |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity-Type不存在的错误信息  
**Data**：null

---

### ET-5-003：删除存在Entity的Entity-Type（业务规则）

#### 设计思路

验证删除Entity-Type时，如果该类型下存在Entity，接口应返回冲突错误。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_has_entity", level=1
- 预先创建Entity：name="test_entity_has_type", type="test_type_has_entity"

#### 执行步骤

1. 创建Entity-Type
2. 创建使用该类型的Entity
3. 发送 DELETE 请求到 `/open-api/v1/entity-types/{type_name}`
4. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | test_type_has_entity |

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：Param Illegal: cannot delete entity type with associated entities  
**Data**：null

---

# 查询单个Entity-Type - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 查询单个Entity-Type |
| 方法 | GET |
| 路径 | /open-api/v1/entity-types/{type_name} |
| 说明 | 获取单个实体类型详情 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| type_name | string | 类型名 |
| description | string | 描述 |
| level | int | 层级级别 |
| create_time | int64 | 创建时间，Unix时间戳（秒） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-3-001 | 查询已存在的Entity-Type | 正常参数 | 返回完整信息 |
| ET-3-002 | 查询不存在的Entity-Type | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### ET-3-001：查询已存在的Entity-Type（正常参数）

#### 设计思路

验证查询单个Entity-Type接口的基本功能：查询已存在的Entity-Type，确认返回完整的类型信息。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_detail", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 GET 请求到 `/open-api/v1/entity-types/{type_name}`
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_detail" | Equals |
| level | 1 | Equals |
| create_time | 非空int64 | NotEmpty |

---

### ET-3-002：查询不存在的Entity-Type（异常参数）

#### 设计思路

验证查询不存在的Entity-Type时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity-Type "non_existent_type" 不存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types/non_existent_type`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | non_existent_type |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity-Type不存在的错误信息  
**Data**：null

---

# 查询Entity-Type列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 查询Entity-Type列表 |
| 方法 | GET |
| 路径 | /open-api/v1/entity-types |
| 说明 | 获取实体类型列表 |

---

## 二、接口参数说明

### 请求参数

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认1 |
| page_size | int | N | 每页条数，默认20，最大100 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []object | Entity-Type列表 |
| list[].type_name | string | 类型名 |
| list[].description | string | 描述 |
| list[].level | int | 层级级别 |
| list[].create_time | int64 | 创建时间 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-2-001 | 获取Entity-Type列表（默认参数） | 正常参数 | 默认分页参数 |
| ET-2-002 | 获取Entity-Type列表（自定义分页） | 正常参数 | 指定page和page_size |

---

## 四、测试场景详细设计

---

### ET-2-001：获取Entity-Type列表（默认参数）

#### 设计思路

验证查询Entity-Type列表接口的基本功能：使用默认分页参数获取列表，确认接口返回成功并返回列表数据。

#### 前提数据准备

- 确保存在至少一个Entity-Type

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types`
2. 验证响应状态码和返回结构

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 非空数组 | NotEmpty |
| list[0].type_name | 非空字符串 | NotEmpty |
| list[0].level | int类型 | NotEmpty |

> 注：当前列表接口不返回 pagination 字段

---

### ET-2-002：获取Entity-Type列表（自定义分页）

#### 设计思路

验证查询Entity-Type列表接口支持自定义分页参数：指定page和page_size，确认返回正确的数据范围。

#### 前提数据准备

- 确保存在至少3个Entity-Type

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types?page=1&page_size=2`
2. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| page | 1 |
| page_size | 2 |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组长度<=2 | LengthLessThanOrEqual(2) |

> 注：当前列表接口不返回 pagination 字段

---

# 更新Entity-Type - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 更新Entity-Type |
| 方法 | PATCH |
| 路径 | /open-api/v1/entity-types/{type_name} |
| 说明 | 更新实体类型描述 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | N | 类型描述 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| type_name | string | 类型名 |
| description | string | 描述 |
| level | int | 层级级别 |
| create_time | int64 | 创建时间，Unix时间戳（秒） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-4-001 | 更新Entity-Type描述 | 正常参数 | 更新description |
| ET-4-002 | 更新不存在的Entity-Type | 异常参数 | 验证 ErrNum=404 |
| ET-4-003 | 更新空描述 | 边界值 | description为空字符串 |

---

## 四、测试场景详细设计

---

### ET-4-001：更新Entity-Type描述（正常参数）

#### 设计思路

验证更新Entity-Type接口的基本功能：更新已存在的Entity-Type的描述，确认接口返回成功并返回更新后的信息。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_update", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 PATCH 请求到 `/open-api/v1/entity-types/{type_name}`，更新description
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

**Body 参数**：

```json
{
    "description": "更新后的描述"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_update" | Equals |
| description | "更新后的描述" | Equals |
| level | 1 | Equals（保持不变） |

---

### ET-4-002：更新不存在的Entity-Type（异常参数）

#### 设计思路

验证更新不存在的Entity-Type时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity-Type "non_existent_update" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entity-types/non_existent_update`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | non_existent_update |

**Body 参数**：

```json
{
    "description": "更新描述"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity-Type不存在的错误信息  
**Data**：null

---

### ET-4-003：更新空描述（边界值）

#### 设计思路

验证更新Entity-Type时传入空描述的场景，确认接口能正确处理空字符串。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_empty_desc", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 PATCH 请求到 `/open-api/v1/entity-types/{type_name}`，传入空description
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

**Body 参数**：

```json
{
    "description": ""
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_empty_desc" | Equals |
| description | "" | Equals |

---

