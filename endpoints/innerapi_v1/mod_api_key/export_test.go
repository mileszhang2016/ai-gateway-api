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

package mod_api_key

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iai_route"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imods"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeAPIKeyStoragerForRule struct{}

func (f *fakeAPIKeyStoragerForRule) FetchAPIKeyList(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
	return nil, nil
}

func (f *fakeAPIKeyStoragerForRule) CreateAPIKey(ctx context.Context, param *icluster_conf.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRule) UpdateAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter, param *icluster_conf.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRule) DeleteAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter) error {
	return nil
}

func (f *fakeAPIKeyStoragerForRule) CreateAPIKeyToken(ctx context.Context, param *icluster_conf.APIKeyTokenParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRule) UpdateAPIKeyToken(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter, param *icluster_conf.APIKeyTokenParam) error {
	return nil
}

func (f *fakeAPIKeyStoragerForRule) FetchAPIKeyTokenList(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter) ([]*icluster_conf.APIKeyTokenParam, error) {
	return nil, nil
}

type fakeAIRouteRuleStoragerForRule struct{}

func (f *fakeAIRouteRuleStoragerForRule) FetchAIRouteRules(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
	return nil, nil
}

func (f *fakeAIRouteRuleStoragerForRule) CreateAIRouteRules(ctx context.Context, param []*iai_route.Rule) error {
	return nil
}

type fakeQuotaPlanStoragerForRule struct{}

func (f *fakeQuotaPlanStoragerForRule) CreateQuotaPlan(ctx context.Context, param *quota.QuotaPlanParam) (int64, error) {
	return 0, nil
}

func (f *fakeQuotaPlanStoragerForRule) FetchQuotaPlan(ctx context.Context, filter *quota.QuotaPlanFilter) (*quota.QuotaPlanParam, error) {
	return nil, nil
}

func (f *fakeQuotaPlanStoragerForRule) FetchQuotaPlanList(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
	return nil, nil
}

func (f *fakeQuotaPlanStoragerForRule) UpdateQuotaPlan(ctx context.Context, filter *quota.QuotaPlanFilter, param *quota.QuotaPlanParam) (int64, error) {
	return 0, nil
}

func (f *fakeQuotaPlanStoragerForRule) DeleteQuotaPlan(ctx context.Context, filter *quota.QuotaPlanFilter) error {
	return nil
}

type fakeEntityStoragerForRule struct{}

func (f *fakeEntityStoragerForRule) CreateEntity(ctx context.Context, param *quota.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForRule) FetchEntity(ctx context.Context, filter *quota.EntityFilter) (*quota.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForRule) FetchEntityList(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForRule) UpdateEntity(ctx context.Context, filter *quota.EntityFilter, param *quota.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForRule) DeleteEntity(ctx context.Context, filter *quota.EntityFilter) error {
	return nil
}

func setupAPIKeyRuleManager(version string) func() {
	origConfig := stateful.DefaultConfig
	stateful.DefaultConfig = &stateful.Config{
		RunTime: stateful.RunTimeConfig{
			AIRouteInnerProductName: "AI_product",
		},
	}

	old := container.APIKeyRuleManager
	container.APIKeyRuleManager = imods.NewAPIKeyRuleManager(
		&testutil.FakeTxn{},
		testutil.NewVersionControlManager(version),
		&fakeAPIKeyStoragerForRule{},
		&fakeAIRouteRuleStoragerForRule{},
		&fakeQuotaPlanStoragerForRule{},
		&fakeEntityStoragerForRule{},
	)
	return func() {
		container.APIKeyRuleManager = old
		stateful.DefaultConfig = origConfig
	}
}

func TestExportAction(t *testing.T) {
	defer setupAPIKeyRuleManager("v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/mod-api-key?version=", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*imods.ModAPIKeyRuleConf)
	require.True(t, ok)
	assert.Equal(t, "v2", *conf.Version)
	assert.Contains(t, conf.Config, "AI_product")
}

func TestExportAction_VersionNotChanged(t *testing.T) {
	defer setupAPIKeyRuleManager("v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/mod-api-key?version=v1", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
