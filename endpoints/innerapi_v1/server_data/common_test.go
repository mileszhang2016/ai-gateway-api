// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package server_data

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iroute_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeRouteRuleStorager struct {
	fetchRoutRulesFn func(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error)
}

func (f *fakeRouteRuleStorager) UpsertProductRule(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error {
	return nil
}

func (f *fakeRouteRuleStorager) FetchProductRule(ctx context.Context, product *ibasic.Product, clusterList []*icluster_conf.Cluster) (*iroute_conf.ProductRouteRule, error) {
	return nil, nil
}

func (f *fakeRouteRuleStorager) FetchRoutRules(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
	if f.fetchRoutRulesFn != nil {
		return f.fetchRoutRulesFn(ctx, products, clusters)
	}
	return nil, nil
}

type fakeDomainStorager struct {
	fetchDomainsFn func(ctx context.Context, param *iroute_conf.DomainFilter) ([]*iroute_conf.Domain, error)
}

func (f *fakeDomainStorager) FetchDomains(ctx context.Context, param *iroute_conf.DomainFilter) ([]*iroute_conf.Domain, error) {
	if f.fetchDomainsFn != nil {
		return f.fetchDomainsFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeDomainStorager) CreateDomain(ctx context.Context, product *ibasic.Product, param *iroute_conf.DomainParam) error {
	return nil
}

func (f *fakeDomainStorager) DeleteDomain(ctx context.Context, product *ibasic.Product, domain *iroute_conf.Domain) error {
	return nil
}

type fakeProductStoragerForRoute struct {
	fetchProductsFn func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error)
}

func (f *fakeProductStoragerForRoute) FetchProducts(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
	if f.fetchProductsFn != nil {
		return f.fetchProductsFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeProductStoragerForRoute) DeleteProduct(ctx context.Context, p *ibasic.Product) error {
	return nil
}

func (f *fakeProductStoragerForRoute) CreateProduct(ctx context.Context, p *ibasic.ProductParam) error {
	return nil
}

func (f *fakeProductStoragerForRoute) UpdateProduct(ctx context.Context, p *ibasic.Product, newVal *ibasic.ProductParam) error {
	return nil
}

type fakeClusterStoragerForRoute struct {
	fetchClusterListFn func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error)
}

func (f *fakeClusterStoragerForRoute) FetchCluster(ctx context.Context, param *icluster_conf.ClusterFilter) (*icluster_conf.Cluster, error) {
	return nil, nil
}

func (f *fakeClusterStoragerForRoute) FetchClusterList(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
	if f.fetchClusterListFn != nil {
		return f.fetchClusterListFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeClusterStoragerForRoute) ClusterUpdate(ctx context.Context, product *ibasic.Product, old *icluster_conf.Cluster, param *icluster_conf.ClusterParam) error {
	return nil
}

func (f *fakeClusterStoragerForRoute) ClusterCreate(ctx context.Context, product *ibasic.Product, param *icluster_conf.ClusterParam, subClusters []*icluster_conf.SubCluster) (int64, error) {
	return 0, nil
}

func (f *fakeClusterStoragerForRoute) ClusterDelete(ctx context.Context, product *ibasic.Product, cluster *icluster_conf.Cluster) error {
	return nil
}

func (f *fakeClusterStoragerForRoute) BindSubCluster(ctx context.Context, cluster *icluster_conf.Cluster, appendSubClusters, unbindSubClusters []*icluster_conf.SubCluster) error {
	return nil
}

func (f *fakeClusterStoragerForRoute) FetchLBMatrixList(ctx context.Context) (map[int64]map[string]map[string]int, error) {
	return nil, nil
}

func newTestCluster() *icluster_conf.Cluster {
	return &icluster_conf.Cluster{
		Name: "cluster1",
		Basic: &icluster_conf.ClusterBasic{
			Protocol: lib.PString("http"),
			Connection: &icluster_conf.ClusterBasicConnection{
				MaxIdleConnPerRs: 2,
			},
			Retries: &icluster_conf.ClusterBasicRetries{},
			Timeouts: &icluster_conf.ClusterBasicTimeouts{},
			Buffers: &icluster_conf.ClusterBasicBuffers{},
		},
		StickySessions: &icluster_conf.ClusterStickySessions{},
	}
}

func setupRouteRuleManager(
	routeRuleStorager iroute_conf.RouteRuleStorager,
	domainStorager iroute_conf.DomainStorager,
	clusterStorager icluster_conf.ClusterStorager,
	productStorager ibasic.ProductStorager,
	version string,
) func() {
	old := container.RouteRuleManager
	container.RouteRuleManager = iroute_conf.NewRouteRuleManager(
		&testutil.FakeTxn{},
		routeRuleStorager,
		clusterStorager,
		productStorager,
		testutil.NewVersionControlManager(version),
		domainStorager,
	)
	return func() {
		container.RouteRuleManager = old
	}
}
