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

	"github.com/yf-networks/ai-gateway-api/lib/xerror"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// EntityTypeManager 定义 Entity 类型管理器
type EntityTypeManager struct {
	txn      itxn.TxnStorager
	storager EntityTypeStorager
}

// NewEntityTypeManager 创建 Entity 类型管理器
func NewEntityTypeManager(txn itxn.TxnStorager, storager EntityTypeStorager) *EntityTypeManager {
	return &EntityTypeManager{
		txn:      txn,
		storager: storager,
	}
}

// CreateEntityType 创建 Entity 类型
func (m *EntityTypeManager) CreateEntityType(ctx context.Context, param *EntityTypeParam) (int64, error) {
	if param.TypeName == nil || *param.TypeName == "" {
		return 0, xerror.WrapParamErrorWithMsg("type_name is required")
	}

	existing, err := m.storager.FetchEntityType(ctx, &EntityTypeFilter{TypeName: param.TypeName})
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, xerror.WrapRecordExisted("Entity-Type")
	}

	return m.storager.CreateEntityType(ctx, param)
}

// FetchEntityType 查询单个 Entity 类型
func (m *EntityTypeManager) FetchEntityType(ctx context.Context, filter *EntityTypeFilter) (*EntityTypeParam, error) {
	return m.storager.FetchEntityType(ctx, filter)
}

// FetchEntityTypeList 查询 Entity 类型列表
func (m *EntityTypeManager) FetchEntityTypeList(ctx context.Context, filter *EntityTypeFilter) ([]*EntityTypeParam, error) {
	return m.storager.FetchEntityTypeList(ctx, filter)
}

// UpdateEntityType 更新 Entity 类型
func (m *EntityTypeManager) UpdateEntityType(ctx context.Context, filter *EntityTypeFilter, param *EntityTypeParam) (int64, error) {
	return m.storager.UpdateEntityType(ctx, filter, param)
}

// DeleteEntityType 删除 Entity 类型
func (m *EntityTypeManager) DeleteEntityType(ctx context.Context, filter *EntityTypeFilter) error {
	return m.storager.DeleteEntityType(ctx, filter)
}
