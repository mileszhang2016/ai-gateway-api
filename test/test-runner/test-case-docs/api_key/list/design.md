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