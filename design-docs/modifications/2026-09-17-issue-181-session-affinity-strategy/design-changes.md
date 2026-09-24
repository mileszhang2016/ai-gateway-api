# Issue #181：会话亲和编译产物缺失 `strategy`——设计变更说明

## 1. 当前问题定位

### 1.1 编译侧（ai-gateway-api）

```text
model/epp_pool/compiler.go:220-236
    if conf.EffectiveSessionAffinityEnabled() && conf.SessionAffinityHeader != nil {
        plugins = append(plugins, &PluginConfig{
            Name: pluginNameSessionScorer,          // "session-scorer"
            Type: pluginTypeSessionScorer,          // "session-affinity-scorer"
            Parameters: map[string]interface{}{
                "sessionIdConfig": { "sources": [{ "header": <session_affinity_header> }] },
                // ↑ 缺 "strategy"：插件侧回落默认值
            },
        })
    }
```

### 1.2 插件侧（llm-d-router，EPP 内嵌）

`pkg/epp/framework/plugins/scheduling/scorer/sessionaffinity/session_affinity.go`：

| 位置 | 行为 |
|------|------|
| `:55-57` `defaultParameters()` | `strategy` 缺省时为 `encoded_endpoint_header`（无状态：响应头回传编码端点，客户端后续请求带回） |
| `:61-68` `applyDefaults()` | 仅对**选中策略**对应的配置套默认值 |
| `:125-130` `newStrategy()` | 仅当 `params.Strategy == StrategySessionID` 才构造 `sessionIDHeaderStrategy`（使用 `sessionIdConfig`）；否则一律走 `encodedEndpointHeaderStrategy`（读 `encodedEndpointHeaderConfig`，默认头 `x-session-token`） |

### 1.3 结论

编译产物写了 `sessionIdConfig`，但插件因 `strategy` 缺省而**忽略**它，实际执行 `encoded_endpoint_header` 策略：请求头 `x-session-id` 无人消费，会话亲和完全失效。这是一个"配置写了但被静默忽略"的契约不匹配缺陷。

### 1.4 测试缺口

`model/epp_pool/compiler_test.go:116-140` `TestCompileEppConfig_SessionAffinity` 只断言了 `sessionIdConfig.sources[0].header`，未断言 `strategy`，因此编译模板与插件契约的偏离没有被任何测试拦截。

### 1.5 文档现状（代码偏离文档，而非文档偏离设计）

| 文档 | 内容 | 结论 |
|------|------|------|
| `design-docs/api-define/OpenAPI接口定义/clusters.md:217` | `session_affinity_enabled` 字段说明注明 "session-affinity-scorer，**session_id 策略**……（有 binding 状态）" | 与修复方向一致，无需改 |
| `design-docs/modifications/2026-09-08-epp-scheduling-integration/api-changes.md:234` | 编译规则明确 "scorer 链追加 `session-affinity-scorer`（**strategy=session_id**，`sessionIdConfig.sources=[{header: <session_affinity_header>}]`，权重固定 1.0）" | 与修复方向一致，无需改 |
| `ai-gateway-epp/docs/zh_cn/configuration/EPP配置定义说明-epp_config.md:105`（关联仓库） | 参数形状示例同样只写了 `sessionIdConfig`，未写 `strategy` | 建议同步补 `strategy`，避免再次误导 |

## 2. 方案对比

| 方案 | 思路 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| **A. 编译产物显式写 `strategy: session_id`（推荐）** | `compiler.go` 的 session-scorer 参数补一个键 | 一行常量 + 一处 map 键，与设计文档完全一致；控制平面对自己下发的契约显式负责；不影响 llm-d 上游默认值语义 | 无 | **采用** |
| B. 改 llm-d-router 插件默认值 | `defaultParameters()` 默认改为 `session_id` | ai-gateway-api 无需改 | 影响所有 llm-d 使用方；`encoded_endpoint_header` 作为上游通用默认（无状态、免绑定表）有其合理场景；跨仓库联动，升级顺序耦合 | 未采用 |
| C. 只改文档 | 承认现状，文档改为 `encoded_endpoint_header` | 零代码改动 | 用户配置继续不生效，违背 issue 目标；`session_affinity_header` 字段将无任何意义 | 未采用 |
| D. strategy 做成用户可配字段 | epp_config 增加 `session_affinity_strategy` | 灵活性最高 | 超出 issue 范围；产品当前只暴露一种有状态策略，YAGNI | 未采用 |

## 3. 推荐方案：编译产物显式声明 `strategy: session_id`

### 3.1 编译产物契约变化

修复前：

```json
{
  "name": "session-scorer",
  "type": "session-affinity-scorer",
  "parameters": {
    "sessionIdConfig": { "sources": [ { "header": "x-session-id" } ] }
  }
}
```

修复后：

```json
{
  "name": "session-scorer",
  "type": "session-affinity-scorer",
  "parameters": {
    "strategy": "session_id",
    "sessionIdConfig": { "sources": [ { "header": "x-session-id" } ] }
  }
}
```

- `strategy` 取值为 llm-d-router `sessionaffinity.StrategySessionID`（`"session_id"`），在 ai-gateway-api 侧以包内常量固定，不引入 llm-d 依赖（与现有 "Minimal struct set mirroring the llm-d apix JSON shape (no llm-d dependency)" 的风格一致）。
- 注入条件、profile 权重（固定 1.0）、插件顺序均不变。

