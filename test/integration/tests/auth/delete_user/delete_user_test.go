package auth_test

import (
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestAuth_DeleteUser(t *testing.T) {
	t.Run("AUTH-2-001 删除用户", func(t *testing.T) {
		userName := testutil.UniqueUserName()
		if err := testutil.CreateUser(userName, "password@123"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/auth/users/" + userName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		resp, _ = testutil.GetClient().Get("/open-api/v1/auth/users/" + userName)
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("AUTH-2-002 删除不存在的用户", func(t *testing.T) {
		resp, err := testutil.GetClient().Delete("/open-api/v1/auth/users/non_existent_user")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})
}
