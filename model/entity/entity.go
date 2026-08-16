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
//limitations under the License.

package entity

import (
	"context"

	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

// EntityParam 定义 Entity 参数
type EntityParam struct {
	InnerID           *int64   `json:"-"`
	EntityID          *string  `json:"id"`
	Name              *string  `json:"name"`
	Type              *string  `json:"type"`
	ParentID          *string  `json:"parent_id"`
	AllowModels       []string `json:"allow_models"`
	BlockModels       []string `json:"block_models"`
	QuotaPlanID       *int64   `json:"-"`
	RateLimitPolicyID *int64   `json:"-"`
	RouteRulesID      *int64   `json:"-"`
	CreateTime        *int64   `json:"create_time,omitempty"`
	UpdateTime        *int64   `json:"update_time,omitempty"`

	QuotaPlan       *shared.QuotaPlanParam       `json:"quota_plan,omitempty"`
	RateLimitPolicy *shared.RateLimitPolicyParam `json:"rate_limit_policy,omitempty"`
	RouteRules      *shared.RouteRulesParam      `json:"route_rules,omitempty"`
}

// EntityFilter 定义 Entity 过滤条件
type EntityFilter struct {
	EntityID     *string `form:"id"`
	Name         *string `form:"name"`
	Type         *string `form:"type"`
	ParentID     *string `form:"parent_id"`
	QuotaPlanID  *int64  `form:"quota_plan_id"`
	RouteRulesID *int64  `form:"route_rules_id"`
	Page         *int    `form:"page"`
	PageSize     *int    `form:"page_size"`
}

// EntityStorager 定义 Entity 存储接口
type EntityStorager interface {
	CreateEntity(ctx context.Context, param *EntityParam) (int64, error)
	FetchEntity(ctx context.Context, filter *EntityFilter) (*EntityParam, error)
	FetchEntityList(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error)
	UpdateEntity(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error)
	DeleteEntity(ctx context.Context, filter *EntityFilter) error
}
