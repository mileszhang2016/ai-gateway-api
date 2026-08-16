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

package route_tables

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

var Endpoints = []*xreq.Endpoint{
	RouteTablesListRoute,
}

var RouteTablesListRoute = &xreq.Endpoint{
	Path:       "/route-tables",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(RouteTablesListAction),
	Authorizer: iauth.FA(iauth.FeatureRoute, iauth.ActionReadAll),
}

type RouteTablesListResponse struct {
	List       []*shared.RouteTableParam `json:"list"`
	Pagination struct {
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
		Total    int64 `json:"total"`
	} `json:"pagination"`
}

func RouteTablesListAction(req *http.Request) (interface{}, error) {
	filter := &shared.RouteRulesFilter{}
	if err := xreq.BindForm(req, filter); err != nil {
		return nil, err
	}

	page := 1
	pageSize := 20
	if filter.Page != nil && *filter.Page > 0 {
		page = *filter.Page
	}
	if filter.PageSize != nil && *filter.PageSize > 0 {
		pageSize = *filter.PageSize
		if pageSize > 100 {
			pageSize = 100
		}
	}
	filter.Page = &page
	filter.PageSize = &pageSize

	if filter.Type != nil && *filter.Type == "" {
		filter.Type = nil
	}
	if filter.Owner != nil && *filter.Owner == "" {
		filter.Owner = nil
	}

	list, total, err := container.RouteRulesManager.ListRouteTables(req.Context(), filter)
	if err != nil {
		return nil, err
	}

	resp := &RouteTablesListResponse{
		List: list,
	}
	resp.Pagination.Page = page
	resp.Pagination.PageSize = pageSize
	resp.Pagination.Total = total

	return resp, nil
}
