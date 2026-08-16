package openapi

import "github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"

// RouteTableSchema 路由表元素 schema
var RouteTableSchema = &testutil.ObjectSchema{
	Required: []string{"id", "type", "owner", "enabled"},
	Fields: map[string]testutil.FieldSpec{
		"id":      {Type: testutil.TypeInt},
		"type":    {Type: testutil.TypeString, Enum: []interface{}{"global", "entity", "api_key"}},
		"owner":   {Type: testutil.TypeString},
		"enabled": {Type: testutil.TypeBool},
	},
}
