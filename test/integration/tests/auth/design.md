# Auth 模块测试用例设计文档

## 模块概述

Auth 模块负责用户认证与授权管理，包括：
- 用户管理（创建、删除、重置密码、列表、设置管理员、绑定产品线）
- Session Key 管理（创建、删除）
- Token 管理（创建、删除、详情、列表、按产品线查询）

## 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| AUTH-1 | 创建用户 | POST | `/open-api/v1/users` | 创建新用户 |
| AUTH-2 | 删除用户 | DELETE | `/open-api/v1/users/{user_name}` | 删除指定用户 |
| AUTH-3 | 重置密码 | PATCH | `/open-api/v1/users/{user_name}/passwd` | 修改用户密码 |
| AUTH-4 | 用户列表 | GET | `/open-api/v1/users` | 获取所有用户列表 |
| AUTH-5 | 设置管理员 | PATCH | `/open-api/v1/users/{user_name}/is_admin` | 设置用户管理员权限 |
| AUTH-6 | 绑定产品线 | POST | `/open-api/v1/users/{user_name}/products/{product_name}` | 为用户绑定产品线 |
| AUTH-7 | 解除产品线绑定 | DELETE | `/open-api/v1/users/{user_name}/products/{product_name}` | 解除用户产品线绑定 |
| AUTH-8 | 按产品线查用户 | GET | `/open-api/v1/users/actions/search-by-product/{product_name}` | 查询指定产品线的授权用户 |
| AUTH-9 | 创建Session Key | POST | `/open-api/v1/session-keys` | 用户名密码登录获取session key |
| AUTH-10 | 删除Session Key | DELETE | `/open-api/v1/session-keys/{session_key}` | 删除session key |
| AUTH-11 | 创建Token | POST | `/open-api/v1/tokens` | 创建Token并绑定产品线 |
| AUTH-12 | 删除Token | DELETE | `/open-api/v1/tokens/{token_name}` | 删除指定Token |
| AUTH-13 | Token详情 | GET | `/open-api/v1/tokens/{token_name}` | 查询Token详情 |
| AUTH-14 | Token列表 | GET | `/open-api/v1/tokens` | 获取所有Token列表 |
| AUTH-15 | 按产品线查Token | GET | `/open-api/v1/tokens/actions/search-by-product/{product_name}` | 查询指定产品线的Token |

## 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建用户 | 7 |
| 删除用户 | 2 |
| 重置密码 | 5 |
| 用户列表 | 2 |
| 设置管理员 | 4 |
| 绑定产品线 | 3 |
| 解除产品线绑定 | 2 |
| 按产品线查用户 | 2 |
| 创建Session Key | 4 |
| 删除Session Key | 2 |
| 创建Token | 5 |
| 删除Token | 2 |
| Token详情 | 2 |
| Token列表 | 2 |
| 按产品线查Token | 2 |
| **合计** | **45** |

## 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 目录结构

```

├── README.md
├── create_user/design.md
├── delete_user/design.md
├── reset_password/design.md
├── list_users/design.md
├── set_admin/design.md
├── bind_product/design.md
├── unbind_product/design.md
├── search_by_product/design.md
├── create_session_key/design.md
├── delete_session_key/design.md
├── create_token/design.md
├── delete_token/design.md
├── token_detail/design.md
├── token_list/design.md
└── search_token_by_product/design.md
```

---

# 绑定产品线 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 绑定产品线 |
| 方法 | POST |
| 路径 | /open-api/v1/auth/users/{user_name}/products/{product_name} |
| 说明 | 为用户增加某个产品线的授权 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 用户名 |
| product_name | string | 是 | 产品线名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-6-001 | 正常绑定产品线 | 正常参数 | 验证 ErrNum=200 |
| AUTH-6-002 | 绑定不存在的产品线 | 异常参数 | 验证 ErrNum=404 |
| AUTH-6-003 | 为不存在的用户绑定产品线 | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-6-001：正常绑定产品线（正常参数）

#### 设计思路

验证为用户绑定产品线接口的基本功能：绑定已存在的用户到指定产品线，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_bind", password="password@123", is_admin=false
- 确保产品线 "product_demo" 存在

#### 执行步骤

1. 先创建用户
2. 发送 POST 请求到 `/open-api/v1/auth/users/test_user_bind/products/product_demo`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_bind |
| product_name | product_demo |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-6-002：绑定不存在的产品线（异常参数）

