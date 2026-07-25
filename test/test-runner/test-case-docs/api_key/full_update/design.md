# 全量更新API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 全量更新API-Key |
| 方法 | PUT |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 全量替换指定 API-Key 的所有字段，description 为必填项 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | **是** | API-Key 唯一标识（UUID 格式） |

### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | **是** | 描述文字，长度 < 512 |
| expired_time | int64 | 否 | 过期时间戳，-1 表示永不过期 |
| enabled | bool | 否 | 是否启用 |
| unlimited_quota | bool | 否 | 是否无限配额 |
| models | []string | 否 | 模型白名单 |
| subnet | []string | 否 | 子网白名单（CIDR 格式） |
| entity_id | string | 否 | 关联 Entity ID |
| quota_plan | object | 否 | 配额计划（含 unlimited、quota、unit、reset_period 等） |
| quota_plan.unlimited | bool | 否 | 是否无限配额 |
| quota_plan.pass_when_no_enough_quota | bool | 否 | 配额不足时是否放行 |
| quota_plan.quota | int64 | 否 | 配额值 |
| quota_plan.unit | string | 否 | 配额单位 |
| quota_plan.reset_period | string | 否 | 重置周期 |
| rate_limit_policy | object | 否 | 限流策略 |
| rate_limit_policy.enabled | bool | 否 | 是否启用限流 |
| rate_limit_policy.rules | object | 否 | 限流规则（含 tpm、rpm、max_concurrency） |

**约束**：
- `id` 必须为有效的 UUID 格式
- 被更新的 API-Key 必须存在，否则返回 404
- `description` 为必填，不能为空，长度 < 512
- `expired_time` 必须 >= 当前时间，或为 -1（永不过期）
- `subnet` 中的 CIDR 必须格式正确
- `rate_limit_policy.enabled=true` 时必须提供有效的 rules

---

## 三、测试场景总览

### 正常参数（9）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-4-001 | 全量更新 description | 正常参数 | description | 验证 description 更新成功 |
| AK-4-002 | 全量更新 expired_time | 正常参数 | expired_time | 验证过期时间更新 |
| AK-4-003 | 全量更新 enabled 状态 | 正常参数 | enabled | 验证启用/禁用切换 |
| AK-4-004 | 全量更新 unlimited_quota | 正常参数 | unlimited_quota | 验证无限配额开关 |
| AK-4-005 | 全量更新 models | 正常参数 | models | 验证模型白名单更新 |
| AK-4-006 | 全量更新 subnet | 正常参数 | subnet | 验证子网白名单更新 |
| AK-4-007 | 全量更新 quota_plan | 正常参数 | quota_plan | 验证配额计划更新 |
| AK-4-008 | 全量更新 rate_limit_policy | 正常参数 | rate_limit_policy | 验证限流策略更新 |
| AK-4-009 | 全量更新所有字段 | 正常参数 | 全部字段 | 验证一次性更新所有字段 |

### 必填校验（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-4-010 | 缺少 description（空 Body） | 必填校验 | description | 验证 ErrNum=422 |
| AK-4-011 | description 为空字符串 | 必填校验 | description | 验证 ErrNum=422 |

### 边界值（4）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-4-012 | description 最大合法长度（511） | 边界值 | description | 验证 511 字符通过 |
| AK-4-013 | expired_time=-1（永不过期） | 边界值 | expired_time | 验证 -1 设置成功 |
| AK-4-014 | 更新超长 ID（256 字符） | 边界值 | id | 验证触发参数校验，返回 422 |
| AK-4-015 | 更新空 models 数组 | 边界值 | models | 验证空数组的默认处理 |

### 异常参数（5）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-4-016 | 更新不存在的 API-Key | 异常参数 | id | 验证返回 404 |
| AK-4-017 | 更新无效 UUID 格式的 ID | 异常参数 | id | 验证返回 404 |
| AK-4-018 | expired_time 为过去时间 | 异常参数 | expired_time | 验证返回 422 |
| AK-4-019 | description 超长（512） | 异常参数 | description | 验证返回 422 |
| AK-4-020 | subnet 为无效 CIDR | 异常参数 | subnet | 验证返回 422 |

