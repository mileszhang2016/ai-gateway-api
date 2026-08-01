# OpenAPI 接口定义与模型对齐优化（2026-07-31）

## 1. 概述

### 1.1 背景

本次变更围绕 `design-docs/api-define/OpenAPI接口定义/clusters.md` 中 `instance_pool` 数据结构的长期不合理展开：字段命名与通用网络术语不一致（`hostname`/`ip`/`ports`）、`ports` 使用 map 结构在 AI 网关单端口场景下过度复杂、内部模型与外部接口字段混用导致响应泄露内部状态。

同时，本次变更修复了四个生产环境问题：

- **Issue #34**：`api_keys.api_key` / `api_key_tokens.api_key` 字段过长导致 MySQL utf8mb4 唯一索引超限；
- **Issue #37**：`/clusters` 系列接口返回的 `instance_pool` 字段与文档定义不一致；
- **Issue #38**：`/clusters` 创建/更新时 `instance_pool` 的 `hostname` 与 `ip` 均被限制为必填，导致只能填写其一的场景无法提交（已随 `instance_pool` 重设计一并解决）；
- **Issue #39**：API-Key 导出 BFE `token_rule.data` 时 `UnlimitedQuota` 与 `QuotaPlans` 冲突，导致 BFE 加载失败。

### 1.2 目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-31 |
| 涉及范围 | `clusters.md`、`api-keys.md`、相关代码实现、`db_ddl.sql`、`sys-design` |
| 变更类型 | 接口定义优化 + 数据模型重设计 + 数据库字段收缩 + 导出层兜底修复 |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **接口与实现一致** | 接口文档、代码实现、数据库约束三者字段名、类型、长度保持一致。 |
| **内部状态不外泄** | 响应模型与内部模型分离，避免将数据库/运行时字段暴露给调用方。 |
| **数据库约束贴合业务** | 字段长度与接口层合法性条件一致，避免索引超限或过窄截断。 |
| **向后兼容旧数据** | 数据库中旧格式 JSON 通过兼容反序列化继续可读，不强制迁移。 |
| **兜底不修改原始配置** | 导出层防御性修正只影响输出配置，不反向修改用户数据。 |

---

## 2. 设计决策

### 2.1 `instance_pool` 数据模型重设计

旧结构：

```json
{
  "hostname": "backend-1",
  "ip": "10.0.0.1",
  "weight": 50,
  "ports": { "Default": 8080 }
}
```

旧结构将 `hostname` 与 `ip` 同时设为必填，导致业务上只需填写其一的配置无法通过校验（Issue #38）；且两者语义重叠，均可能为主机名或 IP。

新结构：

```json
{
  "name": "backend-1",
  "addr": "10.0.0.1",
  "weight": 50,
  "port": 8080
}
```

| 旧字段 | 新字段 | 设计决策 |
|--------|--------|----------|
| `hostname` | `name` | 实例名称，选填，未传入时默认与 `addr` 相同；长度 1-128 字符。 |
| `ip` | `addr` | 实例地址，必填，类型为公共 `Hostname`，支持主机名或 IP。 |
| `ports` | `port` | 实例端口，必填，类型为公共 `Port`；单一整数替代原 map。 |
| - | `weight` | 权重，范围 [0,100]；`0` 表示不接收流量。 |

合法性约束：

- 至少包含 1 个实例；
- 同一集群内，`name` 非空时 `name` 不能重复；
- 同一集群内，`(name, addr)` 组合不能重复；
- 至少有一个实例 `weight > 0`。

### 2.2 内部模型与响应模型分离

为避免内部字段外泄，引入两层模型：

- **内部模型** `icluster_conf.Instance`：包含 `Name`、`Addr`、`Port`、`Weight`、`Disable`，用于数据库存储和运行时处理。
- **响应模型** `product_cluster.InstanceData`：仅包含 `Name`、`Addr`、`Port`、`Weight`，用于 `/clusters` 系列接口返回。

`clusterModel2Control` 负责将内部模型转换为响应模型，显式丢弃 `Disable`。

### 2.3 旧数据兼容

数据库中 `pools.instance_detail` 可能仍保存旧版 PascalCase JSON（`Name`、`Addr`、`Port`、`Ports`、`Weight`、`Disable`）。`Instance.UnmarshalJSON` 通过兼容结构优先读取新字段，回退读取旧字段，确保历史数据无需迁移即可继续使用。

