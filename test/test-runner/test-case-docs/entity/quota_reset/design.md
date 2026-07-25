# 重置配额余额 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 重置配额余额 |
| 方法 | POST |
| 路径 | /open-api/v1/entities/{id}/quota-plan/reset |
| 说明 | 重置Entity的配额余额 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int64 | N | 重置后的配额总量，不传则按当前quota重置 |
| reason | string | N | 重置原因，用于审计 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity标识 |
| previous_quota | int64 | 重置前配额 |
| new_quota | int64 | 重置后配额 |
| balance | object | 余额变更详情 |
| balance.previous_remaining | int64 | 重置前剩余量 |
| balance.new_remaining | int64 | 重置后剩余量 |
| balance.used | int64 | 当前已用量，重置后为0 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-8-001 | 重置配额余额（不传新配额） | 正常参数 | 按当前quota重置 |
| ENT-8-002 | 重置配额余额（传新配额） | 正常参数 | 更新quota并重置 |

---

## 四、测试场景详细设计

---

### ENT-8-001：重置配额余额（不传新配额）

#### 设计思路

验证重置配额余额接口的基本功能：不传新配额，按当前quota重置余额，确认接口返回成功并返回余额变更详情。

#### 前提数据准备

- 预先创建Entity：name="test_entity_reset1", type="dep"，且配置非无限配额

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`，不传quota
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "reason": "月度重置"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 创建时返回的id | Equals |
| previous_quota | > 0 | GreaterThan(0) |
| new_quota | == previous_quota | Equals |
| balance.used | 0 | Equals |

---

### ENT-8-002：重置配额余额（传新配额）

#### 设计思路

验证重置配额余额时传入新配额的场景：更新quota并同步重置余额，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_reset2", type="dep"，且配置非无限配额

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`，传入新quota
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "quota": 50000000,
    "reason": "调整配额"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 创建时返回的id | Equals |
| new_quota | 50000000 | Equals |
| balance.new_remaining | 50000000 | Equals |
| balance.used | 0 | Equals |