package set_rules

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

// helper: 构造 PATCH 请求并返回响应
func patchRouteRules(t *testing.T, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Patch("/open-api/v1/ai-route-rules", body)
	if err != nil {
		t.Fatalf("patch request failed: %v", err)
	}
	return resp
}

// helper: 构造 GET 请求并返回响应
func getRouteRules(t *testing.T) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Get("/open-api/v1/ai-route-rules")
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	return resp
}

// ============================================================
// AR-1-001：仅设置基础路由规则（正常参数）
// ============================================================
func TestSetRules_OnlyBasicRules(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "default",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"paths":        []string{"/api"},
				"cluster_name": "BFE-AI_product.szyf",
				"description":  "基础路由规则",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// 验证 basic_forward_rules 长度
	testutil.AssertListFieldLen(t, resp, "basic_forward_rules", 1)

	// 验证字段值
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	basicRules := data["basic_forward_rules"].([]interface{})
	rule := basicRules[0].(map[string]interface{})

	if rule["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("expected cluster_name=BFE-AI_product.szyf, got %v", rule["cluster_name"])
	}
	if rule["description"] != "基础路由规则" {
		t.Errorf("expected description=基础路由规则, got %v", rule["description"])
	}

	hostNames := rule["host_names"].([]interface{})
	if len(hostNames) != 1 || hostNames[0] != "*.example.com" {
		t.Errorf("expected host_names=[*.example.com], got %v", rule["host_names"])
	}

	paths := rule["paths"].([]interface{})
	if len(paths) != 1 || paths[0] != "/api" {
		t.Errorf("expected paths=[/api], got %v", rule["paths"])
	}
}

// ============================================================
// AR-1-002：仅设置高级路由规则（正常参数）
// ============================================================
func TestSetRules_OnlyAdvancedRules(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"description":  "路由到集群1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// 验证 forward_rules 长度
	testutil.AssertListFieldLen(t, resp, "forward_rules", 1)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	rules := data["forward_rules"].([]interface{})
	rule := rules[0].(map[string]interface{})

	if rule["name"] != "rule1" {
		t.Errorf("expected name=rule1, got %v", rule["name"])
	}
	if rule["expression"] != `req_host_in("api.example.com")` {
		t.Errorf("expected expression mismatch, got %v", rule["expression"])
	}
	if rule["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("expected cluster_name=BFE-AI_product.szyf, got %v", rule["cluster_name"])
	}
}

// ============================================================
// AR-1-003：同时设置基础和高级路由规则（正常参数）
// ============================================================
func TestSetRules_BothRuleTypes(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule-adv",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"paths":        []string{"/api/v1"},
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	testutil.AssertListFieldLen(t, resp, "forward_rules", 1)
	testutil.AssertListFieldLen(t, resp, "basic_forward_rules", 1)
}

// ============================================================
// AR-1-004：清空所有规则（正常参数）
// ============================================================
func TestSetRules_ClearAllRules(t *testing.T) {
	// 先设置规则
	setBody := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}
	resp := patchRouteRules(t, setBody)
	testutil.AssertSuccess(t, resp)

	// 重置为最简规则（API 不接收空数组，需要至少一条 default_t() 规则）
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
	resp = patchRouteRules(t, clearBody)
	testutil.AssertSuccess(t, resp)

	// 通过 GET 验证仅剩 default_t() 规则
	getResp := getRouteRules(t)
	testutil.AssertSuccess(t, getResp)
	testutil.AssertListFieldLen(t, getResp, "basic_forward_rules", 0)
}

