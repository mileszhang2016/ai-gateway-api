# 部分更新Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 部分更新Entity |
| 方法 | PATCH |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 部分更新Entity，仅传需修改字段 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

同创建Entity的Body参数，仅传需修改字段。

### 返回数据字段

同创建接口返回（不含balance）。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-5-001 | 部分更新Entity名称 | 正常参数 | 仅传name字段 |
| ENT-5-002 | 更新不存在的Entity | 异常参数 | 验证 ErrNum=404 |
| ENT-5-003 | 更新配额计划 | 正常参数 | 修改quota_plan |

---

## 四、测试场景详细设计

---

### ENT-5-001：部分更新Entity名称（正常参数）

#### 设计思路

验证部分更新Entity接口的基本功能：仅传入name字段进行更新，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_patch", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PATCH 请求到 `/open-api/v1/entities/{id}`，仅传name字段
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "name": "test_entity_patch_updated"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_entity_patch_updated" | Equals |
| type | "dep" | Equals（保持不变） |

---

### ENT-5-002：更新不存在的Entity（异常参数）

#### 设计思路

验证部分更新不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_patch" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entities/non_existent_patch`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_patch |

**Body 参数**：

```json
{
    "name": "test_entity_patch_update"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

### ENT-5-003：更新配额计划（正常参数）

#### 设计思路

验证部分更新Entity时修改配额计划的场景，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_quota_patch", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PATCH 请求到 `/open-api/v1/entities/{id}`，更新quota_plan
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "quota_plan": {
        "unlimited": false,
        "quota": 50000000,
        "unit": "total_token",
        "reset_period": "weekly"
    }
}
```

#### 预期返回结果

**ErrNum**：500  
**ErrMsg**：system error  
**Data**：null

> 注：当前部分更新 quota_plan 功能存在系统错误，无法正常更新