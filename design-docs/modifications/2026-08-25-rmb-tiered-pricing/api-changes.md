# RMB 配额分时段定价——API 接口变更说明

## 1. `/providers` 资源增强

### 1.1 新增字段

在现有 provider 数据模型基础上新增：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `time_zone` | string | 否 | 计算时段所使用的时区，默认 `Asia/Shanghai`，须为 IANA 时区名 |
| `tiers` | []object | 否 | 时段 tier 定义列表，元素包含 `name` + `time_ranges`；**初期 `name` 只支持 `peak`** |

`tiers` 元素结构示例：

```json
{
  "name": "peak",
  "time_ranges": [
    { "weekdays": [1, 2, 3, 4, 5], "start": "09:00", "end": "12:00" },
    { "weekdays": [1, 2, 3, 4, 5], "start": "14:00", "end": "18:00" }
  ]
}
```

字段约束：

- `weekdays`：0=周日，1=周一，...，6=周六；为空表示每天。
- `start` / `end`：格式 `HH:MM`，`end` 必须大于 `start`；跨午夜请拆成两段。
- 同一 tier 内的多个 `time_ranges` 为"或"关系；`tiers` 列表按顺序匹配，命中第一个即停止。

### 1.2 新增接口

| 方法 | 端点 | 含义 |
|------|------|------|
| PUT | `/providers/{provider_name}/pricing-tiers` | 设置/更新指定 provider 的高峰/闲时模板 |

#### 输入参数（URI）

| 参数名 | 类型 | 参数含义 | 必填 |
|--------|------|----------|------|
| `provider_name` | string | Provider 名称 | 是 |

#### 输入参数（Body）

支持两种提交方式：

1. **JSON 格式**（`Content-Type: application/json`）：

```json
{
    "time_zone": "Asia/Shanghai",
    "tiers": [
        {
            "name": "peak",
            "time_ranges": [
                { "weekdays": [1, 2, 3, 4, 5], "start": "09:00", "end": "12:00" },
                { "weekdays": [1, 2, 3, 4, 5], "start": "14:00", "end": "18:00" }
            ]
        }
    ]
}
```

2. **YAML 文件格式**（`Content-Type: text/yaml` 或 `multipart/form-data` 上传文件）：

```yaml
time_zone: "Asia/Shanghai"
tiers:
  - name: "peak"
    time_ranges:
      - weekdays: [1, 2, 3, 4, 5]
        start: "09:00"
        end: "12:00"
      - weekdays: [1, 2, 3, 4, 5]
        start: "14:00"
        end: "18:00"
```

#### 返回数据

返回更新后的完整 provider 记录（包含基础信息 + `time_zone` / `tiers`）。

### 1.3 读取接口响应变化

现有 `GET /providers/{provider_name}` 与 `GET /providers` 接口保持不变，响应中同时返回基础信息与 `time_zone` / `tiers`：

```json
{
    "name": "deepseek",
    "description": "DeepSeek 官方 API",
    "model_endpoint": { "schema": "https", "uri": "/v1/models" },
    "models": ["deepseek-v4-pro", "deepseek-v4-flash"],
    "keys": [...],
    "instance_pool": [...],
    "model_protocols": ["openai"],
    "time_zone": "Asia/Shanghai",
    "tiers": [
        { "name": "peak", "time_ranges": [...] }
    ],
    "create_time": 1716883200,
    "update_time": 1716883200
}
```

### 1.4 配置流程

1. **创建 provider**：调用 `POST /providers`，只提交基础信息，不提交高峰/闲时信息。
2. **设置高峰/闲时模板**：调用 `PUT /providers/{provider_name}/pricing-tiers`，提交 JSON 或 YAML 格式的高峰/闲时信息。
3. **读取 provider**：调用 `GET /providers/{provider_name}`，返回完整信息。

## 2. `/model-prices` 资源增强

### 2.1 新增字段

