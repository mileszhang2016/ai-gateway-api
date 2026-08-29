# issue-95：provider 表 `keys` 列 MySQL 保留字冲突修复

## 背景与目标

在 `POST /open-api/v1/providers` 创建 provider 时，服务端始终返回 HTTP 500，错误为 MySQL Error 1064（SQL 语法错误）。

根因：`providers` 表存在名为 `keys` 的列，而 `keys` 是 MySQL 保留关键字。建表语句使用反引号将列名括起，因此建表成功；但 DAO 使用 gendry/builder 生成 INSERT/UPDATE 时不给列名加反引号，导致任何写入 `providers` 表的操作都确定性地触发语法错误。

本次修复目标：

1. 将 `providers` 表的列名从 MySQL 保留字 `keys` 改为非保留字 `api_keys`。
2. 同步更新 DAO 层字段 tag，使 SQL 生成合法。
3. 保持 OpenAPI/InnerAPI 的 JSON 字段名仍为 `keys`，不破坏接口契约。
4. 提供存量 MySQL 库的迁移脚本。

## 主要改动点

### 1. 数据库 schema 变更

- `db_ddl.sql`：`providers.keys` → `providers.api_keys`。
- `db_ddl_sqlite.sql`：`providers.keys TEXT` → `providers.api_keys TEXT`。

### 2. DAO 层字段 tag 变更

- `storage/rdb/internal/dao/table_providers.go`：
  - `TProvider.Keys` 的 tag 由 `db:"keys"` 改为 `db:"api_keys"`。
  - `TProviderParam.Keys` 的 tag 由 `db:"keys"` 改为 `db:"api_keys"`。

### 3. 存储转换层无需改动

- `storage/rdb/provider/provider.go` 中 `TProvider.Keys` / `TProviderParam.Keys` 仍通过 `marshalJSON` / `unmarshalKeys` 与 `[]iprovider.ProviderKey` 转换，字段名不变，仅依赖 db tag 映射到新列名。
- `model/iprovider/provider.go` 中 JSON tag 保持 `json:"keys"`，API 契约不变。

### 4. 存量数据迁移

上线前对已有 MySQL 实例执行：

```sql
ALTER TABLE providers CHANGE COLUMN `keys` `api_keys` JSON COMMENT 'API key 列表';
```

SQLite 为新建库场景，直接重新执行 `db_ddl_sqlite.sql` 即可。

## 兼容性说明

- **OpenAPI/InnerAPI 契约不变**：请求/响应 JSON 中字段名仍为 `keys`。
- **数据库列名变化**：新增环境按新 DDL 建表；存量环境需执行一次 `ALTER TABLE`。
- **BFE 数据面影响**：provider 导出字段名仍为 `keys`，数据面无感知。
- **集成测试**：现有 provider 创建/更新/查询用例无需修改，修复后应直接通过。

## 影响范围

- 修复 `POST /open-api/v1/providers` 500 错误。
- 修复 `PUT /open-api/v1/providers/{id}` 等所有对 `providers` 表的写入/更新操作。
- 修复按 `keys` 列过滤/查询时可能触发的语法错误。
- 解除 0.5.0 legacy-contract provider 迁移 120 个 E2E 用例的阻塞。

## 未选择的替代方案

**给 DAO 生成器统一加反引号**：gendry/builder 不自动转义列名，若要在 DAO 层统一包裹反引号，需要改动 `storage/rdb/internal/dao/internal/builder.go` 中所有表的构建逻辑，影响面大、回归成本高；且当前全库仅 `keys` 一个列名命中 MySQL 保留字，因此采用最小改动——重命名列名。

## 验证结果

- [ ] 本地使用 MySQL 8.0 启动服务，执行最小 payload 创建 provider，返回 HTTP 201（需本地 MySQL 环境，未执行）。
- [ ] 执行 provider 更新、列表查询、详情查询，均返回 200（需本地 MySQL 环境，未执行）。
- [x] `go test ./model/iprovider/...` 通过。
- [x] `go test ./...` 通过（ai-gateway-api 全量单元测试；环境无 `make`，使用 `go test ./...` 替代）。
- [ ] 集成测试 `test/integration/tests/provider/create/...` 与 `test/integration/tests/provider/update/...` 通过（需完整依赖环境，未执行）。
- [ ] 存量 MySQL 执行迁移脚本后，旧数据可正常读取、新 provider 可正常创建（需本地 MySQL 环境，未执行）。
