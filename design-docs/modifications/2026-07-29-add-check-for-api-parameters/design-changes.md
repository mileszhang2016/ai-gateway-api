# 接口参数合法性条件设计增强（2026-07-29）

## 1. 概述

### 1.1 变更背景

`design-docs/api-define/` 目录下的 OpenAPI/InnerAPI 接口定义文档，对大多数输入参数仅描述了“必填/非必填”和简单取值范围，缺乏严格的合法性条件说明。这导致：

- 前后端、测试、运维对参数边界理解不一致；
- 代码实现中的参数校验缺少统一依据，容易出现校验遗漏或标准不统一；
- 新增接口时常重复描述同类约束（如用户名、密码、集群名、模型名等），维护成本高。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-29 |
| 涉及文件 | `design-docs/api-define/OpenAPI接口定义/*.md`<br>`design-docs/api-define/InnerAPI接口定义/*.md`<br>`ai-gateway-api/endpoints/openapi_v1/route/expression_verify.go` |
| 变更类型 | 以文档增强为主，少量接口语义修正 |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **合法性条件显性化** | 每个输入参数都应在文档中明确写出可执行的合法性条件，作为后续代码校验的依据。 |
| **公共类型集中化** | 跨接口复用的参数类型（如 Hostname、IP、Port、CIDR、AIModel、RouteRule、用户名、密码、集群名等）统一收敛到 `00-common.md`，具体接口仅引用公共类型。 |
| **最小侵入** | 尽量不修改已有接口的字段名和语义；必须修改时（如明显不合理的字段命名），保持向后兼容的说明。 |
| **文档与代码一致** | 发现文档与代码实现不一致时（如 `/expression/verify` 的鉴权状态），同步修正代码。 |

---

## 2. 公共参数类型设计

### 2.1 为什么需要公共类型

在增强合法性条件的过程中，发现大量接口使用相同或相似的参数：

- `hostname`、`ip`、`port` 出现在 ALB Pool、Clusters 等模块；
- `model` 名称出现在 API-Key、Entity、Cluster、RouteRule 等模块；
- `user_name`、`password`、`token_name` 出现在 Auth 模块；
- 路由规则、限流策略、配额计划在多个资源中重复出现。

如果每个接口都重复写一遍合法性条件，会造成：

1. 同一规则在不同文件中描述不一致；
2. 修改规则时需要改动多处，容易遗漏；
3. 文档冗长，可读性差。

因此，将通用参数和复杂结构抽取为 `00-common.md` 中的公共类型，各接口通过链接引用。

### 2.2 公共类型分层

| 层级 | 类型示例 | 说明 |
|------|----------|------|
| 基础网络类型 | Hostname、IP Address、Port、CIDR | 与网络相关的原子参数，引用 RFC 规范并附加系统特殊要求。 |
| 业务原子类型 | AIModel、ClusterName、EntityTypeName、UserName、Password、TokenName | 与业务对象命名相关的原子参数，定义长度、字符集、唯一性、保留字等规则。 |
| 复杂结构类型 | RouteRule、RouteRules、QuotaPlan、RateLimitPolicy、TPMConfig、RPMConfig | 由多个字段组成的配置对象，给出完整结构说明、跨字段约束和示例。 |

### 2.3 公共类型引用方式

具体接口文档中，参数合法性条件采用如下写法：

```markdown
| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 | 合法性条件 |
|--------|------|----------|------|----------|------------|
| hostname | string | 实例主机名 | Y | - | 必填；类型为 [Hostname](./00-common.md#1-主机名hostname) |
```

这样既保留了参数在接口中的上下文，又将具体规则收敛到公共类型。

---

## 3. 关键接口语义修正

### 3.1 `clusters.md` 模型映射字段重命名

**问题**：`model_mappings` 数组元素使用 `key` / `value` 作为字段名，表意不清，容易与 Map 的键值对概念混淆。

**决策**：改为 `source_model` / `target_model`，明确表达“请求模型名 → 后端实际模型名”的映射关系。

### 3.2 `clusters.md` 实例权重 `weight=0` 语义澄清

**问题**：原文档写“为0时后端按默认值1处理”，与运维场景矛盾——有时需要主动将某实例权重置为0以切断流量。

**决策**：`weight=0` 表示该实例不接收流量；保留“同一集群至少有一个实例 `weight > 0`”的约束，避免所有实例同时被置0导致无可用后端。

### 3.3 `certificates.md` 数据模型简化

**问题**：`cert_file_name`、`key_file_name` 仅作为上传时的展示名，不参与 TLS 握手、不用于索引；`expired_date` 可以从证书内容中解析，无需用户填写。

**决策**：

