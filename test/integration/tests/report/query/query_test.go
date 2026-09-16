// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package query_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
)

// 本用例组在真实 MySQL 8.x 上验证 /report/* 五个查询端点。
//
// 数据源：环境变量 REPORT_MYSQL_DSN（格式 user:pass@tcp(host:port)/，参照
// log-reader LR03 的 LR_MYSQL_DSN 约定）。未设置时全部用例 Skip。
// 测试自建专用随机名数据库 report_it_<ns>，套用项目 db_ddl_report_mysql.sql
// （${INIT_DATE} 替换为今天+3 天，历史时间戳落入 p_init 分区），灌确定性种子
// 数据后启动带 [Report] 装配的 api 进程，结束后 DROP DATABASE 清理。

var (
	sm        *testutil.ServerManager
	mysqlDB   *sql.DB // 种子数据库连接
	serverDSN string  // user:pass@tcp(addr)/，用于清理时重建连接
	dbName    string
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("REPORT_MYSQL_DSN")
	if dsn == "" {
		fmt.Println("REPORT_MYSQL_DSN not set; skipping report query integration tests.")
		fmt.Println("To enable: REPORT_MYSQL_DSN=\"root:****@tcp(127.0.0.1:3306)/\" go test ./tests/report/query/")
		os.Exit(0)
	}

	cfg, err := parseServerDSN(dsn)
	if err != nil {
		fmt.Println("invalid REPORT_MYSQL_DSN:", err)
		os.Exit(1)
	}
	serverDSN = fmt.Sprintf("%s:%s@tcp(%s)/", cfg.user, cfg.pass, cfg.addr)

	dbName = fmt.Sprintf("report_it_%d", time.Now().UnixNano())

	// 1. 创建专用数据库并应用 DDL。
	serverDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		fmt.Println("open mysql failed:", err)
		os.Exit(1)
	}
	if _, err := serverDB.Exec("DROP DATABASE IF EXISTS " + dbName); err != nil {
		fmt.Println("pre-drop test database failed:", err)
		serverDB.Close()
		os.Exit(1)
	}
	if _, err := serverDB.Exec("CREATE DATABASE " + dbName); err != nil {
		fmt.Println("create test database failed (check REPORT_MYSQL_DSN privileges):", err)
		serverDB.Close()
		os.Exit(1)
	}
	serverDB.Close()

	mysqlDB, err = sql.Open("mysql", serverDSN+dbName)
	if err != nil {
		cleanupDB()
		fmt.Println("open test database failed:", err)
		os.Exit(1)
	}
	if err := applyReportDDL(mysqlDB); err != nil {
		cleanupDB()
		fmt.Println("apply report ddl failed:", err)
		os.Exit(1)
	}
	if err := seedReportData(mysqlDB); err != nil {
		cleanupDB()
		fmt.Println("seed report data failed:", err)
		os.Exit(1)
	}

	// 2. 注入 [Report] 配置并启动 api 进程（聚合/分区 JOB 关闭，仅验证查询端点）。
	extraTOML := fmt.Sprintf(`
[Databases.report_db]
Driver = "mysql"
DBName = "%s"
Addr = "%s"
User = "%s"
Passwd = %s
MaxOpenConns = 10
MaxIdleConns = 5

[Report]
Backend = "mysql"
Datasource = "report_db"
EnableAggregateJob = false
EnablePartitionMgmt = false
`, dbName, cfg.addr, cfg.user, tomlString(cfg.pass))

	sm, err = testutil.StartServerWithExtraConfig(extraTOML)
	if err != nil {
		cleanupDB()
		fmt.Println("failed to start server:", err)
		os.Exit(1)
	}

	code := m.Run()

	sm.Shutdown()
	mysqlDB.Close()
	cleanupDB()
	os.Exit(code)
}

func cleanupDB() {
	serverDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return
	}
	defer serverDB.Close()
	serverDB.Exec("DROP DATABASE IF EXISTS " + dbName)
}

type serverConfig struct {
	user string
	pass string
	addr string
}

// parseServerDSN 解析 "user:pass@tcp(host:port)/" 形式的服务器级 DSN
// （允许不带库名；忽略 query 参数）。
func parseServerDSN(dsn string) (*serverConfig, error) {
	cfg := &serverConfig{}

	at := strings.LastIndex(dsn, "@")
	if at <= 0 {
		return nil, fmt.Errorf("missing userinfo: %q", dsn)
	}
	userinfo := dsn[:at]
	rest := dsn[at+1:]

	colon := strings.Index(userinfo, ":")
	if colon < 0 {
		cfg.user = userinfo
	} else {
		cfg.user = userinfo[:colon]
		cfg.pass = userinfo[colon+1:]
	}
	if cfg.user == "" {
		return nil, fmt.Errorf("empty user")
	}

	if !strings.HasPrefix(rest, "tcp(") {
		return nil, fmt.Errorf("only tcp(host:port) network is supported: %q", dsn)
	}
	end := strings.Index(rest, ")")
	if end < 0 {
		return nil, fmt.Errorf("malformed network: %q", dsn)
	}
	cfg.addr = rest[len("tcp("):end]
	if cfg.addr == "" {
		return nil, fmt.Errorf("empty addr")
	}
	return cfg, nil
}

