# 全量更新AI路由规则 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | AI 路由规则 |
| 接口名称 | 全量更新AI路由规则 |
| 方法 | PATCH |
| 路径 | /open-api/v1/ai-route-rules |
| 说明 | 全量替换当前产品线的 AI 路由规则，支持高级路由规则（forward_rules）和基础路由规则（basic_forward_rules） |

---

## 二、接口参数说明

### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| forward_rules | []ForwardRule | 否 | 高级路由规则列表，为空数组代表清空规则 | [] |
| basic_forward_rules | []BasicForwardRule | 否 | 基础路由规则列表，为空数组代表清空规则 | [] |

**ForwardRule 元素**：

| 参数名 | 类型 | 必填 | 说明 | 校验规则 |
|--------|------|------|------|---------|
| name | string | 是 | 规则名称 | 以字母或数字开头，允许 [a-zA-Z0-9_-]，长度 1~128 |
| description | string | 否 | 规则描述 | - |
| expression | string | **是** | 条件表达式 | BFE 条件表达式语法，`validate:"required,min=1"` |
| cluster_name | string | **是** | 目标集群名称 | `validate:"required,min=1"` |

**BasicForwardRule 元素**：

| 参数名 | 类型 | 必填 | 说明 | 校验规则 |
|--------|------|------|------|---------|
| host_names | []string | 否 | 域名列表 | 如 `["*.example.com"]` |
| paths | []string | 否 | 路径列表 | 如 `["/api"]` |
| cluster_name | string | **是** | 目标集群名称 | `validate:"required,min=1"` |
| description | string | 否 | 规则描述 | - |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| forward_rules | []ForwardRule | 高级路由规则列表（镜像请求参数） |
| basic_forward_rules | []BasicForwardRule | 基础路由规则列表（镜像请求参数） |
| forward_cases_code | int | 路由用例代码（仅在存在路由用例时返回） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AR-1-001 | 仅设置基础路由规则 | 正常参数 | 传入 basic_forward_rules 数组 |
| AR-1-002 | 仅设置高级路由规则 | 正常参数 | 传入 forward_rules 数组 |
| AR-1-003 | 同时设置基础和高级路由 | 正常参数 | 同时传入两种规则 |
| AR-1-004 | 清空所有规则 | 正常参数 | 传入空数组 |
| AR-1-005 | 设置多条高级路由规则 | 正常参数 | 传入多条 forward_rules |
| AR-1-006 | 设置多条基础路由规则 | 正常参数 | 传入多条 basic_forward_rules |
| AR-1-007 | 缺少 forward_rules[].expression | 必填校验 | 不传 expression |
| AR-1-008 | 缺少 forward_rules[].cluster_name | 必填校验 | 不传 cluster_name |
| AR-1-009 | 缺少 basic_forward_rules[].cluster_name | 必填校验 | 不传 cluster_name |
| AR-1-010 | forward_rules[].expression 为空字符串 | 边界值 | expression="" |
| AR-1-011 | forward_rules[].cluster_name 为空字符串 | 边界值 | cluster_name="" |
| AR-1-012 | basic_forward_rules[].cluster_name 为空字符串 | 边界值 | cluster_name="" |
| AR-1-013 | forward_rules 数组元素为 null | nil 校验 | 传入 [null] |
| AR-1-014 | basic_forward_rules 数组元素为 null | nil 校验 | 传入 [null] |
| AR-1-015 | 空 Body | 边界值 | 不传任何字段 |
| AR-1-016 | 非法 JSON Body | 异常输入 | 传入非 JSON 字符串 |
| AR-1-017 | forward_rules[].description 可选 | 可选字段 | 不传 description |
| AR-1-018 | basic_forward_rules[].host_names 可选 | 可选字段 | 不传 host_names |
| AR-1-019 | basic_forward_rules[].paths 可选 | 可选字段 | 不传 paths |
| AR-1-020 | 返回数据镜像请求 | 返回数据校验 | 验证响应 Data 结构与请求一致 |
| AR-1-021 | 最后一条自动追加 default_t() | 业务规则 | 通过 GET 验证末尾追加 default_t() |

---

## 四、测试场景详细设计

---

### AR-1-001：仅设置基础路由规则（正常参数）

#### 设计思路

验证仅传入 basic_forward_rules 时接口能正常处理，返回 ErNum=200。

#### 前提数据准备

- 无需预先创建数据（使用内置 AI_product 产品线）

#### 执行步骤

1. 构造请求 Body，仅包含 basic_forward_rules
2. 发送 PATCH 请求到 `/open-api/v1/ai-route-rules`
3. 验证 ErNum=200

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
            "cluster_name": "BFE-AI_product.szyf",
            "description": "基础路由规则"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data.basic_forward_rules | 长度=1 | LenEquals(1) |
