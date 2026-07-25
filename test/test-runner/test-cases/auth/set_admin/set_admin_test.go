package set_admin

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

func TestSetAdmin_Normal_SetToAdmin(t *testing.T) {
	// AUTH-5-001: 设置为管理员
	client := testutil.GetClient()

	// 创建普通用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_promote",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 设置为管理员
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_promote/is_admin", map[string]interface{}{
		"is_admin": true,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestSetAdmin_Normal_RemoveAdmin(t *testing.T) {
	// AUTH-5-002: 取消管理员权限
	client := testutil.GetClient()

	// 创建管理员用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_admin_demote",
		"password":  "password@123",
		"is_admin":  true,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 取消管理员权限
	resp, err = client.Patch("/open-api/v1/auth/users/test_admin_demote/is_admin", map[string]interface{}{
		"is_admin": false,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestSetAdmin_Required_MissingIsAdmin(t *testing.T) {
	// AUTH-5-003: 缺少 is_admin
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_noadmin",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 设置管理员，缺少 is_admin
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_noadmin/is_admin", map[string]interface{}{})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestSetAdmin_Abnormal_UserNotFound(t *testing.T) {
	// AUTH-5-004: 修改不存在的用户
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/auth/users/non_existent_admin/is_admin", map[string]interface{}{
		"is_admin": true,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}
