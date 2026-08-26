# QuotaPlan `reset_period` 回归测试能力增强设计变更说明

## 1. 概述

### 1.1 问题现象

`00-common.md` 中 `QuotaPlan.reset_period` 支持三种周期：

- `never`：不自动重置；
- `weekly`：自然周，每周一 00:00:00 重置；
- `monthly`：自然月，每月 1 日 00:00:00 重置。

当前实现中，`BalanceSyncManager.ResetExpiredBalances` 直接使用 `time.Now()` 判断当前时间是否跨越了周期边界：

```go
// model/quota/balance_sync.go
func (m *BalanceSyncManager) ResetExpiredBalances(ctx context.Context) error {
    // ...
    now := time.Now()
    for _, plan := range plans {
        // ...
        shouldReset := m.shouldResetByPeriod(balance.LastResetAt, *plan.ResetPeriod, now)
        // ...
    }
}
```

这导致集成测试和回归测试面临以下问题：

1. 真实周期太长，等待一周或一个月才能验证显然不可行；
2. 无法在不修改系统时间的情况下构造“上周一刚重置过，这周一又该重置了”的场景；
3. `shouldResetByPeriod` 虽有单元测试，但仅能验证“给定两个时间，判断是否正确”，无法验证调度器在真实运行到周期边界时是否会触发重置。

### 1.2 变更目标

1. 为 `model/quota` 引入可注入的时钟抽象；
2. 测试代码能够任意设定当前时间，快速验证 `weekly`/`monthly` 跨周期自动重置；
3. 保持生产环境行为不变；
4. 为后续可能扩展更多时间粒度（如 `daily`、`hourly`）预留可测试基础。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/quota`、`stateful/container/rdb` |
| 接口契约 | 无变化 |
| 数据迁移 | 无 |

---

## 2. 现状代码分析

### 2.1 调度器流程

```text
model/quota/scheduler.go
  └─ QuotaResetScheduler.run()
       ├─ 每分钟 ticker 触发
       └─ resetQuotas()
            ├─ balanceSyncMgr.SyncAllBalances()
            └─ balanceSyncMgr.ResetExpiredBalances()  ← 使用 time.Now()
```

### 2.2 周期判断流程

```text
model/quota/balance_sync.go
  └─ BalanceSyncManager.ResetExpiredBalances()
       ├─ 查询 reset_period 为 weekly/monthly 且非 unlimited 的 quota_plan
       ├─ 获取对应 quota_balances.last_reset_at
       ├─ shouldResetByPeriod(lastResetAt, resetPeriod, time.Now())
       │    ├─ weekly: 比较当前周周一 vs 上次重置周周一
       │    └─ monthly: 比较当前月 1 日 vs 上次重置月 1 日
       └─ 若需要重置：重置 Redis + 更新 quota_balances
```

### 2.3 已有测试

- `model/quota/balance_sync_test.go:31`：`TestBalanceSyncManager_shouldResetByPeriod` 已对纯判断函数做单元测试；
- `model/quota/balance_sync_test.go:237`：`ResetExpiredBalances skips when not expired` 已覆盖未到期跳过场景；
- 缺少“跨周期时确实触发重置”的单元/集成测试。

---

## 3. 详细设计

### 3.1 核心原则

- **不修改业务语义**：只把 `time.Now()` 替换为可注入的 `Clock.Now()`。
- **默认行为不变**：生产环境使用真实时钟 `realClock`。
- **测试可控**：测试环境使用 `fakeClock`，可任意设定时间。

### 3.2 新增 Clock 抽象

在 `model/quota/clock.go` 中定义：

```go
package quota

import "time"

// Clock 提供当前时间，便于测试注入固定或可推进的时间。
type Clock interface {
    Now() time.Time
}

