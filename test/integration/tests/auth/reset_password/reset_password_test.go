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

func TestAuth_ResetPassword(t *testing.T) {
	userName := testutil.UniqueUserName()
	if err := testutil.CreateUser(userName, "password@123"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AUTH-3-001 管理员重置他人密码", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/auth/users/"+userName+"/passwd", map[string]interface{}{
			"password": "newpassword@456",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
	})

	t.Run("AUTH-3-002 缺少 password", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/auth/users/"+userName+"/passwd", map[string]interface{}{})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("AUTH-3-003 密码过短", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/auth/users/"+userName+"/passwd", map[string]interface{}{
			"password": "short1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("AUTH-3-004 修改不存在用户的密码", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/auth/users/non_existent_user/passwd", map[string]interface{}{
			"password": "newpassword@456",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteUser(userName)
	})
}
