# gslb 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 GSLB 调度配置 | 供 BFE `gslb.data` 使用，按指定 BFE 集群返回各集群到子集群的调度权重 |
| 端点 | `/configs/gslb_data/gslb` | - |
| Method | GET | - |
| 鉴权 | `FeatureRoute + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |
| bfe_cluster | string | 是 | 指定 BFE 集群名称 | 必填；长度至少为 1；必须对应实际存在且在 GSLB 调度中被引用的 BFE 集群名称 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/gslb_data/gslb?version=00010101000000&bfe_cluster=bfe-cluster-1" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

与 BFE 动态配置文件 `gslb.data` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "Ts": "00010101000000",
        "Hostname": "gslb.manual.com",
        "Clusters": {
            "my-cluster": {
                "sub-cluster-1": 100
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号 |
| Ts | string | 时间戳，与 Version 一致 |
| Hostname | string | GSLB 调度器域名，固定为 `gslb.manual.com` |
| Clusters | object | 集群调度配置，key 为集群名称 |

### 3.2 Clusters 结构

```json
{
    "Clusters": {
        "my-cluster": {
            "sub-cluster-1": 100
        }
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| my-cluster | object | 集群内各子集群的调度权重 |
| sub-cluster-1 | int | 子集群调度权重；当前每个集群只包含一个子集群，权重固定为 100 |

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
