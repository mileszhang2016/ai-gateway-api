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
	"strconv"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// ListEndpoint lists providers.
var ListEndpoint = &xreq.Endpoint{
	Path:       "/providers",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListAction),
	Authorizer: iauth.FA(iauth.FeatureProvider, iauth.ActionRead),
}

type pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type listResponse struct {
	List       []*iprovider.Provider `json:"list"`
	Pagination *pagination           `json:"pagination"`
}

// ListAction handles GET /providers.
func ListAction(req *http.Request) (interface{}, error) {
	filter := queryFilter(req)
	page, pageSize, isPagination := pageFilter(req)
	if isPagination {
		filter.Page = &page
		filter.PageSize = &pageSize
	}

	list, total, err := container.ProviderManager.FetchProviderList(req.Context(), filter)
	if err != nil {
		return nil, err
	}

	if !isPagination {
		page = 1
		pageSize = int(total)
		if pageSize < 1 {
			pageSize = 0
		}
	}

	return &listResponse{
		List: list,
		Pagination: &pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

func queryFilter(req *http.Request) *iprovider.ProviderFilter {
	q := req.URL.Query()
	filter := &iprovider.ProviderFilter{}
	if v := q.Get("model_protocol"); v != "" {
		filter.ModelProtocol = &v
	}
	return filter
}

func pageFilter(req *http.Request) (page int, pageSize int, isPagination bool) {
	q := req.URL.Query()
	pageStr := q.Get("page")
	pageSizeStr := q.Get("page_size")
	if pageStr == "" && pageSizeStr == "" {
		return 0, 0, false
	}

	page = 1
	pageSize = 50
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > 1000 {
				ps = 1000
			}
			pageSize = ps
		}
	}
	return page, pageSize, true
}
