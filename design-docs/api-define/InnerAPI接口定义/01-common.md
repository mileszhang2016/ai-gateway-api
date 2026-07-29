# 通用说明

## 1. 基本信息

| 项目 | 值 | 说明 |
|------|------|------|
| URL 格式 | `http://api_server:port/inner-api/v1/{endpoint}?{arg=value}` | 例：`http://127.0.0.1:8086/inner-api/v1/configs/mod-api-key` |
| 鉴权方式 | McUserProbe | HTTP Header 中间件鉴权 |
| 版本控制 | 基于 version 参数 | 配置未变化时返回空响应 |

## 2. 返回值格式

所有 Inner-API 的返回值格式：

```json
{
    "ErrNum": 200,
    "Data": {},
    "ErrMsg": "success",
    "WorkMode": "current mode"
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| ErrNum | int | 返回码，200 表示成功 |
| Data | object | 返回的数据结构，配置未变化时返回 null |
| ErrMsg | string | 文本消息，成功时为 "success"，失败时为错误信息 |
| WorkMode | string | 控制台工作模式 |

## 3. 版本控制机制

所有导出接口支持 `version` 查询参数：

- **首次拉取**：不传 `version` 或传空字符串，返回完整配置
- **增量拉取**：传入上次返回的 `version`，若配置未变化返回 `Data: null`
- **配置变化**：返回新的配置数据和新的 `version`

---

