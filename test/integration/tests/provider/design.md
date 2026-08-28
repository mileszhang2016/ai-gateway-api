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
| PV-2 | 查询 Provider 列表 | GET | `/open-api/v1/providers` | 未携带 `page`/`page_size` 时返回全部；携带时按分页返回；支持 `model_protocol` 过滤 |
| PV-3 | 查询 Provider 详情 | GET | `/open-api/v1/providers/{provider_name}` | - |
| PV-4 | 更新 Provider | PATCH | `/open-api/v1/providers/{provider_name}` | 全量替换 `keys`、`models` 等数组字段 |
| PV-5 | 删除 Provider | DELETE | `/open-api/v1/providers/{provider_name}` | 删除前检查 cluster/model-price 引用 |
| PV-6 | 触发模型发现 | POST | `/open-api/v1/providers/tools/discover-models` | 无状态工具接口，不绑定 Provider |
| PV-7 | 获取所有 Provider 名称 | GET | `/providers/actions/get-provider-names` | 返回全量 provider 名称列表 |
| PV-8 | 设置高峰/闲时模板 | PUT | `/providers/{provider_name}/pricing-tiers` | 支持 JSON / YAML / multipart 上传 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建 Provider | 14 |
| 查询 Provider 列表 | 3 |
| 查询 Provider 详情 | 2 |
| 更新 Provider | 4 |
| Provider instance_pool 同步到 Inner API | 1 |
| 删除 Provider | 3 |
| 触发模型发现 | 6 |
| 获取所有 Provider 名称 | 1 |
| 设置高峰/闲时模板 | 10 |
| **合计** | **38** |

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
├── instance_pool_sync/
│   └── instance_pool_sync_test.go
├── delete/
│   └── delete_test.go
└── pricing_tiers/
    └── pricing_tiers_test.go
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
| `model_protocols` | []string | Y | 支持的模型访问协议，至少 1 个元素，枚举：`openai`、`anthropic` |
| `time_zone` | string | N | 计算时段所使用的时区，默认 `Asia/Shanghai` |
| `tiers` | []object | N | 时段 tier 定义列表，**初期 `name` 只支持 `peak`** |

##### `instance_pool` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `addr` | string | Y | 实例地址 |
| `weight` | int | Y | 权重，取值范围 [0,100]；至少有一个实例 `weight > 0` |
| `port` | int | Y | 实例端口 |

##### `keys` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | Y | Key 名称，同一 Provider 内唯一，长度 1-128 |
| `key` | string | Y | API-Key 明文，长度 1-512 |

##### `tiers` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `name` | string | Y | Tier 名称，**初期只支持 `peak`** |
| `time_ranges` | []object | Y | 时段范围列表，命中任意一个即属于该 tier |

##### `time_ranges` 元素结构

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `weekdays` | []int | N | 星期几，0=周日，...，6=周六；为空表示每天 |
| `start` | string | Y | 开始时间，格式 `HH:MM` |
| `end` | string | Y | 结束时间，格式 `HH:MM`；`end` 必须大于 `start` |

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
| `time_zone` | string | 时区 |
| `tiers` | []object | 时段 tier 列表 |
| `create_time` | int64 | 创建时间 |
| `update_time` | int64 | 更新时间 |

### 6.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-1-001 | 最小参数创建 Provider | 200，返回的 `description` 为空字符串，`models`/`keys` 为空数组 |
| PV-1-002 | 完整参数创建 Provider | 200，返回的 `models`、`keys`、`instance_pool` 与输入一致 |
| PV-1-002a | 创建 anthropic 协议 Provider | 200，返回的 `model_protocols` 为 `["anthropic"]` |
| PV-1-003 | 重复 Provider 名称 | 555 |
| PV-1-004 | 缺少 `instance_pool` | 422 |
| PV-1-005 | 缺少 `model_protocols` | 422 |
| PV-1-006 | 非法 `model_protocols` | 422 |
| PV-1-007 | `instance_pool` 为空数组 | 422 |
| PV-1-008 | 实例 `port` 非法 | 422 |
| PV-1-009 | 非法 `name` | 422 |
| PV-1-010 | `weight` 超过 100 | 422 |
| PV-1-011 | 重复实例 `(addr, port)` 组合 | 422 |
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
| `page` | int | N | 页码；未传时与 `page_size` 同时缺失，返回全部记录 |
| `page_size` | int | N | 每页条数，最大 1000；未传时与 `page` 同时缺失，返回全部记录 |
| `model_protocol` | string | N | 按协议过滤，须为 `model_protocols` 枚举值 |

### 7.3 响应参数

始终返回 `{ list, pagination }` 结构：

- 未携带 `page`/`page_size` 时，`list` 包含全部匹配记录，`pagination.page=1`，`pagination.page_size=total`；
- 携带分页参数时，`list` 包含当前页记录，`pagination` 为对应分页信息。

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
| PV-2-001 | 无分页参数返回全部 | 200，`pagination.total >= 2`，`list` 长度等于 `pagination.total` |
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
| PV-3-001 | 查询存在的 Provider | 200，返回的 `name` 与请求一致；`time_zone` 为 `Asia/Shanghai`；`tiers` 为 `[]` |
| PV-3-002 | 查询不存在的 Provider | 404 |
| PV-3-003 | 查询已设置 pricing-tiers 的 Provider | 200，返回的 `time_zone` 和 `tiers` 与设置一致 |

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

