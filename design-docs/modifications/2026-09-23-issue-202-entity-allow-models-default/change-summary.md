# Issue #202：创建 Entity 省略 `allow_models` 默认值错误（`[]` 应为 `["*"]`）修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api#202](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/202)（P1，SC1203-TC009 A-02 create-fixtures，失败指纹 `64e64377`，verified_commit `ef68995`）：

> 创建 Entity 时省略 `allow_models` 字段，产品设计规定默认值为"允许访问所有模型"，即响应与持久化值应为 `["*"]`。实际产品在**创建响应**与**查询回读**两个通道均返回 `allow_models: []`（空数组），违反 §2.1 的默认值契约。

违反的产品条款（均为 released / normative，**修复的规范性依据，无需改 api-define 设计文档**）：

- `entities.md` §1 字段说明（line 51）："`allow_models` | []string | 允许访问的模型白名单 | 包含"*"表示允许访问所有模型，**默认值为允许访问所有模型**"。
- `entities.md` §2.1 创建 Entity 输入参数（line 84）："`allow_models` | []string | 非必填(N) | …，**默认值为允许访问所有模型**"。
- `entities.md` §2.1 创建成功返回示例（line 177）：`"allow_models": ["*"]`（成功响应固定示例）。
- 对照条款：`block_models` 默认值**为空数组**（line 52/85），实际返回 `[]` 符合，本次不动。

## 2. 根因（源码定向定位）

**三层链路无一处兑现 `["*"]` 默认值**：

1. **接口层无默认回填**。`endpoints/openapi_v1/entity/create.go:85-116` 对 `QuotaPlan`/`RateLimitPolicy`/`RouteRules` 三个省略字段逐一构造默认值，**唯独没有 `AllowModels`**——省略时保持 Go 零值 nil 切片原样下发。
2. **存储层创建转换把"省略"固化成 `"[]"`**。`storage/rdb/entity/entity.go:149-155`（`entityDataToParam`，Create 专用转换，#151 拆分 create/update 路径时引入）：

   ```go
   // 转换 AllowModels 为 JSON 字符串（创建省略时默认空数组）
   if len(param.AllowModels) > 0 {
       allowModelsJSON, _ := json.Marshal(param.AllowModels)
       data.AllowModels = lib.PString(string(allowModelsJSON))
   } else {
       data.AllowModels = lib.PString("[]")   // ← 省略与显式 [] 一律落库 "[]"，即本缺陷
   }
   ```

   注释"创建省略时默认空数组"即错误默认值的产品化表述；省略（nil）与显式 `[]`（len=0）在此不可区分，统一落库 `"[]"`。
3. **读回路径原样透传**。`entityParamToData`（`storage/rdb/entity/entity.go:216-219`）把列内 `"[]"` 反序列化为空切片，JSON 响应即 `[]`——创建响应（create.go:122 创建后 `FetchEntity` 回读）与 GET 回读两个通道因此同时失守，与 issue 证据（entity-871 双通道均为 `[]`）吻合。

**对照基线（同构字段的正确实现）**：API-Key 的 `models` 字段语义与 Entity `allow_models` 完全同构（白名单、含 `"*"` 表全放行、创建省略默认 `["*"]`），其实现在 storager 创建转换 `storage/rdb/api_key/api_key.go:47-52`：

```go
models := []string{"*"}
if len(param.Models) > 0 {
    models = param.Models
}
```

且其**读路径同样归一化**（`api_key.go:129-137`：列空或反序列化结果为空时回填 `["*"]`）。Entity 的 `allow_models` 从引入起就缺了这两处，属实现与契约的原始偏差，而非后续回归。

**git 考古**：`9fa8c62`（issue #151，SC1101-TC019 家族）拆分 `entityDataToParam`/`entityDataToParamForUpdate` 时在注释中写下"创建省略时默认空数组"，但 `"[]"` else 分支在拆分前已存在（diff 中为上下文行、非新增），#151 只是把它隔离进 Create 专用路径并留注，不是引入者。

## 3. 影响面核实（Issue"待确认"项的裁决）

### 3.1 数据面执行语义：`[]` 与 `["*"]` 等价，无"全拒"风险

issue 提出"`[]` 在数据面是全拒还是无约束待确认"。经导出链路代码核实，**当前实现中二者等价，均为"无约束"**：

