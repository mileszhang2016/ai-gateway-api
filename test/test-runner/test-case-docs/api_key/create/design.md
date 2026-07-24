# 创建API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 创建API-Key |
| 方法 | POST |
| 路径 | /open-api/v1/api-keys |
| 说明 | 创建一个新的 API-Key，支持配置配额计划、限流策略、模型白名单、子网限制等 |

---

## 二、接口参数说明

### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| description | string | **是** | API-Key 描述，最小长度1 | - |
| expired_time | int64 | 否 | 过期时间（Unix时间戳秒），-1永不过期 | -1 |
| enabled | bool | 否 | 是否启用 | true |
| unlimited_quota | bool | 否 | 是否无限配额 | false |
| models | []string | 否 | 允许访问的模型白名单 | ["*"] |
| subnet | []string | 否 | 允许的客户端子网（CIDR格式） | ["*"] |
| quota_plan | object | 否 | 配额计划（不含balance） | 使用默认值 |
| rate_limit_policy | object | 否 | 限流策略 | enabled=false |
| entity_id | string | 否 | 挂载的 Entity ID | 空 |

**quota_plan 子字段**：

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| unlimited | bool | 否 | 配额计划是否无限 | false |
| pass_when_no_enough_quota | bool | 否 | 配额不足时是否放通 | false |
| quota | int64 | 否 | 配额总量（>=0） | 0 |
| unit | string | 否 | 配额单位（如 "token"、"total_token"） | "" |
| reset_period | string | 否 | 重置周期（如 "daily"、"monthly"） | "" |

**rate_limit_policy 子字段**：

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| enabled | bool | 否 | 是否启用限流 | false |
| rules | object | 否 | 限流规则（含tpm/rpm/max_concurrency） | - |

**rules 子字段**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| tpm | []object | 否 | TPM规则列表，每项含 name/model/window_minutes/max_tokens/step_minutes |
| rpm | []object | 否 | RPM规则列表，每项含 name/model/window_minutes/max_requests |
| max_concurrency | int | 否 | 最大并发数 |

**约束**：
- 若传入 `rate_limit_policy` 且 `enabled` 为 `true`，则 `rules` 中 `tpm`、`rpm`、`max_concurrency` 三者至少配置其一
- 若 `entity_id` 不为空，该 Entity 必须存在
- `expired_time` 必须 >= 当前时间（或为 -1 表示永不过期）
- `quota_plan.quota` >= 0
- `tpm.window_minutes` 范围 [1, 360]，且 `step_minutes <= window_minutes`
- `rpm.window_minutes` 范围 [1, 360]
- `subnet` 中每个元素需为合法CIDR格式

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 唯一标识 |
| key | string | API-Key 值（用于请求头鉴权） |
| description | string | 描述 |
| enabled | bool | 是否启用 |
| create_time | int64 | 创建时间（Unix时间戳秒） |
| update_time | int64 | 更新时间 |
| expired_time | int64 | 过期时间 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 允许的模型白名单 |
| subnet | []string | 允许的客户端子网 |
| quota_plan | object | 配额计划（不含balance） |
| rate_limit_policy | object | 限流策略 |
| entity_id | string | 挂载的 Entity ID |
| entity | object/null | 挂载的 Entity 摘要 |

---

## 三、测试场景总览

