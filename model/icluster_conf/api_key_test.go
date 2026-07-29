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

package icluster_conf

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/model/shared"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

func ptrBool(b bool) *bool       { return &b }
func ptrInt64(i int64) *int64    { return &i }
func ptrString(s string) *string { return &s }

func TestGetRemainingQuota(t *testing.T) {
	t.Run("unlimited quota", func(t *testing.T) {
		remain, err := GetRemainingQuota(&APIKeyParam{UnlimitedQuota: ptrBool(true)})
		require.NoError(t, err)
		assert.Nil(t, remain)
	})

	t.Run("unlimited plan", func(t *testing.T) {
		remain, err := GetRemainingQuota(&APIKeyParam{
			QuotaPlan: &shared.QuotaPlanParam{Unlimited: ptrBool(true)},
		})
		require.NoError(t, err)
		assert.Nil(t, remain)
	})

	t.Run("nil quota plan", func(t *testing.T) {
		remain, err := GetRemainingQuota(&APIKeyParam{})
		require.NoError(t, err)
		assert.Nil(t, remain)
	})

	t.Run("no redis", func(t *testing.T) {
		stateful.DefaultClientSet = nil
		remain, err := GetRemainingQuota(&APIKeyParam{
			Key:       ptrString("key"),
			QuotaPlan: &shared.QuotaPlanParam{Quota: ptrInt64(100)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(100), *remain)
	})

	t.Run("redis returns remain", func(t *testing.T) {
		redisClient := stateful.NewMockRedisClient()
		redisClient.IncrBy(stateful.AIUsedQuotaKey("key"), 30)
		stateful.DefaultClientSet = &stateful.ClientSet{RedisClient: redisClient}
		defer func() { stateful.DefaultClientSet = nil }()

		remain, err := GetRemainingQuota(&APIKeyParam{
			Key:       ptrString("key"),
			QuotaPlan: &shared.QuotaPlanParam{Quota: ptrInt64(100)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(30), *remain)
	})

	t.Run("redis negative remain", func(t *testing.T) {
		redisClient := stateful.NewMockRedisClient()
		redisClient.IncrBy(stateful.AIUsedQuotaKey("key"), -5)
		stateful.DefaultClientSet = &stateful.ClientSet{RedisClient: redisClient}
		defer func() { stateful.DefaultClientSet = nil }()

		remain, err := GetRemainingQuota(&APIKeyParam{
			Key:       ptrString("key"),
			QuotaPlan: &shared.QuotaPlanParam{Quota: ptrInt64(100)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), *remain)
	})
}

func newAPIKeyManager(storager *fakeAPIKeyStorager) *APIKeyManager {
	return NewAPIKeyManager(&fakeTxn{}, storager, &fakeClusterStorager{}, &fakeQuotaPlanStorager{}, &fakeRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeEntityStorager{}, &fakeQuotaBalanceStorager{})
}

func TestAPIKeyManager_FetchAPIKeyList(t *testing.T) {
	ctx := context.Background()
	expected := []*APIKeyParam{{Key: ptrString("k1")}}
	store := &fakeAPIKeyStorager{
		fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
			return expected, nil
		},
	}
	m := newAPIKeyManager(store)
	got, err := m.FetchAPIKeyList(ctx, &APIKeyFilter{})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestAPIKeyManager_FetchAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{Key: ptrString("k1"), InnerID: ptrInt64(1)}}, nil
			},
		}
		m := newAPIKeyManager(store)
		got, err := m.FetchAPIKey(ctx, &APIKeyFilter{})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "k1", *got.Key)
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
		}
		m := newAPIKeyManager(store)
		got, err := m.FetchAPIKey(ctx, &APIKeyFilter{})
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestAPIKeyManager_DeleteAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.DeleteAPIKey(ctx, &APIKeyFilter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APIKey Record Not Exist")
	})

	t.Run("success", func(t *testing.T) {
		deleted := false
		quotaDeleted := false
		balanceDeleted := false
		rateLimitDeleted := false
		routeDeleted := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{
					Key:               ptrString("k1"),
					QuotaPlanID:       ptrInt64(1),
					RateLimitPolicyID: ptrInt64(2),
					RouteRulesID:      ptrInt64(3),
				}}, nil
			},
			deleteAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter) error {
				deleted = true
				return nil
			},
		}
		m := NewAPIKeyManager(&fakeTxn{}, store, &fakeClusterStorager{},
			&fakeQuotaPlanStorager{
				deleteQuotaPlanFn: func(ctx context.Context, id int64) error {
					quotaDeleted = true
					return nil
				},
			},
			&fakeRateLimitPolicyStorager{
				deleteRateLimitPolicyFn: func(ctx context.Context, id int64) error {
					rateLimitDeleted = true
					return nil
				},
			},
			&fakeRouteRulesStorager{
				deleteRouteRulesFn: func(ctx context.Context, id int64) error {
					routeDeleted = true
					return nil
				},
			},
			&fakeEntityStorager{},
			&fakeQuotaBalanceStorager{
				deleteQuotaBalanceFn: func(ctx context.Context, quotaPlanID int64) error {
					balanceDeleted = true
					return nil
				},
			},
		)
		err := m.DeleteAPIKey(ctx, &APIKeyFilter{})
		require.NoError(t, err)
		assert.True(t, deleted)
		assert.True(t, balanceDeleted)
		assert.True(t, quotaDeleted)
		assert.True(t, rateLimitDeleted)
		assert.True(t, routeDeleted)
	})
}

