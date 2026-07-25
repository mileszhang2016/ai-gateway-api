# 更新Entity-Type - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 更新Entity-Type |
| 方法 | PATCH |
| 路径 | /open-api/v1/entity-types/{type_name} |
| 说明 | 更新实体类型描述 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | N | 类型描述 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| type_name | string | 类型名 |
| description | string | 描述 |
| level | int | 层级级别 |
| create_time | int64 | 创建时间，Unix时间戳（秒） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-4-001 | 更新Entity-Type描述 | 正常参数 | 更新description |
| ET-4-002 | 更新不存在的Entity-Type | 异常参数 | 验证 ErrNum=404 |
| ET-4-003 | 更新空描述 | 边界值 | description为空字符串 |

---

## 四、测试场景详细设计

---

### ET-4-001：更新Entity-Type描述（正常参数）

#### 设计思路

验证更新Entity-Type接口的基本功能：更新已存在的Entity-Type的描述，确认接口返回成功并返回更新后的信息。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_update", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 PATCH 请求到 `/open-api/v1/entity-types/{type_name}`，更新description
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

**Body 参数**：

```json
{
    "description": "更新后的描述"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_update" | Equals |
| description | "更新后的描述" | Equals |
| level | 1 | Equals（保持不变） |

---

### ET-4-002：更新不存在的Entity-Type（异常参数）

#### 设计思路

验证更新不存在的Entity-Type时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity-Type "non_existent_update" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entity-types/non_existent_update`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | non_existent_update |

**Body 参数**：

```json
{
    "description": "更新描述"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity-Type不存在的错误信息  
**Data**：null

---

### ET-4-003：更新空描述（边界值）

#### 设计思路

验证更新Entity-Type时传入空描述的场景，确认接口能正确处理空字符串。

#### 前提数据准备

- 预先创建Entity-Type：type_name="test_type_empty_desc", level=1

#### 执行步骤

1. 先创建Entity-Type
2. 提取返回的type_name
3. 发送 PATCH 请求到 `/open-api/v1/entity-types/{type_name}`，传入空description
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| type_name | 创建时返回的type_name |

**Body 参数**：

```json
{
    "description": ""
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "test_type_empty_desc" | Equals |
| description | "" | Equals |