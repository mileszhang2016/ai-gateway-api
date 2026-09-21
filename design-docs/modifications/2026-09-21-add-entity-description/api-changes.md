# /entities 增加 description 字段 —— API 契约变更

## 1. 变更概览

| 变更类型 | 端点 | 说明 |
|----------|------|------|
| 修改 | `POST /entities` | 请求体新增可选 `description`；响应新增 `description` |
| 修改 | `GET /entities` | 列表元素新增 `description` 字段 |
| 修改 | `GET /entities/{id}` | 返回新增 `description` 字段（随数据模型） |
| 修改 | `PUT /entities/{id}` | 请求体新增可选 `description`（全量语义，省略即清空） |
| 修改 | `PATCH /entities/{id}` | 请求体新增可选 `description`（部分更新语义，省略保留、`""` 清空） |

不涉及：`DELETE /entities/{id}`（Data 为 null）、`GET /entities/{id}/quota-plan`、`POST /entities/{id}/quota-plan/reset`。

**字段统一定义**（对齐 `entity-types.md`）：

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `description` | string | Entity描述 | 自定义 | 非必填；若传入，长度 0-255 字符；不能包含控制字符 |

## 2. entities.md 逐节修改明细

### 2.1 §1 数据模型

**JSON 示例**：`name` 之后增加一行——

```json
{
  "id": "ent-001",
  "name": "op",
  "description": "运营部，负责线上业务",
  "type": "dep",
  ...
}
```

**字段说明表**：`name` 行之后新增——

```markdown
| `description` | string | Entity描述 | 自定义 | 非必填；若传入，长度 0-255 字符；不能包含控制字符 |
```

### 2.2 §2.1 创建Entity

**输入参数（Body）表**：`name` 行之后新增——

```markdown
| description | string | Entity描述 | N | - | 非必填；若传入，长度 0-255 字符；不能包含控制字符 |
```

**约束**：新增一条——

```markdown
- `description` 非必填；若传入，长度 0-255 字符；不能包含控制字符。
```

**HTTP BODY参数示例**、**成功返回示例**：JSON 中 `"name": "op"` 之后增加 `"description": "运营部，负责线上业务",`。

**执行逻辑**：无需新增步骤——"1. 校验参数合法性"自然涵盖 `description` 校验；"8. 写入entity"自然包含该字段。

### 2.3 §2.2 查询Entity列表

**list 对象字段说明表**：`name` 行之后新增——

```markdown
| description | string | Entity描述 | - |
```

**成功返回示例**：两个 list 元素的 `"name"` 之后各增加 `"description": "...",`（第二个元素 `bfe` 可用 `"description": "BFE 网关团队"`）。

### 2.4 §2.3 查询单个Entity

**返回数据（Data内容）** 当前描述为"字段同4.1数据模型"（实际指向 §1 数据模型），随数据模型自动包含 `description`，**文字无需改动**。

> 注：文档中 §2.3/§2.4/§2.5 的"同4.1数据模型"引用编号与实际章节号（§1）不一致，属历史遗留笔误，本期可顺手修正为"同1. 数据模型"，但不属于本变更的必要内容。

### 2.5 §2.4 全量更新Entity

**输入参数（Body）**：同创建接口，自动包含 `description`，无需单独列表。

**约束**：新增——

```markdown
- `description` 非必填；若传入，长度 0-255 字符；不能包含控制字符。**全量语义下省略 `description` 将清空已有描述（重置为空字符串）**。
```

**执行逻辑**："4. 更新Entity其他字段"自然涵盖。

### 2.6 §2.5 部分更新Entity

**输入参数（Body）**：同创建接口、仅传需修改字段，自动支持单独修改 `description`。

**约束**：新增——

```markdown
- `description` 非必填；若传入，长度 0-255 字符；不能包含控制字符。省略时保持原值；显式传入空字符串 `""` 表示清空描述。
```

### 2.7 §2.6 / §2.7 / §2.8

