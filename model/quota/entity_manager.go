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

	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// EntityManager 定义 Entity 管理器
type EntityManager struct {
	txn      itxn.TxnStorager
	storager EntityStorager
}

// NewEntityManager 创建 Entity 管理器
func NewEntityManager(txn itxn.TxnStorager, storager EntityStorager) *EntityManager {
	return &EntityManager{
		txn:      txn,
		storager: storager,
	}
}

// CreateEntity 创建 Entity
func (m *EntityManager) CreateEntity(ctx context.Context, param *EntityParam) (int64, error) {
	return m.storager.CreateEntity(ctx, param)
}

// FetchEntity 查询单个 Entity
func (m *EntityManager) FetchEntity(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
	return m.storager.FetchEntity(ctx, filter)
}

// FetchEntityList 查询 Entity 列表
func (m *EntityManager) FetchEntityList(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
	return m.storager.FetchEntityList(ctx, filter)
}

// UpdateEntity 更新 Entity
func (m *EntityManager) UpdateEntity(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
	return m.storager.UpdateEntity(ctx, filter, param)
}

// DeleteEntity 删除 Entity
func (m *EntityManager) DeleteEntity(ctx context.Context, filter *EntityFilter) error {
	return m.storager.DeleteEntity(ctx, filter)
}
