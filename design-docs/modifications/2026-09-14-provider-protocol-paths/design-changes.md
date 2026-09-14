# Provider 协议路径改写（protocol_paths）：ai-gateway-api 控制面设计变更说明

> 本文档只描述 `ai-gateway-api` 控制面的修改。BFE 数据面修改**已实施完毕**：`AIConf.ProtocolPaths` 扩展与 `AIConfCheck` 加载期校验（`bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`）、改写执行（`bfe_server/ai_path_rewrite.go`、`doSingleAIForward` 接入）、SC17 集成测试 6 例全绿，见 `bfe/docs/zh_cn/modifications/2026-09-14-ai-protocol-paths-rewrite/design-changes.md` 与 `bfe/docs/zh_cn/sys_design/ai_protocol_paths.md`。

## 1. 概述

### 1.1 变更背景

provider 上游路径各异且同一 provider 不同协议前缀不同（百炼 `/compatible-mode/v1` vs `/apps/anthropic`、Kimi Code `/coding/v1` vs `/coding`、火山 `/api/v3` vs `/api/coding` 等，完整调研表见总体方案 §1.1）。BFE 按 `AIConf.ProtocolPaths` 在转发时改写路径，控制面需提供该配置的载体与导出。

**语义约定**（与 BFE 公式自洽，实施时不得偏离）：

- `protocol_paths[protocol]` = 该协议官方 SDK `base_url` 的 path 部分：openai 含 `/v1` 尾（`/compatible-mode/v1`、`/api/v3`、`/coding/v1`）；anthropic 不含 `/v1`（`/apps/anthropic`、`/coding`、`/anthropic`）；
- 仅对标准入口 `/v1/...` 生效；未配置/未命中协议 = 透传；非标准入口永不改写。

### 1.2 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `model/iprovider`、存储层（provider DAO）、`model/icluster_conf`（导出）、`endpoints/openapi_v1/provider`、接口定义文档 |
| 变更类型 | Provider 新增可选字段 + 校验 + 导出透传 |
| 变更规模 | 5 处代码改动（含 3 个易漏的手工字段列举点，见 §2.4）+ DDL + 文档 + 测试 |

### 1.3 无需改动部分

- **导出框架**：`model/imods`、`model/iversion_control` 对 `AIConf` 新字段无感知（随结构体透传），conf-agent 无改动；
- **cluster 侧**：`LLMConfig`、`validateClusterLLMConfigAgainstProvider` 不变——路径配置是 provider 级能力，cluster 恒透传（与 `ModelProtocols` 同模式）；
- **密钥/配额/计费/限流/路由**：无路径分支，天然兼容。

---

## 2. 具体改动

### 2.1 `model/iprovider/provider.go`

**(a) 结构体加字段**（`Provider` :83 / `ProviderParam` :99）：

```go
// ProtocolPaths maps a model protocol to the provider's upstream base path
// (the path part of the protocol SDK base_url), e.g.
// {"openai": "/compatible-mode/v1", "anthropic": "/apps/anthropic"}.
// Empty/nil means disabled (BFE forwards request paths unchanged).
ProtocolPaths map[string]string `json:"protocol_paths"`
```

**(b) `ValidateProviderParam` 新增校验**（委托辅助函数，风格同 `validateProviderInstancePool`）：

```go
func validateProviderProtocolPaths(paths map[string]string, modelProtocols []string) error {
    declared := map[string]bool{}
    for _, p := range modelProtocols {
        declared[p] = true
    }
    for proto, base := range paths {
        if proto != "openai" && proto != "anthropic" {
            return xerror.WrapParamErrorWithMsg("protocol_paths: unsupported protocol %q (expect openai or anthropic)", proto)
        }
        if !declared[proto] {
            return xerror.WrapParamErrorWithMsg("protocol_paths: protocol %q not declared in model_protocols", proto)
        }
        // base：非空、len <= 128、"/" 开头、不以 "/" 结尾、不含 ".." / "?" / "#"
        ...
    }
}
```

**(c) `FillDefaults` 不加默认值**：nil 即未配置，与 BFE "空 = 透传" 语义一致。

### 2.2 存储层

- `storage/rdb/internal/dao/table_providers.go`：`TProvider`（:29 起）与 param 结构增 `ProtocolPaths *string \`db:"protocol_paths"\``，JSON 序列化存储——**完全照搬 `model_protocols` 列的读写模式**（`table_providers.go:37`、`:184`）；
- `storage/rdb/provider/provider.go`：Fetch/Create/Update 的字段映射补 `protocol_paths`；
- `db_ddl.sql` / `db_ddl_sqlite.sql`：`ALTER TABLE providers ADD COLUMN protocol_paths TEXT`（新装建表语句同步加列）。

