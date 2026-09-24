# Report 成本字段定点→金额换算（issue #207）设计变更说明

> Step 4/5 执行依据：sys-design 同步项 + 代码修改点（含文件与函数级位置）。

## 1. 换算模型

定点精度在整个报表链路统一为 1e8（BFE `quota.RmbPrecision`、报表库、00-common.md:208），本特性在报表模型层定义唯一换算出口：

```go
// model/ireport/types.go（新增）
// CostFixedPointScale is the scale of the fixed-point cost values stored
// in the report tables and produced by BFE: one unit equals 1e-8 of the
// currency amount (1e-8 yuan for RMB, 1e-8 dollar for USD).
const CostFixedPointScale = 1e8

// CostFixedPointToAmount converts a raw fixed-point cost integer into the
// currency amount returned by the report API.
func CostFixedPointToAmount(value int64) float64 { return float64(value) / CostFixedPointScale }
```

- 选择除法而非乘法逆运算：`float64(v) / 1e8`，1e8 与业务范围内全部定点值（≤ 9×10^15 < 2^53）均可被 float64 精确表示，换算无精度损失；
- 不做舍入：定点值本身整数精确，除法结果的最短十进制表示（如 0.000669）即业务可读值；
- 未来若出现非 1e8 精度的币种，仅需将 helper 扩展为按币种查表，调用点不变。

## 2. 类型变更（model/ireport/types.go）

| 位置 | 变更前 | 变更后 |
|------|--------|--------|
| `CostItem.Value`（:117-120） | `int64`，注释"fixed-point integer … frontend formats it" | `float64`，注释"cost amount in the currency unit (yuan/dollar); converted from the fixed-point value by CostFixedPointToAmount" |
| `LogRow.CostValue`（:208） | `*int64` json `ai_cost_value` | `*float64`，同 json 名；语义变为金额，nil 语义不变 |
| `MetricCost` 常量注释（:31） | "cost growth (fixed-point integer per second)" | "cost growth (amount per second in the currency unit)" |

`OverviewResult`、`MetricPoint`、`ReportStorager` 接口签名均不涉及成本类型，无需改动；manager 层（`model/ireport/manager.go`）与 endpoints（`endpoints/openapi_v1/report/`）对成本字段纯透传，零改动。

## 3. storager 修改点（MySQL / Doris 对称，共 3 处/后端）

### 3.1 `queryCost`（overview 成本桶）

`storage/mysqlreport/report.go:431`、`storage/dorisreport/report.go:485`：

```go
// 变更前：rows.Scan(&item.Currency, &item.Value)
// 变更后：
var raw int64
if err := rows.Scan(&item.Currency, &raw); err != nil { ... }
item.Value = ireport.CostFixedPointToAmount(raw)
```

SQL（`SUM(ai_cost_value_sum)`）与聚合表口径不变，仅扫描出口换算。

### 3.2 timeseries cost 点（rowToMetricPoint）

`storage/mysqlreport/report.go:623-626`、`storage/dorisreport/report.go:701-704`：

```go
// 变更前：
point.Value = float64Ptr(float64(v.value) / float64(bucketSec))
// 变更后：
point.Value = float64Ptr(float64(v.value) / float64(bucketSec) / ireport.CostFixedPointScale)
```

即"定点/秒"→"金额/秒"，先按桶宽求速率、再换算金额（与先换算再求速率数学等价，复用现有速率先算的顺序）。

### 3.3 logs 明细行

`storage/mysqlreport/report.go:801,843,888`、`storage/dorisreport/report.go:915,957`（Doris 组装行同区域）：

- 扫描目标 `costValue sql.NullInt64` 不变（列类型不变）；
- 组装处新增指针换算 helper（两后端各一个，或下沉到 ireport）：

```go
func nullCostAmountPtr(v sql.NullInt64) *float64 {
    if !v.Valid { return nil }
    f := ireport.CostFixedPointToAmount(v.Int64)
    return &f
}
// CostValue: nullCostAmountPtr(costValue),
```

## 4. 测试更新

### 4.1 单元测试

| 文件 | 内容 |
|------|------|
| `model/ireport/`（新增 types 单测） | `CostFixedPointToAmount`：0→0、66900→0.000669、15230000→0.1523、9e15 边界值无精度损失 |
| `storage/mysqlreport/report_test.go` | `queryCost` 断言 `Value` 为金额（现有 `CostItem{Currency:"USD", Value:5000}` → `0.00005`）；`rowToMetricPoint` cost 用例补 ÷1e8；logs 行成本断言 int→float |
| `storage/dorisreport/` 对应测试 | 同上对称更新 |

验证命令：`make test-model` + `make test-model-cover-gate`（model 覆盖率 ≥70% 门槛）。

### 4.2 集成测试（`test/integration/tests/report/`）

组 B（query，`REPORT_MYSQL_DSN` 门控）种子数据继续以定点整数灌库（库表语义不变），仅响应断言改金额口径：

| 文件 | 修改内容 |
|------|----------|
| `query/query_test.go:304-307` | `overviewData.Cost[].Value` int64 → float64（结构注释"与 model/ireport 的 JSON 形状一致"）；`logItem` 结构补充 `ai_cost_value`（`*float64`）与 `ai_cost_currency`（`*string`）字段 |
| `query/cases_test.go:65-71` | overview 成本断言：`map[string]int64` → `map[string]float64`；定点 350/300（种子）→ 金额 `3.5e-6` / `3e-6`；浮点断言用 `assert.InDelta`（与现有 latency 断言风格一致），不用 `Equal`，避免除法舍入与字面量 1 ulp 差异造成偶发失败 |
| `query/seed_test.go` | 种子值不变；头部注释"cost: USD=350 RMB=300"标注为定点口径（响应期望值见 design.md） |
| `design.md:140` | 期望口径 "cost：USD=350、RMB=300" → "cost：USD=3.5e-6、RMB=3e-6（= 定点种子 350/300 ÷ 1e8）" |