- 删除 `cert_file_name`、`key_file_name` 两个字段；
- `expired_date` 保留在数据模型中，但改为只读字段，由服务端从 `cert_file_content` 自动解析。

### 3.4 `/expression/verify` 废弃标记与鉴权修正

**问题**：文档写“已废弃，无需鉴权”，但业务上仍需使用该接口校验路由表达式，且开放无鉴权存在安全风险。

**决策**：

- 文档中移除“已废弃”说明；
- 代码中 `Authorizer` 由 `nil` 改为 `iauth.FA(iauth.FeatureRoute, iauth.ActionRead)`；
- 文档明确鉴权要求为 `FeatureRoute + ActionRead`。

### 3.5 `entities.md` `block_models` 校验放宽

**问题**：黑名单中的模型名若强制要求为已配置的 `AIModel`，则无法封禁未配置或外部模型名。

**决策**：`block_models` 元素仅需为非空字符串，无需出现在 `/clusters` 的 `llm_config.models` 中；`allow_models` 仍要求为 `AIModel` 或 `"*"`。

---

## 4. InnerAPI 文档补齐

### 4.1 补齐动机

`02-interface-list.md` 列出了 9 个 InnerAPI 导出接口，但此前只有 4 个有详细文档（`mod-api-key`、`mod-body-process`、`rate-limit-policy`、`ai-route`）。缺少的接口包括：

- `/configs/tls_conf/server_data_conf`
- `/configs/gslb_data/gslb`
- `/configs/gslb_data/cluster_table`
- `/configs/protocol/server_cert_conf`
- `/configs/extra_files/{filename}`

### 4.2 补齐内容

为每个缺失接口补充：

1. 接口信息与鉴权说明；
2. 请求参数（`version`、`bfe_cluster`、路径参数等）；
3. 返回数据结构、字段说明、JSON 示例；
4. 配置未变化时的返回示例。

其中 `server-data-conf.md` 和 `cluster-table.md` 特别说明 OpenAPI 到 BFE 配置的映射关系（如 `llm_config` → `AIConf`，`weight=0` 为不接收流量）。

---

## 5. 影响范围

| 维度 | 影响 |
|------|------|
| **接口语义** | `/expression/verify` 由无鉴权改为需 `FeatureRoute` 读权限；`certificates` 创建/更新不再上传文件名和过期时间；`clusters.model_mappings` 字段重命名；`weight=0` 语义澄清。 |
| **代码实现** | 仅 `expression_verify.go` 一处代码变更；其余为文档增强，后续代码可按文档补充校验逻辑。 |
| **数据库/数据面** | 无变化；InnerAPI 导出内容本身不变，仅文档补齐。 |
| **文档协作** | 新增公共类型和跨文件引用规范，后续新增接口应优先引用公共类型。 |
| **系统级设计文档** | `sys-design` 中接口层、模型层、数据库设计、InnerAPI 导出细节已同步更新，确保与接口定义一致。 |

---

## 6. sys-design 同步说明

接口定义增强后，系统级设计文档需保持一致，避免“接口文档已更新、设计文档仍描述旧语义”的问题。本次同步遵循以下原则：

1. **接口层**：《接口层设计文档.md》已修正 `/expression/verify` 鉴权说明，并在证书、集群子包清单中补充语义变更注释；同时新增“参数合法性条件与公共类型”章节，把接口定义中的校验规范下沉到设计层。
2. **模型层**：《模型层设计文档.md》中 `Certificate`/`CertificateParam` 删除 `cert_file_name`、`key_file_name`，`expired_date` 标记为只读；集群对外模型示例与变更表同步 `source_model`/`target_model`。
3. **数据库层**：《数据库设计文档.md》中 `certificates` 表结构删除两个文件名字段，并明确 `expired_date` 的只读来源。
4. **InnerAPI 导出细节**：《InnerAPI配置导出与版本控制.md》补充 `cluster_table` 中 `Weight=0` 的语义说明。

> 说明：`sys-design` 的核心架构（三层分层、Manager/Storager 抽象、版本控制、配置导出流程）未发生变更，仅对上述受接口语义变化影响的局部进行同步。

---

## 7. 后续协作约定

1. **新增参数**必须补充 `合法性条件`，优先引用 `00-common.md` 中的公共类型。
2. **新增公共类型**若跨 2 个及以上接口复用，应加入 `00-common.md`，并更新各接口引用。
3. **修改接口语义**（如字段名、必填性、取值范围）时，同步更新 `api-changes.md`/`design-changes.md` 并评估代码影响。
4. **InnerAPI 新接口**应按现有模块文档格式独立成文件，并在 `02-interface-list.md` 与 `README.md` 中追加索引。
