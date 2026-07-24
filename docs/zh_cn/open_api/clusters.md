# 集群

> **优化说明：**
> - `sub_clusters` 字段替换为 `instance_pool`（实例列表），系统自动创建实例池和子集群
> - `scheduler` 字段移除，系统自动生成（默认子集群权重=100，GSLB_BLACKHOLE=0）
> - `llm_config.enable` 字段移除，默认开启（`true`）
> - 删除集群时自动级联清理关联的实例池和子集群
> - 废弃 `/clusters/{cluster_name}/sub-clusters` 接口（子集群绑定已自动处理）

## 1 创建集群

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
|端点 |	/clusters | |
|动作 |	POST  | |
|含义 |	创建集群（一键创建实例池 + 子集群 + 绑定 + 自动调度） | - |

### 输入参数

#### Body参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| name| string |  集群名 | Y | 集群名必须全局唯一 |
| description| string |  集群描述信息| N |  |
| instance_pool| []Instance |  实例列表 | Y | 系统自动据此创建实例池和子集群 |
| instance_pool[].hostname| string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instance_pool[].ip| string | 实例 IP 地址 | Y | |
| instance_pool[].weight| int | 实例权重，范围 [0,100] | Y | |
| instance_pool[].ports| map[string]int | 实例端口 | Y | 至少包含 Default 端口 |
| instance_pool[].tags| map[string]string | 实例标签 | N | |
| basic| object |  基本参数| Y |  |
| basic.connection| object |  连接管理| Y | 内容见 [表：连接设置](#connection) |
| basic.retries| object |  重试次数| Y | 内容见 [表：重试设置](#retries) |
| basic.buffers| object |  缓冲设置| Y |  |
| basic.buffers.req_write_buffer_size| int |  接受请求的缓冲字节数| Y |  |
| basic.timeouts| object |  超时设置| Y |  内容见 [表：超时设置](#timeouts) |
| basic.protocol| string |  集群支持的协议| Y |  取值为http、https。 |
| sticky_sessions| object |  会话保持| Y | 内容见 [表：会话保持](#sticky_sessions)|
| passive_health_check| object |  被动健康检查| Y | 具体字段见 [表：被动健康检查](#passive_health_check) |
| llm_config| object |  AI LLM服务配置| N | 开启AI网关能力时必填，具体字段见 [表：LLM配置](#llm_config) |

**移除字段说明：**

| 移除字段 | 说明 |
|----------|------|
| `sub_clusters` | 替换为 `instance_pool`，不再手动指定子集群名称，系统自动创建 |
| `scheduler` | 自动生成：默认子集群权重=100，GSLB_BLACKHOLE=0 |
| `llm_config.enable` | 默认 `true`，不再需要手动传入 |

<a id="connection">表：连接设置</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_idle_conn_per_rs| int | 连接池| Y | 每个BFE实例，为集群中每个RS维持的空闲长连接数。一般情况下，无需特别维持，设置为0 。<br/>设置为非0时，可以提升转发性能 |
| cancel_on_client_close| bool |  连接是否级联关闭 | Y | 设置为true时，当客户端关闭连接后，BFE同时关闭对应RS的连接 <br/>设置为false时，当客户端关闭连接后，BFE按默认策略关闭对应RS的连接 |

<a id="retries">表： 重试设置</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| max_retry_in_subcluster| int |  同一个子集群内重试次数| Y |  |
| max_retry_cross_subcluster| int |  跨子集群重试次数| Y | - |

<a id="sticky_sessions">表：会话保持</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| session_sticky_type| string |  会话保持的粒度 | Y | INSTANCE，实例级会话保持 <br/>	SUB_CLUSTER，子集群级别会话保持|
| hash_strategy| string |  会话保持策略  | N | CLIENT_IP_ONLY，根据client ip做会话保持 <br/>	CLIENT_ID_ONLY，根据请求中header做会话保持(默认值) <br>	CLIENT_ID_PERFERED，优先基于特定header，如果请求中没有对应header，则使用client ip|
| hash_header| string |  指定CLIENT_ID使用的header | N |	当使用cookie作为会话保持的哈希key时，数据格式为Cookie:${key} |

<a id="timeouts">表：超时设置</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| timeout_conn_serv| int |  连接后端超时(ms)| Y |  |
| timeout_response_header| int |  读后端响应头部超时(ms)| Y |  |
| timeout_readbody_client| int |  读请求body超时(ms)| Y |  |
| timeout_read_client_again| int |  与用户的长连接超时(ms) | Y |  |
| timeout_write_client| int |  写响应超时(ms)| Y | - |

<a id="passive_health_check">表: 被动健康检查</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| failnum| int |  进入健康检查的失败次数阈值 | Y | 连续转发失败多次后，BFE进入健康检查状态，对下游RS发起探活 |
| interval| int |  连续健康检查的时间间隔 | Y | 单位ms |
| host| string |  健康检查请求的域名| Y | 域名后的部分 |
| uri| string |  健康检查请求的URI  | Y |  |
| statuscode| int |  期望的健康检查返回码 | Y | 如果需要忽略返回码，此处可以填0 |

<a id="llm_config">表: LLM配置</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| service_name| string |  服务名称 | Y | 最长255字符 |
| group| string |  分组名称 | N | 最长255字符 |
| model_endpoint| object |  模型列表端点配置 | Y | 用于调用第三方AI模型提供商的模型列表接口，具体字段见 [表：Endpoint](#endpoint) |
| models| []string |  支持的模型名称列表 | Y | 指定该集群支持的AI模型名称 |
| model_mappings| []object |  模型名称映射 | N | 用于将用户请求的模型名映射为后端实际使用的模型名，具体字段见 [表：模型映射](#model_mapping) |
| key| string |  服务认证密钥 | N | 用于后端AI服务的认证 |
| provider_type| string |  AI模型提供商类型 | N | 取值如：deepseek、openai、qwen 等 |

> **注意：** `enable` 字段已移除，设置 `llm_config` 时默认开启 AI 网关能力。

<a id="endpoint">表: Endpoint</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| schema| string |  请求协议 | Y |  取值为 http、https |
| uri| string |  请求URI | Y |  例如：/v1/models |
| headers| map[string]string |  请求头参数 | N | 自定义请求头 |

<a id="model_mapping">表: 模型映射</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| key| string |  用户请求的模型名 | Y |  |
| value| string |  映射后的实际模型名 | Y |  |

#### HTTP BODY中参数示例

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
            },
            "tags": {
                "region": "bj"
            }
        },
        {
            "hostname": "backend-2",
            "ip": "10.0.0.2",
            "weight": 50,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "region": "sh"
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
            "max_retry_in_subcluster": 2,
            "max_retry_cross_subcluster": 0
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
        "session_sticky_type": "INSTANCE",
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 10,
        "host": "health.example.com",
        "uri": "/health",
        "statuscode": 200
    },
    "llm_config": {
        "service_name": "deepseek-service",
        "group": "llm-group",
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

### 处理逻辑

创建集群时，系统自动执行以下步骤：

1. 校验请求参数（instance_pool、basic 等必填项）
2. 若 llm_config 不为 nil，内部自动设置 `enable = true`
3. 自动生成 scheduler：
   - 从配置读取 `DefaultAIClusterName`（如 `BFE-AI_product.szyf`）
   - 设置 `{cluster_name}: 100`，`GSLB_BLACKHOLE: 0`
4. 创建集群
5. 自动创建实例池（名称格式：`{product_name}.{cluster_name}`）
6. 自动创建子集群（名称：`{cluster_name}`，绑定实例池）
7. 自动绑定子集群到集群

### 返回数据(Data内容)

| 参数名 | 类型 |参数含义 |
| - | -  | - |
| name | string | 集群名 |
| description | string | 集群描述信息 |
| ready | bool | 集群是否就绪 |
| instance_pool | []Instance | 实例列表 |
| sub_clusters | []string | 子集群列表 |
| scheduler | object | 自动生成的调度配置 |
| llm_config | object | LLM 配置 |
| basic | object | 基本参数 |
| sticky_sessions | object | 会话保持 |
| passive_health_check | object | 被动健康检查 |

#### 成功返回数据示例

```json
{
    "name": "my-cluster",
    "description": "示例集群",
    "ready": false,
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {"Default": 8080},
            "tags": {"region": "bj"}
        }
    ],
    "sub_clusters": ["my-cluster"],
    "scheduler": {
        "BFE-AI_product.szyf": {
            "my-cluster": 100,
            "GSLB_BLACKHOLE": 0
        }
    },
    "llm_config": { "...": "..." },
    "basic": { "...": "..." },
    "sticky_sessions": { "...": "..." },
    "passive_health_check": { "...": "..." }
}
```

## 2 集群列表

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 端点	 | /clusters ||
| 动作	 | GET ||
| 含义	 | 所有集群列表 | - |

### 输入参数
无

### 返回数据(Data内容)
数组，单元素同创建接口

## 3 集群详情

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 端点 |	/clusters/{cluster_name} ||
| method |	GET  ||
| 含义 |	单个集群详情 | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| cluster_name | string | 集群名字|  Y | - |

### 返回数据(Data内容)
同创建接口（含 `instance_pool` 和 `scheduler` 字段）

## 4 更新集群基本配置

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义 |	更新集群基本信息 | 可编辑描述信息、Basic配置段、sticky_sessions配置段、healthcheck配置段、instance_pool、llm_config |
| 端点 |	/clusters/{cluster_name} | |
| method |	PATCH | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| cluster_name | string | 集群名字 |  Y | - |

#### Body参数

可修改字段含义同创建接口。若传入 `instance_pool` 字段，系统会自动同步更新对应的实例池。

> **注意：** `scheduler` 由系统自动管理，更新时不支持手动修改。`sub_clusters` 字段已不再支持，请使用 `instance_pool`。

#### HTTP BODY中参数示例

```json
{
    "name": "my-cluster",
    "description": "更新后的集群描述",
    "ready": false,
    "instance_pool": [
        {
            "hostname": "backend-1",
            "ip": "10.0.0.1",
            "weight": 100,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "region": "bj"
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
            "max_retry_in_subcluster": 2,
            "max_retry_cross_subcluster": 0
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
        "session_sticky_type": "INSTANCE",
        "hash_strategy": "CLIENT_ID_ONLY",
        "hash_header": "Cookie:USERID"
    },
    "passive_health_check": {
        "interval": 1000,
        "failnum": 10,
        "host": "health.example.com",
        "uri": "/health",
        "statuscode": 200
    },
    "llm_config": {
        "service_name": "deepseek-service",
        "model_endpoint": {
            "schema": "https",
            "uri": "/v1/models"
        },
        "models": ["deepseek-chat"],
        "provider_type": "deepseek"
    }
}
```

### 返回数据(Data内容)
同创建接口

## 5 删除集群

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义 | 	删除集群（自动级联清理关联的实例池和子集群） ||
| 端点 | 	/clusters/{cluster_name} ||
| method | 	DELETE | - |

### 输入参数
#### URI 参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| cluster_name | string | 集群名字|  Y | - |

### 处理逻辑
删除集群时，系统自动执行以下级联清理：
1. 解绑集群关联的子集群
2. 删除子集群
3. 删除子集群关联的实例池
4. 删除集群

### 返回数据(Data内容)
同创建接口

## 6 集群就绪状态获取

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 端点 | 	/clusters/{cluster_name}/ready | |
| method | 	GET  ||
| 含义 | 	获取集群是否就绪的状态(可以承接线上流量)  | 当前，集群默认是就绪的 |

### 输入参数
#### URI 参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| cluster_name | string | 集群名字|  Y | - |

### 返回数据(Data内容)

```json
{
    "name": "my-cluster",
    "ready": false
}
```

## 7 获取AI模型提供商列表

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义	| 获取 AI 模型提供商列表 | |
| 端点 | /model-providers| |
| 版本 | v1 |  |
| method | GET | - |
| Content-Type | application/x-www-form-urlencoded | - |

### 输入参数

#### Body 参数
无

#### 请求示例
curl -X GET 'http://api-server:port/open-api/v1/model-providers' -H 'Authorization:Token token_string'

### 返回数据(Data内容)
状态码200为成功。

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        "deepseek",
        "qwen",
        "openai"
    ],
    "Version": "",
    "Sign": "",
    "WorkMode": "ModeNormal"
}
```

#### 错误返回
无。

## 8 获取AI模型列表

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义	| 获取 AI 模型列表 | |
| 端点 | /models| |
| 版本 | v1 |  |
| method | POST | - |
| Content-Type | application/json | - |

### 输入参数

#### Body 参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| schema | string | http、https | Y |  |
| uri | string | 请求的uri | N | 路径前面可以有/，也可以无/。例如：/models或者models。 |
| hosts | []string | 请求的ip、port组合或者域名。 | Y | 支持ipv4、ipv6。ipv4："1.1.1.1:8080" ipv6:"[2001:db8::1]:8080"|
| headers | map[string]string | 请求的header参数列表 | N | - |
| provider_type | string | AI模型提供商类型 | N | 取值为：deepseek，openai，qwen。 |

#### 请求示例
curl -X POST 'http://api-server:port/open-api/v1/models' -H 'Authorization:Token token_string' -d "@data.json" -H 'Content-Type:application/json'

data.json示例：
```json
{
    "schema":"http",
    "uri":"/models",
    "hosts":["1.1.1.1:8080", "[2001:db8::1]:8080","www.a.com","www.b.com:8080"],
    "headers":{
        "Content-type":"application/json"
    },
    "provider_type":"deepseek"
}
```

### 返回数据(Data内容)
状态码200为成功。返回数据为列表结构。

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| id | string | 模型ID | Y |  |
| name | string | 名称 | N | -|

#### 成功返回参数示例
```json
[{
    "id": "model1",
    "name": "Model 1"
}]
```

#### 错误返回
| **错误码** | 错误信息 |
| ---------------------- | -------- |
| 422 | 参数不合法|
| 513 | 调用AI模型提供商API出错|