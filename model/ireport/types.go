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

package ireport

import (
	"context"
	"time"
)

// Metric names accepted by the /report/timeseries endpoint. The metric
// calibers follow the existing Grafana dashboard panels (see
// design-docs/modifications/2026-09-15-report-query-api/api-changes.md §2.2).
const (
	MetricQPS     = "qps"     // requests per second
	MetricTokens  = "tokens"  // token throughput (tokens per second)
	MetricLatency = "latency" // request latency in milliseconds
	MetricTTFT    = "ttft"    // time to first token in milliseconds (stream requests)
	MetricTPOT    = "tpot"    // time per output token in milliseconds (stream requests)
	MetricCost    = "cost"    // cost growth (amount per second in the currency unit), per currency
)

// CostFixedPointScale is the scale of the fixed-point cost values stored
// in the report tables and produced by BFE: one unit equals 1e-8 of the
// currency amount (1e-8 yuan for RMB, 1e-8 dollar for USD).
const CostFixedPointScale = 1e8

// CostFixedPointToAmount converts a raw fixed-point cost integer into the
// currency amount returned by the report API.
func CostFixedPointToAmount(value int64) float64 { return float64(value) / CostFixedPointScale }

// TimeSeriesMetrics is the validation set of metric names.
var TimeSeriesMetrics = map[string]bool{
	MetricQPS:     true,
	MetricTokens:  true,
	MetricLatency: true,
	MetricTTFT:    true,
	MetricTPOT:    true,
	MetricCost:    true,
}

// Dimension names accepted by the /report/rankings endpoint.
const (
	DimensionModel          = "model"           // ai_target_model
	DimensionRequestedModel = "requested_model" // ai_requested_model
	DimensionProvider       = "provider"        // ai_provider
	DimensionAPIKey         = "apikey"          // ai_apikey_id
	DimensionHost           = "host"            // hostid
	DimensionStatus         = "status"          // res_status_code
	DimensionProtocol       = "protocol"        // ai_protocol
	DimensionMode           = "mode"            // ai_mode
	DimensionStream         = "stream"          // ai_stream
)

// RankingDimensions is the validation set of dimensions for rankings.
var RankingDimensions = map[string]bool{
	DimensionModel:          true,
	DimensionRequestedModel: true,
	DimensionProvider:       true,
	DimensionAPIKey:         true,
	DimensionHost:           true,
	DimensionStatus:         true,
	DimensionProtocol:       true,
	DimensionMode:           true,
}

// DistributionDimensions is the validation set of dimensions for distribution.
var DistributionDimensions = map[string]bool{
	DimensionStatus:   true,
	DimensionProtocol: true,
	DimensionMode:     true,
	DimensionStream:   true,
}

// DimensionColumns maps dimension names to the aggregate table columns.
// It is shared by both MySQL and Doris storagers so the two backends keep
// an identical caliber.
var DimensionColumns = map[string]string{
	DimensionModel:          "ai_target_model",
	DimensionRequestedModel: "ai_requested_model",
	DimensionProvider:       "ai_provider",
	DimensionAPIKey:         "ai_apikey_id",
	DimensionHost:           "hostid",
	DimensionStatus:         "res_status_code",
	DimensionProtocol:       "ai_protocol",
	DimensionMode:           "ai_mode",
	DimensionStream:         "ai_stream",
}

// Filter carries the filter criteria shared by all report endpoints.
type Filter struct {
	Start       time.Time // required, window start (inclusive)
	End         time.Time // required, window end (exclusive), End > Start
	Models      []string  // ai_target_model
	ApikeyIDs   []string  // ai_apikey_id
	Providers   []string  // ai_provider
	Hosts       []string  // hostid
	Stream      *bool     // ai_stream
	StatusCodes []int     // res_status_code
}

// LogFilter extends Filter with log-list-only criteria.
type LogFilter struct {
	Filter
	RequestedModels []string // ai_requested_model
	ErrOnly         bool     // only rows with non-empty err_code
	Keyword         string   // fuzzy match against err_msg
	Page            int      // 1-based
	PageSize        int
}

// CostItem is one currency bucket of the aggregated cost. Value is the
// cost amount in the currency unit (yuan for RMB, dollar for USD),
// converted from the fixed-point value by CostFixedPointToAmount.
type CostItem struct {
	Currency string  `json:"currency"`
	Value    float64 `json:"value"`
}

// OverviewResult is the data of GET /report/overview. LatencyP50Ms /
// LatencyP90Ms / LatencyP99Ms only exist on the Doris backend; the MySQL
// backend leaves them nil so the fields are omitted from the JSON payload
// and the frontend degrades to avg/max.
type OverviewResult struct {
	RequestTotal int64   `json:"request_total"`
	ErrorTotal   int64   `json:"error_total"`
	ErrorRate    float64 `json:"error_rate"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`

	LatencyAvgMs float64  `json:"latency_avg_ms"`
	LatencyMaxMs float64  `json:"latency_max_ms"`
	LatencyP50Ms *float64 `json:"latency_p50_ms,omitempty"`
	LatencyP90Ms *float64 `json:"latency_p90_ms,omitempty"`
	LatencyP99Ms *float64 `json:"latency_p99_ms,omitempty"`

	TtftAvgMs float64 `json:"ttft_avg_ms"`
	TpotAvgMs float64 `json:"tpot_avg_ms"`

	Cost          []*CostItem `json:"cost"`
	RateLimitHits int64       `json:"rate_limit_hits"`
	AuthRejects   int64       `json:"auth_rejects"`
	LogsTotal     int64       `json:"logs_total"`
}

