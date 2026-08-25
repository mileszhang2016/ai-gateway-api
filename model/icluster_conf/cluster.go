// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
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

package icluster_conf

import (
	"context"
	"fmt"
	"net"

	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
	"github.com/bfenetworks/bfe/bfe_config/bfe_route_conf/route_rule_conf"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
)

var (
	ClusterHashStrategyClientIDOnlyI     int32 = 0
	ClusterHashStrategyClientIPOnlyI     int32 = 1
	ClusterHashStrategyClientIDPreferedI int32 = 2

	ClusterHealthCheckHTTP      = "http"
	ClusterHealthCheckTCP       = "tcp"
	ClusterHealthCheckSchemaMap = map[string]bool{
		ClusterHealthCheckHTTP: true,
		ClusterHealthCheckTCP:  true,
	}

	ClusterStickTypeSubCluster = "SUB_CLUSTER"
	ClusterStickTypeInstance   = "INSTANCE"

	ClusterDefaultReqFlushInterval int32 = 0
	ClusterDefaultResFlushInterval int32 = -1 // -1: write response directly without using timing refresh
)

type ClusterBasicConnectionParam struct {
	MaxIdleConnPerRs    *int16
	CancelOnClientClose *bool
}

type ClusterBasicBuffersParam struct {
	ReqWriteBufferSize *int32
	ReqFlushInterval   *int32
	ResFlushInterval   *int32
}

type ClusterBasicRetriesParam struct {
	MaxRetryInSubcluster    *int8
	MaxRetryCrossSubcluster *int8
}

type ClusterBasicTimeoutsParam struct {
	TimeoutConnServ        *int32
	TimeoutResponseHeader  *int32
	TimeoutReadbodyClient  *int32
	TimeoutReadClientAgain *int32
	TimeoutWriteClient     *int32
}

type ClusterBasicParam struct {
	Connection *ClusterBasicConnectionParam
	Retries    *ClusterBasicRetriesParam
	Buffers    *ClusterBasicBuffersParam
	Timeouts   *ClusterBasicTimeoutsParam
	Protocol   *string
}

type ClusterStickySessionsParam struct {
	SessionSticky *bool
	HashStrategy  *int32
	HashHeader    *string
}

type ClusterPassiveHealthCheckParam struct {
	Schema     *string
	Interval   *int32
	Failnum    *int32
	Statuscode *int32
	Host       *string
	Uri        *string
}

type ClusterFilter struct {
	ID  *int64
	IDs []int64

	Names []string
	Name  *string

	Product *ibasic.Product
}

type ClusterParam struct {
	ID *int64

	Name *string

	ProductID *int64

	Description *string

	Basic *ClusterBasicParam

	StickySessions *ClusterStickySessionsParam

	SubClusters []string

	Scheduler map[string]map[string]int

	PassiveHealthCheck *ClusterPassiveHealthCheckParam

	LLMConfig *LLMConfig

	// InstancePool auto-creates instance-pool and sub-cluster from these instances
	InstancePool []Instance
}

type ClusterBasicConnection struct {
	MaxIdleConnPerRs    int16
	CancelOnClientClose bool
}

type ClusterBasicBuffers struct {
	ReqWriteBufferSize int32
	ReqFlushInterval   int32
	ResFlushInterval   int32
}

type ClusterBasicRetries struct {
	MaxRetryInSubcluster    int8
	MaxRetryCrossSubcluster int8
}

type ClusterBasicTimeouts struct {
	TimeoutConnServ        int32
	TimeoutResponseHeader  int32
	TimeoutReadbodyClient  int32
	TimeoutReadClientAgain int32
	TimeoutWriteClient     int32
}

type ClusterBasic struct {
	Connection *ClusterBasicConnection
	Retries    *ClusterBasicRetries
	Buffers    *ClusterBasicBuffers
	Timeouts   *ClusterBasicTimeouts
	Protocol   *string
}

type ClusterStickySessions struct {
	SessionSticky bool
	HashStrategy  int32
	HashHeader    string
}

type ClusterKeyRef struct {
	Name   *string `json:"name"`   // required; references provider key name; unique within keys
	Weight *int    `json:"weight"` // required; range [0,100]
}

// APIKey is the legacy key structure kept for backward compatibility in storage only.
type APIKey struct {
	Name   *string `json:"name"`   // required; length 1-128; unique within keys
	Key    *string `json:"key"`    // required; non-empty; length 1-512
	Weight *int    `json:"weight"` // required; range [0,100]
}

type KeyPolicy struct {
	Strategy            *string `json:"strategy"`              // default weighted_random
	MaxRetries          *int    `json:"max_retries"`           // default 0
	RetryBackoffInitial *int    `json:"retry_backoff_initial"` // default 500
	RetryBackoffMax     *int    `json:"retry_backoff_max"`     // default 5000
}

