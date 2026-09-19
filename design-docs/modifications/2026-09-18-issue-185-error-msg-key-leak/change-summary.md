# Issue #185：审计日志 error_msg 回显 API Key 明文 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/185](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/185)

> 对已存在的 API Key 发起重复创建（`POST /open-api/v1/api-keys`，提交与现存 Key 相同的 `key` 值）时，网关返回失败并写入一条 `action=create, status=2` 的 operation-log 审计记录。该失败审计记录的 `error_msg` 字段以明文形式回显了提交的完整 API Key 值（错误消息形如 `"API-Key value <裸 Key> already exists"`），而 `change_summary` 的脱敏机制（`MaskSensitiveFields`）只作用于 `before/after` 摘要，不覆盖 `error_msg` 字段，导致裸 API Key 长期留存于追加式审计日志中，低权限审计角色即可读取有效凭证。

与 issue #162（change_summary 经 token 键泄漏，已修复）同族但为**不同泄漏向量**：#162 = `change_summary` 键清单遗漏；本卡 = `error_msg` 自由文本回显绕过脱敏。

违反的产品条款（`design-docs/api-define/OpenAPI接口定义/operation-logs.md`）：

- 第 47 行：`error_msg` 语义为「失败时的简要错误信息」（成功时为空）；
- 第 55 行：「`change_summary` 中已对 api-key token、密码、证书私钥等敏感字段进行脱敏，**不会记录原始敏感信息**」——总括语义覆盖审计记录全字段，`error_msg` 回显完整凭证构成违约（对应 SC2101-TC046 第 69 行「原始 API Key 不进入日志」断言，fp `34e8e993`）。

## 2. 根因（源码定向定位）

**写入侧拼了原文**（`model/api_key/api_key.go`，CreateAPIKey 全局唯一性检查段）：

- `:686`：`return xerror.WrapParamErrorWithMsg("API-Key value %s already exists", *param.Key)` —— 裸 `*param.Key` 拼入 param 错误，经 `lib/xerror/resolve.go:57`（`Msg = errors.Cause(err)`）**同时回显在 HTTP 422 响应体与审计 `error_msg`**；
- `:678`：`return xerror.WrapDirtyDataErrorWithMsg("%s", fmt.Sprintf("API-Key-Token:%s", *param.Key))` —— 脏数据路径同构泄漏（消息形态 `Model.DirtyData: API-Key-Token:<裸 Key>`）。

**持久化侧只截断不脱敏**（`model/api_key/api_key_operation_log.go:50`）：`recordAPIKeyOperation` 将 err 经 `ioperlog.TruncateErrorMessageDefault(err)` 写入 `error_msg`——该函数仅做 1024 字符截断（`model/ioperlog/helper.go:25`），无任何脱敏；`:64` 的 `BuildChangeSummary` → `MaskSensitiveFields` 只覆盖 `change_summary` map。脱敏咽喉点对自由文本 `error_msg` 不存在。

**泄漏扇出（比 issue 描述更广）**：创建失败时同一个 err 还被传给嵌套审计——`api_key.go:729-730` `auditNestedQuotaPlanCreate` / `auditNestedRateLimitPolicyCreate`，二者分别在 `model/quota/operation_log.go:35`、`model/rate_limit_policy/operation_log.go:35` 以同构 `TruncateErrorMessageDefault(err)` 写入各自资源的失败审计。即一次「重复创建（body 含 quota_plan / rate_limit_policy）」最多产生 **3 条**携带同一裸 Key 的审计记录（api_key + quota_plan + rate_limit_policy）。

**Update 路径不受影响**：`UpdateAPIKey` 在 `api_key.go:549` 将 `param.Key = nil`（key 经更新接口不可变），唯一性检查仅存在于 Create 路径。

**同类排查结论（全 model 域 `WrapParamErrorWithMsg`/`WrapDirtyDataErrorWithMsg` 逐一核对）**：其余域错误消息只回显名称/ID/协议等非凭证值，如 provider `duplicate key name: %s`（`model/iprovider/provider.go:534`）只含 key **名**而非 key 值；`validate.APIKeyValue`（`lib/validate/validate.go:331`）格式错误不回显提交值。**当前唯一把凭证值拼入错误消息的是 api_key 域这两处**。但 DAO 层未来若返回含提交值的驱动错误（如唯一索引竞态下 MySQL duplicate-entry），`error_msg` 仍可能被污染——这是汇点兜底的必要性所在。

## 3. 目标

