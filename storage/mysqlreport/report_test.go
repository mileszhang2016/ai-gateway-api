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

package mysqlreport

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ireport"
)

var (
	testStart = time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	testEnd   = time.Date(2026, 9, 15, 11, 0, 0, 0, time.UTC)
)

func fullFilter() *ireport.Filter {
	stream := true
	return &ireport.Filter{
		Start:       testStart,
		End:         testEnd,
		Models:      []string{"gpt-4o", "gpt-4"},
		ApikeyIDs:   []string{"key-1"},
		Providers:   []string{"openai"},
		Hosts:       []string{"gw-01"},
		Stream:      &stream,
		StatusCodes: []int{200, 500},
	}
}

// assertSnapshot compares the builder output against the expected SQL and
// args. The expected SQL is the verbatim gendry rendering: conditions are
// grouped by operator (=, IN, !=, >=, <, LIKE) and lexicographically sorted
// within a group, and the whole WHERE clause is wrapped in parentheses.
func assertSnapshot(t *testing.T, query string, args []interface{}, err error,
	expectSQL string, expectArgs []interface{}) {
	t.Helper()
	require.NoError(t, err)
	assert.Equal(t, expectSQL, query)
	assert.Equal(t, expectArgs, args)
}

func fullFilterArgs() []interface{} {
	return []interface{}{
		int8(1), "key-1", "openai", "gpt-4o", "gpt-4", "gw-01", 200, 500,
		testStart, testEnd,
	}
}

const fullWhere = " WHERE (ai_stream=? AND ai_apikey_id IN (?) AND ai_provider IN (?)" +
	" AND ai_target_model IN (?,?) AND hostid IN (?) AND res_status_code IN (?,?)" +
	" AND ts_min>=? AND ts_min<?)"

const overviewFields = "SELECT IFNULL(SUM(request_count),0) AS request_total," +
	"IFNULL(SUM(error_count),0) AS error_total," +
	"IFNULL(SUM(input_tokens),0) AS input_tokens," +
	"IFNULL(SUM(output_tokens),0) AS output_tokens," +
	"IFNULL(SUM(total_tokens),0) AS total_tokens," +
	"IFNULL(SUM(all_time_sum),0) AS all_time_sum," +
	"IFNULL(MAX(all_time_sum/request_count),0) AS latency_max," +
	"IFNULL(SUM(ttft_us_sum),0) AS ttft_us_sum," +
	"IFNULL(SUM(CASE WHEN ai_stream=1 THEN request_count ELSE 0 END),0) AS stream_requests," +
	"IFNULL(SUM(tpot_us_sum),0) AS tpot_us_sum," +
	"IFNULL(SUM(rate_limit_hits),0) AS rate_limit_hits," +
	"IFNULL(SUM(auth_reject_count),0) AS auth_rejects"

func TestBuildOverviewMetricsSQL_FullFilter(t *testing.T) {
	query, args, err := buildOverviewMetricsSQL("bfe_ai_metrics_1m", fullFilter())

	assertSnapshot(t, query, args, err,
		overviewFields+" FROM bfe_ai_metrics_1m"+fullWhere, fullFilterArgs())
}

func TestBuildOverviewMetricsSQL_TimeOnly(t *testing.T) {
	query, args, err := buildOverviewMetricsSQL("bfe_ai_metrics_1m", &ireport.Filter{Start: testStart, End: testEnd})

	require.NoError(t, err)
	assert.Equal(t,
		overviewFields+" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?)",
		query)
	assert.Equal(t, []interface{}{testStart, testEnd}, args)
}

func TestBuildOverviewCostSQL(t *testing.T) {
	query, args, err := buildOverviewCostSQL("bfe_ai_metrics_1m", fullFilter())

	assertSnapshot(t, query, args, err,
		"SELECT ai_cost_currency AS currency,IFNULL(SUM(ai_cost_value_sum),0) AS value"+
			" FROM bfe_ai_metrics_1m"+
			" WHERE (ai_stream=? AND ai_apikey_id IN (?) AND ai_provider IN (?)"+
			" AND ai_target_model IN (?,?) AND hostid IN (?) AND res_status_code IN (?,?)"+
			" AND IFNULL(ai_cost_currency,'')!=? AND ts_min>=? AND ts_min<?)"+
			" GROUP BY ai_cost_currency",
		append(fullFilterArgs()[:8], append([]interface{}{""}, fullFilterArgs()[8:]...)...))
}

