# API-Key 模块测试

> 本目录存放 API-Key 模块的测试用例设计文档。  
> 对应的 Go 测试文件位于 `test/integration/tests/` 目录。

## 接口列表

| 接口 | 方法 | 路径 | 设计文档 | 用例数 |
|------|------|------|---------|--------|
| 创建API-Key | POST | `/open-api/v1/api-keys` | [create/design.md](./create/design.md) | 43 |
| 查询API-Key列表 | GET | `/open-api/v1/api-keys` | [list/design.md](./list/design.md) | 7 |
| 查询单个API-Key | GET | `/open-api/v1/api-keys/{id}` | [detail/design.md](./detail/design.md) | 3 |
| 全量更新API-Key | PUT | `/open-api/v1/api-keys/{id}` | [full_update/design.md](./full_update/design.md) | 5 |
| 部分更新API-Key | PATCH | `/open-api/v1/api-keys/{id}` | [partial_update/design.md](./partial_update/design.md) | 4 |
| 删除API-Key | DELETE | `/open-api/v1/api-keys/{id}` | [delete/design.md](./delete/design.md) | 3 |
| 查询配额计划 | GET | `/open-api/v1/api-keys/{id}/quota-plan` | [quota_query/design.md](./quota_query/design.md) | 3 |
| 重置配额余额 | POST | `/open-api/v1/api-keys/{id}/quota-plan/reset` | [quota_reset/design.md](./quota_reset/design.md) | 4 |
| **合计** | | | | **72** |

## 测试覆盖维度

- 正常参数：验证各接口最基本功能
- 必填校验：缺少必填字段时返回 422
- 边界值：空字符串、最大值、特殊值
- 异常参数：非法JSON、不存在资源
- 返回数据校验：逐字段验证返回结构完整性
- 业务规则：级联删除、配额重置等

## 目录结构

```
          ← 设计文档（本目录）
├── README.md                    ← 本文件
├── create/design.md             ← 创建API-Key 设计文档
├── list/design.md               ← 查询列表 设计文档
├── detail/design.md             ← 查询单个 设计文档
├── full_update/design.md        ← 全量更新 设计文档
├── partial_update/design.md     ← 部分更新 设计文档
├── delete/design.md             ← 删除 设计文档
├── quota_query/design.md        ← 查询配额 设计文档
└── quota_reset/design.md        ← 重置配额 设计文档

              ← Go 测试文件
├── create/create_test.go
├── list/list_test.go
├── detail/detail_test.go
├── full_update/full_update_test.go
├── partial_update/partial_update_test.go
├── delete/delete_test.go
├── quota_query/quota_query_test.go
└── quota_reset/quota_reset_test.go
```

## 运行命令

```bash
# 编译
cd ai-gateway-api
$env:CGO_ENABLED="0"; go build -o ai-gateway-api.exe ./main.go

# 运行测试
cd test/integration
go test -v -count=1 -timeout 120s ./create/
go test -v -count=1 -timeout 120s ./list/
go test -v -count=1 -timeout 120s ./detail/
go test -v -count=1 -timeout 120s ./full_update/
go test -v -count=1 -timeout 120s ./partial_update/
go test -v -count=1 -timeout 120s ./delete/
go test -v -count=1 -timeout 120s ./quota_query/
go test -v -count=1 -timeout 120s ./quota_reset/
```

---

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

---

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

---

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

---

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

---

# 查询API-Key列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 查询API-Key列表 |
| 方法 | GET |
| 路径 | /open-api/v1/api-keys |
| 说明 | 分页查询 API-Key 列表，支持 enabled、entity_id、unlimited_quota 过滤 |

---

## 二、接口参数说明

