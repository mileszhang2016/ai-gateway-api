package unbind_product

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

func TestUnbindProduct_Normal_Success(t *testing.T) {
	// AUTH-7-001: 正常解除绑定
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_unbind",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 先绑定产品线
	resp, err = client.Post("/open-api/v1/auth/users/test_user_unbind/products/product_demo", nil)
	if err != nil {
		t.Fatalf("bind product failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 解除绑定
	resp, err = client.Delete("/open-api/v1/auth/users/test_user_unbind/products/product_demo")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestUnbindProduct_Abnormal_NotBound(t *testing.T) {
	// AUTH-7-002: 解除未绑定的产品线
	client := testutil.GetClient()

	// 创建用户（不绑定产品线）
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_unbind_none",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 解除未绑定的产品线
	resp, err = client.Delete("/open-api/v1/auth/users/test_user_unbind_none/products/product_unbind")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}