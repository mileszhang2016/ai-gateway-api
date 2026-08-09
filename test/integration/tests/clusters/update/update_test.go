package clusters_test

import (
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}