// realClock 使用系统真实时间。
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// NewRealClock 返回基于系统真实时间的 Clock。
func NewRealClock() Clock { return realClock{} }
```

### 3.3 给 BalanceSyncManager 注入 Clock

修改 `model/quota/balance_sync.go`：

```go
type BalanceSyncManager struct {
    txn             itxn.TxnStorager
    apiKeyStorager  api_key.APIKeyStorager
    balanceStorager QuotaBalanceStorager
    planStorager    QuotaPlanStorager
    entityStorager  entity.EntityStorager
    quotaCache      quotacache.QuotaCache
    clock           Clock   // 新增
}
```

构造器增加 `Clock` 参数，并为 `nil` 提供默认真实时钟，保证向后兼容：

```go
func NewBalanceSyncManager(
    txn itxn.TxnStorager,
    apiKeyStorager api_key.APIKeyStorager,
    balanceStorager QuotaBalanceStorager,
    planStorager QuotaPlanStorager,
    entityStorager entity.EntityStorager,
    quotaCache quotacache.QuotaCache,
    clock Clock,
) *BalanceSyncManager {
    if clock == nil {
        clock = NewRealClock()
    }
    return &BalanceSyncManager{
        txn:             txn,
        apiKeyStorager:  apiKeyStorager,
        balanceStorager: balanceStorager,
        planStorager:    planStorager,
        entityStorager:  entityStorager,
        quotaCache:      quotaCache,
        clock:           clock,
    }
}
```

`ResetExpiredBalances` 中：

```go
now := m.clock.Now()
```

### 3.4 生产环境初始化

在 `stateful/container/rdb/components.go` 中：

```go
container.BalanceSyncManager = quota.NewBalanceSyncManager(
    container.TxnStoragerSingleton,
    container.APIKeyStorager,
    container.QuotaBalanceStorager,
    container.QuotaPlanStorager,
    container.EntityStorager,
    container.QuotaCacheSingleton,
    quota.NewRealClock(),   // 新增
)
```

### 3.5 测试用 fakeClock

在 `model/quota` 测试文件中新增：

```go
type fakeClock struct {
    t time.Time
}

func (f *fakeClock) Now() time.Time { return f.t }
```

### 3.6 单元测试设计

#### 跨周重置

```go
func TestResetExpiredBalances_WeeklyResetsOnNextMonday(t *testing.T) {
    lastReset := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC) // 周一
    clock := &fakeClock{t: time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)} // 下周一

    // 构造 plan: reset_period=weekly, unlimited=false, quota=1000
    // 构造 balance: last_reset_at=lastReset, used=500, remaining=500
    // 构造 api_key/entity 关联到该 plan
    // 构造 mockRedis，其中 api_key 剩余 500，entity 剩余 0

    m := NewBalanceSyncManager(
        &fakeTxn{},
        apiKeyStorager,
        balanceStorager,
        planStorager,
        entityStorager,
        quotacache.NewRedisQuotaCache(mockRedis),
        clock,
    )

    err := m.ResetExpiredBalances(context.Background())
    require.NoError(t, err)

    // 断言 balance 被更新：used=0, remaining=1000, last_reset_at=clock.Now()
    // 断言 Redis 中 api_key/entity 的剩余量都被重置为 1000
}
```

#### 同周内不重置

```go
func TestResetExpiredBalances_WeeklySameWeekNoReset(t *testing.T) {
    lastReset := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC) // 周一
    clock := &fakeClock{t: time.Date(2026, 7, 29, 9, 0, 0, 0, time.UTC)} // 周三

    // ...

    err := m.ResetExpiredBalances(context.Background())
    require.NoError(t, err)

    // 断言 balance 未被更新
}
```

#### 跨月重置

与跨周类似，把 `clock` 调到下个月 1 日即可。

#### 跨年度重置

验证 `lastResetAt` 为去年 12 月，`clock` 为今年 1 月 1 日时，`monthly` 能正确触发。

#### never 不重置

`reset_period=never` 的 plan 不应出现在 `ResetExpiredBalances` 的查询条件中，已存在逻辑保证，可补充断言确认。

### 3.7 调度器接口抽象

为了让调度器可测试，`QuotaResetScheduler` 不再依赖具体的 `*BalanceSyncManager`，而是依赖一个 `BalanceSyncer` 接口：

```go
type BalanceSyncer interface {
    SyncAllBalances(ctx context.Context) error
    ResetExpiredBalances(ctx context.Context) error
}
```

`QuotaResetScheduler` 的字段和构造器相应改为：

```go
type QuotaResetScheduler struct {
    txn            itxn.TxnStorager
    balanceSyncMgr BalanceSyncer
    stopCh         chan struct{}
}

