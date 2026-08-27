# Issue #99：API-Key ID 分配 lock wait timeout 设计变更说明

## 1. 当前问题定位

### 1.1 错误表象

调用 `POST /open-api/v1/api-keys` 创建 API-Key 时，接口耗时约 50 秒后返回 500：

```text
Biz Exception: allocate api-key id sequence: Error 1205:
Lock wait timeout exceeded; try restarting transaction
```

- `cost_ms[50211]` 与 MySQL 默认 `innodb_lock_wait_timeout = 50s` 一致；
- 错误消息中的 `allocate api-key id sequence` 明确指向 `storage/rdb/api_key/id_generator.go` 中的 `RDBAPIKeyIDGenerator.Generate`；
- 出问题前接口可以正常创建 api-key。

### 1.2 引入问题的变更

该问题由 commit `e19df24a4ec7dd69a097f45c28958277569410c7`（PR #80）引入。

PR #80 为了修复 issue #80 的并发 duplicate id，新增了 `api_key_id_seq` 表，并实现了 CAS（compare-and-set）式序号分配：

```text
endpoints/openapi_v1/api_key/create.go
    └── container.APIKeyIDGenerator.Generate(ctx, product.Name)
        └── storage/rdb/api_key/id_generator.go
            └── dao.TAPIKeyIDSeqAllocate(dbCtx, productName)
                └── allocateAPIKeyIDSeqOnce(conn, productName)
                    └── 短事务内:
                        1. SELECT next_seq FROM api_key_id_seq WHERE product_name = ?
                        2. UPDATE api_key_id_seq
                           SET next_seq = next_seq + 1
                           WHERE product_name = ? AND next_seq = ?
                        3. 若 affected = 0，则同一事务内重读并重试
```

### 1.3 根因分析

PR #80 的实现基于一个错误前提：MySQL 默认隔离级别是 READ COMMITTED。但实际上 **MySQL InnoDB 的默认隔离级别是 REPEATABLE READ**。

在 REPEATABLE READ 下，当前 CAS 逻辑会失效：

1. 事务 A、B 同时执行普通 `SELECT next_seq`，由于普通 SELECT 不加锁，两者都读到 `current = 100`；
2. 事务 A 的 `UPDATE ... WHERE next_seq = 100` 获得行锁，成功将值改为 101，随后提交并释放锁；
3. 事务 B 的 `UPDATE ... WHERE next_seq = 100` 被阻塞等待行锁；A 提交后 B 获得锁，但当前值已是 101，`WHERE next_seq = 100` 不匹配，`affected = 0`；
4. B 在同一事务内 `continue` 重试，重新 `SELECT next_seq`。在 REPEATABLE READ 下，事务内快照读仍然返回 100（事务开始时的快照）；
5. B 再次尝试 `UPDATE ... WHERE next_seq = 100`，再次 `affected = 0`，陷入**事务内死循环**；
6. 每次循环都竞争同一行的锁。并发请求增多时，部分事务会在等待行锁的过程中达到 `innodb_lock_wait_timeout`，触发 `Error 1205`。

这就是 issue #99 中 50 秒超时 500 的根因。

### 1.4 为什么单元测试没发现

`storage/rdb/internal/dao/table_api_key_id_seq_test.go` 使用 SQLite 内存库测试，并设置了：

```go
db.SetMaxOpenConns(1)
```

这强制所有 goroutine 串行使用同一连接，无法模拟 MySQL 下多事务并发竞争行锁的场景，因此现有测试未能暴露该缺陷。

## 2. 目标

1. 消除 MySQL 默认 REPEATABLE READ 隔离级别下的 lock wait timeout；
2. 保持 `api-key-{seq}` 格式不变；
3. 复用已有 `api_key_id_seq` 表，不引入新的数据迁移；
4. 保持 SQLite 兼容性，使现有单元测试和集成测试继续通过。

## 3. 方案对比

| 方案 | 思路 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. 原子自增 UPDATE + 读取新值（推荐） | `UPDATE ... SET next_seq = next_seq + 1`，然后读取新值 | 单次写操作，自然原子；无 CAS 重试死循环 | 需要 MySQL/SQLite 两条 SQL 分支读取返回值 | **采用** |
| B. SELECT ... FOR UPDATE + UPDATE | MySQL 加行锁后更新，SQLite 单连接串行 | 保留 CAS 思路，容易理解 | SQLite 不支持 FOR UPDATE；MySQL 仍有多事务排队问题 | 不采用 |
| C. 数据库原生 AUTO_INCREMENT | 每个 product 维护一个自增序列 | 完全依赖数据库，无应用层锁 | 需要按 product 分桶的序列，改动大 | 不采用 |
| D. 调大 `innodb_lock_wait_timeout` | 运维侧临时缓解 | 不改代码 | 不解决死循环，只是推迟超时 | 不采用 |

