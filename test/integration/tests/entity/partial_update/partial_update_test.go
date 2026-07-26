package partial_update

import (
	"encoding/json"
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

func TestPartialUpdateEntity_Normal_UpdateName(t *testing.T) {
	// ENT-5-001: 部分更新Entity名称
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_patch",
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

	// 部分更新Entity（仅传name）
	resp, err = client.Patch("/open-api/v1/entities/"+entityID.(string), map[string]interface{}{
		"name": "test_entity_patch_updated",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "name", "test_entity_patch_updated")
	testutil.AssertDataFieldEquals(t, resp, "type", "dep")
}

func TestPartialUpdateEntity_Abnormal_NotFound(t *testing.T) {
	// ENT-5-002: 更新不存在的Entity
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/entities/non_existent_patch", map[string]interface{}{
		"name": "test_entity_patch_update",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestPartialUpdateEntity_Normal_UpdateQuotaPlan(t *testing.T) {
	// ENT-5-003: 更新配额计划（当前功能存在系统错误）
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_quota_patch",
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

	// 更新配额计划
	resp, err = client.Patch("/open-api/v1/entities/"+entityID.(string), map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        50000000,
			"unit":         "total_token",
			"reset_period": "weekly",
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证配额计划字段
	var updatedData map[string]interface{}
	if err := json.Unmarshal(resp.Data, &updatedData); err != nil {
		t.Fatalf("unmarshal updated data: %v", err)
	}
	quotaPlan, ok := updatedData["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatalf("quota_plan not found or not object in response")
	}
	if quotaPlan["unlimited"] != false {
		t.Errorf("expected quota_plan.unlimited=false, got %v", quotaPlan["unlimited"])
	}
	if quotaPlan["quota"] != float64(50000000) {
		t.Errorf("expected quota_plan.quota=50000000, got %v", quotaPlan["quota"])
	}
	if quotaPlan["unit"] != "total_token" {
		t.Errorf("expected quota_plan.unit=total_token, got %v", quotaPlan["unit"])
	}
	if quotaPlan["reset_period"] != "weekly" {
		t.Errorf("expected quota_plan.reset_period=weekly, got %v", quotaPlan["reset_period"])
	}
}
