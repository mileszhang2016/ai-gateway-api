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

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var EntityTypeDeleteRoute = &xreq.Endpoint{
	Path:       "/entity-types/{type_name}",
	Method:     http.MethodDelete,
	Handler:    xreq.Convert(EntityTypeDeleteAction),
	Authorizer: iauth.FA(iauth.FeatureEntityType, iauth.ActionDelete),
}

type DeleteReq struct {
	TypeName *string `uri:"type_name" validate:"required"`
}

func EntityTypeDeleteAction(req *http.Request) (interface{}, error) {
	delReq := &DeleteReq{}
	if err := xreq.BindURI(req, delReq); err != nil {
		return nil, err
	}

	one, err := container.EntityTypeManager.FetchEntityType(req.Context(), &quota.EntityTypeFilter{
		TypeName: delReq.TypeName,
	})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, xerror.WrapRecordNotExist("Entity-Type")
	}

	entities, err := container.EntityManager.FetchEntityList(req.Context(), &quota.EntityFilter{
		Type: delReq.TypeName,
	})
	if err != nil {
		return nil, err
	}
	if len(entities) > 0 {
		return nil, xerror.WrapParamErrorWithMsg("cannot delete entity type with associated entities")
	}

	return nil, container.EntityTypeManager.DeleteEntityType(req.Context(), &quota.EntityTypeFilter{
		TypeName: delReq.TypeName,
	})
}
