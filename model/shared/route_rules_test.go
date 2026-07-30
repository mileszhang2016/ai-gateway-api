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

package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// fakeTxn implements itxn.TxnStorager for unit tests
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

var _ itxn.TxnStorager = (*fakeTxn)(nil)

// fakeRouteRulesStorager implements RouteRulesStorager
type fakeRouteRulesStorager struct {
	createFn    func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error)
	fetchFn     func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error)
	listFn      func(ctx context.Context, filter *RouteRulesFilter) ([]*RouteTableParam, int64, error)
	updateFn    func(ctx context.Context, id int64, param *RouteRulesParam) (int64, error)
	deleteFn    func(ctx context.Context, id int64) error
	fetchByIDFn func(ctx context.Context, id int64) (*RouteRulesParam, error)

	created     []createRouteRulesCall
	fetched     []fetchRouteRulesCall
	updated     []updateRouteRulesCall
	deleted     []int64
	fetchedByID []int64
}

type createRouteRulesCall struct {
	ruleType string
	owner    *string
	param    *RouteRulesParam
}

type fetchRouteRulesCall struct {
	ruleType string
	owner    *string
}

type updateRouteRulesCall struct {
	id    int64
	param *RouteRulesParam
}

func (s *fakeRouteRulesStorager) CreateRouteRules(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
	s.created = append(s.created, createRouteRulesCall{ruleType: ruleType, owner: owner, param: param})
	if s.createFn != nil {
		return s.createFn(ctx, ruleType, owner, param)
	}
	return 0, nil
}

func (s *fakeRouteRulesStorager) FetchRouteRules(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
	s.fetched = append(s.fetched, fetchRouteRulesCall{ruleType: ruleType, owner: owner})
	if s.fetchFn != nil {
		return s.fetchFn(ctx, ruleType, owner)
	}
	return nil, nil
}

func (s *fakeRouteRulesStorager) FetchRouteRulesList(ctx context.Context, filter *RouteRulesFilter) ([]*RouteTableParam, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, filter)
	}
	return nil, 0, nil
}

func (s *fakeRouteRulesStorager) UpdateRouteRules(ctx context.Context, id int64, param *RouteRulesParam) (int64, error) {
	s.updated = append(s.updated, updateRouteRulesCall{id: id, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, id, param)
	}
	return 0, nil
}

func (s *fakeRouteRulesStorager) DeleteRouteRules(ctx context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s *fakeRouteRulesStorager) FetchRouteRulesByID(ctx context.Context, id int64) (*RouteRulesParam, error) {
	s.fetchedByID = append(s.fetchedByID, id)
	if s.fetchByIDFn != nil {
		return s.fetchByIDFn(ctx, id)
	}
	return nil, nil
}

var _ RouteRulesStorager = (*fakeRouteRulesStorager)(nil)

func TestRouteRulesManager_CreateRouteRules(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			createFn: func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		id, err := m.CreateRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1"), validRouteRulesParam())
		require.NoError(t, err)
		assert.Equal(t, int64(1), id)
		require.Len(t, store.created, 1)
		assert.Equal(t, RouteRulesTypeAPIKey, store.created[0].ruleType)
	})

	t.Run("validation error", func(t *testing.T) {
		m := NewRouteRulesManager(&fakeTxn{}, &fakeRouteRulesStorager{})

		_, err := m.CreateRouteRules(ctx, RouteRulesTypeAPIKey, nil, &RouteRulesParam{
			Rules: []*AiRouteRuleParam{{}},
		})
		require.Error(t, err)
	})
}

