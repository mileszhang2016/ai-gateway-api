# server_cert_conf 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出 BFE 服务端证书配置 | 供 BFE `protocol/server_cert_conf` 使用，包含证书名到证书/密钥文件路径的映射 |
| 端点 | `/configs/protocol/server_cert_conf` | - |
| Method | GET | - |
| 鉴权 | `FeatureCert + ActionExport` | - |

## 2. 请求参数

**Query 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | 否 | 上次返回的版本号，用于增量同步 | 可选；无强制格式/长度校验；为空或未传时按首次拉取处理 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/protocol/server_cert_conf?version=00010101000000" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

### 3.1 顶层结构

与 BFE 动态配置文件 `server_cert_conf` 格式保持一致：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "Version": "00010101000000",
        "Config": {
            "Default": "cert_demo",
            "CertConf": {
                "cert_demo": {
                    "ServerCertFile": "tls_conf_00010101000000/cert_demo/server.crt",
                    "ServerKeyFile": "tls_conf_00010101000000/cert_demo/server.key",
                    "OcspResponseFile": ""
                }
            }
        }
    },
    "WorkMode": "ModeNormal"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| Version | string | 配置版本号 |
| Config | object | 证书配置集合 |

### 3.2 Config 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| Default | string | 默认证书名称；全局必须有且只有一个默认证书 |
| CertConf | object | 证书名到证书配置的映射，key 为 `cert_name` |

### 3.3 单个证书配置（ServerCertConf）结构

| 字段 | 类型 | 说明 |
|------|------|------|
| ServerCertFile | string | 服务端证书文件路径；导出时会追加版本前缀 `tls_conf_{version}` |
| ServerKeyFile | string | 服务端私钥文件路径；导出时会追加版本前缀 `tls_conf_{version}` |
| OcspResponseFile | string | OCSP 响应文件路径；当前未使用，固定为空字符串 |

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
