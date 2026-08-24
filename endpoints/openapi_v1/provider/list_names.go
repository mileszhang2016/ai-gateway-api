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
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// ListNamesEndpoint returns all provider names.
var ListNamesEndpoint = &xreq.Endpoint{
	Path:       "/providers/actions/get-provider-names",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListNamesAction),
	Authorizer: iauth.FA(iauth.FeatureProvider, iauth.ActionRead),
}

type listNamesResponse struct {
	Names []string `json:"names"`
}

// ListNamesAction handles GET /providers/actions/get-provider-names.
func ListNamesAction(req *http.Request) (interface{}, error) {
	names, err := container.ProviderManager.ListProviderNames(req.Context())
	if err != nil {
		return nil, err
	}

	return &listNamesResponse{Names: names}, nil
}
