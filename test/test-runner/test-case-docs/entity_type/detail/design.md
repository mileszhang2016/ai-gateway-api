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