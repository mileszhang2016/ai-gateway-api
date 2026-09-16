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

package report

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ireport"
)

// ReportOverviewRoute serves GET /open-api/v1/report/overview.
var ReportOverviewRoute = &xreq.Endpoint{
	Path:       "/report/overview",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ReportOverviewAction),
	Authorizer: iauth.FA(iauth.FeatureReport, iauth.ActionReadAll),
}

// ReportOverviewAction returns the overview metric card.
func ReportOverviewAction(req *http.Request) (interface{}, error) {
	query, err := bindCommonQuery(req)
	if err != nil {
		return nil, err
	}
	base, err := query.toBaseQuery()
	if err != nil {
		return nil, err
	}

	manager, err := reportManager(req)
	if err != nil {
		return nil, err
	}

	return manager.Overview(req.Context(), &ireport.OverviewQuery{BaseQuery: *base})
}
