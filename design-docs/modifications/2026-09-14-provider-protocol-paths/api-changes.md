# protocol_paths：API 接口变更说明

## 1. 变更范围

| 接口类型 | 变更内容 |
|----------|----------|
| OpenAPI `/providers` | 新增 `protocol_paths` 字段（create / get / update / patch / list 全链路） |
| OpenAPI `/clusters` | 无变更。`llm_config` 契约不变，不新增 cluster 级路径配置字段（路径能力保持 provider 级单一定义，cluster 恒透传） |
| InnerAPI `/configs/...` | 导出的 `AIConf` 新增 `ProtocolPaths`，恒等于所引用 provider 的 `protocol_paths` 透传值 |

---

## 2. OpenAPI 变更：`/providers`

### 2.1 新增字段 `protocol_paths`

**变更原因**：不同 provider（及同一 provider 的不同协议）上游路径前缀各异，需要按协议声明上游 base path，由 BFE 在转发时将标准入口 `/v1/...` 改写到 provider 前缀。

**字段定义**：

| 属性 | 说明 |
|------|------|
| 类型 | object（map[string]string） |
| 必填 | 否；缺省/null = 不启用（请求路径原样转发） |
| 键 | 模型协议，取值 `openai` 或 `anthropic`；必须是该 provider `model_protocols` 已声明的子集 |
| 值 | 上游 base path = 该协议官方 SDK `base_url` 的 path 部分。`openai` 含 `/v1` 尾（如 `/compatible-mode/v1`、`/api/v3`）；`anthropic` 不含 `/v1`（如 `/apps/anthropic`、`/coding`）。须 `/` 开头、不以 `/` 结尾、不含 `..`/`?`/`#`、长度 ≤ 128 |

**请求示例**（百炼形态）：

```json
{
  "name": "bailian",
  "model_protocols": ["openai", "anthropic"],
  "protocol_paths": {
    "openai": "/compatible-mode/v1",
    "anthropic": "/apps/anthropic"
  }
}
```

**响应示例**（GET /providers/{name}）：

```json
{
  "id": 42,
  "name": "bailian",
  "model_protocols": ["openai", "anthropic"],
  "protocol_paths": {
    "openai": "/compatible-mode/v1",
    "anthropic": "/apps/anthropic"
  }
}
```

**校验失败响应**（400 参数错误）示例：

```json
{ "error": "protocol_paths: unsupported protocol \"gemini\" (expect openai or anthropic)" }
```

```json
{ "error": "protocol_paths: protocol \"anthropic\" not declared in model_protocols" }
```

**PATCH 语义**：`protocol_paths` 遵循部分更新约定——请求中不显式携带该字段则保持原值；显式传 `null` 清空（恢复透传）。由 `applyProviderUpdate` 合并逻辑保证。

### 2.2 常见 provider 参考值

| provider | protocol_paths |
|----------|----------------|
| 百炼 DashScope | `{"openai": "/compatible-mode/v1", "anthropic": "/apps/anthropic"}` |
| Kimi 开放平台（api.moonshot.cn） | `{"openai": "/v1", "anthropic": "/anthropic"}` |
| Kimi Code 会员（api.kimi.com） | `{"openai": "/coding/v1", "anthropic": "/coding"}` |
| DeepSeek | `{"openai": "/v1", "anthropic": "/anthropic"}` |
| 火山方舟·按量 | `{"openai": "/api/v3", "anthropic": "/api/compatible"}` |
| 火山方舟·Coding Plan | `{"openai": "/api/coding/v3", "anthropic": "/api/coding"}` |

### 2.3 转发行为说明（BFE 侧，已实施）

- 仅标准入口路径被改写：anthropic 请求 `/v1/messages` → `{anthropic 值}/v1/messages`；openai 请求 `/v1/chat/completions` → `{openai 值}/chat/completions`；
- 未配置 `protocol_paths` 或对应协议无条目 → 原样透传（与现状行为逐字节一致）；
- 非标准入口路径（provider 原生路径、`/v10/xxx`、`/v1beta/...`）永不改写——客户端以 provider 原生路径访问的透传模式不受影响。

---

## 3. OpenAPI 无变更说明：`/clusters`

`/clusters` 的 `llm_config` 结构不变，仍通过 `llm_config.provider` 引用 provider；路径改写能力由被引用 provider 的 `protocol_paths` 表达，经导出链路透传至 `AIConf.ProtocolPaths`。

**不引入 cluster 级路径配置的决策**：单一事实来源，避免 provider/cluster 双源漂移；per-cluster 路径差异（如未来真有需求）通过拆分独立 provider 实现，与 `ModelProtocols` 的既定决策一致。

---

## 4. InnerAPI 变更说明

导出的 BFE cluster 配置中 `AIConf` 新增 `ProtocolPaths` 字段：

- 取值规则：恒等于 cluster 所引用 provider 的 `protocol_paths`（未配置时为 null/缺省，BFE 按关闭处理）；
- BFE 加载期对 `ProtocolPaths` 做 key 白名单（{openai, anthropic}）与 value 格式校验，非法配置拒绝加载（与 `ValidateProtocols` 同先例），控制面校验被绕过时由 BFE 兜底。

---

## 5. 依赖的 BFE 侧变更（已实施）

| BFE 变更 | 状态 |
|----------|------|
| `AIConf.ProtocolPaths` 扩展 + `AIConfCheck` 加载期校验 | 已实施（`bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`） |
| `rewriteUpstreamPath` / `applyAIProtocolPathRewrite`（`bfe_server/ai_path_rewrite.go`） | 已实施 |
| `doSingleAIForward` 接入（改写走私有 URL 拷贝，fallback 每 attempt 重算） | 已实施（`bfe_server/reverseproxy.go`） |
| SC17 集成测试（6 例：双协议改写 / fallback 重算 / 透传兼容） | 已通过 |
