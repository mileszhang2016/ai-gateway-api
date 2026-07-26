package quota_query

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

// helper: 创建带配额计划的 API-Key 并返回 ID
func createAPIKeyWithQuota(t *testing.T, description string) string {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        100000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 创建不带配额计划的 API-Key
func createAPIKeyNoQuota(t *testing.T, description string) string {
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

// helper: 创建无限配额的 API-Key
func createAPIKeyUnlimitedQuota(t *testing.T, description string) string {
	t.Helper()
	body := map[string]interface{}{
		"description":     description,
		"unlimited_quota": true,
		"quota_plan": map[string]interface{}{
			"unlimited": true,
		},
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 查询配额计划
func getQuotaPlan(t *testing.T, id string) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Get(fmt.Sprintf("/open-api/v1/api-keys/%s/quota-plan", id))
	if err != nil {
		t.Fatalf("get quota plan failed: %v", err)
	}
	return resp
}

// ============================================================
// AK-7-001：查询有配额的 API-Key
// ============================================================
func TestQuotaQuery_ExistingKey(t *testing.T) {
	id := createAPIKeyWithQuota(t, "quota-query-test")

	resp := getQuotaPlan(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证配额字段
	if uq, ok := data["unlimited"].(bool); !ok || uq {
		t.Errorf("expected unlimited=false, got %v", data["unlimited"])
	}
	if q, ok := data["quota"].(float64); !ok || q != 100000000 {
		t.Errorf("expected quota=100000000, got %v", data["quota"])
	}
	if unit, ok := data["unit"].(string); !ok || unit != "total_token" {
		t.Errorf("expected unit='total_token', got %v", data["unit"])
	}
	if rp, ok := data["reset_period"].(string); !ok || rp != "monthly" {
		t.Errorf("expected reset_period='monthly', got %v", data["reset_period"])
	}

	// 验证 balance
	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if used, ok := balance["used"].(float64); !ok || used != 0 {
		t.Errorf("expected balance.used=0, got %v", balance["used"])
	}
	if rem, ok := balance["remaining"].(float64); !ok || rem != 100000000 {
		t.Errorf("expected balance.remaining=100000000, got %v", balance["remaining"])
	}
}

// ============================================================
// AK-7-002：查询不存在的 API-Key
// ============================================================
func TestQuotaQuery_NonExistent(t *testing.T) {
	resp := getQuotaPlan(t, "nonexistent-id")
	testutil.AssertErrCode(t, resp, 404)
	if !strings.Contains(resp.ErrMsg, "API-Key") {
		t.Errorf("expected error message containing 'API-Key', got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-7-003：验证 balance 字段结构
// ============================================================
func TestQuotaQuery_BalanceStructure(t *testing.T) {
	id := createAPIKeyWithQuota(t, "balance-struct-test")

	resp := getQuotaPlan(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}

	// 验证 used 字段存在且为数字
	used, ok := balance["used"].(float64)
	if !ok {
		t.Error("balance.used should be a number")
	}
	if used < 0 {
		t.Errorf("balance.used should be >= 0, got %v", used)
	}

	// 验证 remaining 字段存在且为数字
	remaining, ok := balance["remaining"].(float64)
	if !ok {
		t.Error("balance.remaining should be a number")
	}
	if remaining < 0 {
		t.Errorf("balance.remaining should be >= 0, got %v", remaining)
	}

	// 验证 used + remaining = quota
	quota := data["quota"].(float64)
	if used+remaining != quota {
		t.Errorf("used(%.0f) + remaining(%.0f) should equal quota(%.0f)", used, remaining, quota)
	}
}

// ============================================================
// AK-7-004：查询无限配额的 API-Key
// ============================================================
func TestQuotaQuery_Unlimited(t *testing.T) {
	id := createAPIKeyUnlimitedQuota(t, "unlimited-quota-test")

	resp := getQuotaPlan(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if ul, ok := data["unlimited"].(bool); !ok || !ul {
		t.Errorf("expected unlimited=true, got %v", data["unlimited"])
	}
}

// ============================================================
// AK-7-005：查询无配额计划的 API-Key（系统自动创建默认配额计划）
// ============================================================
func TestQuotaQuery_NoQuotaPlan(t *testing.T) {
	id := createAPIKeyNoQuota(t, "no-quota-test")

	resp := getQuotaPlan(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 系统自动创建默认配额计划：unlimited=true, quota=0, unit="total_token", reset_period="never"
	if ul, ok := data["unlimited"].(bool); !ok || !ul {
		t.Errorf("expected unlimited=true, got %v", data["unlimited"])
	}
	if q, ok := data["quota"].(float64); !ok || q != 0 {
		t.Errorf("expected quota=0, got %v", data["quota"])
	}
	if unit, ok := data["unit"].(string); !ok || unit != "total_token" {
		t.Errorf("expected unit='total_token', got %v", data["unit"])
	}
	if rp, ok := data["reset_period"].(string); !ok || rp != "never" {
		t.Errorf("expected reset_period='never', got %v", data["reset_period"])
	}

	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if used, ok := balance["used"].(float64); !ok || used != 0 {
		t.Errorf("expected balance.used=0, got %v", balance["used"])
	}
	if rem, ok := balance["remaining"].(float64); !ok || rem != 0 {
		t.Errorf("expected balance.remaining=0, got %v", balance["remaining"])
	}
}

// ============================================================
// AK-7-006：id 超长（>255字符）
// ============================================================
func TestQuotaQuery_IDTooLong(t *testing.T) {
	longID := strings.Repeat("a", 256)
	resp := getQuotaPlan(t, longID)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AK-7-007：验证返回字段完整性
// ============================================================
func TestQuotaQuery_FieldCompleteness(t *testing.T) {
	id := createAPIKeyWithQuota(t, "fields-test")

	resp := getQuotaPlan(t, id)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证顶层字段
	requiredFields := map[string]string{
		"unlimited":                "bool",
		"pass_when_no_enough_quota": "bool",
		"quota":                    "number",
		"unit":                     "string",
		"reset_period":             "string",
		"balance":                  "object",
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
		case "object":
			if _, ok := data[field].(map[string]interface{}); !ok {
				t.Errorf("field %s should be object, got %T", field, data[field])
			}
		}
	}

	// 验证 unit 和 reset_period 非空
	if unit, ok := data["unit"].(string); !ok || unit == "" {
		t.Error("unit should be non-empty string")
	}
	if rp, ok := data["reset_period"].(string); !ok || rp == "" {
		t.Error("reset_period should be non-empty string")
	}

	// 验证 balance 子字段
	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if _, ok := balance["used"].(float64); !ok {
		t.Error("balance.used should be a number")
	}
	if _, ok := balance["remaining"].(float64); !ok {
		t.Error("balance.remaining should be a number")
	}
}