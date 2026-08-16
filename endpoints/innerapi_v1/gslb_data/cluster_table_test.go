// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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
	"github.com/rainway-ai-gateway/ai-gateway-api/model/icluster_conf"
)

func TestExportClusterTableAction(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{
				{
					Name: "cluster1",
					SubClusters: []*icluster_conf.SubCluster{
						{
							Name: "sub1",
							InstancePool: &icluster_conf.Pool{
								Instances: []icluster_conf.Instance{
									{Name: "host1", Addr: "10.0.0.1", Port: 8080, Weight: 100},
								},
							},
						},
					},
				},
			}, nil
		},
		fetchLBMatrixListFn: func(ctx context.Context) (map[int64]map[string]map[string]int, error) {
			return nil, nil
		},
	}, "v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/cluster_table?version=", nil)
	data, err := ExportClusterTableAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*icluster_conf.ClusterTableConf)
	require.True(t, ok)
	assert.Equal(t, "v2", *conf.Version)
	require.NotNil(t, conf.Config)
	assert.Contains(t, *conf.Config, "cluster1")
}

func TestExportClusterTableAction_VersionNotChanged(t *testing.T) {
	defer setupClusterManager(&fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{
				{
					Name: "cluster1",
					SubClusters: []*icluster_conf.SubCluster{
						{
							Name: "sub1",
							InstancePool: &icluster_conf.Pool{
								Instances: []icluster_conf.Instance{
									{Name: "host1", Addr: "10.0.0.1", Port: 8080, Weight: 100},
								},
							},
						},
					},
				},
			}, nil
		},
		fetchLBMatrixListFn: func(ctx context.Context) (map[int64]map[string]map[string]int, error) {
			return nil, nil
		},
	}, "v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/gslb_data/cluster_table?version=v1", nil)
	data, err := ExportClusterTableAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
