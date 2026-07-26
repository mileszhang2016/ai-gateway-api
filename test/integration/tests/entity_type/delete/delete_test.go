package delete

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

	// 预先创建Entity-Type用于测试删除
	createEntityType("test_type_delete", "删除测试", 1)
	createEntityType("test_type_has_entity", "包含Entity的类型", 1)

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

func TestDeleteEntityType_Normal_Success(t *testing.T) {
	// ET-5-001: 正常删除Entity-Type
	client := testutil.GetClient()

	// 删除Entity-Type
	resp, err := client.Delete("/open-api/v1/entity-types/test_type_delete")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证已删除
	resp, err = client.Get("/open-api/v1/entity-types/test_type_delete")
	if err != nil {
		t.Fatalf("verify delete failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestDeleteEntityType_Abnormal_NotFound(t *testing.T) {
	// ET-5-002: 删除不存在的Entity-Type
	client := testutil.GetClient()
	resp, err := client.Delete("/open-api/v1/entity-types/non_existent_delete")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestDeleteEntityType_Business_HasEntity(t *testing.T) {
	// ET-5-003: 删除存在Entity的Entity-Type
	client := testutil.GetClient()

	// 创建使用该类型的Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_has_type",
		"type": "test_type_has_entity",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 尝试删除Entity-Type
	resp, err = client.Delete("/open-api/v1/entity-types/test_type_has_entity")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}
