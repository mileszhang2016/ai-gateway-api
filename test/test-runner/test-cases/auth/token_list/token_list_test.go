package token_list

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestTokenList_Normal_GetList(t *testing.T) {
	// AUTH-14-001: 获取 Token 列表
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_list",
		"scope":        "Product",
		"product_name": "product_token_list",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取 Token 列表
	resp, err = client.Get("/open-api/v1/auth/tokens")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回数组
	var tokens []interface{}
	if err := json.Unmarshal(resp.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens failed: %v", err)
	}
	assert.NotEmpty(t, tokens, "tokens list should not be empty")
}

func TestTokenList_Data_FieldCompleteness(t *testing.T) {
	// AUTH-14-002: 验证返回字段完整性
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_list2",
		"scope":        "Product",
		"product_name": "product_token_list2",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 获取 Token 列表
	resp, err = client.Get("/open-api/v1/auth/tokens")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证每个 Token 包含必要字段
	var tokens []interface{}
	if err := json.Unmarshal(resp.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens failed: %v", err)
	}

	for _, tkn := range tokens {
		token, ok := tkn.(map[string]interface{})
		if !ok {
			continue
		}
		assert.Contains(t, token, "name", "token should have name")
		assert.Contains(t, token, "token", "token should have token")
		assert.Contains(t, token, "scope", "token should have scope")
	}
}