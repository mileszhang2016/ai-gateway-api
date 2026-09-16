# Issue #170：GET /model-prices 组合键查询单记录缺失修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/170](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/170)

> API 文档 §3.6「按组合键查询单条记录」定义 GET /model-prices?provider=&model=&mode=（三参均必填）应返回 Data = 单个 ModelPrice 对象。但实际 GET /model-prices 仅注册 ListEndpoint，ListAction 无条件返回列表载荷 `{"list":[...],"pagination":{...}}`，无组合键单记录分支。三参被 queryFilter 当作普通列表过滤，Data 永远是列表包装对象。
>
> 客户端把 Data 当单记录解码时，list/pagination 键与 ModelPrice 字段不匹配被 Go json.Unmarshal 静默忽略，所有业务字段保持零值（id=0 provider="" ...）。

危害：组合键查询"看起来正确"（记录确实在 `Data.list[0]` 中），但按 §3.6 契约解码的客户端拿到全零值，属于静默契约违例。PUT / DELETE 已实现 `UpdateByQueryEndpoint` / `DeleteByQueryEndpoint` 组合键路径，唯独 GET 缺失等价路径。E2E SC1301-TC029 保持待真验终态等待本修复。

## 2. 根因（代码核验）

| 证据 | 位置 | 说明 |
|------|------|------|
| §3.6 单记录契约 | `design-docs/api-define/OpenAPI接口定义/model-prices.md:539-560` | 三参均必填(Y)；"返回数据 Data内容：字段同 1. 数据模型" = 单对象 |
| §3.4 列表契约 | `model-prices.md:487-513` | Data = `{list:[], pagination:{}}`；过滤参数仅 provider/mode，**不含 model** |
| GET handler 恒返回列表 | `endpoints/openapi_v1/model_price/list.go:46-65` | `ListAction` 无条件 `return &listResponse{List, Pagination}` |
| 三参仅当过滤 | `endpoints/openapi_v1/model_price/helper.go:39-53` | `queryFilter` 把 provider/model/mode 当可选列表 filter（model 过滤 §3.4 从未提供） |
| 路由缺 GET ByQuery | `endpoints/openapi_v1/model_price/endpoints.go:22-32` | 有 `UpdateByQueryEndpoint`/`DeleteByQueryEndpoint`，无 GET 等价物 |
| PUT 组合键参照 | `endpoints/openapi_v1/model_price/update.go:37-42,74-99` | `UpdateByQueryAction` 缺参拒绝 + 查存合并回读 |
| DELETE 组合键参照 | `endpoints/openapi_v1/model_price/delete.go:36-41,65-83` | `DeleteByQueryAction` 同上 |

补充技术约束：gorilla/mux **不允许同 path+method 注册两个路由**。`UpdateByQueryEndpoint` 能独立存在是因为 method 不同（PUT/DELETE 与 GET 不冲突）；GET 的单记录路径与 `ListEndpoint` 同为 `GET /model-prices`，无法照搬"独立 ByQueryEndpoint"的模式——这是 GET 侧"缺失"的直接结构原因，也决定了修复只能落在 `ListAction` 内部做参数分发（issue 路径 A 允许此实现方式）。

## 3. 裁定：路径 A（落实 §3.6 单记录契约）

issue 要求产品裁定 A/B，本方案裁定 **路径 A**，理由：

1. 合同 §3.6 已发布且语义明确（三参必填、返回单对象），问题属"代码落后于设计"，应修代码而非降合同；
2. PUT/DELETE 组合键路径已落地，GET 补齐后 CRUD 四元组对称，API 家族风格一致；
3. 路径 B（改合同为列表语义）需要修订合同 + 改写 SC1301-TC029 断言，且把"单记录意图"永远绑在易静默解出零值的列表形状上，对 Go 客户端是持续陷阱；
4. 前端已具备兼容前提（见 §8），路径 A 落地零前端改动。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（合同文档同步）；`ai-gateway-web` **无需改动**（已兼容双形状） |
| 主要文件 | `endpoints/openapi_v1/model_price/list.go`（`ListAction` 增加分发分支）；`helper.go`（视实现微调）；新增 `list_test.go` |
| 接口契约 | 三参齐全时 GET `/model-prices` 响应从列表包装修正为单 ModelPrice 对象（§3.6 原承诺）；其余参数组合行为不变 |
| 收紧项 | `model` 单独（或与 provider/mode 部分组合）出现从"静默当列表过滤"变为"参数错误拒绝"；§3.4 合同从未提供 model 过滤，属收紧未文档化行为 |
| 合同对齐项 | §3.4 `page_size` 合同由「默认20，最大100」对齐为「默认50，最大1000」，与实现 `helper.go:55-73` 一致（改合同不改代码，产品已确认随本次修复一并落地） |
| 数据迁移 | 无 |

