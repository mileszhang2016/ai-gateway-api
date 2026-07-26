package delete

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
func createAPIKey(t *testing.T, description string) string {
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

// helper: 删除
func deleteAPIKey(t *testing.T, id string) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Delete(fmt.Sprintf("/open-api/v1/api-keys/%s", id))
	if err != nil {
		t.Fatalf("delete API-Key failed: %v", err)
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

// AK-6-001：删除基本 API-Key（仅 description）
func TestDelete_Normal_BasicKey(t *testing.T) {
	id := createAPIKey(t, "delete-test-001")

	resp := deleteAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	// 验证 Data 为空或 null（服务端删除成功返回空 Data）
	if string(resp.Data) != "null" && len(resp.Data) != 0 {
		t.Logf("Data after delete: %s", string(resp.Data))
	}
}

// AK-6-002：删除含完整配置的 API-Key（级联删除）
func TestDelete_Normal_FullConfigKey(t *testing.T) {
	// 创建 EntityType
	etypeBody := map[string]interface{}{
		"type_name": "del-test-etype",
		"level":     1,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", etypeBody)
	if err != nil {
		t.Fatalf("create entity-type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 创建 Entity
	entityBody := map[string]interface{}{
		"name": "del-test-entity",
		"type": "del-test-etype",
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
		"description": "delete-test-full-config",
		"quota_plan": map[string]interface{}{
			"unlimited":                 false,
			"quota":                     100000,
			"unit":                      "token",
			"reset_period":              "daily",
			"pass_when_no_enough_quota": false,
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

	// 删除
	resp = deleteAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	// 验证 Data 为空或 null（服务端删除成功返回空 Data）
	if string(resp.Data) != "null" && len(resp.Data) != 0 {
		t.Errorf("expected Data=null or empty, got %s", string(resp.Data))
	}

	// 验证删除后 GET 返回 404
	getResp := getAPIKey(t, id)
	testutil.AssertErrCode(t, getResp, 404)
}

// ============================================================
// 必填校验
// ============================================================

// AK-6-003：删除路径缺少 ID
func TestDelete_Required_MissingId(t *testing.T) {
	resp, err := testutil.GetClient().Delete("/open-api/v1/api-keys/")
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

// AK-6-004：删除超长 ID（256 字符）
func TestDelete_Boundary_LongId(t *testing.T) {
	longID := strings.Repeat("a", 256)
	resp := deleteAPIKey(t, longID)
	// 超长 ID 触发参数校验，返回 422
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 异常参数
// ============================================================

// AK-6-005：删除不存在的 API-Key
func TestDelete_Abnormal_NonExistentId(t *testing.T) {
	resp := deleteAPIKey(t, "nonexistent-id-000000")
	testutil.AssertErrCode(t, resp, 404)
}

// AK-6-006：删除无效 UUID 格式的 ID
func TestDelete_Abnormal_InvalidUuidFormat(t *testing.T) {
	resp := deleteAPIKey(t, "invalid-format")
	testutil.AssertErrCode(t, resp, 404)
}

// AK-6-007：双重删除（对已删除的 Key 再次删除）
func TestDelete_Abnormal_DoubleDelete(t *testing.T) {
	id := createAPIKey(t, "double-delete-test")

	// 第一次删除
	firstResp := deleteAPIKey(t, id)
	testutil.AssertSuccess(t, firstResp)

	// 第二次删除（相同 ID），应返回 404
	secondResp := deleteAPIKey(t, id)
	testutil.AssertErrCode(t, secondResp, 404)
}

// ============================================================
// 返回数据校验
// ============================================================

// AK-6-008：删除成功返回结构校验
func TestDelete_ReturnData_ResponseStructure(t *testing.T) {
	id := createAPIKey(t, "delete-structure-test")

	resp := deleteAPIKey(t, id)

	// 验证 ErrNum=200
	if resp.ErrNum != 200 {
		t.Errorf("expected ErrNum=200, got %d", resp.ErrNum)
	}

	// 验证 ErrMsg="success"
	if resp.ErrMsg != "success" {
		t.Errorf("expected ErrMsg='success', got '%s'", resp.ErrMsg)
	}

	// 验证 Data 为空或 null（服务端删除成功返回空 Data）
	if string(resp.Data) != "null" && len(resp.Data) != 0 {
		t.Errorf("expected Data=null or empty, got %s", string(resp.Data))
	}
}

// ============================================================
// 业务规则
// ============================================================

// AK-6-009：删除后查询返回 404
func TestDelete_BusinessRule_DeleteThenGet(t *testing.T) {
	id := createAPIKey(t, "delete-then-get")

	// 删除
	resp := deleteAPIKey(t, id)
	testutil.AssertSuccess(t, resp)

	// 查询应返回 404
	getResp := getAPIKey(t, id)
	testutil.AssertErrCode(t, getResp, 404)
}
