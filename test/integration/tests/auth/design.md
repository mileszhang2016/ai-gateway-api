# Auth 测试用例设计文档

## 1. 模块概述

Auth 模块负责管理用户（User）、Session Key、Token 及系统导航配置。v0.3.0 起用户 `is_admin` 仅支持 `true`，Token `scope` 仅支持 `System`/`Support`，并删除 Product Scope 及产品线的绑定/查询接口。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| AUTH-1 | 创建用户 | POST | `/open-api/v1/auth/users` | `password` 必填，`is_admin` 仅支持 true |
| AUTH-2 | 删除用户 | DELETE | `/open-api/v1/auth/users/{user_name}` | - |
| AUTH-3 | 重置密码 | PATCH | `/open-api/v1/auth/users/{user_name}/passwd` | 管理员重置无需旧密码 |
| AUTH-4 | 用户列表 | GET | `/open-api/v1/auth/users` | 返回 user_name、is_admin |
| AUTH-5 | 设置管理员 | PATCH | `/open-api/v1/auth/users/{user_name}/is_admin` | `is_admin` 固定为 true |
| AUTH-6 | 查询单个用户 | GET | `/open-api/v1/auth/users/{user_name}` | - |
| AUTH-7 | 创建 Session Key | POST | `/open-api/v1/auth/session-keys` | 用户名密码登录 |
| AUTH-8 | 删除 Session Key | DELETE | `/open-api/v1/auth/session-keys/{session_key}` | - |
| AUTH-9 | 创建 Token | POST | `/open-api/v1/auth/tokens` | `scope` 仅 System/Support |
| AUTH-10 | 删除 Token | DELETE | `/open-api/v1/auth/tokens/{token_name}` | - |
| AUTH-11 | Token 详情 | GET | `/open-api/v1/auth/tokens/{token_name}` | 返回 name/token/scope |
| AUTH-12 | Token 列表 | GET | `/open-api/v1/auth/tokens` | 数组，元素同详情 |
| AUTH-13 | 获取系统导航配置 | GET | `/open-api/v1/meta` | 无需鉴权 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建用户 | 5 |
| 删除用户 | 2 |
| 重置密码 | 3 |
| 用户列表 | 1 |
| 设置管理员 | 1 |
| 查询单个用户 | 2 |
| 创建 Session Key | 3 |
| 删除 Session Key | 2 |
| 创建 Token | 5 |
| 删除 Token | 2 |
| Token 详情 | 1 |
| Token 列表 | 1 |
| 获取系统导航配置 | 1 |
| **合计** | **29** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，OpenAPI 请求无需携带认证头；Session Key 用例验证鉴权中间件行为。

## 5. 目录结构

```
auth/
├── design.md
├── create_user/
│   └── create_user_test.go
├── delete_user/
│   └── delete_user_test.go
├── reset_password/
│   └── reset_password_test.go
├── list_users/
│   └── list_users_test.go
├── set_admin/
│   └── set_admin_test.go
├── user_detail/
│   └── user_detail_test.go
├── create_session_key/
│   └── create_session_key_test.go
├── delete_session_key/
│   └── delete_session_key_test.go
├── create_token/
│   └── create_token_test.go
├── delete_token/
│   └── delete_token_test.go
├── token_detail/
│   └── token_detail_test.go
├── token_list/
│   └── token_list_test.go
└── meta/
    └── meta_test.go
```

## 6. 创建用户

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建用户 |
| 方法 | POST |
| 路径 | `/open-api/v1/auth/users` |
| 说明 | 创建用户，`is_admin` 固定为 true |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 用户名 |
| password | string | Y | 用户密码 |
| is_admin | bool | N | 固定为 true，默认填充 true |

#### 6.2.2 返回数据字段

Data 为 null。

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-1-001 | 创建用户（完整参数） | 正常参数 | is_admin=true |
| AUTH-1-002 | 创建用户（省略 is_admin） | 正常参数 | 默认填充 true |
| AUTH-1-003 | 创建用户缺少 user_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-004 | 创建用户缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-005 | 重复创建用户 | 业务规则 | 验证 ErrNum=555/556 |

### 6.4 测试场景详细设计

#### 6.4.1 AUTH-1-001：创建用户（完整参数）

##### 设计思路

