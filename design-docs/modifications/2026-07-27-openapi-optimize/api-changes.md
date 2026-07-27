# OpenAPI 接口变更说明

## 版本对比说明

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-27 |
| 旧版本文件 | `ai-gateway-api/design-docs/api-define/OpenAPI接口定义.md` |
| 新版本文件 | `ai-gateway-api/design-docs/modifications/2026-07-27-openapi-optimize/瑛菲AI网关-OpenAPI接口定义v0.3.0.md` |

> 注：本次变更是对当前基线文档的一次优化收敛，目标版本号为 **v0.3.0**。下文以“旧版”指代基线文档，“新版”指代 v0.3.0 优化稿。

---

## 一、目录/模块结构调整

### 1.1 新增模块

| 章节 | 模块 | 说明 |
|------|------|------|
| 11 | `/model-provider-types` | 独立章节，提供 AI 模型提供商类型列表查询 |
| 12 | `/tools` | 独立章节，提供“从指定提供商获取 AI 模型列表”工具接口 |

### 1.2 移除模块

以下整章从接口定义中移除：

| 章节 | 模块 | 说明 |
|------|------|------|
| 11 | `/ai-route-rules` | AI 路由规则模块整体移除 |
| 12 | `/global-models` | 全局模型模块整体移除 |
| 13 | `/products/{product_name}/models` | 产品可用模型别名查询模块整体移除 |
| 14 | `/general/actions/exec-api` | 代理执行外部 API 模块整体移除 |

---

## 二、通用规范变更

### 2.1 返回值格式

所有 API 返回值中 **移除 `WorkMode` 字段**。

**变更前：**

```json
{
    "ErrNum": 200,
    "Data": json_object,
    "ErrMsg": "string message",
    "WorkMode": "current mode"
}
```

**变更后：**

```json
{
    "ErrNum": 200,
    "Data": json_object,
    "ErrMsg": "string message"
}
```

### 2.2 错误码变化

| ErrNum | 旧版含义 | 新版含义 | 变化 |
|--------|----------|----------|------|
| 401 | 未定义 | 鉴权失败 | **新增** |
| 402 | 没有调用权限 | 没有调用权限造成的失败 | 语义细化 |
| 404 | 查询/修改/删除不存在的对象 | 查询/修改/删除不存在的对象时 | 语义细化 |
| 409 | 资源冲突 | 资源依赖冲突时 | 语义细化 |
| 422 | 参数不合法 | 参数不合法造成的失败 | 语义细化 |
| 500 | 其他业务逻辑错误 | 其他业务逻辑错误，一律返回 500 | 语义细化 |
| 555 | 产品线内重复 | 创建重复对象时 | 语义泛化 |
| 556 | 全局重复 | 数据重复时 | 语义泛化 |

> **影响：** 调用方需新增对 `401`（鉴权失败）的处理逻辑，并注意错误码描述文本的调整。

### 2.3 Method 约定与通用 Query 参数

- **Method 约定**：无变化（GET/POST/PUT/PATCH/DELETE 语义保持一致）
- **通用 Query 参数（列表接口）**：`page`、`page_size`、`sort_by`、`sort_order` 无变化

---

## 三、数据模型变更

### 3.1 /api-keys 数据模型

#### 3.1.1 支持导入外部 API-Key

创建 API-Key 时新增可选字段 `key`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | N | 若传入则使用该值作为 API-Key；若未传则由后台生成。用于从其他系统导入 API-Key。传入值需在系统中全局唯一，重复返回 422。 |

> 更新接口（`PUT /api-keys/{id}`、`PATCH /api-keys/{id}`）中传入 `key` 不生效，API-Key 值不可通过更新接口修改。

#### 3.1.2 列表/详情返回中 quota_plan 包含 balance

根据 `ai-gateway-api` 代码实现（`endpoints/openapi_v1/api_key/list.go`、`model/icluster_conf/api_key.go`）：

- `GET /api-keys`：返回**分页结构** `{list, pagination}`，列表元素中 `quota_plan` 在非 `unlimited` 情况下会包含 `balance` 字段（`used`、`remaining`）。
- `GET /api-keys/{id}`：返回中 `quota_plan` 同样包含 `balance` 字段。

> **说明**：旧版接口定义中“列表接口 `quota_plan` 不含 `balance`”的描述与代码实现不符；新版接口定义已修正为明确声明列表返回包含 `balance`。

### 3.2 /clusters 数据模型（重大变更）

#### 3.2.1 顶层字段调整

| 字段 | 变更 | 说明 |
|------|------|------|
| `ready` | **删除** | 不再对外暴露集群就绪状态 |
| `sub_clusters` | **删除** | 子集群为系统内部概念，不再对外暴露；每个集群只包含一个子集群 |
| `scheduler` | **删除** | 调度为系统内部概念，不再对外暴露；固定 `GSLB_BLACKHOLE=0` |

#### 3.2.2 Instance 结构

| 字段 | 变更 | 说明 |
|------|------|------|
| `tags` | **删除** | 实例标签不再对外暴露 |

#### 3.2.3 basic / retries 调整

