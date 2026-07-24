package detail

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
// AK-3-001：查询存在的 API-Key
// ============================================================
func TestDetail_ExistingKey(t *testing.T) {
	// 先创建
	id := createAPIKeyWithID(t, "detail-test-001")

	// 查询
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

// ============================================================
// AK-3-002：查询不存在的 API-Key
// ============================================================
func TestDetail_NonExistent(t *testing.T) {
	resp := getAPIKey(t, "nonexistent-id")
	testutil.AssertErrCode(t, resp, 404)
}

// ============================================================
// AK-3-003：验证返回字段完整性（含 balance）
// ============================================================
func TestDetail_ResponseStructureValidation(t *testing.T) {
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
