package entity_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
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

func TestEntity_QuotaPlan(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 更新为非无限配额
	_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	})
	if err != nil {
		t.Fatalf("setup quota failed: %v", err)
	}

	t.Run("E-7-001 查询 Entity 配额计划", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID + "/quota-plan")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "unlimited", false)
		testutil.AssertDataFieldNotEmpty(t, resp, "balance")
	})

	t.Run("E-7-002 查询 RMB 配额余额精度", func(t *testing.T) {
		rmbEntityName := testutil.UniqueEntityName()
		rmbEntityID, err := testutil.CreateEntity(rmbEntityName, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(rmbEntityID)

		patchResp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+rmbEntityID, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        2000.12345678,
				"unit":         "RMB",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}
		t.Logf("patch resp: ErrNum=%d ErrMsg=%s Data=%s", patchResp.ErrNum, patchResp.ErrMsg, string(patchResp.Data))

		if err := updateEntityBalance(rmbEntityID, 2.34567890, 1997.77777788); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Get("/open-api/v1/entities/" + rmbEntityID + "/quota-plan")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance, err := testutil.GetDataField(resp, "balance")
		if err != nil {
			t.Fatalf("get balance failed: %v", err)
		}
		balanceMap, ok := balance.(map[string]interface{})
		if !ok {
			t.Fatalf("balance is not an object")
		}
		assert.InDelta(t, float64(1997.77777788), balanceMap["remaining"], 0.00000001)
		assert.InDelta(t, float64(2.34567890), balanceMap["used"], 0.00000001)
	})

	t.Cleanup(func() {
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
	})
}

func updateEntityBalance(entityID string, used, remaining float64) error {
	db, err := sql.Open("sqlite-strip", sm.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var planID int64
	err = db.QueryRow("SELECT quota_plan_id FROM entities WHERE entity_id = ?", entityID).Scan(&planID)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE quota_balances SET used = ?, remaining = ? WHERE quota_plan_id = ?", used, remaining, planID)
	return err
}
