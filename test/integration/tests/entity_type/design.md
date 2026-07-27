# Entity-Type 测试用例设计文档

## 1. 模块概述

Entity-Type 模块用于定义 Entity 的类型及层级级别，是创建 Entity 的前置条件。v0.3.0 无结构性变更。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| ET-1 | 创建 Entity-Type | POST | `/open-api/v1/entity-types` | - |
| ET-2 | 查询 Entity-Type 列表 | GET | `/open-api/v1/entity-types` | 分页 |
| ET-3 | 查询单个 Entity-Type | GET | `/open-api/v1/entity-types/{type_name}` | - |
| ET-4 | 更新 Entity-Type | PATCH | `/open-api/v1/entity-types/{type_name}` | 仅更新描述 |
| ET-5 | 删除 Entity-Type | DELETE | `/open-api/v1/entity-types/{type_name}` | 被引用时禁止删除 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建 Entity-Type | 6 |
| 查询 Entity-Type 列表 | 2 |
| 查询单个 Entity-Type | 2 |
| 更新 Entity-Type | 3 |
| 删除 Entity-Type | 2 |
| **合计** | **15** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
entity_type/
├── design.md
├── create/
│   └── create_test.go
├── list/
│   └── list_test.go
├── detail/
│   └── detail_test.go
├── update/
│   └── update_test.go
└── delete/
    └── delete_test.go
```

## 6. 创建 Entity-Type

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 创建 Entity-Type |
| 方法 | POST |
| 路径 | `/open-api/v1/entity-types` |
| 说明 | 创建 Entity 类型定义 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名，全局唯一，1-32 字符，仅含小写字母、数字、下划线、连字符 |
| description | string | N | 类型描述 |
| level | int | Y | 层级级别，取值范围 1-5 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| type_name | string | 类型名 |
| description | string | 描述 |
| level | int | 层级级别 |
| create_time | int64 | 创建时间 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-1-001 | 创建 Entity-Type（完整参数） | 正常参数 | 返回完整字段 |
| ET-1-002 | 创建 Entity-Type（仅必填） | 正常参数 | description 为空 |
| ET-1-003 | 缺少 type_name | 必填校验 | 验证 ErrNum=422 |
| ET-1-004 | 缺少 level | 必填校验 | 验证 ErrNum=422 |
| ET-1-005 | 重复创建同名 Entity-Type | 业务规则 | 验证 ErrNum=555/556 |
| ET-1-006 | level 超出范围 | 边界值 | 验证 ErrNum=422 |

### 6.4 测试场景详细设计

#### 6.4.1 ET-1-001：创建 Entity-Type（完整参数）

##### 设计思路

验证传入完整参数时成功创建 Entity-Type。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/entity-types`。
2. 验证响应状态码和返回结构。

##### 请求参数

```json
{
    "type_name": "department",
    "description": "一级部门",
    "level": 1
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "department" | Equals |
| description | "一级部门" | Equals |
| level | 1 | Equals |
| create_time | 大于 0 的整数 | Gt(0) |

---

#### 6.4.2 ET-1-002：创建 Entity-Type（仅必填）

##### 设计思路

验证仅传必填字段时，`description` 为空或 null，其他字段正确。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，仅传 `type_name` 和 `level`。
2. 验证返回字段。

##### 请求参数

```json
{
    "type_name": "team",
    "level": 2
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "team" | Equals |
| description | 空字符串或 null | EmptyOrNull |
| level | 2 | Equals |

---

#### 6.4.3 ET-1-003：缺少 type_name（必填校验）

##### 设计思路

验证 `type_name` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `type_name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "level": 1
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "type_name" 的错误信息  
**Data**：null

---

#### 6.4.4 ET-1-004：缺少 level（必填校验）

##### 设计思路

验证 `level` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `level`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "type_name": "project"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "level" 的错误信息  
**Data**：null

---

#### 6.4.5 ET-1-005：重复创建同名 Entity-Type（业务规则）

##### 设计思路

验证 `type_name` 全局唯一，重复创建时返回错误。

##### 前提数据准备

已创建 `department`。

##### 执行步骤

1. 发送 POST 请求，使用重复的 `type_name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "type_name": "department",
    "level": 1
}
```

##### 预期返回结果

**ErrNum**：555 或 556  
**ErrMsg**：类型名已存在的错误信息  
**Data**：null

---

#### 6.4.6 ET-1-006：level 超出范围（边界值）

##### 设计思路

验证 `level` 取值范围为 1-5，超出时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`level=6`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "type_name": "bad_level",
    "level": 6
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "level" 非法的错误信息  
**Data**：null

---

## 7. 查询 Entity-Type 列表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 查询 Entity-Type 列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/entity-types` |
| 说明 | 分页查询 Entity-Type 列表 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认 1 |
| page_size | int | N | 每页条数，默认 20，最大 100 |
| id | int64 | N | 按内部 ID 过滤 |
| type_name | string | N | 按类型名过滤 |
| level | int | N | 按层级级别过滤 |

#### 7.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []EntityType | Entity-Type 列表 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-2-001 | 查询 Entity-Type 列表 | 正常参数 | 返回分页结构 |
| ET-2-002 | 分页参数边界 | 边界值 | page=1&page_size=1 |

### 7.4 测试场景详细设计

#### 7.4.1 ET-2-001：查询 Entity-Type 列表（正常参数）

##### 设计思路

验证列表接口返回分页结构，元素字段完整。

##### 前提数据准备

已创建至少 1 个 Entity-Type。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types`。
2. 验证返回结构和字段。

##### 请求参数

```
page=1&page_size=20
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[0].type_name | 非空字符串 | NotEmpty |
| list[0].level | 1-5 的整数 | Range[1,5] |
| pagination.page | 1 | Equals |
| pagination.page_size | 20 | Equals |
| pagination.total | ≥ 1 | Gte |

