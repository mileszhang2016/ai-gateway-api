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

package entity

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
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

	if param.RateLimitPolicy != nil && param.RateLimitPolicy.Rules != nil {
		for _, tpm := range param.RateLimitPolicy.Rules.TpmConfigs {
			if tpm.WindowMinutes < 1 || tpm.WindowMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm window_minutes must be between 1 and 360, got %d", tpm.WindowMinutes))
			}

			if tpm.StepMinutes < 1 || tpm.StepMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes must be between 1 and 360, got %d", tpm.StepMinutes))
			}

			if tpm.StepMinutes > tpm.WindowMinutes {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("tpm step_minutes (%d) must be <= window_minutes (%d)", tpm.StepMinutes, tpm.WindowMinutes))
			}
		}

		for _, rpm := range param.RateLimitPolicy.Rules.RpmConfigs {
			if rpm.WindowMinutes < 1 || rpm.WindowMinutes > 360 {
				return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("rpm window_minutes must be between 1 and 360, got %d", rpm.WindowMinutes))
			}
		}
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

	// 检查配额计划，如果非无限制则需要确保 quota_balance 和 Redis key 存在
	if param.QuotaPlan != nil && (param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		updated != nil && updated.QuotaPlanID != nil {
		balance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *updated.QuotaPlanID)
		if err != nil {
			return nil, err
		}
		if balance == nil {
			if err := container.QuotaPlanManager.CreateQuotaBalance(req.Context(), *updated.QuotaPlanID, param.QuotaPlan.Quota); err != nil {
				return nil, err
			}
		}

		if updated.EntityID != nil {
			redisKey := stateful.AIUsedQuotaKey(*updated.EntityID)
			currentVal, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
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
			} else if currentVal == 0 {
				quotaVal := int64(0)
				if param.QuotaPlan.Quota != nil {
					quotaVal = *param.QuotaPlan.Quota
				}
				_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, quotaVal)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return updated, nil
}
