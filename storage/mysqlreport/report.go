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
	"context"
	"database/sql"
	"strconv"

	"github.com/didi/gendry/builder"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ireport"
)

const (
	tableDetail   = "bfe_ai_request_log"
	tableMetrics  = "bfe_ai_metrics_1m"
	secondsPerDay = 86400
)

// ReportStorager implements ireport.ReportStorager against the MySQL
// report database (lightweight deployment, see design-docs
// modifications/2026-09-15-report-query-api). Time-series / rankings /
// distribution only read the aggregate table bfe_ai_metrics_1m; only the
// log list (and the overview logs_total) read the detail table.
type ReportStorager struct {
	db       *sql.DB
	database string // optional schema override used as table prefix
}

// New creates a new MySQL report storager. database may be empty (tables
// are then resolved within the connection's default schema) or a schema
// name used to prefix table names.
func New(db *sql.DB, database string) *ReportStorager {
	return &ReportStorager{db: db, database: database}
}

var _ ireport.ReportStorager = (*ReportStorager)(nil)

func (s *ReportStorager) table(name string) string {
	if s.database == "" {
		return name
	}
	return s.database + "." + name
}

// toInterfaces converts a string slice into the []interface{} shape gendry
// requires for IN conditions.
func toInterfaces(ss []string) []interface{} {
	vals := make([]interface{}, 0, len(ss))
	for _, one := range ss {
		vals = append(vals, one)
	}
	return vals
}

// toIntInterfaces converts an int slice into the []interface{} shape
// gendry requires for IN conditions.
func toIntInterfaces(nums []int) []interface{} {
	vals := make([]interface{}, 0, len(nums))
	for _, one := range nums {
		vals = append(vals, one)
	}
	return vals
}

// metricsWhere builds the shared WHERE map for the aggregate table. All
// values are bound parameters.
func metricsWhere(f *ireport.Filter) map[string]interface{} {
	where := map[string]interface{}{
		"ts_min >=": f.Start,
		"ts_min <":  f.End,
	}
	if len(f.Models) > 0 {
		where["ai_target_model in"] = toInterfaces(f.Models)
	}
	if len(f.ApikeyIDs) > 0 {
		where["ai_apikey_id in"] = toInterfaces(f.ApikeyIDs)
	}
	if len(f.Providers) > 0 {
		where["ai_provider in"] = toInterfaces(f.Providers)
	}
	if len(f.Hosts) > 0 {
		where["hostid in"] = toInterfaces(f.Hosts)
	}
	if f.Stream != nil {
		stream := int8(0)
		if *f.Stream {
			stream = 1
		}
		where["ai_stream ="] = stream
	}
	if len(f.StatusCodes) > 0 {
		where["res_status_code in"] = toIntInterfaces(f.StatusCodes)
	}
	return where
}

// detailWhere builds the WHERE map for the detail table from the shared
// filter fields (the detail table uses the same column names).
func detailWhere(f *ireport.Filter) map[string]interface{} {
	where := map[string]interface{}{
		"log_time >=": f.Start,
		"log_time <":  f.End,
	}
	if len(f.Models) > 0 {
		where["ai_target_model in"] = toInterfaces(f.Models)
	}
	if len(f.ApikeyIDs) > 0 {
		where["ai_apikey_id in"] = toInterfaces(f.ApikeyIDs)
	}
	if len(f.Providers) > 0 {
		where["ai_provider in"] = toInterfaces(f.Providers)
	}
	if len(f.Hosts) > 0 {
		where["hostid in"] = toInterfaces(f.Hosts)
	}
	if f.Stream != nil {
		stream := int8(0)
		if *f.Stream {
			stream = 1
		}
		where["ai_stream ="] = stream
	}
	if len(f.StatusCodes) > 0 {
		where["res_status_code in"] = toIntInterfaces(f.StatusCodes)
	}
	return where
}

// logWhere builds the WHERE map for the log list, extending detailWhere
// with the logs-only criteria (requested model, err_only, keyword).
func logWhere(f *ireport.LogFilter) map[string]interface{} {
	where := detailWhere(&f.Filter)
	if len(f.RequestedModels) > 0 {
		where["ai_requested_model in"] = toInterfaces(f.RequestedModels)
	}
	if f.ErrOnly {
		where["IFNULL(err_code,'') !="] = ""
	}
	if f.Keyword != "" {
		where["err_msg like"] = "%" + f.Keyword + "%"
	}
	return where
}

