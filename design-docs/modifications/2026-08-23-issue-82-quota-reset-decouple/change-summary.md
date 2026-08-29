# API-Key / Entity 配额使用量随属性更新被清零问题修复（Issue #82）变更摘要

## 1. 背景

当前修改 API-Key 或 Entity 的任意属性（例如描述、模型列表、限流策略等）时，只要请求体中携带了 `quota_plan`，控制面就会无条件重置该配额计划的余额：

- `quota_balances.used` 被清零；
- Redis 中所有关联 API-Key / Entity 的剩余配额被设为 `quota_plan.quota`。

这导致用户的配额使用量被错误清零，影响计费与限流的准确性。

## 2. 目标

- 将“配额调整”与“API-Key / Entity 属性修改”解耦。
- 只有在真正影响配额计量的字段发生变化时才调整余额：
  - `quota`（配额总额）变化；
  - `unit`（配额单位，如 `total_tokens` ↔ `RMB`）变化。
- 修改配额总额时，保留已有使用量，按 `新总额 - 原使用额 = 新剩余额` 换算后同步到 Redis。
- 修改配额单位时，由于旧单位下的剩余量无法直接换算，按全新配额重置。
- 显式的“重置配额”接口与定期重置调度保持原有语义不变。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要模块 | `model/api_key`、`model/entity`、`model/quota` |
| 涉及接口 | `PATCH /api-keys/{id}`、`PUT /api-keys/{id}`、`PATCH /entities/{id}`、`PUT /entities/{id}` |
| 接口契约 | 请求/响应字段不变，仅后台余额处理逻辑变更 |
| 数据迁移 | 无 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| diff 逻辑下沉到 `model/quota` | 新增 `QuotaPlanManager.ApplyQuotaPlanChange`，由它判断新旧 `QuotaPlan` 是否变化，并决定是否调用内部调整逻辑；4 个 Endpoint 共用，避免重复代码。 |
| 新增 `QuotaPlanManager.AdjustQuota` | 专门处理“总额/单位变化”场景，与既有的 `ResetBalance`（清零式重置）语义分离。 |
| 保留 `ResetBalance` | 显式重置接口（`/api-keys/{id}/reset-quota`、`/entities/{id}/reset-quota`）与定时重置调度继续调用 `ResetBalance`，保持清零语义。 |
| 移除 Model 层事务外无条件 `SetRemaining` | `APIKeyManager.UpdateAPIKey` 与 `EntityManager.UpdateEntity` 中，事务成功后不再无条件写 Redis；Redis 同步由 `ApplyQuotaPlanChange` 按需触发。 |
| 总额变化使用“delta 增量”同步 Redis | 通过 `quotaCache.SetRemaining` 内部已有的 `IncrBy(delta)` 机制原子调整，避免并发消费窗口下的 race。 |

## 5. 关联文档

- 详细设计：`design-changes.md`
- 上游 Issue：[rainway-ai-gateway/ai-gateway-api#82](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/82)
