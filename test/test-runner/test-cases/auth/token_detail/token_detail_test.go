package token_detail

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

func TestTokenDetail_Normal_Success(t *testing.T) {
	// AUTH-13-001: 查询已存在的 Token
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_detail",
		"scope":        "Product",
		"product_name": "product_token_detail",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 查询 Token 详情
	resp, err = client.Get("/open-api/v1/auth/tokens/test_token_detail")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldEquals(t, resp, "name", "test_token_detail")
	testutil.AssertDataFieldEquals(t, resp, "product_name", "product_token_detail")
	testutil.AssertDataFieldEquals(t, resp, "scope", "Product")
	testutil.AssertDataFieldNotEmpty(t, resp, "token")
}

func TestTokenDetail_Abnormal_NotFound(t *testing.T) {
	// AUTH-13-002: 查询不存在的 Token
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/auth/tokens/non_existent_token_detail")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}