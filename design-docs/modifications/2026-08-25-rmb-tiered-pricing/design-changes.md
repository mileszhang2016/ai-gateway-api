# RMB 配额分时段定价——设计变更说明

## 1. 概念定义

| 概念 | 定义 |
|------|------|
| **Tier** | 按时间维度划分的价格层级，如 `peak`（高峰）。请求发生时刻命中某个 tier 时，使用对应价格；否则 fallback 到默认 `prices`。 |
| **TimeRange** | 一个时段定义，包含 `weekdays`（星期几）、`start` / `end`（HH:MM）。 |
| **Provider 时段模板** | 定义在 `/providers` 上的 `time_zone` 和 `tiers`，同一 provider 下所有模型共享。 |
| **Model tier 价格** | 定义在 `/model-prices` 上的 `tier_prices`，描述某个模型在某个 tier 下的价格。 |

## 2. 配置归属与下发链路

### 2.1 provider/cluster 分离后的配置归属

按"同一 provider 内所有模型共享时段规则"的原则，将时段模板与价格数据拆分到两个资源：

- **`/providers`**：新增 `time_zone`、`tiers` 字段，描述"该 provider 在什么时区、哪些时段属于哪个 tier"。
- **`/model-prices`**：新增 `tier_prices` 字段，描述"某个 provider 的某个模型在某个 tier 下按什么价格计费"；默认 `prices` 继续作为非 tier 命中或 tier 未覆盖键时的 fallback 价格。

### 2.2 控制面 → 数据面转换

```
/provider deepseek
    ├── time_zone: Asia/Shanghai
    └── tiers: [peak]
    ├── model-prices deepseek-v4-pro chat
    │       ├── prices: 默认/空闲价
    │       └── tier_prices.peak: 高峰价
    └── model-prices deepseek-v4-flash chat
            └── ...

/cluster my-cluster
    ├── llm_config.provider: deepseek
    └── AIConf.ModelTable（导出时由 control-plane 把 provider 的 time_zone/tiers 与 model-prices 的 prices/tier_prices 拼接）
```

多个 cluster 引用同一个 provider 时，会各自得到一份相同的 `ModelTable` 数据；provider 的时段规则变更后，所有引用它的 cluster 在下一次配置导出时自动生效。

## 3. 数据模型改动

### 3.1 ai-gateway-api `/providers` 数据模型

新增字段：

| 字段 | 类型 | 说明 | 合法性条件 |
|------|------|------|------------|
| `time_zone` | string | 计算时段所使用的时区 | 默认 `Asia/Shanghai`；须为 IANA 时区名 |
| `tiers` | []object | 时段 tier 定义列表 | 可选；元素包含 `name` + `time_ranges`；**初期 `name` 只支持 `peak`** |

### 3.2 ai-gateway-api `/model-prices` 数据模型

新增字段：

| 字段 | 类型 | 说明 | 合法性条件 |
|------|------|------|------------|
| `tier_prices` | object | tier name -> 价格表 | 可选；**初期键名只支持 `peak`**；内部键名须为 `prices` 枚举；与 provider 的 `tiers` 不做强制引用校验 |

### 3.3 数据库表结构

#### `providers` 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `time_zone` | varchar(255) | 默认 `"Asia/Shanghai"`；IANA 时区名 |
| `tiers` | text (JSON) | 时段 tier 定义列表 |

示例：

```sql
ALTER TABLE `providers`
    ADD COLUMN `time_zone` varchar(255) NOT NULL DEFAULT 'Asia/Shanghai',
    ADD COLUMN `tiers` text;
```

#### `model_prices` 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `tier_prices` | text (JSON) | tier name -> 价格表 |

示例：

```sql
ALTER TABLE `model_prices`
    ADD COLUMN `tier_prices` text;
```

### 3.4 BFE `AIConf.ModelTable` 配置结构

BFE 侧保持 v0.4 方案的结构，仅把配置来源明确为"由 ai-gateway-api 根据 provider 的 `time_zone` / `tiers` 与 model-prices 的 `prices` / `tier_prices` 拼接后填充"。

