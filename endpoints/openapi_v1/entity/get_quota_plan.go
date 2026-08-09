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

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var EntityGetQuotaPlanRoute = &xreq.Endpoint{
	Path:       "/entities/{id}/quota-plan",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(EntityGetQuotaPlanAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionRead),
}

type EntityGetQuotaPlanResponse struct {
	Unlimited             *bool     `json:"unlimited"`
	PassWhenNoEnoughQuota *bool     `json:"pass_when_no_enough_quota"`
	Quota                 *int64    `json:"quota"`
	Unit                  *string   `json:"unit"`
	ResetPeriod           *string   `json:"reset_period"`
	Balance               *Balance  `json:"balance"`
}

type Balance struct {
	Used      int64 `json:"used"`
	Remaining int64 `json:"remaining"`
}

type GetQuotaPlanReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

func EntityGetQuotaPlanAction(req *http.Request) (interface{}, error) {
	getReq := &GetQuotaPlanReq{}
	if err := xreq.BindURI(req, getReq); err != nil {
		return nil, err
	}

	entity, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{
		EntityID: getReq.EntityID,
	})
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	if entity.QuotaPlanID == nil || entity.QuotaPlan == nil {
		return nil, xerror.WrapParamErrorWithMsg("Entity has no quota plan")
	}

	var balance *Balance
	if entity.QuotaPlanID != nil {
		balanceData, err := container.QuotaPlanManager.FetchQuotaBalance(req.Context(), *entity.QuotaPlanID)
		if err != nil {
			return nil, err
		}
		if balanceData != nil {
			used := int64(0)
			if balanceData.Used != nil {
				used = *balanceData.Used
			}
			remaining := int64(0)
			if balanceData.Remaining != nil {
				remaining = *balanceData.Remaining
			}
			balance = &Balance{
				Used:      used,
				Remaining: remaining,
			}
		} else {
			remaining := int64(0)
			if entity.QuotaPlan.Quota != nil {
				remaining = *entity.QuotaPlan.Quota
			}
			balance = &Balance{
				Used:      0,
				Remaining: remaining,
			}
		}
	}

	return &EntityGetQuotaPlanResponse{
		Unlimited:             entity.QuotaPlan.Unlimited,
		PassWhenNoEnoughQuota: entity.QuotaPlan.PassWhenNoEnoughQuota,
		Quota:                 entity.QuotaPlan.Quota,
		Unit:                  entity.QuotaPlan.Unit,
		ResetPeriod:           entity.QuotaPlan.ResetPeriod,
		Balance:               balance,
	}, nil
}