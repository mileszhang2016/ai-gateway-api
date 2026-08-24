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

var EntityUpdateRoute = &xreq.Endpoint{
	Path:       "/entities/{id}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(EntityUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionUpdate),
}

type UpdateReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

func EntityUpdateAction(req *http.Request) (interface{}, error) {
	updateReq := &UpdateReq{}
	if err := xreq.BindURI(req, updateReq); err != nil {
		return nil, err
	}

	existing, err := container.EntityManager.FetchEntity(req.Context(), &entity.EntityFilter{
		EntityID: updateReq.EntityID,
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
		EntityID: updateReq.EntityID,
	}, param); err != nil {
		return nil, err
	}

	// 获取更新后的 Entity
	updated, err := container.EntityManager.FetchEntity(req.Context(), &entity.EntityFilter{EntityID: updateReq.EntityID})
	if err != nil {
		return nil, err
	}

	// 当 quota_plan 发生变更时，按需调整配额余额（普通属性修改不重置使用量）。
	if param.QuotaPlan != nil && updated != nil && updated.QuotaPlanID != nil {
		if err := container.QuotaPlanManager.ApplyQuotaPlanChange(req.Context(), *updated.QuotaPlanID, existing.QuotaPlan, param.QuotaPlan); err != nil {
			return nil, err
		}
	}

	return updated, nil
}
