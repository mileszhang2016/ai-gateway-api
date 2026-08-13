package api_key_test

import (
	"encoding/json"
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

func TestAPIKey_QuotaReset(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("quota-reset-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AK-8-001 重置配额余额", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/api-keys/"+apiKeyID+"/quota-plan/reset", map[string]interface{}{
			"reason": "test reset",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "id", apiKeyID)
	})

	t.Run("AK-8-002 重置并修改 quota", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/api-keys/"+apiKeyID+"/quota-plan/reset", map[string]interface{}{
			"quota":  500000,
			"reason": "adjust",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "new_quota", float64(500000))
	})

	t.Run("AK-8-003 重置 RMB 配额余额", func(t *testing.T) {
		rmbID, err := testutil.CreateAPIKey("rmb-reset-key", "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteAPIKey(rmbID)

		_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+rmbID, map[string]interface{}{
			"quota_plan": map[string]interface{}{
				"unlimited":    false,
				"quota":        100.5,
				"unit":         "RMB",
				"reset_period": "monthly",
			},
		})
		if err != nil {
			t.Fatalf("setup quota failed: %v", err)
		}

		resp, err := testutil.GetClient().Post("/open-api/v1/api-keys/"+rmbID+"/quota-plan/reset", map[string]interface{}{
			"quota":  200.8888,
			"reason": "adjust rmb quota",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
		assert.InDelta(t, float64(100.5), data["previous_quota"], 0.00001)
		assert.InDelta(t, float64(200.8888), data["new_quota"], 0.00001)
		balance := data["balance"].(map[string]interface{})
		assert.InDelta(t, float64(0), balance["used"], 0.00001)
		assert.InDelta(t, float64(200.8888), balance["new_remaining"], 0.00001)
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
