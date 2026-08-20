# 将 Redis 配额操作从 endpoints 下沉到 model 层

## 1. 概述

### 1.1 变更背景

当前 `ai-gateway-api` 中，API-Key 与 Entity 的配额（quota）Redis 操作大量散落在 `endpoints/openapi_v1/api_key` 与 `endpoints/openapi_v1/entity` 包中。端点层直接调用 `stateful.DefaultClientSet.RedisClient` 进行 `GetInt64 / IncrBy` 等操作，并自行处理 Redis Key 生成、定点数转换、`redigo: nil returned` 错误等细节。

这种分层方式导致：

- **端点层混入存储细节**：endpoints 需要了解配额缓存的存储语义。
- **业务逻辑重复**：同样的“重置 Redis 剩余量为 quota 总量”逻辑在 8 个端点文件中重复实现。
- **测试与复用困难**：端点层直接依赖全局 `stateful.DefaultClientSet`，难以在单元测试中 mock 或替换存储实现。
- **事务边界模糊**：数据库更新在 model 层事务内完成，Redis 操作在端点层事务外执行，存在部分失败风险。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-08-13 |
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `endpoints/openapi_v1/api_key`、`endpoints/openapi_v1/entity`、`model/quota`、`model/icluster_conf` |
| 变更类型 | 架构重构：将 Redis 缓存操作下沉到 model 层，端点层只负责参数校验与调用 model |

本次变更的目标：

1. 在 `model/quota` 中引入 `QuotaCache` 接口，封装 API-Key / Entity 的实时配额缓存操作。
2. 端点层不再直接调用 `stateful.DefaultClientSet.RedisClient`。
3. `APIKeyManager` / `EntityManager` / `QuotaPlanManager` 通过 `QuotaCache` 维护 Redis 缓存一致性。
4. 统一 Redis Key 规则、定点数转换、nil 错误处理等逻辑。

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **接口隔离** | model 层不直接依赖 `stateful.DefaultClientSet.RedisClient`，只依赖 `quotacache.QuotaCache` 接口。 |
| **单一职责** | endpoints 只负责 HTTP 参数解析、校验、调用 model；model 负责业务逻辑与缓存一致性。 |
| **集中封装** | Redis Key 生成、定点数转换、nil 处理全部封装在 `model/quotacache/quota_cache_redis.go` 中。 |
| **测试友好** | 通过 `QuotaCache` mock，端点层单元测试不再需要真实 Redis。 |
| **保持语义** | 仍使用 `GetInt64 + IncrBy(delta)` 而非 `SET`，以兼容并发请求扣减。 |
| **最终一致** | Redis 作为缓存，失败时记录日志但不回滚 DB（与当前语义保持一致）。 |

---

## 2. 现状问题

### 2.1 端点层直接操作 Redis 的文件

| 文件路径 | 操作说明 |
|---|---|
| `endpoints/openapi_v1/api_key/create.go` | 创建 API-Key 后初始化 Redis 剩余量 |
| `endpoints/openapi_v1/api_key/update.go` | 更新 quota_plan 后同步 Redis |
| `endpoints/openapi_v1/api_key/full_update.go` | 全量更新 quota_plan 后同步 Redis |
| `endpoints/openapi_v1/api_key/reset_quota.go` | 手动重置 Redis 剩余量 |
| `endpoints/openapi_v1/entity/create.go` | 创建 Entity 后初始化 Redis 剩余量 |
| `endpoints/openapi_v1/entity/update.go` | 更新 quota_plan 后同步 Redis |
| `endpoints/openapi_v1/entity/full_update.go` | 全量更新 quota_plan 后同步 Redis |
| `endpoints/openapi_v1/entity/reset_quota.go` | 手动重置 Redis 剩余量 |

### 2.2 反模式示例

以 `endpoints/openapi_v1/api_key/create.go` 为例：

