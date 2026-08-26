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

package quota

import (
	"context"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceSyncManager_shouldResetByPeriod(t *testing.T) {
	m := &BalanceSyncManager{}

	t.Run("nil lastResetAt returns true", func(t *testing.T) {
		assert.True(t, m.shouldResetByPeriod(nil, "weekly", time.Now()))
	})

	t.Run("unknown period returns false", func(t *testing.T) {
		last := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.False(t, m.shouldResetByPeriod(&last, "daily", time.Now()))
	})

	t.Run("weekly same week returns false", func(t *testing.T) {
		// 2026-07-28 is Tuesday
		last := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
		now := time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)
		assert.False(t, m.shouldResetByPeriod(&last, "weekly", now))
	})

	t.Run("weekly different week returns true", func(t *testing.T) {
		last := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC) // Monday
		now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)   // next Monday
		assert.True(t, m.shouldResetByPeriod(&last, "weekly", now))
	})

	t.Run("monthly same month returns false", func(t *testing.T) {
		last := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
		now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
		assert.False(t, m.shouldResetByPeriod(&last, "monthly", now))
	})

	t.Run("monthly different month returns true", func(t *testing.T) {
		last := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
		now := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
		assert.True(t, m.shouldResetByPeriod(&last, "monthly", now))
	})
}

func TestGetWeekStart(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Monday",
			input:    time.Date(2026, 7, 27, 15, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Sunday",
			input:    time.Date(2026, 8, 2, 15, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Wednesday",
			input:    time.Date(2026, 7, 29, 15, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getWeekStart(tt.input))
		})
	}
}

