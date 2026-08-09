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
//limitations under the License.

package api_key

import (
	"context"
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var _ xreq.Handler = GetQuotaPlanAction

var GetQuotaPlanRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}/quota-plan",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(GetQuotaPlanAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionRead),
}

type GetQuotaPlanResponse struct {
	Unlimited             *bool    `json:"unlimited"`
	PassWhenNoEnoughQuota *bool    `json:"pass_when_no_enough_quota"`
	Quota                 *int64   `json:"quota"`
	Unit                  *string  `json:"unit"`
	ResetPeriod           *string  `json:"reset_period"`
	Balance               *Balance `json:"balance"`
}

type Balance struct {
	Used      int64 `json:"used"`
	Remaining int64 `json:"remaining"`
}

func GetQuotaPlanAction(req *http.Request) (interface{}, error) {
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	return GetQuotaPlanProcess(req.Context(), *oneReq.ID, defaultProduct())
}

func GetQuotaPlanProcess(ctx context.Context, id string, product *ibasic.Product) (*GetQuotaPlanResponse, error) {
	apiKey, err := container.APIKeyManager.FetchAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          &id,
		ProductName: &product.Name,
	})
	if err != nil {
		return nil, err
	}

	if apiKey == nil {
		return nil, xerror.WrapRecordNotExist("API-Key")
	}

	if apiKey.QuotaPlan == nil {
		return nil, nil
	}

	// TODO: Get balance from quota balance storage
	balance := &Balance{
		Used:      0,
		Remaining: *apiKey.QuotaPlan.Quota,
	}

	return &GetQuotaPlanResponse{
		Unlimited:             apiKey.QuotaPlan.Unlimited,
		PassWhenNoEnoughQuota: apiKey.QuotaPlan.PassWhenNoEnoughQuota,
		Quota:                 apiKey.QuotaPlan.Quota,
		Unit:                  apiKey.QuotaPlan.Unit,
		ResetPeriod:           apiKey.QuotaPlan.ResetPeriod,
		Balance:               balance,
	}, nil
}