### 正常参数（18）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-1-001 | 最小参数创建（仅description） | 正常参数 | description | 验证所有默认值正确 |
| AK-1-002 | 设置expired_time为未来时间戳 | 正常参数 | expired_time | 验证expired_time参数 |
| AK-1-003 | 设置enabled=false | 正常参数 | enabled | 验证enabled参数 |
| AK-1-004 | 设置unlimited_quota=true | 正常参数 | unlimited_quota | 验证unlimited_quota参数 |
| AK-1-005 | 设置models指定白名单 | 正常参数 | models | 验证models参数 |
| AK-1-006 | 设置subnet指定子网 | 正常参数 | subnet | 验证subnet参数 |
| AK-1-007 | 设置quota_plan全部字段 | 正常参数 | quota_plan | 验证quota_plan参数 |
| AK-1-008 | quota_plan.unlimited=true | 正常参数 | quota_plan.unlimited | 验证quota_plan.unlimited子参数 |
| AK-1-009 | quota_plan.pass_when_no_enough_quota=true | 正常参数 | quota_plan.pass_when_no_enough_quota | 验证pass_when_no_enough_quota子参数 |
| AK-1-010 | quota_plan.unit="token" | 正常参数 | quota_plan.unit | 验证quota_plan.unit子参数 |
| AK-1-011 | quota_plan.reset_period="daily" | 正常参数 | quota_plan.reset_period | 验证quota_plan.reset_period子参数 |
| AK-1-012 | quota_plan.quota=50000000 | 正常参数 | quota_plan.quota | 验证quota_plan.quota子参数 |
| AK-1-013 | rate_limit_policy完整配置（tpm+rpm+max_concurrency=50） | 正常参数 | rate_limit_policy | 验证rate_limit_policy参数 |
| AK-1-014 | rate_limit_policy仅配置tpm | 正常参数 | rules.tpm | 验证rules.tpm子参数 |
| AK-1-015 | rate_limit_policy仅配置rpm | 正常参数 | rules.rpm | 验证rules.rpm子参数 |
| AK-1-016 | rate_limit_policy仅配置max_concurrency=50 | 正常参数 | rules.max_concurrency | 验证rules.max_concurrency子参数 |
| AK-1-017 | rate_limit_policy.max_concurrency=0 | 正常参数 | rules.max_concurrency | 验证max_concurrency=0边界行为 |
| AK-1-018 | 挂载到有效Entity（需先创建Entity和EntityType） | 正常参数 | entity_id | 验证entity_id参数 |

### 必填校验（2）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-1-019 | 缺少description字段 | 必填校验 | description | 验证description必填，ErrNum=422 |
| AK-1-020 | description为空字符串"" | 必填校验 | description | 验证description非空，ErrNum=422 |

### 边界值（8）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-1-021 | expired_time=-1永久有效 | 边界值 | expired_time | 验证expired_time=-1，ErrNum=200 |
| AK-1-022 | description长度511字符（最大合法值） | 边界值 | description | 验证description长度边界，ErrNum=200 |
| AK-1-023 | models=[]空数组 | 边界值 | models | 验证models空数组，ErrNum=200 |
| AK-1-024 | subnet=[]空数组 | 边界值 | subnet | 验证subnet空数组，ErrNum=200 |
| AK-1-025 | quota_plan.quota=0 | 边界值 | quota_plan.quota | 验证quota=0边界，ErrNum=200 |
| AK-1-026 | tpm.window_minutes=1（最小值） | 边界值 | rules.tpm.window_minutes | 验证tpm窗口最小值，ErrNum=200 |
| AK-1-027 | tpm.window_minutes=360（最大值） | 边界值 | rules.tpm.window_minutes | 验证tpm窗口最大值，ErrNum=200 |
| AK-1-028 | rpm.window_minutes=1（最小值） | 边界值 | rules.rpm.window_minutes | 验证rpm窗口最小值，ErrNum=200 |

