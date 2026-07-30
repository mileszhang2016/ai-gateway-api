# Tool 测试用例设计文档

## 1. 模块概述

Tool 模块提供从指定 AI 模型提供商代理拉取模型列表的工具接口，用于集群创建前预览可用模型。该接口由 v0.3.0 从 `/clusters` 的 `/models` 迁移为 `/tools/get-models-from-provider`。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| TOOL-1 | 从指定提供商获取模型列表 | POST | `/open-api/v1/tools/get-models-from-provider` | 代理请求提供商模型列表端点 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 从指定提供商获取模型列表 | 6 |
| **合计** | **6** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
tool/
├── design.md
└── get_models_from_provider/
    └── get_models_from_provider_test.go
```

## 6. 从指定提供商获取模型列表

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | Tool |
| 接口名称 | 从指定提供商获取模型列表 |
| 方法 | POST |
| 路径 | `/open-api/v1/tools/get-models-from-provider` |
| 说明 | 根据提供商信息代理拉取 AI 模型列表 |

### 6.2 接口参数说明

#### 6.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| schema | string | Y | 请求协议 | 仅允许 `http`/`https` |
| uri | string | N | 请求 URI，路径前面可以有 `/`，也可以无 `/` | - |
| hosts | []string | Y | 请求的 IP、Port 组合或域名 | 必填，数组长度 ≥1 |
| headers | map[string]string | N | 请求的 Header 参数列表 | - |
| provider_type | string | N | AI 模型提供商类型，如 deepseek、openai、qwen | - |

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | 模型 ID |
| name | string | 模型名称 |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| TOOL-1-001 | 正常获取模型列表 | 正常参数 | 返回模型数组 |
| TOOL-1-002 | uri 不带斜杠 | 正常参数 | 兼容两种 uri 写法 |
| TOOL-1-003 | 缺少 schema | 必填校验 | 验证 ErrNum=422 |
| TOOL-1-004 | 缺少 hosts | 必填校验 | 验证 ErrNum=422 |
| TOOL-1-005 | 非法 schema | 异常参数 | 验证 ErrNum=422 |
| TOOL-1-006 | 返回数据结构校验 | 返回数据 | 每个模型对象含 id、name |

### 6.4 测试场景详细设计

#### 6.4.1 TOOL-1-001：正常获取模型列表（正常参数）

##### 设计思路

验证传入完整参数时，接口成功代理返回模型列表。

##### 前提数据准备

存在可访问的提供商端点或本地 Mock 服务，返回固定模型列表 JSON。

##### 执行步骤

1. 发送 POST 请求到 `/open-api/v1/tools/get-models-from-provider`。
2. 验证响应状态码和返回结构。
3. 验证 `Data` 为模型数组。

##### 请求参数

```json
{
    "schema": "http",
    "uri": "/models",
    "hosts": ["127.0.0.1:8080"],
    "headers": {
        "Authorization": "Bearer sk-xxx"
    },
    "provider_type": "deepseek"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[0].id | 非空字符串 | IsString / NotEmpty |
| Data[0].name | 非空字符串 | IsString / NotEmpty |

---

#### 6.4.2 TOOL-1-002：uri 不带斜杠（正常参数）

##### 设计思路

验证 `uri` 以 `/` 开头与不以 `/` 开头均被兼容。

##### 前提数据准备

存在可访问的提供商端点或本地 Mock 服务。

##### 执行步骤

1. 发送 POST 请求，`uri` 为 `models`（不带 `/`）。
2. 验证响应与带 `/` 时行为一致。

##### 请求参数

```json
{
    "schema": "http",
    "uri": "models",
    "hosts": ["127.0.0.1:8080"],
    "provider_type": "deepseek"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[*].id | 非空字符串 | IsString / NotEmpty |
| Data[*].name | 非空字符串 | IsString / NotEmpty |

---

#### 6.4.3 TOOL-1-003：缺少 schema（必填校验）

##### 设计思路

验证 `schema` 为必填字段，缺少时接口应返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 构造请求 Body：缺少 `schema` 字段。
2. 发送 POST 请求。
3. 验证返回错误码。

##### 请求参数

```json
{
    "hosts": ["127.0.0.1:8080"],
    "provider_type": "deepseek"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "schema" 的错误信息  
**Data**：null

---

#### 6.4.4 TOOL-1-004：缺少 hosts（必填校验）

##### 设计思路

验证 `hosts` 为必填字段，缺少时接口应返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 构造请求 Body：缺少 `hosts` 字段。
2. 发送 POST 请求。
3. 验证返回错误码。

##### 请求参数

```json
{
    "schema": "http",
    "provider_type": "deepseek"
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "hosts" 的错误信息  
**Data**：null

---

#### 6.4.5 TOOL-1-005：非法 schema（异常参数）

##### 设计思路

验证 `schema` 仅允许 `http`/`https`，传入其他值时返回参数校验错误。

##### 前提数据准备

无

##### 执行步骤

1. 构造请求 Body：`schema` 为 `ftp`。
2. 发送 POST 请求。
3. 验证返回错误码。

##### 请求参数

```json
{
    "schema": "ftp",
    "hosts": ["127.0.0.1:8080"]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "schema" 或协议非法的错误信息  
**Data**：null

---

#### 6.4.6 TOOL-1-006：返回数据结构校验（返回数据）

##### 设计思路

验证返回的每个模型对象均包含 `id` 与 `name` 字段。

##### 前提数据准备

存在可访问的提供商端点或本地 Mock 服务。

##### 执行步骤

1. 发送 POST 请求。
2. 遍历 `Data` 数组，校验每个元素字段。

##### 请求参数

```json
{
    "schema": "http",
    "uri": "/models",
    "hosts": ["127.0.0.1:8080"],
    "provider_type": "deepseek"
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| Data | 数组 | IsArray |
| Data[*].id | 非空字符串 | IsString / NotEmpty |
| Data[*].name | 非空字符串 | IsString / NotEmpty |

---

## 7. 依赖与数据准备

1. 该接口依赖外部网络或本地 Mock 提供商服务。
2. 建议在 `testutil` 中提供简单的 HTTP Mock 服务，返回固定模型列表 JSON，以保证离线可运行。

## 8. 注意事项

1. 该接口替代了旧版 `POST /models`。
2. 由于涉及外部 HTTP 调用，测试用例需要考虑超时与 Mock。
3. `provider_type` 为可选，主要用于解析不同提供商的响应格式；测试时可优先使用 `deepseek`。
4. 测试环境 `SkipTokenValidate=true`，无需认证头。
