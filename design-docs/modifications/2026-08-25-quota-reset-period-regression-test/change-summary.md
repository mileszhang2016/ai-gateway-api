# QuotaPlan `reset_period` 回归测试能力增强

## 变更概述

为 `QuotaPlan.reset_period`（`never`/`weekly`/`monthly`）的周期性自动重置逻辑引入可测试性改造：在 `model/quota` 层注入 `Clock` 接口，使测试代码能够控制当前时间，从而在不等待真实自然周/自然月的情况下验证周期重置行为。

## 变更原因

- `reset_period` 的真实周期太长（一周或一个月），无法通过常规集成测试覆盖。
- 当前 `BalanceSyncManager.ResetExpiredBalances` 直接调用 `time.Now()`，测试无法模拟跨周/跨月场景。
- 需要为后续可能新增的更多时间粒度（如 `daily`、`hourly`）预留可测试的基础能力。

## 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/quota`、`stateful/container` |
| 接口契约 | 无变化，OpenAPI 字段定义保持不变 |
| 数据迁移 | 无 |

## 关键改动

1. 新增 `model/quota/clock.go`，定义 `Clock` 接口及真实实现 `realClock`。
2. `BalanceSyncManager` 增加 `clock Clock` 字段，构造器注入，默认使用真实时钟。
3. `ResetExpiredBalances` 中 `time.Now()` 替换为 `m.clock.Now()`。
4. `stateful/container/rdb/components.go` 初始化时传入 `quota.NewRealClock()`。
5. `QuotaResetScheduler` 改为依赖 `BalanceSyncer` 接口，便于注入 spy 测试调度调用链。
6. 补充 `model/quota` 单元测试，覆盖跨周、跨月、同周期内不重置、跨年、nil last_reset_at 等场景。

## 不涉及的改动

- 不新增 `minutely` 等测试专用 `reset_period` 值，避免污染生产配置。
- 不修改 OpenAPI 接口定义。
- 不修改自然周/自然月的重置语义。