func TestAPIKeyManager_UpdateAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.UpdateAPIKey(ctx, &APIKeyFilter{}, &APIKeyParam{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API-Key Record Not Exist")
	})

	t.Run("success", func(t *testing.T) {
		updated := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{
					Key:       ptrString("k1"),
					InnerID:   ptrInt64(1),
					QuotaPlan: &shared.QuotaPlanParam{Unlimited: ptrBool(true)},
				}}, nil
			},
			updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
				updated = true
				assert.Nil(t, param.Key)
				return 1, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.UpdateAPIKey(ctx, &APIKeyFilter{}, &APIKeyParam{Key: ptrString("should-be-cleared")})
		require.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("update quota plan", func(t *testing.T) {
		updated := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{
					Key:         ptrString("k1"),
					InnerID:     ptrInt64(1),
					QuotaPlanID: ptrInt64(10),
				}}, nil
			},
			updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
				updated = true
				return 1, nil
			},
		}
		quotaPlanStore := &fakeQuotaPlanStorager{
			updateQuotaPlanFn: func(ctx context.Context, id int64, param *shared.QuotaPlanParam) (int64, error) {
				assert.Equal(t, int64(10), id)
				return 1, nil
			},
		}
		m := NewAPIKeyManager(&fakeTxn{}, store, &fakeClusterStorager{}, quotaPlanStore, &fakeRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeEntityStorager{}, &fakeQuotaBalanceStorager{})
		err := m.UpdateAPIKey(ctx, &APIKeyFilter{}, &APIKeyParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: ptrInt64(100)},
		})
		require.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("create new quota plan", func(t *testing.T) {
		created := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{
					Key:     ptrString("k1"),
					InnerID: ptrInt64(1),
				}}, nil
			},
			updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
				assert.Equal(t, int64(20), *param.QuotaPlanID)
				return 1, nil
			},
		}
		quotaPlanStore := &fakeQuotaPlanStorager{
			createQuotaPlanFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				created = true
				return 20, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			createQuotaBalanceFn: func(ctx context.Context, quotaPlanID int64, remaining *int64) error {
				assert.Equal(t, int64(20), quotaPlanID)
				assert.Equal(t, int64(100), *remaining)
				return nil
			},
		}
		m := NewAPIKeyManager(&fakeTxn{}, store, &fakeClusterStorager{}, quotaPlanStore, &fakeRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeEntityStorager{}, balanceStore)
		err := m.UpdateAPIKey(ctx, &APIKeyFilter{}, &APIKeyParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: ptrInt64(100)},
		})
		require.NoError(t, err)
		assert.True(t, created)
	})
}

