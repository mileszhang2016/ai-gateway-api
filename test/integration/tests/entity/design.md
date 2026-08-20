# Entity 测试用例设计文档

## 1. 模块概述

Entity 模块用于管理组织架构实体（部门、团队、项目、个人等），支持层级关系、模型黑白名单、配额计划、限流策略、路由规则。v0.3.0 列表接口明确为分页结构 `{list, pagination}`，详情/创建/更新返回不含 `balance`。配额计划 `unit` 支持 `total_token` 与 `RMB`，金额型配额使用 `DECIMAL(18,8)` 存储。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| E-1 | 创建 Entity | POST | `/open-api/v1/entities` | - |
| E-2 | 查询 Entity 列表 | GET | `/open-api/v1/entities` | 分页 |
| E-3 | 查询单个 Entity | GET | `/open-api/v1/entities/{id}` | - |
| E-4 | 全量更新 Entity | PUT | `/open-api/v1/entities/{id}` | type 不可改 |
| E-5 | 部分更新 Entity | PATCH | `/open-api/v1/entities/{id}` | - |
| E-6 | 删除 Entity | DELETE | `/open-api/v1/entities/{id}` | 有子节点或被挂载时禁止删除 |
| E-7 | 查询配额计划 | GET | `/open-api/v1/entities/{id}/quota-plan` | 含 balance |
| E-8 | 重置配额余额 | POST | `/open-api/v1/entities/{id}/quota-plan/reset` | - |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建 Entity | 11 |
| 查询 Entity 列表 | 3 |
| 查询单个 Entity | 2 |
| 全量更新 Entity | 6 |
| 部分更新 Entity | 4 |
| 删除 Entity | 3 |
| 查询配额计划 | 2 |
| 重置配额余额 | 3 |
| **合计** | **32** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
entity/
├── design.md
├── create/
│   └── create_test.go
├── list/
│   └── list_test.go
├── detail/
│   └── detail_test.go
├── full_update/
│   └── full_update_test.go
├── partial_update/
│   └── partial_update_test.go
├── delete/
│   └── delete_test.go
├── quota_plan/
│   └── quota_plan_test.go
└── quota_reset/
    └── quota_reset_test.go
```

## 6. 创建 Entity

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 创建 Entity |
| 方法 | POST |
| 路径 | `/open-api/v1/entities` |
| 说明 | 创建组织架构实体 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| name | string | Y | Entity 名称，全局唯一 | 长度 1-64；不能包含控制字符；不能有首尾空白；全局唯一 |
| type | string | Y | Entity 类型，必须引用已定义的 Entity-Type | 必须为已存在的 EntityTypeName |
| parent_id | string | N | 父 Entity ID，为空表示根节点 | 若非空，父 Entity 必须存在，且其父类型的 level 必须小于当前类型的 level |
| allow_models | []string | N | 允许访问的模型白名单，默认 ["*"] | 每个元素为 AIModel |
| block_models | []string | N | 禁止访问的模型黑名单，默认 [] | 每个元素为非空字符串 |
| quota_plan | object | N | 配额计划，同 API-Key quota_plan 结构（不含 balance） | `quota` ≥0；`unit` ∈ {`total_token`, `RMB`}；`reset_period` ∈ {never, weekly, monthly} |
| rate_limit_policy | object | N | 限流策略 | 同 RateLimitPolicy 类型 |
| route_rules | object | N | 路由规则 | 同 RouteRules 类型 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity 唯一标识 |
| name | string | Entity 名称 |
| type | string | Entity 类型 |
| parent_id | string | 父 Entity ID |
| allow_models | []string | 允许访问的模型白名单 |
| block_models | []string | 禁止访问的模型黑名单 |
| quota_plan | object | 配额计划（不含 balance） |
| rate_limit_policy | object | 限流策略 |
| route_rules | object | 路由规则 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-1-001 | 创建 Entity（仅必填） | 正常参数 | 验证默认值 |
| E-1-002 | 创建 Entity（含 quota_plan） | 正常参数 | 验证嵌套结构 |
| E-1-003 | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| E-1-004 | 缺少 type | 必填校验 | 验证 ErrNum=422 |
| E-1-005 | type 不存在 | 异常参数 | 验证 ErrNum=404 或 422 |
| E-1-006 | 重复 name | 业务规则 | 验证 ErrNum=555/556 |
| E-1-007 | 创建层级 Entity（合法 parent） | 正常参数 | 父 level 小于子 |
| E-1-008 | 创建层级 Entity（非法 parent level） | 异常参数 | 父 level 必须小于子 |
| E-1-009 | type 格式非法（含大写） | 合法性条件 | 验证 ErrNum=422 |
| E-1-010 | Entity name 包含首尾空白 | 合法性条件 | 验证 ErrNum=422 |

### 6.4 测试场景详细设计

#### 6.4.1 E-1-001：创建 Entity（仅必填）

##### 设计思路

验证仅传必填字段时，默认值正确填充，并返回系统生成的 id。

##### 前提数据准备

已创建 Entity-Type `department`。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/entities`。
2. 验证返回结构和默认值。

