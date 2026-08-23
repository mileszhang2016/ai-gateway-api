package openapi

import (
	"fmt"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/require"
)

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
	var err error
	sm, err = testutil.StartServer()
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

func TestOpenAPI_Schema(t *testing.T) {
	t.Run("entity_types", testEntityTypeSchema)
	t.Run("entities", testEntitySchema)
	t.Run("api_keys", testAPIKeySchema)
	t.Run("providers", testProviderSchema)
	t.Run("clusters", testClusterSchema)
	t.Run("certificates", testCertificateSchema)
	t.Run("auth", testAuthSchema)
	t.Run("model_prices", testModelPriceSchema)
	t.Run("route_tables", testRouteTableSchema)
	t.Run("global_route_rules", testGlobalRouteRulesSchema)
}

// ---------- entity-types ----------

func testEntityTypeSchema(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()

	resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name":   typeName,
		"description": "schema test",
		"level":       1,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	testutil.AssertSchema(t, resp, EntityTypeSchema)

	listResp, err := testutil.GetClient().Get("/open-api/v1/entity-types")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertPagedListSchema(t, listResp, EntityTypeSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/entity-types/" + typeName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, EntityTypeSchema)

	patchResp, err := testutil.GetClient().Patch("/open-api/v1/entity-types/"+typeName, map[string]interface{}{
		"description": "updated schema test",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, patchResp)
	testutil.AssertSchema(t, patchResp, EntityTypeSchema)

	t.Cleanup(func() {
		testutil.DeleteEntityType(typeName)
	})
}

// ---------- entities ----------

func testEntitySchema(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	_, err := testutil.CreateEntityType(typeName, 1)
	require.NoError(t, err)

	entityName := testutil.UniqueEntityName()
	createResp, err := testutil.GetClient().Post("/open-api/v1/entities", map[string]interface{}{
		"name": entityName,
		"type": typeName,
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, EntitySchema)

	id, err := testutil.GetDataField(createResp, "id")
	require.NoError(t, err)
	entityID := id.(string)

	listResp, err := testutil.GetClient().Get("/open-api/v1/entities")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertPagedListSchema(t, listResp, EntityListItemSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, EntitySchema)

	putResp, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"name": entityName + "-updated",
		"type": typeName,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, putResp)
	testutil.AssertSchema(t, putResp, EntitySchema)

	patchResp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"allow_models": []string{"*"},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, patchResp)
	testutil.AssertSchema(t, patchResp, EntitySchema)

	qpResp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID + "/quota-plan")
	require.NoError(t, err)
	testutil.AssertSuccess(t, qpResp)
	testutil.AssertSchema(t, qpResp, QuotaPlanWithBalanceSchema)

	resetResp, err := testutil.GetClient().Post("/open-api/v1/entities/"+entityID+"/quota-plan/reset", map[string]interface{}{})
	require.NoError(t, err)
	testutil.AssertSuccess(t, resetResp)
	testutil.AssertSchema(t, resetResp, QuotaResetResultSchema)

	t.Cleanup(func() {
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
	})
}

// ---------- api-keys ----------

func testAPIKeySchema(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	_, err := testutil.CreateEntityType(typeName, 1)
	require.NoError(t, err)

	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	require.NoError(t, err)

	clusterName := testutil.UniqueClusterName()
	_, err = testutil.CreateCluster(clusterName)
	require.NoError(t, err)

	createResp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "schema test",
		"entity_id":   entityID,
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{"name": "tpm-1m", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1},
				},
				"rpm": []interface{}{
					map[string]interface{}{"name": "rpm-1m", "model": "*", "window_minutes": 1, "max_requests": 100},
				},
				"max_concurrency": 50,
			},
		},
		"route_rules": map[string]interface{}{
			"enabled": true,
			"rules": []interface{}{
				map[string]interface{}{
					"name":  "default",
					"cond":  "default_t()",
					"targets": []interface{}{
						map[string]interface{}{"cluster_name": clusterName, "model": "", "weight": 100},
					},
					"fallbacks": []interface{}{},
				},
			},
		},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, APIKeySchema)

	id, err := testutil.GetDataField(createResp, "id")
	require.NoError(t, err)
	apiKeyID := id.(string)

	listResp, err := testutil.GetClient().Get("/open-api/v1/api-keys")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertPagedListSchema(t, listResp, APIKeyListItemSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, APIKeyListItemSchema)

	putResp, err := testutil.GetClient().Put("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
		"description": "schema test updated",
		"enabled":     true,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, putResp)
	testutil.AssertSchema(t, putResp, APIKeySchema)

	patchResp, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
		"description": "schema test patched",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, patchResp)
	testutil.AssertSchema(t, patchResp, APIKeySchema)

	qpResp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID + "/quota-plan")
	require.NoError(t, err)
	testutil.AssertSuccess(t, qpResp)
	testutil.AssertSchema(t, qpResp, QuotaPlanWithBalanceSchema)

	resetResp, err := testutil.GetClient().Post("/open-api/v1/api-keys/"+apiKeyID+"/quota-plan/reset", map[string]interface{}{})
	require.NoError(t, err)
	testutil.AssertSuccess(t, resetResp)
	testutil.AssertSchema(t, resetResp, QuotaResetResultSchema)

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
		testutil.DeleteCluster(clusterName)
	})
}

