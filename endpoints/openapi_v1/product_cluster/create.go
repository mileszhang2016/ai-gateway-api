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

// HealthCheckParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type HealthCheckParam struct {
	HealthcheckInterval   *int32  `json:"healthcheck_interval" validate:"required,min=1"`
	HealthcheckFailnum    *int32  `json:"healthcheck_failnum" validate:"required,min=1"`
	HealthcheckStatuscode *int32  `json:"healthcheck_statuscode" validate:"required,min=0"`
	HealthcheckHost       *string `json:"healthcheck_host" validate:"required,min=1"`
	HealthcheckUri        *string `json:"healthcheck_uri" validate:"required,min=1,startswith=/"`
}

type PassiveHealthCheckParam struct {
	Schema     *string `json:"schema"`
	Interval   *int32  `json:"interval" validate:"required,min=1"`
	Failnum    *int32  `json:"failnum" validate:"required,min=1"`
	Statuscode *int32  `json:"statuscode" validate:"required,min=0"`
	Host       *string `json:"host" validate:"required,min=1"`
	Uri        *string `json:"uri" validate:"required,min=1,startswith=/"`
}

// Instance Request Param
type Instance struct {
	Hostname string            `json:"hostname" validate:"required,min=2"`
	IP       string            `json:"ip" validate:"required"`
	Weight   int64             `json:"weight" validate:"min=0,max=100"`
	Ports    map[string]int    `json:"ports" validate:"required,min=1"`
	Tags     map[string]string `json:"tags"`
}

// UpsertParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type UpsertParam struct {
	Name           *string              `json:"name" uri:"cluster_name"`
	Description    *string              `json:"description"`
	Basic          *BasicParam          `json:"basic"`
	StickySessions *StickySessionsParam `json:"sticky_sessions"`

	// InstancePool replaces SubClusters - system auto-creates instance-pool and sub-cluster from this
	InstancePool []*Instance `json:"instance_pool"`

	// Scheduler is auto-generated, no longer required in request
	// Scheduler map[string]map[string]int `json:"scheduler"`

	PassiveHealthCheck *PassiveHealthCheckParam `json:"passive_health_check"`
	LLMConfig          *icluster_conf.LLMConfig `json:"llm_config"`
}

// ConnectionParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type ConnectionParam struct {
	MaxIdleConnPerRs    *int16 `json:"max_idle_conn_per_rs" validate:"required,min=0"`
	CancelOnClientClose *bool  `json:"cancel_on_client_close" validate:"required"`
}

// BuffersParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type BuffersParam struct {
	ReqWriteBufferSize *int32 `json:"req_write_buffer_size" validate:"required,min=0"`
}

// TimeoutsParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type TimeoutsParam struct {
	TimeoutConnServ        *int32 `json:"timeout_conn_serv" validate:"required,min=1"`
	TimeoutResponseHeader  *int32 `json:"timeout_response_header" validate:"required,min=1"`
	TimeoutReadbodyClient  *int32 `json:"timeout_readbody_client" validate:"required,min=1"`
	TimeoutReadClientAgain *int32 `json:"timeout_read_client_again" validate:"required,min=1"`
	TimeoutWriteClient     *int32 `json:"timeout_write_client" validate:"required,min=1"`
}

// BasicParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type BasicParam struct {
	Connection *ConnectionParam `json:"connection" validate:"required"`
	Retries    *RetriesParam    `json:"retries" validate:"required"`
	Buffers    *BuffersParam    `json:"buffers" validate:"required"`
	Timeouts   *TimeoutsParam   `json:"timeouts" validate:"required"`
	Protocol   *string          `json:"protocol"`
}

// RetriesParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type RetriesParam struct {
	MaxRetryInSubcluster    *int8 `json:"max_retry_in_subcluster" validate:"required,min=0"`
	MaxRetryCrossSubcluster *int8 `json:"max_retry_cross_subcluster" validate:"required,min=0"`
}

