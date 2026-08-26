# quota-plan 余额同步机制移除设计变更说明

> 对应变更：`ai-gateway-api` quota-plan 余额从 "Redis 实时计数 + 数据库定时同步" 改为 "直接读 Redis"，并删除 `quota_balances` 表。

## 1. 概述

### 1.1 问题现象

当前 `ai-gateway-api` 对 `quota_plan` 余额采用混合架构：

- **Redis**：由 BFE 数据面实时扣减，是热数据。
- **`quota_balances` 表**：由 `BalanceSyncManager.SyncAllBalances` 每分钟从 Redis 汇总回写，用于管理面查询。

该机制带来以下问题：

1. **数据滞后**：`quota_balances` 最多滞后 1 分钟，管理面看到的余额不是最新的。
2. **读性能差**：OpenAPI 从数据库读取余额，性能不如直接读 Redis。
3. **写开销大**：同步复杂度随 API-Key / Entity 数量线性增长，达到数千数万时 Redis + DB 压力很大。
4. **资源浪费**：每分钟全量同步，但管理面查询余额的频率通常并不高。
5. **架构冗余**：`last_reset_at` 保存在 `quota_balances` 表，而配额计划本身在 `quota_plans` 表，导致重置逻辑需要跨表维护。

### 1.2 变更目标

1. OpenAPI 查询余额直接读取 Redis。
2. 删除 `BalanceSyncManager.SyncAllBalances` 定时同步。
3. 把 `last_reset_at` 从 `quota_balances` 迁移到 `quota_plans` 表。
4. 彻底删除 `quota_balances` 表及相关代码。
5. 保持 OpenAPI 响应字段与旧行为兼容：
   - 仅 `quota` 变化（单位不变）时保留 `used`；
   - `unit` 变化或 `unlimited` 切换时重置余额；
   - 无限配额仍返回 sentinel balance（`used=0`, `remaining=100000000`）。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`、`bfe` |
| 涉及模块 | `model/api_key`、`model/entity`、`model/quota`、`model/quotacache`、`model/shared`、`stateful`、`storage/rdb/quota`、`endpoints/openapi_v1/api_key`、`endpoints/openapi_v1/entity` |
| 接口契约 | OpenAPI `/api-keys/{id}/quota-plan`、`/entities/{id}/quota-plan`、列表/详情接口的 `quota_plan.balance` 字段不变 |
| 数据迁移 | 上线前一次性将 `quota_balances.last_reset_at` 回写到 `quota_plans.last_reset_at` |

---

## 2. 现状代码分析

### 2.1 余额查询路径

```
endpoints/openapi_v1/api_key/get_quota_plan.go
  └─ container.APIKeyManager.FetchAPIKey(...)
       └─ model/api_key/api_key.go
            ├─ 事务内读取 api_keys / quota_plans
            ├─ 事务内读取 quota_balances（used / remaining）
            └─ 事务外可选读取 Redis（如果 quota_balances 不存在）

endpoints/openapi_v1/entity/get_quota_plan.go
  └─ container.EntityManager.FetchEntity(...)
       └─ model/entity/entity_manager.go
            ├─ 事务内读取 entities / quota_plans
            ├─ 事务内读取 quota_balances
            └─ 事务外可选读取 Redis
```

列表接口同样先读 `quota_balances`，Redis 仅作为兜底。

### 2.2 定时同步路径

```
model/quota/scheduler.go
  └─ 定时触发 BalanceSyncManager.SyncAllBalances(...)
       └─ model/quota/balance_sync.go
            ├─ 遍历所有 quota_plan
            ├─ 对每个 plan 下所有 API-Key / Entity 读取 Redis 剩余量
            ├─ 汇总 used / remaining
            └─ 写回 quota_balances 表
```

`SyncAllBalances` 每分钟执行一次，全量扫描 Redis 与 DB。

### 2.3 定期重置路径

```
model/quota/balance_sync.go
  └─ ResetExpiredBalances
       ├─ 遍历 quota_plans
       ├─ 读取 quota_balances.last_reset_at 判断是否到期
       ├─ 重置 Redis
       └─ 更新 quota_balances.last_reset_at
