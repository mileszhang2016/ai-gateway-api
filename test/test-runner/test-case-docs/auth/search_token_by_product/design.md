# 按产品线查Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 按产品线查Token |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens/actions/search-by-product/{product_name} |
| 说明 | 获取对指定产品线有权限的Token列表 |

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
| name | string | Token名字 |
| token | string | Token的值 |
| scope | string | 权限范围 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-15-001 | 查询有 Token 绑定的产品线 | 正常参数 | 返回 Token 列表 |
| AUTH-15-002 | 查询无 Token 绑定的产品线 | 正常参数 | 返回空列表 |

---

## 四、测试场景详细设计

---

### AUTH-15-001：查询有 Token 绑定的产品线（正常参数）

#### 设计思路

验证按产品线查询 Token 列表接口的基本功能：查询有绑定 Token 的产品线，确认返回正确的 Token 列表。

#### 前提数据准备

- 确保产品线 "product_token_search" 存在
- 预先创建 Token：name="test_token_search", scope="Product", product_name="product_token_search"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens/actions/search-by-product/product_token_search`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_token_search |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空数组 | IsArray, NotEmpty |
| Data[].name | test_token_search | Contains |
| Data[].scope | "Product" | Equals |

---

### AUTH-15-002：查询无 Token 绑定的产品线（正常参数）

#### 设计思路

验证查询无绑定 Token 的产品线时，接口应返回空数组。

#### 前提数据准备

- 确保产品线 "product_token_empty" 没有绑定任何 Token

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens/actions/search-by-product/product_token_empty`
2. 验证返回空数组

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| product_name | product_token_empty |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 空数组 | IsArray, Len=0 |