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

type RateLimitPolicyStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewRateLimitPolicyStorager(dbCtxFactory lib.DBContextFactory) *RateLimitPolicyStorager {
	return &RateLimitPolicyStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ quota.RateLimitPolicyStorager = &RateLimitPolicyStorager{}

func (s *RateLimitPolicyStorager) CreateRateLimitPolicy(ctx context.Context, param *quota.RateLimitPolicyParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := rateLimitPolicyDataToParam(param)
	data.CreatedAt = lib.PTimeNow()

	return dao.TRateLimitPolicyCreate(dbCtx, data)
}

func (s *RateLimitPolicyStorager) FetchRateLimitPolicy(ctx context.Context, filter *quota.RateLimitPolicyFilter) (*quota.RateLimitPolicyParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := rateLimitPolicyFilterToParam(filter)
	one, err := dao.TRateLimitPolicyOne(dbCtx, where)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	return rateLimitPolicyParamToData(one), nil
}

func (s *RateLimitPolicyStorager) FetchRateLimitPolicyList(ctx context.Context, filter *quota.RateLimitPolicyFilter) ([]*quota.RateLimitPolicyParam, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	where := rateLimitPolicyFilterToParam(filter)
	list, err := dao.TRateLimitPolicyList(dbCtx, where)
	if err != nil {
		return nil, err
	}

	var rst []*quota.RateLimitPolicyParam
	for _, one := range list {
		rst = append(rst, rateLimitPolicyParamToData(one))
	}

	return rst, nil
}

func (s *RateLimitPolicyStorager) UpdateRateLimitPolicy(ctx context.Context, filter *quota.RateLimitPolicyFilter, param *quota.RateLimitPolicyParam) (int64, error) {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := rateLimitPolicyDataToParam(param)
	data.UpdatedAt = lib.PTimeNow()

	return dao.TRateLimitPolicyUpdate(dbCtx, data, rateLimitPolicyFilterToParam(filter))
}

func (s *RateLimitPolicyStorager) DeleteRateLimitPolicy(ctx context.Context, filter *quota.RateLimitPolicyFilter) error {
	dbCtx, err := s.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TRateLimitPolicyDelete(dbCtx, rateLimitPolicyFilterToParam(filter))
	return err
}

func rateLimitPolicyFilterToParam(filter *quota.RateLimitPolicyFilter) *dao.TRateLimitPolicyParam {
	if filter == nil {
		return nil
	}

	return &dao.TRateLimitPolicyParam{
		ID: filter.ID,
	}
}

func rateLimitPolicyDataToParam(param *quota.RateLimitPolicyParam) *dao.TRateLimitPolicyParam {
	data := &dao.TRateLimitPolicyParam{
		Enabled:        param.Enabled,
		MaxConcurrency: param.MaxConcurrency,
	}

	// 转换 TpmConfigs 为 JSON 字符串
	if len(param.TpmConfigs) > 0 {
		tpmConfigsJSON, _ := json.Marshal(param.TpmConfigs)
		data.TpmConfigs = lib.PString(string(tpmConfigsJSON))
	} else {
		data.TpmConfigs = lib.PString("[]")
	}

	// 转换 RpmConfigs 为 JSON 字符串
	if len(param.RpmConfigs) > 0 {
		rpmConfigsJSON, _ := json.Marshal(param.RpmConfigs)
		data.RpmConfigs = lib.PString(string(rpmConfigsJSON))
	} else {
		data.RpmConfigs = lib.PString("[]")
	}

	return data
}

func rateLimitPolicyParamToData(one *dao.TRateLimitPolicy) *quota.RateLimitPolicyParam {
	param := &quota.RateLimitPolicyParam{
		ID:             &one.ID,
		Enabled:        &one.Enabled,
		MaxConcurrency: &one.MaxConcurrency,
	}

	// 解析 TpmConfigs
	if one.TpmConfigs != "" {
		json.Unmarshal([]byte(one.TpmConfigs), &param.TpmConfigs)
	}

	// 解析 RpmConfigs
	if one.RpmConfigs != "" {
		json.Unmarshal([]byte(one.RpmConfigs), &param.RpmConfigs)
	}

	return param
}
