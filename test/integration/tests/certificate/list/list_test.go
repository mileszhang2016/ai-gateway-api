package certificate_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCertificate_List(t *testing.T) {
	defaultCert := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(defaultCert, true); err != nil {
		t.Fatalf("setup default cert failed: %v", err)
	}
	certName := testutil.UniqueCertName()
	if _, err := testutil.CreateCertificate(certName, false); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/open-api/v1/certificates")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	var list []interface{}
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	assert.GreaterOrEqual(t, len(list), 1)
	for _, item := range list {
		cert := item.(map[string]interface{})
		assert.NotEmpty(t, cert["cert_name"])
		assert.NotContains(t, cert, "cert_file_content")
		assert.NotContains(t, cert, "key_file_content")
	}

	t.Cleanup(func() {
		testutil.DeleteCertificate(certName)
		testutil.DeleteCertificate(defaultCert)
	})
}
