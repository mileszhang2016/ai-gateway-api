# Issue #106 后续：Provider keys/models 变更对 Cluster 的引用完整性影响

## 1. 背景

[Issue #106](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/106) 及其修复解决了 `PATCH /providers` 修改 `instance_pool` 后，引用该 provider 的 cluster 实例池未同步的问题。修复采用 **ProviderManager 同步 hook** 机制：在 `UpdateProvider` 事务内调用 `ClusterManager.ProviderInstancePoolSyncer`，将 provider 的最新实例池刷新到所有引用 cluster 的 sub-cluster pool。

完成该修复后，进一步检查 provider 其余字段对 cluster 及导出配置的影响，发现 `keys` 和 `models` 字段在更新时存在类似的引用完整性风险。

## 2. 问题定位

### 2.1 Provider 字段与 Cluster/导出关系总览

| Provider 字段 | 是否已同步/实时投影 | 说明 |
|---------------|--------------------|------|
| `instance_pool` | ✅ 已修复 | 通过 `ProviderInstancePoolSyncer` 同步到 cluster sub-cluster pool |
| `model_protocols` | ✅ 实时投影 | `newAIConf` 导出时直接取 provider 当前 `model_protocols` |
| `keys`（仅改 key 值，name 不变） | ✅ 实时投影 | `newAIConf` 按 name 从 provider 取最新 key 值 |
| `keys`（删除/重命名 name） | ❌ **存在隐患** | cluster 仍引用旧 name，导出后对应 `AIConf.Keys[].Key` 为空字符串 |
| `models`（新增） | ✅ 无影响 | cluster 原本就是 provider models 子集，新增不破坏约束 |
| `models`（删除） | ❌ **存在隐患** | cluster `llm_config.models` 可能包含已被移除的 model，违反子集约束 |
| `tiers` / `time_zone` | ✅ 实时投影 | 通过 `providerPricingTable` 实时填充到 `ModelTable` |
| `model_endpoint` | 无影响 | 仅用于模型发现工具 `/providers/tools/discover-models` |
| `description` | 无影响 | 纯展示字段 |

### 2.2 Provider `keys` 删除/重命名导致 cluster key 悬空

#### 复现路径

1. 创建 provider：
   ```json
   {
     "name": "deepseek",
     "keys": [{"name": "k1", "key": "sk-old"}],
     "instance_pool": [...],
     "model_protocols": ["openai"]
   }
   ```
2. 创建 cluster：
   ```json
   {
     "name": "c1",
     "llm_config": {
       "provider": "deepseek",
       "models": ["deepseek-chat"],
       "keys": [{"name": "k1", "weight": 100}]
     }
   }
   ```
3. 更新 provider，将 keys 改为 `[{"name": "k2", "key": "sk-new"}]`（删除/重命名 k1）。
4. 调用 `GET /inner-api/v1/configs/tls_conf/server_data_conf` 导出。

#### 导出结果

cluster `c1` 的 `AIConf.Keys` 会变成：

```json
[{"Name": "k1", "Key": "", "Weight": 100}]
```

#### 根因

`model/icluster_conf/cluster.go:1119-1130` 的 `newAIConf` 函数：

```go
keyMap := map[string]string{}
for _, k := range providerKeys {
    keyMap[k.Name] = k.Key
}
for _, k := range llmConfig.Keys {
    name := derefString(k.Name, "")
    aiConf.Keys = append(aiConf.Keys, cluster_conf.AIKey{
        Name:   name,
        Key:    keyMap[name],  // name 不在 provider keys 中时，Key 为空字符串
        Weight: derefInt(k.Weight, 0),
    })
}
```

当 cluster 引用的 key name 已从 provider 中删除时，`keyMap[name]` 返回空字符串，导致 BFE 侧无法构造有效鉴权头。

### 2.3 Provider `models` 删除导致 cluster model 越界

#### 复现路径

1. 创建 provider：
   ```json
   {"models": ["deepseek-chat", "deepseek-coder"], ...}
   ```
2. 创建 cluster：
   ```json
   {"llm_config": {"models": ["deepseek-coder"], "provider": "deepseek"}}
   ```
3. 更新 provider：
   ```json
   {"models": ["deepseek-chat"], ...}
   ```
4. 导出配置仍包含 `deepseek-coder`。

#### 根因

- `model/icluster_conf/cluster.go:526` 的 `validateClusterLLMConfigAgainstProvider` 仅在 **cluster 创建/更新** 时校验 model 子集关系。
- `ProviderManager.UpdateProvider` 更新 provider models 时，不会重新校验引用该 provider 的 cluster。
- cluster 的 `LLMConfig.Models` 是持久化快照，provider 变更不会触发清理或校验。

### 2.4 Route rules 的间接影响

`model/route_rules/route_rules.go:226` 提供了 `ClusterModelUpdateChecker`，用于在 cluster 更新时阻止删除被路由规则引用的 model。但该 checker 仅在 cluster 更新路径触发；当 provider 删除 model 导致 cluster model 越界时，没有对应机制拦截。

因此，provider 删除 model 可能产生级联失效：

```
provider.models --(remove)--> cluster.llm_config.models 越界
                                    |
                                    v
                          route rules 引用的 (cluster, model) 失效
```

在 provider 层增加 `ProviderModelRefChecker` 可同时覆盖该级联风险。

## 3. 目标

1. 防止 provider `keys` 更新导致 cluster 引用悬空的 key name。
2. 防止 provider `models` 更新导致 cluster 引用已被移除的 model。
3. 保持与现有 `ProviderDeleteChecker` 一致的校验风格。
4. 不破坏 `instance_pool` 已修复的同步路径。

## 4. 方案对比

| 方案 | 思路 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. Provider 更新时拦截（推荐） | 在 `UpdateProvider` 事务内校验所有引用 cluster，若 key/model 引用失效则返回 409 | 与删除 provider 的引用检查一致；不会静默修改用户显式配置的 cluster | 用户需先修改 cluster 再改 provider | **采用** |
| B. 自动清理 cluster 引用 | provider 删除 key/model 时，自动从引用 cluster 的 `llm_config.keys` / `models` 中移除 | 对用户透明 | 会静默丢失用户配置；多 sub-cluster/路由规则场景难以一致处理 | 不采用 |
| C. 导出时忽略无效引用 | `newAIConf` 中跳过无效的 key/model | 不修改 provider 更新逻辑 | 导致配置与 DB 不一致，排查困难；key 为空仍可能产生请求 | 不采用 |

## 5. 推荐方案：引用完整性校验 hook

### 5.1 核心思路

复用 `ProviderManager.UpdateProvider` 的 hook 机制（已为 `instance_pool` 同步引入），新增两个**校验型 hook**：

```go
func (cm *ClusterManager) ProviderKeyRefChecker(
    ctx context.Context,
    oldProvider, newProvider *iprovider.Provider,
) error

func (cm *ClusterManager) ProviderModelRefChecker(
    ctx context.Context,
    oldProvider, newProvider *iprovider.Provider,
) error
```

在 `UpdateProvider` 中，当请求体显式包含 `keys` / `models` 且内容发生变化时调用对应 checker；任一 checker 失败则整个事务回滚，返回 `409 Conflict`。

### 5.2 `ProviderKeyRefChecker` 实现

文件：`ai-gateway-api/model/icluster_conf/cluster.go`

```go
// ProviderKeyRefChecker returns a hook that verifies all clusters referencing
// the provider still have valid llm_config.keys names after a provider update.
func (cm *ClusterManager) ProviderKeyRefChecker(ctx context.Context,
    oldProvider, newProvider *iprovider.Provider) error {

    if newProvider == nil {
        return nil
    }

    newKeyNames := map[string]bool{}
    for _, k := range newProvider.Keys {
        newKeyNames[k.Name] = true
    }

    clusters, err := cm.storager.FetchClusterList(ctx, nil)
    if err != nil {
        return err
    }

    for _, cluster := range clusters {
        if cluster.LLMConfig == nil || cluster.LLMConfig.Provider == nil {
            continue
        }
        if *cluster.LLMConfig.Provider != newProvider.Name {
            continue
        }
        for _, k := range cluster.LLMConfig.Keys {
            if k.Name == nil || *k.Name == "" {
                continue
            }
            if !newKeyNames[*k.Name] {
                return xerror.WrapConflictErrorWithMsg(
                    "provider %s key %s is referenced by cluster %s",
                    newProvider.Name, *k.Name, cluster.Name)
            }
        }
    }

    return nil
}
```

### 5.3 `ProviderModelRefChecker` 实现

文件：`ai-gateway-api/model/icluster_conf/cluster.go`

```go
// ProviderModelRefChecker returns a hook that verifies all clusters referencing
// the provider still have valid llm_config.models after a provider update.
func (cm *ClusterManager) ProviderModelRefChecker(ctx context.Context,
    oldProvider, newProvider *iprovider.Provider) error {

    if newProvider == nil {
        return nil
    }

    newModelSet := map[string]bool{}
    for _, m := range newProvider.Models {
        newModelSet[m] = true
    }

    clusters, err := cm.storager.FetchClusterList(ctx, nil)
    if err != nil {
        return err
    }

    for _, cluster := range clusters {
        if cluster.LLMConfig == nil || cluster.LLMConfig.Provider == nil {
            continue
        }
        if *cluster.LLMConfig.Provider != newProvider.Name {
            continue
        }
        for _, m := range cluster.LLMConfig.Models {
            if !newModelSet[m] {
                return xerror.WrapConflictErrorWithMsg(
                    "provider %s model %s is referenced by cluster %s",
                    newProvider.Name, m, cluster.Name)
            }
        }
    }

    return nil
}
```

### 5.4 `ProviderManager.UpdateProvider` 调用点

文件：`ai-gateway-api/endpoints/openapi_v1/provider/update.go`

```go
if err := container.ProviderManager.UpdateProvider(req.Context(), name, param,
    container.ClusterManager.ProviderInstancePoolSyncer,
    container.ClusterManager.ProviderKeyRefChecker,
    container.ClusterManager.ProviderModelRefChecker,
); err != nil {
    return nil, err
}
```

### 5.5 触发条件

在 `ProviderManager.UpdateProvider` 中，由于存储层 `toDAOParam` 会调用 `FillDefaults` 把未传的 `models`/`keys` 从 `nil` 改成空切片，因此需要在调用 `storager.UpdateProvider` 前保存原始值：

```go
origInstancePool := param.InstancePool
origKeys := param.Keys
origModels := param.Models
```

校验 hook 的触发条件为：

- `ProviderInstancePoolSyncer`：当 `origInstancePool != nil` 且与 `existing.InstancePool` 不同时触发。
- `ProviderKeyRefChecker`：当 `origKeys != nil` 且与 `existing.Keys` 不同时触发。
- `ProviderModelRefChecker`：当 `origModels != nil` 且与 `existing.Models` 不同时触发。

构建 hook 快照前，再把未变更字段恢复为 `nil`，避免 `applyProviderUpdate` 误把空切片当作“清空”处理。这样可避免无意义的遍历，同时确保用户显式修改这些字段时一定做完整性检查。

## 6. 边界与风险

| 场景 | 行为 | 说明 |
|------|------|------|
| 仅新增 provider keys/models | 通过 | 不破坏现有 cluster |
| 仅修改 provider key 值（name 不变） | 通过 | key 值导出时实时投影 |
| 删除/重命名 provider key | 若被 cluster 引用则 409 | 用户需先更新 cluster |
| 删除 provider model | 若被 cluster/route rule 引用则 409 | 同时保护路由规则 |
| provider 未被任何 cluster 引用 | 任意更新均通过 | 无约束 |
| cluster `llm_config.keys` 为空 | 不检查 key 引用 | 空数组是合法状态 |

## 7. 测试计划

### 7.1 单元测试

#### `ai-gateway-api/model/icluster_conf/cluster_test.go`

新增 `TestProviderKeyRefChecker`：

1. 构造 provider old keys `[k1, k2]`，new keys `[k2, k3]`。
2. 构造两个 cluster：一个引用 `k1`，一个不引用该 provider。
3. 调用 `ProviderKeyRefChecker`。
4. 断言：引用 `k1` 的 cluster 触发 `409 Conflict`；不引用的 cluster 无影响。

新增 `TestProviderModelRefChecker`：

1. 构造 provider old models `[m1, m2]`，new models `[m2, m3]`。
2. 构造两个 cluster：一个引用 `m1`，一个不引用该 provider。
3. 调用 `ProviderModelRefChecker`。
4. 断言：引用 `m1` 的 cluster 触发 `409 Conflict`。

#### `ai-gateway-api/model/iprovider/provider_test.go`

新增 `TestUpdateProvider_KeyRefCheckFails` 与 `TestUpdateProvider_ModelRefCheckFails`：

1. mock provider storager 返回旧 provider。
2. mock cluster storager 返回引用旧 key/model 的 cluster。
3. 调用 `UpdateProvider` 并传入对应 checker。
4. 断言返回错误，且 provider 行未更新（事务回滚）。

### 7.2 集成测试

#### `ai-gateway-api/test/integration/tests/provider/instance_pool_sync/instance_pool_sync_test.go` 或新目录

新增两个子用例：

- `PV-SYNC-1-002 删除 provider key 被 cluster 引用时返回 409`
- `PV-SYNC-1-003 删除 provider model 被 cluster 引用时返回 409`

验证 Open API 行为符合预期。

## 8. 实施状态

- [x] `model/icluster_conf/cluster.go`：新增 `ProviderKeyRefChecker`
- [x] `model/icluster_conf/cluster.go`：新增 `ProviderModelRefChecker`
- [x] `model/iprovider/provider.go`：修正 hook 触发条件，避免 `FillDefaults` 导致的误判
- [x] `endpoints/openapi_v1/provider/update.go`：注入两个 hook
- [x] `model/iprovider/provider_test.go`：新增事务回滚单元测试
- [x] `model/icluster_conf/cluster_test.go`：新增 checker 单元测试
- [x] 集成测试补充

## 9. 参考文档

- [Issue #106](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/106)
- `ai-gateway-api/design-docs/modifications/2026-08-28-issue-106-provider-instance-pool-sync/design-changes.md`
- `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/providers.md`
- `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/clusters.md`
- `ai-gateway-api/model/icluster_conf/cluster.go`
- `ai-gateway-api/model/iprovider/provider.go`
- `ai-gateway-api/endpoints/openapi_v1/provider/update.go`
