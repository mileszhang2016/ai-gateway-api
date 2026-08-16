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

package quota

import (
	"context"
	"errors"
	"testing"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/entity"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/rate_limit_policy"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuotaPlanStoragerAdapter(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateQuotaPlan", func(t *testing.T) {
		inner := &fakeQuotaPlanStorager{
			createFn: func(ctx context.Context, param *QuotaPlanParam) (int64, error) {
				require.NotNil(t, param.Quota)
				assert.Equal(t, float64(100), *param.Quota)
				return 7, nil
			},
		}
		adapter := NewQuotaPlanStoragerAdapter(inner)
		id, err := adapter.CreateQuotaPlan(ctx, &shared.QuotaPlanParam{Quota: lib.PFloat64(100)})
		require.NoError(t, err)
		assert.Equal(t, int64(7), id)
	})

	t.Run("UpdateQuotaPlan", func(t *testing.T) {
		inner := &fakeQuotaPlanStorager{
			updateFn: func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
				assert.Equal(t, int64(7), *filter.ID)
				assert.Equal(t, float64(200), *param.Quota)
				return 1, nil
			},
		}
		adapter := NewQuotaPlanStoragerAdapter(inner)
		affected, err := adapter.UpdateQuotaPlan(ctx, 7, &shared.QuotaPlanParam{Quota: lib.PFloat64(200)})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("DeleteQuotaPlan", func(t *testing.T) {
		inner := &fakeQuotaPlanStorager{
			deleteFn: func(ctx context.Context, filter *QuotaPlanFilter) error {
				assert.Equal(t, int64(7), *filter.ID)
				return nil
			},
		}
		adapter := NewQuotaPlanStoragerAdapter(inner)
		require.NoError(t, adapter.DeleteQuotaPlan(ctx, 7))
	})

	t.Run("FetchQuotaPlan", func(t *testing.T) {
		inner := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				assert.Equal(t, int64(7), *filter.ID)
				return &QuotaPlanParam{Quota: lib.PFloat64(300)}, nil
			},
		}
		adapter := NewQuotaPlanStoragerAdapter(inner)
		result, err := adapter.FetchQuotaPlan(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(300), *result.Quota)
	})

	t.Run("FetchQuotaPlan not found", func(t *testing.T) {
		inner := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return nil, nil
			},
		}
		adapter := NewQuotaPlanStoragerAdapter(inner)
		result, err := adapter.FetchQuotaPlan(ctx, 7)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestRateLimitPolicyStoragerAdapter(t *testing.T) {
	ctx := context.Background()

	buildSharedPolicy := func() *shared.RateLimitPolicyParam {
		return &shared.RateLimitPolicyParam{
			Enabled: lib.PBool(true),
			Rules: &shared.RateLimitRules{
				MaxConcurrency: lib.PInt(10),
				TpmConfigs: []shared.TPMConfig{
					{Name: "tpm-1", Model: "*", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
				},
				RpmConfigs: []shared.RPMConfig{
					{Name: "rpm-1", Model: "*", WindowMinutes: 1, MaxRequests: 10},
				},
			},
		}
	}

	t.Run("CreateRateLimitPolicy", func(t *testing.T) {
		inner := &fakeRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
				assert.True(t, *param.Enabled)
				assert.Equal(t, 10, *param.MaxConcurrency)
				assert.Len(t, param.TpmConfigs, 1)
				assert.Len(t, param.RpmConfigs, 1)
				return 8, nil
			},
		}
		adapter := NewRateLimitPolicyStoragerAdapter(inner)
		id, err := adapter.CreateRateLimitPolicy(ctx, buildSharedPolicy())
		require.NoError(t, err)
		assert.Equal(t, int64(8), id)
	})

	t.Run("UpdateRateLimitPolicy", func(t *testing.T) {
		inner := &fakeRateLimitPolicyStorager{
			updateFn: func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter, param *rate_limit_policy.RateLimitPolicyParam) (int64, error) {
				assert.Equal(t, int64(8), *filter.ID)
				assert.Equal(t, 10, *param.MaxConcurrency)
				return 1, nil
			},
		}
		adapter := NewRateLimitPolicyStoragerAdapter(inner)
		affected, err := adapter.UpdateRateLimitPolicy(ctx, 8, buildSharedPolicy())
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("DeleteRateLimitPolicy", func(t *testing.T) {
		inner := &fakeRateLimitPolicyStorager{
			deleteFn: func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) error {
				assert.Equal(t, int64(8), *filter.ID)
				return nil
			},
		}
		adapter := NewRateLimitPolicyStoragerAdapter(inner)
		require.NoError(t, adapter.DeleteRateLimitPolicy(ctx, 8))
	})

	t.Run("FetchRateLimitPolicy", func(t *testing.T) {
		inner := &fakeRateLimitPolicyStorager{
			fetchFn: func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) (*rate_limit_policy.RateLimitPolicyParam, error) {
				assert.Equal(t, int64(8), *filter.ID)
				return &rate_limit_policy.RateLimitPolicyParam{
					Enabled:        lib.PBool(true),
					MaxConcurrency: lib.PInt(10),
					TpmConfigs:     []rate_limit_policy.TPMConfig{{Name: "tpm-1", Model: "*", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1}},
					RpmConfigs:     []rate_limit_policy.RPMConfig{{Name: "rpm-1", Model: "*", WindowMinutes: 1, MaxRequests: 10}},
				}, nil
			},
		}
		adapter := NewRateLimitPolicyStoragerAdapter(inner)
		result, err := adapter.FetchRateLimitPolicy(ctx, 8)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Rules)
		assert.Equal(t, 10, *result.Rules.MaxConcurrency)
	})

	t.Run("FetchRateLimitPolicy error propagates", func(t *testing.T) {
		inner := &fakeRateLimitPolicyStorager{
			fetchFn: func(ctx context.Context, filter *rate_limit_policy.RateLimitPolicyFilter) (*rate_limit_policy.RateLimitPolicyParam, error) {
				return nil, errors.New("db error")
			},
		}
		adapter := NewRateLimitPolicyStoragerAdapter(inner)
		_, err := adapter.FetchRateLimitPolicy(ctx, 8)
		require.Error(t, err)
	})
}

