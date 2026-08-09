# InnerAPI 测试用例设计文档

## 1. 模块概述

InnerAPI 模块为 BFE/Conf Agent 提供只读配置导出接口，支持基于 `version` 的增量同步。所有接口均为 GET 请求，返回值包含 `WorkMode` 字段。v0.3.0 对应 OpenAPI 的模块精简不影响 InnerAPI 导出结构（底层表结构未变）。

v0.0.7 起，Cluster 表导出（`/configs/gslb_data/cluster_table`）中的 `AIConf` 字段不再使用单 `Key`，而是输出 `Keys`（`[]{Name, Key, Weight}`）和 `KeyPolicy`（`{Strategy, MaxRetries, RetryBackoffInitial, RetryBackoffMax}`），以支持多 API-Key 路由。InnerAPI 应验证该导出结构与 OpenAPI 写入的多 Key 配置一致。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| IN-1 | 导出 TLS/Server 配置 | GET | `/inner-api/v1/configs/tls_conf/server_data_conf` | version 可选 |
| IN-2 | 导出 GSLB 配置 | GET | `/inner-api/v1/configs/gslb_data/gslb` | 必填 bfe_cluster |
| IN-3 | 导出集群表配置 | GET | `/inner-api/v1/configs/gslb_data/cluster_table` | version 可选 |
| IN-4 | 导出证书配置 | GET | `/inner-api/v1/configs/protocol/server_cert_conf` | version 可选 |
| IN-5 | 导出额外文件 | GET | `/inner-api/v1/configs/extra_files/{filename}` | - |
| IN-6 | 导出 API-Key 配置 | GET | `/inner-api/v1/configs/mod-api-key` | version 可选 |
| IN-7 | 导出请求体处理配置 | GET | `/inner-api/v1/configs/mod-body-process` | version 可选 |
| IN-8 | 导出限流策略配置 | GET | `/inner-api/v1/configs/rate-limit-policy` | version 可选 |
| IN-9 | 导出 AI 路由配置 | GET | `/inner-api/v1/configs/ai-route` | version 可选 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 导出 TLS/Server 配置 | 2 |
| 导出 GSLB 配置 | 2 |
| 导出集群表配置 | 2 |
| 导出证书配置 | 1 |
| 导出额外文件 | 1 |
| 导出 API-Key 配置 | 3 |
| 导出请求体处理配置 | 1 |
| 导出限流策略配置 | 1 |
| 导出 AI 路由配置 | 2 |
| **合计** | **14** |

## 4. 认证方式

InnerAPI 鉴权为 `McUserProbe`，测试环境需配置为可跳过或使用 Support Token。

## 5. 目录结构

```
innerapi/
├── design.md
├── tls_conf/
│   └── tls_conf_test.go
├── gslb/
│   └── gslb_test.go
├── cluster_table/
│   └── cluster_table_test.go
├── server_cert_conf/
│   └── server_cert_conf_test.go
├── extra_files/
│   └── extra_files_test.go
├── mod_api_key/
│   └── mod_api_key_test.go
├── mod_body_process/
│   └── mod_body_process_test.go
├── rate_limit_policy/
│   └── rate_limit_policy_test.go
└── ai_route/
    └── ai_route_test.go
```

## 6. 导出 TLS/Server 配置

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出 TLS/Server 配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/tls_conf/server_data_conf` |
| 说明 | 导出 TLS/Server 配置，支持 version 增量同步 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 配置数据，未变化时为 null |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

> 具体 TLS/Server 配置结构以 BFE 配置文件格式为准。

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-1-001 | 首次导出 TLS/Server 配置 | 正常参数 | 返回完整配置与 version |
| IN-1-002 | 导出 ClusterConf 含多 Key AIConf | 返回数据 | 验证 ClusterConf.AIConf.Keys/KeyPolicy 与 OpenAPI 写入一致 |

### 6.4 测试场景详细设计

#### 6.4.1 IN-1-001：首次导出 TLS/Server 配置（正常参数）

##### 设计思路

验证不传 version 时返回完整 TLS/Server 配置与 version。

##### 前提数据准备

已创建证书或存在默认 TLS 配置。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/tls_conf/server_data_conf`。
2. 验证返回包含 `Data`、`version` 和 `WorkMode`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空对象 | IsObject |
| version | 非空字符串 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

