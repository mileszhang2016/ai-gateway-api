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

package iroute_conf

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
)

func TestNewHostTableConf(t *testing.T) {
	productMap := map[int64]string{1: "p1", 2: "p2"}
	domains := []*Domain{
		{ProductID: 2, Name: "b.example.com"},
		{ProductID: 1, Name: "a.example.com"},
	}
	conf := newHostTableConf("v1", productMap, domains)
	require.NotNil(t, conf)
	assert.Equal(t, "v1", *conf.Version)
	assert.NotNil(t, conf.Hosts)
	assert.NotNil(t, conf.HostTags)
}

func TestNewDomainManager(t *testing.T) {
	m := NewDomainManager(&fakeTxn{}, &fakeDomainStorager{}, nil)
	require.NotNil(t, m)
}

func TestDomainManager_DomainList(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1}
	expected := []*Domain{{Name: "a.example.com"}}
	store := &fakeDomainStorager{
		fetchDomainsFn: func(ctx context.Context, param *DomainFilter) ([]*Domain, error) {
			assert.Equal(t, product, param.Product)
			return expected, nil
		},
	}
	m := NewDomainManager(&fakeTxn{}, store, nil)

	got, err := m.DomainList(ctx, &DomainFilter{Product: product})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestDomainManager_CreateDomain(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1}

	t.Run("success", func(t *testing.T) {
		called := false
		store := &fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *DomainFilter) ([]*Domain, error) {
				return nil, nil
			},
			createDomainFn: func(ctx context.Context, product *ibasic.Product, param *DomainParam) error {
				called = true
				assert.Equal(t, int64(1), *param.ProductID)
				return nil
			},
		}
		m := NewDomainManager(&fakeTxn{}, store, nil)
		require.NoError(t, m.CreateDomain(ctx, product, &DomainParam{Name: ptrString("a.example.com")}))
		assert.True(t, called)
	})

	t.Run("duplicate", func(t *testing.T) {
		store := &fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *DomainFilter) ([]*Domain, error) {
				return []*Domain{{Name: "a.example.com"}}, nil
			},
		}
		m := NewDomainManager(&fakeTxn{}, store, nil)
		err := m.CreateDomain(ctx, product, &DomainParam{Name: ptrString("a.example.com")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Domain Record Existed")
	})

	t.Run("wildcard covers plain", func(t *testing.T) {
		store := &fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *DomainFilter) ([]*Domain, error) {
				return []*Domain{{Name: "*.example.com"}}, nil
			},
		}
		m := NewDomainManager(&fakeTxn{}, store, nil)
		err := m.CreateDomain(ctx, product, &DomainParam{Name: ptrString("a.example.com")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Domain Name a.example.com Be Covered By Wildcard Domain *.example.com")
	})

	t.Run("plain covers wildcard", func(t *testing.T) {
		store := &fakeDomainStorager{
			fetchDomainsFn: func(ctx context.Context, param *DomainFilter) ([]*Domain, error) {
				return []*Domain{{Name: "a.example.com"}}, nil
			},
		}
		m := NewDomainManager(&fakeTxn{}, store, nil)
		err := m.CreateDomain(ctx, product, &DomainParam{Name: ptrString("*.example.com")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Domain Name a.example.com Be Covered By Wildcard Domain *.example.com")
	})
}

func TestDomainManager_DeleteDomain(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 1}

	t.Run("used by https config", func(t *testing.T) {
		m := NewDomainManager(&fakeTxn{}, &fakeDomainStorager{}, nil)
		err := m.DeleteDomain(ctx, product, &Domain{UsingAdvancedHsts: 1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Domain  Be Used By HTTPS Config")
	})

	t.Run("fetch rule error", func(t *testing.T) {
		rrm := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
				return nil, errors.New("db down")
			},
		}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		m := NewDomainManager(&fakeTxn{}, &fakeDomainStorager{}, rrm)
		err := m.DeleteDomain(ctx, product, &Domain{Name: "a.example.com"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})

	t.Run("used by route rule", func(t *testing.T) {
		rrm := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{
			fetchRoutRulesFn: func(ctx context.Context, products []*ibasic.Product, clusterList []*icluster_conf.Cluster) (map[int64]*ProductRouteRule, error) {
				return map[int64]*ProductRouteRule{1: {
					BasicRouteRules: []*BasicRouteRule{{HostNames: []string{"a.example.com"}}},
				}}, nil
			},
		}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		m := NewDomainManager(&fakeTxn{}, &fakeDomainStorager{}, rrm)
		err := m.DeleteDomain(ctx, product, &Domain{Name: "a.example.com"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Domain a.example.com Be Used By")
	})

	t.Run("success", func(t *testing.T) {
		called := false
		store := &fakeDomainStorager{
			deleteDomainFn: func(ctx context.Context, product *ibasic.Product, domain *Domain) error {
				called = true
				return nil
			},
		}
		rrm := NewRouteRuleManager(&fakeTxn{}, &fakeRouteRuleStorager{}, &fakeClusterStorager{}, &fakeProductStorager{}, nil, &fakeDomainStorager{})
		m := NewDomainManager(&fakeTxn{}, store, rrm)
		require.NoError(t, m.DeleteDomain(ctx, product, &Domain{Name: "a.example.com"}))
		assert.True(t, called)
	})
}

func TestDomainBeUsedInfo_String_Dependent(t *testing.T) {
	dbui := &DomainBeUsedInfo{
		domain:  &Domain{Name: "a.example.com"},
		RoutRule: &HostUsedInfo{Type: "BasicConditionExpression", Detail: "a.example.com"},
	}
	assert.Equal(t, "Domain a.example.com Be Used By BasicConditionExpression Rule a.example.com", dbui.String())
	typ, name := dbui.Dependent()
	assert.Equal(t, "BasicConditionExpression", typ)
	assert.Equal(t, "a.example.com", name)

	dbui2 := &DomainBeUsedInfo{domain: &Domain{Name: "b.example.com"}, hasHTTPSConfig: true}
	assert.Equal(t, "Domain b.example.com Be Used By HTTPS Config", dbui2.String())
	typ2, name2 := dbui2.Dependent()
	assert.Equal(t, "DomainHttpsConfig", typ2)
	assert.Empty(t, name2)
}
