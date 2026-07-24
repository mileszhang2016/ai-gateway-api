# 获取AI路由规则列表 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | AI 路由规则 |
| 接口名称 | 获取AI路由规则列表 |
| 方法 | GET |
| 路径 | /open-api/v1/ai-route-rules |
| 说明 | 获取当前产品线（AI_product）的 AI 路由规则配置，无规则时返回空数组 |

---

## 二、接口参数说明

### 请求参数

无 Body 参数，无 URI 参数。

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| forward_rules | []ForwardRule | 高级路由规则列表，无规则时为空数组 |
| basic_forward_rules | []BasicForwardRule | 基础路由规则列表，无规则时为空数组 |
| forward_cases_code | int | 路由用例代码（仅在存在路由用例时返回） |

**ForwardRule 元素**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | 规则名称 |
| description | string | 规则描述 |
| expression | string | 条件表达式 |
| cluster_name | string | 目标集群名称 |

**BasicForwardRule 元素**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| host_names | []string | 域名列表 |
| paths | []string | 路径列表 |
| cluster_name | string | 目标集群名称 |
| description | string | 规则描述 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AR-2-001 | 获取已设置的规则 | 正常参数 | 先设置规则，再获取，验证返回数据一致 |
| AR-2-002 | 获取未设置时的列表 | 空数据 | 不设置规则，直接获取，验证返回空数组 |
| AR-2-003 | 返回数据结构校验 | 返回数据校验 | 验证返回字段完整且类型正确 |

---

## 四、测试场景详细设计

---

### AR-2-001：获取已设置的规则（正常参数）

#### 设计思路

验证获取已设置的路由规则时，返回的数据与设置的规则一致。同时验证自动追加的 default_t() 规则在 GET 返回中存在。

#### 前提数据准备

- 先调用 PATCH 设置规则（包含 forward_rules 和 basic_forward_rules）

#### 执行步骤

1. 调用 PATCH 设置路由规则
2. 调用 GET 获取路由规则
3. 验证返回的规则与设置的一致
4. 验证 forward_rules 末尾包含自动追加的 default_t()

#### 期望（PATCH 设置）

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
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
| Data.forward_rules | 长度 ≥ 1 | GreaterThanOrEqual(1) |
| Data.forward_rules[0].name | "rule1" | Equals |
| Data.forward_rules[0].expression | "req_host_in(\"api.example.com\")" | Equals |
| Data.forward_rules[0].cluster_name | "BFE-AI_product.szyf" | Equals |
| Data.forward_rules 末尾 | expression="default_t()" | 末尾有 default_t() |
| Data.basic_forward_rules | 长度=1 | LenEquals(1) |
| Data.basic_forward_rules[0].host_names | ["*.example.com"] | Equals |
| Data.basic_forward_rules[0].paths | ["/api"] | Equals |
| Data.basic_forward_rules[0].cluster_name | "BFE-AI_product.szyf" | Equals |

---

### AR-2-002：获取未设置时的列表（空数据）

#### 设计思路

验证当前产品线没有任何路由规则时，GET 接口返回空数组而非 null 或错误。

#### 前提数据准备

- 确保规则已清空（先调用 PATCH 传入空数组，或直接在新测试环境中执行）

#### 执行步骤

1. 调用 PATCH 清空规则（传入空数组）
2. 调用 GET 获取路由规则
3. 验证返回空数组结构

#### 预期返回结果

**ErNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data.forward_rules | [] | LenEquals(0) |
| Data.basic_forward_rules | [] | LenEquals(0) |

---

### AR-2-003：返回数据结构校验（返回数据校验）

#### 设计思路

验证 GET 返回的 Data 结构完整，包含所有应有字段，且类型正确。

#### 前提数据准备

- 先调用 PATCH 设置完整的规则（含 forward_rules 和 basic_forward_rules）

#### 执行步骤

1. 调用 PATCH 设置完整的路由规则
2. 调用 GET 获取路由规则
3. 逐字段校验返回数据的结构和类型

#### 期望（PATCH 设置）

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "description": "测试描述",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "BFE-AI_product.szyf"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com", "api.test.com"],
            "paths": ["/api/v1", "/api/v2"],
            "cluster_name": "BFE-AI_product.szyf",
            "description": "基础路由描述"
        }
    ]
}
```

#### 预期返回结果

**ErNum**：200  

**Data 顶层键校验**：

| 键名 | 预期类型 | 预期值 |
|------|---------|--------|
| forward_rules | array | 非空 |
| basic_forward_rules | array | 非空 |

**Data.forward_rules[0] 字段校验**：

| 键名 | 预期类型 | 预期值 |
|------|---------|--------|
| name | string | "rule1" |
| description | string | "测试描述" |
| expression | string | "req_host_in(\"api.example.com\")" |
| cluster_name | string | "BFE-AI_product.szyf" |

**Data.basic_forward_rules[0] 字段校验**：

| 键名 | 预期类型 | 预期值 |
|------|---------|--------|
| host_names | []string | ["*.example.com", "api.test.com"] |
| paths | []string | ["/api/v1", "/api/v2"] |
| cluster_name | string | "BFE-AI_product.szyf" |
| description | string | "基础路由描述" |

**额外验证 forward_rules 末尾**：

| 检查项 | 预期 |
|--------|------|
| forward_rules 最后一条 expression | "default_t()" |