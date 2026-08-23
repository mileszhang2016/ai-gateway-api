# `/model-prices` 与 `/providers` 解耦及新增 provider 列表接口——变更摘要

## 1. 背景

在 `2026-08-22-provider-cluster-separation` 的初始设计中，`/model-prices` 的 `provider` 字段被收紧为“必须引用 `/providers` 中已存在的 provider”。实际维护中发现：

- `/providers` 与 `/model-prices` 的数据往往由不同流程/人员维护，很难严格控制提交顺序。
- 强制引用会导致 model-prices 数据写入受阻，增加配置和迁移成本。
- 删除 provider 时，`/model-prices` 中的同名记录也会阻塞删除，进一步加剧维护难度。

因此，决定将 `/model-prices` 与 `/providers` 之间的关系从**强引用**调整为**弱引用**（按名称关联），并新增一个辅助接口用于发现当前 model-prices 数据中已有的 provider 名称。

## 2. 目标

1. 解除 `/model-prices` 中 `provider` 字段必须引用已存在 provider 的约束。
2. 删除 `/providers/{provider_name}` 时，不再因 `/model-prices` 中的同名记录返回 `409 Conflict`。
3. 新增 `GET /model-prices/actions/get-providers` 接口，返回 model-prices 数据中所有 provider 名称的去重列表。
4. 保持 `/clusters` 对 `/providers` 的强引用约束不变。

## 3. 范围

- **涉及面**：`ai-gateway-api` 控制面及其 OpenAPI/InnerAPI；`model-prices` 相关校验、导入、删除约束；`providers` 删除约束。
- **不涉及面**：`bfe` 数据面；`/clusters` 对 provider 的强引用；`AIConf` 生成逻辑。
- **数据面影响**：无。

## 4. 关键决策

| 决策 | 说明 |
|------|------|
| `/model-prices` 弱引用 `/providers` | `provider` 字段仅作为价格归集标识，不强制校验存在性。 |
| 删除 provider 不再检查 model-prices | 仅保留 `/clusters` 引用的 `409 Conflict` 阻塞。 |
| 新增 `GET /model-prices/actions/get-providers` | 按名称聚合去重，返回 model-prices 中所有 provider 名称。 |
| 配置顺序由强制改为推荐 | 推荐 `/providers → /model-prices → /clusters → 路由规则`，但 `/model-prices` 可独立维护。 |
| `model-list.yaml` 导入放宽 | 未知 provider 可正常导入，不再报错跳过。 |

## 5. 关联文档

- 上游设计来源：`document-ai-gateway/迭代系统设计/v0.5/provider和cluster概念分离/provider和cluster概念分离-设计与实施方案.md`
- 原始变更记录：`ai-gateway-api/design-docs/modifications/2026-08-22-provider-cluster-separation/`
- 接口变更：`api-changes.md`
- 设计变更：`design-changes.md`

## 6. 实施阶段

| 阶段 | 内容 | 预计周期 |
|------|------|----------|
| 1 | 更新 api-define 与 sys-design 文档 | 0.5 周 |
| 2 | 调整 model-prices 校验与 provider 删除约束 | 0.5-1 周 |
| 3 | 实现 `GET /model-prices/actions/get-providers` | 0.5 周 |
| 4 | 更新 model-list.yaml 导入逻辑 | 0.5 周 |
| 5 | 测试（含独立创建/导入 model-prices 的场景） | 0.5-1 周 |