```go
if err == nil && param.Key != nil && param.QuotaPlan != nil &&
    (param.QuotaPlan.Unlimited == nil || !*param.QuotaPlan.Unlimited) &&
    param.QuotaPlan.Quota != nil &&
    stateful.DefaultClientSet != nil && stateful.DefaultClientSet.RedisClient != nil {
    redisKey := stateful.AIUsedQuotaKey(*param.Key)
    targetValue := golibquota.PtrToRedisValue(param.QuotaPlan.Quota, param.QuotaPlan.Unit)
    currentValue, errGet := stateful.DefaultClientSet.RedisClient.GetInt64(redisKey)
    if errGet != nil {
        _, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, targetValue)
    } else {
        delta := targetValue - currentValue
        _, err = stateful.DefaultClientSet.RedisClient.IncrBy(redisKey, delta)
    }
}
```

这段逻辑存在以下问题：

- 端点层需要知道 `stateful.AIUsedQuotaKey` 的 Key 生成规则。
- 端点层需要调用 `golibquota.PtrToRedisValue` 做定点数转换。
- 端点层需要处理 Redis 客户端 nil、Get 错误等细节。
- 同样的代码在 8 个端点中重复出现。

### 2.3 model 层仍直接操作 Redis 的情况

除端点层外，以下 model 文件也直接访问 Redis：

- `model/icluster_conf/api_key.go`：`GetRemainingQuota` 直接读取 Redis。
- `model/quota/balance_sync.go`：`BalanceSyncManager` 多处直接调用 `stateful.DefaultClientSet.RedisClient`。

这些也需要一并改造为通过 `QuotaCache` 接口访问。

---

## 3. 目标架构

### 3.1 分层职责

```
endpoints/openapi_v1/api_key/*.go
endpoints/openapi_v1/entity/*.go
    │
    ▼
model/icluster_conf/APIKeyManager
model/quota/EntityManager
model/quota/QuotaPlanManager
    │
    ▼
model/quotacache/QuotaCache (interface)
    │
    ▼
model/quotacache/RedisQuotaCache (implementation)
    │
    ▼
stateful.DefaultClientSet.RedisClient
```

### 3.2 新增/改造的包结构

本次将 Redis 配额缓存抽象独立为 `model/quotacache` 包，避免 `model/quota` 包过于臃肿，同时便于后续复用于其他需要缓存的业务。

```
model/
├── quota/
│   ├── entity.go
│   ├── entity_manager.go          # 改造：依赖 quotacache.QuotaCache
│   ├── quota_plan.go
│   ├── quota_plan_manager.go
│   ├── quota_balance.go
│   ├── balance_sync.go            # 改造：依赖 quotacache.QuotaCache
│   └── scheduler.go
└── quotacache/
    ├── quota_cache.go             # QuotaCache 接口定义
    └── quota_cache_redis.go       # Redis 实现
```

---

## 4. 详细设计

### 4.1 QuotaCache 接口

新增 `model/quotacache/quota_cache.go`：

```go
package quotacache

import "context"

// QuotaCache 封装 API-Key / Entity 的实时配额缓存操作。
// 实现者应保证对 Redis 的访问是线程安全的。
type QuotaCache interface {
    // GetRemaining 获取指定对象的实时剩余配额（已按 unit 转换为浮点数）。
    GetRemaining(ctx context.Context, key string, unit *string) (float64, error)

    // SetRemaining 将指定对象的 Redis 剩余量设置为 target。
    // target 为 nil 或 unlimited 时不执行任何操作。
    SetRemaining(ctx context.Context, key string, quota *float64, unit *string) error

    // ResetToQuota 将 Redis 剩余量重置为 quota 总量。
    // 内部使用 GetInt64 + IncrBy(delta) 保证并发安全。
    ResetToQuota(ctx context.Context, key string, quota *float64, unit *string) error
}
```

### 4.2 Redis 实现

新增 `model/quotacache/quota_cache_redis.go`：

