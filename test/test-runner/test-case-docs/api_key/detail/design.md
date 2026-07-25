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
| id | string | **是** | API-Key 唯一标识（UUID 格式） |

### 返回数据字段

返回完整的 API-Key 对象，主要字段：

| 字段路径 | 类型 | 说明 |
|---------|------|------|
| id | string | API-Key 唯一标识 |
| description | string | 描述 |
| expired_time | int64 | 过期时间戳 |
| enabled | bool | 是否启用 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 模型白名单 |
| subnet | []string | 子网白名单 |
| entity_id | string | 关联 Entity ID |
| quota_plan | object | 配额计划（含 balance） |
| quota_plan.balance | object | 配额使用情况 |
| quota_plan.balance.used | number | 已使用量 |
| quota_plan.balance.remaining | number | 剩余量 |
| rate_limit_policy | object | 限流策略 |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

**约束**：
- `id` 必须为有效的 UUID 格式
- 被查询的 API-Key 必须存在，否则返回 404
- 查询接口返回的 quota_plan 中额外包含 balance 字段（创建接口不返回）

---

## 三、测试场景总览

### 正常参数（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-3-001 | 查询基本 API-Key（仅 description） | 正常参数 | id | 验证查询成功，返回 id 和 description 一致 |
| AK-3-002 | 查询含完整配置的 API-Key | 正常参数 | id | 验证查询含 quota_plan、rate_limit_policy、entity_id 的 Key，返回完整数据 |

### 必填校验（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-3-003 | 查询路径缺少 ID（空路径） | 必填校验 | id | 验证缺少 id 参数时路由不匹配，返回 404 |

### 边界值（1）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-3-004 | 查询超长 ID（256 字符） | 边界值 | id | 验证超长 ID 触发参数校验，返回 422 |

### 异常参数（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-3-005 | 查询不存在的 API-Key | 异常参数 | id | 验证不存在的 ID 返回 404 |
| AK-3-006 | 查询无效 UUID 格式的 ID | 异常参数 | id | 验证 ID 格式无效时返回 404 |

### 返回数据校验（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-3-007 | 返回顶层字段完整性校验 | 返回数据 | 全部顶层字段 | 验证返回包含 id、description、enabled、quota_plan 等所有顶层字段 |
| AK-3-008 | 返回 quota_plan.balance 结构校验 | 返回数据 | quota_plan.balance | 验证 balance 包含 used 和 remaining 字段，类型正确 |

---

## 四、参数覆盖矩阵

| 参数层级 | 参数路径 | 覆盖方式 |
|---------|---------|---------|
| URI参数 | `id` | 正常参数(AK-3-001,AK-3-002) + 必填校验(AK-3-003) + 边界值(AK-3-004) + 异常参数(AK-3-005,AK-3-006) |
| 返回字段 | 顶层字段 | 返回数据校验(AK-3-007) |
| 返回字段 | `quota_plan.balance` | 返回数据校验(AK-3-008) |

---

## 五、测试场景详细设计

---

### 正常参数

---

### AK-3-001：查询基本 API-Key

#### 设计思路

验证查询仅含 description 的 API-Key，确认返回数据中 id 和 description 与创建时一致。

#### 前提数据准备

- 先创建一个基本 API-Key（仅传入 description）

#### 执行步骤

1. 调用创建 API-Key 接口，传入 `{"description": "detail-test-001"}`
2. 获取创建的 API-Key ID
3. 发送 GET 请求到 `/open-api/v1/api-keys/{id}`
4. 验证 ErrNum=200
5. 验证 Data.id 与创建时一致
6. 验证 Data.description 与创建时一致
7. 验证 Data.quota_plan 字段不为 null

#### 请求参数

```
GET /open-api/v1/api-keys/{预先创建的API-Key ID}
```

#### 预期返回结果

**ErrNum**：200  
**Data.id**：与创建时返回的 ID 一致  
**Data.description**：与创建时传入的 description 一致  
**Data.quota_plan**：不为 null

---

### AK-3-002：查询含完整配置的 API-Key

#### 设计思路

验证查询带有 quota_plan、rate_limit_policy 和 entity_id 的 API-Key，确认所有字段被正确返回。

#### 前提数据准备

