# Issue #99：创建 API-Key 偶发 500 / Lock wait timeout 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/99](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/99)

> 调接口创建 api-key 失败，`access.log` 报 500，耗时约 50 秒：
> ```
> cost_ms[50211] method[POST] pattern[/api-keys] path[/open-api/v1/api-keys]
> status_code[500] ret_msg[Biz Exception: allocate api-key id sequence: Error 1205:
> Lock wait timeout exceeded; try restarting transaction]
> ```
> 此前可以正常创建 api-key。

## 2. 目标

1. 消除高并发创建 API-Key 时的 `Error 1205: Lock wait timeout exceeded`；
2. 保留现有 `api-key-{seq}` 的可读 ID 格式；
3. 不引入新的数据迁移或回滚成本（复用已存在的 `api_key_id_seq` 表）；
4. 修复后的分配逻辑在 MySQL 默认隔离级别（REPEATABLE READ）和 SQLite 下均正确工作。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `storage/rdb/internal/dao/table_api_key_id_seq.go`、`storage/rdb/api_key/id_generator.go` |
| 数据库 | 复用已有 `api_key_id_seq` 表，不修改表结构 |
| 接口契约 | OpenAPI 请求/响应字段不变；仅消除内部 500 错误 |
| 数据迁移 | 无需新的迁移；但需确认旧迁移已执行（`api_key_id_seq` 已初始化） |

## 4. 最终方案概览

**将 `TAPIKeyIDSeqAllocate` 从“先 SELECT 再 CAS UPDATE”改为“原子自增 UPDATE + 读取新值”。**

- MySQL：使用 `UPDATE ... SET next_seq = LAST_INSERT_ID(next_seq + 1)` + `SELECT LAST_INSERT_ID()`，利用会话级 `LAST_INSERT_ID()` 原子地返回新序号；
- SQLite：使用 `UPDATE ... RETURNING next_seq`（SQLite 3.35+），或在同一事务内做原子自增后读取；
- 移除事务内的 CAS 重试循环，从根本上避免 REPEATABLE READ 下快照读导致的死循环和行锁竞争。

## 5. 预期收益与风险

| 项目 | 说明 |
|------|------|
| 收益 | 根除并发场景下的 lock wait timeout；批量创建稳定；错误码统一为成功或明确的参数错误 |
| 主要风险 | 需要为 MySQL 和 SQLite 维护两条 SQL 分支；旧版本 SQLite（<3.35）不支持 `RETURNING` |
| 兼容性 | `api_keys.id` 唯一索引保留；新创建的 ID 仍保持 `api-key-{seq}` 格式；旧数据无需修改 |

## 6. 上线前检查清单

1. 确认 `api_key_id_seq` 表已在目标数据库创建；
2. 确认 `api_key_id_seq` 已按 [2026-08-23-issue-80-apikey-id-race](../2026-08-23-issue-80-apikey-id-race/change-summary.md) 的迁移脚本完成初始化；
3. 在 MySQL 环境下进行并发创建压测（如 50~100 并发），验证无 500 / lock wait timeout；
4. 在 SQLite 环境下跑全量集成测试，确认无回归。

## 7. 参考文档

- [Issue #99](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/99)
- [2026-08-23-issue-80-apikey-id-race](../2026-08-23-issue-80-apikey-id-race/change-summary.md)
- `ai-gateway-api/storage/rdb/internal/dao/table_api_key_id_seq.go`
- `ai-gateway-api/storage/rdb/api_key/id_generator.go`