```go
package quotacache

import (
    "context"
    "strings"

    "github.com/bfenetworks/go-lib/quota"
    "github.com/bfenetworks/bfe/bfe_util/redis_client"
    "github.com/rainway-ai-gateway/ai-gateway-api/stateful"
)

type redisQuotaCache struct {
    client redis_client.Client
}

func NewRedisQuotaCache(client redis_client.Client) QuotaCache {
    return &redisQuotaCache{client: client}
}

func (c *redisQuotaCache) GetRemaining(ctx context.Context, key string, unit *string) (float64, error) {
    if c.client == nil {
        return 0, nil
    }
    redisKey := stateful.AIUsedQuotaKey(key)
    remain, err := c.client.GetInt64(redisKey)
    if err != nil {
        if strings.Contains(err.Error(), "redigo: nil returned") {
            return 0, nil
        }
        return 0, err
    }
    return quota.PtrFromRedisValue(remain, unit), nil
}

func (c *redisQuotaCache) SetRemaining(ctx context.Context, key string, quotaValue *float64, unit *string) error {
    if c.client == nil || quotaValue == nil {
        return nil
    }
    redisKey := stateful.AIUsedQuotaKey(key)
    targetValue := quota.PtrToRedisValue(quotaValue, unit)
    currentValue, err := c.client.GetInt64(redisKey)
    if err != nil {
        if !strings.Contains(err.Error(), "redigo: nil returned") {
            return err
        }
        currentValue = 0
    }
    delta := targetValue - currentValue
    _, err = c.client.IncrBy(redisKey, delta)
    return err
}

func (c *redisQuotaCache) ResetToQuota(ctx context.Context, key string, quotaValue *float64, unit *string) error {
    return c.SetRemaining(ctx, key, quotaValue, unit)
}
```

> 说明：`stateful.AIUsedQuotaKey` 仍保留在 `stateful` 包中，由 `quota_cache_redis.go` 统一调用，避免端点层和 model 其他位置再次散播 Key 规则。

### 4.3 Manager 注入 QuotaCache

#### APIKeyManager

在 `model/icluster_conf/api_key.go` 中：

```go
import "github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"

type APIKeyManager struct {
    txn             itxn.TxnStorager
    storager        APIKeyStorager
    quotaPlanMgr    *quota.QuotaPlanManager
    quotaCache      quotacache.QuotaCache  // 新增
}

func NewAPIKeyManager(
    txn itxn.TxnStorager,
    storager APIKeyStorager,
    quotaPlanMgr *quota.QuotaPlanManager,
    quotaCache quotacache.QuotaCache,  // 新增
) *APIKeyManager {
    return &APIKeyManager{
        txn:          txn,
        storager:     storager,
        quotaPlanMgr: quotaPlanMgr,
        quotaCache:   quotaCache,
    }
}
```

`CreateAPIKey` / `UpdateAPIKey` 在 DB 事务成功后，调用 `m.quotaCache.SetRemaining(...)` 初始化或同步 Redis。

`GetRemainingQuota` 改为调用 `m.quotaCache.GetRemaining(...)`。

#### EntityManager

在 `model/quota/entity_manager.go` 中类似注入 `quotacache.QuotaCache`，并在 `CreateEntity` / `UpdateEntity` 中维护 Redis。

#### QuotaPlanManager

`ResetBalance` 可保持仅操作数据库，但新增一个组合方法或在上层 Manager 中调用：

```go
// 在 APIKeyManager 或一个专门的 QuotaResetService 中
func (m *APIKeyManager) resetQuotaCache(ctx context.Context, apiKey *APIKeyParam, quota *float64, unit *string) error {
    if apiKey.Key == nil {
        return nil
    }
    return m.quotaCache.ResetToQuota(ctx, *apiKey.Key, quota, unit)
}
```

若希望完全消除 Manager 对缓存细节的关注，可在 `model/quotacache` 中新增 `QuotaResetService`，将 DB + Redis 组合逻辑集中管理。

---

## 5. 端点层改造

