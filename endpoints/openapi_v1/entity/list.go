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

func EntityListAction(req *http.Request) (interface{}, error) {
	filter := &quota.EntityFilter{}
	if err := xreq.Bind(req, filter); err != nil {
		return nil, err
	}

	return container.EntityManager.FetchEntityList(req.Context(), filter)
}
