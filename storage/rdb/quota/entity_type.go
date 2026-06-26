// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package quota

import (
	"context"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/internal/dao"
)

type EntityTypeStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewEntityTypeStorager(dbCtxFactory lib.DBContextFactory) *EntityTypeStorager {
	return &EntityTypeStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ quota.EntityTypeStorager = &EntityTypeStorager{}

func (s *EntityTypeStorager) CreateEntityType(ctx context.Context, param *quota.EntityTypeParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := &dao.TEntityTypeParam{
		TypeName:    param.TypeName,
		Description: param.Description,
		Level:       param.Level,
		CreatedAt:   lib.PTimeNow(),
	}

	return dao.TEntityTypeCreate(dbCtx, data)
}

func (s *EntityTypeStorager) FetchEntityType(ctx context.Context, filter *quota.EntityTypeFilter) (*quota.EntityTypeParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := entityTypeFilterToParam(filter)
	one, err := dao.TEntityTypeOne(dbCtx, where)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return entityTypeParamToData(one), nil
}

func (s *EntityTypeStorager) FetchEntityTypeList(ctx context.Context, filter *quota.EntityTypeFilter) ([]*quota.EntityTypeParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := entityTypeFilterToParam(filter)
	list, err := dao.TEntityTypeList(dbCtx, where)
	if err != nil {
		return nil, err
	}

	var rst []*quota.EntityTypeParam
	for _, one := range list {
		rst = append(rst, entityTypeParamToData(one))
	}

	return rst, nil
}

func (s *EntityTypeStorager) UpdateEntityType(ctx context.Context, filter *quota.EntityTypeFilter, param *quota.EntityTypeParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := &dao.TEntityTypeParam{
		Description: param.Description,
		Level:       param.Level,
		UpdatedAt:   lib.PTimeNow(),
	}

	return dao.TEntityTypeUpdate(dbCtx, data, entityTypeFilterToParam(filter))
}

func (s *EntityTypeStorager) DeleteEntityType(ctx context.Context, filter *quota.EntityTypeFilter) error {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TEntityTypeDelete(dbCtx, entityTypeFilterToParam(filter))
	return err
}

func entityTypeFilterToParam(filter *quota.EntityTypeFilter) *dao.TEntityTypeParam {
	if filter == nil {
		return nil
	}

	return &dao.TEntityTypeParam{
		ID:       filter.ID,
		TypeName: filter.TypeName,
		Level:    filter.Level,
	}
}

func entityTypeParamToData(one *dao.TEntityType) *quota.EntityTypeParam {
	return &quota.EntityTypeParam{
		ID:          &one.ID,
		TypeName:    &one.TypeName,
		Description: &one.Description,
		Level:       &one.Level,
	}
}
