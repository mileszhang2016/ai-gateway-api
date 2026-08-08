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

package bfe_pool

import (
	"net/http"
	"strings"

	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/openapi_v1/product_pool"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// CreateRoute route - deprecated, endpoint registration removed per optimization plan v1.2
// AUTO GEN BY ctrl, MODIFY AS U NEED
// var CreateEndpoint = &xreq.Endpoint{
// 	Path:       "/alb-pools",
// 	Method:     http.MethodPost,
// 	Handler:    xreq.Convert(CreateAction),
// 	Authorizer: iauth.FA(iauth.FeatureBFEPool, iauth.ActionReadAll),
// }

var _ xreq.Handler = CreateAction

// CreateAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func CreateAction(req *http.Request) (interface{}, error) {
	param, err := product_pool.NewUpsertParam(req)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(*param.Name, ibasic.BuildinProduct.Name+".") {
		return nil, xerror.WrapParamErrorWithMsg("Want Prefix %s.", ibasic.BuildinProduct.Name)
	}
	if len(*param.Name) == len(ibasic.BuildinProduct.Name)+1 {
		return nil, xerror.WrapParamErrorWithMsg("Want Pool Name")
	}

	oneData, err := container.PoolManager.CreateBFEPool(req.Context(), &icluster_conf.PoolParam{
		Name:      param.Name,
		Instances: product_pool.Instancesc2i(param.Instances),
	})
	if err != nil {
		return nil, err
	}

	return product_pool.NewOneData(oneData), nil
}
