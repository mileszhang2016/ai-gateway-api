package get

import (
	"encoding/json"
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

func TestGetAlbPool_Normal_GetDetail(t *testing.T) {
	// ALB-1-001: 获取实例池详情
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/alb-pool")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)

	// 验证返回字段
	testutil.AssertDataFieldNotEmpty(t, resp, "name")
}

func TestGetAlbPool_Data_FieldCompleteness(t *testing.T) {
	// ALB-1-002: 验证返回字段完整性
	client := testutil.GetClient()
	resp, err := client.Get("/open-api/v1/alb-pool")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 解析实例列表
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data failed: %v", err)
	}

	instances, ok := data["instances"].([]interface{})
	if !ok || len(instances) == 0 {
		t.Skip("no instances in pool, skipping field completeness test")
	}

	// 验证每个实例包含必要字段
	for _, inst := range instances {
		instance, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}
		assert.Contains(t, instance, "hostname", "instance should have hostname")
		assert.Contains(t, instance, "ip", "instance should have ip")
		assert.Contains(t, instance, "weight", "instance should have weight")
		assert.Contains(t, instance, "ports", "instance should have ports")

		// 验证 ports 包含 Default
		ports, ok := instance["ports"].(map[string]interface{})
		if ok {
			assert.Contains(t, ports, "Default", "ports should contain Default")
		}
	}
}
