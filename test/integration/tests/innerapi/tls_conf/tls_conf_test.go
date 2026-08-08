package innerapi_test

import (
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestInnerAPI_TlsConf(t *testing.T) {
	// 创建默认证书以确保证书配置非空
	certName := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(certName, true); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "Version")
	testutil.AssertDataFieldNotEmpty(t, resp, "HostTable")
	testutil.AssertDataFieldNotEmpty(t, resp, "RouteTable")
	testutil.AssertDataFieldNotEmpty(t, resp, "ClusterConf")

	t.Cleanup(func() {
		testutil.DeleteCertificate(certName)
	})
}