func TestRouteRulesManager_UpdateRouteRules(t *testing.T) {
	ctx := context.Background()

	t.Run("create when not exists", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
				return 2, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		id, err := m.UpdateRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1"), validRouteRulesParam())
		require.NoError(t, err)
		assert.Equal(t, int64(2), id)
		assert.Len(t, store.created, 1)
	})

	t.Run("update when exists", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return &RouteRulesParam{ID: lib.PInt64(3)}, nil
			},
			updateFn: func(ctx context.Context, id int64, param *RouteRulesParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		id, err := m.UpdateRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1"), validRouteRulesParam())
		require.NoError(t, err)
		assert.Equal(t, int64(3), id)
		assert.Len(t, store.updated, 1)
	})

	t.Run("validation error", func(t *testing.T) {
		m := NewRouteRulesManager(&fakeTxn{}, &fakeRouteRulesStorager{})

		_, err := m.UpdateRouteRules(ctx, RouteRulesTypeAPIKey, nil, &RouteRulesParam{
			Rules: []*AiRouteRuleParam{{}},
		})
		require.Error(t, err)
	})
}

func TestRouteRulesManager_FetchRouteRules(t *testing.T) {
	ctx := context.Background()
	store := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
			return &RouteRulesParam{Enabled: lib.PBool(true)}, nil
		},
	}
	m := NewRouteRulesManager(&fakeTxn{}, store)

	rule, err := m.FetchRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1"))
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.True(t, *rule.Enabled)
}

func TestRouteRulesManager_DeleteRouteRules(t *testing.T) {
	ctx := context.Background()

	t.Run("delete existing", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return &RouteRulesParam{ID: lib.PInt64(5)}, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		require.NoError(t, m.DeleteRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1")))
		assert.Equal(t, []int64{5}, store.deleted)
	})

	t.Run("delete not existing", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return nil, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		require.NoError(t, m.DeleteRouteRules(ctx, RouteRulesTypeAPIKey, lib.PString("owner-1")))
		assert.Empty(t, store.deleted)
	})
}

func TestRouteRulesManager_FetchRouteRulesByID(t *testing.T) {
	ctx := context.Background()
	store := &fakeRouteRulesStorager{
		fetchByIDFn: func(ctx context.Context, id int64) (*RouteRulesParam, error) {
			return &RouteRulesParam{ID: lib.PInt64(7)}, nil
		},
	}
	m := NewRouteRulesManager(&fakeTxn{}, store)

	rule, err := m.FetchRouteRulesByID(ctx, 7)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, int64(7), *rule.ID)
}

func TestRouteRulesManager_ListRouteTables(t *testing.T) {
	ctx := context.Background()
	store := &fakeRouteRulesStorager{
		listFn: func(ctx context.Context, filter *RouteRulesFilter) ([]*RouteTableParam, int64, error) {
			return []*RouteTableParam{{Type: "apikey"}}, 1, nil
		},
	}
	m := NewRouteRulesManager(&fakeTxn{}, store)

	list, total, err := m.ListRouteTables(ctx, &RouteRulesFilter{})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), total)
}

func TestRouteRulesManager_GetGlobalRouteRules(t *testing.T) {
	ctx := context.Background()
	store := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
			assert.Equal(t, RouteRulesTypeGlobal, ruleType)
			return &RouteRulesParam{Enabled: lib.PBool(true)}, nil
		},
	}
	m := NewRouteRulesManager(&fakeTxn{}, store)

	rule, err := m.GetGlobalRouteRules(ctx)
	require.NoError(t, err)
	require.NotNil(t, rule)
}

func TestRouteRulesManager_SetGlobalRouteRules(t *testing.T) {
	ctx := context.Background()
	store := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
			return 9, nil
		},
		fetchByIDFn: func(ctx context.Context, id int64) (*RouteRulesParam, error) {
			return &RouteRulesParam{ID: lib.PInt64(9)}, nil
		},
	}
	m := NewRouteRulesManager(&fakeTxn{}, store)

	rule, err := m.SetGlobalRouteRules(ctx, validRouteRulesParam())
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, int64(9), *rule.ID)
}

