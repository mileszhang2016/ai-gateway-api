package expression_verify_test

import (
	"encoding/json"
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

func TestExpressionVerify(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		body     map[string]interface{}
		wantCode int
		wantNull bool
	}{
		{"EV-1-001 default_t", "default_t()", nil, 200, true},
		{"EV-1-002 req_path_prefix", "req_path_prefix(\"/open-api/v1\")", nil, 500, false},
		{"EV-1-003 组合表达式", "and(req_method_in(\"POST\"), req_path_prefix(\"/v1\"))", nil, 500, false},
		{"EV-1-004 缺少 expression", "", map[string]interface{}{}, 422, false},
		{"EV-1-005 空字符串", "", nil, 422, false},
		{"EV-1-006 括号不匹配", "default_t(", nil, 500, false},
		{"EV-1-007 未知函数", "unknown_func()", nil, 500, false},
		{"EV-1-008 缺少引号", "req_path_prefix(/v1)", nil, 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			if body == nil && tt.expr != "" {
				body = map[string]interface{}{"expression": tt.expr}
			} else if body == nil {
				body = map[string]interface{}{"expression": tt.expr}
			}
			resp, err := testutil.GetClient().Patch("/open-api/v1/expression/verify", body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.wantNull {
				if string(resp.Data) != "null" {
					t.Errorf("expected Data=null, got %s", string(resp.Data))
				}
			}
			if !tt.wantNull && resp.ErrNum == 500 {
				var data map[string]interface{}
				if err := json.Unmarshal(resp.Data, &data); err == nil {
					if msg, ok := data["message"].(string); !ok || msg == "" {
						t.Error("expected non-empty message in verify result")
					}
				}
			}
		})
	}
}
