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
| page | int | 否 | 页码 | 1 |
| page_size | int | 否 | 每页条数 | 20（最大100） |
| enabled | bool | 否 | 是否启用过滤 | - |
| entity_id | string | 否 | 按挂载的 Entity ID 过滤 | - |
| unlimited_quota | bool | 否 | 是否无限配额过滤 | - |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| list | []object | API-Key 列表（quota_plan 含 balance） |
| pagination | object | 分页信息 |
| pagination.page | int | 当前页码 |
| pagination.page_size | int | 每页条数 |
| pagination.total | int | 总条数 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-2-001 | 默认分页查询（无参数） | 正常参数 | 验证默认分页参数生效 |
| AK-2-002 | 指定分页参数 | 正常参数 | page=1, page_size=5 |
| AK-2-003 | 按 enabled 过滤 | 正常参数 | 只查询启用的 API-Key |
| AK-2-004 | page_size=100（最大值） | 边界值 | 验证最大分页支持 |
| AK-2-005 | page_size=101（超最大值） | 边界值 | 验证超出限制时的行为 |
| AK-2-006 | 验证分页返回结构 | 返回数据校验 | 验证 list 和 pagination 结构 |
| AK-2-007 | 查询空列表 | 正常参数 | 新建环境无数据时查询 |

---

## 四、测试场景详细设计

---

### AK-2-001：默认分页查询（无参数）

#### 设计思路

验证不带任何参数时，接口使用默认分页（page=1, page_size=20）返回数据。

#### 前提数据准备

- 先创建 3 个 API-Key

#### 执行步骤

1. 创建 3 个 API-Key
2. 发送 GET 请求（无参数）
3. 验证返回 list 长度和分页信息

#### 预期返回结果

**ErrNum**：200  
**Data.list**：长度 ≥ 3  
**Data.pagination.page**：1  
**Data.pagination.page_size**：20  
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
**Data.list**：长度 ≥ 1，所有元素 enabled=true

---

### AK-2-004：page_size=100（最大值）

#### 设计思路

验证 page_size 取最大值 100 时接口正常。

#### 执行步骤

1. 发送 GET 请求：`?page_size=100`
2. 验证返回成功

#### 预期返回结果

**ErrNum**：200  
**Data.pagination.page_size**：100

---

### AK-2-005：page_size=101（超最大值）

#### 设计思路

验证 page_size 超过最大值 100 时的行为（可能被截断为 100 或返回错误）。

#### 执行步骤

1. 发送 GET 请求：`?page_size=101`
2. 验证接口行为

#### 预期返回结果

**ErrNum**：200（可能被截断为 100）或 422

---

### AK-2-006：验证分页返回结构

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

**Data 顶层校验**：

| 键名 | 预期类型 |
|------|---------|
| list | array |
| pagination | object |
| pagination.page | number |
| pagination.page_size | number |
| pagination.total | number |

**Data.list[0] 字段**：

| 键名 | 预期类型 |
|------|---------|
| id | string |
| key | string |
| description | string |
| enabled | bool |
| create_time | number |
| quota_plan | object（含 balance） |

---

### AK-2-007：查询空列表

#### 设计思路

验证新环境中没有 API-Key 时，接口返回空列表。

#### 前提数据准备

- 无需预先创建数据（新测试环境）

#### 执行步骤

1. 直接发送 GET 请求
2. 验证返回空列表

#### 预期返回结果

**ErrNum**：200  
**Data.list**：长度 = 0  
**Data.pagination.total**：0