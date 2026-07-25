package delete

import (
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

	// 预先创建 Entity-Type
	createEntityType("dep", "一级部门", 1)

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
	// 忽略重复创建错误（555），只处理其他错误
	if resp.ErrNum != 200 && resp.ErrNum != 555 {
		panic("failed to create entity-type " + typeName + ": " + resp.ErrMsg)
	}
}

func TestDeleteEntity_Normal_Success(t *testing.T) {
	// ENT-6-001: 正常删除Entity
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_del",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 提取id
	entityID, err := testutil.GetDataField(resp, "id")
	if err != nil {
		t.Fatalf("get id failed: %v", err)
	}

	// 删除Entity
	resp, err = client.Delete("/open-api/v1/entities/" + entityID.(string))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestDeleteEntity_Abnormal_NotFound(t *testing.T) {
	// ENT-6-002: 删除不存在的Entity
	client := testutil.GetClient()
	resp, err := client.Delete("/open-api/v1/entities/non_existent_del")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}
