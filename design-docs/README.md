# design-docs 使用说明

本目录用于集中管理 `ai-gateway-api` 的设计文档与代码变更流程。所有代码变更都应遵循以下 6 步方法，确保设计文档与代码实现保持一致。

---

## 目录结构

```
design-docs/
├── README.md              # 本文件：变更方法论
├── api-define/            # API 接口定义（OpenAPI / InnerAPI）
├── sys-design/            # 系统设计文档
│   ├── summary.md         # 文档索引
│   └── details/           # 关键模块细节设计
└── modifications/         # 历次代码变更说明（按日期+目的组织）
```

---

## 代码变更六步法

### Step 1：创建变更说明目录

在 `design-docs/modifications/` 下创建一个目录，目录名称格式为：

```
YYYYMMDD-<变更目的简述>
```

示例：

```
modifications/
├── 20260725-add-rate-limit-policy/
├── 20260726-refactor-quota-sync/
└── 20260728-support-entity-hierarchy/
```

目录名称应使用英文或拼音，避免特殊字符，尽量概括本次变更核心目标。

---

### Step 2：编写本次变更说明

在 Step 1 创建的目录内，放置本次代码变更的修改说明文件。建议至少包含以下内容：

| 文件 | 说明 |
|------|------|
| `change-summary.md` | 变更摘要：背景、目标、影响范围、关键决策 |
| `api-changes.md`（如需要） | API 接口变更说明：新增/修改/删除的接口、字段、返回值 |
| `design-changes.md`（如需要） | 设计变更说明：数据模型、流程、算法、约束变化 |

变更说明应**基于大模型生成**，但需由开发者审核确认。生成时可提供：

- 变更背景与目标
- 相关代码路径
- 预期影响的功能模块
- 需要特别关注的边界情况

---

### Step 3：更新并 review api-define

基于 Step 2 的变更说明，对 `api-define/` 下的接口定义进行修改：

1. 新增或修改 OpenAPI 接口定义；
2. 新增或修改 InnerAPI 接口定义；
3. 更新请求/响应字段、错误码、示例；
4. 检查接口版本兼容性。

修改完成后，需进行 review，确认：

- [ ] 接口路径、方法、参数与变更说明一致；
- [ ] 字段命名符合现有规范；
- [ ] 错误码覆盖新增失败场景；
- [ ] 不影响已有接口的兼容性（若涉及）。

---

### Step 4：更新并 review sys-design

基于 Step 2 的变更说明和 Step 3 的 api-define 修改，对 `sys-design/` 下的系统设计文档进行更新：

1. 更新总体设计文档（如涉及架构或流程变化）；
2. 更新接口层/模型层/存储层/数据库设计文档（如涉及分层变化）；
3. 更新或新增 `sys-design/details/` 下的细节设计文档（如涉及关键模块）；
4. 同步更新 `sys-design/summary.md` 索引。

修改完成后，需进行 review，确认：

- [ ] 设计与 api-define 一致；
- [ ] 数据模型、表结构、JSON 字段描述准确；
- [ ] 新增的细节文档已加入 summary.md；
- [ ] 文档中的代码路径和文件名与实际代码一致。

---

### Step 5：基于设计文档修改代码

基于 `design-docs/` 中的最终设计，对 `ai-gateway-api/` 代码进行修改：

1. 按接口层 → 模型层 → 存储层的顺序实现；
2. 新增或修改 endpoints、model、storage/rdb 代码；
3. 补充或更新单元测试、集成测试；
4. 本地运行测试，确保通过。

代码修改应始终与设计文档保持一致。若实现过程中发现设计需要调整，应**回到 Step 3/Step 4 更新设计文档**，而不是直接偏离设计。

---

### Step 6：总结并沉淀到 details

代码变更完成后，请大模型基于以下内容进行总结：

- 本次变更的修改说明（Step 2）
- api-define 的变更（Step 3）
- sys-design 的变更（Step 4）
- 实际代码变更（Step 5）

判断是否有**可复用、可沉淀的设计知识**，值得放入 `sys-design/details/` 中供后续使用。适合沉淀的内容包括：

- 新的核心机制（如配额同步、限流策略、路由规则）
- 复杂的继承/合并逻辑（如 Entity 层级、模型黑白名单）
- 关键的导出/版本控制流程
- 重要的边界情况与设计权衡

若决定沉淀，应：

1. 在 `sys-design/details/` 下新建细节文档；
2. 在 `sys-design/summary.md` 中补充索引；
3. 确保细节文档基于实际代码，而非仅参考旧版设计文档。

---

## 快速检查清单

每次变更完成后，建议对照以下清单确认流程完整：

- [ ] 已在 `modifications/` 下创建本次变更目录；
- [ ] 变更说明文件已填写并审核；
- [ ] `api-define/` 已更新并 review；
- [ ] `sys-design/` 已更新并 review；
- [ ] `sys-design/summary.md` 索引已同步；
- [ ] 代码已按设计文档实现并通过测试；
- [ ] 已评估是否需要沉淀新的 `details/` 文档。

---

## 示例

假设需要为 AI 网关增加 API-Key 级别的限流策略：

```
design-docs/modifications/
└── 20260728-apikey-rate-limit/
    ├── change-summary.md      # 背景、目标、影响范围
    ├── api-changes.md         # 新增 /rate-limit-policies 接口
    └── design-changes.md      # 限流策略数据模型、导出格式、层级合并逻辑
```

变更流程：

1. 创建 `20260728-apikey-rate-limit` 目录；
2. 编写 `change-summary.md`、`api-changes.md`、`design-changes.md`；
3. 更新 `api-define/` 中限流相关接口定义；
4. 更新 `sys-design/` 中模型层、存储层、数据库设计文档，新增 `details/限流策略与导出.md`；
5. 实现 `model/quota/rate_limit_policy*.go`、`storage/rdb/quota/rate_limit_policy.go`、endpoints 等代码；
6. 请大模型总结，确认限流策略机制可沉淀为长期细节文档。
