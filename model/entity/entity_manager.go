// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package entity

import (
	"context"
	"fmt"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
)

// EntityManager 定义 Entity 管理器
type EntityManager struct {
	txn                     itxn.TxnStorager
	storager                EntityStorager
	entityTypeStorager      EntityTypeStorager
	quotaPlanStorager       shared.QuotaPlanStorager
	rateLimitPolicyStorager shared.RateLimitPolicyStorager
	routeRulesStorager      shared.RouteRulesStorager
	quotaBalanceStorager    shared.QuotaBalanceStorager
	quotaCache              quotacache.QuotaCache
}

// NewEntityManager 创建 Entity 管理器
func NewEntityManager(txn itxn.TxnStorager, storager EntityStorager,
	entityTypeStorager EntityTypeStorager,
	quotaPlanStorager shared.QuotaPlanStorager,
	rateLimitPolicyStorager shared.RateLimitPolicyStorager,
	routeRulesStorager shared.RouteRulesStorager,
	quotaBalanceStorager shared.QuotaBalanceStorager,
	quotaCache quotacache.QuotaCache) *EntityManager {
	return &EntityManager{
		txn:                     txn,
		storager:                storager,
		entityTypeStorager:      entityTypeStorager,
		quotaPlanStorager:       quotaPlanStorager,
		rateLimitPolicyStorager: rateLimitPolicyStorager,
		routeRulesStorager:      routeRulesStorager,
		quotaBalanceStorager:    quotaBalanceStorager,
		quotaCache:              quotaCache,
	}
}

// CreateEntity 创建 Entity
func (m *EntityManager) CreateEntity(ctx context.Context, param *EntityParam) (int64, error) {
	if param.Type != nil && m.entityTypeStorager != nil {
		entityTypeInfo, err := m.entityTypeStorager.FetchEntityType(ctx, &EntityTypeFilter{TypeName: param.Type})
		if err != nil {
			return 0, err
		}
		if entityTypeInfo == nil {
			return 0, xerror.WrapParamError(fmt.Errorf("entity type not found: %s", *param.Type))
		}
	}

	if param.Name != nil && *param.Name != "" {
		existing, err := m.storager.FetchEntity(ctx, &EntityFilter{Name: param.Name})
		if err != nil {
			return 0, err
		}
		if existing != nil {
			return 0, xerror.WrapRecordExisted("Entity")
		}
	}

	if param.ParentID != nil && *param.ParentID != "" && param.Type != nil && m.entityTypeStorager != nil {
		if err := m.checkEntityLevel(ctx, *param.Type, *param.ParentID); err != nil {
			return 0, err
		}
	}

	var id int64
	err := m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		var err error

		if param.QuotaPlan != nil && m.quotaPlanStorager != nil {
			quotaPlanID, err := m.quotaPlanStorager.CreateQuotaPlan(ctx, param.QuotaPlan)
			if err != nil {
				return err
			}
			param.QuotaPlanID = &quotaPlanID
		}

		if param.RateLimitPolicy != nil && m.rateLimitPolicyStorager != nil {
			rateLimitPolicyID, err := m.rateLimitPolicyStorager.CreateRateLimitPolicy(ctx, param.RateLimitPolicy)
			if err != nil {
				return err
			}
			param.RateLimitPolicyID = &rateLimitPolicyID
		}

		if param.RouteRules != nil && m.routeRulesStorager != nil {
			routeRulesID, err := m.routeRulesStorager.CreateRouteRules(ctx, shared.RouteRulesTypeEntity, param.EntityID, param.RouteRules)
			if err != nil {
				return err
			}
			param.RouteRulesID = &routeRulesID
		}

		id, err = m.storager.CreateEntity(ctx, param)
		if err != nil {
			return err
		}

		if param.QuotaPlanID != nil && m.quotaBalanceStorager != nil {
			remaining := lib.PFloat64(0)
			if param.QuotaPlan != nil && param.QuotaPlan.Quota != nil {
				remaining = param.QuotaPlan.Quota
			}
			err = m.quotaBalanceStorager.CreateQuotaBalance(ctx, *param.QuotaPlanID, remaining)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// Sync Redis remaining quota after DB transaction commits (best-effort).
	if param.EntityID != nil && param.QuotaPlan != nil && m.quotaCache != nil {
		if param.QuotaPlan.Unlimited != nil && *param.QuotaPlan.Unlimited {
			defaultQuota := float64(100000000)
			if cacheErr := m.quotaCache.SetRemaining(ctx, *param.EntityID, &defaultQuota, param.QuotaPlan.Unit); cacheErr != nil {
				stateful.AccessLogger.Warn("failed to set quota cache for entity %s: %v", *param.EntityID, cacheErr)
			}
		} else if param.QuotaPlan.Quota != nil {
			if cacheErr := m.quotaCache.SetRemaining(ctx, *param.EntityID, param.QuotaPlan.Quota, param.QuotaPlan.Unit); cacheErr != nil {
				stateful.AccessLogger.Warn("failed to set quota cache for entity %s: %v", *param.EntityID, cacheErr)
			}
		}
	}

	return id, nil
}

// FetchEntity 查询单个 Entity
func (m *EntityManager) FetchEntity(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
	var one *EntityParam
	err := m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		var err error
		one, err = m.storager.FetchEntity(ctx, filter)
		if err != nil {
			return err
		}
		if one == nil {
			return nil
		}
		return m.populateAssociatedData(ctx, one)
	})

	return one, err
}

