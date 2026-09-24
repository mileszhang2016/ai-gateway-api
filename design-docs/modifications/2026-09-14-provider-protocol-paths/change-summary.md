# Provider 协议路径改写（protocol_paths）变更摘要

## 1. 背景

对主流 provider 上游端点的调研（详见总体设计方案 §1.1）表明：各家路径各不相同，且**同一 provider 的不同协议（OpenAI 兼容 / Anthropic 兼容）挂在不同前缀下**——百炼 `/compatible-mode/v1` vs `/apps/anthropic`、Kimi Code `/coding/v1` vs `/coding`、火山 `/api/v3` vs `/api/coding`。BFE 数据面转发纯透传，客户端必须按 provider 原生路径发起请求，同一套客户端配置无法复用。

本期在 Provider 上新增 `protocol_paths`（协议 → 上游 base path 的声明式映射），导出至 BFE `AIConf.ProtocolPaths`，由 BFE 在转发时按检测协议改写路径（标准入口 `/v1/...` → provider 前缀）。**BFE 数据面已先行实施**（`bfe/bfe_server/ai_path_rewrite.go`、`AIConf` 扩展与加载期校验、`doSingleAIForward` 接入、SC17 集成测试 6 例全绿，见 `bfe/docs/zh_cn/modifications/2026-09-14-ai-protocol-paths-rewrite/design-changes.md`），本变更补齐控制面：字段、校验、存储、导出与接口契约。

语义约定：`protocol_paths[protocol]` = 该协议官方 SDK `base_url` 的 path 部分（openai 含 `/v1` 尾；anthropic 不含，SDK 自拼 `/v1/messages`），与 BFE 改写公式自洽。

## 2. 目标

| # | 目标 | 验证标准 |
|---|------|----------|
| G1 | `Provider`/`ProviderParam` 增加 `protocol_paths`（map[string]string） | create/get/update/patch/list 全链路携带该字段 |
| G2 | 创建/更新校验 | key ∈ {openai, anthropic} 且 ⊆ `model_protocols`；value 以 `/` 开头、不以 `/` 结尾、不含 `..`/`?`/`#`、长度 ≤ 128；非法输入返回参数错误 |
| G3 | 导出链路透传至 BFE | `newAIConf` 输出 `AIConf.ProtocolPaths` = provider `protocol_paths` 恒透传值 |
| G4 | 部分更新与操作日志语义正确 | `UpdatePricingTiers`、`applyProviderUpdate`、`providerParamToMap`/`providerToMap` 不丢字段、不误清空 |
| G5 | dashboard 配套 | 创建/编辑表单支持按协议输入（`ai-gateway-web` 仓，配套发布） |

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（本说明）；dashboard 见 `ai-gateway-web`；BFE 数据面已实施 |
| 主要文件 | `model/iprovider/provider.go`、`model/iprovider/provider_operation_log.go`、`storage/rdb/internal/dao/table_providers.go`、`storage/rdb/provider/provider.go`、`model/icluster_conf/cluster.go`、`endpoints/openapi_v1/provider/` |
| 接口契约 | OpenAPI `/providers` 新增 `protocol_paths`；`/clusters` 契约不变；InnerAPI 导出新增 `AIConf.ProtocolPaths` |
| 数据迁移 | `TProvider` 增列 `protocol_paths` TEXT（JSON）；`db_ddl.sql` / `db_ddl_sqlite.sql` 同步；存量数据为 NULL = 未配置，无迁移动作 |
| BFE 影响 | 无新增——BFE 侧（含加载期 key 白名单校验）已先行发布 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 不引入 `provider_type` 概念 | 同一品牌的不同套餐/产品线（Kimi 开放平台 vs Kimi Code 会员）建模为**不同 provider 实例**（各自命名、keys、`protocol_paths`），网关不感知类型；评审已否决类型枚举方案 |
| 不引入通用 `path_rewrite` 规则 | 调研 provider 100% 可被"协议 → 单 base path"公式表达；通用改写机制属投机设计（YAGNI），且每多一条自由改写通道就多一个计费域错配口子（火山 path 挂计费语义）；未来表达不了时按"协议内规则列表"增量扩展 |
| `protocol_paths` 取 SDK base_url 语义 | 配置值可直接照抄 provider 官方文档的 base_url 一栏，与 BFE 改写公式（anthropic 追加完整路径 / openai 去 `/v1` 前缀）自洽 |
| cluster 恒透传 provider 值 | 与 `ModelProtocols` 同模式：`LLMConfig` 不新增路径字段，单一事实来源 |
| 双端校验 | 控制面校验配置合法性；BFE 加载期白名单校验（`AIConfCheck`，同 `ValidateProtocols` 先例）不因此放宽。新增协议时需双端同步放开，与 `ModelProtocols` 同节奏 |
| gemini 不进 `protocol_paths` | gemini 原生路径即标准路径、透传已可用且无改写目标；BFE 白名单兜底拒绝，开放问题记入总体方案 §9.2 |

## 5. 关联文档

- 详细设计：`design-changes.md`
- 接口变更：`api-changes.md`
- 总体设计方案：`document-ai-gateway/迭代系统设计/v0.7/rewrite支持/rewrite支持-设计方案.md`
- 竞品调研：`document-ai-gateway/迭代系统设计/v0.7/rewrite支持/tokenhub-newapi-rewrite能力调研.md`（new-api 硬编码特例表 / TokenHub 不支持，佐证配置化价值）
- BFE 侧记录：`bfe/docs/zh_cn/sys_design/ai_protocol_paths.md`、`bfe/docs/zh_cn/modifications/2026-09-14-ai-protocol-paths-rewrite/design-changes.md`（已实施）
- 接口定义：`design-docs/api-define/OpenAPI接口定义/providers.md`
