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

func TestAuth_UserDetail(t *testing.T) {
	userName := testutil.UniqueUserName()
	if err := testutil.CreateUser(userName, "password@123"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AUTH-6-001 查询单个用户", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/auth/users/" + userName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "user_name", userName)
		testutil.AssertDataFieldEquals(t, resp, "is_admin", true)
	})

	t.Run("AUTH-6-002 查询不存在用户", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/auth/users/non_existent_user")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteUser(userName)
	})
}