可修改字段与创建接口一致。`provider_name` 由 URI 路径指定，请求体中无需再传 `name`（若包含 `name` 返回 422）；更新时仍需携带 `instance_pool` 和 `model_protocols`。

### 9.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-4-001 | 更新 `description` | 200，`description` 更新成功 |
| PV-4-002 | 更新 `models` | 200，`models` 长度变为 3 |
| PV-4-003 | 更新 `keys`（全量替换） | 200，`keys` 被替换为新的完整列表 |
| PV-4-004 | 更新不存在的 Provider | 404 |
| PV-4-005 | 请求体不包含 `name` | 200，返回的 `name` 与 URI 一致 |
| PV-4-006 | 请求体包含 `name` | 422 |

## 10. Provider instance_pool 同步到 Inner API

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider / Inner API |
| 接口名称 | 修改 Provider instance_pool 后验证 cluster_table 导出 |
| 方法 | PATCH + GET |
| 路径 | `/open-api/v1/providers/{provider_name}` + `/inner-api/v1/configs/gslb_data/cluster_table` |
| 说明 | 验证 Open API 修改 provider 实例池后，Inner API 的 cluster_table 配置导出会同步更新 |

### 10.2 测试步骤

1. 创建 Provider，初始 `instance_pool` 为 `[{name: "backend-old", addr: "10.0.0.1", port: 8080, weight: 100}]`。
2. 创建 Cluster，通过 `llm_config.provider` 引用该 Provider。
3. 调用 `GET /inner-api/v1/configs/gslb_data/cluster_table`，断言 cluster 的 backends 包含 `backend-old`。
4. 调用 `PATCH /open-api/v1/providers/{provider_name}`，将 `instance_pool` 修改为 `[{name: "backend-new", addr: "10.0.0.2", port: 8081, weight: 100}]`。
5. 再次调用 `GET /inner-api/v1/configs/gslb_data/cluster_table`，断言 backends 包含 `backend-new`，且不再包含 `backend-old`。

### 10.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PV-SYNC-1-001 | 修改 provider instance_pool 后 cluster_table 同步更新 | cluster_table 中 cluster 的 backend 变为新实例 |

## 11. 删除 Provider

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

## 12. 设置高峰/闲时模板（PV-8）

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Provider |
| 接口名称 | 设置高峰/闲时模板 |
| 方法 | PUT |
| 路径 | `/providers/{provider_name}/pricing-tiers` |
| 说明 | 单独维护 provider 的 `time_zone` 和 `tiers`；支持 JSON / YAML / multipart 上传 |

### 11.2 请求参数

#### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `provider_name` | string | Y | Provider 名称，必须已存在 |

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `time_zone` | string | N | 时区，默认 `Asia/Shanghai`；须为合法 IANA 时区名 |
| `tiers` | []object | N | 时段 tier 定义列表，结构与创建 Provider 的 `tiers` 一致 |

### 11.3 测试用例

| 用例编号 | 用例名称 | 预期结果 |
|----------|----------|----------|
| PT-1-001 | JSON 设置高峰模板 | 200，返回的 `time_zone`/`tiers` 与输入一致 |
| PT-1-002 | text/yaml 设置高峰模板 | 200，返回的 `time_zone`/`tiers` 与输入一致 |
| PT-1-003 | multipart/form-data YAML 设置高峰模板 | 200，返回的 `time_zone`/`tiers` 与输入一致 |
| PT-1-004 | provider 不存在 | 404 |
| PT-1-005 | 非法 `time_zone` | 422 |
| PT-1-006 | 非法 tier name | 422 |
| PT-1-007 | 同一 tier 时间重叠 | 422 |
| PT-1-008 | `end <= start` | 422 |
| PT-1-009 | `weekdays` 越界 | 422 |
| PT-1-010 | GET provider 返回 tiers | 200，响应中包含 `time_zone`/`tiers` |

## 12. 依赖与数据准备

1. 测试环境数据库需包含产品线初始化数据，以支持 cluster 创建。
2. 创建 Provider 后如需验证删除冲突，需先创建引用该 Provider 的 Cluster。
3. 测试环境 `SkipTokenValidate=true`，无需认证头。

## 13. 注意事项

1. Provider 名称全局唯一，测试中使用 `testutil.UniqueProviderName()` 生成唯一名称。
2. `keys` 作为数组，按**全量替换**处理；更新时如需保留旧 Key，需传入完整列表。
3. `instance_pool` 中 `(name, addr, port)` 组合不能重复；`name` 为空字符串时系统会默认填充为 `addr`。
4. `model_protocols` 当前枚举值为 `openai`、`anthropic`。
5. `/providers/tools/discover-models` 集成测试见 `discover/discover_test.go`，覆盖 OpenAI/Anthropic 协议、默认 URI、参数校验等场景。
6. `PUT /providers/{provider_name}/pricing-tiers` 只更新 `time_zone`/`tiers`，不会覆盖 provider 的其他字段。
