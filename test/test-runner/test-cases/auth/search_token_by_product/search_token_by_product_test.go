package search_token_by_product

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

func TestSearchTokenByProduct_Normal_HasTokens(t *testing.T) {
	// AUTH-15-001: 查询有 Token 绑定的产品线
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_search",
		"scope":        "Product",
		"product_name": "product_token_search",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 按产品线查询 Token
	resp, err = client.Get("/open-api/v1/auth/tokens/actions/search-by-product/product_token_search")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回结果
	var tokens []interface{}
	if err := json.Unmarshal(resp.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens failed: %v", err)
	}
	assert.NotEmpty(t, tokens, "tokens list should not be empty")

	found := false
	for _, tkn := range tokens {
		token, ok := tkn.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := token["name"].(string); ok && name == "test_token_search" {
			found = true
			if scope, ok := token["scope"].(string); ok {
				assert.Equal(t, "Product", scope, "scope should be Product")
			}
			break
		}
	}
	assert.True(t, found, "test_token_search should be in tokens list")
}

func TestSearchTokenByProduct_Normal_NoTokens(t *testing.T) {
	// AUTH-15-002: 查询无 Token 绑定的产品线
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/auth/tokens/actions/search-by-product/product_token_empty")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 验证返回空数组
	var tokens []interface{}
	if err := json.Unmarshal(resp.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens failed: %v", err)
	}
	assert.Empty(t, tokens, "tokens list should be empty")
}