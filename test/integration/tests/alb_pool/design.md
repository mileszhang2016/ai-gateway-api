# ALB Pool 模块测试用例设计文档

## 1. 模块概述

ALB Pool 模块负责 AI 网关实例池的管理，包括获取默认实例池详情和全量更新实例池。v0.3.0 起，实例结构中不再暴露 `tags` 字段。

## 2. 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| BP-1 | 获取默认实例池详情 | GET | `/open-api/v1/alb-pool` | 获取默认 AI 网关实例池的详情 |
| BP-2 | 更新默认实例池 | PATCH | `/open-api/v1/alb-pool` | 全量更新默认 AI 网关实例池的实例列表 |

## 3. 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 获取默认实例池详情 | 2 |
| 更新默认实例池 | 11 |
| **合计** | **13** |

## 4. 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 5. 目录结构

```
alb_pool/
├── design.md
├── get/
│   └── get_test.go
└── update/
    └── update_test.go
```

## 6. 获取默认实例池详情

### 6.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | ALB Pool |
| 接口名称 | 获取默认实例池详情 |
| 方法 | GET |
| 路径 | /open-api/v1/alb-pool |
| 说明 | 从配置文件读取 `DefaultAIInstancePoolName`，返回默认 AI 网关实例池详情（包含实例列表） |

### 6.2 接口参数说明

#### 6.2.1 请求参数

无

#### 6.2.2 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | 实例池完整名称 |
| instances | []Instance | 实例列表 |
| instances[].hostname | string | 实例所在主机名 |
| instances[].ip | string | 实例 IP 地址 |
| instances[].weight | int | 实例权重，范围 [0,100] |
| instances[].ports | map[string]int | 实例端口，至少包含 Default |

### 6.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| BP-1-001 | 获取实例池详情 | 正常参数 | 返回 name、instances 数组 |
| BP-1-002 | 验证返回字段完整性 | 返回数据 | 验证每个实例包含 hostname、ip、weight、ports，且不包含 tags |

### 6.4 测试场景详细设计

#### 6.4.1 BP-1-001：获取实例池详情（正常参数）

##### 设计思路

验证获取默认实例池详情接口的基本功能：确认接口返回实例池名称和实例列表。

##### 前提数据准备

- 测试配置文件 `ai_gateway_api.toml` 中已配置 `DefaultAIInstancePoolName`。
- 数据库初始化后，该默认实例池记录存在（由 DDL/种子数据保证）。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/alb-pool`。
2. 验证响应状态码和返回结构。
3. 验证 `name` 与配置项一致。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | 非空字符串，等于 `DefaultAIInstancePoolName` | NotEmpty / Equals |
| instances | 数组（可为空） | IsArray |

---

#### 6.4.2 BP-1-002：验证返回字段完整性（返回数据）

##### 设计思路

验证返回的实例列表中每个实例包含完整字段，且 v0.3.0 起不再包含 `tags` 字段。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在且包含至少 1 个实例。

##### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/alb-pool`。
2. 验证返回的每个实例包含 hostname、ip、weight、ports 字段。
3. 验证实例对象中不存在 `tags` 字段。

##### 请求参数

无

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| instances[].hostname | 非空字符串 | NotEmpty |
| instances[].ip | 非空字符串 | NotEmpty |
| instances[].weight | int 类型，范围 [0,100] | IsInt / RangeCheck |
| instances[].ports | map 类型，包含 Default | ContainsKey("Default") |
| instances[].tags | 不存在 | NotExists |

---

## 7. 更新默认实例池

### 7.1 接口信息

| 项目 | 值 |
|------|-----|
| 模块 | ALB Pool |
| 接口名称 | 更新默认实例池 |
| 方法 | PATCH |
| 路径 | /open-api/v1/alb-pool |
| 说明 | 全量更新默认 AI 网关实例池的实例列表 |

### 7.2 接口参数说明

