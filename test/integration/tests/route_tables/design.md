# Route Tables 测试用例设计文档

## 1. 模块概述

Route Tables 模块用于查询系统中各类路由表的元数据列表（global/entity/apikey）。本模块为只读列表接口，路由表的实际创建/更新由其对应模块（Global Route Rules、Entity、API-Key）间接维护。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| RT-1 | 查询路由表列表 | GET | `/open-api/v1/route-tables` | 分页返回路由表元数据 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 查询路由表列表 | 10 |
| **合计** | **10** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
route_tables/
├── design.md
└── list/
    └── list_test.go
```

## 6. 查询路由表列表

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Route Tables |
| 接口名称 | 查询路由表列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/route-tables` |
| 说明 | 分页查询系统中各类路由表元数据 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | N | 页码，默认 1 |
| page_size | int | N | 每页条数，默认 20，最大 100 |
| sort_by | string | N | 排序字段 |
| sort_order | string | N | 排序方向，asc/desc，默认 desc |
| type | string | N | 按路由表类型过滤：global、entity、apikey |
| owner | string | N | 按所有者标识精确匹配过滤 |
| enabled | bool | N | 按启用状态过滤 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []RouteTable | 路由表列表 |
| list[].type | string | 路由表类型：global、entity、apikey |
| list[].owner | string | 所有者标识 |
| list[].enabled | bool | 是否启用 |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| RT-1-001 | 无参数查询路由表列表 | 正常参数 | 返回所有路由表元数据 |
| RT-1-002 | 按 type=global 过滤 | 正常参数 | 仅返回 global 类型 |
| RT-1-003 | 按 type=entity 过滤 | 正常参数 | 仅返回 entity 类型 |
| RT-1-004 | 按 type=apikey 过滤 | 正常参数 | 仅返回 apikey 类型 |
| RT-1-005 | 按 owner 精确匹配过滤 | 正常参数 | 验证 owner 过滤 |
| RT-1-006 | 按 enabled 过滤 | 正常参数 | 验证状态过滤 |
| RT-1-007 | 分页参数边界 | 边界值 | page=1&page_size=1 |
| RT-1-008 | page_size 超过最大值 | 异常参数 | page_size=101 |
| RT-1-009 | 非法 type 值 | 异常参数 | type=unknown |
| RT-1-010 | 空列表返回 | 边界值 | 无任何路由表时返回空列表 |

### 6.4 测试场景详细设计

#### 6.4.1 RT-1-001：无参数查询路由表列表（正常参数）

##### 设计思路

验证无查询参数时返回所有类型的路由表元数据。

##### 前提数据准备

- 已通过 `PUT /global-route-rules` 创建 global 路由表。
- 已通过 `POST /entities` 创建 entity 路由表。
- 已通过 `POST /api-keys` 创建 apikey 路由表。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/route-tables`。
2. 验证响应状态码和返回结构。
3. 验证 `list` 中至少包含 global、entity、apikey 三类元素。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].type | 包含 global/entity/apikey | Contains |
| pagination.total | ≥ 3 | Gte |

---

#### 6.4.2 RT-1-002：按 type=global 过滤（正常参数）

##### 设计思路

验证按 `type=global` 过滤后，列表中仅返回 global 类型路由表。

##### 前提数据准备

已存在 global 路由表。

##### 执行步骤

1. 发送 GET 请求，`type=global`。
2. 验证列表中所有元素的 `type` 均为 `global`。

##### 请求参数

```
type=global
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].type | 全部为 "global" | Equals |

---

#### 6.4.3 RT-1-003：按 type=entity 过滤（正常参数）

##### 设计思路

验证按 `type=entity` 过滤后，列表中仅返回 entity 类型路由表。

##### 前提数据准备

已存在 entity 路由表。

##### 执行步骤

1. 发送 GET 请求，`type=entity`。
2. 验证列表中所有元素的 `type` 均为 `entity`。

