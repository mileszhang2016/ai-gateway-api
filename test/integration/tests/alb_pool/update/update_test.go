package update

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

func TestUpdateAlbPool_Normal_Success(t *testing.T) {
	// ALB-2-001: 更新实例列表
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/alb-pool", map[string]interface{}{
		"instances": []interface{}{
			map[string]interface{}{
				"hostname": "127.0.0.1",
				"ip":       "127.0.0.1",
				"weight":   1,
				"ports": map[string]interface{}{
					"Default": 8080,
				},
				"tags": map[string]interface{}{
					"key": "value",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldNotEmpty(t, resp, "name")
}

func TestUpdateAlbPool_Abnormal_EmptyList(t *testing.T) {
	// ALB-2-002: 更新为空列表（异常参数）
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/alb-pool", map[string]interface{}{
		"instances": []interface{}{},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestUpdateAlbPool_Required_MissingInstances(t *testing.T) {
	// ALB-2-003: 缺少 instances
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/alb-pool", map[string]interface{}{})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestUpdateAlbPool_Required_MissingInstanceField(t *testing.T) {
	// ALB-2-004: 缺少实例必填字段（缺少 ip）
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/alb-pool", map[string]interface{}{
		"instances": []interface{}{
			map[string]interface{}{
				"hostname": "127.0.0.1",
				"weight":   1,
				"ports": map[string]interface{}{
					"Default": 8080,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

func TestUpdateAlbPool_Boundary_WeightOutOfRange(t *testing.T) {
	// ALB-2-005: 实例权重超出范围（101）
	client := testutil.GetClient()
	resp, err := client.Patch("/open-api/v1/alb-pool", map[string]interface{}{
		"instances": []interface{}{
			map[string]interface{}{
				"hostname": "127.0.0.1",
				"ip":       "127.0.0.1",
				"weight":   101,
				"ports": map[string]interface{}{
					"Default": 8080,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}