```

重置周期（weekly/monthly）的判断依赖 `quota_balances` 表。

### 2.4 存在的问题

- `quota_balances` 是冷数据副本，与 Redis 存在秒级到分钟级延迟。
- 管理面查询余额需要先走 DB 事务，性能低于 Redis。
- `SyncAllBalances` 全量同步的写放大随 API-Key / Entity 数量线性增长。
- `last_reset_at` 与 quota plan 分离，导致重置逻辑需要维护两张表的一致性。

---

## 3. 详细设计

### 3.1 核心原则

- **Redis 是余额的唯一真实来源**：OpenAPI 查询余额直接读 Redis。
- **删除冷数据副本**：不再维护 `quota_balances` 表，彻底移除 `SyncAllBalances`。
- **`last_reset_at` 上移至 `quota_plans`**：重置逻辑只操作一张表。
- **批量读取优化**：列表接口使用 `MGET` 批量读取 Redis，减少网络往返。
- **兼容性**：无限配额仍返回 sentinel balance；quota 变化保留 used 语义不变。

### 3.2 扩展 Redis 客户端接口

在 `bfe/bfe_util/redis_client/client.go` 新增批量读取接口：

```go
type Client interface {
    // ... 原有方法
    GetInt64Batch(keys []string) ([]int64, error)
}
```

`bfe/bfe_util/redis_client/redis_bns.go` 的 `RedisClient` 实现：

```
1. 按 Redis 集群 ID 对 keys 分组。
2. 对每个集群执行 MGET。
3. 合并结果并返回 []int64。
```

`ai-gateway-api/stateful/mock_redis.go` 的 `MockRedisClient` 同步实现 `GetInt64Batch`，用于单元测试与集成测试。

### 3.3 扩展 QuotaCache 接口

在 `model/quotacache/quotacache.go` 新增：

```go
type QuotaCache interface {
    // ... 原有方法
    BatchGetRemaining(ctx context.Context, keys []string, unit *string) (map[string]float64, error)
}
```

`model/quotacache/redis.go` 的 `RedisQuotaCache` 实现：

```
1. 将 owner keys 转换为 Redis keys（QUOTA_<key>）。
2. 调用 client.GetInt64Batch。
3. 按单位转换 Redis 固定点整数为业务浮点数。
4. 返回 map[ownerKey]remaining。
```

### 3.4 API-Key / Entity 查询路径改造

#### `model/api_key/api_key.go`

- 从 `APIKeyManager` 移除 `quotaBalanceStorager` 依赖。
- `populateAssociatedData` / `FetchAPIKeyList` 不再读取 `quota_balances`。
- 新增 `populateQuotaBalance`（单对象）与 `populateQuotaBalances`（列表）：
  - 非无限配额：从 Redis 读取剩余量，填充 `QuotaPlan.Balance`；
  - 无限配额：填充 sentinel balance（`used=0`, `remaining=100000000`）。
- `fillQuotaBalance` 根据 `quota - remaining` 计算 `used`。

#### `model/entity/entity_manager.go`

- 与 API-Key 侧完全对称的改造。

#### 列表查询优化

```
for each API-Key in list:
    if unlimited: fillUnlimitedQuotaBalance
    else: collect key -> item mapping, grouped by unit

for each unit group:
    result = quotaCache.BatchGetRemaining(keys, unit)
    for each item: fillQuotaBalance(item.one.QuotaPlan, result[item.key])
