package api_key_test

import (
	"os"
	"testing"

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

func TestAPIKey_Create(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	entityName := testutil.UniqueEntityName()
	entityID, err := testutil.CreateEntity(entityName, typeName, "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	importedKey := "my-imported-key-" + testutil.RandomString(6)

	// 预先创建重复 key
	if _, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": "dup-key",
		"key":         importedKey,
	}); err != nil {
		t.Fatalf("setup dup key failed: %v", err)
	}

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
		check    func(t *testing.T, resp *testutil.APIResponse)
	}{
		{
			name:     "AK-1-001 最小参数创建",
			body:     map[string]interface{}{"description": "test-key-min"},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				testutil.AssertDataFieldEquals(t, resp, "description", "test-key-min")
				testutil.AssertDataFieldEquals(t, resp, "enabled", true)
				testutil.AssertDataFieldNotEmpty(t, resp, "quota_plan")
			},
		},
		{
			name: "AK-1-002 完整参数创建",
			body: map[string]interface{}{
				"description": "test-key-full",
				"expired_time": -1,
				"enabled": true,
				"unlimited_quota": false,
				"models": []string{"gpt-4"},
				"subnet": []string{"10.0.0.0/8"},
				"quota_plan": map[string]interface{}{
					"unlimited": false,
					"quota": 1000000,
					"unit": "total_token",
					"reset_period": "monthly",
				},
				"rate_limit_policy": map[string]interface{}{
					"enabled": true,
					"rules": map[string]interface{}{
						"tpm": []interface{}{
							map[string]interface{}{
								"name": "1m", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1,
							},
						},
						"max_concurrency": 50,
					},
				},
				"entity_id": entityID,
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				testutil.AssertDataFieldEquals(t, resp, "description", "test-key-full")
				testutil.AssertDataFieldNotEmpty(t, resp, "entity")
			},
		},
		{
			name:     "AK-1-003 导入外部 key",
			body:     map[string]interface{}{"description": "test-key-import", "key": "my-imported-key-001"},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				testutil.AssertDataFieldEquals(t, resp, "key", "my-imported-key-001")
			},
		},
		{
			name:     "AK-1-004 导入重复 key",
			body:     map[string]interface{}{"description": "test-key-dup", "key": importedKey},
			wantCode: 422,
		},
		{
			name:     "AK-1-005 缺少 description",
			body:     map[string]interface{}{},
			wantCode: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", tt.body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.check != nil && resp.ErrNum == 200 {
				tt.check(t, resp)
			}
		})
	}

	t.Cleanup(func() {
		testutil.DeleteEntity(entityID)
		testutil.DeleteEntityType(typeName)
	})
}
