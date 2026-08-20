package api_key_test

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

func TestAPIKey_Delete(t *testing.T) {
	t.Run("AK-6-001 删除 API-Key", func(t *testing.T) {
		apiKeyID, err := testutil.CreateAPIKey("delete-key", "")
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/api-keys/" + apiKeyID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		resp, _ = testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("AK-6-002 删除不存在的 API-Key", func(t *testing.T) {
		resp, err := testutil.GetClient().Delete("/open-api/v1/api-keys/non-existent-id")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})
}