// FetchEntityList 查询 Entity 列表
func (m *EntityManager) FetchEntityList(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
	var list []*EntityParam
	err := m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		var err error
		list, err = m.storager.FetchEntityList(ctx, filter)
		if err != nil {
			return err
		}
		for _, one := range list {
			if err := m.populateAssociatedData(ctx, one); err != nil {
				return err
			}
		}
		return nil
	})

	return list, err
}

// UpdateEntity 更新 Entity
func (m *EntityManager) UpdateEntity(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
	if param.ParentID != nil && *param.ParentID != "" && param.Type != nil && m.entityTypeStorager != nil {
		if err := m.checkEntityLevel(ctx, *param.Type, *param.ParentID); err != nil {
			return 0, err
		}
	} else if param.ParentID != nil && *param.ParentID != "" && m.entityTypeStorager != nil {
		list, err := m.storager.FetchEntityList(ctx, filter)
		if err != nil {
			return 0, err
		}
		if len(list) == 0 {
			return 0, xerror.WrapRecordNotExist("Entity")
		}
		if list[0].Type != nil {
			if err := m.checkEntityLevel(ctx, *list[0].Type, *param.ParentID); err != nil {
				return 0, err
			}
		}
	}

	var affected int64
	err := m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		var err error

		list, err := m.storager.FetchEntityList(ctx, filter)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return xerror.WrapRecordNotExist("Entity")
		}

		one := list[0]

		if param.QuotaPlan != nil && m.quotaPlanStorager != nil {
			if one.QuotaPlanID != nil {
				_, err = m.quotaPlanStorager.UpdateQuotaPlan(ctx, *one.QuotaPlanID, param.QuotaPlan)
				if err != nil {
					return err
				}
			} else {
				quotaPlanID, err := m.quotaPlanStorager.CreateQuotaPlan(ctx, param.QuotaPlan)
				if err != nil {
					return err
				}
				param.QuotaPlanID = &quotaPlanID

				if m.quotaBalanceStorager != nil {
					err = m.quotaBalanceStorager.CreateQuotaBalance(ctx, quotaPlanID, param.QuotaPlan.Quota)
					if err != nil {
						return err
					}
				}
			}
		}

		if param.RateLimitPolicy != nil && m.rateLimitPolicyStorager != nil {
			if one.RateLimitPolicyID != nil {
				_, err = m.rateLimitPolicyStorager.UpdateRateLimitPolicy(ctx, *one.RateLimitPolicyID, param.RateLimitPolicy)
				if err != nil {
					return err
				}
			} else {
				rateLimitPolicyID, err := m.rateLimitPolicyStorager.CreateRateLimitPolicy(ctx, param.RateLimitPolicy)
				if err != nil {
					return err
				}
				param.RateLimitPolicyID = &rateLimitPolicyID
			}
		}

		if param.RouteRules != nil && m.routeRulesStorager != nil {
			if one.RouteRulesID != nil {
				_, err = m.routeRulesStorager.UpdateRouteRules(ctx, *one.RouteRulesID, param.RouteRules)
				if err != nil {
					return err
				}
			} else {
				routeRulesID, err := m.routeRulesStorager.CreateRouteRules(ctx, shared.RouteRulesTypeEntity, one.EntityID, param.RouteRules)
				if err != nil {
					return err
				}
				param.RouteRulesID = &routeRulesID
			}
		}

		affected, err = m.storager.UpdateEntity(ctx, &EntityFilter{EntityID: one.EntityID}, param)
		return err
	})
	return affected, err
}

