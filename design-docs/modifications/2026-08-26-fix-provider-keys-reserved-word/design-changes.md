# issue-95：provider 表 `keys` 列保留字冲突——设计变更说明

## 问题定位

### 复现路径

```bash
curl -X POST http://<gateway>:8183/open-api/v1/providers \
  -H "Authorization: Token <system_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "provider-demo",
    "instance_pool": [{"name": "b1", "addr": "127.0.0.1", "port": 8080, "weight": 100}],
    "model_protocols": ["openai"]
  }'
```

响应：

```json
{"ErrNum":500,"ErrMsg":"Database Exception: Error 1064: You have an error in your SQL syntax; ... near 'keys,model_endpoint,...' at line 1"}
```

### 代码路径

1. `endpoints/openapi_v1/provider/create.go` 接收请求并调用 `model/iprovider`。
2. `model/iprovider/provider.go` 校验后调用 `storage/rdb/provider/provider.go`。
3. `storage/rdb/provider/provider.go` 将 `iprovider.ProviderParam` 转换为 `dao.TProviderParam`。
4. `storage/rdb/internal/dao/internal/builder.go` 使用 `github.com/didi/gendry/builder.BuildInsert` 生成 SQL：

```sql
INSERT INTO providers (keys, model_endpoint, model_protocols, models, name, tiers, time_zone, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
```

其中 `keys` 未加反引号，触发 MySQL 语法错误。

## 数据模型变更

### 数据库 schema

#### MySQL (`db_ddl.sql`)

变更前：

```sql
CREATE TABLE `providers` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `name` VARCHAR(255) NOT NULL COMMENT 'Provider 标识',
  ...
  `keys` JSON COMMENT 'API key 列表',
  ...
);
```

变更后：

```sql
CREATE TABLE `providers` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `name` VARCHAR(255) NOT NULL COMMENT 'Provider 标识',
  ...
  `api_keys` JSON COMMENT 'API key 列表',
  ...
);
```

#### SQLite (`db_ddl_sqlite.sql`)

变更前：

```sql
CREATE TABLE providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  ...
  keys TEXT,
  ...
);
```

变更后：

```sql
CREATE TABLE providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  ...
  api_keys TEXT,
  ...
);
```

### DAO 结构体

`storage/rdb/internal/dao/table_providers.go`：

变更前：

```go
type TProvider struct {
    ...
    Keys           string    `db:"keys"`
    ...
}

type TProviderParam struct {
    ...
    Keys           *string    `db:"keys"`
    ...
}
```

变更后：

```go
type TProvider struct {
    ...
    Keys           string    `db:"api_keys"`
    ...
}

type TProviderParam struct {
    ...
    Keys           *string    `db:"api_keys"`
    ...
}
```

## 分层映射关系

| 层级 | 字段名 | 说明 |
|------|--------|------|
| OpenAPI/InnerAPI JSON | `keys` | 保持对外契约不变 |
| `model/iprovider.Provider` / `ProviderParam` | `Keys` (`json:"keys"`) | 业务模型不变 |
| `storage/rdb/internal/dao.TProvider` / `TProviderParam` | `Keys` (`db:"api_keys"`) | Go 字段名不变，仅 tag 映射到新列名 |
| MySQL/SQLite 列 | `api_keys` | 非保留字，避免语法错误 |

## 存量迁移

MySQL：

```sql
ALTER TABLE providers CHANGE COLUMN `keys` `api_keys` JSON COMMENT 'API key 列表';
```

说明：

- `CHANGE COLUMN` 会保留原有 JSON 数据。
- 不需要更新记录，因为只是列名变更，数据类型/约束不变。
- 执行前建议备份 `providers` 表。

SQLite：

SQLite 通常用于本地开发/测试，直接重新初始化数据库即可。若需保留数据，可：

```sql
ALTER TABLE providers RENAME COLUMN keys TO api_keys;
```

（需 SQLite 3.25.0+）

## 设计权衡

### 方案 A：重命名列名（本次采用）

- **优点**：改动最小，只影响 `providers` 表；不改动 DAO 生成器，回归风险低。
- **缺点**：需要存量库迁移。

### 方案 B：在 DAO 生成器中统一包裹反引号

- **优点**：无需改列名，未来新增保留字列也不怕。
- **缺点**：`gendry/builder` 本身不输出反引号，需要手写包装层或在 BuildInsert 前后处理，影响全库所有表；改动面大、回归成本高。

结论：当前仅 `keys` 一个列命中保留字，采用方案 A。

## 回归检查项

- [ ] `db_ddl.sql` 中 `providers` 表仅 `api_keys` 一处出现，无残留 `keys` 列定义。
- [ ] `db_ddl_sqlite.sql` 中 `providers` 表仅 `api_keys` 一处出现。
- [ ] `storage/rdb/internal/dao/table_providers.go` 中 `TProvider` 与 `TProviderParam` 的 `Keys` tag 均为 `db:"api_keys"`。
- [ ] 全仓搜索 `db:"keys"` 无残留（其他表如有同名 tag 需确认是否为 `providers` 表）。
- [ ] 启动服务后，MySQL 下 provider CRUD 全部成功。
