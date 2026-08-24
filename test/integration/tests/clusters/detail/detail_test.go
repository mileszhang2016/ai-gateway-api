package clusters_test

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

func TestClusters_Detail(t *testing.T) {
	clusterName := testutil.UniqueClusterName()
	if _, err := testutil.CreateCluster(clusterName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/open-api/v1/clusters/" + clusterName)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	testutil.AssertDataFieldEquals(t, resp, "name", clusterName)

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data failed: %v", err)
	}
	assert.NotContains(t, data, "ready")
	assert.NotContains(t, data, "sub_clusters")
	assert.NotContains(t, data, "scheduler")
	assert.NotContains(t, data, "instance_pool")
	assert.Contains(t, data, "llm_config")

	sticky, ok := data["sticky_sessions"].(map[string]interface{})
	if assert.True(t, ok, "sticky_sessions should be an object") {
		assert.Equal(t, false, sticky["enabled"])
		assert.Equal(t, "CLIENT_IP_ONLY", sticky["hash_strategy"])
		assert.Equal(t, "", sticky["hash_header"])
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}
