# Global Route Rules 测试用例设计文档

## 1. 模块概述

Global Route Rules 模块负责全局（global）路由表的管理，包括全量更新和查询。该路由表 `type` 固定为 `global`、`owner` 固定为 `global`，用于网关未命中 API-Key/Entity 级路由时的默认路由决策。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| GRR-1 | 更新 Global 路由表 | PUT | `/open-api/v1/global-route-rules` | 全量替换 global 路由表 |
| GRR-2 | 查询 Global 路由表 | GET | `/open-api/v1/global-route-rules` | 返回当前 global 路由表；不存在则 Data 为 null |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 更新 Global 路由表 | 9 |
| 查询 Global 路由表 | 3 |
| **合计** | **12** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
global_route_rules/
├── design.md
├── update/
│   └── update_test.go
└── get/
    └── get_test.go
```

## 6. 更新 Global 路由表

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Global Route Rules |
| 接口名称 | 更新 Global 路由表 |
| 方法 | PUT |
| 路径 | `/open-api/v1/global-route-rules` |
| 说明 | 全量替换 global 路由表 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| enabled | bool | N | 是否启用，默认 true | - |
| rules | []Rule | Y | 规则列表 | 必填 |
| rules[].name | string | Y | 规则名称，同一 global 路由表内唯一 | 必填、非空；同一组 `rules` 内唯一 |
| rules[].Cond | string | Y | BFE 条件表达式 | 必填、非空；须为合法 BFE 条件表达式 |
| rules[].targets | []Target | Y | 转发目标列表，权重总和为 100 | 必填，至少 1 个元素 |
| rules[].targets[].ClusterName | string | Y | 后端集群名称 | 必填；符合 ClusterName 类型；须为已存在集群 |
| rules[].targets[].Model | string | N | 模型名称，空字符串表示透传原始模型 | 非空时须为对应集群已配置的模型名称 |
| rules[].targets[].Weight | int | Y | 权重 | 0-100；同一规则内总和必须等于 100 |
| rules[].fallbacks | []Fallback | N | 降级目标列表 | 可选；允许为空 |
| rules[].fallbacks[].ClusterName | string | Y | 后端集群名称 | 必填；符合 ClusterName 类型；须为已存在集群 |
| rules[].fallbacks[].Model | string | N | 模型名称，空字符串表示透传原始模型 | 非空时须为对应集群已配置的模型名称 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| enabled | bool | 是否启用 |
| rules | []Rule | 规则列表 |
| rules[].name | string | 规则名称 |
| rules[].Cond | string | BFE 条件表达式 |
| rules[].targets | []Target | 转发目标列表 |
| rules[].fallbacks | []Fallback | 降级目标列表 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| GRR-1-001 | 最小参数更新 | 正常参数 | 仅传 rules，enabled 默认 true |
| GRR-1-002 | 完整参数更新 | 正常参数 | enabled=false + 多 target + fallbacks |
| GRR-1-003 | 更新后查询一致性 | 返回数据 | PUT 后立即 GET，验证数据一致 |
| GRR-1-004 | 规则名称重复 | 异常参数 | 验证 ErrNum=422 |
| GRR-1-005 | targets 权重总和不等于 100 | 异常参数 | 验证 ErrNum=422 |
| GRR-1-006 | fallbacks ClusterName 为空 | 异常参数 | 验证 ErrNum=422 |
| GRR-1-007 | Cond 表达式非法 | 异常参数 | 验证 ErrNum=422（暂跳过） |
| GRR-1-008 | 重复 target (ClusterName+Model) | 合法性条件 | 验证 ErrNum=422 |
| GRR-1-009 | target ClusterName 格式非法 | 合法性条件 | 验证 ErrNum=422 |

### 6.4 测试场景详细设计

#### 6.4.1 GRR-1-001：最小参数更新 Global 路由表（正常参数）

##### 设计思路

验证仅传入 `rules` 时，`enabled` 默认填充为 `true`，并返回完整路由表。

##### 前提数据准备

已创建后端 Cluster（如 `cluster_global`）。

##### 执行步骤

1. 发送 PUT 请求到 `/open-api/v1/global-route-rules`，传入 rules。
2. 验证响应状态码和返回结构。
3. 验证 `enabled=true`。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "global-default",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "cluster_global",
                    "Model": "",
                    "Weight": 100
                }
            ],
            "fallbacks": []
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | true | Equals |
| rules | 长度为 1 | Len=1 |
| rules[0].name | "global-default" | Equals |
| rules[0].Cond | "default_t()" | Equals |
| rules[0].targets[0].ClusterName | "cluster_global" | Equals |
| rules[0].targets[0].Weight | 100 | Equals |

---

#### 6.4.2 GRR-1-002：完整参数更新 Global 路由表（正常参数）

##### 设计思路

验证完整参数更新，包括 `enabled=false`、多 targets、fallbacks。

##### 前提数据准备

已创建多个后端 Cluster（如 `cluster_a`、`cluster_b`、`cluster_fallback`）。

##### 执行步骤

1. 发送 PUT 请求，传入完整参数。
2. 验证返回结构与输入一致。

##### 请求参数

```json
{
    "enabled": false,
    "rules": [
        {
            "name": "rule-a",
            "Cond": "req_path_prefix(\"/v1\")",
            "targets": [
                {
                    "ClusterName": "cluster_a",
                    "Model": "gpt-4",
                    "Weight": 70
                },
                {
                    "ClusterName": "cluster_b",
                    "Model": "",
                    "Weight": 30
                }
            ],
            "fallbacks": [
                {
                    "ClusterName": "cluster_fallback",
                    "Model": ""
                }
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | false | Equals |
| rules[0].targets | 长度为 2 | Len=2 |
| rules[0].targets[0].Weight + rules[0].targets[1].Weight | 100 | Sum=100 |
| rules[0].fallbacks[0].ClusterName | "cluster_fallback" | Equals |

---

#### 6.4.3 GRR-1-003：更新后查询一致性（返回数据）

##### 设计思路

验证 PUT 更新成功后，立即通过 GET 查询，返回的 global 路由表与更新请求一致。

##### 前提数据准备

已创建后端 Cluster。

##### 执行步骤

1. 发送 PUT 请求更新 global 路由表。
2. 发送 GET 请求查询 global 路由表。
3. 对比两次返回的 `enabled` 和 `rules` 内容是否一致。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "global-default",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "cluster_global",
                    "Model": "",
                    "Weight": 100
                }
            ],
            "fallbacks": []
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | 与请求一致 | Equals |
| rules | 与请求一致 | DeepEquals |

---

#### 6.4.4 GRR-1-004：规则名称重复（异常参数）

##### 设计思路

验证 `rules` 中规则名称在同一 global 路由表内必须唯一，重复时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，传入两条同名的规则。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "dup",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "c1",
                    "Weight": 100
                }
            ]
        },
        {
            "name": "dup",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "c2",
                    "Weight": 100
                }
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含规则名称重复的错误信息  
**Data**：null

