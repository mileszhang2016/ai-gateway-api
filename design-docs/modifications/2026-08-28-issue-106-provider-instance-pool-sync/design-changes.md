# Issue #106：Provider instance_pool 变更后 Cluster 实例池未同步 — 修复设计

## 1. 问题定位

### 1.1 现象

复现步骤（来自 [Issue #106](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/106)）：

1. provider 实例池成员为 `172.19.1.222:18087` 和 `172.19.1.222:18088` 时，请求能成功转发到这两个实例。
2. 通过 `PATCH /providers/{provider_name}` 将实例池成员修改为 `172.19.1.222:18100` 后，请求仍然转发到旧的 `18087` 和 `18088`，新实例 `18100` 未生效。

集成测试 `SC2101-TC032` 因此失败，表现为 `cluster_table backends=[18087 18088] want=[18100]` 长时间不收敛。

### 1.2 根因分析

**契约存在但实现缺失。**

`design-docs/api-define/OpenAPI接口定义/providers.md` 第 334 行明确说明：

> 若传入 `instance_pool` 字段，系统会自动同步更新被引用该 provider 的所有 cluster 所生成的实例池。

但代码中不存在该同步路径：

1. **`model/iprovider/provider.go:175` 的 `ProviderManager.UpdateProvider`**
   - 仅完成参数校验、存在性校验、调用 `storager.UpdateProvider` 更新 provider 行。
   - 没有遍历引用该 provider 的 cluster，也没有更新任何实例池。

2. **`endpoints/openapi_v1/provider/update.go`**
   - 仅负责参数绑定与调用 `ProviderManager.UpdateProvider`，无补偿逻辑。

3. **`model/icluster_conf/cluster.go:699` 的 `ClusterManager.UpdateCluster`**
   - 存在实例池重快照分支，但触发条件是：
     ```go
     oldProvider != *param.LLMConfig.Provider
     ```
   - 即只有当 cluster 引用的 provider **名称发生变化** 时才刷新实例池。
   - `SC2101-TC032` 中通过 `PATCH /providers` 修改实例池时，cluster 侧并未修改 provider 名称，因此该分支不会触发。

4. **配置导出忠实反映 DB**
   - `cluster_table` 由 `model/icluster_conf/exporter.go:43` 的 `clusterTableConfGenerator` 从 `cluster.SubClusters[*].InstancePool.Instances` 生成。
   - `model/iversion_control/version_control.go:84` 采用内容签名驱动版本控制：每次导出全量重生成，内容变化才发新版本。
   - 由于 DB 中 cluster 的实例池行未变，导出永远使用创建时的快照 `[18087 18088]`。

### 1.3 关键辨析：为什么 keys 收敛了而实例没有

`server_data_conf` 的 `AIConf.Keys` 在导出时由 `newAIConf`（`cluster.go:1049`）从 provider **实时投影**，而实例池是在建 cluster 时**一次性快照**。两条数据路径不同，因此 `AIConf.Keys` 收敛不能证明实例池已同步。

## 2. 修复目标

1. 实现 `PATCH /providers` 修改 `instance_pool` 后，自动同步更新所有引用该 provider 的 cluster 的实例池。
2. 同步操作与 provider 更新处于同一事务，保证原子性。
3. 不破坏现有 OpenAPI 契约和响应语义。
4. 导出版本随 DB 内容签名自动 bump，无需额外处理。
5. 顺带修复 `UpdateCluster` 中同步条件过严的问题，使其与契约语义一致。

## 3. 方案对比

| 方案 | 思路 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. Endpoint 层顺序调用 | `UpdateProvider` 成功后，再调用 `ClusterManager` 的同步方法 | 改动最小 | 两次 `AtomExecute` 非原子；provider 更新成功但 cluster 同步失败会导致数据不一致 | 不采用 |
| B. `ProviderManager` 直接依赖 `icluster_conf` | 在 `ProviderManager` 中注入 `ClusterStorager` / `PoolStorager` | 单事务内完成 | `iprovider` 与 `icluster_conf` 互相引用，产生 Go 包循环依赖（`icluster_conf` 已导入 `iprovider`） | 不可行 |
| C. `ProviderManager` 同步 hook（推荐） | 给 `UpdateProvider` 增加可选 hook，由 `ClusterManager` 实现并注入 | 无循环依赖；复用现有 `DeleteProvider` 的 `refCheckers` 模式；同一事务内完成 | 需要 endpoint 层注入 hook | **采用** |
| D. 放宽 `UpdateCluster` 同步条件 | 只要请求携带 `llm_config.provider` 就重快照实例池 | 覆盖 cluster 侧直接修改场景 | 无法替代 provider PATCH 的主路径；TC032 走 provider PATCH 不走 cluster PATCH | 顺带采用 |

## 4. 推荐方案：ProviderManager 同步 hook

### 4.1 核心思路

在 `iprovider.ProviderManager.UpdateProvider` 中引入可选的同步 hook：

```go
func (m *ProviderManager) UpdateProvider(
    ctx context.Context,
    name string,
    param *ProviderParam,
    syncHooks ...func(ctx context.Context, oldProvider, newProvider *Provider) error,
) error
```

- 当请求体中包含 `instance_pool`（即 `param.InstancePool != nil`）且与旧 pool 不同时，在更新 provider 的事务内调用 `syncHooks`。
- hook 由 `icluster_conf.ClusterManager` 实现，内部遍历引用该 provider 的 cluster 并刷新 sub-cluster 实例池。
- 由于所有 manager 共用同一个 `TxnStoragerSingleton`，且 `NewBFEDBContext` 会在 context 中复用已存在的 `*lib.DBContext`，hook 调用与 provider 更新处于同一数据库事务。

### 4.2 `ProviderManager.UpdateProvider` 修改

文件：`ai-gateway-api/model/iprovider/provider.go`

```go
// UpdateProvider updates an existing provider.
// Optional syncHooks are invoked within the same transaction when instance_pool
// is changed, allowing downstream managers (e.g. cluster_conf) to propagate the
// new pool to clusters that reference this provider.
func (m *ProviderManager) UpdateProvider(ctx context.Context, name string,
    param *ProviderParam,
    syncHooks ...func(ctx context.Context, oldProvider, newProvider *Provider) error) error {

    if err := ValidateProviderParam(param); err != nil {
        return err
    }

    return m.txn.AtomExecute(ctx, func(ctx context.Context) error {
        existing, err := m.storager.FetchProvider(ctx, &ProviderFilter{Name: &name})
        if err != nil {
            return err
        }
        if existing == nil {
            return xerror.WrapRecordNotExist("provider")
        }

        if err := m.storager.UpdateProvider(ctx, name, param); err != nil {
            return err
        }

        // Only propagate when instance_pool is explicitly provided and changed.
        if param.InstancePool != nil && !providerInstancePoolEqual(existing.InstancePool, param.InstancePool) {
            newProvider := &Provider{
                ID:             existing.ID,
                Name:           name,
                Description:    derefString(param.Description, existing.Description),
                ModelEndpoint:  firstNonNil(param.ModelEndpoint, existing.ModelEndpoint),
                Models:         firstNonNilSlice(param.Models, existing.Models),
                Keys:           firstNonNilSlice(param.Keys, existing.Keys),
                InstancePool:   param.InstancePool,
                ModelProtocols: firstNonNilSlice(param.ModelProtocols, existing.ModelProtocols),
                TimeZone:       derefString(param.TimeZone, existing.TimeZone),
                Tiers:          firstNonNilSlice(param.Tiers, existing.Tiers),
            }
            for _, hook := range syncHooks {
                if err := hook(ctx, existing, newProvider); err != nil {
                    return err
                }
            }
        }

        return nil
    })
}
```

辅助函数：

```go
func providerInstancePoolEqual(a, b []ProviderInstance) bool {
    if len(a) != len(b) {
        return false
    }
    // Order matters because the pool is a snapshot; if the UI/CLI reorders
    // instances we should still propagate the new order.
    for i := range a {
        if a[i].Name != b[i].Name ||
            a[i].Addr != b[i].Addr ||
            a[i].Port != b[i].Port ||
            a[i].Weight != b[i].Weight ||
            a[i].Disable != b[i].Disable {
            return false
        }
    }
    return true
}
```

> 说明：实际实现时可根据项目现有 helper 函数（如 `lib.PString` 等）调整 `newProvider` 构造方式，也可直接从 storager 重新 fetch 得到最新 provider，避免字段合并的样板代码。

### 4.3 `ClusterManager.ProviderInstancePoolSyncer` 实现

文件：`ai-gateway-api/model/icluster_conf/cluster.go`

```go
// ProviderInstancePoolSyncer returns a hook that, when given the old/new provider,
// updates the instance pool of every sub-cluster that references the provider.
// It is intended to be passed to iprovider.ProviderManager.UpdateProvider.
func (cm *ClusterManager) ProviderInstancePoolSyncer(ctx context.Context,
    oldProvider, newProvider *iprovider.Provider) error {

    if newProvider == nil || len(newProvider.InstancePool) == 0 {
        return nil
    }

    clusters, err := cm.storager.FetchClusterList(ctx, nil)
    if err != nil {
        return err
    }

    newInstances := providerInstancesToClusterInstances(newProvider.InstancePool)
    for _, cluster := range clusters {
        if cluster.LLMConfig == nil || cluster.LLMConfig.Provider == nil {
            continue
        }
        if *cluster.LLMConfig.Provider != newProvider.Name {
            continue
        }

        for _, sc := range cluster.SubClusters {
            if sc.InstancePool == nil {
                continue
            }
            if err := cm.poolStorager.UpdatePool(ctx, sc.InstancePool, &PoolParam{
                Instances: newInstances,
            }); err != nil {
                return err
            }
        }
    }

    return nil
}
```

设计要点：

- 遍历方式参考 `ProviderDeleteChecker`（`cluster.go:745`）。
- 池名规则已在创建时确定为 `product.Name + "." + cluster.Name`，每个 cluster 的 pool 独立，不存在命名冲突。
- `UpdatePool` 仅更新 `Instances` 字段，不影响 pool 的元数据（`Tag`、`Role` 等）。

### 4.4 Endpoint 层注入 hook

文件：`ai-gateway-api/endpoints/openapi_v1/provider/update.go`

```go
if err := container.ProviderManager.UpdateProvider(req.Context(), name, param,
    container.ClusterManager.ProviderInstancePoolSyncer,
); err != nil {
    return nil, err
}
```

`UpdatePricingTiers`  endpoint 不应注入该 hook，因为 pricing tiers 变更不涉及实例池。

## 5. 顺带修复：放宽 `UpdateCluster` 的实例池同步条件

文件：`ai-gateway-api/model/icluster_conf/cluster.go:681`

当前逻辑：

```go
if oldProvider != *param.LLMConfig.Provider {
    // 重快照实例池
}
```

建议改为：只要 `PATCH /clusters` 请求体显式携带 `llm_config.provider`，即认为调用方希望刷新与该 provider 相关的实例池，触发重快照：

```go
if oldProvider != *param.LLMConfig.Provider || providerExplicitlyProvided {
    // 重快照实例池
}
```

具体判断方式可结合 `param.LLMConfig.Provider` 是否非空（ endpoint 层已保证引用的是存在的 provider）。该改动与 `providers.md §2.4` 的语义对齐：任何通过 provider 通道修改实例池后，cluster 都能重新同步。

> 注意：此改动是“顺带修复”，不能替代 `PATCH /providers` 主路径的修复，因为 TC032 仅修改 provider 而不修改 cluster。

## 6. 事务与一致性说明

1. `ProviderManager.UpdateProvider` 和 `ClusterManager.ProviderInstancePoolSyncer` 都通过 `m.txn.AtomExecute` 执行。
2. `AtomExecute` 内部调用 `NewBFEDBContext(ctx, lib.OpenTxn())`，该工厂函数会检查 context 是否已经是 `*lib.DBContext`；若是，则复用已有事务。
3. 因此 hook 调用与 provider 更新处于**同一事务**：cluster pool 刷新失败时整个事务回滚，不会留下 provider 已更新但 cluster 未更新的中间状态。
4. 配置导出由内容签名驱动：事务提交后，下一次导出会发现 `cluster.SubClusters[*].InstancePool.Instances` 内容变化，自动 bump 版本号并下发给 BFE。

## 7. 数据迁移

本次修复**不需要数据迁移**。数据库 schema 无变化，仅修改业务逻辑。

对于已存在的、处于“provider 新实例池与 cluster 旧实例池不一致”状态的存量数据，修复上线后：
- 可引导用户重新执行一次 `PATCH /providers/{name}` 并传入当前期望的 `instance_pool`，触发同步；
- 或在升级脚本/初始化任务中增加一次全量对齐（可选，不在本次范围内）。

## 8. 测试计划

### 8.1 单元测试

#### `ai-gateway-api/model/iprovider/provider_test.go`

新增 `TestUpdateProviderSyncsInstancePoolToClusters`：

1. 构造 mock `ProviderStorager` 返回旧 provider（含实例 `A:18087`）。
2. 构造 mock sync hook，记录是否被调用以及传入的 old/new provider。
3. 调用 `UpdateProvider(ctx, name, &ProviderParam{InstancePool: newPool})`。
4. 断言：
   - `storager.UpdateProvider` 被调用；
   - sync hook 被调用一次；
   - new provider 的 `InstancePool` 等于请求中的新 pool；
   - 当 `InstancePool` 与旧 pool 完全相同时，sync hook 不被调用。

#### `ai-gateway-api/model/icluster_conf/cluster_test.go`

新增 `TestProviderInstancePoolSyncer`：

1. 构造两个 cluster：一个引用目标 provider，一个不引用。
2. 构造 mock `ClusterStorager.FetchClusterList` 返回上述 cluster。
3. 构造 mock `PoolStorager.UpdatePool`，记录被更新的 pool 与实例列表。
4. 调用 `ProviderInstancePoolSyncer`。
5. 断言：
   - 仅引用目标 provider 的 cluster 的 sub-cluster pool 被更新；
   - 更新后的实例列表等于 `providerInstancesToClusterInstances(newProvider.InstancePool)`；
   - 不引用目标 provider 的 cluster 的 pool 未被更新。

#### `ai-gateway-api/endpoints/openapi_v1/provider/update_test.go`

新增/更新测试：

1. 验证 endpoint 调用 `ProviderManager.UpdateProvider` 时注入了 `ClusterManager.ProviderInstancePoolSyncer`。
2. 验证返回结果与之前一致（返回更新后的 provider）。

### 8.2 集成测试

运行 `integration-test` 中的端到端用例：

```bash
cd integration-test
go test ./test-cases/implementation/scenario-SC2101-provider-instance-pool-sync/... -run TestTC032 -v
```

预期：

- `PATCH /providers` 返回 200；
- 随后 `GET /inner-api/v1/configs/tls_conf/cluster_table` 中对应 cluster 的 backends 收敛为 `[18100]`；
- BFE reload settle 后请求转发到新实例。

## 9. 风险与缓解

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 同步失败导致 provider 更新回滚 | hook 与 provider 更新在同一事务 | 这是期望行为，保证一致性；失败时返回错误给调用方 |
| 大量 cluster 引用同一 provider 时同步耗时 | 每次 provider 更新需遍历全量 cluster 并更新 pool | provider/cluster 数量在控制平面通常可控；如后续出现性能问题，可增加按 provider 索引的 cluster 查询 |
| 误更新非 provider 派生的 pool | `ProviderInstancePoolSyncer` 仅更新 `LLMConfig.Provider == provider.Name` 的 cluster 的 sub-cluster pool | 创建 cluster 时 provider 派生的 pool 与 cluster 绑定，逻辑一致 |
| UpdateProvider 签名变更影响现有调用 | 使用变长参数 `...func` 保持向后兼容 | 除 endpoint 外无其他调用方需要修改 |

## 10. 实施状态

- [x] `model/iprovider/provider.go`：`UpdateProvider` 增加同步 hook 支持
- [x] `model/icluster_conf/cluster.go`：新增 `ProviderInstancePoolSyncer`
- [x] `endpoints/openapi_v1/provider/update.go`：注入 hook
- [x] `model/icluster_conf/cluster.go`：放宽 `UpdateCluster` 实例池同步条件
- [x] 单元测试覆盖（`go test ./...` 全绿）
- [x] 新增集成测试 `PV-SYNC-1-001`：Open API 修改 provider instance_pool 后 Inner API `cluster_table` 同步更新
- [ ] 集成测试 SC2101-TC032 通过（需真实环境/集成测试门禁跑）

## 11. 参考文档

- [Issue #106](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/106)
- `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/providers.md`
- `ai-gateway-api/model/iprovider/provider.go`
- `ai-gateway-api/model/icluster_conf/cluster.go`
- `ai-gateway-api/model/icluster_conf/pool.go`
- `ai-gateway-api/endpoints/openapi_v1/provider/update.go`
- `ai-gateway-api/stateful/config_database.go`
- `ai-gateway-api/lib/xdb.go`
