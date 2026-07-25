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