#### 7.2.1 请求参数

##### Body 参数

| 参数名 | 类型 | 必填 | 说明 | 合法性条件 |
|--------|------|------|------|------------|
| instances | []Instance | Y | 实例列表，全量替换当前实例池中的实例 | 必填，≥1 个元素 |
| instances[].hostname | string | Y | 实例所在主机名，无 DNS 时可填写 IP 地址 | 符合 Hostname 类型 |
| instances[].ip | string | Y | 实例 IP 地址 | 符合 IP Address 类型 |
| instances[].weight | int | Y | 实例权重，范围 [0,100] | 0-100；0 时按 1 处理 |
| instances[].ports | map[string]int | Y | 实例端口，至少包含 Default 端口 | ≥1 个键值对；必须包含 `Default`；端口值唯一 |

#### 7.2.2 返回数据字段

同 GET 接口。

### 7.3 测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| BP-2-001 | 更新实例列表 | 正常参数 | 全量替换实例列表 |
| BP-2-002 | 更新后查询一致性 | 返回数据 | PATCH 后立即 GET，验证数据一致 |
| BP-2-003 | 更新为空列表 | 异常参数 | instances=[]，验证 ErrNum=422 |
| BP-2-004 | 缺少 instances | 必填校验 | 验证 ErrNum=422 |
| BP-2-005 | 缺少实例必填字段 | 必填校验 | 验证 ErrNum=422 |
| BP-2-006 | ports 不含 Default | 异常参数 | 验证 ErrNum=422 |
| BP-2-007 | 实例权重超出范围 | 边界值 | weight > 100 或 < 0 |
| BP-2-008 | 非法 hostname | 合法性条件 | 验证 ErrNum=422 |
| BP-2-009 | 非法 IP | 合法性条件 | 验证 ErrNum=422 |
| BP-2-010 | 重复端口值 | 合法性条件 | 验证 ErrNum=422 |
| BP-2-011 | weight 为 0 时按默认值 1 处理 | 正常参数 | 验证返回 weight=1 |

### 7.4 测试场景详细设计

#### 7.4.1 BP-2-001：更新实例列表（正常参数）

##### 设计思路

验证全量更新实例池接口的基本功能：传入完整的实例列表，确认接口返回成功并返回更新后的实例池详情。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/alb-pool`，传入实例列表。
2. 验证响应状态码和返回结构。
3. 验证返回的实例池与请求一致。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | 非空字符串，等于 `DefaultAIInstancePoolName` | NotEmpty / Equals |
| instances | 包含一个实例 | Len=1 |
| instances[0].hostname | "127.0.0.1" | Equals |
| instances[0].ip | "127.0.0.1" | Equals |
| instances[0].weight | 1 | Equals |
| instances[0].ports.Default | 8080 | Equals |
| instances[0].tags | 不存在 | NotExists |

---

#### 7.4.2 BP-2-002：更新后查询一致性（返回数据）

##### 设计思路

验证 PATCH 更新成功后，立即通过 GET 查询，返回的实例池数据与更新请求一致。

##### 前提数据准备

- 默认实例池存在。

##### 执行步骤

1. 发送 PATCH 请求更新实例池。
2. 发送 GET 请求查询实例池。
3. 对比两次返回的 `instances` 内容是否一致。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "host-a",
            "ip": "10.0.0.1",
            "weight": 50,
            "ports": {
                "Default": 8090
            }
        },
        {
            "hostname": "host-b",
            "ip": "10.0.0.2",
            "weight": 50,
            "ports": {
                "Default": 8091
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| instances | 长度为 2 | Len=2 |
| instances[*].hostname | 与请求一致 | Equals |
| instances[*].ip | 与请求一致 | Equals |
| instances[*].weight | 与请求一致 | Equals |
| instances[*].ports.Default | 与请求一致 | Equals |

---

#### 7.4.3 BP-2-003：更新为空列表（异常参数）

##### 设计思路

验证更新为空列表的场景，API 要求实例列表至少包含 1 个元素，传入空列表时接口应返回参数校验错误。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/alb-pool`，传入空实例列表。
2. 验证返回错误码。