func TestBuildLogsTotalSQL(t *testing.T) {
	query, args, err := buildLogsTotalSQL("bfe_ai_request_log", fullFilter())

	assertSnapshot(t, query, args, err,
		"SELECT COUNT(*) FROM bfe_ai_request_log"+
			" WHERE (ai_stream=? AND ai_apikey_id IN (?) AND ai_provider IN (?)"+
			" AND ai_target_model IN (?,?) AND hostid IN (?) AND res_status_code IN (?,?)"+
			" AND log_time>=? AND log_time<?)",
		fullFilterArgs())
}

func TestBuildTimeSeriesSQL_AllMetrics(t *testing.T) {
	f := &ireport.Filter{Start: testStart, End: testEnd}

	cases := []struct {
		metric    string
		bucket    int
		expectSQL string
	}{
		{
			ireport.MetricQPS, 60,
			"SELECT CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/60)*60 AS SIGNED) AS time,SUM(request_count) AS total" +
				" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?) GROUP BY time ORDER BY time ASC",
		},
		{
			ireport.MetricTokens, 300,
			"SELECT CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/300)*300 AS SIGNED) AS time,SUM(input_tokens) AS input," +
				"SUM(output_tokens) AS output,SUM(total_tokens) AS total" +
				" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?) GROUP BY time ORDER BY time ASC",
		},
		{
			ireport.MetricLatency, 1800,
			"SELECT CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/1800)*1800 AS SIGNED) AS time,SUM(all_time_sum) AS all_time_sum," +
				"SUM(request_count) AS request_count,MAX(all_time_sum/request_count) AS latency_max" +
				" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?) GROUP BY time ORDER BY time ASC",
		},
		{
			ireport.MetricTTFT, 60,
			"SELECT CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/60)*60 AS SIGNED) AS time,SUM(ttft_us_sum) AS ttft_us_sum," +
				"SUM(tpot_us_sum) AS tpot_us_sum,SUM(CASE WHEN ai_stream=1 THEN request_count ELSE 0 END) AS stream_requests" +
				" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?) GROUP BY time ORDER BY time ASC",
		},
		{
			ireport.MetricCost, 60,
			"SELECT CAST(FLOOR(TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)/60)*60 AS SIGNED) AS time,ai_cost_currency AS currency," +
				"SUM(ai_cost_value_sum) AS value" +
				" FROM bfe_ai_metrics_1m WHERE (ts_min>=? AND ts_min<?) GROUP BY time,ai_cost_currency ORDER BY time ASC",
		},
	}
	for _, tc := range cases {
		t.Run(tc.metric, func(t *testing.T) {
			query, args, err := buildTimeSeriesSQL("bfe_ai_metrics_1m", tc.metric, f, tc.bucket)
			assertSnapshot(t, query, args, err, tc.expectSQL, []interface{}{testStart, testEnd})
		})
	}
}

func TestBuildTimeSeriesSQL_InvalidMetric(t *testing.T) {
	_, _, err := buildTimeSeriesSQL("bfe_ai_metrics_1m", "bogus", &ireport.Filter{Start: testStart, End: testEnd}, 60)
	require.Error(t, err)
}

