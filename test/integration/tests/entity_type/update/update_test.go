package entity_type_test

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
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

func TestEntityType_Update(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("ET-4-001 更新 Entity-Type 描述", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/entity-types/"+typeName, map[string]interface{}{
			"description": "更新后的描述",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "type_name", typeName)
		testutil.AssertDataFieldEquals(t, resp, "description", "更新后的描述")
		testutil.AssertDataFieldEquals(t, resp, "level", float64(1))
	})

	t.Run("ET-4-002 更新后查询一致性", func(t *testing.T) {
		_, err := testutil.GetClient().Patch("/open-api/v1/entity-types/"+typeName, map[string]interface{}{
			"description": "一致性校验描述",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/entity-types/" + typeName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "description", "一致性校验描述")
	})

	t.Run("ET-4-003 更新不存在的 Entity-Type", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/entity-types/non_existent_type", map[string]interface{}{
			"description": "x",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteEntityType(typeName)
	})
}
