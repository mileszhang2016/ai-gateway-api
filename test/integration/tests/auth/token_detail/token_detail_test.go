package auth_test

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

func TestAuth_TokenDetail(t *testing.T) {
	tokenName := testutil.UniqueTokenName()
	if _, err := testutil.CreateToken(tokenName, "System"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AUTH-11-001 Token 详情", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/auth/tokens/" + tokenName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", tokenName)
		testutil.AssertDataFieldNotEmpty(t, resp, "token")
		testutil.AssertDataFieldEquals(t, resp, "scope", "System")
	})

	t.Cleanup(func() {
		testutil.DeleteToken(tokenName)
	})
}
