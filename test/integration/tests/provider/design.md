# Provider 测试用例设计文档

## 1. 模块概述

Provider 模块负责 AI 网关**模型提供商**的管理，包括创建、查询、更新、删除，以及触发模型发现。

Provider 与 Cluster 概念分离后：

- Provider 承载后端实例池 `instance_pool`、真实 API-Key `keys`、支持的模型 `models`、模型访问协议 `model_protocols`、模型发现端点 `model_endpoint`。
- Cluster 通过 `llm_config.provider` 引用已有 Provider，并通过 `llm_config.keys` 声明 `{name, weight}` 数组选择 Provider 下要使用的 Key。
- 删除 Provider 前，会检查是否有 Cluster 或 Model-Price 记录引用该 Provider；若存在引用，则删除被拒绝（返回 409）。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| PV-1 | 创建 Provider | POST | `/open-api/v1/providers` | 创建模型提供商 |
| PV-2 | 查询 Provider 列表 | GET | `/open-api/v1/providers` | 支持分页与 `model_protocol` 过滤 |
| PV-3 | 查询 Provider 详情 | GET | `/open-api/v1/providers/{provider_name}` | - |
| PV-4 | 更新 Provider | PATCH | `/open-api/v1/providers/{provider_name}` | 全量替换 `keys`、`models` 等数组字段 |
| PV-5 | 删除 Provider | DELETE | `/open-api/v1/providers/{provider_name}` | 删除前检查 cluster/model-price 引用 |
| PV-6 | 触发模型发现 | POST | `/open-api/v1/providers/{provider_name}/discover-models` | 当前集成测试未覆盖 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建 Provider | 13 |
| 查询 Provider 列表 | 3 |
| 查询 Provider 详情 | 2 |
| 更新 Provider | 4 |
| 删除 Provider | 3 |
| **合计** | **25** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
provider/
├── design.md
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

## 6. 创建 Provider

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 创建 Provider |
| 方法 | POST |
| 路径 | `/open-api/v1/providers` |
| 说明 | 创建模型提供商，返回完整 Provider 对象 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | Y | Provider 唯一标识，全局唯一 |
| `description` | string | N | 描述信息，长度 0-256 |
| `model_endpoint` | object | N | 模型发现端点，默认 `{schema:"https", uri:"/v1/models"}` |
| `models` | []string | N | 支持的模型列表，元素非空且不可重复 |
| `keys` | []object | N | API-Key 列表，元素为 `{name, key}`；同一 Provider 内 `name` 唯一 |
| `instance_pool` | []object | Y | 后端实例池，至少 1 个元素 |
| `model_protocols` | []string | Y | 支持的模型访问协议，至少 1 个元素，枚举：`openai`、`anthropic`、`gemini` |

##### `instance_pool` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | N | 实例名称，未传时默认与 `addr` 相同；若传则长度 1-128 |
| `addr` | string | Y | 实例地址 |
| `weight` | int | Y | 权重，取值范围 [0,100]；至少有一个实例 `weight > 0` |
| `port` | int | Y | 实例端口 |

##### `keys` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | Y | Key 名称，同一 Provider 内唯一，长度 1-128 |
| `key` | string | Y | API-Key 明文，长度 1-512 |

#### 6.2.2 响应参数

| 参数名 | 类型 | 说明 |
|--------|------|------|
| `id` | int64 | Provider ID |
| `name` | string | Provider 名称 |
| `description` | string | 描述 |
| `model_endpoint` | object | 模型发现端点 |
| `models` | []string | 模型列表 |
| `keys` | []object | Key 列表（含明文） |
| `instance_pool` | []object | 实例池 |
| `model_protocols` | []string | 协议列表 |
| `create_time` | int64 | 创建时间 |
| `update_time` | int64 | 更新时间 |