// StickySessionsParam Request Param
// AUTO GEN BY ctrl, MODIFY AS U NEED
type StickySessionsParam struct {
	SessionStickyType *string `json:"session_sticky_type" validate:"required,oneof=INSTANCE SUB_CLUSTER"`
	HashStrategy      *string `json:"hash_strategy" validate:"required,oneof=CLIENT_IP_ONLY CLIENT_ID_ONLY CLIENT_ID_PREFERED"`
	HashHeader        *string `json:"hash_header"`
}

// CreateRoute route
// AUTO GEN BY ctrl, MODIFY AS U NEED
var CreateEndpoint = &xreq.Endpoint{
	Path:       "/clusters",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(CreateAction),
	Authorizer: iauth.FAP(iauth.FeatureProductCluster, iauth.ActionCreate),
}

// AUTO GEN BY ctrl, MODIFY AS U NEED
func newCreateParam4Create(req *http.Request) (*UpsertParam, error) {
	param := &UpsertParam{
		StickySessions: &StickySessionsParam{
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
		},
	}
	err := xreq.BindJSON(req, param)
	if err != nil {
		return nil, err
	}

	// Scheduler is auto-generated from DefaultAIClusterName config

	if len(param.InstancePool) == 0 {
		return nil, xerror.WrapParamErrorWithMsg("InstancePool Want Be Set")
	}

	if param.Basic == nil {
		return nil, xerror.WrapParamErrorWithMsg("Basic Want Be Set")
	}

	if param.PassiveHealthCheck == nil {
		return nil, xerror.WrapParamErrorWithMsg("PassiveHealthCheck Want Be Set")
	}

	if param.StickySessions == nil {
		return nil, xerror.WrapParamErrorWithMsg("StickySessions Want Be Set")
	}
	if *param.StickySessions.HashStrategy != clusterHashStrategyClientIPOnly && param.StickySessions.HashHeader == nil {
		return nil, xerror.WrapParamErrorWithMsg("StickySessions.HashHeader Want Be Set")
	}

	if err := checkLLMConfig(param.LLMConfig); err != nil {
		return nil, err
	}

	if param.Basic.Protocol == nil {
		return nil, xerror.WrapParamErrorWithMsg("Basic.Protocol Want Be Set")
	}

	switch *param.Basic.Protocol {
	case "http":
	case "https":
	default:
		param.Basic.Protocol = lib.PString("http")
	}

	return param, err
}

var (
	clusterHashStrategyClientIDOnly     = "CLIENT_ID_ONLY"
	clusterHashStrategyClientIPOnly     = "CLIENT_IP_ONLY"
	clusterHashStrategyClientIDPrefered = "CLIENT_ID_PREFERED"
	defaultWeight                       = int64(1)
)

