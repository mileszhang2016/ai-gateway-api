# 报表查询 API（Report Query API）设计变更

> 接口定义见同目录 `api-changes.md`

## 1. 代码结构（沿用现有分层惯例）

```
endpoints/openapi_v1/report/          # 路由与 Handler（xreq.Endpoint 声明式注册）
├── overview.go                       # GET /report/overview
├── timeseries.go                     # GET /report/timeseries
├── rankings.go                       # GET /report/rankings
├── distribution.go                   # GET /report/distribution
└── logs.go                           # GET /report/logs

model/ireport/
├── types.go                          # Filter/LogFilter/Result 类型 + ReportStorager 接口
└── manager.go                        # ReportManager：参数绑定校验、bucket 计算、口径常量、委托 storager

storage/mysqlreport/                  # MySQL 实现（gendry builder，DAO 惯例）
├── report.go                         # 五类查询的 SQL 构建与扫描
└── job.go                            # 分钟聚合 JOB + 分区管理 JOB

storage/dorisreport/                  # Doris 实现（同样走 database/sql + mysql 协议驱动）
└── report.go

stateful/
├── config.go                         # 新增 [Report] 配置段解析
├── config_database.go                # 复用 Databases map，无需改动
└── container/rdb/components.go       # 按 [Report].Backend 装配 storager

model/iauth/features.go               # 新增 FeatureReport + scope2permission 映射
endpoints/openapi_v1/endpoints.go     # 注册 /report/* 路由（Report 模块未装配时整体不注册）
```

## 2. 数据源与配置

复用现有 `Databases` map（`stateful/config_database.go` 已支持多数据源，`DbGet(name)` 按名取）。MySQL 与 Doris 都只是 map 里的一个条目；Doris 走 FE 的 MySQL 协议端口（9030），不引入新客户端。

`conf/ai_gateway_api.toml` 新增：

```toml
# 轻量形态
[Databases.report_db]
Driver = "mysql"
DBName = "bfe_report"
Addr = "127.0.0.1:3306"
User = "report_read"          # 查询账号，仅 SELECT + 聚合 JOB 所在实例需要 DELETE/INSERT
Passwd = "******"
MaxOpenConns = 20
MaxIdleConns = 5

# 标准形态（Doris FE 的 MySQL 协议端口）
[Databases.doris_db]
Driver = "mysql"
DBName = "bfe_observability"
Addr = "172.18.1.244:9030"
User = "root"
Passwd = ""
MaxOpenConns = 20

[Report]
Backend = "mysql"             # mysql | doris；缺省时 Report 模块整体不装配，端点 404
Datasource = "report_db"      # Databases 中的数据源名（doris 时填 doris_db）
Database = ""                 # 库名覆盖（可选，默认取数据源的 DBName）
EnableAggregateJob = true     # backend=mysql 时启用分钟聚合 JOB
AggregateIntervalSec = 60
RetentionDays = 7             # 明细保留天数（分区 DROP / DELETE 窗口）
EnablePartitionMgmt = true    # 分区管理 JOB（非分区表自动降级为 DELETE）
```

账号权限矩阵（MySQL 形态，最小权限原则）：

| 账号 | 用途 | 权限 |
|------|------|------|
| log-reader 写入账号 | 插件直写明细表 | report 库明细表 INSERT/UPDATE（幂等覆盖需要 UPDATE），无 DDL/DELETE |
| api 查询/聚合账号 | 查询 API + 聚合/分区 JOB | report 库明细表/聚合表 SELECT、INSERT/DELETE（聚合表）、明细表 DELETE（非分区降级清理）、ALTER TABLE（分区管理） |

## 3. DDL：`db_ddl_report_mysql.sql`（新增文件）

两张表统一放本仓库发布，部署流程执行；**schema-first**（先 DDL 后启用 log-reader 插件）。插件不自动建表。

### 3.1 明细表 `bfe_ai_request_log`（89 列，与 Doris 同名同列）

与 Doris 明细表同名同列：`ARRAY<STRUCT<...>>` 列转 `JSON`（7 个：req_headers、res_headers、ai_route_rule_hits、ai_cluster_key_names、ai_rate_limit_hits、两个 quota plans），`log_time` 用 `DATETIME`。

