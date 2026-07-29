package alb_pool_test

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

func TestAlbPool_Get(t *testing.T) {
	resp, err := testutil.GetClient().Get("/open-api/v1/alb-pool")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "name")

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data failed: %v", err)
	}

	instances, ok := data["instances"].([]interface{})
	if !ok {
		t.Fatal("instances should be an array")
	}

	for _, inst := range instances {
		instance, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}
		assert.Contains(t, instance, "hostname")
		assert.Contains(t, instance, "ip")
		assert.Contains(t, instance, "weight")
		assert.Contains(t, instance, "ports")
		assert.NotContains(t, instance, "tags", "instance should not contain tags in v0.3.0")
		if ports, ok := instance["ports"].(map[string]interface{}); ok {
			assert.Contains(t, ports, "Default")
		}
		if w, ok := instance["weight"].(float64); ok {
			assert.True(t, w >= 0 && w <= 100, "weight should be in [0,100]")
		}
	}
}
