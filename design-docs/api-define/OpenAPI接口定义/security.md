# /security/reencrypt-sweeps

密钥静态加密（DB 落盘加密）的**收敛任务**接口：对库内敏感列（`providers.api_keys`、
`api_keys.api_key`）做批量重加密/解密。任务异步执行，POST 创建任务（202），GET
按 task_id 查询状态与进度。执行语义见 `design-docs/modifications/2026-10-05-db-encryption-at-rest/`。

## 1. 数据模型

任务（只读，经 POST 创建、GET 查询；无更新/删除接口）。异步性说明：创建后立即返回（HTTP 200，body `status=running`），进度经 GET 轮询：

| 字段 | 类型 | 说明 | 可能取值 | 合法性条件 |
|------|------|------|----------|------------|
| `task_id` | string | 任务标识 | `rsp-…` | 系统生成；GET 路径入参 |
| `status` | string | 任务状态 | `running` / `succeeded` / `failed` | 系统维护；`failed` 含 `executor lost`（执行实例失联被接管） |
| `mode` | string | 行级变换方向 | `reencrypt` / `decrypt` | 创建时传入；`reencrypt`（默认）= 非目标形态行（明文/旧 keyID 密文）重写为 active 钥密文；`decrypt` = 全部密文行解密为明文，**仅用于回滚预案** |
| `dry_run` | bool | 是否只扫描统计 | true / false | 创建时传入，默认 false；`true` 不写库 |
| `scope` | string | 扫描范围 | `all` / `providers` / `api_keys` | 创建时传入，默认 `all` |
| `active_key_id` | int | 任务启动时 keyring 的 active 钥 | 0–255 | 系统记录；`decrypt` 模式为 0。**触发响应携带以便立即核对"重写目标钥"** |
| `started_at` / `finished_at` | string | 起止时间（RFC3339） | - | 系统维护；running 时 `finished_at` 为空 |
| `duration_ms` | int | 任务耗时（毫秒） | ≥0 | running 时为当前已耗时 |
| `summary` | object | 按表分组计数，结构见下 | - | 系统维护 |
| `error` | string | 失败归因 | - | 仅 `failed` 时非空；**不含任何密钥或密文材料** |

`summary` 结构：

```json
{
  "providers": { "scanned": 120,  "rewritten": 118,  "skipped": 2 },
  "api_keys":  { "scanned": 2400, "rewritten": 2390, "skipped": 10 }
}
```

| 字段 | 说明 |
|------|------|
| `scanned` | 已扫描行数 |
| `rewritten` | 已重写行数；`dry_run` 下为"将要重写"数 |
| `skipped` | 已是目标形态（reencrypt：active-keyID 密文；decrypt：明文） |

**约束**

- **集群级单例互斥**：任一时刻仅一个任务运行；运行中重复触发返回 409 并携带
  持有者 `task_id`（直接轮询该任务，不排队）。
- **幂等**：任务完成后可重触发（空跑 `rewritten=0` 即收敛确认）；执行实例失联
  （心跳超时）由后续触发自动接管，旧任务标记 `executor lost`。
- 任务记录持久化于 `keyrotate_sweep_tasks` 表，历史保留（无 TTL）；多实例部署下
  任意实例均可查询任意任务。
- 权限：Feature `FeatureSecurity`（scope=System），POST 需 `update`、GET 需
  `read`；触发与完成均写操作日志（含 `mode`/`dry_run`/`scope` 与结果计数）。

## 2. POST /open-api/v1/security/reencrypt-sweeps

触发收敛任务。

**请求体**（JSON，全可选）：

| 字段 | 类型 | 必填 | 默认 | 合法性条件 |
|------|------|------|------|------------|
| `mode` | string | N | `reencrypt` | 仅 `reencrypt` \| `decrypt` |
| `dry_run` | bool | N | `false` | - |
| `scope` | string | N | `all` | 仅 `all` \| `providers` \| `api_keys` |

**成功响应**：HTTP 200，`ErrNum=200`（`xreq` 将 2xx 统一归一为 200，异步语义由 body `status=running` 表达），`Data` 为任务对象（此时 `status=running`，
含 `task_id`、`mode`、`dry_run`、`scope`、`active_key_id`）。

**错误码**：401（未认证）；402（无 FeatureSecurity 权限）；409（已有任务运行，
归因含持有者 `task_id`）；422（`mode`/`scope` 非法，字段级归因）。

## 3. GET /open-api/v1/security/reencrypt-sweeps/{task_id}

查询任务状态与实时进度。

**路径参数**：

| 参数 | 类型 | 合法性条件 |
|------|------|------------|
| `task_id` | string | 非空；POST 响应获得 |

**成功响应**：HTTP 200，`Data` 为任务对象（字段见 §1，含实时 `summary` 与
`duration_ms`）。

**错误码**：401 / 402 同 POST；404（`task_id` 不存在）。

## 4. 示例

```json
// POST → 200（异步，body status=running）
{"ErrNum": 200, "ErrMsg": "success",
 "Data": {"task_id": "rsp-01J8XQ…", "status": "running", "mode": "reencrypt",
          "dry_run": false, "scope": "all", "active_key_id": 2}}

// GET（succeeded）→ 200
{"ErrNum": 200, "ErrMsg": "success",
 "Data": {"task_id": "rsp-01J8XQ…", "status": "succeeded", "mode": "reencrypt",
          "dry_run": false, "scope": "all", "active_key_id": 2,
          "started_at": "2026-10-05T10:00:00Z", "finished_at": "2026-10-05T10:00:08Z",
          "duration_ms": 8300,
          "summary": {"providers": {"scanned": 120, "rewritten": 118, "skipped": 2},
                      "api_keys":  {"scanned": 2400, "rewritten": 2390, "skipped": 10}},
          "error": ""}}
```
