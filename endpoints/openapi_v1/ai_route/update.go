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

package ai_route

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/iroute_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

func routeRuleParam2routeRule(p *ProductRouteRuleParam) *iroute_conf.ProductRouteRule {
	afrs := []*iroute_conf.AdvanceRouteRule{}
	for _, one := range p.AdvanceRouteRules {
		afrs = append(afrs, &iroute_conf.AdvanceRouteRule{
			Description: one.Description,
			ClusterName: one.ClusterName,
			Expression:  one.Expression,
			Name:        one.Name,
		})
	}

	// append default catch-all rule if not present, route to first advance rule's cluster
	if len(afrs) > 0 && afrs[len(afrs)-1].Expression != "default_t()" {
		afrs = append(afrs, &iroute_conf.AdvanceRouteRule{
			Description: "default catch-all rule",
			ClusterName: afrs[len(afrs)-1].ClusterName,
			Expression:  "default_t()",
			Name:        "default",
		})
	}

	bfrs := []*iroute_conf.BasicRouteRule{}
	for _, one := range p.BasicRouteRules {
		clusterName := one.ClusterName
		if clusterName == icluster_conf.RouteAdvancedModeClusterName {
			clusterName = icluster_conf.RouteAdvancedModeClusterName4DP
		}
		bfrs = append(bfrs, &iroute_conf.BasicRouteRule{
			HostNames:   one.HostNames,
			Paths:       one.Paths,
			Description: one.Description,
			ClusterName: clusterName,
		})
	}

	return &iroute_conf.ProductRouteRule{
		BasicRouteRules:   bfrs,
		AdvanceRouteRules: afrs,
	}
}

// UpdateRoute route
// AUTO GEN BY ctrl, MODIFY AS U NEED
var UpdateRoute = &xreq.Endpoint{
	Path:       "/ai-route-rules",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(UpdateAction),
	Authorizer: iauth.FAP(iauth.FeatureAIRoute, iauth.ActionUpdate),
}

// AUTO GEN BY ctrl, MODIFY AS U NEED
func newRuleInfoFromReq(req *http.Request) (*ProductRouteRuleParam, error) {
	rule := &ProductRouteRuleParam{}
	err := xreq.BindJSON(req, rule)
	if err != nil {
		return nil, err
	}

	return rule, err
}

func UpdateActionProcess(req *http.Request, rule *ProductRouteRuleParam) (*ProductRouteRuleData, error) {
	product, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}

	ipfr := routeRuleParam2routeRule(rule)

	err = container.RouteRuleManager.UpsertProductRule(req.Context(), product, ipfr)
	if err != nil {
		return nil, err
	}

	return newProductRouteRuleData(rule), nil
}

var _ xreq.Handler = UpdateAction

// UpdateAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func UpdateAction(req *http.Request) (interface{}, error) {
	rule, err := newRuleInfoFromReq(req)
	if err != nil {
		return nil, err
	}

	return UpdateActionProcess(req, rule)
}