**三处修正已经过 log-reader LR03 集成测试（真实 MySQL 8.4）验证**，实现时严格照此落地：

1. **长文本列用 TEXT**：origin_uri/final_uri/x_forward_for/authorization/referrer/user_agent/cookie/res_location 若按 Doris 的 VARCHAR 长度建列，utf8mb4 下超过 MySQL 65535 字节行上限（Error 1118）；
2. **幂等键字符串列可空**：`ai_apikey_id`/`ai_requested_model` 允许 NULL——写入侧零值规则把空字符串映射为 NULL（未认证请求即如此），`NOT NULL` 使这类写入整批失败（Error 1048）；`hostid`/`log_time` 由 log-reader 构造保证非空，保持 `NOT NULL`；副作用：NULL 不参与唯一键去重（与 Doris UNIQUE KEY 对 NULL 语义一致），可接受；
3. **按天 RANGE 分区**：MySQL 无动态分区，分区管理 JOB 必须先于数据到达建好分区（Error 1526 防护见 §5）。

```sql
-- 要求 MySQL >= 5.7.8（JSON 列），推荐 8.0
CREATE TABLE IF NOT EXISTS bfe_ai_request_log (
    hostid                  VARCHAR(256)  NOT NULL DEFAULT '',
    log_time                DATETIME      NOT NULL,
    ai_apikey_id            VARCHAR(256)  DEFAULT NULL,
    ai_requested_model      VARCHAR(128)  DEFAULT NULL,
    logid                   BIGINT        DEFAULT NULL,
    product                 VARCHAR(64)   DEFAULT NULL,
    log_tag                 VARCHAR(64)   DEFAULT NULL,
    -- 客户端连接
    client_ip               VARCHAR(64)   DEFAULT NULL,
    client_network          VARCHAR(16)   DEFAULT NULL,
    is_trust_src_ip         TINYINT       DEFAULT NULL,
    req_num                 INT           DEFAULT NULL,
    session_id              BIGINT        DEFAULT NULL,
    bfe_ip                  VARCHAR(64)   DEFAULT NULL,
    sock_src_ip             VARCHAR(64)   DEFAULT NULL,
    vip                     VARCHAR(64)   DEFAULT NULL,
    vip6                    VARCHAR(128)  DEFAULT NULL,
    -- 错误
    err_code                VARCHAR(64)   DEFAULT NULL,
    err_msg                 VARCHAR(512)  DEFAULT NULL,
    -- 请求（长文本列用 TEXT，见修正 1）
    proto                   VARCHAR(16)   DEFAULT NULL,
    header_host             VARCHAR(256)  DEFAULT NULL,
    origin_uri              TEXT          DEFAULT NULL,
    final_uri               TEXT          DEFAULT NULL,
    method                  VARCHAR(16)   DEFAULT NULL,
    content_type            VARCHAR(128)  DEFAULT NULL,
    x_forward_for           TEXT          DEFAULT NULL,
    accept_language         VARCHAR(256)  DEFAULT NULL,
    authorization           TEXT          DEFAULT NULL,
    transfer_encoding       VARCHAR(64)   DEFAULT NULL,
    referrer                TEXT          DEFAULT NULL,
    user_agent              TEXT          DEFAULT NULL,
    delegation              VARCHAR(256)  DEFAULT NULL,
    uid                     VARCHAR(256)  DEFAULT NULL,
    cookie                  TEXT          DEFAULT NULL,
    req_headers             JSON          DEFAULT NULL,
    req_header_len          INT           DEFAULT NULL,
    req_body_len            INT           DEFAULT NULL,
    -- 路由
    cluster                 VARCHAR(256)  DEFAULT NULL,
    sub_cluster             VARCHAR(256)  DEFAULT NULL,
    backend_info            VARCHAR(256)  DEFAULT NULL,
    backend_retry           TINYINT       DEFAULT NULL,
    -- 响应
    res_status_code         SMALLINT      DEFAULT NULL,
    res_header_len          INT           DEFAULT NULL,
    res_body_len            INT           DEFAULT NULL,
    res_content_type        VARCHAR(128)  DEFAULT NULL,
    res_location            TEXT          DEFAULT NULL,
    res_transfer_encoding   VARCHAR(64)   DEFAULT NULL,
    res_headers             JSON          DEFAULT NULL,
    -- 耗时（毫秒）
    all_time                INT           DEFAULT NULL,
    read_client_time        INT           DEFAULT NULL,
    cluster_serve_time      INT           DEFAULT NULL,
    backend_serve_time      INT           DEFAULT NULL,
    write_client_time       INT           DEFAULT NULL,
    connect_backend_time    INT           DEFAULT NULL,
    proxy_delay_time        INT           DEFAULT NULL,
    session_offset_time     INT           DEFAULT NULL,
    -- API Key 标签（写入时由 log-reader 从 ai_apikeytags 打平）
    level1Name              VARCHAR(128)  DEFAULT NULL,
    level1                  VARCHAR(128)  DEFAULT NULL,
    level2Name              VARCHAR(128)  DEFAULT NULL,
    level2                  VARCHAR(128)  DEFAULT NULL,
    level3Name              VARCHAR(128)  DEFAULT NULL,
    level3                  VARCHAR(128)  DEFAULT NULL,
    level4Name              VARCHAR(128)  DEFAULT NULL,
    level4                  VARCHAR(128)  DEFAULT NULL,
    level5Name              VARCHAR(128)  DEFAULT NULL,
    level5                  VARCHAR(128)  DEFAULT NULL,
    -- AI 可观测
    ai_target_model         VARCHAR(128)  DEFAULT NULL,
    ai_stream               TINYINT       DEFAULT NULL,
    ai_input_tokens         BIGINT        DEFAULT NULL,
    ai_output_tokens        BIGINT        DEFAULT NULL,
    ai_total_tokens         BIGINT        DEFAULT NULL,
    ai_cache_read_tokens    BIGINT        DEFAULT NULL,
    ai_cache_write_tokens   BIGINT        DEFAULT NULL,
    ai_audio_input_tokens   BIGINT        DEFAULT NULL,
    ai_audio_output_tokens  BIGINT        DEFAULT NULL,
    ai_image_count          BIGINT        DEFAULT NULL,
    ai_ttft_us              BIGINT        DEFAULT NULL,
    ai_tpot_us              BIGINT        DEFAULT NULL,
    ai_provider             VARCHAR(64)   DEFAULT NULL,
    ai_protocol             VARCHAR(64)   DEFAULT NULL,
    ai_mode                 VARCHAR(64)   DEFAULT NULL,
    ai_retry_count          INT           DEFAULT NULL,
    ai_cost_value           BIGINT        DEFAULT NULL,
    ai_cost_currency        VARCHAR(16)   DEFAULT NULL,
    ai_route_rule_hits      JSON          DEFAULT NULL,
    ai_cluster_key_names    JSON          DEFAULT NULL,
    ai_rate_limit_hits      JSON          DEFAULT NULL,
    ai_auth_reject_reason   VARCHAR(256)  DEFAULT NULL,
    ai_auth_reject_quota_plans JSON       DEFAULT NULL,
    ai_auth_hit_quota_plans JSON          DEFAULT NULL,
    -- 幂等键：与 Doris UNIQUE KEY 一致，log-reader 重发/-b 补读安全
    UNIQUE KEY uk_dedup (hostid, log_time, ai_apikey_id, ai_requested_model),
    KEY idx_model_time (ai_target_model, log_time),
    KEY idx_apikey_time (ai_apikey_id, log_time),
    KEY idx_provider_time (ai_provider, log_time),
    KEY idx_host_time (hostid, log_time),
    KEY idx_status_time (res_status_code, log_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
PARTITION BY RANGE (TO_DAYS(log_time)) (
    PARTITION p_init VALUES LESS THAN (TO_DAYS('${INIT_DATE}'))
);
```

