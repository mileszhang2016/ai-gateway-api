# 创建Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 创建Token |
| 方法 | POST |
| 路径 | /open-api/v1/auth/tokens |
| 说明 | 创建Token（同时完成产品线绑定） |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | Token名字，必须全局唯一 |
| scope | string | 是 | 权限范围（Product/Support/System） |
| product_name | string | 是 | 产品线名（scope为Product时必填） |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | Token 值 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-11-001 | 创建 Product scope Token | 正常参数 | 返回 token |
| AUTH-11-002 | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-003 | 缺少 scope | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-004 | scope=Product 缺少 product_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-005 | 重复创建同名 Token | 业务规则 | 验证 ErrNum=555 |

---

## 四、测试场景详细设计

---

### AUTH-11-001：创建 Product scope Token（正常参数）

#### 设计思路

验证创建 Token 接口的基本功能：传入完整参数，创建 Product scope 的 Token，确认接口返回成功并返回 token 值。

#### 前提数据准备

- 确保产品线 "product_token" 存在

#### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/auth/tokens`，传入完整参数
2. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "name": "test_token_001",
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| token | 非空字符串 | NotEmpty |

---

### AUTH-11-002：缺少 name（必填校验）

#### 设计思路

验证 `name` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "name" 的错误信息  
**Data**：null

---

### AUTH-11-003：缺少 scope（必填校验）

#### 设计思路

验证 `scope` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：缺少 scope 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_002",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "scope" 的错误信息  
**Data**：null

---

### AUTH-11-004：scope=Product 缺少 product_name（必填校验）

#### 设计思路

验证当 scope=Product 时，`product_name` 为必填字段，不传该字段时接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body：scope=Product 但缺少 product_name 字段
2. 发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_003",
    "scope": "Product"
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "product_name" 的错误信息  
**Data**：null

---

### AUTH-11-005：重复创建同名 Token（业务规则）

#### 设计思路

验证 Token name 必须全局唯一，当尝试创建已存在的 Token name 时，接口应返回业务错误。

#### 前提数据准备

- 预先创建 Token：name="test_token_dup", scope="Product", product_name="product_token"

#### 执行步骤

1. 先创建 Token
2. 使用相同 name 再次发送 POST 请求到 `/open-api/v1/auth/tokens`
3. 验证返回错误码

#### 请求参数

```json
{
    "name": "test_token_dup",
    "scope": "Product",
    "product_name": "product_token"
}
```

#### 预期返回结果

**ErrNum**：555  
**ErrMsg**：包含重复 Token name 的错误信息  
**Data**：null