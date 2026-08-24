# cluster_table 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 BFE 集群表配置 | 供 BFE `cluster_table.data` 使用，包含所有集群的后端实例（RS）列表 |
| 端点 | `/configs/gslb_data/cluster_table` | - |
| Method | GET | - |
| 鉴权 | `FeatureRoute + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/gslb_data/cluster_table?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

与 BFE 动态配置文件 `cluster_table.data` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "Config": {
            "my-cluster": {
                "sub-cluster-1": [
                    {
                        "Name": "backend-1",
                        "Addr": "10.0.0.1",
                        "Port": 8080,
                        "Weight": 50
                    },
                    {
                        "Name": "backend-2",
                        "Addr": "10.0.0.2",
                        "Port": 8080,
                        "Weight": 50
                    }
                ]
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号 |
| Config | object | 集群表配置，key 为集群名称 |

### 3.2 Config 结构

```json
{
    "Config": {
        "my-cluster": {
            "sub-cluster-1": [ /* 后端实例列表 */ ]
        }
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| my-cluster | object | 集群内的子集群映射；当前每个集群只包含一个子集群 |
| sub-cluster-1 | array | 后端实例列表 |

### 3.3 后端实例（BackendConf）结构

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | 实例名称，对应所属 provider `instance_pool[].name` |
| Addr | string | 实例 IP 地址；IPv6 地址会被 `[]` 包裹 |
| Port | int | 实例端口，对应所属 provider `instance_pool[].port` |
| Weight | int | 实例权重，对应所属 provider `instance_pool[].weight`；`0` 表示该实例不接收流量 |

## 4. 配置未变化返回示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": null,
    "WorkMode": "ModeNormal"
}
```

---
