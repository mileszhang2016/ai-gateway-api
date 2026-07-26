package full_update

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

	// 预先创建 Entity-Type
	createEntityType("dep", "一级部门", 1)
	createEntityType("team", "二级团队", 2)

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

func TestFullUpdateEntity_Normal_UpdateName(t *testing.T) {
	// ENT-4-001: 全量更新Entity名称
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_put",
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

	// 全量更新Entity
	resp, err = client.Put("/open-api/v1/entities/"+entityID.(string), map[string]interface{}{
		"name": "test_entity_put_updated",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "name", "test_entity_put_updated")
	testutil.AssertDataFieldEquals(t, resp, "type", "dep")
}

func TestFullUpdateEntity_Abnormal_NotFound(t *testing.T) {
	// ENT-4-002: 更新不存在的Entity
	client := testutil.GetClient()
	resp, err := client.Put("/open-api/v1/entities/non_existent_put", map[string]interface{}{
		"name": "test_entity_update",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestFullUpdateEntity_Business_NameConflict(t *testing.T) {
	// ENT-4-003: 更新后名称与其他Entity冲突
	client := testutil.GetClient()

	// 创建Entity1
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_conflict1",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity1 failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	entityID1, err := testutil.GetDataField(resp, "id")
	if err != nil {
		t.Fatalf("get entity1 id failed: %v", err)
	}

	// 创建Entity2
	resp, err = client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_conflict2",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity2 failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 更新Entity1的名称为Entity2的名称
	resp, err = client.Put("/open-api/v1/entities/"+entityID1.(string), map[string]interface{}{
		"name": "test_entity_conflict2",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 500)
}
