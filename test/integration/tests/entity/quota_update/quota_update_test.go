package entity_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestEntity_QuotaUpdate(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 初始化为有限 total_token 配额
	_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	})
	if err != nil {
		t.Fatalf("setup quota failed: %v", err)
	}

	t.Run("E-9-001 仅修改 quota（单位不变）保留 used", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000,
				"unit":         "total_token",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateEntityBalanceFloat(id, 200, 800); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     500,
				"unit":      "total_token",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(200), balance["used"], 0.00001)
		assert.InDelta(t, float64(300), balance["remaining"], 0.00001)
	})

	t.Run("E-9-002 RMB 配额仅修改 quota 保留 used", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000.1234,
				"unit":         "RMB",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateEntityBalanceFloat(id, 123.4567, 876.6667); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     800.0000,
				"unit":      "RMB",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(123.4567), balance["used"], 0.00001)
		assert.InDelta(t, float64(676.5433), balance["remaining"], 0.00001)
	})

	t.Run("E-9-003 修改 unit 重置 used 与 remaining", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000,
				"unit":         "total_token",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateEntityBalanceFloat(id, 200, 800); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     888.88,
				"unit":      "RMB",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(0), balance["used"], 0.00001)
		assert.InDelta(t, float64(888.88), balance["remaining"], 0.00001)
	})

	t.Run("E-9-004 unlimited 由 false 改为 true 重置为 sentinel", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000,
				"unit":         "total_token",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateEntityBalanceFloat(id, 200, 800); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": true,
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(0), balance["used"], 0.00001)
		assert.InDelta(t, float64(100000000), balance["remaining"], 0.00001)
	})

	t.Run("E-9-005 unlimited 由 true 改为 false 按新 quota 初始化", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": true,
			},
		})
		if err != nil {
			t.Fatalf("setup unlimited failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     500,
				"unit":      "total_token",
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(0), balance["used"], 0.00001)
		assert.InDelta(t, float64(500), balance["remaining"], 0.00001)
	})

	t.Run("E-9-006 普通属性修改不影响配额余额", func(t *testing.T) {
		name := testutil.UniqueEntityName()
		id, err := testutil.CreateEntity(name, typeName, "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteEntity(id)

		_, err = testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000,
				"unit":         "total_token",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateEntityBalanceFloat(id, 200, 800); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Patch("/open-api/v1/entities/"+id, map[string]interface{}{
			"allow_models": []string{"gpt-4"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		balance := fetchEntityBalance(t, id)
		assert.InDelta(t, float64(200), balance["used"], 0.00001)
		assert.InDelta(t, float64(800), balance["remaining"], 0.00001)
	})

	t.Cleanup(func() {
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
	})
}

func updateEntityBalanceFloat(entityID string, used, remaining float64) error {
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

func fetchEntityBalance(t *testing.T, entityID string) map[string]interface{} {
	resp, err := testutil.GetClient().Get("/open-api/v1/entities/" + entityID + "/quota-plan")
	if err != nil {
		t.Fatalf("query quota-plan failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal quota-plan data: %v", err)
	}
	balance, ok := data["balance"].(map[string]interface{})
	if !ok {
		t.Fatalf("balance is not an object")
	}
	return balance
}
