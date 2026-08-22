// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// CreateEndpoint creates a provider.
var CreateEndpoint = &xreq.Endpoint{
	Path:       "/providers",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(CreateAction),
	Authorizer: iauth.FA(iauth.FeatureProvider, iauth.ActionCreate),
}

// CreateAction handles POST /providers.
func CreateAction(req *http.Request) (interface{}, error) {
	param := &iprovider.ProviderParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	iprovider.FillDefaults(param)

	id, err := container.ProviderManager.CreateProvider(req.Context(), param)
	if err != nil {
		return nil, err
	}

	return container.ProviderManager.FetchProvider(req.Context(), &iprovider.ProviderFilter{ID: &id})
}
