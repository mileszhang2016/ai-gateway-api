package api_key_test

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

func TestAPIKey_QuotaQuery(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("quota-query-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AK-7-001 查询配额计划含 balance", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID + "/quota-plan")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldNotEmpty(t, resp, "balance")
	})

	t.Run("AK-7-002 查询无限配额 API-Key 的 quota-plan", func(t *testing.T) {
		unlimitedID, err := testutil.CreateAPIKey("unlimited-key", "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+unlimitedID, map[string]interface{}{
			"quota_plan": map[string]interface{}{"unlimited": true},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + unlimitedID + "/quota-plan")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 200 && resp.ErrNum != 404 {
			t.Errorf("expected ErrNum=200 or 404, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
		t.Cleanup(func() {
			testutil.DeleteAPIKey(unlimitedID)
		})
	})

	t.Run("AK-7-003 配额计划余额反映真实使用情况", func(t *testing.T) {
		limitedID, err := testutil.CreateAPIKey("limited-key", "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+limitedID, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     100,
				"unit":      "total_token",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateAPIKeyBalance(limitedID, 10, 90); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + limitedID + "/quota-plan")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "quota", float64(100))
		testutil.AssertDataFieldNotEmpty(t, resp, "balance")

		balance, err := testutil.GetDataField(resp, "balance")
		if err != nil {
			t.Fatalf("get balance failed: %v", err)
		}
		balanceMap, ok := balance.(map[string]interface{})
		if !ok {
			t.Fatalf("balance is not an object")
		}
		if used, ok := balanceMap["used"].(float64); !ok || used != 10 {
			t.Errorf("expected balance.used=10, got %v", balanceMap["used"])
		}
		if remaining, ok := balanceMap["remaining"].(float64); !ok || remaining != 90 {
			t.Errorf("expected balance.remaining=90, got %v", balanceMap["remaining"])
		}

		t.Cleanup(func() {
			testutil.DeleteAPIKey(limitedID)
		})
	})

	t.Run("AK-7-004 查询 RMB 配额余额精度", func(t *testing.T) {
		rmbID, err := testutil.CreateAPIKey("rmb-balance-key", "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+rmbID, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        1000.12345678,
				"unit":         "RMB",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		if err := updateAPIKeyBalanceFloat(rmbID, 1.23456789, 998.88888889); err != nil {
			t.Fatalf("update balance failed: %v", err)
		}

		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + rmbID + "/quota-plan")
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
		assert.InDelta(t, float64(998.88888889), balanceMap["remaining"], 0.00000001)
		assert.InDelta(t, float64(1.23456789), balanceMap["used"], 0.00000001)

		t.Cleanup(func() {
			testutil.DeleteAPIKey(rmbID)
		})
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}

func updateAPIKeyBalance(apiKeyID string, used, remaining int64) error {
	db, err := sql.Open("sqlite-strip", sm.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var planID int64
	err = db.QueryRow("SELECT quota_plan_id FROM api_keys WHERE id = ?", apiKeyID).Scan(&planID)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE quota_balances SET used = ?, remaining = ? WHERE quota_plan_id = ?", used, remaining, planID)
	return err
}

func updateAPIKeyBalanceFloat(apiKeyID string, used, remaining float64) error {
	db, err := sql.Open("sqlite-strip", sm.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	var planID int64
	err = db.QueryRow("SELECT quota_plan_id FROM api_keys WHERE id = ?", apiKeyID).Scan(&planID)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE quota_balances SET used = ?, remaining = ? WHERE quota_plan_id = ?", used, remaining, planID)
	return err
}
