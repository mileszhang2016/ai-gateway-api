package model_provider_type_test

import (
	"encoding/json"
	"os"
	"testing"

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

func TestModelProviderType_List(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		wantCode int
	}{
		{"MPT-1-001 获取提供商类型列表", "GET", 200},
		{"MPT-1-003 非法 Method POST", "POST", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *testutil.APIResponse
			var err error
			client := testutil.GetClient()
			if tt.method == "GET" {
				resp, err = client.Get("/open-api/v1/model-provider-types")
			} else {
				resp, err = client.Post("/open-api/v1/model-provider-types", map[string]interface{}{})
			}
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.wantCode == 200 {
				var list []interface{}
				if err := json.Unmarshal(resp.Data, &list); err != nil {
					t.Fatalf("unmarshal data list failed: %v", err)
				}
				if len(list) == 0 {
					t.Error("expected non-empty provider type list")
				}
				for _, item := range list {
					s, ok := item.(string)
					if !ok || s == "" {
						t.Errorf("expected non-empty string element, got %v", item)
					}
				}
			}
		})
	}
}