要点：

- **唯一键长度** `(256+256+128)*4+5 ≈ 2565` 字节 < 3072 InnoDB 上限（DYNAMIC 行格式），无需前缀索引；
- **JSON 列只用于明细页展示**，不做查询过滤条件（避免生成列/函数索引的兼容性问题）；
- 建表时 `p_init` 的 `${INIT_DATE}` 由部署流程填为建表当日之后（建议 +3 天），后续分区由 JOB 维护。

### 3.2 聚合表 `bfe_ai_metrics_1m`（37 维 + 24 指标）

维度集合与指标列与 Doris `bfe_ai_metrics_1m` 完全一致（Doris DDL 见 `ai-gateway-observability/doris/sqls/bfe_ai_metrics_1m.sql`）。MySQL 形态为普通 InnoDB 表：

- **不设唯一键/主键**：37 个维度列（含多个 VARCHAR(256)）无法构成 InnoDB 唯一键（3072 字节上限），且幂等性由聚合 JOB 的「DELETE 窗口 + INSERT SELECT 事务」保证，不依赖约束；InnoDB 隐式 rowid 作为主键。
- 索引：`KEY idx_ts (ts_min)` + 热维度列索引（`ai_target_model`、`ai_apikey_id`、`ai_provider` + ts_min），与明细表索引策略一致。
- 同样按 `ts_min` RANGE 分区按天滚动，由分区管理 JOB 维护；非分区形态降级 `DELETE ... LIMIT` 分批。
- 维度 NULL 归一：聚合 JOB 写入时 `IFNULL(col,'')`（与 Doris JOB 的 COALESCE 语义一致），保证 GROUP BY 与过滤谓词行为对齐 Doris。