### 返回数据校验（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-4-021 | 全量更新返回结构校验 | 返回数据 | 返回字段 | 验证 ErrNum=200，返回完整 API-Key 对象 |
| AK-4-022 | 全量更新后 GET 验证 | 返回数据 | 全部字段 | 验证更新后 GET 查询结果与 PUT 返回一致 |

---

## 四、参数覆盖矩阵

| 参数层级 | 参数路径 | 覆盖方式 |
|---------|---------|---------|
| URI参数 | `id` | 正常参数(AK-4-009) + 边界值(AK-4-014) + 异常参数(AK-4-016,AK-4-017) |
| 顶层 | `description` | 正常参数(AK-4-001) + 必填校验(AK-4-010,AK-4-011) + 边界值(AK-4-012) + 异常参数(AK-4-019) |
| 顶层 | `expired_time` | 正常参数(AK-4-002) + 边界值(AK-4-013) + 异常参数(AK-4-018) |
| 顶层 | `enabled` | 正常参数(AK-4-003) |
| 顶层 | `unlimited_quota` | 正常参数(AK-4-004) |
| 顶层 | `models` | 正常参数(AK-4-005) + 边界值(AK-4-015) |
| 顶层 | `subnet` | 正常参数(AK-4-006) + 异常参数(AK-4-020) |
| 子结构 | `quota_plan` | 正常参数(AK-4-007) |
| 子结构 | `rate_limit_policy` | 正常参数(AK-4-008) |
| 返回字段 | 全部字段 | 返回数据校验(AK-4-021,AK-4-022) |

---

## 五、测试场景详细设计

---

### 正常参数

---

### AK-4-001：全量更新 description

#### 设计思路

验证 PUT 全量更新时 description 字段被正确更新。

#### 前提数据准备

- 先创建一个 API-Key（description="original-desc"）

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，修改 description 为 "updated-desc"
3. 验证 ErrNum=200
4. 验证 Data.description="updated-desc"

#### 请求参数

