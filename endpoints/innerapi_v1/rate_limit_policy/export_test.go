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

package rate_limit_policy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rainway-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/rate_limit_policy"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeRateLimitPolicyStorager struct{}

func (f *fakeRateLimitPolicyStorager) CreateRateLimitPolicy(ctx context.Context, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
	return 0, nil
}

func (f *fakeRateLimitPolicyStorager) FetchRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) (*rate_limit_policy.RateLimitPolicyParam, error) {
	return nil, nil
}

func (f *fakeRateLimitPolicyStorager) FetchRateLimitPolicyList(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) ([]*rate_limit_policy.RateLimitPolicyParam, error) {
	return nil, nil
}

func (f *fakeRateLimitPolicyStorager) UpdateRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
	return 0, nil
}

func (f *fakeRateLimitPolicyStorager) DeleteRateLimitPolicy(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) error {
	return nil
}

type fakeAPIKeyStoragerForRateLimit struct{}

func (f *fakeAPIKeyStoragerForRateLimit) FetchAPIKeyList(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
	return nil, nil
}

func (f *fakeAPIKeyStoragerForRateLimit) CreateAPIKey(ctx context.Context, param *api_key.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRateLimit) UpdateAPIKey(ctx context.Context, filter *api_key.APIKeyFilter, param *api_key.APIKeyParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRateLimit) DeleteAPIKey(ctx context.Context, filter *api_key.APIKeyFilter) error {
	return nil
}

func (f *fakeAPIKeyStoragerForRateLimit) CreateAPIKeyToken(ctx context.Context, param *api_key.APIKeyTokenParam) (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeyStoragerForRateLimit) UpdateAPIKeyToken(ctx context.Context, filter *api_key.APIKeyTokenFilter, param *api_key.APIKeyTokenParam) error {
	return nil
}

func (f *fakeAPIKeyStoragerForRateLimit) FetchAPIKeyTokenList(ctx context.Context, filter *api_key.APIKeyTokenFilter) ([]*api_key.APIKeyTokenParam, error) {
	return nil, nil
}

type fakeEntityStoragerForRateLimit struct{}

func (f *fakeEntityStoragerForRateLimit) CreateEntity(ctx context.Context, param *entity.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForRateLimit) FetchEntity(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForRateLimit) FetchEntityList(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
	return nil, nil
}

func (f *fakeEntityStoragerForRateLimit) UpdateEntity(ctx context.Context, filter *entity.EntityFilter, param *entity.EntityParam) (int64, error) {
	return 0, nil
}

func (f *fakeEntityStoragerForRateLimit) DeleteEntity(ctx context.Context, filter *entity.EntityFilter) error {
	return nil
}

func setupRateLimitPolicyManager(version string) func() {
	old := container.RateLimitPolicyManager
	container.RateLimitPolicyManager = rate_limit_policy.NewRateLimitPolicyManager(
		&testutil.FakeTxn{},
		&fakeRateLimitPolicyStorager{},
		&fakeAPIKeyStoragerForRateLimit{},
		&fakeEntityStoragerForRateLimit{},
		testutil.NewVersionControlManager(version),
	)
	return func() {
		container.RateLimitPolicyManager = old
	}
}

func TestExportAction(t *testing.T) {
	defer setupRateLimitPolicyManager("v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/rate-limit-policy?version=", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*rate_limit_policy.ExportRateLimitPolicyConfig)
	require.True(t, ok)
	assert.Equal(t, "v2", conf.Version)
	assert.Contains(t, conf.Config, "AI_product")
}

func TestExportAction_VersionNotChanged(t *testing.T) {
	defer setupRateLimitPolicyManager("v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/rate-limit-policy?version=v1", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
