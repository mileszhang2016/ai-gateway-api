# Issue #106 后续：Provider keys/models 变更对 Cluster 的引用完整性影响

## 变更概述

在修复 `PATCH /providers` 修改 `instance_pool` 后 cluster 实例池未同步的问题后，进一步检查发现：provider 的 `keys` 和 `models` 字段在更新时也存在类似的同步/引用完整性隐患——删除或重命名 provider 的 key、删除 provider 的 model 会导致引用该 provider 的 cluster 配置失效或产生无效导出。

本目录记录问题分析、影响面评估与推荐修复方案，为后续实现提供设计依据。

## 关键发现

1. **Provider `keys` 删除/重命名**：cluster 通过 `llm_config.keys[].name` 引用 provider key。provider 更新后若旧 name 不存在，导出时 `AIConf.Keys` 中对应条目的 `Key` 字段为空字符串，导致请求无法正常鉴权。
2. **Provider `models` 删除**：cluster `llm_config.models` 必须是 provider `models` 的子集（`clusters.md` 约束）。provider 删除 model 后，已有 cluster 可能违反该约束，且导出仍包含无效 model。
3. **Route rules 间接影响**：route rule 引用 `(cluster, model)` 对，cluster model 越界会间接导致路由规则失效；修复 provider 层校验可同时拦截此类风险。
4. **其它字段无影响**：`model_protocols`、`tiers`、`time_zone` 均为导出时实时投影；`description`、`model_endpoint` 不影响 cluster；model-prices 允许孤儿记录（`PV-5-004` 已确认）。

## 推荐方案

在 `ProviderManager.UpdateProvider` 中新增两类引用完整性校验 hook（风格同 `ProviderDeleteChecker`）：

- `ClusterManager.ProviderKeyRefChecker`：provider `keys` 变更时，校验所有引用 cluster 的 `llm_config.keys` name 仍存在于新 provider keys 中。
- `ClusterManager.ProviderModelRefChecker`：provider `models` 变更时，校验所有引用 cluster 的 `llm_config.models` 仍是新 provider models 的子集。

任一校验失败返回 `409 Conflict`，阻止会导致 cluster 配置失效的 provider 更新。

## 涉及文件（预期）

- `ai-gateway-api/model/icluster_conf/cluster.go`（新增两个 checker）
- `ai-gateway-api/endpoints/openapi_v1/provider/update.go`（注入 hook）
- `ai-gateway-api/model/iprovider/provider.go`（如 hook 签名需调整）
- `ai-gateway-api/model/icluster_conf/cluster_test.go`（单元测试）
- `ai-gateway-api/test/integration/tests/provider/update/update_test.go` 或新目录（集成测试）

## 状态

- [x] 代码实现
- [x] 单元测试通过
- [x] 集成测试通过（目标用例通过；全量集成测试中仅有既有的 `innerapi/rate_limit_policy` 用例失败，与本变更无关）