func clusterParamControlModel(param *UpsertParam) *icluster_conf.ClusterParam {
	rst := &icluster_conf.ClusterParam{
		Name:        param.Name,
		Description: param.Description,
		LLMConfig:   param.LLMConfig,
	}
	// SubClusters and Scheduler are auto-generated in CreateActionProcess

	if basic := param.Basic; basic != nil {
		rst.Basic = &icluster_conf.ClusterBasicParam{
			Protocol: basic.Protocol,
		}

		if conn := basic.Connection; conn != nil {
			rst.Basic.Connection = &icluster_conf.ClusterBasicConnectionParam{
				MaxIdleConnPerRs:    conn.MaxIdleConnPerRs,
				CancelOnClientClose: conn.CancelOnClientClose,
			}
		}

		if retries := basic.Retries; retries != nil {
			rst.Basic.Retries = &icluster_conf.ClusterBasicRetriesParam{
				MaxRetryInSubcluster:    retries.MaxRetryInSubcluster,
				MaxRetryCrossSubcluster: retries.MaxRetryCrossSubcluster,
			}
		}

		if buffers := basic.Buffers; buffers != nil {
			rst.Basic.Buffers = &icluster_conf.ClusterBasicBuffersParam{
				ReqWriteBufferSize: buffers.ReqWriteBufferSize,
				ReqFlushInterval:   &icluster_conf.ClusterDefaultReqFlushInterval,
				ResFlushInterval:   &icluster_conf.ClusterDefaultResFlushInterval,
			}
		}

		if timeouts := basic.Timeouts; timeouts != nil {
			rst.Basic.Timeouts = &icluster_conf.ClusterBasicTimeoutsParam{
				TimeoutConnServ:        timeouts.TimeoutConnServ,
				TimeoutResponseHeader:  timeouts.TimeoutResponseHeader,
				TimeoutReadbodyClient:  timeouts.TimeoutReadbodyClient,
				TimeoutReadClientAgain: timeouts.TimeoutReadClientAgain,
				TimeoutWriteClient:     timeouts.TimeoutWriteClient,
			}
		}
	}

	stickySessionConvert := func(s *string) *bool {
		if s == nil {
			return nil
		}

		return lib.PBool(*s == icluster_conf.ClusterStickTypeInstance)
	}

	hashStrategyConvert := func(s *string) *int32 {
		if s == nil {
			return nil
		}

		return lib.PInt32(map[string]int32{
			clusterHashStrategyClientIDOnly:     icluster_conf.ClusterHashStrategyClientIDOnlyI,
			clusterHashStrategyClientIPOnly:     icluster_conf.ClusterHashStrategyClientIPOnlyI,
			clusterHashStrategyClientIDPrefered: icluster_conf.ClusterHashStrategyClientIDPreferedI,
		}[*s])
	}

	if stickySession := param.StickySessions; stickySession != nil {
		rst.StickySessions = &icluster_conf.ClusterStickySessionsParam{
			SessionSticky: stickySessionConvert(stickySession.SessionStickyType),
			HashStrategy:  hashStrategyConvert(stickySession.HashStrategy),
			HashHeader:    stickySession.HashHeader,
		}
	}

	if passiveHealthCheck := param.PassiveHealthCheck; passiveHealthCheck != nil {
		rst.PassiveHealthCheck = PassiveHealthCheckParamC2M(passiveHealthCheck)
	}

	return rst
}

func PassiveHealthCheckParamC2M(passiveHealthCheck *PassiveHealthCheckParam) *icluster_conf.ClusterPassiveHealthCheckParam {
	if passiveHealthCheck == nil {
		return nil
	}

	return &icluster_conf.ClusterPassiveHealthCheckParam{
		Schema:     &icluster_conf.ClusterHealthCheckHTTP,
		Interval:   passiveHealthCheck.Interval,
		Failnum:    passiveHealthCheck.Failnum,
		Statuscode: passiveHealthCheck.Statuscode,
		Host:       passiveHealthCheck.Host,
		Uri:        passiveHealthCheck.Uri,
	}
}

func Instancesc2i(is []*Instance) []icluster_conf.Instance {
	rst := []icluster_conf.Instance{}
	for _, instance := range is {
		port := 0
		if instance.Ports != nil {
			port = instance.Ports["Default"]
		}

		weight := instance.Weight
		if weight == 0 {
			weight = defaultWeight
		}

		rst = append(rst, icluster_conf.Instance{
			HostName: instance.Hostname,
			IP:       instance.IP,
			Weight:   weight,
			Ports:    instance.Ports,
			Port:     port,
			Tags:     instance.Tags,
		})
	}

	return rst
}

func CreateActionProcess(req *http.Request, _param *UpsertParam) (*ClusterData, error) {
	product, err := getDefaultProduct(req.Context())
	if err != nil {
		return nil, err
	}

	param := clusterParamControlModel(_param)
	// InstancePool is auto-handled by model layer (creates pool + sub-cluster + scheduler in one transaction)
	if len(_param.InstancePool) == 0 {
		return nil, xerror.WrapParamErrorWithMsg("instance_pool is required")
	}
	param.InstancePool = Instancesc2i(_param.InstancePool)

	err = container.ClusterManager.CreateCluster(req.Context(), product, param)
	if err != nil {
		return nil, err
	}

	cluster, err := container.ClusterManager.FetchCluster(req.Context(), &icluster_conf.ClusterFilter{
		Name: param.Name,
	})
	if err != nil {
		return nil, err
	}

	return clusterModel2Control(cluster), nil
}

var _ xreq.Handler = CreateAction

// CreateAction action
// AUTO GEN BY ctrl, MODIFY AS U NEED
func CreateAction(req *http.Request) (interface{}, error) {
	param, err := newCreateParam4Create(req)
	if err != nil {
		return nil, err
	}

	return CreateActionProcess(req, param)
}
