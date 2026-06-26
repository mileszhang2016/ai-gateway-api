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
	"context"
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

var _ xreq.Handler = APIKeyUpdateAction

var APIKeyUpdateRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(APIKeyUpdateAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionUpdate),
}

func APIKeyUpdateAction(req *http.Request) (interface{}, error) {
	// uri param
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	// body param
	param := &icluster_conf.APIKeyParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	param.ID = lib.PString(*oneReq.ID)

	return APIKeyUpdateProcess(req.Context(), param, defaultProduct())
}

func APIKeyUpdateProcess(ctx context.Context, param *icluster_conf.APIKeyParam, product *ibasic.Product) (*ibasic.Product, error) {
	if err := checkUpdateAPIKey(param, product.Name); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	return nil, container.APIKeyManager.UpdateAPIKey(ctx, &icluster_conf.APIKeyFilter{
		ID:          param.ID,
		ProductName: &product.Name,
	}, &icluster_conf.APIKeyParam{
		Enable:            param.Enable,
		Key:               param.Key,
		Description:       param.Description,
		IsLimit:           param.IsLimit,
		UnlimitedQuota:    param.UnlimitedQuota,
		Limit:             param.Limit,
		ExpiredTime:       param.ExpiredTime,
		AllowedModels:     param.AllowedModels,
		AllowedCIDR:       param.AllowedCIDR,
		Subnet:            param.Subnet,
		EntityID:          param.EntityID,
		QuotaPlanID:       param.QuotaPlanID,
		RateLimitPolicyID: param.RateLimitPolicyID,
		ProductName:       &product.Name,
	})
}
