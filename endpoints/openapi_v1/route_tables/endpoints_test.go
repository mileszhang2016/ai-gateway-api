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

package route_tables

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/endpoints/openapi_v1/internal/testutil"
	"github.com/yf-networks/ai-gateway-api/model/shared"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

type fakeRouteRulesStoragerForTables struct {
	fetchRouteRulesListFn func(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error)
}

func (f *fakeRouteRulesStoragerForTables) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
	return 0, nil
}

func (f *fakeRouteRulesStoragerForTables) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
	return nil, nil
}

func (f *fakeRouteRulesStoragerForTables) FetchRouteRulesList(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error) {
	if f.fetchRouteRulesListFn != nil {
		return f.fetchRouteRulesListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (f *fakeRouteRulesStoragerForTables) UpdateRouteRules(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
	return 0, nil
}

func (f *fakeRouteRulesStoragerForTables) DeleteRouteRules(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeRouteRulesStoragerForTables) FetchRouteRulesByID(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
	return nil, nil
}

func setupRouteRulesManagerForTables(storager shared.RouteRulesStorager) func() {
	old := container.RouteRulesManager
	container.RouteRulesManager = shared.NewRouteRulesManager(&testutil.FakeTxn{}, storager)
	return func() {
		container.RouteRulesManager = old
	}
}

func TestEndpoints(t *testing.T) {
	require.Len(t, Endpoints, 1)
	for _, ep := range Endpoints {
		require.NotNil(t, ep)
		assert.NotEmpty(t, ep.Path)
		assert.NotEmpty(t, ep.Method)
	}
}

func TestRouteTablesListAction(t *testing.T) {
	defer setupRouteRulesManagerForTables(&fakeRouteRulesStoragerForTables{
		fetchRouteRulesListFn: func(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error) {
			assert.NotNil(t, filter.Page)
			assert.NotNil(t, filter.PageSize)
			return []*shared.RouteTableParam{
				{Type: "global", Owner: "global", Enabled: true},
			}, 1, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/route-tables?page=2&page_size=50", nil)
	data, err := RouteTablesListAction(req)

	require.NoError(t, err)
	resp, ok := data.(*RouteTablesListResponse)
	require.True(t, ok)
	assert.Len(t, resp.List, 1)
	assert.Equal(t, int64(1), resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.Page)
	assert.Equal(t, 50, resp.Pagination.PageSize)
}
