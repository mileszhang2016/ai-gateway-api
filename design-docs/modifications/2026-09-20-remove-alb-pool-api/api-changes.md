# /alb-pool 接口移除 —— API 契约变更

## 1. 变更概览

| 变更类型 | 端点 | 说明 |
|----------|------|------|
| 删除 | `GET /open-api/v1/alb-pool` | 获取默认 AI 网关实例池详情 |
| 删除 | `PATCH /open-api/v1/alb-pool` | 全量更新默认 AI 网关实例池实例列表 |

- 鉴权：原需 `FeatureBFEPool + ActionReadAll` / `FeatureBFEPool + ActionUpdate`；feature 本身保留（`/epp-pool`、`/epp-assignments` 继续使用）。
- 移除后请求这两个路径返回 404（路由不再注册）。

## 2. 配置项变更

| 变更类型 | 配置项 | 说明 |
|----------|--------|------|
| 删除 | `RunTime.DefaultAIInstancePoolName` | 原默认 `BFE.aipool`，仅被 `/alb-pool` 两个 handler 使用；涉及 `stateful/config.go`、`conf/ai_gateway_api.toml`、`test/integration/conf/ai_gateway_api.toml` |

`RunTime.DefaultAIClusterName`（默认 `BFE-AI_product.szyf`）**保留**，创建集群时仍用作默认调度矩阵键。

## 3. 数据与兼容性

- **数据库**：`pools` 表结构与种子行 `BFE.aipool` 保留不变（无迁移）；该行失去 OpenAPI 写入口，成为纯遗留数据。
- **数据面**：全部 InnerAPI 导出 topic（`cluster_table`、`gslb`、`server_data_conf`、`epp_data` 等）内容不变，BFE / conf-agent 无感知。
- **对外兼容性**：breaking change。已知外部调用方：
  - `ai-gateway-web/src/modules/AIInstancePool/index.vue`（本期保留，页面将失效，后续版本清理）；
  - 若有第三方脚本调用，需同步移除。

## 4. 测试变更

| 变更类型 | 位置 | 说明 |
|----------|------|------|
| 删除 | `test/integration/tests/alb_pool/` | GET 2 例 + PATCH 11 例（共 13 例）随端点一并移除 |
| 修改 | `test/docs/README.md` | 端点索引移除 ALB Pool 行；模块缩写移除 `BP`；统计表移除"BP 实例池"行并更新总计（接口 71→69，用例 226→218）；示例配置移除 `DefaultAIInstancePoolName` |
| 修改 | `test/integration/README.md` | 目录树移除 `alb_pool/` |