func TestBuildRankingsSQL(t *testing.T) {
	cases := []struct {
		dimension string
		expectSel string
		expectGrp string
	}{
		{ireport.DimensionModel, "SELECT ai_target_model AS name,", " GROUP BY ai_target_model "},
		{ireport.DimensionRequestedModel, "SELECT ai_requested_model AS name,", " GROUP BY ai_requested_model "},
		{ireport.DimensionProvider, "SELECT ai_provider AS name,", " GROUP BY ai_provider "},
		{ireport.DimensionAPIKey, "SELECT ai_apikey_id AS name,", " GROUP BY ai_apikey_id "},
		{ireport.DimensionHost, "SELECT hostid AS name,", " GROUP BY hostid "},
		{ireport.DimensionStatus, "SELECT CAST(res_status_code AS CHAR) AS name,", " GROUP BY res_status_code "},
		{ireport.DimensionProtocol, "SELECT ai_protocol AS name,", " GROUP BY ai_protocol "},
		{ireport.DimensionMode, "SELECT ai_mode AS name,", " GROUP BY ai_mode "},
	}
	for _, tc := range cases {
		t.Run(tc.dimension, func(t *testing.T) {
			query, args, err := buildRankingsSQL("bfe_ai_metrics_1m", tc.dimension, &ireport.Filter{Start: testStart, End: testEnd}, 10)
			require.NoError(t, err)
			assert.Contains(t, query, tc.expectSel)
			assert.Contains(t, query, "IFNULL(")
			assert.Contains(t, query, tc.expectGrp)
			assert.Contains(t, query, " ORDER BY request_count DESC LIMIT ?,?")
			assert.Equal(t, []interface{}{"", testStart, testEnd, 0, 10}, args)
		})
	}
}

func TestBuildRankingsSQL_InvalidDimension(t *testing.T) {
	_, _, err := buildRankingsSQL("bfe_ai_metrics_1m", "bogus", &ireport.Filter{Start: testStart, End: testEnd}, 10)
	require.Error(t, err)
}

func TestBuildDistributionSQL(t *testing.T) {
	cases := []struct {
		dimension string
		expectSel string
	}{
		{ireport.DimensionStatus, "SELECT CASE WHEN res_status_code=0 THEN 'unknown' ELSE CAST(res_status_code AS CHAR) END AS name,"},
		{ireport.DimensionProtocol, "SELECT CASE WHEN IFNULL(ai_protocol,'')='' THEN 'unknown' ELSE ai_protocol END AS name,"},
		{ireport.DimensionMode, "SELECT CASE WHEN IFNULL(ai_mode,'')='' THEN 'unknown' ELSE ai_mode END AS name,"},
		{ireport.DimensionStream, "SELECT CAST(ai_stream AS CHAR) AS name,"},
	}
	for _, tc := range cases {
		t.Run(tc.dimension, func(t *testing.T) {
			query, args, err := buildDistributionSQL("bfe_ai_metrics_1m", tc.dimension, &ireport.Filter{Start: testStart, End: testEnd})
			require.NoError(t, err)
			assert.Contains(t, query, tc.expectSel)
			assert.Contains(t, query, " GROUP BY name ORDER BY request_count DESC")
			assert.Equal(t, []interface{}{testStart, testEnd}, args)
		})
	}
}

func TestBuildDistributionSQL_InvalidDimension(t *testing.T) {
	_, _, err := buildDistributionSQL("bfe_ai_metrics_1m", "model", &ireport.Filter{Start: testStart, End: testEnd})
	require.Error(t, err)
}

func TestBuildLogsCountSQL_WithLogFilters(t *testing.T) {
	f := &ireport.LogFilter{
		Filter:          *fullFilter(),
		RequestedModels: []string{"gpt-4"},
		ErrOnly:         true,
		Keyword:         "timeout",
		Page:            2,
		PageSize:        20,
	}
	query, args, err := buildLogsCountSQL("bfe_ai_request_log", f)

	require.NoError(t, err)
	assert.Equal(t, "SELECT COUNT(*) FROM bfe_ai_request_log"+
		" WHERE (ai_stream=? AND ai_apikey_id IN (?) AND ai_provider IN (?) AND ai_requested_model IN (?)"+
		" AND ai_target_model IN (?,?) AND hostid IN (?) AND res_status_code IN (?,?)"+
		" AND IFNULL(err_code,'')!=? AND log_time>=? AND log_time<? AND err_msg LIKE ?)", query)
	assert.Equal(t, []interface{}{
		int8(1), "key-1", "openai", "gpt-4", "gpt-4o", "gpt-4", "gw-01", 200, 500,
		"", testStart, testEnd, "%timeout%",
	}, args)
}

