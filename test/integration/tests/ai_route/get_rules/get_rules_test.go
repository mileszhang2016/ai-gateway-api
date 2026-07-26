package get_rules

import (
	"encoding/json"
	"os"
	"testing"

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

// helper: 构造 PATCH 请求
func patchRouteRules(t *testing.T, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Patch("/open-api/v1/ai-route-rules", body)
	if err != nil {
		t.Fatalf("patch request failed: %v", err)
	}
	return resp
}

// helper: 构造 GET 请求
func getRouteRules(t *testing.T) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Get("/open-api/v1/ai-route-rules")
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	return resp
}

// ============================================================
// AR-2-001：获取已设置的规则（正常参数）
// ============================================================
func TestGetRules_AfterSet(t *testing.T) {
	// 先设置规则
	setBody := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"paths":        []string{"/api"},
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	patchResp := patchRouteRules(t, setBody)
	testutil.AssertSuccess(t, patchResp)

	// 获取规则
	getResp := getRouteRules(t)
	testutil.AssertSuccess(t, getResp)

	var data map[string]interface{}
	if err := json.Unmarshal(getResp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	// 验证 forward_rules 存在且非空
	fr, ok := data["forward_rules"].([]interface{})
	if !ok || len(fr) == 0 {
		t.Fatal("forward_rules should not be empty after setting rules")
	}

	// 验证第一条规则
	fr0 := fr[0].(map[string]interface{})
	if fr0["name"] != "rule1" {
		t.Errorf("expected name=rule1, got %v", fr0["name"])
	}
	if fr0["expression"] != `req_host_in("api.example.com")` {
		t.Errorf("expected expression mismatch, got %v", fr0["expression"])
	}
	if fr0["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("expected cluster_name=BFE-AI_product.szyf, got %v", fr0["cluster_name"])
	}

	// 验证最后一条是 default_t()（自动追加）
	lastRule := fr[len(fr)-1].(map[string]interface{})
	expr, _ := lastRule["expression"].(string)
	if expr != "default_t()" {
		t.Errorf("expected last rule expression=default_t(), got %s", expr)
	}

	// 验证 basic_forward_rules
	br, ok := data["basic_forward_rules"].([]interface{})
	if !ok || len(br) == 0 {
		t.Fatal("basic_forward_rules should not be empty after setting rules")
	}

	br0 := br[0].(map[string]interface{})
	if br0["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("expected basic cluster_name=BFE-AI_product.szyf, got %v", br0["cluster_name"])
	}

	hostNames := br0["host_names"].([]interface{})
	if len(hostNames) != 1 || hostNames[0] != "*.example.com" {
		t.Errorf("expected host_names=[*.example.com], got %v", br0["host_names"])
	}

	paths := br0["paths"].([]interface{})
	if len(paths) != 1 || paths[0] != "/api" {
		t.Errorf("expected paths=[/api], got %v", br0["paths"])
	}
}

// ============================================================
// AR-2-002：获取未设置时的列表（空数据）
// ============================================================
func TestGetRules_EmptyWhenNotSet(t *testing.T) {
	// 先重置为最简规则
	clearBody := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "default",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []interface{}{},
	}
	patchResp := patchRouteRules(t, clearBody)
	testutil.AssertSuccess(t, patchResp)

	// 获取规则
	getResp := getRouteRules(t)
	testutil.AssertSuccess(t, getResp)

	// 验证 basic_forward_rules 为空
	testutil.AssertListFieldLen(t, getResp, "basic_forward_rules", 0)
}

// ============================================================
// AR-2-003：返回数据结构校验（返回数据校验）
// ============================================================
func TestGetRules_DataStructureValidation(t *testing.T) {
	// 先设置完整的规则
	setBody := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"description":  "测试描述",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com", "api.test.com"},
				"paths":        []string{"/api/v1", "/api/v2"},
				"cluster_name": "BFE-AI_product.szyf",
				"description":  "基础路由描述",
			},
		},
	}

	patchResp := patchRouteRules(t, setBody)
	testutil.AssertSuccess(t, patchResp)

	// 获取规则
	getResp := getRouteRules(t)
	testutil.AssertSuccess(t, getResp)

	var data map[string]interface{}
	if err := json.Unmarshal(getResp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	// ======== 顶层键校验 ========
	// forward_rules 必须存在且为数组
	fr, ok := data["forward_rules"].([]interface{})
	if !ok {
		t.Fatal("forward_rules must be an array")
	}
	if len(fr) == 0 {
		t.Fatal("forward_rules should not be empty")
	}

	// basic_forward_rules 必须存在且为数组
	br, ok := data["basic_forward_rules"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules must be an array")
	}
	if len(br) == 0 {
		t.Fatal("basic_forward_rules should not be empty")
	}

	// ======== forward_rules[0] 字段校验 ========
	fr0 := fr[0].(map[string]interface{})

	if name, ok := fr0["name"].(string); !ok || name != "rule1" {
		t.Errorf("forward_rules[0].name: expected 'rule1', got %v", fr0["name"])
	}
	if desc, ok := fr0["description"].(string); !ok || desc != "测试描述" {
		t.Errorf("forward_rules[0].description: expected '测试描述', got %v", fr0["description"])
	}
	if expr, ok := fr0["expression"].(string); !ok || expr != `req_host_in("api.example.com")` {
		t.Errorf("forward_rules[0].expression mismatch, got %v", fr0["expression"])
	}
	if cn, ok := fr0["cluster_name"].(string); !ok || cn != "BFE-AI_product.szyf" {
		t.Errorf("forward_rules[0].cluster_name: expected 'BFE-AI_product.szyf', got %v", fr0["cluster_name"])
	}

	// ======== basic_forward_rules[0] 字段校验 ========
	br0 := br[0].(map[string]interface{})

	if cn, ok := br0["cluster_name"].(string); !ok || cn != "BFE-AI_product.szyf" {
		t.Errorf("basic_forward_rules[0].cluster_name: expected 'BFE-AI_product.szyf', got %v", br0["cluster_name"])
	}
	if desc, ok := br0["description"].(string); !ok || desc != "基础路由描述" {
		t.Errorf("basic_forward_rules[0].description: expected '基础路由描述', got %v", br0["description"])
	}

	// host_names 校验
	hostNames, ok := br0["host_names"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules[0].host_names must be an array")
	}
	if len(hostNames) != 2 {
		t.Errorf("expected host_names length=2, got %d", len(hostNames))
	} else {
		if hostNames[0] != "*.example.com" {
			t.Errorf("host_names[0]: expected '*.example.com', got %v", hostNames[0])
		}
		if hostNames[1] != "api.test.com" {
			t.Errorf("host_names[1]: expected 'api.test.com', got %v", hostNames[1])
		}
	}

	// paths 校验
	paths, ok := br0["paths"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules[0].paths must be an array")
	}
	if len(paths) != 2 {
		t.Errorf("expected paths length=2, got %d", len(paths))
	} else {
		if paths[0] != "/api/v1" {
			t.Errorf("paths[0]: expected '/api/v1', got %v", paths[0])
		}
		if paths[1] != "/api/v2" {
			t.Errorf("paths[1]: expected '/api/v2', got %v", paths[1])
		}
	}

	// ======== 额外验证：forward_rules 末尾是 default_t() ========
	lastRule := fr[len(fr)-1].(map[string]interface{})
	lastExpr, _ := lastRule["expression"].(string)
	if lastExpr != "default_t()" {
		t.Errorf("expected last forward_rules expression=default_t(), got '%s'", lastExpr)
	}
}
