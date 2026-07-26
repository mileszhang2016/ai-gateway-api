package bind_product

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

func TestBindProduct_Normal_Success(t *testing.T) {
	// AUTH-6-001: 正常绑定产品线
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_bind",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 绑定产品线
	resp, err = client.Post("/open-api/v1/auth/users/test_user_bind/products/product_demo", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
}

func TestBindProduct_Abnormal_ProductNotFound(t *testing.T) {
	// AUTH-6-002: 绑定不存在的产品线
	client := testutil.GetClient()

	// 创建用户
	resp, err := client.Post("/open-api/v1/auth/users", map[string]interface{}{
		"user_name": "test_user_bind_none",
		"password":  "password@123",
		"is_admin":  false,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 绑定不存在的产品线
	resp, err = client.Post("/open-api/v1/auth/users/test_user_bind_none/products/non_existent_product", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}

func TestBindProduct_Abnormal_UserNotFound(t *testing.T) {
	// AUTH-6-003: 为不存在的用户绑定产品线
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/auth/users/non_existent_user_bind/products/product_demo", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}