// epochLiteral is the wall-clock UTC epoch used to turn a stored DATETIME
// into Unix seconds without any session-timezone interpretation:
// TIMESTAMPDIFF is pure calendar arithmetic, unlike UNIX_TIMESTAMP which
// reads the DATETIME in the session zone (log-reader writes UTC wall
// clock, so the stored value IS the UTC wall clock).
const epochLiteral = "'1970-01-01 00:00:00'"

// bucketExpr is the MySQL time-bucket expression. The bucket width is
// server-controlled (one of 60/300/1800 computed by the manager), so it is
// inlined as an integer literal: gendry cannot bind parameters inside
// SELECT fields.
func bucketExpr(bucketSec int) string {
	return "CAST(FLOOR(TIMESTAMPDIFF(SECOND, " + epochLiteral + ", ts_min)/" +
		strconv.Itoa(bucketSec) + ")*" + strconv.Itoa(bucketSec) + " AS SIGNED) AS time"
}

// overviewMetricFields are the aggregate SUM columns of the overview card.
var overviewMetricFields = []string{
	"IFNULL(SUM(request_count),0) AS request_total",
	"IFNULL(SUM(error_count),0) AS error_total",
	"IFNULL(SUM(input_tokens),0) AS input_tokens",
	"IFNULL(SUM(output_tokens),0) AS output_tokens",
	"IFNULL(SUM(total_tokens),0) AS total_tokens",
	"IFNULL(SUM(all_time_sum),0) AS all_time_sum",
	"IFNULL(MAX(all_time_sum/request_count),0) AS latency_max",
	"IFNULL(SUM(ttft_us_sum),0) AS ttft_us_sum",
	"IFNULL(SUM(CASE WHEN ai_stream=1 THEN request_count ELSE 0 END),0) AS stream_requests",
	"IFNULL(SUM(tpot_us_sum),0) AS tpot_us_sum",
	"IFNULL(SUM(rate_limit_hits),0) AS rate_limit_hits",
	"IFNULL(SUM(auth_reject_count),0) AS auth_rejects",
}

func buildOverviewMetricsSQL(metricsTable string, f *ireport.Filter) (string, []interface{}, error) {
	return builder.BuildSelect(metricsTable, metricsWhere(f), overviewMetricFields)
}

func buildOverviewCostSQL(metricsTable string, f *ireport.Filter) (string, []interface{}, error) {
	where := metricsWhere(f)
	where["IFNULL(ai_cost_currency,'') !="] = ""
	where["_groupby"] = "ai_cost_currency"
	return builder.BuildSelect(metricsTable, where, []string{
		"ai_cost_currency AS currency",
		"IFNULL(SUM(ai_cost_value_sum),0) AS value",
	})
}

func buildLogsTotalSQL(detailTable string, f *ireport.Filter) (string, []interface{}, error) {
	return builder.BuildSelect(detailTable, detailWhere(f), []string{"COUNT(*)"})
}

// buildTimeSeriesSQL builds the per-metric time-series query over the
// aggregate table. Value fields are raw per-bucket sums; the per-second
// division happens in rowToMetricPoint so the SQL stays integer-only.
func buildTimeSeriesSQL(metricsTable, metric string, f *ireport.Filter, bucketSec int) (string, []interface{}, error) {
	where := metricsWhere(f)
	where["_groupby"] = "time"
	where["_orderby"] = "time ASC"

	switch metric {
	case ireport.MetricQPS:
		return builder.BuildSelect(metricsTable, where, []string{
			bucketExpr(bucketSec),
			"SUM(request_count) AS total",
		})
	case ireport.MetricTokens:
		return builder.BuildSelect(metricsTable, where, []string{
			bucketExpr(bucketSec),
			"SUM(input_tokens) AS input",
			"SUM(output_tokens) AS output",
			"SUM(total_tokens) AS total",
		})
	case ireport.MetricLatency:
		return builder.BuildSelect(metricsTable, where, []string{
			bucketExpr(bucketSec),
			"SUM(all_time_sum) AS all_time_sum",
			"SUM(request_count) AS request_count",
			"MAX(all_time_sum/request_count) AS latency_max",
		})
	case ireport.MetricTTFT, ireport.MetricTPOT:
		return builder.BuildSelect(metricsTable, where, []string{
			bucketExpr(bucketSec),
			"SUM(ttft_us_sum) AS ttft_us_sum",
			"SUM(tpot_us_sum) AS tpot_us_sum",
			"SUM(CASE WHEN ai_stream=1 THEN request_count ELSE 0 END) AS stream_requests",
		})
	case ireport.MetricCost:
		where["_groupby"] = "time,ai_cost_currency"
		return builder.BuildSelect(metricsTable, where, []string{
			bucketExpr(bucketSec),
			"ai_cost_currency AS currency",
			"SUM(ai_cost_value_sum) AS value",
		})
	default:
		return "", nil, xerror.WrapParamErrorWithMsg("invalid metric: %s", metric)
	}
}

