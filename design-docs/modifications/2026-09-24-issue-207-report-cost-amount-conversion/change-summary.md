# Report 成本字段定点→金额换算（issue #207）变更摘要

> 本文档覆盖 report 接口成本字段单位换算的 ai-gateway-api 仓库改动方案。关联卡片：ai-gateway-api [#207](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/207)；前端显示缺陷根因见 ai-gateway-web [#115](https://github.com/rainway-ai-gateway/ai-gateway-web/issues/115)。

## 1. 背景

报表链路的成本字段 `ai_cost_value` 在全链路上保持**定点整数**（1 单位 = 1e-8 元/美元）：

- BFE 计费侧 `quota.CalcCostUnits = round(用量 × 元单价 × 1e8)`（go-lib `quota.RmbPrecision=1e8`），pb 访问日志 `ai_cost_value` 原样写入定点值（`bfe/bfe_modules/mod_access_pb3/request_log.go:448`）；
- log-reader 原样透传落库，报表库明细表/聚合表均为定点整数；
- 当前 ai-gateway-api 的 `/open-api/v1/report/*` 接口将定点整数**原值**返回（`storage/mysqlreport/report.go:431 queryCost`、`storage/dorisreport/report.go:485 queryCost` 直接返回 `SUM` 结果），API 契约（`design-docs/api-define/OpenAPI接口定义/report.md:112`）约定"前端按 `currency` 格式化"。

ai-gateway-web #115 实测：`/report/overview` 返回 `cost=[{"currency":"RMB","value":66900}]`（= 0.000669 元，与 API Key `balance.used` 一致），前端未做 ÷1e8 直接 K/M 缩写显示为 "66.9K RMB"，显示虚增 1e8 倍。底层扣费、报表库、接口返回值均正确，纯展示单位缺陷。

## 2. 目标

- **换算职责从后段（report.md:112）改为本仓库**：`/open-api/v1/report/*` 三个暴露成本字段的端点（overview / timeseries / logs）在**服务端完成 ÷1e8 换算**，直接返回按币种单位的金额（元/美元）；
- 换算内聚在报表模型层单一 helper，MySQL / Doris 两个 storager 共用，保证两种后端口径一致；
- 报表库、聚合 JOB、BFE、log-reader 保持定点存储不变，本特性纯查询出口改造。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（本仓库）；ai-gateway-web 前端适配见 #115 联动 |
| 修改模块 | `model/ireport/`（类型与换算 helper）、`storage/mysqlreport/`、`storage/dorisreport/`（扫描/组装处换算） |
| 测试适配 | 单元测试（`model/ireport`、`storage/*report`）+ 集成测试组 B（`test/integration/tests/report/query/`，种子仍灌定点、断言改金额，并补 timeseries cost 与 logs 成本两处缺口用例）+ **schema 套件新增 report 用例组**（`test/integration/tests/schema/report/`，api-define 契约形状守卫 + 金额口径回归锁，见 design-changes.md §4.3） |
| 涉及接口 | `GET /open-api/v1/report/overview`、`GET /open-api/v1/report/timeseries`（metric=cost）、`GET /open-api/v1/report/logs` |
| 不涉及接口 | rankings / distribution 无成本字段，不变 |
| 数据迁移 | 无；报表库 schema、聚合表口径、聚合 JOB 均不变（继续存定点整数） |
| 数据面影响 | 无（BFE / log-reader 零改动，pb 日志与落库语义不变） |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 换算放在 storager 扫描/组装出口，SQL 与库表不动 | 定点值在 DB 内 SUM、按桶聚合均保持整数精确；只在返回结构组装处除 1e8，一处转换、两种后端天然同口径 |
| 换算系数内聚为 `ireport.CostFixedPointScale = 1e8` + `ireport.CostFixedPointToAmount()` | RMB/USD 定点精度统一 1e8（report.md:131、00-common.md:208）；未来若引入不同精度币种，单点扩展 |
| `CostItem.Value` / `LogRow.CostValue` 类型 int64 → float64 | JSON 输出由定点整数变为金额浮点数（如 66900 → 0.000669），字段名不变；**破坏性变更**，见 §5 |
| 直接除法不做舍入 | 业务余额上限 9000 万元 ⇒ 定点值 ≤ 9×10^15 < 2^53，float64 精确表示无损失；Go/JS JSON 最短表示可正常输出 0.000669 量级 |
| 明细行无成本时保持 null | `LogRow.CostValue` 为 `*float64`，NULL 行不变 nil，不输出 0 |
| 空币种桶继续由 SQL 过滤 | `ai_cost_currency != ''` 条件不变，不会出现空币种金额桶 |

## 5. 兼容性影响（Breaking Change）

- `cost[].value`（overview）、timeseries cost 序列 `value`、`logs items[].ai_cost_value` 由 **int64 定点整数** 变为 **float64 金额**（元/美元），JSON 类型与数值语义同时变化；
- 唯一已知消费者为 ai-gateway-web Report 模块（尚未发布到 web 主干），随 #115 联动适配：直接展示金额并保留 K/M 缩写 formatter，**不再做任何定点换算**；
- 需在 release note 中声明该 breaking change；OpenAPI 路径仍为 v1（Report API 随本特性首次随版本发布前完成口径收敛，不升版本号）。

## 6. 关联文档

- `ai-gateway-api/design-docs/modifications/2026-09-24-issue-207-report-cost-amount-conversion/api-changes.md`
- `ai-gateway-api/design-docs/modifications/2026-09-24-issue-207-report-cost-amount-conversion/design-changes.md`
- 历史契约：`design-docs/modifications/2026-09-15-report-query-api/`（Report API 首次引入时约定"定点原值 + 前端格式化"，本次将其变更）
