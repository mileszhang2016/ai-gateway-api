# Issue #102：model-prices 8 位小数价格序列化精度设计变更说明

## 1. 当前问题定位

### 1.1 错误表象

在 `ai-gateway-api` 上提交如下 model-prices 记录：

```json
{
  "provider": "mock-provider-8-decimal",
  "model": "deepseek-chat",
  "base_model": "deepseek-chat",
  "mode": "chat",
  "prices": {
    "input_cost_per_token": 0.0000015,
    "output_cost_per_token": 0.0000045
  }
}
```

经 InnerAPI `/inner-api/v1/configs/tls_conf/server_data_conf` 导出后，BFE 侧 `cluster_conf.data` 中对应字段可能被表示为：

```json
"prices": {
    "input_cost_per_token": 1.5e-6,
    "output_cost_per_token": 4.5e-6
}
```

虽然 `encoding/json` 反序列化后 `float64` 值仍为 `0.0000015`，但：

1. 配置文本可读性差，运维排查时容易误判为精度丢失；
2. 当链路中存在中间层再次序列化时，科学计数法可能被进一步截断或解析错误；
3. BFE 若对配置做文本对比/签名，相同语义的不同表示会导致不必要的 diff。

### 1.2 根因分析

Go 标准库 `encoding/json` 对 `float64` 编码时默认使用 6 位有效数字（`fmt` 的 `%g` 默认精度）。当数值绝对值很小（如 `0.0000015`）时，`%g` 会自动切换为科学计数法 `1.5e-6`，以控制输出长度。

`PriceMap map[string]float64` 与 `TierPriceMap map[string]map[string]float64` 原本依赖默认 encoder，因此所有经过 JSON 序列化的路径（OpenAPI 响应、InnerAPI 导出、`ImportModelPrices` 导入后回写、BFE 重新导出）都会受影响。

### 1.3 为什么单元测试没发现

早期测试用例价格多使用 `0.000001`、`0.000002` 等 6 位有效数字以内的值，这些值在默认 encoder 下仍以十进制输出（`1e-6` 被表示为 `0.000001`），未触发科学计数法。直到引入 8 位小数价格（`0.0000015`、`0.00000005`）后问题才暴露。

## 2. 目标

1. 保证 `ai-gateway-api` 中 `PriceMap` / `TierPriceMap` 序列化时使用十进制表示法；
2. 保证 `bfe` 中同名类型序列化时使用十进制表示法；
3. 保持数值语义不变，反序列化结果与修改前完全一致；
4. 不引入新的依赖或复杂的数值类型。

## 3. 方案对比

| 方案 | 思路 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. 自定义 `MarshalJSON`（推荐） | 为 `PriceMap` / `TierPriceMap` 实现 `MarshalJSON`，使用 `strconv.FormatFloat(v, 'f', -1, 64)` | 改动小、无依赖、语义不变、同时适用于 ai-gateway-api 与 bfe | 需要为两个仓库的同构类型分别实现 | **采用** |
| B. 使用 `json.Number` 或字符串存储价格 | 将价格改为字符串类型 | 彻底避免浮点问题 | 需改动 OpenAPI 契约、数据库、BFE 解析，影响面大 | 不采用 |
| C. 使用 `math/big.Float` | 高精度浮点类型 | 精度更高 | 引入复杂类型，与现有 `float64` 定点转换逻辑不兼容 | 不采用 |
| D. 调整 `json.Encoder` 精度 | 全局设置 `SetEscapeHTML` 等无法解决；`fmt` 精度参数不能兼顾所有小数位 | 全局生效 | 会改变所有 float64 输出格式，风险不可控 | 不采用 |

## 4. 推荐方案：自定义 `MarshalJSON`

### 4.1 核心思路

在 `PriceMap` 和 `TierPriceMap` 上实现 `json.Marshaler` 接口，手动拼接 JSON 对象，数值部分使用 `strconv.FormatFloat(v, 'f', -1, 64)`。

`FormatFloat` 参数说明：

- `'f'`：普通十进制表示法，非科学计数法；
- `-1`：按必要精度输出，不截断也不补零；
- `64`：按 `float64` 精度处理。

示例：

| 原始值 | 默认 JSON 输出 | `FormatFloat('f', -1, 64)` |
|--------|----------------|----------------------------|
| `0.0000015` | `1.5e-6` | `0.0000015` |
| `0.0000045` | `4.5e-6` | `0.0000045` |
| `0.00000005` | `5e-8` | `0.00000005` |
| `0.000001` | `0.000001` | `0.000001` |