| 字段 | 变更 | 说明 |
|------|------|------|
| `basic.retries.max_retry_in_subcluster` | **重命名** | 改为 `max_retry_in_cluster`（底层仍对应原字段） |
| `basic.retries.max_retry_cross_subcluster` | **删除** | 跨子集群重试字段移除 |

#### 3.2.4 sticky_sessions 调整

| 字段 | 变更 | 说明 |
|------|------|------|
| `sticky_sessions.session_sticky_type` | **删除** | 默认且仅支持实例级会话保持 |
| `sticky_sessions.enabled` | **新增** | bool，默认 `false`，用于控制会话保持开关 |

#### 3.2.5 passive_health_check 调整

所有字段由**必填改为选填**，并补充默认值：

| 字段 | 新版默认值 |
|------|-----------|
| `failnum` | 3 |
| `interval` | 1000 ms |
| `host` | 空（使用 `instance_pool` 首个实例的 `hostname`） |
| `uri` | `/` |
| `statuscode` | 0 |

#### 3.2.6 llm_config 调整

| 字段 | 变更 | 说明 |
|------|------|------|
| `service_name` | **删除** | - |
| `group` | **删除** | - |
| `model_endpoint` | **改为可选** | `schema` 默认 `https`，`uri` 默认 `/v1/models` |
| `models` | 不变 | 仍为必填 |

> `llm_config.enable` 字段继续保持已移除状态；设置 `llm_config` 即默认开启 AI 网关能力。

### 3.3 /auth 数据模型（重大变更）

#### 3.3.1 User 模型

| 字段 | 旧版 | 新版 |
|------|------|------|
| `is_admin` | `true`：System 权限；`false`：Product 权限 | **仅支持 `true`（System 权限）**，`false` 暂不支持 |

#### 3.3.2 Token 模型

| 字段 | 旧版 | 新版 |
|------|------|------|
| `product_name` | Scope 为 Product 时必填 | **删除** |
| `scope` | `System` / `Product` / `Support` | **仅保留 `System` / `Support`**，移除 `Product` |

> 鉴权方式（Session Key / Token）和 Scope 的通用说明同步精简，不再涉及产品线权限校验。

### 3.4 /certificates 数据模型

数据模型本身无字段变化，但新增 **证书详情查询接口** `GET /certificates/{cert_name}`。

---

## 四、接口级变更

### 4.1 新增接口

| 模块 | 接口 | Method | 端点 | 说明 |
|------|------|--------|------|------|
| `/model-provider-types` | 获取AI模型提供商类型列表 | GET | `/model-provider-types` | 新增独立模块，返回字符串数组 |
| `/tools` | 从指定提供商获取AI模型列表 | POST | `/tools/get-models-from-provider` | 新增独立模块，替代原 `/models` |
| `/certificates` | 证书详情 | GET | `/certificates/{cert_name}` | 新增单个证书查询接口 |

### 4.2 删除接口

#### 4.2.1 整模块删除

| 模块 | 删除接口 | 说明 |
|------|----------|------|
| `/ai-route-rules` | `GET /ai-route-rules`<br>`PATCH /ai-route-rules` | 模块整体移除 |
| `/global-models` | `GET /global-models` | 模块整体移除 |
| `/products/{product_name}/models` | `GET /products/{product_name}/models` | 模块整体移除 |
| `/general/actions/exec-api` | `POST /general/actions/exec-api` | 模块整体移除 |

#### 4.2.2 /api-keys 删除

| 接口 | 端点 | 说明 |
|------|------|------|
| 生成API-Key | `GET /api-keys/actions/generate` | 创建前预览/生成 Key 的接口移除 |

#### 4.2.3 /clusters 删除/迁移

| 接口 | 端点 | 说明 |
|------|------|------|
| 集群就绪状态获取 | `GET /clusters/{cluster_name}/ready` | 删除 |
| 获取AI模型提供商列表 | `GET /model-providers` | 迁移至 `/model-provider-types` |
| 获取AI模型列表 | `POST /models` | 迁移至 `/tools/get-models-from-provider` |

#### 4.2.4 /auth 删除

| 接口 | 端点 | 说明 |
|------|------|------|
| 为用户增加某个产品线的授权 | `POST /auth/users/{user_name}/products/{product_name}` | 删除 |
| 对用户取消某个产品线的授权 | `DELETE /auth/users/{user_name}/products/{product_name}` | 删除 |
| 获取对指定产品线有权限的用户列表 | `GET /auth/users/actions/search-by-product/{product_name}` | 删除 |
| 获取对指定产品线有权限的Token列表 | `GET /auth/tokens/actions/search-by-product/{product_name}` | 删除 |

### 4.3 参数/返回调整的接口

#### 4.3.1 /api-keys

| 接口 | 变更项 | 旧版 | 新版 |
|------|--------|------|------|
| `POST /api-keys` | Body 参数 | 无 `key` | 新增可选 `key` |
| `GET /api-keys` | 返回结构 | 分页结构；误写 `quota_plan` 不含 `balance` | 分页结构；明确声明 `quota_plan` 含 `balance` |
| `GET /api-keys/{id}` | 返回结构 | `quota_plan` 不含 `balance`（文档描述，与代码不符） | 明确声明 `quota_plan` 含 `balance` |
| `GET /api-keys/{id}/quota-plan` | 端点 | `/api-keys/{id}/quota-plan`（URI 参数名为 `key`） | 接口定义正文为 `/api-keys/{id}/quota-plan`；历史版本记录中残留 `/api-keys/{key}/quota-plan` 为笔误 |

