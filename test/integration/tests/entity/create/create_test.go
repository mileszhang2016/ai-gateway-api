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

func TestCreateEntity_Normal_MinimalFields(t *testing.T) {
	// ENT-1-001: 创建Entity（仅必填字段）
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_001",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldNotEmpty(t, resp, "id")
	testutil.AssertDataFieldEquals(t, resp, "name", "test_entity_001")
	testutil.AssertDataFieldEquals(t, resp, "type", "dep")
}

func TestCreateEntity_Normal_WithQuotaPlan(t *testing.T) {
	// ENT-1-002: 创建Entity（含配额计划）
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_quota",
		"type": "dep",
		"quota_plan": map[string]interface{}{
			"unlimited":                 false,
			"pass_when_no_enough_quota": false,
			"quota":                     100000000,
			"unit":                      "total_token",
			"reset_period":              "monthly",
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段（创建接口不返回 quota_plan 字段）
	testutil.AssertDataFieldEquals(t, resp, "name", "test_entity_quota")
	testutil.AssertDataFieldEquals(t, resp, "type", "dep")

	// 通过查询配额计划接口验证配额是否正确保存
	entityID, err := testutil.GetDataField(resp, "id")
	if err != nil {
		t.Fatalf("get id failed: %v", err)
	}
	resp, err = client.Get("/open-api/v1/entities/" + entityID.(string) + "/quota-plan")
	if err != nil {
		t.Fatalf("get quota plan failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataFieldEquals(t, resp, "quota", float64(100000000))
	testutil.AssertDataFieldEquals(t, resp, "unit", "total_token")
}

func TestCreateEntity_Required_MissingName(t *testing.T) {
	// ENT-1-003: 缺少 name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateEntity_Required_MissingType(t *testing.T) {
	// ENT-1-004: 缺少 type
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_notype",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateEntity_Business_DuplicateName(t *testing.T) {
	// ENT-1-005: 重复创建同名Entity
	client := testutil.GetClient()

	// 先创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_dup",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 再次创建同名Entity
	resp, err = client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_dup",
		"type": "team",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 555)
}
