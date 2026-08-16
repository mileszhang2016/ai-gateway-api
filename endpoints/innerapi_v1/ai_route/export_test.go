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

package ai_route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/imods"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAPIKeyStoragerForAIRoute struct{}

func (f *fakeAPIKeyStoragerForAIRoute) FetchAPIKeyList(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
	return nil, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) CreateAPIKey(ctx context.Context, param *api_key.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) UpdateAPIKey(ctx context.Context, filter *api_key.APIKeyFilter, param *api_key.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) DeleteAPIKey(ctx context.Context, filter *api_key.APIKeyFilter) error {
	return nil
}

func (f *fakeAPIKeyStoragerForAIRoute) CreateAPIKeyToken(ctx context.Context, param *api_key.APIKeyTokenParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForAIRoute) UpdateAPIKeyToken(ctx context.Context, filter *api_key.APIKeyTokenFilter, param *api_key.APIKeyTokenParam) error {
	return nil
}

func (f *fakeAPIKeyStoragerForAIRoute) FetchAPIKeyTokenList(ctx context.Context, filter *api_key.APIKeyTokenFilter) ([]*api_key.APIKeyTokenParam, error) {
	return nil, nil
}

type fakeEntityStoragerForAIRoute struct{}

func (f *fakeEntityStoragerForAIRoute) CreateEntity(ctx context.Context, param *entity.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForAIRoute) FetchEntity(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForAIRoute) FetchEntityList(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForAIRoute) UpdateEntity(ctx context.Context, filter *entity.EntityFilter, param *entity.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForAIRoute) DeleteEntity(ctx context.Context, filter *entity.EntityFilter) error {
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

func (f *fakeRouteRulesStorager) FetchAllRouteRules(ctx context.Context) ([]*shared.RouteRulesParam, error) {
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