func TestGetMonthStart(t *testing.T) {
	input := time.Date(2026, 7, 28, 15, 30, 0, 0, time.UTC)
	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, getMonthStart(input))
}
func TestBalanceSyncManager_WithMockRedis(t *testing.T) {
	// 保存并恢复全局 Redis 客户端，避免影响其他测试
	origClientSet := stateful.DefaultClientSet
	defer func() {
		stateful.DefaultClientSet = origClientSet
	}()
	mockRedis := stateful.NewMockRedisClient()
	stateful.DefaultClientSet = &stateful.ClientSet{RedisClient: mockRedis}

	ctx := context.Background()

	t.Run("SyncAllBalances calculates used from Redis", func(t *testing.T) {
		mockRedis.Reset()
		planID := int64(1)
		quota := float64(1000)
		apiKey := "ak-1"
		entityID := "ent-1"

		planStorager := &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{
					ID:        &planID,
					Quota:     &quota,
					Unlimited: lib.PBool(false),
				}}, nil
			},
		}
		apiKeyStorager := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				createdAt := time.Now()
				return []*api_key.APIKeyParam{{
					Key:         &apiKey,
					KeyCreateAt: &createdAt,
				}}, nil
			},
		}
		entityStorager := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
				return []*entity.EntityParam{{EntityID: &entityID}}, nil
			},
		}
		balanceStorager := &fakeQuotaBalanceStorager{
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}

		// 设置 Redis 中 API-Key 剩余 800，Entity 剩余 100
		_, err := mockRedis.IncrBy(stateful.AIUsedQuotaKey(apiKey), 800)
		require.NoError(t, err)
		_, err = mockRedis.IncrBy(stateful.AIUsedQuotaKey(entityID), 100)
		require.NoError(t, err)

		m := NewBalanceSyncManager(&fakeTxn{}, apiKeyStorager, balanceStorager, planStorager, entityStorager, quotacache.NewRedisQuotaCache(mockRedis), nil)
		require.NoError(t, m.SyncAllBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(100), *balanceStorager.updated[0].param.Used)      // 1000 - 900
		assert.Equal(t, float64(900), *balanceStorager.updated[0].param.Remaining) // 800 + 100
	})

	t.Run("ResetExpiredBalances resets balance and Redis", func(t *testing.T) {
		mockRedis.Reset()
		planID := int64(1)
		quota := float64(1000)
		apiKey := "ak-1"
		entityID := "ent-1"

		planStorager := &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{
					ID:          &planID,
					Quota:       &quota,
					Unlimited:   lib.PBool(false),
					ResetPeriod: lib.PString("monthly"),
				}}, nil
			},
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: &planID, Quota: &quota}, nil
			},
		}
		apiKeyStorager := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				createdAt := time.Now()
				return []*api_key.APIKeyParam{{
					ID:          lib.PString("ak-id-1"),
					Key:         &apiKey,
					KeyCreateAt: &createdAt,
				}}, nil
			},
		}
		entityStorager := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
				return []*entity.EntityParam{{EntityID: &entityID}}, nil
			},
		}
		balanceStorager := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				lastReset := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
				return &QuotaBalanceParam{
					ID:          lib.PInt64(1),
					QuotaPlanID: &planID,
					Used:        lib.PFloat64(500),
					Remaining:   lib.PFloat64(500),
					LastResetAt: &lastReset,
				}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}

		// 模拟 Redis 中已使用 500
		_, err := mockRedis.IncrBy(stateful.AIUsedQuotaKey(apiKey), 500)
		require.NoError(t, err)
		_, err = mockRedis.IncrBy(stateful.AIUsedQuotaKey(entityID), 0)
		require.NoError(t, err)

		m := NewBalanceSyncManager(&fakeTxn{}, apiKeyStorager, balanceStorager, planStorager, entityStorager, quotacache.NewRedisQuotaCache(mockRedis), nil)
		require.NoError(t, m.ResetExpiredBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(0), *balanceStorager.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStorager.updated[0].param.Remaining)
		assert.NotNil(t, balanceStorager.updated[0].param.LastResetAt)

		// Redis 应该被重置为配额总量
		apiKeyRemaining, _ := mockRedis.GetInt64(stateful.AIUsedQuotaKey(apiKey))
		assert.Equal(t, int64(1000), apiKeyRemaining)
		entityRemaining, _ := mockRedis.GetInt64(stateful.AIUsedQuotaKey(entityID))
		assert.Equal(t, int64(1000), entityRemaining)
	})

	t.Run("ResetExpiredBalances skips when not expired", func(t *testing.T) {
		mockRedis.Reset()
		planID := int64(1)
		quota := float64(1000)

		planStorager := &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{
					ID:          &planID,
					Quota:       &quota,
					Unlimited:   lib.PBool(false),
					ResetPeriod: lib.PString("monthly"),
				}}, nil
			},
		}
		balanceStorager := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				lastReset := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
				return &QuotaBalanceParam{QuotaPlanID: &planID, LastResetAt: &lastReset}, nil
			},
		}
		m := NewBalanceSyncManager(&fakeTxn{}, &fakeAPIKeyStorager{}, balanceStorager, planStorager, &fakeEntityStorager{}, quotacache.NewRedisQuotaCache(mockRedis), nil)

		require.NoError(t, m.ResetExpiredBalances(ctx))
		assert.Empty(t, balanceStorager.updated)
	})

	t.Run("ResetExpiredBalances skips when balance not found", func(t *testing.T) {
		mockRedis.Reset()
		planID := int64(1)
		quota := float64(1000)

		planStorager := &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{
					ID:          &planID,
					Quota:       &quota,
					Unlimited:   lib.PBool(false),
					ResetPeriod: lib.PString("monthly"),
				}}, nil
			},
		}
		balanceStorager := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return nil, nil
			},
		}
		m := NewBalanceSyncManager(&fakeTxn{}, &fakeAPIKeyStorager{}, balanceStorager, planStorager, &fakeEntityStorager{}, quotacache.NewRedisQuotaCache(mockRedis), nil)

		require.NoError(t, m.ResetExpiredBalances(ctx))
		assert.Empty(t, balanceStorager.updated)
	})
}

// fakeClock is a test Clock that returns a fixed time.
type fakeClock struct {
	t time.Time
}

func (f *fakeClock) Now() time.Time { return f.t }

// newResetTestManager creates a BalanceSyncManager configured for reset-period
// tests with a fake clock, mock Redis, and the supplied storagers.
func newResetTestManager(
	planStorager QuotaPlanStorager,
	balanceStorager QuotaBalanceStorager,
	apiKeyStorager api_key.APIKeyStorager,
	entityStorager entity.EntityStorager,
	clock Clock,
) *BalanceSyncManager {
	if clock == nil {
		clock = NewRealClock()
	}
	return NewBalanceSyncManager(
		&fakeTxn{},
		apiKeyStorager,
		balanceStorager,
		planStorager,
		entityStorager,
		quotacache.NewRedisQuotaCache(stateful.NewMockRedisClient()),
		clock,
	)
}

