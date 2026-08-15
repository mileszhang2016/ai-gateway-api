package model_price_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestModelPrice_Update(t *testing.T) {
	t.Run("MP-6-001 按 id 部分更新 prices", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		model := "update-model"
		id, err := testutil.CreateModelPrice(map[string]interface{}{
			"provider":   provider,
			"model":      model,
			"base_model": model,
			"mode":       "chat",
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteModelPrice(id)

		resp, err := testutil.GetClient().Put("/open-api/v1/model-prices/"+fmt.Sprintf("%d", id), map[string]interface{}{
			"prices": map[string]interface{}{
				"input_cost_per_token":  0.00001,
				"output_cost_per_token": 0.00002,
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.Equal(t, provider, data["provider"])
		prices := data["prices"].(map[string]interface{})
		assert.InDelta(t, 0.00001, prices["input_cost_per_token"], 0.0000001)
		assert.InDelta(t, 0.00002, prices["output_cost_per_token"], 0.0000001)

		getResp, err := testutil.GetClient().Get("/open-api/v1/model-prices/" + fmt.Sprintf("%d", id))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		testutil.AssertSuccess(t, getResp)
		testutil.AssertDataFieldEquals(t, getResp, "provider", provider)
	})

	t.Run("MP-6-002 按 id 更新不存在的记录", func(t *testing.T) {
		resp, err := testutil.GetClient().Put("/open-api/v1/model-prices/999999999", map[string]interface{}{
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("MP-6-003 按 id 更新为非法 mode", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		model := "update-invalid"
		id, err := testutil.CreateModelPrice(map[string]interface{}{
			"provider":   provider,
			"model":      model,
			"base_model": model,
			"mode":       "chat",
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteModelPrice(id)

		resp, err := testutil.GetClient().Put("/open-api/v1/model-prices/"+fmt.Sprintf("%d", id), map[string]interface{}{
			"mode": "invalid_mode",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("MP-7-001 按组合键更新 prices", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		model := "update-query-model"
		id, err := testutil.CreateModelPrice(map[string]interface{}{
			"provider":   provider,
			"model":      model,
			"base_model": model,
			"mode":       "chat",
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		defer testutil.DeleteModelPrice(id)

		resp, err := testutil.GetClient().PutWithQuery("/open-api/v1/model-prices", map[string]string{
			"provider": provider,
			"model":    model,
			"mode":     "chat",
		}, map[string]interface{}{
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.00005,
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var data map[string]interface{}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		prices := data["prices"].(map[string]interface{})
		assert.InDelta(t, 0.00005, prices["input_cost_per_token"], 0.0000001)
	})

	t.Run("MP-7-002 按组合键更新缺少 query 参数", func(t *testing.T) {
		resp, err := testutil.GetClient().PutWithQuery("/open-api/v1/model-prices", map[string]string{
			"provider": testutil.UniqueName("provider"),
			"model":    "m",
		}, map[string]interface{}{
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})
}
