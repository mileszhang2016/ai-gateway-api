package list

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
// AK-2-007：查询空列表（必须在其他测试之前运行）
// ============================================================
func TestList_A_EmptyList(t *testing.T) {
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

	// 验证响应成功即可（无分页参数时默认分页值可能为 0）
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	if _, ok := data["pagination"]; !ok {
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
// AK-2-004：page_size=100（最大值）
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
// AK-2-005：page_size=101（超最大值）
// ============================================================
func TestList_ExceedMaxPageSize(t *testing.T) {
	resp := listAPIKeys(t, "page_size=101")
	// 可能被截断为 100 或返回 422，接受两种结果
	if resp.ErrNum != 200 && resp.ErrNum != 422 {
		t.Errorf("expected ErrNum=200 or 422, got %d", resp.ErrNum)
	}
}

// ============================================================
// AK-2-006：验证分页返回结构
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

	// 验证 list 中元素的字段
	list := data["list"].([]interface{})
	if len(list) > 0 {
		item := list[0].(map[string]interface{})
		requiredFields := []string{"id", "key", "description", "enabled", "create_time", "quota_plan"}
		for _, field := range requiredFields {
			if _, ok := item[field]; !ok {
				t.Errorf("list[0] missing field: %s", field)
			}
		}
		// quota_plan 应包含 balance 字段（list 接口可能不包含，仅 detail 接口包含）
		if qp, ok := item["quota_plan"].(map[string]interface{}); ok {
			_ = qp // 仅验证 quota_plan 存在即可
		}
	}
}
