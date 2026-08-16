package model_price_test

import (
	"encoding/json"
	"fmt"
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

type pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type listResponse struct {
	List       []modelPriceDTO `json:"list"`
	Pagination pagination      `json:"pagination"`
}

type modelPriceDTO struct {
	ID        int64              `json:"id"`
	Provider  string             `json:"provider"`
	Model     string             `json:"model"`
	BaseModel string             `json:"base_model"`
	Mode      string             `json:"mode"`
	Prices    map[string]float64 `json:"prices"`
}

func TestModelPrice_Import(t *testing.T) {
	t.Run("MP-1-001 replace 模式全量替换", func(t *testing.T) {
		// 先创建一条旧记录
		oldProvider := testutil.UniqueName("provider")
		_, err := testutil.CreateModelPrice(map[string]interface{}{
			"provider":   oldProvider,
			"model":      "old-model",
			"base_model": "old-model",
			"mode":       "chat",
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.001,
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		provider1 := testutil.UniqueName("provider")
		provider2 := testutil.UniqueName("provider")
		yaml := buildYAML([]map[string]interface{}{
			{
				"provider": provider1,
				"model":    "gpt-4o",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0001,
				},
			},
			{
				"provider": provider2,
				"model":    "deepseek-v3",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.000002,
				},
			},
		})

		result, resp, err := testutil.ImportModelPricesWithResult([]byte(yaml), "replace")
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		assert.Equal(t, 2, result.ImportedCount)
		assert.Equal(t, 0, result.SkippedCount)

		list, err := fetchModelPriceList()
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		assert.Equal(t, int64(2), list.Pagination.Total)

		providers := map[string]bool{}
		for _, item := range list.List {
			providers[item.Provider] = true
		}
		assert.True(t, providers[provider1], "expected provider1 in list")
		assert.True(t, providers[provider2], "expected provider2 in list")
		assert.False(t, providers[oldProvider], "expected old record removed")

		// cleanup via replace with empty
		_, _, _ = testutil.ImportModelPricesWithResult([]byte(buildYAML(nil)), "replace")
	})

	t.Run("MP-1-002 merge 模式增量合并", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		model := "gpt-4o"
		id, err := testutil.CreateModelPrice(map[string]interface{}{
			"provider":   provider,
			"model":      model,
			"base_model": model,
			"mode":       "chat",
			"prices": map[string]interface{}{
				"input_cost_per_token": 0.0001,
			},
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		newProvider := testutil.UniqueName("provider")
		yaml := buildYAML([]map[string]interface{}{
			{
				"provider": provider,
				"model":    model,
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0002,
				},
			},
			{
				"provider": newProvider,
				"model":    "new-model",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0003,
				},
			},
		})

		result, resp, err := testutil.ImportModelPricesWithResult([]byte(yaml), "merge")
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		assert.Equal(t, 2, result.ImportedCount)

		one, err := testutil.GetClient().Get("/open-api/v1/model-prices/" + fmt.Sprintf("%d", id))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		testutil.AssertSuccess(t, one)

		var data map[string]interface{}
		if err := json.Unmarshal(one.Data, &data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		prices := data["prices"].(map[string]interface{})
		assert.InDelta(t, 0.0002, prices["input_cost_per_token"], 0.0000001)

		list, err := fetchModelPriceList()
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		assert.GreaterOrEqual(t, list.Pagination.Total, int64(2))

		_, _, _ = testutil.ImportModelPricesWithResult([]byte(buildYAML(nil)), "replace")
	})

	t.Run("MP-1-003 默认 replace 模式", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		yaml := buildYAML([]map[string]interface{}{
			{
				"provider": provider,
				"model":    "default-model",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0001,
				},
			},
		})

		resp, err := testutil.GetClient().PostMultipartFile("/open-api/v1/model-prices/import", "file", "model-list.yaml", []byte(yaml), nil)
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var result testutil.ImportModelPricesResult
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			t.Fatalf("unmarshal result: %v", err)
		}
		assert.Equal(t, 1, result.ImportedCount)

		list, err := fetchModelPriceList()
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		assert.Equal(t, int64(1), list.Pagination.Total)

		_, _, _ = testutil.ImportModelPricesWithResult([]byte(buildYAML(nil)), "replace")
	})

	t.Run("MP-1-004 非法 mode", func(t *testing.T) {
		yaml := buildYAML([]map[string]interface{}{
			{
				"provider": testutil.UniqueName("provider"),
				"model":    "m",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0001,
				},
			},
		})

		resp, err := testutil.GetClient().PostMultipartFile("/open-api/v1/model-prices/import", "file", "model-list.yaml", []byte(yaml), map[string]string{
			"mode": "invalid",
		})
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("MP-1-005 非法 YAML", func(t *testing.T) {
		resp, err := testutil.GetClient().PostMultipartFile("/open-api/v1/model-prices/import", "file", "model-list.yaml", []byte("not a yaml: ["), map[string]string{
			"mode": "replace",
		})
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("MP-1-006 重复三元组", func(t *testing.T) {
		provider := testutil.UniqueName("provider")
		yaml := buildYAML([]map[string]interface{}{
			{
				"provider": provider,
				"model":    "dup-model",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0001,
				},
			},
			{
				"provider": provider,
				"model":    "dup-model",
				"mode":     "chat",
				"prices": map[string]float64{
					"input_cost_per_token": 0.0002,
				},
			},
		})

		resp, err := testutil.GetClient().PostMultipartFile("/open-api/v1/model-prices/import", "file", "model-list.yaml", []byte(yaml), map[string]string{
			"mode": "replace",
		})
		if err != nil {
			t.Fatalf("import failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})
}

func buildYAML(models []map[string]interface{}) string {
	yaml := "version: v1.0\ndefault_currency: RMB\n"
	if len(models) == 0 {
		yaml += "models: []\n"
		return yaml
	}
	yaml += "models:\n"
	for _, m := range models {
		yaml += fmt.Sprintf("  - provider: %s\n", m["provider"])
		yaml += fmt.Sprintf("    model: %s\n", m["model"])
		yaml += fmt.Sprintf("    base_model: %s\n", m["model"])
		yaml += fmt.Sprintf("    mode: %s\n", m["mode"])
		prices := m["prices"].(map[string]float64)
		yaml += "    prices:\n"
		for k, v := range prices {
			yaml += fmt.Sprintf("      %s: %v\n", k, v)
		}
	}
	return yaml
}

func fetchModelPriceList() (*listResponse, error) {
	resp, err := testutil.GetClient().Get("/open-api/v1/model-prices")
	if err != nil {
		return nil, err
	}
	if resp.ErrNum != 200 {
		return nil, fmt.Errorf("list failed: %d %s", resp.ErrNum, resp.ErrMsg)
	}
	var list listResponse
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	return &list, nil
}
