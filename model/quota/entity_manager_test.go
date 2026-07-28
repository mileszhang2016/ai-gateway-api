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

package quota

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/shared"
)

func TestEntityManager_CreateEntity(t *testing.T) {
	ctx := context.Background()

	t.Run("success with all associated data", func(t *testing.T) {
		entityID := "ent-1"
		entityName := "entity-one"
		entityType := "tenant"

		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				return &EntityTypeParam{TypeName: lib.PString(entityType), Level: lib.PInt(1)}, nil
			},
		}
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return nil, nil // no duplicate
			},
			createFn: func(ctx context.Context, param *EntityParam) (int64, error) {
				return 100, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			createFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				return 200, nil
			},
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) {
				return 300, nil
			},
		}
		routeRulesStore := &fakeRouteRulesStorager{
			createFn: func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
				return 400, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			createFn: func(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
				return 500, nil
			},
		}

		m := NewEntityManager(&fakeTxn{}, entityStore, entityTypeStore, quotaPlanStore, rateLimitStore, routeRulesStore, balanceStore)

		id, err := m.CreateEntity(ctx, &EntityParam{
			EntityID: &entityID,
			Name:     &entityName,
			Type:     &entityType,
			QuotaPlan: &shared.QuotaPlanParam{
				Quota: lib.PInt64(1000),
			},
			RateLimitPolicy: &shared.RateLimitPolicyParam{},
			RouteRules:      &shared.RouteRulesParam{},
		})

		require.NoError(t, err)
		assert.Equal(t, int64(100), id)

		// quota plan created and linked
		require.Len(t, quotaPlanStore.created, 1)
		assert.Equal(t, int64(1000), *quotaPlanStore.created[0].Quota)

		// rate limit policy created and linked
		require.Len(t, rateLimitStore.created, 1)

		// route rules created with correct type and owner
		require.Len(t, routeRulesStore.created, 1)
		assert.Equal(t, shared.RouteRulesTypeEntity, routeRulesStore.created[0].ruleType)
		assert.Equal(t, entityID, *routeRulesStore.created[0].owner)

		// balance created with remaining = quota
		require.Len(t, balanceStore.created, 1)
		assert.Equal(t, int64(1000), *balanceStore.created[0].Remaining)
		assert.Equal(t, int64(200), *balanceStore.created[0].QuotaPlanID)
	})

	t.Run("entity type not found", func(t *testing.T) {
		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, &fakeEntityStorager{}, entityTypeStore,
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		entityType := "unknown"
		_, err := m.CreateEntity(ctx, &EntityParam{Type: &entityType})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity type not found")
	})

	t.Run("duplicate name", func(t *testing.T) {
		entityName := "entity-one"
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{Name: &entityName}, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.CreateEntity(ctx, &EntityParam{Name: &entityName})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Entity Record Existed")
	})

	t.Run("parent level too low", func(t *testing.T) {
		parentID := "parent-1"
		childType := "project"

		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				if filter.TypeName != nil && *filter.TypeName == childType {
					return &EntityTypeParam{TypeName: lib.PString(childType), Level: lib.PInt(2)}, nil
				}
				return &EntityTypeParam{TypeName: filter.TypeName, Level: lib.PInt(2)}, nil
			},
		}
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				if filter.EntityID != nil && *filter.EntityID == parentID {
					parentType := "tenant"
					return &EntityParam{EntityID: &parentID, Type: &parentType}, nil
				}
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, entityTypeStore,
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.CreateEntity(ctx, &EntityParam{
			Type:     &childType,
			ParentID: &parentID,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity type level (2) must be higher than parent entity type level (2)")
	})
}