在现有 model-prices 数据模型基础上新增：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `tier_prices` | object | 否 | tier name -> 价格表；**初期键名只支持 `peak`**；内部键名须为 `prices` 枚举；与 provider 的 `tiers` 不做强制引用校验 |

`tier_prices` 示例：

```json
{
  "tier_prices": {
    "peak": {
      "input_cost_per_token": 9.0e-06,
      "output_cost_per_token": 2.7e-05,
      "cache_read_input_token_cost": 3.0e-07
    }
  }
}
```

### 2.2 字段说明

- `prices` 作为默认价格，用于未命中任何 tier 的场景（如 DeepSeek 的空闲时段）。
- `tier_prices` 只列出与默认价格不同的 tier（如 `peak`）；如果某个 tier 未配置某个价格键，自动 fallback 到 `prices` 中的对应键。
- `tier_prices` 中的 tier name 建议与 provider 的 `tiers` 保持一致。
- 若 provider 的 `tiers` 中不存在该 tier name，则对应价格永远不会被命中。

### 2.3 对 `model-list.yaml` 源格式的影响

`model-list.yaml` 仍只维护模型价格数据，不引入 `providers` 段落；provider 的高峰/闲时模板通过独立的 `/providers/{provider_name}/pricing-tiers` 接口维护。

```yaml
version: v1.0
default_currency: "RMB"

models:
  - provider: "deepseek"
    model: "deepseek-v4-pro"
    base_model: "deepseek-v4-pro"
    mode: "chat"
    prices:
      input_cost_per_token: 0.0000045
      output_cost_per_token: 0.0000135
      cache_read_input_token_cost: 0.00000015
    tier_prices:
      peak:
        input_cost_per_token: 0.000009
        output_cost_per_token: 0.000027
        cache_read_input_token_cost: 0.0000003
```

## 3. 校验规则

### 3.1 `/providers` 校验

1. `time_zone` 须为合法 IANA 时区名；为空时默认 `Asia/Shanghai`。
2. `tiers` 中每个 tier 必须包含非空 `name` 和至少一个 `time_range`。
3. **`tiers` 中每个 tier 的 `name` 初期只支持 `peak`**。
4. `time_ranges` 中 `weekdays` 元素必须在 0-6 之间；`start` / `end` 格式为 `HH:MM`，且 `end` > `start`。
5. 同一 tier 内部 `time_ranges` 不得重叠；不同 tier 之间允许重叠（按列表顺序匹配）。

### 3.2 `/model-prices` 校验

1. `tier_prices.<tier>.<key>` 必须是 `prices` 枚举中的合法键名，且值为非负数。
2. **`tier_prices` 的键名初期只支持 `peak`**。
3. `tier_prices` 中的 tier name 与 provider 的 `tiers` 不做强制引用校验；若引用了一个不存在的 tier name，可记录告警，但不阻塞写入。

## 4. Inner API 变更

### 4.1 `/configs/tls_conf/server_data_conf` 导出结构变化

受影响的端点：

| 端点 | 方法 | 变更点 |
|------|------|--------|
| `/configs/tls_conf/server_data_conf` | GET | 返回的 `ClusterConf.Config.<cluster_name>.AIConf.ModelTable` 结构增强 |

该 Inner API 的 URL、Method、鉴权均无变化，但导出逻辑与返回结构均发生变化：

- **导出逻辑**：`model/iroute_conf/exporter.go` 在导出 route rule 时，会先全量查询 `/providers`，按 `provider_name` 构建 `providerPricingTable`（含 `time_zone`、`tiers`），再传入 `model/icluster_conf.NewBfeClusterConf`。
- **结构增强**：返回的 `ClusterConf.Config.<cluster_name>.AIConf.ModelTable` 结构增强：

