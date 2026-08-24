# API-Key / Entity 配额清零与属性修改解耦设计变更说明

> 对应上游 Issue：[rainway-ai-gateway/ai-gateway-api#82](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/82)

## 1. 概述

### 1.1 问题现象

修改 API-Key / Entity 的任意属性（如 `description`、`models`、`route_rules` 等）时，只要请求体包含 `quota_plan`，控制面就会执行以下动作：

1. `model/api_key/api_key.go` 的 `UpdateAPIKey` 在事务成功后无条件调用 `quotaCache.SetRemaining`；
2. `model/entity/entity_manager.go` 的 `UpdateEntity` 在事务成功后同样无条件调用 `quotaCache.SetRemaining`；
3. 各 Endpoint（`update.go` / `full_update.go`）在更新后无条件调用 `QuotaPlanManager.ResetBalance`。

`ResetBalance` 会：

- 将 `quota_balances.used` 置 0；
- 将 `quota_balances.remaining` 置为 `quota_plan.quota`；
- 遍历所有关联 API-Key / Entity，将 Redis 剩余量重置为 `quota_plan.quota`。

结果是：即使只是修改了描述字段，配额使用量也会被清零。

### 1.2 变更目标

1. 普通属性修改不再影响配额余额。
2. 仅当 `quota_plan.quota` 或 `quota_plan.unit` 发生变化时才调整余额。
3. `quota` 变化时保留历史使用量：
   - 控制面：`新剩余额 = max(0, 新总额 - 原使用额)`；
   - Redis：通过原子 `IncrBy(delta)` 调整剩余量。
4. `unit` 变化时按全新配额重置（旧单位下剩余量语义不兼容）。
5. 无限制配额（`unlimited`）与有限额配额互转时，按新状态初始化余额。
6. 显式重置接口与定期重置调度保持清零语义不变。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/api_key`、`model/entity`、`model/quota`、`endpoints/openapi_v1/api_key`、`endpoints/openapi_v1/entity` |
| 接口契约 | `/api-keys/{id}`（PATCH/PUT）、`/entities/{id}`（PATCH/PUT）请求/响应字段不变 |
| 数据迁移 | 无 |

---

## 2. 现状代码分析

### 2.1 API-Key 更新路径

```
endpoints/openapi_v1/api_key/update.go
  └─ APIKeyUpdateProcess
       ├─ container.APIKeyManager.UpdateAPIKey(...)        // 写 DB
       ├─ fetch updated API-Key
       └─ if quotaPlanChanged: QuotaPlanManager.ResetBalance(...)  // 清零余额 + 重置 Redis

model/api_key/api_key.go
  └─ UpdateAPIKey
       ├─ 事务内更新 quota_plan / api_key 记录
       └─ 事务外 quotaCache.SetRemaining(...)             // 无条件再写一次 Redis
```

`full_update.go` 的 `APIKeyFullUpdateProcess` 逻辑相同。

### 2.2 Entity 更新路径

```
endpoints/openapi_v1/entity/update.go
  └─ EntityUpdateAction
       ├─ container.EntityManager.UpdateEntity(...)        // 写 DB
       ├─ fetch updated Entity
       └─ if QuotaPlan != nil: QuotaPlanManager.ResetBalance(...)

model/entity/entity_manager.go
  └─ UpdateEntity
       ├─ 事务内更新 quota_plan / entity 记录
       └─ 事务外 quotaCache.SetRemaining(...)             // 无条件再写一次 Redis
```

`full_update.go` 的 `EntityFullUpdateAction` 逻辑相同。

### 2.3 `ResetBalance` 语义

`model/quota/quota_plan_manager.go` 中的 `ResetBalance`：

- 可选更新 `quota_plans.quota`；
- 将 `quota_balances.used` 置 0，`remaining` 置为 `resetQuota`；
- 遍历所有关联 API-Key / Entity，调用 `quotaCache.ResetToQuota` 重置 Redis。

该语义适合“显式重置”和“周期重置”，但不适合“仅总额变化”场景。

---

## 3. 详细设计

### 3.1 核心原则

- **Model 层只负责持久化属性**，不再在事务外无条件同步 Redis。
- **Endpoint 层负责判断是否需要调整配额**，并调用专门的 QuotaPlanManager 方法。
- **QuotaPlanManager 提供两种语义**：
  - `ResetBalance`：清零式重置（保留给显式重置接口和调度器）。
  - `AdjustQuota`：差额/换单位调整（给更新接口使用）。

### 3.2 新增 `QuotaPlanManager.AdjustQuota`

在 `model/quota/quota_plan_manager.go` 新增方法：

```go
// AdjustQuota 在 quota_plan 的 quota 或 unit 发生变化时调整余额。
// 与 ResetBalance 不同：
//   - quota 变化时保留 used，仅调整 remaining；
//   - unit 变化时按全新 quota 重置（used 清零）。
func (m *QuotaPlanManager) AdjustQuota(
    ctx context.Context,
    planID int64,
    oldPlan, newPlan *QuotaPlanParam,
) error
```

内部逻辑：

```
1. 校验 planID 存在。
2. 计算 oldQuota / newQuota / oldUnit / newUnit（注意 nil 与 unlimited 处理）。
3. 事务内：
   a. 更新 quota_plans 表（quota/unit/unlimited 等）。
   b. 获取或创建 quota_balances 记录。
   c. 分场景计算新的 used/remaining：
      - 总额变化、单位不变：
          used = oldUsed（保留）
          remaining = max(0, newQuota - oldUsed)
      - 单位变化：
          used = 0
          remaining = newQuota
      - unlimited -> limited：
          used = 0
          remaining = newQuota
      - limited -> unlimited：
          used = 0
          remaining = 一个足够大的 sentinel（例如 1e8），与 EntityManager 当前行为一致
   d. 更新 quota_balances 表。
4. 事务外：
   - 单位不变且仍为有限额：
       对每个 owner 调用 quotaCache.SetRemaining(ctx, ownerKey, &newRemaining, newUnit)
       SetRemaining 内部使用 IncrBy(delta) 原子调整。
   - 单位变化或 unlimited 切换：
       对每个 owner 调用 quotaCache.ResetToQuota(ctx, ownerKey, &newRemaining, newUnit)
       直接覆盖 Redis 值。
```

> **为什么 `unit` 变化要清零？** `total_tokens` 与 `RMB` 是两种不同计量维度，旧单位下的“剩余 100 tokens”无法换算为新单位下的“剩余多少 RMB”。为避免计费歧义，按新单位全新重置。

### 3.3 移除 Model 层无条件 Redis 同步

#### `model/api_key/api_key.go` 的 `UpdateAPIKey`

删除事务结束后的这段代码：

```go
// 删除以下代码块
if updatedKey != nil && updatedQuotaPlan != nil &&
    (updatedQuotaPlan.Unlimited == nil || !*updatedQuotaPlan.Unlimited) &&
    updatedQuotaPlan.Quota != nil && rppm.quotaCache != nil {
    if cacheErr := rppm.quotaCache.SetRemaining(ctx, *updatedKey, updatedQuotaPlan.Quota, updatedQuotaPlan.Unit); ...
}
```

> 保留 `updatedKey` 与 `updatedQuotaPlan` 变量的收集逻辑可以一并清理，但 minimal change 只需删除事务外同步块。

#### `model/entity/entity_manager.go` 的 `UpdateEntity`

删除事务结束后的这段代码：

```go
// 删除以下代码块
if updatedEntityID != nil && updatedQuotaPlan != nil && m.quotaCache != nil {
    if updatedQuotaPlan.Unlimited != nil && *updatedQuotaPlan.Unlimited {
        defaultQuota := float64(100000000)
        if cacheErr := m.quotaCache.SetRemaining(ctx, *updatedEntityID, &defaultQuota, updatedQuotaPlan.Unit); ...
    } else if updatedQuotaPlan.Quota != nil {
        if cacheErr := m.quotaCache.SetRemaining(ctx, *updatedEntityID, updatedQuotaPlan.Quota, updatedQuotaPlan.Unit); ...
    }
}
```

### 3.4 `QuotaPlanManager.ApplyQuotaPlanChange`

将“是否需要调整配额”的判断逻辑下沉到 `model/quota`，Endpoint 只需传入新旧 `QuotaPlan`。

在 `model/quota/quota_plan_manager.go` 新增方法：

```go
// ApplyQuotaPlanChange 比较新旧配额计划，仅在 quota / unit / unlimited 发生变化时调整余额。
func (m *QuotaPlanManager) ApplyQuotaPlanChange(
    ctx context.Context,
    planID int64,
    oldPlan, newPlan *QuotaPlanParam,
) error
```

内部逻辑：

```
1. 如果 newPlan == nil，直接返回 nil（不修改配额计划）。
2. 比较 oldPlan 与 newPlan 的 quota / unit / unlimited：
   - 三者均未变化：直接返回 nil，不调整余额。
   - 任一发生变化：调用内部 AdjustQuota 逻辑执行差额/换单位调整。
3. 兼容 nil 与默认值的比较（例如 PUT 请求中未传 unlimited 视为 false）。
```

> 如果 `unlimited` 从 `false` 变为 `true` 或反向，也视为需要调整配额。

### 3.5 Endpoint 层调用调整

#### API-Key `update.go`

在调用 `UpdateAPIKey` 之后、返回之前：

```go
if param.QuotaPlan != nil && updated != nil && updated.QuotaPlanID != nil {
    if err := container.QuotaPlanManager.ApplyQuotaPlanChange(
        ctx, *updated.QuotaPlanID, existing.QuotaPlan, param.QuotaPlan,
    ); err != nil {
        return nil, err
    }
}
```

#### API-Key `full_update.go`

与 `update.go` 相同逻辑。PUT 语义下 `param.QuotaPlan` 为完整对象，`ApplyQuotaPlanChange` 内部负责将 nil 字段与旧值比较。

#### Entity `update.go` / `full_update.go`

与 API-Key 相同逻辑：

```go
if param.QuotaPlan != nil && updated != nil && updated.QuotaPlanID != nil {
    if err := container.QuotaPlanManager.ApplyQuotaPlanChange(
        ctx, *updated.QuotaPlanID, existing.QuotaPlan, param.QuotaPlan,
    ); err != nil {
        return nil, err
    }
}
```

### 3.6 显式重置与调度器不变

以下逻辑保持调用 `ResetBalance`：

- `endpoints/openapi_v1/api_key/reset_quota.go` 的 `ResetQuotaAction`
- `endpoints/openapi_v1/entity/reset_quota.go` 的 `EntityResetQuotaAction`
- `model/quota/balance_sync.go` 的 `ResetExpiredBalances`

它们的语义就是“清零式重置”，不应改为 `AdjustQuota`。

---

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `model/quota/quota_plan_manager.go` | 新增 `ApplyQuotaPlanChange`（含 diff + 调整）与内部 `AdjustQuota`；`ResetBalance` 保持不变。 |
| `model/api_key/api_key.go` | 删除 `UpdateAPIKey` 事务结束后的无条件 `quotaCache.SetRemaining`。 |
| `model/entity/entity_manager.go` | 删除 `UpdateEntity` 事务结束后的无条件 `quotaCache.SetRemaining`。 |
| `endpoints/openapi_v1/api_key/update.go` | 更新后调用 `ApplyQuotaPlanChange(existing.QuotaPlan, param.QuotaPlan)`。 |
| `endpoints/openapi_v1/api_key/full_update.go` | 同上。 |
| `endpoints/openapi_v1/entity/update.go` | 同上。 |
| `endpoints/openapi_v1/entity/full_update.go` | 同上。 |
| `design-docs/api-define/OpenAPI接口定义/api-keys.md` | 更新 2.4/2.5 节关于 `quota_plan` 变更对 balance 影响的说明。 |
| `design-docs/api-define/OpenAPI接口定义/entities.md` | 更新 2.4/2.5 节关于 `quota_plan` 变更对 balance 影响的说明。 |
| `model/api_key/api_key_test.go` | 调整/新增测试：普通属性更新不重置配额；quota 变化保留 used；unit 变化清零。 |
| `model/entity/entity_manager_test.go` | 同上。 |
| `model/quota/quota_plan_manager_test.go` | 新增 `ApplyQuotaPlanChange` 单元测试。 |

---

## 5. 测试计划

### 5.1 单元测试

#### `model/quota` 层

1. **`AdjustQuota` 总额增加**
   - old: quota=100, used=30, remaining=70
   - new: quota=150
   - 期望: used=30, remaining=120；Redis owner 剩余量由 70 增加到 120。

2. **`AdjustQuota` 总额减少（低于已用量）**
   - old: quota=100, used=80, remaining=20
   - new: quota=50
   - 期望: used=80, remaining=0；Redis owner 剩余量置 0。

3. **`AdjustQuota` 单位变化**
   - old: unit=`total_tokens`, quota=100, used=30
   - new: unit=`RMB`, quota=10
   - 期望: used=0, remaining=10；Redis 值按新单位直接覆盖。

4. **`AdjustQuota` 普通字段变化（pass_when_no_enough_quota / reset_period）**
   - old: quota=100, used=30
   - new: pass_when_no_enough_quota 变化，quota/unit 不变
   - 期望: 不调整余额，Redis 不变。

5. **`AdjustQuota` unlimited 切换**
   - limited -> unlimited: Redis 置 sentinel，used=0，remaining=sentinel。
   - unlimited -> limited: Redis 置新 quota，used=0，remaining=newQuota。

#### `model/api_key` 与 `model/entity` 层

1. **更新 description 不重置配额**
   - 构造已有配额的 API-Key / Entity；
   - 仅更新 `description`；
   - 断言 `quotaCache.SetRemaining` / `ResetBalance` 未被调用。

2. **更新 quota 保留 used**
   - 构造已有配额的 API-Key / Entity，Redis 已有使用量；
   - 修改 `quota_plan.quota`；
   - 断言 `QuotaPlanManager.AdjustQuota` 被调用，`used` 不变。

3. **更新 unit 清零 used**
   - 修改 `quota_plan.unit`；
   - 断言 `used=0`，Redis 被重置。

### 5.2 集成测试

1. **API-Key 更新描述后配额使用量不变**
   - 先通过若干请求产生配额使用；
   - PATCH `/api-keys/{id}` 仅修改 `description`；
   - 调用 `/api-keys/{id}` 详情，`quota_plan.balance.used` 应保持原值。

2. **API-Key 更新 quota 后 remaining 正确**
   - 产生 30 使用量后，PATCH `quota_plan.quota` 从 100 改为 150；
   - 断言 `remaining` 为 120，`used` 仍为 30。

3. **Entity 更新 unit 后清零**
   - Entity 已有使用量；
   - PATCH `quota_plan.unit`；
   - 断言 `used=0`，`remaining` 为新 quota。

---

## 6. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| Redis 与 DB 余额短暂不一致 | `AdjustQuota` 在事务外同步 Redis，失败时只打日志 | 与现有 `ResetBalance` 行为一致；后续 `BalanceSyncManager` 会兜底同步 |
| 总额减少导致 Redis 剩余为负 | `SetRemaining` 内部计算 `delta = target - current`，若 current > target 会得到负增量 | Redis 值可表示负余额；控制面取 `max(0, newQuota - oldUsed)` 保证 DB `remaining` 非负 |
| unlimited 切换语义 | limited -> unlimited 时，数据面可能仍会继续扣减 sentinel 值 | 与现有 EntityManager 行为保持一致（sentinel = 1e8）；数据面应识别 unlimited 配置直接放行 |
| PUT full_update 的 nil 语义 | PUT 请求中未传 `quota_plan` 视为不修改；传了则视为完整对象 | diff 函数需将 nil quota/unit 与旧值比较，避免误判 |
| 并发修改 quota_plan | 两个管理员同时修改同一 API-Key 的 quota | 依赖 DB 事务与 Redis 原子 `IncrBy`；极端并发下以最后一次 AdjustQuota 结果为准 |

---

## 7. 后续可优化（本期不做）

- 在 `QuotaPlanManager.ApplyQuotaPlanChange` / `AdjustQuota` 返回后增加异步重试机制，提高 Redis 同步成功率。
- 考虑将 quota 调整逻辑进一步抽象为独立 `QuotaAdjustmentService`，与 `QuotaPlanManager` 解耦。
