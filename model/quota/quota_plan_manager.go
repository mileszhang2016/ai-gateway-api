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
	"time"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
)

// QuotaPlanManager 定义配额计划管理器
type QuotaPlanManager struct {
	txn             itxn.TxnStorager
	storager        QuotaPlanStorager
	balanceStorager QuotaBalanceStorager
}

// NewQuotaPlanManager 创建配额计划管理器
func NewQuotaPlanManager(txn itxn.TxnStorager, storager QuotaPlanStorager, balanceStorager QuotaBalanceStorager) *QuotaPlanManager {
	return &QuotaPlanManager{
		txn:             txn,
		storager:        storager,
		balanceStorager: balanceStorager,
	}
}

// CreateQuotaPlan 创建配额计划
func (m *QuotaPlanManager) CreateQuotaPlan(ctx context.Context, param *QuotaPlanParam) (int64, error) {
	return m.storager.CreateQuotaPlan(ctx, param)
}

// FetchQuotaPlan 查询单个配额计划
func (m *QuotaPlanManager) FetchQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
	return m.storager.FetchQuotaPlan(ctx, filter)
}

// FetchQuotaPlanList 查询配额计划列表
func (m *QuotaPlanManager) FetchQuotaPlanList(ctx context.Context, filter *QuotaPlanFilter) ([]*QuotaPlanParam, error) {
	return m.storager.FetchQuotaPlanList(ctx, filter)
}

// UpdateQuotaPlan 更新配额计划
func (m *QuotaPlanManager) UpdateQuotaPlan(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
	return m.storager.UpdateQuotaPlan(ctx, filter, param)
}

// DeleteQuotaPlan 删除配额计划
func (m *QuotaPlanManager) DeleteQuotaPlan(ctx context.Context, filter *QuotaPlanFilter) error {
	return m.storager.DeleteQuotaPlan(ctx, filter)
}

// ResetBalance 重置配额余额
// updateLastResetAt: 是否更新 last_reset_at 字段
// - true: 用于定期重置调度，会更新 last_reset_at，影响下次重置判断
// - false: 用于手动重置接口，不更新 last_reset_at，避免影响定期重置调度
func (m *QuotaPlanManager) ResetBalance(ctx context.Context, planID int64, newQuota *int64, updateLastResetAt bool) error {
	return m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		// 1. 获取 QuotaPlan
		plan, err := m.storager.FetchQuotaPlan(ctx, &QuotaPlanFilter{ID: &planID})
		if err != nil {
			return err
		}
		if plan == nil {
			return fmt.Errorf("quota_plan not found")
		}

		// 2. 如果是无限配额，返回错误
		if plan.Unlimited != nil && *plan.Unlimited {
			return fmt.Errorf("cannot reset balance for unlimited quota")
		}

		// 3. 确定重置后的配额总量
		resetQuota := plan.Quota
		if newQuota != nil {
			resetQuota = newQuota
		}

		// 4. 获取或创建 Balance
		balance, err := m.balanceStorager.FetchQuotaBalance(ctx, &QuotaBalanceFilter{QuotaPlanID: &planID})
		if err != nil {
			return err
		}

		// 5. 准备更新参数
		updateParam := &QuotaBalanceParam{
			Used:      lib.PInt64(0),
			Remaining: resetQuota,
		}

		// 6. 根据参数决定是否更新 last_reset_at
		if updateLastResetAt {
			now := time.Now()
			updateParam.LastResetAt = &now
		}

		// 7. 更新余额
		if balance == nil {
			// 创建新的余额记录
			_, err = m.balanceStorager.CreateQuotaBalance(ctx, &QuotaBalanceParam{
				QuotaPlanID: &planID,
				Used:        lib.PInt64(0),
				Remaining:   resetQuota,
				LastResetAt: updateParam.LastResetAt,
			})
		} else {
			// 更新现有余额记录
			_, err = m.balanceStorager.UpdateQuotaBalance(ctx, &QuotaBalanceFilter{ID: balance.ID}, updateParam)
		}

		return err
	})
}

// FetchQuotaBalance 获取配额余额
func (m *QuotaPlanManager) FetchQuotaBalance(ctx context.Context, planID int64) (*QuotaBalanceParam, error) {
	return m.balanceStorager.FetchQuotaBalance(ctx, &QuotaBalanceFilter{QuotaPlanID: &planID})
}
