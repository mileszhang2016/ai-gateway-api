# OpenAPI / InnerAPI 接口参数合法性条件增强（2026-07-29）

## 1. 变更范围

本次变更为 **接口定义文档增强**，主要对 `design-docs/api-define/` 下各接口的输入参数补充严格的合法性条件说明，并在 `00-common.md` 中集中定义可复用的公共参数类型。补充后的合法性条件将用于后续增强 API 代码实现中的参数校验。

少量伴随变更：

- `certificates.md` 数据模型简化；
- `clusters.md` 中模型映射字段重命名；
- `expression/verify` 接口移除废弃标记并恢复鉴权。

---

## 2. 公共参数类型扩展

在 `OpenAPI接口定义/00-common.md` 中新增/强化了以下公共类型：

| 序号 | 公共类型 | 核心合法性条件 |
|------|----------|----------------|
| 1 | Hostname | 长度 ≥2；符合 RFC 1123 / RFC 5890 主机名规范 |
| 2 | IP Address | 符合 RFC 791（IPv4）或 RFC 8200（IPv6）规范的合法地址 |
| 3 | Port | 1-65535 的整数 |
| 4 | CIDR | 符合 RFC 4632 / RFC 4291 的 IPv4/IPv6 CIDR 表示 |
| 5 | AIModel | `"*"` 或 `/clusters` 中某集群 `llm_config.models` 已配置的模型名 |
| 6 | RouteRule | 复杂结构，详见文档；新增 `(ClusterName, Model)` 在 `targets` 内唯一等约束 |
| 7 | RouteRules | `rules` 数组，元素为 RouteRule |
| 8 | QuotaPlan | 复杂结构，详见文档 |
| 9 | RateLimitPolicy | 新增 `tpm`/`rpm` 跨元素唯一性约束 |
| 10 | TPMConfig | 复杂结构，详见文档 |
| 11 | RPMConfig | 复杂结构，详见文档 |
| 12 | UserName | 长度 1-64；字母/数字/`_`/`-`/`.`；首尾不能为 `.`/`-`/`_`；全局唯一（大小写不敏感）；禁止保留名 |
| 13 | Password | 长度 8-128；不能含空白字符；不能等于对应 `user_name` 或其逆序 |
| 14 | TokenName | 长度 1-64；字母/数字/`_`/`-`/`.`；首尾不能为 `.`/`-`/`_`；全局唯一；禁止保留名 |
| 15 | ClusterName | 长度 1-64；字母/数字/`_`/`-`/`.`；首尾不能为 `.`/`-`/`_`；全局唯一 |
| 16 | EntityTypeName | 长度 1-32；小写字母/数字/`_`/`-`；首尾不能为 `-`/`_`；全局唯一 |

---

## 3. OpenAPI 接口文档变更

### 3.1 共性增强

各接口 Body / URI / Query 参数表统一增加或细化了 `合法性条件` 列，复杂公共类型统一引用 `00-common.md` 中的定义，避免重复描述。

### 3.2 按接口变更

| 文件 | 主要变更 |
|------|----------|
| `api-keys.md` | `key` 长度 1-128；`models`/`TPMConfig.model`/`RPMConfig.model` 引用 `AIModel`；`subnet` 引用 `CIDR`；`quota_plan`/`rate_limit_policy`/`route_rules` 引用公共类型；示例中 `route_rules.rules` 给出具体规则。 |
| `entities.md` | `name` 增加长度/控制字符/空白校验；`type` 引用 `EntityTypeName`；`allow_models` 仍引用 `AIModel`；`block_models` 放宽为任意非空字符串（无需是已配置模型）。 |
| `entity-types.md` | `type_name` 引用 `EntityTypeName`；`description`、`level` 增加详细合法性条件；URI 中的 `type_name` 增加存在性校验。 |
| `global-route-rules.md` | `rules` 引用公共 `RouteRule` 类型，删除本地重复结构。 |
| `route-tables.md` | 同 `global-route-rules.md`，引用公共 `RouteRule`。 |
| `alb-pool.md` | `instances[].hostname`/`ip`/`ports` 分别引用 `Hostname`/`IP Address`/`Port` 公共类型。 |
| `auth.md` | `user_name`/`password` 提取为公共类型 `UserName`/`Password`；`name`（Token）提取为公共类型 `TokenName`。 |
| `certificates.md` | 删除冗余字段 `cert_file_name`、`key_file_name`；`expired_date` 改为只读字段，由服务端从 `cert_file_content` 解析；增强 `cert_name`、文件内容、时间格式等校验。 |
| `clusters.md` | `name` 引用 `ClusterName`；增强 `description`、`instance_pool`、各 `basic` 子字段、`sticky_sessions`、`passive_health_check`、`llm_config` 的校验；`model_mappings` 字段由 `key`/`value` 重命名为 `source_model`/`target_model`；`weight=0` 明确为不接收流量。 |
| `expression-verify.md` | 移除“已废弃、无需鉴权”说明；明确鉴权为 `FeatureRoute + ActionRead`。 |