func TestBuildLogsSQL_Pagination(t *testing.T) {
	cases := []struct {
		name   string
		page   int
		size   int
		offset int
	}{
		{"first page", 1, 20, 0},
		{"third page", 3, 50, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &ireport.LogFilter{
				Filter:   ireport.Filter{Start: testStart, End: testEnd},
				Page:     tc.page,
				PageSize: tc.size,
			}
			query, args, err := buildLogsSQL("bfe_ai_request_log", f)

			require.NoError(t, err)
			assert.Contains(t, query, " ORDER BY log_time DESC LIMIT ?,?")
			assert.Equal(t, []interface{}{testStart, testEnd, tc.offset, tc.size}, args)
		})
	}
}

func TestBuildLogsSQL_Projection(t *testing.T) {
	f := &ireport.LogFilter{
		Filter:   ireport.Filter{Start: testStart, End: testEnd},
		Page:     1,
		PageSize: 20,
	}
	query, _, err := buildLogsSQL("bfe_ai_request_log", f)
	require.NoError(t, err)
	assert.Contains(t, query, "TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', log_time) AS log_time")
	assert.Contains(t, query, "ai_auth_reject_quota_plans")
	assert.Contains(t, query, "req_headers")
	assert.Contains(t, query, "res_headers")
}

func TestTablePrefix(t *testing.T) {
	s := New(nil, "bfe_report")
	assert.Equal(t, "bfe_report.bfe_ai_request_log", s.table(tableDetail))
	assert.Equal(t, "bfe_report.bfe_ai_metrics_1m", s.table(tableMetrics))

	s = New(nil, "")
	assert.Equal(t, "bfe_ai_request_log", s.table(tableDetail))
}

func TestOverviewResultFromRow(t *testing.T) {
	row := &overviewMetricsRow{
		requestTotal:   sql.NullInt64{Int64: 100, Valid: true},
		errorTotal:     sql.NullInt64{Int64: 5, Valid: true},
		inputTokens:    sql.NullInt64{Int64: 1000, Valid: true},
		outputTokens:   sql.NullInt64{Int64: 200, Valid: true},
		totalTokens:    sql.NullInt64{Int64: 1200, Valid: true},
		allTimeSum:     sql.NullInt64{Int64: 120000, Valid: true},
		latencyMax:     sql.NullFloat64{Float64: 3000, Valid: true},
		ttftUsSum:      sql.NullInt64{Int64: 40_000_000, Valid: true},
		streamRequests: sql.NullInt64{Int64: 50, Valid: true},
		tpotUsSum:      sql.NullInt64{Int64: 2_000_000, Valid: true},
		rateLimitHits:  sql.NullInt64{Int64: 7, Valid: true},
		authRejects:    sql.NullInt64{Int64: 3, Valid: true},
	}
	cost := []*ireport.CostItem{{Currency: "USD", Value: 5000}}

	result := overviewResultFromRow(row, cost, 123)

	assert.Equal(t, int64(100), result.RequestTotal)
	assert.InDelta(t, 0.05, result.ErrorRate, 1e-9)
	assert.InDelta(t, 1200, result.LatencyAvgMs, 1e-9)
	assert.InDelta(t, 3000, result.LatencyMaxMs, 1e-9)
	assert.InDelta(t, 800, result.TtftAvgMs, 1e-9) // 40s/50 / 1000
	assert.InDelta(t, 40, result.TpotAvgMs, 1e-9)  // 2s/50 / 1000
	assert.Equal(t, int64(7), result.RateLimitHits)
	assert.Equal(t, int64(3), result.AuthRejects)
	assert.Equal(t, int64(123), result.LogsTotal)
	require.Len(t, result.Cost, 1)
	assert.Nil(t, result.LatencyP50Ms)

	// zero requests: rates stay zero without division by zero.
	empty := overviewResultFromRow(&overviewMetricsRow{}, nil, 0)
	assert.Equal(t, float64(0), empty.ErrorRate)
	assert.Equal(t, float64(0), empty.LatencyAvgMs)
	assert.Equal(t, float64(0), empty.TtftAvgMs)
}

