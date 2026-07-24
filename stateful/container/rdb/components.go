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

package rdb

import (
	"context"

	"github.com/yf-networks/ai-gateway-api/model/iai_route"
	"github.com/yf-networks/ai-gateway-api/model/iauth"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/imods"
	"github.com/yf-networks/ai-gateway-api/model/iprotocol"
	"github.com/yf-networks/ai-gateway-api/model/iroute_conf"
	"github.com/yf-networks/ai-gateway-api/model/iversion_control"
	"github.com/yf-networks/ai-gateway-api/model/quota"
	"github.com/yf-networks/ai-gateway-api/stateful"
	"github.com/yf-networks/ai-gateway-api/stateful/container"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/ai_route"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/auth"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/basic"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/cluster_conf"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/protocol"
	quotaStorage "github.com/yf-networks/ai-gateway-api/storage/rdb/quota"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/route_conf"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/txn"
	"github.com/yf-networks/ai-gateway-api/storage/rdb/version_control"
)

func Init() {
	container.TxnStoragerSingleton = txn.NewRDBTxnStorager(stateful.NewBFEDBContext)
	container.VersionControlStoragerSingleton = version_control.NewVersionControllerStorage(stateful.NewBFEDBContext)
	container.RouteRuleStoragerSingleton = route_conf.NewRouteRuleStorager(
		stateful.NewBFEDBContext,
		container.VersionControlStoragerSingleton)

	container.ProductStoragerSingleton = basic.NewProductManager(stateful.NewBFEDBContext)
	container.BFEClusterStoragerSingleton = basic.NewRDBBFEClusterStorager(stateful.NewBFEDBContext)
	container.PoolStoragerSingleton = cluster_conf.NewRDBPoolStorager(
		stateful.NewBFEDBContext,
		container.ProductStoragerSingleton)
	container.SubClusterStoragerSingleton = cluster_conf.NewRDBSubClusterStorager(
		stateful.NewBFEDBContext,
		container.PoolStoragerSingleton,
		container.ProductStoragerSingleton)
	container.ClusterStoragerSingleton = cluster_conf.NewRDBClusterStorager(
		stateful.NewBFEDBContext,
		container.SubClusterStoragerSingleton)

	container.APIKeyStorager = cluster_conf.NewAPIKeyStorager(
		stateful.NewBFEDBContext,
	)

	container.AIRouteRuleStorager = ai_route.NewRDBAIRouteRuleStorager(
		stateful.NewBFEDBContext,
	)
	container.CertificateStoragerSingleton = protocol.NewCertificateStorager(stateful.NewBFEDBContext)
	container.AuthenticateStoragerSingleton = auth.NewAuthenticateStorager(stateful.NewBFEDBContext)
	container.AuthorizeStoragerSingleton = auth.NewAuthorizeStorager(stateful.NewBFEDBContext,
		container.ProductStoragerSingleton,
		container.AuthenticateStoragerSingleton,
	)

	container.DomainStoragerSingleton = route_conf.NewDomainStorager(stateful.NewBFEDBContext)
	container.ExtraFileStoragerSingleton = basic.NewRDBExtraFileStorager(stateful.NewBFEDBContext)

	container.ExtraFileManager = ibasic.NewExtraFileManager(container.ExtraFileStoragerSingleton)
	container.VersionControlManager = iversion_control.NewVersionControllerManager(
		container.TxnStoragerSingleton,
		container.VersionControlStoragerSingleton)

	container.BFEClusterManager = ibasic.NewBFEClusterManager(
		container.TxnStoragerSingleton,
		container.BFEClusterStoragerSingleton)

	container.CertificateManager = iprotocol.NewCertificateManager(
		container.TxnStoragerSingleton,
		container.CertificateStoragerSingleton,
		container.VersionControlManager,
		container.ExtraFileStoragerSingleton)

	container.ProductManager = ibasic.NewProductManager(
		container.TxnStoragerSingleton,
		container.ProductStoragerSingleton)

	container.AIRouteRuleManager = iai_route.NewAIRouteRuleManager(
		container.TxnStoragerSingleton,
		container.AIRouteRuleStorager,
		container.VersionControlManager,
		container.RouteRuleStoragerSingleton,
	)
	container.RouteRuleManager = iroute_conf.NewRouteRuleManager(
		container.TxnStoragerSingleton,
		container.RouteRuleStoragerSingleton,
		container.ClusterStoragerSingleton,
		container.ProductStoragerSingleton,
		container.VersionControlManager,
		container.DomainStoragerSingleton)

	container.ClusterManager = icluster_conf.NewClusterManager(
		container.TxnStoragerSingleton,
		container.ClusterStoragerSingleton,
		container.SubClusterStoragerSingleton,
		container.BFEClusterStoragerSingleton,
		container.PoolStoragerSingleton,
		container.VersionControlManager,
		map[string]func(context.Context, *ibasic.Product, *icluster_conf.Cluster) error{
			"rules": container.RouteRuleManager.ClusterDeleteChecker,
		})

	container.SubClusterManager = icluster_conf.NewSubClusterManager(
		container.TxnStoragerSingleton,
		container.SubClusterStoragerSingleton,
		container.ProductStoragerSingleton,
		container.PoolStoragerSingleton,
		container.ClusterStoragerSingleton)

	container.DomainManager = iroute_conf.NewDomainManager(
		container.TxnStoragerSingleton,
		container.DomainStoragerSingleton,
		container.RouteRuleManager)

	container.AuthenticateManager = iauth.NewAuthenticateManager(
		container.TxnStoragerSingleton,
		container.AuthenticateStoragerSingleton,
		container.AuthorizeStoragerSingleton,
	)
	container.AuthorizeManager = iauth.NewAuthorizeManager(
		container.TxnStoragerSingleton,
		container.AuthorizeStoragerSingleton)

	container.PoolManager = icluster_conf.NewPoolManager(
		container.TxnStoragerSingleton,
		container.PoolStoragerSingleton,
		container.BFEClusterStoragerSingleton,
		container.SubClusterStoragerSingleton)

	// Initialize quota management components
	container.EntityTypeStorager = quotaStorage.NewEntityTypeStorager(stateful.NewBFEDBContext)
	container.EntityStorager = quotaStorage.NewEntityStorager(stateful.NewBFEDBContext)
	container.QuotaPlanStorager = quotaStorage.NewQuotaPlanStorager(stateful.NewBFEDBContext)
	container.QuotaBalanceStorager = quotaStorage.NewQuotaBalanceStorager(stateful.NewBFEDBContext)
	container.RateLimitPolicyStorager = quotaStorage.NewRateLimitPolicyStorager(stateful.NewBFEDBContext)

	container.EntityTypeManager = quota.NewEntityTypeManager(
		container.TxnStoragerSingleton,
		container.EntityTypeStorager)

	container.EntityManager = quota.NewEntityManager(
		container.TxnStoragerSingleton,
		container.EntityStorager,
		container.EntityTypeStorager,
		quota.NewQuotaPlanStoragerAdapter(container.QuotaPlanStorager),
		quota.NewRateLimitPolicyStoragerAdapter(container.RateLimitPolicyStorager),
		container.QuotaBalanceStorager)

	container.APIKeyRuleManager = imods.NewAPIKeyRuleManager(
		container.TxnStoragerSingleton,
		container.VersionControlManager,
		container.APIKeyStorager,
		container.AIRouteRuleStorager,
		container.QuotaPlanStorager,
		container.EntityStorager,
	)

	container.ModBodyProcessManager = imods.NewModBodyProcessManager(
		container.VersionControlManager,
	)

	container.QuotaPlanManager = quota.NewQuotaPlanManager(
		container.TxnStoragerSingleton,
		container.QuotaPlanStorager,
		container.QuotaBalanceStorager)

	container.RateLimitPolicyManager = quota.NewRateLimitPolicyManager(
		container.TxnStoragerSingleton,
		container.RateLimitPolicyStorager,
		container.APIKeyStorager,
		container.EntityStorager,
		container.VersionControlManager)

	container.APIKeyManager = icluster_conf.NewAPIKeyManager(
		container.TxnStoragerSingleton,
		container.APIKeyStorager,
		container.ClusterStoragerSingleton,
		quota.NewQuotaPlanStoragerAdapter(container.QuotaPlanStorager),
		quota.NewRateLimitPolicyStoragerAdapter(container.RateLimitPolicyStorager),
		quota.NewEntityStoragerAdapter(container.EntityStorager),
		quota.NewQuotaBalanceStoragerAdapter(container.QuotaBalanceStorager),
	)

	// Initialize quota balance sync and reset components
	container.BalanceSyncManager = quota.NewBalanceSyncManager(
		container.TxnStoragerSingleton,
		container.APIKeyStorager,
		container.QuotaBalanceStorager,
		container.QuotaPlanStorager,
		container.EntityStorager)

	container.QuotaResetScheduler = quota.NewQuotaResetScheduler(
		container.TxnStoragerSingleton,
		container.BalanceSyncManager)

	container.QuotaResetScheduler.Start()
}
