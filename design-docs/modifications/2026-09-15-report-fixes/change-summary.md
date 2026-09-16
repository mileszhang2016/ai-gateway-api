# 报表模块缺陷修复（时区口径 / 分区边界解析）变更摘要

> 本变更是对 `2026-09-15-report-query-api` 首次实现的两个缺陷修复；实现范围仅限 ai-gateway-api 仓库。

## 1. 背景

报表查询 API（`/report/*`）首次实现后，在真实组件全链路集成测试（BFE → pb 日志 → log-reader → MySQL → 报表 API）中暴露两个缺陷：

### 缺陷 1：`log_time` / 时间桶的 Unix 秒渲染存在会话时区偏移

- **现象**：`GET /report/logs` 返回的 `log_time`（及 `timeseries` 的桶 `time`）在 CST 会话的 MySQL 上比真实 Unix 时间戳偏移 8 小时。
- **根因**：log-reader `mod_log_mysql` 写入 `DATETIME` 使用 go-sql-driver 默认 `loc=UTC`，即**按 UTC 墙钟存储**；查询侧渲染用 `UNIX_TIMESTAMP(log_time)`，该函数按**会话时区**解读墙钟——会话默认 `SYSTEM`（CST）时，同一墙钟被解读为 CST 时刻，换算出的 Unix 秒偏移 8 小时。`timeseries` 的桶边界 `FLOOR(UNIX_TIMESTAMP(ts_min)/b)*b` 同样受影响。
- **影响面**：MySQL 后端与 Doris 后端（Doris 的 `UNIX_TIMESTAMP` 同样会话时区敏感）；任何会话时区非 UTC 的部署均出现偏移，属于跨组件口径 bug。

### 缺陷 2：分区管理 JOB 周期报 Error 1493

- **现象**：`EnablePartitionMgmt=true` 时，JOB 启动报一次 `Error 1493`（重复创建分区），之后每 6h 周期重复。
- **根因**：建表 DDL 的分区定义为 `VALUES LESS THAN (TO_DAYS('YYYY-MM-DD'))`，而 MySQL 在 `information_schema.PARTITIONS.PARTITION_DESCRIPTION` 中回显的是**求值后的整数**（如 `738886`），不是 `TO_DAYS('...')` 文本。`parseBoundaryDate` 只认字面日期文本 → 现有分区被误判为"缺失" → 规划重复 `ADD PARTITION`。
- **连带影响**：`p_init` 这类整数回显分区原先永远解析不出边界日期，不参与过期 DROP 决策，超期分区会无限滞留。

## 2. 修复方案

| 缺陷 | 方案 |
|------|------|
| 时区偏移 | 所有「DATETIME → Unix 秒」渲染改为无时区算术表达式 `TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', <col>)`（纯日期时间差，不涉及时区解读；存储墙钟即 UTC 墙钟，结果即真实 Unix 秒）。MySQL/Doris 两后端同步修改，口径保持一致 |
| 1493 噪音 | `parseBoundaryDate` 重写，支持三种 `PARTITION_DESCRIPTION` 形态：①纯整数（可带引号）按 TO_DAYS 值反解日期（`epoch + (n-719528)` 天；TO_DAYS/FROM_DAYS 与 Go `time` 同为 proleptic Gregorian 历，直接换算；带 `n >= 700000` 下界 sanity 防止垃圾输入导致级联误建）；②字面日期文本（原行为保留兼容）；③`MAXVALUE` 显式跳过（不参与 DROP，保守）。不可解析的值维持"跳过 DROP 决策"的保守语义 |

分区规划函数（`planPartitionAdds` / `planPartitionDrops`）无需改动：整数反解落地后，`p_init` 等分区可被正确识别——不再重复 ADD，且首次纳入过期 DROP 范围（顺带修复了超期分区滞留问题）。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（本仓库）；log-reader 无改动；本仓库 `test/integration` 集成测试断言同步简化 |
| 修改模块 | `storage/mysqlreport/report.go`、`storage/mysqlreport/job.go`、`storage/dorisreport/report.go` 及对应单测；`test/integration/tests/report/` 断言 |
| 接口形状 | 无变化（字段、参数、结构均不变）；仅 `log_time` / 时间桶 `time` 的**取值语义**修正为真实 Unix 秒 |
| 数据迁移 | 无 |

## 4. 验证

- 时区修复的端到端锚点：CST 会话下 `UNIX_TIMESTAMP('2026-09-15 10:00:00')` = 1789437600（= 02:00 UTC，即偏移 8h），真实 Unix 秒为 1789466400。集成测试用例直接断言真实 Unix 秒常量——旧 SQL 必失败、新 SQL 通过。
- 分区解析单测覆盖 8 种输入（整数 738886→2023-01-01、740242→2026-09-18、字面文本、MAXVALUE、非法值、过小整数、空串等），并验证"全整数回显布局下规划为空（不重复 ADD）"与"整数反解后 DROP 决策生效"。
- `go build` / `vet` 干净；storage/model/endpoints 单测全绿；集成组 A（离线）与组 B（真实 MySQL 8.4，16/16 PASS）通过。

## 5. 关联文档

- 首次实现：[2026-09-15-report-query-api](../2026-09-15-report-query-api/change-summary.md)
- 详细设计：[design-changes.md](./design-changes.md)
