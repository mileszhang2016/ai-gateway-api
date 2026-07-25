# Entity 模块测试用例设计文档

## 模块概述

Entity 模块负责实体的管理，包括创建、查询、更新、删除实体，以及配额计划的查询和重置。

## 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| ENT-1 | 创建Entity | POST | `/open-api/v1/entities` | 创建新实体 |
| ENT-2 | 查询Entity列表 | GET | `/open-api/v1/entities` | 获取实体列表 |
| ENT-3 | 查询单个Entity | GET | `/open-api/v1/entities/{id}` | 查询指定实体详情 |
| ENT-4 | 全量更新Entity | PUT | `/open-api/v1/entities/{id}` | 全量更新实体 |
| ENT-5 | 部分更新Entity | PATCH | `/open-api/v1/entities/{id}` | 部分更新实体 |
| ENT-6 | 删除Entity | DELETE | `/open-api/v1/entities/{id}` | 删除指定实体 |
| ENT-7 | 查询配额计划 | GET | `/open-api/v1/entities/{id}/quota-plan` | 查询实体配额计划（含余额） |
| ENT-8 | 重置配额余额 | POST | `/open-api/v1/entities/{id}/quota-plan/reset` | 重置实体配额余额 |

## 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建Entity | 5 |
| 查询Entity列表 | 2 |
| 查询单个Entity | 2 |
| 全量更新Entity | 3 |
| 部分更新Entity | 3 |
| 删除Entity | 2 |
| 查询配额计划 | 2 |
| 重置配额余额 | 2 |
| **合计** | **21** |

## 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 目录结构

```
entity/
├── README.md
├── create/design.md
├── list/design.md
├── detail/design.md
├── full_update/design.md
├── partial_update/design.md
├── delete/design.md
├── quota_plan/design.md
└── quota_reset/design.md
```