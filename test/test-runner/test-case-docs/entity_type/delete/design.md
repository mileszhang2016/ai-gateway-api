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