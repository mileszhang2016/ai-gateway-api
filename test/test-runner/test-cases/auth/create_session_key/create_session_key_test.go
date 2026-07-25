package create_session_key

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

func TestCreateSessionKey_Normal_SuccessLogin(t *testing.T) {
	// AUTH-9-001: 正确用户名密码登录
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_login",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 创建 Session Key
	resp, err = client.Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"user_name": "test_user_login",
		"password":  "password@123",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "user_name", "test_user_login")
	testutil.AssertDataFieldEquals(t, resp, "is_admin", false)
	testutil.AssertDataFieldNotEmpty(t, resp, "session_key")
}

func TestCreateSessionKey_Abnormal_WrongPassword(t *testing.T) {
	// AUTH-9-002: 密码错误
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_wrongpwd",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 使用错误密码登录
	resp, err = client.Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"user_name": "test_user_wrongpwd",
		"password":  "wrongpassword",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 401)
}

func TestCreateSessionKey_Abnormal_UserNotFound(t *testing.T) {
	// AUTH-9-003: 用户不存在
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"user_name": "non_existent_login",
		"password":  "password@123",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 401)
}

func TestCreateSessionKey_Required_MissingUserName(t *testing.T) {
	// AUTH-9-004: 缺少 user_name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"password": "password@123",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}