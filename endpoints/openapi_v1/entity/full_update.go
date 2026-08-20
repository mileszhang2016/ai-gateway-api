// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package entity

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
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

	existing, err := container.EntityManager.FetchEntity(req.Context(), &entity.EntityFilter{
		EntityID: fullUpdateReq.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	param := &entity.EntityParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if err := validateEntityParam(param, false); err != nil {
		return nil, err
	}

	if _, err := container.EntityManager.UpdateEntity(req.Context(), &entity.EntityFilter{
		EntityID: fullUpdateReq.EntityID,
	}, param); err != nil {
		return nil, err
	}

	// 获取更新后的 Entity
	updated, err := container.EntityManager.FetchEntity(req.Context(), &entity.EntityFilter{EntityID: fullUpdateReq.EntityID})
	if err != nil {
		return nil, err
	}

	// 当 quota_plan 发生变更且非无限制时，重置 quota_balance（Redis 同步由 Manager 在事务外完成）
	if param.QuotaPlan != nil && (param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
		updated != nil && updated.QuotaPlanID != nil {
		if err := container.QuotaPlanManager.ResetBalance(req.Context(), *updated.QuotaPlanID, param.QuotaPlan.Quota, false); err != nil {
			return nil, err
		}
	}

	return updated, nil
}
