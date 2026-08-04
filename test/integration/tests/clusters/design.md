# Cluster 测试用例设计文档

## 1. 模块概述

Cluster 模块负责 AI 网关后端集群的管理，包括创建、查询、更新、删除。v0.3.0 是本次变更最大模块：删除 `ready`、`sub_clusters`、`scheduler` 等内部字段对外暴露；`Instance` 删除 `tags`；`llm_config` 必填；`llm_config.models` 为字符串数组；`llm_config.key` 为必填敏感字段；不再通过 OpenAPI 设置/获取 `DefaultAIClusterName`。

另外，删除集群时会检查 `route_rules` 表中的全局、Entity、API-Key 路由规则：若任意规则的 `targets` 或 `fallbacks` 引用了该集群，则删除被拒绝。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| CL-1 | 创建集群 | POST | `/open-api/v1/clusters` | 一键创建集群、实例池、子集群 |
| CL-2 | 查询集群列表 | GET | `/open-api/v1/clusters` | 数组 |
| CL-3 | 查询集群详情 | GET | `/open-api/v1/clusters/{cluster_name}` | - |
| CL-4 | 更新集群 | PATCH | `/open-api/v1/clusters/{cluster_name}` | 可更新描述、实例池、各配置段 |
| CL-5 | 删除集群 | DELETE | `/open-api/v1/clusters/{cluster_name}` | 级联清理实例池、子集群；删除前检查 global/entity/apikey 路由规则引用 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建集群 | 12 |
| 查询集群列表 | 1 |
| 查询集群详情 | 1 |
| 更新集群 | 5 |
| 删除集群 | 8 |
| **合计** | **26** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
clusters/
├── design.md
├── create/
│   └── create_test.go
├── list/
│   └── list_test.go
├── detail/
│   └── detail_test.go
├── update/
│   └── update_test.go
└── delete/
    └── delete_test.go
