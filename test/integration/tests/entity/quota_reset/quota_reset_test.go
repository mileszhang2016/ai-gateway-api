package quota_reset

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

func TestQuotaReset_Normal_WithoutNewQuota(t *testing.T) {
	// ENT-8-001: 重置配额余额（不传新配额）
	client := testutil.GetClient()

	// 创建Entity（配置非无限配额）
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_reset1",
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
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 提取id
	entityID, err := testutil.GetDataField(resp, "id")
	if err != nil {
		t.Fatalf("get id failed: %v", err)
	}

	// 重置配额余额（不传新配额）
	resp, err = client.Post("/open-api/v1/entities/"+entityID.(string)+"/quota-plan/reset", map[string]interface{}{
		"reason": "月度重置",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "id", entityID.(string))
}

func TestQuotaReset_Normal_WithNewQuota(t *testing.T) {
	// ENT-8-002: 重置配额余额（传新配额）
	client := testutil.GetClient()

	// 创建Entity（配置非无限配额）
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_reset2",
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
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 提取id
	entityID, err := testutil.GetDataField(resp, "id")
	if err != nil {
		t.Fatalf("get id failed: %v", err)
	}

	// 重置配额余额（传新配额）
	resp, err = client.Post("/open-api/v1/entities/"+entityID.(string)+"/quota-plan/reset", map[string]interface{}{
		"quota":  50000000,
		"reason": "调整配额",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "id", entityID.(string))
	testutil.AssertDataFieldEquals(t, resp, "new_quota", float64(50000000))
}
