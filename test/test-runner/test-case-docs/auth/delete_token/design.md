# 删除Token - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | 删除Token |
| 方法 | DELETE |
| 路径 | /open-api/v1/auth/tokens/{token_name} |
| 说明 | 删除指定的Token |

---

## 二、接口参数说明

### 请求参数

#### URL 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token_name | string | 是 | 待删除的Token名称 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| - | - | 无数据返回 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AUTH-12-001 | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-12-002 | 删除不存在的 Token | 异常参数 | 验证 ErrNum=404 |

---

## 四、测试场景详细设计

---

### AUTH-12-001：正常删除（正常参数）

#### 设计思路

验证删除 Token 接口的基本功能：删除已存在的 Token，确认接口返回成功。

#### 前提数据准备

- 确保产品线 "product_token_del" 存在
- 预先创建 Token：name="test_token_del", scope="Product", product_name="product_token_del"

#### 执行步骤

1. 先创建 Token
2. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/test_token_del`
3. 验证响应状态码和返回结构

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | test_token_del |

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：null

---

### AUTH-12-002：删除不存在的 Token（异常参数）

#### 设计思路

验证删除不存在的 Token 时，接口应返回资源不存在错误。

#### 前提数据准备

- 确保 Token "non_existent_token" 不存在

#### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/auth/tokens/non_existent_token`
2. 验证返回错误码

#### 请求参数

**URL 参数**：

| 参数名 | 值 |
|--------|-----|
| token_name | non_existent_token |

#### 预期返回结果

**ErrNum**：404  
**ErrMsg**：包含 Token 不存在的错误信息  
**Data**：null