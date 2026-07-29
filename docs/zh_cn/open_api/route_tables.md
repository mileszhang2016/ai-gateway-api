# 路由表列表

## 1 查询路由表列表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询路由表列表 | |
| 端点 | /route-tables | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### Query 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| sort_by | string | 排序字段 | N | - |
| sort_order | string | 排序方向 | N | asc/desc，默认desc |
| type | string | 按路由表类型过滤 | N | 可选值：global、entity、api_key |
| owner | string | 按所有者标识过滤 | N | 支持精确匹配 |
| enabled | bool | 按启用状态过滤 | N | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/route-tables?page=1&page_size=20" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| list | []object | 路由表列表 | 元素字段见下方Route Table结构 |
| pagination | object | 分页信息 | - |
| pagination.page | int | 当前页码 | - |
| pagination.page_size | int | 每页条数 | - |
| pagination.total | int | 总条数 | - |

#### list 对象字段说明
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| type | string | 路由表类型 | global（全局路由表）、entity（Entity路由表）、api_key（API-Key路由表） |
| owner | string | 所有者标识 | global类型固定为global；entity类型为entity_id；api_key类型为apikey_id |
| enabled | bool | 是否启用该路由表 | - |

#### 成功返回参数示例
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
    },
    "WorkMode": "ModeNormal"
}
```

---

## 数据模型

### Route Table 结构
| 字段 | 类型 | 说明 |
| - | - | - |
| type | string | 路由表类型，可选值：global、entity、api_key |
| owner | string | 所有者标识；global类型固定为global；entity类型为entity_id；api_key类型为apikey_id |
| enabled | bool | 是否启用该路由表 |