## 4. 推荐方案：原子自增 UPDATE

### 4.1 核心思路

不再先 SELECT 当前值，而是直接让数据库原子地把 `next_seq` 加 1，并返回增加后的值。由于 `UPDATE` 本身会加行锁，所有并发请求自然排队，每个请求拿到不同的序号，无需 CAS 重试。

### 4.2 MySQL 实现

MySQL 使用单条 `INSERT ... ON DUPLICATE KEY UPDATE` 配合 `LAST_INSERT_ID()`：

```sql
INSERT INTO api_key_id_seq (product_name, next_seq)
VALUES (?, LAST_INSERT_ID(2))
ON DUPLICATE KEY UPDATE next_seq = LAST_INSERT_ID(next_seq + 1);

SELECT LAST_INSERT_ID();
```

说明：
- 若记录不存在：插入 `(product_name, 2)`，同时 `LAST_INSERT_ID()` 被设为 `2`；
- 若记录存在：`next_seq` 原子加 1，`LAST_INSERT_ID()` 被设为新值；
- `SELECT LAST_INSERT_ID()` 读取到的是新的 `next_seq`，减 1 即为本次分配的序号；
- 单条语句完成初始化和自增，避免分步 `INSERT IGNORE + UPDATE` 在高并发下产生死锁；
- `LAST_INSERT_ID()` 是会话级变量，不受 REPEATABLE READ 快照语义影响。

### 4.3 SQLite 实现

SQLite 3.35+ 支持 `RETURNING`：

```sql
-- 初始化（首次创建 product 时）
INSERT OR IGNORE INTO api_key_id_seq (product_name, next_seq) VALUES (?, 1);

-- 原子自增并返回新序号
UPDATE api_key_id_seq
SET next_seq = next_seq + 1
WHERE product_name = ?
RETURNING next_seq;
```

说明：
- `INSERT OR IGNORE` 用于首次分配时兜底初始化；
- `RETURNING next_seq` 直接返回更新后的新值；
- 对于不支持 `RETURNING` 的旧版 SQLite，可退化为同一事务内 `UPDATE` 后 `SELECT next_seq`（SQLite 写锁是数据库级，串行执行，安全）。

### 4.4 Go 代码变更

#### `storage/rdb/internal/dao/table_api_key_id_seq.go`

新增/替换 `TAPIKeyIDSeqAllocate` 实现：

```go
// TAPIKeyIDSeqAllocate atomically allocates and returns the next available
// API-Key sequence number for the given product. It uses a single atomic
// UPDATE to avoid the CAS retry loop that caused lock wait timeouts under
// MySQL REPEATABLE READ.
func TAPIKeyIDSeqAllocate(dbCtx lib.DBContexter, productName string) (int64, error) {
    conn := dbCtx.Conn()
    return allocateAPIKeyIDSeq(conn, productName)
}

func allocateAPIKeyIDSeq(conn *sql.DB, productName string) (int64, error) {
    var allocated int64
    useMySQL := isMySQL(conn)
    err := lib.Transaction(conn, func(tx *sql.Tx) error {
        if useMySQL {
            // MySQL: single INSERT ... ON DUPLICATE KEY UPDATE atomically
            // initializes the row or increments next_seq. LAST_INSERT_ID is
            // set to the new next_seq value, so the allocation result is
            // LAST_INSERT_ID() - 1.
            _, err := tx.ExecContext(context.Background(),
                "INSERT INTO api_key_id_seq (product_name, next_seq) VALUES (?, LAST_INSERT_ID(2)) "+
                    "ON DUPLICATE KEY UPDATE next_seq = LAST_INSERT_ID(next_seq + 1)",
                productName)
            if err != nil {
                return err
            }
            row := tx.QueryRowContext(context.Background(), "SELECT LAST_INSERT_ID()")
            if err := row.Scan(&allocated); err != nil {
                return err
            }
            allocated--
            return nil
        }

        // SQLite: ensure the row exists, then atomically increment and read
        // back the new value inside the same transaction.
        if _, err := tx.ExecContext(context.Background(),
            "INSERT OR IGNORE INTO api_key_id_seq (product_name, next_seq) VALUES (?, 1)",
            productName); err != nil {
            return err
        }
        if _, err := tx.ExecContext(context.Background(),
            "UPDATE api_key_id_seq SET next_seq = next_seq + 1 WHERE product_name = ?",
            productName); err != nil {
            return err
        }
        row := tx.QueryRowContext(context.Background(),
            "SELECT next_seq FROM api_key_id_seq WHERE product_name = ?",
            productName)
        if err := row.Scan(&allocated); err != nil {
            return err
        }
        allocated--
        return nil
    })
    return allocated, err
}
```

说明：
- `isMySQL` 通过 `conn.Driver()` 类型断言判断是否为 MySQL 驱动；
- 移除原 `allocateAPIKeyIDSeqOnce` 和 `isRetryableSeqError`；
- 外层 10 次重试循环也已移除，因为不再需要 CAS 重试。

