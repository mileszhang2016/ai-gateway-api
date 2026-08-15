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

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// CreateEndpoint creates a single model price record.
var CreateEndpoint = &xreq.Endpoint{
	Path:       "/model-prices",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(CreateAction),
	Authorizer: iauth.FA(iauth.FeatureModelPrice, iauth.ActionCreate),
}

// CreateAction handles POST /model-prices.
func CreateAction(req *http.Request) (interface{}, error) {
	param := &imodel_price.ModelPrice{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	imodel_price.NormalizeCurrency(param, "RMB")

	id, err := container.ModelPriceManager.CreateModelPrice(req.Context(), param)
	if err != nil {
		return nil, err
	}

	return container.ModelPriceManager.FetchModelPrice(req.Context(), &imodel_price.ModelPriceFilter{ID: &id})
}
