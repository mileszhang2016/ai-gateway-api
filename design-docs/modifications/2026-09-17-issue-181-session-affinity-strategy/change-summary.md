# Issue #181：会话亲和编译产物缺失 `strategy` 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/181](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/181)

> `CompileEppConfig()` 在生成 `session-affinity-scorer`（session-scorer）插件参数时仅写入 `sessionIdConfig`，未显式设置 `strategy`。而 llm-d-router 的该插件默认 `strategy = "encoded_endpoint_header"`（无状态：客户端回传编码端点头），并非文档所述的 `session_id`。用户配置 `session_affinity_enabled=true` + `session_affinity_header=x-session-id` 下发后，携带 `x-session-id` 的请求**不会产生会话粘性**。

实测数据（issue 报告，3 个推理后端环境）：

- 显式补 `"strategy": "session_id"` 后：同一 `session-id` 连续 12 个请求 **12/12 收敛到同一端点**；不同 `session-id` 各自收敛；无 header 时正常弃权。
- 当前编译产物（无 `strategy`）：同一 `session-id` 连续 12 个请求分布 25% / 41.7% / 33.3%，**无粘性**。

## 2. 目标

1. 编译产物显式声明 `"strategy": "session_id"`，使 `session_affinity_enabled=true` 的集群真正产生有状态会话粘性（EPP 内存维护 session→pod 映射）；
2. 代码与既有设计文档对齐（`api-define/OpenAPI接口定义/clusters.md` 与 `2026-09-08-epp-scheduling-integration` 均已写明编译产物为 `strategy=session_id`，本次**无需改设计文档，只改代码**）；
3. 补上编译期断言测试，防止回归；
4. 不影响未开启会话亲和的集群与既有导出配置的其余部分。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（本仓库）；`ai-gateway-epp` 文档同步为可选协调项 |
| 主要文件 | `model/epp_pool/compiler.go`、`model/epp_pool/compiler_test.go` |
| 数据库 | 无变更 |
| 接口契约 | OpenAPI 请求/响应字段不变（`session_affinity_enabled` / `session_affinity_header` 语义不变）；仅 InnerAPI 导出的编译后 epp_config JSON 增加一个已知字段 `strategy` |
| 数据迁移 | 无 |
| 数据兼容 | 旧导出配置无需改写；EPP 拉取新配置后自动生效 |

## 4. 最终方案概览

**编译模板显式写入 `strategy: session_id`**，与 llm-d-router 插件约定（"strategy 决定使用哪套配置"）匹配：

```go
// model/epp_pool/compiler.go（session-scorer 插件参数）
Parameters: map[string]interface{}{
    "strategy": "session_id",   // 新增：显式声明有状态会话亲和策略
    "sessionIdConfig": map[string]interface{}{
        "sources": []map[string]interface{}{
            {"header": *conf.SessionAffinityHeader},
        },
    },
},
```

同时：

- 新增包内常量（如 `sessionAffinityStrategySessionID = "session_id"`），取值与 llm-d-router `pkg/epp/framework/plugins/scheduling/scorer/sessionaffinity/session_affinity.go:41` 的 `StrategySessionID` 保持一致；
- 更新 `compiler.go` 头部示例 JSON 注释（`:22-61`）中 session-scorer 的参数形状；
- `TestCompileEppConfig_SessionAffinity` 增加 `plugin.Parameters["strategy"] == "session_id"` 断言。

> **备选方案说明：**
> - **修改 llm-d-router 插件默认值**（把默认 strategy 改为 `session_id`）：未采用。`encoded_endpoint_header` 是 llm-d 上游的通用默认（无状态、无需 EPP 维护绑定表），改动影响所有使用该插件的组合根；控制平面显式声明自己依赖的策略更可靠，也是配置即代码的最佳实践。
> - **只改文档、不改代码**：未采用。用户功能仍然不生效，与 issue 目标相悖。
> - **strategy 做成用户可配字段**：未采用。超出 issue 范围；当前产品只暴露 `session_id` 一种有状态策略，YAGNI。

## 5. 预期收益与风险

| 项目 | 说明 |
|------|------|
| 收益 | `session_affinity_enabled=true` 的集群会话亲和真正生效（有状态、按 `session_affinity_header` 粘性路由）；编译产物与设计文档一致；测试覆盖该契约 |
| 主要风险 | 无。`strategy` 是插件参数结构体的合法字段，EPP 严格解码（未知字段拒绝）不会报错；未开启会话亲和的集群编译产物完全不变 |
| 兼容性 | OpenAPI 契约不变；导出 JSON 仅新增字段，conf-agent / EPP 消费端无需改动；修复后行为从"无粘性"变为"有粘性"，正是该开关的文档语义，属于缺陷修复而非行为变更 |

## 6. 验证方式

1. `make test-model` 通过（含更新后的 `TestCompileEppConfig_SessionAffinity`）；
2. `make test-model-cover-gate` 通过；
3. 按 issue 复现步骤实测：3 后端环境开启 `session_affinity_enabled=true` + `session_affinity_header=x-session-id`，同一 `session-id` 连续 12 个请求应全部收敛到同一端点；无 header 请求分布分散（正常弃权）。

## 7. 参考文档

- [Issue #181](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/181)
- `design-docs/api-define/OpenAPI接口定义/clusters.md`（§ 会话亲和字段定义，已写明 session_id 策略）
- `design-docs/modifications/2026-09-08-epp-scheduling-integration/`（EPP 调度集成设计，api-changes.md 已写明编译产物 strategy=session_id）
- `ai-gateway-api/model/epp_pool/compiler.go`（`CompileEppConfig`）
- `llm-d-router/pkg/epp/framework/plugins/scheduling/scorer/sessionaffinity/session_affinity.go`（插件默认值与策略分发）