func TestEntityStoragerAdapter(t *testing.T) {
	ctx := context.Background()

	t.Run("FetchEntity", func(t *testing.T) {
		entityID := "ent-1"
		entityName := "entity-one"
		entityType := "tenant"
		inner := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
				assert.Equal(t, entityID, *filter.EntityID)
				return &entity.EntityParam{
					EntityID: &entityID,
					Name:     &entityName,
					Type:     &entityType,
				}, nil
			},
		}
		adapter := NewEntityStoragerAdapter(inner)
		result, err := adapter.FetchEntity(ctx, &shared.EntityFilter{EntityID: &entityID})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, entityID, *result.ID)
		assert.Equal(t, entityName, *result.Name)
		assert.Equal(t, entityType, *result.Type)
	})

	t.Run("FetchEntity not found", func(t *testing.T) {
		entityID := "not-exist"
		inner := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
				return nil, nil
			},
		}
		adapter := NewEntityStoragerAdapter(inner)
		result, err := adapter.FetchEntity(ctx, &shared.EntityFilter{EntityID: &entityID})
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("FetchEntity error propagates", func(t *testing.T) {
		entityID := "ent-1"
		inner := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
				return nil, errors.New("db error")
			},
		}
		adapter := NewEntityStoragerAdapter(inner)
		_, err := adapter.FetchEntity(ctx, &shared.EntityFilter{EntityID: &entityID})
		require.Error(t, err)
	})
}

func TestQuotaBalanceStoragerAdapter(t *testing.T) {
	ctx := context.Background()

	t.Run("FetchQuotaBalance", func(t *testing.T) {
		inner := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				assert.Equal(t, int64(7), *filter.QuotaPlanID)
				return &QuotaBalanceParam{Used: lib.PFloat64(10), Remaining: lib.PFloat64(90)}, nil
			},
		}
		adapter := NewQuotaBalanceStoragerAdapter(inner)
		result, err := adapter.FetchQuotaBalance(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(10), *result.Used)
		assert.Equal(t, float64(90), *result.Remaining)
	})

	t.Run("FetchQuotaBalance not found", func(t *testing.T) {
		inner := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return nil, nil
			},
		}
		adapter := NewQuotaBalanceStoragerAdapter(inner)
		result, err := adapter.FetchQuotaBalance(ctx, 7)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("CreateQuotaBalance", func(t *testing.T) {
		inner := &fakeQuotaBalanceStorager{
			createFn: func(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
				assert.Equal(t, int64(7), *param.QuotaPlanID)
				assert.Equal(t, float64(0), *param.Used)
				assert.Equal(t, float64(100), *param.Remaining)
				assert.NotNil(t, param.LastResetAt)
				return 9, nil
			},
		}
		adapter := NewQuotaBalanceStoragerAdapter(inner)
		require.NoError(t, adapter.CreateQuotaBalance(ctx, 7, lib.PFloat64(100)))
	})

	t.Run("DeleteQuotaBalance", func(t *testing.T) {
		inner := &fakeQuotaBalanceStorager{
			deleteFn: func(ctx context.Context, filter *QuotaBalanceFilter) error {
				assert.Equal(t, int64(7), *filter.QuotaPlanID)
				return nil
			},
		}
		adapter := NewQuotaBalanceStoragerAdapter(inner)
		require.NoError(t, adapter.DeleteQuotaBalance(ctx, 7))
	})
}
