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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
)

func TestQuotaPlanManager_ResetBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("plan not found", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return nil, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota_plan not found")
	})

	t.Run("unlimited plan cannot reset", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{Unlimited: lib.PBool(true)}, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot reset balance for unlimited quota")
	})

	t.Run("reset existing balance and update last_reset_at", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
				return 1, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{ID: lib.PInt64(10), QuotaPlanID: lib.PInt64(1), Used: lib.PFloat64(100), Remaining: lib.PFloat64(900)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, balanceStore, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, true)
		require.NoError(t, err)

		require.Len(t, balanceStore.updated, 1)
		assert.Equal(t, int64(10), *balanceStore.updated[0].filter.ID)
		assert.Equal(t, float64(0), *balanceStore.updated[0].param.Used)
		assert.Equal(t, float64(1000), *balanceStore.updated[0].param.Remaining)
		assert.NotNil(t, balanceStore.updated[0].param.LastResetAt)
	})

	t.Run("reset existing balance without updating last_reset_at", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{ID: lib.PInt64(10), QuotaPlanID: lib.PInt64(1)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, balanceStore, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, false)
		require.NoError(t, err)

		require.Len(t, balanceStore.updated, 1)
		assert.Nil(t, balanceStore.updated[0].param.LastResetAt)
	})

	t.Run("create balance when not exists", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
				return 20, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, balanceStore, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, false)
		require.NoError(t, err)

		require.Len(t, balanceStore.created, 1)
		assert.Equal(t, int64(1), *balanceStore.created[0].QuotaPlanID)
		assert.Equal(t, float64(0), *balanceStore.created[0].Used)
		assert.Equal(t, float64(1000), *balanceStore.created[0].Remaining)
		assert.NotNil(t, balanceStore.created[0].LastResetAt)
	})

	t.Run("reset with new quota updates plan and balance", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
				return 1, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{ID: lib.PInt64(10), QuotaPlanID: lib.PInt64(1)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, balanceStore, nil, nil, nil)

		newQuota := float64(2000)
		err := m.ResetBalance(ctx, 1, &newQuota, false)
		require.NoError(t, err)

		require.Len(t, planStore.updated, 1)
		assert.Equal(t, float64(2000), *planStore.updated[0].param.Quota)

		require.Len(t, balanceStore.updated, 1)
		assert.Equal(t, float64(2000), *balanceStore.updated[0].param.Remaining)
	})

	t.Run("propagate balance update error", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{ID: lib.PInt64(10)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaBalanceFilter, param *QuotaBalanceParam) (int64, error) {
				return 0, errors.New("balance update failed")
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, balanceStore, nil, nil, nil)

		err := m.ResetBalance(ctx, 1, nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "balance update failed")
	})
}

func TestQuotaPlanManager_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateQuotaPlan", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			createFn: func(ctx context.Context, param *QuotaPlanParam) (int64, error) {
				return 7, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		id, err := m.CreateQuotaPlan(ctx, &QuotaPlanParam{Quota: lib.PFloat64(100)})
		require.NoError(t, err)
		assert.Equal(t, int64(7), id)
	})

	t.Run("FetchQuotaPlan", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(7), Quota: lib.PFloat64(100)}, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		plan, err := m.FetchQuotaPlan(ctx, &QuotaPlanFilter{ID: lib.PInt64(7)})
		require.NoError(t, err)
		require.NotNil(t, plan)
		assert.Equal(t, float64(100), *plan.Quota)
	})

	t.Run("FetchQuotaPlanList", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			listFn: func(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
				return []*QuotaPlanParam{{ID: lib.PInt64(7)}}, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		list, err := m.FetchQuotaPlanList(ctx, &QuotaPlanFilter{})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("DeleteQuotaPlan", func(t *testing.T) {
		planStore := &fakeQuotaPlanStorager{
			deleteFn: func(ctx context.Context, filter *QuotaPlanFilter) error {
				return nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, &fakeQuotaBalanceStorager{}, nil, nil, nil)

		require.NoError(t, m.DeleteQuotaPlan(ctx, &QuotaPlanFilter{ID: lib.PInt64(7)}))
		assert.Len(t, planStore.deleted, 1)
	})
}
func TestQuotaPlanManager_UpdateQuotaPlan(t *testing.T) {
	ctx := context.Background()
	store := &fakeQuotaPlanStorager{
		updateFn: func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
			assert.Equal(t, int64(7), *filter.ID)
			assert.Equal(t, float64(500), *param.Quota)
			return 1, nil
		},
	}
	m := NewQuotaPlanManager(&fakeTxn{}, store, &fakeQuotaBalanceStorager{}, nil, nil, nil)

	affected, err := m.UpdateQuotaPlan(ctx, &QuotaPlanFilter{ID: lib.PInt64(7)}, &QuotaPlanParam{Quota: lib.PFloat64(500)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
}

func TestQuotaPlanManager_FetchQuotaBalance(t *testing.T) {
	ctx := context.Background()
	store := &fakeQuotaBalanceStorager{
		fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
			assert.Equal(t, int64(7), *filter.QuotaPlanID)
			return &QuotaBalanceParam{Used: lib.PFloat64(10), Remaining: lib.PFloat64(90)}, nil
		},
	}
	m := NewQuotaPlanManager(&fakeTxn{}, &fakeQuotaPlanStorager{}, store, nil, nil, nil)

	balance, err := m.FetchQuotaBalance(ctx, 7)
	require.NoError(t, err)
	require.NotNil(t, balance)
	assert.Equal(t, float64(10), *balance.Used)
	assert.Equal(t, float64(90), *balance.Remaining)
}

func TestQuotaPlanManager_CreateQuotaBalance(t *testing.T) {
	ctx := context.Background()
	store := &fakeQuotaBalanceStorager{
		createFn: func(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
			assert.Equal(t, int64(7), *param.QuotaPlanID)
			assert.Equal(t, float64(0), *param.Used)
			assert.Equal(t, float64(100), *param.Remaining)
			assert.NotNil(t, param.LastResetAt)
			return 9, nil
		},
	}
	m := NewQuotaPlanManager(&fakeTxn{}, &fakeQuotaPlanStorager{}, store, nil, nil, nil)

	require.NoError(t, m.CreateQuotaBalance(ctx, 7, lib.PFloat64(100)))
}
