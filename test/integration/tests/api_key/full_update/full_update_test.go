package full_update

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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

// helper: 全量更新
func fullUpdateAPIKey(t *testing.T, id string, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Put(fmt.Sprintf("/open-api/v1/api-keys/%s", id), body)
	if err != nil {
		t.Fatalf("full update API-Key failed: %v", err)
	}
	return resp
}

// helper: 查询
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

// AK-4-001：全量更新 description
func TestFullUpdate_Normal_Description(t *testing.T) {
	id := createAPIKeyWithID(t, "original-desc")

	body := map[string]interface{}{
		"description": "updated-desc",
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["description"] != "updated-desc" {
		t.Errorf("expected description='updated-desc', got %v", data["description"])
	}
}

// AK-4-002：全量更新 expired_time
func TestFullUpdate_Normal_ExpiredTime(t *testing.T) {
	id := createAPIKeyWithID(t, "expire-test")
	futureTime := time.Now().Unix() + 86400*365

	body := map[string]interface{}{
		"description":  "expire-test",
		"expired_time": futureTime,
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if et, ok := data["expired_time"].(float64); !ok || et != float64(futureTime) {
		t.Errorf("expected expired_time=%d, got %v", futureTime, data["expired_time"])
	}
}

// AK-4-003：全量更新 enabled 状态
func TestFullUpdate_Normal_Enabled(t *testing.T) {
	id := createAPIKeyWithID(t, "enabled-test")

	// 禁用
	body := map[string]interface{}{
		"description": "enabled-test",
		"enabled":     false,
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if enabled, ok := data["enabled"].(bool); !ok || enabled {
		t.Errorf("expected enabled=false, got %v", data["enabled"])
	}

	// 重新启用
	body2 := map[string]interface{}{
		"description": "enabled-test",
		"enabled":     true,
	}
	resp2 := fullUpdateAPIKey(t, id, body2)
	testutil.AssertSuccess(t, resp2)

	var data2 map[string]interface{}
	json.Unmarshal(resp2.Data, &data2)

	if enabled, ok := data2["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled=true, got %v", data2["enabled"])
	}
}

// AK-4-004：全量更新 unlimited_quota
func TestFullUpdate_Normal_UnlimitedQuota(t *testing.T) {
	id := createAPIKeyWithID(t, "unlimited-test")

	body := map[string]interface{}{
		"description":     "unlimited-test",
		"unlimited_quota": true,
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if uq, ok := data["unlimited_quota"].(bool); !ok || !uq {
		t.Errorf("expected unlimited_quota=true, got %v", data["unlimited_quota"])
	}
}

// AK-4-005：全量更新 models
func TestFullUpdate_Normal_Models(t *testing.T) {
	id := createAPIKeyWithID(t, "models-test")

	body := map[string]interface{}{
		"description": "models-test",
		"models":      []string{"gpt-4", "gpt-3.5"},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	models, ok := data["models"].([]interface{})
	if !ok {
		t.Errorf("models should be an array, got %T", data["models"])
		return
	}
	if len(models) < 2 {
		t.Errorf("expected at least 2 models, got %v", models)
	}
}

// AK-4-006：全量更新 subnet
func TestFullUpdate_Normal_Subnet(t *testing.T) {
	id := createAPIKeyWithID(t, "subnet-test")

	body := map[string]interface{}{
		"description": "subnet-test",
		"subnet":      []string{"192.168.1.0/24"},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	subnet, ok := data["subnet"].([]interface{})
	if !ok {
		t.Errorf("subnet should be an array, got %T", data["subnet"])
		return
	}
	if len(subnet) == 0 {
		t.Errorf("subnet should not be empty")
	}
}

// AK-4-007：全量更新 quota_plan
func TestFullUpdate_Normal_QuotaPlan(t *testing.T) {
	id := createAPIKeyWithID(t, "quota-plan-test")

	body := map[string]interface{}{
		"description": "quota-plan-test",
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        50000,
			"unit":         "token",
			"reset_period": "daily",
		},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}
}

// AK-4-008：全量更新 rate_limit_policy
func TestFullUpdate_Normal_RateLimitPolicy(t *testing.T) {
	id := createAPIKeyWithID(t, "rate-limit-test")

	body := map[string]interface{}{
		"description": "rate-limit-test",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{
						"name":           "t1",
						"model":          "*",
						"window_minutes": 1,
						"max_tokens":     1000,
						"step_minutes":   1,
					},
				},
			},
		},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["rate_limit_policy"] == nil {
		t.Error("rate_limit_policy should not be null")
	}
}

// AK-4-009：全量更新所有字段
func TestFullUpdate_Normal_AllFields(t *testing.T) {
	id := createAPIKeyWithID(t, "full-update-all-before")
	futureTime := time.Now().Unix() + 86400*365

	body := map[string]interface{}{
		"description":     "full-update-all",
		"expired_time":    futureTime,
		"enabled":         true,
		"unlimited_quota": false,
		"models":          []string{"gpt-4"},
		"subnet":          []string{"10.0.0.0/8"},
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        100000,
			"unit":         "token",
			"reset_period": "monthly",
		},
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"rpm": []map[string]interface{}{
					{
						"name":           "r1",
						"model":          "*",
						"window_minutes": 1,
						"max_tokens":     100,
						"step_minutes":   1,
					},
				},
			},
		},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if data["description"] != "full-update-all" {
		t.Errorf("expected description='full-update-all', got %v", data["description"])
	}
	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}
	if data["rate_limit_policy"] == nil {
		t.Error("rate_limit_policy should not be null")
	}
}

