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

const tEntityTypeTableName = "entity_types"

type TEntityType struct {
	ID          int64     `db:"id"`
	TypeName    string    `db:"type_name"`
	Description string    `db:"description"`
	Level       int       `db:"level"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// TEntityTypeOne Query One
// return nil, nil if record not existed
func TEntityTypeOne(dbCtx lib.DBContexter, where *TEntityTypeParam) (*TEntityType, error) {
	t := &TEntityType{}
	err := internal.QueryOne(dbCtx, tEntityTypeTableName, where, t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

// TEntityTypeList Query Multiple
func TEntityTypeList(dbCtx lib.DBContexter, where *TEntityTypeParam) ([]*TEntityType, error) {
	t := []*TEntityType{}
	err := internal.QueryList(dbCtx, tEntityTypeTableName, where, &t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

type TEntityTypeParam struct {
	ID          *int64     `db:"id"`
	TypeName    *string    `db:"type_name"`
	Description *string    `db:"description"`
	Level       *int       `db:"level"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`

	OrderBy *string `db:"_orderby"`
	Limit   []uint  `db:"_limit"`
}

// TEntityTypeCreate One/Multiple
func TEntityTypeCreate(dbCtx lib.DBContexter, data ...*TEntityTypeParam) (int64, error) {
	if len(data) == 1 {
		if data[0].CreatedAt == nil {
			data[0].CreatedAt = internal.PTimeNow()
		}
		return internal.Create(dbCtx, tEntityTypeTableName, data[0])
	}

	list := make([]interface{}, len(data))
	for i, one := range data {
		if one.CreatedAt == nil {
			one.CreatedAt = internal.PTimeNow()
		}
		list[i] = one
	}

	return internal.Create(dbCtx, tEntityTypeTableName, list...)
}

// TEntityTypeUpdate Update One
func TEntityTypeUpdate(dbCtx lib.DBContexter, val, where *TEntityTypeParam) (int64, error) {
	return internal.Update(dbCtx, tEntityTypeTableName, where, val)
}

// TEntityTypeDelete Delete One/Multiple
func TEntityTypeDelete(dbCtx lib.DBContexter, where *TEntityTypeParam) (int64, error) {
	return internal.Delete(dbCtx, tEntityTypeTableName, where)
}