```json
{
    "description": "updated-desc"
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.description**："updated-desc"

---

### AK-4-002：全量更新 expired_time

#### 设计思路

验证 PUT 更新 expired_time 生效。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 构造未来时间戳（当前时间 + 365 天）
3. 发送 PUT 请求，传入 description + expired_time
4. 验证 ErrNum=200
5. 验证 Data.expired_time 等于传入值

#### 请求参数

```json
{
    "description": "expire-test",
    "expired_time": <未来时间戳>
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.expired_time**：等于传入的未来时间戳

---

### AK-4-003：全量更新 enabled 状态

#### 设计思路

验证 PUT 可以切换 enabled 状态（禁用后重新启用）。

#### 前提数据准备

- 先创建一个 API-Key（默认 enabled=true）

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 enabled=false
3. 验证 ErrNum=200，Data.enabled=false
4. 再次发送 PUT 请求，设置 enabled=true
5. 验证 ErrNum=200，Data.enabled=true

#### 请求参数

第一次：
```json
{
    "description": "enabled-test",
    "enabled": false
}
```

第二次：
```json
{
    "description": "enabled-test",
    "enabled": true
}
```

#### 预期返回结果

**ErrNum**：200  
**第一次**：Data.enabled=false  
**第二次**：Data.enabled=true

---

### AK-4-004：全量更新 unlimited_quota

#### 设计思路

验证 PUT 更新无限配额开关。

#### 前提数据准备

- 先创建一个 API-Key（默认 unlimited_quota=false）

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 unlimited_quota=true
3. 验证 ErrNum=200
4. 验证 Data.unlimited_quota=true

#### 请求参数

```json
{
    "description": "unlimited-test",
    "unlimited_quota": true
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.unlimited_quota**：true

---

### AK-4-005：全量更新 models

#### 设计思路

验证 PUT 更新模型白名单。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 models=["gpt-4", "gpt-3.5"]
3. 验证 ErrNum=200
4. 验证 Data.models 包含设置的模型

#### 请求参数

```json
{
    "description": "models-test",
    "models": ["gpt-4", "gpt-3.5"]
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.models**：包含 "gpt-4" 和 "gpt-3.5"

---

### AK-4-006：全量更新 subnet

#### 设计思路

验证 PUT 更新子网白名单。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 subnet=["192.168.1.0/24"]
3. 验证 ErrNum=200
4. 验证 Data.subnet 包含设置的 CIDR

#### 请求参数

```json
{
    "description": "subnet-test",
    "subnet": ["192.168.1.0/24"]
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.subnet**：包含 "192.168.1.0/24"

---

### AK-4-007：全量更新 quota_plan

#### 设计思路

验证 PUT 更新配额计划。

#### 前提数据准备

- 先创建一个 API-Key（不带 quota_plan）

#### 执行步骤

1. 创建 API-Key（仅 description），获取 ID
2. 发送 PUT 请求，传入完整的 quota_plan
3. 验证 ErrNum=200
4. 验证 Data.quota_plan 不为 null

#### 请求参数

```json
{
    "description": "quota-plan-test",
    "quota_plan": {
        "unlimited": false,
        "quota": 50000,
        "unit": "token",
        "reset_period": "daily"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan**：不为 null，包含 quota、unit、reset_period

---

### AK-4-008：全量更新 rate_limit_policy

#### 设计思路

验证 PUT 更新限流策略。

#### 前提数据准备

- 先创建一个 API-Key（不带 rate_limit_policy）

#### 执行步骤

1. 创建 API-Key（仅 description），获取 ID
2. 发送 PUT 请求，传入完整的 rate_limit_policy
3. 验证 ErrNum=200
4. 验证 Data.rate_limit_policy 不为 null

#### 请求参数

```json
{
    "description": "rate-limit-test",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [{"name": "t1", "model": "*", "window_minutes": 1, "max_tokens": 1000, "step_minutes": 1}]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy**：不为 null，enabled=true

---

### AK-4-009：全量更新所有字段

#### 设计思路

验证 PUT 一次性更新所有字段，确认全量替换语义。

#### 前提数据准备

- 先创建一个基本 API-Key（仅 description）

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，传入所有字段
3. 验证 ErrNum=200
4. 验证所有字段更新为传入值

#### 请求参数

```json
{
    "description": "full-update-all",
    "expired_time": <未来时间戳>,
    "enabled": true,
    "unlimited_quota": false,
    "models": ["gpt-4"],
    "subnet": ["10.0.0.0/8"],
    "quota_plan": {
        "unlimited": false,
        "quota": 100000,
        "unit": "token",
        "reset_period": "monthly"
    },
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "rpm": [{"name": "r1", "model": "*", "window_minutes": 1, "max_tokens": 100, "step_minutes": 1}]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.description**："full-update-all"  
**Data.quota_plan**：不为 null  
**Data.rate_limit_policy**：不为 null

---

### 必填校验

---

### AK-4-010：缺少 description（空 Body）

#### 设计思路

验证 PUT 时 description 为必填，空 Body 返回 422。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，Body 为空对象 {}
3. 验证 ErrNum=422

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-4-011：description 为空字符串

#### 设计思路

验证 PUT 时 description 不能为空字符串。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，description=""
3. 验证 ErrNum=422

#### 请求参数

```json
{
    "description": ""
}
```

#### 预期返回结果

**ErrNum**：422

---

### 边界值

---

### AK-4-012：description 最大合法长度（511）

#### 设计思路

验证 description 长度为 511 时通过校验。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 构造 511 字符的 description 字符串
3. 发送 PUT 请求
4. 验证 ErrNum=200

#### 请求参数

```json
{
    "description": "<511个'a'字符>"
}
```

#### 预期返回结果

**ErrNum**：200

---

### AK-4-013：expired_time=-1（永不过期）

#### 设计思路

验证 expired_time 设置为 -1 表示永不过期。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 expired_time=-1
3. 验证 ErrNum=200
4. 验证 Data.expired_time=-1

#### 请求参数

```json
{
    "description": "never-expire",
    "expired_time": -1
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.expired_time**：-1

---

### AK-4-014：更新超长 ID（256 字符）

#### 设计思路

验证传入超长 ID 时触发参数长度校验，返回 422。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 构造 256 字符的 ID 字符串
2. 发送 PUT 请求
3. 验证返回 422

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-4-015：更新空 models 数组

#### 设计思路

验证 models 设置为空数组时的行为（服务端可能默认填充为 ["*"]）。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 models=[]
3. 验证 ErrNum=200
4. 检查 models 返回（允许空数组或默认填充 ["*"]）

#### 请求参数

```json
{
    "description": "empty-models",
    "models": []
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.models**：空数组或 ["*"]（服务端默认填充）

---

### 异常参数

---

### AK-4-016：更新不存在的 API-Key

#### 设计思路

验证 PUT 不存在的 ID 时返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/api-keys/nonexistent-id`
2. 验证返回 404

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404

---

### AK-4-017：更新无效 UUID 格式的 ID

#### 设计思路

验证传入非 UUID 格式的 ID 时返回 404。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/api-keys/invalid-format`
2. 验证返回 404

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404

---

### AK-4-018：expired_time 为过去时间

#### 设计思路

验证 expired_time 设置为过去时间时返回 422。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 expired_time=<过去时间戳>
3. 验证 ErrNum=422

#### 请求参数

```json
{
    "description": "past-expire",
    "expired_time": 1000000000
}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-4-019：description 超长（512）

#### 设计思路

验证 description 长度为 512 时返回 422。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 构造 512 字符的 description
3. 发送 PUT 请求
4. 验证 ErrNum=422

#### 请求参数

```json
{
    "description": "<512个'a'字符>"
}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-4-020：subnet 为无效 CIDR

#### 设计思路

验证 subnet 中包含无效 CIDR 时返回 422。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，设置 subnet=["invalid-cidr"]
3. 验证 ErrNum=422

#### 请求参数

```json
{
    "description": "bad-subnet",
    "subnet": ["invalid-cidr"]
}
```

#### 预期返回结果

**ErrNum**：422

---

### 返回数据校验

---

### AK-4-021：全量更新返回结构校验

#### 设计思路

验证全量更新成功后，返回完整的 API-Key 对象，包含所有顶层字段。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，传入完整字段
3. 验证 ErrNum=200
4. 验证返回包含 id、description、enabled、expired_time、quota_plan、rate_limit_policy 等字段

#### 请求参数

```json
{
    "description": "structure-test",
    "enabled": true,
    "quota_plan": {
        "unlimited": false,
        "quota": 100000,
        "unit": "token",
        "reset_period": "daily"
    },
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [{"name": "t1", "model": "*", "window_minutes": 1, "max_tokens": 1000, "step_minutes": 1}]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200

| 字段 | 预期类型 | 校验方式 |
|------|---------|---------|
| id | string | 非空 |
| description | string | ="structure-test" |
| expired_time | number | 存在 |
| enabled | bool | =true |
| quota_plan | object | 非 null |
| rate_limit_policy | object | 非 null |

---

### AK-4-022：全量更新后 GET 验证

#### 设计思路

验证全量更新后，通过 GET 接口查询到的数据与 PUT 返回的数据一致。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，更新 description
3. 记录 PUT 返回的 description 值
4. 发送 GET 请求查询该 ID
5. 验证 GET 返回的 description 与 PUT 返回一致

#### 请求参数

```json
{
    "description": "put-get-verify"
}
```

#### 预期返回结果

**PUT**：ErrNum=200，description="put-get-verify"  
**GET**：ErrNum=200，description="put-get-verify"