### Query 参数

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| page | int | 否 | 页码（必须 > 0） | 不传则不分页 |
| page_size | int | 否 | 每页条数（1-100） | 20（最大100） |
| enabled | bool | 否 | 是否启用过滤 | - |
| entity_id | string | 否 | 按挂载的 Entity ID 过滤（最大64字符） | - |
| unlimited_quota | bool | 否 | 是否无限配额过滤 | - |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []object | API-Key 列表（quota_plan 含 balance） |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

**list[0] 字段详情**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 唯一标识 |
| key | string | API-Key 值（脱敏或完整） |
| description | string | 描述 |
| enabled | bool | 是否启用 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |
| expired_time | int64 | 过期时间 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 允许访问的模型白名单 |
| subnet | []string | 允许的客户端子网 |
| quota_plan | object | 配额计划（含 balance） |
| rate_limit_policy | object | 限流策略 |
| entity_id | string | 挂载的 Entity ID |
| entity | object | 挂载的 Entity 摘要（可选） |
| remaining_quota | int64 | 剩余配额 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-2-001 | 默认分页查询（无参数） | 正常参数 | 验证默认分页参数生效 |
| AK-2-002 | 指定分页参数 | 正常参数 | page=1, page_size=5 |
| AK-2-003 | 按 enabled 过滤 | 正常参数 | 只查询启用的 API-Key |
| AK-2-004 | 按 entity_id 过滤 | 正常参数 | 验证过滤结果正确 |
| AK-2-005 | 按 unlimited_quota 过滤 | 正常参数 | 验证过滤结果正确 |
| AK-2-006 | page_size=100（最大值） | 边界值 | 验证最大分页支持 |
| AK-2-007 | page_size=101（超最大值） | 边界值 | 验证超出限制时的行为 |
| AK-2-008 | page=0（非法值） | 边界值 | 验证 ErrNum=422 |
| AK-2-009 | page=-1（负数） | 边界值 | 验证 ErrNum=422 |
| AK-2-010 | page_size=0（非法值） | 边界值 | 验证 ErrNum=422 |
| AK-2-011 | page_size=-1（负数） | 边界值 | 验证 ErrNum=422 |
| AK-2-012 | entity_id 超长（>64字符） | 异常参数 | 验证 ErrNum=422 |
| AK-2-013 | 查询空列表 | 正常参数 | 新建环境无数据时查询 |
| AK-2-014 | 验证分页返回结构 | 返回数据 | 验证 list 和 pagination 结构 |
| AK-2-015 | 验证 list 元素字段完整性 | 返回数据 | 验证所有字段存在且类型正确 |

---

## 四、测试场景详细设计

---

### AK-2-001：默认分页查询（无参数）

#### 设计思路

验证不带任何参数时，接口返回所有数据（不分页模式），此时 pagination 的 page 和 page_size 为 0。

#### 前提数据准备

- 先创建 3 个 API-Key

#### 执行步骤

1. 创建 3 个 API-Key
2. 发送 GET 请求（无参数）
3. 验证返回 list 长度和分页信息

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 ≥ 3  
**Data.pagination.page**：0（不分页模式）  
**Data.pagination.page_size**：0（不分页模式）  
**Data.pagination.total**：≥ 3

---

### AK-2-002：指定分页参数

#### 设计思路

验证指定 page 和 page_size 时，分页参数生效。

#### 前提数据准备

- 先创建 10 个 API-Key

#### 执行步骤

1. 创建 10 个 API-Key
2. 发送 GET 请求：`?page=1&page_size=5`
3. 验证返回 5 条记录

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 = 5  
**Data.pagination.page**：1  
**Data.pagination.page_size**：5  
**Data.pagination.total**：≥ 10

---

### AK-2-003：按 enabled 过滤

#### 设计思路

验证按 enabled 状态过滤能正确筛选。

#### 前提数据准备

- 创建 1 个 enabled=true 的 API-Key
- 创建 1 个 enabled=false 的 API-Key

#### 执行步骤

1. 创建 2 个不同状态的 API-Key
2. 发送 GET 请求：`?enabled=true`
3. 验证返回的 list 中所有记录 enabled=true

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 ≥ 1，所有元素 enabled=true