func TestRowToMetricPoint(t *testing.T) {
	qps := rowToMetricPoint(ireport.MetricQPS, 1000, 60, metricRowValues{total: 300})
	require.NotNil(t, qps.Value)
	assert.InDelta(t, 5, *qps.Value, 1e-9)
	assert.Nil(t, qps.P50)

	tokens := rowToMetricPoint(ireport.MetricTokens, 1000, 300, metricRowValues{input: 600, output: 300, total: 900})
	assert.InDelta(t, 2, *tokens.Input, 1e-9)
	assert.InDelta(t, 1, *tokens.Output, 1e-9)
	assert.InDelta(t, 3, *tokens.Total, 1e-9)

	latency := rowToMetricPoint(ireport.MetricLatency, 1000, 60, metricRowValues{allTimeSum: 900, requestCount: 3, latencyMax: 450})
	assert.InDelta(t, 300, *latency.Avg, 1e-9)
	assert.InDelta(t, 450, *latency.Max, 1e-9)

	ttft := rowToMetricPoint(ireport.MetricTTFT, 1000, 60, metricRowValues{ttftUsSum: 2_000_000, streamRequests: 4})
	assert.InDelta(t, 500, *ttft.Value, 1e-9)

	noStream := rowToMetricPoint(ireport.MetricTPOT, 1000, 60, metricRowValues{})
	require.NotNil(t, noStream.Value)
	assert.Equal(t, float64(0), *noStream.Value)

	cost := rowToMetricPoint(ireport.MetricCost, 1000, 60, metricRowValues{value: 600, currency: "USD"})
	assert.InDelta(t, 10, *cost.Value, 1e-9)
	assert.Equal(t, "USD", cost.Currency)
}

func TestDistributionRatios(t *testing.T) {
	items := []*ireport.DistItem{
		{Name: "200", RequestCount: 80},
		{Name: "500", RequestCount: 20},
	}
	distributionRatios(items)
	assert.InDelta(t, 0.8, items[0].Ratio, 1e-9)
	assert.InDelta(t, 0.2, items[1].Ratio, 1e-9)

	empty := []*ireport.DistItem{{Name: "unknown", RequestCount: 0}}
	distributionRatios(empty)
	assert.Equal(t, float64(0), empty[0].Ratio)
}

func TestNullPtrHelpers(t *testing.T) {
	assert.Nil(t, nullInt64Ptr(sql.NullInt64{}))
	assert.Nil(t, nullInt16Ptr(sql.NullInt64{}))
	assert.Nil(t, nullStringPtr(sql.NullString{}))

	v := sql.NullInt64{Int64: 42, Valid: true}
	require.NotNil(t, nullInt64Ptr(v))
	assert.Equal(t, int64(42), *nullInt64Ptr(v))
	require.NotNil(t, nullInt16Ptr(v))
	assert.Equal(t, int16(42), *nullInt16Ptr(v))

	s := sql.NullString{String: "x", Valid: true}
	require.NotNil(t, nullStringPtr(s))
	assert.Equal(t, "x", *nullStringPtr(s))
}

// TestUnixSecondRendering_TimezoneNeutral 是时区口径问题的说明性测试：
// 所有「DATETIME → Unix 秒」的渲染必须走 TIMESTAMPDIFF（日历算术，不依赖
// MySQL 会话时区），不得使用按会话时区解读墙钟的 UNIX_TIMESTAMP。
// log-reader 以 UTC 墙钟写入 log_time，存储值即 UTC 墙钟。
func TestUnixSecondRendering_TimezoneNeutral(t *testing.T) {
	for _, bucket := range []int{60, 300, 1800} {
		expr := bucketExpr(bucket)
		assert.Contains(t, expr, "TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', ts_min)")
		assert.NotContains(t, expr, "UNIX_TIMESTAMP")
	}
	for _, field := range logRowFields {
		assert.NotContains(t, field, "UNIX_TIMESTAMP", "field: %s", field)
	}
	assert.Contains(t, logRowFields[1], "TIMESTAMPDIFF(SECOND, '1970-01-01 00:00:00', log_time)")

	// 口径自洽性：TIMESTAMPDIFF 对 UTC 墙钟做日历算术，结果即真实 Unix 秒
	//（该值与任何会话时区无关；UNIX_TIMESTAMP 在 CST 会话下会把同一墙钟
	// 解读为 1789437600 = 2026-09-15 02:00:00 UTC，正是本修复消除的 8h 偏移）。
	want := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC).Unix()
	assert.Equal(t, int64(1789466400), want)
}