1. api_key 创建失败（含 param/脏数据两条路径）的审计记录 `error_msg` **不含裸 Key 全量明文**，掩码形态与 `change_summary` 契约一致（首 4 + `****` + 尾 4，短值全掩码 `******`），明文不可还原；
2. 嵌套 quota_plan / rate_limit_policy 失败审计的 `error_msg` 同步不再含裸 Key（扇出一并收口）；
3. `error_msg` 获得与 `change_summary` 同级的汇点脱敏能力（值级替换），兜住未来新增的错误回显路径与 DAO 驱动错误；
4. HTTP 422 响应 `msg` 由裸 Key 变为掩码形态——调用方提交者本人可见性无损失，且与脱敏契约一致；
5. 回归验证通过：本地单测与 model 覆盖率门禁（≥70%）、SC2101-TC046 requeue real-verification（失败审计日志任意字段不含裸 Key）。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `model/api_key/api_key.go`（源点掩码 2 处）、`model/api_key/api_key_operation_log.go`（汇点值级脱敏）、`model/ioperlog/mask.go`（新增 `MaskErrorMessage` 助手）、对应 `_test.go` |
| 文档同步 | `operation-logs.md:55` 契约注记补 `error_msg` 同等脱敏表述；`design-docs/sys-design/details/操作日志模块.md` 脱敏规则表加 `error_msg` 行 |
| 接口契约 | POST /open-api/v1/api-keys 422 响应 `msg` 中 Key 由明文变为部分掩码——属兑现既有脱敏契约，非新契约；响应结构与错误码（etParam 422 / dirty-data）不变 |
| 不在范围 | 其他域的汇点值级脱敏推广（当前无泄漏写入方，记录为可选后续加固模式）；ai-gateway-web 展示适配（前端原样渲染）；历史审计行明文清洗（见 §8，运维定夺） |
| 数据迁移 | 无（可选一次性历史清洗，见 §8） |

## 5. 最终方案

两层修复：**源点掩码（根因，必做）+ 汇点值级脱敏（纵深，必做）**。仅做源点无法兜 DAO 驱动错误等未来路径；仅做汇点则 HTTP 响应仍回显裸 Key 且嵌套审计链条依赖各域自觉——两层互补，缺一不可。

### 5.1 源点掩码（`model/api_key/api_key.go`）

`:686`：

```go
return xerror.WrapParamErrorWithMsg("API-Key value %s already exists", ioperlog.MaskAPIKeyToken(*param.Key))
```

`:678`：

```go
return xerror.WrapDirtyDataErrorWithMsg("%s", fmt.Sprintf("API-Key-Token:%s", ioperlog.MaskAPIKeyToken(*param.Key)))
```

- `ioperlog.MaskAPIKeyToken` 为既有导出函数（`model/ioperlog/mask.go:27`，首 4+`****`+尾 4，≤8 字符返回 `******`），与 `change_summary` 对 `key` 的部分掩码契约逐字一致；`api_key` 包已依赖 `model/ioperlog`（`api_key_operation_log.go:22`），无新依赖；
- **决策：部分掩码而非全掩码/删除值**。理由：① 与 change_summary 既有 `key` 契约同源同形，避免同一凭证在一条审计记录里两种掩码形态；② 管理员可凭首尾 4 位辨识「哪一把 Key 冲突」，排障信息保留；③ 掩码后明文不可还原（中间字符全量遮蔽），满足「原始 API Key 不进入日志」条款；
- 效果：HTTP 422 响应 `msg`、api_key 审计 `error_msg`、嵌套 quota_plan / rate_limit_policy 审计 `error_msg` 四处同一次修复全部收口（嵌域拿到的 err 本身已脱敏）。

### 5.2 汇点值级脱敏（`model/ioperlog/mask.go` 新增助手）

```go
// MaskErrorMessage redacts known sensitive values (e.g. API-Key values) from a
// free-text error message by replacing each occurrence with the masked token
// form (first 4 + "****" + last 4, "******" for short values), mirroring the
// change_summary masking contract. Values that are empty are skipped.
func MaskErrorMessage(msg string, values ...string) string {
	for _, v := range values {
		if v == "" {
			continue
		}
		msg = strings.ReplaceAll(msg, v, MaskAPIKeyToken(v))
	}
	return msg
}
```

- **值级精确替换**（非模式匹配）：调用方持有确切敏感值（`apiKey.Key`），`ReplaceAll` 只命中真实回显点，不依赖「API Key 长什么样」的正则启发式——自由文本脱敏在键名清单范式内无解（#162 §6.2 已知限制的镜像问题），值级传参是本范式下的可靠解；
- 短值走 `MaskAPIKeyToken` 的 `******` 全掩码分支，不留「≤8 字符 Key 明文残留」角落；空值跳过避免 `ReplaceAll` 空串插入行为；
- 脱敏在截断**之后**执行（只缩短、不回退截断语义）。

### 5.3 汇点接线（`model/api_key/api_key_operation_log.go:48-51`）

```go
if err != nil {
	status = ioperlog.StatusFailed
	errorMsg = ioperlog.TruncateErrorMessageDefault(err)
	if apiKey.Key != nil && *apiKey.Key != "" {
		errorMsg = ioperlog.MaskErrorMessage(errorMsg, *apiKey.Key)
	}
}
```

- `recordAPIKeyOperation` 的 `apiKey` 形参在三处调用点分别承载 `param`（创建）与 `oldAPIKey`（更新/删除），`apiKey.Key` 恰好覆盖「创建提交的 Key」与「存量 Key」两个值域；更新路径 `param.Key` 虽被置 nil，存量 Key 的替换仍作为对未来错误路径的兜底；
- 效果：即便未来新增的错误消息（含 DAO 驱动错误）回显该 Key，落库前必经此咽喉点脱敏。

