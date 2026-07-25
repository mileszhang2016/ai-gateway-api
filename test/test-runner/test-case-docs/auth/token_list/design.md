# Token列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Auth |
| 接口名称 | Token列表 |
| 方法 | GET |
| 路径 | /open-api/v1/auth/tokens |
| 说明 | 获取所有Token列表 |

---

## 二、接口参数说明

### 请求参数

无

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
| AUTH-14-001 | 获取 Token 列表 | 正常参数 | 返回数组 |
| AUTH-14-002 | 验证返回字段完整性 | 返回数据 | 验证每个元素包含 name、token、scope |

---

## 四、测试场景详细设计

---

### AUTH-14-001：获取 Token 列表（正常参数）

#### 设计思路

验证获取 Token 列表接口的基本功能：确认接口返回 Token 数组。

#### 前提数据准备

- 确保产品线 "product_token_list" 存在
- 预先创建 Token：name="test_token_list", scope="Product", product_name="product_token_list"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens`
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

---

### AUTH-14-002：验证返回字段完整性（返回数据）

#### 设计思路

验证 Token 列表返回数据的字段完整性，确认每个元素包含 name、token、scope 字段。

#### 前提数据准备

- 确保产品线 "product_token_list2" 存在
- 预先创建 Token：name="test_token_list2", scope="Product", product_name="product_token_list2"

#### 执行步骤

1. 先创建 Token
2. 发送 GET 请求到 `/open-api/v1/auth/tokens`
3. 验证返回列表中每个元素包含完整字段

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data[].name | 非空字符串 | NotEmpty |
| Data[].token | 非空字符串 | NotEmpty |
| Data[].scope | 非空字符串 | NotEmpty |