```

## 6. 创建集群

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Cluster |
| 接口名称 | 创建集群 |
| 方法 | POST |
| 路径 | `/open-api/v1/clusters` |
| 说明 | 一键创建集群、实例池、子集群 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| name | string | Y | 集群名，全局唯一 | 长度 1-64；仅允许字母、数字、`_`、`-`、`.`；不能以 `.`、`-`、`_` 开头或结尾；全局唯一；不能包含空白字符 |
| description | string | N | 集群描述 | 长度 ≤256；不能包含控制字符 |
| instance_pool | []Instance | Y | 实例列表 | ≥1 个元素；(hostname, ip) 组合唯一；至少一个实例 weight>0 |
| instance_pool[].hostname | string | Y | 实例主机名 | 符合 Hostname 类型 |
| instance_pool[].ip | string | Y | 实例 IP 地址 | 符合 IP Address 类型 |
| instance_pool[].weight | int | Y | 实例权重，范围 [0,100] | 0-100 |
| instance_pool[].ports | map[string]int | Y | 实例端口，至少包含 Default | ≥1 个键值对；必须包含 `Default`；端口值唯一 |
| basic | object | N | 基本参数 | `protocol` ∈ {http, https}；超时 >0；重试/连接数 ≥0 |
| sticky_sessions | object | N | 会话保持 | `hash_strategy` ∈ {CLIENT_IP_ONLY, CLIENT_ID_ONLY, CLIENT_ID_PREFERED} |
| passive_health_check | object | N | 被动健康检查 | `uri` 非空且以 `/` 开头；`statuscode` 为 0 或 100-599 |
| llm_config | object | Y | AI LLM 服务配置 | 必填；`models` ≥1 个非空唯一字符串；`key` ≤512；`model_endpoint.schema` ∈ {http, https}；`model_mappings.source_model` 唯一 |
| llm_config.model_endpoint | object | N | 模型列表端点配置 | `schema` ∈ {http, https} |
| llm_config.models | []string | Y | 支持的模型名称列表 | ≥1 个非空唯一字符串 |
| llm_config.model_mappings | []object | N | 模型名称映射 | `source_model`/`target_model` 必填；`source_model` 唯一 |
| llm_config.key | string | Y | 服务认证密钥 | 必填；长度 ≤512 |
| llm_config.provider_type | string | Y | AI 模型提供商类型 | 必填 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | 集群名 |
| description | string | 集群描述 |
| instance_pool | []Instance | 实例列表 |
| basic | object | 基本参数 |
| sticky_sessions | object | 会话保持 |
| passive_health_check | object | 被动健康检查 |
| llm_config | object | LLM 配置 |

> 返回不含 `ready`、`sub_clusters`、`scheduler`、`Instance.tags`。

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| CL-1-001 | 最小参数创建集群 | 正常参数 | 验证默认值填充 |
| CL-1-002 | 完整参数创建集群 | 正常参数 | 验证完整结构 |
| CL-1-003 | 缺少 llm_config | 必填校验 | 验证 ErrNum=422 |
| CL-1-004 | 缺少 instance_pool | 必填校验 | 验证 ErrNum=422 |
| CL-1-005 | 重复集群名 | 业务规则 | 验证 ErrNum=555/556 |
| CL-1-006 | instance_pool 为空数组 | 异常参数 | 验证 ErrNum=422 |
| CL-1-007 | 实例不含 Default 端口 | 异常参数 | 验证 ErrNum=422 |
| CL-1-008 | 非法 hostname | 合法性条件 | 验证 ErrNum=422 |
| CL-1-009 | 非法 IP | 合法性条件 | 验证 ErrNum=422 |
| CL-1-010 | weight 超过 100 | 合法性条件 | 验证 ErrNum=422 |
| CL-1-011 | 重复实例 (hostname+ip) | 合法性条件 | 验证 ErrNum=422 |
| CL-1-012 | llm_config 模型重复 | 合法性条件 | 验证 ErrNum=422 |

### 6.4 测试场景详细设计

#### 6.4.1 CL-1-001：最小参数创建集群（正常参数）

##### 设计思路

验证最小参数创建集群时，`basic`、`sticky_sessions`、`passive_health_check` 使用默认值填充，且不包含内部字段。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/clusters`，传入最小参数。
2. 验证响应状态码和返回结构。
3. 验证返回不含 `ready`、`sub_clusters`、`scheduler`。

##### 请求参数

