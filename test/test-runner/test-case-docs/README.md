# 测试用例设计文档

> 本目录存放各模块的测试用例设计文档（`design.md`）。  
> 对应的 Go 测试文件（`*_test.go`）位于 `test/test-runner/test-cases/` 目录。

## 目录约定

```
test-runner/
├── test-case-docs/              ← 设计文档（本目录）
│   ├── README.md                ← 本文件
│   ├── api_key/                 ← API-Key 模块设计文档
│   │   ├── README.md
│   │   ├── create/design.md
│   │   ├── list/design.md
│   │   ├── detail/design.md
│   │   ├── full_update/design.md
│   │   ├── partial_update/design.md
│   │   ├── delete/design.md
│   │   ├── quota_query/design.md
│   │   └── quota_reset/design.md
│   └── ai_route/                ← AI 路由规则模块设计文档
│       ├── README.md
│       ├── set_rules/design.md
│       └── get_rules/design.md
│
└── test-cases/                  ← Go 测试文件
    ├── api_key/
    │   ├── create/create_test.go
    │   ├── list/list_test.go
    │   ├── detail/detail_test.go
    │   ├── full_update/full_update_test.go
    │   ├── partial_update/partial_update_test.go
    │   ├── delete/delete_test.go
    │   ├── quota_query/quota_query_test.go
    │   └── quota_reset/quota_reset_test.go
    └── ai_route/
        ├── set_rules/set_rules_test.go
        └── get_rules/get_rules_test.go
```

## 模块总览

| 模块 | 设计文档 | 接口数 | 用例数 |
|------|---------|--------|--------|
| API-Key | [api_key/](./api_key/) | 8 | 43 |
| AI 路由规则 | [ai_route/](./ai_route/) | 2 | 24 |

## 新增测试用例规范

1. 在 `test-case-docs/<module>/<interface>/` 下创建 `design.md` 设计文档
2. 在 `test-cases/<module>/<interface>/` 下创建 `*_test.go` 测试文件
3. 更新 `test-case-docs/<module>/README.md` 中的用例列表
4. 更新本文件中的模块总览表

## 设计文档模板

每个 `design.md` 应包含以下章节：

```markdown
# <接口名称> - 测试用例设计

## 一、接口信息
（接口路径、方法、说明）

## 二、接口参数说明
（请求参数表、返回数据字段表、约束条件）

## 三、测试场景总览
（编号、场景名、测试类型、简要说明的汇总表）

## 四、测试场景详细设计
（每个场景的：设计思路、前提数据准备、执行步骤、请求参数、预期返回结果）
```