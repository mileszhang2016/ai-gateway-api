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

package global_route_rules

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/openapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/route_rules"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeRouteRulesStorager struct {
	fetchRouteRulesFn   func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error)
	createRouteRulesFn  func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error)
	updateRouteRulesFn  func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error)
	fetchRouteRulesByID func(ctx context.Context, id int64) (*shared.RouteRulesParam, error)
}

func (f *fakeRouteRulesStorager) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
	if f.createRouteRulesFn != nil {
		return f.createRouteRulesFn(ctx, ruleType, owner, param)
	}
	return 1, nil
}

func (f *fakeRouteRulesStorager) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
	if f.fetchRouteRulesFn != nil {
		return f.fetchRouteRulesFn(ctx, ruleType, owner)
	}
	return nil, nil
}

func (f *fakeRouteRulesStorager) FetchRouteRulesList(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error) {
	return nil, 0, nil
}

func (f *fakeRouteRulesStorager) UpdateRouteRules(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
	if f.updateRouteRulesFn != nil {
		return f.updateRouteRulesFn(ctx, id, param)
	}
	return id, nil
}

func (f *fakeRouteRulesStorager) DeleteRouteRules(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeRouteRulesStorager) FetchRouteRulesByID(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
	if f.fetchRouteRulesByID != nil {
		return f.fetchRouteRulesByID(ctx, id)
	}
	return nil, nil
}

func setupRouteRulesManager(storager shared.RouteRulesStorager) func() {
	old := container.RouteRulesManager
	container.RouteRulesManager = route_rules.NewRouteRulesManager(&testutil.FakeTxn{}, storager)
	return func() {
		container.RouteRulesManager = old
	}
}

func validRouteRulesParam() *shared.RouteRulesParam {
	enabled := true
	weight := 100
	return &shared.RouteRulesParam{
		Enabled: &enabled,
		Rules: []*shared.AiRouteRuleParam{
			{
				Name: lib.PString("rule1"),
				Cond: lib.PString("default_t()"),
				Targets: []*shared.AiRouteTargetParam{
					{ClusterName: lib.PString("cluster1"), Weight: &weight},
				},
			},
		},
	}
}

func TestGlobalRouteRulesGetAction(t *testing.T) {
	expected := validRouteRulesParam()
	expected.ID = lib.PInt64(1)
	defer setupRouteRulesManager(&fakeRouteRulesStorager{
		fetchRouteRulesFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			assert.Equal(t, shared.RouteRulesTypeGlobal, ruleType)
			return expected, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/global-route-rules", nil)
	data, err := GlobalRouteRulesGetAction(req)

	require.NoError(t, err)
	result, ok := data.(*shared.RouteRulesParam)
	require.True(t, ok)
	assert.Nil(t, result.ID)
	assert.Equal(t, expected.Rules[0].Name, result.Rules[0].Name)
}

func TestGlobalRouteRulesGetAction_NotFound(t *testing.T) {
	defer setupRouteRulesManager(&fakeRouteRulesStorager{
		fetchRouteRulesFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			return nil, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/global-route-rules", nil)
	data, err := GlobalRouteRulesGetAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestGlobalRouteRulesUpdateAction(t *testing.T) {
	updated := validRouteRulesParam()
	updated.ID = lib.PInt64(1)
	defer setupRouteRulesManager(&fakeRouteRulesStorager{
		fetchRouteRulesFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			return &shared.RouteRulesParam{ID: lib.PInt64(1)}, nil
		},
		updateRouteRulesFn: func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
			assert.Equal(t, int64(1), id)
			return id, nil
		},
		fetchRouteRulesByID: func(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
			return updated, nil
		},
	})()

	body := `{
		"enabled": true,
		"rules": [
			{
				"name": "rule1",
				"Cond": "default_t()",
				"targets": [
					{"ClusterName": "cluster1", "Weight": 100}
				]
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPut, "/global-route-rules", strings.NewReader(body))
	data, err := GlobalRouteRulesUpdateAction(req)

	require.NoError(t, err)
	result, ok := data.(*shared.RouteRulesParam)
	require.True(t, ok)
	assert.Nil(t, result.ID)
	assert.Len(t, result.Rules, 1)
}

func TestGlobalRouteRulesUpdateAction_InvalidJSON(t *testing.T) {
	defer setupRouteRulesManager(&fakeRouteRulesStorager{})()

	req := httptest.NewRequest(http.MethodPut, "/global-route-rules", strings.NewReader("not-json"))
	_, err := GlobalRouteRulesUpdateAction(req)

	require.Error(t, err)
}

func TestGlobalRouteRulesUpdateAction_NotFound(t *testing.T) {
	defer setupRouteRulesManager(&fakeRouteRulesStorager{
		fetchRouteRulesFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			return &shared.RouteRulesParam{ID: lib.PInt64(1)}, nil
		},
		updateRouteRulesFn: func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
			return id, nil
		},
		fetchRouteRulesByID: func(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
			return nil, nil
		},
	})()

	body := `{
		"enabled": true,
		"rules": [
			{
				"name": "rule1",
				"Cond": "default_t()",
				"targets": [
					{"ClusterName": "cluster1", "Weight": 100}
				]
			}
		]
	}`
	req := httptest.NewRequest(http.MethodPut, "/global-route-rules", strings.NewReader(body))
	_, err := GlobalRouteRulesUpdateAction(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Record Not Exist")
}
