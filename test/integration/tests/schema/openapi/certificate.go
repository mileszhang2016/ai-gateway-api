package openapi

import "github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"

// CertificateSchema 证书数据模型 schema（创建/更新/详情/列表返回，不包含 cert_file_content/key_file_content）
var CertificateSchema = &testutil.ObjectSchema{
	Required: []string{"cert_name", "description", "is_default", "expired_date"},
	Fields: map[string]testutil.FieldSpec{
		"cert_name":   {Type: testutil.TypeString},
		"description": {Type: testutil.TypeString},
		"is_default":  {Type: testutil.TypeBool},
		"expired_date":{Type: testutil.TypeString},
	},
}
