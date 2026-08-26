# quota-plan 余额同步机制移除变更摘要

## 背景与目标

当前 `ai-gateway-api` 对 `quota_plan` 余额采用 **"Redis 实时计数 + 数据库定时同步"** 的混合架构：

- Redis 由 BFE 数据面实时扣减，是热数据。
- `quota_balances` 表由 `BalanceSyncManager.SyncAllBalances` 每分钟从 Redis 汇总回写，用于管理面查询。

该机制存在以下问题：

1. **数据滞后**：`quota_balances` 最多滞后 1 分钟，管理面看到的余额不是最新的。
2. **读性能差**：OpenAPI 从数据库读取余额，性能不如直接读 Redis。
3. **写开销大**：同步复杂度随 API-Key / Entity 数量线性增长，达到数千数万时 Redis + DB 压力很大。
4. **资源浪费**：每分钟全量同步，但管理面查询余额的频率通常并不高。

本次重构目标：

1. OpenAPI 查询余额直接读取 Redis。
2. 删除 `BalanceSyncManager.SyncAllBalances` 定时同步。
3. 把 `last_reset_at` 从 `quota_balances` 迁移到 `quota_plans` 表。
4. 彻底删除 `quota_balances` 表及相关代码。

## 主要改动点

### 1. OpenAPI 余额查询改为读取 Redis

- `model/api_key/api_key.go` 的 `populateAssociatedData` / `FetchAPIKeyList` 优先从 Redis 读取实时余额。
- `model/entity/entity_manager.go` 的 `populateAssociatedData` / `FetchEntityList` 优先从 Redis 读取实时余额。
- 列表查询使用 `QuotaCache.BatchGetRemaining`，内部通过扩展后的 `redis_client.Client.GetInt64Batch` 批量读取（按 Redis 集群分组执行 MGET）。
- Redis 读取失败时返回错误，不降级到已废弃的 `quota_balances`。
- `endpoints/openapi_v1/api_key/get_quota_plan.go` 与 `endpoints/openapi_v1/entity/get_quota_plan.go` 的 `Balance` 字段增加 `omitempty`，无余额时不再返回 `null`；无限配额仍返回 sentinel balance（`used=0`, `remaining=100000000`）。

### 2. 扩展 `redis_client.Client` 接口支持批量读取

- 在 `bfe/bfe_util/redis_client/client.go` 新增 `GetInt64Batch(keys []string) ([]int64, error)`。
- 在 `bfe/bfe_util/redis_client/redis_bns.go` 的 `RedisClient` 中实现：按 Redis 集群 ID 分组，对每个分组执行 `MGET`。
- 在 `ai-gateway-api/stateful/mock_redis.go` 的 `MockRedisClient` 中实现批量读取。

### 3. 扩展 `QuotaCache` 接口

- `model/quotacache/quotacache.go` 新增 `BatchGetRemaining(ctx, keys, unit)`。
- `model/quotacache/redis.go` 的 `RedisQuotaCache` 实现批量读取及单位转换。

### 4. 删除 `SyncAllBalances`

- 删除 `model/quota/balance_sync.go` 中的 `SyncAllBalances` 和 `syncPlanBalance`。
- 删除 `model/quota/scheduler.go` 中对 `SyncAllBalances` 的调用。
- 删除 `model/quota/balance_sync_test.go` 中相关测试。

### 5. 迁移 `last_reset_at` 到 `quota_plans`

- `db_ddl.sql` / `db_ddl_sqlite.sql`：在 `quota_plans` 表新增 `last_reset_at` 字段。
- `storage/rdb/internal/dao/table_quota_plans.go`：`TQuotaPlan` / `TQuotaPlanParam` 新增 `LastResetAt`。
- `storage/rdb/quota/quota_plan.go`：转换层处理 `LastResetAt`。
- `model/quota/quota_plan.go`：`QuotaPlanParam` 新增 `LastResetAt`。
- `model/quota/balance_sync.go`：`ResetExpiredBalances` 从 `plan.LastResetAt` 判断，重置后更新 `quota_plans.last_reset_at`。
- `model/quota/quota_plan_manager.go`：`ResetBalance` 手动重置时更新 `quota_plans.last_reset_at`。

### 6. 删除 `quota_balances` 表及相关代码

- 删除 `storage/rdb/quota/quota_balance.go`。
- 删除 `model/quota/quota_balance.go`。
- 删除 `model/quota/adapters.go` 中的 `quotaBalanceStoragerAdapter`。
- 删除 `model/shared/types.go` 中的 `QuotaBalanceStorager` 接口。
- 删除 `stateful/container/rdb/components.go` 中的 `QuotaBalanceStorager` 初始化与注入。
- 从 `APIKeyManager`、`EntityManager`、`QuotaPlanManager` 中移除 `quotaBalanceStorager` 依赖。
- 删除 `db_ddl.sql` / `db_ddl_sqlite.sql` 中的 `quota_balances` 表定义。