func TestEntityManager_FetchEntity(t *testing.T) {
	ctx := context.Background()

	t.Run("populate associated data", func(t *testing.T) {
		entityID := "ent-1"
		quotaPlanID := int64(200)
		rateLimitID := int64(300)
		routeRulesID := int64(400)

		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{
					EntityID:          &entityID,
					QuotaPlanID:       &quotaPlanID,
					RateLimitPolicyID: &rateLimitID,
					RouteRulesID:      &routeRulesID,
				}, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			fetchFn: func(ctx context.Context, id int64) (*shared.QuotaPlanParam, error) {
				return &shared.QuotaPlanParam{Quota: lib.PInt64(1000)}, nil
			},
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			fetchFn: func(ctx context.Context, id int64) (*shared.RateLimitPolicyParam, error) {
				return &shared.RateLimitPolicyParam{Enabled: lib.PBool(true)}, nil
			},
		}
		routeRulesStore := &fakeRouteRulesStorager{
			fetchByIDFn: func(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
				return &shared.RouteRulesParam{Enabled: lib.PBool(true)}, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return &QuotaBalanceParam{Used: lib.PInt64(100), Remaining: lib.PInt64(900)}, nil
			},
		}

		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore, rateLimitStore, routeRulesStore, balanceStore)

		entity, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.NoError(t, err)
		require.NotNil(t, entity)
		require.NotNil(t, entity.QuotaPlan)
		assert.Equal(t, int64(1000), *entity.QuotaPlan.Quota)
		require.NotNil(t, entity.QuotaPlan.Balance)
		assert.Equal(t, int64(100), *entity.QuotaPlan.Balance.Used)
		assert.Equal(t, int64(900), *entity.QuotaPlan.Balance.Remaining)
		require.NotNil(t, entity.RateLimitPolicy)
		assert.True(t, *entity.RateLimitPolicy.Enabled)
		require.NotNil(t, entity.RouteRules)
		assert.True(t, *entity.RouteRules.Enabled)
	})

	t.Run("not found", func(t *testing.T) {
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		entityID := "not-exist"
		entity, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.NoError(t, err)
		assert.Nil(t, entity)
	})
}

func TestEntityManager_DeleteEntity(t *testing.T) {
	ctx := context.Background()

	t.Run("success deletes cascade", func(t *testing.T) {
		entityID := "ent-1"
		quotaPlanID := int64(200)
		rateLimitID := int64(300)
		routeRulesID := int64(400)

		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				if filter.EntityID != nil && *filter.EntityID == entityID {
					return []*EntityParam{{
						EntityID:          &entityID,
						QuotaPlanID:       &quotaPlanID,
						RateLimitPolicyID: &rateLimitID,
						RouteRulesID:      &routeRulesID,
					}}, nil
				}
				// children check
				return nil, nil
			},
			deleteFn: func(ctx context.Context, filter *EntityFilter) error {
				return nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{}
		routeRulesStore := &fakeRouteRulesStorager{}
		balanceStore := &fakeQuotaBalanceStorager{}

		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore, rateLimitStore, routeRulesStore, balanceStore)

		require.NoError(t, m.DeleteEntity(ctx, &EntityFilter{EntityID: &entityID}))
		assert.Len(t, balanceStore.deleted, 1)
		assert.Equal(t, quotaPlanID, *balanceStore.deleted[0].QuotaPlanID)
		assert.Len(t, quotaPlanStore.deleted, 1)
		assert.Equal(t, quotaPlanID, quotaPlanStore.deleted[0])
		assert.Len(t, rateLimitStore.deleted, 1)
		assert.Equal(t, rateLimitID, rateLimitStore.deleted[0])
		assert.Len(t, routeRulesStore.deleted, 1)
		assert.Equal(t, routeRulesID, routeRulesStore.deleted[0])
	})

	t.Run("cannot delete entity with children", func(t *testing.T) {
		entityID := "ent-1"
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				if filter.EntityID != nil && *filter.EntityID == entityID {
					return []*EntityParam{{EntityID: &entityID}}, nil
				}
				return []*EntityParam{{EntityID: lib.PString("child-1")}}, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		err := m.DeleteEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete entity with children")
	})

	t.Run("entity not found", func(t *testing.T) {
		entityID := "not-exist"
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		err := m.DeleteEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Entity Record Not Exist")
	})
}

