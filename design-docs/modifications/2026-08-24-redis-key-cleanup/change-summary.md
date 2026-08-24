# AI Gateway Redis Key 清理机制变更摘要

## 1. 背景

当前 `ai-gateway-api` 在管理 API-Key、Entity 及其 quota-plan、rate-limit 策略时，只负责 Redis Key 的创建与更新，**没有任何地方主动删除已经失效的 Redis Key**。随着 API-Key、Entity 被删除，或 rate-limit 规则被移除/变更，Redis 中会残留大量无用 Key：

- Quota Key：`QUOTA_<api-key-token>`、`QUOTA_<entity-id>`
- Rate-Limit Key：`RL_TPM_rlp-<policyID>_<idx>`、`RL_RPM_rlp-<policyID>_<idx>`（BFE 实际使用 `default_bfe_rlp-<policyID>_<suffix>`）

这些残留 Key 无法被回收，且可能在后续被复用时导致旧额度/计数被“继承”。

## 2. 目标

- 在 API-Key / Entity 删除时，清理其 quota-plan 与 rate-limit 对应的 Redis Key。
- 在 rate-limit 规则被删除或变更时，清理旧规则对应的 Redis Key。
- 明确配额 plan 常规变更（改配额、改单位、切换 unlimited）不会删除 Quota Key，只会覆盖其值。
- 采用最简单可行的实现方式：控制面在删除/更新成功后立即执行 Redis Key 删除。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要模块 | `model/api_key`、`model/entity`、`model/quotacache`、`model/rate_limit_policy` |
| 涉及接口 | `DELETE /api-keys/{id}`、`DELETE /entities/{id}`、`PATCH /api-keys/{id}`、`PUT /api-keys/{id}`、`PATCH /entities/{id}`、`PUT /entities/{id}` |
| 接口契约 | 请求/响应字段不变，仅后台增加 Redis Key 清理逻辑 |
| 数据迁移 | 无 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 控制面立即清理 | 基于风险评估，Quota Key 提前删除只发生在 API-Key / Entity 删除场景，实际业务影响有限；Rate-Limit Key 提前删除会导致短暂限流失效，但窗口可控。为简化实现，采用立即删除，不引入延迟队列或 BFE 侧清理。 |
| Quota Key 删除范围 | API-Key 删除时只删除自身的 Quota Key；Entity 删除时只删除自身的 Quota Key；不清理父 Entity 或所属 Entity 的 Key。 |
| Rate-Limit Key 删除范围 | API-Key / Entity 删除时只删除自身直接绑定的 Rate-Limit Policy 的 Key；Entity 更新时只清理当前 Entity 自身 Policy 的 Key，不清理父 Entity。 |
| quota-plan 变更不清理 Key | 改配额、改单位、切换 unlimited 等场景调用 `ResetToQuota` / `SetRemaining` 覆盖 Key 值，不触发删除。 |
| rate-limit 规则匹配 | 将 `name` 作为 policy 内唯一且不可变的规则标识，用 `name` 替代数组下标生成 Redis Key，实现单条规则的精确清理。 |

## 5. 改动点

| 仓库 | 文件 | 修改内容 |
|------|------|----------|
| `ai-gateway-api` | `model/quotacache/quotacache.go` | `QuotaCache` 接口新增 `DeleteKeys(ctx, keys []string) error` 方法。 |
| `ai-gateway-api` | `model/quotacache/redis.go` | 实现 `DeleteKeys`，内部使用 Redis Pipeline 批量执行 `UNLINK`/`DEL`。 |
| `ai-gateway-api` | `model/api_key/api_key.go` | `DeleteAPIKey` 事务提交后，生成并删除该 API-Key 的 Quota Key 与 Rate-Limit Key。 |
| `ai-gateway-api` | `model/entity/entity_manager.go` | `DeleteEntity` 事务提交后，生成并删除该 Entity 的 Quota Key 与 Rate-Limit Key。 |
| `ai-gateway-api` | `model/api_key/api_key.go` | `UpdateAPIKey` / `FullUpdateAPIKey` 中，若 rate-limit 规则发生删除/变更，生成旧规则 Redis Key 并删除。 |
| `ai-gateway-api` | `model/entity/entity_manager.go` | `UpdateEntity` / `FullUpdateEntity` 中，若 rate-limit 规则发生删除/变更，生成旧规则 Redis Key 并删除。 |
| `ai-gateway-api` | `model/rate_limit_policy/rate_limit_policy.go` | 将规则 `name` 作为不可变标识，用 `name` 替代数组下标生成 Redis Key。 |
| `ai-gateway-api` | `lib/validate/validate.go` | 收紧 rate-limit 规则 `name` 字符集为 `[a-zA-Z0-9_-]`，确保 Redis Key 安全；更新时校验 `name` 不可修改。 |

## 6. 影响面

| 项目 | 说明 |
|------|------|
| Quota Key 删除 | API-Key / Entity 删除后，对应 `QUOTA_*` Key 被立即清理，避免内存泄漏和 Key 复用。 |
| Rate-Limit Key 删除 | 规则/policy 删除或变更后，旧 `RL_*` Key 被立即清理。在 BFE 热加载完成前，旧配置可能访问已删除 Key，导致短暂限流失效。 |
| 配额变更 | 改配额、改单位、切换 unlimited 不删除 Quota Key，只覆盖值，不影响现有请求。 |
| 兼容性 | Open API 请求/响应不变；BFE 侧无需修改。 |
| 测试 | 需补充 API-Key / Entity 删除后的 Redis Key 清理单元/集成测试。 |

