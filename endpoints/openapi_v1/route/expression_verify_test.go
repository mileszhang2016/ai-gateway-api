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

package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/openapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iroute_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeRouteRuleStorager struct{}

func (f *fakeRouteRuleStorager) UpsertProductRule(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error {
	return nil
}

func (f *fakeRouteRuleStorager) FetchProductRule(ctx context.Context, product *ibasic.Product, clusterList []*icluster_conf.Cluster) (*iroute_conf.ProductRouteRule, error) {
	return nil, nil
}

func (f *fakeRouteRuleStorager) FetchRoutRules(ctx context.Context, products []*ibasic.Product, clusters []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
	return nil, nil
}

func setupRouteRuleManager() func() {
	old := container.RouteRuleManager
	container.RouteRuleManager = iroute_conf.NewRouteRuleManager(
		&testutil.FakeTxn{},
		&fakeRouteRuleStorager{},
		nil,
		nil,
		nil,
		nil,
	)
	return func() {
		container.RouteRuleManager = old
	}
}

func TestExpressionVerifyActionProcess_Success(t *testing.T) {
	defer setupRouteRuleManager()()

	req := httptest.NewRequest(http.MethodPatch, "/expression/verify", strings.NewReader(`{"expression": "default_t()"}`))
	data, err := ExpressionVerifyAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestExpressionVerifyActionProcess_InvalidExpression(t *testing.T) {
	defer setupRouteRuleManager()()

	req := httptest.NewRequest(http.MethodPatch, "/expression/verify", strings.NewReader(`{"expression": "invalid(@"}`))
	data, err := ExpressionVerifyAction(req)

	require.Error(t, err)
	result, ok := data.(*VerifyResult)
	require.True(t, ok)
	assert.Equal(t, 500, result.Code)
	assert.NotEmpty(t, result.Message)
}

func TestExpressionVerifyAction_InvalidJSON(t *testing.T) {
	defer setupRouteRuleManager()()

	req := httptest.NewRequest(http.MethodPatch, "/expression/verify", strings.NewReader(`not-json`))
	_, err := ExpressionVerifyAction(req)

	require.Error(t, err)
}
