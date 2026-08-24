# `/model-prices` 与 `/providers` 解耦及新增 provider 列表接口——API 变更说明

## 1. 新增接口

### 1.1 `GET /model-prices/actions/get-providers`

**功能**

查询 `/model-prices` 数据中包含的所有 `provider` 名称的去重列表。

**响应示例**

```json
{
  "providers": [
    "deepseek",
    "openai",
    "qwen"
  ]
}
```

**处理逻辑**

1. 从 `model-prices` 表中按 `provider` 字段聚合去重。
2. 返回按字典序排列的 provider 名称列表。

**使用场景**

- 前端创建 model-prices 记录或配置 cluster 时，快速查看已有价格数据的 provider。
- 与 `GET /providers` 结果对比，识别“已有价格记录但尚未创建 provider”的 provider，便于补录 `/providers`。

---

## 2. `/model-prices` 校验变更

### 2.1 `provider` 字段约束放宽

| 项目 | 变更前 | 变更后 |
|------|--------|--------|
| 含义 | Provider / Cluster 标识 | 保持为 Provider / Cluster 标识，仅用于价格归集和查找 |
| 存在性校验 | 必须引用 `/providers` 中已存在的 provider | 不再强制校验存在性 |
| 删除 provider 影响 | 若存在同名 model-prices 记录，删除 provider 返回 `409 Conflict` | 不再作为删除 provider 的阻塞条件 |

### 2.2 新增/导入/更新行为

- **新增/更新 `/model-prices`**：不再校验 `provider` 是否存在于 `/providers`。
- **`model-list.yaml` 导入**：未知 provider 可正常写入，不再作为 error 返回或跳过。

---

## 3. `/providers` 删除约束变更

### 3.1 删除 `/providers/{provider_name}`

| 被引用资源 | 变更前 | 变更后 |
|------------|--------|--------|
| `/clusters` | `409 Conflict` | `409 Conflict`（保持不变） |
| `/model-prices` | `409 Conflict` | 不再阻塞删除 |

---

## 4. 配置顺序

变更前为强制顺序：

```
/providers → /model-prices → /clusters → 路由规则
```

变更后为推荐顺序：

```
/providers → /model-prices → /clusters → 路由规则
```

其中 `/model-prices` 不再强制依赖 `/providers`，允许独立维护。