#### 6.4.2 IN-1-002：导出 ClusterConf 含多 Key AIConf（返回数据）

##### 设计思路

验证通过 OpenAPI 创建的 Cluster（含 `llm_config.keys`/`key_policy`）经 InnerAPI `/configs/tls_conf/server_data_conf` 导出后，`ClusterConf.Config.{cluster_name}.AIConf` 中正确输出 `Keys` 和 `KeyPolicy`。

##### 前提数据准备

1. 通过 POST `/open-api/v1/clusters` 创建集群 `cluster_inner_multi_keys`，其 `llm_config` 包含：
   - `models`: ["deepseek-chat"]
   - `keys`: 2 个 Key（name/key/weight）
   - `key_policy`: strategy="weighted_random", max_retries=3, retry_backoff_initial=100, retry_backoff_max=5000

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/tls_conf/server_data_conf`。
2. 在返回的 `Data.ClusterConf.Config` 中找到目标集群。
3. 校验该集群的 `AIConf.Keys` 和 `AIConf.KeyPolicy` 与写入一致。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.Keys | 长度为 2 | Len=2 |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.Keys[0].Name | "primary" | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.Keys[0].Key | "sk-aaaaaaaaaaaa" | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.Keys[0].Weight | 70 | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.KeyPolicy.Strategy | "weighted_random" | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.KeyPolicy.MaxRetries | 3 | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.KeyPolicy.RetryBackoffInitial | 100 | Equals |
| Data.ClusterConf.Config.cluster_inner_multi_keys.AIConf.KeyPolicy.RetryBackoffMax | 5000 | Equals |

---

## 7. 导出 GSLB 配置

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出 GSLB 配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/gslb_data/gslb` |
| 说明 | 导出 GSLB 配置，`bfe_cluster` 必填 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| version | string | N | 上次返回的版本号 | - |
| bfe_cluster | string | Y | BFE 集群名称 | 必填 |

#### 7.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | GSLB 配置数据 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-2-001 | 导出 GSLB 缺少 bfe_cluster | 必填校验 | 验证 ErrNum=422 或 404 |
| IN-2-002 | 正常导出 GSLB | 正常参数 | 返回 GSLB 配置与 version |

### 7.4 测试场景详细设计

#### 7.4.1 IN-2-001：导出 GSLB 缺少 bfe_cluster（必填校验）

##### 设计思路

验证 `bfe_cluster` 为必填参数。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/gslb_data/gslb`，不带 `bfe_cluster`。
2. 验证返回错误码。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：422 或 404  
**ErrMsg**：缺少 bfe_cluster 的错误信息  
**Data**：null

---

#### 7.4.2 IN-2-002：正常导出 GSLB（正常参数）

##### 设计思路

验证传入正确的 `bfe_cluster` 后返回 GSLB 配置与 version。

##### 前提数据准备

已创建 Cluster。

##### 执行步骤

1. 发送 GET 请求，`bfe_cluster=BFE-AI_product.szyf`。
2. 验证返回结构和 `version`。

##### 请求参数

```
bfe_cluster=BFE-AI_product.szyf
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空对象 | IsObject |
| version | 非空字符串 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

## 8. 导出集群表配置

### 8.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出集群表配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/gslb_data/cluster_table` |
| 说明 | 导出集群表配置 |

### 8.2 接口参数说明

#### 8.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 8.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 集群表配置数据 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 8.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-3-001 | 导出 cluster_table | 正常参数 | 返回集群表配置 |

### 8.4 测试场景详细设计

#### 8.4.1 IN-3-001：导出 cluster_table（正常参数）

##### 设计思路

验证首次导出集群表配置成功。

##### 前提数据准备

已创建 Cluster。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/gslb_data/cluster_table`。
2. 验证返回结构和 `version`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空对象 | IsObject |
| version | 非空字符串 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

## 9. 导出证书配置

### 9.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出证书配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/protocol/server_cert_conf` |
| 说明 | 导出证书配置 |

### 9.2 接口参数说明

#### 9.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 9.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 证书配置数据 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 9.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-4-001 | 导出证书配置 | 正常参数 | 返回证书配置 |

### 9.4 测试场景详细设计

#### 9.4.1 IN-4-001：导出证书配置（正常参数）

##### 设计思路

验证首次导出证书配置成功。

##### 前提数据准备

