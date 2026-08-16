// Package innerapi 定义 InnerAPI v1 各接口返回值的 schema。
// 由于 InnerAPI 大量字段为动态 key 的 map，本包优先校验顶层固定字段的类型，
// 对动态 map 仅校验其为 object。
package innerapi

import "github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"

// ---------- 公共/复用 schema ----------

// InnerAPIBaseSchema 所有 InnerAPI 导出配置的顶层基础字段（Version 为 string）
var InnerAPIBaseSchema = &testutil.ObjectSchema{
	Required: []string{"Version"},
	Fields: map[string]testutil.FieldSpec{
		"Version": {Type: testutil.TypeString},
	},
}

// ServerDataConfSchema /configs/tls_conf/server_data_conf 返回 schema
var ServerDataConfSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "HostTable", "RouteTable", "ClusterConf"},
	Fields: map[string]testutil.FieldSpec{
		"Version":    {Type: testutil.TypeString},
		"HostTable":  {Type: testutil.TypeObject},
		"RouteTable": {Type: testutil.TypeObject},
		"ClusterConf":{Type: testutil.TypeObject},
	},
}

// GSLBSchema /configs/gslb_data/gslb 返回 schema
var GSLBSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "Ts", "Hostname", "Clusters"},
	Fields: map[string]testutil.FieldSpec{
		"Version":  {Type: testutil.TypeString},
		"Ts":       {Type: testutil.TypeString},
		"Hostname": {Type: testutil.TypeString},
		"Clusters": {Type: testutil.TypeObject},
	},
}

// ClusterTableSchema /configs/gslb_data/cluster_table 返回 schema
var ClusterTableSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "Config"},
	Fields: map[string]testutil.FieldSpec{
		"Version": {Type: testutil.TypeString},
		"Config":  {Type: testutil.TypeObject},
	},
}

// ServerCertConfSchema /configs/protocol/server_cert_conf 返回 schema
var ServerCertConfSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "Config"},
	Fields: map[string]testutil.FieldSpec{
		"Version": {Type: testutil.TypeString},
		"Config":  {Type: testutil.TypeObject},
	},
}

// ModAPIKeySchema /configs/mod-api-key 返回 schema
var ModAPIKeySchema = &testutil.ObjectSchema{
	Required: []string{"version", "config", "QuotaPlans", "tokens"},
	Fields: map[string]testutil.FieldSpec{
		"version":    {Type: testutil.TypeString},
		"config":     {Type: testutil.TypeObject},
		"QuotaPlans": {Type: testutil.TypeObject},
		"tokens":     {Type: testutil.TypeObject},
	},
}

// ModBodyProcessSchema /configs/mod-body-process 返回 schema
var ModBodyProcessSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "Config"},
	Fields: map[string]testutil.FieldSpec{
		"Version": {Type: testutil.TypeString},
		"Config":  {Type: testutil.TypeObject},
	},
}

// RateLimitPolicySchema /configs/rate-limit-policy 返回 schema
var RateLimitPolicySchema = &testutil.ObjectSchema{
	Required: []string{"Config", "RateLimitPolicies", "ApikeyRateLimitPolicyBindings", "Version"},
	Fields: map[string]testutil.FieldSpec{
		"Config":                        {Type: testutil.TypeObject},
		"RateLimitPolicies":             {Type: testutil.TypeObject},
		"ApikeyRateLimitPolicyBindings": {Type: testutil.TypeObject},
		"Version":                       {Type: testutil.TypeString},
	},
}

// AIRouteSchema /configs/ai-route 返回 schema
var AIRouteSchema = &testutil.ObjectSchema{
	Required: []string{"Version", "RouteRules", "ApikeyRouteTableBindings"},
	Fields: map[string]testutil.FieldSpec{
		"Version":                  {Type: testutil.TypeString},
		"RouteRules":               {Type: testutil.TypeObject},
		"ApikeyRouteTableBindings": {Type: testutil.TypeObject},
	},
}
