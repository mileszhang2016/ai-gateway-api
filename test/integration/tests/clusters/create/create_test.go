package clusters_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yf-networks/ai-gateway-api/integration/testutil"
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

func minClusterBody(name string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
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
			"key":           "sk-xxx",
			"provider_type": "deepseek",
		},
	}
}

func assertNoInternalFields(t *testing.T, data map[string]interface{}) {
	assert.NotContains(t, data, "ready")
	assert.NotContains(t, data, "sub_clusters")
	assert.NotContains(t, data, "scheduler")
}

func TestClusters_Create(t *testing.T) {
	clusterMin := testutil.UniqueClusterName()
	clusterFull := testutil.UniqueClusterName()
	clusterDup := testutil.UniqueClusterName()

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
		skip     string
		check    func(t *testing.T, resp *testutil.APIResponse)
	}{
		{
			name:     "CL-1-001 最小参数创建集群",
			body:     minClusterBody(clusterMin),
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				var data map[string]interface{}
				json.Unmarshal(resp.Data, &data)
				assertNoInternalFields(t, data)
				testutil.AssertDataFieldEquals(t, resp, "name", clusterMin)

				sticky, ok := data["sticky_sessions"].(map[string]interface{})
				if assert.True(t, ok, "sticky_sessions should be an object") {
					assert.Equal(t, false, sticky["enabled"])
					assert.Equal(t, "CLIENT_IP_ONLY", sticky["hash_strategy"])
					assert.Equal(t, "", sticky["hash_header"])
				}
			},
		},
		{
			name: "CL-1-002 完整参数创建集群",
			body: map[string]interface{}{
				"name":        clusterFull,
				"description": "完整集群",
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 50,
						"port":   8080,
					},
					map[string]interface{}{
						"name":   "backend-2",
						"addr":   "10.0.0.2",
						"weight": 50,
						"port":   8080,
					},
				},
				"llm_config": map[string]interface{}{
					"models":        []string{"deepseek-chat", "deepseek-coder"},
					"key":           "sk-xxx",
					"provider_type": "deepseek",
				},
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				var data map[string]interface{}
				json.Unmarshal(resp.Data, &data)
				assertNoInternalFields(t, data)
				testutil.AssertDataFieldEquals(t, resp, "name", clusterFull)
				insts, _ := data["instance_pool"].([]interface{})
				assert.Len(t, insts, 2)
			},
		},
		{
			name: "CL-1-003 缺少 llm_config",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 100,
						"port":   8080,
					},
				},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-004 缺少 instance_pool",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"llm_config": map[string]interface{}{
					"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek",
				},
			},
			wantCode: 422,
		},
		{
			name:     "CL-1-005 重复集群名",
			body:     minClusterBody(clusterDup),
			wantCode: 555,
		},
		{
			name: "CL-1-006 instance_pool 为空数组",
			body: map[string]interface{}{
				"name":          testutil.UniqueClusterName(),
				"instance_pool": []interface{}{},
				"llm_config":    map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-007 实例 port 非法",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 100,
						"port":   0,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-008 非法 name",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   strings.Repeat("a", 129),
						"addr":   "10.0.0.1",
						"weight": 100,
						"port":   8080,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-009 非法 addr",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "-bad",
						"weight": 100,
						"port":   8080,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-010 weight 超过 100",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 101,
						"port":   8080,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-011 重复实例 (name+addr)",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 50,
						"port":   8080,
					},
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 50,
						"port":   8081,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
		{
			name: "CL-1-012 llm_config 模型重复",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"name":   "backend-1",
						"addr":   "10.0.0.1",
						"weight": 100,
						"port":   8080,
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m", "m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
		},
	}

	// 预先创建重复集群
	if _, err := testutil.GetClient().Post("/open-api/v1/clusters", minClusterBody(clusterDup)); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			resp, err := testutil.GetClient().Post("/open-api/v1/clusters", tt.body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.check != nil && resp.ErrNum == 200 {
				tt.check(t, resp)
			}
		})
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterMin)
		testutil.DeleteCluster(clusterFull)
		testutil.DeleteCluster(clusterDup)
	})
}
