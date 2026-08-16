// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package global_route_rules

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/validate"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

var Endpoints = []*xreq.Endpoint{
	GlobalRouteRulesGetRoute,
	GlobalRouteRulesUpdateRoute,
}

var GlobalRouteRulesGetRoute = &xreq.Endpoint{
	Path:       "/global-route-rules",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(GlobalRouteRulesGetAction),
	Authorizer: iauth.FA(iauth.FeatureRoute, iauth.ActionRead),
}

var GlobalRouteRulesUpdateRoute = &xreq.Endpoint{
	Path:       "/global-route-rules",
	Method:     http.MethodPut,
	Handler:    xreq.Convert(GlobalRouteRulesUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureRoute, iauth.ActionUpdate),
}

func GlobalRouteRulesGetAction(req *http.Request) (interface{}, error) {
	rules, err := container.RouteRulesManager.GetGlobalRouteRules(req.Context())
	if err != nil {
		return nil, err
	}
	if rules == nil {
		return nil, nil
	}

	// Hide internal id from response
	rules.ID = nil
	return rules, nil
}

func GlobalRouteRulesUpdateAction(req *http.Request) (interface{}, error) {
	param := &shared.RouteRulesParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if err := validate.RouteRules(param); err != nil {
		return nil, err
	}

	if param.Enabled == nil {
		enabled := true
		param.Enabled = &enabled
	}

	if param.Rules == nil {
		param.Rules = []*shared.AiRouteRuleParam{}
	}

	updated, err := container.RouteRulesManager.SetGlobalRouteRules(req.Context(), param)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, xerror.WrapRecordNotExist("global route rules")
	}

	// Hide internal id from response
	updated.ID = nil
	return updated, nil
}