已创建证书。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/protocol/server_cert_conf`。
2. 验证返回结构。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空对象 | IsObject |
| version | 非空字符串 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

## 10. 导出额外文件

### 10.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出额外文件 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/extra_files/{filename}` |
| 说明 | 导出额外文件 |

### 10.2 接口参数说明

#### 10.2.1 请求参数

##### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| filename | string | Y | 文件名 |

#### 10.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object/string | 文件内容 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 10.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-5-001 | 导出额外文件 | 正常参数 | 返回文件内容 |

### 10.4 测试场景详细设计

#### 10.4.1 IN-5-001：导出额外文件（正常参数）

##### 设计思路

验证导出额外文件成功。

##### 前提数据准备

存在额外文件 `example.data`。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/extra_files/example.data`。
2. 验证返回文件内容。

##### 请求参数

URI：`example.data`

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 非空 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

## 11. 导出 API-Key 配置

### 11.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出 API-Key 配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/mod-api-key` |
| 说明 | 导出 API-Key 及配额配置 |

### 11.2 接口参数说明

#### 11.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 11.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 配置数据，未变化时为 null |
| Data.version | string | 配置版本号 |
| Data.config | object | 按产品线分组的 API-Key 路由规则 |
| Data.QuotaPlans | object | 按产品线分组的配额计划定义 |
| Data.tokens | object | 按产品线分组的 Token 配置 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 11.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-6-001 | 首次导出 mod-api-key | 正常参数 | 返回 config/QuotaPlans/tokens |
| IN-6-002 | 增量同步未变化 | 正常参数 | Data=null |
| IN-6-003 | 增量同步配置变化 | 正常参数 | 返回新配置与新 version |

### 11.4 测试场景详细设计

#### 11.4.1 IN-6-001：首次导出 mod-api-key（正常参数）

##### 设计思路

验证不传 version 时返回完整 API-Key 配置。

##### 前提数据准备

已创建 API-Key。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/mod-api-key`。
2. 验证返回包含 `version`、`config`、`QuotaPlans`、`tokens`。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| version | 非空字符串 | NotEmpty |
| config | 对象 | IsObject |
| QuotaPlans | 对象 | IsObject |
| tokens | 对象 | IsObject |
| WorkMode | 非空字符串 | NotEmpty |

---

#### 11.4.2 IN-6-002：增量同步未变化（正常参数）

##### 设计思路

验证传入上次 version 且配置未变化时，返回 `Data=null`。

##### 前提数据准备

已获取 version=XXX 且未变更配置。

##### 执行步骤

1. 发送 GET 请求，`version=XXX`。
2. 验证返回 `Data=null`。

##### 请求参数

```
version=XXX
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | null | IsNull |
| WorkMode | 非空字符串 | NotEmpty |

---

#### 11.4.3 IN-6-003：增量同步配置变化（正常参数）

##### 设计思路

验证配置变化后传入旧 version，返回新配置与新 version。

##### 前提数据准备

已获取 version=XXX，随后创建新 API-Key。

##### 执行步骤

1. 创建新 API-Key。
2. 发送 GET 请求，`version=XXX`。
3. 验证返回新配置与新 version。

##### 请求参数

```
version=XXX
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| version | 非空字符串，且不等于 XXX | NotEmpty / NotEquals(XXX) |
| tokens | 对象 | IsObject |
| WorkMode | 非空字符串 | NotEmpty |

---

## 12. 导出请求体处理配置

### 12.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出请求体处理配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/mod-body-process` |
| 说明 | 导出请求体处理模块配置 |

### 12.2 接口参数说明

#### 12.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 12.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 配置数据，未变化时为 null |
| Data.Version | string | 配置版本号 |
| Data.Config | object | 按产品线组织的请求体处理配置 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 12.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-7-001 | 导出 mod-body-process | 正常参数 | 返回 Version、Config |

### 12.4 测试场景详细设计

#### 12.4.1 IN-7-001：导出 mod-body-process（正常参数）

##### 设计思路

验证首次导出请求体处理配置成功。

##### 前提数据准备

无

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/mod-body-process`。
2. 验证返回结构和字段。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Version | 非空字符串 | NotEmpty |
| Config | 对象 | IsObject |
| WorkMode | 非空字符串 | NotEmpty |

---

## 13. 导出限流策略配置

### 13.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出限流策略配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/rate-limit-policy` |
| 说明 | 导出限流策略配置 |

