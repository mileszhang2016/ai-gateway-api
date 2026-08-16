package clusters_test

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

func TestClusters_Update(t *testing.T) {
	clusterName := testutil.UniqueClusterName()
	if _, err := testutil.CreateCluster(clusterName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("CL-4-001 更新实例池", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-2",
					"addr":   "10.0.0.2",
					"weight": 100,
					"port":   9090,
				},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", clusterName)
	})

	t.Run("CL-4-002 更新 llm_config", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models":        []string{"qwen-turbo"},
				"provider_type": "qwen",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", clusterName)
	})

	t.Run("CL-4-003 更新后查询一致性", func(t *testing.T) {
		_, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"description": "更新后的集群描述",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/clusters/" + clusterName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "description", "更新后的集群描述")
	})

	t.Run("CL-4-004 更新不存在的集群", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/non_existent_cluster", map[string]interface{}{
			"description": "x",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("CL-4-005 更新非法 instance_pool（非法 addr）", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"instance_pool": []interface{}{
				map[string]interface{}{
					"name":   "backend-1",
					"addr":   "-bad",
					"weight": 100,
					"port":   8080,
				},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("CL-4-006 更新 keys（全量替换）", func(t *testing.T) {
		clusterUpdateKeys := testutil.UniqueClusterName()
		if _, err := testutil.CreateCluster(clusterUpdateKeys); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterUpdateKeys)

		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterUpdateKeys, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models": []string{"deepseek-chat"},
				"keys": []interface{}{
					map[string]interface{}{
						"name":   "new-primary",
						"key":    "sk-new-primary",
						"weight": 60,
					},
					map[string]interface{}{
						"name":   "new-secondary",
						"key":    "sk-new-secondary",
						"weight": 40,
					},
				},
				"key_policy": map[string]interface{}{
					"strategy":    "weighted_random",
					"max_retries": 5,
				},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		resp, err = testutil.GetClient().Get("/open-api/v1/clusters/" + clusterUpdateKeys)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", clusterUpdateKeys)
	})

	t.Run("CL-4-007 更新 key_policy", func(t *testing.T) {
		clusterUpdatePolicy := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterUpdatePolicy,
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
						"key":    "sk-primary",
						"weight": 60,
					},
					map[string]interface{}{
						"name":   "secondary",
						"key":    "sk-secondary",
						"weight": 40,
					},
				},
				"provider_type": "deepseek",
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteCluster(clusterUpdatePolicy)

		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterUpdatePolicy, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models": []string{"deepseek-chat"},
				"key_policy": map[string]interface{}{
					"strategy":              "weighted_random",
					"retry_backoff_initial": 200,
					"retry_backoff_max":     2000,
				},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		resp, err = testutil.GetClient().Get("/open-api/v1/clusters/" + clusterUpdatePolicy)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", clusterUpdatePolicy)
	})

	t.Run("CL-4-008 更新 match_prefix / strip_prefix", func(t *testing.T) {
		clusterUpdatePrefix := testutil.UniqueClusterName()
		_, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
			"name": clusterUpdatePrefix,
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
		defer testutil.DeleteCluster(clusterUpdatePrefix)

		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterUpdatePrefix, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models":        []string{"deepseek-chat"},
				"match_prefix":  "deepseek/",
				"strip_prefix":  false,
				"provider_type": "deepseek",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		resp, err = testutil.GetClient().Get("/open-api/v1/clusters/" + clusterUpdatePrefix)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		llm, _ := data["llm_config"].(map[string]interface{})
		assert.Equal(t, "deepseek/", llm["match_prefix"])
		assert.Equal(t, false, llm["strip_prefix"])

		resp, err = testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		if err != nil {
			t.Fatalf("inner api request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var innerData map[string]interface{}
		if err := json.Unmarshal(resp.Data, &innerData); err != nil {
			t.Fatalf("unmarshal inner data failed: %v", err)
		}
		clusterConf, _ := innerData["ClusterConf"].(map[string]interface{})
		config, _ := clusterConf["Config"].(map[string]interface{})
		cluster, _ := config[clusterUpdatePrefix].(map[string]interface{})
		aiconf, _ := cluster["AIConf"].(map[string]interface{})
		assert.Equal(t, "deepseek/", aiconf["MatchPrefix"])
		assert.Equal(t, false, aiconf["StripPrefix"])
	})

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}

func TestClusters_Update_PreventDeleteReferencedModel(t *testing.T) {
	clusterName := testutil.UniqueClusterName()
	resp, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
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
			"models":        []string{"test-model"},
			"provider_type": "test-provider",
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// Create API-Key with route rule referencing the cluster model.
	apiKeyResp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "model-ref-key",
		"route_rules": map[string]interface{}{
			"enabled": true,
			"rules": []interface{}{
				map[string]interface{}{
					"name": "rule-ref-model",
					"cond": "default_t()",
					"targets": []interface{}{
						map[string]interface{}{
							"cluster_name": clusterName,
							"model":       "test-model",
							"weight":      100,
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create api-key failed: %v", err)
	}
	testutil.AssertSuccess(t, apiKeyResp)
	apiKeyID, _ := testutil.GetDataField(apiKeyResp, "id")

	t.Run("CL-4-009 删除被路由引用的模型应被拦截", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models":        []string{"other-model"},
				"provider_type": "test-provider",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 500)
		assert.Contains(t, resp.ErrMsg, "Rule rule-ref-model Refer To Model test-model In Cluster")
	})

	t.Run("CL-4-010 清理路由引用后可删除模型", func(t *testing.T) {
		// Remove the route rule reference to test-model.
		resp, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID.(string), map[string]interface{}{
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("update api-key failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		resp, err = testutil.GetClient().Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
			"llm_config": map[string]interface{}{
				"models":        []string{"other-model"},
				"provider_type": "test-provider",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
	})

	t.Cleanup(func() {
		testutil.GetClient().Delete("/open-api/v1/api-keys/" + apiKeyID.(string))
		testutil.DeleteCluster(clusterName)
	})
}
