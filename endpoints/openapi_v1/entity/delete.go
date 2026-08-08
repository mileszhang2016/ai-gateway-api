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
	"fmt"
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var EntityDeleteRoute = &xreq.Endpoint{
	Path:       "/entities/{id}",
	Method:     http.MethodDelete,
	Handler:    xreq.Convert(EntityDeleteAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionDelete),
}

type DeleteReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

func EntityDeleteAction(req *http.Request) (interface{}, error) {
	deleteReq := &DeleteReq{}
	if err := xreq.BindURI(req, deleteReq); err != nil {
		return nil, err
	}

	entity, err := container.EntityManager.FetchEntity(req.Context(), &quota.EntityFilter{
		EntityID: deleteReq.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	children, err := container.EntityManager.FetchEntityList(req.Context(), &quota.EntityFilter{ParentID: deleteReq.EntityID})
	if err != nil {
		return nil, err
	}
	if len(children) > 0 {
		return nil, xerror.WrapConflictErrorWithMsg("entity has children, cannot delete")
	}

	apiKeys, err := container.APIKeyManager.FetchAPIKeyList(req.Context(), &icluster_conf.APIKeyFilter{
		EntityID: deleteReq.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if len(apiKeys) > 0 {
		return nil, xerror.WrapParamErrorWithMsg(fmt.Sprintf("cannot delete entity with associated api-keys"))
	}

	return nil, container.EntityManager.DeleteEntity(req.Context(), &quota.EntityFilter{
		EntityID: deleteReq.EntityID,
	})
}
