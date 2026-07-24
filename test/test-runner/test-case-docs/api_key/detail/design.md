# 查询单个API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询单个API-Key |
| 方法 | GET |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 查询指定 API-Key 的详细信息，quota_plan 中包含 balance 字段 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### 返回数据字段

同创建接口返回，quota_plan 中包含 balance 字段。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-3-001 | 查询存在的 API-Key | 正常参数 | 验证返回完整信息，含 balance |
| AK-3-002 | 查询不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-3-003 | 验证返回字段完整性 | 返回数据校验 | 验证 quota_plan 含 balance 结构 |

---

## 四、测试场景详细设计

---

### AK-3-001：查询存在的 API-Key

#### 设计思路

验证查询已创建的 API-Key 时返回完整信息。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 GET 请求到 `/open-api/v1/api-keys/{id}`
3. 验证返回字段

#### 预期返回结果

**ErrNum**：200  
**Data.id**：与创建时返回的 ID 一致  
**Data.quota_plan.balance**：存在，包含 used 和 remaining

---

### AK-3-002：查询不存在的 API-Key

#### 设计思路

验证查询不存在的 ID 时返回 404。

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/nonexistent-id`
2. 验证返回 404

#### 预期返回结果

**ErrNum**：404  
**Data**：null

---

### AK-3-003：验证返回字段完整性

#### 设计思路

验证查询单个 API-Key 时，返回的 quota_plan 包含 balance 字段。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 执行步骤

1. 创建 API-Key（带 quota_plan）
2. 通过 ID 查询
3. 验证 quota_plan.balance 结构

#### 预期返回结果

**ErrNum**：200  

**Data.quota_plan.balance 校验**：

| 字段 | 预期类型 |
|------|---------|
| balance.used | number |
| balance.remaining | number |