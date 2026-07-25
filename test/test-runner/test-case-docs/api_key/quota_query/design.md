# 查询配额计划 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询配额计划（含实时余额） |
| 方法 | GET |
| 路径 | /open-api/v1/api-keys/{id}/quota-plan |
| 说明 | 查询指定 API-Key 的配额计划，含实时余额（balance） |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识（最大255字符） |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| unlimited | bool | 是否无限配额 |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 |
| quota | int64 | 配额总量 |
| unit | string | 配额单位 |
| reset_period | string | 配额重置周期 |
| balance | object | 余额状态 |
| balance.used | int64 | 已用量 |
| balance.remaining | int64 | 剩余量 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-7-001 | 查询有配额的 API-Key | 正常参数 | 验证返回完整配额信息 |
| AK-7-002 | 查询不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-7-003 | 验证 balance 字段结构 | 返回数据校验 | 验证 used/remaining 类型和关系 |
| AK-7-004 | 查询无限配额的 API-Key | 正常参数 | 验证 unlimited=true |
| AK-7-005 | 查询无配额计划的 API-Key | 正常参数 | 验证返回 nil |
| AK-7-006 | id 超长（>255字符） | 边界值 | 验证 ErrNum=422 |
| AK-7-007 | 验证返回字段完整性 | 返回数据校验 | 验证所有字段存在且类型正确 |

---

## 四、测试场景详细设计

---

### AK-7-001：查询有配额的 API-Key

#### 设计思路

验证查询配额计划时返回完整信息。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 执行步骤

1. 创建 API-Key（带 quota_plan）
2. 发送 GET 请求到 `/open-api/v1/api-keys/{id}/quota-plan`
3. 验证返回字段

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 |
|------|--------|
| unlimited | false |
| quota | 100000000 |
| unit | "total_token" |
| reset_period | "monthly" |
| balance.used | 0 |
| balance.remaining | 100000000 |

---

### AK-7-002：查询不存在的 API-Key

#### 设计思路

验证查询不存在的 ID 时返回 404。

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 "API-Key" 和 "not exist"

---

### AK-7-003：验证 balance 字段结构

#### 设计思路

验证 balance 对象包含 used 和 remaining 两个字段，且为数字类型，used + remaining = quota。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data.balance 校验**：

| 字段 | 预期类型 | 预期值 |
|------|---------|--------|
| balance.used | number | ≥ 0 |
| balance.remaining | number | ≥ 0 |
| used + remaining | — | = quota |

---

### AK-7-004：查询无限配额的 API-Key

#### 设计思路

验证查询 unlimited=true 的 API-Key 时，返回 unlimited=true。

#### 前提数据准备

- 先创建一个 unlimited_quota=true 的 API-Key

#### 执行步骤

1. 创建 API-Key（unlimited_quota=true, quota_plan.unlimited=true）
2. 发送 GET 请求
3. 验证 unlimited=true

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.unlimited**：true

---

### AK-7-005：查询无配额计划的 API-Key

#### 设计思路

验证查询未设置 quota_plan 的 API-Key 时，系统自动创建默认配额计划（unlimited=true, quota=0, unit="total_token", reset_period="never"）。

#### 前提数据准备

- 先创建一个不传 quota_plan 的 API-Key

#### 执行步骤

1. 创建 API-Key（仅传 description）
2. 发送 GET 请求
3. 验证返回默认配额计划

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 |
|------|--------|
| unlimited | true |
| quota | 0 |
| unit | "total_token" |
| reset_period | "never" |
| balance.used | 0 |
| balance.remaining | 0 |

---

### AK-7-006：id 超长（>255字符）

#### 设计思路

验证 URI 参数 id 超过 255 字符时返回参数错误。

#### 请求参数

URI：`/open-api/v1/api-keys/<256字符>/quota-plan`

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "id" 和 "invalid"

---

### AK-7-007：验证返回字段完整性

#### 设计思路

验证返回数据结构包含所有字段（unlimited, pass_when_no_enough_quota, quota, unit, reset_period, balance）且类型正确。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 必填字段校验**：

| 键名 | 预期类型 |
|------|---------|
| unlimited | bool |
| pass_when_no_enough_quota | bool |
| quota | number（>0） |
| unit | string（非空） |
| reset_period | string（非空） |
| balance | object（非空） |
| balance.used | number |
| balance.remaining | number |