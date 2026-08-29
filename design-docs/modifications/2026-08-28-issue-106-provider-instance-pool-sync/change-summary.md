# Issue #106：Provider instance_pool 变更后 Cluster 实例池未同步

## 变更概述

修复 `PATCH /providers/{provider_name}` 修改 `instance_pool` 后，引用该 provider 的 cluster 仍在向旧后端实例转发请求的问题。

根据 `providers.md §2.4` 的契约承诺：
> 若传入 `instance_pool` 字段，系统会自动同步更新被引用该 provider 的所有 cluster 所生成的实例池。

当前实现中该同步路径缺失，导致 cluster 创建时实例池被一次性快照固化，后续 provider 侧变更无法传播到数据平面。

## 关键变更点

1. `model/iprovider/provider.go`
   - `ProviderManager.UpdateProvider` 增加可选的同步 hook 参数，在更新 provider 的事务内触发下游同步。

2. `model/icluster_conf/cluster.go`
   - 新增 `ClusterManager.ProviderInstancePoolSyncer` 方法，遍历引用该 provider 的所有 cluster，并将其 sub-cluster 实例池重刷为 provider 最新实例池。

3. `endpoints/openapi_v1/provider/update.go`
   - 调用 `UpdateProvider` 时注入 `ClusterManager.ProviderInstancePoolSyncer` hook。

4. `model/icluster_conf/cluster.go`（顺带修复）
   - `ClusterManager.UpdateCluster` 中实例池重快照条件放宽：只要请求携带 `llm_config.provider` 即触发刷新，不再要求 provider 名称发生变化。

## 涉及文件

- `ai-gateway-api/model/iprovider/provider.go`
- `ai-gateway-api/model/icluster_conf/cluster.go`
- `ai-gateway-api/endpoints/openapi_v1/provider/update.go`
- `ai-gateway-api/model/iprovider/provider_test.go`（新增单元测试）
- `ai-gateway-api/model/icluster_conf/cluster_test.go`（新增单元测试）
- `ai-gateway-api/test/integration/tests/provider/instance_pool_sync/instance_pool_sync_test.go`（新增集成测试）
- `ai-gateway-api/test/integration/tests/provider/design.md`（更新测试用例文档）

## 状态

- [x] 代码实现
- [x] 单元测试通过（`go test ./...` 全绿）
- [x] 新增集成测试 `PV-SYNC-1-001` 通过
- [ ] 集成测试 SC2101-TC032 通过（需真实环境/集成测试门禁跑）
