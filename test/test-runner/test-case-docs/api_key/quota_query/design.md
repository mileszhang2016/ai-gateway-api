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
| id | string | 是 | API-Key 唯一标识 |

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
| AK-7-003 | 验证 balance 字段结构 | 返回数据校验 | 验证 used 和 remaining 字段 |

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

---

### AK-7-003：验证 balance 字段结构

#### 设计思路

验证 balance 对象包含 used 和 remaining 两个字段，且为数字类型。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  

**Data.balance 校验**：

| 字段 | 预期类型 | 预期值 |
|------|---------|--------|
| balance.used | number | 0 |
| balance.remaining | number | = quota（初始剩余=配额） |