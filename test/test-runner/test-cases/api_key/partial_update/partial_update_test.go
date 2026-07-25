package partial_update

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yf-networks/ai-gateway-api/test-runner/testutil"
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
func createAPIKey(t *testing.T, description string, enabled bool) (string, map[string]interface{}) {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
		"enabled":     enabled,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string), data
}

// helper: 部分更新
func partialUpdateAPIKey(t *testing.T, id string, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Patch(fmt.Sprintf("/open-api/v1/api-keys/%s", id), body)
	if err != nil {
		t.Fatalf("partial update API-Key failed: %v", err)
	}
	return resp
}

// helper: 创建 Entity-Type
func createEntityType(t *testing.T, typeName string, level int) {
	t.Helper()
	body := map[string]interface{}{
		"type_name": typeName,
		"level":     level,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", body)
	if err != nil {
		t.Fatalf("create Entity-Type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

// helper: 创建 Entity
func createEntity(t *testing.T, name, entityType string) string {
	t.Helper()
	body := map[string]interface{}{
		"name": name,
		"type": entityType,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/entities", body)
	if err != nil {
		t.Fatalf("create Entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// ============================================================
// AK-5-001：仅修改 description
// ============================================================
func TestPartialUpdate_Description(t *testing.T) {
	id, _ := createAPIKey(t, "original-desc", true)

	body := map[string]interface{}{
		"description": "patched-desc",
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["description"] != "patched-desc" {
		t.Errorf("expected description='patched-desc', got %v", data["description"])
	}
	// ID 不变
	if data["id"] != id {
		t.Errorf("expected id=%s unchanged, got %v", id, data["id"])
	}
	// enabled 不变
	if enabled, ok := data["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled=true unchanged, got %v", data["enabled"])
	}
}

// ============================================================
// AK-5-002：禁用 API-Key
// ============================================================
func TestPartialUpdate_Disable(t *testing.T) {
	id, _ := createAPIKey(t, "disable-test", true)

	body := map[string]interface{}{
		"enabled": false,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if enabled, ok := data["enabled"].(bool); !ok || enabled {
		t.Errorf("expected enabled=false, got %v", data["enabled"])
	}
}

// ============================================================
// AK-5-003：启用 API-Key
// ============================================================
func TestPartialUpdate_Enable(t *testing.T) {
	id, _ := createAPIKey(t, "enable-test", false)

	body := map[string]interface{}{
		"enabled": true,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if enabled, ok := data["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled=true, got %v", data["enabled"])
	}
}

// ============================================================
// AK-5-004：更新不存在的 API-Key
// ============================================================
func TestPartialUpdate_NonExistent(t *testing.T) {
	body := map[string]interface{}{
		"description": "test",
	}
	resp := partialUpdateAPIKey(t, "nonexistent-id", body)
	testutil.AssertErrCode(t, resp, 404)
	if !strings.Contains(resp.ErrMsg, "API-Key") {
		t.Errorf("expected error message containing 'API-Key', got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-005：更新 expired_time（永不过期）
// ============================================================
func TestPartialUpdate_ExpiredTimeUnlimited(t *testing.T) {
	id, _ := createAPIKey(t, "expired-test", true)

	body := map[string]interface{}{
		"expired_time": -1,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if et, ok := data["expired_time"].(float64); !ok || et != -1 {
		t.Errorf("expected expired_time=-1, got %v", data["expired_time"])
	}
}

// ============================================================
// AK-5-006：更新 expired_time（未来时间）
// ============================================================
func TestPartialUpdate_ExpiredTimeFuture(t *testing.T) {
	id, _ := createAPIKey(t, "expired-future", true)

	futureTime := time.Now().Unix() + 3600 // 当前时间+1小时
	body := map[string]interface{}{
		"expired_time": futureTime,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if et, ok := data["expired_time"].(float64); !ok || int64(et) != futureTime {
		t.Errorf("expected expired_time=%d, got %v", futureTime, data["expired_time"])
	}
}

// ============================================================
// AK-5-007：更新 quota_plan
// ============================================================
func TestPartialUpdate_QuotaPlan(t *testing.T) {
	id, _ := createAPIKey(t, "quota-test", true)

	body := map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"quota":        10000,
			"unit":         "token",
			"reset_period": "daily",
		},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if qp, ok := data["quota_plan"].(map[string]interface{}); ok {
		if quota, ok := qp["quota"].(float64); !ok || quota != 10000 {
			t.Errorf("expected quota_plan.quota=10000, got %v", qp["quota"])
		}
	} else {
		t.Error("quota_plan should be an object")
	}
}

// ============================================================
// AK-5-008：更新 rate_limit_policy
// ============================================================
func TestPartialUpdate_RateLimitPolicy(t *testing.T) {
	id, _ := createAPIKey(t, "rate-limit-test", true)

	body := map[string]interface{}{
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{
						"window_minutes": 1,
						"step_minutes":   1,
						"max_tokens":     1000,
					},
				},
			},
		},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if rlp, ok := data["rate_limit_policy"].(map[string]interface{}); ok {
		if enabled, ok := rlp["enabled"].(bool); !ok || !enabled {
			t.Errorf("expected rate_limit_policy.enabled=true, got %v", rlp["enabled"])
		}
	} else {
		t.Error("rate_limit_policy should be an object")
	}
}

// ============================================================
// AK-5-009：更新 models
// ============================================================
func TestPartialUpdate_Models(t *testing.T) {
	id, _ := createAPIKey(t, "models-test", true)

	body := map[string]interface{}{
		"models": []string{"gpt-3.5-turbo", "gpt-4"},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if models, ok := data["models"].([]interface{}); ok {
		modelStrs := make([]string, len(models))
		for i, m := range models {
			modelStrs[i] = m.(string)
		}
		if len(modelStrs) != 2 || modelStrs[0] != "gpt-3.5-turbo" || modelStrs[1] != "gpt-4" {
			t.Errorf("expected models=['gpt-3.5-turbo', 'gpt-4'], got %v", modelStrs)
		}
	} else {
		t.Error("models should be an array")
	}
}

// ============================================================
// AK-5-010：更新 subnet
// ============================================================
func TestPartialUpdate_Subnet(t *testing.T) {
	id, _ := createAPIKey(t, "subnet-test", true)

	body := map[string]interface{}{
		"subnet": []string{"192.168.1.0/24"},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if subnet, ok := data["subnet"].([]interface{}); ok {
		if len(subnet) != 1 || subnet[0].(string) != "192.168.1.0/24" {
			t.Errorf("expected subnet=['192.168.1.0/24'], got %v", subnet)
		}
	} else {
		t.Error("subnet should be an array")
	}
}

// ============================================================
// AK-5-011：更新 entity_id
// ============================================================
func TestPartialUpdate_EntityID(t *testing.T) {
	// 创建 Entity-Type
	createEntityType(t, "test_type_patch", 1)
	// 创建 Entity
	entityID := createEntity(t, "test_entity_patch", "test_type_patch")
	// 创建 API-Key（不挂载 Entity）
	id, _ := createAPIKey(t, "entity-test", true)

	body := map[string]interface{}{
		"entity_id": entityID,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["entity_id"] != entityID {
		t.Errorf("expected entity_id=%s, got %v", entityID, data["entity_id"])
	}
}

// ============================================================
// AK-5-012：description 超长（512字符）
// ============================================================
func TestPartialUpdate_DescriptionTooLong(t *testing.T) {
	id, _ := createAPIKey(t, "desc-test", true)

	longDesc := strings.Repeat("a", 512)
	body := map[string]interface{}{
		"description": longDesc,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "description must be less than 512 characters") {
		t.Errorf("expected error message about description length, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-013：expired_time=-2（非法值）
// ============================================================
func TestPartialUpdate_ExpiredTimeInvalid(t *testing.T) {
	id, _ := createAPIKey(t, "expired-invalid", true)

	body := map[string]interface{}{
		"expired_time": -2,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "Invalid expired_time") {
		t.Errorf("expected error message about invalid expired_time, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-014：expired_time=过去时间
// ============================================================
func TestPartialUpdate_ExpiredTimePast(t *testing.T) {
	id, _ := createAPIKey(t, "expired-past", true)

	pastTime := time.Now().Unix() - 3600 // 当前时间-1小时
	body := map[string]interface{}{
		"expired_time": pastTime,
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "expired_time must be >= current time") {
		t.Errorf("expected error message about past expired_time, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-015：subnet 格式错误
// ============================================================
func TestPartialUpdate_SubnetInvalidFormat(t *testing.T) {
	id, _ := createAPIKey(t, "subnet-invalid", true)

	body := map[string]interface{}{
		"subnet": []string{"invalid-subnet"},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "invalid subnet format") {
		t.Errorf("expected error message about subnet format, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-016：rate_limit_policy.enabled=true 但无规则
// ============================================================
func TestPartialUpdate_RateLimitPolicyNoRules(t *testing.T) {
	id, _ := createAPIKey(t, "rate-limit-no-rules", true)

	body := map[string]interface{}{
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
		},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "when rate_limit_policy.enabled is true, rules must be set") {
		t.Errorf("expected error message about rate limit rules, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-017：quota_plan.quota<0
// ============================================================
func TestPartialUpdate_QuotaPlanNegative(t *testing.T) {
	id, _ := createAPIKey(t, "quota-negative", true)

	body := map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"quota": -100,
		},
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "quota must be >= 0") {
		t.Errorf("expected error message about quota >= 0, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-018：entity_id 指向不存在的 Entity
// ============================================================
func TestPartialUpdate_EntityIDNotFound(t *testing.T) {
	id, _ := createAPIKey(t, "entity-not-found", true)

	body := map[string]interface{}{
		"entity_id": "nonexistent-entity-id",
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "Entity not found") {
		t.Errorf("expected error message about Entity not found, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-5-019：id 超长（>256字符）
// ============================================================
func TestPartialUpdate_IDTooLong(t *testing.T) {
	longID := strings.Repeat("a", 257)
	body := map[string]interface{}{
		"description": "test",
	}
	resp := partialUpdateAPIKey(t, longID, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AK-5-020：验证返回结构完整性
// ============================================================
func TestPartialUpdate_ResponseStructure(t *testing.T) {
	id, _ := createAPIKey(t, "response-structure", true)

	body := map[string]interface{}{
		"description": "patched",
	}
	resp := partialUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证必填字段存在且类型正确
	requiredFields := map[string]string{
		"id":                "string",
		"key":               "string",
		"description":       "string",
		"enabled":           "bool",
		"create_time":       "number",
		"update_time":       "number",
		"expired_time":      "number",
		"unlimited_quota":   "bool",
		"models":            "array",
		"subnet":            "array",
		"quota_plan":        "object",
		"rate_limit_policy": "object",
	}

	for field, expectedType := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("data missing field: %s", field)
			continue
		}

		switch expectedType {
		case "string":
			if _, ok := data[field].(string); !ok {
				t.Errorf("field %s should be string, got %T", field, data[field])
			}
		case "bool":
			if _, ok := data[field].(bool); !ok {
				t.Errorf("field %s should be bool, got %T", field, data[field])
			}
		case "number":
			if _, ok := data[field].(float64); !ok {
				t.Errorf("field %s should be number, got %T", field, data[field])
			}
		case "array":
			if _, ok := data[field].([]interface{}); !ok {
				t.Errorf("field %s should be array, got %T", field, data[field])
			}
		case "object":
			if _, ok := data[field].(map[string]interface{}); !ok {
				t.Errorf("field %s should be object, got %T", field, data[field])
			}
		}
	}

	// 验证 id 和 key 非空
	if idVal, ok := data["id"].(string); !ok || idVal == "" {
		t.Error("id should be non-empty string")
	}
	if keyVal, ok := data["key"].(string); !ok || keyVal == "" {
		t.Error("key should be non-empty string")
	}
	if ct, ok := data["create_time"].(float64); !ok || ct <= 0 {
		t.Error("create_time should be > 0")
	}
	if ut, ok := data["update_time"].(float64); !ok || ut <= 0 {
		t.Error("update_time should be > 0")
	}
}
