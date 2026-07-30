# ai-route 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 AI 网关路由配置 | 供 BFE 的 `mod_ai_route` 模块执行 apikey → entity → global 三级路由查找 |
| 端点 | `/configs/ai-route` | - |
| Method | GET | - |
| 鉴权 | `FeatureRoute + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/ai-route?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

与 BFE 动态配置文件 `ai_route.data` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "RouteRules": {
            "apikey_ak_user_a": {/* API-Key 路由表 */},
            "entity_dept_ai": {/* Entity 路由表 */},
            "global_default": {/* Global 路由表 */}
        },
        "ApikeyRouteTableBindings": {
            "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"]
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号，由版本控制机制生成 |
| RouteRules | object | 所有路由表的集合，key 为 `<type>_<owner>`，保证全局唯一 |
| ApikeyRouteTableBindings | object | API-Key 到路由表查找顺序的映射 |

**说明**：仅导出 `route_rules.enabled = true` 的路由表。若 Global、API-Key 或 Entity 路由表的 `enabled = false`，则不生成该路由表，也不会加入 `ApikeyRouteTableBindings`。

### 3.2 RouteRules 结构（路由表集合）

```json
{
    "RouteRules": {
        "apikey_ak_user_a": {
            "type": "apikey",
            "owner": "ak_user_a",
            "rules": [
                {
                    "name": "user_a-deepseek",
                    "Cond": "req_host_in(\"api.example.org\")",
                    "targets": [
                        {
                            "ClusterName": "cluster_deepseek_a",
                            "Model": "deepseek-v4-pro",
                            "Weight": 70
                        },
                        {
                            "ClusterName": "cluster_deepseek_b",
                            "Model": "deepseek-v4-pro",
                            "Weight": 30
                        }
                    ],
                    "fallbacks": [
                        {
                            "ClusterName": "cluster_deepseek_c",
                            "Model": "deepseek-v3.2"
                        }
                    ]
                }
            ]
        },
        "entity_dept_ai": {
            "type": "entity",
            "owner": "dept_ai",
            "rules": [
                {
                    "name": "dept_ai-default",
                    "Cond": "default_t()",
                    "targets": [
                        {
                            "ClusterName": "cluster_dept_ai",
                            "Model": "",
                            "Weight": 100
                        }
                    ],
                    "fallbacks": []
                }
            ]
        },
        "global_default": {
            "type": "global",
            "owner": "global",
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
    }
}
```

**RouteTable 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 路由表类型，枚举值：`apikey`、`entity`、`global` |
| owner | string | 路由表属主。`apikey` 类型为 API-Key 标识；`entity` 类型为 Entity 名称；`global` 类型固定为 `global` |
| rules | array | 该路由表下的规则列表，按顺序匹配 |

**RouteRule 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 规则名称，用于日志、监控和问题定位 |
| Cond | string | BFE 条件表达式，命中则使用该规则 |
| targets | array | 转发目标列表 |
| fallbacks | array | 降级目标列表，允许为空 |

**Target 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传请求原始模型 |
| Weight | int | 权重，单个 target 时为 100，多个 target 时总和必须为 100 |

**Fallback 字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ClusterName | string | 后端集群名称 |
| Model | string | 模型名称，空字符串表示透传请求原始模型 |

### 3.3 ApikeyRouteTableBindings 结构（绑定关系）

```json
{
    "ApikeyRouteTableBindings": {
        "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"],
        "ak_user_b": ["apikey_ak_user_b", "entity_dept_ai", "global_default"]
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | API-Key 字符串 |
| value | []string | 该 API-Key 对应的路由表查找顺序，典型顺序为 `apikey → entity → global` |

## 4. 成功返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "RouteRules": {
            "apikey_ak_user_a": {
                "type": "apikey",
                "owner": "ak_user_a",
                "rules": [
                    {
                        "name": "user_a-deepseek",
                        "Cond": "req_host_in(\"api.example.org\")",
                        "targets": [
                            {
                                "ClusterName": "cluster_deepseek_a",
                                "Model": "deepseek-v4-pro",
                                "Weight": 70
                            },
                            {
                                "ClusterName": "cluster_deepseek_b",
                                "Model": "deepseek-v4-pro",
                                "Weight": 30
                            }
                        ],
                        "fallbacks": [
                            {
                                "ClusterName": "cluster_deepseek_c",
                                "Model": "deepseek-v3.2"
                            }
                        ]
                    },
                    {
                        "name": "user_a-default",
                        "Cond": "default_t()",
                        "targets": [
                            {
                                "ClusterName": "cluster_default",
                                "Model": "",
                                "Weight": 100
                            }
                        ],
                        "fallbacks": []
                    }
                ]
            },
            "entity_dept_ai": {
                "type": "entity",
                "owner": "dept_ai",
                "rules": [
                    {
                        "name": "dept_ai-default",
                        "Cond": "default_t()",
                        "targets": [
                            {
                                "ClusterName": "cluster_dept_ai",
                                "Model": "",
                                "Weight": 100
                            }
                        ],
                        "fallbacks": []
                    }
                ]
            },
            "global_default": {
                "type": "global",
                "owner": "global",
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
        },
        "ApikeyRouteTableBindings": {
            "ak_user_a": ["apikey_ak_user_a", "entity_dept_ai", "global_default"],
            "ak_user_b": ["apikey_ak_user_b", "entity_dept_ai", "global_default"]
        }
    },
    "WorkMode": "ModeNormal"
}
```

## 5. 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---

