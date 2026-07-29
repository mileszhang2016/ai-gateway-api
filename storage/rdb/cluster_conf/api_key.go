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

package cluster_conf

import (
	"context"
	"encoding/json"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/internal/dao"
)

type APIKeyStorager struct {
	dbCtxFactory lib.DBContextFactory
}

func NewAPIKeyStorager(dbCtxFactory lib.DBContextFactory) *APIKeyStorager {
	return &APIKeyStorager{
		dbCtxFactory: dbCtxFactory,
	}
}

var _ icluster_conf.APIKeyStorager = &APIKeyStorager{}

func (rpps *APIKeyStorager) CreateAPIKey(ctx context.Context,
	param *icluster_conf.APIKeyParam) (int64, error) {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := newAPIKeyDataToParam(param)

	models := []string{"*"}
	if len(param.Models) > 0 {
		models = param.Models
	}
	modelsValue, _ := json.Marshal(models)
	data.AllowedModels = lib.PString(string(modelsValue))

	data.CreatedAt = lib.PTimeNow()

	subnet := []string{"*"}
	if len(param.Subnet) > 0 {
		subnet = param.Subnet
	}
	subnetValue, _ := json.Marshal(subnet)
	data.Subnet = lib.PString(string(subnetValue))

	return dao.TAPIKeyCreate(dbCtx, data)
}

func newAPIKeyDataToParam(param *icluster_conf.APIKeyParam) *dao.TAPIKeyParam {
	return &dao.TAPIKeyParam{
		ID:                param.ID,
		Enable:            param.Enable,
		Key:               param.Key,
		Description:       param.Description,
		UnlimitedQuota:    param.UnlimitedQuota,
		ExpiredTime:       param.ExpiredTime,
		ProductName:       param.ProductName,
		EntityID:          param.EntityID,
		QuotaPlanID:       param.QuotaPlanID,
		RateLimitPolicyID: param.RateLimitPolicyID,
		RouteRulesID:      param.RouteRulesID,
		UpdatedAt:         lib.PTimeNow(),
	}
}

func newAPIKeyFilterToParam(filter *icluster_conf.APIKeyFilter) *dao.TAPIKeyParam {
	if filter == nil {
		return nil
	}

	param := &dao.TAPIKeyParam{
		ProductName:     filter.ProductName,
		ID:              filter.ID,
		Key:             filter.Key,
		InnerID:         filter.InnerID,
		QuotaPlanID:     filter.QuotaPlanID,
		RouteRulesID:    filter.RouteRulesID,
		Enable:          filter.Enabled,
		EntityID:        filter.EntityID,
		UnlimitedQuota:  filter.UnlimitedQuota,
	}

	if filter.Page != nil && filter.PageSize != nil {
		offset := (*filter.Page - 1) * *filter.PageSize
		param.Limit = []uint{uint(offset), uint(*filter.PageSize)}
	}

	return param
}

func (rpps *APIKeyStorager) FetchAPIKeyList(ctx context.Context,
	filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	list, err := dao.TAPIKeyList(dbCtx, newAPIKeyFilterToParam(filter))
	if err != nil {
		return nil, err
	}

	var rst []*icluster_conf.APIKeyParam
	for _, one := range list {
		rst = append(rst, apiKeyParamToData(one))
	}

	return rst, nil
}

func apiKeyParamToData(one *dao.TAPIKey) *icluster_conf.APIKeyParam {
	models := []string{"*"}
	if one.AllowedModels != "" {
		json.Unmarshal([]byte(one.AllowedModels), &models)
	}
	if len(models) == 0 {
		models = []string{"*"}
	}

	subnet := []string{"*"}
	if one.Subnet != "" {
		json.Unmarshal([]byte(one.Subnet), &subnet)
	}
	if len(subnet) == 0 {
		subnet = []string{"*"}
	}

	createTime := one.CreatedAt.Unix()
	updateTime := one.UpdatedAt.Unix()

	return &icluster_conf.APIKeyParam{
		InnerID:           &one.InnerID,
		ID:                &one.ID,
		Enable:            &one.Enable,
		Key:               &one.Key,
		Description:       &one.Description,
		UnlimitedQuota:    &one.UnlimitedQuota,
		ExpiredTime:       &one.ExpiredTime,
		Models:            models,
		Subnet:            subnet,
		ProductName:       &one.ProductName,
		KeyCreateAt:       &one.CreatedAt,
		CreateTime:        &createTime,
		UpdatedTime:       &updateTime,
		EntityID:          one.EntityID,
		QuotaPlanID:       one.QuotaPlanID,
		RateLimitPolicyID: one.RateLimitPolicyID,
		RouteRulesID:      one.RouteRulesID,
	}
}

func (rpps *APIKeyStorager) DeleteAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter) error {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TAPIKeyDelete(dbCtx, newAPIKeyFilterToParam(filter))

	return err
}

func (rpps *APIKeyStorager) UpdateAPIKey(ctx context.Context, filter *icluster_conf.APIKeyFilter, param *icluster_conf.APIKeyParam) (int64, error) {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	data := newAPIKeyDataToParam(param)

	models := []string{"*"}
	if len(param.Models) > 0 {
		models = param.Models
	}
	modelsValue, _ := json.Marshal(models)
	data.AllowedModels = lib.PString(string(modelsValue))

	subnet := []string{"*"}
	if len(param.Subnet) > 0 {
		subnet = param.Subnet
	}
	subnetValue, _ := json.Marshal(subnet)
	data.Subnet = lib.PString(string(subnetValue))

	return dao.TAPIKeyUpdate(dbCtx, data, newAPIKeyFilterToParam(filter))
}

func (rpps *APIKeyStorager) CreateAPIKeyToken(ctx context.Context,
	param *icluster_conf.APIKeyTokenParam) (int64, error) {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return 0, err
	}

	return dao.TAPIKeyTokenCreate(dbCtx, &dao.TAPIKeyTokenParam{
		Key:       param.Key,
		CreatedAt: lib.PTimeNow(),
		UpdatedAt: lib.PTimeNow(),
	})
}

func (rpps *APIKeyStorager) UpdateAPIKeyToken(ctx context.Context, filter *icluster_conf.APIKeyTokenFilter, param *icluster_conf.APIKeyTokenParam) error {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return err
	}

	_, err = dao.TAPIKeyTokenUpdate(dbCtx, &dao.TAPIKeyTokenParam{
		Key: param.Key,
	}, &dao.TAPIKeyTokenParam{ID: filter.ID})
	return err
}

func (rpps *APIKeyStorager) FetchAPIKeyTokenList(ctx context.Context,
	filter *icluster_conf.APIKeyTokenFilter) ([]*icluster_conf.APIKeyTokenParam, error) {
	dbCtx, err := rpps.dbCtxFactory(ctx)
	if err != nil {
		return nil, err
	}

	list, err := dao.TAPIKeyTokenList(dbCtx, &dao.TAPIKeyTokenParam{
		Key: filter.Key,
	})
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}

	return newAPIKeyTokensDataToParam(list), nil
}

func newAPIKeyTokensDataToParam(list []*dao.TAPIKeyToken) []*icluster_conf.APIKeyTokenParam {
	results := make([]*icluster_conf.APIKeyTokenParam, len(list))
	for i, one := range list {
		results[i] = &icluster_conf.APIKeyTokenParam{
			Key:       &one.Key,
			CreatedAt: &one.CreatedAt,
		}
	}

	return results
}
