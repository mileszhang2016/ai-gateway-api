# /alb-pool

## 1. 数据模型

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
      }
    }
  ]
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `name` | string | 实例池完整名称 | 如 `BFE.aipool` |
| `instances` | []Instance | 实例列表 | - |

**Instance 结构**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `hostname` | string | 实例所在主机名 | 无 DNS 时可填写 IP 地址 |
| `ip` | string | 实例 IP 地址 | - |
| `weight` | int | 实例权重 | 范围 [0,100] |
| `ports` | map[string]int | 实例端口 | 至少包含 `Default` 端口 |

**约束**

- AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置。
- 实例池名称由配置项 `DefaultAIInstancePoolName` 提供，请求中无需传入 `name`。

## 2. 接口清单

### 2.1 获取默认 AI 网关实例池详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取默认 AI 网关实例池的详情 | - |
| 端点 | /alb-pool | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**执行逻辑**

1. 从配置文件 `RunTime` 中读取 `DefaultAIInstancePoolName`（默认值：`BFE.aipool`）。
2. 查询实例池，若不存在则返回错误。
3. 返回实例池详情（包含实例列表）。

**返回数据（Data内容）**

字段同 [1. 数据模型](#1-数据模型)。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "name": "BFE.aipool",
        "instances": [
            {
                "hostname": "127.0.0.1",
                "ip": "127.0.0.1",
                "weight": 1,
                "ports": {
                    "Default": 8080
                }
            }
        ]
    }
}
```

### 2.2 更新默认 AI 网关实例池

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 全量更新默认 AI 网关实例池的实例列表 | 该更新是全量更新，不支持仅添加部分数据 |
| 端点 | /alb-pool | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| instances | []Instance | 实例列表 | Y | 全量替换当前实例池中的实例 |
| instances[].hostname | string | 实例所在主机名 | Y | 无 DNS 时可填写 IP 地址 |
| instances[].ip | string | 实例 IP 地址 | Y | - |
| instances[].weight | int | 实例权重 | Y | 范围 [0,100] |
| instances[].ports | map[string]int | 实例端口 | Y | 至少包含 `Default` 端口 |

**约束**

- 实例池名称由配置项 `DefaultAIInstancePoolName` 提供，请求中无需传入 `name`。
- AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置。

**HTTP BODY参数示例**

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

**执行逻辑**

1. 校验请求参数合法性。
2. 使用配置项 `DefaultAIInstancePoolName` 定位默认 AI 网关实例池。
3. 全量替换实例池中的实例列表。
4. 返回更新后的实例池详情。

**返回数据（Data内容）**

同 [2.1 获取默认 AI 网关实例池详情](#21-获取默认-ai-网关实例池详情)。

**成功返回示例**

同 [2.1 获取默认 AI 网关实例池详情](#21-获取默认-ai-网关实例池详情)。

---