## 4. ReportStorager 接口与 Manager 职责

```go
type Filter struct {
    Start, End  time.Time // 必填，时间窗（≤7d）
    Models      []string  // ai_target_model
    ApikeyIDs   []string
    Providers   []string
    Hosts       []string
    Stream      *bool
    StatusCodes []int
}

type ReportStorager interface {
    Overview(ctx context.Context, f *Filter) (*OverviewResult, error)
    TimeSeries(ctx context.Context, metric string, f *Filter, bucketSec int) ([]*MetricPoint, error)
    Rankings(ctx context.Context, dimension string, f *Filter, limit int) ([]*RankingItem, error)
    Distribution(ctx context.Context, dimension string, f *Filter) ([]*DistItem, error)
    Logs(ctx context.Context, f *LogFilter) (*LogQueryResult, error)
}
```

Manager（`model/ireport/manager.go`）：

- 参数绑定校验：`go-playground/validator`（同 operation-logs 惯例）；metric/dimension 枚举校验；窗口 ≤7 天；
- bucket 计算：≤6h→60s，≤3d→300s，≤7d→1800s；
- 指标口径常量（见 `api-changes.md` §2.1）；维度枚举映射；
- 委托 storager 执行，自身不含 SQL。

装配（`stateful/container/rdb/components.go`）：`[Report].Backend = "mysql"` 时装配 `mysqlreport.New(DbGet(Datasource))` 并启动 JOB；`= "doris"` 时装配 `dorisreport.New(...)`，不启动 JOB（Doris 的聚合由既有 INSERT JOB 完成）；配置缺省则 Report 组件整体不装配，`/report/*` 路由不注册（404）。

## 5. 后台 JOB（`storage/mysqlreport/job.go`，仅 backend=mysql）

Go ticker 内嵌单协程，`AggregateIntervalSec` 周期；多副本部署用 MySQL `GET_LOCK('report_agg_job', 0)` 抢锁防重（连接同一个 MySQL 即可）。

**聚合 JOB**（每分钟）：

```
BEGIN
  DELETE FROM bfe_ai_metrics_1m WHERE ts_min = <上一整分钟>;      -- 窗口重放幂等
  INSERT INTO bfe_ai_metrics_1m
    SELECT DATE_FORMAT(log_time, '%Y-%m-%d %H:%i:00') AS ts_min,
           IFNULL(hostid,''), IFNULL(ai_apikey_id,''), ...          -- 维度 COALESCE 归一
           COUNT(1),
           SUM(err_code != '' AND err_code IS NOT NULL),
           SUM(ai_input_tokens), ...                                 -- 24 指标
    FROM bfe_ai_request_log
    WHERE log_time >= <上一分钟起点> AND log_time < <本分钟起点>
    GROUP BY ts_min, <全部 37 维度>;
COMMIT
```

