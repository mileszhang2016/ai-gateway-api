# Auth 模块测试用例设计文档

## 模块概述

Auth 模块负责用户认证与授权管理，包括：
- 用户管理（创建、删除、重置密码、列表、设置管理员、绑定产品线）
- Session Key 管理（创建、删除）
- Token 管理（创建、删除、详情、列表、按产品线查询）

## 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| AUTH-1 | 创建用户 | POST | `/open-api/v1/auth/users` | 创建新用户 |
| AUTH-2 | 删除用户 | DELETE | `/open-api/v1/auth/users/{user_name}` | 删除指定用户 |
| AUTH-3 | 重置密码 | PATCH | `/open-api/v1/auth/users/{user_name}/passwd` | 修改用户密码 |
| AUTH-4 | 用户列表 | GET | `/open-api/v1/auth/users` | 获取所有用户列表 |
| AUTH-5 | 设置管理员 | PATCH | `/open-api/v1/auth/users/{user_name}/is_admin` | 设置用户管理员权限 |
| AUTH-6 | 绑定产品线 | POST | `/open-api/v1/auth/users/{user_name}/products/{product_name}` | 为用户绑定产品线 |
| AUTH-7 | 解除产品线绑定 | DELETE | `/open-api/v1/auth/users/{user_name}/products/{product_name}` | 解除用户产品线绑定 |
| AUTH-8 | 按产品线查用户 | GET | `/open-api/v1/auth/users/actions/search-by-product/{product_name}` | 查询指定产品线的授权用户 |
| AUTH-9 | 创建Session Key | POST | `/open-api/v1/auth/session-keys` | 用户名密码登录获取session key |
| AUTH-10 | 删除Session Key | DELETE | `/open-api/v1/auth/session-keys/{session_key}` | 删除session key |
| AUTH-11 | 创建Token | POST | `/open-api/v1/auth/tokens` | 创建Token并绑定产品线 |
| AUTH-12 | 删除Token | DELETE | `/open-api/v1/auth/tokens/{token_name}` | 删除指定Token |
| AUTH-13 | Token详情 | GET | `/open-api/v1/auth/tokens/{token_name}` | 查询Token详情 |
| AUTH-14 | Token列表 | GET | `/open-api/v1/auth/tokens` | 获取所有Token列表 |
| AUTH-15 | 按产品线查Token | GET | `/open-api/v1/auth/tokens/actions/search-by-product/{product_name}` | 查询指定产品线的Token |

## 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 创建用户 | 7 |
| 删除用户 | 2 |
| 重置密码 | 5 |
| 用户列表 | 2 |
| 设置管理员 | 4 |
| 绑定产品线 | 3 |
| 解除产品线绑定 | 2 |
| 按产品线查用户 | 2 |
| 创建Session Key | 4 |
| 删除Session Key | 2 |
| 创建Token | 5 |
| 删除Token | 2 |
| Token详情 | 2 |
| Token列表 | 2 |
| 按产品线查Token | 2 |
| **合计** | **45** |

## 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 目录结构

```
auth/
├── README.md
├── create_user/design.md
├── delete_user/design.md
├── reset_password/design.md
├── list_users/design.md
├── set_admin/design.md
├── bind_product/design.md
├── unbind_product/design.md
├── search_by_product/design.md
├── create_session_key/design.md
├── delete_session_key/design.md
├── create_token/design.md
├── delete_token/design.md
├── token_detail/design.md
├── token_list/design.md
└── search_token_by_product/design.md
```