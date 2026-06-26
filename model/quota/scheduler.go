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
	"time"

	"github.com/yf-networks/ai-gateway-api/model/itxn"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

// QuotaResetScheduler 配额重置调度器
type QuotaResetScheduler struct {
	txn            itxn.TxnStorager
	balanceSyncMgr *BalanceSyncManager
	stopCh         chan struct{}
}

// NewQuotaResetScheduler 创建配额重置调度器
func NewQuotaResetScheduler(
	txn itxn.TxnStorager,
	balanceSyncMgr *BalanceSyncManager,
) *QuotaResetScheduler {
	return &QuotaResetScheduler{
		txn:            txn,
		balanceSyncMgr: balanceSyncMgr,
		stopCh:         make(chan struct{}),
	}
}

// Start 启动定时任务
func (s *QuotaResetScheduler) Start() {
	go s.run()
}

// Stop 停止定时任务
func (s *QuotaResetScheduler) Stop() {
	close(s.stopCh)
}

// run 运行定时任务循环
func (s *QuotaResetScheduler) run() {
	// 每小时检查一次需要重置的配额
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.resetQuotas()
		case <-s.stopCh:
			return
		}
	}
}

// resetQuotas 重置配额
func (s *QuotaResetScheduler) resetQuotas() {
	ctx := context.Background()

	// 使用 BalanceSyncManager 的重置方法
	if err := s.balanceSyncMgr.ResetExpiredBalances(ctx); err != nil {
		stateful.AccessLogger.Error("Failed to reset expired balances: %v", err)
	}
}