func TestEntityManager_FetchEntitySummary(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"
	entityName := "entity-one"
	entityType := "tenant"

	entityStore := &fakeEntityStorager{
		fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
			return &EntityParam{
				EntityID: &entityID,
				Name:     &entityName,
				Type:     &entityType,
			}, nil
		},
	}
	m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
		&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

	summary, err := m.FetchEntitySummary(ctx, entityID)
	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, entityID, *summary.ID)
	assert.Equal(t, entityName, *summary.Name)
	assert.Equal(t, entityType, *summary.Type)
}

func TestEntityManager_populateAssociatedData_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("quota plan fetch error propagates", func(t *testing.T) {
		entityID := "ent-1"
		quotaPlanID := int64(200)
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{EntityID: &entityID, QuotaPlanID: &quotaPlanID}, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			fetchFn: func(ctx context.Context, id int64) (*shared.QuotaPlanParam, error) {
				return nil, errors.New("quota plan db error")
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore,
			&fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota plan db error")
	})
}
func TestEntityManager_UpdateEntity(t *testing.T) {
	ctx := context.Background()

	t.Run("success update with new quota plan", func(t *testing.T) {
		entityID := "ent-1"
		innerID := int64(100)

		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{
					InnerID:  &innerID,
					EntityID: &entityID,
				}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			createFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				return 200, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			createFn: func(ctx context.Context, param *QuotaBalanceParam) (int64, error) {
				return 300, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore,
			&fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, balanceStore)

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PInt64(1000)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		require.Len(t, quotaPlanStore.created, 1)
		assert.Equal(t, int64(1000), *quotaPlanStore.created[0].Quota)

		require.Len(t, balanceStore.created, 1)
		assert.Equal(t, int64(200), *balanceStore.created[0].QuotaPlanID)
		assert.Equal(t, int64(1000), *balanceStore.created[0].Remaining)

		require.Len(t, entityStore.updated, 1)
		assert.Equal(t, int64(200), *entityStore.updated[0].param.QuotaPlanID)
	})

	t.Run("success update existing quota plan", func(t *testing.T) {
		entityID := "ent-1"
		innerID := int64(100)
		quotaPlanID := int64(200)

		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{
					InnerID:     &innerID,
					EntityID:    &entityID,
					QuotaPlanID: &quotaPlanID,
				}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			updateFn: func(ctx context.Context, id int64, param *shared.QuotaPlanParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore,
			&fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PInt64(2000)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		require.Len(t, quotaPlanStore.updated, 1)
		assert.Equal(t, quotaPlanID, quotaPlanStore.updated[0].id)
		assert.Equal(t, int64(2000), *quotaPlanStore.updated[0].param.Quota)
	})

	t.Run("success update route rules", func(t *testing.T) {
		entityID := "ent-1"
		innerID := int64(100)
		routeRulesID := int64(400)

		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{
					InnerID:      &innerID,
					EntityID:     &entityID,
					RouteRulesID: &routeRulesID,
				}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		routeRulesStore := &fakeRouteRulesStorager{
			updateFn: func(ctx context.Context, id int64, param *shared.RouteRulesParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
			&fakeSharedRateLimitPolicyStorager{}, routeRulesStore, &fakeQuotaBalanceStorager{})

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			RouteRules: &shared.RouteRulesParam{Enabled: lib.PBool(true)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		require.Len(t, routeRulesStore.updated, 1)
		assert.Equal(t, routeRulesID, routeRulesStore.updated[0].id)
	})

	t.Run("entity not found", func(t *testing.T) {
		entityID := "not-exist"
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Entity Record Not Exist")
	})

	t.Run("parent level check with existing entity type", func(t *testing.T) {
		entityID := "ent-1"
		parentID := "parent-1"
		entityType := "tenant"
		innerID := int64(100)

		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				return &EntityTypeParam{TypeName: lib.PString(entityType), Level: lib.PInt(1)}, nil
			},
		}
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{
					InnerID:  &innerID,
					EntityID: &entityID,
					Type:     &entityType,
				}}, nil
			},
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				if filter.EntityID != nil && *filter.EntityID == parentID {
					return &EntityParam{EntityID: &parentID, Type: &entityType}, nil
				}
				return nil, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, entityTypeStore,
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{ParentID: &parentID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity type level (1) must be higher than parent entity type level (1)")
	})
}

