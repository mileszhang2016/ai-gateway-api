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
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

var DeleteRoute = &xreq.Endpoint{
	Path:       "/api-keys/{id}",
	Method:     http.MethodDelete,
	Handler:    xreq.Convert(DeleteAction),
	Authorizer: iauth.FA(iauth.FeatureAPIKey, iauth.ActionDelete),
}

var _ xreq.Handler = DeleteAction

func DeleteAction(req *http.Request) (interface{}, error) {
	oneReq, err := newReq4One(req)
	if err != nil {
		return nil, err
	}

	productName := defaultProductName

	// Fetch API key to get quotaPlanID for deleting associated quota balances
	apiKey, err := container.APIKeyManager.FetchAPIKey(req.Context(), &icluster_conf.APIKeyFilter{
		ID:          oneReq.ID,
		ProductName: &productName,
	})
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, xerror.WrapRecordNotExist("API-Key")
	}

	// Delete quota balances associated with the quota plan
	if apiKey.QuotaPlanID != nil {
		if err := container.QuotaBalanceStorager.DeleteQuotaBalance(req.Context(), &quota.QuotaBalanceFilter{QuotaPlanID: apiKey.QuotaPlanID}); err != nil {
			return nil, err
		}
	}

	err = container.APIKeyManager.DeleteAPIKey(req.Context(), &icluster_conf.APIKeyFilter{
		ID:          oneReq.ID,
		ProductName: &productName,
	})
	return nil, err
}
