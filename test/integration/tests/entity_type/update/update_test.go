package update

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

	// 预先创建Entity-Type用于测试更新
	createEntityType("test_type_update", "原始描述", 1)
	createEntityType("test_type_empty_desc", "原始描述", 2)

	code := m.Run()

	sm.Shutdown()
	os.Exit(code)
}

func createEntityType(typeName, description string, level int) {
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name":   typeName,
		"description": description,
		"level":       level,
	})
	if err != nil {
		panic("failed to create entity-type " + typeName + ": " + err.Error())
	}
	if resp.ErrNum != 200 && resp.ErrNum != 555 {
		panic("failed to create entity-type " + typeName + ": " + resp.ErrMsg)
	}
}

func TestUpdateEntityType_Normal_UpdateDescription(t *testing.T) {
	// ET-4-001: 更新Entity-Type描述
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/entity-types/test_type_update", map[string]interface{}{
		"description": "更新后的描述",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldEquals(t, resp, "type_name", "test_type_update")
	testutil.AssertDataFieldEquals(t, resp, "description", "更新后的描述")
	testutil.AssertDataFieldEquals(t, resp, "level", float64(1))
}

func TestUpdateEntityType_Abnormal_NotFound(t *testing.T) {
	// ET-4-002: 更新不存在的Entity-Type
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/entity-types/non_existent_update", map[string]interface{}{
		"description": "更新描述",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestUpdateEntityType_Boundary_EmptyDescription(t *testing.T) {
	// ET-4-003: 更新空描述
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/entity-types/test_type_empty_desc", map[string]interface{}{
		"description": "",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldEquals(t, resp, "type_name", "test_type_empty_desc")
	testutil.AssertDataFieldEquals(t, resp, "description", "")
}