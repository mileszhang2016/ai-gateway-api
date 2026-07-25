package reset_password

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

func TestResetPassword_Normal_AdminResetOther(t *testing.T) {
	// AUTH-3-001: 管理员重置他人密码
	client := testutil.GetClient()

	// 创建管理员和普通用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_admin_reset",
		"password":  "password@123",
		"is_admin":  true,
	})
	if err != nil {
		t.Fatalf("create admin failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	resp, err = client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_reset",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 管理员重置普通用户密码（无需旧密码）
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_reset/passwd", map[string]interface{}{
		"password": "newpassword@456",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestResetPassword_Normal_SelfReset(t *testing.T) {
	// AUTH-3-002: 用户自己重置密码
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_self",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 用户自己重置密码（需要旧密码）
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_self/passwd", map[string]interface{}{
		"old_password": "password@123",
		"password":     "newpassword@456",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestResetPassword_Required_MissingPassword(t *testing.T) {
	// AUTH-3-003: 缺少 password
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_nopass",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 重置密码，缺少 password
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_nopass/passwd", map[string]interface{}{
		"old_password": "password@123",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestResetPassword_Abnormal_UserNotFound(t *testing.T) {
	// AUTH-3-004: 修改不存在的用户密码
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/auth/users/non_existent_pass/passwd", map[string]interface{}{
		"password": "newpassword@456",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestResetPassword_Abnormal_WrongOldPassword(t *testing.T) {
	// AUTH-3-005: old_password 错误
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_wrongpass",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 用错误的旧密码重置
	resp, err = client.Patch("/open-api/v1/auth/users/test_user_wrongpass/passwd", map[string]interface{}{
		"old_password": "wrongpassword",
		"password":     "newpassword@456",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}