### 7. 调整配额计划变更时的余额行为

- 当 `quota_plan` 的 `quota` / `unit` / `unlimited` 任一发生变化时，按以下规则调整该计划下所有 API-Key / Entity 的 Redis 余额：
  - **仅 `quota` 变化、单位不变**：保留已使用量 `used`，根据 Redis 实时剩余量计算 `used = old_quota - remaining`，然后 `new_remaining = max(0, new_quota - used)`，使用 `SetRemaining` 下发。
  - **`unit` 变化、`unlimited` 切换或新建计划**：`used` 清零，`remaining` 置为新配额（`unlimited` 时用 sentinel），使用 `ResetToQuota`。
- 普通属性修改（如 `enabled`、`description`、`allow_models`）不触发余额变更。

### 8. 更新集成测试

- 重写 `test/integration/tests/api_key/quota_query/quota_query_test.go`：移除对 `quota_balances` 的直接 SQL 更新，改为验证 Redis 实时读取的初始余额。
- 重写 `test/integration/tests/api_key/quota_update/quota_update_test.go`：按"变化即重置"的新行为调整断言；无限配额场景验证 `balance` 字段不存在；新增非零 used 场景。
- 重写 `test/integration/tests/entity/quota_plan/quota_plan_test.go`：移除对 `quota_balances` 的直接 SQL 更新。
- 重写 `test/integration/tests/entity/quota_update/quota_update_test.go`：同 API-Key 调整逻辑；新增非零 used 场景。
- 引入 `miniredis` 替换集成测试中的内存 Mock Redis：
  - `test/integration/testutil/server.go`：每个测试进程启动一个 `miniredis` 实例，生成 `name_conf.data` 并修改 `ai_gateway_api.toml`，让 ai-gateway-api 子进程通过 BNS 连接到该 miniredis。
  - `test/integration/testutil/db.go`：修复 seed provider 的列名 `keys` -> `api_keys`。
  - `stateful/config_redis.go`：`name_conf.data` 加载路径优先使用 `DefaultConfig.ConfigDir`，确保子进程从临时配置目录读取。
  - 新增 `ServerManager.SetQuotaRemaining` / `GetQuotaRemaining` 辅助方法，支持按 API-Key value / Entity ID 直接写入 Redis。

## 数据迁移

上线前一次性执行：

```sql
UPDATE quota_plans qp
JOIN quota_balances qb ON qp.id = qb.quota_plan_id
SET qp.last_reset_at = qb.last_reset_at;
```

SQLite 使用对应 UPDATE 语法。

## 兼容性说明

- OpenAPI 响应字段保持不变，余额语义从"准实时（最多滞后 1 分钟）"变为"实时"。
- Redis 读取失败时接口返回错误，不再返回可能滞后的数据库余额。
- `quota_balances` 表删除后，依赖其 `used` / `remaining` 的外部系统需提前迁移到 Redis 或内部接口。
- 配额计划仅 `quota` 变化（单位不变）时，仍保留原有"保留 used"语义；`unit` 变化或 `unlimited` 切换时余额重置，与旧行为一致。
- 无限配额的 `quota-plan` 接口继续返回 sentinel balance（`used=0`, `remaining=100000000`），保持 OpenAPI 响应兼容性。

## 数据面影响

- 不影响 BFE 数据面的配额扣减逻辑。
- 不影响 Redis Key 生成规则（`QUOTA_<key>`）。

## 验证结果

- `go test ./model/...`：通过。
- `go test ./...`（ai-gateway-api 单元测试）：通过。
- `go test ./tests/api_key/quota_query/... ./tests/api_key/quota_update/... ./tests/entity/quota_plan/... ./tests/entity/quota_update/... ./tests/api_key/quota_reset/... ./tests/entity/quota_reset/...`：通过。
- `go test ./tests/schema/openapi/...`：通过。
- `go test ./tests/...`（全量集成测试）：除预存在的 `innerapi/rate_limit_policy` 的 `TestInnerAPI_RateLimitPolicy_RedisKeyStable` 失败外，其余全部通过；该失败与 quota 重构及 miniredis 改造均无关（未修改 rate-limit 相关代码）。

## 待实现清单

- [x] 扩展 `redis_client.Client` 接口及实现
- [x] 扩展 `MockRedisClient`
- [x] 扩展 `QuotaCache` 接口及实现
- [x] API-Key / Entity 查询路径改为读 Redis
- [x] 删除 `SyncAllBalances`
- [x] `quota_plans` 表新增 `last_reset_at`
- [x] `ResetExpiredBalances` 改从 `quota_plans` 读写 `last_reset_at`
- [x] 删除 `quota_balances` 表及相关代码
- [x] 更新单元测试
- [x] 更新集成测试
- [x] 引入 miniredis 替换集成测试 Mock Redis
- [x] 更新 design-docs
