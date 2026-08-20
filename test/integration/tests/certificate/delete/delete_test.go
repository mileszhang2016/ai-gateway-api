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

func TestCertificate_Delete(t *testing.T) {
	defaultCert := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(defaultCert, true); err != nil {
		t.Fatalf("setup default cert failed: %v", err)
	}

	t.Run("CERT-6-001 删除非默认证书", func(t *testing.T) {
		certName := testutil.UniqueCertName()
		if _, err := testutil.CreateCertificate(certName, false); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/certificates/" + certName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		resp, _ = testutil.GetClient().Get("/open-api/v1/certificates/" + certName)
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("CERT-6-002 删除默认证书", func(t *testing.T) {
		certName := testutil.UniqueCertName()
		if _, err := testutil.CreateCertificate(certName, true); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/certificates/" + certName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 409 && resp.ErrNum != 422 && resp.ErrNum != 500 {
			t.Errorf("expected ErrNum=409/422/500, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
		t.Cleanup(func() {
			testutil.DeleteCertificate(certName)
		})
	})

	t.Cleanup(func() {
		testutil.DeleteCertificate(defaultCert)
	})
}
