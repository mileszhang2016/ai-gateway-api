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