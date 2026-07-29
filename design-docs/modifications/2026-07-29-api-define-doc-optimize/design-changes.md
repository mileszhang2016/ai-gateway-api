# api-define 接口定义文档拆分（2026-07-29）

## 1. 概述

### 1.1 变更背景

`design-docs/api-define/` 目录下的 `OpenAPI接口定义.md` 与 `InnerAPI接口定义.md` 是两份单体 Markdown 文件，分别包含所有 OpenAPI v1 与 InnerAPI v1 的接口定义。多人同时在这两份文件中不同模块下新增/修改接口时，极易出现合并冲突，且冲突解决成本高。

### 1.2 变更目标

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-29 |
| 涉及文件 | `design-docs/api-define/OpenAPI接口定义.md`（删除）<br>`design-docs/api-define/InnerAPI接口定义.md`（删除）<br>`design-docs/api-define/OpenAPI接口定义/*.md`（新增）<br>`design-docs/api-define/InnerAPI接口定义/*.md`（新增） |
| 变更类型 | 文档结构重组，无 API 语义变化 |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **按模块/主题拆分** | 每个业务模块或主题独立成文件，降低多人编辑冲突概率。 |
| **索引与内容分离** | `README.md` 仅作为目录索引，不放置具体接口定义。 |
| **最小侵入** | 仅调整文档文件组织方式，不修改接口字段、路径、返回值等定义。 |
| **链接可维护** | 修复拆分时产生的跨文件引用，保留文件内锚点。 |

---

## 2. OpenAPI 接口定义拆分

### 2.1 目录结构调整

原文件：

```text
design-docs/api-define/
├── OpenAPI接口定义.md      # 3547 行，包含所有模块
```

变更后：

```text
design-docs/api-define/
├── OpenAPI接口定义/
│   ├── README.md                  # 索引 + 全局约定入口
│   ├── 00-common.md               # 通用说明（URL、鉴权、返回值、Method、Query 参数）
│   ├── api-keys.md                # /api-keys
│   ├── entity-types.md            # /entity-types
│   ├── entities.md                # /entities
│   ├── global-route-rules.md      # /global-route-rules
│   ├── route-tables.md            # /route-tables
│   ├── alb-pool.md                # /alb-pool
│   ├── auth.md                    # /auth
│   ├── certificates.md            # /certificates
│   ├── clusters.md                # /clusters
│   ├── model-provider-types.md    # /model-provider-types
│   ├── tools.md                   # /tools
│   ├── expression-verify.md       # /expression/verify
│   ├── workflows.md               # 关键业务流程
│   └── object-relations.md        # 对象关系图
```

### 2.2 模块文件编号

每个模块文件保留原文件中的对应章节内容，章节主标题从 `## N. /xxx` 调整为 `# /xxx`（不再依赖全局序号）。小节标题在每个模块文件内独立重新编号：

- 一级小节：`## 1. xxx`、`## 2. xxx`
- 二级小节：`### 2.1 xxx`、`### 2.2 xxx`
- 三级小节：`#### 2.2.1 xxx`（如有）

例如原 `### 7.1 数据模型` / `### 7.2 接口清单` / `#### 7.2.1 xxx` 在 `alb-pool.md` 中重新编号为 `## 1. 数据模型` / `## 2. 接口清单` / `### 2.1 xxx`。

### 2.3 跨文件链接修复

拆分时发现一处跨模块引用：

- `api-keys.md` 中 `rules` / `targets` / `fallbacks` 元素结构原本指向 `#51-数据模型`（即原文件中的 `/global-route-rules` 章节）。
- 已修正为 `./global-route-rules.md#1-数据模型`（经独立重新编号后，`global-route-rules.md` 中的对应小节为 `## 1. 数据模型`）。

---

## 3. InnerAPI 接口定义拆分

### 3.1 目录结构调整

原文件：

```text
design-docs/api-define/
├── InnerAPI接口定义.md      # 1196 行，包含所有主题
```

变更后：

```text
design-docs/api-define/
├── InnerAPI接口定义/
│   ├── README.md                  # 索引 + 全局约定入口
│   ├── 00-overview.md             # 概述（文档目的、接口特点、与 OpenAPI 关系）
│   ├── 01-common.md               # 通用说明（URL、鉴权、返回值、版本控制）
│   ├── 02-interface-list.md       # 现有接口清单
│   ├── mod-api-key.md             # mod-api-key 接口
│   ├── mod-body-process.md        # mod-body-process 接口
│   ├── rate-limit-policy.md       # rate-limit-policy 接口
│   ├── ai-route.md                # ai-route 接口
│   ├── data-models.md             # 数据模型定义
│   └── appendix.md                # 附录
```

### 3.2 模块文件编号

每个主题文件保留原文件中的对应章节内容，章节主标题从 `## N、主题` 调整为 `# 主题`（不再依赖全局序号）。小节标题在每个文件内独立重新编号：

- 一级小节：`## 1. xxx`、`## 2. xxx`
- 二级小节：`### 2.1 xxx`、`### 2.2 xxx`
- 三级小节：`#### 2.2.1 xxx`（如有）

例如原 `### 4.1 接口信息` / `### 4.2 请求参数` / `#### 4.3.1 顶层结构` 在 `mod-api-key.md` 中重新编号为 `## 1. 接口信息` / `## 2. 请求参数` / `### 3.1 顶层结构`。

---

## 4. README.md 内容

`OpenAPI接口定义/README.md` 与 `InnerAPI接口定义/README.md` 均仅包含：

- 目录说明
- 全局约定入口链接
- 各模块/主题接口定义的索引表格

不放置任何具体接口的请求/响应示例、数据模型或字段说明。

---

## 5. 外部引用更新

以下文档中指向原单体文件的链接已同步更新：

- `design-docs/sys-design/总体设计文档.md`：OpenAPI/InnerAPI 链接均指向新 `README.md`
- `test/docs/README.md`：OpenAPI/InnerAPI 链接均指向新 `README.md`

---

## 6. 影响范围

| 维度 | 影响 |
|------|------|
| **OpenAPI/InnerAPI 接口语义** | 无变化 |
| **代码/数据库** | 无变化 |
| **其他设计文档** | 仅更新引用路径 |
| **文档协作方式** | 后续多人编辑时，冲突范围从整份 `OpenAPI接口定义.md` / `InnerAPI接口定义.md` 缩小到单个模块 `.md` 文件 |

---

## 7. 后续协作约定

1. **新增接口**统一追加到对应模块 `.md` 文件末尾，不要插入已有接口中间。
2. **不要直接修改 `README.md` 以外的公共文件**；若需调整全局约定，请单独提 PR。
3. 如需新增模块/主题，新建 `<module>.md` 并在对应 `README.md` 索引表中追加一行。
4. 跨模块引用使用相对文件链接，例如 `./global-route-rules.md#1-数据模型`。