---

### AK-2-004：按 entity_id 过滤

#### 设计思路

验证按 entity_id 过滤能正确筛选挂载到指定 Entity 的 API-Key。

#### 前提数据准备

- 创建 1 个 Entity-Type（type_name=test_type, level=1）
- 创建 1 个 Entity（type=test_type）
- 创建 1 个挂载到该 Entity 的 API-Key
- 创建 1 个不挂载任何 Entity 的 API-Key

#### 执行步骤

1. 创建 Entity-Type
2. 创建 Entity，获取 entity_id
3. 创建 API-Key-A（挂载到该 Entity）
4. 创建 API-Key-B（不挂载 Entity）
5. 发送 GET 请求：`?entity_id={entity_id}`
6. 验证返回结果只包含 API-Key-A

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 = 1  
**Data.list[0].entity_id**：等于步骤2获取的 entity_id

---

### AK-2-005：按 unlimited_quota 过滤

#### 设计思路

验证按 unlimited_quota 过滤能正确筛选无限配额的 API-Key。

#### 前提数据准备

- 创建 1 个 unlimited_quota=true 的 API-Key
- 创建 1 个 unlimited_quota=false 的 API-Key

#### 执行步骤

1. 创建 API-Key-A（unlimited_quota=true）
2. 创建 API-Key-B（unlimited_quota=false）
3. 发送 GET 请求：`?unlimited_quota=true`
4. 验证返回结果只包含 API-Key-A

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 = 1  
**Data.list[0].unlimited_quota**：true

---

### AK-2-006：page_size=100（最大值）

#### 设计思路

验证 page_size 取最大值 100 时接口正常。

#### 执行步骤

1. 发送 GET 请求：`?page_size=100`
2. 验证返回成功

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.pagination.page_size**：100

---

### AK-2-007：page_size=101（超最大值）

#### 设计思路

验证 page_size 超过最大值 100 时的行为（应返回参数错误）。

#### 执行步骤

1. 发送 GET 请求：`?page_size=101`
2. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "page_size must be between 1 and 100"

---

### AK-2-008：page=0（非法值）

#### 设计思路

验证 page=0 时返回参数错误。

#### 执行步骤

1. 发送 GET 请求：`?page=0`
2. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "page must be > 0"

---

### AK-2-009：page=-1（负数）

#### 设计思路

验证 page=-1 时返回参数错误。

#### 执行步骤

1. 发送 GET 请求：`?page=-1`
2. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "page must be > 0"

---

### AK-2-010：page_size=0（非法值）

#### 设计思路

验证 page_size=0 时返回参数错误。

#### 执行步骤

1. 发送 GET 请求：`?page_size=0`
2. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "page_size must be between 1 and 100"

---

### AK-2-011：page_size=-1（负数）

#### 设计思路

验证 page_size=-1 时返回参数错误。

#### 执行步骤

1. 发送 GET 请求：`?page_size=-1`
2. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "page_size must be between 1 and 100"

---

### AK-2-012：entity_id 超长（>64字符）

#### 设计思路

验证 entity_id 超过最大长度 64 字符时返回参数错误。

#### 执行步骤

1. 构造超过 64 字符的 entity_id（如 100 个 'a'）
2. 发送 GET 请求：`?entity_id={超长字符串}`
3. 验证接口返回参数错误

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "entity_id must be <= 64 characters"

---

### AK-2-013：查询空列表

#### 设计思路

验证新环境中没有 API-Key 时，接口返回空列表。

#### 前提数据准备

- 无需预先创建数据（新测试环境）

#### 执行步骤

1. 直接发送 GET 请求
2. 验证返回空列表

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.list**：长度 = 0  
**Data.pagination.total**：0

---

### AK-2-014：验证分页返回结构

#### 设计思路