### 5.4 单元测试

1. `model/ioperlog/mask_test.go` 新增 `TestMaskErrorMessage`：长值部分掩码、≤8 字符短值全掩码、空值跳过、多值依次替换、无命中原文不动、掩码结果不含原值子串；
2. `model/api_key/api_key_test.go` 既有重复创建用例（`:505-521`）加强：现断言 `err.Error()` 含 `"already exists"`，追加断言**不含裸 Key**、含 `MaskAPIKeyToken` 形态；新增脏数据路径（`FetchAPIKeyTokenList` 返回 >1 条）同构断言；
3. 仿 `nested_audit_test.go` 的 fake auditor / mocks 模式新增集成单测：捕获 `OperationLogEntry`，走「重复创建（含 quota_plan + rate_limit_policy）」失败路径，断言 api_key、quota_plan、rate_limit_policy 三条记录的 `ErrorMsg` 序列化后均不含裸 Key；
4. 更新失败路径用例：断言经 `oldAPIKey.Key` 汇点脱敏后 `error_msg` 不含存量 Key 明文（正常更新错误本不含 Key，验证接线不误伤、不 panic）。

### 5.5 文档同步

- `operation-logs.md:55` 注记扩展为：「`change_summary` 中已对 api-key token、密码、证书私钥等敏感字段进行脱敏；`error_msg` 同样不回显原始敏感信息（凭证值经掩码后记录）。不会记录原始敏感信息。」——将 #185 判定的总括语义显式化，防回归；
- `design-docs/sys-design/details/操作日志模块.md` 脱敏规则表（#162 已补 token 键与数组递归行）再加一行：`error_msg` 自由文本 | api_key 域经 `MaskErrorMessage` 值级替换（首 4+尾 4 部分掩码），其余域暂无凭证值写入方。

## 6. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`go test -cover ./model/...`（≥70% 门禁）全部通过；
2. 单元级：§5.4 全部新用例通过，既有 mask/change_summary/api_key 用例不回退（`api_key_test.go:520` 的 `"already exists"` 断言在掩码后依然成立）；
3. 部署门禁：SC2101-TC046 requeue real-verification——
   - 重复创建失败审计记录整体 JSON 序列化后 `bytes.Contains(..., K)` 不命中（A-failed-writes 断言转绿）；
   - `change_summary.after.key` 保持部分掩码形态不回退；
   - HTTP 422 响应 `msg` 呈掩码形态、错误码 etParam/422 不变；
   - 带 quota_plan / rate_limit_policy 的重复创建：三条失败审计（api_key、quota_plan、rate_limit_policy）`error_msg` 均不含裸 Key；
4. 上线后抽样：`GET /open-api/v1/operation-logs?resource_type=api_key&action=create&status=2` 近 N 日记录，人工抽查 `error_msg` 无裸 Key 形态（明文特征：长度 >9 且首尾 4 位外含非 `*` 字符）。

## 7. 备选方案（不采纳）

| 备选 | 不采纳理由 |
|------|-----------|
| 仅源点掩码，不建汇点助手 | DAO 驱动错误、未来新增错误路径的 `error_msg` 污染无兜底；嵌套审计链依赖各域自觉，与本仓库「脱敏收敛在咽喉点」的既定架构原则（#162 §6）相悖 |
| 仅汇点脱敏，不改源点 | HTTP 422 响应持续回显裸 Key；响应与审计两套形态不一致；脏数据路径的嵌套扇出仍在 |
| 全掩码 `******`（不留首尾 4 位） | 与 change_summary 对 `key` 的既有部分掩码契约分裂；管理员失去冲突 Key 辨识能力；若产品改判凭证连首尾都不保留，应在 masker 引入按域配置并另立 issue（同 #162 §6.2 结论） |
| 在 `TruncateErrorMessageDefault` 内做正则模式脱敏 | API Key 为任意 `[A-Za-z0-9_-]{1,128}` 字符串（`validate.go:332-337`），无可靠模式可匹配；误掩/漏掩双高，值级传参才是自由文本的正解 |

## 8. 历史数据处置（建议，可选）

代码修复只阻断新增泄漏；已落库的 `error_msg` 明文（`resource_type=api_key/quota_plan/rate_limit_policy`、`action=create`、`status=2`、消息形态 `API-Key value <K> already exists` 或 `API-Key-Token:<K>`）随审计行无限期留存。`operation_logs.error_msg` 为 varchar(1024)（`db_ddl.sql`）。建议：

- 由运维定夺执行一次性清洗：按消息形态正则定位含疑似 Key 的失败行，对该行执行值级掩码回写（清洗脚本直接 import 本仓库 `ioperlog.MaskAPIKeyToken`，保证规则与线上代码一致）；或按数据敏感度直接删除相应历史行；
- 执行前备份该表；清洗动作本身产生的新审计噪音可忽略；
- 若评估历史明文暴露面可接受（审计库访问已受控），也可选择仅修复新增、留存历史，但需在 issue #185 中显式记录该决策。
