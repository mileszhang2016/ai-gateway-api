---
name: ai-gateway-api-rd-workflow
description: 引导用户在 ai-gateway-api 代码库中完成一次完整的功能研发流程，包括需求对齐、设计文档修改、代码实现、单元测试、集成测试、schema 校验与回归验证。
---

# AI Gateway API 代码库研发流程

本 Skill 适用于在 `ai-gateway-api/` 目录中新增或修改控制面功能（例如 AI Gateway 的 cluster 配置、API-Key 路由、限流、配额、模型定价、会话级 Key 亲和性配置导出等）的完整研发流程。

## 触发语

当用户提出以下类型请求时启用本流程：

- “我要实现 xxx 功能”
- “请帮我完成 xxx 的代码与测试”
- “请在 AI Gateway API 中增加/修改 xxx”
- “请修改 ai-gateway-api 的 xxx”

## 执行原则

1. **分阶段暂停确认**：本流程将研发拆分为多个 Phase。每个 Phase 执行完毕后，必须暂停并等待用户确认“可以继续”后，再进入下一个 Phase。不要在未获得用户确认的情况下自动推进到下一阶段。
2. **Git 提交前确认**：在任何 `git commit`、`git push` 或其他会改变 Git 仓库状态的操作之前，必须先向用户说明变更内容并取得明确同意。禁止自动执行 Git 提交或推送。
3. **Git push 默认目标为 origin**：如果获得用户授权执行 `git push`，默认推送到 `origin` 远程仓库，而不是 `upstream`。除非用户明确指定其他 remote，否则不使用 `upstream`。

## 研发阶段

### Phase 0. 需求澄清与范围界定

1. 让用户明确：
   - 功能目标（一句话描述）
   - 验收标准（必须通过的测试/行为）
   - 影响范围（是否改 OpenAPI、是否改 InnerAPI 导出、是否改数据库表、是否影响 BFE 数据面消费格式）
2. 判断是否为“非平凡改动”（多文件、有架构选择、用户偏好影响实现）：
   - 是 → 调用 `EnterPlanMode`，先出设计文档/plan 再执行。
   - 否 → 直接进入 Phase 1。

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 1. 修改 modifications 文档

`ai-gateway-api` 侧的非平凡改动必须在 `design-docs/modifications/` 留下修改说明，便于后续维护与审计。

1. 在 `design-docs/modifications/` 下新建日期化目录，例如：
   ```
   design-docs/modifications/2026-08-26-ai-key-session-affinity/
   ```
2. 在该目录下创建：
   - `change-summary.md`：背景与目标、主要改动点、数据面影响、兼容性说明、待实现清单
   - `api-changes.md`：OpenAPI/InnerAPI 字段变更、请求/响应示例、默认值、校验规则
   - `design-changes.md`：数据模型变更、核心代码改动点、配置映射关系
3. 如已有同主题目录，则更新其中的文档，不要重复创建。

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 2. 更新 api-define 文档

如果改动涉及 OpenAPI 或 InnerAPI 接口字段，必须同步更新 `design-docs/api-define/`：

1. OpenAPI 接口定义：`design-docs/api-define/OpenAPI接口定义/`
   - 新增/修改请求体字段、响应字段、参数表、示例
2. InnerAPI 接口定义：`design-docs/api-define/InnerAPI接口定义/`
   - 新增/修改导出结构字段、示例、字段映射说明
3. 更新原则：
   - 新增字段要说明类型、默认值、是否必填、合法性条件
   - 修改字段要说明前后行为差异
   - OpenAPI 与 InnerAPI 的字段映射关系要清晰

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 3. 更新 sys-design 文档

如果改动涉及系统设计、数据模型或配置导出，需要更新 `design-docs/sys-design/`：

1. 查找现有相关文档（例如 `模型层设计文档.md`、`数据库设计文档.md`、`接口层设计文档.md`）。
2. 在文档中新增/修改对应章节，说明：
   - 数据模型变更（`model/icluster_conf/` 等结构体）
   - 数据库存储变更（`db_ddl.sql`、`db_ddl_sqlite.sql`）
   - 接口层变更（请求参数、校验规则）
   - InnerAPI 导出映射（OpenAPI 字段如何转换为 BFE 配置字段）
   - 边界情况与默认值
3. 如需新增独立文档，直接创建 `.md` 文件，并在相关文档中建立链接。

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 4. 代码实现

1. 阅读相关源码，确认修改范围：
   - API 层：`endpoints/openapi_v1/<domain>/`、`endpoints/innerapi_v1/<domain>/`
   - 模型层：`model/<domain>/`
   - 存储层：`storage/rdb/<domain>/`
   - 校验层：`lib/validate/`
   - 公共工具：`lib/`
2. 做最小改动，优先匹配现有代码风格：
   - 不引入未使用的依赖。
   - 不修改与本次需求无关的逻辑。
3. 关键实现完成后，先跑相关单元测试：
   ```bash
   cd ai-gateway-api
   go test ./model/<相关模块>/... ./lib/validate/... ./endpoints/<相关模块>/...
   ```

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 5. 补充单元测试