#### 4.3.2 /clusters

| 接口 | 变更项 | 说明 |
|------|--------|------|
| `POST /clusters` | 必填项调整 | `basic`、`sticky_sessions`、`passive_health_check` 改为**可选**；`llm_config` 改为**必填**；未传字段使用 AI 网关场景默认推荐值 |
| `GET /clusters` | 返回字段 | 移除 `ready`、`sub_clusters`、`scheduler` |
| `GET /clusters/{cluster_name}` | 返回字段 | 移除 `ready`、`sub_clusters`、`scheduler` |
| `PATCH /clusters/{cluster_name}` | 可修改字段 | 不再支持修改 `sub_clusters` / `scheduler`，通过 `instance_pool` 调整实例 |
| `DELETE /clusters/{cluster_name}` | 返回字段 | 返回创建接口同结构（已移除系统内部字段） |

#### 4.3.3 /auth

| 接口 | 变更项 | 旧版 | 新版 |
|------|--------|------|------|
| `POST /auth/users` | Body 参数 | `password` 条件必填；`is_admin` 可选；有 `type` 字段 | `password` **必填**；`is_admin` **选填且固定为 `true`**；**删除 `type` 字段** |
| `GET /auth/users` | 返回字段 | 含 `user_name`、`is_admin`、`products` | 仅含 `user_name`、`is_admin`（固定 `true`） |
| `PATCH /auth/users/{user_name}/is_admin` | Body 参数 | `is_admin` 可设置为 `true`/`false` | `is_admin` **固定为 `true`** |
| `POST /auth/session-keys` | 返回字段 | 含 `is_admin`、`products` | `is_admin` 固定返回 `true`，**删除 `products`** |
| `POST /auth/tokens` | Body 参数 | 含 `name`、`scope`、`product_name` | **删除 `product_name`**，`scope` 仅支持 `System`/`Support` |
| `GET /auth/tokens/{token_name}` | 返回字段 | 含 `name`、`product_name`、`token`、`scope` | **删除 `product_name`** |
| `GET /auth/tokens` | 返回字段 | 元素含 `product_name` | **删除 `product_name`** |

---

## 五、关键业务流程变更

关键业务流程的核心逻辑保持不变：

1. **运行时执行顺序**仍为：模型访问控制检查 → 限流检查 → 配额扣减。
2. **创建 API-Key 流程**和 **创建 Entity 流程**步骤不变。
3. **运行时模型访问控制**、**限流检查**、**配额扣减**逻辑不变。
4. **配置变更的级联与隔离**内容不变。
5. 对象关系图同步更新：删除 `/ai-route-rules`、`/global-models` 相关对象，新增 `RouteRules` 统一结构以反映 API-Key、Entity、Global 路由表均通过 `RouteRules` 管理规则。

---

## 六、版本修改记录（v0.3.0 主要变更点）

### 6.1 模块增删（当前版本 → v0.3.0）

- **新增**：`/model-provider-types`、`/tools`
  - `/model-provider-types`：从 `/clusters` 中独立出来，提供 AI 模型提供商类型列表查询
  - `/tools`：新增工具接口，提供“从指定提供商获取 AI 模型列表”能力
- **删除**：`/ai-route-rules`、`/global-models`、`/products/{product_name}/models`、`/general/actions/exec-api`

### 6.2 通用规范

- 返回值中**移除 `WorkMode` 字段**
- **新增错误码 401**，并细化既有错误码描述

### 6.3 API-Key

- 创建接口**支持导入外部 `key`**
- 列表/详情接口返回中 `quota_plan` **包含 `balance`**

### 6.4 Auth

- 用户 `is_admin` **仅支持 `true`（System 权限）**
- Token **删除 `product_name`**，`scope` 仅保留 `System`/`Support`
- 删除产品线授权相关接口

### 6.5 Clusters

- 数据模型**删除 `ready`、`sub_clusters`、`scheduler`**
- `Instance` 删除 `tags`
- `sticky_sessions` 删除 `session_sticky_type`，新增 `enabled`
- `retries` 中 `max_retry_in_subcluster` 重命名为 `max_retry_in_cluster`，删除 `max_retry_cross_subcluster`
- `passive_health_check` 全字段改为选填并补充默认值
- `llm_config` 删除 `service_name`、`group`，`model_endpoint` 改为可选
- 创建集群接口：`basic`/`sticky_sessions`/`passive_health_check` 改为可选，`llm_config` 改为必填
- 删除集群就绪状态接口；模型提供商/模型列表接口迁移至独立模块

### 6.6 Certificates

- 新增 **证书详情查询接口** `GET /certificates/{cert_name}`

### 6.7 对象关系图

- 对象关系图同步重构，反映模块精简后的对象关系

---

*文档生成日期：2026-07-27*