func NewQuotaResetScheduler(
    txn itxn.TxnStorager,
    balanceSyncMgr BalanceSyncer,
) *QuotaResetScheduler {
    // ...
}
```

`*BalanceSyncManager` 自然实现该接口，因此生产初始化代码无需修改。测试中可传入 spy 实现，直接调用 `s.resetQuotas()` 验证 `SyncAllBalances` 和 `ResetExpiredBalances` 均被调用，无需等待 1 分钟 ticker。

---

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `model/quota/clock.go` | 新增 `Clock` 接口、`realClock`、`NewRealClock`。 |
| `model/quota/balance_sync.go` | `BalanceSyncManager` 新增 `clock` 字段；构造器注入；`ResetExpiredBalances` 使用 `m.clock.Now()`。 |
| `stateful/container/rdb/components.go` | 初始化 `BalanceSyncManager` 时传入 `quota.NewRealClock()`。 |
| `model/quota/balance_sync_test.go` | 新增 `fakeClock`；补充跨周/跨月/同周期/跨年/never 等测试。 |
| `model/quota/scheduler.go` | `QuotaResetScheduler` 改为依赖 `BalanceSyncer` 接口，便于测试注入 spy。 |
| `model/quota/scheduler_test.go` | 新增，使用 spy 验证 `resetQuotas()` 调用链。 |

---

## 5. 测试计划

### 5.1 单元测试

| 用例 | 场景 | 期望 |
|------|------|------|
| `weekly` 跨周 | `last_reset_at` 上周一，`clock` 下周一 | `used=0`，`remaining=quota`，Redis 重置为 quota |
| `weekly` 同周 | `last_reset_at` 周一，`clock` 周三 | 不更新 balance，不重置 Redis |
| `monthly` 跨月 | `last_reset_at` 上月末，`clock` 本月 1 日 | `used=0`，`remaining=quota`，Redis 重置为 quota |
| `monthly` 同月 | `last_reset_at` 月初，`clock` 月中 | 不更新 balance，不重置 Redis |
| 跨年 | `last_reset_at` 去年 12 月，`clock` 今年 1 月 | `monthly` 正常触发重置 |
| `never` | `reset_period=never` | 不参与 `ResetExpiredBalances` 查询 |
| nil `last_reset_at` | 新创建 plan 无 balance | 首次应触发重置 |

### 5.2 集成测试（可选）

若需覆盖完整 HTTP 链路，可在集成测试中：

1. 创建 API-Key/Entity，设置 `reset_period: weekly`；
2. 直接修改测试数据库 `quota_balances.last_reset_at` 到上周；
3. 重启测试进程时传入测试专用时钟（需把 clock 配置化到进程启动参数或环境变量）；
4. 触发一次调度后查询详情，断言 `balance.used = 0`。

更轻量的做法是在单元测试中覆盖 `BalanceSyncManager` 层即可，因为 OpenAPI 端点本身不涉及周期判断逻辑。

---

## 6. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 生产行为变化 | 若 `clock` 注入后默认实现有误，可能影响周期重置 | 默认使用 `NewRealClock()`，与现有 `time.Now()` 行为一致；保留现有测试 |
| 并发测试污染 | `fakeClock` 是单个实例，多个测试共享可能互相影响 | 每个 `t.Run` 独立构造 `fakeClock` 和 `BalanceSyncManager` |
| 时间精度与时区 | `shouldResetByPeriod` 依赖本地时区 | `fakeClock` 使用 `time.UTC` 构造固定时间，避免本地时区差异 |
| 对 `QuotaResetScheduler` 的 ticker 测试耗时 | 真实 ticker 为 1 分钟 | 测试直接调用 `resetQuotas()`，不依赖 ticker |

---

## 7. 后续可优化（本期不做）

- 把 `QuotaResetScheduler` 的 ticker 间隔也做成可配置，便于集成测试缩短调度周期。
- 考虑把 `Clock` 抽象推广到 `rate_limit`、`api_key` 等其他需要时间判断的模块。
- 若后续新增 `daily`、`hourly` 等更细粒度，`shouldResetByPeriod` 可扩展为通用的时间窗口判断函数。