// MetricPoint is one time-series point. Value is set for single-value
// metrics (qps / ttft / tpot / cost); Input/Output/Total for tokens;
// Avg/Max for latency; P50/P90/P99 only on the Doris latency series.
// Currency distinguishes per-currency cost points.
type MetricPoint struct {
	Time     int64    `json:"time"`
	Value    *float64 `json:"value,omitempty"`
	Input    *float64 `json:"input,omitempty"`
	Output   *float64 `json:"output,omitempty"`
	Total    *float64 `json:"total,omitempty"`
	Avg      *float64 `json:"avg,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	P50      *float64 `json:"p50,omitempty"`
	P90      *float64 `json:"p90,omitempty"`
	P99      *float64 `json:"p99,omitempty"`
	Currency string   `json:"currency,omitempty"`
}

// RankingItem is one row of GET /report/rankings.
type RankingItem struct {
	Name         string `json:"name"`
	RequestCount int64  `json:"request_count"`
	ErrorCount   int64  `json:"error_count"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
}

// DistItem is one bucket of GET /report/distribution. Name is "unknown"
// when the dimension value is NULL or empty.
type DistItem struct {
	Name         string  `json:"name"`
	RequestCount int64   `json:"request_count"`
	Ratio        float64 `json:"ratio"`
}

// LogRow is the display projection of one bfe_ai_request_log row; field
// names align with the table columns. JSON columns are returned verbatim
// as strings for the frontend to expand. Nullable columns are pointers.
// CostValue is the cost amount in the currency unit (converted from the
// fixed-point value), nil when the row has no cost.
type LogRow struct {
	LogID               *int64  `json:"logid"`
	LogTime             int64   `json:"log_time"`
	Hostid              *string `json:"hostid"`
	Product             *string `json:"product"`
	APIKeyID            *string `json:"ai_apikey_id"`
	RequestedModel      *string `json:"ai_requested_model"`
	TargetModel         *string `json:"ai_target_model"`
	Provider            *string `json:"ai_provider"`
	Protocol            *string `json:"ai_protocol"`
	Mode                *string `json:"ai_mode"`
	Stream              *int16  `json:"ai_stream"`
	StatusCode          *int16  `json:"res_status_code"`
	ErrCode             *string `json:"err_code"`
	ErrMsg              *string `json:"err_msg"`
	InputTokens         *int64  `json:"ai_input_tokens"`
	OutputTokens        *int64  `json:"ai_output_tokens"`
	TotalTokens         *int64  `json:"ai_total_tokens"`
	AllTime             *int64  `json:"all_time"`
	TTFTUs              *int64  `json:"ai_ttft_us"`
	TPOTUs              *int64  `json:"ai_tpot_us"`
	CostValue           *float64 `json:"ai_cost_value"`
	CostCurrency        *string `json:"ai_cost_currency"`
	RateLimitHits       *string `json:"ai_rate_limit_hits"`
	AuthRejectQuotaPlan *string `json:"ai_auth_reject_quota_plans"`
	Level1Name          *string `json:"level1Name"`
	Level1              *string `json:"level1"`
	Level2Name          *string `json:"level2Name"`
	Level2              *string `json:"level2"`
	Level3Name          *string `json:"level3Name"`
	Level3              *string `json:"level3"`
	Level4Name          *string `json:"level4Name"`
	Level4              *string `json:"level4"`
	Level5Name          *string `json:"level5Name"`
	Level5              *string `json:"level5"`
	ClientIP            *string `json:"client_ip"`
	HeaderHost          *string `json:"header_host"`
	OriginURI           *string `json:"origin_uri"`
	ReqHeaders          *string `json:"req_headers"`
	ResHeaders          *string `json:"res_headers"`
}

// LogQueryResult is the data of GET /report/logs.
type LogQueryResult struct {
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Items    []*LogRow `json:"items"`
}

// ReportStorager defines the storage operations backing the report queries.
// Implementations exist for MySQL (storage/mysqlreport) and Doris
// (storage/dorisreport); the manager never contains SQL.
type ReportStorager interface {
	Overview(ctx context.Context, f *Filter) (*OverviewResult, error)
	TimeSeries(ctx context.Context, metric string, f *Filter, bucketSec int) ([]*MetricPoint, error)
	Rankings(ctx context.Context, dimension string, f *Filter, limit int) ([]*RankingItem, error)
	Distribution(ctx context.Context, dimension string, f *Filter) ([]*DistItem, error)
	Logs(ctx context.Context, f *LogFilter) (*LogQueryResult, error)
}

// BucketSeconds returns the time-bucket width in seconds for a query
// window: <=6h -> 60s, <=3d -> 300s, <=7d -> 1800s. It is a pure function.
func BucketSeconds(window time.Duration) int {
	switch {
	case window <= 6*time.Hour:
		return 60
	case window <= 72*time.Hour:
		return 300
	default:
		return 1800
	}
}