### 5.1 改造原则

端点层只保留以下职责：

1. HTTP 参数解析与校验
2. 调用 `APIKeyManager` / `EntityManager` / `QuotaPlanManager`
3. 构造响应

删除所有 `stateful.DefaultClientSet.RedisClient` 的直接调用、Redis Key 生成、定点数转换逻辑。

### 5.2 api_key/create.go 改造示例

**改造前：**

```go
result, err := container.APIKeyManager.CreateAPIKey(ctx, param)
// ... 成功后直接操作 Redis ...
```

**改造后：**

```go
result, err := container.APIKeyManager.CreateAPIKey(ctx, param)
if err != nil {
    return nil, err
}
// Redis 初始化已在 APIKeyManager.CreateAPIKey 内部完成
return result, nil
```

### 5.3 api_key/reset_quota.go 改造示例

**改造前：**

```go
if err := container.QuotaPlanManager.ResetBalance(ctx, *apiKey.QuotaPlanID, &resetQuota, true); err != nil {
    return nil, err
}
// ... 端点层直接操作 Redis ...
```

**改造后：**

```go
if err := container.QuotaPlanManager.ResetBalance(ctx, *apiKey.QuotaPlanID, &resetQuota, true); err != nil {
    return nil, err
}
if err := container.APIKeyManager.ResetQuotaCache(ctx, apiKey, &resetQuota, unit); err != nil {
    return nil, err
}
```

或者将 DB + Redis 组合逻辑进一步封装到 `QuotaResetService`：

```go
if err := container.QuotaResetService.ResetAPIKeyQuota(ctx, apiKey, &resetQuota, unit); err != nil {
    return nil, err
}
```

---

## 6. model 层改造

### 6.1 model/quota/balance_sync.go

`BalanceSyncManager` 改为依赖 `quotacache.QuotaCache` 接口：

```go
import "github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"

type BalanceSyncManager struct {
    txn             itxn.TxnStorager
    apiKeyStorager  icluster_conf.APIKeyStorager
    balanceStorager QuotaBalanceStorager
    planStorager    QuotaPlanStorager
    entityStorager  EntityStorager
    quotaCache      quotacache.QuotaCache  // 新增
}
```

所有 `stateful.DefaultClientSet.RedisClient.GetInt64 / IncrBy` 调用替换为 `m.quotaCache.GetRemaining / ResetToQuota`。

### 6.2 model/icluster_conf/api_key.go

- `GetRemainingQuota` 改为调用 `m.quotaCache.GetRemaining`。
- `CreateAPIKey` / `UpdateAPIKey` 在 DB 操作成功后调用 `m.quotaCache.SetRemaining`。
- 通过 `NewAPIKeyManager` 注入 `quotacache.QuotaCache`。

### 6.3 model/quota/quota_plan_manager.go

保持 `ResetBalance` 仅操作数据库。Redis 同步由调用方（`APIKeyManager` / `EntityManager` / `QuotaResetService`）通过 `QuotaCache` 完成。

---

## 7. 依赖注入

在 `stateful/container/rdb/components.go`（或类似的依赖注入入口）中初始化 `QuotaCache`：

```go
import (
    "github.com/rainway-ai-gateway/ai-gateway-api/model/quotacache"
)

func NewQuotaCache(client redis_client.Client) quotacache.QuotaCache {
    return quotacache.NewRedisQuotaCache(client)
}

// 注入到各 Manager
apiKeyManager := icluster_conf.NewAPIKeyManager(
    txn,
    apiKeyStorager,
    quotaPlanManager,
    quotaCache,
)

entityManager := quota.NewEntityManager(
    txn,
    entityStorager,
    quotaPlanManager,
    quotaCache,
)

balanceSyncManager := quota.NewBalanceSyncManager(
    txn,
    apiKeyStorager,
    balanceStorager,
    planStorager,
    entityStorager,
    quotaCache,
)
```

---

## 8. 改动文件清单

