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
	"github.com/infinity-ai-gateway/ai-gateway-api/model/entity"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var EntityOneRoute = &xreq.Endpoint{
	Path:       "/entities/{id}",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(EntityOneAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionRead),
}

type OneReq struct {
	EntityID *string `uri:"id" validate:"required"`
}

func EntityOneAction(req *http.Request) (interface{}, error) {
	oneReq := &OneReq{}
	if err := xreq.BindURI(req, oneReq); err != nil {
		return nil, err
	}

	one, err := container.EntityManager.FetchEntity(req.Context(), &entity.EntityFilter{
		EntityID: oneReq.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, xerror.WrapRecordNotExist("Entity")
	}

	return one, nil
}
