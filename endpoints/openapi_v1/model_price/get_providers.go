// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// GetProvidersEndpoint returns all distinct provider names from model_prices.
var GetProvidersEndpoint = &xreq.Endpoint{
	Path:       "/model-prices/actions/get-providers",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(GetProvidersAction),
	Authorizer: iauth.FA(iauth.FeatureModelPrice, iauth.ActionRead),
}

type getProvidersResponse struct {
	Providers []string `json:"providers"`
}

// GetProvidersAction handles GET /model-prices/actions/get-providers.
func GetProvidersAction(req *http.Request) (interface{}, error) {
	providers, err := container.ModelPriceManager.ListProviders(req.Context())
	if err != nil {
		return nil, err
	}

	return &getProvidersResponse{
		Providers: providers,
	}, nil
}