## 5. 最终方案

### 5.1 分发规则（精确判定）

`ListAction` 入口按 query 参数做三分支：

| 参数组合 | 走向 | 行为 |
|----------|------|------|
| provider、model、mode **均非空** | §3.6 单记录分支 | mode 须 ∈ `imodel_price.ValidModes`，否则参数错误；`FetchModelPrice` 查组合键，nil 则 `WrapRecordNotExist("ModelPrice")`；命中返回单个 `*ModelPrice`（形状同 `OneAction`，`one.go:36-50`） |
| **带 model** 但三参不齐（如 model 单独、provider+model 缺 mode） | 参数错误拒绝 | `xerror.WrapParamErrorWithMsg("provider, model and mode are required")`——与 `update.go:77`、`delete.go:68` 同款文案，非 2xx，可归因 |
| **不带 model** | §3.4 列表分支（现状不动） | provider/mode 可选过滤 + page/page_size 分页，返回 `{list, pagination}`；`provider+mode` 无 model 的合法 §3.4 过滤保持列表语义 |

选择"带 model 即拒绝缺参"而非"缺参回落列表"：缺参回落会让 `?model=xxx` 静默走列表，正是本 issue 要消除的"看起来成功"陷阱；§3.4 合同的过滤参数本就不含 model。

### 5.2 实现要点

`list.go` 的 `ListAction` 改为：

```go
func ListAction(req *http.Request) (interface{}, error) {
    filter := queryFilter(req)

    // §3.6：三参齐全 → 单记录
    if filter.Provider != nil && filter.Model != nil && filter.Mode != nil {
        if !imodel_price.ValidModes[*filter.Mode] {
            return nil, xerror.WrapParamErrorWithMsg("invalid mode: %s", *filter.Mode)
        }
        one, err := container.ModelPriceManager.FetchModelPrice(req.Context(), filter)
        if err != nil {
            return nil, err
        }
        if one == nil {
            return nil, xerror.WrapRecordNotExist("ModelPrice")
        }
        return one, nil
    }

    // 带 model 但三参不齐 → 拒绝（不留静默列表fallback）
    if filter.Model != nil {
        return nil, xerror.WrapParamErrorWithMsg(
            "provider, model and mode are required for single-record query")
    }

    // §3.4：列表（现状逻辑原样保留）
    page, pageSize := pageFilter(req)
    ...
}
```

- mode 枚举复用 `model/imodel_price/validate.go:28-42` 的 `ValidModes`，不新建枚举表；
- 错误类型复用 `xerror.WrapParamErrorWithMsg` / `xerror.WrapRecordNotExist`，与组合键 PUT/DELETE、OneAction 一致；
- `helper.go` 的 `queryFilter` 保持原样（仍解析三参），分发逻辑收敛在 `ListAction` 一处；
- 不改路由表（`endpoints.go`）、不改 model/storage 层。

### 5.3 修复后行为对照（issue 复现路径）

`GET /open-api/v1/model-prices?provider=deepseek&model=deepseek-v4-pro&mode=chat`：

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 组合键存在 | 200，Data 为 `{list:[...],pagination:{...}}`，单记录客户端解码出全零值 | 200，Data 为单个 ModelPrice 对象，字段直接匹配 |
| 组合键不存在 | 200，Data 为 `{list:[],pagination:{total:0}}` | RecordNotExist（404 语义），客户端可按"不存在"处理 |
| 缺任一参（带 model） | 200，静默当列表过滤 | 参数错误（非 2xx），ErrMsg 可归因 |
| mode 非法 | 200，当过滤条件查列表 | 参数错误（非 2xx） |

## 6. 合同文档变更（model-prices.md）

§3.6 的合同语义本身正确，无需改写，仅补充分发说明避免再歧义：

- §3.4 输入参数表下补一行说明：「当 `model` 参数出现时，本端点改按 §3.6 单记录语义处理；`provider`+`mode`（不带 `model`）为列表过滤组合」；
- §3.6 补一行说明：「三参任一缺失或 `mode` 非法时返回参数错误，不回落列表语义；组合键不存在时返回 Record Not Exist」。