**趁口径变更补两处缺口用例**（现有集成测试未覆盖）：

1. **timeseries `metric=cost`**：锁定金额/秒口径——窗口 [09:59,10:05)、60s 桶下，USD 桶 10:00 = 定点 300 → `300/60/1e8 = 5e-8` 元/秒，RMB 桶 10:00 = 定点 300 → `5e-8`，USD 桶 10:01 = 定点 50 → `8.33e-9`（InDelta）；`series[]` 按 `currency` 分序列；
2. **logs 明细 `ai_cost_value`**：logid 1001 定点 100 → `1e-6`（USD）；logid 1004（currency 为 NULL）、1005（currency 为 `''`）锁定零成本/空币种行的 `ai_cost_value` 序列化行为（`0`/字段缺省）不回归。

组 A（not_assembled）无成本断言，不受影响；组 C（partition）测分区 JOB，与成本口径无关。组 B/组 C 需真实 MySQL 8.x（`REPORT_MYSQL_DSN`），CI 未设环境变量时自动 Skip，**实现本卡后须在带 MySQL 的环境补跑组 B**。

### 4.3 集成测试：schema 套件新增 report 用例组（本卡新增范围）

**新增 `test/integration/tests/schema/report/` 子包**，为 `/report/*` 补齐 api-define 契约级形状守卫。不放入现有 `tests/schema/openapi/` 包的原因：该包 `TestMain` 以 `testutil.StartServer()` 启动（无 `[Report]` 装配，离线单形态），而 report 端点必须注入 `[Report]` 配置 + 报表库种子，需 `REPORT_MYSQL_DSN` 门控——独立子包各自保持单一启动形态。

| 文件 | 内容 |
|------|------|
| `tests/schema/report/schema.go` | 声明式 ObjectSchema（风格对齐 `tests/schema/openapi/schema.go`）：`OverviewResultSchema`（`cost[].value` TypeNumber、`currency` TypeString）、`MetricPointSchema`（cost 序列）、`LogRowSchema`（`ai_cost_value` TypeNumber + Optional 允许 null、`ai_cost_currency` TypeString + Optional） |
| `tests/schema/report/report_schema_test.go` | TestMain：`REPORT_MYSQL_DSN` 门控（未设置打印提示后 Skip，同组 B）→ 建专用库 → 套 `db_ddl_report_mysql.sql` → 灌**最小种子**（聚合表 1–2 行 + 明细表 2–3 行）→ `testutil.StartServerWithExtraConfig` 注入 `[Databases.report_db]` + `[Report]`（两 JOB 关闭） |
| 种子设计 | 定点成本刻意取**非整数金额**：`ai_cost_value=66900`（RMB）→ 金额 `0.000669`（#115 实测值）、`15230000`（USD）→ `0.1523`（report.md:131 示例值）；覆盖 currency 为 NULL / 空串的零成本行 |
| 断言 | ① `AssertSchema` 校验三个端点响应形状（§4.2 的 `logItem` 补字段与此处 `LogRowSchema` 一致）；② 金额语义断言（`InDelta`）：overview `cost[]` 精确值、timeseries cost 金额/秒、logs 明细 `ai_cost_value`；③ **口径回归锁**：断言 `0 < value < 1`（旧定点口径下同种子必为 ≥1 的整数形态，此断言直接锁死 #115 类回归） |
| 可选 | 顺带断言无 FeatureReport 权限时 report 端点返回 402，与契约错误码表对齐 |

**实现提示**：组 B TestMain 中的 DSN 解析 / 建库 / 套 DDL 骨架（`tests/report/query/query_test.go:45-128`）随本卡抽取为 testutil 公共 helper（如 `StartReportServer`），组 B 与新子包共用，避免第二份复制；抽取不改组 B 行为。

**文档联动**：`test/integration/README.md` 测试布局与运行命令小节补 `tests/schema/report/` 条目（注明 `REPORT_MYSQL_DSN` 门控）；`tests/report/design.md` "集成测试分三个用例组" 概述补第四组（schema 形状守卫）指向新目录。

**验证**：带 MySQL 环境执行 `go test -v -count=1 -timeout 300s ./tests/schema/report/`；无 DSN 环境确认 Skip 不失败。

## 5. sys-design 同步（Step 4）

- `design-docs/sys-design/details/报表查询模块.md:62`：口径常量描述"成本按币种分组返回定点整数原值"→"成本按币种分组返回金额（定点值 ÷1e8，元/美元），换算由 storager 出口统一完成"；
- 同文档若含明细行 `ai_cost_value` 字段表/示例，同步为 number 金额口径；
- `design-docs/sys-design/summary.md` 索引无需新增条目（细节文档已存在，仅内容修订）；
- 变更完成后按六步法 Step 6 评估：本变更为出口换算的单点职责调整，机制简单，**不新增 details 沉淀文档**。

## 6. 明确不变项（防止误改）

| 组件 | 不变原因 |
|------|----------|
| BFE pb 日志 `ai_cost_value`（定点整数） | proto 语义（bfe_access.proto:371 "fixed-point integer"）与计费链路同源，BFE 不做换算 |
| log-reader 透传逻辑 | 落库值必须与扣费定点值一致，供对账 |
| 报表库明细表/聚合表列类型与聚合 JOB | DB 内 SUM/聚合保持整数精确；换算是纯出口行为 |
| `/report/rankings`、`/report/distribution` | 无成本字段 |
| quota 计费链路（Redis 定点、1e8 精度、9000 万元上限） | 与本特性无关 |
