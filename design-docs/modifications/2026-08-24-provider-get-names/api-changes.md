# Provider 全量名称查询接口：API 变更说明

## 1. 变更范围

在 `design-docs/api-define/OpenAPI接口定义/providers.md` 新增 2.7 节。

## 2. 新增接口

### 2.1 基本信息

| 项目 | 值 |
|------|-----|
| 含义 | 获取所有 Provider 名称列表 |
| 端点 | `/providers/actions/get-provider-names` |
| 版本 | v1 |
| method | GET |

### 2.2 输入参数

无。

### 2.3 返回数据

Data 字段：

| 参数名 | 类型 | 参数含义 |
|--------|------|----------|
| `names` | []string | 所有 Provider 名称列表，按字典序升序排列 |

### 2.4 执行逻辑

1. 查询所有 provider 的 `name` 字段。
2. 去重（数据库唯一约束已保证）、按字典序升序排列。
3. 返回 `{ names: [...] }`。

### 2.5 请求 / 响应示例

**请求**

```http
GET /v1/providers/actions/get-provider-names
```

**响应**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "names": ["anthropic", "deepseek", "openai"]
    }
}
```

## 3. 影响面

| 影响项 | 说明 |
|--------|------|
| OpenAPI 文档 | 新增 2.7 节 |
| 代码实现 | 新增 handler、model、storage 方法 |
| 调用方 | 新增全量名称获取能力；现有 `GET /providers` 分页接口行为不变 |