func buildRankingsSQL(metricsTable, dimension string, f *ireport.Filter, limit int) (string, []interface{}, error) {
	column, ok := ireport.DimensionColumns[dimension]
	if !ok {
		return "", nil, xerror.WrapParamErrorWithMsg("invalid dimension: %s", dimension)
	}

	field := column + " AS name"
	if dimension == ireport.DimensionStatus || dimension == ireport.DimensionStream {
		field = "CAST(" + column + " AS CHAR) AS name"
	}

	where := metricsWhere(f)
	where["IFNULL("+column+",'') !="] = ""
	where["_groupby"] = column
	where["_orderby"] = "request_count DESC"
	where["_limit"] = []uint{0, uint(limit)}

	return builder.BuildSelect(metricsTable, where, []string{
		field,
		"SUM(request_count) AS request_count",
		"SUM(error_count) AS error_count",
		"SUM(input_tokens) AS input_tokens",
		"SUM(output_tokens) AS output_tokens",
	})
}

// distributionNameExpr renders the dimension value as its display name,
// normalizing NULL/empty to 'unknown'. Numeric dimensions are cast to
// strings so the response shape is uniform.
func distributionNameExpr(dimension string) (string, error) {
	switch dimension {
	case ireport.DimensionStatus:
		return "CASE WHEN res_status_code=0 THEN 'unknown' ELSE CAST(res_status_code AS CHAR) END AS name", nil
	case ireport.DimensionStream:
		return "CAST(ai_stream AS CHAR) AS name", nil
	case ireport.DimensionProtocol:
		return "CASE WHEN IFNULL(ai_protocol,'')='' THEN 'unknown' ELSE ai_protocol END AS name", nil
	case ireport.DimensionMode:
		return "CASE WHEN IFNULL(ai_mode,'')='' THEN 'unknown' ELSE ai_mode END AS name", nil
	default:
		return "", xerror.WrapParamErrorWithMsg("invalid dimension: %s", dimension)
	}
}

func buildDistributionSQL(metricsTable, dimension string, f *ireport.Filter) (string, []interface{}, error) {
	nameExpr, err := distributionNameExpr(dimension)
	if err != nil {
		return "", nil, err
	}

	where := metricsWhere(f)
	where["_groupby"] = "name"
	where["_orderby"] = "request_count DESC"

	return builder.BuildSelect(metricsTable, where, []string{
		nameExpr,
		"SUM(request_count) AS request_count",
	})
}

func buildLogsCountSQL(detailTable string, f *ireport.LogFilter) (string, []interface{}, error) {
	return builder.BuildSelect(detailTable, logWhere(f), []string{"COUNT(*)"})
}

// logRowFields is the display projection of the detail row; JSON columns
// are returned verbatim. log_time is converted to unix seconds with
// TIMESTAMPDIFF (calendar arithmetic, session-timezone neutral) so the
// scan does not depend on the driver's ParseTime setting or the MySQL
// session zone.
var logRowFields = []string{
	"logid",
	"TIMESTAMPDIFF(SECOND, " + epochLiteral + ", log_time) AS log_time",
	"hostid",
	"product",
	"ai_apikey_id",
	"ai_requested_model",
	"ai_target_model",
	"ai_provider",
	"ai_protocol",
	"ai_mode",
	"ai_stream",
	"res_status_code",
	"err_code",
	"err_msg",
	"ai_input_tokens",
	"ai_output_tokens",
	"ai_total_tokens",
	"all_time",
	"ai_ttft_us",
	"ai_tpot_us",
	"ai_cost_value",
	"ai_cost_currency",
	"ai_rate_limit_hits",
	"ai_auth_reject_quota_plans",
	"level1Name",
	"level1",
	"level2Name",
	"level2",
	"level3Name",
	"level3",
	"level4Name",
	"level4",
	"level5Name",
	"level5",
	"client_ip",
	"header_host",
	"origin_uri",
	"req_headers",
	"res_headers",
}

