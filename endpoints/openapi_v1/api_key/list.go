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

package api_key

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"

	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
)

var ListRoute = &xreq.Endpoint{
	Path:       "/api-keys",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListAction),
	Authorizer: iauth.FAP(iauth.FeatureAPIKey, iauth.ActionReadAll),
}

var _ xreq.Handler = OneAction

type APIKeyListReq struct {
	Page           *int    `form:"page"`
	PageSize       *int    `form:"page_size"`
	Enabled        *bool   `form:"enabled"`
	EntityID       *string `form:"entity_id"`
	UnlimitedQuota *bool   `form:"unlimited_quota"`
}

type APIKeyListResponse struct {
	List       []*icluster_conf.APIKeyParam `json:"list"`
	Pagination struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Total    int `json:"total"`
	} `json:"pagination"`
}

func ListAction(req *http.Request) (interface{}, error) {
	listReq := &APIKeyListReq{}
	if err := xreq.BindForm(req, listReq); err != nil {
		return nil, err
	}

	page := 1
	if listReq.Page != nil && *listReq.Page > 0 {
		page = *listReq.Page
	}

	pageSize := 20
	if listReq.PageSize != nil && *listReq.PageSize > 0 {
		pageSize = *listReq.PageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}

	productName := defaultProductName

	list, err := container.APIKeyManager.FetchAPIKeyList(req.Context(), &icluster_conf.APIKeyFilter{
		ProductName: &productName,
	})
	if err != nil {
		return nil, err
	}

	list, err = newResponse(list)
	if err != nil {
		return nil, err
	}

	total := len(list)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		list = []*icluster_conf.APIKeyParam{}
	} else if end > total {
		list = list[start:]
	} else {
		list = list[start:end]
	}

	resp := &APIKeyListResponse{
		List: list,
	}
	resp.Pagination.Page = page
	resp.Pagination.PageSize = pageSize
	resp.Pagination.Total = total

	return resp, nil
}

func newResponse(list []*icluster_conf.APIKeyParam) ([]*icluster_conf.APIKeyParam, error) {
	for i, one := range list {
		if one.Key != nil {
			remainingQuota, err := icluster_conf.GetRemainingQuota(one)
			if err != nil {
				return nil, err
			}
			list[i].RemainingQuota = remainingQuota
		}
	}

	return list, nil
}