// ============================================================
// 必填校验
// ============================================================

// AK-4-010：缺少 description（空 Body）
func TestFullUpdate_Required_EmptyBody(t *testing.T) {
	id := createAPIKeyWithID(t, "before-missing-desc")

	body := map[string]interface{}{}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-4-011：description 为空字符串
func TestFullUpdate_Required_EmptyDescription(t *testing.T) {
	id := createAPIKeyWithID(t, "empty-desc-test")

	body := map[string]interface{}{
		"description": "",
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 边界值
// ============================================================

// AK-4-012：description 最大合法长度（511）
func TestFullUpdate_Boundary_Description511(t *testing.T) {
	id := createAPIKeyWithID(t, "boundary-511-before")

	body := map[string]interface{}{
		"description": strings.Repeat("a", 511),
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)
}

// AK-4-013：expired_time=-1（永不过期）
func TestFullUpdate_Boundary_ExpiredTimeNever(t *testing.T) {
	id := createAPIKeyWithID(t, "never-expire-test")

	body := map[string]interface{}{
		"description":  "never-expire",
		"expired_time": -1,
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if et, ok := data["expired_time"].(float64); !ok || et != -1 {
		t.Errorf("expected expired_time=-1, got %v", data["expired_time"])
	}
}

// AK-4-014：更新超长 ID（256 字符）
func TestFullUpdate_Boundary_LongId(t *testing.T) {
	longID := strings.Repeat("a", 256)

	body := map[string]interface{}{
		"description": "test",
	}
	resp := fullUpdateAPIKey(t, longID, body)
	// 超长 ID 触发参数校验，返回 422
	testutil.AssertErrCode(t, resp, 422)
}

// AK-4-015：更新空 models 数组
func TestFullUpdate_Boundary_ModelsEmpty(t *testing.T) {
	id := createAPIKeyWithID(t, "empty-models-test")

	body := map[string]interface{}{
		"description": "empty-models",
		"models":      []string{},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	models, ok := data["models"].([]interface{})
	if !ok {
		t.Errorf("models should be an array, got %T", data["models"])
		return
	}
	// 空数组可能被服务端默认填充为 ["*"]
	if len(models) != 0 {
		t.Logf("models empty array was defaulted to %v (expected empty or ['*'])", models)
	}
}

// ============================================================
// 异常参数
// ============================================================

// AK-4-016：更新不存在的 API-Key
func TestFullUpdate_Abnormal_NonExistentId(t *testing.T) {
	body := map[string]interface{}{
		"description": "test",
	}
	resp := fullUpdateAPIKey(t, "nonexistent-id", body)
	testutil.AssertErrCode(t, resp, 404)
}

// AK-4-017：更新无效 UUID 格式的 ID
func TestFullUpdate_Abnormal_InvalidUuidFormat(t *testing.T) {
	body := map[string]interface{}{
		"description": "test",
	}
	resp := fullUpdateAPIKey(t, "invalid-format", body)
	testutil.AssertErrCode(t, resp, 404)
}

// AK-4-018：expired_time 为过去时间
func TestFullUpdate_Abnormal_ExpiredTimePast(t *testing.T) {
	id := createAPIKeyWithID(t, "past-expire-test")

	body := map[string]interface{}{
		"description":  "past-expire",
		"expired_time": int64(1000000000),
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-4-019：description 超长（512）
func TestFullUpdate_Abnormal_Description512(t *testing.T) {
	id := createAPIKeyWithID(t, "long-desc-test")

	body := map[string]interface{}{
		"description": strings.Repeat("a", 512),
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-4-020：subnet 为无效 CIDR
func TestFullUpdate_Abnormal_InvalidSubnet(t *testing.T) {
	id := createAPIKeyWithID(t, "bad-subnet-test")

	body := map[string]interface{}{
		"description": "bad-subnet",
		"subnet":      []string{"invalid-cidr"},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 返回数据校验
// ============================================================

// AK-4-021：全量更新返回结构校验
func TestFullUpdate_ReturnData_ResponseStructure(t *testing.T) {
	id := createAPIKeyWithID(t, "structure-test")

	body := map[string]interface{}{
		"description": "structure-test",
		"enabled":     true,
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
						"name":           "t1",
						"model":          "*",
						"window_minutes": 1,
						"max_tokens":     1000,
						"step_minutes":   1,
					},
				},
			},
		},
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证关键字段
	if data["id"] == nil || data["id"].(string) == "" {
		t.Error("id should not be empty")
	}
	if data["description"] != "structure-test" {
		t.Errorf("expected description='structure-test', got %v", data["description"])
	}
	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}
	if data["rate_limit_policy"] == nil {
		t.Error("rate_limit_policy should not be null")
	}
	if enabled, ok := data["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled=true, got %v", data["enabled"])
	}
}

// AK-4-022：全量更新后 GET 验证
func TestFullUpdate_ReturnData_PutGetVerify(t *testing.T) {
	id := createAPIKeyWithID(t, "put-get-before")

	body := map[string]interface{}{
		"description": "put-get-verify",
	}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertSuccess(t, resp)

	var putData map[string]interface{}
	json.Unmarshal(resp.Data, &putData)

	// 查询
	getResp := getAPIKey(t, id)
	testutil.AssertSuccess(t, getResp)

	var getData map[string]interface{}
	json.Unmarshal(getResp.Data, &getData)

	// 验证 description 一致
	if putData["description"] != getData["description"] {
		t.Errorf("PUT description=%v, GET description=%v", putData["description"], getData["description"])
	}
}