边界：JOB 启动后首个周期补聚合「上一整分钟」；进程重启不补历史窗口（接受 ≤1 分钟空洞，与 Doris INSERT JOB 语义对称）。

**分区管理 JOB**（每 6h）：向前建 3 天分区、`DROP PARTITION` 超 `RetentionDays` 的；`information_schema.PARTITIONS` 探测目标表为非分区形态时自动降级为 `DELETE ... WHERE log_time < 阈值 LIMIT 分批`。

## 6. 两后端 SQL 差异

| 点 | Doris | MySQL |
|----|-------|-------|
| 时间桶 | `DATE_TRUNC(ts_min, 'minute')` / `FLOOR(EPOCH(hour)/bucket)` | `FROM_UNIXTIME(FLOOR(UNIX_TIMESTAMP(ts_min)/?)*?)` |
| 分位数延迟 | `PERCENTILE_APPROX(all_time, 0.5/0.9/0.99)`（明细表） | 无原生分位数：v1 只返回 avg/max（分位数近似方案待 v1 上线后按反馈决策） |
| 限流命中展开 | `CROSS JOIN UNNEST(ai_rate_limit_hits)` | MySQL 8.0 可 `JSON_TABLE`（v1 明细页不展开，仅展示 JSON 原文） |
| 过滤空值 | `ai_provider != ''` | 同（写入侧 NULL 已在聚合 JOB 归一为 ''，明细过滤用 `IFNULL(ai_provider,'') != ''`） |

## 7. 鉴权

- `model/iauth/features.go` 新增 `FeatureReport`；
- `scope2permission`：System scope → `ReadAll`（当前 web 仅管理员，报表为管理员功能）；Product scope → `Read`（为租户自助用量报表预留）；
- 所有端点 `Authorizer: iauth.FA(iauth.FeatureReport, iauth.ActionReadAll)`，无权限 402。

## 8. 测试计划

| 层 | 测试 | 要点 |
|----|------|------|
| 单元 | manager 参数校验/bucket 计算；两后端 SQL builder | 方言模板快照测试；JOB 窗口边界（跨分钟/跨天/空窗口） |
| 接口 | 两后端各起测试库灌 demo 数据（复用 `doris/demo/*.json` 同源构造） | 五端点结构与口径比对；分页/过滤/空结果/越权（402）；未装配 404 |
| 数据一致性 | 同一份 pb 日志同时跑「mod_log_mysql→MySQL」与「mod_kafka→Doris」，API 与 Grafana 同窗口比对 | QPS/Token/错误率误差为 0（同一份源数据） |
| 聚合 JOB | 分钟边界、窗口重放（重复执行同一周期结果不变）、GET_LOCK 单实例 | DELETE+INSERT 事务幂等 |

`make test-model-cover-gate`（model/ 覆盖率 ≥70%）需包含 `model/ireport`。

## 9. 发布顺序

1. 本仓库发布含 DDL（`db_ddl_report_mysql.sql`）、Report 模块的 api 版本——`[Report]` 缺省时纯增量，端点 404，不影响现有接口；
2. 部署流程执行 DDL（schema-first），建表 `p_init` 覆盖建表日后 3 天；
3. 启用 log-reader `mod_log_mysql` 插件（log-reader 侧独立发布，默认不启用）；
4. 配置 `[Report].Backend = "mysql"` 装配查询与 JOB；标准形态配置 `Backend = "doris"` 接存量 Doris；
5. ai-gateway-web 报表页**同步发布**（老前端 + 未装配新后端时菜单不含该项）。

## 10. 风险（api 侧）

- **聚合 JOB 多副本**：`GET_LOCK` 防重；多实例独立 MySQL 时各跑各的，无冲突；
- **MySQL 容量**：建议 MySQL 形态目标量级 ≤ 日均百万级；超量走 Doris 形态；
- **分区先于数据**：分区管理 JOB 必须先于数据到达建好分区（Error 1526）；JOB 每 6h 巡检 + 启动时立即补建一次；
- **分位数缺失**：MySQL 后端无 P50/P90/P99（待上线后按反馈决策是否做明细抽样近似）；
- **双链路并存口径**：同一集群同时启用两种落库形态时两套报表不一致（log-reader 无持久化位点，既有短板）；部署文档明确单集群单形态。