无修改。

## 3. 校验与错误码

| 场景 | HTTP status | ErrNum | 说明 |
|------|-------------|--------|------|
| `description` 长度 0-255、无控制字符 | 200 | 200 | 正常写入 |
| `description` 长度超过 255 字符 | 422 | 422 | `Param Illegal`，由 `validateEntityParam` 扩展校验 |
| `description` 包含控制字符 | 422 | 422 | `Param Illegal`（控制字符定义与 entity-type 现有实现对齐） |

## 4. 语义边界速查

| 场景 | 行为 |
|------|------|
| POST 不传 `description` | 创建为空描述（`""`），响应返回 `""` |
| POST 传合法 `description` | 写入并返回 |
| GET 列表 / 单个 | 始终返回字符串，存量数据为空字符串 |
| PUT 不传 `description` | **清空**已有描述（全量语义，同创建默认值） |
| PUT 传合法 `description` | 覆盖写入 |
| PATCH 省略 `description` | 保持原值 |
| PATCH 传 `""` | 清空描述 |
| PATCH 传合法非空 `description` | 覆盖写入 |

## 5. 代码与 DDL 变更清单（Step 5 实施参考）

| 层 | 位置 | 变更 |
|----|------|------|
| DDL | `db_ddl.sql`、`db_ddl_sqlite.sql` | `entities` 表新增 `description VARCHAR(255) NOT NULL DEFAULT ''`；sys-design 数据库设计文档补充存量 `ALTER TABLE` 迁移说明 |
| 接口层 | `endpoints/openapi_v1/entity/validator.go` | `validateEntityParam` 增加 `description` 长度/控制字符校验 |
| 接口层 | `endpoints/openapi_v1/entity/create.go`、`full_update.go`、`update.go` | 绑参与回包结构透传 `description` |
| 模型层 | `model/entity/entity_manager.go` | `EntityParam` 增加 `Description`；`entityParamToMap`（操作日志快照）纳入该字段 |
| 存储层 | `storage/rdb/entity/entity.go` | 结构体映射与 `entityDataToParamForUpdate` 透传；DAO nil-skip 语义下 PATCH 省略自动保留原值，与既有可选字段一致 |
| 导出 | `model/imods/` 等 | **不改**：description 不进入任何导出 topic |

## 6. 测试变更建议（Step 5）

| 类型 | 位置 | 用例 |
|------|------|------|
| 单测 | `endpoints/openapi_v1/entity/validator_test.go` | description 合法 / 255 字符边界 / 超长 422 / 含控制字符 422 |
| 单测 | `model/entity/entity_manager_test.go` | 创建/更新携带 description 的 mock 断言；操作日志快照包含 description |
| 集成测试 | `test/integration/tests/entity/` | 创建带 description → GET 回读一致；PATCH 修改 / `""` 清空 → 回读；PUT 省略 → 清空；PUT/PATCH 超长 → 422 |
| 集成测试（schema） | `test/integration/tests/schema/openapi/entity.go` | `EntitySchema`/`EntityListItemSchema`：`Required` 增加 `"description"`，`Fields` 增加 `description: TypeString`。对齐 api_key/cluster/provider/entity_type schema 的既有约定——description 恒返回（空字符串），故放 `Required` 而非 `Optional` |
| 集成测试（schema） | `test/integration/tests/schema/openapi/openapi_schema_test.go` | `testEntitySchema`：POST 创建与 PATCH 请求体携带 `description`（参照 api_keys 用例的写法），使创建/列表/单个/PUT/PATCH 各环节的 schema 断言覆盖新字段 |
| 回归 | 全量 `go test ./...`、`make test-model-cover-gate` | 既有用例无回归，model 覆盖率 ≥70% |

> **innerapi schema 无需变更**：`test/integration/tests/schema/innerapi/` 无 entity 相关 schema（description 不进入任何导出 topic，数据面零影响），与 §5 导出决策一致。
