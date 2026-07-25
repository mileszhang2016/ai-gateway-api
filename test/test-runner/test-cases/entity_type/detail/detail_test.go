package detail

import (
	"os"
	"testing"

	"github.com/yf-networks/ai-gateway-api/test-runner/testutil"
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

func TestDetailEntityType_Normal_Success(t *testing.T) {
	// ET-3-001: 查询已存在的Entity-Type
	client := testutil.GetClient()

	// 创建Entity-Type
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name": "test_type_detail",
		"level":     1,
	})
	if err != nil {
		t.Fatalf("create entity-type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 查询该Entity-Type
	resp, err = client.Get("/open-api/v1/entity-types/test_type_detail")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldEquals(t, resp, "type_name", "test_type_detail")
	testutil.AssertDataFieldEquals(t, resp, "level", float64(1))
	testutil.AssertDataFieldNotEmpty(t, resp, "create_time")
}

func TestDetailEntityType_Abnormal_NotFound(t *testing.T) {
	// ET-3-002: 查询不存在的Entity-Type
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/entity-types/non_existent_type")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 404)
}