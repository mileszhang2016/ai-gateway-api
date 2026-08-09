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

package product_cluster

import (
	"net/http"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/validate"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib/xreq"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

// PassiveHealthCheckParam Request Param
type PassiveHealthCheckParam struct {
	Interval   *int32  `json:"interval"`
	Failnum    *int32  `json:"failnum"`
	Statuscode *int32  `json:"statuscode"`
	Host       *string `json:"host"`
	Uri        *string `json:"uri"`
}

// Instance Request Param
type Instance struct {
	Name   string `json:"name"`
	Addr   string `json:"addr" validate:"required"`
	Port   int    `json:"port" validate:"required"`
	Weight int64  `json:"weight" validate:"min=0,max=100"`
}

// UpsertParam Request Param
type UpsertParam struct {
	Name           *string                  `json:"name" uri:"cluster_name"`
	Description    *string                  `json:"description"`
	Basic          *BasicParam              `json:"basic"`
	StickySessions *StickySessionsParam     `json:"sticky_sessions"`
	InstancePool   []*Instance              `json:"instance_pool"`
	PassiveHealthCheck *PassiveHealthCheckParam `json:"passive_health_check"`
	LLMConfig          *icluster_conf.LLMConfig `json:"llm_config"`
}

// ConnectionParam Request Param
type ConnectionParam struct {
	MaxIdleConnPerRs    *int16 `json:"max_idle_conn_per_rs"`
	CancelOnClientClose *bool  `json:"cancel_on_client_close"`
}

// BuffersParam Request Param
type BuffersParam struct {
	ReqWriteBufferSize *int32 `json:"req_write_buffer_size"`
}

// TimeoutsParam Request Param
type TimeoutsParam struct {
	TimeoutConnServ        *int32 `json:"timeout_conn_serv"`
	TimeoutResponseHeader  *int32 `json:"timeout_response_header"`
	TimeoutReadbodyClient  *int32 `json:"timeout_readbody_client"`
	TimeoutReadClientAgain *int32 `json:"timeout_read_client_again"`
	TimeoutWriteClient     *int32 `json:"timeout_write_client"`
}

// BasicParam Request Param
type BasicParam struct {
	Connection *ConnectionParam `json:"connection"`
	Retries    *RetriesParam    `json:"retries"`
	Buffers    *BuffersParam    `json:"buffers"`
	Timeouts   *TimeoutsParam   `json:"timeouts"`
	Protocol   *string          `json:"protocol"`
}

// RetriesParam Request Param
type RetriesParam struct {
	MaxRetryInCluster *int8 `json:"max_retry_in_cluster"`
}

// StickySessionsParam Request Param
type StickySessionsParam struct {
	Enabled      *bool   `json:"enabled"`
	HashStrategy *string `json:"hash_strategy"`
	HashHeader   *string `json:"hash_header"`
}

// Validate performs centralized business validation on the request parameters.
func (p *UpsertParam) Validate() error {
	if p.Name == nil || *p.Name == "" {
		return xerror.WrapParamErrorWithMsg("name is required")
	}
	if err := validate.ClusterName(*p.Name); err != nil {
		return err
	}
	if len(p.InstancePool) > 0 {
		// Default instance name to addr when not provided.
		for _, inst := range p.InstancePool {
			if inst.Name == "" {
				inst.Name = inst.Addr
			}
		}
		if err := validate.InstancePool(Instancesc2i(p.InstancePool)); err != nil {
			return err
		}
	}
	if err := validateStickySessions(p.StickySessions); err != nil {
		return err
	}
	if p.LLMConfig != nil {
		return validate.LLMConfig(p.LLMConfig)
	}
	return nil
}

// CreateRoute route
var CreateEndpoint = &xreq.Endpoint{
	Path:       "/clusters",
	Method:     http.MethodPost,
	Handler:    xreq.Convert(CreateAction),
	Authorizer: iauth.FA(iauth.FeatureProductCluster, iauth.ActionCreate),
}

var (
	clusterHashStrategyClientIDOnly     = "CLIENT_ID_ONLY"
	clusterHashStrategyClientIPOnly     = "CLIENT_IP_ONLY"
	clusterHashStrategyClientIDPrefered = "CLIENT_ID_PREFERED"
)

func validateStickySessions(ss *StickySessionsParam) error {
	if ss == nil || ss.Enabled == nil || !*ss.Enabled {
		return nil
	}

	hashStrategy := clusterHashStrategyClientIDOnly
	if ss.HashStrategy != nil && *ss.HashStrategy != "" {
		hashStrategy = *ss.HashStrategy
	}

	switch hashStrategy {
	case clusterHashStrategyClientIDOnly, clusterHashStrategyClientIDPrefered:
		if ss.HashHeader == nil || *ss.HashHeader == "" {
			return xerror.WrapParamErrorWithMsg("sticky_sessions.hash_header is required when enabled and hash_strategy is %s", hashStrategy)
		}
	case clusterHashStrategyClientIPOnly:
	default:
		return xerror.WrapParamErrorWithMsg("sticky_sessions.hash_strategy must be one of %s, %s, %s",
			clusterHashStrategyClientIPOnly, clusterHashStrategyClientIDOnly, clusterHashStrategyClientIDPrefered)
	}

	return nil
}

func newCreateParam4Create(req *http.Request) (*UpsertParam, error) {
	param := &UpsertParam{}
	if err := xreq.BindJSON(req, param); err != nil {
		return nil, err
	}

	if len(param.InstancePool) == 0 {
		return nil, xerror.WrapParamErrorWithMsg("instance_pool is required")
	}

	if param.LLMConfig == nil {
		return nil, xerror.WrapParamErrorWithMsg("llm_config is required")
	}

	return param, nil
}

func checkInstancePool(instances []*Instance) error {
	if len(instances) == 0 {
		return xerror.WrapParamErrorWithMsg("instance_pool is required")
	}

	hasPositiveWeight := false
	for i, instance := range instances {
		if instance.Weight > 0 {
			hasPositiveWeight = true
		}
		if instance.Port == 0 {
			return xerror.WrapParamErrorWithMsg("instance_pool[%d].port is required", i)
		}
	}

	if !hasPositiveWeight {
		return xerror.WrapParamErrorWithMsg("instance_pool must have at least one instance with weight > 0")
	}

	return nil
}

func clusterParamControlModel(param *UpsertParam) *icluster_conf.ClusterParam {
	rst := &icluster_conf.ClusterParam{
		Name:        param.Name,
		Description: param.Description,
		LLMConfig:   normalizeLLMConfig(param.LLMConfig),
	}

	basic := normalizeBasic(param.Basic)
	rst.Basic = &icluster_conf.ClusterBasicParam{
		Protocol: basic.Protocol,
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

	sticky := normalizeStickySessions(param.StickySessions)
	rst.StickySessions = &icluster_conf.ClusterStickySessionsParam{
		SessionSticky: sticky.Enabled,
		HashStrategy:  hashStrategyConvert(sticky.HashStrategy),
		HashHeader:    sticky.HashHeader,
	}

	phc := normalizePassiveHealthCheck(param.PassiveHealthCheck, param.InstancePool)
	rst.PassiveHealthCheck = PassiveHealthCheckParamC2M(phc)

	return rst
}

func normalizeBasic(basic *BasicParam) *BasicParam {
	if basic == nil {
		basic = &BasicParam{}
	}

	if basic.Protocol == nil {
		basic.Protocol = lib.PString("https")
	}

	switch *basic.Protocol {
	case "http", "https":
	default:
		basic.Protocol = lib.PString("https")
	}

	if basic.Connection == nil {
		basic.Connection = &ConnectionParam{}
	}
	if basic.Connection.MaxIdleConnPerRs == nil {
		basic.Connection.MaxIdleConnPerRs = lib.PInt16(0)
	}
	if basic.Connection.CancelOnClientClose == nil {
		basic.Connection.CancelOnClientClose = lib.PBool(false)
	}

	if basic.Retries == nil {
		basic.Retries = &RetriesParam{}
	}
	if basic.Retries.MaxRetryInCluster == nil {
		basic.Retries.MaxRetryInCluster = lib.PInt8(2)
	}

	if basic.Buffers == nil {
		basic.Buffers = &BuffersParam{}
	}
	if basic.Buffers.ReqWriteBufferSize == nil {
		basic.Buffers.ReqWriteBufferSize = lib.PInt32(512)
	}

	if basic.Timeouts == nil {
		basic.Timeouts = &TimeoutsParam{}
	}
	if basic.Timeouts.TimeoutConnServ == nil {
		basic.Timeouts.TimeoutConnServ = lib.PInt32(50000)
	}
	if basic.Timeouts.TimeoutResponseHeader == nil {
		basic.Timeouts.TimeoutResponseHeader = lib.PInt32(50000)
	}
	if basic.Timeouts.TimeoutReadbodyClient == nil {
		basic.Timeouts.TimeoutReadbodyClient = lib.PInt32(30000)
	}
	if basic.Timeouts.TimeoutReadClientAgain == nil {
		basic.Timeouts.TimeoutReadClientAgain = lib.PInt32(30000)
	}
	if basic.Timeouts.TimeoutWriteClient == nil {
		basic.Timeouts.TimeoutWriteClient = lib.PInt32(60000)
	}

	return basic
}

func normalizeStickySessions(ss *StickySessionsParam) *StickySessionsParam {
	if ss == nil {
		ss = &StickySessionsParam{}
	}
	if ss.Enabled == nil {
		ss.Enabled = lib.PBool(false)
	}
	if ss.HashStrategy == nil {
		// CLIENT_IP_ONLY does not require hash_header, so it is safe as the default.
		ss.HashStrategy = lib.PString(clusterHashStrategyClientIPOnly)
	}
	if ss.HashHeader == nil {
		ss.HashHeader = lib.PString("")
	}
	return ss
}

func normalizePassiveHealthCheck(phc *PassiveHealthCheckParam, instances []*Instance) *PassiveHealthCheckParam {
	if phc == nil {
		phc = &PassiveHealthCheckParam{}
	}
	if phc.Interval == nil {
		phc.Interval = lib.PInt32(1000)
	}
	if phc.Failnum == nil {
		phc.Failnum = lib.PInt32(3)
	}
	if phc.Statuscode == nil {
		phc.Statuscode = lib.PInt32(0)
	}
	if phc.Uri == nil {
		phc.Uri = lib.PString("/")
	}
	if phc.Host == nil {
		if len(instances) > 0 {
			phc.Host = &instances[0].Addr
		} else {
			empty := ""
			phc.Host = &empty
		}
	}
	return phc
}

func normalizeLLMConfig(llm *icluster_conf.LLMConfig) *icluster_conf.LLMConfig {
	if llm == nil {
		return nil
	}

	rst := &icluster_conf.LLMConfig{
		Models:        llm.Models,
		ModelMappings: llm.ModelMappings,
		Keys:          llm.Keys,
		KeyPolicy:     llm.KeyPolicy,
		ProviderType:  llm.ProviderType,
	}
	if rst.Keys == nil {
		rst.Keys = []icluster_conf.APIKey{}
	}

	if llm.ModelEndpoint != nil {
		rst.ModelEndpoint = &icluster_conf.Endpoint{
			Schema:  llm.ModelEndpoint.Schema,
			URI:     llm.ModelEndpoint.URI,
			Headers: llm.ModelEndpoint.Headers,
		}
	} else {
		rst.ModelEndpoint = &icluster_conf.Endpoint{
			Schema: "https",
			URI:    "/v1/models",
		}
	}

	if rst.ModelEndpoint.Schema == "" {
		rst.ModelEndpoint.Schema = "https"
	}
	if rst.ModelEndpoint.URI == "" {
		rst.ModelEndpoint.URI = "/v1/models"
	}

	return rst
}

func hashStrategyConvert(s *string) *int32 {
	if s == nil {
		return nil
	}

	return lib.PInt32(map[string]int32{
		clusterHashStrategyClientIDOnly:     icluster_conf.ClusterHashStrategyClientIDOnlyI,
		clusterHashStrategyClientIPOnly:     icluster_conf.ClusterHashStrategyClientIPOnlyI,
		clusterHashStrategyClientIDPrefered: icluster_conf.ClusterHashStrategyClientIDPreferedI,
	}[*s])
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
		name := instance.Name
		if name == "" {
			name = instance.Addr
		}

		rst = append(rst, icluster_conf.Instance{
			Name:   name,
			Addr:   instance.Addr,
			Port:   instance.Port,
			Weight: instance.Weight,
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
func CreateAction(req *http.Request) (interface{}, error) {
	param, err := newCreateParam4Create(req)
	if err != nil {
		return nil, err
	}

	return CreateActionProcess(req, param)
}
