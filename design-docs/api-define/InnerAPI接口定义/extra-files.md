# extra_files 接口

## 1. 接口信息

| 项目 | 值 | 说明 |
|------|------|------|
| 含义 | 导出额外文件内容 | 供 BFE 或 conf-agent 拉取证书、密钥等二进制/文本文件 |
| 端点 | `/configs/extra_files/{filename}` | `filename` 为文件路径标识 |
| Method | GET | - |
| 鉴权 | `FeatureExtraFile + ActionExport` | - |

## 2. 请求参数

**Path 参数**

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| filename | string | 是 | 要导出的文件标识 | 必填；长度至少为 1；必须对应实际存在的额外文件 |

**请求示例**

```shell
curl -X GET "http://api-server:port/inner-api/v1/configs/extra_files/cert_demo/server.crt" \
  -H "Authorization:Token TOKEN_STRING"
```

## 3. 返回数据结构

本接口**直接返回文件原始内容**，不包装为标准 `ErrNum/ErrMsg/Data` JSON 结构。

返回的 `Content-Type` 由文件内容决定，常见为：

- 证书/密钥文件：`application/x-pem-file` 或 `text/plain`
- 二进制文件：`application/octet-stream`

## 4. 错误返回示例

当文件不存在时，接口返回 HTTP 404 及标准错误响应：

```json
{
    "ErrNum": 404,
    "ErrMsg": "record not exist"
}
```

---
