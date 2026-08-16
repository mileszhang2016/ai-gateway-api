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

func TestAuth_ListUsers(t *testing.T) {
	userName := testutil.UniqueUserName()
	if err := testutil.CreateUser(userName, "password@123"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/open-api/v1/auth/users")
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
		user := item.(map[string]interface{})
		assert.NotEmpty(t, user["user_name"])
		assert.Equal(t, true, user["is_admin"])
		assert.NotContains(t, user, "products")
	}

	t.Cleanup(func() {
		testutil.DeleteUser(userName)
	})
}