1. 创建 EntityType（如 `{"type_name": "detail-test-etype", "level": 1}`）
2. 创建 Entity（基于上述 EntityType）
3. 创建 API-Key，传入完整配置

#### 执行步骤

1. 创建 EntityType
2. 创建 Entity
3. 创建 API-Key，传入完整配置：
   ```json
   {
       "description": "detail-test-full-config",
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
4. 发送 GET 请求查询该 API-Key
5. 验证 ErrNum=200
6. 验证 quota_plan、rate_limit_policy、entity_id 字段均存在且正确

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan**：包含 quota=100000、unit="token"、reset_period="daily"  
**Data.rate_limit_policy**：包含 enabled=true 和 rules 结构  
**Data.entity_id**：与创建时传入的 entity_id 一致

---

### 必填校验

---

### AK-3-003：查询路径缺少 ID

#### 设计思路

验证 GET 请求路径中缺少 ID 参数时，路由无法匹配，返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/`（路径末尾无 ID）
2. 验证返回 404

#### 请求参数

```
GET /open-api/v1/api-keys/
```

#### 预期返回结果

**ErrNum**：404

---

### 边界值

---

### AK-3-004：查询超长 ID（256 字符）

#### 设计思路

验证传入超长 ID（256 字符）时，接口触发参数长度校验，返回 422。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 构造一个 256 字符的 ID 字符串
2. 发送 GET 请求到 `/open-api/v1/api-keys/{超长ID}`
3. 验证返回 422（非 500）

#### 请求参数

```
GET /open-api/v1/api-keys/aaaa...（256个'a'字符）
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "ID" 或 "Param Illegal"

---

### 异常参数

---

### AK-3-005：查询不存在的 API-Key

#### 设计思路

验证传入一个不存在的 ID 时，接口返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/nonexistent-id-000000`
2. 验证返回 404

#### 请求参数

```
GET /open-api/v1/api-keys/nonexistent-id-000000
```

#### 预期返回结果

**ErrNum**：404

---

### AK-3-006：查询无效 UUID 格式的 ID

#### 设计思路

验证传入非 UUID 格式的 ID 时，接口返回 404（数据库查询不到该记录）。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/api-keys/invalid-format`
2. 验证返回 404

#### 请求参数

```
GET /open-api/v1/api-keys/invalid-format
```

#### 预期返回结果

**ErrNum**：404

---

### 返回数据校验

---

### AK-3-007：返回顶层字段完整性校验

#### 设计思路

验证查询单个 API-Key 时，返回数据包含所有顶层字段，且类型正确。

#### 前提数据准备

- 先创建一个带完整配置的 API-Key

#### 执行步骤

1. 创建 API-Key（含 quota_plan 和 rate_limit_policy）
2. 发送 GET 请求查询该 API-Key
3. 验证 ErrNum=200
4. 验证以下顶层字段存在且类型正确

#### 预期返回结果

**ErrNum**：200

| 字段 | 预期类型 | 校验方式 |
|------|---------|---------|
| id | string | 非空 |
| description | string | 非空 |
| expired_time | number | 存在 |
| enabled | bool | 存在 |
| unlimited_quota | bool | 存在 |
| models | array | 存在 |
| subnet | array | 存在 |
| entity_id | string | 存在 |
| quota_plan | object | 非 null |
| rate_limit_policy | object | 非 null |

---

### AK-3-008：返回 quota_plan.balance 结构校验

#### 设计思路

验证查询单个 API-Key 时，返回的 quota_plan 中包含 balance 字段（used 和 remaining），且类型正确。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 执行步骤

1. 创建 API-Key，传入：
   ```json
   {
       "description": "detail-balance-test",
       "quota_plan": {
           "unlimited": false,
           "quota": 100000000,
           "unit": "total_token",
           "reset_period": "monthly"
       }
   }
   ```
2. 发送 GET 请求查询该 API-Key
3. 验证 ErrNum=200
4. 验证 quota_plan.balance 存在且为 object
5. 验证 balance.used 为 number 类型
6. 验证 balance.remaining 为 number 类型

#### 预期返回结果

**ErrNum**：200

**Data.quota_plan.balance 校验**：

| 字段 | 预期类型 | 校验方式 |
|------|---------|---------|
| balance | object | 存在且非 null |
| balance.used | number | 类型为 float64 |
| balance.remaining | number | 类型为 float64 |