func TestEntityManager_FetchEntityList(t *testing.T) {
	ctx := context.Background()
	entityID1 := "ent-1"
	entityID2 := "ent-2"
	quotaPlanID := int64(200)

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
			return []*EntityParam{
				{EntityID: &entityID1, QuotaPlanID: &quotaPlanID},
				{EntityID: &entityID2},
			}, nil
		},
	}
	quotaPlanStore := &fakeSharedQuotaPlanStorager{
		fetchFn: func(ctx context.Context, id int64) (*shared.QuotaPlanParam, error) {
			return &shared.QuotaPlanParam{Quota: lib.PInt64(100)}, nil
		},
	}
	balanceStore := &fakeQuotaBalanceStorager{
		fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
			return &QuotaBalanceParam{Used: lib.PInt64(10), Remaining: lib.PInt64(90)}, nil
		},
	}
	m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore,
		&fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, balanceStore)

	list, err := m.FetchEntityList(ctx, &EntityFilter{})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, entityID1, *list[0].EntityID)
	assert.Equal(t, int64(100), *list[0].QuotaPlan.Quota)
	assert.Equal(t, entityID2, *list[1].EntityID)
}
func TestEntityManager_UpdateEntity_RateLimitPolicy(t *testing.T) {
	ctx := context.Background()

	t.Run("create new rate limit policy", func(t *testing.T) {
		entityID := "ent-1"
		innerID := int64(100)
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{InnerID: &innerID, EntityID: &entityID}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) {
				return 300, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
			rateLimitStore, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			RateLimitPolicy: &shared.RateLimitPolicyParam{Enabled: lib.PBool(true)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		require.Len(t, rateLimitStore.created, 1)
		assert.Equal(t, int64(300), *entityStore.updated[0].param.RateLimitPolicyID)
	})

	t.Run("update existing rate limit policy", func(t *testing.T) {
		entityID := "ent-1"
		innerID := int64(100)
		rateLimitID := int64(300)
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{InnerID: &innerID, EntityID: &entityID, RateLimitPolicyID: &rateLimitID}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			updateFn: func(ctx context.Context, id int64, param *shared.RateLimitPolicyParam) (int64, error) {
				assert.Equal(t, rateLimitID, id)
				return 1, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
			rateLimitStore, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			RateLimitPolicy: &shared.RateLimitPolicyParam{Enabled: lib.PBool(false)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		require.Len(t, rateLimitStore.updated, 1)
	})
}

func TestEntityManager_UpdateEntity_RouteRulesCreate(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"
	innerID := int64(100)
	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
			return []*EntityParam{{InnerID: &innerID, EntityID: &entityID}}, nil
		},
		updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
			return 1, nil
		},
	}
	routeRulesStore := &fakeRouteRulesStorager{
		createFn: func(ctx context.Context, ruleType string, owner *string, param *shared.RouteRulesParam) (int64, error) {
			assert.Equal(t, shared.RouteRulesTypeEntity, ruleType)
			assert.Equal(t, entityID, *owner)
			return 400, nil
		},
	}
	m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
		&fakeSharedRateLimitPolicyStorager{}, routeRulesStore, &fakeQuotaBalanceStorager{})

	affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
		RouteRules: &shared.RouteRulesParam{Enabled: lib.PBool(true)},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	require.Len(t, routeRulesStore.created, 1)
	assert.Equal(t, int64(400), *entityStore.updated[0].param.RouteRulesID)
}

