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

package dao

import (
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/storage/rdb/internal/dao/internal"
)

const tEntityTableName = "entities"

type TEntity struct {
	ID               int64     `db:"id"`
	EntityID         string    `db:"entity_id"`
	Name             string    `db:"name"`
	Type             string    `db:"type"`
	ParentID         *string   `db:"parent_id"`
	AllowModels      string    `db:"allow_models"`
	BlockModels      string    `db:"block_models"`
	QuotaPlanID       *int64    `db:"quota_plan_id"`
	RateLimitPolicyID *int64    `db:"rate_limit_policy_id"`
	RouteRulesID      *int64    `db:"route_rules_id"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// TEntityOne Query One
// return nil, nil if record not existed
func TEntityOne(dbCtx lib.DBContexter, where *TEntityParam) (*TEntity, error) {
	t := &TEntity{}
	err := internal.QueryOne(dbCtx, tEntityTableName, where, t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

// TEntityList Query Multiple
func TEntityList(dbCtx lib.DBContexter, where *TEntityParam) ([]*TEntity, error) {
	t := []*TEntity{}
	err := internal.QueryList(dbCtx, tEntityTableName, where, &t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

type TEntityParam struct {
	ID               *int64     `db:"id"`
	EntityID         *string    `db:"entity_id"`
	Name             *string    `db:"name"`
	Type             *string    `db:"type"`
	ParentID         *string    `db:"parent_id"`
	AllowModels      *string    `db:"allow_models"`
	BlockModels      *string    `db:"block_models"`
	QuotaPlanID       *int64     `db:"quota_plan_id"`
	RateLimitPolicyID *int64     `db:"rate_limit_policy_id"`
	RouteRulesID      *int64     `db:"route_rules_id"`
	CreatedAt         *time.Time `db:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at"`

	OrderBy *string `db:"_orderby"`
	Limit   []uint  `db:"_limit"`
}

// TEntityCreate One/Multiple
func TEntityCreate(dbCtx lib.DBContexter, data ...*TEntityParam) (int64, error) {
	if len(data) == 1 {
		if data[0].CreatedAt == nil {
			data[0].CreatedAt = internal.PTimeNow()
		}
		return internal.Create(dbCtx, tEntityTableName, data[0])
	}

	list := make([]interface{}, len(data))
	for i, one := range data {
		if one.CreatedAt == nil {
			one.CreatedAt = internal.PTimeNow()
		}
		list[i] = one
	}

	return internal.Create(dbCtx, tEntityTableName, list...)
}

// TEntityUpdate Update One
func TEntityUpdate(dbCtx lib.DBContexter, val, where *TEntityParam) (int64, error) {
	return internal.Update(dbCtx, tEntityTableName, where, val)
}

// TEntityDelete Delete One/Multiple
func TEntityDelete(dbCtx lib.DBContexter, where *TEntityParam) (int64, error) {
	return internal.Delete(dbCtx, tEntityTableName, where)
}
