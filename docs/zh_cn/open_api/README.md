# OpenAPI 接口文档

欢迎使用瑛菲AI网关的 OpenAPI 接口文档。本文档提供了所有对外暴露的 RESTful API 的详细说明。

> **优化说明：** 已删除以下独立接口文档，相关功能已集成到集群接口中或不再对外暴露：[bfe_cluster.md]、[domains.md]、[subclusters.md]、[product_pools.md]、[traffic.md]、[forward_rule.md]、[products.md]。

---

## 一、规范说明

API 规范、鉴权机制等通用说明请参考 [norms.md](norms.md)。

---

## 二、接口文档列表

### 2.1 全局管理

| 文档 | 说明 |
|------|------|
| [auth.md](auth.md) | 用户和鉴权机制说明 |

### 2.2 API-KEY和Entity管理

| 文档 | 说明 |
|------|------|
| [api_key.md](api_key.md) | API-Key 管理（创建、查询、更新、删除、配额计划、限流策略、路由规则） |
| [entity_type.md](entity_type.md) | Entity-Type 管理（创建、查询、更新、删除Entity类型定义） |
| [entity.md](entity.md) | Entity 管理（创建、查询、更新、删除实体，支持层级结构、配额配置和路由规则） |

### 2.3 实例池管理

| 文档 | 说明 |
|------|------|
| [bfe_pools.md](bfe_pools.md) | AI 网关实例池管理（获取详情、更新实例） |

### 2.4 集群管理

| 文档 | 说明 |
|------|------|
| [clusters.md](clusters.md) | 集群管理（一键创建、更新、删除、就绪检查、模型管理） |

### 2.5 路由规则

| 文档 | 说明 |
|------|------|
| [ai_route_rule.md](ai_route_rule.md) | AI 路由规则管理 |
| [global_route_rules.md](global_route_rules.md) | Global 路由表管理（全量更新、查询全局默认路由表） |
| [route_tables.md](route_tables.md) | 路由表列表查询（global/entity/api_key） |

### 2.6 证书

| 文档 | 说明 |
|------|------|
| [certificate.md](certificate.md) | 证书管理 |

---

## 三、快速开始

### 3.1 基础 URL

```
http://ai_gateway_api:port/open-api/v1/{endpoint}
```

示例：`http://127.0.0.1:8086/open-api/v1/api-keys`

### 3.2 认证方式

在 HTTP Authorization Header 中加入 Token：

```
Authorization: Token YOUR_TOKEN
```

### 3.3 返回格式

所有 API 返回统一格式：

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {}
}
```

详细规范请参考 [norms.md](norms.md)。