验证返回结构包含 list 数组和 pagination 对象，且每个元素包含必要字段。

#### 前提数据准备

- 先创建 1 个 API-Key

#### 执行步骤

1. 创建 1 个 API-Key
2. 发送 GET 请求
3. 逐字段校验返回结构

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 顶层校验**：

| 键名 | 预期类型 |
|------|---------|
| list | array |
| pagination | object |
| pagination.page | number |
| pagination.page_size | number |
| pagination.total | number |

---

### AK-2-015：验证 list 元素字段完整性

#### 设计思路

验证 list 中每个元素的必填字段都存在且类型正确，entity_id 为可选字段（未挂载 Entity 时可能不存在）。

#### 前提数据准备

- 先创建 1 个 API-Key

#### 执行步骤

1. 创建 1 个 API-Key
2. 发送 GET 请求
3. 逐字段校验 list[0] 的字段

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data.list[0] 必填字段校验**：

| 键名 | 预期类型 |
|------|---------|
| id | string（非空） |
| key | string（非空） |
| description | string |
| enabled | bool |
| create_time | number（>0） |
| update_time | number（>0） |
| expired_time | number |
| unlimited_quota | bool |
| models | array |
| subnet | array |
| quota_plan | object（非空） |
| rate_limit_policy | object（非空） |
| remaining_quota | number |

**Data.list[0] 可选字段**：

| 键名 | 预期类型 | 说明 |
|------|---------|------|
| entity_id | string | 挂载 Entity 时存在，否则可能不存在 |
| entity | object | 挂载 Entity 时存在，否则可能不存在 |

---

# 部分更新API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 部分更新API-Key |
| 方法 | PATCH |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 部分更新指定 API-Key 的字段，仅传需修改的字段 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | 否 | 描述，最大511字符 |
| enabled | bool | 否 | 是否启用 |
| key | string | 否 | API-Key 值，必须以 productName- 开头 |
| unlimited_quota | bool | 否 | 是否无限配额 |
| expired_time | int64 | 否 | 过期时间（-1=永不过期） |
| models | []string | 否 | 允许访问的模型白名单 |
| subnet | []string | 否 | 允许的客户端子网（CIDR格式） |
| entity_id | string | 否 | 挂载的 Entity ID |
| quota_plan | object | 否 | 配额计划 |
| rate_limit_policy | object | 否 | 限流策略 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 唯一标识 |
| key | string | API-Key 值 |
| description | string | 描述 |
| enabled | bool | 是否启用 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |
| expired_time | int64 | 过期时间 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 允许访问的模型白名单 |
| subnet | []string | 允许的客户端子网 |
| quota_plan | object | 配额计划（含 balance） |
| rate_limit_policy | object | 限流策略 |
| entity_id | string | 挂载的 Entity ID（可选） |
| remaining_quota | int64 | 剩余配额 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-5-001 | 仅修改 description | 正常参数 | 验证部分更新 description |
| AK-5-002 | 禁用 API-Key | 正常参数 | 验证 enabled=false |
| AK-5-003 | 启用 API-Key | 正常参数 | 验证 enabled=true |
| AK-5-004 | 更新不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-5-005 | 更新 expired_time（永不过期） | 正常参数 | expired_time=-1 |
| AK-5-006 | 更新 expired_time（未来时间） | 正常参数 | 设置未来过期时间 |
| AK-5-007 | 更新 quota_plan | 正常参数 | 更新配额计划 |
| AK-5-008 | 更新 rate_limit_policy | 正常参数 | 更新限流策略 |
| AK-5-009 | 更新 models | 正常参数 | 更新允许的模型列表 |
| AK-5-010 | 更新 subnet | 正常参数 | 更新允许的子网 |
| AK-5-011 | 更新 entity_id | 正常参数 | 挂载到指定 Entity |
| AK-5-012 | description 超长（512字符） | 边界值 | 验证 ErrNum=422 |
| AK-5-013 | expired_time=-2（非法值） | 边界值 | 验证 ErrNum=422 |
| AK-5-014 | expired_time=过去时间 | 边界值 | 验证 ErrNum=422 |
| AK-5-015 | subnet 格式错误 | 异常参数 | 验证 ErrNum=422 |
| AK-5-016 | rate_limit_policy.enabled=true 但无规则 | 异常参数 | 验证 ErrNum=422 |
| AK-5-017 | quota_plan.quota<0 | 异常参数 | 验证 ErrNum=422 |
| AK-5-018 | entity_id 指向不存在的 Entity | 异常参数 | 验证 ErrNum=422 |
| AK-5-019 | id 超长（>256字符） | 异常参数 | 验证 ErrNum=422 |
| AK-5-020 | 验证返回结构完整性 | 返回数据校验 | 验证所有字段存在 |