#### 设计思路

验证绑定不存在的产品线时，接口应返回资源不存在错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_bind_none", password="password@123", is_admin=false
- 确保产品线 "non_existent_product" 不存在

#### 执行步骤

1. 先创建用户
2. 发送 POST 请求到 `/open-api/v1/auth/users/test_user_bind_none/products/non_existent_product`
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_bind_none |
| product_name | non_existent_product |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含产品线不存在的错误信息  
**Data**：null

---

### AUTH-6-003：为不存在的用户绑定产品线（异常参数）

#### 设计思路

验证为不存在的用户绑定产品线时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保用户 "non_existent_user_bind" 不存在

#### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/users/non_existent_user_bind/products/product_demo`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | non_existent_user_bind |
| product_name | product_demo |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含用户不存在的错误信息  
**Data**：null

---

# 创建Session Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建Session Key |
| 方法 | POST |
| 路径 | /open-api/v1/auth/session-keys |
| 说明 | 使用用户名和密码登录，获取会话密钥 |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 用户名 |
| password | string | 是 | 用户密码 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| session_key | string | 会话密钥 |
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-9-001 | 正确用户名密码登录 | 正常参数 | 返回 session_key、user_name、is_admin |
| AUTH-9-002 | 密码错误 | 异常参数 | 验证 ErrNum=401 |
| AUTH-9-003 | 用户不存在 | 异常参数 | 验证 ErrNum=401 |
| AUTH-9-004 | 缺少 user_name | 必填校验 | 验证 ErrNum=422 |

---

## 四、测试场景详细设计

---

### AUTH-9-001：正确用户名密码登录（正常参数）

#### 设计思路

验证使用正确的用户名和密码登录，获取 Session Key 的基本功能，确认返回完整的字段信息。

#### 前提数据准备

- 预先创建用户：user_name="test_user_login", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 POST 请求到 `/open-api/v1/auth/session-keys`，传入正确的用户名和密码
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "user_name": "test_user_login",
    "password": "password@123"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| session_key | 非空字符串 | NotEmpty |
| user_name | "test_user_login" | Equals |
| is_admin | false | Equals |

---

### AUTH-9-002：密码错误（异常参数）

#### 设计思路

