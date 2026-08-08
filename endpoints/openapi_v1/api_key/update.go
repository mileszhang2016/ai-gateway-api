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

package api_key

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var _ xreq.Handler = APIKeyUpdateAction

var APIKeyUpdateRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(APIKeyUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionUpdate),
}

func APIKeyUpdateAction(req *http.Request) (interface{}, error) {
	// uri param
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	// body param
	param := &icluster_conf.APIKeyParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	param.ID = lib.PString(*oneReq.ID)

	return APIKeyUpdateProcess(req.Context(), param, defaultProduct())
}

func APIKeyUpdateProcess(ctx context.Context, param *icluster_conf.APIKeyParam, product *ibasic.Product) (*icluster_conf.APIKeyParam, error) {
	if err := checkUpdateAPIKey(param, product.Name); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	existing, err := container.APIKeyManager.FetchAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          param.ID,
		ProductName: &product.Name,
	})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, xerror.WrapRecordNotExist("API-Key")
	}

	// 检查 entity_id 是否存在（如果传入的话）
	if param.EntityID != nil && *param.EntityID != "" {
		entity, err := container.EntityManager.FetchEntity(ctx, &quota.EntityFilter{EntityID: param.EntityID})
		if err != nil {
			return nil, err
		}
		if entity == nil {
			return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("Entity not found: %s", *param.EntityID))
		}
	}

	// 从最早获取的 existing 中取旧配额值
	var oldQuota int64
	if existing.QuotaPlanID != nil {
		oldPlan, errPlan := container.QuotaPlanManager.FetchQuotaPlan(ctx, &quota.QuotaPlanFilter{ID: existing.QuotaPlanID})
		if errPlan == nil && oldPlan != nil && oldPlan.Quota != nil {
			oldQuota = *oldPlan.Quota
		}
	}

	err = container.APIKeyManager.UpdateAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          param.ID,
		ProductName: &product.Name,
	}, &icluster_conf.APIKeyParam{
		Enable:          param.Enable,
		Key:             param.Key,
		Description:     param.Description,
		UnlimitedQuota:  param.UnlimitedQuota,
		ExpiredTime:     param.ExpiredTime,
		Models:          param.Models,
		Subnet:          param.Subnet,
		EntityID:        param.EntityID,
		ProductName:     &product.Name,
		QuotaPlan:       param.QuotaPlan,
		RateLimitPolicy: param.RateLimitPolicy,
		RouteRules:      param.RouteRules,
	})

	if err != nil {
		return nil, err
	}

	// 获取更新后的 API-Key
	updated, err := container.APIKeyManager.FetchAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          param.ID,
		ProductName: &product.Name,
	})
	if err != nil {
		return nil, err
	}

	// 检查配额计划，如果非无限制则需要确保 quota_balance 和 Redis key 存在
	if param.QuotaPlan != nil && (param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		updated != nil && updated.QuotaPlanID != nil {
		// 检查 quota_balance 是否存在，不存在则创建
		balance, err := container.QuotaPlanManager.FetchQuotaBalance(ctx, *updated.QuotaPlanID)
		if err != nil {
			return nil, err
		}
		if balance == nil {
			if err := container.QuotaPlanManager.CreateQuotaBalance(ctx, *updated.QuotaPlanID, param.QuotaPlan.Quota); err != nil {
				return nil, err
			}
		}

		// 检查 Redis key 是否存在，不存在则创建，存在则更新差值（新请求参数 quota - existing 旧 quota）
		if updated.Key != nil && stateful.DefaultClientSet != nil && stateful.DefaultClientSet.RedisClient != nil {
			redisKey := stateful.AIUsedQuotaKey(*updated.Key)
			_, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
			if errGet != nil {
				if strings.Contains(errGet.Error(), "redigo: nil returned") {
					quotaVal := int64(0)
					if param.QuotaPlan.Quota != nil {
						quotaVal = *param.QuotaPlan.Quota
					}
					_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, quotaVal)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, errGet
				}
			} else if param.QuotaPlan.Quota != nil {
				delta := *param.QuotaPlan.Quota - oldQuota
				_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return updated, nil
}
