package openapi

import "github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"

// EntitySchema Entity 数据模型 schema（不含 balance）
// parent_id 对根节点为 null，因此设为可选。
var EntitySchema = &testutil.ObjectSchema{
	Required: []string{
		"id", "name", "type",
		"allow_models", "block_models",
		"quota_plan", "rate_limit_policy", "route_rules",
		"create_time", "update_time",
	},
	Optional: []string{"parent_id"},
	Fields: map[string]testutil.FieldSpec{
		"id":                {Type: testutil.TypeString},
		"name":              {Type: testutil.TypeString},
		"type":              {Type: testutil.TypeString},
		"parent_id":         {Type: testutil.TypeString},
		"allow_models":      {Type: testutil.TypeArray},
		"block_models":      {Type: testutil.TypeArray},
		"quota_plan":        {Type: testutil.TypeObject, Nested: QuotaPlanWithoutBalanceSchema},
		"rate_limit_policy": {Type: testutil.TypeObject, Nested: RateLimitPolicySchema},
		"route_rules":       {Type: testutil.TypeObject, Nested: RouteRulesSchema},
		"create_time":       {Type: testutil.TypeInt},
		"update_time":       {Type: testutil.TypeInt},
	},
}

// EntityListItemSchema Entity 列表元素 schema（不含 balance）
var EntityListItemSchema = EntitySchema
