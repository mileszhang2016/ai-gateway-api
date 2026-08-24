# Provider 与 Cluster 概念分离——变更摘要

## 1. 背景

当前 `ai-gateway-api` 的 `/clusters` 资源同时承担了两种职责：

1. **Provider 职责**：下游模型提供商标识、后端实例池、模型端点、API Key 明文等。
2. **Cluster 职责**：路由/转发策略、连接与超时参数、健康检查等 BFE 集群参数。

这种混合导致：

- 同一 provider 被多个 cluster 引用时，`instance_pool`、`keys`、`model_endpoint` 重复配置。
- cluster 接口暴露 API Key 明文，且无法通过引用复用 provider 的 key。
- `model-prices` 的 `provider` 字段语义不清。
- 新增 provider 类型或协议时，cluster 模型不断膨胀。

## 2. 目标

在 `ai-gateway-api` 控制面中把 `provider`（模型提供方）与 `cluster`（转发集群）解耦：

- 建立独立的 `/providers` 资源，集中管理接入能力（实例池、模型、key、协议）。
- `cluster` 只保留“转发策略”相关配置，通过引用 provider 获取后端能力。
- `model-prices` 的 `provider` 字段必须引用已存在的 provider。

## 3. 范围

- **涉及面**：`ai-gateway-api` 控制面及其 OpenAPI/InnerAPI。
- **不涉及面**：`bfe` 数据面；`ai-gateway-api` 生成的 BFE 相关配置（集群、子集群、实例池、`AIConf` 等）保持原有结构和语义不变。
- **数据面影响**：无；变化仅在 `ai-gateway-api` 内部做“provider → 老配置”的转换。

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| Provider 独立成资源 | 新增 `/providers` CRUD 及 `/providers/{name}/discover-models` 接口。 |
| Cluster 瘦身 | 移除 `instance_pool`、`llm_config.model_endpoint`、`llm_config.provider_type`；`llm_config.keys` 改为 `{name, weight}` 引用。 |
| Model-prices 引用收紧 | `provider` 字段必须引用 `/providers` 中已存在的 provider。 |
| 配置顺序 | `/providers` → `/model-prices` → `/clusters` → 路由规则。 |
| 数据迁移 | 提供自动迁移脚本，将存量 cluster 中的 provider 相关字段抽取为 provider 记录。 |
| API 版本策略 | OpenAPI 层面为破坏性变更；建议以产品大版本或 API 版本切换方式发布，必要时保留只读兼容层。 |

## 5. 关联文档

- 详细设计：`design-changes.md`
- 接口变更：`api-changes.md`
- 上游设计来源：`document-ai-gateway/迭代系统设计/v0.5/provider和cluster概念分离/provider和cluster概念分离-设计与实施方案.md`

## 6. 实施阶段

| 阶段 | 内容 | 预计周期 |
|------|------|----------|
| 1 | 设计与文档冻结 | 1 周 |
| 2 | 控制面数据模型与存储 | 1-2 周 |
| 3 | 接口与校验 | 1-2 周 |
| 4 | BFE 配置生成适配 | 1 周 |
| 5 | 迁移与测试 | 1-2 周 |
| 6 | 发布 | 1 周 |
