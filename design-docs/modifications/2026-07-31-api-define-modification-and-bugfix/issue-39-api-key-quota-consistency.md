# Issue #39 修复方案：API-Key 导出 BFE 配置时 UnlimitedQuota 与 QuotaPlans 冲突

## 问题描述

创建 API-Key 时如果只传 `description`：

```
POST /open-api/v1/api-keys
{"description":"test-debug"}
```

数据库中会出现：

- `api_keys.unlimited_quota = 0`
- `quota_plans.unlimited = 1`

导出给 BFE 的 `token_rule.data` 中：

- `token.UnlimitedQuota = false`
- `token.QuotaPlans = []`（因为该 api-key 的 quota plan 是 unlimited，被跳过导出）

BFE 的 `mod_ai_token_auth` 加载时校验失败：

```
if UnlimitedQuota is false, QuotaPlans must be non-empty
```

## 根因分析

对于一个 API-Key，`UnlimitedQuota` 是人工设定的标志，含义是：该 API-Key 是否要受到各级 quota plan 的约束。

- `UnlimitedQuota = false`：该 API-Key 应受 quota plan 约束。
- `UnlimitedQuota = true`：该 API-Key 不受 quota plan 约束。

`UnlimitedQuota` 不应该被系统自动修改。

但是，可能出现这样的情况：用户将 API-Key 的 `UnlimitedQuota` 设为 `false`，但该 API-Key 关联的各级 quota plan（API-Key 自身级别、Entity 继承级别）全部设置为了 `unlimited = true`。此时导出逻辑会跳过所有 unlimited 的 quota plan，导致最终的 `QuotaPlans` 列表为空。

BFE 将 `UnlimitedQuota = false` 且 `QuotaPlans` 为空视为非法配置，从而加载失败。

## 修复方案

修改 BFE 配置文件导出逻辑（`model/imods/exporter.go`）：

对于一个 API-Key，在构造 `TokenFile` 时：

```go
tokenFile := &TokenFile{
    ...
    UnlimitedQuota: one.UnlimitedQuota != nil && *one.UnlimitedQuota,
    QuotaPlans:     quotaPlanIDs,
    ...
}
```

增加以下兜底处理：

```go
// Defense: if all related quota plans are unlimited and QuotaPlans is empty,
// treat this token as unlimited to avoid BFE load failure.
if !tokenFile.UnlimitedQuota && len(tokenFile.QuotaPlans) == 0 {
    tokenFile.UnlimitedQuota = true
}
```

即：**如果 API-Key 的 `UnlimitedQuota` 为 `false`，但最终导出的 `QuotaPlans` 为空，则在输出配置中将该 API-Key 的 `UnlimitedQuota` 改为 `true`**。

这样可以在不修改用户原始 `UnlimitedQuota` 配置的前提下，保证导出的 BFE 配置始终合法。

## 影响范围

- `model/imods/exporter.go`：导出 `mod_api_key_rule` 配置时的 `TokenFile` 构造逻辑。

## 回归测试

1. 最小参数创建 API-Key（只传 `description`）：
   - 导出的 `token_rule.data` 中该 token 的 `UnlimitedQuota = true`。
   - BFE 能正常加载配置。

2. 手动设置 `unlimited_quota = false`，但所有 quota plan 均为 `unlimited = true`：
   - 导出的 `token_rule.data` 中该 token 的 `UnlimitedQuota = true`。
   - BFE 能正常加载配置。

3. 手动设置 `unlimited_quota = false`，且存在非 unlimited 的 quota plan：
   - 导出的 `token_rule.data` 中该 token 的 `UnlimitedQuota = false`，`QuotaPlans` 非空。
   - BFE 能正常加载配置。
