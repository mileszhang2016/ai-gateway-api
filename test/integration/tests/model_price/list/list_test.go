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

type listResponse struct {
	Total int64                    `json:"total"`
	Items []map[string]interface{} `json:"items"`
}

func TestModelPrice_List(t *testing.T) {
	providerA := testutil.UniqueName("provider")
	providerB := testutil.UniqueName("provider")

	idA, err := testutil.CreateModelPrice(map[string]interface{}{
		"provider":   providerA,
		"model":      "model-a",
		"base_model": "model-a",
		"mode":       "chat",
		"prices": map[string]interface{}{
			"input_cost_per_token": 0.001,
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer testutil.DeleteModelPrice(idA)

	idB, err := testutil.CreateModelPrice(map[string]interface{}{
		"provider":   providerB,
		"model":      "model-b",
		"base_model": "model-b",
		"mode":       "completion",
		"prices": map[string]interface{}{
			"input_cost_per_token": 0.002,
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer testutil.DeleteModelPrice(idB)

	t.Run("MP-3-001 默认分页", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list listResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Total, int64(2))
		assert.NotEmpty(t, list.Items)
	})

	t.Run("MP-3-002 自定义分页", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"page":      "1",
			"page_size": "1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list listResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Total, int64(2))
		assert.Len(t, list.Items, 1)
	})

	t.Run("MP-3-003 按 provider 过滤", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"provider": providerA,
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list listResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Total, int64(1))
		for _, item := range list.Items {
			assert.Equal(t, providerA, item["provider"])
		}
	})

	t.Run("MP-3-004 按 mode 过滤", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"mode": "chat",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list listResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Total, int64(1))
		for _, item := range list.Items {
			assert.Equal(t, "chat", item["mode"])
		}
	})
}

func int64ToStr(v int64) string {
	return fmt.Sprintf("%d", v)
}
