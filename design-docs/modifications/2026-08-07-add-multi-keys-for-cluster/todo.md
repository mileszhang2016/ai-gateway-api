# TODO：BFE 升级后切回 BFE 原生 `cluster_conf.AIConf`

## 状态

**已完成**。BFE `v1.8.5` 已包含原生 `AIConf.Keys` / `AIConf.KeyPolicy` 支持，`ai-gateway-api` 已删除临时自定义结构并切回 BFE 原生定义。

---

## 背景

此前 `ai-gateway-api` 为了实现 Cluster 多 API-Key（`Keys` + `KeyPolicy`）的 InnerAPI 导出，临时在 `model/icluster_conf/server_data_conf.go` 中定义了以下自定义结构：

- `ServerDataAIConf`
- `ServerDataClusterConf`
- `ServerDataBfeClusterConf`

这些自定义结构复用了 BFE `cluster_conf` 中的子配置，仅重写了 `AIConf` 字段，以绕过当时 BFE `cluster_conf.AIConf` 不支持多 Key 的限制。

---

## 已完成的变更

### BFE 侧（已完成并推送至 `v1.8.5`）

- `bfe/bfe_config/bfe_cluster_conf/cluster_conf_load.go` 中的 `AIConf` 已新增：
  - `Keys []AIKey`
  - `KeyPolicy *AIKeyPolicy`
  - `ModelTable *ModelTable`

### ai-gateway-api 侧（已完成并推送至 `v0.0.7`）

1. **删除临时自定义导出结构**
   - 删除文件 `model/icluster_conf/server_data_conf.go`。

2. **修改 `model/icluster_conf/cluster.go`**
   - `NewBfeClusterConf` 返回类型改为 `*cluster_conf.BfeClusterConf`。
   - 内部使用 `cluster_conf.ClusterConf`、`cluster_conf.AIConf`、`cluster_conf.AIKey`、`cluster_conf.AIKeyPolicy`、`cluster_conf.ModelTable`、`cluster_conf.ModelPrice`。
   - `newServerDataAIConf` 重命名为 `newAIConf`，返回 `*cluster_conf.AIConf`。

3. **修改 `model/iroute_conf/exporter.go`**
   - `RouteRuleExportData.ClusterConf` 类型改为 `*cluster_conf.BfeClusterConf`。

4. **更新单元测试**
   - `model/icluster_conf/cluster_test.go`
   - `model/iroute_conf/exporter_test.go`

5. **升级 BFE 依赖**
   - `go.mod` / `go.sum` 已更新到包含新 `AIConf` 的 BFE 版本。

6. **修复 mock 接口**
   - `stateful/mock_redis.go` 补充 `NewScript` 方法，以满足 BFE `v1.8.5` 新增的 `redis_client.Client` 接口。

---

## 回归验证

已执行并通过：

```bash
cd ai-gateway-api
go build ./...
go test ./model/icluster_conf/... ./model/iroute_conf/...
```

---

## 兼容性说明

- BFE 原生 `AIConf` 的 JSON 字段名（`Keys`、`KeyPolicy`、`ModelTable` 等）与此前自定义输出保持一致。
- 历史数据：`clusters.llm_config` 中旧的单 `key` 数据在读取后 `Keys` 为空；切换 BFE 结构不影响该行为。
