# Provider 模型发现接口调整摘要

## 1. 背景

`providers.md` 中 2.6 节“触发模型发现”当前与 Provider 资源强耦合：

- 端点路径含 `provider_name`；
- 需读取 provider 的 `model_endpoint`、`instance_pool`、`keys`、`model_protocols`；
- 发现成功后需写回 provider 的 `models` 字段。

该设计不利于在创建 Provider 前预先探测可用模型，且将认证、持久化等职责混入发现逻辑。

## 2. 目标

将模型发现能力拆分为独立的无状态工具接口：

- 端点改为 `/providers/tools/discover-models`；
- 调用方显式传入 `model_protocol`、`schema`、`addr`、`port`，可选传入 `uri`、`apikey`；
- 接口只负责调用模型列表接口并返回模型名列表，不读写 Provider 资源。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及文档 | `design-docs/api-define/OpenAPI接口定义/providers.md`、`design-docs/sys-design/` 相关文档 |
| 涉及模块 | `endpoints/openapi_v1/providers`、模型协议解析相关模块（如 `model/iprotocol`） |
| 数据迁移 | 无 |
| 接口兼容 | 旧接口 `/providers/{provider_name}/discover-models` 不再保留 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 无状态工具接口 | 不与 provider 资源绑定，支持创建 provider 前预探模型。 |
| 显式参数 | 连接信息由调用方传入，避免接口内部读取 provider 配置。 |
| 认证头按协议生成 | 根据 `model_protocol` 自动生成认证头：`openai` 使用 `Authorization: Bearer {apikey}`，`anthropic` 使用 `x-api-key: {apikey}`。 |
| 不写回 provider | 返回结果不持久化，需要回填时调用方再调用 `PATCH /providers/{provider_name}`。 |
| 单协议 | 一次请求指定一个 `model_protocol`，按该协议选择解析器。 |

## 5. 改动点

| 仓库 | 文件/模块 | 修改内容 |
|------|-----------|----------|
| `ai-gateway-api` | `design-docs/api-define/OpenAPI接口定义/providers.md` | 更新 2.6 节端点、参数、执行逻辑、返回示例 |
| `ai-gateway-api` | `design-docs/sys-design/接口层设计文档.md` | 更新接口清单中的模型发现端点 |
| `ai-gateway-api` | `design-docs/sys-design/总体设计文档.md` | 更新 Provider 管理功能描述 |
| `ai-gateway-api` | `design-docs/sys-design/模型层设计文档.md` | 更新 `DiscoverModels` 与 Provider/协议章节描述 |
| `ai-gateway-api` | `design-docs/sys-design/details/provider与cluster概念分离.md` | 更新接口列表、实现要点、待确认事项 |
| `ai-gateway-api` | `design-docs/sys-design/details/Claude协议转发支持.md` | 更新模型发现端点引用 |
| `ai-gateway-api` | `endpoints/openapi_v1/providers/` | 新增/调整 `discover-models` handler，改为无状态实现 |
| `ai-gateway-api` | `endpoints/openapi_v1/endpoints.go` | 注册 `/providers/tools/discover-models`，移除旧路由 |
| `ai-gateway-api` | 模型协议解析模块 | 提供按 `model_protocol` 选择解析器的能力 |
