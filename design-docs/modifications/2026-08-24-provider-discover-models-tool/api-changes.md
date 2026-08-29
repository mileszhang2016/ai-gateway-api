# Provider 模型发现接口：API 变更说明

## 1. 变更范围

修改 `design-docs/api-define/OpenAPI接口定义/providers.md` 的 2.6 节“触发模型发现”。

## 2. 变更内容

### 2.1 端点变更

| 项目 | 变更前 | 变更后 |
|------|--------|--------|
| 端点 | `/providers/{provider_name}/discover-models` | `/providers/tools/discover-models` |
| 类型 | Provider 资源操作 | 无状态工具接口 |

### 2.2 Method / 版本

| 项目 | 值 |
|------|-----|
| method | POST |
| 版本 | v1 |

### 2.3 输入参数变更

由 URI 参数 `provider_name` 改为 Body 参数：

| 参数名 | 类型 | 参数含义 | 必填 | 合法性条件 |
|--------|------|----------|------|------------|
| `model_protocol` | string | 模型访问协议 | Y | 枚举值：`openai`、`anthropic` |
| `schema` | string | 请求协议 | Y | `http`、`https` |
| `addr` | string | 目标实例地址 | Y | 类型为 [Hostname](./00-common.md#1-主机名hostname) |
| `port` | int | 目标实例端口 | Y | 类型为 [Port](./00-common.md#3-网络端口port) |
| `uri` | string | 模型列表接口 URI | N | 为空时默认使用 `/v1/models` | 非空时须以 `/` 开头 |
| `apikey` | string | 调用模型列表接口的 API Key | N | - | 非空时长度 1-512 字符 |

### 2.4 返回数据

返回数据由“更新后的 provider 对象中的 `models`”改为只返回模型名列表：

| 参数名 | 类型 | 参数含义 |
|--------|------|----------|
| `models` | []string | 发现到的模型名列表 |

### 2.5 执行逻辑变更

1. 若 `uri` 为空，默认使用 `/v1/models`；构造请求 URL：`{schema}://{addr}:{port}{uri}`。
2. 若 `apikey` 非空，根据 `model_protocol` 生成认证头：
   - `openai`：`Authorization: Bearer {apikey}`
   - `anthropic`：`x-api-key: {apikey}`
3. 携带认证头（若有）调用第三方模型列表接口。
4. 根据 `model_protocol` 选择对应的响应解析器（如 `openai`、`anthropic`），提取模型名列表。
5. 返回模型名列表。

**不再执行以下操作**：

- 读取 provider 配置；
- 将发现结果写回 provider 的 `models` 字段；
- 更新 provider 的 `update_time`。

## 3. 请求 / 响应示例

### 3.1 请求示例

```http
POST /v1/providers/tools/discover-models
Content-Type: application/json

{
    "model_protocol": "openai",
    "schema": "https",
    "addr": "api.deepseek.com",
    "port": 443,
    "uri": "/v1/models",
    "apikey": "sk-aaaaaaaaaaaa"
}
```

### 3.2 响应示例

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "models": ["deepseek-chat", "deepseek-coder", "deepseek-reasoner"]
    }
}
```

## 4. 影响面

| 影响项 | 说明 |
|--------|------|
| OpenAPI 文档 | 需要更新 `providers.md` 2.6 节 |
| 代码实现 | 需要新增/调整 `endpoints/openapi_v1/providers/` 对应 handler 与协议解析逻辑 |
| 调用方 | 原先通过 `provider_name` 触发的调用方需要改为传入完整连接参数 |
| 数据持久化 | 本接口不再写回 provider，如需回填，调用方需再调用 `PATCH /providers/{provider_name}` |
