# Token详情 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token详情 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens/{token_name} |
| 说明 | 查询指定Token的详细信息 |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | 是 | Token名称 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | Token名字 |
| product_name | string | 产品线名 |
| token | string | Token的值 |
| scope | string | 权限范围 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-13-001 | 查询已存在的 Token | 正常参数 | 返回 name、token、scope、product_name |
| AUTH-13-002 | 查询不存在的 Token | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-13-001：查询已存在的 Token（正常参数）

#### 设计思路

验证查询 Token 详情接口的基本功能：查询已存在的 Token，确认返回完整的字段信息。

#### 前提数据准备

- 确保产品线 "product_token_detail" 存在
- 预先创建 Token：name="test_token_detail", scope="Product", product_name="product_token_detail"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens/test_token_detail`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | test_token_detail |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "test_token_detail" | Equals |
| product_name | "product_token_detail" | Equals |
| token | 非空字符串 | NotEmpty |
| scope | "Product" | Equals |

---

### AUTH-13-002：查询不存在的 Token（异常参数）

#### 设计思路

验证查询不存在的 Token 时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保 Token "non_existent_token_detail" 不存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/auth/tokens/non_existent_token_detail`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | non_existent_token_detail |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 Token 不存在的错误信息  
**Data**：null