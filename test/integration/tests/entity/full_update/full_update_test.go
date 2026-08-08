package entity_test

import (
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestEntity_FullUpdate(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	otherName := testutil.UniqueEntityName()
	_, err = testutil.CreateEntity(otherName, typeName, "")
	if err != nil {
		t.Fatalf("setup other failed: %v", err)
	}

	t.Run("E-4-001 全量更新 Entity name", func(t *testing.T) {
		newName := testutil.UniqueEntityName()
		resp, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"name":         newName,
			"type":         typeName,
			"allow_models": []string{"*"},
			"block_models": []string{},
			"quota_plan":   map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{
				"enabled": false,
			},
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", newName)
		testutil.AssertDataFieldEquals(t, resp, "type", typeName)
	})

	t.Run("E-4-002 全量更新后查询一致性", func(t *testing.T) {
		newName := testutil.UniqueEntityName()
		_, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"name":         newName,
			"type":         typeName,
			"allow_models": []string{"gpt-4"},
			"block_models": []string{},
			"quota_plan":   map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{
				"enabled": false,
			},
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", newName)
		testutil.AssertDataFieldEquals(t, resp, "allow_models", []interface{}{"gpt-4"})
	})

	t.Run("E-4-003 全量更新冲突 name", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"name":         otherName,
			"type":         typeName,
			"allow_models": []string{"*"},
			"block_models": []string{},
			"quota_plan":   map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{
				"enabled": false,
			},
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 555 && resp.ErrNum != 556 && resp.ErrNum != 500 {
			t.Errorf("expected conflict error, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
	})

	t.Run("E-4-004 全量更新修改 type", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"name":         testutil.UniqueEntityName(),
			"type":         typeName,
			"allow_models": []string{"*"},
			"block_models": []string{},
			"quota_plan":   map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{
				"enabled": false,
			},
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "type", typeName)
	})

	t.Run("E-4-005 全量更新非法 name（含首尾空白）", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"name":         " badname ",
			"type":         typeName,
			"allow_models": []string{"*"},
			"block_models": []string{},
			"quota_plan":   map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{
				"enabled": false,
			},
			"route_rules": map[string]interface{}{
				"enabled": false,
				"rules":   []interface{}{},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Cleanup(func() {
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
	})
}
