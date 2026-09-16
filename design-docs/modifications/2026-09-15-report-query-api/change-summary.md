# 报表查询 API（Report Query API）变更摘要

> 本文档覆盖报表查询 API 特性中 ai-gateway-api 仓库的改动；log-reader 侧见 `log-reader/doc/modifications/2026-09-15-add-mod-log-mysql-plugin/`

## 1. 背景

当前数据报表链路为 `BFE → pb 日志 → log-reader(mod_kafka) → Kafka → Doris → Grafana`，存在三个痛点：

1. **组件重**：Kafka、Doris、Grafana 缺一不可，小部署/私有化场景的运维成本远超网关本身；
2. **Grafana 依赖**：报表展示依赖独立 Grafana，与网关控制台体验割裂；
3. **无自有 API**：报表数据没有 REST API，无法集成到 ai-gateway-web，也无法被第三方系统消费。

v0.7 方案形成两种部署形态：**轻量形态**（log-reader mod_log_mysql 插件 → MySQL → ai-gateway-api → ai-gateway-web）与**标准形态**（存量 Kafka → Doris 链路保留）。ai-gateway-api 在本特性中承担：

- 提供**统一的报表查询 API**（总览/时序/排行/分布/明细五类端点），报表直接在 ai-gateway-web 展示；
- 查询层通过配置切换 **MySQL / Doris 两种后端**，两种形态共用一套 API 与前端；
- 持有两张 MySQL 报表表的 **DDL**（`db_ddl_report_mysql.sql`），并运行 **分钟聚合 JOB** 与 **分区管理 JOB**（MySQL 形态）。

## 2. 目标

- 新增 `/open-api/v1/report/*` 五个端点，支持按时间窗/模型/API Key/提供商/Host/状态码/流式等过滤；
- 各接口返回口径与现有 Grafana Dashboard 面板口径一致（同窗口数据比对一致）；
- 同一套接口在 MySQL 与 Doris 两种后端配置下返回结构一致的数据；
- MySQL 侧提供与 Doris 对齐的分钟级预聚合能力（时序/下钻类查询读聚合表，不扫明细大表）。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（本仓库）；log-reader / ai-gateway-web 改动见各自仓库文档 |
| 新增模块 | `endpoints/openapi_v1/report/`、`model/ireport/`、`storage/mysqlreport/`、`storage/dorisreport/` |
| 修改模块 | `stateful/`（`[Report]` 配置段、数据源装配）、`model/iauth/features.go`（新增 Feature）、`endpoints/openapi_v1/endpoints.go`（路由注册） |
| 新增文件 | `db_ddl_report_mysql.sql`（两张报表表 DDL，明细表 89 列 + 聚合表 37 维 + 24 指标） |
| 涉及接口 | 新增 `GET /open-api/v1/report/overview` / `timeseries` / `rankings` / `distribution` / `logs` |
| 数据迁移 | 新增两张报表表；**不影响**既有 `db_ddl.sql` 体系（报表库独立于控制面库） |
| 数据面影响 | 无（BFE / conf-agent 零改动） |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 查询层抽象 `ReportStorager` 接口，MySQL/Doris 各一套 SQL | 方言有差异（时间桶函数、分位数、UNNEST/JSON_TABLE），无法用同一套 SQL 通吃；Manager 只做参数校验与口径组装 |
| Doris 后端复用现有 `Databases` 数据源（MySQL 协议驱动连 FE 9030） | 复用现有连接池/配置体系，不引入新客户端 |
| MySQL 预聚合用 api 内定时 JOB（DELETE 窗口 + INSERT SELECT 事务） | 与 Doris INSERT JOB 语义对称（同一 SQL 逻辑）；窗口重放幂等；log-reader 保持无状态 |
| 时序/排行/分布只读聚合表，仅明细查询读明细表 | 对齐 Doris 形态，避免扫描明细大表 |
| DDL 归本仓库 `db_ddl_report_mysql.sql`，插件不自动建表 | ai-gateway-api 对 schema 依赖面最广（聚合 JOB / 分区管理 JOB / 查询），且是唯一有 DDL 管理传统的仓库；log-reader 账号按最小权限仅 INSERT/UPDATE |
| MySQL 形态无原生分位数 | P50/P90/P99 仅 Doris 后端返回；MySQL 后端 v1 只出 avg/max，由返回字段有无决定前端降级展示 |
| 鉴权新增 Feature `FeatureReport` | System scope 给 `ReadAll`（当前 web 仅管理员）；Product scope 给 `Read` 为租户自助报表预留 |
| `[Report]` 配置缺省时模块不装配 | 端点 404，纯增量发布，不影响现有接口 |

## 5. 关联文档

- `ai-gateway-api/design-docs/modifications/2026-09-15-report-query-api/api-changes.md`
- `ai-gateway-api/design-docs/modifications/2026-09-15-report-query-api/design-changes.md`
- log-reader 侧：`log-reader/doc/modifications/2026-09-15-add-mod-log-mysql-plugin/design-changes.md`（写入插件；其集成测试 LR03 验证了本文档 DDL 的三处修正）