- `model/imods/exporter.go:514`：`fetchEntityModelHierarchy` 仅收集 `len(entity.AllowModels) > 0 && !containsStar(...)` 的白名单——空数组与 `["*"]` 一样被跳过，对交集零贡献；
- `exporter.go:317-320`（Rule 1）：API-Key 与 Entity 层级均无白名单约束时导出 `Models: ""`（全放行）。

因此 issue 中"新建 Entity 默认访问语义从全放行翻转成全不放行"的风险**在当前数据面不兑现**；本缺陷的实际危害是：① 管理面响应/持久化值违反 §2.1 明文契约；② 控制台/脚本/下游消费者按字段语义（白名单为空 = 不放行）误读 `[]`，可能在管理面做出错误处置（例如告警或手动改配），属契约完整性缺陷而非流量阻断缺陷——与 P1 定级吻合。

### 3.2 修复对数据面零影响

把创建默认值改为 `["*"]` 后，导出路径对 `["*"]` 的处置（containsStar 跳过）与对 `[]` 的处置逐字相同（exporter.go:514），存量与增量配置的导出产物**逐字节不变**，无需联动 BFE / conf-agent。

### 3.3 PUT/PATCH 不在影响面内

`entityDataToParamForUpdate`（update 专用转换）对省略的 `allow_models` 保持 nil 走 DAO nil-skip（#151/#152 契约：PATCH 省略保留原值），读路径归一化**不改变** param→DAO 的写入判定，PUT/PATCH 行为零变化。唯一可见差异：PUT/PATCH 的响应体来自更新后 `FetchEntity` 回读，存量 `"[]"` 行在响应中显示为 `["*"]`（见 §6.2），与契约一致。

## 4. 目标

1. `POST /entities` 省略 `allow_models` 时，创建响应与 `GET /entities/{id}` 回读均返回 `allow_models: ["*"]`，持久化值为 `["*"]`；
2. `block_models` 默认值保持 `[]` 不变；
3. 显式传入非空 `allow_models`（含含 `"*"` 的列表）行为不变；
4. 存量"省略创建"的 Entity 行（持久化 `"[]"`）回读契约合规（`["*"]`），无需数据迁移即可闭环；
5. 修复具备单测覆盖（storager 层，进 `make test` 门禁），SC1203-TC009 A-02 requeue 转绿，SC1203-TC028（#151 回归）与 entity 全模块集成测试零回退。

## 5. 范围

| 范围 | 说明 |
| - | - |
| 涉及仓库 | `ai-gateway-api` |
| 主要文件 | `storage/rdb/entity/entity.go`（唯一代码改动点：创建转换 + 读回归一化）、`storage/rdb/entity/entity_test.go`（既有用例纠偏 + 新增）、`test/integration/tests/entity/create/`（新增断言用例） |
| 设计文档 | api-define 契约（entities.md §1/§2.1）已明确且现行，**无需修改**；sys-design `details/部分更新语义与DAO-nil-skip约定.md` §2/§3 表述纠偏（该文档当前把 Entity 创建默认值误记为 `[]`，见 §6.5）；sys-design/summary.md 该条目的"默认值语义差异"描述仍成立，无需改 |
| 接口契约 | 无变化：省略字段响应值由 `[]` 修正为契约规定的 `["*"]`；成功/失败码型不变 |
| 数据面联动 | 无（§3.2 已证导出产物不变），无需联动 BFE |
| 不在范围 | ① PUT/PATCH 省略 `allow_models` 的语义（§8 备选）；② "显式 `[]`"与"省略"的区分（nil vs 空切片，§4 已知限制，需另立契约）；③ 存量数据物理迁移（可选，见 §6.4） |
| 数据迁移 | 无必需迁移；读回归一化已使存量行契约合规，物理对齐 SQL 为可选项 |

## 6. 最终方案

单层修复、两处改动，全部位于 `storage/rdb/entity/entity.go`，**逐字对称 API-Key `models` 的既有先例**（`storage/rdb/api_key/api_key.go:47-52` 与 `:129-137`）。

### 6.1 创建转换默认值纠偏（`entityDataToParam`，主修复）

`storage/rdb/entity/entity.go:149-155` 的 else 分支由 `"[]"` 改为 `["*"]`：

```go
// 转换 AllowModels 为 JSON 字符串（创建省略时默认 ["*]，即允许访问所有模型，api-define entities.md §2.1）
if len(param.AllowModels) > 0 {
    allowModelsJSON, _ := json.Marshal(param.AllowModels)
    data.AllowModels = lib.PString(string(allowModelsJSON))
} else {
    allowModelsJSON, _ := json.Marshal([]string{"*"})
    data.AllowModels = lib.PString(string(allowModelsJSON))
}
```