func buildLogsSQL(detailTable string, f *ireport.LogFilter) (string, []interface{}, error) {
	offset := uint(0)
	if f.Page > 1 {
		offset = uint(f.Page-1) * uint(f.PageSize)
	}

	where := logWhere(f)
	where["_orderby"] = "log_time DESC"
	where["_limit"] = []uint{offset, uint(f.PageSize)}

	return builder.BuildSelect(detailTable, where, logRowFields)
}

// Overview implements ireport.ReportStorager.
func (s *ReportStorager) Overview(ctx context.Context, f *ireport.Filter) (*ireport.OverviewResult, error) {
	metricsTable := s.table(tableMetrics)
	detailTable := s.table(tableDetail)

	metricsSQL, metricsArgs, err := buildOverviewMetricsSQL(metricsTable, f)
	if err != nil {
		return nil, err
	}

	row := &overviewMetricsRow{}
	err = s.db.QueryRowContext(ctx, metricsSQL, metricsArgs...).Scan(
		&row.requestTotal,
		&row.errorTotal,
		&row.inputTokens,
		&row.outputTokens,
		&row.totalTokens,
		&row.allTimeSum,
		&row.latencyMax,
		&row.ttftUsSum,
		&row.streamRequests,
		&row.tpotUsSum,
		&row.rateLimitHits,
		&row.authRejects,
	)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	costSQL, costArgs, err := buildOverviewCostSQL(metricsTable, f)
	if err != nil {
		return nil, err
	}
	cost, err := s.queryCost(ctx, costSQL, costArgs)
	if err != nil {
		return nil, err
	}

	logsSQL, logsArgs, err := buildLogsTotalSQL(detailTable, f)
	if err != nil {
		return nil, err
	}
	var logsTotal int64
	if err := s.db.QueryRowContext(ctx, logsSQL, logsArgs...).Scan(&logsTotal); err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	return overviewResultFromRow(row, cost, logsTotal), nil
}