**合同对齐（本次修复一并落地）**：§3.4 `page_size` 参数行由「默认20，最大100 / 取值范围 1-100」改为「默认50，最大1000 / 取值范围 1-1000」，与实现 `helper.go:55-73`（默认 50、上限 1000）一致。采用"改合同不改代码"方向：默认值/上限收紧会影响现存调用方与 web 前端分页习惯，放宽合同表述则零行为变更。

**sys-design 同步**：`sys-design/接口层设计文档.md` 路由表 4.2.6 节已载「GET /model-prices：分页列表查询 / 按组合键查询」，与修复后行为一致，无需改动；`sys-design/总体设计文档.md` 能力概述行补充「按 `(provider, model, mode)` 组合键查询/修改/删除」；`summary.md` 索引描述仍准确，无需改动。本次为单端点分发修正，规模不达新增 `details/` 长期文档门槛（参照 issue #140 同族修改仅落在 modifications/ 与 api-define）。

## 7. 回归防护

1. **handler 单测**：新增 `endpoints/openapi_v1/model_price/list_test.go`。模式：手写 fake `ModelPriceStorager`，经 `imodel_price.NewManager(txn, storager)` 构造 Manager 注入 `container.ModelPriceManager`（容器变量直换模式参照 `endpoints/openapi_v1/endpoints_test.go:88`）。覆盖矩阵：
   - 三参齐全 + 命中 → 返回 `*ModelPrice`（非 listResponse，断言无 `list`/`pagination` 键）；
   - 三参齐全 + 未命中 → `WrapRecordNotExist`；
   - 缺参矩阵：model 单独、provider+model 缺 mode、model+mode 缺 provider → 参数错误；
   - mode 非法（如 `mode=foo`）→ 参数错误；
   - 不带 model：无参 / 仅 provider / 仅 mode / provider+mode → 列表形状与过滤行为不变（§3.4 回归护栏）；
   - 行为分支不依赖注册顺序，直接调 `ListAction` 验证。
2. **本地集成测试对齐**（`test/integration/tests/model_price/`，随 issue #172 修复批次执行时发现并修正）：`one/one_test.go` MP-5-001~003 原断言旧列表语义（三参齐全返回列表包装 / 缺参回落列表 / 未命中空列表 200），已改写为新契约——MP-5-001 断言单对象形状（无 `list`/`pagination` 键、字段直配）、MP-5-002 断言 422、MP-5-003 断言 404；`design.md` MP-5 章节（接口说明、场景总览、详细设计）同步改写，MP-3 列表用例（仅 provider/mode/page 过滤）不受影响。
3. **E2E**：SC1301-TC029 在修复部署后走 requeue-real-verification 真验转 PASSED（三记录回读 id/provider/model/mode/价格匹配 + 缺参矩阵拒绝）。

## 8. 风险与兼容性

| 项目 | 说明 |
|------|------|
| web 前端（已核验） | `ModelPriceUpsert.vue:701-720` 组合唯一性校验已同时兼容两种 Data 形状（`data.id \|\| data.list.length>0`）且把 404 视为"不存在可提交"——路径 A 落地后两种结局与今天一致（存在→单对象带 id→判重；不存在→404→放行），**无需前端改动**；`ModelPrices/index.vue:222-232` 列表搜索仅传 provider（及 mode），§3.4 分支不受影响 |
| 其他客户端 | 三参齐全时 Data 形状变化属"修正为合同承诺形状"，是缺陷修复而非行为变更；此前按列表解码的调用方本就解出零值，无实际可用行为被破坏 |
| 收紧项 | `model` 单独出现从静默列表过滤变为参数错误。合同 §3.4 从未承诺 model 过滤；若存在未知调用方依赖该行为，错误信息可引导其改用三参单记录查询 |
| page_size 合同对齐 | 纯合同表述变更（20/100 → 50/1000），实现 `helper.go:55-73` 行为不变，对所有现存调用方零影响 |
| 主要风险 | 低。代码变更面收敛在 `ListAction` 一个函数；model/storage/路由层零改动 |

## 9. 备选方案记录（路径 B，不采纳）

删除 §3.6、组合键查询并入 §3.4 列表语义，改写 SC1301-TC029 断言为「total=1 且 list[0] 字段匹配」。不采纳原因：降合同迁就缺陷实现；与已落地的 PUT/DELETE 组合键路径风格断裂；列表形状对单记录意图的 Go 客户端是持续的静默零值陷阱（issue 描述的核心危害原样保留）。

---

*文档生成日期：2026-09-16*
