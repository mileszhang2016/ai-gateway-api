# Issue #115：Provider `models` 改为必填修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/115](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/115)

> models 在 provider & clusters 接口定义不一致，一个是非必填，一个是必填。
> 结果容易使得在创建 cluster 时才发现 provider 没有配置 model 而无法创建 cluster，又需要返回修改 provider 配置。
> 建议：provider.models 改为必填。

**不一致的双方**：

| 接口 | 字段 | 合同条款 |
|------|------|---------|
| `/providers`（providers.md 第 1 部分） | `models` | `非必填；元素非空且不可重复；可通过模型发现接口自动填充`（providers.md:55） |
| `/clusters`（clusters.md 第 1 部分） | `llm_config.models` | `必填；至少1个元素；每个模型名非空；元素不能重复；每个元素必须存在于 provider 的 models 中`（clusters.md:163） |

由于 `llm_config.models` 必须是 provider `models` 的子集且至少 1 个元素，provider `models` 为空的 provider 永远无法被 cluster 引用——缺陷被推迟到创建 cluster 时才暴露，违背"尽早失败"。

## 2. 现状核验（文档 + 代码）

| 证据 | 位置 | 说明 |
|------|------|------|
| 本次修改点（主） | `design-docs/api-define/OpenAPI接口定义/providers.md:55` | `models` 合法性条件写作"非必填；元素非空且不可重复；**可通过模型发现接口自动填充**" |
| 同源过时描述 ① | providers.md:191（2.1 执行逻辑第 6 步） | "若请求中携带 models 且非空，直接保存；否则可在创建后调用 /providers/tools/discover-models 接口探测模型列表，再回填到 provider"——基于非必填语义 |
| 同源过时描述 ② | providers.md:606（第 3 部分校验规则 6） | 仅"`models` 元素非空且不可重复"，缺"必填、至少 1 个元素" |
| 自动填充确已不存在 | providers.md:442（2.6 说明）+ modifications/2026-08-24-provider-discover-models-tool | discover-models 自 2026-08-24 起为**无状态工具接口**，不写回 provider；providers.md:55 的"自动填充"与 providers.md:442 自相矛盾，确已过时 |
| 代码与旧合同一致 | `model/iprovider/provider.go:497-508` | `ValidateProviderParam` 仅当 `len(param.Models) > 0` 时校验元素，空/缺省直接放行 |
| 默认值填充 | `model/iprovider/provider.go:815-817` `FillDefaults` | `param.Models == nil` → 填充 `[]string{}`（POST 创建路径 create.go:41 先 FillDefaults 再校验） |
| 校验挂点 | `model/iprovider/provider.go:162`（CreateProvider）、`:202`（UpdateProvider） | 创建/更新共用 `ValidateProviderParam`；PATCH 部分更新语义为"省略保留原值"（providers.md:354-356），storager nil-skip（storage/rdb/provider/provider.go:253-264 `toDAOParamForUpdate` → `marshalJSONPtr`） |
| 集成测试固化旧合同 | `test/integration/tests/provider/create/create_test.go:69-83` | `PV-1-001 最小参数创建 Provider` 不带 models 创建并断言 200 + models 为空 |

## 3. 裁定：采纳 issue 建议

provider `models` 改为**必填、至少 1 个元素**，删除"可通过模型发现接口自动填充"。理由：

1. 与 clusters.md:163 对齐，消除接口间矛盾，provider 配置缺失在创建时即暴露；
2. "自动填充"与 2026-08-24 之后 discover-models 的无状态语义矛盾（providers.md:442 自身已写明"不读写任何 Provider 资源"），同一文档自相矛盾；
3. 无冲突：provider PATCH 的部分更新语义（providers.md:354-356）不受影响——"必填"约束的是**资源本身与创建（POST）请求**，PATCH 省略 `models` 仍表示保留原值。

**"必填"的适用范围（关键语义裁定）**：