| Data.basic_forward_rules[0].host_names | ["*.example.com"] | Equals |
| Data.basic_forward_rules[0].paths | ["/api"] | Equals |
| Data.basic_forward_rules[0].cluster_name | "BFE-AI_product.szyf" | Equals |
| Data.basic_forward_rules[0].description | "基础路由规则" | Equals |
| Data.forward_rules | 空数组 | LenEquals(0) |

---

### AR-1-002：仅设置高级路由规则（正常参数）

#### 设计思路

验证仅传入 forward_rules 时接口能正常处理，返回 ErNum=200。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 构造请求 Body，仅包含 forward_rules
2. 发送 PATCH 请求
3. 验证 ErNum=200

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "description": "路由到集群1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data.forward_rules | 长度=1 | LenEquals(1) |
| Data.forward_rules[0].name | "rule1" | Equals |
| Data.forward_rules[0].expression | "req_host_in(\"api.example.com\")" | Equals |
| Data.forward_rules[0].cluster_name | "BFE-AI_product.szyf" | Equals |
| Data.basic_forward_rules | 空数组 | LenEquals(0) |

---

### AR-1-003：同时设置基础和高级路由规则（正常参数）

#### 设计思路

验证同时传入两种路由规则时接口能正常处理。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 构造请求 Body，同时包含 forward_rules 和 basic_forward_rules
2. 发送 PATCH 请求
3. 验证 ErNum=200

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule-adv",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api/v1"],
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data.forward_rules | 长度=1 | LenEquals(1) |
| Data.basic_forward_rules | 长度=1 | LenEquals(1) |

---

### AR-1-004：清空所有规则（正常参数）

#### 设计思路

验证传入空数组可以清空已设置的规则。

#### 前提数据准备

- 先设置一条规则（调用 PATCH 设置 forward_rules）

#### 执行步骤

1. 先设置一条规则
2. 再发送 PATCH 请求，传入空数组
3. 通过 GET 验证规则已清空

#### 请求参数

```json
{
    "forward_rules": [],
    "basic_forward_rules": []
}
```

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  

**后续 GET 验证**：

| 字段 | 预期值 |
|------|--------|
| Data.forward_rules | [] |
| Data.basic_forward_rules | [] |

---

### AR-1-005：设置多条高级路由规则（正常参数）

#### 设计思路

验证传入多条 forward_rules 时接口能正常处理。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule-1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        },
        {
            "name": "rule-2",
            "expression": "req_host_in(\"test.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        },
        {
            "name": "rule-3",
            "expression": "default_t()",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**Data.forward_rules**：长度=3

---

### AR-1-006：设置多条基础路由规则（正常参数）

#### 设计思路

验证传入多条 basic_forward_rules 时接口能正常处理。

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
            "cluster_name": "BFE-AI_product.szyf"
        },
        {
            "host_names": ["*.test.com"],
            "paths": ["/v2"],
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**Data.basic_forward_rules**：长度=2

---

### AR-1-007：缺少 forward_rules[].expression（必填校验）

#### 设计思路

验证 forward_rules 数组中每个元素的 expression 字段为必填（`validate:"required,min=1"`，缺失时 go-playground/validator 报错，返回 422。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "expression" 或 "Param Illegal" 的错误信息

---

### AR-1-008：缺少 forward_rules[].cluster_name（必填校验）

#### 设计思路

验证 forward_rules 数组中每个元素的 cluster_name 字段为必填。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "req_host_in(\"api.example.com\")"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "cluster_name" 或 "Param Illegal" 的错误信息

---

### AR-1-009：缺少 basic_forward_rules[].cluster_name（必填校验）

#### 设计思路

验证 basic_forward_rules 数组中每个元素的 cluster_name 字段为必填。

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"]
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "cluster_name" 或 "Param Illegal" 的错误信息

---

### AR-1-010：forward_rules[].expression 为空字符串（边界值）

#### 设计思路

验证 expression 传递空字符串时，由于 `validate:"required,min=1"` 约束，validator 会报错返回 422。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "expression" 或 "Param Illegal" 的错误信息

---

### AR-1-011：forward_rules[].cluster_name 为空字符串（边界值）

#### 设计思路

验证 cluster_name 传递空字符串时的行为。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": ""
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "cluster_name" 或 "Param Illegal" 的错误信息

---

### AR-1-012：basic_forward_rules[].cluster_name 为空字符串（边界值）

#### 设计思路

验证 basic_forward_rules 中 cluster_name 传递空字符串时的行为。

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "cluster_name": ""
        }
    ]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "cluster_name" 或 "Param Illegal" 的错误信息

