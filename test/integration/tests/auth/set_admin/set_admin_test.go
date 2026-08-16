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

func TestAuth_SetAdmin(t *testing.T) {
	userName := testutil.UniqueUserName()
	if err := testutil.CreateUser(userName, "password@123"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AUTH-5-001 设置管理员为 true", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/auth/users/"+userName+"/is_admin", map[string]interface{}{
			"is_admin": true,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		getResp, err := testutil.GetClient().Get("/open-api/v1/auth/users/" + userName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, getResp)
		testutil.AssertDataFieldEquals(t, getResp, "is_admin", true)
	})

	t.Cleanup(func() {
		testutil.DeleteUser(userName)
	})
}
