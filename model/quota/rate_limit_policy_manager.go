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

package quota

import (
	"context"
	"fmt"

	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// RateLimitPolicyManager 定义限流策略管理器
type RateLimitPolicyManager struct {
	txn            itxn.TxnStorager
	storager       RateLimitPolicyStorager
	apiKeyStorager icluster_conf.APIKeyStorager
}

// NewRateLimitPolicyManager 创建限流策略管理器
func NewRateLimitPolicyManager(txn itxn.TxnStorager, storager RateLimitPolicyStorager, apiKeyStorager icluster_conf.APIKeyStorager) *RateLimitPolicyManager {
	return &RateLimitPolicyManager{
		txn:            txn,
		storager:       storager,
		apiKeyStorager: apiKeyStorager,
	}
}

// CreateRateLimitPolicy 创建限流策略
func (m *RateLimitPolicyManager) CreateRateLimitPolicy(ctx context.Context, param *RateLimitPolicyParam) (int64, error) {
	return m.storager.CreateRateLimitPolicy(ctx, param)
}

// FetchRateLimitPolicy 查询单个限流策略
func (m *RateLimitPolicyManager) FetchRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) (*RateLimitPolicyParam, error) {
	return m.storager.FetchRateLimitPolicy(ctx, filter)
}

// FetchRateLimitPolicyList 查询限流策略列表
func (m *RateLimitPolicyManager) FetchRateLimitPolicyList(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
	return m.storager.FetchRateLimitPolicyList(ctx, filter)
}

// UpdateRateLimitPolicy 更新限流策略
func (m *RateLimitPolicyManager) UpdateRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter, param *RateLimitPolicyParam) (int64, error) {
	return m.storager.UpdateRateLimitPolicy(ctx, filter, param)
}

// DeleteRateLimitPolicy 删除限流策略
func (m *RateLimitPolicyManager) DeleteRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) error {
	return m.storager.DeleteRateLimitPolicy(ctx, filter)
}

// ConfigExport 导出限流策略配置供 BFE 使用
func (m *RateLimitPolicyManager) ConfigExport(ctx context.Context, lastVersion string) (*ExportRateLimitPolicyConfig, error) {
	// 获取所有 API Key
	apiKeys, err := m.apiKeyStorager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{})
	if err != nil {
		return nil, fmt.Errorf("fetch api keys error: %s", err.Error())
	}

	// 获取所有限流策略
	policies, err := m.storager.FetchRateLimitPolicyList(ctx, &RateLimitPolicyFilter{})
	if err != nil {
		return nil, fmt.Errorf("fetch rate limit policies error: %s", err.Error())
	}

	// 构建策略 ID 到策略的映射
	policyMap := make(map[int64]*RateLimitPolicyParam)
	for _, p := range policies {
		if p.ID != nil {
			policyMap[*p.ID] = p
		}
	}

	// 构建导出结构
	config := make(map[string][]*ExportRouteRule)
	rateLimitPolicies := make(map[string]*ExportRateLimitPolicy)
	bindings := make(map[string][]string)

	// 默认路由规则
	productName := "AI_product"
	config[productName] = []*ExportRouteRule{
		{
			Cond: "default_t()",
			HitAction: &ExportHitAction{
				Cmd:    "FINISH",
				Params: []string{},
			},
		},
	}

	// 构建策略导出和绑定关系
	for _, apiKey := range apiKeys {
		if apiKey.Key == nil || apiKey.RateLimitPolicyID == nil {
			continue
		}

		policyID := *apiKey.RateLimitPolicyID
		policy, exists := policyMap[policyID]
		if !exists {
			continue
		}

		// 构建策略导出
		policyKey := fmt.Sprintf("rlp-%d", policyID)
		if _, exists := rateLimitPolicies[policyKey]; !exists {
			exportPolicy := &ExportRateLimitPolicy{
				Name:           policyKey,
				Enabled:        policy.Enabled != nil && *policy.Enabled,
				MaxConcurrency: 0,
				TPM:            []ExportTPMConfig{},
				RPM:            []ExportRPMConfig{},
			}

			if policy.MaxConcurrency != nil {
				exportPolicy.MaxConcurrency = *policy.MaxConcurrency
			}

			// 转换 TPM 配置
			for _, tpm := range policy.TpmConfigs {
				models := []string{"*"}
				if tpm.Model != "" && tpm.Model != "*" {
					models = []string{tpm.Model}
				}
				exportPolicy.TPM = append(exportPolicy.TPM, ExportTPMConfig{
					Name:          tpm.Name,
					Models:        models,
					WindowMinutes: tpm.WindowMinutes,
					MaxTokens:     tpm.MaxTokens,
					StepMinutes:   tpm.StepMinutes,
				})
			}

			// 转换 RPM 配置
			for _, rpm := range policy.RpmConfigs {
				models := []string{"*"}
				if rpm.Model != "" && rpm.Model != "*" {
					models = []string{rpm.Model}
				}
				exportPolicy.RPM = append(exportPolicy.RPM, ExportRPMConfig{
					Name:          rpm.Name,
					Models:        models,
					WindowMinutes: rpm.WindowMinutes,
					MaxRequests:   rpm.MaxRequests,
					Burst:         1,
				})
			}

			rateLimitPolicies[policyKey] = exportPolicy
		}

		// 构建绑定关系
		if _, exists := bindings[*apiKey.Key]; !exists {
			bindings[*apiKey.Key] = []string{}
		}
		bindings[*apiKey.Key] = append(bindings[*apiKey.Key], policyKey)
	}

	// 生成版本号（简单使用策略数量）
	version := fmt.Sprintf("v1.0.%d", len(policies))

	return &ExportRateLimitPolicyConfig{
		Config:                        config,
		RateLimitPolicies:             rateLimitPolicies,
		ApikeyRateLimitPolicyBindings: bindings,
		Version:                       version,
	}, nil
}
