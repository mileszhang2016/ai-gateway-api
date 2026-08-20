package openapi

import "github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"

// APIKeySchema API-Key 数据模型 schema（不含 balance）
var APIKeySchema = &testutil.ObjectSchema{
	Required: []string{
		"id", "key", "description", "enabled",
		"create_time", "update_time", "expired_time",
		"unlimited_quota", "models", "subnet",
		"quota_plan", "rate_limit_policy", "route_rules",
		"entity_id",
	},
	Optional: []string{"entity"},
	Fields: map[string]testutil.FieldSpec{
		"id":              {Type: testutil.TypeString},
		"key":             {Type: testutil.TypeString},
		"description":     {Type: testutil.TypeString},
		"enabled":         {Type: testutil.TypeBool},
		"create_time":     {Type: testutil.TypeInt},
		"update_time":     {Type: testutil.TypeInt},
		"expired_time":    {Type: testutil.TypeInt},
		"unlimited_quota": {Type: testutil.TypeBool},
		"models":          {Type: testutil.TypeArray},
		"subnet":          {Type: testutil.TypeArray},
		"quota_plan":      {Type: testutil.TypeObject, Nested: QuotaPlanWithoutBalanceSchema},
		"rate_limit_policy": {Type: testutil.TypeObject, Nested: RateLimitPolicySchema},
		"route_rules":     {Type: testutil.TypeObject, Nested: RouteRulesSchema},
		"entity_id":       {Type: testutil.TypeString},
		"entity":          {Type: testutil.TypeObject, Nested: EntitySummarySchema},
	},
}

// APIKeyListItemSchema API-Key 列表元素 schema（含 balance）
var APIKeyListItemSchema = &testutil.ObjectSchema{
	Required: []string{
		"id", "key", "description", "enabled",
		"create_time", "update_time", "expired_time",
		"unlimited_quota", "models", "subnet",
		"quota_plan", "rate_limit_policy", "route_rules",
		"entity_id",
	},
	Optional: []string{"entity"},
	Fields: map[string]testutil.FieldSpec{
		"id":              {Type: testutil.TypeString},
		"key":             {Type: testutil.TypeString},
		"description":     {Type: testutil.TypeString},
		"enabled":         {Type: testutil.TypeBool},
		"create_time":     {Type: testutil.TypeInt},
		"update_time":     {Type: testutil.TypeInt},
		"expired_time":    {Type: testutil.TypeInt},
		"unlimited_quota": {Type: testutil.TypeBool},
		"models":          {Type: testutil.TypeArray},
		"subnet":          {Type: testutil.TypeArray},
		"quota_plan":      {Type: testutil.TypeObject, Nested: QuotaPlanWithBalanceSchema},
		"rate_limit_policy": {Type: testutil.TypeObject, Nested: RateLimitPolicySchema},
		"route_rules":     {Type: testutil.TypeObject, Nested: RouteRulesSchema},
		"entity_id":       {Type: testutil.TypeString},
		"entity":          {Type: testutil.TypeObject, Nested: EntitySummarySchema},
	},
}