### 异常参数（12）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-1-029 | 非法JSON Body（text/plain） | 异常参数 | Body | 验证JSON解析失败，ErrNum=422 |
| AK-1-030 | rate_limit_policy.enabled=true但rules为空对象 | 异常参数 | rate_limit_policy.rules | 验证规则必填，ErrNum=422 |
| AK-1-031 | expired_time为过去时间戳（如1700000000） | 异常参数 | expired_time | 验证过期时间校验，ErrNum=422 |
| AK-1-032 | expired_time=-2（小于-1） | 异常参数 | expired_time | 验证expired_time下限，ErrNum=422 |
| AK-1-033 | entity_id指向不存在的Entity | 异常参数 | entity_id | 验证Entity存在性，ErrNum=404或422 |
| AK-1-034 | subnet格式无效（如"invalid"） | 异常参数 | subnet | 验证CIDR格式校验，ErrNum=422 |
| AK-1-035 | quota_plan.quota=-1（负数） | 异常参数 | quota_plan.quota | 验证quota>=0，ErrNum=422 |
| AK-1-036 | description长度512字符（超出上限） | 异常参数 | description | 验证description长度上限，ErrNum=422 |
| AK-1-037 | tpm.window_minutes=0（无效） | 异常参数 | rules.tpm.window_minutes | 验证tpm窗口下限，ErrNum=422 |
| AK-1-038 | tpm.window_minutes=361（无效） | 异常参数 | rules.tpm.window_minutes | 验证tpm窗口上限，ErrNum=422 |
| AK-1-039 | tpm.step_minutes=10 > window_minutes=5 | 异常参数 | rules.tpm.step_minutes | 验证step≤window约束，ErrNum=422 |
| AK-1-040 | rpm.window_minutes=0（无效） | 异常参数 | rules.rpm.window_minutes | 验证rpm窗口下限，ErrNum=422 |

### 返回数据校验（3）

| 编号 | 场景 | 测试类型 | 覆盖参数 | 简要说明 |
|------|------|---------|---------|---------|
| AK-1-041 | 顶层字段完整性校验 | 返回数据校验 | 全部顶层字段 | 验证所有13个顶层字段存在且类型正确，ErrNum=200 |
| AK-1-042 | quota_plan返回结构校验 | 返回数据校验 | quota_plan子字段 | 验证quota_plan子字段完整性，ErrNum=200 |
| AK-1-043 | rate_limit_policy返回结构校验 | 返回数据校验 | rate_limit_policy子字段 | 验证rate_limit_policy子字段完整性，ErrNum=200 |

---

## 四、测试场景详细设计

---

### 正常参数

---

### AK-1-001：最小参数创建（仅description）

#### 设计思路

验证 API-Key 创建接口的最基本功能：仅传入必填字段 description，所有可选字段使用默认值。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：`{"description": "test-key-001"}`
2. 发送 POST 请求到 `/open-api/v1/api-keys`
3. 验证 ErrNum=200
4. 验证返回字段值

#### 请求参数