func TestResetExpiredBalances_WithFakeClock(t *testing.T) {
	ctx := context.Background()
	planID := int64(1)
	quota := float64(1000)
	apiKey := "ak-reset-1"
	entityID := "ent-reset-1"

	makePlanStorager := func(period string) *fakeQuotaPlanStorager {
		return &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{
					ID:          &planID,
					Quota:       &quota,
					Unlimited:   lib.PBool(false),
					ResetPeriod: lib.PString(period),
				}}, nil
			},
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: &planID, Quota: &quota}, nil
			},
		}
	}

	makeAPIKeyStorager := func() *fakeAPIKeyStorager {
		return &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				createdAt := time.Now()
				return []*api_key.APIKeyParam{{
					ID:          lib.PString("ak-id-1"),
					Key:         &apiKey,
					KeyCreateAt: &createdAt,
				}}, nil
			},
		}
	}

	makeEntityStorager := func() *fakeEntityStorager {
		return &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
				return []*entity.EntityParam{{EntityID: &entityID}}, nil
			},
		}
	}

	makeBalanceStorager := func(lastResetAt time.Time) *fakeQuotaBalanceStorager {
		return &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{
					ID:          lib.PInt64(1),
					QuotaPlanID: &planID,
					Used:        lib.PFloat64(500),
					Remaining:   lib.PFloat64(500),
					LastResetAt: &lastResetAt,
				}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}
	}

	t.Run("weekly resets on next Monday", func(t *testing.T) {
		lastReset := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC) // Monday
		clock := &fakeClock{t: time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)} // next Monday

		balanceStorager := makeBalanceStorager(lastReset)
		m := newResetTestManager(makePlanStorager("weekly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(0), *balanceStorager.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStorager.updated[0].param.Remaining)
		assert.NotNil(t, balanceStorager.updated[0].param.LastResetAt)
		assert.Equal(t, clock.Now(), *balanceStorager.updated[0].param.LastResetAt)
	})

	t.Run("weekly does not reset within same week", func(t *testing.T) {
		lastReset := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC) // Monday
		clock := &fakeClock{t: time.Date(2026, 7, 29, 9, 0, 0, 0, time.UTC)} // Wednesday

		balanceStorager := makeBalanceStorager(lastReset)
		m := newResetTestManager(makePlanStorager("weekly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))
		assert.Empty(t, balanceStorager.updated)
	})

	t.Run("monthly resets on next month first day", func(t *testing.T) {
		lastReset := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC) // Jul 31
		clock := &fakeClock{t: time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)} // Aug 1

		balanceStorager := makeBalanceStorager(lastReset)
		m := newResetTestManager(makePlanStorager("monthly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(0), *balanceStorager.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStorager.updated[0].param.Remaining)
		assert.Equal(t, clock.Now(), *balanceStorager.updated[0].param.LastResetAt)
	})

	t.Run("monthly does not reset within same month", func(t *testing.T) {
		lastReset := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC) // Aug 1
		clock := &fakeClock{t: time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)} // Aug 15

		balanceStorager := makeBalanceStorager(lastReset)
		m := newResetTestManager(makePlanStorager("monthly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))
		assert.Empty(t, balanceStorager.updated)
	})

	t.Run("monthly resets across year boundary", func(t *testing.T) {
		lastReset := time.Date(2025, 12, 31, 10, 0, 0, 0, time.UTC) // Dec 31
		clock := &fakeClock{t: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)} // Jan 1

		balanceStorager := makeBalanceStorager(lastReset)
		m := newResetTestManager(makePlanStorager("monthly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(0), *balanceStorager.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStorager.updated[0].param.Remaining)
		assert.Equal(t, clock.Now(), *balanceStorager.updated[0].param.LastResetAt)
	})

	t.Run("nil last_reset_at triggers reset", func(t *testing.T) {
		clock := &fakeClock{t: time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)}

		balanceStorager := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{
					ID:          lib.PInt64(1),
					QuotaPlanID: &planID,
					Used:        lib.PFloat64(500),
					Remaining:   lib.PFloat64(500),
					LastResetAt: nil,
				}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}
		m := newResetTestManager(makePlanStorager("weekly"), balanceStorager, makeAPIKeyStorager(), makeEntityStorager(), clock)

		require.NoError(t, m.ResetExpiredBalances(ctx))

		require.Len(t, balanceStorager.updated, 1)
		assert.Equal(t, float64(0), *balanceStorager.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStorager.updated[0].param.Remaining)
	})
}
