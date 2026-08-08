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

package quota

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/internal/dao"
)

type QuotaBalanceStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewQuotaBalanceStorager(dbCtxFactory lib.DBContextFactory) *QuotaBalanceStorager {
	return &QuotaBalanceStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ quota.QuotaBalanceStorager = &QuotaBalanceStorager{}

func (s *QuotaBalanceStorager) CreateQuotaBalance(ctx context.Context, param *quota.QuotaBalanceParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := quotaBalanceDataToParam(param)
	data.CreatedAt = lib.PTimeNow()

	return dao.TQuotaBalanceCreate(dbCtx, data)
}

func (s *QuotaBalanceStorager) FetchQuotaBalance(ctx context.Context, filter *quota.QuotaBalanceFilter) (*quota.QuotaBalanceParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := quotaBalanceFilterToParam(filter)
	one, err := dao.TQuotaBalanceOne(dbCtx, where)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return quotaBalanceParamToData(one), nil
}

func (s *QuotaBalanceStorager) FetchQuotaBalanceList(ctx context.Context, filter *quota.QuotaBalanceFilter) ([]*quota.QuotaBalanceParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := quotaBalanceFilterToParam(filter)
	list, err := dao.TQuotaBalanceList(dbCtx, where)
	if err != nil {
		return nil, err
	}

	var rst []*quota.QuotaBalanceParam
	for _, one := range list {
		rst = append(rst, quotaBalanceParamToData(one))
	}

	return rst, nil
}

func (s *QuotaBalanceStorager) UpdateQuotaBalance(ctx context.Context, filter *quota.QuotaBalanceFilter, param *quota.QuotaBalanceParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := quotaBalanceDataToParam(param)
	data.UpdatedAt = lib.PTimeNow()

	return dao.TQuotaBalanceUpdate(dbCtx, data, quotaBalanceFilterToParam(filter))
}

func (s *QuotaBalanceStorager) DeleteQuotaBalance(ctx context.Context, filter *quota.QuotaBalanceFilter) error {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TQuotaBalanceDelete(dbCtx, quotaBalanceFilterToParam(filter))
	return err
}

func quotaBalanceFilterToParam(filter *quota.QuotaBalanceFilter) *dao.TQuotaBalanceParam {
	if filter == nil {
		return nil
	}

	return &dao.TQuotaBalanceParam{
		ID:          filter.ID,
		QuotaPlanID: filter.QuotaPlanID,
	}
}

func quotaBalanceDataToParam(param *quota.QuotaBalanceParam) *dao.TQuotaBalanceParam {
	return &dao.TQuotaBalanceParam{
		QuotaPlanID: param.QuotaPlanID,
		Used:        param.Used,
		Remaining:   param.Remaining,
		LastResetAt: param.LastResetAt,
	}
}

func quotaBalanceParamToData(one *dao.TQuotaBalance) *quota.QuotaBalanceParam {
	return &quota.QuotaBalanceParam{
		ID:          &one.ID,
		QuotaPlanID: &one.QuotaPlanID,
		Used:        &one.Used,
		Remaining:   &one.Remaining,
		LastResetAt: one.LastResetAt,
	}
}
