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

package api_key

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
)

// OneRoute route
var OneRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(OneAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionRead),
}

type OneReq struct {
	ID *string `uri:"id" validate:"required,max=255"`
}

var _ xreq.Handler = OneAction

// OneAction action
func OneAction(req *http.Request) (interface{}, error) {
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	productName := defaultProductName

	one, err := container.APIKeyManager.FetchAPIKey(req.Context(), &api_key.APIKeyFilter{
		ID:          oneReq.ID,
		ProductName: &productName,
	})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, xerror.WrapRecordNotExist("API-Key")
	}

	response, err := newResponse(req.Context(), []*api_key.APIKeyParam{one})
	if err != nil {
		return nil, err
	}
	return response[0], nil
}

func newReq4One(req *http.Request) (*OneReq, error) {
	reqParam := &OneReq{}
	err := xreq.BindURI(req, reqParam)
	return reqParam, err
}