```go
type TimeRange struct {
    Weekdays []int  // 0=周日, 1=周一 ... 6=周六；为空表示每天
    Start    string // "HH:MM"
    End      string // "HH:MM"，必须 > Start；跨午夜请拆成两段
}

type PriceTier struct {
    Name       string      // 初期只支持 "peak"
    TimeRanges []TimeRange // 命中任意一个即属于该 Tier
}

type ModelPrice struct {
    Provider            string
    Model               string
    BaseModel           string
    Mode                string
    Capabilities        []string
    SupportedParameters []string
    Limits              map[string]interface{}
    Prices              map[string]float64            // 默认价格（标准 prices 键名）
    TierPrices          map[string]map[string]float64 // tier name -> 价格表（标准 prices 键名）
    Metadata            map[string]interface{}

    // 运行时字段：配置加载阶段预计算定点整数
    pricesInt     map[string]int64            // 默认价格定点整数
    tierPricesInt map[string]map[string]int64 // tier 价格定点整数
}

type ModelTable struct {
    Currency string      // 仍是 "RMB"
    TimeZone string      // 默认 "Asia/Shanghai"
    Tiers    []PriceTier // 时段定义
    Models   []ModelPrice

    priceIndex map[string]map[string]*ModelPrice // model -> mode -> *ModelPrice
    tierIndex  map[string]*PriceTier             // tier name -> *PriceTier
    tz         *time.Location                    // 加载时解析的时区
}
```

## 4. 控制面改造点

### 4.1 配置导出（`model/icluster_conf/exporter.go`）

生成 `AIConf` 时：

1. 根据 `cluster.llm_config.provider` 查询 `/providers`，获取 `time_zone`、`tiers`。
2. 根据同一 `provider` 查询 `/model-prices`，把每条记录转换成 `ModelPrice`。
3. 将 `time_zone` / `tiers` 填入 `AIConf.ModelTable`，将 `prices` / `tier_prices` 填入对应 `ModelPrice`。
4. 若 provider 未配置 `time_zone` / `tiers`，则 `AIConf.ModelTable` 中 `TimeZone` / `Tiers` 为空，BFE 按固定价格处理。

### 4.2 多 provider 时段差异

provider/cluster 分离后，时段规则天然跟随 provider：

- `/providers` 中每个 provider 可以有自己的 `time_zone` 和 `tiers`。
- 导出到 BFE 时，每个 cluster 的 `AIConf.ModelTable` 携带其引用的 provider 的时区与时段规则。
- 不同 provider 的高峰/空闲定义可以完全不同。

## 5. BFE 数据面改动

BFE 侧与 v0.4 方案基本一致，因为 `AIConf.ModelTable` 结构未变，仅配置来源发生变化。

### 5.1 加载阶段增强

在 `ModelTableCheck` 里：

1. 解析 `TimeZone`，默认 `Asia/Shanghai`。
2. 校验 `Tiers`：名字非空、**初期 `name` 只支持 `peak`**、时间格式正确、`Weekdays` 合法、同一 Tier 内时间不重叠。
3. 将默认 `Prices` 和每个 tier 的价格表分别转成定点整数；`TierPrices` 中的 tier name 不强制要求在 `Tiers` 中已定义（运行时只有 `Tiers` 中命中的 tier 才会被使用），但**初期 `TierPrices` 的键名只支持 `peak`**。

### 5.2 运行时时段匹配

按请求发生时刻、模型表绑定的时区，依次匹配 `Tiers` 列表：

- 命中 tier 后，成本计算优先从该 tier 的价格表中取价。
- 未命中 tier 或 tier 未覆盖某个价格键时，fallback 到默认 `Prices`。
- 时间区间采用左闭右开语义（`start <= cur < end`）。

### 5.3 运行时成本计算

`calcCostUnits` 根据当前时间匹配 tier，再取对应价格；同时支持缓存命中 / 未命中分层：

- `TokenUsage` 增加 `CachedTokens` 字段，用于命中缓存的输入 token 数。
- `UpdateCtxByUsage` 增加缓存 token 解析（不同 provider 字段不同，如 `usage.prompt_tokens_details.cached_tokens` 或 `usage.prompt_cache_hit_tokens`）。
- 成本计算区分 `input_cost_per_token` 与 `cache_read_input_token_cost`。

## 6. 向后兼容