##### 请求参数

```json
{
    "instances": []
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 instances 或至少 1 个元素的错误信息  
**Data**：null

---

#### 7.4.4 BP-2-004：缺少 instances（必填校验）

##### 设计思路

验证 `instances` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：缺少 instances 字段。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "instances" 的错误信息  
**Data**：null

---

#### 7.4.5 BP-2-005：缺少实例必填字段（必填校验）

##### 设计思路

验证实例对象的必填字段校验，当实例缺少 hostname、ip、weight、ports 中的任一必填字段时，接口应返回参数校验错误。本用例以缺少 `ip` 为例。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例缺少 ip 字段。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "ip" 的错误信息  
**Data**：null

---

#### 7.4.6 BP-2-006：ports 不含 Default（异常参数）

##### 设计思路

验证实例端口必须包含 `Default` 键，否则接口应返回参数校验错误。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例 ports 中不包含 `Default`。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Other": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Default" 或 "ports" 的错误信息  
**Data**：null

---

#### 7.4.7 BP-2-007：实例权重超出范围（边界值）

##### 设计思路

验证实例权重的范围校验，当 weight > 100 或 weight < 0 时，接口应返回参数校验错误。本用例以 weight=101 为例。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例权重为 101（超出范围）。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 101,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "weight" 的错误信息  
**Data**：null

---

#### 7.4.8 BP-2-008：非法 hostname（合法性条件）

##### 设计思路

验证实例 `hostname` 必须符合 Hostname 类型。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例 `hostname` 以 `-` 开头。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "-bad",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 hostname 非法的错误信息  
**Data**：null

---

#### 7.4.9 BP-2-009：非法 IP（合法性条件）

##### 设计思路

验证实例 `ip` 必须是合法 IPv4/IPv6 地址。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例 `ip` 为非法字符串。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "host-a",
            "ip": "not-an-ip",
            "weight": 1,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 ip 非法的错误信息  
**Data**：null

---

#### 7.4.10 BP-2-010：重复端口值（合法性条件）

##### 设计思路

验证同一实例内端口值不能重复。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例 `ports` 中两个键对应相同端口值。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回错误码。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "host-a",
            "ip": "10.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080,
                "Admin": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含重复端口值的错误信息  
**Data**：null

---

#### 7.4.11 BP-2-011：weight 为 0 时按默认值 1 处理（正常参数）

##### 设计思路

验证实例 `weight=0` 时，服务端按 1 处理。

##### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在。

##### 执行步骤

1. 构造请求 Body：实例 `weight=0`。
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`。
3. 验证返回 `weight=1`。

##### 请求参数

```json
{
    "instances": [
        {
            "hostname": "host-a",
            "ip": "10.0.0.1",
            "weight": 0,
            "ports": {
                "Default": 8080
            }
        }
    ]
}
```

##### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| instances[0].weight | 1 | Equals |

---

## 8. 依赖与数据准备

1. 测试配置文件 `ai_gateway_api.toml` 中需包含 `DefaultAIInstancePoolName`。
2. 数据库初始化后需存在默认实例池记录（通常由 DDL/初始化脚本保证）。
3. 更新默认实例池的测试可能会影响后续依赖 ALB Pool 的测试，建议在每个更新用例结束后恢复默认值，或采用独立测试进程隔离。

## 9. 注意事项

1. v0.3.0 实例结构已移除 `tags`，返回数据校验应断言该字段不存在。
2. 更新为全量替换，测试后应恢复默认实例或确保不影响其他模块。
3. 实例池名由配置决定，请求中无需也不应传入 `name`。
4. 测试环境 `SkipTokenValidate=true`，无需认证头。
