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

func TestAuth_CreateSessionKey(t *testing.T) {
	userName := testutil.UniqueUserName()
	password := "password@123"
	if err := testutil.CreateUser(userName, password); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name:     "AUTH-7-001 正确登录创建 Session Key",
			body:     map[string]interface{}{"user_name": userName, "password": password},
			wantCode: 200,
		},
		{
			name:     "AUTH-7-002 密码错误",
			body:     map[string]interface{}{"user_name": userName, "password": "wrong"},
			wantCode: 401,
		},
		{
			name:     "AUTH-7-003 缺少 user_name",
			body:     map[string]interface{}{"password": password},
			wantCode: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testutil.GetClient().Post("/open-api/v1/auth/session-keys", tt.body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
		})
	}

	t.Cleanup(func() {
		testutil.DeleteUser(userName)
	})
}