| 变更项 | 说明 |
|--------|------|
| 新增 `ModelTable.Currency` | 固定为 `"RMB"` |
| 新增 `ModelTable.TimeZone` | 来源为 `/providers` 的 `time_zone`；未配置时默认 `"Asia/Shanghai"` |
| 新增 `ModelTable.Tiers` | 来源为 `/providers` 的 `tiers`；为空时 BFE 按固定价格处理 |
| 新增 `ModelTable.Models[].TierPrices` | 来源为 `/model-prices` 的 `tier_prices`；key 为 tier name，value 为价格对象 |
| `ModelTable.Models[].Prices` 语义调整 | 从"唯一价格"变为"默认 / fallback 价格"；未命中 tier 或 tier 未覆盖某个键时使用 |

导出后的 `ModelTable` 示例：

```json
{
    "ModelTable": {
        "Currency": "RMB",
        "TimeZone": "Asia/Shanghai",
        "Tiers": [
            {
                "Name": "peak",
                "TimeRanges": [
                    { "Weekdays": [1, 2, 3, 4, 5], "Start": "09:00", "End": "12:00" },
                    { "Weekdays": [1, 2, 3, 4, 5], "Start": "14:00", "End": "18:00" }
                ]
            }
        ],
        "Models": [
            {
                "Provider": "deepseek",
                "Model": "deepseek-v4-pro",
                "BaseModel": "deepseek-v4-pro",
                "Mode": "chat",
                "Prices": {
                    "input_cost_per_token": 4.5e-06,
                    "output_cost_per_token": 1.35e-05,
                    "cache_read_input_token_cost": 1.5e-07
                },
                "TierPrices": {
                    "peak": {
                        "input_cost_per_token": 9.0e-06,
                        "output_cost_per_token": 2.7e-05,
                        "cache_read_input_token_cost": 3.0e-07
                    }
                }
            }
        ]
    }
}
```

### 4.2 配置导出逻辑

`ai-gateway-api` 在生成 BFE 配置时，关键实现位于 `model/iroute_conf/exporter.go` 与 `model/icluster_conf/cluster.go`：

1. `exportRouteRule` 全量查询 `/providers`，构建 `map[string]ProviderPricingInfo`（`providerPricingTable`），包含每个 provider 的 `time_zone`、`tiers`。
2. `NewBfeClusterConf` 接收 `providerPricingTable`，根据 `cluster.llm_config.provider` 取出对应 `ProviderPricingInfo`。
3. 按同一 `provider` 查询 `/model-prices`，把每条记录转换成 `ModelPrice`。
4. 将 `time_zone` / `tiers` 填入 `AIConf.ModelTable`，将 `prices` / `tier_prices` 填入对应 `ModelPrice`；`Currency` 固定为 `"RMB"`。
5. 若 provider 未配置 `time_zone` / `tiers`，则 `AIConf.ModelTable` 中 `TimeZone` / `Tiers` 为空，BFE 按固定价格处理。

### 4.3 对 BFE 的兼容要求

- BFE 侧 `AIConf.ModelTable` 结构复用 v0.4 方案，字段名保持 PascalCase（`Currency`、`TimeZone`、`Tiers`、`TierPrices`、`Weekdays`、`Start`、`End`）。
- `Currency` 固定为 `"RMB"`；BFE 加载阶段需校验。
- `TimeZone` 为空时，BFE 默认使用 `"Asia/Shanghai"`。
- `Tiers` 为空时，BFE 直接按 `Prices` 固定价格计费。
- `TierPrices` 中引用了未在 `Tiers` 中定义的 tier name 时，BFE 不报错，对应价格仅永不命中。

## 5. 实现状态

- 接口层：`PUT /providers/{provider_name}/pricing-tiers` 已在 `endpoints/openapi_v1/provider/pricing_tiers.go` 实现并注册。
- 模型层：`model/iprovider` 完成 `time_zone`/`tiers` 校验与 `UpdatePricingTiers`；`model/imodel_price` 完成 `tier_prices` 校验。
- 导出层：`model/iroute_conf/exporter.go` + `model/icluster_conf/cluster.go` 完成 `ModelTable` 拼接。
- 测试：`go test ./...` 全量通过。
