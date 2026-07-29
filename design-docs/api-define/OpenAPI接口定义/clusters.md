# /clusters

## 1. 数据模型

```json
{
    "name": "my-cluster",
    "description": "示例集群",
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {"Default": 8080}
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
            {"key": "gpt-4", "value": "deepseek-chat"}
        ],
        "key": "sk-xxxxxxxxxxxx",
        "provider_type": "deepseek"
    }
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 集群名 | 全局唯一 |
| `description` | string | 集群描述信息 | - |
| `instance_pool` | []Instance | 实例列表 | 系统自动据此创建实例池和子集群 |
| `basic` | object | 基本参数 | 见下方 表：连接设置、表：重试设置、表：超时设置 |
| `sticky_sessions` | object | 会话保持 | 见下方 表：会话保持 |
| `passive_health_check` | object | 被动健康检查 | 见下方 表：被动健康检查 |
| `llm_config` | object | AI LLM 服务配置 | 见下方 表：LLM配置 |

**Instance 结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `hostname` | string | 实例所在主机名 | 无 DNS 时可填写 IP 地址 |
| `ip` | string | 实例 IP 地址 | - |
| `weight` | int | 实例权重 | 范围 [0,100] |
| `ports` | map[string]int | 实例端口 | 至少包含 `Default` 端口 |

**表：连接设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_idle_conn_per_rs| int | 连接池| N | 每个BFE实例，为集群中每个RS维持的空闲长连接数。一般情况下，无需特别维持，设置为0 。<br/>设置为非0时，可以提升转发性能 |
| cancel_on_client_close| bool |  连接是否级联关闭 | N | 设置为true时，当客户端关闭连接后，BFE同时关闭对应RS的连接 <br/>设置为false时，当客户端关闭连接后，BFE按默认策略关闭对应RS的连接 |

**表：重试设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_retry_in_cluster| int |  同一个集群内重试次数 | N | 底层对应 `max_retry_in_subcluster`，**默认值为2** |

**表：会话保持**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| enabled| bool |  是否开启会话保持 | N | **默认false**。为true时开启会话保持；为false时关闭 |
| hash_strategy| string |  会话保持策略  | N | CLIENT_IP_ONLY，根据client ip做会话保持 <br/>	CLIENT_ID_ONLY，根据请求中header做会话保持(默认值) <br>	CLIENT_ID_PERFERED，优先基于特定header，如果请求中没有对应header，则使用client ip|
| hash_header| string |  指定CLIENT_ID使用的header | N | 当使用cookie作为会话保持的哈希key时，数据格式为Cookie:${key} |

**表：超时设置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| timeout_conn_serv| int |  连接后端超时(ms)| N |  |
| timeout_response_header| int |  读后端响应头部超时(ms)| N |  |
| timeout_readbody_client| int |  读请求body超时(ms)| N |  |
| timeout_read_client_again| int |  与用户的长连接超时(ms) | N |  |
| timeout_write_client| int |  写响应超时(ms)| N | - |

**表：被动健康检查**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| failnum| int |  进入健康检查的失败次数阈值 | N | 连续转发失败多次后，BFE进入健康检查状态，对下游RS发起探活；**默认值为3** |
| interval| int |  连续健康检查的时间间隔 | N | 单位ms；**默认值为1000** |
| host| string |  健康检查请求的域名| N | 域名后的部分；为空时使用 `instance_pool` 中第一个实例的 `hostname` |
| uri| string |  健康检查请求的URI  | N | **默认值为 `/`** |
| statuscode| int |  期望的健康检查返回码 | N | 如果需要忽略返回码，此处可以填0；**默认值为0** |

**表：LLM配置**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| model_endpoint| object |  模型列表端点配置 | N | 用于调用第三方AI模型提供商的模型列表接口，具体字段见下方 表：Endpoint；未设置时使用默认值 |
| models| []string |  支持的模型名称列表 | Y | 指定该集群支持的AI模型名称 |
| model_mappings| []object |  模型名称映射 | N | 用于将用户请求的模型名映射为后端实际使用的模型名，具体字段见下方 表：模型映射 |
| key| string |  服务认证密钥 | N | 用于后端AI服务的认证 |
| provider_type| string |  AI模型提供商类型 | N | 取值如：deepseek、openai、qwen 等。数据来自 `/model-provider-types` |

> **注意：** `enable` 字段已移除，设置 `llm_config` 时默认开启 AI 网关能力。

**表：Endpoint**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| schema| string |  请求协议 | N |  取值为 http、https；**默认值为 https** |
| uri| string |  请求URI | N |  **默认值为 `/v1/models`** |
| headers| map[string]string |  请求头参数 | N | 自定义请求头 |

**表：模型映射**

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| key| string |  用户请求的模型名 | Y |  |
| value| string |  映射后的实际模型名 | Y |  |

**约束**

- `sub_clusters` 与 `scheduler` 为系统内部自动生成数据，不再对外暴露；每个集群只包含一个子集群，调度设置固定为 `GSLB_BLACKHOLE=0`。
- `llm_config.enable` 字段已移除，设置 `llm_config` 时默认开启 AI 网关能力。
- 删除集群时自动级联清理关联的实例池和子集群。

## 2. 接口清单

### 2.1 创建集群

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建集群（一键创建实例池 + 子集群 + 绑定） | - |
| 端点 | /clusters | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| name| string |  集群名 | Y | 集群名必须全局唯一 |
| description| string |  集群描述信息| N |  |
| instance_pool| []Instance |  实例列表 | Y | 系统自动据此创建实例池和子集群 |
| instance_pool[].hostname| string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instance_pool[].ip| string | 实例 IP 地址 | Y | |
| instance_pool[].weight| int | 实例权重，范围 [0,100] | Y | |
| instance_pool[].ports| map[string]int | 实例端口 | Y | 至少包含 Default 端口 |
| basic| object |  基本参数| N | AI网关场景默认推荐值：connection.max_idle_conn_per_rs=0、cancel_on_client_close=false；retries.max_retry_in_cluster=2；timeouts.timeout_conn_serv=50000、timeout_response_header=50000、timeout_readbody_client=30000、timeout_read_client_again=30000、timeout_write_client=60000 |
| sticky_sessions| object |  会话保持| N | AI网关场景默认推荐值：enabled=false；若开启，hash_strategy=CLIENT_ID_ONLY、hash_header为空 |
| passive_health_check| object |  被动健康检查| N | AI网关场景默认推荐值：failnum=3、interval=1000ms、host为空（使用instance_pool首个实例hostname）、uri="/"、statuscode=0 |
| llm_config| object |  AI LLM服务配置| Y | 见 10.1 数据模型中表：LLM配置 |

**HTTP BODY参数示例**

```json
{
    "name": "my-cluster",
    "description": "示例集群",
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
        "models": [
            "deepseek-chat",
            "deepseek-coder"
        ],
        "model_mappings": [
            {
                "key": "gpt-4",
                "value": "deepseek-chat"
            }
        ],
        "key": "sk-xxxxxxxxxxxx",
        "provider_type": "deepseek"
    }
}
```

**执行逻辑**

创建集群时，系统自动执行以下步骤：

1. 校验请求参数（`name`、`instance_pool`、`llm_config` 等必填项；`basic`、`sticky_sessions`、`passive_health_check` 若未传则使用 AI 网关场景默认推荐值）
2. 若 `llm_config` 不为 nil，内部自动设置 `enable = true`
3. 创建集群
4. 自动创建实例池（名称格式：`{product_name}.{cluster_name}`）
5. 自动创建子集群（名称：`{cluster_name}`，绑定实例池）
6. 自动绑定子集群到集群

**返回数据（Data内容）**

| 参数名 | 类型 |参数含义 |
| - | -  | - |
| name | string | 集群名 |
| description | string | 集群描述信息 |
| instance_pool | []Instance | 实例列表 |
| llm_config | object | LLM 配置 |
| basic | object | 基本参数 |
| sticky_sessions | object | 会话保持 |
| passive_health_check | object | 被动健康检查 |

**成功返回示例**
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "name": "my-cluster",
        "description": "示例集群",
        "instance_pool": [
            {
                "hostname": "backend-1",
                "ip": "10.0.0.1",
                "weight": 50,
                "ports": {"Default": 8080}
            }
        ],
        "llm_config": { "...": "..." },
        "basic": { "...": "..." },
        "sticky_sessions": { "...": "..." },
        "passive_health_check": { "...": "..." }
    }
}
```

