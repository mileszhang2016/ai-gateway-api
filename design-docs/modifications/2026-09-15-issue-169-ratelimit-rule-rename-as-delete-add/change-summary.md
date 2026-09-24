# Issue #169：限流规则名变更语义决策记录（不修复：改名 = 删旧规则 + 增新规则）

## 1. 问题来源与结论

[rainway-ai-gateway/ai-gateway-api/issues/169](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/169)

> 按照接口定义，rpm name 不允许修改，但实际调用 openapi 接口可以修改成功。

复现（E2E SC1302-TC004 真实环境 + 手工确认）：`PATCH /open-api/v1/api-keys/{id}` 仅提交 `rate_limit_policy`、将 RPM 规则名由旧名改为新名（其余字段不变），产品返回 HTTP 200 静默接受；而契约（当时 `00-common.md` §10/§11）写明 name「创建后不可修改」，TC004 期望 4xx。

**结论（2026-09-15 产品决策）：不修复，按"works as designed"处理。** 用户传上来的规则名与既有规则不一致时，即认为用户在**删除老规则、并添加新规则**。实现零改动（现行代码已经是这个语义），仅需把契约文档从"不可修改"订正为"无改名语义，按删旧+增新处理"，并同步修订 SC1302-TC004 用例预期（FAILED_PRODUCT → FAILED_TEST）。

## 2. 根因分析（保留备查）

### 2.1 直接原因：校验链缺失

调用链（api_key 侧）：`APIKeyUpdateProcess`（`endpoints/openapi_v1/api_key/update.go:60-62`）→ `checkUpdateAPIKey`（`checker.go:59-94`，不接触 storager）→ `validate.RateLimitPolicy`（`lib/validate/validate.go:432-531`，只有请求体入参）→ `APIKeyManager.UpdateAPIKey`（`model/api_key/api_key.go:525-628`，事务内 `:572` 取出旧策略但仅用于 `:581-583` 的 Redis key 清理，`:577` 整行覆盖）。Entity 同族（`entity_manager.go:336-357`），PUT FullUpdate 与 PATCH 汇入同一 model 出口。

「更新时校验 name 不可修改」在 8d1b310 的设计文档（`2026-08-24-redis-key-cleanup/design-changes.md:214`）中列项但未落地；全仓唯一实现的 immutable 约定是 API Key token 值（`api_key.go:548`）。

### 2.2 结构性根因：改名与"删旧+增新"在全替换语义下不可区分

`rate_limit_policy` 的 PATCH/PUT 是整对象替换，负载只含目标状态、不含操作意图。"改名"（`{a}→{b}`）与"删旧规则+增新规则"（`{a}→{b}`）产生完全相同的状态差分，仅比对新旧状态在信息上无法区分。背后是标识设计：规则没有服务端生成的稳定标识（rule id），`name` 兼任展示字段与标识字段，Redis key 基于 `(policy_id, name)`（`model/shared/rate_limit_redis_key.go:10`）。

这一根因决定了四个可选方向（评估详见本文档历史版本与 §7）：

| 方案 | 对根因问题的处理 | 成本 |
|------|------------------|------|
| 改名签名拒绝（单请求内"消失集+新增集"均非空 → 422） | 不区分，把歧义组合整体拒绝；可拆两步绕过 | 小，但治标且误伤合法删旧+增新 |
| 名集严格相等（增删改全禁） | 靠冻结规则集合杜绝改名 | 规则集合创建后不可增删，与 sys-design §8 冲突 |
| 规则 ID 改造（rule id 作标识、name 回归展示字段） | 精确区分 | 大：存储/导出/契约/前端联动 + 存量迁移 |
| **契约重定义（本决策）** | **消解问题：承认改名 = 删旧 + 增新** | **极小：代码零改动，改文档与用例预期** |

## 3. 决策语义规格：改名 = 删旧规则 + 增新规则

按 `name` 匹配新旧规则（`DiffRateLimitRedisKeys`，`model/shared/rate_limit_redis_key.go:32`）：

| 场景 | 名集变化 | 行为 |
|------|----------|------|
| 仅调参数（`model`/`window_*`/`max_*`，name 不变） | `{a} → {a}` | 规则延续，Redis Key 不变，**计数连续** |
| 改名（本 issue 场景） | `{a} → {b}` | 视为删除规则 a + 新增规则 b：a 的 Key 被清理，b 生成新 Key、**从零计数** |
| 删除规则 | `{a,b} → {a}` | a 的 Key 被清理 |
| 新增规则 | `{a} → {a,b}` | b 生成新 Key |
| 首次创建策略 | `∅ → {a}` | 全新创建 |
| 修改 `model`（name 不变） | `{a} → {a}` | Key 不变，计数连续（issue #84 修复语义，不受影响） |