### 2.3 导出链路 `model/icluster_conf/cluster.go`

- `newAIConf`（:1341）签名增 `providerProtocolPaths map[string]string`，构造 `AIConf{ProtocolPaths: providerProtocolPaths, ...}`；调用点（:1330）与 `providerProtocolTable` 同模式取 provider 新字段传入；
- `applyProviderUpdate`（`model/iprovider/provider.go:848`）增加 `if param.ProtocolPaths != nil` 合并分支（patch 语义：不传 = 不变）。

### 2.4 易漏点清单（手工列举字段的三处，必须同步，否则字段被静默清空）

| # | 位置 | 现状 | 必须做的事 |
|---|------|------|-----------|
| 1 | `model/iprovider/provider.go` `UpdatePricingTiers`（updateParam 构造） | 手工构造 `updateParam`，逐字段复制存量 provider 值 | 补 `ProtocolPaths: existing.ProtocolPaths`，否则更新 pricing tiers 会清空 `protocol_paths` |
| 2 | `model/iprovider/provider.go:848` `applyProviderUpdate` | 手工合并 patch 字段 | 增加 ProtocolPaths 分支（见 2.3） |
| 3 | `model/iprovider/provider_operation_log.go:52/70` `providerParamToMap`/`providerToMap` | 手工列举审计字段 | 补 `protocol_paths`，保证操作日志记录该字段变更 |

### 2.5 `endpoints/openapi_v1/provider/`

请求/响应模型跟随 `ProviderParam`/`Provider` 自动携带；PATCH 部分更新的语义由 2.3/2.4-2 保证。无需新增端点。

---

## 3. 文档与测试

| # | 位置 | 动作 |
|---|------|------|
| 1 | `design-docs/api-define/OpenAPI接口定义/providers.md` | `protocol_paths` 字段定义、校验规则、常见 provider 参考值表 |
| 2 | `test/integration/tests/schema/openapi/provider.go` | schema 用例补 `protocol_paths` 字段往返 |
| 3 | 集成测试 | provider create/get/update/patch：字段往返、非法 key/非法 value 拒绝、`UpdatePricingTiers` 后字段不被清空（回归 2.4-1） |
| 4 | `model/iprovider/provider_test.go` | 校验全分支（key 越界、未声明协议、value 格式）；`applyProviderUpdate` 合并语义 |
| 5 | `model/icluster_conf` 测试 | 导出内容断言 `AIConf.ProtocolPaths` 恒等于 provider 配置值 |

---

## 4. 发布顺序与回滚

- **发布顺序**：BFE 数据面**已先行发布**（本变更的前提已满足）；本变更与 `ai-gateway-web` 表单配套发布后即可配置使用。期间不存在"api 先行"窗口（字段为纯新增，api 不发则配置入口不存在）。
- **回滚**：api 回滚后 `protocol_paths` 列保留（不读即忽略），存量 provider 数据无影响；已下发 BFE 的 `AIConf.ProtocolPaths` 在 BFE 侧行为不变（BFE 独立消费该配置）。

---

## 5. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 双端白名单漂移 | BFE `AIConfCheck` 白名单 {openai, anthropic} 与本变更校验独立维护，未来新增协议需双端同步 | 与 `ModelProtocols` 同节奏；BFE 加载期拒绝未知 key 为兜底（响亮失败，非静默） |
| 手工字段列举漏加 | §2.4 三处漏加会导致 `protocol_paths` 被静默清空（更新 tiers / patch / 审计丢失） | 按清单实施；集成测试 G4 用例（更新 pricing tiers 后字段往返）固化 |
| 计费域错配 | path 决定计费（火山 `/api/v3` 按量 vs `/api/coding` 订阅），订阅 Key 配按量前缀会错扣费 | 本期靠文档标注；连通性探测（创建时鉴权校验）列控制面二期（总体方案 §9.1） |
| 模型发现端点不联动 | `model_endpoint`（模型发现）URI 不随 `protocol_paths` 自动生成，百炼 anthropic 发现需手工配 `/apps/anthropic/v1/models` | 已知 gap，列总体方案 §9.3 开放问题；discover 失败不影响转发链路 |
