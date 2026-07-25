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