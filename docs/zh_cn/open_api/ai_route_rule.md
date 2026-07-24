# AI路由规则配置

## 1 全量更新AI路由规则

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 端点|	/ai-route-rules ||
| 动作|	PATCH | |
| 含义|	全量更新AI路由规则 |  |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
无。

#### Body参数
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| forward_rules | []ForwardRule | 高级路由规则列表 | N | 为空代表清空规则。详见[表1：ForwardRule数据结构](#forward_rule) |
| basic_forward_rules | []BasicForwardRule | 基础路由规则列表 | N | 为空代表清空规则。详见[表2：BasicForwardRule数据结构](#basic_forward_rule) |

<a id="forward_rule">表1：ForwardRule 数据结构</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| name | string | 路由规则名称 | Y | 需要以字母或数字开头，允许数字、大小写字母、下划线、中划线组合且长度大于1，小于128 |
| description | string | 路由规则描述 | N | |
| expression | string | 条件表达式 | Y | BFE 条件表达式语法 |
| cluster_name | string | 目标集群名称 | Y | 转发到的目标集群名称 |

<a id="basic_forward_rule">表2：BasicForwardRule 数据结构</a>

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| host_names | []string | 域名列表 | N | 如 `["*.example.com", "api.example.com"]` |
| paths | []string | 路径列表 | N | 如 `["/aaa", "/abc"]` |
| cluster_name | string | 目标集群名称 | Y | 转发到的目标集群名称 |
| description | string | 路由规则描述 | N | |

##### 请求示例
```
curl -X PATCH "http://{api_server}/open-api/v1/ai-route-rules" -d @data.json -H "Authorization:Token token_string" -H "Content-Type:application/json"
```

##### 输入参数示例
data.json如下：
```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "description": "路由到集群1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "cluster-demo"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
            "cluster_name": "cluster-demo",
            "description": "基础路由"
        }
    ]
}
```

### 返回数据(Data内容)

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| forward_rules | []ForwardRule | 高级路由规则列表 | Y | |
| basic_forward_rules | []BasicForwardRule | 基础路由规则列表 | Y | |
| forward_cases_code | int | 路由用例代码 | N | 仅在存在路由用例时返回 |

#### 返回数据示例

```json
{
    "forward_rules": [
        {
            "name": "rule1",
            "description": "路由到集群1",
            "expression": "req_host_in(\"api.example.com\")",
            "cluster_name": "cluster-demo"
        }
    ],
    "basic_forward_rules": [
        {
            "host_names": ["*.example.com"],
            "paths": ["/api"],
            "cluster_name": "cluster-demo",
            "description": "基础路由"
        }
    ]
}
```

##### 错误返回
| **错误码** | 错误信息 |
| ---------------------- | -------- |
| 422 | 参数不合法 |
| 511 | 数据库异常 |

## 2 获取AI路由规则列表

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义 |	获取AI路由规则列表 ||
| 端点 |	/ai-route-rules  ||
| 动作 |	GET | - |
| Content-Type | application/x-www-form-urlencoded | - |

### 输入参数

#### Body参数
无。

#### 请求示例
```
curl "http://api-server:port/open-api/v1/ai-route-rules" -H "Authorization:Token token_string" -H "Content-Type:application/x-www-form-urlencoded"
```

### 返回数据(Data内容)
| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| forward_rules | []ForwardRule | 高级路由规则列表 | Y | 为空代表无规则 |
| basic_forward_rules | []BasicForwardRule | 基础路由规则列表 | Y | 为空代表无规则 |
| forward_cases_code | int | 路由用例代码 | N | 仅在存在路由用例时返回 |

#### 返回数据示例
同 全量更新AI路由规则 接口返回数据示例。

##### 错误返回
| **错误码** | 错误信息 |
| ---------------------- | -------- |
| 422 | 参数不合法 |
| 511 | 数据库异常 |