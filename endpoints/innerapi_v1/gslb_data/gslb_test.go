// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gslb_data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
)

func TestExportGSLBAction(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{
				{
					Name: "cluster1",
					Scheduler: map[string]map[string]int{
						"bfe-cluster": {"sub1": 100},
					},
				},
			}, nil
		},
	}, "v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/gslb?version=&bfe_cluster=bfe-cluster", nil)
	data, err := ExportGSLBAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*icluster_conf.GSLBConf)
	require.True(t, ok)
	assert.Equal(t, "v2", conf.Version)
	require.NotNil(t, conf.Clusters)
	assert.Contains(t, *conf.Clusters, "cluster1")
}

func TestExportGSLBAction_VersionNotChanged(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{
				{
					Name: "cluster1",
					Scheduler: map[string]map[string]int{
						"bfe-cluster": {"sub1": 100},
					},
				},
			}, nil
		},
	}, "v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/gslb?version=v1&bfe_cluster=bfe-cluster", nil)
	data, err := ExportGSLBAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestExportGSLBAction_MissingBFECluster(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{}, "v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/gslb?version=", nil)
	_, err := ExportGSLBAction(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "BFECluster")
}

func TestExportGSLBAction_BFEClusterNotExist(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{
				{
					Name: "cluster1",
					Scheduler: map[string]map[string]int{
						"other-cluster": {"sub1": 100},
					},
				},
			}, nil
		},
	}, "v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/gslb?version=&bfe_cluster=bfe-cluster", nil)
	_, err := ExportGSLBAction(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "BFECluster bfe-cluster Not Exist")
}
