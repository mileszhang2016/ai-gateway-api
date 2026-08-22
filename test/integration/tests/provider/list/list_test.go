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

type providerListResponse struct {
	List       []map[string]interface{} `json:"list"`
	Pagination struct {
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
		Total    int64 `json:"total"`
	} `json:"pagination"`
}

func TestProvider_List(t *testing.T) {
	providerA := testutil.UniqueProviderName()
	providerB := testutil.UniqueProviderName()

	_, err := testutil.CreateProvider(providerA)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.CreateProvider(providerB, map[string]interface{}{
		"model_protocols": []string{"anthropic"},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("PV-2-001 默认分页", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/providers")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list providerListResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Pagination.Total, int64(2))
		assert.NotEmpty(t, list.List)
	})

	t.Run("PV-2-002 自定义分页", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/providers", map[string]string{
			"page":      "1",
			"page_size": "1",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list providerListResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Pagination.Total, int64(2))
		assert.Len(t, list.List, 1)
	})

	t.Run("PV-2-003 按 model_protocol 过滤", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/providers", map[string]string{
			"model_protocol": "openai",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)

		var list providerListResponse
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.GreaterOrEqual(t, list.Pagination.Total, int64(1))
		for _, item := range list.List {
			protocols, ok := item["model_protocols"].([]interface{})
			assert.True(t, ok, "model_protocols should be an array")
			found := false
			for _, p := range protocols {
				if p == "openai" {
					found = true
					break
				}
			}
			assert.True(t, found, "each returned provider must contain openai in model_protocols")
		}
	})

	t.Cleanup(func() {
		testutil.DeleteProvider(providerA)
		testutil.DeleteProvider(providerB)
	})
}
