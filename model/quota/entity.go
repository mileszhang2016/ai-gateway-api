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

// EntityParam 定义 Entity 参数
type EntityParam struct {
	ID                *int64    `json:"id"`
	EntityID          *string   `json:"entity_id"`
	Name              *string   `json:"name"`
	Type              *string   `json:"type"`
	ParentID          *string   `json:"parent_id"`
	AllowModels       []string  `json:"allow_models"`
	BlockModels       []string  `json:"block_models"`
	QuotaPlanID       *int64    `json:"quota_plan_id"`
	RateLimitPolicyID *int64    `json:"rate_limit_policy_id"`
}

// EntityFilter 定义 Entity 过滤条件
type EntityFilter struct {
	ID       *int64
	EntityID *string
	Name     *string
	Type     *string
	ParentID *string
}

// EntityStorager 定义 Entity 存储接口
type EntityStorager interface {
	CreateEntity(ctx context.Context, param *EntityParam) (int64, error)
	FetchEntity(ctx context.Context, filter *EntityFilter) (*EntityParam, error)
	FetchEntityList(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error)
	UpdateEntity(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error)
	DeleteEntity(ctx context.Context, filter *EntityFilter) error
}