- `/providers` 不填 `time_zone` / `tiers` 时，BFE 的 `ModelTable.TimeZone` / `ModelTable.Tiers` 为空，行为与现在完全一致（固定价格）。
- `/model-prices` 不填 `tier_prices` 时，按默认 `Prices` 固定价格计费。
- 命中 tier 但该 tier 未配置某个价格键时，自动 fallback 到默认 `Prices` 中的对应键。
- `TokenUsage.UsedCost`、Lua 扣减逻辑、Redis 定点数存储都不需要改。

## 7. 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| tier name 首期约束 | 仅支持 `peak`，业务上若需要表达 `off_peak` / 周末等需后续放开 | 设计时已通过 `Weekdays` 预留扩展能力，后续只需放开命名约束。 |
| provider 与 model-prices 的 tier 松耦合 | `tier_prices` 可能引用 provider 未定义的 tier，导致价格永不命中 | 可记录告警，但不做强制引用校验以保持灵活性。 |
| 时区解析依赖系统时区数据库 | 部署环境需包含 IANA 时区数据 | 使用 Go 标准库 `time.LoadLocation`，确保容器镜像包含时区数据。 |
| 高峰/空闲切换时刻的边界 | 时间区间左闭右开，切换瞬间价格突变 | 产品侧明确规则，BFE 按配置精确匹配。 |
| 多 cluster 共享 provider | 修改 provider 时段规则会影响所有引用 cluster | 更新时给出引用列表确认。 |

## 8. 实现与测试

### 8.1 已实现的关键代码

| 模块 | 文件 | 说明 |
|------|------|------|
| 数据库 DDL | `db_ddl.sql`、`db_ddl_sqlite.sql` | `providers` 新增 `time_zone`、`tiers`；`model_prices` 新增 `tier_prices` |
| 存储层 | `storage/rdb/internal/dao/table_providers.go`、`table_model_prices.go` | DAO struct 新增字段 |
| 存储层 | `storage/rdb/provider/provider.go`、`storage/rdb/model_price/model_price.go` | JSON 序列化/反序列化 |
| 模型层 | `model/iprovider/provider.go`、`model/iprovider/validate.go` | `Provider`/`ProviderParam` 新增字段、`UpdatePricingTiers`、时区/时段校验 |
| 模型层 | `model/imodel_price/model_price.go`、`model/imodel_price/validate.go` | `ModelPrice` 新增 `TierPrices`、`tier_prices` 校验 |
| 导出层 | `model/icluster_conf/cluster.go`、`model/icluster_conf/cluster_test.go` | `NewBfeClusterConf` 新增 `providerPricingTable`、拼接 `ModelTable` |
| 导出层 | `model/iroute_conf/exporter.go` | 构建 `providerPricingTable` 并传入 cluster conf |
| 接口层 | `endpoints/openapi_v1/provider/pricing_tiers.go`、`endpoints/openapi_v1/provider/endpoints.go` | `PUT /providers/{provider_name}/pricing-tiers` |
| 集成测试 | `test/integration/tests/provider/pricing_tiers/pricing_tiers_test.go` | pricing-tiers 接口 JSON/YAML/multipart 正例与异常校验 |
| 集成测试 | `test/integration/tests/model_price/tier_prices/tier_prices_test.go` | model-prices `tier_prices` 正例与异常校验、导入覆盖 |
| 集成测试 | `test/integration/tests/innerapi/tls_conf/tiered_pricing_test.go` | InnerAPI 导出 `ModelTable` 含 `Currency`/`TimeZone`/`Tiers`/`TierPrices` |
| 集成测试 | `test/integration/tests/schema/openapi/provider.go` | ProviderSchema 增加 `time_zone`、`tiers` 定义 |
| 集成测试 | `test/integration/tests/schema/openapi/model_price.go` | ModelPriceSchema 增加 `tier_prices` 定义 |
| 集成测试 | `test/integration/tests/schema/openapi/openapi_schema_test.go` | schema 测试数据携带 `time_zone`/`tiers`/`tier_prices` |
| 测试工具 | `test/integration/testutil/client.go`、`test/integration/testutil/api_helpers.go` | `PutMultipartFile`、pricing-tiers/YAML 辅助函数 |

### 8.2 测试期间发现并修复的问题

