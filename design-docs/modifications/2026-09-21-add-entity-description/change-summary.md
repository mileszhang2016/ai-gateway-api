# Entity 增加 description 字段变更摘要

## 1. 背景

`/entities` 的 Entity 数据模型目前缺少描述字段，用户无法对 Entity（部门、团队、项目等组织节点）添加备注说明，在实体数量增多后不利于辨识与管理。

系统中其他同类资源均已支持描述字段，Entity 是组织管理接口中唯一缺少该字段的资源：

| 资源 | 文件 | description 合法性条件 |
|------|------|------------------------|
| Entity-Type | `entity-types.md` | 非必填；若传入，长度 0-255 字符；不能包含控制字符 |
| Provider | `providers.md` | 非必填；若传入，长度 0-256 字符；不能包含控制字符 |
| API-Key | `api-keys.md` | 必填、非空；长度不超过 512 字符 |
| Cluster / Certificate | `clusters.md` / `certificates.md` | 已有描述字段 |

其中 Entity-Type 与 Entity 同属组织元数据，语义最接近，本变更**直接对齐 entity-types.md 的合法性条件**。

## 2. 目标

1. 为 Entity 数据模型增加 `description` 字段，创建、查询（列表/单个）、全量更新、部分更新接口全链路支持；
2. 合法性条件与 entity-types.md 完全一致：**非必填；若传入，长度 0-255 字符；不能包含控制字符**；
3. 向后兼容：存量数据与存量调用方不传 `description` 时行为不变（视为空描述）；
4. `description` 为纯控制面元数据，不进入任何导出 topic，数据面（conf-agent / BFE）零感知。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 接口契约 | `design-docs/api-define/OpenAPI接口定义/entities.md`：§1 数据模型、§2.1 创建、§2.2 查询列表、§2.3 查询单个、§2.4 全量更新、§2.5 部分更新 |
| 不涉及接口 | §2.6 删除（Data 为 null）、§2.7 查询配额计划、§2.8 重置配额余额，均与 description 无关 |
| 不新增 | 列表查询不新增按 `description` 过滤的参数（保持最小变更，后续有需求再单独立项） |
| 代码实现（Step 5） | `endpoints/openapi_v1/entity/`（create、full_update、update、validator、查询回包）、`model/entity/`（EntityParam、`entityParamToMap`、操作日志快照）、`storage/rdb/entity/`（字段映射） |
| 数据库 | `db_ddl.sql` / `db_ddl_sqlite.sql`：`entities` 表新增 `description` VARCHAR(255) 列，并补充存量部署的 `ALTER TABLE` 迁移说明 |
| 数据面影响 | 无：`model/imods` 导出结构不变，BFE / conf-agent 无感知 |
| Dashboard | `ai-gateway-web` 可后续在实体列表/详情展示，本期不强制 |
| 数据迁移 | 存量行 `description` 默认为空字符串，无需数据回填 |

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| 合法性条件对齐 entity-types.md（0-255 字符、禁控制字符），而非 api-keys.md 的必填 512 | Entity 与 Entity-Type 同属组织维度元数据，描述均为"自定义"可选备注；API-Key 的必填语义不适用于组织节点 |
| 默认值 | 未传入时为空字符串 `""`；DB 列默认 `''`，API 一律返回字符串（不出现 null），与 `entity-types.md` 字段说明中"自定义"的宽松语义一致 |
| PUT 全量更新语义 | 请求体同创建接口：`description` 省略时按默认值（空字符串）处理，即**清空已有描述**——与 §2.4 "Body 同创建"及 `quota_plan` 等字段"若未设置则使用默认值"的全量语义一致 |
| PATCH 部分更新语义 | 省略 `description` 保持原值；显式传 `""` 表示清空描述（0-255 允许 0 长度，空字符串合法） |
| 校验失败错误码 | 长度超限或含控制字符 → 422 `Param Illegal`（`xerror.WrapParamError`），与既有 `validateEntityParam` 参数校验同类别、同位置 |
| 操作日志 | `description` 纳入创建/更新操作日志的 before/after 快照（`entityParamToMap` 增加该字段），审计语义与既有字段一致 |
| 数据面零影响 | `description` 不进入 `cluster_table` / `gslb` / `server_data_conf` 等任何导出 topic，仅控制面存储与展示 |

## 5. 兼容性

- **API 向后兼容**：旧客户端不传 `description` → 创建默认为空；查询响应新增字段不影响按名取值的客户端。
- **数据库迁移**：新增列对存量部署需执行 `ALTER TABLE entities ADD COLUMN description VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Entity描述';`（SQLite 对应语句不含 COMMENT），在 sys-design 数据库设计文档中补充迁移说明。

## 6. 实施步骤（对照六步法）

1. **Step 1/2（本目录）**：变更摘要（本文件）+ 接口契约变更明细（`api-changes.md`）；
2. **Step 3**：按 `api-changes.md` 逐节修改 `entities.md`；
3. **Step 4**：同步 `sys-design/`（数据库设计文档 `entities` 表结构、存储层设计文档实体字段、接口层设计文档参数表）及 `sys-design/summary.md` 索引（如涉及）；
4. **Step 5**：按 §3 范围实现代码（接口层 → 模型层 → 存储层），DDL 同步，补单测/集成测试；
5. **Step 6**：评估是否沉淀 details 文档（预计不需要，属常规加字段变更）。

