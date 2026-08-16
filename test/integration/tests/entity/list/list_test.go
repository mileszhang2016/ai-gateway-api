package entity_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestEntity_List(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	ent1 := testutil.UniqueEntityName()
	ent2 := testutil.UniqueEntityName()
	id1, err := testutil.CreateEntity(ent1, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.CreateEntity(ent2, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("E-2-001 Entity 列表分页", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/entities", map[string]string{"page": "1", "page_size": "10"})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertPagination(t, resp, 1, 10, 2)
	})

	t.Run("E-2-002 按 type 过滤", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/entities", map[string]string{"type": typeName})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		for _, item := range data["list"].([]interface{}) {
			assert.Equal(t, typeName, item.(map[string]interface{})["type"])
		}
	})

	t.Run("E-2-003 分页参数边界", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/entities", map[string]string{"page": "1", "page_size": "1"})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertListFieldLen(t, resp, "list", 1)
		testutil.AssertPagination(t, resp, 1, 1, 2)
	})

	t.Cleanup(func() {
		testutil.DeleteEntity(id1)
		testutil.DeleteEntityType(typeName)
	})
}
