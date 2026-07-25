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