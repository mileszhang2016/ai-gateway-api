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