```json
{
    "name": "cluster_min",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["deepseek-chat"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "cluster_min" | Equals |
| instance_pool | 长度为 1 | Len=1 |
| basic | 非空对象 | IsObject |
| sticky_sessions | 非空对象 | IsObject |
| passive_health_check | 非空对象 | IsObject |
| llm_config.models | ["deepseek-chat"] | Equals |
| llm_config.key | "sk-xxx" | Equals |
| ready | 不存在 | NotExists |
| sub_clusters | 不存在 | NotExists |
| scheduler | 不存在 | NotExists |

---

#### 6.4.2 CL-1-002：完整参数创建集群（正常参数）

##### 设计思路

验证完整参数创建集群时返回结构与输入一致。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，传入完整参数。
2. 验证返回结构与输入一致。

##### 请求参数

```json
{
    "name": "cluster_full",
    "description": "完整集群",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {
                "Default": 8080
            }
        },
        {
            "hostname": "backend-2",
            "ip": "10.0.0.2",
            "weight": 50,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "basic": {
        "protocol": "http",
        "connection": {
            "max_idle_conn_per_rs": 0,
            "cancel_on_client_close": false
        },
        "retries": {
            "max_retry_in_cluster": 2
        },
        "buffers": {
            "req_write_buffer_size": 512
        },
        "timeouts": {
            "timeout_conn_serv": 50000,
            "timeout_response_header": 50000,
            "timeout_readbody_client": 30000,
            "timeout_read_client_again": 30000,
            "timeout_write_client": 60000
        }
    },
    "sticky_sessions": {
        "enabled": false,
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 3,
        "host": "",
        "uri": "/",
        "statuscode": 0
    },
    "llm_config": {
        "model_endpoint": {
            "schema": "https",
            "uri": "/v1/models",
            "headers": {
                "Authorization": "Bearer ${API_KEY}"
            }
        },
        "models": ["deepseek-chat", "deepseek-coder"],
        "model_mappings": [
            {
                "key": "gpt-4",
                "value": "deepseek-chat"
            }
        ],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "cluster_full" | Equals |
| description | "完整集群" | Equals |
| instance_pool | 长度为 2 | Len=2 |
| llm_config.models | ["deepseek-chat", "deepseek-coder"] | Equals |
| ready | 不存在 | NotExists |
| sub_clusters | 不存在 | NotExists |
| scheduler | 不存在 | NotExists |

---

#### 6.4.3 CL-1-003：缺少 llm_config（必填校验）

##### 设计思路

验证 `llm_config` 为必填字段，缺少时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `llm_config`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_no_llm",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "llm_config" 的错误信息  
**Data**：null

---

#### 6.4.4 CL-1-004：缺少 instance_pool（必填校验）

##### 设计思路

验证 `instance_pool` 为必填字段，缺少时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，Body 中缺少 `instance_pool`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_no_pool",
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "instance_pool" 的错误信息  
**Data**：null

---

#### 6.4.5 CL-1-005：重复集群名（业务规则）

##### 设计思路

验证集群名全局唯一，重复创建时返回错误。

##### 前提数据准备

已创建 `cluster_dup`。

##### 执行步骤

1. 发送 POST 请求，使用重复集群名。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_dup",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：555 或 556  
**ErrMsg**：集群名已存在的错误信息  
**Data**：null

---

#### 6.4.6 CL-1-006：instance_pool 为空数组（异常参数）

##### 设计思路

验证 `instance_pool` 不能为空数组。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`instance_pool` 为空数组。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_empty_pool",
    "instance_pool": [],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "instance_pool" 或至少 1 个元素的错误信息  
**Data**：null

---

#### 6.4.7 CL-1-007：实例不含 Default 端口（异常参数）

##### 设计思路

验证实例端口必须包含 `Default` 键。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，实例 `ports` 中不包含 `Default`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_bad_port",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Other": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Default" 或 "ports" 的错误信息  
**Data**：null

---

#### 6.4.8 CL-1-008：非法 hostname（合法性条件）

##### 设计思路

验证实例 `hostname` 必须符合 Hostname 类型（不能以 `-` 开头等）。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，实例 `hostname` 以 `-` 开头。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_bad_hostname",
    "instance_pool": [
        {
            "hostname": "-bad",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 hostname 非法的错误信息  
**Data**：null

---

#### 6.4.9 CL-1-009：非法 IP（合法性条件）

##### 设计思路

验证实例 `ip` 必须是合法 IPv4/IPv6 地址。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，实例 `ip` 为非法字符串。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_bad_ip",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "not-an-ip",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 ip 非法的错误信息  
**Data**：null

---

#### 6.4.10 CL-1-010：weight 超过 100（合法性条件）

##### 设计思路

验证实例 `weight` 取值范围为 0-100。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，实例 `weight=101`。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_bad_weight",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 101,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 weight 非法的错误信息  
**Data**：null

---

#### 6.4.11 CL-1-011：重复实例 (hostname+ip)（合法性条件）

##### 设计思路

验证同一集群内 `(hostname, ip)` 组合不能重复。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`instance_pool` 包含两个相同 `(hostname, ip)` 的实例。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_dup_instance",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {
                "Default": 8080
            }
        },
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {
                "Default": 8081
            }
        }
    ],
    "llm_config": {
        "models": ["m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含重复实例的错误信息  
**Data**：null

---

#### 6.4.12 CL-1-012：llm_config 模型重复（合法性条件）

##### 设计思路

验证 `llm_config.models` 中的模型名称不能重复。

##### 前提数据准备

无

##### 执行步骤

1. 发送 POST 请求，`llm_config.models` 包含重复模型名。
2. 验证返回错误码。

##### 请求参数

```json
{
    "name": "cluster_dup_model",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ],
    "llm_config": {
        "models": ["m", "m"],
        "key": "sk-xxx",
        "provider_type": "deepseek"
    }
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含重复 model 的错误信息  
**Data**：null

---

## 7. 查询集群列表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Cluster |
| 接口名称 | 查询集群列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/clusters` |
| 说明 | 返回所有集群列表 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

无

#### 7.2.2 返回数据字段

数组，单元素同创建接口返回字段。

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| CL-2-001 | 查询集群列表 | 正常参数 | 返回数组不含内部字段 |

### 7.4 测试场景详细设计

#### 7.4.1 CL-2-001：查询集群列表（正常参数）

##### 设计思路

验证列表接口返回所有集群，且元素不含内部字段。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/clusters`。
2. 验证返回数组元素字段完整且不含 `ready`/`sub_clusters`/`scheduler`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[*].name | 非空字符串 | NotEmpty |
| Data[*].ready | 不存在 | NotExists |
| Data[*].sub_clusters | 不存在 | NotExists |
| Data[*].scheduler | 不存在 | NotExists |

---

## 8. 查询集群详情

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Cluster |
| 接口名称 | 查询集群详情 |
| 方法 | GET |
| 路径 | `/open-api/v1/clusters/{cluster_name}` |
| 说明 | 返回单个集群详情 |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| cluster_name | string | Y | 集群名字 |

#### 8.2.2 返回数据字段

同创建接口。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| CL-3-001 | 查询集群详情 | 正常参数 | 字段完整且无内部字段 |

### 8.4 测试场景详细设计

#### 8.4.1 CL-3-001：查询集群详情（正常参数）

##### 设计思路

验证详情接口返回完整集群字段，且不包含内部字段。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/clusters/{cluster_name}`。
2. 验证返回字段完整且不含 `ready`/`sub_clusters`/`scheduler`。

##### 请求参数

URI：`cluster_full`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | 与目标集群一致 | Equals |
| instance_pool | 数组 | IsArray |
| llm_config | 非空对象 | IsObject |
| ready | 不存在 | NotExists |
| sub_clusters | 不存在 | NotExists |
| scheduler | 不存在 | NotExists |

---

## 9. 更新集群

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Cluster |
| 接口名称 | 更新集群 |
| 方法 | PATCH |
| 路径 | `/open-api/v1/clusters/{cluster_name}` |
| 说明 | 更新集群基本信息，可编辑描述、实例池、各配置段 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| cluster_name | string | Y | 集群名字 |

##### Body 参数

可修改字段含义同创建接口。若传入 `instance_pool` 字段，系统会自动同步更新对应的实例池。

#### 9.2.2 返回数据字段

同创建接口。

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| CL-4-001 | 更新实例池 | 正常参数 | instance_pool 更新，其余字段不变 |
| CL-4-002 | 更新 llm_config | 正常参数 | llm_config.models 更新 |
| CL-4-003 | 更新后查询一致性 | 返回数据 | PATCH 后立即 GET，验证数据一致 |
| CL-4-004 | 更新不存在的集群 | 异常参数 | 验证 ErrNum=404 |
| CL-4-005 | 更新非法 instance_pool（非法 IP） | 合法性条件 | 验证 ErrNum=422 |

### 9.4 测试场景详细设计

#### 9.4.1 CL-4-001：更新实例池（正常参数）

##### 设计思路

验证部分更新 `instance_pool` 成功，其余字段保持不变。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/clusters/{cluster_name}`，传入新的 `instance_pool`。
2. 验证返回的 `instance_pool` 已更新。

##### 请求参数

```json
{
    "instance_pool": [
        {
            "hostname": "backend-2",
            "ip": "10.0.0.2",
            "weight": 100,
            "ports": {
                "Default": 9090
            }
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
| instance_pool | 长度为 1 | Len=1 |
| instance_pool[0].hostname | "backend-2" | Equals |
| instance_pool[0].ip | "10.0.0.2" | Equals |
| instance_pool[0].ports.Default | 9090 | Equals |

---

#### 9.4.2 CL-4-002：更新 llm_config（正常参数）

##### 设计思路

验证部分更新 `llm_config` 成功。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 PATCH 请求，传入新的 `llm_config`。
2. 验证返回的 `llm_config.models` 已更新。

##### 请求参数

```json
{
    "llm_config": {
        "models": ["qwen-turbo"],
        "key": "sk-xxx",
        "provider_type": "qwen"
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| llm_config.models | ["qwen-turbo"] | Equals |
| llm_config.provider_type | "qwen" | Equals |

---

#### 9.4.3 CL-4-003：更新后查询一致性（返回数据）

##### 设计思路

验证 PATCH 更新成功后，立即通过 GET 查询，返回的集群数据与更新请求一致。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 PATCH 请求更新集群描述。
2. 发送 GET 请求查询集群。
3. 对比两次返回的描述是否一致。

##### 请求参数

```json
{
    "description": "更新后的集群描述"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| description | "更新后的集群描述" | Equals |

---

#### 9.4.4 CL-4-004：更新不存在的集群（异常参数）

##### 设计思路

验证更新不存在的集群时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/clusters/non_existent_cluster`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_cluster`  
Body：
```json
{
    "description": "x"
}
```

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：集群不存在的错误信息  
**Data**：null

---

#### 9.4.5 CL-4-005：更新非法 instance_pool（非法 IP）

##### 设计思路

验证更新集群时传入的 `instance_pool` 同样受实例合法性条件约束。

##### 前提数据准备

已创建集群。

##### 执行步骤

1. 发送 PATCH 请求，`instance_pool` 中实例 `ip` 为非法字符串。
2. 验证返回错误码。

##### 请求参数

```json
{
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "not-an-ip",
            "weight": 100,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 ip 非法的错误信息  
**Data**：null

---

## 10. 删除集群

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Cluster |
| 接口名称 | 删除集群 |
| 方法 | DELETE |
| 路径 | `/open-api/v1/clusters/{cluster_name}` |
| 说明 | 删除集群，自动级联清理关联的实例池和子集群；若集群被 global/entity/apikey 路由规则引用，则拒绝删除 |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| cluster_name | string | Y | 集群名字 |

#### 10.2.2 返回数据字段

同创建接口。

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| CL-5-001 | 删除集群 | 正常参数 | 级联清理，再次查询返回 404 |
| CL-5-002 | 删除不存在的集群 | 异常参数 | 验证 ErrNum=404 |
| CL-5-003 | 删除被 global 路由规则 target 引用的集群 | 业务规则 | 验证 ErrNum=500，错误信息包含规则名 |
| CL-5-004 | 删除被 global 路由规则 fallback 引用的集群 | 业务规则 | 验证 ErrNum=500，fallback 同样会拦截删除 |
| CL-5-005 | 删除被 entity 路由规则 target 引用的集群 | 业务规则 | 验证 ErrNum=500 |
| CL-5-006 | 删除被 apikey 路由规则 target 引用的集群 | 业务规则 | 验证 ErrNum=500 |
| CL-5-007 | 解除引用后可删除集群 | 正常参数 | 更新 global 路由规则移除引用后删除成功 |
| CL-5-008 | 路由规则引用其他集群时可删除 | 正常参数 | 规则引用 cluster_other，删除 cluster_unref 成功 |

### 10.4 测试场景详细设计

#### 10.4.1 CL-5-001：删除集群（正常参数）

##### 设计思路

验证删除集群成功，并级联清理实例池和子集群。

##### 前提数据准备

已创建集群 `cluster_to_del`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_to_del`。
2. 验证返回被删除集群对象。
3. 再次查询集群，验证返回 404。

##### 请求参数

URI：`cluster_to_del`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | "cluster_to_del" | Equals |

---

#### 10.4.2 CL-5-002：删除不存在的集群（异常参数）

##### 设计思路

验证删除不存在的集群时返回 404。

##### 前提数据准备

无

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/non_existent_cluster`。
2. 验证返回错误码。

##### 请求参数

URI：`non_existent_cluster`

##### 预期返回结果

**ErrNum**：404  
**ErrMsg**：集群不存在的错误信息  
**Data**：null

---

#### 10.4.3 CL-5-003：删除被 global 路由规则 target 引用的集群

##### 设计思路

验证 global 路由规则的 `targets` 引用集群时，删除该集群会被拒绝。

##### 前提数据准备

1. 创建集群 `cluster_global_ref`。
2. 创建另一集群 `cluster_other` 用于后续解除引用场景。
3. 通过 `PUT /open-api/v1/global-route-rules` 设置 global 路由规则，使其 `targets` 引用 `cluster_global_ref`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_global_ref`。
2. 验证返回业务异常。

##### 请求参数

URI：`cluster_global_ref`

Global 路由规则准备请求：

```json
{
    "rules": [
        {
            "name": "global-ref",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "cluster_global_ref", "Model": "", "Weight": 100}
            ],
            "fallbacks": []
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含 "Rule global-ref Refer To This Cluster" 或本地化的“集群被转发规则 global-ref 引用”  
**Data**：null

---

#### 10.4.4 CL-5-004：删除被 global 路由规则 fallback 引用的集群

##### 设计思路

验证 global 路由规则的 `fallbacks` 引用集群时，删除该集群同样会被拒绝。

##### 前提数据准备

1. 创建集群 `cluster_global_fb`。
2. 创建另一集群 `cluster_target`。
3. 通过 `PUT /open-api/v1/global-route-rules` 设置 global 路由规则，target 指向 `cluster_target`，fallback 指向 `cluster_global_fb`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_global_fb`。
2. 验证返回业务异常。

##### 请求参数

URI：`cluster_global_fb`

Global 路由规则准备请求：

```json
{
    "rules": [
        {
            "name": "global-fb-ref",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "cluster_target", "Model": "", "Weight": 100}
            ],
            "fallbacks": [
                {"ClusterName": "cluster_global_fb", "Model": ""}
            ]
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含 "Rule global-fb-ref Refer To This Cluster" 或本地化的“集群被转发规则 global-fb-ref 引用”  
**Data**：null

---

#### 10.4.5 CL-5-005：删除被 entity 路由规则 target 引用的集群

##### 设计思路

验证 Entity 级别的路由规则引用集群时，删除该集群会被拒绝。

##### 前提数据准备

1. 创建 Entity-Type `type_entity_ref`。
2. 创建 Entity `entity_ref`。
3. 创建集群 `cluster_entity_ref`。
4. 通过 `PATCH /open-api/v1/entities/{id}` 为 Entity 设置路由规则，target 引用 `cluster_entity_ref`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_entity_ref`。
2. 验证返回业务异常。

##### 请求参数

URI：`cluster_entity_ref`

Entity 路由规则准备请求：

```json
{
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "entity-ref",
                "Cond": "default_t()",
                "targets": [
                    {"ClusterName": "cluster_entity_ref", "Model": "", "Weight": 100}
                ],
                "fallbacks": []
            }
        ]
    }
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含 "Rule entity-ref Refer To This Cluster" 或本地化的“集群被转发规则 entity-ref 引用”  
**Data**：null

---

#### 10.4.6 CL-5-006：删除被 apikey 路由规则 target 引用的集群

##### 设计思路

验证 API-Key 级别的路由规则引用集群时，删除该集群会被拒绝。

##### 前提数据准备

1. 创建 API-Key `apikey_ref`（不绑定 Entity）。
2. 创建集群 `cluster_apikey_ref`。
3. 通过 `PATCH /open-api/v1/api-keys/{id}` 为 API-Key 设置路由规则，target 引用 `cluster_apikey_ref`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_apikey_ref`。
2. 验证返回业务异常。

##### 请求参数

URI：`cluster_apikey_ref`

API-Key 路由规则准备请求：

```json
{
    "route_rules": {
        "enabled": true,
        "rules": [
            {
                "name": "apikey-ref",
                "Cond": "default_t()",
                "targets": [
                    {"ClusterName": "cluster_apikey_ref", "Model": "", "Weight": 100}
                ],
                "fallbacks": []
            }
        ]
    }
}
```

##### 预期返回结果

**ErrNum**：500  
**ErrMsg**：包含 "Rule apikey-ref Refer To This Cluster" 或本地化的“集群被转发规则 apikey-ref 引用”  
**Data**：null

---

#### 10.4.7 CL-5-007：解除引用后可删除集群

##### 设计思路

验证当路由规则不再引用目标集群后，删除可以成功。

##### 前提数据准备

1. 创建集群 `cluster_after_unref` 和 `cluster_other2`。
2. 通过 `PUT /open-api/v1/global-route-rules` 设置 global 路由规则引用 `cluster_after_unref`。
3. 确认删除 `cluster_after_unref` 被拒绝。
4. 更新 global 路由规则，将 target 改为引用 `cluster_other2`（或清空 rules）。

##### 执行步骤

1. 再次发送 DELETE 请求到 `/open-api/v1/clusters/cluster_after_unref`。
2. 验证返回 200，并再次查询返回 404。

##### 请求参数

URI：`cluster_after_unref`

更新后的 Global 路由规则（示例）：

```json
{
    "rules": [
        {
            "name": "global-ref",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "cluster_other2", "Model": "", "Weight": 100}
            ],
            "fallbacks": []
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：返回被删除集群对象

---

#### 10.4.8 CL-5-008：路由规则引用其他集群时可删除

##### 设计思路

验证路由规则引用的是其他集群时，不会误拦截当前集群的删除。

##### 前提数据准备

1. 创建集群 `cluster_unref` 和 `cluster_referred`。
2. 通过 `PUT /open-api/v1/global-route-rules` 设置 global 路由规则引用 `cluster_referred`。

##### 执行步骤

1. 发送 DELETE 请求到 `/open-api/v1/clusters/cluster_unref`。
2. 验证返回 200，并再次查询返回 404。

##### 请求参数

URI：`cluster_unref`

Global 路由规则准备请求：

```json
{
    "rules": [
        {
            "name": "global-other-ref",
            "Cond": "default_t()",
            "targets": [
                {"ClusterName": "cluster_referred", "Model": "", "Weight": 100}
            ],
            "fallbacks": []
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**Data**：返回被删除集群对象

---

## 11. 依赖与数据准备

1. 模型提供商类型 `provider_type` 取值可参考 `/model-provider-types`。
2. 创建集群会级联写入实例池、子集群、lb_matrices，测试后需清理。
3. 测试环境数据库需包含产品线初始化数据，以支持 `{product_name}.{cluster_name}` 实例池命名。
4. 涉及 global/entity/apikey 路由规则的用例，需先通过对应接口写入规则，测试结束后清理规则或清空 `rules` 数组，避免影响其他用例。

## 12. 注意事项

1. v0.3.0 已删除 `/clusters/{cluster_name}/ready`、`/model-providers`、`/models`。
2. 返回中不应出现 `ready`、`sub_clusters`、`scheduler`、`Instance.tags`、`llm_config.service_name`、`llm_config.group`。
3. `basic.retries.max_retry_in_cluster` 对应底层 `max_retry_in_subcluster`。
4. 更新时不支持修改 `sub_clusters`/`scheduler`，需通过 `instance_pool` 调整实例。
5. 测试环境 `SkipTokenValidate=true`，无需认证头。
