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

package rdb

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iai_route"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iauth"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imods"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iprotocol"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iroute_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quotacache"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/rate_limit_policy"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/route_rules"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/ai_route"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/auth"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/basic"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/cluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/model_price"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/protocol"
	quotaStorage "github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/route_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/txn"
	"github.com/infinity-ai-gateway/ai-gateway-api/storage/rdb/version_control"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/entity"
)

func Init() error {
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
	// Initialize model pricing before route rule manager so InnerAPI exports can
	// attach ModelTable to AIConf.
	container.ModelPriceStorager = model_price.NewRDBModelPriceStorager(stateful.NewBFEDBContext)
	container.ModelPriceManager = imodel_price.NewManager(
		container.TxnStoragerSingleton,
		container.ModelPriceStorager)

	container.RouteRuleManager = iroute_conf.NewRouteRuleManager(
		container.TxnStoragerSingleton,
		container.RouteRuleStoragerSingleton,
		container.ClusterStoragerSingleton,
		container.ProductStoragerSingleton,
		container.VersionControlManager,
		container.DomainStoragerSingleton)
	container.RouteRuleManager.SetModelPriceStorager(container.ModelPriceStorager)

	// Initialize route rules components before cluster manager because the
	// cluster delete checker depends on RouteRulesManager.
	container.RouteRulesStorager = quotaStorage.NewRouteRulesStorager(stateful.NewBFEDBContext)
	container.RouteRulesManager = route_rules.NewRouteRulesManager(
		container.TxnStoragerSingleton,
		container.RouteRulesStorager)

	container.ClusterManager = icluster_conf.NewClusterManager(
		container.TxnStoragerSingleton,
		container.ClusterStoragerSingleton,
		container.SubClusterStoragerSingleton,
		container.BFEClusterStoragerSingleton,
		container.PoolStoragerSingleton,
		container.VersionControlManager,
		map[string]func(context.Context, *ibasic.Product, *icluster_conf.Cluster) error{
			"rules":       container.RouteRuleManager.ClusterDeleteChecker,
			"route_rules": container.RouteRulesManager.ClusterDeleteChecker,
		},
		map[string]func(context.Context, *ibasic.Product, *icluster_conf.Cluster, *icluster_conf.ClusterParam) error{
			"route_rules": container.RouteRulesManager.ClusterModelUpdateChecker,
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

	container.QuotaCacheSingleton = quotacache.NewRedisQuotaCache(
		stateful.DefaultClientSet.RedisClient,
	)

	container.EntityTypeManager = entity.NewEntityTypeManager(
		container.TxnStoragerSingleton,
		container.EntityTypeStorager)

	container.EntityManager = entity.NewEntityManager(
		container.TxnStoragerSingleton,
		container.EntityStorager,
		container.EntityTypeStorager,
		quota.NewQuotaPlanStoragerAdapter(container.QuotaPlanStorager),
		rate_limit_policy.NewRateLimitPolicyStoragerAdapter(container.RateLimitPolicyStorager),
		container.RouteRulesStorager,
		quota.NewQuotaBalanceStoragerAdapter(container.QuotaBalanceStorager),
		container.QuotaCacheSingleton)

	container.APIKeyRuleManager = imods.NewAPIKeyRuleManager(
		container.TxnStoragerSingleton,
		container.VersionControlManager,
		container.APIKeyStorager,
		container.AIRouteRuleStorager,
		container.QuotaPlanStorager,
		container.EntityStorager,
		container.QuotaCacheSingleton,
	)

	container.ModBodyProcessManager = imods.NewModBodyProcessManager(
		container.VersionControlManager,
	)

	container.QuotaPlanManager = quota.NewQuotaPlanManager(
		container.TxnStoragerSingleton,
		container.QuotaPlanStorager,
		container.QuotaBalanceStorager,
		container.APIKeyStorager,
		container.EntityStorager,
		container.QuotaCacheSingleton)

	container.RateLimitPolicyManager = rate_limit_policy.NewRateLimitPolicyManager(
		container.TxnStoragerSingleton,
		container.RateLimitPolicyStorager,
		container.APIKeyStorager,
		container.EntityStorager,
		container.VersionControlManager)

	container.AIRouteExporter = imods.NewAIRouteExporter(
		container.APIKeyStorager,
		container.EntityStorager,
		container.RouteRulesStorager,
		container.VersionControlManager)

	container.APIKeyManager = api_key.NewAPIKeyManager(
		container.TxnStoragerSingleton,
		container.APIKeyStorager,
		quota.NewQuotaPlanStoragerAdapter(container.QuotaPlanStorager),
		quota.NewRateLimitPolicyStoragerAdapter(container.RateLimitPolicyStorager),
		container.RouteRulesStorager,
		quota.NewEntityStoragerAdapter(container.EntityStorager),
		quota.NewQuotaBalanceStoragerAdapter(container.QuotaBalanceStorager),
		container.QuotaCacheSingleton,
	)

	// Initialize quota balance sync and reset components
	container.BalanceSyncManager = quota.NewBalanceSyncManager(
		container.TxnStoragerSingleton,
		container.APIKeyStorager,
		container.QuotaBalanceStorager,
		container.QuotaPlanStorager,
		container.EntityStorager,
		container.QuotaCacheSingleton)

	container.QuotaResetScheduler = quota.NewQuotaResetScheduler(
		container.TxnStoragerSingleton,
		container.BalanceSyncManager)

	container.QuotaResetScheduler.Start()

	// Ensure the global route table exists on startup.
	if err := container.RouteRulesManager.EnsureGlobalRouteRules(context.Background()); err != nil {
		return err
	}

	return nil
}