### 6.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-1-001 | 最小参数创建 Provider | 200，返回的 `description` 为空字符串，`models`/`keys` 为空数组 |
| PV-1-002 | 完整参数创建 Provider | 200，返回的 `models`、`keys`、`instance_pool` 与输入一致 |
| PV-1-003 | 重复 Provider 名称 | 555 |
| PV-1-004 | 缺少 `instance_pool` | 422 |
| PV-1-005 | 缺少 `model_protocols` | 422 |
| PV-1-006 | 非法 `model_protocols` | 422 |
| PV-1-007 | `instance_pool` 为空数组 | 422 |
| PV-1-008 | 实例 `port` 非法 | 422 |
| PV-1-009 | 非法 `name` | 422 |
| PV-1-010 | `weight` 超过 100 | 422 |
| PV-1-011 | 重复实例 `(name, addr)` 组合 | 422 |
| PV-1-012 | `models` 元素重复 | 422 |
| PV-1-013 | `keys` 中 `name` 重复 | 422 |

## 7. 查询 Provider 列表

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 查询 Provider 列表 |
| 方法 | GET |
| 路径 | `/open-api/v1/providers` |
| 说明 | 支持分页与 `model_protocol` 过滤 |

### 7.2 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `page` | int | N | 页码，默认 1 |
| `page_size` | int | N | 每页条数，默认 50，最大 1000 |
| `model_protocol` | string | N | 按协议过滤，须为 `model_protocols` 枚举值 |

### 7.3 响应参数

返回分页结构：

```json
{
  "list": [...],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total": 100
  }
}
```

### 7.4 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-2-001 | 默认分页 | 200，`pagination.total >= 2`，`list` 非空 |
| PV-2-002 | 自定义分页 | 200，`list` 长度为 1 |
| PV-2-003 | 按 `model_protocol` 过滤 | 200，返回的 Provider 都包含 `openai` 协议 |

## 8. 查询 Provider 详情

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 查询 Provider 详情 |
| 方法 | GET |
| 路径 | `/open-api/v1/providers/{provider_name}` |

### 8.2 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-3-001 | 查询存在的 Provider | 200，返回的 `name` 与请求一致 |
| PV-3-002 | 查询不存在的 Provider | 404 |

## 9. 更新 Provider

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 更新 Provider |
| 方法 | PATCH |
| 路径 | `/open-api/v1/providers/{provider_name}` |
| 说明 | `keys`、`models` 等数组字段按全量替换处理 |

### 9.2 请求参数

可修改字段与创建接口一致。当前实现要求请求体中必须携带 `name`、`instance_pool`、`model_protocols`。

### 9.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-4-001 | 更新 `description` | 200，`description` 更新成功 |
| PV-4-002 | 更新 `models` | 200，`models` 长度变为 3 |
| PV-4-003 | 更新 `keys`（全量替换） | 200，`keys` 被替换为新的完整列表 |
| PV-4-004 | 更新不存在的 Provider | 404 |

## 10. 删除 Provider

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 删除 Provider |
| 方法 | DELETE |
| 路径 | `/open-api/v1/providers/{provider_name}` |
| 说明 | 删除前检查是否有 Cluster 或 Model-Price 引用 |

### 10.2 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-5-001 | 删除 Provider | 200，删除后查询返回 404 |
| PV-5-002 | 删除不存在的 Provider | 404 |
| PV-5-003 | 删除被 Cluster 引用的 Provider | 409 |

## 11. 依赖与数据准备

1. 测试环境数据库需包含产品线初始化数据，以支持 cluster 创建。
2. 创建 Provider 后如需验证删除冲突，需先创建引用该 Provider 的 Cluster。
3. 测试环境 `SkipTokenValidate=true`，无需认证头。

## 12. 注意事项

1. Provider 名称全局唯一，测试中使用 `testutil.UniqueProviderName()` 生成唯一名称。
2. `keys` 作为数组，按**全量替换**处理；更新时如需保留旧 Key，需传入完整列表。
3. `instance_pool` 中 `(name, addr)` 组合不能重复；`name` 为空字符串时系统会默认填充为 `addr`。
4. `model_protocols` 当前枚举值为 `openai`、`anthropic`、`gemini`。
5. 当前集成测试未覆盖 `/providers/{provider_name}/discover-models` 接口。
