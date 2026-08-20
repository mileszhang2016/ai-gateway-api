package clusters_test

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

func TestClusters_List(t *testing.T) {
	clusterName := testutil.UniqueClusterName()
	if _, err := testutil.CreateCluster(clusterName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := testutil.GetClient().Get("/open-api/v1/clusters")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	var list []interface{}
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		t.Fatalf("unmarshal data failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty cluster list")
	}
	for _, item := range list {
		cluster := item.(map[string]interface{})
		assert.NotEmpty(t, cluster["name"])
		assert.NotContains(t, cluster, "ready")
		assert.NotContains(t, cluster, "sub_clusters")
		assert.NotContains(t, cluster, "scheduler")
	}

	t.Cleanup(func() {
		testutil.DeleteCluster(clusterName)
	})
}
