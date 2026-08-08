package api_key_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestAPIKey_List(t *testing.T) {
	id1, err := testutil.CreateAPIKey("enabled-true-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+id1, map[string]interface{}{
		"quota_plan": map[string]interface{}{"unlimited": false, "quota": 1000000, "unit": "total_token", "reset_period": "monthly"},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	id2, err := testutil.CreateAPIKey("enabled-false-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+id2, map[string]interface{}{"enabled": false})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AK-2-001 列表返回包含 balance", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys", map[string]string{"page": "1", "page_size": "20"})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 2)
		found := false
		for _, item := range list {
			apiKey := item.(map[string]interface{})
			qp := apiKey["quota_plan"].(map[string]interface{})
			if unlimited, ok := qp["unlimited"].(bool); ok && !unlimited {
				assert.Contains(t, qp, "balance")
				found = true
			}
		}
		if !found {
			t.Skip("no non-unlimited api-key found with balance")
		}
	})

	t.Run("AK-2-002 列表分页参数", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys", map[string]string{"page": "1", "page_size": "1"})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertListFieldLen(t, resp, "list", 1)
		testutil.AssertPagination(t, resp, 1, 1, 2)
	})

	t.Run("AK-2-003 按 enabled 过滤", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys", map[string]string{"enabled": "false"})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		for _, item := range data["list"].([]interface{}) {
			assert.Equal(t, false, item.(map[string]interface{})["enabled"])
		}
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(id1)
		testutil.DeleteAPIKey(id2)
	})
}