1. **`model/iprovider/provider.go` `UpdatePricingTiers`**：原实现只构造含 `time_zone`/`tiers` 的 `ProviderParam`，导致 `UpdateProvider` 把 provider 的 `models`、`keys`、`instance_pool` 等字段覆盖为空。已修复为从 existing provider 回填全部字段，仅更新时区与 tiers。
2. **`endpoints/openapi_v1/model_price/update.go` `mergeModelPrice`**：原合并逻辑遗漏 `TierPrices`，导致 `PUT /model-prices/{id}` 更新 `tier_prices` 不生效。已补充合并分支。
3. **`model/iprovider/provider.go` 时段重叠校验**：设计文档要求“同一 tier 内部 `time_ranges` 不得重叠”，但实现仅校验重复。已补充 `timeRangesOverlap` 检测。
4. **`tests/schema/innerapi/innerapi_schema_test.go` `redis_key` 前缀断言**：限流策略的 `redis_key` 实际包含产品线/集群前缀（如 `default_bfe_rlp-2_RL_TPM_rlp-2_tpm-1m`），原断言使用严格前缀导致失败。已改为校验关键片段 `RL_<TYPE>_<policy>_`。

- **单元测试**：`model/iprovider`、`model/imodel_price`、`model/icluster_conf`、`model/iroute_conf`、`endpoints/openapi_v1/provider`、`endpoints/openapi_v1/model_price` 均新增或更新用例。
- **全量测试**：`go test ./...` 全量通过。

### 8.3 测试建议（已落地的验证点）

1. **Provider 校验**：`time_zone` 非法时拒绝；`tiers` 中 `name` 非 `peak` 时拒绝；时间重叠、`weekdays` 越界、`end <= start` 时拒绝。
2. **Model-prices 校验**：`tier_prices.<tier>.<key>` 使用非枚举键名时拒绝；`tier_prices` 中出现非 `peak` 的 tier name 时拒绝。
3. **Pricing-tiers 接口**：JSON / YAML / multipart 均可正确解析并写入；`GET /providers/{provider_name}` 可返回完整信息。
4. **配置导出**：引用同一 provider 的多个 cluster 导出的 `ModelTable` 一致；provider 时段规则变更后同步更新；未配置时段规则时固定价格行为正确。
5. **BFE 数据面**：BFE 侧结构已就绪，加载、时段匹配与成本计算由 BFE 现有逻辑承载。

## 9. 设计文档更新

为与代码和测试保持一致，已同步更新以下设计文档：

| 文档 | 更新内容 |
|------|----------|
| `test/integration/tests/provider/design.md` | 新增 PV-8 `PUT /providers/{provider_name}/pricing-tiers` 接口说明、目录结构、测试用例（JSON/YAML/multipart、异常校验、GET 返回验证） |
| `test/integration/tests/model_price/design.md` | 在模块概述、接口参数、数据示例中补充 `tier_prices`；新增第 17 节“分时段价格字段测试（MTP-1）”，覆盖创建、更新、导入、非法 tier name/价格键/负数价格等 6 个场景 |
| `test/integration/tests/innerapi/design.md` | 在模块概述中补充分时段定价下发链路；更新测试用例统计；在 `tls_conf` 章节新增 IN-TIER-1-001/IN-TIER-1-002，验证 `ModelTable.TimeZone/Tiers/TierPrices` 导出及向后兼容 |
| `test/integration/tests/schema/openapi/provider.go` | ProviderSchema 增加 `time_zone`、`tiers` 字段定义 |
| `test/integration/tests/schema/openapi/model_price.go` | ModelPriceSchema 增加 `tier_prices` 字段定义 |
| `test/integration/tests/schema/openapi/openapi_schema_test.go` | schema 测试数据携带 `time_zone`/`tiers`/`tier_prices`，验证字段生效 |
| `test/integration/tests/schema/innerapi/schema.go` | ServerDataConfSchema 增加 `ClusterConf` 嵌套校验；新增 `ModelTableSchema` 与 `ModelPriceSchema`，覆盖 `Currency`/`TimeZone`/`Tiers`/`Models`/`Prices`/`TierPrices` 等字段 |
| `test/integration/tests/schema/innerapi/innerapi_schema_test.go` | 新增 `server_data_conf_tiered_pricing` schema 测试，验证 InnerAPI 导出 `AIConf.ModelTable` 含分时段定价字段 |
