# Global 路由表

## 1 全量更新Global路由表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 全量更新Global路由表 | |
| 端点 | /global-route-rules | |
| 版本 | v1 | |
| method | PUT | 全量更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| enabled | bool | 是否启用该Global路由表 | N | 默认值为true |
| rules | array | 规则列表 | Y | 同Global Route Rules数据模型中rules结构 |

**约束**
- `rules` 中规则的 `name` 在同一 `global` 路由表内必须唯一
- 每个规则的 `targets` 中所有 `Weight` 之和必须等于100
- 每个规则的 `fallbacks` 中的元素 `ClusterName` 不能为空

#### 请求示例
```shell
curl -X PUT "http://api-server:port/open-api/v1/global-route-rules" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "enabled": true,
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
            "fallbacks": [
                {
                    "ClusterName": "cluster_global_fallback",
                    "Model": ""
                }
            ]
        }
    ]
}
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| enabled | bool | 是否启用该Global路由表 | - |
| rules | array | 规则列表 | - |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "enabled": true,
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
                "fallbacks": [
                    {
                        "ClusterName": "cluster_global_fallback",
                        "Model": ""
                    }
                ]
            }
        ]
    },
    "WorkMode": "ModeNormal"
}
```

---

## 2 查询Global路由表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询Global路由表 | |
| 端点 | /global-route-rules | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数
无。

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/global-route-rules" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
字段同全量更新接口返回。若Global路由表不存在，返回Data为null。

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "enabled": true,
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
                "fallbacks": [
                    {
                        "ClusterName": "cluster_global_fallback",
                        "Model": ""
                    }
                ]
            }
        ]
    },
    "WorkMode": "ModeNormal"
}
```

---

## 数据模型

### Global Route Rules 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| enabled | bool | 是否启用该Global路由表，默认值为true |
| rules | []AiRouteRule | 规则列表，按顺序匹配 |

### AiRouteRule 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| name | string | 规则名称，同一路由表内唯一 |
| Cond | string | BFE条件表达式，命中则使用该规则 |
| targets | []AiRouteTarget | 转发目标列表，至少1个元素 |
| fallbacks | []AiRouteFallback | 降级目标列表，允许为空 |

### AiRouteTarget 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传原始模型 |
| Weight | int | 权重，单个target时为100，多个target时总和为100 |

### AiRouteFallback 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传原始模型 |

### 约束
- 每个 `AiRouteRule` 的 `targets` 中所有 `Weight` 之和必须等于100
- 每个 `AiRouteRule` 的 `fallbacks` 中的元素 `ClusterName` 不能为空