实现现状已完全覆盖上表：`UpdateAPIKey` / `UpdateEntity` 整行替换策略，并在事务提交后按 name diff 清理被删规则的 Redis Key（8d1b310）。**本次无任何代码改动。**

## 4. 已落地的契约文档修订

| 文档 | 位置 | 修订内容 |
|------|------|----------|
| `design-docs/api-define/OpenAPI接口定义/00-common.md` | §10/§11 `name` 行 | 删除「创建后不可修改」，指向 §9 更新语义 |
| 同上 | §9 RateLimitPolicy | 新增"更新（PATCH/PUT）语义"段落：整对象替换、无单独改名、不一致即删旧+增新、参数调整计数连续 |
| `design-docs/sys-design/details/限流策略与导出.md` | §4.1 校验规则 | 删除「`name` 创建后不可修改」校验条款 |
| 同上 | §6.6 语义 | 「不可修改/修改规则名不被允许」改为「不提供单独改名语义：视为删旧+增新，旧 Key 清理、新规则从零计数」 |
| 同上 | §8 边界情况 | "修改规则名"行同步为无改名语义 |
| `design-docs/sys-design/details/Redis Key 清理机制.md` | §6.3.2/§6.4 | 「创建后不可修改」改为本决策语义；文件清单中未落地的"更新时校验 name 不可修改"标注不再实施 |

## 5. 测试与 issue 处理

- **SC1302-TC004**（测试侧修订，不在本仓库）：rename-rejected 用例预期由"400/422 拒绝"改为"200 接受 + 旧规则 Key 被清理 + 新规则新 Key 生效"，FAILED_PRODUCT → FAILED_TEST；same-name-update 与 rule-removal 用例预期不变。
- **回归点**：改名后旧 Key 清理（`model/api_key/api_key_test.go:395-438` 已有同构断言）；仅调参数时 Key 与计数连续；`make test-model-cover-gate` 保持通过（无代码变更）。
- **issue #169**：以"契约重定义，非产品缺陷"关闭，引用本文档。

## 6. 接受的风险

- **管理面可重置计数**：持有管理面权限者可通过"删旧+增新"使某规则计数从零开始（issue #84 同类风险的明确接受）。仅限管理面操作，数据面调用方无法利用；旧 Key 被清理而非残留，无孤儿 Key 问题。
- **审计语义**：日志中表现为规则删除 + 新增，无单独"改名"事件。如需改名审计或"不重置计数"的改名能力，须引入 rule id 标识（见 §7），单独立项。

## 7. 演进方向（不实施）

若未来产品需要 first-class 的规则操作（改名不重置计数、按规则粒度的审计、精确的增删改语义），唯一干净的解法是补标识能力：服务端生成不可变 `rule_id`，`name` 回归展示字段，Redis Key 改基于 `(policy_id, rule_id)`；存储（规则 JSON 增加 id 字段 + 存量补 id）、导出（`redis_key` 已 opaque，BFE 无感）、契约、前端联动改造，另立设计文档。

## 8. 上线检查清单

- [x] 产品决策确认：不修复，改名 = 删旧规则 + 增新规则；
- [x] `00-common.md` §9 增补更新语义，§10/§11 移除「创建后不可修改」；
- [x] `限流策略与导出.md` §4.1/§6.6/§8 同步；
- [x] `Redis Key 清理机制.md` §6.3.2/§6.4 同步；
- [ ] SC1302-TC004 rename-rejected 用例预期修订（测试负责人）；
- [ ] issue #169 关闭并引用本文档；
- [ ] 知会 ai-gateway-web：当前改名交互无需改动（语义已放行）；如需"改名不重置计数"等待 rule id 方案。

## 9. 参考文档

- [Issue #169](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/169)
- `design-docs/api-define/OpenAPI接口定义/00-common.md`（§9 限流规则配置 / §10 TPMConfig / §11 RPMConfig）
- `design-docs/sys-design/details/限流策略与导出.md`（§4.1 / §6.6 / §8）
- `design-docs/sys-design/details/Redis Key 清理机制.md`（§6.3 name 稳定标识）
- `design-docs/modifications/2026-08-24-redis-key-cleanup/design-changes.md`（§10 name 稳定标识改造）
- `design-docs/modifications/2026-08-24-issue-84-rpm-rename-reset/change-summary.md`（改名重置计数器的前史）
- `model/shared/rate_limit_redis_key.go`、`model/api_key/api_key.go:525-628`、`model/entity/entity_manager.go:255-357`