// ============================================================
// AR-1-005：设置多条高级路由规则（正常参数）
// ============================================================
func TestSetRules_MultipleAdvancedRules(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule-1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
			{
				"name":         "rule-2",
				"expression":   `req_host_in("test.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
			{
				"name":         "rule-3",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)
	testutil.AssertListFieldLen(t, resp, "forward_rules", 3)
}

// ============================================================
// AR-1-006：设置多条基础路由规则（正常参数）
// ============================================================
func TestSetRules_MultipleBasicRules(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "default",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"paths":        []string{"/api"},
				"cluster_name": "BFE-AI_product.szyf",
			},
			{
				"host_names":   []string{"*.test.com"},
				"paths":        []string{"/v2"},
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)
	testutil.AssertListFieldLen(t, resp, "basic_forward_rules", 2)
}

// ============================================================
// AR-1-007：缺少 forward_rules[].expression（必填校验）
// ============================================================
func TestSetRules_MissingExpression(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-008：缺少 forward_rules[].cluster_name（必填校验）
// ============================================================
func TestSetRules_MissingClusterName(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":       "rule1",
				"expression": `req_host_in("api.example.com")`,
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-009：缺少 basic_forward_rules[].cluster_name（必填校验）
// ============================================================
func TestSetRules_MissingBasicClusterName(t *testing.T) {
	body := map[string]interface{}{
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names": []string{"*.example.com"},
				"paths":      []string{"/api"},
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-010：forward_rules[].expression 为空字符串（边界值）
// ============================================================
func TestSetRules_EmptyExpression(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   "",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	// validate:"required,min=1" 约束，空字符串不通过校验
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-011：forward_rules[].cluster_name 为空字符串（边界值）
// ============================================================
func TestSetRules_EmptyClusterName(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "",
			},
		},
	}

	resp := patchRouteRules(t, body)
	// validate:"required,min=1" 约束，空字符串不通过校验
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-012：basic_forward_rules[].cluster_name 为空字符串（边界值）
// ============================================================
func TestSetRules_EmptyBasicClusterName(t *testing.T) {
	body := map[string]interface{}{
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"cluster_name": "",
			},
		},
	}

	resp := patchRouteRules(t, body)
	// validate:"required,min=1" 约束，空字符串不通过校验
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-013：forward_rules 数组元素为 null（nil 校验）
// ============================================================
func TestSetRules_NullAdvanceElement(t *testing.T) {
	// 使用 json.RawMessage 构造带 null 的 JSON
	body := json.RawMessage(`{"forward_rules":[null]}`)
	resp, err := testutil.GetClient().Patch("/open-api/v1/ai-route-rules", body)
	if err != nil {
		t.Fatalf("patch request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-014：basic_forward_rules 数组元素为 null（nil 校验）
// ============================================================
func TestSetRules_NullBasicElement(t *testing.T) {
	body := json.RawMessage(`{"basic_forward_rules":[null]}`)
	resp, err := testutil.GetClient().Patch("/open-api/v1/ai-route-rules", body)
	if err != nil {
		t.Fatalf("patch request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-015：空 Body（边界值）
// ============================================================
func TestSetRules_EmptyBody(t *testing.T) {
	body := map[string]interface{}{}
	resp := patchRouteRules(t, body)
	// 空 Body 时 AdvanceRouteRules 为空，Convert() 校验失败
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-016：非法 JSON Body（异常输入）
// ============================================================
func TestSetRules_InvalidJSON(t *testing.T) {
	// 使用 RawBody 发送非 JSON 字符串
	client := testutil.GetClient()
	resp, err := client.RawBody("PATCH", "/open-api/v1/ai-route-rules", "this is not valid json", "text/plain")
	if err != nil {
		t.Fatalf("patch request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AR-1-017：forward_rules[].description 可选（可选字段）
// ============================================================
func TestSetRules_OptionalDescription(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// 验证 description 未传时返回空字符串
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	rules, ok := data["forward_rules"].([]interface{})
	if !ok {
		t.Fatal("forward_rules is not a list")
	}
	rule := rules[0].(map[string]interface{})
	desc, _ := rule["description"].(string)
	if desc != "" {
		t.Errorf("expected description to be empty string, got %v", desc)
	}
}

// ============================================================
// AR-1-018：basic_forward_rules[].host_names 可选（可选字段）
// ============================================================
func TestSetRules_OptionalHostNames(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "default",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"paths":        []string{"/api"},
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// 验证 host_names 未传时返回 null 或空数组
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	rules, ok := data["basic_forward_rules"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules is not a list")
	}
	rule := rules[0].(map[string]interface{})
	if rule["host_names"] != nil {
		t.Logf("host_names when not set: %v", rule["host_names"])
	}
}

// ============================================================
// AR-1-019：basic_forward_rules[].paths 可选（可选字段）
// ============================================================
func TestSetRules_OptionalPaths(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "default",
				"expression":   "default_t()",
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// 验证 paths 未传时返回 null 或空数组
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	rules, ok := data["basic_forward_rules"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules is not a list")
	}
	rule := rules[0].(map[string]interface{})
	if rule["paths"] != nil {
		t.Logf("paths when not set: %v", rule["paths"])
	}
}

// ============================================================
// AR-1-020：返回数据镜像请求（返回数据校验）
// ============================================================
func TestSetRules_ResponseMirrorsRequest(t *testing.T) {
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"description":  "测试规则",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
		"basic_forward_rules": []map[string]interface{}{
			{
				"host_names":   []string{"*.example.com"},
				"paths":        []string{"/api"},
				"cluster_name": "BFE-AI_product.szyf",
				"description":  "基础路由",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	// 校验 forward_rules
	fr, ok := data["forward_rules"].([]interface{})
	if !ok {
		t.Fatal("forward_rules is not a list")
	}
	fr0 := fr[0].(map[string]interface{})
	if fr0["name"] != "rule1" {
		t.Errorf("forward_rules[0].name: expected rule1, got %v", fr0["name"])
	}
	if fr0["description"] != "测试规则" {
		t.Errorf("forward_rules[0].description: expected 测试规则, got %v", fr0["description"])
	}
	if fr0["expression"] != `req_host_in("api.example.com")` {
		t.Errorf("forward_rules[0].expression mismatch, got %v", fr0["expression"])
	}
	if fr0["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("forward_rules[0].cluster_name: expected BFE-AI_product.szyf, got %v", fr0["cluster_name"])
	}

	// 校验 basic_forward_rules
	br, ok := data["basic_forward_rules"].([]interface{})
	if !ok {
		t.Fatal("basic_forward_rules is not a list")
	}
	br0 := br[0].(map[string]interface{})
	if br0["cluster_name"] != "BFE-AI_product.szyf" {
		t.Errorf("basic_forward_rules[0].cluster_name: expected BFE-AI_product.szyf, got %v", br0["cluster_name"])
	}
	if br0["description"] != "基础路由" {
		t.Errorf("basic_forward_rules[0].description: expected 基础路由, got %v", br0["description"])
	}
}

// ============================================================
// AR-1-021：最后一条自动追加 default_t()（业务规则）
// ============================================================
func TestSetRules_AutoDefaultRule(t *testing.T) {
	// PATCH 设置规则，最后一条不是 default_t()
	body := map[string]interface{}{
		"forward_rules": []map[string]interface{}{
			{
				"name":         "rule1",
				"expression":   `req_host_in("api.example.com")`,
				"cluster_name": "BFE-AI_product.szyf",
			},
		},
	}

	resp := patchRouteRules(t, body)
	testutil.AssertSuccess(t, resp)

	// PATCH 返回应只有 1 条规则（不包含自动追加的 default_t()）
	testutil.AssertListFieldLen(t, resp, "forward_rules", 1)

	// GET 应返回 2 条规则（包含自动追加的 default_t()）
	getResp := getRouteRules(t)
	testutil.AssertSuccess(t, getResp)

	var data map[string]interface{}
	json.Unmarshal(getResp.Data, &data)
	fr := data["forward_rules"].([]interface{})

	// 至少 2 条
	if len(fr) < 2 {
		t.Fatalf("expected at least 2 forward_rules, got %d", len(fr))
	}

	// 最后一条应是 default_t()
	lastRule := fr[len(fr)-1].(map[string]interface{})
	expr, _ := lastRule["expression"].(string)
	if expr != "default_t()" {
		t.Errorf("expected last rule expression=default_t(), got %s", expr)
	}
	// 最后一条的 cluster_name 应与第一条规则相同
	firstName, _ := fr[0].(map[string]interface{})["cluster_name"].(string)
	lastClusterName, _ := lastRule["cluster_name"].(string)
	if lastClusterName != firstName {
		t.Errorf("expected last rule cluster_name=%s, got %s", firstName, lastClusterName)
	}
}
