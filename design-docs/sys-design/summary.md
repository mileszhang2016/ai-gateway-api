# sys-design 文档索引

本目录存放 `ai-gateway-api` 系统级设计文档，覆盖总体架构、接口层、模型层、存储层、数据库设计以及关键模块的细节设计。

---

## 一、总体与分层设计

| 文档名称 | 相对路径 | 摘要说明 |
|---------|---------|---------|
| 总体设计文档 | [总体设计文档.md](./总体设计文档.md) | 描述 `ai-gateway-api` 在 AI 网关中的定位、功能范围、三层架构（接口层/模型层/存储层）、组件交互关系、核心数据流以及 OpenAPI 与 InnerAPI 的职责划分。 |
| 接口层设计文档 | [接口层设计文档.md](./接口层设计文档.md) | 描述管理面 OpenAPI（`/open-api/v1`）与数据面 InnerAPI（`/inner-api/v1`）的路由组织、统一的 `xreq.Endpoint` 抽象、全局与业务中间件链、子包划分以及典型接口实现模式。 |
| 模型层设计文档 | [模型层设计文档.md](./模型层设计文档.md) | 描述 `model/` 层的子包职责、Manager + Storager 接口的分层模式、Param/Filter 设计、事务管理、典型业务流程以及各业务模型（API-Key、Entity、Quota、RateLimit、Route 等）的交互方式。 |
| 存储层设计文档 | [存储层设计文档.md](./存储层设计文档.md) | 描述 `storage/rdb` 层的 DAO + Storage 两层结构、通用 DAO 设计模式、事务与连接管理、25 张表的 DAO 映射关系以及 Storage 实现如何向上暴露接口供模型层调用。 |
| 数据库设计文档 | [数据库设计文档.md](./数据库设计文档.md) | 描述 `ai-gateway-api` 当前实现中全部 25 张持久化表的字段、约束、索引、JSON 字段结构以及表间逻辑关系，覆盖基础配置、集群、路由、证书、API-Key、Entity、配额、限流等模块。 |

---

## 二、细节设计（details/）

| 文档名称 | 相对路径 | 摘要说明 |
|---------|---------|---------|
| 认证授权机制 | [details/认证授权机制.md](./details/认证授权机制.md) | 描述 `model/iauth` 的认证授权设计，包括用户与 Token 共用 `users` 表、`Visitor` 统一抽象、Scope 作用域、Feature-Action 权限模型、四种认证方式以及中间件集成与数据库表设计。 |
| InnerAPI 配置导出与版本控制 | [details/InnerAPI配置导出与版本控制.md](./details/InnerAPI配置导出与版本控制.md) | 描述面向 BFE/Conf Agent 的 InnerAPI 配置导出机制，包括 `VersionControlManager` 的 MD5 签名比对、版本号生成、`config_versions` 表持久化、9 类配置导出主题以及增量同步流程。 |
| API-Key 与 Entity 关联及模型继承 | [details/API-Key与Entity关联及模型继承.md](./details/API-Key与Entity关联及模型继承.md) | 描述 API-Key 与 Entity 的挂载关系、Entity 层级树约束、模型白名单交集与黑名单继承、配额计划层级合并、限流策略与路由规则的层级收集，以及导出到 BFE 时的最终生效规则。 |
| 限流策略与导出 | [details/限流策略与导出.md](./details/限流策略与导出.md) | 描述限流策略的数据模型（TPM/RPM/并发数）、JSON 配置结构、CRUD 校验、API-Key/Entity 引用关系、按 Entity 层级向上合并导出到 BFE 的流程，以及 BFE 侧预期行为与边界情况。 |
| 路由规则管理 | [details/路由规则管理.md](./details/路由规则管理.md) | 区分产品级 BFE 路由规则与 AI 路由规则，描述 `route_rules` 表的三级（Global/Entity/API-Key）管理、校验规则、与 API-Key/Entity 生命周期的一致性、导出到 BFE 的绑定顺序与文件格式。 |
| 配额余额同步机制 | [details/配额余额同步机制.md](./details/配额余额同步机制.md) | 描述 Redis 实时计数与数据库定时同步的混合架构，包括 `QuotaResetScheduler` 调度器、`BalanceSyncManager` 的余额同步与过期重置、自然周/月重置逻辑、Redis Key 生命周期以及 OpenAPI 查询余额的处理。 |

---

## 三、阅读建议

1. **快速建立全局认知**：先阅读《总体设计文档.md》。
2. **理解代码分层**：依次阅读《接口层设计文档.md》《模型层设计文档.md》《存储层设计文档.md》。
3. **查看具体表结构**：参考《数据库设计文档.md》。
4. **深入关键模块**：根据关注领域选择 `details/` 下的细节文档。
