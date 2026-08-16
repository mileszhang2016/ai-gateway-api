package entity_test

import (
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestEntity_Delete(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	childTypeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := testutil.CreateEntityType(childTypeName, 2); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("E-6-001 删除 Entity", func(t *testing.T) {
		entityName := testutil.UniqueEntityName()
		entityID, err := testutil.CreateEntity(entityName, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/entities/" + entityID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		resp, _ = testutil.GetClient().Get("/open-api/v1/entities/" + entityID)
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("E-6-002 删除存在子节点的 Entity", func(t *testing.T) {
		parentName := testutil.UniqueEntityName()
		parentID, err := testutil.CreateEntity(parentName, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		childName := testutil.UniqueEntityName()
		childID, err := testutil.CreateEntity(childName, childTypeName, parentID)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/entities/" + parentID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 409 && resp.ErrNum != 422 {
			t.Errorf("expected ErrNum=409 or 422, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
		t.Cleanup(func() {
			testutil.DeleteEntity(childID)
			testutil.DeleteEntity(parentID)
		})
	})

	t.Run("E-6-003 删除被 API-Key 挂载的 Entity", func(t *testing.T) {
		entityName := testutil.UniqueEntityName()
		entityID, err := testutil.CreateEntity(entityName, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		apiKeyID, err := testutil.CreateAPIKey(testutil.UniqueAPIKeyDesc(), entityID)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/entities/" + entityID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 409 && resp.ErrNum != 422 {
			t.Errorf("expected ErrNum=409 or 422, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
		t.Cleanup(func() {
			testutil.DeleteAPIKey(apiKeyID)
			testutil.DeleteEntity(entityID)
		})
	})

	t.Cleanup(func() {
		testutil.DeleteEntityType(typeName)
		testutil.DeleteEntityType(childTypeName)
	})
}
