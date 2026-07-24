package partial_update

import (
	"encoding/json"
	"fmt"
	"os"
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
}