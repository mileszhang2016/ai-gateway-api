package certificate_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func assertCertMeta(t *testing.T, resp *testutil.APIResponse, certName string) {
	testutil.AssertDataFieldEquals(t, resp, "cert_name", certName)
	exists, _ := testutil.FieldExists(resp, "cert_file_content")
	assert.False(t, exists, "should not return cert_file_content")
	exists, _ = testutil.FieldExists(resp, "key_file_content")
	assert.False(t, exists, "should not return key_file_content")
}

func TestCertificate_Create(t *testing.T) {
	defaultCert := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(defaultCert, true); err != nil {
		t.Fatalf("setup default cert failed: %v", err)
	}

	cert1 := testutil.UniqueCertName()
	certDefault := testutil.UniqueCertName()

	t.Run("CERT-1-001 创建非默认证书", func(t *testing.T) {
		resp, err := testutil.CreateCertificate(cert1, false)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		_ = resp
		getResp, err := testutil.GetClient().Get("/open-api/v1/certificates/" + cert1)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		testutil.AssertSuccess(t, getResp)
		assertCertMeta(t, getResp, cert1)
		testutil.AssertDataFieldEquals(t, getResp, "is_default", false)
	})

	t.Run("CERT-1-002 创建默认证书", func(t *testing.T) {
		resp, err := testutil.CreateCertificate(certDefault, true)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		_ = resp
		getResp, err := testutil.GetClient().Get("/open-api/v1/certificates/" + certDefault)
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		testutil.AssertSuccess(t, getResp)
		testutil.AssertDataFieldEquals(t, getResp, "is_default", true)
	})

	t.Run("CERT-1-003 证书与密钥不匹配", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/certificates", map[string]interface{}{
			"cert_name":         testutil.UniqueCertName(),
			"description":       "不匹配证书",
			"is_default":        false,
			"cert_file_content": "-----BEGIN CERTIFICATE-----INVALID-----END CERTIFICATE-----",
			"key_file_content":  "-----BEGIN RSA PRIVATE KEY-----INVALID-----END RSA PRIVATE KEY-----",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 422 && resp.ErrNum != 500 {
			t.Errorf("expected ErrNum=422 or 500, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
	})

	t.Cleanup(func() {
		testutil.DeleteCertificate(cert1)
		testutil.DeleteCertificate(certDefault)
		testutil.DeleteCertificate(defaultCert)
	})
}
