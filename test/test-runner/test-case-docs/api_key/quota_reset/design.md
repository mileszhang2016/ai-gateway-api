# 重置配额余额 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 重置配额余额 |
| 方法 | POST |
| 路径 | /open-api/v1/api-keys/{id}/quota-plan/reset |
| 说明 | 重置指定 API-Key 的配额余额，可选传入新的配额总量 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int64 | 否 | 重置后的配额总量，不传则按当前配额重置 |
| reason | string | 否 | 重置原因 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 标识 |
| previous_quota | int64 | 重置前配额 |
| new_quota | int64 | 重置后配额 |
| balance | object | 余额变更详情 |
| balance.previous_remaining | int64 | 重置前剩余量 |
| balance.new_remaining | int64 | 重置后剩余量 |
| balance.used | int64 | 当前已用量（重置后为 0） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-8-001 | 传入 quota 重置 | 正常参数 | 验证传入新配额重置成功 |
| AK-8-002 | 不传 quota 重置（按当前配额） | 正常参数 | 验证不传参数时按当前配额重置 |
| AK-8-003 | 重置不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-8-004 | 验证返回结构完整性 | 返回数据校验 | 验证 previous_quota、new_quota、balance 字段 |

---

## 四、测试场景详细设计

---

### AK-8-001：传入 quota 重置

#### 设计思路

验证传入新的配额总量后，配额被正确更新，balance 被重置。

#### 前提数据准备

- 先创建一个带配额计划（quota=100000000）的 API-Key

#### 执行步骤

1. 创建 API-Key（quota=100000000）
2. 发送 POST 请求，传入 `quota=50000000`
3. 验证返回的 previous_quota、new_quota 和 balance

#### 请求参数

```json
{
    "quota": 50000000,
    "reason": "手动调整"
}
```

#### 预期返回结果

**ErrNum**：200  

**Data 字段校验**：

| 字段 | 预期值 |
|------|--------|
| id | 与 API-Key ID 一致 |
| previous_quota | 100000000 |
| new_quota | 50000000 |
| balance.previous_remaining | 100000000 |
| balance.new_remaining | 50000000 |
| balance.used | 0 |

---

### AK-8-002：不传 quota 重置（按当前配额）

#### 设计思路

验证不传 quota 参数时，按当前配额值重置 balance。

#### 前提数据准备

- 先创建一个带配额计划（quota=100000000）的 API-Key

#### 请求参数

```json
{
    "reason": "月度重置"
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.previous_quota**：100000000  
**Data.new_quota**：100000000  
**Data.balance.used**：0

---

### AK-8-003：重置不存在的 API-Key

#### 设计思路

验证重置不存在的 ID 时返回 404。

#### 请求参数

```json
{
    "quota": 50000000
}
```

#### 预期返回结果

**ErrNum**：404

---

### AK-8-004：验证返回结构完整性

#### 设计思路

验证重置后返回的 Data 包含所有必要字段。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  

**Data 顶层键校验**：

| 键名 | 预期类型 |
|------|---------|
| id | string |
| previous_quota | number |
| new_quota | number |
| balance | object |
| balance.previous_remaining | number |
| balance.new_remaining | number |
| balance.used | number |