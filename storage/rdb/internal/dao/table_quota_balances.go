// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http: //www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package dao

import (
	"time"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/internal/dao/internal"
)

const tQuotaBalanceTableName = "quota_balances"

type TQuotaBalance struct {
	ID          int64      `db:"id"`
	QuotaPlanID int64      `db:"quota_plan_id"`
	Used        int64      `db:"used"`
	Remaining   int64      `db:"remaining"`
	LastResetAt *time.Time `db:"last_reset_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// TQuotaBalanceOne Query One
// return nil, nil if record not existed
func TQuotaBalanceOne(dbCtx lib.DBContexter, where *TQuotaBalanceParam) (*TQuotaBalance, error) {
	t := &TQuotaBalance{}
	err := internal.QueryOne(dbCtx, tQuotaBalanceTableName, where, t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

// TQuotaBalanceList Query Multiple
func TQuotaBalanceList(dbCtx lib.DBContexter, where *TQuotaBalanceParam) ([]*TQuotaBalance, error) {
	t := []*TQuotaBalance{}
	err := internal.QueryList(dbCtx, tQuotaBalanceTableName, where, &t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

type TQuotaBalanceParam struct {
	ID          *int64     `db:"id"`
	QuotaPlanID *int64     `db:"quota_plan_id"`
	Used        *int64     `db:"used"`
	Remaining   *int64     `db:"remaining"`
	LastResetAt *time.Time `db:"last_reset_at"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`

	OrderBy *string `db:"_orderby"`
}

// TQuotaBalanceCreate One/Multiple
func TQuotaBalanceCreate(dbCtx lib.DBContexter, data ...*TQuotaBalanceParam) (int64, error) {
	if len(data) == 1 {
		if data[0].CreatedAt == nil {
			data[0].CreatedAt = internal.PTimeNow()
		}
		return internal.Create(dbCtx, tQuotaBalanceTableName, data[0])
	}

	list := make([]interface{}, len(data))
	for i, one := range data {
		if one.CreatedAt == nil {
			one.CreatedAt = internal.PTimeNow()
		}
		list[i] = one
	}

	return internal.Create(dbCtx, tQuotaBalanceTableName, list...)
}

// TQuotaBalanceUpdate Update One
func TQuotaBalanceUpdate(dbCtx lib.DBContexter, val, where *TQuotaBalanceParam) (int64, error) {
	return internal.Update(dbCtx, tQuotaBalanceTableName, where, val)
}

// TQuotaBalanceDelete Delete One/Multiple
func TQuotaBalanceDelete(dbCtx lib.DBContexter, where *TQuotaBalanceParam) (int64, error) {
	return internal.Delete(dbCtx, tQuotaBalanceTableName, where)
}
