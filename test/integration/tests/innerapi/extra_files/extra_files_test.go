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

func TestInnerAPI_ExtraFiles(t *testing.T) {
	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/extra_files/name_conf.data")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// 该文件可能不存在，接受 200 或 404
	if resp.ErrNum != 200 && resp.ErrNum != 404 {
		t.Errorf("expected ErrNum=200 or 404, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
	}
}
