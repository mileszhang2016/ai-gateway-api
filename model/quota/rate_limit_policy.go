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

import "context"

// TPMConfig 定义 TPM 限制配置
type TPMConfig struct {
	Name          string `json:"name"`
	Model         string `json:"model"`
	WindowMinutes int    `json:"window_minutes"`
	MaxTokens     int    `json:"max_tokens"`
	StepMinutes   int    `json:"step_minutes"`
}

// RPMConfig 定义 RPM 限制配置
type RPMConfig struct {
	Name          string `json:"name"`
	Model         string `json:"model"`
	WindowMinutes int    `json:"window_minutes"`
	MaxRequests   int    `json:"max_requests"`
}

// RateLimitPolicyParam 定义限流策略参数
type RateLimitPolicyParam struct {
	ID             *int64      `json:"id"`
	Enabled        *bool       `json:"enabled"`
	MaxConcurrency *int        `json:"max_concurrency"`
	TpmConfigs     []TPMConfig `json:"tpm_configs"`
	RpmConfigs     []RPMConfig `json:"rpm_configs"`
}

// RateLimitPolicyFilter 定义限流策略过滤条件
type RateLimitPolicyFilter struct {
	ID *int64
}

// RateLimitPolicyStorager 定义限流策略存储接口
type RateLimitPolicyStorager interface {
	CreateRateLimitPolicy(ctx context.Context, param *RateLimitPolicyParam) (int64, error)
	FetchRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) (*RateLimitPolicyParam, error)
	FetchRateLimitPolicyList(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error)
	UpdateRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter, param *RateLimitPolicyParam) (int64, error)
	DeleteRateLimitPolicy(ctx context.Context, filter *RateLimitPolicyFilter) error
}

// ExportRateLimitPolicy 定义导出到 BFE 的限流策略结构
type ExportRateLimitPolicy struct {
	Name           string            `json:"name"`
	Enabled        bool              `json:"enabled"`
	MaxConcurrency int               `json:"max_concurrency"`
	TPM            []ExportTPMConfig `json:"tpm"`
	RPM            []ExportRPMConfig `json:"rpm"`
}

// ExportTPMConfig 定义导出的 TPM 配置
type ExportTPMConfig struct {
	Name          string   `json:"name"`
	Models        []string `json:"models"`
	WindowMinutes int      `json:"window_minutes"`
	MaxTokens     int      `json:"max_tokens"`
	StepMinutes   int      `json:"step_minutes"`
}

// ExportRPMConfig 定义导出的 RPM 配置
type ExportRPMConfig struct {
	Name          string   `json:"name"`
	Models        []string `json:"models"`
	WindowMinutes int      `json:"window_minutes"`
	MaxRequests   int      `json:"max_requests"`
	Burst         int      `json:"burst"`
}

// ExportRateLimitPolicyConfig 定义导出的限流策略配置
type ExportRateLimitPolicyConfig struct {
	Config                        map[string][]*ExportRouteRule     `json:"Config"`
	RateLimitPolicies             map[string]*ExportRateLimitPolicy `json:"RateLimitPolicies"`
	ApikeyRateLimitPolicyBindings map[string][]string               `json:"ApikeyRateLimitPolicyBindings"`
	Version                       string                            `json:"Version"`
}

// ExportRouteRule 定义导出的路由规则
type ExportRouteRule struct {
	Cond      string           `json:"cond"`
	HitAction *ExportHitAction `json:"hit_action"`
}

// ExportHitAction 定义导出的命中动作
type ExportHitAction struct {
	Cmd    string   `json:"cmd"`
	Params []string `json:"params"`
}
