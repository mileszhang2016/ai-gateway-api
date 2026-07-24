# 删除API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 删除API-Key |
| 方法 | DELETE |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 删除指定 API-Key，级联删除其专属配额和限流策略 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | **是** | API-Key 唯一标识（UUID格式） |

### 返回数据

| 参数名 | 类型 | 说明 |
|--------|------|------|
| Data | null | 删除成功返回 Data 为 null 或空 |

**约束**：
- `id` 必须为有效的 UUID 格式
- 被删除的 API-Key 必须存在
- 删除后不可恢复，再次查询应返回 404
- 级联删除关联的配额计划（quota_plan）和限流策略（rate_limit_policy）

---

## 三、测试场景总览

### 正常参数（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-001 | 删除基本 API-Key（仅 description） | 正常参数 | id | 验证删除成功，返回 Data=null |
| AK-6-002 | 删除含完整配置的 API-Key（quota_plan + rate_limit_policy + entity_id） | 正常参数 | id | 验证级联删除，关联配置不残留 |

### 必填校验（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-003 | 删除路径缺少 ID（空路径） | 必填校验 | id | 验证缺少 id 参数时路由不匹配，返回 404 |

### 边界值（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-004 | 删除超长 ID（256 字符） | 边界值 | id | 验证超长 ID 触发参数校验，返回 422 |

### 异常参数（3）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-005 | 删除不存在的 API-Key | 异常参数 | id | 验证不存在的 ID 返回 404 |
| AK-6-006 | 删除无效 UUID 格式的 ID | 异常参数 | id | 验证 ID 格式校验，返回 404 |
| AK-6-007 | 双重删除（对已删除的 Key 再次删除） | 异常参数 | id | 验证幂等性，第二次删除返回 404 |

### 返回数据校验（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-008 | 删除成功返回结构校验 | 返回数据 | Data | 验证 ErrNum=200、ErrMsg="success"、Data=null |

### 业务规则（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-6-009 | 删除后查询返回 404 | 业务规则 | id | 验证删除后 GET 接口不可再查询到该 Key |

---

## 四、参数覆盖矩阵

| 参数层级 | 参数路径 | 覆盖方式 |
|---------|---------|---------|
| URI参数 | `id` | 正常参数(AK-6-001,AK-6-002) + 必填校验(AK-6-003) + 边界值(AK-6-004) + 异常参数(AK-6-005,AK-6-006,AK-6-007) |
| 返回字段 | `Data` | 返回数据校验(AK-6-008) |

---

## 五、测试场景详细设计

---

### 正常参数

---

### AK-6-001：删除基本 API-Key

#### 设计思路

验证删除仅含 description 的 API-Key，确认删除成功返回 Data=null。

#### 前提数据准备

- 先创建一个基本 API-Key（仅传入 description）

#### 执行步骤

1. 调用创建 API-Key 接口，传入 `{"description": "delete-test-001"}`
2. 获取创建的 API-Key ID
3. 发送 DELETE 请求到 `/open-api/v1/api-keys/{id}`
4. 验证 ErrNum=200
5. 验证返回 Data=null

#### 请求参数

```
DELETE /open-api/v1/api-keys/{预先创建的API-Key ID}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AK-6-002：删除含完整配置的 API-Key（级联删除）

#### 设计思路

验证删除一个带有 quota_plan、rate_limit_policy 和 entity_id 的 API-Key，确认级联删除成功，关联的配额和限流策略不残留。

#### 前提数据准备

1. 创建 EntityType（如 `{"name": "test-et-delete", "level": 1}`）
2. 创建 Entity（基于上述 EntityType）
3. 创建 API-Key，传入 `quota_plan`、`rate_limit_policy` 和 `entity_id`

#### 执行步骤

1. 创建 EntityType
2. 创建 Entity
3. 创建 API-Key，传入完整配置：
   ```json
   {
       "description": "delete-test-full-config",
       "quota_plan": {
           "unlimited": false,
           "quota": 100000,
           "unit": "token",
           "reset_period": "daily"
       },
       "rate_limit_policy": {
           "enabled": true,
           "rules": {
               "tpm": [{"name": "test", "model": "*", "window_minutes": 1, "max_tokens": 1000, "step_minutes": 1}]
           }
       },
       "entity_id": "<entity_id>"
   }
   ```
4. 发送 DELETE 请求删除该 API-Key
5. 验证 ErrNum=200
6. 发送 GET 请求确认 API-Key 已删除（返回 404）

#### 预期返回结果

**DELETE**：ErrNum=200，Data=null  
**GET**：ErrNum=404

---

### 必填校验

---

### AK-6-003：删除路径缺少 ID

#### 设计思路

验证 DELETE 请求路径中缺少 ID 参数时，路由无法匹配，返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/api-keys/`（路径末尾无 ID）
2. 验证返回 404

