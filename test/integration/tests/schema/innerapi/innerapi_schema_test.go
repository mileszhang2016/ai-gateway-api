package innerapi

import (
	"encoding/json"
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
		"llm_config": map[string]interface{}{
			"models":   []string{"deepseek-chat"},
			"provider": "deepseek",
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
		assertModAPIKeyFieldDetails(t, innerResp.Data)
	}

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}

// assertModAPIKeyFieldDetails 校验 /configs/mod-api-key 响应中 tokens 和 QuotaPlans 的字段细节：
// token 的 enabled 为 bool，不包含 status/update_time；QuotaPlan 不包含 CreateTime/ResetMode；Tags 包含 TagLevel。
func assertModAPIKeyFieldDetails(t *testing.T, data []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal mod-api-key data: %v", err)
	}

	tokens, ok := payload["tokens"].(map[string]interface{})
	if !ok {
		t.Fatal("tokens is not an object")
	}
	for product, productTokens := range tokens {
		productTokenMap, ok := productTokens.(map[string]interface{})
		if !ok {
			t.Fatalf("tokens.%s is not an object", product)
		}
		for key, token := range productTokenMap {
			tokenMap, ok := token.(map[string]interface{})
			if !ok {
				t.Fatalf("tokens.%s.%s is not an object", product, key)
			}
			if _, ok := tokenMap["status"]; ok {
				t.Errorf("tokens.%s.%s should not contain status", product, key)
			}
			if _, ok := tokenMap["update_time"]; ok {
				t.Errorf("tokens.%s.%s should not contain update_time", product, key)
			}
			if enabled, ok := tokenMap["enabled"].(bool); !ok {
				t.Errorf("tokens.%s.%s.enabled should be bool", product, key)
			} else if !enabled {
				t.Errorf("tokens.%s.%s.enabled should be true", product, key)
			}
			if keyID, ok := tokenMap["key_id"].(string); !ok || keyID == "" {
				t.Errorf("tokens.%s.%s.key_id should be non-empty string", product, key)
			}
			if expiredTime, ok := tokenMap["expired_time"].(float64); !ok {
				t.Errorf("tokens.%s.%s.expired_time should be number", product, key)
			} else if expiredTime != -1 {
				// ok, can be -1 or a timestamp
			}
			if _, ok := tokenMap["unlimited_quota"].(bool); !ok {
				t.Errorf("tokens.%s.%s.unlimited_quota should be bool", product, key)
			}
			if models, ok := tokenMap["allow_models"].(string); ok && models == "*" {
				t.Errorf("tokens.%s.%s.allow_models should not be '*', got %q", product, key, models)
			}
			if quotaPlans, ok := tokenMap["quota_plans"].([]interface{}); ok {
				for i, qp := range quotaPlans {
					if _, ok := qp.(string); !ok {
						t.Errorf("tokens.%s.%s.quota_plans[%d] should be string", product, key, i)
					}
				}
			}
			if tags, ok := tokenMap["Tags"].([]interface{}); ok {
				for i, tag := range tags {
					tagMap, ok := tag.(map[string]interface{})
					if !ok {
						t.Fatalf("tokens.%s.%s.Tags[%d] is not an object", product, key, i)
					}
					if _, ok := tagMap["TagName"].(string); !ok {
						t.Errorf("tokens.%s.%s.Tags[%d].TagName should be string", product, key, i)
					}
					if _, ok := tagMap["TagValue"].(string); !ok {
						t.Errorf("tokens.%s.%s.Tags[%d].TagValue should be string", product, key, i)
					}
					if _, ok := tagMap["TagLevel"].(float64); !ok {
						t.Errorf("tokens.%s.%s.Tags[%d].TagLevel should be int", product, key, i)
					}
				}
			}
		}
	}

	quotaPlans, ok := payload["QuotaPlans"].(map[string]interface{})
	if !ok {
		t.Fatal("QuotaPlans is not an object")
	}
	for product, plans := range quotaPlans {
		planList, ok := plans.([]interface{})
		if !ok {
			t.Fatalf("QuotaPlans.%s is not an array", product)
		}
		for i, plan := range planList {
			planMap, ok := plan.(map[string]interface{})
			if !ok {
				t.Fatalf("QuotaPlans.%s[%d] is not an object", product, i)
			}
			if _, ok := planMap["CreateTime"]; ok {
				t.Errorf("QuotaPlans.%s[%d] should not contain CreateTime", product, i)
			}
			if _, ok := planMap["ResetMode"]; ok {
				t.Errorf("QuotaPlans.%s[%d] should not contain ResetMode", product, i)
			}
			if id, ok := planMap["Id"].(string); !ok || id == "" {
				t.Errorf("QuotaPlans.%s[%d].Id should be non-empty string", product, i)
			}
			if _, ok := planMap["Unlimited"].(bool); !ok {
				t.Errorf("QuotaPlans.%s[%d].Unlimited should be bool", product, i)
			}
			if _, ok := planMap["PassNoQuota"].(bool); !ok {
				t.Errorf("QuotaPlans.%s[%d].PassNoQuota should be bool", product, i)
			}
			if redisKey, ok := planMap["RedisKey"].(string); !ok || redisKey == "" {
				t.Errorf("QuotaPlans.%s[%d].RedisKey should be non-empty string", product, i)
			}
			if _, ok := planMap["ExpiredTime"].(float64); !ok {
				t.Errorf("QuotaPlans.%s[%d].ExpiredTime should be number", product, i)
			}
			if _, ok := planMap["Quota"].(float64); !ok {
				t.Errorf("QuotaPlans.%s[%d].Quota should be number", product, i)
			}
		}
	}
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
