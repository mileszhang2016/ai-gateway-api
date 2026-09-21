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

package partition_test

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 分区断言辅助
// ---------------------------------------------------------------------------

var dateInTextRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// toDays 向 MySQL 询问 TO_DAYS('YYYY-MM-DD')，避免在 Go 侧复刻历法换算。
func toDays(t *testing.T, day time.Time) int {
	t.Helper()
	var n int
	require.NoError(t, mysqlDB.QueryRow("SELECT TO_DAYS(?)", day.Format("2006-01-02")).Scan(&n))
	return n
}

// descToDays 解析 information_schema 回显的 PARTITION_DESCRIPTION：
// MySQL 8.x 对 TO_DAYS 分区回显求值整数；回显文本形态时取其中的日期再
// 向 MySQL 求 TO_DAYS。
func descToDays(t *testing.T, desc string) (int, bool) {
	t.Helper()
	if n, err := strconv.Atoi(strings.TrimSpace(desc)); err == nil {
		return n, true
	}
	d := dateInTextRe.FindString(desc)
	if d == "" {
		return 0, false
	}
	var n int
	if err := mysqlDB.QueryRow("SELECT TO_DAYS(?)", d).Scan(&n); err != nil {
		return 0, false
	}
	return n, true
}

// listPartitions 查询目标库一张表的全部分区（分区名 -> 边界 TO_DAYS 值）。
func listPartitions(t *testing.T, table string) map[string]int {
	t.Helper()
	rows, err := mysqlDB.Query(
		"SELECT PARTITION_NAME, PARTITION_DESCRIPTION"+
			" FROM information_schema.PARTITIONS"+
			" WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND PARTITION_NAME IS NOT NULL"+
			" ORDER BY PARTITION_ORDINAL_POSITION", dbName, table)
	require.NoError(t, err)
	defer rows.Close()

	parts := map[string]int{}
	for rows.Next() {
		var name, desc string
		require.NoError(t, rows.Scan(&name, &desc))
		n, ok := descToDays(t, desc)
		require.True(t, ok, "unparseable PARTITION_DESCRIPTION %q of partition %s", desc, name)
		parts[name] = n
	}
	require.NoError(t, rows.Err())
	return parts
}

type partitionWant struct {
	name     string
	boundary int
}

// aheadWants 计算前瞻窗口 [今天, 今天+3) 应有的三个分区及其边界：
// p<今天+i> 的边界为 TO_DAYS(今天+i+1)。
func aheadWants(t *testing.T) []partitionWant {
	t.Helper()
	var wants []partitionWant
	for i := 0; i < 3; i++ {
		day := today.AddDate(0, 0, i)
		wants = append(wants, partitionWant{
			name:     "p" + day.Format("20060102"),
			boundary: toDays(t, day.AddDate(0, 0, 1)),
		})
	}
	return wants
}

func hasAhead(parts map[string]int, wants []partitionWant) bool {
	for _, w := range wants {
		if parts[w.name] != w.boundary {
			return false
		}
	}
	return true
}

// waitForAhead 轮询等待两表分区就绪。JOB 在 Start() 时立即跑一轮周期，
// 但服务器就绪（TCP 可拨号）与该周期完成之间无时序保证，故轮询而非定长
// sleep；已就绪时立即返回。
func waitForAhead(t *testing.T) {
	t.Helper()
	wants := aheadWants(t)
	deadline := time.Now().Add(30 * time.Second)
	for {
		detail := listPartitions(t, "bfe_ai_request_log")
		metrics := listPartitions(t, "bfe_ai_metrics_1m")
		if hasAhead(detail, wants) && hasAhead(metrics, wants) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("partition job did not create the ahead window within 30s;"+
				" bfe_ai_request_log=%v bfe_ai_metrics_1m=%v want=%+v",
				detail, metrics, wants)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// 用例
// ---------------------------------------------------------------------------

// C-1 启动即建分区：JOB 首轮周期应为今天..今天+2 建分区；两表同验
// （bfe_ai_metrics_1m 与明细表由同一 JOB 循环维护，issue #191 中同病，
// 其写入方是聚合 JOB 自身，越界后同样 Error 1526）。
// 修复前 JOB 误判两表未分区、静默走 DELETE 兜底，本用例恒超时失败。
func TestPartitionJob_CreatesAheadPartitions(t *testing.T) {
	waitForAhead(t)

	// p_init（建表时 INIT_DATE=昨天）保留：边界 TO_DAYS(昨天)，前瞻窗口
	// 内分区未到期（RetentionDays=7）。
	initBoundary := toDays(t, today.AddDate(0, 0, -1))
	assert.Equal(t, initBoundary, listPartitions(t, "bfe_ai_request_log")["p_init"], "detail p_init boundary")
	assert.Equal(t, initBoundary, listPartitions(t, "bfe_ai_metrics_1m")["p_init"], "metrics p_init boundary")
}

// C-2 覆盖期内写入成功：模拟 log-reader 写入路径，锁定 issue #191 的
// 下游症状（Error 1526）消除。
func TestPartitionJob_WritesWithinAheadWindow(t *testing.T) {
	waitForAhead(t)

	within := []time.Time{
		today,
		today.AddDate(0, 0, 2).Add(23*time.Hour + 59*time.Minute + 59*time.Second),
	}
	for _, ts := range within {
		_, err := mysqlDB.Exec(
			"INSERT INTO bfe_ai_request_log (hostid, log_time) VALUES ('it-host', ?)",
			ts.Format("2006-01-02 15:04:05"))
		require.NoError(t, err, "insert into bfe_ai_request_log at %s", ts.Format("2006-01-02 15:04:05"))

		_, err = mysqlDB.Exec(
			"INSERT INTO bfe_ai_metrics_1m (ts_min) VALUES (?)",
			ts.Format("2006-01-02 15:04:05"))
		require.NoError(t, err, "insert into bfe_ai_metrics_1m at %s", ts.Format("2006-01-02 15:04:05"))
	}
}

// C-3 边界外写入拒绝：今天+3 超出 JOB 保有的 3 天前瞻窗口，写入必须以
// Error 1526 失败——固定该契约，防止未来误调小前瞻窗口而无感知。
func TestPartitionJob_RejectsBeyondAheadWindow(t *testing.T) {
	waitForAhead(t)

	ts := today.AddDate(0, 0, 3).Format("2006-01-02 15:04:05")

	_, err := mysqlDB.Exec(
		"INSERT INTO bfe_ai_request_log (hostid, log_time) VALUES ('it-host', ?)", ts)
	require.Error(t, err, "log_time %s must fall into no partition", ts)
	var me *mysql.MySQLError
	require.True(t, errors.As(err, &me), "expect *mysql.MySQLError, got %T: %v", err, err)
	assert.Equal(t, uint16(1526), me.Number)

	_, err = mysqlDB.Exec("INSERT INTO bfe_ai_metrics_1m (ts_min) VALUES (?)", ts)
	require.Error(t, err, "ts_min %s must fall into no partition", ts)
	require.True(t, errors.As(err, &me), "expect *mysql.MySQLError, got %T: %v", err, err)
	assert.Equal(t, uint16(1526), me.Number)
}