type LLMConfig struct {
	Models        []string        `json:"models"`         // model name list
	ModelMappings []*Mapping      `json:"model_mappings"` // model mapping
	Keys          []ClusterKeyRef `json:"keys"`           // references to provider keys with weights
	KeyPolicy     *KeyPolicy      `json:"key_policy"`     // key routing policy
	Provider      *string         `json:"provider"`       // provider name; required

	// MatchPrefix defines the provider/model prefix this cluster matches.
	// Must end with '/' to avoid matching model names themselves.
	MatchPrefix *string `json:"match_prefix"`
	// StripPrefix controls whether to strip MatchPrefix from the request model
	// field before forwarding to the backend.
	StripPrefix *bool `json:"strip_prefix"`
}

type Mapping struct {
	SourceModel *string `json:"source_model"`
	TargetModel *string `json:"target_model"`
}

type Endpoint struct {
	Schema  string            `json:"schema"`  // http、https
	URI     string            `json:"uri"`     // request uri
	Headers map[string]string `json:"headers"` // request headers
}

type ClusterPassiveHealthCheck struct {
	Schema     string
	Interval   int32
	Failnum    int32
	Statuscode int32
	Host       string
	Uri        string
}

func (cphc *ClusterPassiveHealthCheck) toBackendCheck() *cluster_conf.BackendCheck {
	if cphc == nil {
		return nil
	}

	int322intp := func(i int32) *int {
		tmp := int(i)
		return &tmp
	}

	return &cluster_conf.BackendCheck{
		Schem:         &cphc.Schema,
		Uri:           &cphc.Uri,
		Host:          &cphc.Host,
		FailNum:       int322intp(cphc.Failnum),
		CheckInterval: int322intp(cphc.Interval),
		StatusCode:    int322intp(cphc.Statuscode),
	}
}

type Cluster struct {
	ID          int64
	Name        string
	Description string
	Ready       bool
	ProductID   int64

	Basic              *ClusterBasic
	StickySessions     *ClusterStickySessions
	SubClusters        []*SubCluster
	Scheduler          map[string]map[string]int
	PassiveHealthCheck *ClusterPassiveHealthCheck
	LLMConfig          *LLMConfig
}

func (cluster *Cluster) SubClusterNames() []string {
	var names []string
	for _, one := range cluster.SubClusters {
		names = append(names, one.Name)
	}

	return names
}

func (cluster *Cluster) getBalanceMode() string {
	for _, sc := range cluster.SubClusters {
		if sc.Role == ProductPoolRoleEPP {
			return "EPP"
		}
	}
	return "WRR"
}

func ClusterList2MapByName(list []*Cluster) map[string]*Cluster {
	m := map[string]*Cluster{}
	for _, one := range list {
		m[one.Name] = one
	}

	return m
}

const (
	ResourceClusterRule = "cluster_rule"
)

func ClusterList2MapByID(list []*Cluster) map[int64]*Cluster {
	m := map[int64]*Cluster{}
	for _, one := range list {
		m[one.ID] = one
	}

	return m
}

func NewClusterManager(txn itxn.TxnStorager, storager ClusterStorager,
	subClusterStorager SubClusterStorager, bfeClusterStorager ibasic.BFEClusterStorager,
	poolStorager PoolStorager,
	providerStorager iprovider.ProviderStorager,
	versionControlManager *iversion_control.VersionControlManager,
	deleteCheckers map[string]func(context.Context, *ibasic.Product, *Cluster) error,
	updateCheckers map[string]func(context.Context, *ibasic.Product, *Cluster, *ClusterParam) error) *ClusterManager {

	return &ClusterManager{
		txn:                   txn,
		storager:              storager,
		subClusterStorager:    subClusterStorager,
		bfeClusterStorager:    bfeClusterStorager,
		poolStorager:          poolStorager,
		providerStorager:      providerStorager,
		versionControlManager: versionControlManager,

		deleteCheckers: deleteCheckers,
		updateCheckers: updateCheckers,
	}
}

type ClusterStorager interface {
	FetchCluster(ctx context.Context, param *ClusterFilter) (*Cluster, error)
	FetchClusterList(ctx context.Context, param *ClusterFilter) ([]*Cluster, error)
	ClusterUpdate(ctx context.Context, product *ibasic.Product, old *Cluster, param *ClusterParam) error
	ClusterCreate(ctx context.Context, product *ibasic.Product, param *ClusterParam, subClusters []*SubCluster) (int64, error)
	ClusterDelete(ctx context.Context, product *ibasic.Product, cluster *Cluster) error
	BindSubCluster(ctx context.Context, cluster *Cluster, appendSubClusters, unbindSubClusters []*SubCluster) error
	FetchLBMatrixList(ctx context.Context) (map[int64]map[string]map[string]int, error)
}

