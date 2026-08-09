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

package iroute_conf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
)

func ptrString(s string) *string { return &s }

func TestProductRouteRule_HostBeUsed(t *testing.T) {
	prr := &ProductRouteRule{
		BasicRouteRules: []*BasicRouteRule{
			{HostNames: []string{"a.example.com"}},
		},
		AdvanceRouteRules: []*AdvanceRouteRule{
			{Expression: `req_host_in("b.example.com")`},
		},
	}

	t.Run("basic match", func(t *testing.T) {
		info := prr.HostBeUsed("a.example.com")
		require.NotNil(t, info)
		assert.Equal(t, "BasicConditionExpression", info.Type)
		assert.Equal(t, "a.example.com", info.Detail)
	})

	t.Run("advance match", func(t *testing.T) {
		info := prr.HostBeUsed("b.example.com")
		require.NotNil(t, info)
		assert.Equal(t, "AdvanceConditionExpression", info.Type)
	})

	t.Run("no match", func(t *testing.T) {
		assert.Nil(t, prr.HostBeUsed("c.example.com"))
	})
}

func TestProductRouteRule_Convert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		prr := &ProductRouteRule{
			BasicRouteRules: []*BasicRouteRule{
				{HostNames: []string{"a.example.com"}, Paths: []string{"/"}, ClusterName: "c1"},
			},
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: "c2"},
			},
		}
		cr, err := prr.Convert()
		require.NoError(t, err)
		require.NotNil(t, cr)
		assert.Len(t, cr.BasicRouteRuleFiles, 1)
		assert.Len(t, cr.AdvancedRouteRuleFiles, 1)
		assert.ElementsMatch(t, []string{"c1", "c2"}, cr.ReferClusterNames)
	})

	t.Run("missing default expression", func(t *testing.T) {
		prr := &ProductRouteRule{
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "req_host_in(\"a\")", ClusterName: "c1"},
			},
		}
		_, err := prr.Convert()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Last ForwardRule Expression Must Be default_t()")
	})
}

func TestNewRouteRuleManager(t *testing.T) {
	m := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
	require.NotNil(t, m)
}

func TestRouteRuleManager_ExpressionVerify(t *testing.T) {
	m := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})

	t.Run("valid", func(t *testing.T) {
		require.NoError(t, m.ExpressionVerify(context.Background(), "default_t()"))
	})

	t.Run("invalid", func(t *testing.T) {
		err := m.ExpressionVerify(context.Background(), "not_a_valid_expr(")
		require.Error(t, err)
	})
}

func TestRouteRuleManager_UpsertProductRule(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1, Name: "p1"}

	t.Run("convert error", func(t *testing.T) {
		m := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		err := m.UpsertProductRule(ctx, product, &ProductRouteRule{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Last ForwardRule Expression Must Be default_t()")
	})

	t.Run("cluster not exist", func(t *testing.T) {
		store := &fakeRouteRuleStorager{}
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, clusterStore, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		rule := &ProductRouteRule{
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: "c1"},
			},
		}
		err := m.UpsertProductRule(ctx, product, rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cluster c1 Not Exist")
	})

	t.Run("cluster not ready", func(t *testing.T) {
		store := &fakeRouteRuleStorager{}
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{{ID: 1, Name: "c1", Ready: false}}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, clusterStore, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		rule := &ProductRouteRule{
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: "c1"},
			},
		}
		err := m.UpsertProductRule(ctx, product, rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cluster c1 Not Ready")
	})

	t.Run("success", func(t *testing.T) {
		called := false
		store := &fakeRouteRuleStorager{
			upsertProductRuleFn: func(ctx context.Context, product *ibasic.Product, rule *ProductRouteRule) error {
				called = true
				assert.Equal(t, int64(1), rule.AdvanceRouteRules[0].ClusterID)
				return nil
			},
		}
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{{ID: 1, Name: "c1", Ready: true}}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, clusterStore, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		rule := &ProductRouteRule{
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: "c1"},
			},
		}
		require.NoError(t, m.UpsertProductRule(ctx, product, rule))
		assert.True(t, called)
	})

	t.Run("system keep cluster", func(t *testing.T) {
		called := false
		store := &fakeRouteRuleStorager{
			upsertProductRuleFn: func(ctx context.Context, product *ibasic.Product, rule *ProductRouteRule) error {
				called = true
				return nil
			},
		}
		clusterStore := &fakeClusterStorager{
			fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
				return []*icluster_conf.Cluster{}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, clusterStore, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		rule := &ProductRouteRule{
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: icluster_conf.RouteAdvancedModeClusterName4DP},
			},
		}
		require.NoError(t, m.UpsertProductRule(ctx, product, rule))
		assert.True(t, called)
	})
}

func TestRouteRuleManager_ClusterDeleteChecker(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1}
	cluster := &icluster_conf.Cluster{ID: 1, Name: "c1"}

	t.Run("no rule", func(t *testing.T) {
		store := &fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
				return nil, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		require.NoError(t, m.ClusterDeleteChecker(ctx, product, cluster))
	})

	t.Run("advance rule refers", func(t *testing.T) {
		store := &fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
				return map[int64]*ProductRouteRule{1: {
					AdvanceRouteRules: []*AdvanceRouteRule{{Name: "rule1"}},
				}}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		err := m.ClusterDeleteChecker(ctx, product, cluster)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rule rule1 Refer To This Cluster")
	})

	t.Run("basic rule refers", func(t *testing.T) {
		store := &fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
				return map[int64]*ProductRouteRule{1: {
					BasicRouteRules: []*BasicRouteRule{{Description: "rule2"}},
				}}, nil
			},
		}
		m := NewRouteRuleManager(&fakeTxn{}, store, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		err := m.ClusterDeleteChecker(ctx, product, cluster)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Rule rule2 Refer To This Cluster")
	})
}

func TestRouteRuleManager_FetchProductRule(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1}
	expected := &ProductRouteRule{BasicRouteRules: []*BasicRouteRule{{}}}

	store := &fakeRouteRuleStorager{
		fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
			assert.Len(t, products, 1)
			return map[int64]*ProductRouteRule{1: expected}, nil
		},
	}
	clusterStore := &fakeClusterStorager{
		fetchClusterListFn: func(ctx context.Context, param *icluster_conf.ClusterFilter) ([]*icluster_conf.Cluster, error) {
			return []*icluster_conf.Cluster{{ID: 1, Name: "c1"}}, nil
		},
	}
	m := NewRouteRuleManager(&fakeTxn{}, store, clusterStore, &fakeProductStorager{}, nil, &fakeDomainStorager{})

	got, err := m.FetchProductRule(ctx, product)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestNewRouteTableFile(t *testing.T) {
	productMap := map[int64]string{1: "p1", 2: "p2"}
	routeRules := map[int64]*ProductRouteRule{
		1: {
			AdvanceRouteRules: []*AdvanceRouteRule{
				{Expression: "default_t()", ClusterName: "c1"},
			},
		},
	}
	got := newRouteTableFile("v1", productMap, routeRules)
	require.NotNil(t, got)
	assert.Equal(t, "v1", *got.Version)
	require.NotNil(t, got.ProductRule)
	require.Contains(t, *got.ProductRule, "p1")
}
