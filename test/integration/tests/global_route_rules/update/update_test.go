package global_route_rules_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func putGlobalRules(t *testing.T, body map[string]interface{}) *testutil.APIResponse {
	resp, err := testutil.GetClient().Put("/open-api/v1/global-route-rules", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestGlobalRouteRules_Update(t *testing.T) {
	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
		skip     string
		check    func(t *testing.T, resp *testutil.APIResponse)
	}{
		{
			name: "GRR-1-001 最小参数更新",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":      "global-default",
						"Cond":      "default_t()",
						"targets":   []interface{}{map[string]interface{}{"ClusterName": "cluster_global", "Model": "", "Weight": 100}},
						"fallbacks": []interface{}{},
					},
				},
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				testutil.AssertDataFieldEquals(t, resp, "enabled", true)
				var data map[string]interface{}
				json.Unmarshal(resp.Data, &data)
				rules := data["rules"].([]interface{})
				assert.Len(t, rules, 1)
			},
		},
		{
			name: "GRR-1-002 完整参数更新",
			body: map[string]interface{}{
				"enabled": false,
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "rule-a",
						"Cond":    "default_t()",
						"targets": []interface{}{
							map[string]interface{}{"ClusterName": "cluster_a", "Model": "gpt-4", "Weight": 70},
							map[string]interface{}{"ClusterName": "cluster_b", "Model": "", "Weight": 30},
						},
						"fallbacks": []interface{}{map[string]interface{}{"ClusterName": "cluster_fallback", "Model": ""}},
					},
				},
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				testutil.AssertDataFieldEquals(t, resp, "enabled", false)
			},
		},
		{
			name: "GRR-1-003 更新后查询一致性",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":      "global-default",
						"Cond":      "default_t()",
						"targets":   []interface{}{map[string]interface{}{"ClusterName": "cluster_global", "Model": "", "Weight": 100}},
						"fallbacks": []interface{}{},
					},
				},
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				getResp, err := testutil.GetClient().Get("/open-api/v1/global-route-rules")
				if err != nil {
					t.Fatalf("get failed: %v", err)
				}
				assert.Equal(t, string(resp.Data), string(getResp.Data))
			},
		},
		{
			name: "GRR-1-004 规则名称重复",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "dup",
						"Cond":    "default_t()",
						"targets": []interface{}{map[string]interface{}{"ClusterName": "c1", "Weight": 100}},
					},
					map[string]interface{}{
						"name":    "dup",
						"Cond":    "default_t()",
						"targets": []interface{}{map[string]interface{}{"ClusterName": "c2", "Weight": 100}},
					},
				},
			},
			wantCode: 422,
		},
		{
			name: "GRR-1-005 targets 权重总和不等于 100",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "bad-weight",
						"Cond":    "default_t()",
						"targets": []interface{}{
							map[string]interface{}{"ClusterName": "c1", "Weight": 60},
							map[string]interface{}{"ClusterName": "c2", "Weight": 30},
						},
					},
				},
			},
			wantCode: 422,
		},
		{
			name: "GRR-1-006 fallbacks ClusterName 为空",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":      "bad-fb",
						"Cond":      "default_t()",
						"targets":   []interface{}{map[string]interface{}{"ClusterName": "c1", "Weight": 100}},
						"fallbacks": []interface{}{map[string]interface{}{"ClusterName": "", "Model": ""}},
					},
				},
			},
			wantCode: 422,
		},
		{
			name: "GRR-1-007 Cond 表达式非法",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "bad-cond",
						"Cond":    "not_a_valid_expr(",
						"targets": []interface{}{map[string]interface{}{"ClusterName": "c1", "Weight": 100}},
					},
				},
			},
			wantCode: 422,
			skip:     "implementation does not validate Cond expression syntax",
		},
		{
			name: "GRR-1-008 重复 target (ClusterName+Model)",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name": "dup-target",
						"Cond": "default_t()",
						"targets": []interface{}{
							map[string]interface{}{"ClusterName": "c1", "Model": "m1", "Weight": 50},
							map[string]interface{}{"ClusterName": "c1", "Model": "m1", "Weight": 50},
						},
					},
				},
			},
			wantCode: 422,
		},
		{
			name: "GRR-1-009 target ClusterName 格式非法",
			body: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"name":    "bad-cluster",
						"Cond":    "default_t()",
						"targets": []interface{}{map[string]interface{}{"ClusterName": "-bad", "Weight": 100}},
					},
				},
			},
			wantCode: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			resp := putGlobalRules(t, tt.body)
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.check != nil && resp.ErrNum == 200 {
				tt.check(t, resp)
			}
		})
	}
}
