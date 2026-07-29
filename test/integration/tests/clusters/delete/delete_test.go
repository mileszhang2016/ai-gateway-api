package clusters_test

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

func TestClusters_Delete(t *testing.T) {
	t.Run("CL-5-001 删除集群", func(t *testing.T) {
		clusterName := testutil.UniqueClusterName()
		if _, err := testutil.CreateCluster(clusterName); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		resp, err := testutil.GetClient().Delete("/open-api/v1/clusters/" + clusterName)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertSuccess(t, resp)
		resp, _ = testutil.GetClient().Get("/open-api/v1/clusters/" + clusterName)
		testutil.AssertErrCode(t, resp, 404)
	})

	t.Run("CL-5-002 删除不存在的集群", func(t *testing.T) {
		resp, err := testutil.GetClient().Delete("/open-api/v1/clusters/non_existent_cluster")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		testutil.AssertErrCode(t, resp, 404)
	})
}
