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

func TestEntity_PartialUpdate(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("E-5-001 部分更新 allow_models", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"allow_models": []string{"gpt-4"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "type", typeName)
	})

	t.Run("E-5-002 部分更新后查询一致性", func(t *testing.T) {
		_, err := testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"block_models": []string{"gpt-4-32k"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "block_models", []interface{}{"gpt-4-32k"})
	})

	t.Run("E-5-003 部分更新非法 route_rules（规则名重复）", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
			"route_rules": map[string]interface{}{
				"enabled": true,
				"rules": []interface{}{
					map[string]interface{}{
						"name": "dup",
						"Cond": "default_t()",
						"targets": []interface{}{
							map[string]interface{}{"ClusterName": "c1", "Weight": 100},
						},
					},
					map[string]interface{}{
						"name": "dup",
						"Cond": "default_t()",
						"targets": []interface{}{
							map[string]interface{}{"ClusterName": "c2", "Weight": 100},
						},
					},
				},
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
