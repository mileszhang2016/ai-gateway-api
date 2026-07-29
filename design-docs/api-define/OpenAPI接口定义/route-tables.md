# /route-tables

## 1. 数据模型

```json
{
  "type": "global",
  "owner": "global",
  "enabled": true
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `type` | string | 路由表类型 | `global`（全局路由表）、`entity`（Entity路由表）、`api_key`（API-Key路由表） |
| `owner` | string | 所有者标识 | `global` 类型固定为 `global`；`entity` 类型为 `entity_id`；`api_key` 类型为 `apikey_id` |
| `enabled` | bool | 是否启用该路由表 | `true`（启用）、`false`（停用） |

---

## 2. 接口清单

### 2.1 查询路由表列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 查询路由表列表 | - |
| 端点 | /route-tables | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| sort_by | string | 排序字段 | N | - |
| sort_order | string | 排序方向 | N | asc/desc，默认desc |
| type | string | 按路由表类型过滤 | N | 可选值：`global`、`entity`、`api_key` |
| owner | string | 按所有者标识过滤 | N | 支持精确匹配 |
| enabled | bool | 按启用状态过滤 | N | - |

**返回数据（Data内容）**

Data为数组，元素字段同6.1数据模型。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "list": [
            {
                "type": "global",
                "owner": "global",
                "enabled": true
            },
            {
                "type": "entity",
                "owner": "ent-001",
                "enabled": false
            },
            {
                "type": "api_key",
                "owner": "apikey-001",
                "enabled": true
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 3
        }
    }
}
```

---


---

