package list_users

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestListUsers_Normal_GetList(t *testing.T) {
	// AUTH-4-001: 获取用户列表
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_list",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取用户列表
	resp, err = client.Get("/open-api/v1/auth/users")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回数组
	var users []interface{}
	if err := json.Unmarshal(resp.Data, &users); err != nil {
		t.Fatalf("unmarshal users failed: %v", err)
	}
	assert.NotEmpty(t, users, "users list should not be empty")

	// 验证每个用户包含必要字段
	for _, u := range users {
		user, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		assert.Contains(t, user, "user_name", "user should have user_name")
		assert.Contains(t, user, "is_admin", "user should have is_admin")
	}
}

func TestListUsers_Data_ContainsNewUser(t *testing.T) {
	// AUTH-4-002: 创建后验证列表包含新用户
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_check",
		"password":  "password@123",
		"is_admin":  true,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取用户列表
	resp, err = client.Get("/open-api/v1/auth/users")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证列表包含新创建的用户
	var users []interface{}
	if err := json.Unmarshal(resp.Data, &users); err != nil {
		t.Fatalf("unmarshal users failed: %v", err)
	}

	found := false
	var foundIsAdmin bool
	for _, u := range users {
		user, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		if userName, ok := user["user_name"].(string); ok && userName == "test_user_check" {
			found = true
			if isAdmin, ok := user["is_admin"].(bool); ok {
				foundIsAdmin = isAdmin
			}
			break
		}
	}

	assert.True(t, found, "test_user_check should be in users list")
	assert.True(t, foundIsAdmin, "test_user_check should be admin")
}