func (s *ReportStorager) queryCost(ctx context.Context, query string, args []interface{}) ([]*ireport.CostItem, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	defer rows.Close()

	items := make([]*ireport.CostItem, 0, 4)
	for rows.Next() {
		item := &ireport.CostItem{}
		if err := rows.Scan(&item.Currency, &item.Value); err != nil {
			return nil, xerror.WrapDaoError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	return items, nil
}

// overviewMetricsRow is the scan target of the overview aggregate query.
type overviewMetricsRow struct {
	requestTotal   sql.NullInt64
	errorTotal     sql.NullInt64
	inputTokens    sql.NullInt64
	outputTokens   sql.NullInt64
	totalTokens    sql.NullInt64
	allTimeSum     sql.NullInt64
	latencyMax     sql.NullFloat64
	ttftUsSum      sql.NullInt64
	streamRequests sql.NullInt64
	tpotUsSum      sql.NullInt64
	rateLimitHits  sql.NullInt64
	authRejects    sql.NullInt64
}

// overviewResultFromRow assembles the overview card with the documented
// calibers: error_rate = error/request; latency_avg = all_time_sum/request;
// latency_max approximates the peak per-minute average latency
// (MAX(all_time_sum/request_count)) because the aggregate table keeps no
// per-request max column (identical on both backends); ttft/tpot average
// over stream requests with microsecond to millisecond conversion.
func overviewResultFromRow(row *overviewMetricsRow, cost []*ireport.CostItem, logsTotal int64) *ireport.OverviewResult {
	requestTotal := row.requestTotal.Int64
	streamRequests := row.streamRequests.Int64

	result := &ireport.OverviewResult{
		RequestTotal:  requestTotal,
		ErrorTotal:    row.errorTotal.Int64,
		InputTokens:   row.inputTokens.Int64,
		OutputTokens:  row.outputTokens.Int64,
		TotalTokens:   row.totalTokens.Int64,
		LatencyMaxMs:  row.latencyMax.Float64,
		Cost:          cost,
		RateLimitHits: row.rateLimitHits.Int64,
		AuthRejects:   row.authRejects.Int64,
		LogsTotal:     logsTotal,
	}
	if requestTotal > 0 {
		result.ErrorRate = float64(row.errorTotal.Int64) / float64(requestTotal)
		result.LatencyAvgMs = float64(row.allTimeSum.Int64) / float64(requestTotal)
	}
	if streamRequests > 0 {
		result.TtftAvgMs = float64(row.ttftUsSum.Int64) / float64(streamRequests) / 1000
		result.TpotAvgMs = float64(row.tpotUsSum.Int64) / float64(streamRequests) / 1000
	}
	return result
}

// TimeSeries implements ireport.ReportStorager.
func (s *ReportStorager) TimeSeries(ctx context.Context, metric string, f *ireport.Filter, bucketSec int) ([]*ireport.MetricPoint, error) {
	query, args, err := buildTimeSeriesSQL(s.table(tableMetrics), metric, f, bucketSec)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	defer rows.Close()

	points := make([]*ireport.MetricPoint, 0, 128)
	for rows.Next() {
		point, err := scanMetricPoint(rows, metric, bucketSec)
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	return points, nil
}

// metricRowScanner abstracts *sql.Rows for scanMetricPoint.
type metricRowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMetricPoint(scanner metricRowScanner, metric string, bucketSec int) (*ireport.MetricPoint, error) {
	var (
		bucket         sql.NullInt64
		total          sql.NullInt64
		input          sql.NullInt64
		output         sql.NullInt64
		allTimeSum     sql.NullInt64
		requestCount   sql.NullInt64
		latencyMax     sql.NullFloat64
		ttftUsSum      sql.NullInt64
		tpotUsSum      sql.NullInt64
		streamRequests sql.NullInt64
		value          sql.NullInt64
		currency       sql.NullString
	)

	var err error
	switch metric {
	case ireport.MetricQPS:
		err = scanner.Scan(&bucket, &total)
	case ireport.MetricTokens:
		err = scanner.Scan(&bucket, &input, &output, &total)
	case ireport.MetricLatency:
		err = scanner.Scan(&bucket, &allTimeSum, &requestCount, &latencyMax)
	case ireport.MetricTTFT, ireport.MetricTPOT:
		err = scanner.Scan(&bucket, &ttftUsSum, &tpotUsSum, &streamRequests)
	case ireport.MetricCost:
		err = scanner.Scan(&bucket, &currency, &value)
	default:
		return nil, xerror.WrapParamErrorWithMsg("invalid metric: %s", metric)
	}
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	return rowToMetricPoint(metric, bucket.Int64, bucketSec, metricRowValues{
		total:          total.Int64,
		input:          input.Int64,
		output:         output.Int64,
		allTimeSum:     allTimeSum.Int64,
		requestCount:   requestCount.Int64,
		latencyMax:     latencyMax.Float64,
		ttftUsSum:      ttftUsSum.Int64,
		tpotUsSum:      tpotUsSum.Int64,
		streamRequests: streamRequests.Int64,
		value:          value.Int64,
		currency:       currency.String,
	}), nil
}

// metricRowValues carries the scanned raw sums of one bucket.
type metricRowValues struct {
	total          int64
	input          int64
	output         int64
	allTimeSum     int64
	requestCount   int64
	latencyMax     float64
	ttftUsSum      int64
	tpotUsSum      int64
	streamRequests int64
	value          int64
	currency       string
}

// rowToMetricPoint converts raw per-bucket sums into a point, applying the
// per-second rates (qps / tokens / cost, divided by the bucket width) and
// the per-request calibers (latency / ttft / tpot). It is a pure function.
func rowToMetricPoint(metric string, bucket int64, bucketSec int, v metricRowValues) *ireport.MetricPoint {
	point := &ireport.MetricPoint{Time: bucket}
	float64Ptr := func(f float64) *float64 { return &f }

	switch metric {
	case ireport.MetricQPS:
		point.Value = float64Ptr(float64(v.total) / float64(bucketSec))
	case ireport.MetricTokens:
		point.Input = float64Ptr(float64(v.input) / float64(bucketSec))
		point.Output = float64Ptr(float64(v.output) / float64(bucketSec))
		point.Total = float64Ptr(float64(v.total) / float64(bucketSec))
	case ireport.MetricLatency:
		if v.requestCount > 0 {
			point.Avg = float64Ptr(float64(v.allTimeSum) / float64(v.requestCount))
		} else {
			point.Avg = float64Ptr(0)
		}
		point.Max = float64Ptr(v.latencyMax)
	case ireport.MetricTTFT:
		point.Value = float64Ptr(avgLatencyMs(v.ttftUsSum, v.streamRequests))
	case ireport.MetricTPOT:
		point.Value = float64Ptr(avgLatencyMs(v.tpotUsSum, v.streamRequests))
	case ireport.MetricCost:
		point.Value = float64Ptr(float64(v.value) / float64(bucketSec))
		point.Currency = v.currency
	}
	return point
}

// avgLatencyMs averages a microsecond sum over stream requests and
// converts the result to milliseconds; zero when no stream request exists.
func avgLatencyMs(usSum, streamRequests int64) float64 {
	if streamRequests <= 0 {
		return 0
	}
	return float64(usSum) / float64(streamRequests) / 1000
}

// Rankings implements ireport.ReportStorager.
func (s *ReportStorager) Rankings(ctx context.Context, dimension string, f *ireport.Filter, limit int) ([]*ireport.RankingItem, error) {
	query, args, err := buildRankingsSQL(s.table(tableMetrics), dimension, f, limit)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	defer rows.Close()

	items := make([]*ireport.RankingItem, 0, limit)
	for rows.Next() {
		item := &ireport.RankingItem{}
		if err := rows.Scan(&item.Name, &item.RequestCount, &item.ErrorCount, &item.InputTokens, &item.OutputTokens); err != nil {
			return nil, xerror.WrapDaoError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	return items, nil
}

// Distribution implements ireport.ReportStorager.
func (s *ReportStorager) Distribution(ctx context.Context, dimension string, f *ireport.Filter) ([]*ireport.DistItem, error) {
	query, args, err := buildDistributionSQL(s.table(tableMetrics), dimension, f)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	defer rows.Close()

	items := make([]*ireport.DistItem, 0, 16)
	for rows.Next() {
		item := &ireport.DistItem{}
		if err := rows.Scan(&item.Name, &item.RequestCount); err != nil {
			return nil, xerror.WrapDaoError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	distributionRatios(items)
	return items, nil
}

// distributionRatios fills the ratio of each bucket in place: bucket
// requests / window total requests (pure function).
func distributionRatios(items []*ireport.DistItem) {
	var total int64
	for _, item := range items {
		total += item.RequestCount
	}
	if total == 0 {
		return
	}
	for _, item := range items {
		item.Ratio = float64(item.RequestCount) / float64(total)
	}
}

// Logs implements ireport.ReportStorager.
func (s *ReportStorager) Logs(ctx context.Context, f *ireport.LogFilter) (*ireport.LogQueryResult, error) {
	detailTable := s.table(tableDetail)

	countSQL, countArgs, err := buildLogsCountSQL(detailTable, f)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	listSQL, listArgs, err := buildLogsSQL(detailTable, f)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}
	defer rows.Close()

	items := make([]*ireport.LogRow, 0, f.PageSize)
	for rows.Next() {
		item, err := scanLogRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	return &ireport.LogQueryResult{
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
		Items:    items,
	}, nil
}

func nullInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func nullInt16Ptr(n sql.NullInt64) *int16 {
	if !n.Valid {
		return nil
	}
	v := int16(n.Int64)
	return &v
}

func nullStringPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	v := n.String
	return &v
}

// scanLogRow scans one detail projection row. All nullable columns use the
// sql.Null types; JSON columns are returned verbatim as strings.
func scanLogRow(scanner metricRowScanner) (*ireport.LogRow, error) {
	var (
		logid               sql.NullInt64
		logTime             sql.NullInt64
		hostid              sql.NullString
		product             sql.NullString
		apiKeyID            sql.NullString
		requestedModel      sql.NullString
		targetModel         sql.NullString
		provider            sql.NullString
		protocol            sql.NullString
		mode                sql.NullString
		stream              sql.NullInt64
		statusCode          sql.NullInt64
		errCode             sql.NullString
		errMsg              sql.NullString
		inputTokens         sql.NullInt64
		outputTokens        sql.NullInt64
		totalTokens         sql.NullInt64
		allTime             sql.NullInt64
		ttftUs              sql.NullInt64
		tpotUs              sql.NullInt64
		costValue           sql.NullInt64
		costCurrency        sql.NullString
		rateLimitHits       sql.NullString
		authRejectQuotaPlan sql.NullString
		level1Name          sql.NullString
		level1              sql.NullString
		level2Name          sql.NullString
		level2              sql.NullString
		level3Name          sql.NullString
		level3              sql.NullString
		level4Name          sql.NullString
		level4              sql.NullString
		level5Name          sql.NullString
		level5              sql.NullString
		clientIP            sql.NullString
		headerHost          sql.NullString
		originURI           sql.NullString
		reqHeaders          sql.NullString
		resHeaders          sql.NullString
	)

	err := scanner.Scan(
		&logid,
		&logTime,
		&hostid,
		&product,
		&apiKeyID,
		&requestedModel,
		&targetModel,
		&provider,
		&protocol,
		&mode,
		&stream,
		&statusCode,
		&errCode,
		&errMsg,
		&inputTokens,
		&outputTokens,
		&totalTokens,
		&allTime,
		&ttftUs,
		&tpotUs,
		&costValue,
		&costCurrency,
		&rateLimitHits,
		&authRejectQuotaPlan,
		&level1Name,
		&level1,
		&level2Name,
		&level2,
		&level3Name,
		&level3,
		&level4Name,
		&level4,
		&level5Name,
		&level5,
		&clientIP,
		&headerHost,
		&originURI,
		&reqHeaders,
		&resHeaders,
	)
	if err != nil {
		return nil, xerror.WrapDaoError(err)
	}

	row := &ireport.LogRow{
		LogID:               nullInt64Ptr(logid),
		LogTime:             logTime.Int64,
		Hostid:              nullStringPtr(hostid),
		Product:             nullStringPtr(product),
		APIKeyID:            nullStringPtr(apiKeyID),
		RequestedModel:      nullStringPtr(requestedModel),
		TargetModel:         nullStringPtr(targetModel),
		Provider:            nullStringPtr(provider),
		Protocol:            nullStringPtr(protocol),
		Mode:                nullStringPtr(mode),
		Stream:              nullInt16Ptr(stream),
		StatusCode:          nullInt16Ptr(statusCode),
		ErrCode:             nullStringPtr(errCode),
		ErrMsg:              nullStringPtr(errMsg),
		InputTokens:         nullInt64Ptr(inputTokens),
		OutputTokens:        nullInt64Ptr(outputTokens),
		TotalTokens:         nullInt64Ptr(totalTokens),
		AllTime:             nullInt64Ptr(allTime),
		TTFTUs:              nullInt64Ptr(ttftUs),
		TPOTUs:              nullInt64Ptr(tpotUs),
		CostValue:           nullInt64Ptr(costValue),
		CostCurrency:        nullStringPtr(costCurrency),
		RateLimitHits:       nullStringPtr(rateLimitHits),
		AuthRejectQuotaPlan: nullStringPtr(authRejectQuotaPlan),
		Level1Name:          nullStringPtr(level1Name),
		Level1:              nullStringPtr(level1),
		Level2Name:          nullStringPtr(level2Name),
		Level2:              nullStringPtr(level2),
		Level3Name:          nullStringPtr(level3Name),
		Level3:              nullStringPtr(level3),
		Level4Name:          nullStringPtr(level4Name),
		Level4:              nullStringPtr(level4),
		Level5Name:          nullStringPtr(level5Name),
		Level5:              nullStringPtr(level5),
		ClientIP:            nullStringPtr(clientIP),
		HeaderHost:          nullStringPtr(headerHost),
		OriginURI:           nullStringPtr(originURI),
		ReqHeaders:          nullStringPtr(reqHeaders),
		ResHeaders:          nullStringPtr(resHeaders),
	}
	return row, nil
}
