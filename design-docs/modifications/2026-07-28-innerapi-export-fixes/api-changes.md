# InnerAPI 配置导出行为变更说明（2026-07-28）

## 1. 变更范围

本次变更为 **InnerAPI 配置导出层的热修复**，仅影响 `/inner-api/v1/configs/*` 返回的配置内容，**不改变任何 OpenAPI 接口的请求/响应格式**。

涉及端点：

| 端点 | 对应配置主题 | 变化 |
|------|-------------|------|
| `GET /inner-api/v1/configs/mod-api-key` | `mod_api_key_rule` | 精简：移除无限配额计划；修复 token 耗尽状态 |
| `GET /inner-api/v1/configs/ai-route` | `ai_route` | 扩展：API-Key 绑定列表包含祖先 Entity 路由表 |
| `GET /inner-api/v1/configs/rate-limit-policy` | `mod_ai_rate_limit` | 精简：仅导出启用状态的策略及绑定 |

---

## 2. 详细变更

### 2.1 `GET /inner-api/v1/configs/mod-api-key`

#### 变更前行为

- API-Key 自身及 Entity 层级的**无限配额计划**会被加入 `QuotaPlans` 与 `tokens[].quota_plan_ids`；
- 由于导出时未预加载 `QuotaPlan`，`GetRemainingQuota` 无法正确判断 token 是否耗尽，有限配额 token 可能被错误标记为 `exhausted`。

#### 变更后行为

- 导出前预加载每个 API-Key 关联的 `QuotaPlan`；
- `unlimited=true` 的配额计划不再加入导出结果；
- token 状态计算基于正确的 `QuotaPlan.Quota`。

#### 对下游（BFE）影响

- BFE 收到的 `QuotaPlans` 映射中不再包含无意义的无限配额计划；
- token 的 `quota_plan_ids` 仅指向有限配额计划；
- `Enabled` 字段更准确，避免正常 token 被误禁用。

---

### 2.2 `GET /inner-api/v1/configs/ai-route`

#### 变更前行为

每个 API-Key 的 `ApikeyRouteTableBindings` 仅按以下顺序绑定：

1. API-Key 自身路由表（`apikey_<key>`）
2. **直接挂载 Entity** 的路由表（`entity_<name>`）
3. Global 路由表（`global_default`）

#### 变更后行为

在 API-Key 自身路由表与 Global 路由表之间，按 Entity 层级自底向上插入所有祖先 Entity 的路由表：

1. API-Key 自身路由表（`apikey_<key>`）
2. 直接挂载 Entity 的路由表（`entity_<name>`）
3. 父 Entity 的路由表（`entity_<parent_name>`）
4. ……
5. 根 Entity 的路由表
6. Global 路由表（`global_default`）

#### 对下游（BFE）影响

- BFE 按绑定列表顺序匹配，先命中 API-Key 级规则，再依次命中各级 Entity 规则，最后命中 Global 兜底规则；
- 与配额计划、限流策略的层级继承语义保持一致。

---

### 2.3 `GET /inner-api/v1/configs/rate-limit-policy`

#### 变更前行为

所有被 API-Key 或 Entity 引用的限流策略（无论 `enabled` 状态）都会被导出，并生成绑定关系。

#### 变更后行为

仅导出 `enabled=true` 的限流策略；被禁用策略不再生成绑定关系。

#### 对下游（BFE）影响

- BFE 不再加载已禁用的限流策略；
- 导出结构中的 `Enabled` 字段固定为 `true`（因只有启用策略才会被导出）。

---

## 3. 数据字段常量变更

### 3.1 `route_rules.type` 常量

| 常量 | 旧值 | 新值 | 影响 |
|------|------|------|------|
| `RouteRulesTypeAPIKey` | `"api_key"` | `"apikey"` | 新创建的 API-Key 路由表 `type` 字段写为 `"apikey"`；AI 路由导出 key 统一为 `apikey_<key>`。 |

> 数据库中已有 `type="api_key"` 的历史记录仍可读，但建议通过迁移脚本统一为 `"apikey"`。

---

## 4. OpenAPI 接口变更

无。

---

## 5. 错误码变更

无。

---

## 6. 兼容性说明

| 维度 | 兼容性 |
|------|--------|
| OpenAPI 请求参数 | 完全兼容 |
| OpenAPI 响应结构 | 完全兼容 |
| 数据库 schema | 完全兼容 |
| InnerAPI 配置内容 | 语义修正；建议 BFE 全量拉取一次 |
| 历史 `route_rules.type="api_key"` | 数据库可读，建议迁移 |

---

## 7. 相关文档

| 文档 | 路径 |
|------|------|
| 本次设计变更说明 | `design-docs/modifications/2026-07-28-innerapi-export-fixes/design-changes.md` |
| OpenAPI v0.3.0 API 变更 | `design-docs/modifications/2026-07-27-openapi-optimize/api-changes.md` |
| InnerAPI 配置导出与版本控制 | `design-docs/sys-design/details/InnerAPI配置导出与版本控制.md` |

---

*文档生成日期：2026-07-28*