| 场景 | 语义 |
|------|------|
| POST /providers | `models` 必填，至少 1 个元素；缺失或显式 `[]` 均拒绝（ErrNum=422） |
| PATCH /providers/{name}，省略 `models` | 合法，保留原值（部分更新语义不变） |
| PATCH /providers/{name}，显式传 `[]` | 拒绝（422）——显式提供即全量替换，清空 models 违反合同，且可能使已引用 cluster 失效 |
| PATCH 显式传非空数组 | 元素非空、不重复校验后全量替换（删除被引用 model 返回 409 的逻辑不变） |

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（合同文档 + 校验实现 + 测试）；BFE / conf-agent 零改动 |
| 主要文件 | `design-docs/api-define/OpenAPI接口定义/providers.md`（3 处）；`model/iprovider/provider.go`（`ValidateProviderParam` models 分支 + `CreateProvider` + `UpdatePricingTiers`）；`test/integration/tests/provider/`（用例改造与新增） |
| 接口契约 | POST /providers 新增对缺失/空 models 的 422 拒绝；PATCH 新增对显式空数组的 422 拒绝；其余行为不变 |
| 数据迁移 | 无；存量 models 为空的 provider 记录不清理，仍可读可改（见 §7 兼容性） |
| 明确不做 | discover-models 工具接口本身不变（保持无状态）；cluster 侧（clusters.md:163）本就正确，不动；dashboard 前端表单联动另案跟进 |

## 5. 最终方案

### 5.1 合同文档修改（providers.md，共 3 处）

**修改 1（主修改，第 1 部分字段说明，providers.md:55）**：

```diff
-| `models` | []string | 该 provider 支持的模型列表 | - | 非必填；元素非空且不可重复；可通过模型发现接口自动填充 |
+| `models` | []string | 该 provider 支持的模型列表 | - | 必填；至少 1 个元素；元素非空且不可重复 |
```

**修改 2（2.1 创建 Provider 执行逻辑第 6 步，providers.md:191）**：

```diff
-6. 若请求中携带 `models` 且非空，直接保存；否则可在创建后调用 `/providers/tools/discover-models` 接口探测模型列表，再回填到 provider。
+6. 校验 `models` 必填：至少 1 个元素，元素非空且不可重复，通过后直接保存。如需借助模型发现工具（`/providers/tools/discover-models`，无状态接口）确定模型列表，调用方需先调用该工具，再在创建请求中携带其结果。
```

**修改 3（第 3 部分校验规则 6，providers.md:606）**：

```diff
-6. `models` 元素非空且不可重复。
+6. `models` 必填，至少 1 个元素；元素非空且不可重复。（PATCH 部分更新时省略 `models` 表示保留原值，不视为违反必填；显式提供时必须满足本条款。）
```

**sys-design 同步**：经核验，`sys-design/接口层设计文档.md:245` 与 `details/provider与cluster概念分离.md` 仅描述 models 子集关系与引用校验，未出现"非必填/自动填充"表述，与本修复无矛盾，**均无需改动**；本次为单字段合同修正，规模不达新增 `details/` 长期文档门槛。

### 5.2 代码修改（使合同生效）

合同从"非必填"变为"必填"是行为变更，需同步实现，否则合同与实现再次背离。

**修改点 1：`ValidateProviderParam` models 分支**（`model/iprovider/provider.go:497-508`）：

```go
// 旧：仅当 len>0 时校验元素，空/缺省放行
if len(param.Models) > 0 { ... }

// 新：区分"未提供（nil，PATCH 省略）"与"显式提供"
if param.Models != nil {
    if len(param.Models) == 0 {
        return xerror.WrapParamErrorWithMsg("models must have at least one element")
    }
    seenModel := map[string]bool{}
    for i, m := range param.Models {
        if strings.TrimSpace(m) == "" {
            return xerror.WrapParamErrorWithMsg("models[%d] cannot be empty", i)
        }
        if seenModel[m] {
            return xerror.WrapParamErrorWithMsg("duplicate model: %s", m)
        }
        seenModel[m] = true
    }
}
```

- `nil`（PATCH 省略）放行，保持部分更新语义与 storager nil-skip 约定；
- 显式 `[]` 拒绝（ErrNum=422，参数非法）。

**修改点 2：创建路径必填**（`CreateProvider`，`provider.go:161-169`）：

POST 端点先经 `FillDefaults`（create.go:41 → provider.go:815-817）将 `nil` 改写为 `[]`，校验阶段无法区分"未传"与"显式空"——两者在创建语境下均违反必填，统一拒绝即可。在 `CreateProvider` 的 `ValidateProviderParam` 之后追加：

