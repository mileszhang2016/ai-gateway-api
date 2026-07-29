# 整体说明

## API规范

一个典型API包括如下部分：
- 接口名：一句话描述接口名
- 基本信息
- 输入参数
- 返回参数

### 基本信息

**URL格式说明**
- API遵守一般RESTful风格，API的URL格式：
    - http://ai_gateway_api:port/open-api/{ver}/{endpoint}?{arg=value}
    - 例子：http://127.0.0.1:8086/open-api/v1/api-keys
- API URL各部分说明
    - ai_gateway_api: 服务器地址，一般是域名或者IP地址
    - port：API服务的端口号
    - ver：当前API的版本
    - endpoint：REST风格的资源路径
    - arg：参数名
    - value：参数值

若无特殊说明，后续文档的具体API只描述Endpoint.

**method说明**

若无特殊说明，method 遵循如下约定：
- GET：获取一条或多条
- POST: 创建
- PUT：全量更新
- PATCH：部分更新
- DELETE：删除

举例：

| 项目  | 值  | 说明 | 
| - | - | - |
| 含义	| 创建API-Key | |
| 端点 | /api-keys | |
| 版本 | v1 |  |
| method | POST | - |

### 通用Query参数（列表接口）

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| page | int | 页码 | N | 默认1 |
| page_size | int | 每页条数 | N | 默认20，最大100 |
| sort_by | string | 排序字段 | N | - |
| sort_order | string | 排序方向 | N | asc/desc，默认desc |

### 请求参数

请求参数分为：

- URI参数
- Query参数
- Body内容

举例：

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - | 
| description | string | API-Key描述 |  Y | - |
| quota_plan | object | 配额计划 | N | 同Quota Plan结构 |
| quota_plan.unlimited | bool | 是否无限配额 | N | 默认true |

HTTP BODY中参数示例
```json
{
    "description": "BFE项目测试Key",
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```


### 返回数据

所有API的返回值格式为：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": json_object,
    "WorkMode": "current mode"
}
```

- ErrNum: 返回码
    - 200：调用成功
    - 402：没有调用权限
    - 404：查询/修改/删除不存在的对象
    - 409：资源冲突（如存在依赖关系、循环引用等）
    - 422：参数不合法
    - 500：其他业务逻辑错误
    - 555：产品线内重复（如API-Key描述重复）
    - 556：全局重复（如entity-type或entity-name重复）
- Data: 返回的数据结构
    - 调用成功时，返回json格式的数据
    - 调用失败时，返回null
- ErrMsg: 文本消息
    - 调用成功时，ErrMsg是success或空串
    - 调用失败时，ErrMsg是相关的错误信息
- WorkMode: 控制台工作模式


举例：
```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "apikey-001",
        "key": "ak-2v8x9k3m7p",
        "description": "BFE项目测试Key",
        "enabled": true
    },
    "WorkMode": "ModeNormal"
}
```

说明：API文档中API的返回结果，仅给出Data部分。


## 鉴权机制
- API使用Token机制鉴权
- 访问时在HTTP Authorization Header中加入Token
- 鉴权详细机制见 [用户和鉴权](auth.md)
- Token的使用示例：

```
curl http://127.0.0.1:8086/open-api/v1/api-keys -H "Authorization: Token YOUR_TOKEN"
```
