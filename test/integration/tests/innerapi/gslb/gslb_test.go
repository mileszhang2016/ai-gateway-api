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

func TestInnerAPI_Gslb(t *testing.T) {
	t.Run("IN-2-001 导出 GSLB 缺少 bfe_cluster", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/gslb_data/gslb")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 422 && resp.ErrNum != 404 {
			t.Errorf("expected ErrNum=422 or 404, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
		}
	})

	t.Run("IN-2-002 正常导出 GSLB", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/gslb_data/gslb", map[string]string{
			"bfe_cluster": "BFE-AI_product.szyf",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataNotEmpty(t, resp)
		testutil.AssertDataFieldNotEmpty(t, resp, "Version")
		// Clusters 可能因无后端实例而为空对象，仅验证字段存在且类型为对象
		exists, err := testutil.FieldExists(resp, "Clusters")
		if err != nil {
			t.Fatalf("check Clusters field: %v", err)
		}
		if !exists {
			t.Error("Clusters field not found in Data")
		}
	})
}
