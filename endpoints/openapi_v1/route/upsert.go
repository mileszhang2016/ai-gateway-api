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
//limitations under the License. All rights reserved.

// Copyright (c) 2021 The BFE Authors.
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

package route

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
)

type ProductRouteRuleParam struct {
	DefaultRouteRule *DefaultRouteRule `json:"default_forward_rule"`
}

type DefaultRouteRule struct {
	Cmd         string       `json:"cmd"`
	Params      []string     `json:"params"`
	Description string       `json:"description"`
	RouteAction *RouteAction `json:"action,omitempty"`
}

type RouteAction struct {
	Forward           *ActionForward           `json:"forward,omitempty"`
	GoToAdvancedRules *ActionGoToAdvancedRules `json:"go_to_advanced_rules,omitempty"`
	Redirect          *ActionRedirect          `json:"redirect,omitempty"`
	Response          *ActionResponse          `json:"response,omitempty"`
}

type ActionForward struct {
	ClusterName string `json:"cluster_name"`
	URL         string `json:"url"`
}

type ActionResponse struct {
	StatusCode  string `json:"status_code"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

type ActionRedirect struct {
	URL string `json:"url"`
}

type ActionGoToAdvancedRules struct {
}

type ProductRouteRuleData struct {
	DefaultRouteRule *DefaultRouteRule `json:"default_route_rule"`
}

func newProductRouteRuleData(pfr *ProductRouteRuleParam) *ProductRouteRuleData {
	return &ProductRouteRuleData{
		DefaultRouteRule: pfr.DefaultRouteRule,
	}
}

// UpsertRoute route
// AUTO GEN BY ctrl, MODIFY AS U NEED
// deprecated, endpoint registration removed per optimization plan v1.2
// var UpsertEndpoint = &xreq.Endpoint{
// 	Path:       "/routes",
// 	Method:     http.MethodPatch,
// 	Handler:    xreq.Convert(UpsertAction),
// 	Authorizer: iauth.FAP(iauth.FeatureRoute, iauth.ActionUpdate),
// }

// AUTO GEN BY ctrl, MODIFY AS U NEED
func newRuleInfoFromReq(req *http.Request) (*ProductRouteRuleParam, error) {
	rule := &ProductRouteRuleParam{}
	err := xreq.BindJSON(req, rule)
	if err != nil {
		return nil, err
	}

	if rule.DefaultRouteRule == nil {
		return nil, xerror.WrapParamErrorWithMsg("default_forward_rule cant be nil")
	}

	return rule, err
}

func UpsertActionProcess(req *http.Request, rule *ProductRouteRuleParam) (*ProductRouteRuleData, error) {
	_, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}

	return newProductRouteRuleData(rule), nil
}

var _ xreq.Handler = UpsertAction

// UpsertAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func UpsertAction(req *http.Request) (interface{}, error) {
	rule, err := newRuleInfoFromReq(req)
	if err != nil {
		return nil, err
	}

	return UpsertActionProcess(req, rule)
}
