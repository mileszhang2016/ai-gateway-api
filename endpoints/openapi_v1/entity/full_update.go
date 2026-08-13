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

package entity

import (
	"net/http"
	"strings"

	golibquota "github.com/bfenetworks/go-lib/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var EntityFullUpdateRoute = &xreq.Endpoint{
	Path:       "/entities/{id}",
	Method:     http.MethodPut,
	Handler:    xreq.Convert(EntityFullUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionUpdate),
}

type FullUpdateReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

func EntityFullUpdateAction(req *http.Request) (interface{}, error) {
	fullUpdateReq := &FullUpdateReq{}
	if err := xreq.BindURI(req, fullUpdateReq); err != nil {
		return nil, err
	}

	existing, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{
		EntityID: fullUpdateReq.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	param := &quota.EntityParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if err := validateEntityParam(param, false); err != nil {
		return nil, err
	}

	if _, err := container.EntityManager.UpdateEntity(req.Context(), &quota.EntityFilter{
		EntityID: fullUpdateReq.EntityID,
	}, param); err != nil {
		return nil, err
	}

	// 获取更新后的 Entity
	updated, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{EntityID: fullUpdateReq.EntityID})
	if err != nil {
		return nil, err
	}

	// 当 quota_plan 发生变更且非无限制时，重置 quota_balance 并同步 Redis
	if param.QuotaPlan != nil && (param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		updated != nil && updated.QuotaPlanID != nil {
		if err := container.QuotaPlanManager.ResetBalance(req.Context(), *updated.QuotaPlanID, param.QuotaPlan.Quota, false); err != nil {
			return nil, err
		}

		if updated.EntityID != nil {
			redisKey := stateful.AIUsedQuotaKey(*updated.EntityID)
			targetValue := golibquota.PtrToRedisValue(param.QuotaPlan.Quota, param.QuotaPlan.Unit)
			currentValue, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
			if errGet != nil {
				if strings.Contains(errGet.Error(), "redigo: nil returned") {
					currentValue = 0
				} else {
					return nil, errGet
				}
			}
			delta := targetValue - currentValue
			if _, err := stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta); err != nil {
				return nil, err
			}
		}
	}

	return updated, nil
}