func TestRouteRulesManager_EnsureGlobalRouteRules(t *testing.T) {
	ctx := context.Background()

	t.Run("creates default global route table when not exists", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
				return 10, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		require.NoError(t, m.EnsureGlobalRouteRules(ctx))
		require.Len(t, store.created, 1)
		assert.Equal(t, RouteRulesTypeGlobal, store.created[0].ruleType)
		require.NotNil(t, store.created[0].owner)
		assert.Equal(t, RouteRulesTypeGlobal, *store.created[0].owner)
		require.NotNil(t, store.created[0].param.Enabled)
		assert.False(t, *store.created[0].param.Enabled)
		assert.Empty(t, store.created[0].param.Rules)
	})

	t.Run("does nothing when global route table already exists", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return &RouteRulesParam{ID: lib.PInt64(1)}, nil
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		require.NoError(t, m.EnsureGlobalRouteRules(ctx))
		assert.Empty(t, store.created)
	})
}

func TestRouteRulesManager_validateRouteRules(t *testing.T) {
	m := NewRouteRulesManager(&fakeTxn{}, &fakeRouteRulesStorager{})

	t.Run("nil param", func(t *testing.T) {
		assert.NoError(t, m.validateRouteRules(nil))
	})

	t.Run("missing rule name", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{{
				Cond: lib.PString("default_t()"),
				Targets: []*AiRouteTargetParam{{
					ClusterName: lib.PString("cluster-1"),
					Weight:      lib.PInt(100),
				}},
			}},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("duplicate rule name", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{Name: lib.PString("rule-1"), Cond: lib.PString("cond1"), Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("c1"), Weight: lib.PInt(100)}}},
				{Name: lib.PString("rule-1"), Cond: lib.PString("cond2"), Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("c2"), Weight: lib.PInt(100)}}},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("missing cond", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{Name: lib.PString("rule-1"), Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("c1"), Weight: lib.PInt(100)}}},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("empty targets", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{Name: lib.PString("rule-1"), Cond: lib.PString("cond1")},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("missing target weight", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{Name: lib.PString("rule-1"), Cond: lib.PString("cond1"), Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("c1")}}},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("total weight not 100", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{Name: lib.PString("rule-1"), Cond: lib.PString("cond1"), Targets: []*AiRouteTargetParam{
					{ClusterName: lib.PString("c1"), Weight: lib.PInt(60)},
					{ClusterName: lib.PString("c2"), Weight: lib.PInt(30)},
				}},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("empty fallback cluster name", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{
					Name:      lib.PString("rule-1"),
					Cond:      lib.PString("cond1"),
					Targets:   []*AiRouteTargetParam{{ClusterName: lib.PString("c1"), Weight: lib.PInt(100)}},
					Fallbacks: []*AiRouteFallbackParam{{}},
				},
			},
		}
		assert.Error(t, m.validateRouteRules(param))
	})

	t.Run("valid rules", func(t *testing.T) {
		param := &RouteRulesParam{
			Rules: []*AiRouteRuleParam{
				{
					Name:    lib.PString("rule-1"),
					Cond:    lib.PString("cond1"),
					Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("c1"), Weight: lib.PInt(100)}},
				},
			},
		}
		assert.NoError(t, m.validateRouteRules(param))
	})
}

func TestRouteRulesManager_PropagateErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateRouteRules storage error", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			createFn: func(ctx context.Context, ruleType string, owner *string, param *RouteRulesParam) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		_, err := m.CreateRouteRules(ctx, RouteRulesTypeAPIKey, nil, validRouteRulesParam())
		require.Error(t, err)
	})

	t.Run("FetchRouteRules storage error", func(t *testing.T) {
		store := &fakeRouteRulesStorager{
			fetchFn: func(ctx context.Context, ruleType string, owner *string) (*RouteRulesParam, error) {
				return nil, errors.New("db error")
			},
		}
		m := NewRouteRulesManager(&fakeTxn{}, store)

		_, err := m.FetchRouteRules(ctx, RouteRulesTypeAPIKey, nil)
		require.Error(t, err)
	})
}

func validRouteRulesParam() *RouteRulesParam {
	return &RouteRulesParam{
		Enabled: lib.PBool(true),
		Rules: []*AiRouteRuleParam{
			{
				Name:    lib.PString("rule-1"),
				Cond:    lib.PString("default_t()"),
				Targets: []*AiRouteTargetParam{{ClusterName: lib.PString("cluster-1"), Weight: lib.PInt(100)}},
			},
		},
	}
}