### 2.4 数据库字段长度设计

`api_keys` 与 `api_key_tokens` 表使用 `utf8mb4` 字符集，InnoDB 单条唯一索引最大长度为 3072 字节。原 `api_key varchar(1024)` 的索引长度为 `1024 × 4 = 4096` 字节，超出限制。

`api-keys.md` 已规定 `key` 长度为 1-128 字符，因此将数据库字段统一收缩为 `varchar(128)`：`128 × 4 = 512` 字节，既满足索引约束，也与接口层校验一致。

### 2.5 导出层兜底原则

API-Key 的 `unlimited_quota` 是用户设定的“是否受配额约束”开关，而 `QuotaPlans` 是导出到 BFE 的实际配额计划列表。当用户将 `unlimited_quota=false` 但所有关联 quota plan 均为 `unlimited=true` 时，导出逻辑会跳过所有 unlimited plan，导致 `QuotaPlans=[]`。BFE 校验 `UnlimitedQuota=false & QuotaPlans=[]` 为非法配置。

决策：在导出层 `model/imods/exporter.go` 增加防御性修正，若 `!UnlimitedQuota && len(QuotaPlans)==0` 则将输出配置中的 `UnlimitedQuota` 改为 `true`。该修正不修改数据库原始值。

---

## 3. 接口变更

### 3.1 `clusters.md`

- `instance_pool` 元素字段由 `hostname/ip/ports` 改为 `name/addr/port`，同时解决原 `hostname`/`ip` 必须同时填写的校验冲突（Issue #38）。
- `basic.protocol` 默认值由 `http` 改为 `https`。
- `passive_health_check.host` 为空时默认使用 `instance_pool` 首个实例的 `addr`。

### 3.2 `api-keys.md`

- `key` 长度明确为 1-128 字符，仅允许大小写字母、数字、连字符、下划线。

---

## 4. 缺陷修复

### 4.1 Issue #34：api_key 字段索引超限

**问题**：`api_keys.api_key` / `api_key_tokens.api_key` 原定义为 `varchar(1024)`，在 `utf8mb4` 字符集下唯一索引长度超过 InnoDB 3072 字节上限，导致 MySQL DDL 导入失败。

**修复**：将两字段调整为 `varchar(128)`，与 `api-keys.md` 中 1-128 字符限制一致。

### 4.2 Issue #37：/clusters 响应字段与文档不一致

**问题**：`/clusters` 响应直接使用内部 `[]icluster_conf.Instance`，导致返回了文档未定义的 `disable` 字段，且历史字段名与当前文档不一致。

**修复**：新增响应专用结构 `InstanceData`，`ClusterData.InstancePool` 改用 `[]InstanceData`，隐藏 `disable` 字段，字段名统一为小写 `name/addr/port/weight`。

### 4.3 Issue #38：实例池 `hostname`/`ip` 必填冲突

**问题**：`/clusters` 创建与更新接口将 `instance_pool[].hostname` 与 `ip` 同时设为必填，但业务场景中实例通常只需要填写域名 **或** IP，导致只能提供其一的配置无法提交。

**修复**：随 `instance_pool` 数据模型重设计，新结构中 `name`（对应原 `hostname`）改为选填，`addr`（对应原 `ip`）改为必填。既解决了必填冲突，又统一了字段语义。

### 4.4 Issue #39：API-Key 导出 UnlimitedQuota 与 QuotaPlans 冲突

**问题**：最小参数创建 API-Key 时，`api_keys.unlimited_quota=false` 但 quota plan 为 `unlimited=true`，导出时跳过 unlimited plan 后 `QuotaPlans=[]`，BFE 加载失败。

**修复**：在 `model/imods/exporter.go` 中增加兜底逻辑：若 `UnlimitedQuota=false` 但 `QuotaPlans` 为空，则将导出配置中的 `UnlimitedQuota` 修正为 `true`，不修改数据库原始值。

详细修复方案见同目录 [issue-39-api-key-quota-consistency.md](./issue-39-api-key-quota-consistency.md)。

---

## 5. 代码实现要点

### 5.1 `endpoints/openapi_v1/product_cluster/one.go`