### 13.2 接口参数说明

#### 13.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 13.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 配置数据，未变化时为 null |
| Data.Config | object | 按产品线组织的路由规则 |
| Data.RateLimitPolicies | object | 限流策略定义 |
| Data.ApikeyRateLimitPolicyBindings | object | API-Key 到策略 ID 列表的绑定关系 |
| Data.Version | string | 配置版本号 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 13.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-8-001 | 导出 rate-limit-policy | 正常参数 | 返回 Config/RateLimitPolicies/Bindings/Version |

### 13.4 测试场景详细设计

#### 13.4.1 IN-8-001：导出 rate-limit-policy（正常参数）

##### 设计思路

验证首次导出限流策略配置成功。

##### 前提数据准备

已创建含限流策略的 API-Key 或 Entity。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/rate-limit-policy`。
2. 验证返回结构和字段。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Config | 对象 | IsObject |
| RateLimitPolicies | 对象 | IsObject |
| ApikeyRateLimitPolicyBindings | 对象 | IsObject |
| Version | 非空字符串 | NotEmpty |
| WorkMode | 非空字符串 | NotEmpty |

---

## 14. 导出 AI 路由配置

### 14.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | InnerAPI |
| 接口名称 | 导出 AI 路由配置 |
| 方法 | GET |
| 路径 | `/inner-api/v1/configs/ai-route` |
| 说明 | 导出 AI 网关路由配置 |

### 14.2 接口参数说明

#### 14.2.1 请求参数

##### Query 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| version | string | N | 上次返回的版本号 |

#### 14.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| ErrNum | int | 返回码 |
| Data | object | 配置数据，未变化时为 null |
| Data.Version | string | 配置版本号 |
| Data.RouteRules | object | 所有路由表的集合 |
| Data.ApikeyRouteTableBindings | object | API-Key 到路由表查找顺序的映射 |
| ErrMsg | string | 文本消息 |
| WorkMode | string | 控制台工作模式 |

### 14.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| IN-9-001 | 导出 ai-route | 正常参数 | 返回 Version/RouteRules/Bindings |
| IN-9-002 | ai-route 未启用路由表不导出 | 业务规则 | RouteRules 为空 |

### 14.4 测试场景详细设计

#### 14.4.1 IN-9-001：导出 ai-route（正常参数）

##### 设计思路

验证导出 AI 路由配置成功，包含启用路由规则的路由表。

##### 前提数据准备

已创建含启用路由规则的 API-Key/Entity/Global Route。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/ai-route`。
2. 验证返回结构和字段。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Version | 非空字符串 | NotEmpty |
| RouteRules | 对象 | IsObject |
| ApikeyRouteTableBindings | 对象 | IsObject |
| WorkMode | 非空字符串 | NotEmpty |

---

#### 14.4.2 IN-9-002：ai-route 未启用路由表不导出（业务规则）

##### 设计思路

验证仅导出 `route_rules.enabled=true` 的路由表；全部禁用时 `RouteRules` 为空。

##### 前提数据准备

已创建但禁用所有路由规则（API-Key/Entity/Global Route 的 route_rules.enabled=false）。

##### 执行步骤

1. 发送 GET 请求到 `/inner-api/v1/configs/ai-route`。
2. 验证 `RouteRules` 为空对象。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| RouteRules | 空对象 | IsEmpty |
| ApikeyRouteTableBindings | 对象（可能为空） | IsObject |
| WorkMode | 非空字符串 | NotEmpty |

---

## 15. 依赖与数据准备

1. 需要预先通过 OpenAPI 创建 API-Key、Entity、Cluster、证书、Global Route 等数据，才能验证导出内容非空。
2. `/configs/gslb_data/gslb` 依赖正确的 `bfe_cluster` 参数，通常为 `BFE-AI_product.szyf`。
3. InnerAPI 鉴权为 `McUserProbe`，测试环境需配置为可跳过或使用 Support Token。

## 16. 注意事项

1. InnerAPI 返回值仍包含 `WorkMode`（与 OpenAPI v0.3.0 不同，InnerAPI 未移除该字段）。
2. 配置未变化时 `Data=null`，不要断言为空对象 `{}`。
3. 测试用例应尽量验证 `version` 变化：首次非空、未变化时一致且 Data 为 null、变化后更新。
4. 增量同步测试对时间敏感，创建新资源后应立即再次拉取。