### 8.1 新增文件

| 文件 | 说明 |
|---|---|
| `model/quotacache/quota_cache.go` | `QuotaCache` 接口定义 |
| `model/quotacache/quota_cache_redis.go` | 基于 Redis 的实现 |
| `model/quotacache/quota_cache_mock.go` | 单元测试用 mock 实现 |

### 8.2 修改文件

#### 端点层

| 文件 | 改动 |
|---|---|
| `endpoints/openapi_v1/api_key/create.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/api_key/update.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/api_key/full_update.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/api_key/reset_quota.go` | 删除 Redis 直接调用，调用 model 方法 |
| `endpoints/openapi_v1/entity/create.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/entity/update.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/entity/full_update.go` | 删除 Redis 直接调用 |
| `endpoints/openapi_v1/entity/reset_quota.go` | 删除 Redis 直接调用，调用 model 方法 |

#### model 层

| 文件 | 改动 |
|---|---|
| `model/quota/entity_manager.go` | 注入 `QuotaCache`，在创建/更新时同步 Redis |
| `model/quota/quota_plan_manager.go` | 可选：新增组合重置方法 |
| `model/quota/balance_sync.go` | 改为依赖 `QuotaCache` 接口 |
| `model/icluster_conf/api_key.go` | 注入 `QuotaCache`，替换 `GetRemainingQuota` 实现 |
| `model/icluster_conf/api_key_test.go` | 更新 mock 与测试 |
| `model/quota/balance_sync_test.go` | 更新 mock 与测试 |

#### 依赖注入

| 文件 | 改动 |
|---|---|
| `stateful/container/rdb/components.go` | 初始化 `QuotaCache` 并注入到各 Manager |

---

## 9. 兼容性

1. **Redis 存储格式不变**：Key 规则、定点数精度均保持不变，BFE 侧无需改动。
2. **接口行为不变**：OpenAPI 请求/响应字段、状态码、错误码均不变。
3. **DB 事务行为不变**：数据库操作仍在 `AtomExecute` 事务内完成。
4. **Redis 最终一致语义不变**：缓存更新失败仍记录日志，不回滚 DB。

---

## 10. 测试策略

### 10.1 单元测试

1. **QuotaCache mock**：新增 `model/quotacache/quota_cache_mock.go`，实现内存版 `QuotaCache`，支持 `GetRemaining` / `SetRemaining` / `ResetToQuota`。
2. **Manager 层测试**：
   - `APIKeyManager.CreateAPIKey`：验证 DB 成功后调用 `QuotaCache.SetRemaining`。
   - `APIKeyManager.UpdateAPIKey`：验证 quota_plan 变更后调用 `QuotaCache.SetRemaining`。
   - `BalanceSyncManager`：验证使用 `QuotaCache` 完成同步。
3. **端点层测试**：使用 mock 的 Manager 和 `QuotaCache`，不再依赖真实 Redis。

### 10.2 集成测试

1. 启动真实 Redis，验证 `RedisQuotaCache` 的 `GetRemaining` / `SetRemaining` / `ResetToQuota` 行为正确。
2. 验证 API-Key 创建、更新、重置后 Redis 余额与 DB 一致。

### 10.3 回归测试

1. 运行现有 `/api-keys` 和 `/entities` 相关接口测试。
2. 运行 `balance_sync` 定时任务测试。
3. 确保 RMB 与 total_token 两种 unit 的行为一致。

---

## 11. 总结

本次重构通过引入 `model/quotacache.QuotaCache` 接口，将 Redis 配额操作从 `endpoints/openapi_v1` 下沉到 `model` 层。改造后：

- 端点层不再直接操作 Redis，职责更清晰。
- Redis Key 生成、定点数转换、nil 错误处理等逻辑集中在 `model/quotacache` 中管理。
- 单元测试更容易 mock，可维护性提升。
- `model/quotacache` 作为独立包，便于后续复用于其他需要缓存的业务。