```json
{
    "description": "test-key-001"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| key | 非空字符串 | NotEmpty |
| description | "test-key-001" | Equals |
| enabled | true | Equals（默认值） |
| create_time | > 0 | GreaterThan(0) |
| update_time | = create_time | Equals |
| expired_time | -1 | Equals（默认值） |
| unlimited_quota | false | Equals（默认值） |
| models | ["*"] | Equals（默认值） |
| subnet | ["*"] | Equals（默认值） |
| quota_plan | 非 null | NotNull |
| rate_limit_policy | 非 null | NotNull |
| entity_id | "" | Equals（默认值） |
| entity | null | Equals |

---

### AK-1-002：设置expired_time为未来时间戳

#### 设计思路

验证传入具体过期时间戳（1年后），接口正常返回该值。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 计算 `futureTime = time.Now().Unix() + 86400*365`
2. 构造请求 Body，包含 `expired_time` 为 futureTime
3. 发送 POST 请求到 `/open-api/v1/api-keys`
4. 验证 ErrNum=200
5. 验证 `Data.expired_time` 等于传入值

#### 请求参数

```json
{
    "description": "test-key-002",
    "expired_time": 1793260800
}
```

> 注：expired_time 使用 `time.Now().Unix() + 86400*365` 动态计算

#### 预期返回结果

**ErrNum**：200  
**Data.expired_time**：与传入的 expired_time 一致

---

### AK-1-003：设置enabled=false

#### 设计思路

验证创建时设置 enabled=false，返回正确值。

#### 请求参数

```json
{
    "description": "test-key-003",
    "enabled": false
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.enabled**：false

---

### AK-1-004：设置unlimited_quota=true

#### 设计思路

验证无限配额模式，返回 unlimited_quota=true。

#### 请求参数

```json
{
    "description": "test-key-004",
    "unlimited_quota": true
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.unlimited_quota**：true

---

### AK-1-005：设置models指定白名单

#### 设计思路

验证传入自定义模型白名单，接口正常返回。

#### 请求参数

```json
{
    "description": "test-key-005",
    "models": ["gpt-4", "gpt-3.5-turbo"]
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.models**：["gpt-4", "gpt-3.5-turbo"]

---

### AK-1-006：设置subnet指定子网

#### 设计思路

验证传入自定义子网限制（CIDR格式），接口正常返回。

#### 请求参数

```json
{
    "description": "test-key-006",
    "subnet": ["10.0.0.0/8", "192.168.1.0/24"]
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.subnet**：["10.0.0.0/8", "192.168.1.0/24"]

---

### AK-1-007：设置quota_plan全部字段

#### 设计思路

验证创建 API-Key 时传入完整的自定义配额计划，确认所有配额计划字段正确持久化并返回。

#### 请求参数

```json
{
    "description": "test-key-007",
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```

#### 预期返回结果

**ErrNum**：200  

**Data.quota_plan 字段校验**：

| 字段 | 预期值 |
|------|--------|
| quota_plan.unlimited | false |
| quota_plan.pass_when_no_enough_quota | false |
| quota_plan.quota | 100000000 |
| quota_plan.unit | "total_token" |
| quota_plan.reset_period | "monthly" |

---

### AK-1-008：quota_plan.unlimited=true

#### 设计思路

验证 quota_plan 中设置 unlimited=true（配额计划无限），单独验证该子参数。

#### 请求参数

```json
{
    "description": "test-key-008",
    "quota_plan": {
        "unlimited": true
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.unlimited**：true

---

### AK-1-009：quota_plan.pass_when_no_enough_quota=true

#### 设计思路

验证 quota_plan 中设置 pass_when_no_enough_quota=true（配额不足时放通），单独验证该子参数。

#### 请求参数

```json
{
    "description": "test-key-009",
    "quota_plan": {
        "pass_when_no_enough_quota": true
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.pass_when_no_enough_quota**：true

---

### AK-1-010：quota_plan.unit="token"

#### 设计思路

验证 quota_plan 中设置 unit="token"，单独验证该子参数。

#### 请求参数

```json
{
    "description": "test-key-010",
    "quota_plan": {
        "unit": "token"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.unit**："token"

---

### AK-1-011：quota_plan.reset_period="daily"

#### 设计思路

验证 quota_plan 中设置 reset_period="daily"，单独验证该子参数。

#### 请求参数

```json
{
    "description": "test-key-011",
    "quota_plan": {
        "reset_period": "daily"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.reset_period**："daily"

---

### AK-1-012：quota_plan.quota=50000000

#### 设计思路

验证 quota_plan 中设置 quota=50000000，单独验证该子参数。

#### 请求参数

```json
{
    "description": "test-key-012",
    "quota_plan": {
        "quota": 50000000
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.quota**：50000000

---

### AK-1-013：rate_limit_policy完整配置（tpm+rpm+max_concurrency=50）

#### 设计思路

验证 rate_limit_policy.enabled=true 时，同时配置 tpm、rpm、max_concurrency 三条规则，确认所有规则正确持久化并返回。

#### 请求参数

```json
{
    "description": "test-key-013",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "1分钟窗口",
                    "model": "*",
                    "window_minutes": 1,
                    "max_tokens": 10000,
                    "step_minutes": 1
                }
            ],
            "rpm": [
                {
                    "name": "1分钟请求",
                    "model": "*",
                    "window_minutes": 1,
                    "max_requests": 100
                }
            ],
            "max_concurrency": 50
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  

**Data.rate_limit_policy 字段校验**：

| 字段 | 预期值 |
|------|--------|
| rate_limit_policy.enabled | true |
| rate_limit_policy.rules.tpm[0].name | "1分钟窗口" |
| rate_limit_policy.rules.tpm[0].max_tokens | 10000 |
| rate_limit_policy.rules.tpm[0].window_minutes | 1 |
| rate_limit_policy.rules.tpm[0].step_minutes | 1 |
| rate_limit_policy.rules.rpm[0].name | "1分钟请求" |
| rate_limit_policy.rules.rpm[0].max_requests | 100 |
| rate_limit_policy.rules.rpm[0].window_minutes | 1 |
| rate_limit_policy.rules.max_concurrency | 50 |

---

### AK-1-014：rate_limit_policy仅配置tpm

#### 设计思路

验证 rate_limit_policy 仅配置 tpm 规则（不配置 rpm 和 max_concurrency），单独验证 rules.tpm 子参数。

#### 请求参数

```json
{
    "description": "test-key-014",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "仅TPM限制",
                    "model": "*",
                    "window_minutes": 5,
                    "max_tokens": 50000,
                    "step_minutes": 1
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.enabled**：true  
**Data.rate_limit_policy.rules.tpm[0].name**："仅TPM限制"  
**Data.rate_limit_policy.rules.tpm[0].max_tokens**：50000

---

### AK-1-015：rate_limit_policy仅配置rpm

#### 设计思路

验证 rate_limit_policy 仅配置 rpm 规则（不配置 tpm 和 max_concurrency），单独验证 rules.rpm 子参数。

#### 请求参数

```json
{
    "description": "test-key-015",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "rpm": [
                {
                    "name": "仅RPM限制",
                    "model": "*",
                    "window_minutes": 1,
                    "max_requests": 200
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.enabled**：true  
**Data.rate_limit_policy.rules.rpm[0].name**："仅RPM限制"  
**Data.rate_limit_policy.rules.rpm[0].max_requests**：200

---

### AK-1-016：rate_limit_policy仅配置max_concurrency=50

#### 设计思路

验证 rate_limit_policy 仅配置 max_concurrency（不配置 tpm 和 rpm），单独验证 rules.max_concurrency 子参数。

#### 请求参数

```json
{
    "description": "test-key-016",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "max_concurrency": 50
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.enabled**：true  
**Data.rate_limit_policy.rules.max_concurrency**：50

---

### AK-1-017：rate_limit_policy.max_concurrency=0

#### 设计思路

验证 max_concurrency=0 的边界行为，确认接口是否接受该值。0 可能表示无并发限制或无效值。

#### 请求参数

```json
{
    "description": "test-key-017",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "max_concurrency": 0
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200（若0为合法值，表示无并发限制）或 422（若0不合法）  
**Data.rate_limit_policy.rules.max_concurrency**：0（若返回200）

---

### AK-1-018：挂载到有效Entity（需先创建Entity和EntityType）

#### 设计思路

验证创建 API-Key 时传入有效的 entity_id，确认 API-Key 能正确挂载到已有 Entity 上。

#### 前提数据准备

1. 调用创建 EntityType 接口，创建一个 EntityType
2. 调用创建 Entity 接口，基于上述 EntityType 创建一个 Entity
3. 获取 Entity 的 ID 作为 entity_id

#### 执行步骤

1. 创建 EntityType（如 `{"name": "test-entity-type-018"}`）
2. 创建 Entity（如 `{"entity_type_id": "<entity_type_id>", "name": "test-entity-018"}`）
3. 构造 API-Key 请求，传入 `entity_id` 为步骤2获取的 Entity ID
4. 发送 POST 请求到 `/open-api/v1/api-keys`
5. 验证 ErrNum=200
6. 验证 `Data.entity_id` 等于传入值
7. 验证 `Data.entity` 不为 null，包含 Entity 摘要信息

#### 请求参数

```json
{
    "description": "test-key-018",
    "entity_id": "<预先创建的Entity ID>"
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.entity_id**：与传入的 entity_id 一致  
**Data.entity**：非 null，包含 Entity 摘要信息

---

### 必填校验

---

### AK-1-019：缺少description字段

#### 设计思路

验证 description 为必填字段，缺失时返回 422。

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description" 的错误信息  
**Data**：null

---

### AK-1-020：description为空字符串""

#### 设计思路

验证 description 为空字符串时，`validate:"required,min=1"` 约束触发校验失败。

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

### AK-1-021：expired_time=-1永久有效

#### 设计思路

验证 -1 表示永不过期，接口正常返回 -1。

#### 请求参数

```json
{
    "description": "test-key-021",
    "expired_time": -1
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.expired_time**：-1

---

### AK-1-022：description长度511字符（最大合法值）

#### 设计思路

验证 description 长度刚好为 511 字符（最大合法边界），接口正常返回。

#### 请求参数

```json
{
    "description": "aaaaaaaaaa...（511个字符）"
}
```

> 注：description 为 511 个 'a' 字符组成的字符串

#### 预期返回结果

**ErrNum**：200  
**Data.description**：与传入的511字符一致

---

### AK-1-023：models=[]空数组

#### 设计思路

验证 models 传入空数组，确认接口行为。空数组可能表示不限制模型或不允许任何模型。

#### 请求参数

```json
{
    "description": "test-key-023",
    "models": []
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.models**：[]

---

### AK-1-024：subnet=[]空数组

#### 设计思路

验证 subnet 传入空数组，确认接口行为。

#### 请求参数

```json
{
    "description": "test-key-024",
    "subnet": []
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.subnet**：[]

---

### AK-1-025：quota_plan.quota=0

#### 设计思路

验证 quota_plan.quota=0 的边界行为，确认接口是否接受配额为0。

#### 请求参数

```json
{
    "description": "test-key-025",
    "quota_plan": {
        "quota": 0
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.quota_plan.quota**：0

---

### AK-1-026：tpm.window_minutes=1（最小值）

#### 设计思路

验证 tpm 规则中 window_minutes 设为最小值 1，接口正常接受。

#### 请求参数

```json
{
    "description": "test-key-026",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "最小窗口",
                    "model": "*",
                    "window_minutes": 1,
                    "max_tokens": 1000,
                    "step_minutes": 1
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.rules.tpm[0].window_minutes**：1

---

### AK-1-027：tpm.window_minutes=360（最大值）

#### 设计思路

验证 tpm 规则中 window_minutes 设为最大值 360，接口正常接受。

#### 请求参数

```json
{
    "description": "test-key-027",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "最大窗口",
                    "model": "*",
                    "window_minutes": 360,
                    "max_tokens": 1000000,
                    "step_minutes": 60
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.rules.tpm[0].window_minutes**：360

---

### AK-1-028：rpm.window_minutes=1（最小值）

#### 设计思路

验证 rpm 规则中 window_minutes 设为最小值 1，接口正常接受。

#### 请求参数

```json
{
    "description": "test-key-028",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "rpm": [
                {
                    "name": "最小RPM窗口",
                    "model": "*",
                    "window_minutes": 1,
                    "max_requests": 10
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.rate_limit_policy.rules.rpm[0].window_minutes**：1

---

### 异常参数

---

### AK-1-029：非法JSON Body（text/plain）

#### 设计思路

验证传入非 JSON 格式的 Body 时，JSON 解析失败返回 422。

#### 请求参数

```
this is not valid json
```

（Content-Type: text/plain）

#### 预期返回结果

**ErrNum**：422

---

### AK-1-030：rate_limit_policy.enabled=true但rules为空对象

#### 设计思路

验证限流策略启用但未配置任何规则（tpm/rpm/max_concurrency 均未设置）时，接口返回参数校验错误。

#### 请求参数

```json
{
    "description": "test-key-030",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {}
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "rate_limit_policy" 或 "rules" 或 "Param Illegal"

---

### AK-1-031：expired_time为过去时间戳（如1700000000）

#### 设计思路

验证传入已过期的时间戳（如 1700000000，对应 2023-11-15），接口校验过期时间必须大于当前时间。

#### 请求参数

```json
{
    "description": "test-key-031",
    "expired_time": 1700000000
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "expired" 或 "time" 或 "Param Illegal"

---

### AK-1-032：expired_time=-2（小于-1）

#### 设计思路

验证 expired_time 小于 -1 的非法值（如 -2），接口校验 expired_time 必须 >= -1。

#### 请求参数

```json
{
    "description": "test-key-032",
    "expired_time": -2
}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-1-033：entity_id指向不存在的Entity

#### 设计思路

验证传入一个不存在的 entity_id 时，接口返回资源不存在错误。

#### 请求参数

```json
{
    "description": "test-key-033",
    "entity_id": "non-existent-entity-id-000000"
}
```

#### 预期返回结果

**ErrNum**：404 或 422  
**ErrMsg**：包含 "entity" 或 "not found" 或 "Param Illegal"

---

### AK-1-034：subnet格式无效（如"invalid"）

#### 设计思路

验证 subnet 中传入非 CIDR 格式的字符串（如 "invalid"），接口校验 CIDR 格式。

#### 请求参数

```json
{
    "description": "test-key-034",
    "subnet": ["invalid"]
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "subnet" 或 "CIDR" 或 "Param Illegal"

---

### AK-1-035：quota_plan.quota=-1（负数）

#### 设计思路

验证 quota_plan.quota 传入负数，接口校验 quota 必须 >= 0。

#### 请求参数

```json
{
    "description": "test-key-035",
    "quota_plan": {
        "quota": -1
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "quota" 或 "Param Illegal"

---

### AK-1-036：description长度512字符（超出上限）

#### 设计思路

验证 description 长度为 512 字符（超出上限 511），接口校验长度限制。

#### 请求参数

```json
{
    "description": "aaaaaaaaaa...（512个字符）"
}
```

> 注：description 为 512 个 'a' 字符组成的字符串

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description" 或 "length" 或 "max" 或 "Param Illegal"

---

### AK-1-037：tpm.window_minutes=0（无效）

#### 设计思路

验证 tpm.window_minutes 设为 0（小于最小值 1），接口校验失败。

#### 请求参数

```json
{
    "description": "test-key-037",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "无效窗口",
                    "model": "*",
                    "window_minutes": 0,
                    "max_tokens": 1000,
                    "step_minutes": 1
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "window_minutes" 或 "Param Illegal"

---

### AK-1-038：tpm.window_minutes=361（无效）

#### 设计思路

验证 tpm.window_minutes 设为 361（大于最大值 360），接口校验失败。

#### 请求参数

```json
{
    "description": "test-key-038",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "超限窗口",
                    "model": "*",
                    "window_minutes": 361,
                    "max_tokens": 1000,
                    "step_minutes": 1
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "window_minutes" 或 "Param Illegal"

---

### AK-1-039：tpm.step_minutes=10 > window_minutes=5

#### 设计思路

验证 tpm 规则中 step_minutes 大于 window_minutes 时，违反 step <= window 约束，接口校验失败。

#### 请求参数

```json
{
    "description": "test-key-039",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "step大于window",
                    "model": "*",
                    "window_minutes": 5,
                    "max_tokens": 1000,
                    "step_minutes": 10
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "step" 或 "window" 或 "Param Illegal"

---

### AK-1-040：rpm.window_minutes=0（无效）

#### 设计思路

验证 rpm.window_minutes 设为 0（小于最小值 1），接口校验失败。

#### 请求参数

```json
{
    "description": "test-key-040",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "rpm": [
                {
                    "name": "无效RPM窗口",
                    "model": "*",
                    "window_minutes": 0,
                    "max_requests": 100
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "window_minutes" 或 "Param Illegal"

---

### 返回数据校验

---

### AK-1-041：顶层字段完整性校验

#### 设计思路

验证创建成功后返回的 Data 包含所有 13 个顶层字段，且类型正确。

#### 请求参数

```json
{
    "description": "test-key-041",
    "expired_time": -1,
    "unlimited_quota": false,
    "models": ["*"],
    "subnet": ["*"],
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```

#### 预期返回结果

**ErrNum**：200  

**Data 顶层键校验**：

| 键名 | 预期类型 | 说明 |
|------|---------|------|
| id | string | 非空 |
| key | string | 非空 |
| description | string | "test-key-041" |
| enabled | bool | true（默认） |
| create_time | number | > 0 |
| update_time | number | > 0 |
| expired_time | number | -1 |
| unlimited_quota | bool | false |
| models | array | ["*"] |
| subnet | array | ["*"] |
| quota_plan | object | 非 null |
| rate_limit_policy | object | 非 null |
| entity_id | string | "" |
| entity | null | null |

---

### AK-1-042：quota_plan返回结构校验

#### 设计思路

验证创建成功后返回的 quota_plan 子字段完整，包含 unlimited、pass_when_no_enough_quota、quota、unit、reset_period 五个字段。

#### 请求参数

```json
{
    "description": "test-key-042",
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": true,
        "quota": 50000000,
        "unit": "token",
        "reset_period": "daily"
    }
}
```

#### 预期返回结果

**ErrNum**：200  

**Data.quota_plan 字段完整性校验**：

| 键名 | 预期类型 | 预期值 |
|------|---------|--------|
| quota_plan.unlimited | bool | false |
| quota_plan.pass_when_no_enough_quota | bool | true |
| quota_plan.quota | number | 50000000 |
| quota_plan.unit | string | "token" |
| quota_plan.reset_period | string | "daily" |

---

### AK-1-043：rate_limit_policy返回结构校验

#### 设计思路

验证创建成功后返回的 rate_limit_policy 子字段完整，包含 enabled、rules.tpm、rules.rpm、rules.max_concurrency。

#### 请求参数

```json
{
    "description": "test-key-043",
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "name": "结构校验TPM",
                    "model": "*",
                    "window_minutes": 10,
                    "max_tokens": 20000,
                    "step_minutes": 5
                }
            ],
            "rpm": [
                {
                    "name": "结构校验RPM",
                    "model": "*",
                    "window_minutes": 10,
                    "max_requests": 500
                }
            ],
            "max_concurrency": 30
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  

**Data.rate_limit_policy 字段完整性校验**：

| 键名 | 预期类型 | 预期值 |
|------|---------|--------|
| rate_limit_policy.enabled | bool | true |
| rate_limit_policy.rules.tpm | array | 非空，长度1 |
| rate_limit_policy.rules.tpm[0].name | string | "结构校验TPM" |
| rate_limit_policy.rules.tpm[0].model | string | "*" |
| rate_limit_policy.rules.tpm[0].window_minutes | number | 10 |
| rate_limit_policy.rules.tpm[0].max_tokens | number | 20000 |
| rate_limit_policy.rules.tpm[0].step_minutes | number | 5 |
| rate_limit_policy.rules.rpm | array | 非空，长度1 |
| rate_limit_policy.rules.rpm[0].name | string | "结构校验RPM" |
| rate_limit_policy.rules.rpm[0].model | string | "*" |
| rate_limit_policy.rules.rpm[0].window_minutes | number | 10 |
| rate_limit_policy.rules.rpm[0].max_requests | number | 500 |
| rate_limit_policy.rules.max_concurrency | number | 30 |