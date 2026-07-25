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