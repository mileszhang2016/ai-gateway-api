package innerapi_test

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

func TestInnerAPI_ModBodyProcess(t *testing.T) {
	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-body-process")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "Version")
	testutil.AssertDataFieldNotEmpty(t, resp, "Config")
}
