# Expression Verify 测试用例设计文档

## 1. 模块概述

Expression Verify 模块用于校验 BFE 路由表达式是否合法。该接口已废弃，无需鉴权。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| EV-1 | 校验路由表达式 | PATCH | `/open-api/v1/expression/verify` | 校验给定 BFE 表达式合法性 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 校验路由表达式 | 11 |
| **合计** | **11** |

## 4. 认证方式

该接口已废弃，无需鉴权；测试环境 `SkipTokenValidate=true`，无需携带认证头。

## 5. 目录结构

```
expression_verify/
├── design.md
└── verify/
    └── verify_test.go
```

## 6. 校验路由表达式

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Expression Verify |
| 接口名称 | 校验路由表达式 |
| 方法 | PATCH |
| 路径 | `/open-api/v1/expression/verify` |
| 说明 | 校验 BFE 路由表达式合法性，成功返回 null，失败返回 VerifyResult |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| expression | string | Y | 待校验的 BFE 表达式 | 必填、非空；须为合法 BFE 条件表达式 |

#### 6.2.2 返回数据字段

- 校验成功：Data 为 `null`。
- 校验失败：Data 为 VerifyResult 对象。

| 参数名 | 类型 | 说明 |
|--------|------|------|
| code | int | 错误码，固定 500 |
| message | string | 错误信息 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| EV-1-001 | 校验 default_t 表达式 | 正常参数 | 合法表达式返回 Data=null |
| EV-1-002 | 校验 req_path_prefix 表达式 | 正常参数 | 常用路径前缀表达式 |
| EV-1-003 | 校验组合表达式 | 正常参数 | 组合表达式合法 |
| EV-1-004 | 缺少 expression 参数 | 必填校验 | 验证 ErrNum=422 |
| EV-1-005 | expression 为空字符串 | 异常参数 | 验证 ErrNum=500 |
| EV-1-006 | 表达式括号不匹配 | 异常参数 | 语法错误 |
| EV-1-007 | 表达式包含未知函数 | 异常参数 | 非法函数名 |
| EV-1-008 | 表达式缺少引号 | 异常参数 | 字符串参数未加引号 |
| EV-1-009 | 校验 req_body_larger_than 表达式 | 正常参数 | 合法表达式返回 Data=null |
| EV-1-010 | 校验 req_body_less_than 表达式 | 正常参数 | 合法表达式返回 Data=null |
| EV-1-011 | req_body_larger_than 参数非法 | 异常参数 | INT 参数传 STRING |

### 6.4 测试场景详细设计

#### 6.4.1 EV-1-001：校验 default_t 表达式（正常参数）

##### 设计思路

验证最基础的合法表达式 `default_t()` 校验成功。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/expression/verify`。
2. 验证响应状态码和返回结构。
3. 验证 `Data` 为 `null`。

##### 请求参数

```json
{
    "expression": "default_t()"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.2 EV-1-002：校验 req_path_prefix 表达式（正常参数）

##### 设计思路

验证常用的路径前缀表达式合法。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入 `req_path_prefix` 表达式。
2. 验证返回 `Data` 为 `null`。

##### 请求参数

```json
{
    "expression": "req_path_prefix(\"/open-api/v1\")"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.3 EV-1-003：校验组合表达式（正常参数）

##### 设计思路

验证组合表达式（如 `and`）语法合法。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入组合表达式。
2. 验证返回 `Data` 为 `null`。

##### 请求参数

```json
{
    "expression": "and(req_method_in(\"POST\"), req_path_prefix(\"/v1\"))"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.4 EV-1-004：缺少 expression 参数（必填校验）

##### 设计思路

验证 `expression` 为必填字段，缺少时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，Body 为空对象。
2. 验证返回错误码。

##### 请求参数

```json
{}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "expression" 的错误信息  
**Data**：null

---

#### 6.4.5 EV-1-005：expression 为空字符串（异常参数）

##### 设计思路

验证空字符串表达式被视为非法，返回 VerifyResult。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，`expression` 为空字符串。
2. 验证返回错误码和 VerifyResult 结构。

##### 请求参数

```json
{
    "expression": ""
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：错误信息非空

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| code | 500 | Equals |
| message | 非空字符串 | NotEmpty |

---

#### 6.4.6 EV-1-006：表达式括号不匹配（异常参数）

##### 设计思路

验证表达式语法错误（括号不匹配）时返回校验失败。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入括号不匹配的表达式。
2. 验证返回错误码和 VerifyResult 结构。

##### 请求参数

```json
{
    "expression": "default_t("
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：错误信息非空

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| code | 500 | Equals |
| message | 非空字符串 | NotEmpty |

---

#### 6.4.7 EV-1-007：表达式包含未知函数（异常参数）

##### 设计思路

验证表达式包含未知函数名时返回校验失败。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入包含未知函数的表达式。
2. 验证返回错误码和 VerifyResult 结构。

##### 请求参数

```json
{
    "expression": "unknown_func()"
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：错误信息非空

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| code | 500 | Equals |
| message | 非空字符串 | NotEmpty |

---

#### 6.4.8 EV-1-008：表达式缺少引号（异常参数）

##### 设计思路

验证字符串参数未加引号时返回校验失败。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入字符串参数未加引号的表达式。
2. 验证返回错误码和 VerifyResult 结构。

##### 请求参数

```json
{
    "expression": "req_path_prefix(/v1)"
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：错误信息非空

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| code | 500 | Equals |
| message | 非空字符串 | NotEmpty |

---

#### 6.4.9 EV-1-009：校验 req_body_larger_than 表达式（正常参数）

##### 设计思路

验证基于请求体大小的 `req_body_larger_than` 表达式合法。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入 `req_body_larger_than(8192)`。
2. 验证返回 `Data` 为 `null`。

##### 请求参数

```json
{
    "expression": "req_body_larger_than(8192)"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.10 EV-1-010：校验 req_body_less_than 表达式（正常参数）

##### 设计思路

验证基于请求体大小的 `req_body_less_than` 表达式合法。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入 `req_body_less_than(2048)`。
2. 验证返回 `Data` 为 `null`。

##### 请求参数

```json
{
    "expression": "req_body_less_than(2048)"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 6.4.11 EV-1-011：req_body_larger_than 参数非法（异常参数）

##### 设计思路

验证 `req_body_larger_than` 参数类型错误时返回校验失败。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求，传入 `req_body_larger_than("abc")`。
2. 验证返回错误码和 VerifyResult 结构。

##### 请求参数

```json
{
    "expression": "req_body_larger_than(\"abc\")"
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：错误信息非空

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| code | 500 | Equals |
| message | 非空字符串 | NotEmpty |

---

## 7. 依赖与数据准备

- 无需前置数据准备。
- 调用时不需携带认证 Token。

## 8. 注意事项

1. 该接口标记为已废弃，但在 v0.3.0 中仍保留，需保证其行为稳定。
2. 成功时 `Data=null`，失败时 `Data` 为对象 `{ code, message }`。
3. 失败时 HTTP 状态码与 ErrNum 可能仍为 200（以实际实现为准），但文档约定 ErrNum=500 表示校验失败。
