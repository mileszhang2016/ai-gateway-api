package auth_test

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

func TestAuth_DeleteToken(t *testing.T) {
	t.Run("AUTH-10-001 删除 Token", func(t *testing.T) {
		tokenName := testutil.UniqueTokenName()
		if _, err := testutil.CreateToken(tokenName, "System"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/auth/tokens/" + tokenName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
	})

	t.Run("AUTH-10-002 删除不存在 Token", func(t *testing.T) {
		resp, err := testutil.GetClient().Delete("/open-api/v1/auth/tokens/non_existent_token")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})
}
