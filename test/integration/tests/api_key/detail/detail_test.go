package detail

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// helper: 创建 API-Key 并返回 ID
func createAPIKeyWithID(t *testing.T, description string) string {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 创建 API-Key（自定义 body）并返回 ID
func createAPIKeyWithBody(t *testing.T, body map[string]interface{}) string {
	t.Helper()
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 查询单个 API-Key
func getAPIKey(t *testing.T, id string) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Get(fmt.Sprintf("/open-api/v1/api-keys/%s", id))
	if err != nil {
		t.Fatalf("get API-Key failed: %v", err)
	}
	return resp
}

// ============================================================
// 正常参数
// ============================================================

// AK-3-001：查询基本 API-Key（仅 description）
func TestDetail_Normal_BasicKey(t *testing.T) {
	id := createAPIKeyWithID(t, "detail-test-001")

	resp := getAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证 ID 一致
	if data["id"] != id {
		t.Errorf("expected id=%s, got %v", id, data["id"])
	}

	// 验证 description
	if data["description"] != "detail-test-001" {
		t.Errorf("expected description='detail-test-001', got %v", data["description"])
	}

	// 验证 quota_plan 存在
	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}
}

// AK-3-002：查询含完整配置的 API-Key
func TestDetail_Normal_FullConfigKey(t *testing.T) {
	// 创建 EntityType
	etypeBody := map[string]interface{}{
		"type_name": "detail-test-etype",
		"level":     1,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", etypeBody)
	if err != nil {
		t.Fatalf("create entity-type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 创建 Entity
	entityBody := map[string]interface{}{
		"name": "detail-test-entity",
		"type": "detail-test-etype",
	}
	resp, err = testutil.GetClient().Post("/open-api/v1/entities", entityBody)
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	var entityData map[string]interface{}
	json.Unmarshal(resp.Data, &entityData)
	entityID := entityData["id"].(string)

	// 创建 API-Key，含完整配置
	body := map[string]interface{}{
		"description": "detail-test-full-config",
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        100000,
			"unit":         "token",
			"reset_period": "daily",
		},
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{
						"name":           "test",
						"model":          "*",
						"window_minutes": 1,
						"max_tokens":     1000,
						"step_minutes":   1,
					},
				},
			},
		},
		"entity_id": entityID,
	}
	id := createAPIKeyWithBody(t, body)

	// 查询
	resp = getAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证 quota_plan
	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}

	// 验证 rate_limit_policy
	if data["rate_limit_policy"] == nil {
		t.Error("rate_limit_policy should not be null")
	}

	// 验证 entity_id
	if eid, ok := data["entity_id"].(string); !ok || eid != entityID {
		t.Errorf("expected entity_id='%s', got %v", entityID, data["entity_id"])
	}
}

// ============================================================
// 必填校验
// ============================================================

// AK-3-003：查询路径缺少 ID
func TestDetail_Required_MissingId(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/")
	if err != nil {
		// 路由不匹配可能导致 HTTP 请求出错，视为预期行为
		t.Logf("missing ID request failed (expected): %v", err)
		return
	}
	// 如果返回了响应，预期为 404
	testutil.AssertErrCode(t, resp, 404)
}

// ============================================================
// 边界值
// ============================================================

// AK-3-004：查询超长 ID（256 字符）
func TestDetail_Boundary_LongId(t *testing.T) {
	longID := strings.Repeat("a", 256)
	resp := getAPIKey(t, longID)
	// 超长 ID 触发参数校验，返回 422
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 异常参数
// ============================================================

// AK-3-005：查询不存在的 API-Key
func TestDetail_Abnormal_NonExistentId(t *testing.T) {
	resp := getAPIKey(t, "nonexistent-id-000000")
	testutil.AssertErrCode(t, resp, 404)
}

// AK-3-006：查询无效 UUID 格式的 ID
func TestDetail_Abnormal_InvalidUuidFormat(t *testing.T) {
	resp := getAPIKey(t, "invalid-format")
	testutil.AssertErrCode(t, resp, 404)
}

// ============================================================
// 返回数据校验
// ============================================================

// AK-3-007：返回顶层字段完整性校验
func TestDetail_ReturnData_TopLevelFields(t *testing.T) {
	// 创建带完整配置的 API-Key
	body := map[string]interface{}{
		"description": "detail-top-fields-test",
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        100000,
			"unit":         "token",
			"reset_period": "daily",
		},
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{
						"name":           "test",
						"model":          "*",
						"window_minutes": 1,
						"max_tokens":     1000,
						"step_minutes":   1,
					},
				},
			},
		},
	}
	id := createAPIKeyWithBody(t, body)

	// 查询
	resp := getAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证顶层字段存在且类型正确
	topFields := []struct {
		name     string
		expected string // "string", "bool", "number", "array", "object"
	}{
		{"id", "string"},
		{"description", "string"},
		{"expired_time", "number"},
		{"enabled", "bool"},
		{"unlimited_quota", "bool"},
		{"models", "array"},
		{"subnet", "array"},
		{"entity_id", "string"},
		{"quota_plan", "object"},
		{"rate_limit_policy", "object"},
	}

	for _, f := range topFields {
		val, ok := data[f.name]
		if !ok {
			// entity_id 仅在显式设置时返回，作为可选字段处理
			if f.name == "entity_id" {
				t.Logf("field '%s' not present (optional when not set)", f.name)
				continue
			}
			t.Errorf("field '%s' not found in response", f.name)
			continue
		}
		if val == nil {
			t.Errorf("field '%s' is null", f.name)
			continue
		}
		// 类型校验
		switch f.expected {
		case "string":
			if _, ok := val.(string); !ok {
				t.Errorf("field '%s' expected string, got %T", f.name, val)
			}
		case "bool":
			if _, ok := val.(bool); !ok {
				t.Errorf("field '%s' expected bool, got %T", f.name, val)
			}
		case "number":
			if _, ok := val.(float64); !ok {
				t.Errorf("field '%s' expected number, got %T", f.name, val)
			}
		case "array":
			if _, ok := val.([]interface{}); !ok {
				t.Errorf("field '%s' expected array, got %T", f.name, val)
			}
		case "object":
			if _, ok := val.(map[string]interface{}); !ok {
				t.Errorf("field '%s' expected object, got %T", f.name, val)
			}
		}
	}
}

// AK-3-008：返回 quota_plan.balance 结构校验
func TestDetail_ReturnData_BalanceStructure(t *testing.T) {
	// 创建带配额计划的 API-Key
	body := map[string]interface{}{
		"description": "detail-balance-test",
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        100000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	var createData map[string]interface{}
	json.Unmarshal(resp.Data, &createData)
	id := createData["id"].(string)

	// 查询
	detailResp := getAPIKey(t, id)
	testutil.AssertSuccess(t, detailResp)

	var data map[string]interface{}
	json.Unmarshal(detailResp.Data, &data)

	// 验证 quota_plan 包含 balance
	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	balance, ok := qp["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan.balance should be an object in detail response")
	}

	if _, ok := balance["used"].(float64); !ok {
		t.Error("balance.used should be a number")
	}
	if _, ok := balance["remaining"].(float64); !ok {
		t.Error("balance.remaining should be a number")
	}
}

// AK-3-008：空 id 路径应返回 JSON 404，而非静态文件 HTML
func TestDetail_Abnormal_EmptyID(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)

	// 确保返回的是 JSON，不是 HTML
	if strings.HasPrefix(string(resp.Data), "<") {
		t.Errorf("expected JSON error, got HTML: %s", string(resp.Data))
	}
}