---

#### 7.4.2 ET-2-002：分页参数边界（边界值）

##### 设计思路

验证分页参数边界，`page_size=1` 时返回单条记录。

##### 前提数据准备

已创建至少 2 个 Entity-Type。

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

## 8. 查询单个 Entity-Type

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 查询单个 Entity-Type |
| 方法 | GET |
| 路径 | `/open-api/v1/entity-types/{type_name}` |
| 说明 | 按类型名查询单个 Entity-Type |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

#### 8.2.2 返回数据字段

同 6.2.2。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-3-001 | 查询单个 Entity-Type | 正常参数 | 返回字段完整 |
| ET-3-002 | 查询不存在的 Entity-Type | 异常参数 | 验证 ErrNum=404 |

### 8.4 测试场景详细设计

#### 8.4.1 ET-3-001：查询单个 Entity-Type（正常参数）

##### 设计思路

验证按类型名查询 Entity-Type 的基本功能。

##### 前提数据准备

已创建 `department`。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types/department`。
2. 验证返回字段完整。

##### 请求参数

URI：`department`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "department" | Equals |
| description | 与创建时一致 | Equals |
| level | 与创建时一致 | Equals |
| create_time | 大于 0 | Gt(0) |

---

#### 8.4.2 ET-3-002：查询不存在的 Entity-Type（异常参数）

##### 设计思路

验证查询不存在的类型名时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/entity-types/non_existent`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：类型不存在的错误信息  
**Data**：null

---

## 9. 更新 Entity-Type

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 更新 Entity-Type |
| 方法 | PATCH |
| 路径 | `/open-api/v1/entity-types/{type_name}` |
| 说明 | 仅更新 Entity-Type 描述 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| description | string | N | 类型描述 |

#### 9.2.2 返回数据字段

同 6.2.2。

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-4-001 | 更新 Entity-Type 描述 | 正常参数 | description 更新，其余不变 |
| ET-4-002 | 更新后查询一致性 | 返回数据 | PATCH 后立即 GET，验证数据一致 |
| ET-4-003 | 更新不存在的 Entity-Type | 异常参数 | 验证 ErrNum=404 |

### 9.4 测试场景详细设计

#### 9.4.1 ET-4-001：更新 Entity-Type 描述（正常参数）

##### 设计思路

验证部分更新 `description` 成功，其余字段保持不变。

##### 前提数据准备

已创建 `department`。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entity-types/department`。
2. 验证返回的 `description` 已更新，`level` 保持不变。

##### 请求参数

```json
{
    "description": "更新后的描述"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| type_name | "department" | Equals |
| description | "更新后的描述" | Equals |
| level | 与创建时一致 | Equals |

---

#### 9.4.2 ET-4-002：更新后查询一致性（返回数据）

##### 设计思路

验证 PATCH 更新成功后，立即通过 GET 查询，返回的描述与更新请求一致。

##### 前提数据准备

已创建 `department`。

##### 执行步骤

1. 发送 PATCH 请求更新描述。
2. 发送 GET 请求查询该 Entity-Type。
3. 对比两次返回的 `description`。

##### 请求参数

```json
{
    "description": "一致性校验描述"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| description | "一致性校验描述" | Equals |

---

#### 9.4.3 ET-4-003：更新不存在的 Entity-Type（异常参数）

##### 设计思路

验证更新不存在的类型名时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/entity-types/non_existent`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent`  
Body：
```json
{
    "description": "x"
}
```

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：类型不存在的错误信息  
**Data**：null

---

## 10. 删除 Entity-Type

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Entity-Type |
| 接口名称 | 删除 Entity-Type |
| 方法 | DELETE |
| 路径 | `/open-api/v1/entity-types/{type_name}` |
| 说明 | 删除 Entity-Type，被引用时禁止删除 |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type_name | string | Y | 类型名 |

#### 10.2.2 返回数据字段

Data 为 null。

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ET-5-001 | 删除 Entity-Type | 正常参数 | 删除成功，再次查询返回 404 |
| ET-5-002 | 删除被 Entity 引用的 Entity-Type | 业务规则 | 验证 ErrNum=409 |

### 10.4 测试场景详细设计

#### 10.4.1 ET-5-001：删除 Entity-Type（正常参数）

##### 设计思路

验证删除未被引用的 Entity-Type 成功。

##### 前提数据准备

已创建未被引用的 Entity-Type `to_delete`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/entity-types/to_delete`。
2. 验证返回成功。
3. 再次查询，验证返回 404。

##### 请求参数

URI：`to_delete`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 10.4.2 ET-5-002：删除被 Entity 引用的 Entity-Type（业务规则）

##### 设计思路

验证被 Entity 引用的 Entity-Type 不可删除。

##### 前提数据准备

已创建 Entity-Type `referenced_type`，并被某个 Entity 引用。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/entity-types/referenced_type`。
2. 验证返回错误码。

##### 请求参数

URI：`referenced_type`

##### 预期返回结果

**ErrNum**：409  
**ErrMsg**：类型被引用无法删除的错误信息  
**Data**：null

---

## 11. 依赖与数据准备

1. 删除被引用用例需要预先创建对应 Entity。
2. Entity-Type 名称全局唯一，测试间注意清理。

## 12. 注意事项

1. v0.3.0 接口定义与基线一致，重点验证层级约束与删除依赖。
2. `level` 数字越小级别越高，后续 Entity 父子关系校验依赖此规则。
3. 测试环境 `SkipTokenValidate=true`，无需认证头。