##### 请求参数

```
type=entity
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].type | 全部为 "entity" | Equals |

---

#### 6.4.4 RT-1-004：按 type=apikey 过滤（正常参数）

##### 设计思路

验证按 `type=apikey` 过滤后，列表中仅返回 apikey 类型路由表。

##### 前提数据准备

已存在 apikey 路由表。

##### 执行步骤

1. 发送 GET 请求，`type=apikey`。
2. 验证列表中所有元素的 `type` 均为 `apikey`。

##### 请求参数

```
type=apikey
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].type | 全部为 "apikey" | Equals |

---

#### 6.4.5 RT-1-005：按 owner 精确匹配过滤（正常参数）

##### 设计思路

验证按 `owner` 精确匹配过滤后，列表中仅返回指定所有者的路由表。

##### 前提数据准备

已创建 API-Key，其 id 为 `apikey-001`。

##### 执行步骤

1. 发送 GET 请求，`owner=apikey-001`。
2. 验证列表中所有元素的 `owner` 均为 `apikey-001`。

##### 请求参数

```
owner=apikey-001
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].owner | 全部为 "apikey-001" | Equals |

---

#### 6.4.6 RT-1-006：按 enabled 过滤（正常参数）

##### 设计思路

验证按 `enabled` 过滤后，列表中仅返回指定状态的路由表。

##### 前提数据准备

同时存在 enabled=true 与 enabled=false 的路由表。

##### 执行步骤

1. 发送 GET 请求，`enabled=false`。
2. 验证列表中所有元素的 `enabled` 均为 `false`。

##### 请求参数

```
enabled=false
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 数组 | IsArray |
| list[*].enabled | 全部为 false | Equals |

---

#### 6.4.7 RT-1-007：分页参数边界（边界值）

##### 设计思路

验证分页参数边界，`page_size=1` 时返回单条记录且 `pagination.total` 正确。

##### 前提数据准备

已创建至少 2 条路由表。

##### 执行步骤

1. 发送 GET 请求，`page=1&page_size=1`。
2. 验证 `list` 长度为 1，且 `pagination.total ≥ 2`。

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

#### 6.4.8 RT-1-008：page_size 超过最大值（异常参数）

##### 设计思路

验证 `page_size` 超过最大值 100 时的处理行为。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求，`page_size=101`。
2. 验证返回错误码或按最大值 100 截断返回。

##### 请求参数

```
page_size=101
```

##### 预期返回结果

**ErrNum**：422 或 200（以接口实现为准）  
**ErrMsg**：若为 422，提示 page_size 非法；若为 200，按 page_size=100 返回

---

#### 6.4.9 RT-1-009：非法 type 值（异常参数）

##### 设计思路

验证 `type` 传非枚举值时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求，`type=unknown`。
2. 验证返回错误码。

##### 请求参数

```
type=unknown
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "type" 非法的错误信息  
**Data**：null

---

#### 6.4.10 RT-1-010：空列表返回（边界值）

##### 设计思路

验证系统中无任何路由表时返回空列表与零总数。

##### 前提数据准备

全新数据库，无任何路由表。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/route-tables`。
2. 验证 `list` 为空数组，`pagination.total=0`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| list | 空数组 | IsEmpty |
| pagination.total | 0 | Equals |

---

## 7. 依赖与数据准备

- 查询前需通过其他模块预置数据：
  - `global` 类型：调用 `PUT /global-route-rules`。
  - `entity` 类型：调用 `POST /entities`。
  - `apikey` 类型：调用 `POST /api-keys`。

## 8. 注意事项

1. 本模块仅提供列表查询，不暴露创建/更新/删除能力。
2. `type`、`owner`、`enabled` 的组合过滤应同时生效。
3. 返回数据结构为 `{ list: [...], pagination: {...} }`。
4. 测试环境 `SkipTokenValidate=true`，无需认证头。