##### 请求参数

```json
{
    "name": "ent_root",
    "type": "department"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| name | "ent_root" | Equals |
| type | "department" | Equals |
| parent_id | null 或空字符串 | IsNullOrEmpty |
| allow_models | ["*"] | Equals |
| block_models | [] | IsEmpty |
| quota_plan | 非空对象 | IsObject |
| rate_limit_policy | 非空对象 | IsObject |
| route_rules | 非空对象 | IsObject |
| quota_plan.balance | 不存在 | NotExists |

---

#### 6.4.2 E-1-002：创建 Entity（含 quota_plan）

##### 设计思路

验证传入完整嵌套结构时，返回的 `quota_plan` 与输入一致且不含 `balance`。

##### 前提数据准备

已创建 Entity-Type `department`。

##### 执行步骤

1. 发送 POST 请求，传入完整参数。
2. 验证返回结构和字段。

##### 请求参数

```json
{
    "name": "ent_quota",
    "type": "department",
    "quota_plan": {
        "unlimited": false,
        "quota": 1000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "ent_quota" | Equals |
| quota_plan.unlimited | false | Equals |
| quota_plan.quota | 1000000 | Equals |
| quota_plan.unit | "total_token" | Equals |
| quota_plan.reset_period | "monthly" | Equals |
| quota_plan.balance | 不存在 | NotExists |

---

#### 6.4.3 E-1-003：缺少 name（必填校验）

##### 设计思路

验证 `name` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "type": "department"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "name" 的错误信息  
**Data**：null

---

#### 6.4.4 E-1-004：缺少 type（必填校验）

##### 设计思路

验证 `type` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `type`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "ent_notype"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "type" 的错误信息  
**Data**：null

---

#### 6.4.5 E-1-005：type 不存在（异常参数）

##### 设计思路

验证 `type` 必须引用已定义的 Entity-Type。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`type` 为不存在的类型。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "ent_bad_type",
    "type": "not_exist"
}
```

##### 预期返回结果

**ErrNum**：404 或 422  
**ErrMsg**：类型不存在的错误信息  
**Data**：null

---

#### 6.4.6 E-1-006：重复 name（业务规则）

##### 设计思路

验证 `name` 全局唯一。

##### 前提数据准备

已创建 `ent_dup`。

##### 执行步骤

1. 发送 POST 请求，使用重复的 `name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "ent_dup",
    "type": "department"
}
```

##### 预期返回结果

**ErrNum**：555 或 556  
**ErrMsg**：名称已存在的错误信息  
**Data**：null

---

#### 6.4.7 E-1-007：创建层级 Entity（合法 parent）

##### 设计思路

验证父 Entity 的 Entity-Type level 小于子 Entity 的 level 时，创建成功。

##### 前提数据准备

已创建 Entity-Type `department`(level=1)、`team`(level=2)，以及父 Entity。

##### 执行步骤

1. 发送 POST 请求，传入合法的 `parent_id`。
2. 验证返回的 `parent_id` 与输入一致。

##### 请求参数

```json
{
    "name": "ent_team",
    "type": "team",
    "parent_id": "<parent_id>"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "ent_team" | Equals |
| type | "team" | Equals |
| parent_id | "<parent_id>" | Equals |

---

#### 6.4.8 E-1-008：创建层级 Entity（非法 parent level）

##### 设计思路

验证父 Entity 的 Entity-Type level 不满足约束时返回错误。

##### 前提数据准备

已创建 level 较高子类型与 level 较低父类型（如 `department` level=1 的父 Entity 与 `team` level=2 的父 Entity）。

##### 执行步骤

1. 发送 POST 请求，将 `department` 类型 Entity 的 `parent_id` 指向 `team` 类型 Entity。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "ent_bad_parent",
    "type": "department",
    "parent_id": "<team_entity_id>"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：父节点层级非法的错误信息  
**Data**：null

---

#### 6.4.9 E-1-009：type 格式非法（含大写）

##### 设计思路

验证 `type` 必须是已存在的 EntityTypeName（小写等格式约束）。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`type` 包含大写字母。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "ent_bad_type_fmt",
    "type": "BadType"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 type 非法或类型不存在的错误信息  
**Data**：null

---

#### 6.4.10 E-1-010：Entity name 包含首尾空白（合法性条件）

##### 设计思路

验证 `name` 不能有首尾空白字符。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`name` 包含首尾空格。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": " badname ",
    "type": "department"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 name 非法的错误信息  
**Data**：null

---

#### 6.4.11 E-1-011：创建 Entity 并指定 RMB 配额（正常参数）

##### 设计思路

验证 Entity 创建时可指定 `quota_plan.unit=RMB` 与小数配额，返回结构正确。

##### 前提数据准备

已创建 Entity-Type `department`。

##### 执行步骤

1. 发送 POST 请求，`quota_plan.unit=RMB`、`quota=5555.5555`。
2. 验证返回 `quota_plan.unit=RMB`、`quota=5555.5555`、`quota_plan.balance` 不存在。
3. 调用 quota-plan 接口，验证 `balance.remaining=5555.5555`、`balance.used=0`。

##### 请求参数

```json
{
    "name": "ent_rmb_quota",
    "type": "department",
    "quota_plan": {
        "unlimited": false,
        "quota": 5555.5555,
        "unit": "RMB",
        "reset_period": "monthly"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "ent_rmb_quota" | Equals |
| quota_plan.unit | "RMB" | Equals |
| quota_plan.quota | 5555.5555 | Equals |
| quota_plan.balance | 不存在 | NotExists |

---

## 7. 查询 Entity 列表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询 Entity 列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/entities` |
| 说明 | 分页查询 Entity 列表 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认 1 |
| page_size | int | N | 每页条数，默认 20，最大 100 |
| id | string | N | 按 Entity ID 过滤 |
| name | string | N | 按 Entity 名称过滤 |
| type | string | N | 按类型过滤 |
| parent_id | string | N | 按父 Entity 过滤 |
| quota_plan_id | int64 | N | 按配额计划 ID 过滤 |
| route_rules_id | int64 | N | 按路由规则 ID 过滤 |

#### 7.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []Entity | Entity 列表 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

**list 对象字段说明**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity 唯一标识 |
| name | string | Entity 名称 |
| type | string | Entity 类型 |
| parent_id | string | 父 Entity ID |
| allow_models | []string | 允许访问的模型白名单 |
| block_models | []string | 禁止访问的模型黑名单 |
| quota_plan | object | 配额计划（不含 balance） |
| rate_limit_policy | object | 限流策略 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-2-001 | Entity 列表分页 | 正常参数 | 返回 {list, pagination} |
| E-2-002 | 按 type 过滤 | 正常参数 | 仅返回指定类型 |
| E-2-003 | 分页参数边界 | 边界值 | page=1&page_size=1 |

### 7.4 测试场景详细设计

#### 7.4.1 E-2-001：Entity 列表分页（正常参数）

##### 设计思路

验证列表接口返回分页结构，元素字段完整。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities`。
2. 验证返回结构和字段。

##### 请求参数

```
page=1&page_size=10
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[0].id | 非空字符串 | NotEmpty |
| list[0].quota_plan | 非空对象 | IsObject |
| list[0].quota_plan.balance | 不存在 | NotExists |
| pagination.page | 1 | Equals |
| pagination.page_size | 10 | Equals |
| pagination.total | ≥ 1 | Gte |

---

#### 7.4.2 E-2-002：按 type 过滤（正常参数）

##### 设计思路

验证按 `type` 过滤后，列表中仅返回指定类型的 Entity。

##### 前提数据准备

已创建不同类型的 Entity。

##### 执行步骤

1. 发送 GET 请求，`type=department`。
2. 验证列表中所有元素的 `type` 均为 `department`。

##### 请求参数

```
type=department
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].type | 全部为 "department" | Equals |

---

#### 7.4.3 E-2-003：分页参数边界（边界值）

##### 设计思路

验证分页参数边界，`page_size=1` 时返回单条记录。

##### 前提数据准备

已创建至少 2 个 Entity。

##### 执行步骤

1. 发送 GET 请求，`page=1&page_size=1`。
2. 验证 `list` 长度为 1，`pagination.total ≥ 2`。

##### 请求参数

```
page=1&page_size=1
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 长度为 1 | Len=1 |
| pagination.page | 1 | Equals |
| pagination.page_size | 1 | Equals |
| pagination.total | ≥ 2 | Gte |

---

## 8. 查询单个 Entity

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询单个 Entity |
| 方法 | GET |
| 路径 | `/open-api/v1/entities/{id}` |
| 说明 | 按 Entity ID 查询单个 Entity |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

#### 8.2.2 返回数据字段

同 6.2.2，`quota_plan` 不含 `balance`。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-3-001 | 查询单个 Entity | 正常参数 | 字段完整，不含 balance |
| E-3-002 | 查询不存在的 Entity | 异常参数 | 验证 ErrNum=404 |

### 8.4 测试场景详细设计

#### 8.4.1 E-3-001：查询单个 Entity（正常参数）

##### 设计思路

验证按 ID 查询 Entity 的基本功能。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities/{id}`。
2. 验证返回字段完整且 `quota_plan` 不含 `balance`。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| name | 非空字符串 | NotEmpty |
| quota_plan | 非空对象 | IsObject |
| quota_plan.balance | 不存在 | NotExists |

---

#### 8.4.2 E-3-002：查询不存在的 Entity（异常参数）

##### 设计思路

验证查询不存在的 ID 时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities/non_existent_id`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_id`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：Entity 不存在的错误信息  
**Data**：null

---

## 9. 全量更新 Entity

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 全量更新 Entity |
| 方法 | PUT |
| 路径 | `/open-api/v1/entities/{id}` |
| 说明 | 全量更新 Entity，`type` 不可修改 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

##### Body 参数

同创建接口。

#### 9.2.2 返回数据字段

同创建接口。

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-4-001 | 全量更新 Entity name | 正常参数 | name 更新，其余不变 |
| E-4-002 | 全量更新后查询一致性 | 返回数据 | PUT 后立即 GET，验证数据一致 |
| E-4-003 | 全量更新冲突 name | 业务规则 | 验证名称唯一约束 |
| E-4-004 | 全量更新修改 type | 业务规则 | type 不可修改 |
| E-4-005 | 全量更新非法 name（含首尾空白） | 合法性条件 | 验证 ErrNum=422 |

### 9.4 测试场景详细设计

#### 9.4.1 E-4-001：全量更新 Entity name（正常参数）

##### 设计思路

验证全量更新 `name` 成功，其余字段保持不变。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/entities/{id}`，传入完整 Body，`name` 改为新值。
2. 验证返回的 `name` 已更新，其余字段不变。

##### 请求参数

```json
{
    "name": "ent_updated",
    "type": "department",
    "allow_models": ["*"],
    "block_models": [],
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "ent_updated" | Equals |
| type | 与原 Entity 一致 | Equals |
| quota_plan.unlimited | true | Equals |

---

#### 9.4.2 E-4-002：全量更新后查询一致性（返回数据）

##### 设计思路

验证 PUT 更新成功后，立即通过 GET 查询，返回数据与更新请求一致。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PUT 请求更新 Entity。
2. 发送 GET 请求查询该 Entity。
3. 对比两次返回的数据是否一致。

##### 请求参数

```json
{
    "name": "ent_consistency",
    "type": "department",
    "allow_models": ["gpt-4"],
    "block_models": [],
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "ent_consistency" | Equals |
| allow_models | ["gpt-4"] | Equals |

---

#### 9.4.3 E-4-003：全量更新冲突 name（业务规则）

##### 设计思路

验证 `name` 全局唯一，更新为已存在的 name 时返回错误。

##### 前提数据准备

已创建两个 Entity：Entity1 和 Entity2。

##### 执行步骤

1. 发送 PUT 请求，用 Entity1 的 id 更新 name 为 Entity2 的 name。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "<entity2_name>",
    "type": "<entity1_type>",
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：555、556 或 500  
**ErrMsg**：名称冲突的错误信息  
**Data**：null

---

#### 9.4.4 E-4-004：全量更新修改 type（业务规则）

##### 设计思路

验证 `type` 字段不可修改。

##### 前提数据准备

已创建 Entity，类型为 `department`。

##### 执行步骤

1. 发送 PUT 请求，尝试将 `type` 修改为其他类型。
2. 验证返回 `type` 保持原值或返回错误。

##### 请求参数

```json
{
    "name": "<entity_name>",
    "type": "team",
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200（type 保持不变）或 422  
**ErrMsg**：success 或 type 不可修改的错误信息

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type | 原类型 | Equals |

---

#### 9.4.5 E-4-005：全量更新非法 name（含首尾空白）

##### 设计思路

验证全量更新时 `name` 同样受 EntityName 合法性条件约束。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PUT 请求，`name` 包含首尾空格。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": " badname ",
    "type": "department",
    "quota_plan": {
        "unlimited": true
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 name 非法的错误信息  
**Data**：null

---

#### 9.4.6 E-4-006：全量更新 quota_plan 切换为 RMB（正常参数）

##### 设计思路

验证全量更新可将 Entity 的 `quota_plan.unit` 切换为 `RMB`，并按新的金额配额重置余额。

##### 前提数据准备

已创建 `unit=total_token` 的 Entity。

##### 执行步骤

1. 发送 PUT 请求，`quota_plan.unit=RMB`、`quota=1234.56`。
2. 验证返回 `quota_plan.unit=RMB`、`quota=1234.56`。
3. 调用 quota-plan 接口，验证 `balance.remaining=1234.56`、`balance.used=0`。

##### 请求参数

```json
{
    "name": "ent_rmb_update",
    "type": "department",
    "allow_models": ["*"],
    "block_models": [],
    "quota_plan": {
        "unlimited": false,
        "quota": 1234.56,
        "unit": "RMB",
        "reset_period": "monthly"
    },
    "rate_limit_policy": {
        "enabled": false
    },
    "route_rules": {
        "enabled": false,
        "rules": []
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| quota_plan.unit | "RMB" | Equals |
| quota_plan.quota | 1234.56 | Equals |
| quota_plan.balance | 不存在 | NotExists |

---

## 10. 部分更新 Entity

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 部分更新 Entity |
| 方法 | PATCH |
| 路径 | `/open-api/v1/entities/{id}` |
| 说明 | 部分更新 Entity 字段 |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

##### Body 参数

同创建接口，仅传需修改字段。

#### 10.2.2 返回数据字段

同创建接口。

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-5-001 | 部分更新 allow_models | 正常参数 | allow_models 更新 |
| E-5-002 | 部分更新后查询一致性 | 返回数据 | PATCH 后立即 GET，验证数据一致 |
| E-5-003 | 部分更新非法 route_rules（规则名重复） | 合法性条件 | 验证 ErrNum=422 |

### 10.4 测试场景详细设计

#### 10.4.1 E-5-001：部分更新 allow_models（正常参数）

##### 设计思路

验证部分更新 `allow_models` 成功，其余字段保持不变。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entities/{id}`。
2. 验证返回的 `allow_models` 已更新。

##### 请求参数

```json
{
    "allow_models": ["gpt-4"]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| allow_models | ["gpt-4"] | Equals |
| type | 与原 Entity 一致 | Equals |

---

#### 10.4.2 E-5-002：部分更新后查询一致性（返回数据）

##### 设计思路

验证 PATCH 更新成功后，立即通过 GET 查询，返回数据与更新请求一致。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PATCH 请求更新 `block_models`。
2. 发送 GET 请求查询该 Entity。
3. 对比两次返回的 `block_models`。

##### 请求参数

```json
{
    "block_models": ["gpt-4-32k"]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| block_models | ["gpt-4-32k"] | Equals |

---

#### 10.4.3 E-5-003：部分更新非法 route_rules（规则名重复）

##### 设计思路

验证部分更新 `route_rules` 时同样受 RouteRules 合法性条件约束。

##### 前提数据准备

已创建 Entity。

##### 执行步骤

1. 发送 PATCH 请求，`route_rules.rules` 包含同名规则。
2. 验证返回错误码。

##### 请求参数

```json
{
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "dup",
                "Cond": "default_t()",
                "targets": [
                    {"ClusterName": "c1", "Weight": 100}
                ]
            },
            {
                "name": "dup",
                "Cond": "default_t()",
                "targets": [
                    {"ClusterName": "c2", "Weight": 100}
                ]
            }
        ]
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含规则名称重复的错误信息  
**Data**：null

---

#### 10.4.4 E-5-004：部分更新 quota_plan 切换为 RMB（正常参数）

##### 设计思路

验证 PATCH 可单独修改 Entity 的 `quota_plan.unit` 与 `quota`，切换为金额型配额后余额同步重置。

##### 前提数据准备

已创建 `unit=total_token` 的 Entity。

##### 执行步骤

1. 发送 PATCH 请求，`quota_plan.unit=RMB`、`quota=777.7777`。
2. 验证返回 `quota_plan.unit=RMB`、`quota=777.7777`。
3. 调用 quota-plan 接口，验证 `balance.remaining=777.7777`、`balance.used=0`。

##### 请求参数

```json
{
    "quota_plan": {
        "unlimited": false,
        "quota": 777.7777,
        "unit": "RMB",
        "reset_period": "weekly"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| quota_plan.unit | "RMB" | Equals |
| quota_plan.quota | 777.7777 | Equals |
| quota_plan.balance | 不存在 | NotExists |

---

## 11. 删除 Entity

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 删除 Entity |
| 方法 | DELETE |
| 路径 | `/open-api/v1/entities/{id}` |
| 说明 | 删除 Entity，有子节点或被挂载时禁止删除 |

### 11.2 接口参数说明

#### 11.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

#### 11.2.2 返回数据字段

Data 为 null。

### 11.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-6-001 | 删除 Entity | 正常参数 | 删除成功，再次查询返回 404 |
| E-6-002 | 删除存在子节点的 Entity | 业务规则 | 验证 ErrNum=409 |
| E-6-003 | 删除被 API-Key 挂载的 Entity | 业务规则 | 验证 ErrNum=409 |

### 11.4 测试场景详细设计

#### 11.4.1 E-6-001：删除 Entity（正常参数）

##### 设计思路

验证删除无子节点、未被挂载的 Entity 成功。

##### 前提数据准备

已创建无子节点、未被挂载的 Entity。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/entities/{id}`。
2. 验证返回成功。
3. 再次查询，验证返回 404。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 11.4.2 E-6-002：删除存在子节点的 Entity（业务规则）

##### 设计思路

验证存在子 Entity 的节点不可删除。

##### 前提数据准备

已创建父 Entity 及子 Entity。

##### 执行步骤

1. 发送 DELETE 请求到父 Entity id。
2. 验证返回错误码。

##### 请求参数

URI：父 Entity id

##### 预期返回结果

**ErrNum**：409  
**ErrMsg**：存在子节点无法删除的错误信息  
**Data**：null

---

#### 11.4.3 E-6-003：删除被 API-Key 挂载的 Entity（业务规则）

##### 设计思路

验证被 API-Key 挂载的 Entity 不可删除。

##### 前提数据准备

已创建 Entity 并被 API-Key 挂载。

##### 执行步骤

1. 发送 DELETE 请求到该 Entity id。
2. 验证返回错误码。

##### 请求参数

URI：Entity id

##### 预期返回结果

**ErrNum**：409  
**ErrMsg**：Entity 被挂载无法删除的错误信息  
**Data**：null

---

## 12. 查询配额计划

### 12.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 查询配额计划 |
| 方法 | GET |
| 路径 | `/open-api/v1/entities/{id}/quota-plan` |
| 说明 | 查询 Entity 的配额计划，含实时余额 |

### 12.2 接口参数说明

#### 12.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

#### 12.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| unlimited | bool | 是否无限配额 |
| pass_when_no_enough_quota | bool | 配额不足时是否放行 |
| quota | int64 | 配额总量 |
| unit | string | 配额单位 |
| reset_period | string | 配额重置周期 |
| balance | object | 余额状态，包含 used 和 remaining |

### 12.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-7-001 | 查询 Entity 配额计划 | 正常参数 | 返回完整 quota_plan 含 balance |

### 12.4 测试场景详细设计

#### 12.4.1 E-7-001：查询 Entity 配额计划（正常参数）

##### 设计思路

验证独立 quota-plan 接口返回完整配额计划与余额。

##### 前提数据准备

已创建非无限配额 Entity。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entities/{id}/quota-plan`。
2. 验证返回包含 `balance`。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| unlimited | false | Equals |
| quota | 与创建时一致 | Equals |
| balance | 非空对象 | IsObject |
| balance.used | 大于等于 0 | Gte(0) |
| balance.remaining | 大于等于 0 | Gte(0) |

---

#### 12.4.2 E-7-002：查询 RMB 配额余额精度（正常参数）

##### 设计思路

验证 Entity 的 `unit=RMB` 配额余额支持 8 位小数精度，且与数据库真实余额一致。

##### 前提数据准备

已创建 `unit=RMB`、`quota=2000.12345678` 的 Entity。

##### 执行步骤

1. 在测试数据库中将该 Entity 对应 `quota_balances` 的 `used` 更新为 `2.34567890`，`remaining` 更新为 `1997.77777788`。
2. 发送 GET 请求到 `/open-api/v1/entities/{id}/quota-plan`。
3. 验证返回 `unit=RMB`、`quota=2000.12345678`、`balance.used=2.34567890`、`balance.remaining=1997.77777788`。

##### 请求参数

URI：`id`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| unit | "RMB" | Equals |
| quota | 2000.12345678 | Equals |
| balance.used | 2.34567890 | Equals |
| balance.remaining | 1997.77777788 | Equals |

---

## 13. 重置配额余额

### 13.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口名称 | 重置配额余额 |
| 方法 | POST |
| 路径 | `/open-api/v1/entities/{id}/quota-plan/reset` |
| 说明 | 重置 Entity 的配额余额 |

### 13.2 接口参数说明

#### 13.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | Y | Entity 标识 |

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| quota | int64 | N | 重置后的配额总量 |
| reason | string | N | 重置原因 |

#### 13.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | Entity 标识 |
| previous_quota | int64 | 重置前配额 |
| new_quota | int64 | 重置后配额 |
| balance | object | 余额变更详情 |
| balance.previous_remaining | int64 | 重置前剩余量 |
| balance.new_remaining | int64 | 重置后剩余量 |
| balance.used | int64 | 当前已用量 |

### 13.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| E-8-001 | 重置配额余额 | 正常参数 | used=0，new_remaining=previous_quota |
| E-8-002 | 重置并修改 quota | 正常参数 | new_quota 和 new_remaining 同步更新 |

### 13.4 测试场景详细设计

#### 13.4.1 E-8-001：重置配额余额（正常参数）

##### 设计思路

验证不传 quota 时按当前 quota 重置余额。

##### 前提数据准备

已创建非无限配额 Entity。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`。
2. 验证返回的 `balance.used=0`，`new_remaining=previous_quota`。

##### 请求参数

```json
{
    "reason": "test reset"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 与 URI 一致 | Equals |
| previous_quota | 与当前 quota 一致 | Equals |
| new_quota | 与 previous_quota 一致 | Equals |
| balance.used | 0 | Equals |
| balance.new_remaining | 与 new_quota 一致 | Equals |

---

#### 13.4.2 E-8-002：重置并修改 quota（正常参数）

##### 设计思路

验证传入 quota 时同步更新 quota 并重置余额。

##### 前提数据准备

已创建非无限配额 Entity。

##### 执行步骤

1. 发送 POST 请求，传入新的 `quota`。
2. 验证返回的 `new_quota` 和 `new_remaining` 均为新值，`used=0`。

##### 请求参数

```json
{
    "quota": 200000,
    "reason": "reset"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| new_quota | 200000 | Equals |
| balance.new_remaining | 200000 | Equals |
| balance.used | 0 | Equals |

---

#### 13.4.3 E-8-003：重置 RMB 配额余额（正常参数）

##### 设计思路

验证 Entity 的 `unit=RMB` 配额重置可指定小数新配额，余额精度保持 8 位小数。

##### 前提数据准备

已创建 `unit=RMB`、`quota=50.5` 的 Entity，并产生一定用量。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/entities/{id}/quota-plan/reset`，`quota=300.1234`。
2. 验证返回 `previous_quota=50.5`、`new_quota=300.1234`。
3. 验证 `balance.used=0`、`balance.new_remaining=300.1234`。

##### 请求参数

```json
{
    "quota": 300.1234,
    "reason": "adjust entity rmb quota"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| previous_quota | 50.5 | Equals |
| new_quota | 300.1234 | Equals |
| balance.used | 0 | Equals |
| balance.new_remaining | 300.1234 | Equals |

---

## 14. 依赖与数据准备

1. 必须预先创建至少两种不同 `level` 的 Entity-Type 以验证层级约束。
2. 配额/余额用例依赖 Redis Mock。
3. 删除约束用例需要预先准备子 Entity 或挂载的 API-Key。

## 15. 注意事项

1. Entity 详情、创建、更新返回的 `quota_plan` 不含 `balance`，需通过独立 quota-plan 接口验证余额。
2. `name` 全局唯一，测试用例间注意清理。
3. 层级修改必须保证父节点 Entity-Type 的 `level` 小于当前节点。
4. 测试环境 `SkipTokenValidate=true`，无需认证头。