// tomlString 生成 TOML 字符串字面量（无特殊字符时用 literal string，避免转义差异）。
func tomlString(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, "'\"\\") {
		return "'" + s + "'"
	}
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// applyReportDDL 套用项目根 db_ddl_report_mysql.sql，${INIT_DATE} 替换为今天+3 天。
func applyReportDDL(db *sql.DB) error {
	ddlPath, err := findRepoFile("db_ddl_report_mysql.sql")
	if err != nil {
		return err
	}
	data, err := os.ReadFile(ddlPath)
	if err != nil {
		return err
	}
	script := strings.ReplaceAll(string(data), "${INIT_DATE}",
		time.Now().AddDate(0, 0, 3).Format("2006-01-02"))

	for _, stmt := range splitSQLStatements(script) {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("apply ddl: %w\nstmt: %s", err, stmt)
		}
	}
	return nil
}

// splitSQLStatements 按分号切分 SQL 脚本（跳过 -- 行注释）。
func splitSQLStatements(script string) []string {
	var stmts []string
	var sb strings.Builder
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		sb.WriteString(line)
		sb.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, strings.TrimSpace(sb.String()))
			sb.Reset()
		}
	}
	if rest := strings.TrimSpace(sb.String()); rest != "" {
		stmts = append(stmts, rest)
	}
	return stmts
}

// findRepoFile 从当前目录向上查找项目根目录中的文件。
func findRepoFile(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	startDir := dir
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found from %s upwards", name, startDir)
		}
		dir = parent
	}
}

func windowQuery() map[string]string {
	return map[string]string{
		"start": fmt.Sprintf("%d", epoch("2026-09-15 09:59:00")),
		"end":   fmt.Sprintf("%d", epoch("2026-09-15 10:05:00")),
	}
}

// epoch 把 UTC 民用时刻转为 Unix 秒（请求参数；驱动按 UTC 格式化回 DATETIME）。
func epoch(datetime string) int64 {
	tt, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, time.UTC)
	if err != nil {
		panic(err)
	}
	return tt.Unix()
}

func with(base map[string]string, kv ...string) map[string]string {
	out := make(map[string]string, len(base)+len(kv)/2)
	for k, v := range base {
		out[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		out[kv[i]] = kv[i+1]
	}
	return out
}

// ---------------------------------------------------------------------------
// 响应数据结构（与 model/ireport 的 JSON 形状一致）
// ---------------------------------------------------------------------------

type overviewData struct {
	RequestTotal int64   `json:"request_total"`
	ErrorTotal   int64   `json:"error_total"`
	ErrorRate    float64 `json:"error_rate"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`

	LatencyAvgMs float64  `json:"latency_avg_ms"`
	LatencyMaxMs float64  `json:"latency_max_ms"`
	LatencyP50Ms *float64 `json:"latency_p50_ms"`

	TtftAvgMs float64 `json:"ttft_avg_ms"`
	TpotAvgMs float64 `json:"tpot_avg_ms"`

	Cost []struct {
		Currency string `json:"currency"`
		Value    int64  `json:"value"`
	} `json:"cost"`

	RateLimitHits int64 `json:"rate_limit_hits"`
	AuthRejects   int64 `json:"auth_rejects"`
	LogsTotal     int64 `json:"logs_total"`
}

type metricPoint struct {
	Time   int64    `json:"time"`
	Value  *float64 `json:"value"`
	Input  *float64 `json:"input"`
	Output *float64 `json:"output"`
	Total  *float64 `json:"total"`
}

type timeseriesData struct {
	BucketSec int           `json:"bucket_sec"`
	Series    []metricPoint `json:"series"`
}

type rankingItem struct {
	Name         string `json:"name"`
	RequestCount int64  `json:"request_count"`
	ErrorCount   int64  `json:"error_count"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
}

type rankingsData struct {
	Items []rankingItem `json:"items"`
}

type distItem struct {
	Name         string  `json:"name"`
	RequestCount int64   `json:"request_count"`
	Ratio        float64 `json:"ratio"`
}

type distributionData struct {
	Items []distItem `json:"items"`
}

type logItem struct {
	LogID          *int64  `json:"logid"`
	LogTime        int64   `json:"log_time"`
	Hostid         *string `json:"hostid"`
	Product        *string `json:"product"`
	APIKeyID       *string `json:"ai_apikey_id"`
	RequestedModel *string `json:"ai_requested_model"`
	TargetModel    *string `json:"ai_target_model"`
	Provider       *string `json:"ai_provider"`
	Protocol       *string `json:"ai_protocol"`
	Mode           *string `json:"ai_mode"`
	Stream         *int16  `json:"ai_stream"`
	StatusCode     *int16  `json:"res_status_code"`
	ErrCode        *string `json:"err_code"`
	ErrMsg         *string `json:"err_msg"`

	RateLimitHits        *string `json:"ai_rate_limit_hits"`
	AuthRejectQuotaPlans *string `json:"ai_auth_reject_quota_plans"`
	Level1Name           *string `json:"level1Name"`
	Level1               *string `json:"level1"`
	ClientIP             *string `json:"client_ip"`
	HeaderHost           *string `json:"header_host"`
	OriginURI            *string `json:"origin_uri"`
	ReqHeaders           *string `json:"req_headers"`
	ResHeaders           *string `json:"res_headers"`
}

type logsData struct {
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Items    []logItem `json:"items"`
}
