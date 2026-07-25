package delete_user

import (
	"os"
	"testing"

	"github.com/yf-networks/ai-gateway-api/test-runner/testutil"
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

func TestDeleteUser_Normal_Success(t *testing.T) {
	// AUTH-2-001: 正常删除用户
	client := testutil.GetClient()

	// 先创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_del",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 删除用户
	resp, err = client.Delete("/open-api/v1/auth/users/test_user_del")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestDeleteUser_Abnormal_NotFound(t *testing.T) {
	// AUTH-2-002: 删除不存在的用户
	client := testutil.GetClient()
	resp, err := client.Delete("/open-api/v1/auth/users/non_existent_user")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}