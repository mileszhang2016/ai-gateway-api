package delete_token

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

func TestDeleteToken_Normal_Success(t *testing.T) {
	// AUTH-12-001: 正常删除
	client := testutil.GetClient()

	// 创建 Token
	resp, err := client.Post("/open-api/v1/auth/tokens", map[string]interface{}{
		"name":         "test_token_del",
		"scope":        "Product",
		"product_name": "product_token_del",
	})
	if err != nil {
		t.Fatalf("create token failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 删除 Token
	resp, err = client.Delete("/open-api/v1/auth/tokens/test_token_del")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestDeleteToken_Abnormal_NotFound(t *testing.T) {
	// AUTH-12-002: 删除不存在的 Token
	client := testutil.GetClient()
	resp, err := client.Delete("/open-api/v1/auth/tokens/non_existent_token")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}