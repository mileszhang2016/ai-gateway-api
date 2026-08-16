package certificate_test

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

func TestCertificate_Detail(t *testing.T) {
	defaultCert := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(defaultCert, true); err != nil {
		t.Fatalf("setup default cert failed: %v", err)
	}
	certName := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(certName, false); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("CERT-3-001 查询已存在证书", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/certificates/" + certName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "cert_name", certName)
	})

	t.Run("CERT-3-002 查询不存在的证书", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/certificates/non_existent_cert")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Cleanup(func() {
		testutil.DeleteCertificate(certName)
		testutil.DeleteCertificate(defaultCert)
	})
}
