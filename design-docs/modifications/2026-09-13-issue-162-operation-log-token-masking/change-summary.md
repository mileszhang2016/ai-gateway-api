# Issue #162：审计日志 change_summary 泄漏 Token/凭证明文 修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/162](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/162)

> POST /open-api/v1/auth/tokens 创建 Token 与 DELETE /open-api/v1/auth/tokens/{name} 删除 Token 的审计日志中，change_summary.after.token（创建）与 change_summary.before.token（删除）均为未脱敏的 Token 明文，与创建接口响应返回的凭证逐字符一致。密码路径不受影响（password/session_key 键命中脱敏清单）。

手工复现步骤（测试环境实测，2026-09-12）：

1. POST /open-api/v1/auth/tokens，body `{"name":"tc049probe-","scope":"Support"}` → 200，响应 token 字段为 20 字符明文；
2. GET /open-api/v1/operation-logs?resource_type=token&action=create → 日志 id=29198，change_summary.after 键为 [id,name,scope,token]，其中 token 与响应明文逐字符一致（sha256 相同）；
3. DELETE /open-api/v1/auth/tokens/tc049probe- → 日志 id=29200，change_summary.before.token 仍为同一明文——Token 已删除，凭证原文却永久留存在追加式审计日志中。

影响：Token 是有效管理面凭证（scope 支持 System/Support），任何具备 operation-log 查询权限的角色（通常为审计/低权限角色）可从日志直接提取可用 Token 冒用创建者身份，形成经审计通道的提权；删除 Token 不能止损，明文随审计行无限期留存。属 CWE-532（敏感信息写入日志文件），违反 operation-logs.md 自身的敏感字段脱敏契约（`design-docs/api-define/OpenAPI接口定义/operation-logs.md:55`："已对 api-key token、密码、证书私钥等敏感字段进行脱敏"）。

## 2. 根因

与 issue 分析一致，两层链路均实证到代码：

**写入侧带了原文**：`model/iauth/operation_log.go:160` `tokenToMap` 将裸 `token.Token` 以 `"token"` 为键放入 map；`:179` `tokenParamToMap` 对 `param.Token` 同构（创建失败路径 `authentication.go:513` 经此写入，正常创建中 param.Token 为 nil，但内部调用方一旦传值即泄漏）。四处调用点：`authentication.go:435/:439`（删除，作为 before）、`:513`（创建失败）、`:521`（创建成功，作为 after）。

**脱敏侧键清单缺 token**：`model/ioperlog/mask.go:50` 全掩码清单仅含 password/secret/session_key/sessionkey/private_key/privatekey，部分掩码 api_key/apikey/key（`:52`），certificate 族打标（`:56`）——`"token"` 键不在任何分支，字符串值原样透传。

**脱敏咽喉点确认**：所有审计写入统一经 `ioperlog.BuildChangeSummary`（`model/ioperlog/change_summary.go:41-45`）在持久化前调用 `MaskSensitiveFields`；`diff_keys` 仅写键名、不写值（`change_summary.go:57-76`），无值泄漏旁路。对照 userToMap/userParamToMap 的 password/session_key 均命中清单——证明脱敏机制本身有效，纯属键名清单遗漏。

## 3. 同族缺陷（同批修复）

按 issue 修复建议 §2 全量排查了所有 `recordXxxOperation` 写入路径的 map 键（11 个接入域逐一核对），实证到第二处泄漏及若干结构性缺口，一并收口：

### 3.1 Provider 上游凭证泄漏（代码链路实证，比 token 泄漏影响面更大）

`iprovider.Provider` / `ProviderParam` 整结构经 `json.Marshal` 进入审计 map（`model/iprovider/provider_operation_log.go:52-86` 的 `providerParamToMap`/`providerToMap`），其中 `Keys []ProviderKey`（`model/iprovider/provider.go:54-57`，元素为 `{name, key}`）的 `key` 是 **Provider 上游真实凭证**（≤512 字符，数据面按 `Authorization: Bearer …` 使用，`provider.go:786-797`）。写入点覆盖 provider Create/Update/UpdatePricingTiers/Delete 全路径（`provider.go:165,183,187,201,264,268,306,314,342,346`）。

泄漏原因叠加了两层：其一，`keys` 值为 `[]interface{}` 数组，而 `MaskSensitiveFields` 的 default 分支只递归 `map[string]interface{}`（`mask.go:58-63`），**数组永不进入递归**；其二，即便进入递归，元素内 `key` 只配置部分掩码（首 4+尾 4），数组场景下连这部分也未生效。净效果：provider 上游 API key 明文随每次 provider 写操作进入审计库，读取审计日志即可提取可用上游凭证。

### 3.2 排查确认干净的路径（本次不改）