type ClusterManager struct {
	txn itxn.TxnStorager

	storager           ClusterStorager
	subClusterStorager SubClusterStorager
	bfeClusterStorager ibasic.BFEClusterStorager
	poolStorager       PoolStorager
	providerStorager   iprovider.ProviderStorager

	versionControlManager *iversion_control.VersionControlManager

	deleteCheckers map[string]func(context.Context, *ibasic.Product, *Cluster) error
	updateCheckers map[string]func(context.Context, *ibasic.Product, *Cluster, *ClusterParam) error
}

func (rm *ClusterManager) FetchClusterList(ctx context.Context, param *ClusterFilter) (list []*Cluster, err error) {
	err = rm.txn.AtomExecute(ctx, func(ctx context.Context) error {
		list, err = rm.storager.FetchClusterList(ctx, param)
		return err
	})

	return
}

func (cm *ClusterManager) FetchCluster(ctx context.Context, param *ClusterFilter) (*Cluster, error) {
	list, err := cm.FetchClusterList(ctx, param)
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (cm *ClusterManager) CreateCluster(ctx context.Context, product *ibasic.Product, param *ClusterParam) (err error) {
	param.ProductID = &product.ID

	err = cm.txn.AtomExecute(ctx, func(ctx context.Context) error {
		old, err := cm.storager.FetchClusterList(ctx, &ClusterFilter{
			Name: param.Name,
		})
		if err != nil {
			return err
		}
		if len(old) != 0 {
			return xerror.WrapRecordExisted("cluster")
		}

		// Resolve provider and auto-create instance-pool/sub-cluster from provider data.
		if param.LLMConfig != nil && param.LLMConfig.Provider != nil && *param.LLMConfig.Provider != "" {
			provider, err := cm.providerStorager.FetchProvider(ctx, &iprovider.ProviderFilter{Name: param.LLMConfig.Provider})
			if err != nil {
				return err
			}
			if provider == nil {
				return xerror.WrapParamErrorWithMsg("provider %s not exist", *param.LLMConfig.Provider)
			}
			if err := cm.validateClusterLLMConfigAgainstProvider(param.LLMConfig, provider); err != nil {
				return err
			}

			poolName := product.Name + "." + *param.Name
			clusterName := *param.Name

			var pN string
			pN, err = poolNameJudger(product.Name, poolName)
			if err != nil {
				return err
			}
			poolName = pN

			old, err := cm.poolStorager.FetchPool(ctx, poolName)
			if err != nil {
				return err
			}
			if old != nil {
				return xerror.WrapRecordExisted()
			}

			pool, err := cm.poolStorager.CreatePool(ctx, product, &PoolParam{
				Name:      &poolName,
				Instances: providerInstancesToClusterInstances(provider.InstancePool),
				Role:      lib.PString(ProductPoolRoleCommon),
				Tag:       &PoolTagProduct,
			})
			if err != nil {
				return err
			}

			err = cm.subClusterStorager.CreateSubCluster(ctx, &SubClusterParam{
				Name:         &clusterName,
				PoolName:     &poolName,
				Product:      product,
				InstancePool: pool,
				Cluster: &Cluster{
					ID:   -1,
					Name: "unbinding",
				},
			})
			if err != nil {
				return err
			}

			if param.Scheduler == nil {
				defaultClusterName := stateful.DefaultConfig.RunTime.DefaultAIClusterName
				param.Scheduler = map[string]map[string]int{
					defaultClusterName: {
						clusterName: 100,
						BlackHole:   0,
					},
				}
			}

			if len(param.SubClusters) == 0 {
				param.SubClusters = []string{clusterName}
			}
		}

		bindingSubClusters, err := cm.subClusterStorager.FetchSubClusterList(ctx, &SubClusterFilter{
			Names:   param.SubClusters,
			Product: product,
		})
		if err != nil {
			return err
		}

		subClusterCount := len(bindingSubClusters)
		if subClusterCount > 0 {
			oneSubCluster := bindingSubClusters[0]
			switch oneSubCluster.Role {
			case ProductPoolRoleEPP:
				if subClusterCount != 1 {
					return xerror.WrapParamErrorWithMsg(fmt.Sprintf("subcluster is EPP, then must be one subcluster"))
				}
			}
		}

		err = cm.checkBindingSubClusters(ctx, nil, param.SubClusters, bindingSubClusters)
		if err != nil {
			return err
		}
		if err := cm.checkManualLB(ctx, nil, param); err != nil {
			return err
		}

		if param.Scheduler == nil {
			if param.Scheduler, err = cm.constructDefaultScheduler(ctx, bindingSubClusters); err != nil {
				return err
			}
		}

		clusterID, err := cm.storager.ClusterCreate(ctx, product, param, bindingSubClusters)
		if err != nil {
			return err
		}

		return cm.storager.BindSubCluster(ctx, &Cluster{
			ID: clusterID,
		}, bindingSubClusters, nil)
	})

	return
}

const BlackHole = "GSLB_BLACKHOLE"

func (cm *ClusterManager) validateClusterLLMConfigAgainstProvider(llm *LLMConfig, provider *iprovider.Provider) error {
	if llm == nil || provider == nil {
		return nil
	}

	providerModels := map[string]bool{}
	for _, m := range provider.Models {
		providerModels[m] = true
	}
	for _, m := range llm.Models {
		if !providerModels[m] {
			return xerror.WrapParamErrorWithMsg("model %s not in provider %s models", m, provider.Name)
		}
	}

	providerKeys := map[string]bool{}
	for _, k := range provider.Keys {
		providerKeys[k.Name] = true
	}
	if len(llm.Keys) > 0 {
		for _, k := range llm.Keys {
			if k.Name == nil || *k.Name == "" {
				return xerror.WrapParamErrorWithMsg("llm_config.keys.name is required")
			}
			if !providerKeys[*k.Name] {
				return xerror.WrapParamErrorWithMsg("key %s not found in provider %s", *k.Name, provider.Name)
			}
		}
	}

	return nil
}

func providerInstancesToClusterInstances(instances []iprovider.ProviderInstance) []Instance {
	rst := make([]Instance, 0, len(instances))
	for _, inst := range instances {
		rst = append(rst, Instance{
			Name:    inst.Name,
			Addr:    inst.Addr,
			Port:    inst.Port,
			Weight:  inst.Weight,
			Disable: inst.Disable,
		})
	}
	return rst
}

func (cm *ClusterManager) constructDefaultScheduler(ctx context.Context, subClusters []*SubCluster) (map[string]map[string]int, error) {
	bfeClusters, err := cm.bfeClusterStorager.FetchBFEClusters(ctx, nil)
	if err != nil {
		return nil, err
	}

	lbMatrix := map[string]map[string]int{}

	rate := 100 / len(subClusters)
	mod := 100 - rate*len(subClusters)
	for _, bfeCluster := range bfeClusters {
		tmp := map[string]int{
			BlackHole: mod,
		}
		for _, subCluster := range subClusters {
			tmp[subCluster.Name] = rate
		}

		lbMatrix[bfeCluster.Name] = tmp
	}

	return lbMatrix, nil
}

func (cm *ClusterManager) checkManualLB(ctx context.Context, old *Cluster, param *ClusterParam) error {
	if param.Scheduler == nil {
		return nil
	}

	bfeClusters, err := cm.bfeClusterStorager.FetchBFEClusters(ctx, nil)
	if err != nil {
		return err
	}

	lbMatrix := param.Scheduler
	if len(bfeClusters) != len(lbMatrix) {
		return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, Want All BFE Cluster Exist")
	}

	subClusters := param.SubClusters
	if subClusters == nil && old != nil {
		subClusters = old.SubClusterNames()
	}

	bfeClusterMap := ibasic.BFEClusterNameMap(bfeClusters)
	for bfeClusterName, subClusterLbMatrix := range lbMatrix {
		if _, ok := bfeClusterMap[bfeClusterName]; !ok {
			return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, BFE Cluster %s Not Exist", bfeClusterName)
		}

		total := 0
		for subClusterName, rate := range subClusterLbMatrix {
			if subClusterName != BlackHole && !lib.StringSliceHasElement(subClusters, subClusterName) {
				return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, SubCluster %s Not In BFE Cluster %s Config", bfeClusterName, subClusterName)
			}
			if rate < 0 {
				return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, BFE Cluster %s Rate Must Bigger Than 0, Got %d", bfeClusterName, rate)
			}
			total += rate
		}
		if total != 100 {
			return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, BFE Cluster %s Total Rate Is %d, Want 100", bfeClusterName, total)
		}

		for _, subCluster := range subClusters {
			if _, ok := subClusterLbMatrix[subCluster]; !ok {
				return xerror.WrapParamErrorWithMsg("LbMatrix Config Illegal, SubCluster %s Not In BFE Cluster %s Config", bfeClusterName, subCluster)
			}
		}
	}

	return nil
}

func (cm *ClusterManager) checkBindingSubClusters(ctx context.Context, cluster *Cluster,
	bindingSubClusterNames []string, bindSubClusters []*SubCluster) error {

	if len(bindSubClusters) == 0 {
		return xerror.WrapModelErrorWithMsg("Cluster Want At Least On SubCluster")
	}

	var oldClusterID int64
	if cluster != nil {
		oldClusterID = cluster.ID
	}
	subClusterMap := SubClusterList2MapByName(bindSubClusters)
	for _, scName := range bindingSubClusterNames {
		subCluster := subClusterMap[scName]
		if subCluster == nil {
			return xerror.WrapModelErrorWithMsg("SubCluster %s Not Exist", scName)
		}

		if subCluster.ClusterID != oldClusterID && subCluster.ClusterID > 0 {
			return xerror.WrapModelErrorWithMsg("SubCluster %s be Mounted With Cluster %d", scName, subCluster.ClusterID)
		}

		if !stateful.IgnoreBNSStatusCheck && !subCluster.Ready {
			return xerror.WrapDependentUnReadyErrorWithMsg("SubCluster %s Not Ready", scName)
		}
	}

	return nil
}

func (cm *ClusterManager) UpdateCluster(ctx context.Context, product *ibasic.Product, oldData *Cluster,
	param *ClusterParam) (err error) {

	err = cm.txn.AtomExecute(ctx, func(ctx context.Context) error {
		// Validate and sync provider changes.
		if param.LLMConfig != nil && param.LLMConfig.Provider != nil && *param.LLMConfig.Provider != "" {
			provider, err := cm.providerStorager.FetchProvider(ctx, &iprovider.ProviderFilter{Name: param.LLMConfig.Provider})
			if err != nil {
				return err
			}
			if provider == nil {
				return xerror.WrapParamErrorWithMsg("provider %s not exist", *param.LLMConfig.Provider)
			}
			if err := cm.validateClusterLLMConfigAgainstProvider(param.LLMConfig, provider); err != nil {
				return err
			}

			// Sync instance pool if provider changed.
			oldProvider := ""
			if oldData.LLMConfig != nil && oldData.LLMConfig.Provider != nil {
				oldProvider = *oldData.LLMConfig.Provider
			}
			if oldProvider != *param.LLMConfig.Provider {
				for _, sc := range oldData.SubClusters {
					if sc.InstancePool != nil {
						if err = cm.poolStorager.UpdatePool(ctx, sc.InstancePool, &PoolParam{
							Instances: providerInstancesToClusterInstances(provider.InstancePool),
						}); err != nil {
							return err
						}
						break
					}
				}
			}
		}

		if err = cm.checkManualLB(ctx, oldData, param); err != nil {
			return err
		}

		subClusterCount := len(oldData.SubClusters)
		if subClusterCount > 0 {
			oneSubCluster := oldData.SubClusters[0]
			switch oneSubCluster.Role {
			case ProductPoolRoleEPP:
				if subClusterCount != 1 {
					return xerror.WrapParamErrorWithMsg(fmt.Sprintf("subcluster is EPP, then must be one subcluster"))
				}
			}
		}

		for _, checker := range cm.updateCheckers {
			if err = checker(ctx, product, oldData, param); err != nil {
				return err
			}
		}

		return cm.storager.ClusterUpdate(ctx, product, oldData, param)
	})

	return
}

// ProviderDeleteChecker returns an error if any cluster references the provider.
func (cm *ClusterManager) ProviderDeleteChecker(ctx context.Context, providerName string) error {
	clusters, err := cm.storager.FetchClusterList(ctx, nil)
	if err != nil {
		return err
	}
	for _, c := range clusters {
		if c.LLMConfig != nil && c.LLMConfig.Provider != nil && *c.LLMConfig.Provider == providerName {
			return xerror.WrapConflictErrorWithMsg("provider %s is referenced by cluster %s", providerName, c.Name)
		}
	}
	return nil
}

func (cm *ClusterManager) checkLbMatrix(cluster *Cluster, unbindSubClusters, appendSubClusters []string) (map[string]map[string]int, error) {
	unbindSubClusterMap := lib.StringSlice2Map(unbindSubClusters)
	newManualLbMatrix := map[string]map[string]int{}

	for bfeCluster, subClusterRate := range cluster.Scheduler {
		newManualLbMatrix[bfeCluster] = map[string]int{}
		for subClusterName, rate := range subClusterRate {
			if unbindSubClusterMap[subClusterName] {
				if rate := subClusterRate[subClusterName]; rate != 0 {
					return nil, xerror.WrapModelErrorWithMsg(
						"BFE Cluster %s, SubCluster: %s Rate is %d, Set to 0 Before Unbind", bfeCluster, subClusterName, rate)
				}
			} else {
				newManualLbMatrix[bfeCluster][subClusterName] = rate
			}

			for _, subClusterName := range appendSubClusters {
				newManualLbMatrix[bfeCluster][subClusterName] = 0
			}
		}
	}

	return newManualLbMatrix, nil
}

func (cm *ClusterManager) RebindSubCluster(ctx context.Context, product *ibasic.Product, cluster *Cluster,
	bindingSubClusterNames []string) error {

	unbindSubClusterNames := lib.StringSliceSubtract(cluster.SubClusterNames(), bindingSubClusterNames)
	appendSubClusterNames := lib.StringSliceSubtract(bindingSubClusterNames, cluster.SubClusterNames())
	if len(unbindSubClusterNames) == 0 && len(appendSubClusterNames) == 0 {
		return nil
	}

	manualLbMatrix, err := cm.checkLbMatrix(cluster, unbindSubClusterNames, appendSubClusterNames)
	if err != nil {
		return err
	}

	return cm.txn.AtomExecute(ctx, func(ctx context.Context) error {
		bindingSubClusters, err := cm.subClusterStorager.FetchSubClusterList(ctx, &SubClusterFilter{
			Names:   bindingSubClusterNames,
			Product: product,
		})
		if err != nil {
			return err
		}

		if err = cm.checkBindingSubClusters(ctx, cluster, bindingSubClusterNames, bindingSubClusters); err != nil {
			return err
		}

		if manualLbMatrix != nil {
			if err := cm.storager.ClusterUpdate(ctx, product, cluster, &ClusterParam{
				Scheduler: manualLbMatrix,
			}); err != nil {
				return err
			}
		}

		var appendSubClusters, unbindSubClusters []*SubCluster
		if len(unbindSubClusterNames) > 0 {
			scMap := SubClusterList2MapByName(cluster.SubClusters)
			for _, one := range unbindSubClusterNames {
				unbindSubClusters = append(unbindSubClusters, scMap[one])
			}
		}

		scMap := SubClusterList2MapByName(bindingSubClusters)
		for _, one := range appendSubClusterNames {
			appendSubClusters = append(appendSubClusters, scMap[one])
		}

		// U should check param by yourself
		return cm.storager.BindSubCluster(ctx, cluster, appendSubClusters, unbindSubClusters)
	})
}

func (cm *ClusterManager) DeleteCluster(ctx context.Context, product *ibasic.Product, cluster *Cluster) (err error) {
	err = cm.txn.AtomExecute(ctx, func(ctx context.Context) error {
		for _, checker := range cm.deleteCheckers {
			err = checker(ctx, product, cluster)
			if err != nil {
				return err
			}
		}

		// Cascade delete sub-clusters and their instance-pools
		for _, sc := range cluster.SubClusters {
			if err = cm.subClusterStorager.DeleteSubCluster(ctx, sc); err != nil {
				return err
			}
			if sc.InstancePool != nil {
				if err = cm.poolStorager.DeletePool(ctx, sc.InstancePool); err != nil {
					return err
				}
			}
		}

		if err = cm.storager.BindSubCluster(ctx, cluster, nil, cluster.SubClusters); err != nil {
			return err
		}

		if err = cm.storager.ClusterDelete(ctx, product, cluster); err != nil {
			return err
		}

		return nil
	})

	return
}

// IsBFEClusterUsed checks if a BFE cluster is referenced in any lb_matrix (scheduler) of any AI cluster
func (cm *ClusterManager) IsBFEClusterUsed(ctx context.Context, bfeClusterName string) (bool, error) {
	lbMatrixMap, err := cm.storager.FetchLBMatrixList(ctx)
	if err != nil {
		return false, err
	}

	for _, lbMatrix := range lbMatrixMap {
		if _, ok := lbMatrix[bfeClusterName]; ok {
			return true, nil
		}
	}

	return false, nil
}

var (
	UnMountedClusterID int64 = -1

	RouteAdvancedModeClusterName4DP       = route_rule_conf.AdvancedMode
	RouteAdvancedModeClusterName          = "GO_TO_ADVANCED_RULES"
	RouteAdvancedModeClusterID      int64 = -1

	SystemKeepRouteNames = map[string]bool{
		RouteAdvancedModeClusterName:    true,
		RouteAdvancedModeClusterName4DP: true,
	}
)

func AppendAdvancedRuleCluster(list []*Cluster) []*Cluster {
	return append(list, &Cluster{
		ID:   RouteAdvancedModeClusterID,
		Name: RouteAdvancedModeClusterName4DP,
	})
}

func normalizeBFEHashStrategy(sticky *ClusterStickySessions) int32 {
	if sticky == nil {
		return ClusterHashStrategyClientIPOnlyI
	}

	// BFE requires HashHeader to be non-empty when HashStrategy is CLIENT_ID_ONLY
	// or CLIENT_ID_PREFERRED. If session sticky is disabled and no HashHeader is
	// provided, fall back to CLIENT_IP_ONLY to keep the generated config valid.
	if !sticky.SessionSticky && sticky.HashHeader == "" &&
		(sticky.HashStrategy == ClusterHashStrategyClientIDOnlyI ||
			sticky.HashStrategy == ClusterHashStrategyClientIDPreferedI) {
		return ClusterHashStrategyClientIPOnlyI
	}

	return sticky.HashStrategy
}

type ProviderPricingInfo struct {
	TimeZone string
	Tiers    []cluster_conf.PriceTier
}

func NewBfeClusterConf(version string, clusters []*Cluster,
	providerModelTable map[string][]*imodel_price.ModelPrice,
	providerKeyTable map[string][]iprovider.ProviderKey,
	providerProtocolTable map[string][]string,
	providerPricingTable map[string]ProviderPricingInfo) *cluster_conf.BfeClusterConf {
	clusterConfMap := cluster_conf.ClusterToConf{}

	int322intp := func(i int32) *int {
		tmp := int(i)
		return &tmp
	}
	int162intp := func(i int16) *int {
		tmp := int(i)
		return &tmp
	}
	int82intp := func(i int8) *int {
		tmp := int(i)
		return &tmp
	}

	for _, cluster := range clusters {
		if SystemKeepRouteNames[cluster.Name] {
			continue
		}

		clusterConf := cluster_conf.ClusterConf{
			BackendConf: &cluster_conf.BackendBasic{
				Protocol:              cluster.Basic.Protocol,
				TimeoutConnSrv:        int322intp(cluster.Basic.Timeouts.TimeoutConnServ),
				TimeoutResponseHeader: int322intp(cluster.Basic.Timeouts.TimeoutResponseHeader),
				MaxIdleConnsPerHost:   int162intp(cluster.Basic.Connection.MaxIdleConnPerRs),
			},
			CheckConf: cluster.PassiveHealthCheck.toBackendCheck(),
			GslbBasic: &cluster_conf.GslbBasicConf{
				CrossRetry: int82intp(cluster.Basic.Retries.MaxRetryCrossSubcluster),
				RetryMax:   int82intp(cluster.Basic.Retries.MaxRetryInSubcluster),
				HashConf: &cluster_conf.HashConf{
					HashStrategy:  int322intp(normalizeBFEHashStrategy(cluster.StickySessions)),
					HashHeader:    &cluster.StickySessions.HashHeader,
					SessionSticky: &cluster.StickySessions.SessionSticky,
				},
				BalanceMode: lib.PString(cluster.getBalanceMode()),
			},
			ClusterBasic: &cluster_conf.ClusterBasicConf{
				TimeoutReadClient:      int322intp(cluster.Basic.Timeouts.TimeoutReadbodyClient),
				TimeoutWriteClient:     int322intp(cluster.Basic.Timeouts.TimeoutWriteClient),
				TimeoutReadClientAgain: int322intp(cluster.Basic.Timeouts.TimeoutReadClientAgain),
				ReqWriteBufferSize:     int322intp(cluster.Basic.Buffers.ReqWriteBufferSize),
				ReqFlushInterval:       int322intp(cluster.Basic.Buffers.ReqFlushInterval),
				ResFlushInterval:       int322intp(cluster.Basic.Buffers.ResFlushInterval),
				CancelOnClientClose:    &cluster.Basic.Connection.CancelOnClientClose,
			},
		}

		if clusterConf.GslbBasic.BalanceMode != nil && *clusterConf.GslbBasic.BalanceMode == ProductPoolRoleEPP {
			clusterConf.GslbBasic.EPPAddr = buildEPPAddrsFromSubClusters(cluster.SubClusters)
		}

		if cluster.Basic.Protocol != nil && *cluster.Basic.Protocol == "https" {
			clusterConf.HTTPSConf = &cluster_conf.BackendHTTPS{
				RSHost:               lib.PString(""),
				RSInsecureSkipVerify: lib.PBool(true),
				RSCAList:             nil,
				BFECertFile:          nil,
				BFEKeyFile:           nil,
			}
		}

		if isDomainPool(cluster.SubClusters) {
			clusterConf.ClusterBasic.DisableHealthCheck = lib.PBool(true)
			clusterConf.ClusterBasic.DisableHostHeader = lib.PBool(true)
		}

		if cluster.LLMConfig != nil {
			var modelTable *cluster_conf.ModelTable
			provider := ""
			if cluster.LLMConfig.Provider != nil {
				provider = *cluster.LLMConfig.Provider
			}
			if provider != "" {
				if entries, ok := providerModelTable[provider]; ok && len(entries) > 0 {
					models := make([]cluster_conf.ModelPrice, 0, len(entries))
					for _, e := range entries {
						if e != nil {
							models = append(models, cluster_conf.ModelPrice{
								Provider:            e.Provider,
								Model:               e.Model,
								BaseModel:           e.BaseModel,
								Mode:                e.Mode,
								Capabilities:        e.Capabilities,
								SupportedParameters: e.SupportedParameters,
								Limits:              e.Limits,
								Prices:              e.Prices,
								TierPrices:          e.TierPrices,
								Metadata:            e.Metadata,
							})
						}
					}
					pricingInfo := providerPricingTable[provider]
					timeZone := pricingInfo.TimeZone
					if timeZone == "" {
						timeZone = "Asia/Shanghai"
					}
					modelTable = &cluster_conf.ModelTable{
						Currency: "RMB",
						TimeZone: timeZone,
						Tiers:    pricingInfo.Tiers,
						Models:   models,
					}
				}
			}
			providerKeys := providerKeyTable[provider]
			providerProtocols := providerProtocolTable[provider]
			clusterConf.AIConf = newAIConf(cluster.LLMConfig, modelTable, providerKeys, providerProtocols)
		}

		clusterConfMap[cluster.Name] = clusterConf
	}
	return &cluster_conf.BfeClusterConf{
		Version: &version,
		Config:  &clusterConfMap,
	}
}

func newAIConf(llmConfig *LLMConfig, modelTable *cluster_conf.ModelTable,
	providerKeys []iprovider.ProviderKey, providerModelProtocols []string) *cluster_conf.AIConf {
	aiConf := &cluster_conf.AIConf{
		Type:           0,
		ModelMapping:   convertToBFEModelMapping(llmConfig.ModelMappings),
		Keys:           []cluster_conf.AIKey{},
		ModelProtocols: providerModelProtocols,
	}

	if llmConfig.Provider != nil {
		aiConf.Provider = *llmConfig.Provider
	}
	if modelTable != nil {
		aiConf.ModelTable = modelTable
	}
	if llmConfig.MatchPrefix != nil {
		aiConf.MatchPrefix = *llmConfig.MatchPrefix
	}
	if llmConfig.StripPrefix != nil {
		aiConf.StripPrefix = *llmConfig.StripPrefix
	}

	keyMap := map[string]string{}
	for _, k := range providerKeys {
		keyMap[k.Name] = k.Key
	}
	for _, k := range llmConfig.Keys {
		name := derefString(k.Name, "")
		aiConf.Keys = append(aiConf.Keys, cluster_conf.AIKey{
			Name:   name,
			Key:    keyMap[name],
			Weight: derefInt(k.Weight, 0),
		})
	}

	aiConf.KeyPolicy = &cluster_conf.AIKeyPolicy{
		Strategy:            "weighted_random",
		MaxRetries:          0,
		RetryBackoffInitial: 500,
		RetryBackoffMax:     5000,
	}
	if llmConfig.KeyPolicy != nil {
		aiConf.KeyPolicy.Strategy = derefString(llmConfig.KeyPolicy.Strategy, "weighted_random")
		aiConf.KeyPolicy.MaxRetries = derefInt(llmConfig.KeyPolicy.MaxRetries, 0)
		aiConf.KeyPolicy.RetryBackoffInitial = derefInt(llmConfig.KeyPolicy.RetryBackoffInitial, 500)
		aiConf.KeyPolicy.RetryBackoffMax = derefInt(llmConfig.KeyPolicy.RetryBackoffMax, 5000)
	}

	return aiConf
}

func derefString(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}

func derefInt(i *int, defaultValue int) int {
	if i == nil {
		return defaultValue
	}
	return *i
}

func convertToBFEModelMapping(modelMappings []*Mapping) *map[string]string {
	responseMap := make(map[string]string)
	for _, modelMapping := range modelMappings {
		responseMap[*modelMapping.SourceModel] = *modelMapping.TargetModel
	}

	if len(responseMap) == 0 {
		return nil
	}
	return &responseMap
}

func isDomainPool(subClusters []*SubCluster) bool {
	for _, subCluster := range subClusters {
		if subCluster.InstancePool != nil {
			for _, instance := range subCluster.InstancePool.Instances {
				ip := net.ParseIP(instance.Addr)
				if ip == nil {
					return true
				}
			}
		}
	}

	return false
}

func buildEPPAddrsFromSubClusters(subClusters []*SubCluster) *[]string {
	if len(subClusters) == 0 || subClusters[0] == nil || subClusters[0].InstancePool == nil {
		return nil
	}

	eppServer := subClusters[0].InstancePool.EPPServer
	if eppServer == nil {
		return nil
	}

	// Prefer domain+port mode; fallback to endpoints mode.
	if eppServer.Domain != nil && *eppServer.Domain != "" && eppServer.Port != nil && *eppServer.Port > 0 {
		addrs := []string{fmt.Sprintf("%s:%d", *eppServer.Domain, *eppServer.Port)}
		return &addrs
	}

	addrs := make([]string, 0, len(eppServer.Endpoints))
	for _, endpoint := range eppServer.Endpoints {
		if endpoint == nil || endpoint.IP == nil || endpoint.Port == nil {
			continue
		}
		if *endpoint.IP == "" || *endpoint.Port <= 0 {
			continue
		}

		addrs = append(addrs, fmt.Sprintf("%s:%d", *endpoint.IP, *endpoint.Port))
	}

	if len(addrs) == 0 {
		return nil
	}

	return &addrs
}
