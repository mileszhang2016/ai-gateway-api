# 现有接口清单

## 1. 接口总览

| 序号 | 接口路径 | 功能描述 | 特殊参数 |
|------|----------|----------|----------|
| 1 | `/configs/tls_conf/server_data_conf` | 导出 TLS/Server 配置 | `version` |
| 2 | `/configs/gslb_data/gslb` | 导出 GSLB 配置 | `version`、`bfe_cluster`(必填) |
| 3 | `/configs/gslb_data/cluster_table` | 导出集群表配置 | `version` |
| 4 | `/configs/protocol/server_cert_conf` | 导出证书配置 | `version` |
| 5 | `/configs/extra_files/{filename}` | 导出额外文件 | - |
| 6 | `/configs/mod-api-key` | 导出 API-Key 配置 | `version` |
| 7 | `/configs/mod-body-process` | 导出请求体处理配置 | `version` |
| 8 | `/configs/rate-limit-policy` | 导出限流策略配置 | `version` |
| 9 | `/configs/ai-route` | 导出 AI 路由配置 | `version` |

## 2. 特殊参数说明

- **`bfe_cluster`**：调用 `/configs/gslb_data/gslb` 时为必填参数，用于指定 BFE 集群名称，返回该集群对应的 GSLB 调度配置。

---

