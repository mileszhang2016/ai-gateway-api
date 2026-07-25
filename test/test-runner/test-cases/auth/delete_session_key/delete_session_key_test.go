package delete_session_key

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

func TestDeleteSessionKey_Normal_Success(t *testing.T) {
	// AUTH-10-001: 正常删除
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_session_del",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取 Session Key
	resp, err = client.Post("/open-api/v1/auth/session-keys", map[string]interface{}{
		"user_name": "test_user_session_del",
		"password":  "password@123",
	})
	if err != nil {
		t.Fatalf("create session key failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 提取 session_key
	sessionKey, err := testutil.GetDataField(resp, "session_key")
	if err != nil {
		t.Fatalf("get session_key failed: %v", err)
	}

	// 删除 Session Key
	resp, err = client.Delete("/open-api/v1/auth/session-keys/" + sessionKey.(string))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestDeleteSessionKey_Abnormal_NotFound(t *testing.T) {
	// AUTH-10-002: 删除不存在的 key
	client := testutil.GetClient()
	resp, err := client.Delete("/open-api/v1/auth/session-keys/non_existent_session_key")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}