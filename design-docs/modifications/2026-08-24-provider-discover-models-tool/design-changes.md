# Provider 模型发现接口：设计变更说明

## 1. 概述

### 1.1 背景

当前 `/providers/{provider_name}/discover-models` 是一个 Provider 资源操作接口：

- 需要 provider 已存在；
- 从 provider 配置中读取 `model_endpoint`、`instance_pool`、`keys`、`model_protocols`；
- 使用 provider 的第一个 key 构造认证请求；
- 发现成功后写回 provider 的 `models` 字段。

该设计将“模型发现”与“Provider 资源”强耦合，不利于在创建 Provider 前预探模型列表，也增加了认证头风格与 key 管理的复杂度。

### 1.2 目标

将模型发现能力拆分为无状态工具接口 `/providers/tools/discover-models`：

- 调用方显式传入连接参数与协议类型；
- 接口只负责调用模型列表接口并解析模型名；
- 不依赖、不修改任何 Provider 资源状态。

### 1.3 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `endpoints/openapi_v1/providers`、模型协议解析相关模块 |
| 涉及文档 | `design-docs/api-define/OpenAPI接口定义/providers.md`、`design-docs/sys-design/` 相关文档 |
| 数据迁移 | 无 |

---

## 2. 接口设计

### 2.1 端点与 Method

| 项目 | 值 |
|------|-----|
| 端点 | `/providers/tools/discover-models` |
| method | POST |
| 版本 | v1 |

### 2.2 请求参数

请求体（Body）：

| 参数名 | 类型 | 参数含义 | 必填 | 合法性条件 |
|--------|------|----------|------|------------|
| `model_protocol` | string | 模型访问协议 | Y | 枚举：`openai`、`anthropic` |
| `schema` | string | 请求协议 | Y | `http`、`https` |
| `addr` | string | 目标实例地址 | Y | Hostname |
| `port` | int | 目标实例端口 | Y | Port |
| `uri` | string | 模型列表接口 URI | N | 为空时默认 `/v1/models`；非空时须以 `/` 开头 |
| `apikey` | string | 调用模型列表接口的 API Key | N | 非空时长度 1-512 字符 |

### 2.3 响应参数

Data 字段：

| 参数名 | 类型 | 参数含义 |
|--------|------|----------|
| `models` | []string | 发现到的模型名列表 |

### 2.4 执行流程

```text
调用方
  ↓ POST /providers/tools/discover-models
API Handler
  ↓ 校验参数
Model Discovery Manager
  ↓ 根据 model_protocol 选择 Parser 与认证头风格
HTTP Client
  ↓ 若 apikey 非空则携带认证头，请求 {schema}://{addr}:{port}{uri}
第三方模型列表接口
  ↓ 返回原始响应
Parser
  ↓ 提取模型名列表
返回 { models: [...] }
```

---

## 3. 关键设计决策

| 决策 | 说明 |
|------|------|
| 无状态工具接口 | 不与 provider 资源绑定，便于在创建 provider 前预探可用模型。 |
| 显式传入连接参数 | 避免接口内部读取 provider 配置，逻辑更清晰，也便于测试。 |
| 认证头按协议生成 | 根据 `model_protocol` 自动生成认证头：`openai` 使用 `Authorization: Bearer {apikey}`，`anthropic` 使用 `x-api-key: {apikey}`。 |
| 不写回 provider | 发现结果仅返回，不持久化；需要回填时调用方再调用 `PATCH /providers/{provider_name}`。 |
| 单协议解析 | 一次请求只指定一个 `model_protocol`，按该协议选择对应解析器。 |

---

## 4. 实现要点

### 4.1 Handler 层

- 新增/调整 `endpoints/openapi_v1/providers/discover_models.go`（或现有 handler）。
- 绑定路由 `/providers/tools/discover-models`。
- 请求体校验：参数必填、枚举值、hostname、port、uri 格式。

### 4.2 Model 层

- 复用/新增模型列表响应解析器：
  - `openai`：解析 `/v1/models` 标准响应，取 `data[].id`。
  - `anthropic`：解析 Anthropic 模型列表响应。
- 建议的解析器接口：

  ```go
  type ModelDiscoveryParser interface {
      Parse(data []byte) ([]string, error)
  }
  ```

- HTTP 调用逻辑放在 model 层或独立 manager 中，便于单元测试。

### 4.3 路由注册

在 `endpoints/openapi_v1/endpoints.go` 中注册新路由，并移除/废弃旧路由 `/providers/{provider_name}/discover-models`。

### 4.4 设计文档同步

更新 `design-docs/sys-design/` 中所有引用旧端点或旧执行逻辑的文档，确保与 OpenAPI 定义一致：

- `接口层设计文档.md`：更新接口清单。
- `总体设计文档.md`：更新 Provider 管理功能描述。
- `模型层设计文档.md`：更新 `DiscoverModels` 描述与 Provider/协议章节。
- `details/provider与cluster概念分离.md`：更新接口列表、实现要点、待确认事项。
- `details/Claude协议转发支持.md`：更新模型发现端点引用。

---

## 5. 待确认事项

| 事项 | 建议 |
|------|------|
| 认证信息 | 已新增 `apikey` 参数，认证头风格由 `model_protocol` 决定。后续如需支持自定义认证头，可再扩展 `auth_style` 参数。 |
| 错误码 | 第三方接口不可达、返回非 200、解析失败时的错误码需统一约定，建议分别返回 502/422。 |
| 旧接口兼容 | 是否保留旧接口一段时间？建议直接替换并同步更新前端调用。 |
