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