package search_by_product

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

func TestSearchByProduct_Normal_HasUsers(t *testing.T) {
	// AUTH-8-001: 查询有绑定用户的产品线
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_search",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 绑定产品线
	resp, err = client.Post("/open-api/v1/auth/users/test_user_search/products/product_search", nil)
	if err != nil {
		t.Fatalf("bind product failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 按产品线查询用户
	resp, err = client.Get("/open-api/v1/auth/users/actions/search-by-product/product_search")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回结果
	var users []interface{}
	if err := json.Unmarshal(resp.Data, &users); err != nil {
		t.Fatalf("unmarshal users failed: %v", err)
	}
	assert.NotEmpty(t, users, "users list should not be empty")

	found := false
	for _, u := range users {
		user, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		if userName, ok := user["user_name"].(string); ok && userName == "test_user_search" {
			found = true
			if isAdmin, ok := user["is_admin"].(bool); ok {
				assert.False(t, isAdmin, "test_user_search should not be admin")
			}
			break
		}
	}
	assert.True(t, found, "test_user_search should be in users list")
}

func TestSearchByProduct_Normal_NoUsers(t *testing.T) {
	// AUTH-8-002: 查询无绑定用户的产品线
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/auth/users/actions/search-by-product/product_empty")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证返回空数组
	var users []interface{}
	if err := json.Unmarshal(resp.Data, &users); err != nil {
		t.Fatalf("unmarshal users failed: %v", err)
	}
	assert.Empty(t, users, "users list should be empty")
}