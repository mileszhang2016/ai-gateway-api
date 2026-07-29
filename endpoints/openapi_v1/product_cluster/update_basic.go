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

package product_cluster

import (
	"net/http"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
)

// UpdateBasicRoute route
var UpdateBasicEndpoint = &xreq.Endpoint{
	Path:       "/clusters/{cluster_name}",
	Method:     http.MethodPatch,
	Handler:    xreq.Convert(UpdateAction),
	Authorizer: iauth.FAP(iauth.FeatureProductCluster, iauth.ActionUpdate),
}

// AUTO GEN BY ctrl, MODIFY AS U NEED
func newUpdateParam4Update(req *http.Request) (*UpsertParam, error) {
	param := &UpsertParam{}
	if err := xreq.Bind(req, param); err != nil {
		return nil, err
	}

	if param.Name == nil || *param.Name == "" {
		return nil, xerror.WrapParamErrorWithMsg("cluster_name is required")
	}

	if param.LLMConfig != nil {
		if err := checkLLMConfig(param.LLMConfig); err != nil {
			return nil, err
		}
	}

	return param, nil
}

func updateActionProcess(req *http.Request, param *UpsertParam) (*ClusterData, error) {
	product, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}
	cluster, err := container.ClusterManager.FetchCluster(req.Context(), &icluster_conf.ClusterFilter{
		Name:    param.Name,
		Product: product,
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, xerror.WrapRecordNotExist("Cluster")
	}

	_modelParam := clusterParamControlModel4Update(param)
	if len(param.InstancePool) > 0 {
		_modelParam.InstancePool = Instancesc2i(param.InstancePool)
	}

	if err := container.ClusterManager.UpdateCluster(req.Context(), product, cluster, _modelParam); err != nil {
		return nil, err
	}

	cluster, err = container.ClusterManager.FetchCluster(req.Context(), &icluster_conf.ClusterFilter{
		ID: &cluster.ID,
	})
	if err != nil {
		return nil, err
	}
	return clusterModel2Control(cluster), nil
}

func clusterParamControlModel4Update(param *UpsertParam) *icluster_conf.ClusterParam {
	rst := &icluster_conf.ClusterParam{
		Name:        param.Name,
		Description: param.Description,
	}

	if param.LLMConfig != nil {
		rst.LLMConfig = normalizeLLMConfig(param.LLMConfig)
	}

	if param.Basic != nil {
		basic := param.Basic
		rst.Basic = &icluster_conf.ClusterBasicParam{
			Protocol: basic.Protocol,
		}
		if basic.Protocol != nil {
			switch *basic.Protocol {
			case "http", "https":
			default:
				rst.Basic.Protocol = lib.PString("http")
			}
		}
		if basic.Connection != nil {
			rst.Basic.Connection = &icluster_conf.ClusterBasicConnectionParam{
				MaxIdleConnPerRs:    basic.Connection.MaxIdleConnPerRs,
				CancelOnClientClose: basic.Connection.CancelOnClientClose,
			}
		}
		if basic.Retries != nil {
			rst.Basic.Retries = &icluster_conf.ClusterBasicRetriesParam{
				MaxRetryInSubcluster: basic.Retries.MaxRetryInCluster,
			}
		}
		if basic.Buffers != nil {
			rst.Basic.Buffers = &icluster_conf.ClusterBasicBuffersParam{
				ReqWriteBufferSize: basic.Buffers.ReqWriteBufferSize,
				ReqFlushInterval:   &icluster_conf.ClusterDefaultReqFlushInterval,
				ResFlushInterval:   &icluster_conf.ClusterDefaultResFlushInterval,
			}
		}
		if basic.Timeouts != nil {
			rst.Basic.Timeouts = &icluster_conf.ClusterBasicTimeoutsParam{
				TimeoutConnServ:        basic.Timeouts.TimeoutConnServ,
				TimeoutResponseHeader:  basic.Timeouts.TimeoutResponseHeader,
				TimeoutReadbodyClient:  basic.Timeouts.TimeoutReadbodyClient,
				TimeoutReadClientAgain: basic.Timeouts.TimeoutReadClientAgain,
				TimeoutWriteClient:     basic.Timeouts.TimeoutWriteClient,
			}
		}
	}

	if param.StickySessions != nil {
		sticky := param.StickySessions
		rst.StickySessions = &icluster_conf.ClusterStickySessionsParam{
			HashStrategy: hashStrategyConvert(sticky.HashStrategy),
		}
		if sticky.Enabled != nil {
			rst.StickySessions.SessionSticky = sticky.Enabled
		}
		if sticky.HashHeader != nil {
			rst.StickySessions.HashHeader = sticky.HashHeader
		}
	}

	if param.PassiveHealthCheck != nil {
		phc := param.PassiveHealthCheck
		rst.PassiveHealthCheck = &icluster_conf.ClusterPassiveHealthCheckParam{
			Schema:     &icluster_conf.ClusterHealthCheckHTTP,
			Interval:   phc.Interval,
			Failnum:    phc.Failnum,
			Statuscode: phc.Statuscode,
			Host:       phc.Host,
			Uri:        phc.Uri,
		}
	}

	return rst
}

var _ xreq.Handler = UpdateAction

// UpdateAction action
func UpdateAction(req *http.Request) (interface{}, error) {
	param, err := newUpdateParam4Update(req)
	if err != nil {
		return nil, err
	}

	return updateActionProcess(req, param)
}
