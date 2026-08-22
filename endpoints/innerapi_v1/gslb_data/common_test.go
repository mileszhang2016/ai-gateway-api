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

	"github.com/rainway-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeClusterStorager struct {
	fetchClusterListFn  func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error)
	fetchLBMatrixListFn func(ctx context.Context) (map[int64]map[string]map[string]int, error)
}

func (f *fakeClusterStorager) FetchCluster(ctx context.Context, param *icluster_conf.ClusterFilter) (*icluster_conf.Cluster, error) {
	return nil, nil
}

func (f *fakeClusterStorager) FetchClusterList(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
	if f.fetchClusterListFn != nil {
		return f.fetchClusterListFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeClusterStorager) ClusterUpdate(ctx context.Context, product *ibasic.Product, old *icluster_conf.Cluster, param *icluster_conf.ClusterParam) error {
	return nil
}

func (f *fakeClusterStorager) ClusterCreate(ctx context.Context, product *ibasic.Product, param *icluster_conf.ClusterParam, subClusters []*icluster_conf.SubCluster) (int64, error) {
	return 0, nil
}

func (f *fakeClusterStorager) ClusterDelete(ctx context.Context, product *ibasic.Product, cluster *icluster_conf.Cluster) error {
	return nil
}

func (f *fakeClusterStorager) BindSubCluster(ctx context.Context, cluster *icluster_conf.Cluster, appendSubClusters, unbindSubClusters []*icluster_conf.SubCluster) error {
	return nil
}

func (f *fakeClusterStorager) FetchLBMatrixList(ctx context.Context) (map[int64]map[string]map[string]int, error) {
	if f.fetchLBMatrixListFn != nil {
		return f.fetchLBMatrixListFn(ctx)
	}
	return nil, nil
}

type fakeSubClusterStorager struct{}

func (f *fakeSubClusterStorager) FetchSubClusterList(ctx context.Context, param *icluster_conf.SubClusterFilter) ([]*icluster_conf.SubCluster, error) {
	return nil, nil
}

func (f *fakeSubClusterStorager) CreateSubCluster(ctx context.Context, param *icluster_conf.SubClusterParam) error {
	return nil
}

func (f *fakeSubClusterStorager) DeleteSubCluster(ctx context.Context, param *icluster_conf.SubCluster) error {
	return nil
}

func (f *fakeSubClusterStorager) UpdateSubCluster(ctx context.Context, one *icluster_conf.SubCluster, param *icluster_conf.SubClusterParam) error {
	return nil
}

type fakeBFEClusterStorager struct{}

func (f *fakeBFEClusterStorager) DeleteBFECluster(ctx context.Context, cluster *ibasic.BFECluster) error {
	return nil
}

func (f *fakeBFEClusterStorager) CreateBFECluster(ctx context.Context, param *ibasic.BFEClusterParam) error {
	return nil
}

func (f *fakeBFEClusterStorager) FetchBFEClusters(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
	return nil, nil
}

type fakePoolStorager struct{}

func (f *fakePoolStorager) FetchPool(ctx context.Context, name string) (*icluster_conf.Pool, error) {
	return nil, nil
}

func (f *fakePoolStorager) FetchPools(ctx context.Context, param *icluster_conf.PoolFilter) ([]*icluster_conf.Pool, error) {
	return nil, nil
}

func (f *fakePoolStorager) CreatePool(ctx context.Context, product *ibasic.Product, data *icluster_conf.PoolParam) (*icluster_conf.Pool, error) {
	return nil, nil
}

func (f *fakePoolStorager) UpdatePool(ctx context.Context, oldData *icluster_conf.Pool, diff *icluster_conf.PoolParam) error {
	return nil
}

func (f *fakePoolStorager) DeletePool(ctx context.Context, pool *icluster_conf.Pool) error {
	return nil
}

func setupClusterManager(storager icluster_conf.ClusterStorager, version string) func() {
	old := container.ClusterManager
	container.ClusterManager = icluster_conf.NewClusterManager(
		&testutil.FakeTxn{},
		storager,
		&fakeSubClusterStorager{},
		&fakeBFEClusterStorager{},
		&fakePoolStorager{},
		nil,
		testutil.NewVersionControlManager(version),
		nil,
		nil,
	)
	return func() {
		container.ClusterManager = old
	}
}
