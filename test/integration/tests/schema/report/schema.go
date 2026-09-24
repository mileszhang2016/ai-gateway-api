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

// Package report 定义 /open-api/v1/report/* 接口返回值的 schema。
// schema 根据 ai-gateway-api/design-docs/api-define/OpenAPI接口定义/report.md 编写，
// 用于在集成测试中严格校验接口返回字段的存在性与类型；成本字段自 issue #207 起
// 为金额口径（元/美元，服务端已完成 ÷1e8 换算），value 为 number 且通常为非整数。
package report

import "github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"

// CostItemSchema 是 overview cost 数组元素（按币种分桶的成本金额）。
var CostItemSchema = &testutil.ObjectSchema{
	Required: []string{"currency", "value"},
	Fields: map[string]testutil.FieldSpec{
		"currency": {Type: testutil.TypeString},
		"value":    {Type: testutil.TypeNumber},
	},
}

// OverviewResultSchema 是 GET /report/overview 的 Data schema。
var OverviewResultSchema = &testutil.ObjectSchema{
	Required: []string{
		"request_total", "error_total", "error_rate",
		"input_tokens", "output_tokens", "total_tokens",
		"latency_avg_ms", "latency_max_ms",
		"ttft_avg_ms", "tpot_avg_ms",
		"cost", "rate_limit_hits", "auth_rejects", "logs_total",
	},
	Fields: map[string]testutil.FieldSpec{
		"request_total":   {Type: testutil.TypeInt},
		"error_total":     {Type: testutil.TypeInt},
		"error_rate":      {Type: testutil.TypeNumber},
		"input_tokens":    {Type: testutil.TypeInt},
		"output_tokens":   {Type: testutil.TypeInt},
		"total_tokens":    {Type: testutil.TypeInt},
		"latency_avg_ms":  {Type: testutil.TypeNumber},
		"latency_max_ms":  {Type: testutil.TypeNumber},
		"ttft_avg_ms":     {Type: testutil.TypeNumber},
		"tpot_avg_ms":     {Type: testutil.TypeNumber},
		"cost":            {Type: testutil.TypeArray, Elem: CostItemSchema},
		"rate_limit_hits": {Type: testutil.TypeInt},
		"auth_rejects":    {Type: testutil.TypeInt},
		"logs_total":      {Type: testutil.TypeInt},
	},
}

// CostMetricPointSchema 是 timeseries metric=cost 的序列点
// （金额/秒，currency 字段区分币种）。
var CostMetricPointSchema = &testutil.ObjectSchema{
	Required: []string{"time", "value", "currency"},
	Fields: map[string]testutil.FieldSpec{
		"time":     {Type: testutil.TypeInt},
		"value":    {Type: testutil.TypeNumber},
		"currency": {Type: testutil.TypeString},
	},
}

// TimeSeriesCostDataSchema 是 GET /report/timeseries?metric=cost 的 Data schema。
var TimeSeriesCostDataSchema = &testutil.ObjectSchema{
	Required: []string{"bucket_sec", "series"},
	Fields: map[string]testutil.FieldSpec{
		"bucket_sec": {Type: testutil.TypeInt},
		"series":     {Type: testutil.TypeArray, Elem: CostMetricPointSchema},
	},
}

// LogRowSchema 是 /report/logs items 元素的 schema（字段集为明细行投影的子集，
// 覆盖本用例组语义断言涉及的列；可空列均为 Optional，null 时跳过类型校验）。
var LogRowSchema = &testutil.ObjectSchema{
	Required: []string{"logid", "log_time"},
	Optional: []string{
		"ai_apikey_id", "err_code", "err_msg",
		"ai_cost_value", "ai_cost_currency",
	},
	Fields: map[string]testutil.FieldSpec{
		"logid":            {Type: testutil.TypeInt},
		"log_time":         {Type: testutil.TypeInt},
		"ai_apikey_id":     {Type: testutil.TypeString},
		"err_code":         {Type: testutil.TypeString},
		"err_msg":          {Type: testutil.TypeString},
		"ai_cost_value":    {Type: testutil.TypeNumber},
		"ai_cost_currency": {Type: testutil.TypeString},
	},
}

// LogQueryResultSchema 是 GET /report/logs 的 Data schema。
var LogQueryResultSchema = &testutil.ObjectSchema{
	Required: []string{"total", "page", "page_size", "items"},
	Fields: map[string]testutil.FieldSpec{
		"total":     {Type: testutil.TypeInt},
		"page":      {Type: testutil.TypeInt},
		"page_size": {Type: testutil.TypeInt},
		"items":     {Type: testutil.TypeArray, Elem: LogRowSchema},
	},
}
