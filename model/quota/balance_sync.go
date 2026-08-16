package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/itxn"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"
)

// BalanceSyncManager 配额余额同步管理器
type BalanceSyncManager struct {
	txn             itxn.TxnStorager
	apiKeyStorager  api_key.APIKeyStorager
	balanceStorager QuotaBalanceStorager
	planStorager    QuotaPlanStorager
	entityStorager  entity.EntityStorager
	quotaCache      quotacache.QuotaCache
}

// NewBalanceSyncManager 创建配额余额同步管理器
func NewBalanceSyncManager(
	txn itxn.TxnStorager,
	apiKeyStorager api_key.APIKeyStorager,
	balanceStorager QuotaBalanceStorager,
	planStorager QuotaPlanStorager,
	entityStorager entity.EntityStorager,
	quotaCache quotacache.QuotaCache,
) *BalanceSyncManager {
	return &BalanceSyncManager{
		txn:             txn,
		apiKeyStorager:  apiKeyStorager,
		balanceStorager: balanceStorager,
		planStorager:    planStorager,
		entityStorager:  entityStorager,
		quotaCache:      quotaCache,
	}
}

// SyncAllBalances 同步所有配额计划的余额
func (m *BalanceSyncManager) SyncAllBalances(ctx context.Context) error {
	return m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		// 1. 获取所有非无限配额计划
		plans, err := m.planStorager.FetchQuotaPlanList(ctx, &QuotaPlanFilter{
			Unlimited: lib.PBool(false),
		})
		if err != nil {
			return err
		}

		// 2. 遍历每个配额计划，同步其余额
		for _, plan := range plans {
			if err := m.syncPlanBalance(ctx, plan); err != nil {
				// 记录错误但继续处理其他计划
				fmt.Printf("Failed to sync balance for plan %d: %v\n", *plan.ID, err)
			}
		}

		return nil
	})
}

// syncPlanBalance 同步单个配额计划的余额
func (m *BalanceSyncManager) syncPlanBalance(ctx context.Context, plan *QuotaPlanParam) error {
	// 1. 获取关联到此配额计划的所有 API-Key
	apiKeys, err := m.apiKeyStorager.FetchAPIKeyList(ctx, &api_key.APIKeyFilter{
		QuotaPlanID: plan.ID,
	})
	if err != nil {
		return err
	}

	// 2. 获取关联到此配额计划的所有 Entity
	var entities []*entity.EntityParam
	if m.entityStorager != nil {
		entities, err = m.entityStorager.FetchEntityList(ctx, &entity.EntityFilter{
			QuotaPlanID: plan.ID,
		})
		if err != nil {
			return err
		}
	}

	// 3. 从 Redis 读取每个 API-Key 的剩余量（按单位转换回 float64）
	var totalRemaining float64 = 0
	if m.quotaCache != nil {
		for _, apiKey := range apiKeys {
			if apiKey.Key == nil || apiKey.KeyCreateAt == nil {
				continue
			}

			remaining, err := m.quotaCache.GetRemaining(ctx, *apiKey.Key, plan.Unit)
			if err != nil {
				continue
			}

			totalRemaining += remaining
		}

		// 4. 从 Redis 读取每个 Entity 的剩余量
		for _, entity := range entities {
			if entity.EntityID == nil {
				continue
			}

			remaining, err := m.quotaCache.GetRemaining(ctx, *entity.EntityID, plan.Unit)
			if err != nil {
				continue
			}

			totalRemaining += remaining
		}
	}

	// 5. 计算已使用量
	totalUsed := *plan.Quota - totalRemaining
	if totalUsed < 0 {
		totalUsed = 0
	}

	// 6. 更新数据库中的余额记录
	_, err = m.balanceStorager.UpdateQuotaBalance(ctx, &QuotaBalanceFilter{
		QuotaPlanID: plan.ID,
	}, &QuotaBalanceParam{
		Used:      lib.PFloat64(totalUsed),
		Remaining: lib.PFloat64(totalRemaining),
	})

	return err
}

// ResetExpiredBalances 检查并重置所有到期的配额计划
func (m *BalanceSyncManager) ResetExpiredBalances(ctx context.Context) error {
	return m.txn.AtomExecute(ctx, func(ctx context.Context) error {
		// 1. 获取所有需要重置的配额计划（非无限配额且设置了重置周期）
		plans, err := m.planStorager.FetchQuotaPlanList(ctx, &QuotaPlanFilter{
			ResetPeriod: []string{"weekly", "monthly"},
			Unlimited:   lib.PBool(false),
		})
		if err != nil {
			return err
		}

		now := time.Now()

		// 2. 遍历每个配额计划
		for _, plan := range plans {
			balance, err := m.balanceStorager.FetchQuotaBalance(ctx, &QuotaBalanceFilter{
				QuotaPlanID: plan.ID,
			})
			if err != nil {
				fmt.Printf("Failed to fetch balance for plan %d: %v\n", *plan.ID, err)
				continue
			}

			if balance == nil {
				continue
			}

			// 3. 判断是否需要重置（基于自然周/自然月）
			shouldReset := m.shouldResetByPeriod(balance.LastResetAt, *plan.ResetPeriod, now)

			if shouldReset {
				// 4. 重置该配额计划下所有 API-Key 的 Redis 使用量
				if err := m.resetAPIKeysRedisUsage(ctx, *plan.ID); err != nil {
					fmt.Printf("Failed to reset Redis usage for plan %d: %v\n", *plan.ID, err)
					continue
				}

				// 5. 重置数据库余额
				_, err = m.balanceStorager.UpdateQuotaBalance(ctx, &QuotaBalanceFilter{
					QuotaPlanID: plan.ID,
				}, &QuotaBalanceParam{
					Used:        lib.PFloat64(0),
					Remaining:   plan.Quota,
					LastResetAt: lib.PTime(now),
				})
				if err != nil {
					fmt.Printf("Failed to reset balance for plan %d: %v\n", *plan.ID, err)
				} else {
					fmt.Printf("Reset balance for plan %d at %v\n", *plan.ID, now)
				}
			}
		}

		return nil
	})
}