- **cluster**：`LLMConfig.Keys []ClusterKeyRef` 仅 `{name, weight}` 引用（`model/icluster_conf/cluster.go:227`），不携带凭证明文，设计如此。注意 `Cluster`/`ClusterParam` 顶层字段无 json tag，Go 按 PascalCase 输出——masker 匹配前 `ToLower`（`mask.go:47`），本次补入 `token` 键后可兜住未来以 `Token` 形态出现的同类字段。
- **api_key**：`APIKeyParam.Key json:"key"`（`model/api_key/api_key.go:42`）走部分掩码，命中现有契约。
- **certificate**：`CertificateParam` 的 PEM 正文在写审计 map 时被刻意排除（`model/iprotocol/operation_log.go:52-78` 仅记文件路径），inline after-map 只含 cert_name/is_default。
- entity / quota / rate_limit_policy / route_rules / iroute_conf / imodel_price：map 键逐一核对，无凭证类键（imodel_price `operation_log.go:76` 直调 MaskSensitiveFields 的 `{mode, imported, entries}` 亦无敏感值）。

### 3.3 掩码清单同族缺口（无现行写入方，防御性补齐）

现有清单对 secret 族只收精确 `secret`，对 token 族只收无。无当前写入方使用下列键，但属同族凭证键，借本次"一次收口"防御性纳入全掩码清单：`access_token`、`refresh_token`、`session_token`、`id_token`、`secret_key`/`secretkey`、`client_secret`/`clientsecret`、`access_key`/`accesskey`。清单维持"小写化后精确匹配"语义不变（`session_key_create_at` 这类时间戳字段不受影响）。

## 4. 目标

1. Token 创建/删除/失败审计日志中，`token` 值全掩码（`******`），明文不可还原；change_summary 键集合 `[id,name,scope,token]` 保持不变；
2. Provider 审计日志中 `keys[].key` 按 `key` 既有契约部分脱敏（首 4 位+`****`+尾 4 位），明文不可还原，`name`/`weight` 等非敏感字段原样保留；
3. `MaskSensitiveFields` 支持数组递归，数组内 map 的敏感键与顶层同规则命中；
4. 密码/证书/session_key 路径脱敏行为不变，不引入新的落库值旁路；
5. 回归验证通过：本地单测与 model 覆盖率门禁（≥70%）、SC2101-TC049 requeue real-verification（token 创建/删除日志无明文、password 路径保持干净），并补 provider keys 审计断言。

## 5. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `model/ioperlog/mask.go`（键清单 + 数组递归）、`model/ioperlog/mask_test.go`（新增用例）；`model/iauth`、`model/iprovider` **写侧不改**（脱敏在咽喉点统一完成） |
| 文档同步 | `design-docs/sys-design/details/操作日志模块.md` §6 脱敏规则表、`design-docs/sys-design/模型层设计文档.md` §4.25.3 同表，补 `token` 行与"数组递归"行 |
| 接口契约 | GET /operation-logs 的 change_summary 中 `token` 值由明文变为 `******`、`keys[].key` 变为部分掩码——响应结构不变，属兑现既有契约（operation-logs.md:48/55 已承诺"敏感字段已脱敏"），非契约变更 |
| 不在范围 | ai-gateway-web 展示适配（前端原样渲染脱敏值，无需改）；历史审计行明文清洗（见 §8 处置建议，是否执行由运维定夺）；dashboard/BFE 等其余仓库 |
| 数据迁移 | 无（可选一次性历史清洗脚本，见 §8） |

## 6. 最终方案

核心思路：脱敏是持久化前唯一咽喉点（`BuildChangeSummary` → `MaskSensitiveFields`），所有修复收敛在 masker 一处，写侧与 change_summary 结构不动。

### 6.1 全掩码清单补 `token` 及同族键（`model/ioperlog/mask.go:50`）

```go
switch lowerKey {
case "password", "secret", "session_key", "sessionkey", "private_key", "privatekey",
    "token", "access_token", "refresh_token", "session_token", "id_token",
    "secret_key", "secretkey", "client_secret", "clientsecret",
    "access_key", "accesskey":
    data[key] = maskPlaceholder
case "api_key", "apikey", "key":
    ... // 不变
```

**决策：token 走全掩码，不走 `MaskAPIKeyToken` 部分掩码**（issue 修复建议 §1 把选择权留给了方案）。理由：① token 为服务端生成的 20 位高熵凭证，暴露首尾 8 位泄漏约 40% 字符，而消费侧 API key 保留片段是为管理员辨识"哪一把 key"，auth token 的辨识由同对象 `name` 字段承担，无需保留片段；② token 与 password/session_key 同属认证凭证族，应同规则；③ 与 operation-logs.md:55 契约表述（token 与密码、证书私钥并列）一致。

### 6.2 masker 支持数组递归（`model/ioperlog/mask.go` default 分支）

```go
default:
    // Recurse into nested maps and maps inside arrays.
    switch typed := val.(type) {
    case map[string]interface{}:
        data[key] = MaskSensitiveFields(typed)
    case []interface{}:
        data[key] = maskSlice(typed)
    }
}

// maskSlice masks sensitive fields in maps nested inside a slice.
func maskSlice(items []interface{}) []interface{} {
    for i, item := range items {
        switch v := item.(type) {
        case map[string]interface{}:
            items[i] = MaskSensitiveFields(v)
        case []interface{}:
            items[i] = maskSlice(v)
        }
    }
    return items
}
```