### 2.2 集群列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 所有集群列表 | - |
| 端点 | /clusters | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**返回数据（Data内容）**

数组，单元素同创建接口。

### 2.3 集群详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 单个集群详情 | - |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**返回数据（Data内容）**

同创建接口。

### 2.4 更新集群基本配置

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 更新集群基本信息 | 可编辑描述信息、Basic配置段、sticky_sessions配置段、healthcheck配置段、instance_pool、llm_config |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**输入参数（Body）**
可修改字段含义同创建接口。若传入 `instance_pool` 字段，系统会自动同步更新对应的实例池。

> **注意：** `sub_clusters` 与 `scheduler` 为系统内部自动生成，更新时不支持手动修改，请通过 `instance_pool` 调整实例。

**HTTP BODY参数示例**

```json
{
    "name": "my-cluster",
    "description": "更新后的集群描述",
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
            "uri": "/v1/models"
        },
        "models": ["deepseek-chat"],
        "provider_type": "deepseek"
    }
}
```

**返回数据（Data内容）**

同创建接口。

### 2.5 删除集群

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除集群（自动级联清理关联的实例池和子集群） | - |
| 端点 | /clusters/{cluster_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cluster_name | string | 集群名字 | Y | - |

**执行逻辑**
删除集群时，系统自动执行以下级联清理：
1. 解绑集群关联的子集群
2. 删除子集群
3. 删除子集群关联的实例池
4. 删除集群

**返回数据（Data内容）**

同创建接口。

---

