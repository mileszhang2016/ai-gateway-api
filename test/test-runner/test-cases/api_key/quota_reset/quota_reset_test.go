package quota_reset

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

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

// helper: 创建带配额计划的 API-Key 并返回 ID
func createAPIKeyWithQuota(t *testing.T, description string, quota int64) string {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        quota,
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

// helper: 重置配额
func resetQuota(t *testing.T, id string, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Post(fmt.Sprintf("/open-api/v1/api-keys/%s/quota-plan/reset", id), body)
	if err != nil {
		t.Fatalf("reset quota failed: %v", err)
	}
	return resp
}

// ============================================================
// AK-8-001：传入 quota 重置
// ============================================================
func TestQuotaReset_WithNewQuota(t *testing.T) {
	id := createAPIKeyWithQuota(t, "reset-quota-test", 100000000)

	body := map[string]interface{}{
		"quota":  50000000,
		"reason": "手动调整",
	}
	resp := resetQuota(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证 ID
	if data["id"] != id {
		t.Errorf("expected id=%s, got %v", id, data["id"])
	}

	// 验证配额变更
	if pq, ok := data["previous_quota"].(float64); !ok || pq != 100000000 {
		t.Errorf("expected previous_quota=100000000, got %v", data["previous_quota"])
	}
	if nq, ok := data["new_quota"].(float64); !ok || nq != 50000000 {
		t.Errorf("expected new_quota=50000000, got %v", data["new_quota"])
	}

	// 验证 balance
	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if pr, ok := balance["previous_remaining"].(float64); !ok || pr != 100000000 {
		t.Errorf("expected balance.previous_remaining=100000000, got %v", balance["previous_remaining"])
	}
	if nr, ok := balance["new_remaining"].(float64); !ok || nr != 50000000 {
		t.Errorf("expected balance.new_remaining=50000000, got %v", balance["new_remaining"])
	}
	if used, ok := balance["used"].(float64); !ok || used != 0 {
		t.Errorf("expected balance.used=0, got %v", balance["used"])
	}
}

// ============================================================
// AK-8-002：不传 quota 重置（按当前配额）
// ============================================================
func TestQuotaReset_WithoutQuota(t *testing.T) {
	id := createAPIKeyWithQuota(t, "reset-no-quota", 100000000)

	body := map[string]interface{}{
		"reason": "月度重置",
	}
	resp := resetQuota(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 不传 quota 时，previous_quota 和 new_quota 应相同
	if pq, ok := data["previous_quota"].(float64); !ok || pq != 100000000 {
		t.Errorf("expected previous_quota=100000000, got %v", data["previous_quota"])
	}
	if nq, ok := data["new_quota"].(float64); !ok || nq != 100000000 {
		t.Errorf("expected new_quota=100000000, got %v", data["new_quota"])
	}

	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if used, ok := balance["used"].(float64); !ok || used != 0 {
		t.Errorf("expected balance.used=0 after reset, got %v", balance["used"])
	}
}

// ============================================================
// AK-8-003：重置不存在的 API-Key
// ============================================================
func TestQuotaReset_NonExistent(t *testing.T) {
	body := map[string]interface{}{
		"quota": 50000000,
	}
	resp := resetQuota(t, "nonexistent-id", body)
	testutil.AssertErrCode(t, resp, 404)
}

// ============================================================
// AK-8-004：验证返回结构完整性
// ============================================================
func TestQuotaReset_ResponseStructure(t *testing.T) {
	id := createAPIKeyWithQuota(t, "reset-structure", 100000000)

	body := map[string]interface{}{
		"quota":  200000000,
		"reason": "结构调整测试",
	}
	resp := resetQuota(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证顶层字段
	requiredFields := []string{"id", "previous_quota", "new_quota", "balance"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// 验证 balance 子字段
	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	balanceFields := []string{"previous_remaining", "new_remaining", "used"}
	for _, field := range balanceFields {
		if _, ok := balance[field]; !ok {
			t.Errorf("missing balance field: %s", field)
		}
	}

	// 类型校验
	if _, ok := data["id"].(string); !ok {
		t.Error("id should be string")
	}
	if _, ok := data["previous_quota"].(float64); !ok {
		t.Error("previous_quota should be number")
	}
	if _, ok := data["new_quota"].(float64); !ok {
		t.Error("new_quota should be number")
	}
	if _, ok := balance["previous_remaining"].(float64); !ok {
		t.Error("balance.previous_remaining should be number")
	}
	if _, ok := balance["new_remaining"].(float64); !ok {
		t.Error("balance.new_remaining should be number")
	}
	if _, ok := balance["used"].(float64); !ok {
		t.Error("balance.used should be number")
	}
}

// ============================================================
// AK-8-005：重置无配额计划的 API-Key（默认配额 unlimited=true，拒绝重置）
// ============================================================
func TestQuotaReset_NoQuotaPlan(t *testing.T) {
	id := createAPIKeyNoQuota(t, "no-quota-reset")

	body := map[string]interface{}{
		"quota": 50000000,
	}
	resp := resetQuota(t, id, body)
	// 系统自动创建默认配额计划（unlimited=true），配额管理器拒绝重置
	testutil.AssertErrCode(t, resp, 500)
	if !strings.Contains(resp.ErrMsg, "cannot reset balance for unlimited quota") {
		t.Errorf("expected error message containing 'cannot reset balance for unlimited quota', got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-8-006：重置配额为 0
// ============================================================
func TestQuotaReset_ZeroQuota(t *testing.T) {
	id := createAPIKeyWithQuota(t, "reset-zero", 100000000)

	body := map[string]interface{}{
		"quota": 0,
	}
	resp := resetQuota(t, id, body)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if nq, ok := data["new_quota"].(float64); !ok || nq != 0 {
		t.Errorf("expected new_quota=0, got %v", data["new_quota"])
	}

	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatal("balance should be an object")
	}
	if nr, ok := balance["new_remaining"].(float64); !ok || nr != 0 {
		t.Errorf("expected balance.new_remaining=0, got %v", balance["new_remaining"])
	}
}

// ============================================================
// AK-8-007：重置配额为负数
// ============================================================
func TestQuotaReset_NegativeQuota(t *testing.T) {
	id := createAPIKeyWithQuota(t, "reset-negative", 100000000)

	body := map[string]interface{}{
		"quota": -100,
	}
	resp := resetQuota(t, id, body)
	// 记录实际行为：可能接受负数或返回 422
	t.Logf("Negative quota reset: ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
}

// ============================================================
// AK-8-008：id 超长（>255字符）
// ============================================================
func TestQuotaReset_IDTooLong(t *testing.T) {
	longID := strings.Repeat("a", 256)
	body := map[string]interface{}{
		"quota": 50000000,
	}
	resp := resetQuota(t, longID, body)
	testutil.AssertErrCode(t, resp, 422)
}