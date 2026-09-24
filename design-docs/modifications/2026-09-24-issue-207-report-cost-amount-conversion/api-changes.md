# Report 成本字段定点→金额换算（issue #207）API 变更说明

> 对应 `design-docs/api-define/OpenAPI接口定义/report.md` 的修改内容（Step 3 执行）；sys-design 同步见 design-changes.md。

## 1. 变更总览

| 端点 | 字段 | 变更前 | 变更后 |
|------|------|--------|--------|
| GET /report/overview | `cost[].value` | int64，定点整数原值（1 单位 = 1e-8 元/美元） | float64，金额（元/美元）= 定点值 ÷ 1e8 |
| GET /report/timeseries?metric=cost | `series[].value` | float64，定点整数/秒 | float64，金额/秒（元/秒、美元/秒） |
| GET /report/logs | `items[].ai_cost_value` | int64，定点整数原值；无成本为 null | float64，金额；无成本仍为 null |
| 以上三处 | `currency` / `ai_cost_currency` | 不变 | 不变 |

换算在服务端完成（÷1e8），调用方拿到的即为按币种单位的金额，**不再有任何"前端按 currency 格式化"的约定**。

## 2. report.md 具体修改点

### 2.1 §1.2 明细行示例（:48）

```diff
-  "ai_cost_value": 5000,
+  "ai_cost_value": 0.00005,
   "ai_cost_currency": "USD",
```

### 2.2 §1.2 明细行字段表（:78）

```diff
-| `ai_cost_value` / `ai_cost_currency` | int64 / string | 成本固定点整数值 / 币种 |
+| `ai_cost_value` / `ai_cost_currency` | number / string | 成本金额（元/美元，服务端已完成 ÷1e8 换算）/ 币种；无成本为 null |
```

### 2.3 §2.1 总览指标约束（:112）

```diff
-成本按币种分组返回定点整数原值，前端按 `currency` 格式化。
+成本按币种分组返回换算后的金额（定点值 ÷ 1e8，元/美元），换算由服务端完成，调用方直接展示。
```

### 2.4 §2.1 返回示例（:131）

```diff
-  "cost": [{"currency": "USD", "value": 15230000}, {"currency": "RMB", "value": 98000}],
+  "cost": [{"currency": "USD", "value": 0.1523}, {"currency": "RMB", "value": 0.00098}],
```

### 2.5 §2.2 时序 metric 表（:166）

```diff
-| `cost` | 成本增速（定点整数/秒） | 按币种多条序列，`currency` 字段区分 |
+| `cost` | 成本增速（金额/秒，元/秒、美元/秒） | 按币种多条序列，`currency` 字段区分 |
```

## 3. 00-common.md 联动（:206-208 注记段）

RMB 配额定点说明（1e-8 元、9000 万元上限）本身不变（Redis 存储语义未动），在该注记末尾补一句：

> 报表库 `ai_cost_value` 及 BFE 访问日志同为此定点口径（1 单位 = 1e-8 元/美元）；`/open-api/v1/report/*` 接口出口已统一换算为金额（÷1e8），报表链路中唯一直接对外暴露定点值的是 pb 访问日志与库表本身。

## 4. 兼容性

- 破坏性变更：`cost[].value`、`series[].value`（cost 序列）、`ai_cost_value` 的 JSON 类型 int64 → float64，数值语义定点 → 金额；
- 影响消费者：ai-gateway-web Report 模块（#115 联动适配，直接展示金额）；无其他已知消费者；
- 接口路径、方法、参数、权限、错误码均无变化。
