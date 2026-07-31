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

package server_data

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/iroute_conf"
)

func TestExportAction(t *testing.T) {
	defer setupRouteRuleManager(
		&fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
				return map[int64]*iroute_conf.ProductRouteRule{
					1: {},
				}, nil
			},
		},
		&fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *iroute_conf.DomainFilter) ([]*iroute_conf.Domain, error) {
				return []*iroute_conf.Domain{
					{ProductID: 1, Name: "example.com"},
				}, nil
			},
		},
		&fakeClusterStoragerForRoute{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{newTestCluster()}, nil
			},
		},
		&fakeProductStoragerForRoute{
			fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
				return []*ibasic.Product{
					{ID: 1, Name: "AI_product"},
				}, nil
			},
		},
		"v2",
	)()

	req := httptest.NewRequest(http.MethodGet, "/configs/tls_conf/server_data_conf?version=", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*iroute_conf.RouteRuleExportData)
	require.True(t, ok)
	assert.Equal(t, "v2", conf.Version)
	assert.NotNil(t, conf.RouteTable)
	assert.NotNil(t, conf.HostTable)
	assert.NotNil(t, conf.ClusterConf)
}

func TestExportAction_FetchError(t *testing.T) {
	defer setupRouteRuleManager(
		&fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
				return nil, errors.New("db error")
			},
		},
		&fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *iroute_conf.DomainFilter) ([]*iroute_conf.Domain, error) {
				return []*iroute_conf.Domain{
					{ProductID: 1, Name: "example.com"},
				}, nil
			},
		},
		&fakeClusterStoragerForRoute{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{newTestCluster()}, nil
			},
		},
		&fakeProductStoragerForRoute{
			fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
				return []*ibasic.Product{
					{ID: 1, Name: "AI_product"},
				}, nil
			},
		},
		"v2",
	)()

	req := httptest.NewRequest(http.MethodGet, "/configs/tls_conf/server_data_conf?version=", nil)
	_, err := ExportAction(req)

	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestExportAction_VersionNotChanged(t *testing.T) {
	defer setupRouteRuleManager(
		&fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
				return map[int64]*iroute_conf.ProductRouteRule{
					1: {},
				}, nil
			},
		},
		&fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *iroute_conf.DomainFilter) ([]*iroute_conf.Domain, error) {
				return []*iroute_conf.Domain{
					{ProductID: 1, Name: "example.com"},
				}, nil
			},
		},
		&fakeClusterStoragerForRoute{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{newTestCluster()}, nil
			},
		},
		&fakeProductStoragerForRoute{
			fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
				return []*ibasic.Product{
					{ID: 1, Name: "AI_product"},
				}, nil
			},
		},
		"v1",
	)()

	req := httptest.NewRequest(http.MethodGet, "/configs/tls_conf/server_data_conf?version=v1", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
