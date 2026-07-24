package full_update

import (
	"encoding/json"
	"fmt"
	"os"
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

// ============================================================
// AK-4-001：更新 description
// ============================================================
func TestFullUpdate_Description(t *testing.T) {
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

// ============================================================
// AK-4-002：更新不存在的 API-Key
// ============================================================
func TestFullUpdate_NonExistent(t *testing.T) {
	body := map[string]interface{}{
		"description": "test",
	}
	resp := fullUpdateAPIKey(t, "nonexistent-id", body)
	testutil.AssertErrCode(t, resp, 404)
}

// ============================================================
// AK-4-003：缺少 description
// ============================================================
func TestFullUpdate_MissingDescription(t *testing.T) {
	id := createAPIKeyWithID(t, "before-missing-desc")

	body := map[string]interface{}{}
	resp := fullUpdateAPIKey(t, id, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// AK-4-004：更新 expired_time
// ============================================================
func TestFullUpdate_ExpiredTime(t *testing.T) {
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

// ============================================================
// AK-4-005：更新 enabled 状态
// ============================================================
func TestFullUpdate_Enabled(t *testing.T) {
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
