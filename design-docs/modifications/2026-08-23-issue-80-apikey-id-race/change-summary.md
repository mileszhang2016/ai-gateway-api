# Issue #80：并发创建 API-Key 出现 Duplicate id / 500 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/80](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/80)

> 多个 session 同时创建 apikey，有概率出现 422 Duplicate id 错误；并发量大时还会偶发 500。
>
> 根因：`endpoints/openapi_v1/api_key/create.go:143` 的 `generateAPIKeyID` 使用“读全表 → 找 max `api-key-N` → N+1”的非原子方式生成 ID，而 `api_keys.id` 上有唯一索引 `uk_id`。并发时两个请求算到同一个 ID：
> - 先插成功 → 后者的插入前查重命中 → 422 Duplicate id；
> - 或者查重窗口错过 → INSERT 撞唯一约束 → 未捕获 DB 错误 → 500。

## 2. 目标

消除并发创建 API-Key 时的 ID 生成竞态，确保在高并发（如批量创建 700 key）下：

1. 不再出现 422 Duplicate id；
2. 不再因唯一约束冲突产生 500；
3. 创建接口保持幂等、稳定、可预测；
4. **保留现有 `api-key-{seq}` 的 ID 格式**（已确认产品必须保留该可读格式，不采用 UUID 备选方案）。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `endpoints/openapi_v1/api_key/create.go`、`model/api_key/api_key.go`、`storage/rdb/internal/dao/`、`db_ddl.sql`、`db_ddl_sqlite.sql` |
| 数据库 | 新增 `api_key_id_seq` 序列表（推荐方案）或移除手动 ID 生成逻辑（备选方案） |
| 接口契约 | OpenAPI 请求/响应字段不变；仅消除并发错误 |
| 数据迁移 | 需要初始化 `api_key_id_seq`，使新序列从当前最大 ID 之后开始 |

## 4. 最终方案概览

**采用“数据库原子序列表”方案，并确认保留 `api-key-{seq}` 格式。**

- 新增表 `api_key_id_seq(product_name, next_seq)`，按 `product_name` 维护下一个可用序号。
- 使用 compare-and-set（CAS）更新 + 短事务外层重试的方式原子地占用一个序号，确保同一 product 下并发请求拿到不同序号；兼容 MySQL 与 SQLite。
- `generateAPIKeyID` 改为调用新的 `APIKeyIDGenerator`，不再 `SELECT *` 全表扫描。
- 保留 `api-key-{seq}` 格式，兼容现有设计文档、BFE `key_id` 消费方及前端展示。

> **UUID/ULID 备选方案未采用**：虽然代码改动更小，但会改变 `id` 格式，降低可读性与排序性，不符合产品要求。

## 5. 预期收益与风险

| 项目 | 说明 |
|------|------|
| 收益 | 根除恶化并发体验的竞态；批量创建稳定；错误码统一为成功或明确的参数错误 |
| 主要风险 | 新序列表需要一次性迁移；跨 MySQL/SQLite 的 upsert/行锁语法需抽象一致 |
| 兼容性 | 现有 `api_keys.id` 唯一索引保留；新创建的 ID 仍保持 `api-key-{seq}` 格式；旧数据无需修改；模型层对显式传入重复 ID 的查重保护继续生效 |

## 6. 上线前必须执行的迁移

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

> 说明：`next_seq` 必须初始化为当前最大序号 + 1，否则新分配序号可能与旧数据冲突，触发唯一索引错误或生成重复 ID。

## 7. 参考文档

- [Issue #80](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/80)
- `ai-gateway-api/endpoints/openapi_v1/api_key/create.go`
- `ai-gateway-api/model/api_key/api_key.go`
- `ai-gateway-api/db_ddl.sql`（`uk_id` / `uk_api_key` 唯一索引）
- `ai-gateway-api/design-docs/sys-design/模型层设计文档.md`（`api-key-{seq}` ID 生成说明）