---

## 4. 代码变更

### 4.1 `endpoints/openapi_v1/route/expression_verify.go`

- 移除 `// Deprecated` 注释；
- `Authorizer` 由 `nil` 改为 `iauth.FA(iauth.FeatureRoute, iauth.ActionRead)`；
- 调用 `/expression/verify` 时需提供具有 `FeatureRoute` 读权限的凭证。

---

## 5. InnerAPI 接口文档补齐

在 `InnerAPI接口定义/` 目录下新增 5 份接口文档，并在 `02-interface-list.md` 与 `README.md` 中增加索引链接：

| 新增文档 | 对应端点 | 说明 |
|----------|----------|------|
| `server-data-conf.md` | `/configs/tls_conf/server_data_conf` | 导出 HostTable、RouteTable、ClusterConf 综合配置；说明 `llm_config` 导出为集群内 `AIConf` |
| `gslb.md` | `/configs/gslb_data/gslb` | 导出 GSLB 调度配置 |
| `cluster-table.md` | `/configs/gslb_data/cluster_table` | 导出集群后端实例表；说明 `weight=0` 为不接收流量 |
| `server-cert-conf.md` | `/configs/protocol/server_cert_conf` | 导出证书路径映射配置 |
| `extra-files.md` | `/configs/extra_files/{filename}` | 导出证书/密钥等原始文件内容 |

---

## 6. sys-design 同步更新

为保持 `design-docs/sys-design/` 与接口定义一致，同步调整了以下系统级设计文档：

| 文档 | 主要更新 |
|------|----------|
| `sys-design/接口层设计文档.md` | 修正 `/expression/verify` 鉴权说明为 `FeatureRoute + ActionRead`；在证书、集群子包清单中补充语义变更说明；新增 4.4 节“参数合法性条件与公共类型”，说明校验分层、公共类型引用及代码实现建议。 |
| `sys-design/模型层设计文档.md` | `Certificate`/`CertificateParam` 删除 `cert_file_name`、`key_file_name`，`expired_date` 标记为只读；集群对外模型示例中 `model_mappings` 字段改为 `source_model`/`target_model`；主要变更表新增 `llm_config.model_mappings[].key/value` 重命名说明。 |
| `sys-design/数据库设计文档.md` | `certificates` 表删除 `cert_file_name`、`key_file_name` 字段；`expired_date` 增加“从证书内容解析、只读”注释。 |
| `sys-design/details/InnerAPI配置导出与版本控制.md` | 在 `cluster_table` 导出说明中补充 `Weight=0` 表示该实例不接收流量，不会被后端强制改为默认值。 |
| `sys-design/总体设计文档.md` | 证书与集群管理的功能范围描述已提前与语义变更对齐。 |

## 7. 对下游/实现的影响

- **文档层面**：后续 API 代码实现可直接以文档中的 `合法性条件` 作为参数校验依据。
- **接口语义层面**：
  - `/expression/verify` 由“无鉴权”变为“需要 `FeatureRoute` 读权限”；
  - `certificates` 创建/更新接口不再要求上传 `cert_file_name`、`key_file_name`、`expired_date`；
  - `clusters` 的 `model_mappings` 字段名由 `key`/`value` 变为 `source_model`/`target_model`；
  - `clusters` 的 `instance_pool[].weight=0` 明确为不接收流量，不再被后端默认改为 1。
- **BFE 数据面**：InnerAPI 导出内容本身无变化，仅文档补齐。
