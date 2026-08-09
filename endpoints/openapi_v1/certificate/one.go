// Copyright (c) 2021 The BFE Authors.
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

package certificate

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iprotocol"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// OneEndpoint route
var OneEndpoint = &xreq.Endpoint{
	Path:       "/certificates/{cert_name}",
	Method:     http.MethodGet,
	Handler:    xreq.Convert(OneAction),
	Authorizer: iauth.FA(iauth.FeatureCert, iauth.ActionRead),
}

var _ xreq.Handler = OneAction

// OneAction action
func OneAction(req *http.Request) (interface{}, error) {
	param, err := newOneParamFromReq(req)
	if err != nil {
		return nil, err
	}

	list, err := container.CertificateManager.FetchCertificates(req.Context(), &iprotocol.CertificateFilter{
		CertName: param.CertName,
	})
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, xerror.WrapRecordNotExist()
	}

	return newOneData(list[0]), nil
}
