// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package model_provider_type

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
)

var _ xreq.Handler = ListModelProviderTypesAction

var ListModelProviderTypesRoute = &xreq.Endpoint{
	Path:       "/model-provider-types",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListModelProviderTypesAction),
	Authorizer: iauth.FA(iauth.FeatureProductCluster, iauth.ActionRead),
}

var Endpoints = []*xreq.Endpoint{
	ListModelProviderTypesRoute,
}

type ModelProvider struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func ListModelProviderTypesAction(req *http.Request) (interface{}, error) {
	return listModelProviderTypesProcess(req.Context())
}

func listModelProviderTypesProcess(ctx context.Context) (interface{}, error) {
	data, err := os.ReadFile("conf/ai/models.json")
	if err != nil {
		return nil, xerror.WrapParamError(fmt.Errorf("failed to read config conf/ai/models.json: %w", err))
	}

	var providers []ModelProvider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, xerror.WrapParamError(fmt.Errorf("failed to parse config conf/ai/models.json: %w", err))
	}

	types := make([]string, 0, len(providers))
	for _, p := range providers {
		if p.ID != "" {
			types = append(types, p.ID)
		}
	}

	return types, nil
}