```go
if len(param.Models) == 0 {
    err := xerror.WrapParamErrorWithMsg("models is required and must have at least one element")
    m.recordProviderOperation(ctx, string(ioperlog.ActionCreate), name, nil, providerParamToMap(param), err)
    return 0, err
}
```

> 备选实现：为该检查并入 `ValidateProviderParam` 并增加 create/update 模式参数。倾向最小改动——校验函数保持单一签名，创建必填在 `CreateProvider` 内显式表达。

**修改点 3：`UpdatePricingTiers` 防误伤**（`provider.go:306-317`）：

`updateParam.Models` 当前回填 `existing.Models`。存量 provider 的 models 可能是空数组（非 nil，来自 `unmarshalStringSlice`），在修改点 1 的新校验下会导致"维护 pricing tiers 被拒"。改为 `Models: nil`（storager nil-skip 保留存量值；pricing tiers 本就不触碰 models）。

**不做**：`FillDefaults` 的 models nil→`[]` 填充保留——POST 路径由修改点 2 兜住，PATCH 路径不经 FillDefaults；删除该填充会影响操作日志等其他消费方，收益低风险高。

### 5.3 修复后行为对照

| 输入 | 修复前 | 修复后 |
|------|--------|--------|
| POST 创建不带 models | 200，models=`[]` | 422 `models is required...` |
| POST 创建 `"models": []` | 200，models=`[]` | 422 `models must have at least one element` |
| POST 创建 models=`["a",""]` / `["a","a"]` | 422 | 422（不变） |
| POST 创建 models=`["a","b"]` | 200 | 200（不变） |
| PATCH 省略 models | 200 保留原值 | 200 保留原值（不变） |
| PATCH `"models": []`（未被引用） | 200 清空 models | 422 `models must have at least one element` |
| PATCH models=`["a","b"]`（非空） | 200 全量替换 | 200 全量替换（不变；删除被引用 model 仍 409） |
| PUT pricing-tiers（存量空 models provider） | 200 | 200（修改点 3 保证） |

## 6. 回归防护

1. **manager 单测**（`model/iprovider/` 既有测试文件）：
   - `CreateProvider`：缺 models、显式 `[]` 拒绝（错误文案可归因）；≥1 元素通过；
   - `UpdateProvider`：nil models 通过；显式 `[]` 拒绝；非空数组元素校验回归；
   - `UpdatePricingTiers`：存量 models 为空的 provider 可正常更新 tiers。
2. **集成测试**（`test/integration/tests/provider/`，沿用 PV 编号风格续接）：
   - 改造 `PV-1-001 最小参数创建 Provider`（create_test.go:69-83）：最小 body 需携带 models 并断言回显，或将其调整为"缺 models 创建拒绝"用例并另新增带 models 的最小创建用例；
   - 新增：POST 显式 `"models": []` 拒绝（ErrNum=422 + 文案归因）；PATCH 显式 `[]` 拒绝；PATCH 省略 models 保留原值（GET 回读断言）；
   - `design.md` 用例表同步更新。
3. **前端联动（另案）**：dashboard 创建 provider 表单应将 models 标为必填，并支持调用 discover-models 辅助填充（人工确认后提交）——不阻塞本修复。

## 7. 风险与兼容性

| 项目 | 说明 |
|------|------|
| 兼容性 | 行为变更：POST 不带 models 从"200 静默入库"变为 422。属 issue 明确要求的修复目标；调用方需在创建时显式提供模型列表（可先调 discover-models 获取） |
| 存量数据 | 存量 models 为空的 provider 不做数据迁移：仍可查询、PATCH（省略 models）、维护 pricing tiers；但新合同下 cluster 引用前必须先为其补 models（cluster 侧子集校验一直如此，无新增阻塞） |
| 校验边界 | 仅在 OpenAPI 入口校验一次（单一入口原则）；InnerAPI 导出消费已入库数据，不重复校验 |
| BFE 关系 | 零改动：BFE 消费的是 cluster 导出配置，provider models 必填只影响控制面录入 |
| 主要风险 | 低。校验函数单分支修改 + 创建路径一处显式检查 + pricing tiers 一处 nil 化；存储层/导出层零改动 |

---

*文档生成日期：2026-09-16*
