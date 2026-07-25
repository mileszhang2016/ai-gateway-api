# 全量更新Entity - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 全量更新Entity |
| 方法 | PUT |
| 路径 | /open-api/v1/entities/{id} |
| 说明 | 全量更新Entity，同创建接口Body参数 |

---

## 二、接口参数说明

### 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity标识 |

#### Body 参数

同创建Entity的Body参数。

**约束**：
- `type` 不可修改（创建后固定）
- `name` 全局唯一，不可与其他Entity冲突

### 返回数据字段

同创建接口返回（不含balance）。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ENT-4-001 | 全量更新Entity名称 | 正常参数 | 修改name |
| ENT-4-002 | 更新不存在的Entity | 异常参数 | 验证 ErrNum=404 |
| ENT-4-003 | 更新后名称与其他Entity冲突 | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### ENT-4-001：全量更新Entity名称（正常参数）

#### 设计思路

验证全量更新Entity接口的基本功能：更新已存在的Entity名称，确认接口返回成功。

#### 前提数据准备

- 预先创建Entity：name="test_entity_put", type="dep"

#### 执行步骤

1. 先创建Entity
2. 提取返回的id
3. 发送 PUT 请求到 `/open-api/v1/entities/{id}`，修改name
4. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | 创建Entity时返回的id |

**Body 参数**：

```json
{
    "name": "test_entity_put_updated",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_entity_put_updated" | Equals |
| type | "dep" | Equals |

---

### ENT-4-002：更新不存在的Entity（异常参数）

#### 设计思路

验证更新不存在的Entity时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保Entity "non_existent_put" 不存在

#### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/entities/non_existent_put`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | non_existent_put |

**Body 参数**：

```json
{
    "name": "test_entity_update",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含Entity不存在的错误信息  
**Data**：null

---

### ENT-4-003：更新后名称与其他Entity冲突（业务规则）

#### 设计思路

验证更新后名称与其他Entity冲突时，接口应返回业务错误。

#### 前提数据准备

- 预先创建Entity1：name="test_entity_conflict1", type="dep"
- 预先创建Entity2：name="test_entity_conflict2", type="dep"

#### 执行步骤

1. 创建两个Entity
2. 用Entity1的id发送 PUT 请求，将name改为Entity2的name
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| id | Entity1的id |

**Body 参数**：

```json
{
    "name": "test_entity_conflict2",
    "type": "dep"
}
```

#### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含数据库约束异常的错误信息（UNIQUE constraint failed）  
**Data**：null