- 创建响应（create.go:122 回读）与 GET 回读双通道同时修复（值来自持久化，一处落库两处生效）；
- `BlockModels` 的 `"[]"` 默认值**保持不动**（契约规定默认空数组）；
- 调用方唯一（`EntityManager.CreateEntity` 全仓仅 `create.go:118` 一处生产调用），无 innerapi 旁路需要排查。

### 6.2 读回归一化（`entityParamToData`，存量行契约闭环）

`storage/rdb/entity/entity.go:216-219`，反序列化后追加空值归一化：

```go
// 解析 AllowModels（存量 "[]" 行按契约归一化为 ["*]，与 api_key 读路径 api_key.go:129-137 同构）
if one.AllowModels != "" {
    json.Unmarshal([]byte(one.AllowModels), &param.AllowModels)
}
if len(param.AllowModels) == 0 {
    param.AllowModels = []string{"*"}
}
```

- 修复前创建的存量行（持久化 `"[]"`）在 GET/list 响应中回读为 `["*"]`，无需迁移即满足"持久化默认值"契约的读侧；
- 对称 api_key 先例（`api_key.go:129-137`：列空或解析结果为空均回填 `["*"]`），两资源读行为拉齐；
- 对导出链路零影响：`fetchEntityModelHierarchy`（exporter.go:514）对 `["*"]` 与 `[]` 均跳过，归一化前后导出产物逐字节相同；
- 对 PATCH/PUT 写路径零影响：nil-skip 判定在 param→DAO 转换（`entityDataToParamForUpdate`），读路径归一化不触碰任何写入判定；
- 副作用说明：创建审计日志的 after 快照（`entityParamToMap`，`model/entity/operation_log.go:129-130`，仅 `len>0` 才收录）在创建瞬时仍不显示 `allow_models`（默认值在 storager 内回填，接口/模型层 param 未变）——与 api_key `models` 的现状一致，属可接受的一致性行为，不为审计展示单独上移默认值。

### 6.3 显式 `[]` 与省略的区分（维持现状，契约兜底）

创建时显式传 `"allow_models": []` 与省略同样落库 `["*"]`（`len==0` 不可区分）。该限制与 api_key `models`、PATCH 语义一致（《部分更新语义与DAO-nil-skip约定》§4"显式空数组无法与省略区分"），且契约只规定了"省略"的默认值——显式空白的语义（全拒？）本身未有产品定义，**不在本修复扩张**；若未来需要"显式全拒"，应另立契约（nil vs 空切片区分，参照 Provider `models`/`keys` 的 #147 先例）。

### 6.4 存量数据处置（可选，非必需）

读回归一化（§6.2）后，存量行的 API 响应已契约合规，**无需迁移**。若运维希望物理对齐持久化值（消除"库内 `[]`、读侧 `["*"]`"的表象差异），可按以下口径执行（MySQL，表定义 `db_ddl.sql:396`，列无 DDL 默认值、storager 创建时显式写入，直接 UPDATE 安全）：

```sql
UPDATE entities SET allow_models = '["*"]', updated_at = NOW()
WHERE allow_models = '[]';
```

注意：缺陷存续期内**不存在**"显式 `[]` 意图全拒"的存量行（创建路径对显式 `[]` 与省略同处置，§6.3），故该 UPDATE 不会误伤；执行前仍建议先 `SELECT ... WHERE allow_models = '[]'` 抽样确认。SQLite 部署同理。

### 6.5 文档与测试

1. **单测纠偏 + 新增**（`storage/rdb/entity/entity_test.go`，sqlite 内存库，进 `make test`）：
   - `TestCreateEntity_OmittedModelsDefaultToEmpty`（:134-149）为**缺陷固化用例**（`assert.Empty(t, one.AllowModels)`），改写为 `TestCreateEntity_OmittedModelsDefaultToStar`：断言创建省略时读回 `AllowModels == ["*"]`、`BlockModels` 仍为空；
   - 新增存量读回归一化用例：直接向库内写入 `allow_models='[]'` 的行（绕开 storager 创建转换），经 `FetchEntity` 断言读回 `["*"]`；
   - 新增显式非空用例保持既有断言（`AllowModels=["model-a"]` 原样往返，防误伤）；
   - 既有 update 族用例（`TestUpdateEntity_OmittedModelsPreserveValues` 等，#151 引入）不受影响，逐项复跑确认。
