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
	"encoding/json"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/internal/dao"
)

type EntityStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewEntityStorager(dbCtxFactory lib.DBContextFactory) *EntityStorager {
	return &EntityStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ quota.EntityStorager = &EntityStorager{}

func (s *EntityStorager) CreateEntity(ctx context.Context, param *quota.EntityParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := entityDataToParam(param)
	data.CreatedAt = lib.PTimeNow()

	return dao.TEntityCreate(dbCtx, data)
}

func (s *EntityStorager) FetchEntity(ctx context.Context, filter *quota.EntityFilter) (*quota.EntityParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := entityFilterToParam(filter)
	one, err := dao.TEntityOne(dbCtx, where)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return entityParamToData(one), nil
}

func (s *EntityStorager) FetchEntityList(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := entityFilterToParam(filter)
	list, err := dao.TEntityList(dbCtx, where)
	if err != nil {
		return nil, err
	}

	var rst []*quota.EntityParam
	for _, one := range list {
		rst = append(rst, entityParamToData(one))
	}

	return rst, nil
}

func (s *EntityStorager) UpdateEntity(ctx context.Context, filter *quota.EntityFilter, param *quota.EntityParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := entityDataToParam(param)
	data.UpdatedAt = lib.PTimeNow()

	return dao.TEntityUpdate(dbCtx, data, entityFilterToParam(filter))
}

func (s *EntityStorager) DeleteEntity(ctx context.Context, filter *quota.EntityFilter) error {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TEntityDelete(dbCtx, entityFilterToParam(filter))
	return err
}

func entityFilterToParam(filter *quota.EntityFilter) *dao.TEntityParam {
	if filter == nil {
		return nil
	}

	hasCondition := filter.EntityID != nil || filter.Name != nil ||
		filter.Type != nil || filter.ParentID != nil || filter.QuotaPlanID != nil

	if !hasCondition && (filter.Page == nil || filter.PageSize == nil) {
		return nil
	}

	param := &dao.TEntityParam{
		EntityID:    filter.EntityID,
		Name:        filter.Name,
		Type:        filter.Type,
		ParentID:    filter.ParentID,
		QuotaPlanID: filter.QuotaPlanID,
	}

	if filter.Page != nil && filter.PageSize != nil {
		offset := (*filter.Page - 1) * (*filter.PageSize)
		if offset < 0 {
			offset = 0
		}
		param.Limit = []uint{uint(offset), uint(*filter.PageSize)}
	}

	return param
}

func entityDataToParam(param *quota.EntityParam) *dao.TEntityParam {
	data := &dao.TEntityParam{
		EntityID:          param.EntityID,
		Name:              param.Name,
		Type:              param.Type,
		ParentID:          param.ParentID,
		QuotaPlanID:       param.QuotaPlanID,
		RateLimitPolicyID: param.RateLimitPolicyID,
	}

	// 转换 AllowModels 为 JSON 字符串
	if len(param.AllowModels) > 0 {
		allowModelsJSON, _ := json.Marshal(param.AllowModels)
		data.AllowModels = lib.PString(string(allowModelsJSON))
	} else {
		data.AllowModels = lib.PString("[]")
	}

	// 转换 BlockModels 为 JSON 字符串
	if len(param.BlockModels) > 0 {
		blockModelsJSON, _ := json.Marshal(param.BlockModels)
		data.BlockModels = lib.PString(string(blockModelsJSON))
	} else {
		data.BlockModels = lib.PString("[]")
	}

	return data
}

func entityParamToData(one *dao.TEntity) *quota.EntityParam {
	param := &quota.EntityParam{
		InnerID:           &one.ID,
		EntityID:          &one.EntityID,
		Name:              &one.Name,
		Type:              &one.Type,
		ParentID:          one.ParentID,
		QuotaPlanID:       one.QuotaPlanID,
		RateLimitPolicyID: one.RateLimitPolicyID,
	}

	createTime := one.CreatedAt.Unix()
	param.CreateTime = &createTime
	updateTime := one.UpdatedAt.Unix()
	param.UpdateTime = &updateTime

	// 解析 AllowModels
	if one.AllowModels != "" {
		json.Unmarshal([]byte(one.AllowModels), &param.AllowModels)
	}

	// 解析 BlockModels
	if one.BlockModels != "" {
		json.Unmarshal([]byte(one.BlockModels), &param.BlockModels)
	}

	return param
}