#### 请求参数

```
DELETE /open-api/v1/api-keys/
```

#### 预期返回结果

**ErrNum**：404

---

### 边界值

---

### AK-6-004：删除超长 ID（256 字符）

#### 设计思路

验证传入超长 ID（256 字符）时，接口触发参数长度校验，返回 422。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 构造一个 256 字符的 ID 字符串
2. 发送 DELETE 请求到 `/open-api/v1/api-keys/{超长ID}`
3. 验证返回 422（非 500）

#### 请求参数

```
DELETE /open-api/v1/api-keys/aaaa...（256个'a'字符）
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "ID" 或 "Param Illegal"

---

### 异常参数

---

### AK-6-005：删除不存在的 API-Key

#### 设计思路

验证传入一个不存在的 ID 时，接口返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/api-keys/nonexistent-id-000000`
2. 验证返回 404

#### 请求参数

```
DELETE /open-api/v1/api-keys/nonexistent-id-000000
```

#### 预期返回结果

**ErrNum**：404

---

### AK-6-006：删除无效 UUID 格式的 ID

#### 设计思路

验证传入非 UUID 格式的 ID 时，接口返回 404（数据库查询不到该记录）。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/api-keys/invalid-format`
2. 验证返回 404

#### 请求参数

```
DELETE /open-api/v1/api-keys/invalid-format
```

#### 预期返回结果

**ErrNum**：404

---

### AK-6-007：双重删除（删除已删除的 Key）

#### 设计思路

验证对同一个 API-Key 进行两次删除操作，第一次成功，第二次返回 404（幂等性校验）。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 第一次 DELETE 请求，验证 ErrNum=200
3. 第二次 DELETE 请求（相同 ID），验证 ErrNum=404

#### 请求参数

```
第一次：DELETE /open-api/v1/api-keys/{id}
第二次：DELETE /open-api/v1/api-keys/{id}（相同 ID）
```

#### 预期返回结果

**第一次 DELETE**：ErrNum=200，Data=null  
**第二次 DELETE**：ErrNum=404

---

### 返回数据校验

---

### AK-6-008：删除成功返回结构校验

#### 设计思路

验证删除成功后，返回的响应结构完整且正确：ErrNum=200、ErrMsg="success"、Data=null。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 DELETE 请求
3. 验证 ErrNum=200
4. 验证 ErrMsg="success"
5. 验证 Data 为 null（JSON 中的 null）

#### 请求参数

```
DELETE /open-api/v1/api-keys/{预先创建的API-Key ID}
```

#### 预期返回结果

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| ErrNum | 200 | Equals |
| ErrMsg | "success" | Equals |
| Data | null | Equals |

---

### 业务规则

---

### AK-6-009：删除后查询返回 404

#### 设计思路

验证删除 API-Key 后，通过 GET 接口查询该 ID 应返回 404，确认业务规则"删除后不可恢复"。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 DELETE 请求删除该 API-Key
3. 发送 GET 请求查询该 ID
4. 验证 DELETE 返回 200
5. 验证 GET 返回 404

#### 请求参数

```
DELETE /open-api/v1/api-keys/{id}
GET /open-api/v1/api-keys/{id}
```

#### 预期返回结果

**DELETE 响应**：ErrNum=200  
**GET 响应**：ErrNum=404