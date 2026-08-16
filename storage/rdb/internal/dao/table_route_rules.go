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

const tRouteRulesTableName = "route_rules"

type TRouteRules struct {
	ID        int64     `db:"id"`
	Type      string    `db:"type"`
	Owner     string    `db:"owner"`
	Enabled   bool      `db:"enabled"`
	Rules     string    `db:"rules"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TRouteRulesOne Query One
// return nil, nil if record not existed
func TRouteRulesOne(dbCtx lib.DBContexter, where *TRouteRulesParam) (*TRouteRules, error) {
	t := &TRouteRules{}
	err := internal.QueryOne(dbCtx, tRouteRulesTableName, where, t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

// TRouteRulesList Query Multiple
func TRouteRulesList(dbCtx lib.DBContexter, where *TRouteRulesParam) ([]*TRouteRules, error) {
	t := []*TRouteRules{}
	err := internal.QueryList(dbCtx, tRouteRulesTableName, where, &t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

type TRouteRulesParam struct {
	ID        *int64     `db:"id"`
	Type      *string    `db:"type"`
	Owner     *string    `db:"owner"`
	Enabled   *bool      `db:"enabled"`
	Rules     *string    `db:"rules"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`

	OrderBy *string `db:"_orderby"`
	Limit   []uint  `db:"_limit"`
}

// TRouteRulesCreate One/Multiple
func TRouteRulesCreate(dbCtx lib.DBContexter, data ...*TRouteRulesParam) (int64, error) {
	if len(data) == 1 {
		if data[0].CreatedAt == nil {
			data[0].CreatedAt = internal.PTimeNow()
		}
		return internal.Create(dbCtx, tRouteRulesTableName, data[0])
	}

	list := make([]interface{}, len(data))
	for i, one := range data {
		if one.CreatedAt == nil {
			one.CreatedAt = internal.PTimeNow()
		}
		list[i] = one
	}

	return internal.Create(dbCtx, tRouteRulesTableName, list...)
}

// TRouteRulesUpdate Update One
func TRouteRulesUpdate(dbCtx lib.DBContexter, val, where *TRouteRulesParam) (int64, error) {
	return internal.Update(dbCtx, tRouteRulesTableName, where, val)
}

// TRouteRulesDelete Delete One/Multiple
func TRouteRulesDelete(dbCtx lib.DBContexter, where *TRouteRulesParam) (int64, error) {
	return internal.Delete(dbCtx, tRouteRulesTableName, where)
}
