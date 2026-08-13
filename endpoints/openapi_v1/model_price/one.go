// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model_price

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// OneEndpoint fetches a single model price record by id.
var OneEndpoint = &xreq.Endpoint{
	Path:       "/model-prices/{id}",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(OneAction),
	Authorizer: iauth.FA(iauth.FeatureModelPrice, iauth.ActionRead),
}

// OneAction handles GET /model-prices/{id}.
func OneAction(req *http.Request) (interface{}, error) {
	id := idFromURI(req)
	if id == nil {
		return nil, xerror.WrapParamErrorWithMsg("id is required")
	}

	one, err := container.ModelPriceManager.FetchModelPrice(req.Context(), &imodel_price.ModelPriceFilter{ID: id})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, xerror.WrapRecordNotExist("ModelPrice")
	}
	return one, nil
}
