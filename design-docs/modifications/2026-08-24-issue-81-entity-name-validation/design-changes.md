# Issue #81：增强 Entity name 合法性检查，与 Entity-Type 对齐

> 对应上游 Issue：[rainway-ai-gateway/ai-gateway-api#81](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/81)

## 1. 概述

### 1.1 问题现象

当前 `Entity` 的 `name` 字段合法性检查较为宽松，仅校验：

- 长度 1-64 字符；
- 不能包含控制字符；
- 不能包含前导/尾随空白字符；
- 全局唯一（由 endpoint / model 层查重保证）。

而 `Entity-Type` 的 `type_name` 字段校验规则要严格得多：

- 长度 1-32 字符；
- 仅允许小写字母、数字、`_`、`-`；
- 不能以 `-`、`_` 开头或结尾；
- 不能包含空白字符；
- 全局唯一。

由于 `Entity` 与 `Entity-Type` 在语义上强相关（`Entity.type` 必须引用已存在的 `Entity-Type`），`Entity.name` 的宽松规则会导致：

1. 可创建出包含大写字母、空格、中文、特殊符号的 Entity 名称；
2. 这些名称在日志、监控、路由标签、InnerAPI 导出等场景下可能产生解析或展示问题；
3. 与 `Entity-Type` 的命名风格不一致，增加用户理解成本。

### 1.2 变更目标

将 `Entity.name` 的合法性检查收紧到与 `Entity-Type` 同级的规则，仅放宽长度上限（64 字符），使两者在字符集、首尾约束、空白字符处理上保持一致。

### 1.3 变更范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api` |
| 涉及模块 | `lib/validate`、`endpoints/openapi_v1/entity` |
| 接口契约 | `/entities`（POST/PATCH/PUT）请求参数 `name` 的合法性条件变更 |
| 数据迁移 | 无；对历史数据只读，仅影响新建/更新时的校验 |

---

## 2. 现状代码分析

### 2.1 当前 `EntityName` 校验实现

`lib/validate/validate.go:257-269`：

```go
// EntityName validates an entity name.
func EntityName(s string) error {
	if len(s) == 0 || len(s) > 64 {
		return xerror.WrapParamErrorWithMsg("entity name length must be between 1 and 64")
	}
	if err := NoControlChars(s); err != nil {
		return xerror.WrapParamErrorWithMsg("entity name: %v", err)
	}
	if err := NoLeadingTrailingSpace(s); err != nil {
		return xerror.WrapParamErrorWithMsg("entity name: %v", err)
	}
	return nil
}
```

该函数已注册复用名 `validate.EntityName`，并在 `endpoints/openapi_v1/entity/validator.go` 的 `validateEntityParam` 中被调用：

```go
if param.Name != nil && *param.Name != "" {
    if err := validate.EntityName(*param.Name); err != nil {
        return err
    }
}
```

### 2.2 当前 `EntityTypeName` 校验实现

`lib/validate/validate.go:271-285`：

```go
// EntityTypeName validates an entity type name.
func EntityTypeName(s string) error {
	if err := validateName(s, 1, MaxEntityTypeNameLength, "type_name"); err != nil {
		return err
	}
	if err := validateNamePattern(s, entityTypeToken, "type_name"); err != nil {
		return err
	}
	first := s[0]
	last := s[len(s)-1]
	if first == '-' || first == '_' || last == '-' || last == '_' {
		return xerror.WrapParamErrorWithMsg("type_name cannot start or end with '-' or '_'")
	}
	return nil
}
```

其中 `entityTypeToken = regexp.MustCompile(`^[a-z0-9_-]+$`)`。

### 2.3 当前 API 文档描述

`design-docs/api-define/OpenAPI接口定义/entities.md` 中 `name` 字段描述为：

> 必填；长度 1-64 字符；不能包含控制字符；不能包含前导/尾随空白字符；全局唯一

`00-common.md` 中尚未定义 `EntityName` 公共类型，仅定义了 `EntityTypeName`。

---

## 3. 详细设计

### 3.1 核心原则

- **与 `EntityTypeName` 对齐**：`Entity.name` 采用同样的字符集与首尾约束，仅长度上限不同（64 vs 32）。
- **单点修改**：复用已有的 `validate.EntityName` 入口，集中修改 `lib/validate/validate.go`。
- **文档同步**：更新 `entities.md` 及新增 `00-common.md` 公共类型定义，保持接口契约与实现一致。
- **测试覆盖**：补充单元测试与集成测试用例，覆盖合法/非法边界。

### 3.2 新的 `EntityName` 校验规则

| 校验项 | 规则 |
|--------|------|
| 长度 | 1-64 字符 |
| 允许字符 | 小写字母 `a-z`、数字 `0-9`、下划线 `_`、连字符 `-` |
| 首尾限制 | 不能以 `-` 或 `_` 开头或结尾 |
| 空白字符 | 不允许任何空白字符（含空格、Tab、换行等） |
| 控制字符 | 不允许（已由字符集规则覆盖） |
| 全局唯一 | 由 endpoint/model 层查重保证，不在校验函数内 |

### 3.3 `lib/validate/validate.go` 修改

#### 新增常量

```go
const (
    // ... 现有常量 ...
    MaxEntityNameLength = 64
)
```

#### 新增/复用正则

已有 `entityTypeToken = regexp.MustCompile(`^[a-z0-9_-]+$`)`，可直接复用（`EntityName` 与 `EntityTypeName` 字符集相同）。

#### 修改 `EntityName` 函数

```go
// EntityName validates an entity name.
// Rules are aligned with EntityTypeName, except the length limit (64 vs 32).
func EntityName(s string) error {
	if err := validateName(s, 1, MaxEntityNameLength, "name"); err != nil {
		return err
	}
	if err := validateNamePattern(s, entityTypeToken, "name"); err != nil {
		return err
	}
	first := s[0]
	last := s[len(s)-1]
	if first == '-' || first == '_' || last == '-' || last == '_' {
		return xerror.WrapParamErrorWithMsg("name cannot start or end with '-' or '_'")
	}
	return nil
}
```

> 注意：保持错误码使用 `xerror.WrapParamErrorWithMsg`，返回 HTTP 422。

### 3.4 Endpoint 层无需改动

`endpoints/openapi_v1/entity/validator.go` 已经调用 `validate.EntityName`，因此规则收紧后自动生效，无需修改 endpoint 代码。

### 3.5 API 文档更新

#### `design-docs/api-define/OpenAPI接口定义/00-common.md`

在 `## 16. Entity-Type 名称（EntityTypeName）` 之后新增：

```markdown
## 17. Entity 名称（EntityName）

Entity 名称字符串，用于 `/entities` 相关接口。

| 字段 | 类型 | 必填 | 说明 | 合法性条件 |
|------|------|------|------|------------|
| - | string | - | Entity 名称 | 长度 1-64 字符；仅允许小写字母、数字、`_`、`-`；不能以 `-`、`_` 开头或结尾；全局唯一；不能包含空白字符 |
```

#### `design-docs/api-define/OpenAPI接口定义/entities.md`

将文档中所有 `name` 字段的合法性条件从：

> 必填；长度 1-64 字符；不能包含控制字符；不能包含前导/尾随空白字符；全局唯一

统一更新为：

> 必填；类型为 [EntityName](./00-common.md#17-entity-名称entityname)；全局唯一

涉及位置：

- 1 数据模型字段说明表；
- 2.1 创建 Entity 输入参数表及约束说明；
- 2.4 全量更新 Entity 约束说明；
- 2.5 部分更新 Entity 约束说明。

---

## 4. 涉及文件清单

| 文件 | 修改内容 |
|------|----------|
| `lib/validate/validate.go` | 收紧 `EntityName` 校验规则，与 `EntityTypeName` 对齐；新增 `MaxEntityNameLength` 常量。 |
| `lib/validate/validate_test.go` | 新增/调整 `TestEntityName` 用例，覆盖合法、大写、特殊字符、首尾 `-`、空白、超长等场景。 |
| `design-docs/api-define/OpenAPI接口定义/00-common.md` | 新增 `EntityName` 公共类型定义。 |
| `design-docs/api-define/OpenAPI接口定义/entities.md` | 更新 `name` 字段描述，引用 `EntityName` 公共类型。 |
| `test/integration/tests/entity/create/create_test.go` | 新增非法 name 格式集成测试用例（如大写、含空格、首尾 `-`）。 |
| `test/integration/tests/entity/partial_update/partial_update_test.go` | 新增非法 name 格式集成测试用例（如 PATCH 修改 name 为非法值）。 |
| `test/integration/tests/entity/full_update/full_update_test.go` | 新增非法 name 格式集成测试用例（如 PUT 修改 name 为非法值）。 |
| `test/integration/testutil/fixture.go` | 无需修改；现有 `UniqueEntityName` 已使用小写+数字+连字符，符合新规。 |

---

## 5. 测试计划

### 5.1 单元测试（`lib/validate`）

在 `validate_test.go` 中新增/替换 `TestEntityName`：

| 用例 | 输入 | 期望 |
|------|------|------|
| 合法-纯小写 | `"dep"` | 通过 |
| 合法-含数字 | `"dep_01"` | 通过 |
| 合法-含连字符 | `"ai-gateway"` | 通过 |
| 合法-最大长度 | 64 个字符的小写/数字/`-`/`_` 组合 | 通过 |
| 非法-空字符串 | `""` | 失败 |
| 非法-超长 | 65 个字符 | 失败 |
| 非法-大写字母 | `"Dep"` | 失败 |
| 非法-含空格 | `"dep 1"` | 失败 |
| 非法-含中文 | `"部门"` | 失败 |
| 非法-含特殊符号 | `"dep@1"` | 失败 |
| 非法-以 `-` 开头 | `"-dep"` | 失败 |
| 非法-以 `_` 开头 | `"_dep"` | 失败 |
| 非法-以 `-` 结尾 | `"dep-"` | 失败 |
| 非法-以 `_` 结尾 | `"dep_"` | 失败 |

### 5.2 集成测试（`test/integration/tests/entity`）

#### 创建接口（`create/create_test.go`）

新增以下用例：

| 用例编号 | 场景 | 请求体 | 期望 |
|----------|------|--------|------|
| E-1-013 | name 含大写字母 | `{"name":"BadName","type":typeName}` | 422 |
| E-1-014 | name 含空格 | `{"name":"bad name","type":typeName}` | 422 |
| E-1-015 | name 以 `-` 开头 | `{"name":"-badname","type":typeName}` | 422 |
| E-1-016 | name 以 `_` 结尾 | `{"name":"badname_","type":typeName}` | 422 |
| E-1-017 | name 长度为 64 | 64 字符合法字符串 | 200 |
| E-1-018 | name 长度为 65 | 65 字符合法字符串 | 422 |

#### 部分更新接口（`partial_update/partial_update_test.go`）

新增以下用例：

| 用例编号 | 场景 | 请求体 | 期望 |
|----------|------|--------|------|
| E-5-0XX | PATCH name 为含空格字符串 | `{"name":"bad name"}` | 422 |
| E-5-0XX | PATCH name 为以 `_` 开头 | `{"name":"_badname"}` | 422 |

> 具体用例编号在编写测试时按现有顺序递增。

#### 全量更新接口（`full_update/full_update_test.go`）

新增以下用例：

| 用例编号 | 场景 | 请求体 | 期望 |
|----------|------|--------|------|
| E-4-0XX | PUT name 为含大写字母 | `{"name":"BadName", ...}` | 422 |
| E-4-0XX | PUT name 以 `-` 结尾 | `{"name":"badname-", ...}` | 422 |

### 5.3 回归测试

- `go test ./lib/validate/...`
- `go test ./endpoints/openapi_v1/entity/...`
- `go test ./model/entity/...`
- 运行 `test/integration/tests/entity/create/`、`test/integration/tests/entity/partial_update/`、`test/integration/tests/entity/full_update/` 集成测试。

---

## 6. 风险与注意事项

| 风险 | 说明 | 缓解措施 |
|------|------|----------|
| 历史数据包含非法字符 | 升级前已创建的 Entity 名称可能包含大写、空格、中文等；更新这些 Entity 时若请求体再次传入旧 name 会触发 422 | 该变更仅影响写接口校验；历史数据保持只读。若需修改旧 Entity，管理员应重命名为符合新规的名称。 |
| 前端/外部系统展示受限 | 收紧后 Entity name 不再支持中文等本地化字符 | 与产品侧确认：Entity 名称用于系统标识，应使用机器友好格式；展示名称可由未来扩展字段（如 `display_name`）承担。 |
| 集成测试辅助函数兼容性 | `testutil.UniqueEntityName` 生成 `entity-{6位随机小写数字}`，已通过 | 无需调整。 |
| 与 `Entity-Type` 长度上限不一致 | `EntityName` 64 字符，`EntityTypeName` 32 字符 | 已在文档中明确说明，这是唯一差异点，其余规则一致。 |

---

## 7. 实施状态

### 7.1 已完成的修改

- [x] 修改 `lib/validate/validate.go` 中的 `EntityName` 实现，规则与 `EntityTypeName` 对齐；
- [x] 新增 `MaxEntityNameLength = 64` 常量；
- [x] 补充 `lib/validate/validate_test.go` 单元测试 `TestEntityName`；
- [x] 更新 `design-docs/api-define/OpenAPI接口定义/00-common.md`，新增 `EntityName` 公共类型；
- [x] 更新 `design-docs/api-define/OpenAPI接口定义/entities.md` 中所有 `name` 字段描述；
- [x] 补充 `test/integration/tests/entity/create/create_test.go` 非法 name 用例（E-1-013 ~ E-1-018）；
- [x] 补充 `test/integration/tests/entity/partial_update/partial_update_test.go` 非法 name 用例（E-5-004、E-5-005）；
- [x] 补充 `test/integration/tests/entity/full_update/full_update_test.go` 非法 name 用例（E-4-003、E-4-004）。

### 7.2 验证结果

| 测试命令 | 结果 |
|----------|------|
| `go test ./lib/validate/...` | PASS |
| `go test ./endpoints/openapi_v1/entity/...` | PASS |
| `go test ./model/entity/...` | PASS |
| `go test ./model/...` | PASS |
| `go test ./endpoints/... ./lib/... ./storage/...` | PASS |
| `go test ./tests/entity/create/... ./tests/entity/partial_update/... ./tests/entity/full_update/...`（在 `test/integration` 模块执行） | PASS |

> 集成测试前需要先重新编译 `ai-gateway-api.exe`（`go build -o ai-gateway-api.exe .`），因为测试框架会启动该二进制作为子进程。

### 7.3 上线前检查清单

- [x] 代码修改完成；
- [x] 单元测试通过；
- [x] 集成测试通过；
- [x] 设计文档与 API 文档已同步；
- [ ] 提交 PR 并关联 Issue #81。

---

## 8. 后续可优化（本期不做）

- 考虑在 `00-common.md` 中将 `EntityName` 与 `EntityTypeName` 的共性规则抽象为一个统一的“资源标识符”公共类型，减少未来类似资源的重复描述。
- 考虑在 Dashboard 前端同步增加同名校验，避免请求到达后端后再报错。
