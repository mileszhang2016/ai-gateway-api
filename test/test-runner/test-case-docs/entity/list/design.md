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