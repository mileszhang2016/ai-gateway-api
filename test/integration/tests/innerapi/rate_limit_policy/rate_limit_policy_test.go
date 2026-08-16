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

func TestInnerAPI_RateLimitPolicy(t *testing.T) {
	apiKeyID, err := testutil.CreateAPIKey("rate-limit-key", "")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	_, err = testutil.GetClient().Patch("/open-api/v1/api-keys/"+apiKeyID, map[string]interface{}{
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{
						"name": "1m", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/inner-api/v1/configs/rate-limit-policy")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataNotEmpty(t, resp)
	testutil.AssertDataFieldNotEmpty(t, resp, "Config")
	testutil.AssertDataFieldNotEmpty(t, resp, "RateLimitPolicies")
	testutil.AssertDataFieldNotEmpty(t, resp, "ApikeyRateLimitPolicyBindings")
	testutil.AssertDataFieldNotEmpty(t, resp, "Version")

	t.Cleanup(func() {
		testutil.DeleteAPIKey(apiKeyID)
	})
}
