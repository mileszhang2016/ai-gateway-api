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

	"github.com/gorilla/mux"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// UpdateEndpoint updates a provider.
var UpdateEndpoint = &xreq.Endpoint{
	Path:       "/providers/{provider_name}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(UpdateAction),
	Authorizer: iauth.FA(iauth.FeatureProvider, iauth.ActionUpdate),
}

// UpdateAction handles PATCH /providers/{provider_name}.
func UpdateAction(req *http.Request) (interface{}, error) {
	name := mux.Vars(req)["provider_name"]
	if name == "" {
		return nil, xerror.WrapParamErrorWithMsg("provider_name is required")
	}

	param := &iprovider.ProviderParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if err := container.ProviderManager.UpdateProvider(req.Context(), name, param); err != nil {
		return nil, err
	}

	return container.ProviderManager.FetchProvider(req.Context(), &iprovider.ProviderFilter{Name: &name})
}
