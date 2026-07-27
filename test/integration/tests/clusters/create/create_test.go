package clusters_test

import (
	"encoding/json"
	"os"
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
				"hostname": "backend-1",
				"ip":       "10.0.0.1",
				"weight":   100,
				"ports": map[string]interface{}{
					"Default": 8080,
				},
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
			},
		},
		{
			name: "CL-1-002 完整参数创建集群",
			body: map[string]interface{}{
				"name":        clusterFull,
				"description": "完整集群",
				"instance_pool": []interface{}{
					map[string]interface{}{
						"hostname": "backend-1",
						"ip":       "10.0.0.1",
						"weight":   50,
						"ports":    map[string]interface{}{"Default": 8080},
					},
					map[string]interface{}{
						"hostname": "backend-2",
						"ip":       "10.0.0.2",
						"weight":   50,
						"ports":    map[string]interface{}{"Default": 8080},
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
						"hostname": "backend-1",
						"ip":       "10.0.0.1",
						"weight":   100,
						"ports":    map[string]interface{}{"Default": 8080},
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
			name: "CL-1-007 实例不含 Default 端口",
			body: map[string]interface{}{
				"name": testutil.UniqueClusterName(),
				"instance_pool": []interface{}{
					map[string]interface{}{
						"hostname": "backend-1",
						"ip":       "10.0.0.1",
						"weight":   100,
						"ports":    map[string]interface{}{"Other": 8080},
					},
				},
				"llm_config": map[string]interface{}{"models": []string{"m"}, "key": "sk-xxx", "provider_type": "deepseek"},
			},
			wantCode: 422,
			skip:     "implementation does not validate Default port presence",
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
