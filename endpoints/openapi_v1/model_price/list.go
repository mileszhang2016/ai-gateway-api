// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package model_price

import (
	"net/http"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// ListEndpoint lists model price records with optional filters.
var ListEndpoint = &xreq.Endpoint{
	Path:       "/model-prices",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(ListAction),
	Authorizer: iauth.FA(iauth.FeatureModelPrice, iauth.ActionRead),
}

type pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type listResponse struct {
	List       []*imodel_price.ModelPrice `json:"list"`
	Pagination *pagination                `json:"pagination"`
}

// ListAction handles GET /model-prices. The route carries two contract shapes:
// with provider, model and mode all present it returns a single record
// (§3.6); any other combination returns the paginated list (§3.4).
func ListAction(req *http.Request) (interface{}, error) {
	filter := queryFilter(req)

	if filter.Provider != nil && filter.Model != nil && filter.Mode != nil {
		if !imodel_price.ValidModes[*filter.Mode] {
			return nil, xerror.WrapParamErrorWithMsg("invalid mode: %s", *filter.Mode)
		}
		one, err := container.ModelPriceManager.FetchModelPrice(req.Context(), filter)
		if err != nil {
			return nil, err
		}
		if one == nil {
			return nil, xerror.WrapRecordNotExist("ModelPrice")
		}
		return one, nil
	}
	if filter.Model != nil {
		return nil, xerror.WrapParamErrorWithMsg(
			"provider, model and mode are required for single-record query")
	}

	page, pageSize := pageFilter(req)
	filter.Page = &page
	filter.PageSize = &pageSize

	list, total, err := container.ModelPriceManager.FetchModelPriceList(req.Context(), filter)
	if err != nil {
		return nil, err
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
