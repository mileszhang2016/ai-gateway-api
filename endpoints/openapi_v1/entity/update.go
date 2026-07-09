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

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
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

	existing, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{
		EntityID: updateReq.EntityID,
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
		EntityID: updateReq.EntityID,
	}, param); err != nil {
		return nil, err
	}

	return container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{EntityID: updateReq.EntityID})
}
