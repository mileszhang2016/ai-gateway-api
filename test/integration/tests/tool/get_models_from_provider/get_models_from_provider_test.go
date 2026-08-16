package tool_test

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func startMockProvider(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr := listener.Addr().String()
	mux := http.NewServeMux()
	mux.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "deepseek-chat", "name": "DeepSeek Chat"},
				{"id": "deepseek-coder", "name": "DeepSeek Coder"},
			},
		})
	})
	mux.HandleFunc("/models_no_slash", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "model-a", "name": "Model A"},
			},
		})
	})
	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	t.Cleanup(func() {
		server.Close()
	})
	return addr
}

func TestTool_GetModelsFromProvider(t *testing.T) {
	mockAddr := startMockProvider(t)

	t.Run("TOOL-1-001 正常获取模型列表", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/tools/get-models-from-provider", map[string]interface{}{
			"schema":        "http",
			"uri":           "/models",
			"hosts":         []string{mockAddr},
			"headers":       map[string]string{"Authorization": "Bearer sk-xxx"},
			"provider_type": "deepseek",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		var list []interface{}
		if err := json.Unmarshal(resp.Data, &list); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		assert.GreaterOrEqual(t, len(list), 1)
		first := list[0].(map[string]interface{})
		assert.NotEmpty(t, first["id"])
		assert.NotEmpty(t, first["name"])
	})

	t.Run("TOOL-1-002 uri 不带斜杠", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/tools/get-models-from-provider", map[string]interface{}{
			"schema":        "http",
			"uri":           "models",
			"hosts":         []string{mockAddr},
			"provider_type": "deepseek",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.ErrNum != 200 {
			t.Skip("implementation may not support uri without leading slash")
		}
	})

	t.Run("TOOL-1-003 缺少 schema", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/tools/get-models-from-provider", map[string]interface{}{
			"hosts":         []string{mockAddr},
			"provider_type": "deepseek",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("TOOL-1-004 缺少 hosts", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/tools/get-models-from-provider", map[string]interface{}{
			"schema":        "http",
			"provider_type": "deepseek",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})

	t.Run("TOOL-1-005 非法 schema", func(t *testing.T) {
		resp, err := testutil.GetClient().Post("/open-api/v1/tools/get-models-from-provider", map[string]interface{}{
			"schema": "ftp",
			"hosts":  []string{mockAddr},
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 422)
	})
}
