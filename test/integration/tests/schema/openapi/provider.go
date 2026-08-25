package openapi

import "github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"

// ProviderEndpointSchema provider model_endpoint schema
var ProviderEndpointSchema = &testutil.ObjectSchema{
	Required: []string{"schema", "uri"},
	Fields: map[string]testutil.FieldSpec{
		"schema": {Type: testutil.TypeString},
		"uri":    {Type: testutil.TypeString},
	},
}

// ProviderKeySchema provider keys element schema
var ProviderKeySchema = &testutil.ObjectSchema{
	Required: []string{"name", "key"},
	Fields: map[string]testutil.FieldSpec{
		"name": {Type: testutil.TypeString},
		"key":  {Type: testutil.TypeString},
	},
}

// ProviderInstanceSchema provider instance_pool element schema
var ProviderInstanceSchema = &testutil.ObjectSchema{
	Required: []string{"name", "addr", "weight", "port"},
	Fields: map[string]testutil.FieldSpec{
		"name":   {Type: testutil.TypeString},
		"addr":   {Type: testutil.TypeString},
		"weight": {Type: testutil.TypeInt},
		"port":   {Type: testutil.TypeInt},
	},
}

// TimeRangeSchema provider tiers.time_ranges element schema
var TimeRangeSchema = &testutil.ObjectSchema{
	Required: []string{"start", "end"},
	Fields: map[string]testutil.FieldSpec{
		"weekdays": {Type: testutil.TypeArray, Item: &testutil.FieldSpec{Type: testutil.TypeInt}},
		"start":    {Type: testutil.TypeString},
		"end":      {Type: testutil.TypeString},
	},
}

// PricingTierSchema provider tiers element schema
var PricingTierSchema = &testutil.ObjectSchema{
	Required: []string{"name", "time_ranges"},
	Fields: map[string]testutil.FieldSpec{
		"name":        {Type: testutil.TypeString},
		"time_ranges": {Type: testutil.TypeArray, Elem: TimeRangeSchema},
	},
}

// ProviderSchema Provider 数据模型 schema
var ProviderSchema = &testutil.ObjectSchema{
	Required: []string{
		"id", "name", "description", "model_endpoint", "models", "keys",
		"instance_pool", "model_protocols", "time_zone", "tiers",
		"create_time", "update_time",
	},
	Fields: map[string]testutil.FieldSpec{
		"id":              {Type: testutil.TypeInt},
		"name":            {Type: testutil.TypeString},
		"description":     {Type: testutil.TypeString},
		"model_endpoint":  {Type: testutil.TypeObject, Nested: ProviderEndpointSchema},
		"models":          {Type: testutil.TypeArray, Item: &testutil.FieldSpec{Type: testutil.TypeString}},
		"keys":            {Type: testutil.TypeArray, Elem: ProviderKeySchema},
		"instance_pool":   {Type: testutil.TypeArray, Elem: ProviderInstanceSchema},
		"model_protocols": {Type: testutil.TypeArray, Item: &testutil.FieldSpec{Type: testutil.TypeString, Enum: []interface{}{"openai", "anthropic"}}},
		"time_zone":       {Type: testutil.TypeString},
		"tiers":           {Type: testutil.TypeArray, Elem: PricingTierSchema},
		"create_time":     {Type: testutil.TypeInt},
		"update_time":     {Type: testutil.TypeInt},
	},
}
