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