---

### AR-1-013：forward_rules 数组元素为 null（nil 校验）

#### 设计思路

验证 forward_rules 数组中包含 null 元素时，代码中的手动 nil 检查（`newRuleInfoFromReq`）会返回 422 + "AdvanceRouteRules element cant be nil"。

#### 请求参数

```json
{
    "forward_rules": [null]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "element cant be nil" 或 "Param Illegal"

---

### AR-1-014：basic_forward_rules 数组元素为 null（nil 校验）

#### 设计思路

验证 basic_forward_rules 数组中包含 null 元素时，代码中的手动 nil 检查会返回 422。

#### 请求参数

```json
{
    "basic_forward_rules": [null]
}
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "element cant be nil" 或 "Param Illegal"

---

### AR-1-015：空 Body（边界值）

#### 设计思路

验证请求 Body 为空时，JSON 反序列化到结构体，所有字段为空（零值），接口应能正常处理（空 Body 等价于传入空数组）。

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  
**Data.forward_rules**：null 或 []  
**Data.basic_forward_rules**：null 或 []

---

### AR-1-016：非法 JSON Body（异常输入）

#### 设计思路

验证传入非 JSON 格式的 Body 时，`json.NewDecoder` 会反序列化失败，返回 422。

#### 请求参数

```
this is not json
```

#### 预期返回结果

**ErNum**：422  
**ErrMsg**：包含 "Param Illegal" 的错误信息

---

### AR-1-017：forward_rules[].description 可选（可选字段）

#### 设计思路

验证 forward_rules 中 description 字段为可选，不传时接口正常处理。

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**Data.forward_rules[0].description**：空字符串 ""

---

### AR-1-018：basic_forward_rules[].host_names 可选（可选字段）

#### 设计思路

验证 basic_forward_rules 中 host_names 字段为可选。

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "paths": ["/api"],
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**Data.basic_forward_rules[0].host_names**：null 或 []

---

### AR-1-019：basic_forward_rules[].paths 可选（可选字段）

#### 设计思路

验证 basic_forward_rules 中 paths 字段为可选。

#### 请求参数

```json
{
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  
**Data.basic_forward_rules[0].paths**：null 或 []

---

### AR-1-020：返回数据镜像请求（返回数据校验）

#### 设计思路

验证 PATCH 接口返回的 Data 是请求参数的镜像（不包含自动追加的 default_t() 规则）。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 PATCH 请求
2. 验证返回的 Data 结构与请求参数完全一致
3. 通过 GET 验证存储的规则包含自动追加的 default_t()

#### 请求参数

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "description": "测试规则",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
            "cluster_name": "BFE-AI_product.szyf",
            "description": "基础路由"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  

**Data 字段逐项校验**：

| 字段 | 预期值 |
|------|--------|
| Data.forward_rules[0].name | "rule1" |
| Data.forward_rules[0].description | "测试规则" |
| Data.forward_rules[0].expression | "req_host_in(\"api.example.com\")" |
| Data.forward_rules[0].cluster_name | "BFE-AI_product.szyf" |
| Data.basic_forward_rules[0].host_names | ["*.example.com"] |
| Data.basic_forward_rules[0].paths | ["/api"] |
| Data.basic_forward_rules[0].cluster_name | "BFE-AI_product.szyf" |
| Data.basic_forward_rules[0].description | "基础路由" |

---

### AR-1-021：最后一条自动追加 default_t()（业务规则）

#### 设计思路

验证 `routeRuleParam2routeRule` 函数的业务逻辑：当 forward_rules 非空且最后一条 expression 不是 "default_t()" 时，自动追加一条 default_t() 规则，其 cluster_name 取最后一条规则的 cluster_name。通过 GET 接口验证存储的规则包含自动追加的 default_t()。

#### 前提数据准备

- 无需预先创建数据

#### 执行步骤

1. 发送 PATCH 请求，forward_rules 最后一条不是 default_t()
2. 验证 PATCH 返回 ErNum=200
3. 发送 GET 请求获取已存储的规则
4. 验证 GET 返回的 forward_rules 最后一条是 default_t()

#### 请求参数（PATCH）

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ]
}
```

#### 预期 PATCH 返回结果

**ErNum**：200  
**Data.forward_rules**：长度=1（不包含自动追加的 default_t()）

#### 预期 GET 返回结果

**Data.forward_rules**：长度=2  
**Data.forward_rules[1].expression**："default_t()"  
**Data.forward_rules[1].cluster_name**：与第一条规则相同