// shouldResetByPeriod 判断是否应该根据周期重置配额
// 采用自然周/自然月的边界判断
func (m *BalanceSyncManager) shouldResetByPeriod(lastResetAt *time.Time, resetPeriod string, now time.Time) bool {
	if lastResetAt == nil {
		return true // 从未重置过，需要重置
	}

	switch resetPeriod {
	case "weekly":
		// 自然周：每周一 00:00:00
		// 计算上次重置时间所在周的周一
		lastResetWeekStart := getWeekStart(*lastResetAt)
		// 计算当前时间所在周的周一
		currentWeekStart := getWeekStart(now)
		// 如果当前周的开始时间晚于上次重置周的开始时间，说明需要重置
		return currentWeekStart.After(lastResetWeekStart)

	case "monthly":
		// 自然月：每月1日 00:00:00
		// 计算上次重置时间所在月的1日
		lastResetMonthStart := getMonthStart(*lastResetAt)
		// 计算当前时间所在月的1日
		currentMonthStart := getMonthStart(now)
		// 如果当前月的开始时间晚于上次重置月的开始时间，说明需要重置
		return currentMonthStart.After(lastResetMonthStart)
	}

	return false
}

// getWeekStart 获取给定时间所在周的周一 00:00:00
func getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	// Go 的 Weekday() 返回 0=Sunday, 1=Monday, ..., 6=Saturday
	// 转换为 Monday=0, Tuesday=1, ..., Sunday=6
	if weekday == time.Sunday {
		weekday = 6
	} else {
		weekday -= 1
	}

	// 回退到周一
	weekStart := t.AddDate(0, 0, -int(weekday))
	// 设置为 00:00:00
	return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, t.Location())
}

// getMonthStart 获取给定时间所在月的1日 00:00:00
func getMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// resetAPIKeysRedisUsage 重置指定配额计划下所有 API-Key 和 Entity 的 Redis 使用量
// 将 Redis key 的值重置为配额总量
func (m *BalanceSyncManager) resetAPIKeysRedisUsage(ctx context.Context, planID int64) error {
	// 1. 获取配额计划信息
	plan, err := m.planStorager.FetchQuotaPlan(ctx, &QuotaPlanFilter{ID: &planID})
	if err != nil {
		return fmt.Errorf("fetch plan %d error: %v", planID, err)
	}
	if plan == nil || plan.Quota == nil {
		return fmt.Errorf("plan %d not found or quota is nil", planID)
	}

	// 2. 获取该配额计划下的所有 API-Key
	apiKeys, err := m.apiKeyStorager.FetchAPIKeyList(ctx, &api_key.APIKeyFilter{
		QuotaPlanID: &planID,
	})
	if err != nil {
		return fmt.Errorf("fetch api keys for plan %d error: %v", planID, err)
	}

	// 3. 获取该配额计划下的所有 Entity
	var entities []*entity.EntityParam
	if m.entityStorager != nil {
		entities, err = m.entityStorager.FetchEntityList(ctx, &entity.EntityFilter{
			QuotaPlanID: &planID,
		})
		if err != nil {
			return fmt.Errorf("fetch entities for plan %d error: %v", planID, err)
		}
	}

	if m.quotaCache == nil {
		return nil
	}

	// 4. 遍历每个 API-Key，重置其 Redis 值为配额总量
	for _, apiKey := range apiKeys {
		if apiKey.Key == nil || apiKey.KeyCreateAt == nil {
			continue
		}

		if err := m.quotaCache.ResetToQuota(ctx, *apiKey.Key, plan.Quota, plan.Unit); err != nil {
			fmt.Printf("Failed to reset Redis value for API-Key %s (key=%s): %v\n",
				*apiKey.ID, *apiKey.Key, err)
			continue
		}

		fmt.Printf("Reset Redis value for API-Key %s (key=%s) to quota %v\n",
			*apiKey.ID, *apiKey.Key, *plan.Quota)
	}

	// 5. 遍历每个 Entity，重置其 Redis 值为配额总量
	for _, entity := range entities {
		if entity.EntityID == nil {
			continue
		}

		if err := m.quotaCache.ResetToQuota(ctx, *entity.EntityID, plan.Quota, plan.Unit); err != nil {
			fmt.Printf("Failed to reset Redis value for Entity %s: %v\n",
				*entity.EntityID, err)
			continue
		}

		fmt.Printf("Reset Redis value for Entity %s to quota %v\n",
			*entity.EntityID, *plan.Quota)
	}

	return nil
}
