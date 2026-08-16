package innerapi

import (
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestInnerAPI_Schema(t *testing.T) {
	t.Run("server_data_conf", testServerDataConfSchema)
	t.Run("gslb", testGSLBSchema)
	t.Run("cluster_table", testClusterTableSchema)
	t.Run("server_cert_conf", testServerCertConfSchema)
	t.Run("mod_api_key", testModAPIKeySchema)
	t.Run("mod_body_process", testModBodyProcessSchema)
	t.Run("rate_limit_policy", testRateLimitPolicySchema)
	t.Run("ai_route", testAIRouteSchema)
}

// setupCluster 创建一个测试集群并返回名称
func setupCluster(t *testing.T) string {
	clusterName := testutil.UniqueClusterName()
	_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
		"name": clusterName,
		"instance_pool": []interface{}{
			map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080},
		},
		"llm_config": map[string]interface{}{
			"models":        []string{"deepseek-chat"},
			"provider_type": "deepseek",
			"provider":      "deepseek",
		},
	})
	require.NoError(t, err)
	return clusterName
}

// setupAPIKeyWithRoute 创建一个启用 route_rules 的 API-Key
func setupAPIKeyWithRoute(t *testing.T, clusterName string) string {
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "inner-schema-route",
		"quota_plan": map[string]interface{}{
			"unlimited": false,
			"quota":     1000,
			"unit":      "total_token",
		},
		"route_rules": map[string]interface{}{
			"enabled": true,
			"rules": []interface{}{
				map[string]interface{}{
					"name":    "default",
					"cond":    "default_t()",
					"targets": []interface{}{map[string]interface{}{"cluster_name": clusterName, "model": "", "weight": 100}},
					"fallbacks": []interface{}{},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 200, resp.ErrNum, resp.ErrMsg)
	id, err := testutil.GetDataField(resp, "id")
	require.NoError(t, err)
	return id.(string)
}

func testServerDataConfSchema(t *testing.T) {
	clusterName := setupCluster(t)

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, ServerDataConfSchema)
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}

func testGSLBSchema(t *testing.T) {
	clusterName := setupCluster(t)

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/gslb_data/gslb", map[string]string{
		"bfe_cluster": "BFE-AI_product.szyf",
	})
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, GSLBSchema)
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}

func testClusterTableSchema(t *testing.T) {
	clusterName := setupCluster(t)

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/gslb_data/cluster_table")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, ClusterTableSchema)
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}

func testServerCertConfSchema(t *testing.T) {
	certName := testutil.UniqueCertName()
	certPEM, keyPEM, err := testutil.GenerateTestCert(certName)
	require.NoError(t, err)

	_, err = testutil.GetClient().Post("/open-api/v1/certificates", map[string]interface{}{
		"cert_name":         certName,
		"description":       "schema test",
		"is_default":        true,
		"cert_file_content": certPEM,
		"key_file_content":  keyPEM,
	})
	require.NoError(t, err)

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/protocol/server_cert_conf")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, ServerCertConfSchema)
	}

	t.Cleanup(func() {
		_ = testutil.DeleteCertificate(certName)
	})
}

func testModAPIKeySchema(t *testing.T) {
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "inner-schema-mod-api-key",
		"quota_plan": map[string]interface{}{
			"unlimited": false,
			"quota":     1000,
			"unit":      "total_token",
		},
	})
	require.NoError(t, err)
	require.Equal(t, 200, resp.ErrNum, resp.ErrMsg)
	id, err := testutil.GetDataField(resp, "id")
	require.NoError(t, err)
	apiKeyID := id.(string)

	innerResp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key")
	require.NoError(t, err)
	testutil.AssertSuccess(t, innerResp)
	if innerResp.Data != nil && string(innerResp.Data) != "null" {
		testutil.AssertSchema(t, innerResp, ModAPIKeySchema)
	}

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}

func testModBodyProcessSchema(t *testing.T) {
	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-body-process")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, ModBodyProcessSchema)
	}
}

func testRateLimitPolicySchema(t *testing.T) {
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "inner-schema-rate-limit",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{"name": "tpm-1m", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 200, resp.ErrNum, resp.ErrMsg)
	id, err := testutil.GetDataField(resp, "id")
	require.NoError(t, err)
	apiKeyID := id.(string)

	innerResp, err := testutil.GetClient().Get("/inner-api/v1/configs/rate-limit-policy")
	require.NoError(t, err)
	testutil.AssertSuccess(t, innerResp)
	if innerResp.Data != nil && string(innerResp.Data) != "null" {
		testutil.AssertSchema(t, innerResp, RateLimitPolicySchema)
	}

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}

func testAIRouteSchema(t *testing.T) {
	// 设置 global route
	_, err := testutil.GetClient().Put("/open-api/v1/global-route-rules", map[string]interface{}{
		"enabled": true,
		"rules": []interface{}{
			map[string]interface{}{
				"name":      "global-default",
				"cond":      "default_t()",
				"targets":   []interface{}{map[string]interface{}{"cluster_name": "cluster_global", "model": "", "weight": 100}},
				"fallbacks": []interface{}{},
			},
		},
	})
	require.NoError(t, err)

	clusterName := setupCluster(t)
	apiKeyID := setupAPIKeyWithRoute(t, clusterName)

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/ai-route")
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	if resp.Data != nil && string(resp.Data) != "null" {
		testutil.AssertSchema(t, resp, AIRouteSchema)
	}

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
		testutil.DeleteCluster(clusterName)
	})
}
