package api_key_test

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

func TestAPIKey_FullUpdate(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("full-update-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	// get original key
	detailResp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	originalKey, _ := testutil.GetDataField(detailResp, "key")

	t.Run("AK-4-001 全量更新 quota_plan 触发余额重置", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"description": "test-key-updated",
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota": 500000,
				"unit": "total_token",
				"reset_period": "monthly",
			},
			"rate_limit_policy": map[string]interface{}{"enabled": false, "rules": map[string]interface{}{}},
			"route_rules": map[string]interface{}{"enabled": false, "rules": []interface{}{}},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "description", "test-key-updated")
		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		qp := data["quota_plan"].(map[string]interface{})
		assert.Equal(t, float64(500000), qp["quota"])
	})

	t.Run("AK-4-002 全量更新传入 key 被忽略", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"key":         "new-key",
			"description": "test-key-ignore-key",
			"quota_plan":  map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{"enabled": false, "rules": map[string]interface{}{}},
			"route_rules": map[string]interface{}{"enabled": false, "rules": []interface{}{}},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "key", originalKey)
		testutil.AssertDataFieldEquals(t, resp, "description", "test-key-ignore-key")
	})

	t.Run("AK-4-003 全量更新后查询一致性", func(t *testing.T) {
		_, err := testutil.GetClient().Put("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"description": "test-key-consistency",
			"enabled": false,
			"quota_plan":  map[string]interface{}{"unlimited": true},
			"rate_limit_policy": map[string]interface{}{"enabled": false, "rules": map[string]interface{}{}},
			"route_rules": map[string]interface{}{"enabled": false, "rules": []interface{}{}},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "description", "test-key-consistency")
		testutil.AssertDataFieldEquals(t, resp, "enabled", false)
	})

	t.Run("AK-4-004 全量更新非法 quota_plan unit", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"description": "test-key-bad-unit-update",
			"quota_plan": map[string]interface{}{
				"unlimited": false,
				"quota":     100,
				"unit":      "invalid_unit",
			},
			"rate_limit_policy": map[string]interface{}{"enabled": false, "rules": map[string]interface{}{}},
			"route_rules":     map[string]interface{}{"enabled": false, "rules": []interface{}{}},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
