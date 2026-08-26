// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package quota

import (
	"context"
	"runtime"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
)

// BalanceSyncer defines the operations required by QuotaResetScheduler.
type BalanceSyncer interface {
	SyncAllBalances(ctx context.Context) error
	ResetExpiredBalances(ctx context.Context) error
}

// QuotaResetScheduler 配额重置调度器
type QuotaResetScheduler struct {
	txn            itxn.TxnStorager
	balanceSyncMgr BalanceSyncer
	stopCh         chan struct{}
}

// NewQuotaResetScheduler 创建配额重置调度器
func NewQuotaResetScheduler(
	txn itxn.TxnStorager,
	balanceSyncMgr BalanceSyncer,
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
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			stateful.ExceptionLogger.Error("PANIC in QuotaResetScheduler: err=%v\n%s", err, string(stack))
		}
	}()

	s.resetQuotasWithRecover()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.resetQuotasWithRecover()
		case <-s.stopCh:
			return
		}
	}
}

func (s *QuotaResetScheduler) resetQuotasWithRecover() {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			stateful.ExceptionLogger.Error("PANIC in resetQuotas: err=%v\n%s", err, string(stack))
		}
	}()

	s.resetQuotas()
}

// resetQuotas 重置配额
func (s *QuotaResetScheduler) resetQuotas() {
	ctx := context.Background()
	now := time.Now()

	stateful.AccessLogger.Info("Starting quota scheduler tasks at %v", now)

	// 1. 从 Redis 更新每个 quota_plan 对应的 quota_balance 的配额余量
	if err := s.balanceSyncMgr.SyncAllBalances(ctx); err != nil {
		stateful.AccessLogger.Error("Failed to sync all balances: %v", err)
	} else {
		stateful.AccessLogger.Info("Successfully synced all balances from Redis")
	}

	// 2. 查找所有 unlimited=0 的 quota_plan 中是否达到重置的时间条件，达到后执行向 Redis 重置配额余额
	if err := s.balanceSyncMgr.ResetExpiredBalances(ctx); err != nil {
		stateful.AccessLogger.Error("Failed to reset expired balances: %v", err)
	} else {
		stateful.AccessLogger.Info("Successfully checked and reset expired balances")
	}

	stateful.AccessLogger.Info("Quota scheduler tasks completed at %v", time.Now())
}
