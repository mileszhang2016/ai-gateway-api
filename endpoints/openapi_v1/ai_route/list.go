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

package ai_route

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/validate"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iroute_conf"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// AdvanceRouteRule defines the structure for advance routing rules
type AdvanceRouteRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Expression  string `json:"expression" validate:"required,min=1"`
	ClusterName string `json:"cluster_name" validate:"required,min=1"`
}

// BasicRouteRule defines the structure for basic routing rules
type BasicRouteRule struct {
	HostNames   []string `json:"host_names"`
	Paths       []string `json:"paths"`
	ClusterName string   `json:"cluster_name" validate:"required,min=1"`
	Description string   `json:"description"`
}

// ProductRouteRuleParam defines the request parameters for product route rules
type ProductRouteRuleParam struct {
	BasicRouteRules   []*BasicRouteRule   `json:"basic_forward_rules" validate:"dive"`
	AdvanceRouteRules []*AdvanceRouteRule `json:"forward_rules" validate:"dive"`
}

// Validate performs centralized business validation on the request parameters.
func (p *ProductRouteRuleParam) Validate() error {
	for i, one := range p.AdvanceRouteRules {
		if one == nil {
			return xerror.WrapParamErrorWithMsg("AdvanceRouteRules element cant be nil")
		}
		if err := validate.ClusterName(one.ClusterName); err != nil {
			return xerror.WrapParamErrorWithMsg("AdvanceRouteRules[%d].ClusterName: %v", i, err)
		}
	}
	for i, one := range p.BasicRouteRules {
		if one == nil {
			return xerror.WrapParamErrorWithMsg("BasicRouteRules element cant be nil")
		}
		if err := validate.ClusterName(one.ClusterName); err != nil {
			return xerror.WrapParamErrorWithMsg("BasicRouteRules[%d].ClusterName: %v", i, err)
		}
	}
	return nil
}

// ProductRouteRuleData defines the response data for product route rules
type ProductRouteRuleData struct {
	BasicRouteRules   []*BasicRouteRule   `json:"basic_forward_rules"`
	AdvanceRouteRules []*AdvanceRouteRule `json:"forward_rules"`

	RouteCasesCode int `json:"forward_cases_code,omitempty"`
}

func newProductRouteRuleData(pfr *ProductRouteRuleParam) *ProductRouteRuleData {
	return &ProductRouteRuleData{
		BasicRouteRules:   pfr.BasicRouteRules,
		AdvanceRouteRules: pfr.AdvanceRouteRules,
	}
}

func routeRule2routeRuleParam(p *iroute_conf.ProductRouteRule) *ProductRouteRuleParam {
	afrs := []*AdvanceRouteRule{}
	for _, one := range p.AdvanceRouteRules {
		afrs = append(afrs, &AdvanceRouteRule{
			Description: one.Description,
			ClusterName: one.ClusterName,
			Expression:  one.Expression,
			Name:        one.Name,
		})
	}

	bfrs := []*BasicRouteRule{}
	if p.BasicRouteRules != nil {
		for _, one := range p.BasicRouteRules {
			clusterName := one.ClusterName
			if clusterName == icluster_conf.RouteAdvancedModeClusterName4DP {
				clusterName = icluster_conf.RouteAdvancedModeClusterName
			}
			bfrs = append(bfrs, &BasicRouteRule{
				HostNames:   one.HostNames,
				Paths:       one.Paths,
				Description: one.Description,
				ClusterName: clusterName,
			})
		}
	}

	return &ProductRouteRuleParam{
		BasicRouteRules:   bfrs,
		AdvanceRouteRules: afrs,
	}
}

// ListRoute route
// AUTO GEN BY ctrl, MODIFY AS U NEED
var ListRoute = &xreq.Endpoint{
	Path:       "/ai-route-rules",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListAction),
	Authorizer: iauth.FAP(iauth.FeatureAIRoute, iauth.ActionRead),
}

func listActionProcess(req *http.Request) (*ProductRouteRuleData, error) {
	product, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}

	rule, err := container.RouteRuleManager.FetchProductRule(req.Context(), product)
	if err != nil {
		return nil, err
	}

	if rule == nil {
		return nullRule, nil
	}

	return newProductRouteRuleData(routeRule2routeRuleParam(rule)), nil
}

var nullRule = &ProductRouteRuleData{
	BasicRouteRules:   []*BasicRouteRule{},
	AdvanceRouteRules: []*AdvanceRouteRule{},
}

var _ xreq.Handler = ListAction

// ListAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func ListAction(req *http.Request) (interface{}, error) {
	return listActionProcess(req)
}