// DeleteEntity 删除 Entity
func (m *EntityManager) DeleteEntity(ctx context.Context, filter *EntityFilter) error {
	return m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		list, err := m.storager.FetchEntityList(ctx, filter)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return xerror.WrapRecordNotExist("Entity")
		}

		one := list[0]

		children, err := m.storager.FetchEntityList(ctx, &EntityFilter{ParentID: one.EntityID})
		if err != nil {
			return err
		}
		if len(children) > 0 {
			return xerror.WrapConflictErrorWithMsg("cannot delete entity with children")
		}

		if one.QuotaPlanID != nil && m.quotaBalanceStorager != nil {
			if err := m.quotaBalanceStorager.DeleteQuotaBalance(ctx, *one.QuotaPlanID); err != nil {
				return err
			}
		}

		if one.QuotaPlanID != nil && m.quotaPlanStorager != nil {
			if err := m.quotaPlanStorager.DeleteQuotaPlan(ctx, *one.QuotaPlanID); err != nil {
				return err
			}
		}

		if one.RateLimitPolicyID != nil && m.rateLimitPolicyStorager != nil {
			if err := m.rateLimitPolicyStorager.DeleteRateLimitPolicy(ctx, *one.RateLimitPolicyID); err != nil {
				return err
			}
		}

		if one.RouteRulesID != nil && m.routeRulesStorager != nil {
			if err := m.routeRulesStorager.DeleteRouteRules(ctx, *one.RouteRulesID); err != nil {
				return err
			}
		}

		return m.storager.DeleteEntity(ctx, filter)
	})
}

func (m *EntityManager) populateAssociatedData(ctx context.Context, one *EntityParam) error {
	if one.QuotaPlan == nil {
		one.QuotaPlan = &shared.QuotaPlanParam{}
	}
	if one.RateLimitPolicy == nil {
		one.RateLimitPolicy = &shared.RateLimitPolicyParam{}
	}

	if one.QuotaPlanID != nil && m.quotaPlanStorager != nil {
		quotaPlan, err := m.quotaPlanStorager.FetchQuotaPlan(ctx, *one.QuotaPlanID)
		if err != nil {
			return err
		}
		one.QuotaPlan = quotaPlan

		if m.quotaBalanceStorager != nil {
			balance, err := m.quotaBalanceStorager.FetchQuotaBalance(ctx, *one.QuotaPlanID)
			if err != nil {
				return err
			}
			if balance != nil {
				if one.QuotaPlan.Balance == nil {
					one.QuotaPlan.Balance = &shared.BalanceSummary{}
				}
				one.QuotaPlan.Balance.Used = balance.Used
				one.QuotaPlan.Balance.Remaining = balance.Remaining
			}
		}
	}

	if one.RateLimitPolicyID != nil && m.rateLimitPolicyStorager != nil {
		rateLimitPolicy, err := m.rateLimitPolicyStorager.FetchRateLimitPolicy(ctx, *one.RateLimitPolicyID)
		if err != nil {
			return err
		}
		one.RateLimitPolicy = rateLimitPolicy
	}

	if one.RouteRules == nil {
		one.RouteRules = &shared.RouteRulesParam{}
	}
	if one.RouteRules.Enabled == nil {
		enabled := false
		one.RouteRules.Enabled = &enabled
	}
	if one.RouteRules.Rules == nil {
		one.RouteRules.Rules = []*shared.AiRouteRuleParam{}
	}

	if one.RouteRulesID != nil && m.routeRulesStorager != nil {
		routeRules, err := m.routeRulesStorager.FetchRouteRulesByID(ctx, *one.RouteRulesID)
		if err != nil {
			return err
		}
		if routeRules != nil {
			one.RouteRules = routeRules
		}
	}

	return nil
}

