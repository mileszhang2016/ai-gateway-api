package api_key_test

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

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
