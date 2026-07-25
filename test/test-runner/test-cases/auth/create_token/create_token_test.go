package create_token

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

func TestCreateToken_Normal_ProductScope(t *testing.T) {
	// AUTH-11-001: 创建 Product scope Token
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_001",
		"scope":        "Product",
		"product_name": "product_token",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "token")
}

func TestCreateToken_Required_MissingName(t *testing.T) {
	// AUTH-11-002: 缺少 name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"scope":        "Product",
		"product_name": "product_token",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateToken_Required_MissingScope(t *testing.T) {
	// AUTH-11-003: 缺少 scope
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_002",
		"product_name": "product_token",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateToken_Required_MissingProductName(t *testing.T) {
	// AUTH-11-004: scope=Product 缺少 product_name
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":  "test_token_003",
		"scope": "Product",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestCreateToken_Business_DuplicateToken(t *testing.T) {
	// AUTH-11-005: 重复创建同名 Token
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_dup",
		"scope":        "Product",
		"product_name": "product_token",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 再次创建同名 Token
	resp, err = client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_dup",
		"scope":        "Product",
		"product_name": "product_token",
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 555)
}