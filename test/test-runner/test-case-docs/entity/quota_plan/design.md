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