---

## 四、测试场景详细设计

---

### AK-5-001：仅修改 description

#### 设计思路

验证 PATCH 仅传 description 时，其他字段保持不变。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PATCH 请求，仅修改 description
3. 验证 description 更新，其他字段不变

#### 请求参数

```json
{
    "description": "patched-description"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.description**："patched-description"  
**Data.id**：不变  
**Data.enabled**：不变

---

### AK-5-002：禁用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=false 禁用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=true 的 API-Key

#### 请求参数

```json
{
    "enabled": false
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.enabled**：false

---

### AK-5-003：启用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=true 启用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=false 的 API-Key

#### 请求参数

```json
{
    "enabled": true
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.enabled**：true

---

### AK-5-004：更新不存在的 API-Key

#### 设计思路

验证 PATCH 不存在的 ID 时返回 404。

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 "API-Key" 和 "not exist"

---

### AK-5-005：更新 expired_time（永不过期）

#### 设计思路

验证 PATCH 设置 expired_time=-1（永不过期）。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": -1
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.expired_time**：-1

---

### AK-5-006：更新 expired_time（未来时间）

#### 设计思路

验证 PATCH 设置未来过期时间。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key
2. 获取当前时间 + 3600 秒作为过期时间
3. 发送 PATCH 请求

#### 请求参数

```json
{
    "expired_time": <当前时间+3600秒>
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.expired_time**：等于请求中的值

---

### AK-5-007：更新 quota_plan

#### 设计思路

验证 PATCH 更新配额计划。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "quota_plan": {
        "quota": 10000,
        "unit": "token",
        "reset_period": "daily"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.quota_plan.quota**：10000

---

### AK-5-008：更新 rate_limit_policy

#### 设计思路

验证 PATCH 更新限流策略。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "rate_limit_policy": {
        "enabled": true,
        "rules": {
            "tpm": [
                {
                    "window_minutes": 1,
                    "step_minutes": 1,
                    "max_tokens": 1000
                }
            ]
        }
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.rate_limit_policy.enabled**：true

---

### AK-5-009：更新 models

#### 设计思路

验证 PATCH 更新允许访问的模型列表。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "models": ["gpt-3.5-turbo", "gpt-4"]
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.models**：包含 ["gpt-3.5-turbo", "gpt-4"]

---

### AK-5-010：更新 subnet

#### 设计思路

验证 PATCH 更新允许的客户端子网。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "subnet": ["192.168.1.0/24"]
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.subnet**：包含 ["192.168.1.0/24"]

---

### AK-5-011：更新 entity_id

#### 设计思路

验证 PATCH 将 API-Key 挂载到指定 Entity。

#### 前提数据准备

- 创建 Entity-Type（type_name=test_type_patch, level=1）
- 创建 Entity（type=test_type_patch）
- 创建 API-Key（不挂载 Entity）

#### 请求参数

```json
{
    "entity_id": "<entity_id>"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.entity_id**：等于请求中的 entity_id

---

### AK-5-012：description 超长（512字符）

#### 设计思路

验证 description 超过最大长度 511 字符时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "description": "<512个字符>"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description must be less than 512 characters"

---

### AK-5-013：expired_time=-2（非法值）

#### 设计思路

验证 expired_time 小于 -1 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": -2
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Invalid expired_time"

---

### AK-5-014：expired_time=过去时间

#### 设计思路

验证 expired_time 设置为过去时间时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "expired_time": <当前时间-3600秒>
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "expired_time must be >= current time"

---

### AK-5-015：subnet 格式错误

#### 设计思路

验证 subnet 格式不正确时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "subnet": ["invalid-subnet"]
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "invalid subnet format"

---

### AK-5-016：rate_limit_policy.enabled=true 但无规则

#### 设计思路

验证 rate_limit_policy.enabled=true 但未设置 rules 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "rate_limit_policy": {
        "enabled": true
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "when rate_limit_policy.enabled is true, rules must be set"

---

### AK-5-017：quota_plan.quota<0

#### 设计思路

验证 quota_plan.quota 为负数时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "quota_plan": {
        "quota": -100
    }
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "quota must be >= 0"

---

### AK-5-018：entity_id 指向不存在的 Entity

#### 设计思路

验证 entity_id 指向不存在的 Entity 时返回参数错误。

#### 前提数据准备

- 先创建一个 API-Key

#### 请求参数

```json
{
    "entity_id": "nonexistent-entity-id"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Entity not found"

---

### AK-5-019：id 超长（>256字符）

#### 设计思路

验证 URI 参数 id 超长时返回参数错误。

#### 请求参数

URI：`/open-api/v1/api-keys/<257个字符>`

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "id" 和 "invalid"

---

### AK-5-020：验证返回结构完整性

#### 设计思路

验证部分更新成功后返回的结构完整性。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key
2. 发送 PATCH 请求修改 description
3. 验证返回结构包含所有必填字段

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 必填字段校验**：

| 键名 | 预期类型 |
|------|---------|
| id | string（非空） |
| key | string（非空） |
| description | string |
| enabled | bool |
| create_time | number（>0） |
| update_time | number（>0） |
| expired_time | number |
| unlimited_quota | bool |
| models | array |
| subnet | array |
| quota_plan | object（非空） |
| rate_limit_policy | object（非空） |

---

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
| id | string | 是 | API-Key 唯一标识（最大255字符） |

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
| AK-7-003 | 验证 balance 字段结构 | 返回数据校验 | 验证 used/remaining 类型和关系 |
| AK-7-004 | 查询无限配额的 API-Key | 正常参数 | 验证 unlimited=true |
| AK-7-005 | 查询无配额计划的 API-Key | 正常参数 | 验证返回 nil |
| AK-7-006 | id 超长（>255字符） | 边界值 | 验证 ErrNum=422 |
| AK-7-007 | 验证返回字段完整性 | 返回数据校验 | 验证所有字段存在且类型正确 |

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
**ErrMsg**：success

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
**ErrMsg**：包含 "API-Key" 和 "not exist"

---

### AK-7-003：验证 balance 字段结构

#### 设计思路

验证 balance 对象包含 used 和 remaining 两个字段，且为数字类型，used + remaining = quota。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data.balance 校验**：

| 字段 | 预期类型 | 预期值 |
|------|---------|--------|
| balance.used | number | ≥ 0 |
| balance.remaining | number | ≥ 0 |
| used + remaining | — | = quota |

---

### AK-7-004：查询无限配额的 API-Key

#### 设计思路

验证查询 unlimited=true 的 API-Key 时，返回 unlimited=true。

#### 前提数据准备

- 先创建一个 unlimited_quota=true 的 API-Key

#### 执行步骤

1. 创建 API-Key（unlimited_quota=true, quota_plan.unlimited=true）
2. 发送 GET 请求
3. 验证 unlimited=true

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.unlimited**：true

---

### AK-7-005：查询无配额计划的 API-Key

#### 设计思路

验证查询未设置 quota_plan 的 API-Key 时，系统自动创建默认配额计划（unlimited=true, quota=0, unit="total_token", reset_period="never"）。

#### 前提数据准备

- 先创建一个不传 quota_plan 的 API-Key

#### 执行步骤

1. 创建 API-Key（仅传 description）
2. 发送 GET 请求
3. 验证返回默认配额计划

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 |
|------|--------|
| unlimited | true |
| quota | 0 |
| unit | "total_token" |
| reset_period | "never" |
| balance.used | 0 |
| balance.remaining | 0 |

---

### AK-7-006：id 超长（>255字符）

#### 设计思路

验证 URI 参数 id 超过 255 字符时返回参数错误。

#### 请求参数

URI：`/open-api/v1/api-keys/<256字符>/quota-plan`

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "id" 和 "invalid"

---

### AK-7-007：验证返回字段完整性

#### 设计思路

验证返回数据结构包含所有字段（unlimited, pass_when_no_enough_quota, quota, unit, reset_period, balance）且类型正确。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 必填字段校验**：

| 键名 | 预期类型 |
|------|---------|
| unlimited | bool |
| pass_when_no_enough_quota | bool |
| quota | number（>0） |
| unit | string（非空） |
| reset_period | string（非空） |
| balance | object（非空） |
| balance.used | number |
| balance.remaining | number |

---

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
| id | string | 是 | API-Key 唯一标识（最大255字符） |

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
| AK-8-005 | 重置无配额计划的 API-Key | 异常参数 | 验证返回 422 |
| AK-8-006 | 重置配额为 0 | 边界值 | 验证 quota=0 重置成功 |
| AK-8-007 | 重置配额为负数 | 边界值 | 验证 quota<0 行为 |
| AK-8-008 | id 超长（>255字符） | 边界值 | 验证 ErrNum=422 |

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
**ErrMsg**：success

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
**ErrMsg**：success  
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
**ErrMsg**：包含 "API-Key" 和 "not exist"

---

### AK-8-004：验证返回结构完整性

#### 设计思路

验证重置后返回的 Data 包含所有必要字段且类型正确。

#### 前提数据准备

- 先创建一个带配额计划的 API-Key

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

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

---

### AK-8-005：重置无配额计划的 API-Key

#### 设计思路

验证重置未设置配额计划的 API-Key 时，由于系统自动创建了默认配额计划（unlimited=true），配额管理器拒绝重置无限配额，返回 500。

#### 前提数据准备

- 先创建一个不传 quota_plan 的 API-Key

#### 执行步骤

1. 创建 API-Key（仅传 description）
2. 发送 POST 请求
3. 验证返回 500

#### 请求参数

```json
{
    "quota": 50000000
}
```

#### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含 "cannot reset balance for unlimited quota"

---

### AK-8-006：重置配额为 0

#### 设计思路

验证传入 quota=0 时重置成功。

#### 前提数据准备

- 先创建一个带配额计划（quota=100000000）的 API-Key

#### 请求参数

```json
{
    "quota": 0
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data.new_quota**：0  
**Data.balance.new_remaining**：0

---

### AK-8-007：重置配额为负数

#### 设计思路

验证传入负数 quota 时接口行为（可能被接受或返回参数错误）。

#### 前提数据准备

- 先创建一个带配额计划（quota=100000000）的 API-Key

#### 请求参数

```json
{
    "quota": -100
}
```

#### 预期返回结果

**ErrNum**：200（可能接受）或 422

---

### AK-8-008：id 超长（>255字符）

#### 设计思路

验证 URI 参数 id 超过 255 字符时返回参数错误。

#### 请求参数

URI：`/open-api/v1/api-keys/<256字符>/quota-plan/reset`

Body：
```json
{
    "quota": 50000000
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "id" 和 "invalid"

---

