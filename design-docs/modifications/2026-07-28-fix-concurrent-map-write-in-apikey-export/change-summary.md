# 修复 APIKey 规则导出中的并发 map 写入问题

## 变更背景

在测试环境中，`/inner-api/v1/config/mod-api-key/export` 接口触发以下致命错误：

```
fatal error: concurrent map writes

goroutine 1428038 [running]:
github.com/yf-networks/ai-gateway-api/model/imods.(*APIKeyRuleManager).fetchEntityQuotaPlanHierarchy(...)
        .../model/imods/exporter.go:417 +0x6f0
github.com/yf-networks/ai-gateway-api/model/imods.(*APIKeyRuleManager).fetchEntityQuotaPlanHierarchy(...)
        .../model/imods/exporter.go:429 +0x418
...
github.com/yf-networks/ai-gateway-api/model/iversion_control.(*VersionControlManager).ExportConfig.func1(...)
        .../model/iversion_control/version_control.go:88 +0x43
```

该错误导致服务崩溃，需要立即修复。

## 变更目标

消除 `APIKeyRuleManager` 在并发导出配置时的 map 写入竞争，确保多个 goroutine 同时调用 `ConfigExport` 不会触发 `fatal error: concurrent map writes`。

## 影响范围

- `model/imods/mod_api_key_rule.go`
- `model/imods/exporter.go`
- `model/imods/exporter_test.go`
- `model/imods/mocks_test.go`（为并发测试让测试 fake 具备线程安全）

## 根因分析

`APIKeyRuleManager` 结构体中包含一个共享字段：

```go
type APIKeyRuleManager struct {
    ...
    quotaPlanCache map[string][]*QuotaPlan
}
```

`APIKeyRuleGenerator` 在每次导出时执行：

```go
rlm.quotaPlanCache = make(map[string][]*QuotaPlan)
```

并在后续处理 API Key 列表及其 Entity 层级时，多次读写该字段：

- `fetchQuotaPlansWithEntityHierarchy` 写入 `rlm.quotaPlanCache`
- `fetchEntityQuotaPlanHierarchy` 递归写入 `rlm.quotaPlanCache`
- `isQuotaPlanCached` 读取 `rlm.quotaPlanCache`

当多个请求/定时任务同时调用 `ConfigExport` 时，多个 goroutine 会并发地：

1. 重置同一个 `rlm.quotaPlanCache` 字段
2. 向同一个 map 执行写操作（`map[key] = value`、`append` 触发扩容）

Go 的 map 不是线程安全的，并发写会触发 `fatal error: concurrent map writes`。

## 关键决策

### 决策 1：`quotaPlanCache` 不再作为 `APIKeyRuleManager` 的字段

`quotaPlanCache` 的生命周期应当与单次 `APIKeyRuleGenerator` 调用绑定，而不是与 `APIKeyRuleManager` 实例绑定。将其改为 `APIKeyRuleGenerator` 内部的局部变量。

### 决策 2：通过函数参数传递 cache，而非依赖共享状态

涉及 `collectedQuotaPlans` 的辅助函数签名调整如下：

- `fetchQuotaPlansWithEntityHierarchy(ctx, apiKey, productName, collectedQuotaPlans)`
- `fetchEntityQuotaPlanHierarchy(ctx, entity, productName, collectedQuotaPlans)`
- `containsQuotaPlan(collectedQuotaPlans, productName, id)`

这样每次 `ConfigExport` 调用都有自己独立的 cache，不存在跨 goroutine 竞争。

### 决策 3：保持对外接口不变

`ConfigExport` 和 `APIKeyRuleGenerator` 的公开签名不变，仅修改内部实现。调用方无需调整。

## 预期改动

### `model/imods/mod_api_key_rule.go`

从 `APIKeyRuleManager` 结构体中删除 `quotaPlanCache` 字段。

### `model/imods/exporter.go`

1. `APIKeyRuleGenerator` 内部创建局部 `collectedQuotaPlans := make(map[string][]*QuotaPlan)`
2. 将 `collectedQuotaPlans` 作为参数传递给 `fetchQuotaPlansWithEntityHierarchy`
3. `fetchQuotaPlansWithEntityHierarchy` 继续传递给 `fetchEntityQuotaPlanHierarchy`
4. `isQuotaPlanCached` 改名为 `containsQuotaPlan`，并接收 `collectedQuotaPlans` 参数
5. 构建返回的 `ModAPIKeyRuleConf` 时使用局部 `collectedQuotaPlans`

