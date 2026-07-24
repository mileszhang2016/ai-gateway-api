# ALB实例池

> **v1.2 优化说明：** 接口从 `/alb-pools` 改为 `/alb-pool`，移除 `{instance_pool_name}` URI 参数，使用配置文件中的默认 AI 网关实例池名称。移除列表、创建、删除接口。

## 1 获取默认 AI 网关实例池详情

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义	| 获取默认 AI 网关实例池的详情 ||
| 端点	| /alb-pool ||
| 动作	| GET | - |

### 输入参数
无 URI 参数，无 Body 参数。

### 处理逻辑
1. 从配置文件 `RunTime` 中读取 `DefaultAIInstancePoolName`（默认值：`BFE.aipool`）
2. 查询实例池，若不存在则返回错误
3. 返回实例池详情（包含实例列表）

### 返回数据(Data内容)

| 参数名 | 类型 | 参数含义 |
|--------|------|----------|
| name | string | 实例池完整名称 |
| instances | []Instance | 实例列表 |
| instances[].hostname | string | 实例所在主机名 |
| instances[].ip | string | 实例 IP 地址 |
| instances[].weight | int | 实例权重，范围 [0,100] |
| instances[].ports | map[string]int | 实例端口，至少包含 Default |
| instances[].tags | map[string]string | 实例标签 |

#### 成功返回数据示例

```json
{
    "name": "BFE.aipool",
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "key": "value"
            }
        }
    ]
}
```

---

## 2 更新默认 AI 网关实例池

### 基本信息
| 项目  | 值  | 说明 |
| - | - | - |
| 含义 |	全量更新默认 AI 网关实例池的实例列表 | 该更新是全量更新，不支持仅添加部分数据 |
| 端点 |	/alb-pool ||
| method |	PATCH | - |

### 输入参数

无 URI 参数。

#### Body参数

| 参数名 | 类型 |参数含义 | 必填 | 补充描述 |
| - | -  | - | - | - |
| instances | []Instance | 实例列表 | Y | 全量替换当前实例池中的实例 |
| instances[].hostname | string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instances[].ip | string | 实例 IP 地址 | Y | |
| instances[].weight | int | 实例权重，范围 [0,100] | Y | |
| instances[].ports | map[string]int | 实例端口 | Y | 至少包含 Default 端口 |
| instances[].tags | map[string]string | 实例标签 | N | |

**移除字段说明：**

| 移除字段 | 原因 |
|----------|------|
| `name` | 实例池名称由配置项 `DefaultAIInstancePoolName` 提供，无需在请求中传入 |
| `epp_server` | AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置 |

#### HTTP BODY中参数示例

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "key": "value"
            }
        }
    ]
}
```

### 返回数据(Data内容)
同 GET 接口。