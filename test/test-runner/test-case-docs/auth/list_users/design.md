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