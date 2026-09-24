# 报表模块缺陷修复（时区口径 / 分区边界解析）设计变更

## 1. 缺陷 1：DATETIME → Unix 秒渲染的会话时区偏移

### 1.1 问题链路

```
log-reader mod_log_mysql 写入：
  time.Time(Unix ts) --[driver loc=UTC]--> 存储墙钟 = UTC 墙钟
ai-gateway-api 查询：
  UNIX_TIMESTAMP(log_time) --[按 @@session.time_zone 解读墙钟]--> Unix 秒
会话时区 = SYSTEM( CST ) 时：同一墙钟被 +8h 解读 → 返回 Unix 秒偏移 8h
```

`timeseries` 桶表达式 `FLOOR(UNIX_TIMESTAMP(ts_min)/b)*b` 同理会话时区敏感：桶边界按偏移后的秒数取整，输出的 `time` 整体偏移。

### 1.2 修复

所有「DATETIME → Unix 秒」渲染统一改为**无时区算术**：

```sql
-- 单值渲染（/report/logs 的 log_time 输出列）
TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', log_time)

-- 时间桶渲染（/report/timeseries）
CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/<b>)*<b> AS SIGNED)
```

- `TIMESTAMPDIFF` 是两条 DATETIME 的整数秒差，不做任何时区解读；存储墙钟即写入侧 UTC 墙钟，差值即真实 Unix 秒；
- 常量 epoch 字面量 `'1970-01-01 00:00:00'` 为静态 SQL 片段（gendry 对 SELECT 字段按字面量拼接，无注入面；WHERE 条件仍全部参数化）；
- MySQL 与 Doris 两侧同步修改（Doris 的 `UNIX_TIMESTAMP` 同样按会话时区解读），Doris 侧 CAST 保持原 `AS BIGINT` 形状；两后端口径一致；
- 修改点（以代码为准）：`storage/mysqlreport/report.go`（log 投影 + `bucketExpr`）、`storage/dorisreport/report.go`（对应两处），新增 `epochLiteral` 常量与原理注释。

### 1.3 语义影响

接口字段、参数、响应结构不变；`log_time` 与桶 `time` 的取值从"会话时区依赖的错误值"修正为"真实 Unix 秒"。这是行为修正而非兼容变更——原取值在任何非 UTC 会话下即错误值，不存在依赖它的正确用法。

## 2. 缺陷 2：`parseBoundaryDate` 不识别整数形式的 PARTITION_DESCRIPTION

### 2.1 问题

MySQL 对表达式定义的分区边界，在 `information_schema.PARTITIONS.PARTITION_DESCRIPTION` 中回显**求值结果**而非表达式文本：

| DDL 定义 | PARTITION_DESCRIPTION 回显 |
|----------|---------------------------|
| `VALUES LESS THAN (TO_DAYS('2026-09-18'))` | `738886` |
| `VALUES LESS THAN ('2026-09-18')`（RANGE COLUMNS） | `'2026-09-18'`（文本） |
| `VALUES LESS THAN MAXVALUE` | `MAXVALUE` |

本仓库 DDL 用 `TO_DAYS(...)` 形态 → 回显恒为整数。旧解析只认字面日期文本 → 后果：

1. 现有分区全部"解析失败" → 规划误判缺失 → 重复 `ADD PARTITION` → Error 1493（启动一次 + 每 6h 周期一次）；
2. `p_init` 等整数回显分区永远没有边界日期 → 不参与过期 DROP → 超期分区无限滞留。

### 2.2 修复

`parseBoundaryDate` 重写为三形态解析（伪代码）：

```
输入 description:
  MAXVALUE          → skip（显式跳过，不参与 DROP 决策）
  纯整数 n（可带引号）→ 要求 n >= 700000（sanity 下界，防垃圾输入级联误建）；
                        日期 = epochDate + (n - 719528) 天      // TO_DAYS/FROM_DAYS 与 Go time 同为 proleptic Gregorian
  'YYYY-MM-DD' 文本   → 原解析逻辑保留（RANGE COLUMNS 等形态兼容）
  其他/解析失败       → skip（维持保守语义：不参与 DROP，不误删）
```

- 整数反解的历制依据：MySQL `TO_DAYS()` 以公元 0 年起算 proleptic Gregorian 历，与 Go `time` 的历制一致，故 `epoch + (n-719528) 天`（719528 = 0000-01-01 到 1970-01-01 的天数）可直接换算，无需走 SQL `FROM_DAYS`；
- 分区规划函数 `planPartitionAdds` / `planPartitionDrops` 不改：反解落地后自然获得两个收益——①整数回显分区被正确识别，规划为空，Error 1493 消除；②`p_init` 等分区首次获得边界日期，超期时纳入 DROP（修复滞留）；
- 独立性说明：`parseBoundaryDate` 纯函数化，输入输出与 MySQL 版本无关，单测直接覆盖。

### 2.3 遗留观察项

整数反解的**实机端到端**验证依赖分区管理 JOB 实际运行（集成测试为控制变量将 `EnablePartitionMgmt=false`）。当前由规划层单测覆盖（整数/文本/MAXVALUE 混合布局下规划为空、DROP 决策生效）；首个启用 `EnablePartitionMgmt=true` 的部署建议观察一个 6h 周期确认 1493 噪音消失。

## 3. 测试设计

| 层 | 测试 | 要点 |
|----|------|------|
| 单测 | `report_test.go` 快照更新 | timeseries 5 处 + logs 投影的期望 SQL 改为 `TIMESTAMPDIFF` 形态 |
| 单测 | `TestUnixSecondRendering_TimezoneNeutral` | 表达式形状断言 + Unix 秒常量自洽（1789466400 档真实值） |
| 单测 | `TestParseBoundaryDate`（8 输入） | 738886→2023-01-01、740242→2026-09-18、740243→2026-09-19、字面文本、MAXVALUE、非法值、过小整数（<700000 拒绝）、空串 |
| 单测 | `TestPlanPartitionAdds_IntegerDescriptions` | 全整数回显 / 文本+整数+MAXVALUE 混合两种布局下规划为空（不重复 ADD）；整数反解后 DROP 决策生效 |
| 集成 | `test/integration/tests/report` 组 B 两用例 | `TestTimeSeries_QPS` / `TestLogs_RowShape` 直接断言真实 Unix 秒常量：CST 会话 MySQL 上旧 SQL 必失败、新 SQL 通过（端到端锚点）；删除原 `unixTS(...)` 时区解耦绕法 |

## 4. 发布说明

- 无配置变更、无 DDL 变更、无接口形状变更；
- 修复随下一次 ai-gateway-api 发布生效；已部署环境升级后 `/report/*` 的 `log_time`/桶 `time` 取值自动修正为真实 Unix 秒；
- `EnablePartitionMgmt=true` 的部署升级后 Error 1493 噪音消除，且历史上滞留的超期分区会在下一个 6h 周期被 DROP 清理。
