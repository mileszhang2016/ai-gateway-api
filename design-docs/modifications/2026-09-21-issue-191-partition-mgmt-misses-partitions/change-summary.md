# 变更摘要：修复分区维护 Job 在配置 `[Report].Database` 后永不创建分区（Issue #191）

## 背景

GitHub Issue [rainway-ai-gateway/ai-gateway-api#191](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/191)：

配置 MySQL 报表库（`[Report] Backend="mysql"`、`Database="open_bfe_222"`、`EnablePartitionMgmt=true`）后，`bfe_ai_request_log` 始终只有建表时的 `p_init` 分区（`VALUES LESS THAN (740244)`，即 `log_time < 2026-09-20`）。日志时间进入 2026-09-21 后，log-reader 写入持续报 `Error 1526 (HY000): Table has no partition for value 740245`，重试 4 次后丢行。两个环境均复现；将 `EnablePartitionMgmt` 改为 `true` 并重启 `ai-gateway-api` 无效。

## 根因定位

分区 JOB 已正常运行，但**元数据查询条件错误**，导致它把已分区的表误判为"未分区表"，永远走不到 ADD PARTITION 分支：

1. `stateful/container/rdb/components.go:385` 以 `cfg.Database`（`open_bfe_222`，非空）构造 `mysqlreport.NewJob`，`Job.database` 非空。
2. `storage/mysqlreport/job.go:80-85` 的 `j.table()` 据此生成 **schema 限定名** `open_bfe_222.bfe_ai_request_log`。
3. `job.go:288-289` 将该限定名传入 `manageTablePartitions` → `listPartitions`（`job.go:364-365`），而查询 SQL（`job.go:359-362`）以 `TABLE_NAME = ?` 过滤 `information_schema.PARTITIONS`——该视图的 `TABLE_NAME` 是**不含库名的裸表名**，限定名恒匹配不到，查询返回 0 行。
4. `manageTablePartitions`（`job.go:328-335`）将 0 行解读为"表未分区"（为兼容不支持分区的 RDS 形态设计的 DELETE 兜底分支），于是只执行 `purgeBatches`（无过期数据时静默 no-op），**从不发起任何 ADD PARTITION DDL**。

由此可完整解释 Issue 中的全部现象：

- **"刚开始可以正常写入，运行数小时后开始报错"**：`db_ddl_report_mysql.sql` 要求建表时 `p_init` 覆盖建表当日 +3 天；边界（2026-09-20）之前写入正常，越过边界即 1526。
- **`EnablePartitionMgmt=true` 重启无效**：门控逻辑（`components.go:376-387`）工作正常，JOB 确实在跑，失败点与开关无关。
- **为何现有单测未拦截**：`job_test.go` 中所有 `RunPartitionMgmtOnce` 用例均以 `NewJob(db, "", ...)` 构造（`job_test.go:210/247/280/293/312`），`database` 为空时 `j.table()` 返回裸名，查询恰好能匹配——测试矩阵缺失"database 非空"这一生产形态。
- **集成测试同样存在盲区**：`test/integration/tests/report/design.md:127-128` 明确 `EnablePartitionMgmt = false`，"JOB 行为由单测覆盖"，而单测恰好漏了该形态。
- **附带影响（非数据丢失）**：兜底 DELETE 按 `RetentionDays` 清旧数据，保留语义未失效；`bfe_ai_metrics_1m` 由同一 JOB 维护（`job.go:289`），同病——其聚合写入（api 自身）同样会在 `ts_min` 越界后 1526 失败。

## 目标

分区维护 JOB 在 `[Report].Database` 非空（跨库/限定表名）时，能正确发现现有分区并执行创建/删除；`Database` 为空时行为保持不变。不再静默降级为 DELETE 兜底。

## 修复方案

### 1. 代码修复（`storage/mysqlreport/job.go`）

- **`listPartitions` 按 (schema, 裸表名) 查询**：SQL 改为 `WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`；`j.database == ""` 时 schema 条件回退为 `DATABASE()`（保持现语义）。等价改法：`TABLE_SCHEMA = COALESCE(?, DATABASE())` 单条 SQL 绑定 `j.database`（空串时 `COALESCE` 落到当前库）。
- **`RunPartitionMgmtOnce` 区分两种名字**（`job.go:284-290`）：遍历目标表时用裸名常量（`tableDetail`/`tableMetrics`，见 `report.go:29-30`）做元数据查询；ADD/DROP/DELETE 等 DDL 仍经 `j.table()` 生成限定名（跨库场景 DDL 必须带库名）。
- **消除静默降级**：`manageTablePartitions` 的 DELETE 兜底分支（`job.go:328-335`）增加 WARN 日志（"table %s has no partitions, fallback to batched DELETE purge"）。分区查询修复后，走入该分支只剩"真未分区"一种情况， Warn 可见可告警；未来若再出现元数据不匹配也能立刻暴露。

### 2. 回归测试（`storage/mysqlreport/job_test.go`）

- 新增 `TestRunPartitionMgmtOnce_QualifiedDatabase`：`NewJob(db, "open_bfe_222", ...)`，断言：
  - information_schema 查询参数为 `("open_bfe_222", "bfe_ai_request_log")` / `("open_bfe_222", "bfe_ai_metrics_1m")`；
  - 返回已有分区行后，执行 `ALTER TABLE open_bfe_222.bfe_ai_request_log ADD PARTITION ...` 限定名 DDL；
  - **不出现** DELETE 兜底语句（`mock.ExpectationsWereMet` 即可保证）。
- 现有 `""` 场景用例（裸名）全部保留，锁定空 database 行为不回归。

### 3. 受影响环境的应急措施（修复上线前，可选但推荐）

先查实际边界（两环境可能不同，两张表分别查）：

```sql
SELECT PARTITION_NAME, PARTITION_DESCRIPTION
FROM information_schema.PARTITIONS
WHERE TABLE_SCHEMA = 'open_bfe_222' AND TABLE_NAME = 'bfe_ai_request_log'
  AND PARTITION_NAME IS NOT NULL ORDER BY PARTITION_ORDINAL_POSITION;
```

按实际边界补齐到当天 +3 天（以边界 740244 = 2026-09-20、今天 2026-09-21 为例；`bfe_ai_metrics_1m` 同理）：

```sql
ALTER TABLE bfe_ai_request_log ADD PARTITION (
    PARTITION p20260920 VALUES LESS THAN (TO_DAYS('2026-09-21')),
    PARTITION p20260921 VALUES LESS THAN (TO_DAYS('2026-09-22')),
    PARTITION p20260922 VALUES LESS THAN (TO_DAYS('2026-09-23')),
    PARTITION p20260923 VALUES LESS THAN (TO_DAYS('2026-09-24'))
);
```

说明：

- 需要维护账号具备 `ALTER` 权限（见 `db_ddl_report_mysql.sql` 头部权限说明）。
- 若不做手工补分区，修复上线后 JOB 首个周期（`Start()` 立即执行一轮，`job.go:131-140`）也会自动补齐：`planPartitionAdds` 以现有最大边界 09-20 为起点追加，首个 `p20260921 VALUES LESS THAN (TO_DAYS('2026-09-22'))` 将覆盖 [09-20, 09-22) 两天（RANGE 追加语义，后续滚动/清理均按边界计算，分区名短暂错位无害）。手工 DDL 的好处是分区名与日期对齐更整洁。
- 被 log-reader 丢弃的历史行不可恢复，也无需补数（明细缺口的分钟聚合会略少，属预期）。

### 4. 集成测试修改方案（`test/integration/tests/report/`，新增组 C：partition）

现状：组 A（not_assembled，离线必跑）验证 `[Report]` 缺省时模块不装配、五端点 404；组 B（query，`REPORT_MYSQL_DSN` 门控的真实 MySQL 8.x）验证五个查询端点，且装配配置为 `EnablePartitionMgmt = false`（`design.md:127-128`，"JOB 行为由单测覆盖"）。JOB 行为此前确实只由 `job_test.go` 单测覆盖，而单测矩阵缺失 `database` 非空形态——#191 正是该缺口在生产形态上被放大。为此新增组 C，把"分区 JOB 在真实 MySQL + schema 限定名形态下自动建分区"的行为固定到集成层。

**目录结构**（组 C 与组 B 并列、自包含）：

```
report/
├── design.md                  # 增补组 C 章节（设计思路、数据构造、用例统计、运行方式）
├── not_assembled/
├── query/
└── partition/                 # 组 C（新增，REPORT_MYSQL_DSN 门控）
    ├── partition_test.go      # TestMain：建库、套用 DDL（INIT_DATE=昨天）、灌入边界数据、启动服务
    └── cases_test.go          # 组 C 用例断言
```

**TestMain 数据构造要点**（沿用组 B 约定：专用随机库 `report_it_<ns>`、先 `DROP DATABASE IF EXISTS` 防御残留、结束后清理）：

1. 套用项目根 `db_ddl_report_mysql.sql` 时 `${INIT_DATE}` 替换为**昨天**（注意不是组 B 的"今天+3"）——这是用例可观测的前提：p_init 边界 < 今天，待建分区窗口 [今天, 今天+3) 非空，JOB 必须有动作；若按组 B 方式建表，p_init 已覆盖前瞻窗口，用例对修复前后都通过，失去回归意义。
2. `[Report]` 装配配置采用 #191 生产同型（回归关键形态）：

```toml
[Report]
Backend = "mysql"
Datasource = "report_db"
Database = "report_it_<ns>"     # 关键：非空，走 schema 限定名路径（本 bug 的触发条件）
EnableAggregateJob = false      # 聚焦分区行为；聚合 JOB 由单测覆盖
EnablePartitionMgmt = true
RetentionDays = 7
```

**用例（cases_test.go）**：

| 用例 | 断言 |
| - | - |
| C-1 启动即建分区 | 服务启动后轮询 `information_schema.PARTITIONS`（总超时 30s、步长 200ms）：`bfe_ai_request_log` 与 `bfe_ai_metrics_1m` 两表均出现 `p<今天>`、`p<今天+1>`、`p<今天+2>` 三个分区，边界值分别为 TO_DAYS(今天+1/+2/+3)（MySQL 8.x 回显求值整数形态）。修复前 JOB 误判表未分区、走 DELETE 兜底，该用例恒超时失败，即回归 red→green |
| C-2 覆盖期内写入成功 | 向两张表插入 `log_time = 今天`、`今天+2 23:59:59` 的行成功——模拟 log-reader 写入路径，锁定 #191 的下游症状（Error 1526）消除 |
| C-3 边界外写入拒绝 | 向两张表插入 `log_time = 今天+3` 的行，期望失败且为 Error 1526——固定"JOB 只保有 3 天前瞻"的契约，防止未来误调小前瞻窗口而无感知 |

实现说明：

- **时序**：JOB 在 `Start()` 时立即跑一轮周期（`job.go:131-140`），用例以轮询等待代替固定 sleep；断言全程秒级完成，跨自然日漂移风险可忽略（如需绝对稳健，可在 TestMain 记录 `today` 并快速完成断言）。
- **环境门控与运行方式**同组 B：`REPORT_MYSQL_DSN="root:****@tcp(127.0.0.1:3306)/" go test -v -count=1 ./tests/report/partition/`；未设置时打印说明后以 0 退出，CI 无新增依赖；账号要求同组 B（需 `CREATE/DROP DATABASE`，维护账号还需 `ALTER`）。
- **辅助代码**：组 C 自建 DDL/连接辅助（与组 B 的 `parseServerDSN`/`applyReportDDL`/`cleanupDB` 同构，`applyReportDDL` 的 INIT_DATE 入参化）；本次不做组 B 重构，后续若再增用例组可抽取共享 helper 包。
- **两表同验**：`bfe_ai_metrics_1m` 与明细表由同一 JOB 循环维护（`job.go:284-294`），#191 中同病（其写入方是聚合 JOB 自身，越界后同样 1526），组 C 必须同时覆盖两张表。
- **design.md 同步**：组 C 加入后更新 `test/integration/tests/report/design.md`——§1 模块概述补组 C 说明、§3 用例统计表增加 3 条、§5 目录结构补 `partition/`、新增"组 C：分区自动维护"章节（环境门控、数据构造、期望值）。

## 影响范围

| 对象 | 影响 |
| - | - |
| `storage/mysqlreport/job.go` | `listPartitions` 查询条件、`RunPartitionMgmtOnce` 传参、兜底分支告警日志（见修复方案） |
| `storage/mysqlreport/job_test.go` | 新增 database 非空回归用例 |
| `test/integration/tests/report/` | 新增 partition 用例组（组 C），并更新组内 `design.md`（见修复方案 §4） |
| `stateful/container/rdb/components.go` | 无需改动（门控/装配逻辑正确） |
| `db_ddl_report_mysql.sql` | 无需改动（"后续分区由 JOB 自动滚动维护"的设计不变，本次是恢复该承诺） |
| `design-docs/sys-design/报表库设计文档.md` | 无需改动（设计意图即 JOB 自动维护，实现 bug 不构成交付设计变更） |
| log-reader | 无需改动（写入失败是分区缺失的下游症状） |

## 验证

1. `make test`（重点 `go test ./storage/mysqlreport/...`）通过；新增回归用例在修复前应能复现"查询参数为限定名 → 0 行 → 走 DELETE 兜底"的错误形态。
2. 真实 MySQL 复现 Issue 步骤：仅 `p_init` 建表 → `[Report] Database="open_bfe_222" EnablePartitionMgmt=true` 启动 → 观察 `information_schema.PARTITIONS` 出现 `today..today+2` 三个新分区 → 写入 `log_time` 为当日的日志成功。
3. 回归空 database 形态（不设 `[Report].Database`）：分区维护行为与修复前一致。
4. 组 C 集成用例：`REPORT_MYSQL_DSN="root:****@tcp(127.0.0.1:3306)/" go test -v -count=1 ./tests/report/partition/`，其中 C-1 在修复前代码上应超时失败（red）、修复后通过（green），构成完整回归证据链。

## 后续事项

1. 按修复方案提交代码与回归测试，PR 关联 Issue #191 并在 Issue 中回复根因摘要；
2. 修复随版本发布到两个受影响环境后，复查两库两张表的分区覆盖情况（应急 DDL 与自动补齐可能叠加，确认无重复/无缺口的最终状态）；
3. 实现组 C 集成用例（见修复方案 §4），补上"JOB 行为仅由单测覆盖"的测试金字塔缺口——本次根因正是单测形态遗漏在生产形态上被放大。
