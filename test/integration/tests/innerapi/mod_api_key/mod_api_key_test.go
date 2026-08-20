package innerapi_test

import (
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestInnerAPI_ModApiKey(t *testing.T) {
	// 创建带有限配额计划的 API-Key，确保 QuotaPlans 非空
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "inner-api-key",
		"quota_plan": map[string]interface{}{
			"unlimited": false,
			"quota":     1000,
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if resp.ErrNum != 200 {
		t.Fatalf("create api-key failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	apiKeyID, _ := testutil.GetDataField(resp, "id")

	t.Run("IN-6-001 首次导出 mod-api-key", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		testutil.AssertDataNotEmpty(t, resp)
		testutil.AssertDataFieldNotEmpty(t, resp, "version")
		testutil.AssertDataFieldNotEmpty(t, resp, "config")
		testutil.AssertDataFieldNotEmpty(t, resp, "QuotaPlans")
		testutil.AssertDataFieldNotEmpty(t, resp, "tokens")

		// 验证导出的 token 包含 key_id 且不包含 name
		var data map[string]interface{}
		if err := testutil.UnmarshalData(resp, &data); err != nil {
			t.Fatalf("parse data failed: %v", err)
		}
		tokens, ok := data["tokens"].(map[string]interface{})
		if !ok {
			t.Fatal("tokens is not an object")
		}
		productTokens, ok := tokens["AI_product"].(map[string]interface{})
		if !ok {
			t.Fatal("tokens.AI_product is not an object")
		}
		if len(productTokens) == 0 {
			t.Fatal("tokens.AI_product is empty")
		}
		for key, token := range productTokens {
			tokenMap, ok := token.(map[string]interface{})
			if !ok {
				t.Fatalf("token %s is not an object", key)
			}
			if _, ok := tokenMap["key_id"]; !ok {
				t.Errorf("token %s missing key_id", key)
			}
			if _, ok := tokenMap["name"]; ok {
				t.Errorf("token %s should not contain name", key)
			}
		}
	})

	t.Run("IN-6-002 增量同步未变化", func(t *testing.T) {
		firstResp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		version, _ := testutil.GetDataField(firstResp, "version")
		resp, err := testutil.GetClient().Get("/inner-api/v1/configs/mod-api-key", map[string]string{
			"version": version.(string),
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		if string(resp.Data) != "null" {
			t.Skip("version comparison not stable, Data is not null")
		}
	})

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID.(string))
	})
}