### 4.2 `ai-gateway-api` 实现

文件：`ai-gateway-api/model/imodel_price/model_price.go`

```go
// PriceMap is a map of price keys to their numeric values.
// It marshals to JSON using decimal notation instead of scientific notation.
type PriceMap map[string]float64

// MarshalJSON serializes PriceMap using decimal notation for all values.
func (p PriceMap) MarshalJSON() ([]byte, error) {
    if p == nil {
        return []byte("null"), nil
    }
    var buf bytes.Buffer
    buf.WriteByte('{')
    first := true
    for k, v := range p {
        if !first {
            buf.WriteByte(',')
        }
        first = false
        keyBytes, err := json.Marshal(k)
        if err != nil {
            return nil, err
        }
        buf.Write(keyBytes)
        buf.WriteByte(':')
        buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

// TierPriceMap is a map of tier names to PriceMap values.
// It marshals to JSON using decimal notation instead of scientific notation.
type TierPriceMap map[string]map[string]float64

// MarshalJSON serializes TierPriceMap using decimal notation for all nested values.
func (t TierPriceMap) MarshalJSON() ([]byte, error) {
    if t == nil {
        return []byte("null"), nil
    }
    var buf bytes.Buffer
    buf.WriteByte('{')
    first := true
    for tier, prices := range t {
        if !first {
            buf.WriteByte(',')
        }
        first = false
        tierBytes, err := json.Marshal(tier)
        if err != nil {
            return nil, err
        }
        buf.Write(tierBytes)
        buf.WriteByte(':')
        if prices == nil {
            buf.WriteString("null")
            continue
        }
        innerFirst := true
        buf.WriteByte('{')
        for k, v := range prices {
            if !innerFirst {
                buf.WriteByte(',')
            }
            innerFirst = false
            keyBytes, err := json.Marshal(k)
            if err != nil {
                return nil, err
            }
            buf.Write(keyBytes)
            buf.WriteByte(':')
            buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
        }
        buf.WriteByte('}')
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}
```

### 4.3 `bfe` 实现

文件：`bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`

BFE 侧同构类型 `PriceMap` / `TierPriceMap` 采用与 `ai-gateway-api` 完全一致的实现，确保 BFE 重新导出配置时不会把已修正的十进制表示再变回科学计数法。

```go
// PriceMap is a map of price keys to their numeric values.
// It marshals to JSON using decimal notation instead of scientific notation.
type PriceMap map[string]float64

// MarshalJSON serializes PriceMap using decimal notation for all values.
func (p PriceMap) MarshalJSON() ([]byte, error) {
    if p == nil {
        return []byte("null"), nil
    }
    var buf bytes.Buffer
    buf.WriteByte('{')
    first := true
    for k, v := range p {
        if !first {
            buf.WriteByte(',')
        }
        first = false
        keyBytes, err := stdjson.Marshal(k)
        if err != nil {
            return nil, err
        }
        buf.Write(keyBytes)
        buf.WriteByte(':')
        buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

// TierPriceMap is a map of tier names to PriceMap values.
// It marshals to JSON using decimal notation instead of scientific notation.
type TierPriceMap map[string]map[string]float64

// MarshalJSON serializes TierPriceMap using decimal notation for all nested values.
func (t TierPriceMap) MarshalJSON() ([]byte, error) {
    if t == nil {
        return []byte("null"), nil
    }
    var buf bytes.Buffer
    buf.WriteByte('{')
    first := true
    for tier, prices := range t {
        if !first {
            buf.WriteByte(',')
        }
        first = false
        tierBytes, err := stdjson.Marshal(tier)
        if err != nil {
            return nil, err
        }
        buf.Write(tierBytes)
        buf.WriteByte(':')
        if prices == nil {
            buf.WriteString("null")
            continue
        }
        innerFirst := true
        buf.WriteByte('{')
        for k, v := range prices {
            if !innerFirst {
                buf.WriteByte(',')
            }
            innerFirst = false
            keyBytes, err := stdjson.Marshal(k)
            if err != nil {
                return nil, err
            }
            buf.Write(keyBytes)
            buf.WriteByte(':')
            buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
        }
        buf.WriteByte('}')
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}
```

