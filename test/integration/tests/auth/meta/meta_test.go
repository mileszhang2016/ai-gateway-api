package auth_test

import (
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestAuth_Meta(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/meta")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "nav")
	// icon/logo 在测试配置中可能为空，仅检查字段存在
	exists, _ := testutil.FieldExists(resp, "icon")
	if !exists {
		t.Error("expected icon field to exist")
	}
	exists, _ = testutil.FieldExists(resp, "logo")
	if !exists {
		t.Error("expected logo field to exist")
	}
}
