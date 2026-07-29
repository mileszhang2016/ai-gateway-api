# /global-route-rules

## 1. 数据模型

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

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `enabled` | bool | 是否启用该Global路由表 | `true`（启用）、`false`（禁用），**默认值为`true`** |
| `rules` | array | 规则列表，按顺序匹配 | - |

**rules元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 规则名称，用于日志和监控 | 同一`global`路由表内唯一 |
| `Cond` | string | BFE条件表达式 | 命中则使用该规则 |
| `targets` | array | 转发目标列表 | 至少1个元素 |
| `fallbacks` | array | 降级目标列表 | 允许为空 |

**targets元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `ClusterName` | string | 后端集群名称 | - |
| `Model` | string | 模型名称，空字符串表示透传原始模型 | - |
| `Weight` | int | 权重 | 单个target时为100，多个target时总和为100 |

**fallbacks元素结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `ClusterName` | string | 后端集群名称 | - |
| `Model` | string | 模型名称，空字符串表示透传原始模型 | - |

**约束**

- `rules` 中规则的 `name` 在同一 `global` 路由表内必须唯一
- 每个规则的 `targets` 中所有 `Weight` 之和必须等于100
- 每个规则的 `fallbacks` 中的元素 `ClusterName` 不能为空

---

## 2. 接口清单

### 2.1 更新Global路由表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新Global路由表 | - |
| 端点 | /global-route-rules | - |
| 版本 | v1 | - |
| method | PUT | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| enabled | bool | 是否启用该Global路由表 | N | **默认值为`true`** |
| rules | array | 规则列表 | Y | 同5.1数据模型中rules结构 |

**HTTP BODY参数示例**

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

**执行逻辑**

1. 校验参数合法性
2. 若未传入 `enabled`，使用默认值 `true`
3. 校验 `rules` 中规则名称是否重复
4. 校验每个规则的 `targets` 权重总和是否为100
5. 校验每个规则的 `Cond` 表达式是否合法
6. 在持久化存储中，将 `type` 固定设置为 `global`，`owner` 固定设置为 `global`
7. 写入或替换 `global_default` 路由表
8. 返回结果

**返回数据（Data内容）**

字段同5.1数据模型。

**成功返回示例**

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
    }
}
```

---

### 2.2 查询Global路由表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询Global路由表 | - |
| 端点 | /global-route-rules | - |
| 版本 | v1 | - |
| method | GET | - |

**返回数据（Data内容）**

字段同5.1数据模型。若Global路由表不存在，返回Data为null。

---