验证创建用户接口基本功能，`is_admin` 为 true 时创建成功。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/users`。
2. 验证返回成功，后续查询 `is_admin=true`。

##### 请求参数

```json
{
    "user_name": "test_user_001",
    "password": "password@123",
    "is_admin": true
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.2 AUTH-1-002：创建用户（省略 is_admin）

##### 设计思路

验证省略 `is_admin` 时默认填充为 true。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中不包含 `is_admin`。
2. 验证返回成功，后续查询 `is_admin=true`。

##### 请求参数

```json
{
    "user_name": "test_user_002",
    "password": "password@123"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.3 AUTH-1-003：创建用户缺少 user_name（必填校验）

##### 设计思路

验证 `user_name` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `user_name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "password": "password@123"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "user_name" 的错误信息  
**Data**：null

---

#### 6.4.4 AUTH-1-004：创建用户缺少 password（必填校验）

##### 设计思路

验证 `password` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `password`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "user_name": "test_user_003"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "password" 的错误信息  
**Data**：null

---

#### 6.4.5 AUTH-1-005：重复创建用户（业务规则）

##### 设计思路

验证用户名全局唯一，重复创建时返回错误。

##### 前提数据准备

已创建 `test_user_dup`。

##### 执行步骤

1. 发送 POST 请求，使用重复的用户名。
2. 验证返回错误码。

##### 请求参数

```json
{
    "user_name": "test_user_dup",
    "password": "password@123"
}
```

##### 预期返回结果

**ErrNum**：555 或 556  
**ErrMsg**：用户名已存在的错误信息  
**Data**：null

---

## 7. 删除用户

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除用户 |
| 方法 | DELETE |
| 路径 | `/open-api/v1/auth/users/{user_name}` |
| 说明 | 删除用户记录 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 待删除的用户名 |

#### 7.2.2 返回数据字段

Data 为 null。

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-2-001 | 删除用户 | 正常参数 | 删除成功，再次查询返回 404 |
| AUTH-2-002 | 删除不存在的用户 | 异常参数 | 验证 ErrNum=404 |

### 7.4 测试场景详细设计

#### 7.4.1 AUTH-2-001：删除用户（正常参数）

##### 设计思路

验证删除用户成功，并级联清理相关记录。

##### 前提数据准备

已创建用户 `test_user_del`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/users/test_user_del`。
2. 验证返回成功。
3. 再次查询该用户，验证返回 404。

##### 请求参数

URI：`test_user_del`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 7.4.2 AUTH-2-002：删除不存在的用户（异常参数）

##### 设计思路

验证删除不存在的用户时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/users/non_existent_user`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_user`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：用户不存在的错误信息  
**Data**：null

---

## 8. 重置密码

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 重置密码 |
| 方法 | PATCH |
| 路径 | `/open-api/v1/auth/users/{user_name}/passwd` |
| 说明 | 重置用户密码，管理员重置无需旧密码 |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 待修改密码的用户名 |

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| old_password | string | N | 旧密码，修改当前登录用户时需要 |
| password | string | Y | 新密码 |

#### 8.2.2 返回数据字段

Data 为 null。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-3-001 | 管理员重置他人密码 | 正常参数 | 无需 old_password |
| AUTH-3-002 | 缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-3-003 | 修改不存在用户的密码 | 异常参数 | 验证 ErrNum=404 |

### 8.4 测试场景详细设计

#### 8.4.1 AUTH-3-001：管理员重置他人密码（正常参数）

##### 设计思路

验证管理员重置他人密码时无需旧密码。

##### 前提数据准备

已创建用户。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/auth/users/{user_name}/passwd`，传入新密码。
2. 验证返回成功。

##### 请求参数

```json
{
    "password": "newpassword@456"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 8.4.2 AUTH-3-002：缺少 password（必填校验）

##### 设计思路

验证 `password` 为必填字段。

##### 前提数据准备

已创建用户。

##### 执行步骤

1. 发送 PATCH 请求，Body 为空对象。
2. 验证返回错误码。

##### 请求参数

```json
{}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "password" 的错误信息  
**Data**：null

---

#### 8.4.3 AUTH-3-003：修改不存在用户的密码（异常参数）

##### 设计思路

验证修改不存在的用户密码时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/auth/users/non_existent_user/passwd`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_user`  
Body：
```json
{
    "password": "newpassword@456"
}
```

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：用户不存在的错误信息  
**Data**：null

---

## 9. 用户列表

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 用户列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/auth/users` |
| 说明 | 返回所有用户列表 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

无

#### 9.2.2 返回数据字段

Data 为数组，每个元素包含：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员，固定返回 true |

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-4-001 | 用户列表 | 正常参数 | 返回数组，元素仅含 user_name、is_admin |

### 9.4 测试场景详细设计

#### 9.4.1 AUTH-4-001：用户列表（正常参数）

##### 设计思路

验证用户列表接口返回字段完整，且不包含 `products` 等已移除字段。

##### 前提数据准备

已创建用户。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/users`。
2. 验证返回数组元素仅含 `user_name`、`is_admin`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[*].user_name | 非空字符串 | NotEmpty |
| Data[*].is_admin | true | Equals |
| Data[*].products | 不存在 | NotExists |

---

## 10. 设置管理员

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 设置管理员 |
| 方法 | PATCH |
| 路径 | `/open-api/v1/auth/users/{user_name}/is_admin` |
| 说明 | 设置用户管理员权限，`is_admin` 固定为 true |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 待修改权限的用户名 |

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| is_admin | bool | Y | 固定为 true |

#### 10.2.2 返回数据字段

Data 为 null。

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-5-001 | 设置管理员为 true | 正常参数 | v0.3.0 仅支持 true |

### 10.4 测试场景详细设计

#### 10.4.1 AUTH-5-001：设置管理员为 true（正常参数）

##### 设计思路

验证设置管理员接口成功，`is_admin` 固定为 true。

##### 前提数据准备

已创建用户。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/auth/users/{user_name}/is_admin`。
2. 验证返回成功，后续查询 `is_admin=true`。

##### 请求参数

```json
{
    "is_admin": true
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

## 11. 查询单个用户

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 查询单个用户 |
| 方法 | GET |
| 路径 | `/open-api/v1/auth/users/{user_name}` |
| 说明 | 查询单个用户详情 |

### 11.2 接口参数说明

#### 11.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 用户名 |

#### 11.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员，固定返回 true |

### 11.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-6-001 | 查询单个用户 | 正常参数 | is_admin=true |
| AUTH-6-002 | 查询不存在用户 | 异常参数 | 验证 ErrNum=404 |

### 11.4 测试场景详细设计

#### 11.4.1 AUTH-6-001：查询单个用户（正常参数）

##### 设计思路

验证查询单个用户接口返回字段完整。

##### 前提数据准备

已创建用户 `test_user_001`。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/users/test_user_001`。
2. 验证返回 `is_admin=true`。

##### 请求参数

URI：`test_user_001`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| user_name | "test_user_001" | Equals |
| is_admin | true | Equals |

---

#### 11.4.2 AUTH-6-002：查询不存在用户（异常参数）

##### 设计思路

验证查询不存在的用户时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/users/non_existent_user`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_user`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：用户不存在的错误信息  
**Data**：null

---

## 12. 创建 Session Key

### 12.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建 Session Key |
| 方法 | POST |
| 路径 | `/open-api/v1/auth/session-keys` |
| 说明 | 使用账号密码得到 session key |

### 12.2 接口参数说明

#### 12.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_name | string | Y | 用户名 |
| password | string | Y | 用户密码 |

#### 12.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| session_key | string | 会话密钥 |
| user_name | string | 用户名 |
| is_admin | bool | 是否为系统管理员，固定返回 true |

### 12.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-7-001 | 正确登录创建 Session Key | 正常参数 | 返回 session_key、is_admin=true |
| AUTH-7-002 | 密码错误 | 异常参数 | 验证 ErrNum=401 |
| AUTH-7-003 | 缺少 user_name | 必填校验 | 验证 ErrNum=422 |

### 12.4 测试场景详细设计

#### 12.4.1 AUTH-7-001：正确登录创建 Session Key（正常参数）

##### 设计思路

验证使用正确用户名密码创建 Session Key 成功。

##### 前提数据准备

已创建用户 `test_user_login`，密码为 `password@123`。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/session-keys`。
2. 验证返回 `session_key` 非空，`is_admin=true`。

##### 请求参数

```json
{
    "user_name": "test_user_login",
    "password": "password@123"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| session_key | 非空字符串 | NotEmpty |
| user_name | "test_user_login" | Equals |
| is_admin | true | Equals |

---

#### 12.4.2 AUTH-7-002：密码错误（异常参数）

##### 设计思路

验证密码错误时返回 401。

##### 前提数据准备

已创建用户 `test_user_login`。

##### 执行步骤

1. 发送 POST 请求，传入错误密码。
2. 验证返回错误码。

##### 请求参数

```json
{
    "user_name": "test_user_login",
    "password": "wrong"
}
```

##### 预期返回结果

**ErrNum**：401  
**ErrMsg**：密码错误或登录失败的错误信息  
**Data**：null

---

#### 12.4.3 AUTH-7-003：缺少 user_name（必填校验）

##### 设计思路

验证 `user_name` 为必填字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `user_name`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "password": "password@123"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "user_name" 的错误信息  
**Data**：null

---

## 13. 删除 Session Key

### 13.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除 Session Key |
| 方法 | DELETE |
| 路径 | `/open-api/v1/auth/session-keys/{session_key}` |
| 说明 | 删除 session key |

### 13.2 接口参数说明

#### 13.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_key | string | Y | 待删除的 session key |

#### 13.2.2 返回数据字段

Data 为 null。

### 13.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-8-001 | 删除 Session Key | 正常参数 | 删除成功 |
| AUTH-8-002 | 删除不存在 Session Key | 异常参数 | 验证 ErrNum=404 |

### 13.4 测试场景详细设计

#### 13.4.1 AUTH-8-001：删除 Session Key（正常参数）

##### 设计思路

验证删除 Session Key 成功。

##### 前提数据准备

已创建 Session Key。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/session-keys/{session_key}`。
2. 验证返回成功。

##### 请求参数

URI：`<session_key>`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 13.4.2 AUTH-8-002：删除不存在 Session Key（异常参数）

##### 设计思路

验证删除不存在的 Session Key 时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/session-keys/non_existent_key`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_key`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：Session Key 不存在的错误信息  
**Data**：null

---

## 14. 创建 Token

### 14.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建 Token |
| 方法 | POST |
| 路径 | `/open-api/v1/auth/tokens` |
| 说明 | 创建 Token，`scope` 仅支持 System/Support |

### 14.2 接口参数说明

#### 14.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | Y | Token 名称，全局唯一 |
| scope | string | Y | 权限范围：System / Support |

#### 14.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | Token 值 |

### 14.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-9-001 | 创建 System Token | 正常参数 | 返回非空 token |
| AUTH-9-002 | 创建 Support Token | 正常参数 | 返回非空 token |
| AUTH-9-003 | 创建 Token 缺少 name | 必填校验 | 验证 ErrNum=422 |
| AUTH-9-004 | 创建 Token 非法 scope | 异常参数 | Product 已移除 |
| AUTH-9-005 | 重复创建 Token | 业务规则 | 验证 ErrNum=555/556 |

### 14.4 测试场景详细设计

#### 14.4.1 AUTH-9-001：创建 System Token（正常参数）

##### 设计思路

验证创建 System scope Token 成功。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/tokens`。
2. 验证返回 `token` 非空。

##### 请求参数

```json
{
    "name": "token_system_001",
    "scope": "System"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| token | 非空字符串 | NotEmpty |

---

#### 14.4.2 AUTH-9-002：创建 Support Token（正常参数）

##### 设计思路

验证创建 Support scope Token 成功。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`scope=Support`。
2. 验证返回 `token` 非空。

##### 请求参数

```json
{
    "name": "token_support_001",
    "scope": "Support"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| token | 非空字符串 | NotEmpty |

---

#### 14.4.3 AUTH-9-003：创建 Token 缺少 name（必填校验）

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
    "scope": "System"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "name" 的错误信息  
**Data**：null

---

#### 14.4.4 AUTH-9-004：创建 Token 非法 scope（异常参数）

##### 设计思路

验证 `scope` 仅支持 `System`/`Support`，`Product` 已移除。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`scope=Product`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "token_bad",
    "scope": "Product"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "scope" 非法的错误信息  
**Data**：null

---

#### 14.4.5 AUTH-9-005：重复创建 Token（业务规则）

##### 设计思路

验证 Token name 全局唯一。

##### 前提数据准备

已创建同名 Token `token_dup`。

##### 执行步骤

1. 发送 POST 请求，使用重复的 name。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "token_dup",
    "scope": "System"
}
```

##### 预期返回结果

**ErrNum**：555 或 556  
**ErrMsg**：Token 名称已存在的错误信息  
**Data**：null

---

## 15. 删除 Token

### 15.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除 Token |
| 方法 | DELETE |
| 路径 | `/open-api/v1/auth/tokens/{token_name}` |
| 说明 | 删除 Token |

### 15.2 接口参数说明

#### 15.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | Y | 待删除的 token name |

#### 15.2.2 返回数据字段

Data 为 null。

### 15.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-10-001 | 删除 Token | 正常参数 | 删除成功 |
| AUTH-10-002 | 删除不存在 Token | 异常参数 | 验证 ErrNum=404 |

### 15.4 测试场景详细设计

#### 15.4.1 AUTH-10-001：删除 Token（正常参数）

##### 设计思路

验证删除 Token 成功。

##### 前提数据准备

已创建 Token `token_system_001`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/token_system_001`。
2. 验证返回成功。

##### 请求参数

URI：`token_system_001`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 15.4.2 AUTH-10-002：删除不存在 Token（异常参数）

##### 设计思路

验证删除不存在的 Token 时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/non_existent_token`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_token`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：Token 不存在的错误信息  
**Data**：null

---

## 16. Token 详情

### 16.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token 详情 |
| 方法 | GET |
| 路径 | `/open-api/v1/auth/tokens/{token_name}` |
| 说明 | 查询 Token 详情 |

### 16.2 接口参数说明

#### 16.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | Y | token name |

#### 16.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | token 名字 |
| token | string | token 的值 |
| scope | string | scope |

### 16.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-11-001 | Token 详情 | 正常参数 | 返回 name/token/scope，无 product_name |

### 16.4 测试场景详细设计

#### 16.4.1 AUTH-11-001：Token 详情（正常参数）

##### 设计思路

验证 Token 详情接口返回字段完整，且不包含 `product_name`。

##### 前提数据准备

已创建 Token `token_system_001`。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens/token_system_001`。
2. 验证返回字段。

##### 请求参数

URI：`token_system_001`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "token_system_001" | Equals |
| token | 非空字符串 | NotEmpty |
| scope | "System" | Equals |
| product_name | 不存在 | NotExists |

---

## 17. Token 列表

### 17.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token 列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/auth/tokens` |
| 说明 | 返回所有 Token 列表 |

### 17.2 接口参数说明

#### 17.2.1 请求参数

无

#### 17.2.2 返回数据字段

Data 为数组，元素同 Token 详情。

### 17.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-12-001 | Token 列表 | 正常参数 | 数组元素不含 product_name |

### 17.4 测试场景详细设计

#### 17.4.1 AUTH-12-001：Token 列表（正常参数）

##### 设计思路

验证 Token 列表接口返回字段完整，且不包含 `product_name`。

##### 前提数据准备

已创建 Token。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens`。
2. 验证返回数组元素字段。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[*].name | 非空字符串 | NotEmpty |
| Data[*].token | 非空字符串 | NotEmpty |
| Data[*].scope | System 或 Support | InEnum |
| Data[*].product_name | 不存在 | NotExists |

---

## 18. 获取系统导航配置

### 18.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 获取系统导航配置 |
| 方法 | GET |
| 路径 | `/open-api/v1/meta` |
| 说明 | 获取系统导航/图标/Logo 配置，无需鉴权 |

### 18.2 接口参数说明

#### 18.2.1 请求参数

无

#### 18.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| nav | object | 导航配置 |
| icon | object | 图标配置 |
| logo | object | Logo 配置 |

### 18.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-13-001 | 获取系统导航配置 | 正常参数 | 返回 nav/icon/logo |

### 18.4 测试场景详细设计

#### 18.4.1 AUTH-13-001：获取系统导航配置（正常参数）

##### 设计思路

验证获取系统导航配置接口基本功能，无需鉴权。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/meta`。
2. 验证返回结构。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| nav | 对象 | IsObject |
| icon | 对象 | IsObject |
| logo | 对象 | IsObject |

---

## 19. 依赖与数据准备

1. 用户、Token、Session Key 均使用全局唯一名称，测试用例间需避免冲突。
2. Session Key 用例依赖预先创建的用户记录。
3. 测试环境 `SkipTokenValidate=true`，无需构造真实 Token 即可调用管理接口。

## 20. 注意事项

1. v0.3.0 已删除以下接口，测试方案不再覆盖：
   - `POST/DELETE /auth/users/{user_name}/products/{product_name}`
   - `GET /auth/users/actions/search-by-product/{product_name}`
   - `GET /auth/tokens/actions/search-by-product/{product_name}`
2. 用户 `is_admin` 仅支持 `true`；创建用户时即使传入 `false` 也应按实现断言为 true 或 422。
3. Token `scope` 仅支持 `System`/`Support`，不再接受 `Product` 或 `product_name`。
4. 用户列表、Token 列表、Token 详情均不再返回 `products`/`product_name`。
5. 测试环境 `SkipTokenValidate=true`，无需认证头。