> 注：BFE 侧将标准库 json 包导入别名 `stdjson`，以避免与本地类型或变量命名冲突。

### 4.4 定点整数转换不受影响

BFE 在配置加载阶段仍按原有逻辑将 `float64` 价格转换为定点整数（1e-8 元/单位、1e-12 / token 等）：

```go
// 示意
func toFixedPoint(price float64) int64 {
    return int64(price*1e12 + 0.5)
}
```

`MarshalJSON` 只影响配置文本输出，不影响 `float64` 解析后的数值，因此运行时扣减精度不变。

## 5. 数据迁移

本次修复**不需要数据迁移**。数据库中价格仍以 `float64` / `REAL` 存储，数值本身未发生变化。仅在序列化到 JSON 文本时使用十进制表示法。

## 6. 测试计划

### 6.1 单元测试

#### `ai-gateway-api/model/imodel_price/import_test.go`

新增/保留 `TestParseModelListYAMLEightDecimalPlaces`：

1. 解析含 8 位小数价格的 YAML；
2. 断言 `Prices` / `TierPrices` 中 `float64` 值正确；
3. `json.Marshal` 后断言输出包含 `"0.00000005"`、`"0.0000001"`；
4. 断言输出不包含 `"5e-"`、`"1e-"`；
5. `json.Unmarshal` 回 `ModelPrice` 后验证数值往返一致。

#### `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load_test.go`

保留现有 `TestModelTableCheck` 中含 8 位小数的子用例，如 `0.000004525`、`0.000022625`、`0.0000004525`、`0.00000565625`，验证：

1. `ModelTableCheck` 通过；
2. `GetPriceInt` 返回预期定点整数（如 45、565）。

### 6.2 集成测试

运行 `integration-test` 中的端到端用例：

```bash
# SC19：普通 RMB 配额 8 位小数精度
go test ./test-cases/implementation/scenario-SC19-rmb-quota-deduction/... -run TestTC08 -v

# SC25：分时段 tier 价格 8 位小数精度
go test ./test-cases/implementation/scenario-SC25-rmb-tiered-pricing/... -run TestTC06 -v
```

预期：

- 请求成功返回 200；
- RMB 余额按精确固定点整数扣减；
- 导出到 BFE 的 `cluster_conf.data` 中价格为十进制表示法。

### 6.3 回归验证命令

```bash
# ai-gateway-api
cd ai-gateway-api
go test ./model/imodel_price/...

# bfe
cd bfe
go test ./bfe_config/bfe_cluster_conf/cluster_conf/...

# integration-test
cd integration-test
go test ./test-cases/implementation/scenario-SC19-rmb-quota-deduction/... -run TestTC08
go test ./test-cases/implementation/scenario-SC25-rmb-tiered-pricing/... -run TestTC06
```

## 7. 风险与缓解

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| JSON 文本长度增加 | 十进制表示比科学计数法更长 | 价格字段数量有限，对整体配置大小影响可忽略 |
| 自定义 Marshal 引入新 bug | 手动拼接 JSON 可能遗漏转义 | key 仍使用 `json.Marshal` 生成，确保转义正确；测试覆盖 nil、空 map、嵌套 map 场景 |
| 反序列化行为变化 | 自定义 Marshal 不影响 Unmarshal | 保留标准 `float64` 解析，已有单元测试验证 |

## 8. 实施状态

- [x] `ai-gateway-api/model/imodel_price/model_price.go` 增加 `PriceMap.MarshalJSON` 与 `TierPriceMap.MarshalJSON`；
- [x] `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go` 增加同名自定义序列化；
- [x] `ai-gateway-api` 单元测试 `TestParseModelListYAMLEightDecimalPlaces` 通过；
- [x] `bfe` 单元测试 `TestModelTableCheck` 含 8 位小数子用例通过；
- [x] `integration-test` SC19 TC-08 与 SC25 TC-06 端到端通过。

## 9. 参考文档

- [Issue #102](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/102)
- `ai-gateway-api/model/imodel_price/model_price.go`
- `ai-gateway-api/model/imodel_price/import_test.go`
- `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go`
- `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load_test.go`
- `integration-test/test-cases/测试设计文档/scenario-SC19-RMB配额扣减/TC-08-8位小数价格端到端精度.md`
- `integration-test/test-cases/测试设计文档/scenario-SC25-RMB分时段计费/TC-06-8位小数分时段价格端到端精度.md`
