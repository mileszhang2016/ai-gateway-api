# Model Price 测试用例设计文档

## 1. 模块概述

Model Price 模块负责模型定价数据的管理，支持：

- 通过 `model-list.yaml` 整表导入（`replace` / `merge` 两种模式）
- 单条记录的增删改查
- 按 `id` 或按 `(provider, model, mode)` 组合键查询/更新/删除

本次新增该模块的集成测试，覆盖 OpenAPI `/model-prices` 下的全部接口。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| MP-1 | 整表导入 | POST | `/open-api/v1/model-prices/import` | 通过 YAML 文件批量导入，支持 replace/merge |
| MP-2 | 新增单条记录 | POST | `/open-api/v1/model-prices` | 创建单条模型定价 |
| MP-3 | 分页列表查询 | GET | `/open-api/v1/model-prices` | 支持 provider、mode 过滤 |
| MP-4 | 按 ID 查询单条 | GET | `/open-api/v1/model-prices/{id}` | - |
| MP-5 | 按组合键查询（列表过滤） | GET | `/open-api/v1/model-prices` | 需传 provider + model + mode，返回列表 |
| MP-6 | 按 ID 修改单条 | PUT | `/open-api/v1/model-prices/{id}` | 支持部分字段更新 |
| MP-7 | 按组合键修改单条 | PUT | `/open-api/v1/model-prices` | 需传 provider + model + mode |
| MP-8 | 按 ID 删除单条 | DELETE | `/open-api/v1/model-prices/{id}` | - |
| MP-9 | 按组合键删除单条 | DELETE | `/open-api/v1/model-prices` | 需传 provider + model + mode |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 整表导入 | 6 |
| 新增单条记录 | 6 |
| 分页列表查询 | 4 |
| 按 ID 查询单条 | 2 |
| 按组合键查询单条 | 3 |
| 按 ID 修改单条 | 3 |
| 按组合键修改单条 | 2 |
| 按 ID 删除单条 | 2 |
| 按组合键删除单条 | 2 |
| **合计** | **30** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
model_price/
├── design.md
├── import/
│   └── import_test.go
├── create/
│   └── create_test.go
├── list/
│   └── list_test.go
├── one/
│   └── one_test.go
├── update/
│   └── update_test.go
└── delete/
    └── delete_test.go
```

## 6. 测试数据约定

### 6.1 合法模型记录示例

```json
{
    "provider": "deepseek",
    "model": "deepseek-v3",
    "base_model": "deepseek-v3",
    "mode": "chat",
    "capabilities": ["chat", "reasoning", "tools"],
    "supported_parameters": ["temperature", "max_tokens"],
    "limits": {
        "context_window": 128000,
        "max_input_tokens": 128000,
        "max_output_tokens": 8192
    },
    "prices": {
        "input_cost_per_token": 0.000002,
        "output_cost_per_token": 0.000008
    },
    "metadata": {
        "source": "https://platform.deepseek.com/pricing",
        "notes": "DeepSeek V3"
    }
}
```

### 6.2 唯一性约束

`(provider, model, mode)` 三元组必须唯一。

### 6.3 测试用例编号规则

```
MP-{接口编号}-{场景编号}
```

例如：`MP-2-001` 表示 MP-2（新增单条记录）的第 1 个场景。

---

## 7. 整表导入（MP-1）

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 整表导入 |
| 方法 | POST |
| 路径 | `/open-api/v1/model-prices/import` |
| 说明 | 通过 `model-list.yaml` 批量导入，支持 replace / merge 模式 |

### 7.2 接口参数说明

#### 请求参数（Form-Data）

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| file | file | Y | YAML 文件 | 须为合法 YAML，版本 v1.0，default_currency=RMB |
| mode | string | N | 导入模式 | `replace`（默认）/ `merge` |

#### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| imported_count | int | 成功导入/更新的记录数 |
| skipped_count | int | 跳过的记录数 |
| errors | []string | 错误列表 |

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-1-001 | replace 模式全量替换 | 正常参数 | 导入后数据库中只有新数据 |
| MP-1-002 | merge 模式增量合并 | 正常参数 | 已存在记录更新，新记录插入 |
| MP-1-003 | 默认 replace 模式 | 正常参数 | 不传 mode 时按 replace 处理 |
| MP-1-004 | 非法 mode | 异常参数 | mode=invalid，返回 422 |
| MP-1-005 | 非法 YAML | 异常参数 | 文件内容非合法 YAML，返回 422 |
| MP-1-006 | 重复三元组 | 异常参数 | YAML 内 (provider,model,mode) 重复，返回 422 |

### 7.4 测试场景详细设计

#### MP-1-001：replace 模式全量替换（正常参数）

##### 设计思路

验证 replace 模式会清空已有数据并写入 YAML 中的全部记录。

##### 前提数据准备

已通过其他接口创建 1 条 model price 记录。

##### 执行步骤

1. 构造包含 2 条记录的 `model-list.yaml`。
2. 以 `mode=replace` 调用导入接口。
3. 验证返回 `imported_count=2`。
4. 调用列表接口，验证总条数为 2，且只有 YAML 中的 2 条记录。

##### 请求参数

```yaml
version: v1.0
default_currency: RMB
models:
  - provider: openai
    model: gpt-4o
    base_model: gpt-4o
    mode: chat
    prices:
      input_cost_per_token: 0.0001
  - provider: deepseek
    model: deepseek-v3
    base_model: deepseek-v3
    mode: chat
    prices:
      input_cost_per_token: 0.000002
