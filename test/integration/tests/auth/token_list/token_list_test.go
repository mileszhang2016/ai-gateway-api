package auth_test

import (
	"encoding/json"
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

func TestAuth_TokenList(t *testing.T) {
	tokenName := testutil.UniqueTokenName()
	if _, err := testutil.CreateToken(tokenName, "System"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/open-api/v1/auth/tokens")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	var list []interface{}
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	assert.GreaterOrEqual(t, len(list), 1)
	for _, item := range list {
		token := item.(map[string]interface{})
		assert.NotEmpty(t, token["name"])
		assert.NotEmpty(t, token["token"])
		assert.Contains(t, []string{"System", "Support"}, token["scope"])
		assert.NotContains(t, token, "product_name")
	}

	t.Cleanup(func() {
		testutil.DeleteToken(tokenName)
	})
}
