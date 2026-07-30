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

package entity_type

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib/validate"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var EntityTypeUpdateRoute = &xreq.Endpoint{
	Path:       "/entity-types/{type_name}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(EntityTypeUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureEntityType, iauth.ActionUpdate),
}

type UpdateReq struct {
	TypeName *string `uri:"type_name" validate:"required"`
}

func EntityTypeUpdateAction(req *http.Request) (interface{}, error) {
	updateReq := &UpdateReq{}
	if err := xreq.BindURI(req, updateReq); err != nil {
		return nil, err
	}
	if err := validate.EntityTypeName(*updateReq.TypeName); err != nil {
		return nil, err
	}

	existing, err := container.EntityTypeManager.FetchEntityType(req.Context(), &quota.EntityTypeFilter{
		TypeName: updateReq.TypeName,
	})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, xerror.WrapRecordNotExist("Entity-Type")
	}

	param := &quota.EntityTypeParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}
	if param.TypeName != nil {
		if err := validate.EntityTypeName(*param.TypeName); err != nil {
			return nil, err
		}
	}
	if param.Description != nil {
		if err := validate.Description(*param.Description, validate.MaxDescriptionLength, "description"); err != nil {
			return nil, err
		}
	}

	_, err = container.EntityTypeManager.UpdateEntityType(req.Context(), &quota.EntityTypeFilter{
		TypeName: updateReq.TypeName,
	}, param)
	if err != nil {
		return nil, err
	}

	return container.EntityTypeManager.FetchEntityType(req.Context(), &quota.EntityTypeFilter{
		TypeName: updateReq.TypeName,
	})
}
