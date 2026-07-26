package create

import (
	"os"
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

func TestCreateEntityType_Normal_FullParams(t *testing.T) {
	// ET-1-001: 创建Entity-Type（完整参数）
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name":   "test_type_full",
		"description": "测试类型完整参数",
		"level":       1,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldEquals(t, resp, "type_name", "test_type_full")
	testutil.AssertDataFieldEquals(t, resp, "description", "测试类型完整参数")
	testutil.AssertDataFieldEquals(t, resp, "level", float64(1))
	testutil.AssertDataFieldNotEmpty(t, resp, "create_time")
}

func TestCreateEntityType_Normal_MinimalParams(t *testing.T) {
	// ET-1-002: 创建Entity-Type（仅必填字段）
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name": "test_type_min",
		"level":     2,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldEquals(t, resp, "type_name", "test_type_min")
	testutil.AssertDataFieldEquals(t, resp, "level", float64(2))
}

func TestCreateEntityType_Required_MissingTypeName(t *testing.T) {
	// ET-1-003: 缺少 type_name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"level": 1,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateEntityType_Required_MissingLevel(t *testing.T) {
	// ET-1-004: 缺少 level
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name": "test_type_nolevel",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateEntityType_Business_DuplicateName(t *testing.T) {
	// ET-1-005: 重复创建同名Entity-Type
	client := testutil.GetClient()

	// 先创建Entity-Type
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name": "test_type_dup",
		"level":     1,
	})
	if err != nil {
		t.Fatalf("create entity-type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 再次创建同名Entity-Type
	resp, err = client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name": "test_type_dup",
		"level":     2,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 556)
}