```

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| imported_count | 2 | Equals |
| skipped_count | 0 | Equals |
| errors | 空数组 | Len=0 |

---

#### MP-1-002：merge 模式增量合并（正常参数）

##### 设计思路

验证 merge 模式对已存在 `(provider, model, mode)` 记录进行更新，不存在的记录插入。

##### 前提数据准备

已存在 `(openai, gpt-4o, chat)` 记录。

##### 执行步骤

1. 构造 YAML：1 条与已存在记录同组合键但价格不同，1 条全新记录。
2. 以 `mode=merge` 调用导入接口。
3. 验证 `imported_count=2`。
4. 查询已存在记录，验证价格已更新；查询新记录，验证已插入。

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| imported_count | 2 | Equals |

---

#### MP-1-004：非法 mode（异常参数）

##### 设计思路

验证 mode 仅支持 replace / merge。

##### 执行步骤

1. 上传任意合法 YAML，mode=invalid。
2. 验证返回 422。

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 import mode 错误信息

---

## 8. 新增单条记录（MP-2）

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 新增单条记录 |
| 方法 | POST |
| 路径 | `/open-api/v1/model-prices` |
| 说明 | 创建单条模型定价 |

### 8.2 接口参数说明

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| provider | string | Y | Provider 标识 | 非空 |
| model | string | Y | 模型名 | 非空 |
| base_model | string | Y | 归一化模型名 | 非空 |
| mode | string | Y | 请求模式 | 枚举值 |
| capabilities | []string | N | 能力列表 | 元素为枚举值 |
| supported_parameters | []string | N | 支持的请求参数 | 元素为枚举值 |
| limits | object | N | 限制对象 | 键名为枚举值 |
| prices | object | Y | 价格对象 | 至少一个键；值为非负数；键名为枚举值 |
| metadata | object | N | 元数据 | 键名为枚举值 |

#### 返回数据字段

返回完整记录，含系统生成的 `id`、`price_currency`（固定 RMB）、`created_at`、`updated_at`。

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-2-001 | 最小参数创建 | 正常参数 | 仅必填字段 |
| MP-2-002 | 完整参数创建 | 正常参数 | 含所有可选字段 |
| MP-2-003 | 缺少 provider | 必填校验 | 返回 422 |
| MP-2-004 | 非法 mode | 合法性条件 | mode=invalid，返回 422 |
| MP-2-005 | prices 包含负数 | 合法性条件 | 返回 422 |
| MP-2-006 | 重复三元组 | 异常参数 | 返回 422 |

### 8.4 测试场景详细设计

#### MP-2-001：最小参数创建（正常参数）

##### 设计思路

验证仅传必填字段即可创建成功，且返回默认值正确。

##### 执行步骤

1. POST `/open-api/v1/model-prices`，Body 仅含必填字段。
2. 验证返回 200，返回记录中 `price_currency=RMB`。

##### 请求参数

```json
{
    "provider": "test-provider",
    "model": "test-model",
    "base_model": "test-model",
    "mode": "chat",
    "prices": {
        "input_cost_per_token": 0.001
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 大于 0 | Gt(0) |
| provider | "test-provider" | Equals |
| price_currency | "RMB" | Equals |
| created_at | 大于 0 | Gt(0) |

---

#### MP-2-005：prices 包含负数（合法性条件）

##### 设计思路

验证 prices 中所有价格必须为非负数。

##### 执行步骤

1. POST 请求，`prices.input_cost_per_token=-1`。
2. 验证返回 422。

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 price 错误信息

---

## 9. 分页列表查询（MP-3）

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 分页列表查询 |
| 方法 | GET |
| 路径 | `/open-api/v1/model-prices` |
| 说明 | 支持 provider、mode 过滤 |

### 9.2 接口参数说明

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 | 默认值/范围 |
|--------|------|------|------|------------|
| provider | string | N | 按 provider 过滤 | - |
| mode | string | N | 按 mode 过滤 | 枚举值 |
| page | int | N | 页码 | 1 |
| page_size | int | N | 每页条数 | 50，最大 1000 |

#### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| total | int64 | 总条数 |
| items | []ModelPrice | 模型定价记录列表 |

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-3-001 | 默认分页 | 正常参数 | page=1, page_size=50 |
| MP-3-002 | 自定义分页 | 边界值 | page=1, page_size=1 |
| MP-3-003 | 按 provider 过滤 | 正常参数 | 仅返回该 provider 记录 |
| MP-3-004 | 按 mode 过滤 | 正常参数 | 仅返回该 mode 记录 |

### 9.4 测试场景详细设计

#### MP-3-001：默认分页（正常参数）

##### 设计思路

验证无参数时返回默认分页数据。

##### 前提数据准备

已创建至少 2 条记录。

##### 执行步骤

1. GET `/open-api/v1/model-prices`。
2. 验证返回 200，`items` 非空，`total ≥ 2`。

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| total | ≥ 2 | Gte(2) |
| items | 非空数组 | NotEmpty |

---

## 10. 按 ID 查询单条（MP-4）

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按 ID 查询单条 |
| 方法 | GET |
| 路径 | `/open-api/v1/model-prices/{id}` |

### 10.2 接口参数说明

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | Y | 记录 ID |

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-4-001 | 查询存在的记录 | 正常参数 | 返回完整记录 |
| MP-4-002 | 查询不存在的记录 | 异常参数 | 返回 404 |

### 10.4 测试场景详细设计

#### MP-4-001：查询存在的记录（正常参数）

##### 设计思路

验证按 id 可返回正确记录。

##### 前提数据准备

已创建 1 条记录，记录其 id。

##### 执行步骤

1. GET `/open-api/v1/model-prices/{id}`。
2. 验证返回 200，Data.id 与请求 id 一致。

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 与请求一致 | Equals |
| provider | 非空 | NotEmpty |

---

#### MP-4-002：查询不存在的记录（异常参数）

##### 设计思路

验证查询不存在的 id 返回 404。

##### 执行步骤

1. GET `/open-api/v1/model-prices/999999999`。
2. 验证返回 404。

##### 预期返回结果

**ErrNum**：404

---

## 11. 按组合键查询（MP-5）

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按组合键查询 |
| 方法 | GET |
| 路径 | `/open-api/v1/model-prices` |
| 说明 | 通过 provider + model + mode 过滤查询；由于当前 GET `/model-prices` 为列表接口，返回符合过滤条件的列表（命中时 1 条，未命中时 0 条） |

### 11.2 接口参数说明

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| provider | string | Y | Provider 标识 |
| model | string | Y | 模型名 |
| mode | string | Y | 请求模式 |

### 11.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-5-001 | 查询存在的组合键 | 正常参数 | 返回只包含 1 条记录的列表 |
| MP-5-002 | 缺少 query 参数 | 边界场景 | 按现有参数进行列表过滤 |
| MP-5-003 | 查询不存在的组合键 | 异常参数 | 返回空列表 |

### 11.4 测试场景详细设计

#### MP-5-001：查询存在的组合键（正常参数）

##### 设计思路

验证通过 provider/model/mode 可唯一定位记录；当前实现走列表接口，返回总条数为 1 的列表。

##### 前提数据准备

已存在 `(deepseek, deepseek-v3, chat)` 记录。

##### 执行步骤

1. GET `/open-api/v1/model-prices?provider=deepseek&model=deepseek-v3&mode=chat`。
2. 验证返回 200，`total=1`，`items[0]` 的字段匹配。

##### 预期返回结果

**ErrNum**：200

---

## 12. 按 ID 修改单条（MP-6）

### 12.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按 ID 修改单条 |
| 方法 | PUT |
| 路径 | `/open-api/v1/model-prices/{id}` |
| 说明 | 支持部分字段更新 |

### 12.2 接口参数说明

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | Y | 记录 ID |

#### Body 参数

同 MP-2，仅传需修改字段。

### 12.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-6-001 | 部分更新 prices | 正常参数 | 修改价格后查询一致 |
| MP-6-002 | 更新不存在的记录 | 异常参数 | 返回 404 |
| MP-6-003 | 更新为非法 mode | 合法性条件 | 返回 422 |

### 12.4 测试场景详细设计

#### MP-6-001：部分更新 prices（正常参数）

##### 设计思路

验证仅传 prices 可更新成功，未传字段保持原值。

##### 前提数据准备

已存在记录。

##### 执行步骤

1. PUT `/open-api/v1/model-prices/{id}`，Body 仅含新 prices。
2. 验证返回 200，Data.prices 为新值，provider/model 保持原值。
3. 再次 GET 该记录，验证一致。

##### 请求参数

```json
{
    "prices": {
        "input_cost_per_token": 0.00001,
        "output_cost_per_token": 0.00002
    }
}
```

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| prices.input_cost_per_token | 0.00001 | Equals |
| prices.output_cost_per_token | 0.00002 | Equals |

---

## 13. 按组合键修改单条（MP-7）

### 13.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按组合键修改单条 |
| 方法 | PUT |
| 路径 | `/open-api/v1/model-prices` |
| 说明 | 通过 provider + model + mode 定位并更新 |

### 13.2 接口参数说明

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| provider | string | Y | Provider 标识 |
| model | string | Y | 模型名 |
| mode | string | Y | 请求模式 |

#### Body 参数

同 MP-6。

### 13.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-7-001 | 按组合键更新 prices | 正常参数 | 更新成功 |
| MP-7-002 | 缺少 query 参数 | 必填校验 | 返回 422 |

---

## 14. 按 ID 删除单条（MP-8）

### 14.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按 ID 删除单条 |
| 方法 | DELETE |
| 路径 | `/open-api/v1/model-prices/{id}` |

### 14.2 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-8-001 | 删除存在的记录 | 正常参数 | 删除后再查询返回 404 |
| MP-8-002 | 删除不存在的记录 | 异常参数 | 返回 404 |

### 14.3 测试场景详细设计

#### MP-8-001：删除存在的记录（正常参数）

##### 设计思路

验证删除后记录不可再查询。

##### 前提数据准备

已创建 1 条记录。

##### 执行步骤

1. DELETE `/open-api/v1/model-prices/{id}`。
2. 验证返回 200，`Data.deleted=true`。
3. 再次 GET 该记录，验证返回 404。

##### 预期返回结果

**ErrNum**：200  
**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| deleted | true | Equals |

---

## 15. 按组合键删除单条（MP-9）

### 15.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Model Price |
| 接口名称 | 按组合键删除单条 |
| 方法 | DELETE |
| 路径 | `/open-api/v1/model-prices` |
| 说明 | 通过 provider + model + mode 定位并删除 |

### 15.2 接口参数说明

#### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| provider | string | Y | Provider 标识 |
| model | string | Y | 模型名 |
| mode | string | Y | 请求模式 |

### 15.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| MP-9-001 | 删除存在的组合键记录 | 正常参数 | 删除成功 |
| MP-9-002 | 缺少 query 参数 | 必填校验 | 返回 422 |

---

## 16. 附录

### 16.1 通用断言说明

- `Equals`：字段值与预期值相等
- `NotEmpty`：字段值非空
- `Gte(n)`：数值大于等于 n
- `Len=n`：数组/字符串长度等于 n
- `IsArray` / `IsObject`：类型校验

### 16.2 返回码约定

| ErrNum | 含义 |
|--------|------|
| 200 | 成功 |
| 404 | 记录不存在 |
| 422 | 参数校验失败 |