// ---------- providers ----------

func testProviderSchema(t *testing.T) {
	providerName := testutil.UniqueProviderName()

	createResp, err := testutil.GetClient().Post("/open-api/v1/providers", map[string]interface{}{
		"name":        providerName,
		"description": "schema test",
		"instance_pool": []interface{}{
			map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080},
		},
		"model_protocols": []string{"openai"},
		"models":          []string{"deepseek-chat"},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, ProviderSchema)

	listResp, err := testutil.GetClient().Get("/open-api/v1/providers")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertPagedListSchema(t, listResp, ProviderSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/providers/" + providerName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, ProviderSchema)

	patchResp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
		"name":            providerName,
		"description":     "schema test updated",
		"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
		"model_protocols": []string{"openai"},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, patchResp)
	testutil.AssertSchema(t, patchResp, ProviderSchema)

	t.Cleanup(func() {
		testutil.DeleteProvider(providerName)
	})
}

// ---------- clusters ----------

func testClusterSchema(t *testing.T) {
	providerName := testutil.UniqueProviderName()
	_, err := testutil.CreateProvider(providerName)
	require.NoError(t, err)

	clusterName := testutil.UniqueClusterName()

	createResp, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
		"name":        clusterName,
		"description": "schema test",
		"llm_config": map[string]interface{}{
			"models":   []string{"deepseek-chat"},
			"provider": providerName,
		},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, ClusterSchema)

	listResp, err := testutil.GetClient().Get("/open-api/v1/clusters")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertListSchema(t, listResp, ClusterSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/clusters/" + clusterName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, ClusterSchema)

	patchResp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
		"description": "schema test updated",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, patchResp)
	testutil.AssertSchema(t, patchResp, ClusterSchema)

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
		testutil.DeleteProvider(providerName)
	})
}

// ---------- certificates ----------

func testCertificateSchema(t *testing.T) {
	certName := testutil.UniqueCertName()
	certPEM, keyPEM, err := testutil.GenerateTestCert(certName)
	require.NoError(t, err)

	createResp, err := testutil.GetClient().Post("/open-api/v1/certificates", map[string]interface{}{
		"cert_name":         certName,
		"description":       "schema test",
		"is_default":        true,
		"cert_file_content": certPEM,
		"key_file_content":  keyPEM,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, CertificateSchema)

	listResp, err := testutil.GetClient().Get("/open-api/v1/certificates")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertListSchema(t, listResp, CertificateSchema)

	oneResp, err := testutil.GetClient().Get("/open-api/v1/certificates/" + certName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, CertificateSchema)

	defaultResp, err := testutil.GetClient().Patch("/open-api/v1/certificates/"+certName+"/default", map[string]interface{}{})
	require.NoError(t, err)
	testutil.AssertSuccess(t, defaultResp)
	testutil.AssertSchema(t, defaultResp, CertificateSchema)

	t.Cleanup(func() {
		// 默认证书不能直接删除，需先创建另一个默认证书再删
		// 这里简单处理：保留默认证书，测试中忽略删除错误
		_ = testutil.DeleteCertificate(certName)
	})
}

// ---------- auth ----------

func testAuthSchema(t *testing.T) {
	userName := testutil.UniqueUserName()
	password := "Password123!"

	createUserResp, err := testutil.GetClient().Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": userName,
		"password":  password,
		"is_admin":  true,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createUserResp)

	listUsersResp, err := testutil.GetClient().Get("/open-api/v1/auth/users")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listUsersResp)
	testutil.AssertListSchema(t, listUsersResp, UserSchema)

	oneUserResp, err := testutil.GetClient().Get("/open-api/v1/auth/users/" + userName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneUserResp)
	testutil.AssertSchema(t, oneUserResp, UserSchema)

	sessionResp, err := testutil.GetClient().Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"user_name": userName,
		"password":  password,
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, sessionResp)
	testutil.AssertSchema(t, sessionResp, SessionKeySchema)

	sessionKeyVal, err := testutil.GetDataField(sessionResp, "session_key")
	require.NoError(t, err)

	tokenName := testutil.UniqueTokenName()
	createTokenResp, err := testutil.GetClient().Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":  tokenName,
		"scope": "System",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createTokenResp)
	testutil.AssertSchema(t, createTokenResp, CreateTokenResponseSchema)

	listTokensResp, err := testutil.GetClient().Get("/open-api/v1/auth/tokens")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listTokensResp)
	testutil.AssertListSchema(t, listTokensResp, TokenSchema)

	oneTokenResp, err := testutil.GetClient().Get("/open-api/v1/auth/tokens/" + tokenName)
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneTokenResp)
	testutil.AssertSchema(t, oneTokenResp, TokenSchema)

	metaResp, err := testutil.GetClient().Get("/open-api/v1/meta")
	require.NoError(t, err)
	testutil.AssertSuccess(t, metaResp)
	testutil.AssertSchema(t, metaResp, MetaSchema)

	t.Cleanup(func() {
		testutil.DeleteToken(tokenName)
		testutil.GetClient().Delete("/open-api/v1/auth/session-keys/" + sessionKeyVal.(string))
		testutil.DeleteUser(userName)
	})
}

