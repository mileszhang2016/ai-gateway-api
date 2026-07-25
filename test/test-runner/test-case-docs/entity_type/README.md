# Entity-Type 模块测试用例设计文档

## 一、模块概述

Entity-Type 模块负责管理 Entity 的类型定义，包括类型的创建、查询、更新和删除操作。Entity-Type 定义了 Entity 的层级结构和属性约束，是创建 Entity 的前提条件。

## 二、接口列表

| 序号 | 接口名称 | 方法 | 路径 | 说明 | 用例数 |
|------|---------|------|------|------|--------|
| 1 | 创建Entity-Type | POST | /open-api/v1/entity-types | 创建新的实体类型定义 | 5 |
| 2 | 查询Entity-Type列表 | GET | /open-api/v1/entity-types | 获取实体类型列表 | 2 |
| 3 | 查询单个Entity-Type | GET | /open-api/v1/entity-types/{type_name} | 获取单个实体类型详情 | 2 |
| 4 | 更新Entity-Type | PATCH | /open-api/v1/entity-types/{type_name} | 更新实体类型描述 | 3 |
| 5 | 删除Entity-Type | DELETE | /open-api/v1/entity-types/{type_name} | 删除实体类型 | 3 |

## 三、测试用例统计

| 测试类型 | 用例数 | 说明 |
|---------|--------|------|
| 正常参数 | 5 | 正常场景下的接口调用 |
| 必填校验 | 2 | 验证必填字段缺失时的处理 |
| 边界值 | 2 | 参数边界值测试 |
| 异常参数 | 5 | 异常场景测试（如不存在的资源） |
| 业务规则 | 1 | 业务逻辑约束测试（删除时存在Entity） |
| **合计** | **15** | - |

## 四、认证方式

所有接口均使用 Token 认证，通过 `Authorization: Token TOKEN_STRING` 请求头传递。

## 五、目录结构

```
test-case-docs/entity_type/
├── README.md              # 模块概述和接口列表
├── create/
│   └── design.md          # 创建Entity-Type接口测试用例设计
├── list/
│   └── design.md          # 查询Entity-Type列表接口测试用例设计
├── detail/
│   └── design.md          # 查询单个Entity-Type接口测试用例设计
├── update/
│   └── design.md          # 更新Entity-Type接口测试用例设计
└── delete/
    └── design.md          # 删除Entity-Type接口测试用例设计
```

## 六、测试用例编号规则

测试用例编号格式：`ET-{接口序号}-{用例序号}`

示例：
- ET-1-001：创建Entity-Type接口的第1个测试用例
- ET-2-002：查询Entity-Type列表接口的第2个测试用例