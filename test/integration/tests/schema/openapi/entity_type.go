package openapi

import "github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"

// EntityTypeSchema Entity-Type 数据模型 schema
var EntityTypeSchema = &testutil.ObjectSchema{
	Required: []string{"type_name", "description", "level", "create_time"},
	Fields: map[string]testutil.FieldSpec{
		"type_name":   {Type: testutil.TypeString},
		"description": {Type: testutil.TypeString},
		"level":       {Type: testutil.TypeInt},
		"create_time": {Type: testutil.TypeInt},
	},
}