## 7. 关联文档

- 接口契约变更明细：`design-docs/modifications/2026-09-21-add-entity-description/api-changes.md`

## 8. 实施记录

- **代码实现**（与 api-changes.md §5 一致，工作区已改、未提交）：
  - DDL：`db_ddl.sql` / `db_ddl_sqlite.sql` `entities` 表新增 `description VARCHAR(255) NOT NULL DEFAULT ''`（位于 `name` 之后）；
  - DAO：`storage/rdb/internal/dao/table_entities.go` `TEntity`/`TEntityParam` 新增 `Description` 字段；
  - 存储层：`storage/rdb/entity/entity.go` `entityBaseDataToParam` 透传（Update 省略保持 nil 交 DAO nil-skip）、`entityParamToData` 回读（NOT NULL 列保证响应恒为字符串、不出现 null）；
  - 模型层：`model/entity/entity.go` `EntityParam` 新增 `Description *string`；`operation_log.go` `entityParamToMap` 纳入操作日志 before/after 快照；
  - 校验：`lib/validate/validate.go` 新增 `MaxEntityDescriptionLength = 255`；`endpoints/openapi_v1/entity/validator.go` `validateEntityParam` 增加 description 校验（超长/控制字符 → 422 PARAM）；
  - PUT 全量语义：`endpoints/openapi_v1/entity/full_update.go` 绑参校验后将 nil 显式置 `""`（省略即清空）；PATCH 不做处理（nil-skip 省略保留、显式 `""` 非 nil 写入清空）。
- **实现说明**：长度校验沿用代码库既有 `validate.Description`（按字节计数），与 entity-type 的实现对齐方式一致（其文档 0-255、代码 256 的存量差异不在本次范围）；本实现按批准的方案取 255。
- **单测**：
  - `endpoints/openapi_v1/entity/validator_test.go` 新增 `TestValidateEntityParamDescription`（7 例：nil 省略 / 空串 / 正常中文 / 255 边界 / 256 超限 / NUL / Tab）；
  - `model/entity/entity_manager_test.go` 新增 `TestEntityManager_EntityDescriptionFlow`（create/update 透传 storager + 审计快照包含 description）；
  - `storage/rdb/entity/entity_test.go` 新增 `TestEntityDescription_RoundTrip`（create 省略默认空 / 显式写入 / update 省略保留 / 显式置空清空），测试内存表结构同步加列。
- **集成测试**（`test/integration`，真实二进制子进程 + SQLite 自动执行新 DDL）：
  - `tests/entity/create` 新增 E-1-023~E-1-026（携带 description 回读一致 / 长度 255 / 长度 256 拒绝 / 控制字符拒绝）；
  - `tests/entity/full_update` 新增 E-4-009~E-4-011（携带写入并回读 / 省略清空 / 长度 256 拒绝）；
  - `tests/entity/partial_update` 新增 E-5-010~E-5-013（修改生效并回读 / 省略保留 / 显式 `""` 清空 / 长度 256 拒绝）；
  - `tests/schema/openapi/entity.go` `EntitySchema`/`EntityListItemSchema`：`Required` 增加 `description`、`Fields` 增加 `TypeString`（对齐 api_key/cluster/provider/entity_type 恒返回约定）；`openapi_schema_test.go` `testEntitySchema` 创建/PATCH 请求体携带 description；innerapi schema 无 entity 引用，确认无需变更；
  - `test/docs/README.md` 统计表：E Entity 用例 25→36，总计 218→229。
- **文档同步**：
  - `api-define/OpenAPI接口定义/entities.md`：§1 数据模型、§2.1 创建、§2.2 列表、§2.4 全量更新、§2.5 部分更新按 api-changes.md §2 逐节落位（共 10 处）；
  - `sys-design/数据库设计文档.md` §6.4：DDL、字段表、存量 `ALTER TABLE` 迁移说明；
  - `sys-design/模型层设计文档.md` §4.10.2：`EntityParam` 结构新增 `Description`；
  - `sys-design/details/部分更新语义与DAO-nil-skip约定.md`：§2 对照表新增 Entity `description` 行，新增 §7「PUT 全量更新的显式默认值（Entity description）」（Step 6 沉淀，复用既有 details 文档，无需新建）；
  - `CHANGELOG.md` `[Unreleased]` 新增 `Added` 条目。
- **验证结果**（2026-09-21，本地）：`go build ./...`、`go vet ./...`、`go test ./...` 全部通过；`go test -cover ./model/...` 中 `model/entity` 78.9%（≥70% 门禁）；集成测试 `tests/entity/...`（9 包）与 `tests/schema/...`（2 包）全部 PASS，含新增 11 个集成用例，既有用例无回归。
- **待办**：提交代码与文档；存量部署执行 `ALTER TABLE entities ADD COLUMN description VARCHAR(255) NOT NULL DEFAULT ''`（SQLite 无 COMMENT/AFTER）。
