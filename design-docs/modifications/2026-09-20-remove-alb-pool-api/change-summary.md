# 移除 /alb-pool OpenAPI 变更摘要

## 1. 背景

`/alb-pool`（`GET` 详情 + `PATCH` 全量替换）管理的是内置 BFE 实例池 `BFE.aipool`（池名由配置项 `RunTime.DefaultAIInstancePoolName` 提供）。经调研确认，该池的数据对 AI 网关运行**没有任何实际影响**：

1. 种子数据（`db_ddl.sql` / `db_ddl_sqlite.sql` / `init.sql`）中 `BFE.aipool` 仅被 `bfe_clusters.pool_name` 引用（内置 BFE 集群 `BFE-AI_product.szyf`），属于 GSLB 层的 BFE 网关实例列表概念；
2. 数据面消费的所有配置中均不包含该池实例：`cluster_table` 导出只读 Cluster → SubCluster → `InstancePool`（产品池，源自 provider `instance_pool` 快照），`gslb.data` 只导出按 BFE 集群名键控的调度权重矩阵，`server_data_conf` / `epp_data` 等 topic 也不读 BFE 池；
3. PATCH 只是 `pools.instance_detail` 列的数据库全量替换（`model/icluster_conf/pool.go` → `storage/rdb/cluster_conf/pool.go`），不触发任何导出 topic 版本变化，conf-agent 无感知，BFE 不重载；
4. `FetchBFEPools` 的唯一调用方是 OpenAPI 列表代码本身，`model/imods`、`model/iversion_control` 中无 pool 相关逻辑；BFE 代码库无任何 `aipool` 引用。

该接口是 BFE 原生 GSLB 多集群体系的遗留概念：完整 BFE 生态中 BFE 实例列表供上层 GSLB 调度客户端流量到具体 BFE 实例；当前 AI Gateway 部署客户端经 K8s Service 直达 BFE，实例列表无人消费。

## 2. 目标

- 删除 `GET /open-api/v1/alb-pool` 与 `PATCH /open-api/v1/alb-pool` 两个端点及其专属代码、配置、测试与文档；
- 保留仍被其他功能依赖的共享物：`FeatureBFEPool` 鉴权 feature（`/epp-pool`、`/epp-assignments` 在用）、`model/icluster_conf` PoolManager 与 `pools` 表（产品实例池在用）、`bfe_clusters` 种子行（`DefaultAIClusterName` 创建集群时用作默认调度矩阵键）。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | 仅 `ai-gateway-api`（按本期决策只删 api 侧；`ai-gateway-web` 的 AIInstancePool 页面调用本接口，本期保留、后续清理） |
| 删除模块 | `endpoints/openapi_v1/bfe_pool/`（`one.go`/`update.go` 及未注册的 `create.go`/`delete.go`/`list.go` 半成品）、`test/integration/tests/alb_pool/`（GET 2 例 + PATCH 11 例，共 13 例） |
| 修改模块 | `endpoints/openapi_v1/endpoints.go`（去注册）、`stateful/config.go`（删 `RunTime.DefaultAIInstancePoolName`）、`conf/ai_gateway_api.toml`、`test/integration/conf/ai_gateway_api.toml`、`test/docs/README.md`、`lib/validate/validate.go`（注释） |
| 删除文档 | `design-docs/api-define/OpenAPI接口定义/alb-pool.md` |
| 修改文档 | `OpenAPI接口定义/README.md`（索引）、`00-common.md`（命名示例）、`sys-design/总体设计文档.md`（功能清单 + 模块表）、`sys-design/接口层设计文档.md`（目录树、注册代码示例、§4.2.8 删除并重新编号后续小节）、`sys-design/存储层设计文档.md`、`sys-design/details/EPP调度对接.md`、`sys-design/数据库设计文档.md`、`test/integration/README.md`（目录树） |
| 数据迁移 | 无；`pools` 表保留，种子行 `BFE.aipool` 兼容保留（无 OpenAPI 可写，数据面不消费） |
| 数据面影响 | 无（BFE / conf-agent 零改动，导出 topic 无变化） |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 只删 api 侧，保留 ai-gateway-web 页面 | 本期最小变更；已知 `ai-gateway-web/src/modules/AIInstancePool/index.vue` 调用 GET/PATCH `alb-pool`，删除 API 后该页面失效（报 404），后续单独清理前端模块与 i18n 词条 |
| `FeatureBFEPool` 鉴权 feature 保留 | `/epp-pool`、`/epp-assignments` 的 Authorizer 仍使用 |
| PoolManager / `pools` 表 / `sub_clusters` 机制保留 | 产品实例池（provider `instance_pool` 快照 → cluster_table 导出 → BFE 后端）仍依赖 |
| `bfe_clusters` 种子行与 `DefaultAIClusterName` 保留 | 创建集群时作为默认调度矩阵键（`model/icluster_conf/cluster.go`），与 `/alb-pool` 无耦合 |
| `pools` 种子行 `BFE.aipool` 保留 | 被 `bfe_clusters.pool_name` 引用；无 API 删除路径后成为纯遗留数据，删除收益低、留待后续随种子数据统一清理 |
| 历史 modifications 文档不回改 | `2026-09-08-epp-scheduling-integration` 等历史变更记录中对 `/alb-pool` 的引用是当时事实，予以保留 |
| Breaking change 对外标注 | 两个端点从 200 变为 404；CHANGELOG 计入 Removed |

## 5. 关联文档

- 接口契约变更明细：`design-docs/modifications/2026-09-20-remove-alb-pool-api/api-changes.md`
