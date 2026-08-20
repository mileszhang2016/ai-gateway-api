package innerapi_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
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

func TestInnerAPI_TlsConf(t *testing.T) {
	// 创建默认证书以确保证书配置非空
	certName := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(certName, true); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("IN-1-001 首次导出 TLS/Server 配置", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataNotEmpty(t, resp)
		testutil.AssertDataFieldNotEmpty(t, resp, "Version")
		testutil.AssertDataFieldNotEmpty(t, resp, "HostTable")
		testutil.AssertDataFieldNotEmpty(t, resp, "RouteTable")
		testutil.AssertDataFieldNotEmpty(t, resp, "ClusterConf")
	})

	t.Run("IN-1-002 导出 ClusterConf 含多 Key AIConf", func(t *testing.T) {
		clusterName := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterName,
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "10.0.0.1",
					"weight": 100,
					"port":   8080,
				},
			},
			"llm_config": map[string]interface{}{
				"models": []string{"deepseek-chat"},
				"keys": []interface{}{
					map[string]interface{}{
						"name":   "primary",
						"key":    "sk-aaaaaaaaaaaa",
						"weight": 70,
					},
					map[string]interface{}{
						"name":   "secondary",
						"key":    "sk-bbbbbbbbbbbb",
						"weight": 30,
					},
				},
				"key_policy": map[string]interface{}{
					"strategy":              "weighted_random",
					"max_retries":           3,
					"retry_backoff_initial": 100,
					"retry_backoff_max":     5000,
				},
				"provider_type": "deepseek",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterName)

		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		clusterConf, ok := data["ClusterConf"].(map[string]interface{})
		if !assert.True(t, ok, "ClusterConf should be an object") {
			return
		}
		config, ok := clusterConf["Config"].(map[string]interface{})
		if !assert.True(t, ok, "ClusterConf.Config should be an object") {
			return
		}
		cluster, ok := config[clusterName].(map[string]interface{})
		if !assert.True(t, ok, "target cluster should exist in ClusterConf.Config") {
			return
		}
		aiconf, ok := cluster["AIConf"].(map[string]interface{})
		if !assert.True(t, ok, "AIConf should be an object") {
			return
		}
		keys, ok := aiconf["Keys"].([]interface{})
		if !assert.True(t, ok, "AIConf.Keys should be an array") {
			return
		}
		assert.Len(t, keys, 2)
		key0, _ := keys[0].(map[string]interface{})
		assert.Equal(t, "primary", key0["Name"])
		assert.Equal(t, "sk-aaaaaaaaaaaa", key0["Key"])
		assert.Equal(t, float64(70), key0["Weight"])

		policy, ok := aiconf["KeyPolicy"].(map[string]interface{})
		if !assert.True(t, ok, "AIConf.KeyPolicy should be an object") {
			return
		}
		assert.Equal(t, "weighted_random", policy["Strategy"])
		assert.Equal(t, float64(3), policy["MaxRetries"])
		assert.Equal(t, float64(100), policy["RetryBackoffInitial"])
		assert.Equal(t, float64(5000), policy["RetryBackoffMax"])
	})

	t.Run("IN-1-003 导出 ClusterConf 含模型定价表", func(t *testing.T) {
		yamlContent := []byte(`version: v1.0
default_currency: RMB
models:
  - provider: openai
    model: gpt-4o
    base_model: gpt-4o
    mode: chat
    prices:
      input_cost_per_token: 0.0001
  - provider: openai
    model: gpt-4o-mini
    base_model: gpt-4o-mini
    mode: chat
    prices:
      input_cost_per_token: 0.00001
`)
		if err := testutil.ImportModelPrices(yamlContent, "replace"); err != nil {
			t.Fatalf("import model prices failed: %v", err)
		}

		clusterName := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterName,
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "10.0.0.1",
					"weight": 100,
					"port":   8080,
				},
			},
			"llm_config": map[string]interface{}{
				"models":        []string{"gpt-4o"},
				"provider_type": "openai",
				"provider":      "openai",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterName)

		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		clusterConf, ok := data["ClusterConf"].(map[string]interface{})
		if !assert.True(t, ok, "ClusterConf should be an object") {
			return
		}
		config, ok := clusterConf["Config"].(map[string]interface{})
		if !assert.True(t, ok, "ClusterConf.Config should be an object") {
			return
		}
		cluster, ok := config[clusterName].(map[string]interface{})
		if !assert.True(t, ok, "target cluster should exist in ClusterConf.Config") {
			return
		}
		aiconf, ok := cluster["AIConf"].(map[string]interface{})
		if !assert.True(t, ok, "AIConf should be an object") {
			return
		}
		modelTable, ok := aiconf["ModelTable"].(map[string]interface{})
		if !assert.True(t, ok, "AIConf.ModelTable should be an object") {
			return
		}
		assert.Equal(t, "RMB", modelTable["Currency"])
		models, ok := modelTable["Models"].([]interface{})
		if !assert.True(t, ok, "ModelTable.Models should be an array") {
			return
		}
		assert.Len(t, models, 2)

		modelNames := make([]string, 0, len(models))
		for _, m := range models {
			mm, ok := m.(map[string]interface{})
			if assert.True(t, ok, "model entry should be an object") {
				modelNames = append(modelNames, mm["Model"].(string))
				assert.Equal(t, "openai", mm["Provider"])
			}
		}
		assert.Contains(t, modelNames, "gpt-4o")
		assert.Contains(t, modelNames, "gpt-4o-mini")
	})

	t.Run("IN-TLS-1-004 AIConf 包含 MatchPrefix / StripPrefix", func(t *testing.T) {
		clusterName := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterName,
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "10.0.0.1",
					"weight": 100,
					"port":   8080,
				},
			},
			"llm_config": map[string]interface{}{
				"models":        []string{"openrouter/anthropic/claude-sonnet-4"},
				"match_prefix":  "openrouter/",
				"strip_prefix":  true,
				"provider_type": "openrouter",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterName)

		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		aiconf := extractAIConf(t, resp, clusterName)
		assert.Equal(t, "openrouter/", aiconf["MatchPrefix"])
		assert.Equal(t, true, aiconf["StripPrefix"])
	})

	t.Run("IN-TLS-1-005 未配置前缀时 AIConf 为默认值", func(t *testing.T) {
		clusterName := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterName,
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "10.0.0.1",
					"weight": 100,
					"port":   8080,
				},
			},
			"llm_config": map[string]interface{}{
				"models":        []string{"deepseek-chat"},
				"provider_type": "deepseek",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterName)

		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		aiconf := extractAIConf(t, resp, clusterName)
		v, ok := aiconf["MatchPrefix"]
		if ok && v != nil {
			assert.Equal(t, "", v, "MatchPrefix should be empty if present")
		}
		assert.Equal(t, false, aiconf["StripPrefix"])
	})

	t.Run("IN-TLS-1-006 仅 match_prefix、strip_prefix=false 时的导出", func(t *testing.T) {
		clusterName := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterName,
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "10.0.0.1",
					"weight": 100,
					"port":   8080,
				},
			},
			"llm_config": map[string]interface{}{
				"models":        []string{"openrouter/anthropic/claude-sonnet-4"},
				"match_prefix":  "openrouter/",
				"strip_prefix":  false,
				"provider_type": "openrouter",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterName)

		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		aiconf := extractAIConf(t, resp, clusterName)
		assert.Equal(t, "openrouter/", aiconf["MatchPrefix"])
		assert.Equal(t, false, aiconf["StripPrefix"])
	})

	t.Cleanup(func() {
		testutil.DeleteCertificate(certName)
	})
}

func extractAIConf(t *testing.T, resp *testutil.APIResponse, clusterName string) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	clusterConf, ok := data["ClusterConf"].(map[string]interface{})
	if !assert.True(t, ok, "ClusterConf should be an object") {
		return nil
	}
	config, ok := clusterConf["Config"].(map[string]interface{})
	if !assert.True(t, ok, "ClusterConf.Config should be an object") {
		return nil
	}
	cluster, ok := config[clusterName].(map[string]interface{})
	if !assert.True(t, ok, "target cluster should exist in ClusterConf.Config") {
		return nil
	}
	aiconf, ok := cluster["AIConf"].(map[string]interface{})
	if !assert.True(t, ok, "AIConf should be an object") {
		return nil
	}
	return aiconf
}
