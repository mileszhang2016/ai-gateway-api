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

func TestAPIKey_PartialUpdate(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("partial-update-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("AK-5-001 部分更新 enabled", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{"enabled": false})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "enabled", false)
	})

	t.Run("AK-5-002 部分更新 route_rules", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"route_rules": map[string]interface{}{
				"enabled": true,
				"rules": []interface{}{
					map[string]interface{}{
						"name":  "default",
						"Cond":  "default_t()",
						"targets": []interface{}{map[string]interface{}{"ClusterName": "cluster1", "Weight": 100}},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		rr := data["route_rules"].(map[string]interface{})
		assert.Equal(t, true, rr["enabled"])
	})

	t.Run("AK-5-003 部分更新后查询一致性", func(t *testing.T) {
		_, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"models": []string{"gpt-3.5-turbo"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp, err := testutil.GetClient().Get("/open-api/v1/api-keys/" + apiKeyID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "models", []interface{}{"gpt-3.5-turbo"})
	})

	t.Run("AK-5-004 部分更新非法 rate_limit_policy（window_minutes 为 0）", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
			"rate_limit_policy": map[string]interface{}{
				"enabled": true,
				"rules": map[string]interface{}{
					"tpm": []interface{}{
						map[string]interface{}{"name": "t1", "model": "*", "window_minutes": 0, "max_tokens": 100, "step_minutes": 1},
					},
				},
			},
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