2. **集成测试**（`test/integration/tests/entity/create/`，需 live 环境）：新增用例——创建省略 `allow_models`，断言创建响应与 GET 回读均为 `["*"]`、POST 显式 `["gpt-4"]` 回读一致、`block_models` 均为 `[]`。该模块 design.md（`test/integration/tests/entity/design.md:91`）本就写着"默认 `["*"]`"，即测试设计文档一直是对的、代码是错的，本次为代码向设计看齐。
3. **E2E**：SC1203-TC009 A-02 requeue 转绿（`default allow_models=[] want [*]` 断言闭合）；回归 SC1203-TC028（PATCH 省略保留原值）。E2E 资产在 integration-test 仓库，无需本仓库改动。
4. **sys-design 文档纠偏**（`design-docs/sys-design/details/部分更新语义与DAO-nil-skip约定.md`）：
   - §2 表格 Entity 行："Create 仍走 `entityDataToParam` 默认 `"[]"`" → 默认 `["*"]`（issue #202）；
   - §3 表格 Entity 行：Create 省略 `allow_models` 默认 `[]` → `["*"]`；`block_models` 行保持 `[]` 不动；
   - §3 表后或 §4 补一句：Entity `allow_models` 创建默认值 `["*"]` 自 issue #202 修复起与 api_key `models` 对齐，"显式 `[]` 无法区分"限制对创建路径同样适用（显式 `[]` 按省略处置）。

## 7. 回归验证

1. 本地：`go build ./...`、`go vet ./...`、`go test ./...`、`go test -cover ./model/...`（≥70% 门禁）全部通过；
2. 单测：§6.5-1 新增/改写用例通过；**关键自证**——仅改 `entityDataToParam` 而未改测试时，`TestCreateEntity_OmittedModelsDefaultToEmpty` 必失败（读回 `["*"]` ≠ empty），可证该用例对修复有回归效力；#151 引入的 update 族用例零回退；
3. 手动对照（本地起服务）：
   - POST 省略 `allow_models` → 200，响应 `allow_models=["*"]`、`block_models=[]`；GET 回读一致；
   - POST 显式 `allow_models=["gpt-4"]` → 响应与回读均为 `["gpt-4"]`；
   - 修复前创建的存量行 GET 回读显示 `["*"]`；
4. 集成测试 §6.5-2 通过；导出冒烟：创建省略的 Entity 挂载 API-Key 后 inner 导出 mod-api-key，修复前后产物一致（`Models: ""`，Rule 1 全放行）；
5. 部署门禁：SC1203-TC009 requeue A-02 转绿 + SC1203-TC028 回归通过。

## 8. 备选方案（不采纳）

| 备选 | 不采纳理由 |
| - | - |
| 仅在接口层 `create.go` 回填（`if param.AllowModels == nil { param.AllowModels = ["*"] }`，仿 QuotaPlan 模式） | 可修复双通道响应，但与 api_key `models` 的先例（storager 创建转换持有默认值）不对称；默认值分散到接口层后，未来 innerapi 等新增 `CreateEntity` 调用方须各自记得回填，否则再次失守；审计快照倒是能显示 `["*"]`，但为此打破分层一致性不值 |
| 仅改 `entityDataToParam`、不加读回归一化 | 增量行合规，但存量 `"[]"` 行 GET 回读仍违契约（issue 预期表明确要求回读 `["*"]`）；且与 api_key 读路径（`api_key.go:129-137` 有空值回填）不对称 |
| 不做读回归一化，改跑物理迁移 UPDATE | 可行（§6.4），但迁移有时序与多部署面（MySQL/SQLite/在途实例）成本；读归一化为零迁移的等价闭环，迁移降级为可选项 |
| 将读回归一化下沉到 `dao` 层或数据库视图 | 越过 storager 分层，DAO 是纯机械映射（与"storager 持有转换语义"的仓库惯例相悖，参见《部分更新语义与DAO-nil-skip约定》§1） |
| 顺带把 PUT 省略 `allow_models` 也默认为 `["*"]`（§2.4"Body 同创建"的严格全量解读） | §2.4 仅对 `description` 明示"省略清空"，对 `allow_models` 未规定省略语义；#151 家族约定 PATCH 省略保留原值，PUT 现状（nil-skip 保留原值）与该约定自洽；属契约歧义，应单独提 issue 由产品裁决，P1 修复不扩张爆破半径 |
| 顺带支持"显式 `[]` = 全拒"（nil vs 空切片区分，仿 Provider #147） | 产品契约未定义"显式全拒"语义，且 DAO 更新转换需同步改 `len>0` 判定（波及 PATCH，属契约变更）；见 §6.3，另立契约时再议 |
| `block_models` 同步纠默认值 | 契约规定其默认即为 `[]`（entities.md line 52/85），现状合规，动它才是违反契约 |

