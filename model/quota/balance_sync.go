package quota

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/icluster_conf"
	"github.com/yf-networks/ai-gateway-api/model/itxn"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

// BalanceSyncManager 配额余额同步管理器
type BalanceSyncManager struct {
	txn             itxn.TxnStorager
	apiKeyStorager  icluster_conf.APIKeyStorager
	balanceStorager QuotaBalanceStorager
	planStorager    QuotaPlanStorager
}

// NewBalanceSyncManager 创建配额余额同步管理器
func NewBalanceSyncManager(
	txn itxn.TxnStorager,
	apiKeyStorager icluster_conf.APIKeyStorager,
	balanceStorager QuotaBalanceStorager,
	planStorager QuotaPlanStorager,
) *BalanceSyncManager {
	return &BalanceSyncManager{
		txn:             txn,
		apiKeyStorager:  apiKeyStorager,
		balanceStorager: balanceStorager,
		planStorager:    planStorager,
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
	apiKeys, err := m.apiKeyStorager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{
		QuotaPlanID: plan.ID,
	})
	if err != nil {
		return err
	}

	// 2. 从 Redis 读取每个 API-Key 的已使用量
	var totalUsed int64 = 0
	for _, apiKey := range apiKeys {
		if apiKey.Key == nil || apiKey.KeyCreateAt == nil {
			continue
		}

		// 使用与 GetRemainingQuota 相同的 Redis Key 格式
		redisKey := stateful.AIUsedQuotaKey(*apiKey.Key, apiKey.KeyCreateAt.Unix())
		used, err := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if err != nil {
			// 如果没有使用记录，使用量为 0
			continue
		}

		totalUsed += used
	}

	// 3. 计算剩余配额
	remaining := *plan.Quota - totalUsed
	if remaining < 0 {
		remaining = 0
	}

	// 4. 更新数据库中的余额记录
	_, err = m.balanceStorager.UpdateQuotaBalance(ctx, &QuotaBalanceFilter{
		QuotaPlanID: plan.ID,
	}, &QuotaBalanceParam{
		Used:      lib.PInt64(totalUsed),
		Remaining: lib.PInt64(remaining),
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
					Used:        lib.PInt64(0),
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

// resetAPIKeysRedisUsage 重置指定配额计划下所有 API-Key 的 Redis 使用量
// 直接更新现有 Redis key 的值为 0，不新建 key
func (m *BalanceSyncManager) resetAPIKeysRedisUsage(ctx context.Context, planID int64) error {
	// 1. 获取该配额计划下的所有 API-Key
	apiKeys, err := m.apiKeyStorager.FetchAPIKeyList(ctx, &icluster_conf.APIKeyFilter{
		QuotaPlanID: &planID,
	})
	if err != nil {
		return fmt.Errorf("fetch api keys for plan %d error: %v", planID, err)
	}

	// 2. 遍历每个 API-Key，重置其 Redis 使用量
	for _, apiKey := range apiKeys {
		if apiKey.Key == nil || apiKey.KeyCreateAt == nil {
			continue
		}

		// 使用与 GetRemainingQuota 相同的 Redis Key 格式
		redisKey := stateful.AIUsedQuotaKey(*apiKey.Key, apiKey.KeyCreateAt.Unix())

		// 获取当前使用量，然后通过 IncrBy 减去该值来重置为 0
		currentUsage, err := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
		if err != nil {
			// 如果 key 不存在，使用量已经是 0，跳过
			if strings.Contains(err.Error(), "redigo: nil returned") {
				continue
			}
			fmt.Printf("Failed to get Redis usage for API-Key %d (key=%s): %v\n",
				*apiKey.ID, *apiKey.Key, err)
			continue
		}
		if currentUsage > 0 {
			_, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, -currentUsage)
			if err != nil {
				fmt.Printf("Failed to reset Redis usage for API-Key %d (key=%s): %v\n",
					*apiKey.ID, *apiKey.Key, err)
				continue
			}
		}

		fmt.Printf("Reset Redis usage for API-Key %d (key=%s, redisKey=%s) to 0\n",
			*apiKey.ID, *apiKey.Key, redisKey)
	}

	return nil
}