### 3.2 兼容性论证

1. **EPP 严格解码安全**：EPP 对插件参数采用严格 JSON 解码（未知字段拒绝），但 `strategy` 是插件 `parameters` 结构体的**合法字段**（`session_affinity.go:47`），不会被拒绝。
2. **仅影响开启会话亲和的集群**：`session_affinity_enabled` 为 false / 缺省时编译产物逐字节不变。
3. **行为变化即缺陷修复**：修复后受影响集群从"无粘性"变为"有粘性"，这正是 `session_affinity_enabled` 开关在设计文档中的既有语义，不属于破坏性变更。
4. **无需下游配合**：conf-agent 透传配置，EPP 全量拉取编译产物，无版本协商问题。

### 3.3 修复后运行时行为（对照验收）

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 携带 `x-session-id: sess-1` 的连续请求 | 按 encoded_endpoint_header 策略处理（该策略读 `x-session-token`，`x-session-id` 被忽略），分散到各端点 | `session_id` 策略生效：EPP 内存维护 sess-1→pod 绑定，请求收敛到同一端点；绑定端点摘除后自动迁移重粘 |
| 不带 session 头的请求 | 正常参与调度 | scorer 弃权（score 全 0），与其他 scorer 加权总分决定，行为不变 |
| `session_affinity_enabled=false` | 不注入 session-scorer | 不变 |

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `model/epp_pool/compiler.go` | ① 常量区新增 `sessionAffinityStrategySessionID = "session_id"`（置于插件常量块附近）；② `:224-230` session-scorer `Parameters` 增加 `"strategy": sessionAffinityStrategySessionID`；③ `:22-61` 头部示例 JSON 注释中 session-scorer 参数形状补上 `strategy` |
| `model/epp_pool/compiler_test.go` | `TestCompileEppConfig_SessionAffinity`（`:116-140`）新增断言：`plugin.Parameters["strategy"] == "session_id"` |
| `ai-gateway-epp/docs/zh_cn/configuration/EPP配置定义说明-epp_config.md`（关联仓库，可选协调项） | `:105` 参数形状示例补上 `"strategy": "session_id"` |
| `design-docs/modifications/2026-09-17-issue-181-session-affinity-strategy/` | 本方案文档 |

> 无需修改：`design-docs/api-define/`、`design-docs/sys-design/`（既有文档已正确描述 `strategy=session_id` 的编译契约）；`db_ddl.sql` / `db_ddl_sqlite.sql`；endpoints 层。

## 5. 测试计划

### 5.1 单元测试

1. **编译断言**（核心）：`TestCompileEppConfig_SessionAffinity` 增加 `strategy` 断言后，直接验证本修复；
2. **回归**：`TestCompileEppConfig_DeterministicJSON` 已覆盖 session 场景的 marshal/roundtrip，确认新增键后 JSON 仍可解码回 `EndpointPickerConfig`；
3. **关闭路径不变**：同测试用例首段（默认关闭时无 session-scorer）保持通过。

### 5.2 命令

```bash
make test-model              # model/... 全量
make test-model-cover-gate   # 覆盖率门槛 ≥70%
```

### 5.3 实测验收（对照 issue 复现步骤）

1. 配置 `balance_mode=EPP` + `epp_config={session_affinity_enabled: true, session_affinity_header: "x-session-id"}`；
2. 下发后查看导出配置，确认 session-scorer 参数含 `"strategy": "session_id"`；
3. 同一 `session-id` 连续 12 个请求应 12/12 收敛同一端点；不同 `session-id` 各自收敛；无 header 请求分散（正常弃权）。

### 5.4 验收对照（issue 验收标准）

| 验收标准 | 验证方式 |
|----------|----------|
| 编译产物显式写入 `"strategy": "session_id"` | 单测断言 + 导出配置目检 |
| 携带 `x-session-id` 的请求稳定路由到同一后端 | 5.3 实测（issue 已验证该修复有效：12/12 收敛） |
| 修复不破坏其余编译产物 | `TestCompileEppConfig_*` 全量回归 + DeterministicJSON |

## 6. 风险与缓解

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 存量 EPP 版本过旧，插件不含 `strategy` 字段 | 若 EPP 内嵌的 sessionaffinity 插件根本没有 `strategy` 字段，严格解码会拒绝 | 该场景下当前配置同样无效（issue 实测环境已含该字段并验证修复有效）；EPP 与 ai-gateway-api 同步发布即可 |
| 修复后粘性突然"出现"被误判为异常 | 开启开关的集群流量分布将发生变化（从分散变为收敛） | 这正是该开关的文档语义，属缺陷修复；发布说明中显式注明行为变化 |
| 常量值与 llm-d 上游漂移 | llm-d 未来若重命名策略值 | 常量集中定义并随 EPP 版本升级核对；单测断言锁定当前契约 |

## 7. 上线步骤

1. 合并代码，跑通 `make test-model` / `make test-model-cover-gate`；
2. 发布 ai-gateway-api；conf-agent 全量拉取后新配置自动带 `strategy`；
3. 任选一台开启会话亲和的集群按 5.3 实测验证；
4. 关闭 issue #181。
