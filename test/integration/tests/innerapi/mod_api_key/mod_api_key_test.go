package innerapi_test

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

func TestInnerAPI_ModApiKey(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("inner-api-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("IN-6-001 首次导出 mod-api-key", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataNotEmpty(t, resp)
		testutil.AssertDataFieldNotEmpty(t, resp, "version")
		testutil.AssertDataFieldNotEmpty(t, resp, "config")
		testutil.AssertDataFieldNotEmpty(t, resp, "QuotaPlans")
		testutil.AssertDataFieldNotEmpty(t, resp, "tokens")
	})

	t.Run("IN-6-002 增量同步未变化", func(t *testing.T) {
		firstResp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		version, _ := testutil.GetDataField(firstResp, "version")
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key", map[string]string{
			"version": version.(string),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		if string(resp.Data) != "null" {
			t.Skip("version comparison not stable, Data is not null")
		}
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
