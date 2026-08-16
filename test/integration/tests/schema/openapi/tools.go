package openapi

import "github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"

// ModelFromProviderItemSchema /tools/get-models-from-provider 返回数组元素 schema
var ModelFromProviderItemSchema = &testutil.ObjectSchema{
	Required: []string{"id", "name"},
	Fields: map[string]testutil.FieldSpec{
		"id":   {Type: testutil.TypeString},
		"name": {Type: testutil.TypeString},
	},
}