func TestAPIKeyManager_CreateAPIKey(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		created := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
			createAPIKeyFn: func(ctx context.Context, param *APIKeyParam) (int64, error) {
				created = true
				return 1, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:          ptrString("id1"),
			ProductName: ptrString("test"),
			Key:         ptrString("key"),
		})
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("duplicate id", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{}}, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:          ptrString("id1"),
			ProductName: ptrString("test"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Duplicate id")
	})

	t.Run("entity not found", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
		}
		entityStore := &fakeEntityStorager{
			fetchEntityFn: func(ctx context.Context, filter *shared.EntityFilter) (*shared.EntitySummary, error) {
				return nil, nil
			},
		}
		m := NewAPIKeyManager(&fakeTxn{}, store, &fakeClusterStorager{}, &fakeQuotaPlanStorager{}, &fakeRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, entityStore, &fakeQuotaBalanceStorager{})
		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:          ptrString("id1"),
			ProductName: ptrString("test"),
			EntityID:    ptrString("e1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Entity not found")
	})

	t.Run("duplicate key value", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				if filter.Key != nil {
					return []*APIKeyParam{{Key: ptrString("key1")}}, nil
				}
				return nil, nil
			},
			fetchAPIKeyTokenListFn: func(ctx context.Context, filter *APIKeyTokenFilter) ([]*APIKeyTokenParam, error) {
				return nil, nil
			},
		}
		m := newAPIKeyManager(store)
		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:          ptrString("id1"),
			ProductName: ptrString("test"),
			Key:         ptrString("key1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("create quota plan and route rules", func(t *testing.T) {
		createdKey := false
		quotaCreated := false
		rateLimitCreated := false
		routeCreated := false
		balanceCreated := false
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, nil
			},
			fetchAPIKeyTokenListFn: func(ctx context.Context, filter *APIKeyTokenFilter) ([]*APIKeyTokenParam, error) {
				return nil, nil
			},
			createAPIKeyFn: func(ctx context.Context, param *APIKeyParam) (int64, error) {
				createdKey = true
				return 1, nil
			},
		}
		quotaPlanStore := &fakeQuotaPlanStorager{
			createQuotaPlanFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				quotaCreated = true
				return 10, nil
			},
		}
		rateLimitStore := &fakeRateLimitPolicyStorager{
			createRateLimitPolicyFn: func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) {
				rateLimitCreated = true
				return 20, nil
			},
		}
		routeRulesStore := &fakeRouteRulesStorager{
			createRouteRulesFn: func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
				routeCreated = true
				assert.Equal(t, shared.RouteRulesTypeAPIKey, ruleType)
				return 30, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			createQuotaBalanceFn: func(ctx context.Context, quotaPlanID int64, remaining *int64) error {
				balanceCreated = true
				return nil
			},
		}
		m := NewAPIKeyManager(&fakeTxn{}, store, &fakeClusterStorager{}, quotaPlanStore, rateLimitStore, routeRulesStore, &fakeEntityStorager{}, balanceStore)
		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:              ptrString("id1"),
			ProductName:     ptrString("test"),
			Key:             ptrString("key1"),
			QuotaPlan:       &shared.QuotaPlanParam{Quota: ptrInt64(100)},
			RateLimitPolicy: &shared.RateLimitPolicyParam{Enabled: ptrBool(true)},
			RouteRules:      &shared.RouteRulesParam{Enabled: ptrBool(true)},
		})
		require.NoError(t, err)
		assert.True(t, createdKey)
		assert.True(t, quotaCreated)
		assert.True(t, rateLimitCreated)
		assert.True(t, routeCreated)
		assert.True(t, balanceCreated)
	})

	t.Run("fetch api key list error", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return nil, errors.New("db down")
			},
		}
		m := newAPIKeyManager(store)
		err := m.CreateAPIKey(ctx, &APIKeyParam{ProductName: ptrString("test")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}

func TestAPIKeyManager_CreateAPIKeyToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		updated := false
		store := &fakeAPIKeyStorager{
			createAPIKeyTokenFn: func(ctx context.Context, param *APIKeyTokenParam) (int64, error) {
				assert.Equal(t, "key1", *param.Key)
				return 5, nil
			},
			updateAPIKeyTokenFn: func(ctx context.Context, filter *APIKeyTokenFilter, param *APIKeyTokenParam) error {
				updated = true
				assert.Equal(t, "key1-5", *param.Key)
				return nil
			},
		}
		m := newAPIKeyManager(store)
		id, err := m.CreateAPIKeyToken(ctx, &APIKeyTokenParam{Key: ptrString("key1"), CreatedAt: &fixedTestTime})
		require.NoError(t, err)
		assert.Equal(t, int64(5), id)
		assert.True(t, updated)
	})

	t.Run("create error", func(t *testing.T) {
		store := &fakeAPIKeyStorager{
			createAPIKeyTokenFn: func(ctx context.Context, param *APIKeyTokenParam) (int64, error) {
				return 0, errors.New("db down")
			},
		}
		m := newAPIKeyManager(store)
		_, err := m.CreateAPIKeyToken(ctx, &APIKeyTokenParam{Key: ptrString("key1"), CreatedAt: &fixedTestTime})
		require.Error(t, err)
	})
}