// FetchEntitySummary 获取 Entity 摘要
func (m *EntityManager) FetchEntitySummary(ctx context.Context, entityID string) (*shared.EntitySummary, error) {
	entity, err := m.FetchEntity(ctx, &EntityFilter{
		EntityID: &entityID,
	})
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, nil
	}

	return &shared.EntitySummary{
		ID:   entity.EntityID,
		Name: entity.Name,
		Type: entity.Type,
	}, nil
}

// checkEntityLevel 检查子 Entity 的 level 是否高于父 Entity 的 level
func (m *EntityManager) checkEntityLevel(ctx context.Context, entityType string, parentID string) error {
	if m.entityTypeStorager == nil {
		return nil
	}

	entityTypeInfo, err := m.entityTypeStorager.FetchEntityType(ctx, &EntityTypeFilter{TypeName: &entityType})
	if err != nil {
		return err
	}
	if entityTypeInfo == nil {
		return xerror.WrapParamErrorWithMsg("entity type not found: " + entityType)
	}
	if entityTypeInfo.Level == nil {
		return xerror.WrapParamErrorWithMsg("entity type level not set: " + entityType)
	}

	parentEntity, err := m.storager.FetchEntity(ctx, &EntityFilter{EntityID: &parentID})
	if err != nil {
		return err
	}
	if parentEntity == nil {
		return xerror.WrapParamErrorWithMsg("parent entity not found: " + parentID)
	}
	if parentEntity.Type == nil {
		return xerror.WrapParamErrorWithMsg("parent entity type not set: " + parentID)
	}

	parentEntityTypeInfo, err := m.entityTypeStorager.FetchEntityType(ctx, &EntityTypeFilter{TypeName: parentEntity.Type})
	if err != nil {
		return err
	}
	if parentEntityTypeInfo == nil {
		return xerror.WrapParamErrorWithMsg("parent entity type not found: " + *parentEntity.Type)
	}
	if parentEntityTypeInfo.Level == nil {
		return xerror.WrapParamErrorWithMsg("parent entity type level not set: " + *parentEntity.Type)
	}

	if *entityTypeInfo.Level <= *parentEntityTypeInfo.Level {
		return xerror.WrapParamErrorWithMsg(fmt.Sprintf("entity type level (%d) must be higher than parent entity type level (%d)",
			*entityTypeInfo.Level, *parentEntityTypeInfo.Level))
	}

	return nil
}

// entityStoragerAdapter 适配 EntityStorager 为 shared.EntityStorager
type entityStoragerAdapter struct {
	entityStorager EntityStorager
}

// NewEntityStoragerAdapter 创建 EntityStorager 适配器
func NewEntityStoragerAdapter(entityStorager EntityStorager) shared.EntityStorager {
	return &entityStoragerAdapter{
		entityStorager: entityStorager,
	}
}

func (a *entityStoragerAdapter) FetchEntity(ctx context.Context, filter *shared.EntityFilter) (*shared.EntitySummary, error) {
	entity, err := a.entityStorager.FetchEntity(ctx, &EntityFilter{
		EntityID: filter.EntityID,
		Name:     filter.Name,
		Type:     filter.Type,
		ParentID: filter.ParentID,
	})
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return &shared.EntitySummary{
		ID:   entity.EntityID,
		Name: entity.Name,
		Type: entity.Type,
	}, nil
}
