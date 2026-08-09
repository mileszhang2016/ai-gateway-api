# TODO：BFE 升级后切回 BFE 原生 `cluster_conf.AIConf`

## 背景

当前 `ai-gateway-api` 为了实现 Cluster 多 API-Key（`Keys` + `KeyPolicy`）的 InnerAPI 导出，临时在 `model/icluster_conf/server_data_conf.go` 中定义了以下自定义结构：

- `ServerDataAIConf`
- `ServerDataClusterConf`
- `ServerDataBfeClusterConf`

这些自定义结构复用了 BFE `cluster_conf` 中的子配置（`BackendBasic`、`BackendCheck`、`GslbBasicConf`、`ClusterBasicConf`、`BackendHTTPS`），仅重写了 `AIConf` 字段，以绕过当前 BFE `cluster_conf.AIConf` 不支持多 Key 的限制。

等 BFE 完成升级后，应删除这些临时结构，切回 BFE 原生定义。

---

## BFE 侧需完成的变更

1. 在 `bfe/bfe_config/bfe_cluster_conf/cluster_conf_load.go` 的 `AIConf` 结构体中：
   - 删除旧字段 `Key`（或保留为兼容字段，但 `ai-gateway-api` 不再使用）。
   - 新增 `Keys []AIKey` 字段。
   - 新增 `KeyPolicy *AIKeyPolicy` 字段（可选）。
   - 示例定义：

     ```go
     type AIKey struct {
         Name   string `json:"Name"`
         Key    string `json:"Key"`
         Weight int    `json:"Weight"`
     }

     type AIKeyPolicy struct {
         Strategy            string `json:"Strategy"`
         MaxRetries          int    `json:"MaxRetries"`
         RetryBackoffInitial int    `json:"RetryBackoffInitial"`
         RetryBackoffMax     int    `json:"RetryBackoffMax"`
     }

     type AIConf struct {
         Type         int                `json:"Type"`
         ModelMapping *map[string]string `json:"ModelMapping"`
         Keys         []AIKey            `json:"Keys"`
         KeyPolicy    *AIKeyPolicy       `json:"KeyPolicy"`
     }
     ```

2. 确保 BFE 的 JSON 序列化/反序列化、配置校验、运行时 Key 选择逻辑均支持新的 `Keys`/`KeyPolicy`。

3. 升级 `ai-gateway-api` 所依赖的 BFE 版本（更新 `go.mod` / `go.sum`）。

---

## ai-gateway-api 侧需完成的变更

### 1. 删除临时自定义导出结构

删除文件 `model/icluster_conf/server_data_conf.go`。

如果 BFE 的 `AIConf` 字段名与当前自定义结构不一致，需要相应调整映射代码。

### 2. 修改 `model/icluster_conf/cluster.go`

- `NewBfeClusterConf` 函数：
  - 返回类型从 `*ServerDataBfeClusterConf` 改回 `*cluster_conf.BfeClusterConf`。
  - 删除对 `ServerDataAIConf`、`ServerDataClusterConf`、`ServerDataBfeClusterConf` 的引用。
  - 将多 Key 数据直接填充到 `cluster_conf.AIConf` 的 `Keys`/`KeyPolicy` 字段。

- `LLMConfig` 模型中的 `Keys`/`KeyPolicy` 类型通常可保持不变（它们已经映射到自定义结构，后续只需把映射目标改为 BFE 原生结构）。

### 3. 修改 `model/iroute_conf/exporter.go`

- `RouteRuleExportData` 结构体中的 `ClusterConf` 字段：
  - 类型从 `*ServerDataBfeClusterConf` 改回 `*cluster_conf.BfeClusterConf`。

### 4. 更新单元测试

- `model/icluster_conf/cluster_test.go`：
  - 更新 `TestNewBfeClusterConf` 中关于 LLM 配置导出的断言。
  - 断言类型改回 `*cluster_conf.BfeClusterConf`。

- `model/iroute_conf/exporter_test.go`：
  - 更新与 `ClusterConf` 类型相关的断言。

### 5. 回归验证

- `go build ./...`
- `go test ./...`
- 重点验证 InnerAPI `/server_data_conf` 导出的 JSON 结构与 BFE 新定义兼容，且 MD5 签名逻辑一致。

---

## 兼容性说明

- 当前阶段：`ai-gateway-api` 输出的 `server_data_conf` 中 `AIConf.Keys`/`AIConf.KeyPolicy` 为自定义格式；BFE 若未升级则无法消费这些字段。
- 切回 BFE 原生结构后：应确保 BFE 新 `AIConf` 的 JSON 字段名（`Keys`、`KeyPolicy`）与当前自定义输出保持一致，避免已下发的配置文件格式发生不兼容变化。
- 历史数据：`clusters.llm_config` 中旧的单 `key` 数据在读取后 `Keys` 为空；切换 BFE 结构不影响该行为。

---

## 优先级

- **P1**：BFE 侧完成 `AIConf` 升级并发布新版本。
- **P2**：`ai-gateway-api` 删除临时结构并切回 BFE 原生结构。
- **P3**：回归测试与线上灰度验证。