### `model/imods/exporter_test.go`

1. 更新直接访问 `m.quotaPlanCache` 的测试用例（改为局部变量 `collectedQuotaPlans`）
2. 更新直接调用 `m.isQuotaPlanCached(...)` 的测试用例（改为 `containsQuotaPlan(...)`）
3. 更新直接调用 `m.fetchQuotaPlansWithEntityHierarchy(...)` 的测试用例
4. 新增并发测试 `TestAPIKeyRuleManager_ConfigExport_Concurrent`：
   - 多个 goroutine 同时调用 `ConfigExport`
   - 使用 `runtime.GOMAXPROCS` 增加竞争概率
   - 验证不触发 panic，且每个返回的配置独立正确

### `model/imods/mocks_test.go`

为使并发测试能在 `-race` 下通过，给以下测试 fake 增加 `sync.Mutex`，保护其内部追踪切片：

- `fakeAIRouteRuleStorager`
- `fakeQuotaPlanStorager`
- `fakeEntityStorager`
- `fakeVersionControlStorager`
- `fakeRouteRulesStorager`

这些 fake 仅用于测试，生产代码不受影响。

## 兼容性说明

- 对外 API 签名不变
- 导出配置的数据结构 `ModAPIKeyRuleConf` 不变
- 单条导出的逻辑和结果不变，仅消除并发竞争

## 测试计划与结果

- [x] 运行 `go test ./model/imods/... -count=1`：通过
- [x] 运行 `go test ./model/... -count=1`：通过
- [x] 运行 `go build ./...`：通过
- [x] 新增并发测试 `TestAPIKeyRuleManager_ConfigExport_Concurrent` 在修复后稳定通过
- [ ] 运行 `go test ./model/imods/... -race`：当前 Windows 环境缺少 gcc，无法启用 cgo，因此无法执行 `-race`；建议在 Linux/CI 环境中补跑

> 说明：`-race` 依赖 cgo，本机未安装 gcc，所以未能执行。并发测试通过大量 goroutine 同时调用 `ConfigExport` 来最大化触发竞争的概率，已在修复后稳定通过。

## 替代方案分析：为什么不采用加锁

### 方案 A：对 `quotaPlanCache` 加锁

最简单的修复思路是在 `APIKeyRuleManager` 中增加一把锁：

```go
type APIKeyRuleManager struct {
    ...
    quotaPlanCache map[string][]*QuotaPlan
    quotaPlanMu    sync.Mutex
}
```

每次读写 `quotaPlanCache` 时加锁：

```go
rlm.quotaPlanMu.Lock()
if _, ok := rlm.quotaPlanCache[productName]; !ok {
    rlm.quotaPlanCache[productName] = make([]*QuotaPlan, 0)
}
rlm.quotaPlanCache[productName] = append(rlm.quotaPlanCache[productName], qp)
rlm.quotaPlanMu.Unlock()
```

### 不采用加锁方案的理由

| 问题 | 说明 |
|---|---|
| **无共享必要** | 每次 `ConfigExport` 生成的是一份完整、独立的配置，A 调用产生的 cache 对 B 调用毫无意义 |
| **串行化瓶颈** | 若锁覆盖整个 `APIKeyRuleGenerator`，所有导出调用会串行执行；导出涉及多次存储访问，串行会成为性能瓶颈 |
| **重置操作也需要锁** | `rlm.quotaPlanCache = make(map[string][]*QuotaPlan)` 替换 map 引用时，必须阻止其他 goroutine 读写旧 map，否则 A 可能写到 B 的新 map 上 |
| **掩盖设计问题** | 加锁只是让共享可变状态“安全”，但没有消除共享状态本身，后续维护仍需时刻注意锁的正确性 |

### 最终选择：方案 B（局部变量）

将 `quotaPlanCache` 改为单次 `APIKeyRuleGenerator` 调用的局部变量，通过参数传递给辅助函数。这样：

- 每个 goroutine 拥有独立的 map，从根本上消除竞争
- 无需任何锁，无串行化开销
- 生命周期清晰，一次导出结束后即释放
- 对外接口 `ConfigExport` / `APIKeyRuleGenerator` 保持不变