- 新增 `InstanceData` 响应结构；
- `ClusterData.InstancePool` 类型改为 `[]InstanceData`；
- `clusterModel2Control` 将内部 `Instance` 转换为 `InstanceData`。

### 5.2 `endpoints/openapi_v1/product_cluster/create.go`

- 请求参数 `Instance` 使用 `name/addr/port/weight`；
- `Instancesc2i` 仅映射上述四个字段；
- `normalizeBasic` 默认 `protocol=https`；
- `normalizePassiveHealthCheck` 默认 `host` 为首个实例 `addr`。

### 5.3 `model/icluster_conf/pool.go`

- `Instance` 结构改为 `name/addr/port/weight/disable`；
- 新增 `UnmarshalJSON` 兼容旧版 PascalCase 字段。

### 5.4 `lib/validate/validate.go`

- `Instance` 校验适配新字段；
- `InstancePool` 校验增加 `name` 重复、`name+addr` 重复、至少一个正权重等约束。

### 5.5 `model/imods/exporter.go`

- 增加 Issue #39 的 `UnlimitedQuota` 兜底逻辑。

### 5.6 `db_ddl.sql`

- `api_keys.api_key`：`varchar(1024)` → `varchar(128)`；
- `api_key_tokens.api_key`：`varchar(1024)` → `varchar(128)`。

---

## 6. sys-design 同步更新

| 文档 | 主要更新 |
|------|----------|
| `sys-design/数据库设计文档.md` | `api_keys` / `api_key_tokens` 的 `api_key` 字段说明补充长度限制及 InnoDB 索引背景。 |
| `sys-design/模型层设计文档.md` | Cluster 模型主要变更表补充 `instance_pool[].disable` 不对外暴露。 |
| `sys-design/接口层设计文档.md` | `/clusters` 响应 `instance_pool` 仅含 `name/addr/port/weight`；`/api-keys` 传入 `key` 长度 1-128 字符。 |
| `sys-design/details/API-Key与Entity关联及模型继承.md` | 边界情况表格补充 Issue #39 导出兜底场景说明。 |

---

## 7. 测试补充

- `endpoints/openapi_v1/product_cluster/one_test.go`：新增 `TestClusterModel2ControlInstancePoolFields`，验证响应 JSON 仅含 `name/addr/port/weight`，不含 `disable`/`hostname`/`ip`/`ports`。
- `lib/validate/validate_test.go`：`TestAPIKeyValue` 增加 128/129 字符边界用例。
- `model/imods/exporter_test.go`：Issue #39 兜底逻辑已有单元测试覆盖。
- `integration-test`：新增 `scenario-SC17-apikey-unlimited-quota-export`，通过真实 BFE + ai-gateway-api + conf-agent 进程验证 Issue #39 修复。

---

## 8. 影响范围

| 维度 | 影响 |
|------|------|
| **接口语义** | `/clusters` 请求与响应中的 `instance_pool` 统一为 `name/addr/port/weight`；解决 Issue #38 中 `hostname`/`ip` 必须同时填写导致的校验冲突；`basic.protocol` 默认 `https`；`/api-keys` 传入 `key` 长度限制与数据库字段一致。 |
| **数据模型** | 内部 `Instance` 模型字段重命名；响应模型 `InstanceData` 与内部模型分离。 |
| **数据库** | `api_keys.api_key` / `api_key_tokens.api_key` 缩短为 `varchar(128)`，解决 MySQL utf8mb4 唯一索引超限问题。 |
| **BFE 数据面** | Issue #39 修复后，最小参数创建的 API-Key 导出配置不再导致 `mod_ai_token_auth` 加载失败。 |
| **兼容性** | 数据库旧数据通过 `Instance.UnmarshalJSON` 兼容读取；ALB Pool（`/alb-pool`）保留旧字段名以保持对外兼容。 |
| **文档协作** | `sys-design` 与 `api-define` 保持一致；`modifications` 过程文档与系统现状文档分离。 |

---

## 9. 后续协作约定

1. **接口字段变更**时，同步检查内部模型、响应模型、数据库字段、合法性校验四者是否一致。
2. **新增内部字段**若不应出现在响应中，必须显式定义响应模型进行隔离。
3. **数据库字段长度**应与接口文档的合法性条件保持一致。
4. **导出层防御性逻辑**应遵循“不修改原始配置”原则，仅修正输出配置。
