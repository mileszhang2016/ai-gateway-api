package list

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

	// 预先创建几个Entity-Type用于测试列表查询
	createEntityType("test_type_list1", "列表测试类型1", 1)
	createEntityType("test_type_list2", "列表测试类型2", 2)
	createEntityType("test_type_list3", "列表测试类型3", 3)

	code := m.Run()

	sm.Shutdown()
	os.Exit(code)
}

func createEntityType(typeName, description string, level int) {
	client := testutil.GetClient()
	resp, err := client.Post("/open-api/v1/entity-types", map[string]interface{}{
		"type_name":   typeName,
		"description": description,
		"level":       level,
	})
	if err != nil {
		panic("failed to create entity-type " + typeName + ": " + err.Error())
	}
	if resp.ErrNum != 200 && resp.ErrNum != 555 {
		panic("failed to create entity-type " + typeName + ": " + resp.ErrMsg)
	}
}

func TestListEntityType_Normal_DefaultParams(t *testing.T) {
	// ET-2-001: 获取Entity-Type列表（默认参数）
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/entity-types")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldNotEmpty(t, resp, "list")
	// 当前列表接口不返回 pagination 字段
}

func TestListEntityType_Normal_CustomPagination(t *testing.T) {
	// ET-2-002: 获取Entity-Type列表（自定义分页）
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/entity-types?page=1&page_size=2")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	testutil.AssertDataFieldNotEmpty(t, resp, "list")
	// 当前列表接口不返回 pagination 字段
}
