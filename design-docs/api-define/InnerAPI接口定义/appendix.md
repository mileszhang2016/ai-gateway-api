# 附录

## 1. 与 OpenAPI 数据对应关系

| Inner-API 字段 | OpenAPI 字段 | 数据来源 |
|----------------|--------------|----------|
| `QuotaPlans.{product}.{id}.Unlimited` | `quota_plan.unlimited` | `quota_plans` 表 |
| `QuotaPlans.{product}.{id}.PassNoQuota` | `quota_plan.pass_when_no_enough_quota` | `quota_plans` 表 |
| `QuotaPlans.{product}.{id}.Quota` | `quota_plan.quota` | `quota_plans` 表 |
| `QuotaPlans.{product}.{id}.ResetMode` | `quota_plan.reset_period` | `quota_plans` 表 |
| `tokens.{key}.quota_plans` | API-Key / Entity 关联的配额计划 ID 列表 | `api_keys.quota_plan_id`、`entities.quota_plan_id` |
| `RateLimitPolicies.{id}.enabled` | `rate_limit_policy.enabled` | `rate_limit_policies` 表 |
| `RateLimitPolicies.{id}.rules.tpm` | `rate_limit_policy.rules.tpm_configs` | `rate_limit_policies` 表 |
| `RateLimitPolicies.{id}.rules.rpm` | `rate_limit_policy.rules.rpm_configs` | `rate_limit_policies` 表 |
| `RateLimitPolicies.{id}.rules.max_concurrency` | `rate_limit_policy.rules.max_concurrency` | `rate_limit_policies` 表 |
| `ApikeyRateLimitPolicyBindings.{api_key}` | API-Key / Entity 关联的限流策略 ID 列表 | `api_keys.rate_limit_policy_id`、`entities.rate_limit_policy_id` |
| `RouteRules.{key}.rules` | `api_keys.route_rules` / `entities.route_rules` / `global-route-rules` | `route_rules` 表 |
| `ApikeyRouteTableBindings.{api_key}` | API-Key 与 Entity 的挂载关系 | `api_keys.entity_id`、`entities.parent_id` |

