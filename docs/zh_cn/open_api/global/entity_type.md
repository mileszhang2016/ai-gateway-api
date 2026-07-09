# Entity-Type

## 1 创建Entity-Type

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 创建Entity类型定义 | |
| 端点 | /entity-types | |
| 版本 | v1 | |
| method | POST | 创建资源 |
| Content-Type | application/json | - |

### 输入参数

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | 全局唯一，1-32字符，仅含小写字母、数字、下划线、连字符 |
| description | string | 类型描述 | N | - |
| level | int | 层级级别 | Y | 取值范围1-5，数字越小级别越高 |

#### 请求示例
```shell
curl -X POST "http://api-server:port/open-api/v1/entity-types" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "type_name": "dep",
    "description": "一级部门",
    "level": 1
}
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| type_name | string | 类型名 | - |
| description | string | 描述 | - |
| level | int | 层级级别 | - |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "type_name": "dep",
        "description": "一级部门",
        "level": 1,
        "create_time": 1716883200
    },
    "WorkMode": "ModeNormal"
}
```

---

## 2 查询Entity-Type列表

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询Entity-Type列表 | |
| 端点 | /entity-types | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### Query 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/entity-types?page=1&page_size=20" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
| 参数名 | 类型 | 参数含义 | 补充描述 |
| - | - | - | - |
| list | []object | Entity-Type列表 | - |
| pagination | object | 分页信息 | - |
| pagination.page | int | 当前页码 | - |
| pagination.page_size | int | 每页条数 | - |
| pagination.total | int | 总条数 | - |

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "list": [
            {
                "type_name": "dep",
                "description": "一级部门",
                "level": 1,
                "create_time": 1716883200
            }
        ],
        "pagination": {
            "page": 1,
            "page_size": 20,
            "total": 1
        }
    },
    "WorkMode": "ModeNormal"
}
```

---

## 3 查询单个Entity-Type

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 查询单个Entity-Type | |
| 端点 | /entity-types/{type_name} | |
| 版本 | v1 | |
| method | GET | 获取资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

#### 请求示例
```shell
curl -X GET "http://api-server:port/open-api/v1/entity-types/dep" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
同创建接口返回。

#### 成功返回参数示例
同创建接口返回示例。

---

## 4 更新Entity-Type

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 更新Entity-Type描述 | |
| 端点 | /entity-types/{type_name} | |
| 版本 | v1 | |
| method | PATCH | 部分更新资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

#### Body 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| description | string | 类型描述 | N | - |

#### 请求示例
```shell
curl -X PATCH "http://api-server:port/open-api/v1/entity-types/dep" \
  -d data.json \
  -H "Authorization:Token TOKEN_STRING" \
  -H 'Content-Type:application/json'
```

data.json 如下：
```json
{
    "description": "更新后的描述"
}
```

### 返回数据(Data内容)
同创建接口返回。

#### 成功返回参数示例
同创建接口返回示例。

---

## 5 删除Entity-Type

### 基本信息
| 项目  | 值  | 说明 | 
| - | - | - |
| 含义 | 删除Entity-Type | |
| 端点 | /entity-types/{type_name} | |
| 版本 | v1 | |
| method | DELETE | 删除资源 |
| Content-Type | application/json | - |

### 输入参数

#### URI 参数
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| type_name | string | 类型名 | Y | - |

#### 请求示例
```shell
curl -X DELETE "http://api-server:port/open-api/v1/entity-types/dep" \
  -H "Authorization:Token TOKEN_STRING"
```

### 返回数据(Data内容)
Data为null。

#### 成功返回参数示例
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

**约束**
- 若该类型下存在Entity，返回ErrNum=409

---

## 数据模型

### Entity-Type 数据模型
| 字段 | 类型 | 说明 | 可能取值 |
| - | - | - | - |
| type_name | string | 类型标识 | 全局唯一，1-32字符，仅含小写字母、数字、下划线、连字符 |
| description | string | 类型描述 | 自定义 |
| level | int | 层级级别 | 取值范围1-5，数字越小级别越高 |
| create_time | int64 | 创建时间 | Unix时间戳（秒） |
