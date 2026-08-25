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
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// PricingTiersEndpoint updates the pricing tiers of a provider.
var PricingTiersEndpoint = &xreq.Endpoint{
	Path:       "/providers/{provider_name}/pricing-tiers",
	Method:     http.MethodPut,
	Handler:    xreq.Convert(PricingTiersAction),
	Authorizer: iauth.FA(iauth.FeatureProvider, iauth.ActionUpdate),
}

// PricingTiersAction handles PUT /providers/{provider_name}/pricing-tiers.
// It accepts both JSON (application/json) and YAML (text/yaml or multipart/form-data file) bodies.
func PricingTiersAction(req *http.Request) (interface{}, error) {
	name := mux.Vars(req)["provider_name"]
	if name == "" {
		return nil, xerror.WrapParamErrorWithMsg("provider_name is required")
	}

	param := &iprovider.PricingTiersParam{}
	contentType := strings.ToLower(req.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		if err := xreq.BindJSON(req, param); err != nil {
			return nil, err
		}
	case strings.HasPrefix(contentType, "text/yaml"), strings.HasPrefix(contentType, "application/x-yaml"):
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, xerror.WrapParamErrorWithMsg("failed to read body: %v", err)
		}
		if err := yaml.Unmarshal(data, param); err != nil {
			return nil, xerror.WrapParamErrorWithMsg("failed to parse yaml: %v", err)
		}
	case strings.HasPrefix(contentType, "multipart/form-data"):
		file, _, err := req.FormFile("file")
		if err != nil {
			return nil, xerror.WrapParamErrorWithMsg("file is required for multipart/form-data")
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, xerror.WrapParamErrorWithMsg("failed to read file: %v", err)
		}
		if err := yaml.Unmarshal(data, param); err != nil {
			return nil, xerror.WrapParamErrorWithMsg("failed to parse yaml: %v", err)
		}
	default:
		return nil, xerror.WrapParamErrorWithMsg("unsupported Content-Type: %s", req.Header.Get("Content-Type"))
	}

	if err := container.ProviderManager.UpdatePricingTiers(req.Context(), name, param); err != nil {
		return nil, err
	}

	return container.ProviderManager.FetchProvider(req.Context(), &iprovider.ProviderFilter{Name: &name})
}