1. 如果修改了模型层，在对应 `_test.go` 中补充单元测试，保持 `model/` 语句覆盖率 ≥ 70%。
2. 如果修改了校验逻辑，在 `lib/validate/validate_test.go` 中补充合法/非法用例。
3. 如果修改了 endpoints，在对应 `*_test.go` 中补充参数校验、归一化、响应字段断言。
4. 运行单元测试：
   ```bash
   cd ai-gateway-api
   go test ./model/... ./lib/... ./endpoints/...
   ```

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 6. 补充集成测试设计文档

集成测试设计文档位于 `test/integration/tests/<module>/design.md`。在写测试代码之前，先补充对应场景的设计文档。

1. 在 `test/integration/tests/<module>/design.md` 中：
   - 模块概述补充新字段/新行为说明
   - 接口参数表补充新增字段
   - 测试场景总览表新增用例编号与简要说明
   - 详细设计章节新增每个测试例（设计思路、前置数据、执行步骤、请求参数、预期结果）
2. 如涉及 InnerAPI 导出，同步更新 `test/integration/tests/innerapi/design.md` 与 `test/integration/tests/innerapi/<submodule>/design.md`。

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 7. 补充集成测试代码

集成测试代码位于 `test/integration/tests/<module>/<interface>/`。

1. 根据设计文档在对应 `_test.go` 中实现用例。
2. 如涉及 OpenAPI 响应字段，在 `test/integration/tests/schema/openapi/<module>.go` 中补充 schema 定义。
3. 如涉及 InnerAPI 导出字段，在 `test/integration/tests/schema/innerapi/schema.go` 与 `innerapi_schema_test.go` 中补充 schema 与断言。
4. 运行相关集成测试：
   ```bash
   cd ai-gateway-api/test/integration
   go test -v -count=1 -timeout 300s ./tests/<module>/...
   ```

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 8. 补充/更新 schema 测试

1. OpenAPI schema：`test/integration/tests/schema/openapi/`
   - 新增/修改对象 schema，确保字段类型、必填/可选关系正确
2. InnerAPI schema：`test/integration/tests/schema/innerapi/`
   - 新增/修改 schema 定义
   - 在测试中创建覆盖新字段的集群/配置，并断言导出字段
3. 运行 schema 测试：
   ```bash
   cd ai-gateway-api/test/integration
   go test -v -count=1 -timeout 300s ./tests/schema/...
   ```

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 9. 回归验证

1. 运行本次新增/修改模块的集成测试。
2. 运行与被改动模块相关的既有集成测试：
   - cluster 相关：`./tests/clusters/...`、`./tests/innerapi/tls_conf/...`
   - api-key 相关：`./tests/api_key/...`、`./tests/innerapi/mod_api_key/...`
   - rate-limit 相关：`./tests/innerapi/rate_limit_policy/...`
3. 运行相关单元测试：
   ```bash
   cd ai-gateway-api
   make test-model
   ```
4. 如有失败，优先修复；修复后再次回归，直到全部通过。

完成本阶段后暂停，等待用户确认后再进入下一阶段。

### Phase 10. 收尾与总结

1. 检查是否有注释/文档描述的是旧行为，及时同步更新。
2. 向用户汇报：
   - 改动了哪些文件
   - 新增/修改了哪些测试
   - 验证结果
   - 仍存在的风险或待决策点（如有）

3. **Git 提交前必须人工确认**：如果用户要求或流程需要执行 `git commit`、`git push` 等 Git 操作，必须先向用户清晰说明本次提交内容（包含文件清单与主要变更摘要），并取得明确同意后再执行。禁止在未获授权的情况下自动提交或推送。获得授权后，`git push` 默认推送到 `origin`，不要推送到 `upstream`，除非用户明确指定。

## 常见陷阱

- **provider 与 cluster 字段混淆**：`llm_config.provider` 引用 Provider 名称，模型必须在 Provider 的 `models` 列表内，Key 必须通过 `llm_config.keys[].name` 引用 Provider 中的 Key。
- **OpenAPI 与 InnerAPI 字段命名不一致**：OpenAPI 使用 snake_case（如 `key_affinity`、`redis_prefix`），InnerAPI/BFE 配置常使用 PascalCase（如 `SessionAffinityRedisPrefix`），注意映射关系。
- **默认值未同步**：修改默认值时，需同时更新模型层默认值、校验层默认值、设计文档、api-define 文档、集成测试断言。
- **Schema 测试遗漏**：新增 OpenAPI/InnerAPI 字段后，记得同步更新 `test/integration/tests/schema/openapi/` 与 `test/integration/tests/schema/innerapi/`。
- **Integration 测试需要重新编译二进制**：运行集成测试前确保已执行 `go build -o ai-gateway-api.exe .`，否则测试会启动旧版本服务。
- **SQLite 日志文件占用**：Windows 上并发运行多个 integration 测试包时，可能出现日志文件被占用导致启动失败，可单独重新运行失败包。

## 推荐命令速查

```bash
# 构建二进制
cd ai-gateway-api
go build -o ai-gateway-api.exe .

# 模块单元测试
cd ai-gateway-api
make test-model

# 相关单元测试
go test ./model/icluster_conf/... ./lib/validate/... ./endpoints/openapi_v1/product_cluster/...

# 集成测试（指定模块）
cd ai-gateway-api/test/integration
go test -v -count=1 -timeout 300s ./tests/clusters/...

# 集成测试（InnerAPI 导出）
go test -v -count=1 -timeout 300s ./tests/innerapi/tls_conf/...

# Schema 测试
go test -v -count=1 -timeout 300s ./tests/schema/...
```
