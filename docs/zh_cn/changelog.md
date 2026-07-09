# Changelog

## 2026-07-07

### Bug Fixes

#### API-Key 接口

- **创建接口** (`endpoints/openapi_v1/api_key/create.go`)
  - 修复创建成功后返回 Data 为 nil 的问题，改为返回创建的 API-Key 数据

- **更新接口** (`endpoints/openapi_v1/api_key/update.go`)
  - 修复更新成功后返回 Data 为 nil 的问题，改为返回更新后的 API-Key 数据

- **全量更新接口** (`endpoints/openapi_v1/api_key/full_update.go`)
  - 修复全量更新成功后返回 Data 为 nil 的问题，改为返回更新后的 API-Key 数据

- **列表接口** (`endpoints/openapi_v1/api_key/list.go`)
  - 修复返回格式错误，改为返回包含 `list` 和 `pagination` 的对象
  - 添加 `enabled`、`entity_id`、`unlimited_quota` 参数过滤支持

- **详情接口** (`endpoints/openapi_v1/api_key/one.go`)
  - 修复不存在的 API-Key 返回 200 的问题，改为返回 404 错误

- **模型字段** (`model/icluster_conf/api_key.go`)
  - `Enable` 字段 JSON tag 从 `enable` 改为 `enabled`
  - `UpdatedTime` 字段 JSON tag 从 `updated_time` 改为 `update_time`

#### Entity 接口

- **创建接口** (`endpoints/openapi_v1/entity/create.go`)
  - 添加 `name` 必填校验
  - 添加类型存在性校验（校验传入的 type 是否存在于 entity_types 表）


- **详情接口** (`endpoints/openapi_v1/entity/one.go`)
  - 修复不存在的 Entity 返回 200 的问题，改为返回 404 错误

- **删除接口** (`model/quota/entity_manager.go`)
  - 添加子实体检查，存在子实体时返回 409 错误

- **模型字段** (`model/quota/entity.go`)
  - 添加 `CreateTime` 和 `UpdateTime` 字段
  - `UpdateTime` 字段 JSON tag 从 `updated_time` 改为 `update_time`

- **类型层级校验** (`model/quota/entity_manager.go`)
  - 添加父子层级校验，确保子实体类型层级高于父实体类型层级

#### Entity-Type 接口

- **创建接口** (`endpoints/openapi_v1/entity_type/create.go`)
  - 修复创建成功后返回 int 的问题，改为返回创建的 Entity-Type 对象
  - 添加 `type_name` 必填校验
  - 添加 `level` 参数校验（范围 1-5）

- **列表接口** (`endpoints/openapi_v1/entity_type/list.go`)
  - 修复返回格式错误，改为返回包含 `list` 和 `pagination` 的对象

- **详情接口** (`endpoints/openapi_v1/entity_type/one.go`)
  - 修复不存在的 Entity-Type 返回 200 的问题，改为返回 404 错误

- **删除接口** (`endpoints/openapi_v1/entity_type/delete.go`)
  - 修复不存在的 Entity-Type 返回 200 的问题，改为返回 404 错误

- **重置配额接口** (`endpoints/openapi_v1/api_key/reset_quota.go`)
  - 修复返回字段格式错误，改为返回包含 `id`、`previous_quota`、`new_quota`、`balance` 的统一格式

- **Entity 列表接口** (`endpoints/openapi_v1/entity/list.go`)
  - 修复列表只返回基础字段的问题，改为返回完整 Entity 字段（包含 allow_models、block_models、quota_plan、rate_limit_policy、create_time、update_time）

- **参数校验** (`endpoints/openapi_v1/api_key/checker.go`)
  - 添加 `description` 必填校验

- **Entity-Type 更新接口** (`endpoints/openapi_v1/entity_type/update.go`)
  - 修复 PATCH 返回 int 的问题，改为返回更新后的 Entity-Type 对象

- **Entity 创建接口** (`endpoints/openapi_v1/entity/create.go`)
  - 添加 `rate_limit_policy.enabled=true` 时必须设置 `rules.tpm` 或 `rules.rpm` 的校验

- **Entity 创建** (`model/quota/entity_manager.go`)
  - 修复 `QuotaPlan` 为 nil 时访问 `QuotaPlan.Quota` 导致的 nil pointer dereference
  - 添加重复名称检查，已存在同名 Entity 时返回 556 错误

- **Entity-Type 创建** (`model/quota/entity_type_manager.go`)
  - 修复重复 type_name 返回 422 的问题，改为返回 556 错误

- **HTTP 状态码处理** (`lib/xreq/result.go`)
  - 修复业务错误码（400-599）返回对应 HTTP 状态码导致测试客户端解析失败的问题，改为统一返回 HTTP 200，通过 `ErrNum` 字段传递业务错误码

- **新增冲突错误** (`lib/xerror/wrap.go`)
  - 添加 `WrapConflictErrorWithMsg` 函数，用于返回 409 冲突错误

- **错误码映射** (`lib/xerror/resolve.go`)
  - 添加 `Model.Conflict` 错误类型映射到 HTTP 409 状态码

### API 兼容性

- API-Key 接口返回字段名变更：`enable` → `enabled`
- API-Key 接口返回字段名变更：`updated_time` → `update_time`
- Entity 接口返回字段名变更：`updated_time` → `update_time`
- Entity 列表接口返回字段精简为基础字段

### Technical Improvements

- Entity 创建时自动初始化 `quota_plan` 和 `rate_limit_policy` 字段为默认值
- Entity 列表和详情接口返回 `create_time` 和 `update_time` 字段（Unix 时间戳）