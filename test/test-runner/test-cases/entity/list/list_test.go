package list

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestListEntity_Normal_GetList(t *testing.T) {
	// ENT-2-001: 获取Entity列表
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_list",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取Entity列表
	resp, err = client.Get("/open-api/v1/entities")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldNotEmpty(t, resp, "list")
	testutil.AssertDataFieldNotEmpty(t, resp, "pagination")
}

func TestListEntity_Data_FieldCompleteness(t *testing.T) {
	// ENT-2-002: 验证返回字段完整性
	client := testutil.GetClient()

	// 创建Entity
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name": "test_entity_check",
		"type": "dep",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取Entity列表
	resp, err = client.Get("/open-api/v1/entities")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证返回结构
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data failed: %v", err)
	}

	list, ok := data["list"].([]interface{})
	assert.True(t, ok && len(list) > 0, "list should be non-empty array")

	// 验证每个Entity包含必要字段
	for _, item := range list {
		entity, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		assert.Contains(t, entity, "id", "entity should have id")
		assert.Contains(t, entity, "name", "entity should have name")
		assert.Contains(t, entity, "type", "entity should have type")
	}
}
