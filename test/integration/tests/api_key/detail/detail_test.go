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

func TestAPIKey_Detail(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("detail-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AK-3-001 详情返回包含 balance", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "id", apiKeyID)
		testutil.AssertDataFieldNotEmpty(t, resp, "quota_plan")
	})

	t.Run("AK-3-002 查询不存在的 API-Key", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/non-existent-id")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
