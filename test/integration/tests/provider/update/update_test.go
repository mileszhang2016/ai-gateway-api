package provider_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
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

func TestProvider_Update(t *testing.T) {
	providerName := testutil.UniqueProviderName()
	if _, err := testutil.CreateProvider(providerName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("PV-4-001 更新 description", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
			"description":     "更新后的 Provider 描述",
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "description", "更新后的 Provider 描述")
	})

	t.Run("PV-4-002 更新 models", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
			"models":          []string{"deepseek-chat", "deepseek-coder", "deepseek-reasoner"},
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		models, _ := data["models"].([]interface{})
		assert.Len(t, models, 3)
	})

	t.Run("PV-4-003 更新 keys（全量替换）", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
			"keys": []interface{}{
				map[string]interface{}{
					"name": "key-primary",
					"key":  "sk-new-primary",
				},
				map[string]interface{}{
					"name": "key-tertiary",
					"key":  "sk-cccccccccccc",
				},
			},
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		keys, _ := data["keys"].([]interface{})
		assert.Len(t, keys, 2)
	})

	t.Run("PV-4-004 更新不存在的 Provider", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/non_existent_provider", map[string]interface{}{
			"description":     "x",
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("PV-4-005 请求体不包含 name", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
			"description":     "请求体未传 name",
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataFieldEquals(t, resp, "name", providerName)
	})

	t.Run("PV-4-006 请求体包含 name", func(t *testing.T) {
		resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
			"name":            providerName,
			"description":     "请求体传了 name",
			"instance_pool":   []interface{}{map[string]interface{}{"name": "backend-1", "addr": "10.0.0.1", "weight": 100, "port": 8080}},
			"model_protocols": []string{"openai"},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Cleanup(func() {
		testutil.DeleteProvider(providerName)
	})
}