```

单对象查询仍使用 `quotaCache.GetRemaining`。

### 3.5 删除 SyncAllBalances

- 删除 `model/quota/balance_sync.go` 中的 `SyncAllBalances` 和 `syncPlanBalance`。
- 删除 `model/quota/scheduler.go` 中对 `SyncAllBalances` 的调用。
- 删除 `model/quota/balance_sync_test.go` 中相关测试。

### 3.6 迁移 last_reset_at

#### 数据库 DDL

- `db_ddl.sql` / `db_ddl_sqlite.sql`：
  - 删除 `quota_balances` 表定义；
  - 在 `quota_plans` 表新增 `last_reset_at` 字段。

#### 存储层

- `storage/rdb/internal/dao/table_quota_plans.go`：`TQuotaPlan` / `TQuotaPlanParam` 新增 `LastResetAt`。
- `storage/rdb/quota/quota_plan.go`：转换层处理 `LastResetAt`。
- 删除 `storage/rdb/quota/quota_balance.go` 与 `table_quota_balances.go`。

#### 模型层

- `model/quota/quota_plan.go`：`QuotaPlanParam` 新增 `LastResetAt`。
- `model/quota/balance_sync.go`：`ResetExpiredBalances` 从 `plan.LastResetAt` 判断，重置后更新 `quota_plans.last_reset_at`。
- `model/quota/quota_plan_manager.go`：`ResetBalance` 手动重置时更新 `quota_plans.last_reset_at`。

### 3.7 删除 quota_balances 相关代码

- 删除 `model/quota/quota_balance.go`。
- 删除 `model/quota/adapters.go` 中的 `quotaBalanceStoragerAdapter`。
- 删除 `model/shared/types.go` 中的 `QuotaBalanceStorager` 接口。
- 删除 `stateful/container/rdb/components.go` 中的 `QuotaBalanceStorager` 初始化与注入。
- 从 `APIKeyManager`、`EntityManager`、`QuotaPlanManager` 中移除 `quotaBalanceStorager` 依赖。

### 3.8 Endpoint 层微调

- `endpoints/openapi_v1/api_key/get_quota_plan.go` 与 `endpoints/openapi_v1/entity/get_quota_plan.go`：
  - `Balance` 字段增加 `omitempty`，避免无余额时返回 `null`；
  - 无限配额由于 Model 层已填充 sentinel，`balance` 仍会返回。

### 3.9 ApplyQuotaPlanChange 行为保持

`model/quota/quota_plan_manager.go` 的 `ApplyQuotaPlanChange` / `adjustQuota` 保持以下语义：

- **仅 `quota` 变化、单位不变**：
  - 批量读取 Redis 当前剩余量；
  - `used = old_quota - current_remaining`；
  - `new_remaining = max(0, new_quota - used)`；
  - 调用 `SetRemaining` 下发。
- **`unit` 变化、`unlimited` 切换或新建计划**：
  - `used = 0`，`remaining = new_quota`（unlimited 时为 sentinel）；
  - 调用 `ResetToQuota` 重置 Redis。
- **普通属性修改**：不触发余额调整。

---

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `bfe/bfe_util/redis_client/client.go` | 新增 `GetInt64Batch` 接口。 |
| `bfe/bfe_util/redis_client/redis_bns.go` | 实现 `GetInt64Batch`：按集群分组 MGET。 |
| `ai-gateway-api/stateful/mock_redis.go` | `MockRedisClient` 实现 `GetInt64Batch`。 |
| `model/quotacache/quotacache.go` | 新增 `BatchGetRemaining` 接口。 |
| `model/quotacache/redis.go` | 实现 `BatchGetRemaining`。 |
| `model/api_key/api_key.go` | 移除 `quotaBalanceStorager`；查询路径改读 Redis；新增 `fillUnlimitedQuotaBalance`。 |
| `model/entity/entity_manager.go` | 与 API-Key 侧对称改造。 |
| `model/quota/balance_sync.go` | 删除 `SyncAllBalances`；`ResetExpiredBalances` 改从 `quota_plans` 读写 `last_reset_at`。 |
| `model/quota/quota_plan_manager.go` | 移除 `balanceStorager`；`ResetBalance` 更新 `quota_plans.last_reset_at`；`adjustQuota` 保留 used 语义。 |
| `model/quota/quota_plan.go` | `QuotaPlanParam` 新增 `LastResetAt`。 |
| `model/quota/adapters.go` | 删除 `quotaBalanceStoragerAdapter`。 |
| `model/quota/quota_balance.go` | 删除。 |
| `model/shared/types.go` | 删除 `QuotaBalanceStorager` 接口。 |
| `stateful/container/components.go` / `stateful/container/rdb/components.go` | 删除 `QuotaBalanceStorager` 注入。 |
| `storage/rdb/internal/dao/table_quota_plans.go` | 新增 `LastResetAt` 字段。 |
| `storage/rdb/internal/dao/table_quota_balances.go` | 删除。 |
| `storage/rdb/quota/quota_plan.go` | 转换层处理 `LastResetAt`。 |
| `storage/rdb/quota/quota_balance.go` | 删除。 |
| `storage/rdb/quota/quota_balance_test.go` | 删除。 |
| `db_ddl.sql` / `db_ddl_sqlite.sql` | 删除 `quota_balances` 表，给 `quota_plans` 增加 `last_reset_at`。 |
| `endpoints/openapi_v1/api_key/get_quota_plan.go` | `Balance` 增加 `omitempty`。 |
| `endpoints/openapi_v1/entity/get_quota_plan.go` | `Balance` 增加 `omitempty`。 |
| `test/integration/tests/api_key/quota_query/quota_query_test.go` | 移除 `quota_balances` SQL；改为验证 Redis 读取。 |
| `test/integration/tests/api_key/quota_update/quota_update_test.go` | 同上；无限配额改回验证 sentinel balance。 |
| `test/integration/tests/entity/quota_plan/quota_plan_test.go` | 移除 `quota_balances` SQL；改为验证 Redis 读取。 |
| `test/integration/tests/entity/quota_update/quota_update_test.go` | 同上；无限配额改回验证 sentinel balance。 |
| `test/integration/tests/api_key/design.md` | 移除 `quota_balances` 描述，更新为 Redis 实时读取。 |
| `test/integration/tests/entity/design.md` | 同上。 |

---

## 5. 测试计划

### 5.1 单元测试

#### `model/quota` 层

1. **`ApplyQuotaPlanChange` quota 增加保留 used**
   - old: quota=100, Redis remaining=40
   - new: quota=150
   - 期望: 调用 `SetRemaining` 传入 90。

2. **`ApplyQuotaPlanChange` quota 减少保留 used**
   - old: quota=100, Redis remaining=40
   - new: quota=50
   - 期望: 调用 `SetRemaining` 传入 0。

3. **`ApplyQuotaPlanChange` unit 变化重置**
   - old: unit=total_token, quota=100
   - new: unit=RMB, quota=10
   - 期望: 调用 `ResetToQuota` 传入 10。

4. **`ApplyQuotaPlanChange` unlimited 切换**
   - limited -> unlimited: 调用 `ResetToQuota` 传入 sentinel。
   - unlimited -> limited: 调用 `ResetToQuota` 传入新 quota。

5. **`ResetExpiredBalances` 从 `quota_plans.last_reset_at` 判断**
   - 构造 `last_reset_at` 到期的 plan；
   - 断言重置后 `quota_plans.last_reset_at` 被更新。

#### `model/api_key` 与 `model/entity` 层

1. **FetchAPIKey / FetchEntity 优先读 Redis**
   - Redis 有值时，`QuotaPlan.Balance` 与 Redis 一致。
   - Redis 无值时，`remaining` 等于 `quota`，`used=0`。

2. **无限配额返回 sentinel balance**
   - `QuotaPlan.Unlimited=true` 时，`Balance.Used=0`、`Balance.Remaining=100000000`。

3. **FetchAPIKeyList / FetchEntityList 批量读 Redis**
   - 多个对象混合 unlimited / 有限额；
   - 断言所有对象的 `Balance` 被正确填充。

### 5.2 集成测试

1. **API-Key quota-plan 查询包含 balance**
   - 创建有限配额 API-Key；
   - 查询 `/api-keys/{id}/quota-plan`；
   - 断言 `balance.used=0`、`balance.remaining=quota`。

2. **Entity quota-plan 查询 RMB 精度**
   - 创建 RMB 配额 Entity；
   - 查询 `/entities/{id}/quota-plan`；
   - 断言 `balance.remaining` 精度正确。

3. **更新 quota_plan 后余额调整**
   - 仅修改 quota（无使用量）：断言 `remaining` 为新 quota。
   - 仅修改 quota（存在使用量）：断言 `used = old_quota - current_remaining`、`remaining = max(0, new_quota - used)`。
   - 修改 unit：断言 `used=0`、`remaining=新 quota`。
   - unlimited false -> true：断言 `remaining=100000000`。

> 集成测试使用 `miniredis` 作为嵌入式 Redis，测试进程可直接写入 Redis 构造非零 used，从而完整覆盖“保留 used”路径。

---

## 6. 数据迁移

上线前一次性执行：

```sql
UPDATE quota_plans qp
JOIN quota_balances qb ON qp.id = qb.quota_plan_id
SET qp.last_reset_at = qb.last_reset_at;
```

SQLite 使用对应 UPDATE 语法。

迁移完成后，确认 `quota_balances` 表无其他系统依赖，即可删除该表。

---

## 7. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| Redis 读取失败时接口报错 | 删除 `quota_balances` 兜底后，Redis 异常会直接影响管理面查询 | 监控 Redis 可用性；Redis 故障期间管理面余额查询不可用，但数据面扣减不受影响 |
| 管理面性能依赖 Redis | 列表接口批量 MGET，Redis 延迟成为瓶颈 | 使用 `GetInt64Batch` 按集群分组 MGET，减少网络往返；Redis 集群本身可水平扩展 |
| `last_reset_at` 迁移遗漏 | 若上线前未正确迁移，定期重置可能误判 | 上线脚本必须执行迁移 SQL，并在迁移后校验 `quota_plans.last_reset_at` 非空率 |
| 外部系统依赖 `quota_balances` | 其他服务或报表直接读取 `quota_balances` | 上线前确认所有下游已迁移到 Redis 或内部接口；删除表前保留备份 |
| 并发更新 quota_plan | 多个管理员同时修改同一配额计划 | `ApplyQuotaPlanChange` 基于 Redis 实时剩余量计算，DB 事务保证 plan 元数据一致性 |

---

## 8. 后续可优化

- [x] 引入 `miniredis` 替换集成测试中的内存 Mock Redis，测试进程可直接写入 Redis，完整覆盖“保留 used”的非零 used 路径。
  - `test/integration/testutil/server.go`：启动 `miniredis`，写入 `name_conf.data` 并将 `RedisConf.Bns` 指向 miniredis。
  - `test/integration/tests/api_key/quota_update/quota_update_test.go`：新增非零 used 场景。
  - `test/integration/tests/entity/quota_update/quota_update_test.go`：新增非零 used 场景。
  - `stateful/config_redis.go`：`name_conf.data` 路径优先从 `DefaultConfig.ConfigDir` 读取，保证子进程能正确解析 miniredis 地址。
- 进一步评估是否将 `QuotaCache` 的批量读取抽象为独立服务，便于后续支持多 Redis 集群、读写分离等场景。
