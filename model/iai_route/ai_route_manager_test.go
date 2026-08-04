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

package iai_route

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/iroute_conf"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// fakeAIRouteRuleStorager implements AIRouteRuleStorager
type fakeAIRouteRuleStorager struct {
	createFn func(ctx context.Context, param []*Rule) error
	fetchFn  func(ctx context.Context, filter *AIRouteFilter) ([]*Rule, error)
}

func (s *fakeAIRouteRuleStorager) FetchAIRouteRules(ctx context.Context, filter *AIRouteFilter) ([]*Rule, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeAIRouteRuleStorager) CreateAIRouteRules(ctx context.Context, param []*Rule) error {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return nil
}

var _ AIRouteRuleStorager = (*fakeAIRouteRuleStorager)(nil)

// fakeTxn implements itxn.TxnStorager for unit tests
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

// fakeRouteRuleStorager implements iroute_conf.RouteRuleStorager
type fakeRouteRuleStorager struct {
	upsertFn func(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error
	fetchFn  func(ctx context.Context, product *ibasic.Product, clusterList []*icluster_conf.Cluster) (*iroute_conf.ProductRouteRule, error)
	listFn   func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error)
}

func (s *fakeRouteRuleStorager) UpsertProductRule(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error {
	if s.upsertFn != nil {
		return s.upsertFn(ctx, product, rule)
	}
	return nil
}

func (s *fakeRouteRuleStorager) FetchProductRule(ctx context.Context, product *ibasic.Product, clusterList []*icluster_conf.Cluster) (*iroute_conf.ProductRouteRule, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, product, clusterList)
	}
	return nil, nil
}

func (s *fakeRouteRuleStorager) FetchRoutRules(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*iroute_conf.ProductRouteRule, error) {
	if s.listFn != nil {
		return s.listFn(ctx, products, clusterList)
	}
	return nil, nil
}

var _ iroute_conf.RouteRuleStorager = (*fakeRouteRuleStorager)(nil)

func TestAIRouteRuleManager_CreateAIRouteRule(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		aiStore := &fakeAIRouteRuleStorager{
			createFn: func(ctx context.Context, param []*Rule) error {
				return nil
			},
		}
		routeStore := &fakeRouteRuleStorager{
			upsertFn: func(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error {
				return nil
			},
		}
		m := NewAIRouteRuleManager(&fakeTxn{}, aiStore, nil, routeStore)

		product := &ibasic.Product{Name: "AI"}
		err := m.CreateAIRouteRule(ctx, []*Rule{{Name: "rule_1"}}, product, nil)
		require.NoError(t, err)
	})

	t.Run("ai route create error", func(t *testing.T) {
		aiStore := &fakeAIRouteRuleStorager{
			createFn: func(ctx context.Context, param []*Rule) error {
				return errors.New("ai db error")
			},
		}
		m := NewAIRouteRuleManager(&fakeTxn{}, aiStore, nil, &fakeRouteRuleStorager{})

		err := m.CreateAIRouteRule(ctx, []*Rule{{Name: "rule_1"}}, &ibasic.Product{Name: "AI"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ai db error")
	})

	t.Run("route upsert error", func(t *testing.T) {
		aiStore := &fakeAIRouteRuleStorager{
			createFn: func(ctx context.Context, param []*Rule) error {
				return nil
			},
		}
		routeStore := &fakeRouteRuleStorager{
			upsertFn: func(ctx context.Context, product *ibasic.Product, rule *iroute_conf.ProductRouteRule) error {
				return errors.New("route db error")
			},
		}
		m := NewAIRouteRuleManager(&fakeTxn{}, aiStore, nil, routeStore)

		err := m.CreateAIRouteRule(ctx, []*Rule{{Name: "rule_1"}}, &ibasic.Product{Name: "AI"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "route db error")
	})
}

func TestAIRouteRuleManager_FetchAIRouteRules(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		aiStore := &fakeAIRouteRuleStorager{
			fetchFn: func(ctx context.Context, filter *AIRouteFilter) ([]*Rule, error) {
				return []*Rule{{Name: "rule_1"}}, nil
			},
		}
		m := NewAIRouteRuleManager(&fakeTxn{}, aiStore, nil, &fakeRouteRuleStorager{})

		rules, err := m.FetchAIRouteRules(ctx, &AIRouteFilter{ProductName: lib.PString("AI")})
		require.NoError(t, err)
		assert.Len(t, rules, 1)
	})

	t.Run("fetch error", func(t *testing.T) {
		aiStore := &fakeAIRouteRuleStorager{
			fetchFn: func(ctx context.Context, filter *AIRouteFilter) ([]*Rule, error) {
				return nil, errors.New("db error")
			},
		}
		m := NewAIRouteRuleManager(&fakeTxn{}, aiStore, nil, &fakeRouteRuleStorager{})

		_, err := m.FetchAIRouteRules(ctx, &AIRouteFilter{})
		require.Error(t, err)
	})
}

