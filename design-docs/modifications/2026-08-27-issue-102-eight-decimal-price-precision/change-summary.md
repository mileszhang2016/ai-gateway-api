# Issue #102：model-prices 8 位小数价格在配置导出时精度丢失

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/102](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/102)

> 在 `ai-gateway-api` 上提交 8 位小数的 model-prices（如 `input_cost_per_token = 0.0000015`），经 InnerAPI 导出到 `bfe` 后，价格被序列化为科学计数法或截断为 6 位有效数字，导致 BFE 按错误价格扣减 RMB 配额。

## 2. 目标

1. 保证 `ai-gateway-api` 在 OpenAPI 响应、InnerAPI 导出、`ImportModelPrices` YAML/JSON 序列化时，8 位小数价格均使用十进制表示法，不丢失精度。
2. 保证 `bfe` 在重新导出/回写 `AIConf.ModelTable` 时，同样使用十进制表示法。
3. 保持对现有 6 位及以内小数价格的完全向后兼容。
4. 不改动 OpenAPI / InnerAPI 的接口契约、请求/响应字段名或数据类型。

## 3. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`、`bfe` |
| 主要文件 | `ai-gateway-api/model/imodel_price/model_price.go`、`bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go` |
| 数据库 | 无需迁移；浮点数值本身存储不变 |
| 接口契约 | 不变；仅 JSON 文本表示方式调整 |
| 数据面影响 | BFE 解析到的 `float64` 数值不变，配置文本可读性提升 |

## 4. 最终方案概览

为 `PriceMap` 与 `TierPriceMap` 添加自定义 `MarshalJSON`，使用 `strconv.FormatFloat(v, 'f', -1, 64)` 替代默认 JSON encoder。

- Go 标准 `encoding/json` 对 `float64` 默认使用 6 位有效数字，小数值（如 `0.0000015`）会被序列化为 `1.5e-6`，在部分场景下造成可读性与精度混淆。
- `strconv.FormatFloat(v, 'f', -1, 64)` 按所需的最小小数位输出十进制字符串：`0.0000015` 保持为 `"0.0000015"`，`0.00000005` 保持为 `"0.00000005"`。
- 该改动同时作用于 `ai-gateway-api` 控制面与 `bfe` 数据面，确保整条链路的配置文本一致。

## 5. 预期收益与风险

| 项目 | 说明 |
|------|------|
| 收益 | 8 位小数价格在 OpenAPI、InnerAPI、BFE 配置文件中均以十进制表示，避免科学计数法导致的精度误解；RMB 配额扣减按精确固定点整数执行 |
| 主要风险 | 自定义 `MarshalJSON` 后 JSON 文本长度略有增加；无功能回退 |
| 兼容性 | 数值反序列化仍由标准库完成，`float64` 解析结果不变；现有测试无需修改即可通过 |

## 6. 上线前检查清单

1. `ai-gateway-api` 单元测试 `TestParseModelListYAMLEightDecimalPlaces` 通过；
2. `bfe` 单元测试 `TestModelTableCheck` 及含 8 位小数价格的子用例通过；
3. 运行 `integration-test` 中 SC19 `TestTC08_EightDecimalPlaces` 与 SC25 `TestTC06_RMBQuotaDeduction_Peak_8DecimalPlaces`，确认端到端 RMB 扣减精度正确。

## 7. 参考文档

- [Issue #102](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/102)
- `ai-gateway-api/model/imodel_price/model_price.go`
- `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`
- `integration-test/test-cases/测试设计文档/scenario-SC19-RMB配额扣减/TC-08-8位小数价格端到端精度.md`
- `integration-test/test-cases/测试设计文档/scenario-SC25-RMB分时段计费/TC-06-8位小数分时段价格端到端精度.md`
