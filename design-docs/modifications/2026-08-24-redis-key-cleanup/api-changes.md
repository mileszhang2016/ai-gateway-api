# Redis Key 清理机制：API 接口变更说明

## 1. 变更范围

| 接口类型 | 变更内容 |
|----------|----------|
| OpenAPI | **无契约变更**。请求/响应字段保持不变，后台增加 Redis Key 清理逻辑。 |
| InnerAPI | **无契约变更**。`rate_limit_policy` 导出的 `RedisKey` 字段值当前仍按 `(policyID, idx)` 生成；若后续启用 `name` 作为稳定标识的长期方案，`RedisKey` 值会变为按 `(policyID, name)` 生成。 |

---

## 2. OpenAPI 无变更说明

### 2.1 `/api-keys`

- `POST /api-keys`、`DELETE /api-keys/{id}`、`PATCH /api-keys/{id}`、`PUT /api-keys/{id}` 的请求/响应字段保持不变。
- `DELETE /api-keys/{id}` 成功后，后台会额外删除该 API-Key 的 Quota Key 与 Rate-Limit Key。
- `PATCH /api-keys/{id}` / `PUT /api-keys/{id}` 中，若 `rate_limit_policy.rules` 发生规则删除/变更，后台会额外删除旧规则对应的 Rate-Limit Key。

### 2.2 `/entities`

- `POST /entities`、`DELETE /entities/{id}`、`PATCH /entities/{id}`、`PUT /entities/{id}` 的请求/响应字段保持不变。
- `DELETE /entities/{id}` 成功后，后台会额外删除该 Entity 的 Quota Key 与 Rate-Limit Key。
- `PATCH /entities/{id}` / `PUT /entities/{id}` 中，若 `rate_limit_policy.rules` 发生规则删除/变更，后台会额外删除旧规则对应的 Rate-Limit Key。

### 2.3 rate-limit 规则 `name` 约束

本次修改采用“用 `name` 作为规则稳定标识”的方案，因此需在 OpenAPI 文档中补充说明：

- `TPMConfig.name` / `RPMConfig.name` 创建后**不允许修改**；
- `name` 在同一 `RateLimitPolicy` 内唯一（已有约束）；
- `name` 字符集限制为 `[a-zA-Z0-9_-]`，确保 Redis Key 安全。

> 说明：原有 OpenAPI 已要求 `name` 必填、非空、同一 policy 内唯一。本次新增“创建后不可修改”和“字符集限制”两个约束。

---

## 3. InnerAPI 无变更说明

### 3.1 `/configs/rate-limit-policy`

响应结构不变，仍包含每条规则的 `redis_key` 字段。本次修改后 `redis_key` 值由按数组下标生成改为按规则 `name` 生成：

```text
# 变更前
RL_TPM_rlp-<policyID>_<idx>
RL_RPM_rlp-<policyID>_<idx>

# 变更后
RL_TPM_rlp-<policyID>_<name>
RL_RPM_rlp-<policyID>_<name>
```

BFE 侧已优先使用控制面导出的 `redis_key` 字段（`bfe/bfe_modules/mod_ai_rate_limit/policy_limiter.go:109-121`），因此格式变更对 BFE 透明。

### 3.2 `/configs/mod-api-key`

响应结构不变。Quota Plan 的 `RedisKey` 仍由控制面生成并下发，BFE 直接使用。

---

## 4. 向后兼容

- OpenAPI 与 InnerAPI 的请求/响应字段均无变化，旧客户端不受影响。
- 新增的 Redis Key 删除逻辑对调用方透明。
- `rate_limit_policy` 导出的 `redis_key` 格式会变化，BFE 加载新配置后会使用新 Key；历史基于数组下标的旧 Key 会由本次新增的清理机制回收。
