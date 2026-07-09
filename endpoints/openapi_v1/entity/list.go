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

package entity

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var EntityListRoute = &xreq.Endpoint{
	Path:       "/entities",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(EntityListAction),
	Authorizer: iauth.FA(iauth.FeatureEntity, iauth.ActionReadAll),
}

type EntityListResponse struct {
	List       []*quota.EntityParam `json:"list"`
	Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Total    int `json:"total"`
	} `json:"pagination"`
}

func EntityListAction(req *http.Request) (interface{}, error) {
	filter := &quota.EntityFilter{}
	if err := xreq.BindForm(req, filter); err != nil {
		return nil, err
	}

	if filter.Page == nil || *filter.Page < 1 {
		defaultPage := 1
		filter.Page = &defaultPage
	}
	if filter.PageSize == nil || *filter.PageSize < 1 {
		defaultPageSize := 20
		filter.PageSize = &defaultPageSize
	}
	if *filter.PageSize > 100 {
		maxPageSize := 100
		filter.PageSize = &maxPageSize
	}

	if filter.Type != nil && *filter.Type == "" {
		filter.Type = nil
	}

	if filter.ParentID != nil && *filter.ParentID == "" {
		filter.ParentID = nil
	}

	list, err := container.EntityManager.FetchEntityList(req.Context(), filter)
	if err != nil {
		return nil, err
	}

	total := len(list)
	page := *filter.Page
	pageSize := *filter.PageSize

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		list = []*quota.EntityParam{}
	} else if end > total {
		list = list[start:]
	} else {
		list = list[start:end]
	}

	resp := &EntityListResponse{
		List: list,
	}
	resp.Pagination.Page = page
	resp.Pagination.PageSize = pageSize
	resp.Pagination.Total = total

	return resp, nil
}
