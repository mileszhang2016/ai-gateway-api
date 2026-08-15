# InnerAPI TlsConf 导出测试设计文档

## 1. 模块概述

`/inner-api/v1/configs/tls_conf/server_data_conf` 是 InnerAPI 的核心导出接口之一，负责将 AI 网关管理面的集群、路由、证书等配置聚合为 BFE 可识别的 `server_data_conf`。其中 `ClusterConf.Config.<cluster_name>.AIConf` 承载 AI 转发所需的模型映射、API-Key、Key 路由策略、模型定价表以及本次新增的 `MatchPrefix` / `StripPrefix` 前缀路由裁剪开关。

本文档覆盖 `AIConf.MatchPrefix` / `StripPrefix` 在 InnerAPI 导出结果中的正确性验证。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| IN-TLS-1 | 导出 Server Data Conf | GET | `/inner-api/v1/configs/tls_conf/server_data_conf` | 导出 TLS/Server/Cluster 等配置 |

## 3. 测试用例统计

| 场景 | 测试用例数 |
|------|-----------|
| AIConf 含 MatchPrefix/StripPrefix | 3（编号 IN-TLS-1-004 ~ IN-TLS-1-006，与现有 IN-1-001 ~ IN-1-003 区分） |
| **合计** | **3** |

## 4. 前置条件

- 测试环境配置 `SkipTokenValidate=true`。
- 通过 OpenAPI `/open-api/v1/clusters` 预先创建测试集群。
- 若验证 `ModelTable` 共存场景，需通过 `/open-api/v1/model-prices/import` 导入模型定价并设置 `llm_config.provider`。

## 5. 目录结构

```
innerapi/tls_conf/
├── design.md          # 本文档
└── tls_conf_test.go   # 现有 InnerAPI TlsConf 测试（含 AIConf 多 Key、ModelTable 等用例）
```

> 新增 `MatchPrefix` / `StripPrefix` 验证用例直接补充到 `tls_conf_test.go` 中，保持同一模块测试文件集中。

## 6. 导出 AIConf 中 MatchPrefix / StripPrefix 验证

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI - TlsConf |
| 接口名称 | 导出 Server Data Conf |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/tls_conf/server_data_conf` |

### 6.2 测试用例

#### IN-TLS-1-004 AIConf 包含 MatchPrefix / StripPrefix

- **前置**：通过 OpenAPI 创建集群，请求体中 `llm_config` 配置：
  ```json
  {
    "match_prefix": "openrouter/",
    "strip_prefix": true
  }
  ```
- **请求**：`GET /inner-api/v1/configs/tls_conf/server_data_conf`
- **预期**：
  - `ErrNum == 200`；
  - `ClusterConf.Config.<cluster>.AIConf.MatchPrefix == "openrouter/"`；
  - `ClusterConf.Config.<cluster>.AIConf.StripPrefix == true`。

#### IN-TLS-1-005 未配置前缀时 AIConf 为默认值

- **前置**：创建集群时不传 `match_prefix` / `strip_prefix`。
- **请求**：`GET /inner-api/v1/configs/tls_conf/server_data_conf`
- **预期**：
  - `AIConf.MatchPrefix` 不存在或为空字符串（BFE `AIConf.MatchPrefix` 使用 `omitempty`）；
  - `AIConf.StripPrefix == false`。

#### IN-TLS-1-006 仅 match_prefix、strip_prefix=false 时的导出

- **前置**：创建集群时配置 `match_prefix="openrouter/"`、`strip_prefix=false`。
- **请求**：`GET /inner-api/v1/configs/tls_conf/server_data_conf`
- **预期**：
  - `AIConf.MatchPrefix == "openrouter/"`；
  - `AIConf.StripPrefix == false`。

## 7. 数据清理

每个用例在 `t.Cleanup` 或 `defer` 中调用 `testutil.DeleteCluster(clusterName)`，确保测试集群被清理。
