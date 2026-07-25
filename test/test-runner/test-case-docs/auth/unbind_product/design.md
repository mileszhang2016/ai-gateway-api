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