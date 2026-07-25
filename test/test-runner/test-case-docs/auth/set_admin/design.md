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