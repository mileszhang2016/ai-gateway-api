package create_user

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

func TestCreateUser_Normal_RegularUser(t *testing.T) {
	// AUTH-1-001: 创建普通用户
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_001",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestCreateUser_Normal_AdminUser(t *testing.T) {
	// AUTH-1-002: 创建管理员用户
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_admin_001",
		"password":  "password@123",
		"is_admin":  true,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestCreateUser_Required_MissingUserName(t *testing.T) {
	// AUTH-1-003: 缺少 user_name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"password": "password@123",
		"is_admin": false,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateUser_Required_MissingPassword(t *testing.T) {
	// AUTH-1-004: 缺少 password
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_002",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateUser_Required_MissingIsAdmin(t *testing.T) {
	// AUTH-1-005: 缺少 is_admin
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_003",
		"password":  "password@123",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateUser_Business_DuplicateUser(t *testing.T) {
	// AUTH-1-006: 重复创建同名用户
	client := testutil.GetClient()

	// 先创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_dup",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 再次创建同名用户
	resp, err = client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_dup",
		"password":  "password@456",
		"is_admin":  true,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 555)
}

func TestCreateUser_Boundary_EmptyUserName(t *testing.T) {
	// AUTH-1-007: user_name 为空字符串
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}