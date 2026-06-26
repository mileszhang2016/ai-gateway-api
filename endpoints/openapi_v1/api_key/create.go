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
	"fmt"
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

const defaultProductName = "AI_product"

func defaultProduct() *ibasic.Product {
	return &ibasic.Product{Name: defaultProductName}
}

var _ xreq.Handler = APIKeyCreateAction

var APIKeyCreateRoute = &xreq.Endpoint{
	Path:       "/api-keys",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(APIKeyCreateAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionCreate),
}

func APIKeyCreateAction(req *http.Request) (interface{}, error) {
	param := &icluster_conf.APIKeyParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	return APIKeyCreateProcess(req.Context(), param, defaultProduct())
}

func APIKeyCreateProcess(ctx context.Context, param *icluster_conf.APIKeyParam, product *ibasic.Product) (*ibasic.Product, error) {
	if err := checkCreateAPIKey(param, product.Name); err != nil {
		return nil, xerror.WrapParamError(err)
	}

	// Auto-generate API-Key ID
	apiKeyID, err := generateAPIKeyID(ctx, product.Name)
	if err != nil {
		return nil, err
	}

	err = container.APIKeyManager.CreateAPIKey(ctx, &icluster_conf.APIKeyParam{
		ID:                lib.PString(apiKeyID),
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

	return nil, err
}

// generateAPIKeyID generates a unique API-Key ID with format "api-key-{sequence}"
func generateAPIKeyID(ctx context.Context, productName string) (string, error) {
	// Get all API keys to find the max sequence number
	list, err := container.APIKeyManager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{
		ProductName: &productName,
	})
	if err != nil {
		return "", err
	}

	maxSeq := 0
	for _, apiKey := range list {
		if apiKey.ID != nil {
			var seq int
			if _, err := fmt.Sscanf(*apiKey.ID, "api-key-%d", &seq); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	return fmt.Sprintf("api-key-%d", maxSeq+1), nil
}
