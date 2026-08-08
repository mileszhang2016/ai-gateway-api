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

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var EntityResetQuotaRoute = &xreq.Endpoint{
	Path:       "/entities/{id}/quota-plan/reset",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(EntityResetQuotaAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionUpdate),
}

type ResetQuotaReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

type ResetQuotaBody struct {
	Quota  *int64  `json:"quota,omitempty"`
	Reason *string `json:"reason,omitempty"`
}

type ResetQuotaResponse struct {
	ID            *string           `json:"id"`
	PreviousQuota *int64            `json:"previous_quota"`
	NewQuota      *int64            `json:"new_quota"`
	Balance       *ResetBalanceInfo `json:"balance"`
}

type ResetBalanceInfo struct {
	PreviousRemaining int64 `json:"previous_remaining"`
	NewRemaining      int64 `json:"new_remaining"`
	Used              int64 `json:"used"`
}

func EntityResetQuotaAction(req *http.Request) (interface{}, error) {
	resetReq := &ResetQuotaReq{}
	if err := xreq.BindURI(req, resetReq); err != nil {
		return nil, err
	}

	bodyReq := &ResetQuotaBody{}
	if err := xreq.BindJSON(req, bodyReq); err != nil {
		return nil, err
	}

	entity, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{
		EntityID: resetReq.EntityID,
	})
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	if entity.QuotaPlanID == nil {
		return nil, xerror.WrapParamErrorWithMsg("Entity has no quota plan")
	}

	var previousQuota *int64
	var previousRemaining int64

	plan, err := container.QuotaPlanManager.FetchQuotaPlan(req.Context(), &quota.QuotaPlanFilter{
		ID: entity.QuotaPlanID,
	})
	if err != nil {
		return nil, err
	}
	if plan != nil {
		previousQuota = plan.Quota
	}

	balance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *entity.QuotaPlanID)
	if err != nil {
		return nil, err
	}
	if balance != nil && balance.Remaining != nil {
		previousRemaining = *balance.Remaining
	}

	err = container.QuotaPlanManager.ResetBalance(req.Context(), *entity.QuotaPlanID, bodyReq.Quota, false)
	if err != nil {
		return nil, err
	}

	// 重置 Redis 中的值为 quota 总量
	if entity.EntityID != nil {
		redisKey := stateful.AIUsedQuotaKey(*entity.EntityID)

		var resetQuota int64
		if bodyReq.Quota != nil {
			resetQuota = *bodyReq.Quota
		} else if plan != nil && plan.Quota != nil {
			resetQuota = *plan.Quota
		} else {
			return nil, xerror.WrapParamErrorWithMsg("quota is required")
		}

		currentValue, err := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if err != nil {
			if !strings.Contains(err.Error(), "redigo: nil returned") {
				return nil, err
			}
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, resetQuota)
		} else {
			delta := resetQuota - currentValue
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
		}
		if err != nil {
			return nil, err
		}
	}

	newPlan, err := container.QuotaPlanManager.FetchQuotaPlan(req.Context(), &quota.QuotaPlanFilter{
		ID: entity.QuotaPlanID,
	})
	if err != nil {
		return nil, err
	}

	newBalance, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *entity.QuotaPlanID)
	if err != nil {
		return nil, err
	}

	newRemaining := int64(0)
	if newBalance != nil && newBalance.Remaining != nil {
		newRemaining = *newBalance.Remaining
	}

	newQuota := previousQuota
	if bodyReq.Quota != nil {
		newQuota = bodyReq.Quota
	} else if newPlan != nil {
		newQuota = newPlan.Quota
	}

	return &ResetQuotaResponse{
		ID:            resetReq.EntityID,
		PreviousQuota: previousQuota,
		NewQuota:      newQuota,
		Balance: &ResetBalanceInfo{
			PreviousRemaining: previousRemaining,
			NewRemaining:      newRemaining,
			Used:              0,
		},
	}, nil
}
