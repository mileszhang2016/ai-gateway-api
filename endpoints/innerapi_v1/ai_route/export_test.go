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

package ai_route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imods"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeAPIKeyStoragerForAIRoute struct{}

func (f *fakeAPIKeyStoragerForAIRoute) FetchAPIKeyList(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
	return nil, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) CreateAPIKey(ctx context.Context, param *icluster_conf.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) UpdateAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter, param *icluster_conf.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) DeleteAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter) error {
	return nil
}

func (f *fakeAPIKeyStoragerForAIRoute) CreateAPIKeyToken(ctx context.Context, param *icluster_conf.APIKeyTokenParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) UpdateAPIKeyToken(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter, param *icluster_conf.APIKeyTokenParam) error {
	return nil
}

func (f *fakeAPIKeyStoragerForAIRoute) FetchAPIKeyTokenList(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter) ([]*icluster_conf.APIKeyTokenParam, error) {
	return nil, nil
}

type fakeEntityStoragerForAIRoute struct{}

func (f *fakeEntityStoragerForAIRoute) CreateEntity(ctx context.Context, param *quota.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForAIRoute) FetchEntity(ctx context.Context, filter *quota.EntityFilter) (*quota.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForAIRoute) FetchEntityList(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForAIRoute) UpdateEntity(ctx context.Context, filter *quota.EntityFilter, param *quota.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForAIRoute) DeleteEntity(ctx context.Context, filter *quota.EntityFilter) error {
	return nil
}

type fakeRouteRulesStorager struct{}

func (f *fakeRouteRulesStorager) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
	return 0, nil
}

func (f *fakeRouteRulesStorager) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
	return nil, nil
}

func (f *fakeRouteRulesStorager) FetchRouteRulesList(ctx context.Context, filter *shared.RouteRulesFilter) ([]*shared.RouteTableParam, int64, error) {
	return nil, 0, nil
}

func (f *fakeRouteRulesStorager) UpdateRouteRules(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
	return 0, nil
}

func (f *fakeRouteRulesStorager) DeleteRouteRules(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeRouteRulesStorager) FetchRouteRulesByID(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
	return nil, nil
}

func setupAIRouteExporter(version string) func() {
	old := container.AIRouteExporter
	container.AIRouteExporter = imods.NewAIRouteExporter(
		&fakeAPIKeyStoragerForAIRoute{},
		&fakeEntityStoragerForAIRoute{},
		&fakeRouteRulesStorager{},
		testutil.NewVersionControlManager(version),
	)
	return func() {
		container.AIRouteExporter = old
	}
}

func TestExportAction(t *testing.T) {
	defer setupAIRouteExporter("v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/ai-route?version=", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*imods.AiRouteDataExport)
	require.True(t, ok)
	assert.Equal(t, "v2", conf.Version)
	assert.NotNil(t, conf.RouteRules)
	assert.NotNil(t, conf.ApikeyRouteTableBindings)
}

func TestExportAction_VersionNotChanged(t *testing.T) {
	defer setupAIRouteExporter("v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/ai-route?version=v1", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
