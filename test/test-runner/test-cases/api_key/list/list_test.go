package list

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

// helper: 创建 API-Key
func createAPIKey(t *testing.T, description string) *testutil.APIResponse {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	return resp
}

// helper: 创建 API-Key 并返回 ID
func createAPIKeyWithID(t *testing.T, description string, enabled bool) string {
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
	return data["id"].(string)
}

// helper: 创建 API-Key 带 unlimited_quota
func createAPIKeyWithUnlimitedQuota(t *testing.T, description string, unlimited bool) string {
	t.Helper()
	body := map[string]interface{}{
		"description":      description,
		"unlimited_quota":  unlimited,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 创建 API-Key 挂载到 Entity
func createAPIKeyWithEntity(t *testing.T, description string, entityID string) string {
	t.Helper()
	body := map[string]interface{}{
		"description": description,
		"entity_id":   entityID,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	return data["id"].(string)
}

// helper: 创建 Entity-Type
func createEntityType(t *testing.T, typeName string, level int) {
	t.Helper()
	body := map[string]interface{}{
		"type_name":  typeName,
		"level":      level,
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

// helper: 查询列表
func listAPIKeys(t *testing.T, query string) *testutil.APIResponse {
	t.Helper()
	path := "/open-api/v1/api-keys"
	if query != "" {
		path = path + "?" + query
	}
	resp, err := testutil.GetClient().Get(path)
	if err != nil {
		t.Fatalf("list API-Keys failed: %v", err)
	}
	return resp
}

// ============================================================
// AK-2-013：查询空列表（必须在其他测试之前运行）
// ============================================================
func TestList_EmptyList(t *testing.T) {
	// 新测试环境，直接查询
	resp := listAPIKeys(t, "")
	testutil.AssertSuccess(t, resp)

	testutil.AssertListFieldLen(t, resp, "list", 0)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if pag, ok := data["pagination"].(map[string]interface{}); ok {
		if total, ok := pag["total"].(float64); ok && total != 0 {
			t.Errorf("expected total=0 for empty list, got %v", total)
		}
	}
}

// ============================================================
// AK-2-001：默认分页查询（无参数）
// ============================================================
func TestList_DefaultPagination(t *testing.T) {
	// 先创建 3 个 API-Key
	for i := 1; i <= 3; i++ {
		resp := createAPIKey(t, fmt.Sprintf("list-test-%d", i))
		testutil.AssertSuccess(t, resp)
	}

	// 查询列表
	resp := listAPIKeys(t, "")
	testutil.AssertSuccess(t, resp)

	// 验证 list 长度
	testutil.AssertListFieldLen(t, resp, "list", 3)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	if pag, ok := data["pagination"].(map[string]interface{}); ok {
		// 不传分页参数时，不分页模式，page 和 page_size 为 0
		if page, ok := pag["page"].(float64); !ok || page != 0 {
			t.Errorf("expected page=0 (no pagination), got %v", pag["page"])
		}
		if ps, ok := pag["page_size"].(float64); !ok || ps != 0 {
			t.Errorf("expected page_size=0 (no pagination), got %v", pag["page_size"])
		}
		if total, ok := pag["total"].(float64); !ok || total < 3 {
			t.Errorf("expected total>=3, got %v", pag["total"])
		}
	} else {
		t.Error("pagination should exist in response")
	}
}

// ============================================================
// AK-2-002：指定分页参数
// ============================================================
func TestList_CustomPagination(t *testing.T) {
	// 先创建 10 个 API-Key
	for i := 1; i <= 10; i++ {
		resp := createAPIKey(t, fmt.Sprintf("page-test-%d", i))
		testutil.AssertSuccess(t, resp)
	}

	// 查询第一页（每页5条）
	resp := listAPIKeys(t, "page=1&page_size=5")
	testutil.AssertSuccess(t, resp)

	testutil.AssertListFieldLen(t, resp, "list", 5)
	testutil.AssertPagination(t, resp, 1, 5, 10)
}

// ============================================================
// AK-2-003：按 enabled 过滤
// ============================================================
func TestList_FilterByEnabled(t *testing.T) {
	// 创建 1 个 enabled=true 和 1 个 enabled=false
	createAPIKeyWithID(t, "enabled-true", true)
	createAPIKeyWithID(t, "enabled-false", false)

	// 查询 enabled=true
	resp := listAPIKeys(t, "enabled=true")
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	list, ok := data["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Fatal("list should not be empty")
	}

	// 验证所有元素的 enabled=true
	for _, item := range list {
		itemMap := item.(map[string]interface{})
		if enabled, ok := itemMap["enabled"].(bool); !ok || !enabled {
			t.Errorf("expected enabled=true for all items, got %v", itemMap["enabled"])
		}
	}
}

// ============================================================
// AK-2-004：按 entity_id 过滤
// ============================================================
func TestList_FilterByEntityID(t *testing.T) {
	// 创建 Entity-Type
	createEntityType(t, "test_type_list", 1)

	// 创建 Entity
	entityID := createEntity(t, "test_entity_list", "test_type_list")

	// 创建 API-Key-A（挂载到该 Entity）
	createAPIKeyWithEntity(t, "key-with-entity", entityID)

	// 创建 API-Key-B（不挂载 Entity）
	createAPIKey(t, "key-without-entity")

	// 按 entity_id 过滤
	resp := listAPIKeys(t, "entity_id="+entityID)
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	list, ok := data["list"].([]interface{})
	if !ok {
		t.Fatal("list should be an array")
	}

	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}

	if len(list) > 0 {
		item := list[0].(map[string]interface{})
		if item["entity_id"] != entityID {
			t.Errorf("expected entity_id=%s, got %v", entityID, item["entity_id"])
		}
	}
}

// ============================================================
// AK-2-005：按 unlimited_quota 过滤
// ============================================================
func TestList_FilterByUnlimitedQuota(t *testing.T) {
	// 创建 API-Key-A（unlimited_quota=true）
	createAPIKeyWithUnlimitedQuota(t, "key-unlimited", true)

	// 创建 API-Key-B（unlimited_quota=false）
	createAPIKeyWithUnlimitedQuota(t, "key-limited", false)

	// 查询 unlimited_quota=true
	resp := listAPIKeys(t, "unlimited_quota=true")
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	list, ok := data["list"].([]interface{})
	if !ok {
		t.Fatal("list should be an array")
	}

	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}

	if len(list) > 0 {
		item := list[0].(map[string]interface{})
		if uq, ok := item["unlimited_quota"].(bool); !ok || !uq {
			t.Errorf("expected unlimited_quota=true, got %v", item["unlimited_quota"])
		}
	}
}

// ============================================================
// AK-2-006：page_size=100（最大值）
// ============================================================
func TestList_MaxPageSize(t *testing.T) {
	resp := listAPIKeys(t, "page_size=100")
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	pag, ok := data["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("pagination should be an object")
	}
	if ps, ok := pag["page_size"].(float64); !ok || ps != 100 {
		t.Errorf("expected page_size=100, got %v", pag["page_size"])
	}
}

// ============================================================
// AK-2-007：page_size=101（超最大值）
// ============================================================
func TestList_ExceedMaxPageSize(t *testing.T) {
	resp := listAPIKeys(t, "page_size=101")
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "page_size must be between 1 and 100") {
		t.Errorf("expected error message about page_size range, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-008：page=0（非法值）
// ============================================================
func TestList_InvalidPageZero(t *testing.T) {
	resp := listAPIKeys(t, "page=0")
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "page must be > 0") {
		t.Errorf("expected error message about page must be > 0, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-009：page=-1（负数）
// ============================================================
func TestList_InvalidPageNegative(t *testing.T) {
	resp := listAPIKeys(t, "page=-1")
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "page must be > 0") {
		t.Errorf("expected error message about page must be > 0, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-010：page_size=0（非法值）
// ============================================================
func TestList_InvalidPageSizeZero(t *testing.T) {
	resp := listAPIKeys(t, "page_size=0")
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "page_size must be between 1 and 100") {
		t.Errorf("expected error message about page_size range, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-011：page_size=-1（负数）
// ============================================================
func TestList_InvalidPageSizeNegative(t *testing.T) {
	resp := listAPIKeys(t, "page_size=-1")
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "page_size must be between 1 and 100") {
		t.Errorf("expected error message about page_size range, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-012：entity_id 超长（>64字符）
// ============================================================
func TestList_EntityIDTooLong(t *testing.T) {
	// 构造 100 个 'a' 的 entity_id
	longEntityID := strings.Repeat("a", 100)
	resp := listAPIKeys(t, "entity_id="+longEntityID)
	testutil.AssertErrCode(t, resp, 422)
	if !strings.Contains(resp.ErrMsg, "entity_id must be <= 64 characters") {
		t.Errorf("expected error message about entity_id length, got: %s", resp.ErrMsg)
	}
}

// ============================================================
// AK-2-014：验证分页返回结构
// ============================================================
func TestList_ResponseStructureValidation(t *testing.T) {
	// 先创建 1 个 API-Key
	createAPIKey(t, "structure-test")

	resp := listAPIKeys(t, "")
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	// 验证顶层结构
	if _, ok := data["list"].([]interface{}); !ok {
		t.Error("list should be an array")
	}

	pag, ok := data["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("pagination should be an object")
	}
	if _, ok := pag["page"].(float64); !ok {
		t.Error("pagination.page should be a number")
	}
	if _, ok := pag["page_size"].(float64); !ok {
		t.Error("pagination.page_size should be a number")
	}
	if _, ok := pag["total"].(float64); !ok {
		t.Error("pagination.total should be a number")
	}
}

// ============================================================
// AK-2-015：验证 list 元素字段完整性
// ============================================================
func TestList_FieldCompleteness(t *testing.T) {
	// 先创建 1 个 API-Key
	createAPIKey(t, "field-test")

	resp := listAPIKeys(t, "")
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)

	list, ok := data["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Fatal("list should not be empty")
	}

	item := list[0].(map[string]interface{})

	// 验证所有必填字段存在且类型正确
	// entity_id 为可选字段（未挂载 Entity 时可能不存在）
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
		"remaining_quota":   "number",
	}

	for field, expectedType := range requiredFields {
		if _, ok := item[field]; !ok {
			t.Errorf("list[0] missing field: %s", field)
			continue
		}

		// 验证类型
		switch expectedType {
		case "string":
			if _, ok := item[field].(string); !ok {
				t.Errorf("field %s should be string, got %T", field, item[field])
			}
		case "bool":
			if _, ok := item[field].(bool); !ok {
				t.Errorf("field %s should be bool, got %T", field, item[field])
			}
		case "number":
			if _, ok := item[field].(float64); !ok {
				t.Errorf("field %s should be number, got %T", field, item[field])
			}
		case "array":
			if _, ok := item[field].([]interface{}); !ok {
				t.Errorf("field %s should be array, got %T", field, item[field])
			}
		case "object":
			if _, ok := item[field].(map[string]interface{}); !ok {
				t.Errorf("field %s should be object, got %T", field, item[field])
			}
		}
	}

	// 验证 id 和 key 非空
	if id, ok := item["id"].(string); !ok || id == "" {
		t.Error("id should be non-empty string")
	}
	if key, ok := item["key"].(string); !ok || key == "" {
		t.Error("key should be non-empty string")
	}
	if ct, ok := item["create_time"].(float64); !ok || ct <= 0 {
		t.Error("create_time should be > 0")
	}
	if ut, ok := item["update_time"].(float64); !ok || ut <= 0 {
		t.Error("update_time should be > 0")
	}
}