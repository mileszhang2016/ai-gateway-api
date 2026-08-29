# RMB 配额支持分时段/分工作日定价——变更摘要

## 1. 背景

随着 DeepSeek 等模型提供商采用"高峰 / 空闲"分时段定价策略，BFE RMB 配额扣费需要具备按请求发生时刻匹配不同价格的能力。本变更是 v0.4《BFE RMB 配额支持 peak/off-peak 时段定价增强方案》的更新版，主要应对以下变化：

1. **ai-gateway-api 已完成 provider 与 cluster 的概念分离**：`provider` 负责下游模型提供商接入能力，`cluster` 负责转发策略并引用 provider。时段规则需要明确归属与下发链路。
2. **DeepSeek 价格体系调整**：官方以北京时间区分工作日高峰时段（周一至周五 09:00-12:00、14:00-18:00）与空闲时段，空闲价格为高峰价格的一半。

由于 v0.4 已预留 `TimeRange.Weekdays` 按星期几分段的能力，BFE 运行时匹配逻辑无需大改。本次重点是在 provider/cluster 分离后的新架构下，重新明确时段配置的归属、下发链路，并更新 DeepSeek 配置示例。

## 2. 目标

1. 支持按请求发生时刻匹配特定价格 tier（如 `peak`），未命中时 fallback 到默认价格；通过 `Weekdays` 保留区分工作日与休息日的扩展能力。
2. 支持不同 provider 配置不同的时区和时段规则。
3. 结合 provider/cluster 分离，明确时段配置归属：时段模板放在 `/providers`，分时段价格放在 `/model-prices`。
4. 保持对现有固定价格的向后兼容。
5. 结构可扩展，自然支持缓存命中 / 未命中、200k+ 长文本等分层价格。
6. 运行时成本计算仍保持纯整数运算，不引入浮点误差。
7. 价格键名符合 `model-prices.md` 的枚举规范。

## 3. 范围

- **涉及面**：`ai-gateway-api` 控制面（`/providers`、`/model-prices`、配置导出）及 `bfe` 数据面（`AIConf.ModelTable` 加载、时段匹配、`calcCostUnits`）。
- **不涉及面**：`TokenUsage.UsedCost`、Lua 扣减逻辑、Redis 定点数存储等已有 RMB 扣费基础设施。
- **数据面影响**：BFE `AIConf.ModelTable` 结构复用 v0.4 设计，仅在配置来源上发生变化。

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 时段模板归属 provider | 同一 provider 下所有模型共享 `time_zone` 和 `tiers`，避免重复配置。 |
| 分时段价格归属 model-prices | 不同模型在同一 tier 下可有不同价格，保留价格灵活性。 |
| 独立 pricing-tiers 接口 | 高峰/闲时模板通过 `PUT /providers/{provider_name}/pricing-tiers` 单独维护，保持创建 provider 时接口简洁。 |
| 默认价格作为 fallback | `prices` 字段继续表示非 tier 命中或 tier 未覆盖键时的默认价格。 |
| 初期只支持 `peak` tier | 简化首期实现，后续按需放开 tier name 约束。 |
| 定点整数运算 | `prices` 与 `tier_prices` 在加载阶段转换为定点整数，运行时避免浮点误差。 |
| 缓存命中价格键名 | 使用 `model-prices.md` 标准键名 `cache_read_input_token_cost`。 |

## 5. 关联文档

- 详细设计：`design-changes.md`
- 接口变更：`api-changes.md`
- 上游设计来源：`document-ai-gateway/迭代系统设计/v0.5/计费能力扩展/bfe-rmb-tiered-pricing-design.md`
- 相关基础设计：`ai-gateway-api/design-docs/sys-design/details/provider与cluster概念分离.md`

## 6. 实施阶段

| 阶段 | 内容 | 状态 | 关键文件 |
|------|------|------|----------|
| 1 | 设计与文档冻结 | ✅ 已完成 | `design-docs/sys-design/details/RMB配额分时段定价.md`、`design-docs/api-define/OpenAPI接口定义/providers.md`、`model-prices.md`、`InnerAPI接口定义/server-data-conf.md` |
| 2 | 控制面数据模型与存储（provider / model-prices 新增字段） | ✅ 已完成 | `db_ddl.sql`、`db_ddl_sqlite.sql`、`storage/rdb/internal/dao/table_providers.go`、`table_model_prices.go`、`storage/rdb/provider/provider.go`、`storage/rdb/model_price/model_price.go` |
| 3 | `pricing-tiers` 接口与校验 | ✅ 已完成 | `endpoints/openapi_v1/provider/pricing_tiers.go`、`model/iprovider/provider.go`、`model/iprovider/validate.go`、`model/imodel_price/model_price.go`、`model/imodel_price/validate.go` |
| 4 | 配置导出适配（`AIConf.ModelTable` 拼接） | ✅ 已完成 | `model/icluster_conf/cluster.go`、`model/iroute_conf/exporter.go` |
| 5 | BFE 数据面加载、匹配与成本计算 | ✅ 已完成（BFE 侧结构已就绪） | `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`（已有 `TimeZone`/`Tiers`/`TierPrices`/`ActiveTierName`） |
| 6 | 测试与发布 | ✅ 已完成 | `go test ./...` 全量通过；新增集成测试 `test/integration/tests/provider/pricing_tiers`、`model_price/tier_prices`、`innerapi/tls_conf/tiered_pricing` |

## 7. 实现结果

- 代码已于 `2026-08-25` 完成并全量测试通过（`go build ./...`、`go test ./...`）。
- 新增 OpenAPI 端点 `PUT /providers/{provider_name}/pricing-tiers` 已注册并支持 `application/json`、`text/yaml`、`multipart/form-data` 三种提交方式。
- Inner API `/configs/tls_conf/server_data_conf` 返回的 `ClusterConf.Config.<cluster>.AIConf.ModelTable` 已自动填充 `Currency`、`TimeZone`、`Tiers`、`Models[].TierPrices`。
- BFE 数据面无需额外改动，其 `AIConf.ModelTable` 结构与 `ActiveTierName` / `calcCostUnits` 已支持分时段定价。
- 集成测试覆盖 pricing-tiers、model-prices tier_prices、InnerAPI 导出三个链路；测试期间修复了 `UpdatePricingTiers` 覆盖其他 provider 字段、`mergeModelPrice` 遗漏 `TierPrices`、时段重叠未校验 3 个问题。