验证使用错误密码登录时，接口应返回认证失败错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_wrongpwd", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 POST 请求到 `/open-api/v1/auth/session-keys`，传入错误的密码
3. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "test_user_wrongpwd",
    "password": "wrongpassword"
}
```

#### 预期返回结果

**ErrNum**：401  
**ErrMsg**：包含认证失败的错误信息  
**Data**：null

---

### AUTH-9-003：用户不存在（异常参数）

#### 设计思路

验证使用不存在的用户名登录时，接口应返回认证失败错误。

#### 前提数据准备

- 确保用户 "non_existent_login" 不存在

#### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/session-keys`，传入不存在的用户名
2. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "non_existent_login",
    "password": "password@123"
}
```

#### 预期返回结果

**ErrNum**：401  
**ErrMsg**：包含认证失败的错误信息  
**Data**：null

---

### AUTH-9-004：缺少 user_name（必填校验）

#### 设计思路

验证 `user_name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 user_name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/session-keys`
3. 验证返回错误码

#### 请求参数

```json
{
    "password": "password@123"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "user_name" 的错误信息  
**Data**：null

---

# 创建Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建Token |
| 方法 | POST |
| 路径 | /open-api/v1/auth/tokens |
| 说明 | 创建Token（同时完成产品线绑定） |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | Token名字，必须全局唯一 |
| scope | string | 是 | 权限范围（Product/Support/System） |
| product_name | string | 是 | 产品线名（scope为Product时必填） |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | Token 值 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-11-001 | 创建 Product scope Token | 正常参数 | 返回 token |
| AUTH-11-002 | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-003 | 缺少 scope | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-004 | scope=Product 缺少 product_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-005 | 重复创建同名 Token | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### AUTH-11-001：创建 Product scope Token（正常参数）

#### 设计思路

验证创建 Token 接口的基本功能：传入完整参数，创建 Product scope 的 Token，确认接口返回成功并返回 token 值。

#### 前提数据准备

- 确保产品线 "product_token" 存在

#### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/tokens`，传入完整参数
2. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "name": "test_token_001",
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| token | 非空字符串 | NotEmpty |

---

### AUTH-11-002：缺少 name（必填校验）

#### 设计思路

验证 `name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "name" 的错误信息  
**Data**：null

---

### AUTH-11-003：缺少 scope（必填校验）

#### 设计思路

验证 `scope` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 scope 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_002",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "scope" 的错误信息  
**Data**：null

---

### AUTH-11-004：scope=Product 缺少 product_name（必填校验）

#### 设计思路

验证当 scope=Product 时，`product_name` 为必填字段，不传该字段时接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：scope=Product 但缺少 product_name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_003",
    "scope": "Product"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "product_name" 的错误信息  
**Data**：null

---

### AUTH-11-005：重复创建同名 Token（业务规则）

#### 设计思路

验证 Token name 必须全局唯一，当尝试创建已存在的 Token name 时，接口应返回业务错误。

#### 前提数据准备

- 预先创建 Token：name="test_token_dup", scope="Product", product_name="product_token"

#### 执行步骤

1. 先创建 Token
2. 使用相同 name 再次发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_dup",
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：555  
**ErrMsg**：包含重复 Token name 的错误信息  
**Data**：null

---

# 创建用户 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建用户 |
| 方法 | POST |
| 路径 | /open-api/v1/auth/users |
| 说明 | 创建新用户，需要提供用户名、密码和是否为管理员 |

---

## 二、接口参数说明

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 用户名，必须全局唯一 |
| password | string | 是 | 用户密码 |
| is_admin | bool | 是 | 是否为系统管理员 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-1-001 | 创建普通用户 | 正常参数 | user_name+password+is_admin=false |
| AUTH-1-002 | 创建管理员用户 | 正常参数 | is_admin=true |
| AUTH-1-003 | 缺少 user_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-004 | 缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-005 | 缺少 is_admin | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-006 | 重复创建同名用户 | 业务规则 | 验证 ErrNum=555 |
| AUTH-1-007 | user_name 为空字符串 | 边界值 | 验证 ErrNum=422 |

---

## 四、测试场景详细设计

---

### AUTH-1-001：创建普通用户（正常参数）

#### 设计思路

验证创建普通用户接口的基本功能：传入完整参数，is_admin=false，确认接口返回成功。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：包含 user_name、password、is_admin=false
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "user_name": "test_user_001",
    "password": "password@123",
    "is_admin": false
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-1-002：创建管理员用户（正常参数）

#### 设计思路

验证创建管理员用户的场景，is_admin=true，确认接口返回成功。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：包含 user_name、password、is_admin=true
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "user_name": "test_admin_001",
    "password": "password@123",
    "is_admin": true
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-1-003：缺少 user_name（必填校验）

#### 设计思路

验证 `user_name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 user_name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证返回错误码

#### 请求参数

```json
{
    "password": "password@123",
    "is_admin": false
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "user_name" 的错误信息  
**Data**：null

---

### AUTH-1-004：缺少 password（必填校验）

#### 设计思路

验证 `password` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 password 字段
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "test_user_002",
    "is_admin": false
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "password" 的错误信息  
**Data**：null

---

### AUTH-1-005：缺少 is_admin（必填校验）

#### 设计思路

验证 `is_admin` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 is_admin 字段
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "test_user_003",
    "password": "password@123"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "is_admin" 的错误信息  
**Data**：null

---

### AUTH-1-006：重复创建同名用户（业务规则）

#### 设计思路

验证用户名必须全局唯一，当尝试创建已存在的用户名时，接口应返回业务错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_dup", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户 test_user_dup
2. 使用相同 user_name 再次发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "test_user_dup",
    "password": "password@456",
    "is_admin": true
}
```

#### 预期返回结果

**ErrNum**：555  
**ErrMsg**：包含重复用户名的错误信息  
**Data**：null

---

### AUTH-1-007：user_name 为空字符串（边界值）

#### 设计思路

验证 user_name 为空字符串时的处理，确认接口返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：user_name 为空字符串
2. 发送 POST 请求到 `/open-api/v1/auth/users`
3. 验证返回错误码

#### 请求参数

```json
{
    "user_name": "",
    "password": "password@123",
    "is_admin": false
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "user_name" 的错误信息  
**Data**：null

---

# 删除Session Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除Session Key |
| 方法 | DELETE |
| 路径 | /open-api/v1/auth/session-keys/{session_key} |
| 说明 | 删除指定的会话密钥 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_key | string | 是 | 待删除的会话密钥 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-10-001 | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-10-002 | 删除不存在的 key | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-10-001：正常删除（正常参数）

#### 设计思路

验证删除 Session Key 接口的基本功能：删除已存在的 Session Key，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_session_del", password="password@123", is_admin=false
- 预先获取 Session Key：通过创建 Session Key 接口获取

#### 执行步骤

1. 先创建用户并获取 Session Key
2. 发送 DELETE 请求到 `/open-api/v1/auth/session-keys/{session_key}`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| session_key | 实际获取的 session_key 值 |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-10-002：删除不存在的 key（异常参数）

#### 设计思路

验证删除不存在的 Session Key 时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保 Session Key "non_existent_session_key" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/session-keys/non_existent_session_key`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| session_key | non_existent_session_key |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 Session Key 不存在的错误信息  
**Data**：null

---

# 删除Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除Token |
| 方法 | DELETE |
| 路径 | /open-api/v1/auth/tokens/{token_name} |
| 说明 | 删除指定的Token |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | 是 | 待删除的Token名称 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-12-001 | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-12-002 | 删除不存在的 Token | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-12-001：正常删除（正常参数）

#### 设计思路

验证删除 Token 接口的基本功能：删除已存在的 Token，确认接口返回成功。

#### 前提数据准备

- 确保产品线 "product_token_del" 存在
- 预先创建 Token：name="test_token_del", scope="Product", product_name="product_token_del"

#### 执行步骤

1. 先创建 Token
2. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/test_token_del`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | test_token_del |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-12-002：删除不存在的 Token（异常参数）

#### 设计思路

验证删除不存在的 Token 时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保 Token "non_existent_token" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/non_existent_token`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | non_existent_token |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 Token 不存在的错误信息  
**Data**：null

---

# 删除用户 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除用户 |
| 方法 | DELETE |
| 路径 | /open-api/v1/auth/users/{user_name} |
| 说明 | 删除指定用户名的用户 |

---

## 二、接口参数说明

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 待删除的用户名（URL参数） |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-2-001 | 正常删除用户 | 正常参数 | 验证 ErrNum=200 |
| AUTH-2-002 | 删除不存在的用户 | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-2-001：正常删除用户（正常参数）

#### 设计思路

验证删除用户接口的基本功能：删除已存在的用户，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_del", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户 test_user_del
2. 发送 DELETE 请求到 `/open-api/v1/auth/users/test_user_del`
3. 验证响应状态码和返回结构

#### 请求参数

| 参数名 | 值 |
|--------|-----|
| user_name (URL) | test_user_del |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-2-002：删除不存在的用户（异常参数）

#### 设计思路

验证删除不存在的用户时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保用户 "non_existent_user" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/users/non_existent_user`
2. 验证返回错误码

#### 请求参数

| 参数名 | 值 |
|--------|-----|
| user_name (URL) | non_existent_user |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含用户不存在的错误信息  
**Data**：null

---

# 用户列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 用户列表 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/users |
| 说明 | 获取所有用户列表 |

---

## 二、接口参数说明

### 请求参数

无

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-4-001 | 获取用户列表 | 正常参数 | 验证返回数组，包含 user_name、is_admin |
| AUTH-4-002 | 创建后验证列表包含新用户 | 返回数据 | 验证列表包含新创建的用户 |

---

## 四、测试场景详细设计

---

### AUTH-4-001：获取用户列表（正常参数）

#### 设计思路

验证获取用户列表接口的基本功能：确认接口返回用户数组，每个用户包含 user_name 和 is_admin 字段。

#### 前提数据准备

- 预先创建至少一个用户：user_name="test_user_list", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 GET 请求到 `/open-api/v1/auth/users`
3. 验证响应状态码和返回结构

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空数组 | IsArray, NotEmpty |
| Data[].user_name | 非空字符串 | NotEmpty |
| Data[].is_admin | bool 类型 | IsBool |

---

### AUTH-4-002：创建后验证列表包含新用户（返回数据）

#### 设计思路

验证创建用户后，用户列表接口能正确返回新创建的用户信息。

#### 前提数据准备

- 预先创建用户：user_name="test_user_check", password="password@123", is_admin=true

#### 执行步骤

1. 先创建用户 test_user_check
2. 发送 GET 请求到 `/open-api/v1/auth/users`
3. 验证返回列表中包含 test_user_check，且 is_admin=true

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 包含 test_user_check | Contains user_name="test_user_check" |
| Data[].is_admin（test_user_check） | true | Equals |

---

# 重置密码 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 重置密码 |
| 方法 | PATCH |
| 路径 | /open-api/v1/auth/users/{user_name}/passwd |
| 说明 | 修改用户密码，管理员重置他人密码无需旧密码，用户自己重置需提供旧密码 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 待修改密码的用户名 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| old_password | string | 否 | 旧密码，用户自己修改时必填 |
| password | string | 是 | 新密码 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-3-001 | 管理员重置他人密码 | 正常参数 | 无需 old_password |
| AUTH-3-002 | 用户自己重置密码 | 正常参数 | 需提供 old_password |
| AUTH-3-003 | 缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-3-004 | 修改不存在的用户密码 | 异常参数 | 验证 ErrNum=404 |
| AUTH-3-005 | old_password 错误 | 异常参数 | 验证 ErrNum=422 |

---

## 四、测试场景详细设计

---

### AUTH-3-001：管理员重置他人密码（正常参数）

#### 设计思路

验证管理员重置普通用户密码的场景，无需提供旧密码，确认接口返回成功。

#### 前提数据准备

- 预先创建管理员用户：user_name="test_admin_reset", password="password@123", is_admin=true
- 预先创建普通用户：user_name="test_user_reset", password="password@123", is_admin=false

#### 执行步骤

1. 先创建管理员和普通用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_reset/passwd`，仅提供新密码
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_reset |

**Body 参数**：

```json
{
    "password": "newpassword@456"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-3-002：用户自己重置密码（正常参数）

#### 设计思路

验证用户自己修改密码的场景，需要提供旧密码，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_self", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_self/passwd`，提供旧密码和新密码
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_self |

**Body 参数**：

```json
{
    "old_password": "password@123",
    "password": "newpassword@456"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-3-003：缺少 password（必填校验）

#### 设计思路

验证 `password` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_nopass", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_nopass/passwd`，缺少 password 字段
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_nopass |

**Body 参数**：

```json
{
    "old_password": "password@123"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "password" 的错误信息  
**Data**：null

---

### AUTH-3-004：修改不存在的用户密码（异常参数）

#### 设计思路

验证修改不存在用户的密码时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保用户 "non_existent_pass" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/auth/users/non_existent_pass/passwd`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | non_existent_pass |

**Body 参数**：

```json
{
    "password": "newpassword@456"
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含用户不存在的错误信息  
**Data**：null

---

### AUTH-3-005：old_password 错误（异常参数）

#### 设计思路

验证用户自己修改密码时，旧密码错误的处理，确认接口返回参数校验错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_wrongpass", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_wrongpass/passwd`，提供错误的旧密码
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_wrongpass |

**Body 参数**：

```json
{
    "old_password": "wrongpassword",
    "password": "newpassword@456"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含旧密码错误的错误信息  
**Data**：null

---

# 按产品线查用户 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 按产品线查用户 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/users/actions/search-by-product/{product_name} |
| 说明 | 获取对指定产品线有权限的用户列表 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| product_name | string | 是 | 产品线名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-8-001 | 查询有绑定用户的产品线 | 正常参数 | 验证返回列表包含该用户 |
| AUTH-8-002 | 查询无绑定用户的产品线 | 正常参数 | 验证返回空列表 |

---

## 四、测试场景详细设计

---

### AUTH-8-001：查询有绑定用户的产品线（正常参数）

#### 设计思路

验证按产品线查询用户列表接口的基本功能：查询有绑定用户的产品线，确认返回正确的用户列表。

#### 前提数据准备

- 预先创建用户：user_name="test_user_search", password="password@123", is_admin=false
- 预先为用户绑定产品线 "product_search"

#### 执行步骤

1. 先创建用户并绑定产品线
2. 发送 GET 请求到 `/open-api/v1/auth/users/actions/search-by-product/product_search`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_search |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空数组 | IsArray, NotEmpty |
| Data[].user_name | test_user_search | Contains |
| Data[].is_admin | false | Equals |

---

### AUTH-8-002：查询无绑定用户的产品线（正常参数）

#### 设计思路

验证查询无绑定用户的产品线时，接口应返回空数组。

#### 前提数据准备

- 确保产品线 "product_empty" 没有绑定任何用户

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/users/actions/search-by-product/product_empty`
2. 验证返回空数组

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_empty |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 空数组 | IsArray, Len=0 |

---

# 按产品线查Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 按产品线查Token |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens/actions/search-by-product/{product_name} |
| 说明 | 获取对指定产品线有权限的Token列表 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| product_name | string | 是 | 产品线名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | Token名字 |
| token | string | Token的值 |
| scope | string | 权限范围 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-15-001 | 查询有 Token 绑定的产品线 | 正常参数 | 返回 Token 列表 |
| AUTH-15-002 | 查询无 Token 绑定的产品线 | 正常参数 | 返回空列表 |

---

## 四、测试场景详细设计

---

### AUTH-15-001：查询有 Token 绑定的产品线（正常参数）

#### 设计思路

验证按产品线查询 Token 列表接口的基本功能：查询有绑定 Token 的产品线，确认返回正确的 Token 列表。

#### 前提数据准备

- 确保产品线 "product_token_search" 存在
- 预先创建 Token：name="test_token_search", scope="Product", product_name="product_token_search"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens/actions/search-by-product/product_token_search`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_token_search |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空数组 | IsArray, NotEmpty |
| Data[].name | test_token_search | Contains |
| Data[].scope | "Product" | Equals |

---

### AUTH-15-002：查询无 Token 绑定的产品线（正常参数）

#### 设计思路

验证查询无绑定 Token 的产品线时，接口应返回空数组。

#### 前提数据准备

- 确保产品线 "product_token_empty" 没有绑定任何 Token

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens/actions/search-by-product/product_token_empty`
2. 验证返回空数组

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_token_empty |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 空数组 | IsArray, Len=0 |

---

# 设置管理员 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 设置管理员 |
| 方法 | PATCH |
| 路径 | /open-api/v1/auth/users/{user_name}/is_admin |
| 说明 | 设置用户是否具有管理员权限 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 待修改权限的用户名 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| is_admin | bool | 是 | 是否为系统管理员 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-5-001 | 设置为管理员 | 正常参数 | is_admin=true |
| AUTH-5-002 | 取消管理员权限 | 正常参数 | is_admin=false |
| AUTH-5-003 | 缺少 is_admin | 必填校验 | 验证 ErrNum=422 |
| AUTH-5-004 | 修改不存在的用户 | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-5-001：设置为管理员（正常参数）

#### 设计思路

验证将普通用户设置为管理员的场景，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_promote", password="password@123", is_admin=false

#### 执行步骤

1. 先创建普通用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_promote/is_admin`，设置 is_admin=true
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_promote |

**Body 参数**：

```json
{
    "is_admin": true
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-5-002：取消管理员权限（正常参数）

#### 设计思路

验证将管理员用户取消管理员权限的场景，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_admin_demote", password="password@123", is_admin=true

#### 执行步骤

1. 先创建管理员用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_admin_demote/is_admin`，设置 is_admin=false
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_admin_demote |

**Body 参数**：

```json
{
    "is_admin": false
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-5-003：缺少 is_admin（必填校验）

#### 设计思路

验证 `is_admin` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_noadmin", password="password@123", is_admin=false

#### 执行步骤

1. 先创建用户
2. 发送 PATCH 请求到 `/open-api/v1/auth/users/test_user_noadmin/is_admin`，缺少 is_admin 字段
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_noadmin |

**Body 参数**：

```json
{}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "is_admin" 的错误信息  
**Data**：null

---

### AUTH-5-004：修改不存在的用户（异常参数）

#### 设计思路

验证设置不存在用户的管理员权限时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保用户 "non_existent_admin" 不存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/auth/users/non_existent_admin/is_admin`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | non_existent_admin |

**Body 参数**：

```json
{
    "is_admin": true
}
```

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含用户不存在的错误信息  
**Data**：null

---

# Token详情 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token详情 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens/{token_name} |
| 说明 | 查询指定Token的详细信息 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | 是 | Token名称 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | Token名字 |
| product_name | string | 产品线名 |
| token | string | Token的值 |
| scope | string | 权限范围 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-13-001 | 查询已存在的 Token | 正常参数 | 返回 name、token、scope、product_name |
| AUTH-13-002 | 查询不存在的 Token | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-13-001：查询已存在的 Token（正常参数）

#### 设计思路

验证查询 Token 详情接口的基本功能：查询已存在的 Token，确认返回完整的字段信息。

#### 前提数据准备

- 确保产品线 "product_token_detail" 存在
- 预先创建 Token：name="test_token_detail", scope="Product", product_name="product_token_detail"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens/test_token_detail`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | test_token_detail |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_token_detail" | Equals |
| product_name | "product_token_detail" | Equals |
| token | 非空字符串 | NotEmpty |
| scope | "Product" | Equals |

---

### AUTH-13-002：查询不存在的 Token（异常参数）

#### 设计思路

验证查询不存在的 Token 时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保 Token "non_existent_token_detail" 不存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens/non_existent_token_detail`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | non_existent_token_detail |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 Token 不存在的错误信息  
**Data**：null

---

# Token列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token列表 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens |
| 说明 | 获取所有Token列表 |

---

## 二、接口参数说明

### 请求参数

无

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | Token名字 |
| product_name | string | 产品线名 |
| token | string | Token的值 |
| scope | string | 权限范围 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-14-001 | 获取 Token 列表 | 正常参数 | 返回数组 |
| AUTH-14-002 | 验证返回字段完整性 | 返回数据 | 验证每个元素包含 name、token、scope |

---

## 四、测试场景详细设计

---

### AUTH-14-001：获取 Token 列表（正常参数）

#### 设计思路

验证获取 Token 列表接口的基本功能：确认接口返回 Token 数组。

#### 前提数据准备

- 确保产品线 "product_token_list" 存在
- 预先创建 Token：name="test_token_list", scope="Product", product_name="product_token_list"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens`
3. 验证响应状态码和返回结构

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空数组 | IsArray, NotEmpty |

---

### AUTH-14-002：验证返回字段完整性（返回数据）

#### 设计思路

验证 Token 列表返回数据的字段完整性，确认每个元素包含 name、token、scope 字段。

#### 前提数据准备

- 确保产品线 "product_token_list2" 存在
- 预先创建 Token：name="test_token_list2", scope="Product", product_name="product_token_list2"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens`
3. 验证返回列表中每个元素包含完整字段

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data[].name | 非空字符串 | NotEmpty |
| Data[].token | 非空字符串 | NotEmpty |
| Data[].scope | 非空字符串 | NotEmpty |

---

# 解除产品线绑定 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 解除产品线绑定 |
| 方法 | DELETE |
| 路径 | /open-api/v1/auth/users/{user_name}/products/{product_name} |
| 说明 | 对用户取消某个产品线的授权 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | 是 | 用户名 |
| product_name | string | 是 | 产品线名 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-7-001 | 正常解除绑定 | 正常参数 | 验证 ErrNum=200 |
| AUTH-7-002 | 解除未绑定的产品线 | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-7-001：正常解除绑定（正常参数）

#### 设计思路

验证解除用户产品线绑定接口的基本功能：解除已绑定的产品线，确认接口返回成功。

#### 前提数据准备

- 预先创建用户：user_name="test_user_unbind", password="password@123", is_admin=false
- 预先为用户绑定产品线 "product_demo"

#### 执行步骤

1. 先创建用户并绑定产品线
2. 发送 DELETE 请求到 `/open-api/v1/auth/users/test_user_unbind/products/product_demo`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_unbind |
| product_name | product_demo |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-7-002：解除未绑定的产品线（异常参数）

#### 设计思路

验证解除未绑定的产品线时，接口应返回资源不存在错误。

#### 前提数据准备

- 预先创建用户：user_name="test_user_unbind_none", password="password@123", is_admin=false
- 确保用户未绑定产品线 "product_unbind"

#### 执行步骤

1. 先创建用户（不绑定产品线）
2. 发送 DELETE 请求到 `/open-api/v1/auth/users/test_user_unbind_none/products/product_unbind`
3. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| user_name | test_user_unbind_none |
| product_name | product_unbind |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含未绑定的错误信息  
**Data**：null

---

