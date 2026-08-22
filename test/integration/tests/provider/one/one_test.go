package provider_test

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

func TestProvider_One(t *testing.T) {
	providerName := testutil.UniqueProviderName()
	if _, err := testutil.CreateProvider(providerName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("PV-3-001 查询存在的 Provider", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/providers/" + providerName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", providerName)
	})

	t.Run("PV-3-002 查询不存在的 Provider", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/providers/non_existent_provider")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteProvider(providerName)
	})
}