#### `storage/rdb/api_key/id_generator.go`

无需修改接口，继续调用 `dao.TAPIKeyIDSeqAllocate` 并格式化 ID：

```go
func (g *RDBAPIKeyIDGenerator) Generate(ctx context.Context, productName string) (string, error) {
    dbCtx, err := g.dbCtxFactory(ctx)
    if err != nil {
        return "", err
    }

    seq, err := dao.TAPIKeyIDSeqAllocate(dbCtx, productName)
    if err != nil {
        return "", fmt.Errorf("allocate api-key id sequence: %w", err)
    }

    return fmt.Sprintf("api-key-%d", seq), nil
}
```

### 4.5 其他文件

- `endpoints/openapi_v1/api_key/create.go`：无需改动，仍调用 `container.APIKeyIDGenerator.Generate`；
- `model/api_key/api_key.go`：无需改动，保留显式 ID 查重保护；
- `db_ddl.sql` / `db_ddl_sqlite.sql`：无需改动，复用已有 `api_key_id_seq` 表。

## 5. 数据迁移

本次修复**不需要新的数据迁移**。但需确认 PR #80 的迁移已执行：

```sql
-- MySQL
INSERT INTO api_key_id_seq (product_name, next_seq)
SELECT product_name, COALESCE(MAX(CAST(SUBSTRING(id, 9) AS UNSIGNED)) + 1, 1)
FROM api_keys
WHERE id LIKE 'api-key-%'
GROUP BY product_name;

-- SQLite
INSERT INTO api_key_id_seq (product_name, next_seq)
SELECT product_name, COALESCE(MAX(CAST(SUBSTR(id, 9) AS INTEGER)) + 1, 1)
FROM api_keys
WHERE id LIKE 'api-key-%'
GROUP BY product_name;
```

## 6. 测试计划

### 6.1 单元测试

1. 保留 `storage/rdb/internal/dao/table_api_key_id_seq_test.go` 现有 SQLite 测试：
   - `Basic`：确保分配语义不变（首次返回 1，后续递增）；
   - `PerProduct`：验证不同 product 的序号独立；
   - `Concurrent`：验证 50 goroutine 并发分配无重复。

2. 新增 `storage/rdb/internal/dao/table_api_key_id_seq_mysql_test.go`（带 `//go:build mysql` 标签）：
   - `MySQL_Basic`：验证 MySQL 下基本分配语义；
   - `MySQL_Concurrent`：100 goroutine 并发分配，断言无 `Error 1205` / deadlock，无重复序号；
   - 默认 `go test` 不执行，需显式带 `-tags mysql` 运行。

### 6.2 集成测试

1. 在 SQLite 环境下跑全量 `tests/api_key/create/...`，确认无回归；
2. 在 MySQL 环境下补充并发创建压测（如 100 并发，持续创建 700 个 key），验证：
   - 无 500 / lock wait timeout；
   - 无 duplicate id；
   - `api_key_id_seq.next_seq` 连续递增。

### 6.3 回归验证命令

```bash
# 默认 SQLite 单元测试
go test ./storage/rdb/internal/dao/...
go test ./endpoints/openapi_v1/api_key/...
go test ./model/api_key/...
go test ./tests/api_key/create/... -run TestAPIKey_Create$

# MySQL 专项测试（需本地 MySQL 已启动）
go test -tags mysql ./storage/rdb/internal/dao/... -run TestTAPIKeyIDSeqAllocate_MySQL
```

## 7. 风险与缓解

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| MySQL/SQLite SQL 分支增加维护成本 | 两种数据库需要不同 SQL 读取返回值 | 将分支逻辑收敛在 `table_api_key_id_seq.go` 一个文件内，并补充明确注释 |
| 并发下仍可能等待行锁 | 所有创建请求更新同一行 | 单条原子语句比 CAS 重试快得多，锁持有时间从数十秒降到毫秒级；必要时可按 product 分桶（已满足） |

## 8. 实施状态

- [x] 修改 `storage/rdb/internal/dao/table_api_key_id_seq.go`，实现原子自增分配；
- [x] 保留现有 SQLite DAO 单元测试（`Basic`、`PerProduct`、`Concurrent` 均通过）；
- [x] 新增 MySQL DAO 单元测试（带 `mysql` build tag，100 goroutine 并发通过）；
- [x] 全量回归测试通过（`go test ./...`）。

## 9. 参考文档

- [Issue #99](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/99)
- [Issue #80](../2026-08-23-issue-80-apikey-id-race/change-summary.md)
- `ai-gateway-api/storage/rdb/internal/dao/table_api_key_id_seq.go`
- `ai-gateway-api/storage/rdb/api_key/id_generator.go`
- `ai-gateway-api/lib/xdb.go`
