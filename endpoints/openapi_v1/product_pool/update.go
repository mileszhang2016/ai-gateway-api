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
//limitations under the License. All rights reserved.
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

package product_pool

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// UpdateRoute route
// AUTO GEN BY ctrl, MODIFY AS U NEED
// deprecated, endpoint registration removed per optimization plan v1.2
// var UpdateEndpoint = &xreq.Endpoint{
// 	Path:       "/instance-pools/{instance_pool_name}",
// 	Method:     http.MethodPatch,
// 	Handler:    xreq.Convert(UpdateAction),
// 	Authorizer: iauth.FAP(iauth.FeatureProductPool, iauth.ActionUpdate),
// }

var _ xreq.Handler = UpdateAction

// UpdateAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func UpdateAction(req *http.Request) (interface{}, error) {
	param, err := NewUpsertParam(req)
	if err != nil {
		return nil, err
	}

	product, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}
	one, err := container.PoolManager.FetchProductPool(req.Context(), product, *param.Name)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, xerror.WrapRecordNotExist("Instance Pool")
	}

	err = container.PoolManager.UpdateProductPool(req.Context(), product, one, &icluster_conf.PoolParam{
		Instances: Instancesc2i(param.Instances),
		Role:      param.Role,
		EPPServer: param.EPPServer,
	})

	one.Instances = Instancesc2i(param.Instances)
	if param.Role != nil {
		one.Role = *param.Role
	}
	one.EPPServer = param.EPPServer

	return NewOneData(one), err
}