---

#### 6.4.5 GRR-1-005：targets 权重总和不等于 100（异常参数）

##### 设计思路

验证每个规则的 `targets` 权重总和必须等于 100。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，传入权重总和不等于 100 的 targets。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "bad-weight",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "c1",
                    "Weight": 60
                },
                {
                    "ClusterName": "c2",
                    "Weight": 30
                }
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含权重总和必须为 100 的错误信息  
**Data**：null

---

#### 6.4.6 GRR-1-006：fallbacks ClusterName 为空（异常参数）

##### 设计思路

验证 `fallbacks` 中元素的 `ClusterName` 不能为空。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，传入 `fallbacks` 中 `ClusterName` 为空字符串的规则。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "bad-fb",
            "Cond": "default_t()",
            "targets": [
                {
                    "ClusterName": "c1",
                    "Weight": 100
                }
            ],
            "fallbacks": [
                {
                    "ClusterName": "",
                    "Model": ""
                }
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 fallback ClusterName 不能为空的错误信息  
**Data**：null

---

#### 6.4.7 GRR-1-007：Cond 表达式非法（异常参数）

##### 设计思路

验证 `Cond` 必须为合法的 BFE 条件表达式。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，传入非法的 `Cond` 表达式。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "bad-cond",
            "Cond": "not_a_valid_expr(",
            "targets": [
                {
                    "ClusterName": "c1",
                    "Weight": 100
                }
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含表达式非法的错误信息  
**Data**：null

---

#### 6.4.8 GRR-1-008：重复 target (ClusterName+Model)（合法性条件）

##### 设计思路

验证同一规则 `targets` 内 `(ClusterName, Model)` 组合不能重复。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，同一规则内包含两条相同 `(ClusterName, Model)` 的 target。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "dup-target",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "c1", "Model": "m1", "Weight": 50},
                {"ClusterName": "c1", "Model": "m1", "Weight": 50}
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含重复 target 的错误信息  
**Data**：null

---

#### 6.4.9 GRR-1-009：target ClusterName 格式非法（合法性条件）

##### 设计思路

验证 `targets` 中 `ClusterName` 必须符合 ClusterName 类型。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PUT 请求，`ClusterName` 以 `-` 开头。
2. 验证返回错误码。

##### 请求参数

```json
{
    "rules": [
        {
            "name": "bad-cluster",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "-bad", "Weight": 100}
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 ClusterName 非法的错误信息  
**Data**：null

---

## 7. 查询 Global 路由表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Global Route Rules |
| 接口名称 | 查询 Global 路由表 |
| 方法 | GET |
| 路径 | `/open-api/v1/global-route-rules` |
| 说明 | 返回当前 global 路由表；不存在则 Data 为 null |

### 7.2 接口参数说明

#### 7.2.1 请求参数

无

#### 7.2.2 返回数据字段

同 6.2.2。

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| GRR-2-001 | 查询已更新的 Global 路由表 | 正常参数 | 返回最近一次 PUT 内容 |
| GRR-2-002 | 查询未配置的 Global 路由表 | 边界值 | Data 为 null（系统初始化后默认已存在，需全新数据库验证） |

### 7.4 测试场景详细设计

#### 7.4.1 GRR-2-001：查询已更新的 Global 路由表（正常参数）

##### 设计思路

验证 GET 接口返回最近一次 PUT 写入的 global 路由表内容。

##### 前提数据准备

已通过 GRR-1-001 或 GRR-1-002 更新 global 路由表。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/global-route-rules`。
2. 验证返回的 `enabled` 和 `rules` 与最近一次 PUT 一致。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | 与最近一次 PUT 一致 | Equals |
| rules | 与最近一次 PUT 一致 | DeepEquals |

---

#### 7.4.2 GRR-2-002：查询未配置的 Global 路由表（边界值）

##### 设计思路

验证数据库中无 global 路由表时，接口返回 `Data=null`。

##### 前提数据准备

数据库中无 global 路由表。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/global-route-rules`。
2. 验证 `Data` 为 `null`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |

---

#### 7.4.3 GRR-2-003：查询系统默认 Global 路由表（正常参数）

##### 设计思路

验证系统初始化后自动创建的默认 global 路由表可被查询到。

##### 前提数据准备

无（依赖系统启动时的自动初始化）。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/global-route-rules`。
2. 验证返回的 `enabled=false`、`rules=[]`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| enabled | false | Equals |
| rules | 空数组 | IsEmpty |

---

## 8. 依赖与数据准备

1. 需要在测试前置中准备至少一个可用的 Cluster，供 `targets` 与 `fallbacks` 引用。
2. 由于 OpenAPI 层不校验 Cluster 是否真实存在，也可使用占位名称验证参数校验逻辑，但建议结合真实 Cluster 验证端到端一致性。

## 9. 注意事项

1. PUT 为全量更新，会替换整个 global 路由表。
2. `type` 与 `owner` 对调用方不可见，由服务端固定写入。
3. 本模块不直接涉及鉴权，测试环境跳过 Token 校验。
4. 更新后应恢复默认全局路由或确保不影响其他模块测试。
