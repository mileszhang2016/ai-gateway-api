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