// ---------- model-prices ----------

func testModelPriceSchema(t *testing.T) {
	schemaProvider := "schema-test-provider"
	schemaProvider2 := "schema-test-provider-2"
	_, err := testutil.CreateProvider(schemaProvider)
	require.NoError(t, err)
	_, err = testutil.CreateProvider(schemaProvider2)
	require.NoError(t, err)

	yamlContent := []byte(`version: v1.0
default_currency: RMB
models:
  - provider: schema-test-provider
    model: schema-test-model
    base_model: schema-test-model
    mode: chat
    capabilities: [chat]
    supported_parameters: [temperature]
    limits:
      context_window: 128000
    prices:
      input_cost_per_token: 0.000002
      output_cost_per_token: 0.000008
    metadata:
      source: test
`)
	importResp, err := testutil.GetClient().PostMultipartFile("/open-api/v1/model-prices/import", "file", "model-list.yaml", yamlContent, map[string]string{
		"mode": "replace",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, importResp)
	testutil.AssertSchema(t, importResp, ModelPriceImportResultSchema)

	createResp, err := testutil.GetClient().Post("/open-api/v1/model-prices", map[string]interface{}{
		"provider":             schemaProvider2,
		"model":                "schema-test-model-2",
		"base_model":           "schema-test-model-2",
		"mode":                 "chat",
		"capabilities":         []string{"chat"},
		"supported_parameters": []string{"temperature"},
		"limits": map[string]interface{}{
			"context_window": 128000,
		},
		"prices": map[string]interface{}{
			"input_cost_per_token":  0.000002,
			"output_cost_per_token": 0.000008,
		},
		"metadata": map[string]interface{}{
			"source": "test",
		},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, createResp)
	testutil.AssertSchema(t, createResp, ModelPriceSchema)

	id, err := testutil.GetDataField(createResp, "id")
	require.NoError(t, err)
	modelPriceID := int64(id.(float64))

	listResp, err := testutil.GetClient().Get("/open-api/v1/model-prices")
	require.NoError(t, err)
	testutil.AssertSuccess(t, listResp)
	testutil.AssertSchema(t, listResp, ModelPriceListResponseSchema)

	oneResp, err := testutil.GetClient().Get(fmt.Sprintf("/open-api/v1/model-prices/%d", modelPriceID))
	require.NoError(t, err)
	testutil.AssertSuccess(t, oneResp)
	testutil.AssertSchema(t, oneResp, ModelPriceSchema)

	updateResp, err := testutil.GetClient().Put(fmt.Sprintf("/open-api/v1/model-prices/%d", modelPriceID), map[string]interface{}{
		"provider":             schemaProvider2,
		"model":                "schema-test-model-2",
		"base_model":           "schema-test-model-2",
		"mode":                 "chat",
		"capabilities":         []string{"chat", "vision"},
		"supported_parameters": []string{"temperature"},
		"limits": map[string]interface{}{
			"context_window": 128000,
		},
		"prices": map[string]interface{}{
			"input_cost_per_token":  0.000003,
			"output_cost_per_token": 0.000009,
		},
		"metadata": map[string]interface{}{
			"source": "test",
		},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, updateResp)
	testutil.AssertSchema(t, updateResp, ModelPriceSchema)

	getProvidersResp, err := testutil.GetClient().Get("/open-api/v1/model-prices/actions/get-providers")
	require.NoError(t, err)
	testutil.AssertSuccess(t, getProvidersResp)
	testutil.AssertSchema(t, getProvidersResp, ModelPriceGetProvidersResponseSchema)

	t.Cleanup(func() {
		testutil.DeleteModelPrice(modelPriceID)
		testutil.DeleteModelPriceByQuery(schemaProvider, "schema-test-model", "chat")
		testutil.DeleteProvider(schemaProvider)
		testutil.DeleteProvider(schemaProvider2)
	})
}

// ---------- route-tables ----------

func testRouteTableSchema(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/route-tables")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	testutil.AssertPagedListSchema(t, resp, RouteTableSchema)
}

// ---------- global-route-rules ----------

func testGlobalRouteRulesSchema(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/global-route-rules")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	testutil.AssertSchema(t, resp, GlobalRouteRulesSchema)

	putResp, err := testutil.GetClient().Put("/open-api/v1/global-route-rules", map[string]interface{}{
		"enabled": true,
		"rules":   []interface{}{},
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, putResp)
	testutil.AssertSchema(t, putResp, GlobalRouteRulesSchema)
}


