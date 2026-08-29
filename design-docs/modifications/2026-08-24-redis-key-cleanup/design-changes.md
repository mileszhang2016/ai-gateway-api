# Redis Key 清理机制：ai-gateway-api 控制面设计变更说明

> 本文档描述 `ai-gateway-api` 控制面需要为 Redis Key 清理做的设计变更。数据面 BFE 无需修改。

## 1. 概述

### 1.1 变更背景

当前 `ai-gateway-api` 在管理 API-Key、Entity 及其 quota-plan、rate-limit 策略时，只负责 Redis Key 的创建与更新，**没有任何地方主动删除已经失效的 Redis Key**。随着 API-Key、Entity 被删除，或 rate-limit 规则被移除/变更，Redis 中会残留大量无用 Key。

### 1.2 变更目标

1. 在 API-Key / Entity 删除时，清理其 quota-plan 与 rate-limit 对应的 Redis Key。
2. 在 rate-limit 规则被删除或变更时，清理旧规则对应的 Redis Key。
3. 明确 quota-plan 常规变更（改配额、改单位、切换 unlimited）不会删除 Quota Key，只会覆盖其值。
4. 采用最简单可行的实现方式：控制面在删除/更新成功后立即执行 Redis Key 删除。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/api_key`、`model/entity`、`model/quotacache`、`model/rate_limit_policy` |
| 变更类型 | 后台 Redis Key 生命周期管理 |
| 数据迁移 | 无 |

---

## 2. Redis Key 格式

### 2.1 Quota Key

控制面定义（`stateful/config_redis.go`）：

```go
func AIUsedQuotaKey(key string) string {
    return fmt.Sprintf("QUOTA_%s", key)
}
```

实际 Redis Key：

```text
QUOTA_<api-key-token>
QUOTA_<entity-id>
```

BFE 侧直接使用控制面导出的 `RedisKey` 字段。

### 2.2 Rate-Limit Key

控制面导出时生成（`model/rate_limit_policy/rate_limit_policy_manager.go`）。本次修改后改为按规则 `name` 生成：

```go
RedisKey: fmt.Sprintf("RL_TPM_rlp-%d_%s", policyID, rule.Name)
RedisKey: fmt.Sprintf("RL_RPM_rlp-%d_%s", policyID, rule.Name)
```

BFE 侧实际访问时拼接前缀：

```go
func buildRedisKey(policyId string, suffix string) string {
    return fmt.Sprintf("%s_%s_%s", "default_bfe", policyId, suffix)
}
```

因此 BFE 实际 Key 形如：

```text
default_bfe_rlp-<policyID>_RL_TPM_rlp-<policyID>_<name>
default_bfe_rlp-<policyID>_RL_RPM_rlp-<policyID>_<name>
```

> 并发限流 Key（`default_bfe_rlp-<policyID>_con`）由 BFE 自行构造，控制面不感知，不在本方案清理范围内。

---

## 3. 删除场景

### 3.1 API-Key 删除

**触发点**：`DELETE /open-api/v1/api-keys/{id}` 数据库事务提交成功后。

**清理范围**：

- 该 API-Key 自身的 Quota Key：`QUOTA_<api-key-token>`
- 该 API-Key 直接绑定的 Rate-Limit Policy 的 Rate-Limit Key

**不清理**：

- 所属 Entity 的 Quota Key
- 所属 Entity 及其父 Entity 的 Rate-Limit Key

原因：Entity 未被删除，仍可能被其他 API-Key 引用。

### 3.2 Entity 删除

**触发点**：`DELETE /open-api/v1/entities/{id}` 数据库事务提交成功后。

**清理范围**：

- 该 Entity 自身的 Quota Key：`QUOTA_<entity-id>`
- 该 Entity 自身 `RateLimitPolicyID` 导出的 Rate-Limit Key

**不清理**：

- 父 Entity 的 Quota Key
- 父 Entity 的 Rate-Limit Key

原因：父 Entity 未被删除，仍可能被其他子 Entity 或 API-Key 引用。

### 3.3 API-Key 更新导致 rate-limit 规则被删除

**触发点**：`PATCH /open-api/v1/api-keys/{id}` / `PUT /open-api/v1/api-keys/{id}` 数据库事务提交成功后。

**清理范围**：当前 API-Key 直接绑定的 Rate-Limit Policy 中，旧规则存在但新规则不存在的 Rate-Limit Key。

由于规则 `name` 在同一 policy 内唯一且不可变，可以按 `name` 精确识别被删除/变更的规则，只清理旧规则存在而新规则不存在的 Rate-Limit Key。

### 3.4 Entity 更新导致 rate-limit 规则被删除

**触发点**：`PATCH /open-api/v1/entities/{id}` / `PUT /open-api/v1/entities/{id}` 数据库事务提交成功后。

**清理范围**：当前被更新的 Entity 的 `RateLimitPolicyID` 对应的旧 Rate-Limit Key。

**不清理**：父 Entity 的 Rate-Limit Key。

---

## 4. 配额 plan 变更不触发 Quota Key 删除

根据 `model/quota/quota_plan_manager.go:206-283` 的 `adjustQuota` 实现：

| 变更类型 | 当前行为 | 是否删除 Quota Key |
|----------|----------|-------------------|
| 有限配额 → 无限配额 | `ResetToQuota(ctx, key, sentinel, unit)` | 否 |
| 无限配额 → 有限配额 | `ResetToQuota(ctx, key, newQuota, unit)` | 否 |
| 单位 `total_token` ↔ `RMB` | `ResetToQuota(ctx, key, newQuota, unit)` | 否 |
| 配额总额变化 | `SetRemaining(ctx, key, remaining, unit)` | 否 |

因此，配额 plan 的常规变更**不会**产生“BFE 旧配置访问已删除 Quota Key”的问题。

---

## 5. 对 BFE 转发的影响

### 5.1 配置下发时序

1. 控制面接收删除/更新请求，先写数据库；
2. 控制面返回 API 响应；
3. `conf-agent` 定期轮询 InnerAPI 拉取最新配置；
4. `conf-agent` 将新配置写入 BFE 并触发热加载；
5. BFE 使用新配置处理后续请求。

因此，控制面的 Redis Key 删除一定发生在 BFE 感知到配置变更之前。

### 5.2 Quota Key 提前删除的影响

BFE 在请求转发前调用 `plan.HasBalance`：

```go
current, err := client.GetInt64(q.RedisKey)
if err != nil {
    return false, 0, err
}
```

若 Quota Key 已被删除，`GetInt64` 返回 Redis 空值错误，BFE 返回 `CodeInternalQuotaError`。

但由于 Quota Key 的删除只发生在 API-Key / Entity 删除场景，而被删除的资源本不应再被使用，因此实际业务影响有限，主要表现为错误码不准确。

### 5.3 Rate-Limit Key 提前删除的影响

BFE 的 TPM/RPM 限流使用 Lua 脚本访问 Redis。若 Key 已被删除，脚本会将其视为该窗口内无任何消耗，请求被放行（相当于限流窗口被重置）。

这会导致旧配置窗口内短暂限流失效、流量突增，但通常窗口仅为 conf-agent 轮询周期 + BFE 热加载时间，可控。

---

## 6. 推荐实现方案：控制面立即清理

基于风险评估，推荐采用最简洁的实现方式：控制面在删除/更新 API 成功后立即执行 Redis Key 删除。

### 6.1 设计

- 在 `APIKeyManager.DeleteAPIKey`、`EntityManager.DeleteEntity` 完成数据库事务提交后，立即删除对应的 Quota Key 和 Rate-Limit Key。
- 在 `APIKeyManager.UpdateAPIKey` / `FullUpdateAPIKey`、`EntityManager.UpdateEntity` / `FullUpdateEntity` 中，若 rate-limit 规则发生删除/变更，立即删除旧规则对应的 Rate-Limit Key。
- quota-plan 的常规变更不触发 Quota Key 清理。

### 6.2 优点

- 实现最简单，不引入延迟队列、worker 等额外组件。
- 内存回收最及时，避免无用 Key 残留。
- 无需修改 BFE 代码。

### 6.3 缺点与缓解

| 缺点 | 缓解措施 |
|------|----------|
| Rate-Limit Key 被提前删除后，BFE 旧配置仍可能访问已删除 Key，导致短暂限流失效 | 窗口短暂，可接受 |
| 删除 API-Key / Entity 后若仍有请求命中 BFE 旧配置，Quota Key 缺失会返回内部错误 | API-Key / Entity 已删除，请求本应被拒绝；错误码不准确，但业务影响有限 |

---

## 7. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `model/quotacache/quotacache.go` | `QuotaCache` 接口新增 `DeleteKeys(ctx, keys []string) error` |
| `model/quotacache/redis.go` | 实现 `DeleteKeys`，使用 Pipeline 批量执行 `UNLINK`/`DEL` |
| `model/api_key/api_key.go` | `DeleteAPIKey` 事务提交后清理 Quota Key 与 Rate-Limit Key；更新时清理旧 Rate-Limit Key |
| `model/entity/entity_manager.go` | `DeleteEntity` 事务提交后清理 Quota Key 与 Rate-Limit Key；更新时清理旧 Rate-Limit Key |
| `model/rate_limit_policy/rate_limit_policy.go` | 将规则 `name` 作为不可变标识，用 `name` 替代数组下标生成 Redis Key |
| `lib/validate/validate.go` | 收紧 rate-limit 规则 `name` 字符集为 `[a-zA-Z0-9_-]`；更新时校验 `name` 不可修改 |

---

## 8. 测试计划

### 8.1 单元测试

1. `QuotaCache.DeleteKeys` 成功删除多个 Key。
2. `APIKeyManager.DeleteAPIKey` 成功后，对应 Quota Key 与 Rate-Limit Key 被删除。
3. `EntityManager.DeleteEntity` 成功后，对应 Quota Key 与 Rate-Limit Key 被删除。
4. `APIKeyManager.UpdateAPIKey` 在 rate-limit 规则变更后，旧规则 Key 被删除。
5. `EntityManager.UpdateEntity` 在 rate-limit 规则变更后，旧规则 Key 被删除。
6. 配额 plan 变更（改配额、改单位、切换 unlimited）不触发 Quota Key 删除。

### 8.2 集成测试

1. 创建 API-Key → 触发配额使用 → 删除 API-Key → 断言对应 `QUOTA_*` Key 不存在。
2. 创建 API-Key 并启用 rate-limit → 触发限流 → 删除 API-Key → 断言对应 `RL_*` Key 不存在。
3. 更新 API-Key 的 rate-limit 规则 → 断言旧规则 Key 不存在、新规则 Key 可正常访问。

---

## 9. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| Rate-Limit 短暂失效 | BFE 旧配置在热加载前仍可能访问已删除的 Rate-Limit Key，导致限流失效 | 窗口短暂，可接受 |
| Redis 压力过大 | 大批量 `UNLINK` 触发 Redis 主从同步压力 | 控制单次删除数量，分批执行 |
| 误删活跃 Key | 代码逻辑错误导致不该删除的 Key 被删除 | 删除前校验 Key 确实属于被删除/变更的资源 |
| 多实例并发 | 多个 ai-gateway-api 实例同时执行删除 | 同一 Key 被多次 `UNLINK` 是幂等的，无需额外同步 |

---

## 10. rate-limit 规则 `name` 稳定标识改造

本次修改将 rate-limit 规则的 `name` 从“可修改的展示字段”提升为“policy 内唯一且不可变的规则标识”，并用于生成稳定的 Redis Key。

### 10.1 改造原因

原方案按数组下标生成 Redis Key：

```go
RedisKey: fmt.Sprintf("RL_TPM_rlp-%d_%d", policyID, idx)
```

导致：
- 删除数组中间规则后，后续规则下标前移，旧 Key 与新 Key 可能重叠；
- 规则顺序调整时，即使规则内容没变，Redis Key 也会变化，造成旧 Key 残留；
- 无法安全地只清理单条被删除规则的 Key。

### 10.2 改造内容

1. **业务约束**：
   - `name` 必填、非空、长度 1-128 字符（已有）；
   - `name` 在同一 `RateLimitPolicy` 内唯一（已有）；
   - `name` 字符集限制为 `[a-zA-Z0-9_-]`（新增）；
   - `name` 创建后不可修改（新增）。

2. **Redis Key 生成**：

   ```go
   RedisKey: fmt.Sprintf("RL_TPM_rlp-%d_%s", policyID, rule.Name)
   RedisKey: fmt.Sprintf("RL_RPM_rlp-%d_%s", policyID, rule.Name)
   ```

3. **更新匹配规则**：
   - 请求体中存在、数据库中也存在的 `name`：视为修改该规则（`model`、`window_minutes` 等可变更，`name` 不可变）；
   - 请求体中存在、数据库中不存在的 `name`：视为新增规则；
   - 数据库中存在、请求体中不存在的 `name`：视为删除规则，触发该规则 Redis Key 的清理。

### 10.3 兼容性

- Open API 请求/响应结构不变，用户无需感知 `name` 的内部标识作用；
- BFE 侧已优先使用控制面导出的 `redis_key` 字段，格式变更对 BFE 透明；
- 历史基于数组下标的旧 Key 会由本次新增的清理机制逐步回收。