func TestEntityManager_UpdateEntity_ParentIDWithoutType(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"
	parentID := "parent-1"
	innerID := int64(100)
	entityType := "tenant"

	entityTypeStore := &fakeEntityTypeStorager{
		fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
			return &EntityTypeParam{TypeName: lib.PString(entityType), Level: lib.PInt(2)}, nil
		},
	}
	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
			return []*EntityParam{{
				InnerID:  &innerID,
				EntityID: &entityID,
				Type:     &entityType,
			}}, nil
		},
		fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
			if filter.EntityID != nil && *filter.EntityID == parentID {
				parentType := "tenant"
				return &EntityParam{EntityID: &parentID, Type: &parentType}, nil
			}
			return nil, nil
		},
		updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
			return 1, nil
		},
	}
	m := NewEntityManager(&fakeTxn{}, entityStore, entityTypeStore,
		&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

	_, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{ParentID: &parentID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "entity type level (2) must be higher than parent entity type level (2)")
}

func TestEntityManager_checkEntityLevel_NilLevel(t *testing.T) {
	ctx := context.Background()
	entityType := "tenant"
	parentID := "parent-1"

	t.Run("entity type level nil", func(t *testing.T) {
		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				return &EntityTypeParam{TypeName: lib.PString(entityType)}, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, &fakeEntityStorager{}, entityTypeStore,
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		err := m.checkEntityLevel(ctx, entityType, parentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity type level not set")
	})

	t.Run("parent entity type level nil", func(t *testing.T) {
		parentType := "parent-type"
		entityTypeStore := &fakeEntityTypeStorager{
			fetchFn: func(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
				if filter.TypeName != nil && *filter.TypeName == entityType {
					return &EntityTypeParam{TypeName: lib.PString(entityType), Level: lib.PInt(2)}, nil
				}
				if filter.TypeName != nil && *filter.TypeName == parentType {
					return &EntityTypeParam{TypeName: lib.PString(parentType)}, nil
				}
				return nil, nil
			},
		}
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{EntityID: &parentID, Type: &parentType}, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, entityTypeStore,
			&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		err := m.checkEntityLevel(ctx, entityType, parentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parent entity type level not set")
	})
}

func TestEntityManager_FetchEntitySummary_Error(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"
	entityStore := &fakeEntityStorager{
		fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
			return nil, errors.New("db error")
		},
	}
	m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
		&fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

	_, err := m.FetchEntitySummary(ctx, entityID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestEntityManager_populateAssociatedData_MoreBranches(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"

	t.Run("rate limit policy fetch error", func(t *testing.T) {
		rateLimitID := int64(300)
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{EntityID: &entityID, RateLimitPolicyID: &rateLimitID}, nil
			},
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			fetchFn: func(ctx context.Context, id int64) (*shared.RateLimitPolicyParam, error) {
				return nil, errors.New("rate limit db error")
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
			rateLimitStore, &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

		_, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit db error")
	})

	t.Run("route rules fetch error", func(t *testing.T) {
		routeRulesID := int64(400)
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{EntityID: &entityID, RouteRulesID: &routeRulesID}, nil
			},
		}
		routeRulesStore := &fakeRouteRulesStorager{
			fetchByIDFn: func(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
				return nil, errors.New("route rules db error")
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, &fakeSharedQuotaPlanStorager{},
			&fakeSharedRateLimitPolicyStorager{}, routeRulesStore, &fakeQuotaBalanceStorager{})

		_, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "route rules db error")
	})

	t.Run("balance not found", func(t *testing.T) {
		quotaPlanID := int64(200)
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
				return &EntityParam{EntityID: &entityID, QuotaPlanID: &quotaPlanID}, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			fetchFn: func(ctx context.Context, id int64) (*shared.QuotaPlanParam, error) {
				return &shared.QuotaPlanParam{Quota: lib.PInt64(100)}, nil
			},
		}
		balanceStore := &fakeQuotaBalanceStorager{
			fetchFn: func(ctx context.Context, filter *QuotaBalanceFilter) (*QuotaBalanceParam, error) {
				return nil, nil
			},
		}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore,
			&fakeSharedRateLimitPolicyStorager{}, &fakeRouteRulesStorager{}, balanceStore)

		entity, err := m.FetchEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.NoError(t, err)
		require.NotNil(t, entity.QuotaPlan)
		assert.Nil(t, entity.QuotaPlan.Balance)
	})
}
