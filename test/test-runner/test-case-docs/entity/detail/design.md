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