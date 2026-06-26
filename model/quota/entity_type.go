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

// EntityTypeParam 定义 Entity 类型参数
type EntityTypeParam struct {
	ID          *int64  `json:"id"`
	TypeName    *string `json:"type_name"`
	Description *string `json:"description"`
	Level       *int    `json:"level"`
}

// EntityTypeFilter 定义 Entity 类型过滤条件
type EntityTypeFilter struct {
	ID       *int64
	TypeName *string
	Level    *int
}

// EntityTypeStorager 定义 Entity 类型存储接口
type EntityTypeStorager interface {
	CreateEntityType(ctx context.Context, param *EntityTypeParam) (int64, error)
	FetchEntityType(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error)
	FetchEntityTypeList(ctx context.Context, filter *EntityTypeFilter) ([]*EntityTypeParam, error)
	UpdateEntityType(ctx context.Context, filter *EntityTypeFilter, param *EntityTypeParam) (int64, error)
	DeleteEntityType(ctx context.Context, filter *EntityTypeFilter) error
}