效果：provider `keys: [{name, key}]` 中元素 map 进入递归，`key` 命中既有部分掩码分支 → `abcd****wxyz`，`name` 保留。该改动对全域生效（任何域的数组内敏感键一并兜住，含数组套数组），不影响非敏感键透传。

备选方案（不推荐）：provider 域在 `providerParamToMap`/`providerToMap` marshal 前将 `ProviderKey.Key` 预置 `******` 全掩码。缺点：`key` 的全/部分掩码语义将按域分裂，且整结构 marshal 方式下新增凭证字段仍会漏——不如递归方案通用，不采纳。若产品后续认定上游凭证连首尾 4 位都不应保留，可在 masker 引入按域配置，另立 issue。

已知限制（接受）：掩码以键名精确匹配为前提，凭证若被写成匿名字符串数组（如 `["sk-xxx"]` 挂在非敏感键下）仍无法兜住——现行写入方无此形态，且在"键名清单"范式内无解，靠代码评审约束 `XxxToMap` 不引入此类结构。

### 6.3 写侧不改（iauth / iprovider 零改动）

`tokenToMap`/`tokenParamToMap` 保留 `token` 键原文写入，经咽喉点脱敏后落库为 `******`。理由：change_summary 键集合稳定（SC2101-TC049 对 after 键集合 [id,name,scope,token] 的断言不受影响），键存在即证明字段已设置；写侧删键会牵连测试断言与下游键集合校验，收益为零。写侧双保险（`tokenToMap` 直接置 `maskPlaceholder`）可作为可选项，非必须。

### 6.4 单元测试（`model/ioperlog/mask_test.go` 新增用例）

1. `token` 顶层键 → `******`，断言结果不含原文子串；
2. 嵌套 map 内 `token` 键 → `******`（验证递归路径不回退）；
3. `keys: []interface{}{map[string]interface{}{"name": "k1", "key": "abcdefghijkl"}}` → `key` 变为 `abcd****ijkl`、`name` 保留（6.2 的 provider 场景）；
4. 数组内嵌 `token` 键、数组嵌数组场景；
5. 既有 password/api_key/certificate 断言保持不动（回归保护）。
另仿照 `model/entity/operation_log_test.go` 等现有模式，为 iauth（CreateToken/DeleteToken/CreateToken 失败路径）与 iprovider（CreateProvider/UpdateProvider）各补 mock recorder 集成单测：捕获 `OperationLogEntry`，断言 `ChangeSummary` JSON 序列化后不含凭证原文。

### 6.5 文档同步

- `design-docs/sys-design/details/操作日志模块.md:146-151` 与 `design-docs/sys-design/模型层设计文档.md:513-518` 脱敏规则表：全掩码行补 `` `token`、`access_token`、`client_secret` 等凭证键 ``，并加一行"`[]interface{}` 数组 | 逐元素递归，元素为 map 时同规则脱敏"；
- `operation-logs.md:55` 契约措辞已成立，无需改动。

## 7. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`go test -cover ./model/...`（≥70% 门禁）全部通过；
2. 单元级：§6.4 全部新用例通过，既有 mask/change_summary 用例不回退；
3. 部署门禁（issue 修复建议 §4）：SC2101-TC049 requeue real-verification——
   - token 创建日志：change_summary.after.token 为 `******`，与响应明文 sha256 不一致；
   - token 删除日志：change_summary.before.token 为 `******`；
   - password/session_key 路径保持 `******`，行为不回退；
   - 新增断言：provider 创建/更新日志 change_summary 中 `keys[].key` 不含上游凭证明文（部分掩码形态）；
4. 上线后抽样：GET /open-api/v1/operation-logs?resource_type=token|provider 近 N 日日志，人工抽查无明文。

## 8. 历史数据处置（建议，可选）

代码修复只阻断新增泄漏；已落库明文（测试环境实证 id=29198/29200 的 token 明文，及历史上全部 provider keys[].key 明文）随审计行无限期留存，删除 Token/Provider 不能止损。`operation_logs.change_summary` 为 mediumtext（`db_ddl.sql:516`）。建议：

- 各环境由运维定夺执行一次性清洗：扫描 `resource_type ∈ {token, provider}` 的行，`json.Unmarshal` → 复用新版 `ioperlog.MaskSensitiveFields` → 回写（可写临时 tools 脚本直接 import 本仓库 masker，保证规则与线上代码一致）；或按数据敏感度直接删除相应历史行；
- 清洗动作本身会产生新的审计噪音可忽略；执行前备份该表；
- 生产环境若评估历史明文暴露面可接受（审计库访问已受控），也可选择仅修复新增、留存历史，但需在 issue 中显式记录该决策。