## 9. 实施记录（2026-09-23，已完成）

按 §6 方案完成实施，改动集中于 storager 层，逐字对称 api_key `models` 先例：

| 文件 | 改动 |
| - | - |
| `storage/rdb/entity/entity.go` | §6.1：`entityDataToParam` else 分支 `"[]"` → `["*"]`（marshal `[]string{"*"}`），注释更新为契约出处；§6.2：`entityParamToData` 反序列化后追加 `len(param.AllowModels) == 0 → ["*"]` 归一化（存量 `"[]"`/NULL 行），注释注明与 api_key 读路径同构。`block_models` 的 `"[]"` 默认值未动 |
| `storage/rdb/entity/entity_test.go` | §6.5-1：`TestCreateEntity_OmittedModelsDefaultToEmpty`（缺陷固化用例）改写为 `TestCreateEntity_OmittedModelsDefaultToStar`——断言省略创建读回 `AllowModels == ["*"]`、`BlockModels` 为空，并内联显式非空往返断言（`["gpt-4"]`/`["gpt-3"]` 原样读回）；新增 `TestFetchEntity_LegacyEmptyAllowModelsNormalizedToStar`——raw SQL 构造存量 `"[]"`、NULL、非空三种遗留行，断言读回归一化行为；`setupTestStorager` 拆出 `setupTestStoragerWithDB`（返回 db 句柄供遗留数据构造），既有用例调用点签名不变 |
| `test/integration/tests/entity/create/create_test.go` | §6.5-2：新增 `E-1-027`（省略 `allow_models` → 创建响应与 GET 回读双通道 `["*"]`、`block_models` 为 `[]`）与 `E-1-028`（显式 `["gpt-4"]` 原样双通道回读），E-1-027 对修复具备回归效力（仅回滚 storager 修复而保留该用例必失败） |
| `test/integration/tests/entity/design.md` | §6.3 总览表登记 E-1-027/028；新增 §6.4.14/§6.4.15 详细设计（含设计思路、步骤、请求参数、预期校验表）；§3 统计表创建用例 24 → 30、合计 52 → 58 |
| `design-docs/sys-design/details/部分更新语义与DAO-nil-skip约定.md` | §6.5-4：§2 表格 Entity 行"Create 仍走 `entityDataToParam` 默认 `"[]"`"更正为"`allow_models` 默认 `["*"]`（issue #202）、`block_models` 默认 `"[]"`"；§3 表格 Entity 行拆分（allow_models 默认 `["*"]` 含修复说明 / block_models 保持 `[]`），表后补充与 api_key 对齐及"显式 `[]` 按省略处置"的创建路径适用说明 |

验证结果（本节替代 §7 计划清单，逐项已执行）：

1. `go build ./...`、`go vet ./...` 通过（本环境无 `make`，按 AGENTS.md 等价使用底层 go 命令完成 `make test` / `test-model-cover-gate` 的同项验证）；
2. 主模块 `go test ./...` 全量通过，零失败；`storage/rdb/entity`、`model/entity`、`model/imods` 等重点包单独复跑通过；
3. model 覆盖率：`model/entity` 78.9%，整体远高于 70% 门禁线（本次改动在 storage 层，未触碰 model 层代码）；
4. 关键自证成立：改写后的 `TestCreateEntity_OmittedModelsDefaultToStar` 断言 `["*"]`，修复前代码读回 `[]` 必失败——即该用例对修复有真实回归效力；#151 引入的 update 族用例（`TestUpdateEntity_OmittedModelsPreserveValues`/`TestUpdateEntity_ProvidedModelsAreWritten`）零回退，PATCH nil-skip 语义未受影响；
5. 集成模块 `test/integration` `go vet ./...` 编译通过；E-1-027/028 需 live 环境执行（本工作区无起服务依赖，未实跑），部署后随 `tests/entity/...` 模块运行；
6. 导出零影响已按 §3.2 代码核实（exporter.go:514 对 `[]`/`["*"]` 同跳过），未改动 `model/imods`，其测试全绿即为佐证。

待办：E2E SC1203-TC009 A-02 requeue 与 SC1203-TC028 回归在 integration-test 仓库执行；存量物理对齐 SQL（§6.4）由运维按需定夺。
