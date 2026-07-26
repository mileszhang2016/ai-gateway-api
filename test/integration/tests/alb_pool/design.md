# ALB Pool 模块测试用例设计文档

## 模块概述

ALB Pool 模块负责 AI 网关实例池的管理，包括获取默认实例池详情和全量更新实例池。

## 接口列表

| 编号 | 接口名称 | 方法 | 路径 | 说明 |
|------|----------|------|------|------|
| ALB-1 | 获取默认实例池详情 | GET | `/open-api/v1/alb-pool` | 获取默认 AI 网关实例池的详情 |
| ALB-2 | 更新默认实例池 | PATCH | `/open-api/v1/alb-pool` | 全量更新默认 AI 网关实例池的实例列表 |

## 测试用例统计

| 接口 | 测试用例数 |
|------|-----------|
| 获取默认实例池详情 | 2 |
| 更新默认实例池 | 5 |
| **合计** | **7** |

## 认证方式

测试环境配置 `SkipTokenValidate=true`，所有请求无需携带认证头。

## 目录结构

```

├── README.md
├── get/design.md
└── update/design.md
```

---

# 获取默认实例池详情 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | ALB Pool |
| 接口名称 | 获取默认实例池详情 |
| 方法 | GET |
| 路径 | /open-api/v1/alb-pool |
| 说明 | 获取默认 AI 网关实例池的详情（包含实例列表） |

---

## 二、接口参数说明

### 请求参数

无

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | 实例池完整名称 |
| instances | []Instance | 实例列表 |
| instances[].hostname | string | 实例所在主机名 |
| instances[].ip | string | 实例 IP 地址 |
| instances[].weight | int | 实例权重，范围 [0,100] |
| instances[].ports | map[string]int | 实例端口，至少包含 Default |
| instances[].tags | map[string]string | 实例标签 |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ALB-1-001 | 获取实例池详情 | 正常参数 | 返回 name、instances 数组 |
| ALB-1-002 | 验证返回字段完整性 | 返回数据 | 验证每个实例包含 hostname、ip、weight、ports |

---

## 四、测试场景详细设计

---

### ALB-1-001：获取实例池详情（正常参数）

#### 设计思路

验证获取默认实例池详情接口的基本功能：确认接口返回实例池名称和实例列表。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/alb-pool`
2. 验证响应状态码和返回结构

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | 非空字符串 | NotEmpty |
| instances | 数组（可为空） | IsArray |

---

### ALB-1-002：验证返回字段完整性（返回数据）

#### 设计思路

验证返回的实例列表中每个实例包含完整的字段信息。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在且包含实例

#### 执行步骤

1. 发送 GET 请求到 `/open-api/v1/alb-pool`
2. 验证返回的每个实例包含 hostname、ip、weight、ports 字段

#### 请求参数

无

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| instances[].hostname | 非空字符串 | NotEmpty |
| instances[].ip | 非空字符串 | NotEmpty |
| instances[].weight | int 类型，范围 [0,100] | IsInt |
| instances[].ports | map 类型，包含 Default | ContainsKey("Default") |

---

# 更新默认实例池 - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | ALB Pool |
| 接口名称 | 更新默认实例池 |
| 方法 | PATCH |
| 路径 | /open-api/v1/alb-pool |
| 说明 | 全量更新默认 AI 网关实例池的实例列表 |

---

## 二、接口参数说明

### 请求参数

#### Body 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| instances | []Instance | Y | 实例列表，全量替换当前实例池中的实例 |
| instances[].hostname | string | Y | 实例所在主机名，无 DNS 时可填写 IP 地址 |
| instances[].ip | string | Y | 实例 IP 地址 |
| instances[].weight | int | Y | 实例权重，范围 [0,100] |
| instances[].ports | map[string]int | Y | 实例端口，至少包含 Default 端口 |
| instances[].tags | map[string]string | N | 实例标签 |

### 返回数据字段

同 GET 接口。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| ALB-2-001 | 更新实例列表 | 正常参数 | 全量替换实例列表 |
| ALB-2-002 | 更新为空列表 | 异常参数 | instances=[]，验证 ErrNum=422 |
| ALB-2-003 | 缺少 instances | 必填校验 | 验证 ErrNum=422 |
| ALB-2-004 | 缺少实例必填字段 | 必填校验 | 验证 ErrNum=422 |
| ALB-2-005 | 实例权重超出范围 | 边界值 | weight > 100 或 < 0 |

---

## 四、测试场景详细设计

---

### ALB-2-001：更新实例列表（正常参数）

#### 设计思路

验证全量更新实例池接口的基本功能：传入完整的实例列表，确认接口返回成功并返回更新后的实例池详情。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/alb-pool`，传入实例列表
2. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "instances": [
        {
            "hostname": "127.0.0.1",
            "ip": "127.0.0.1",
            "weight": 1,
            "ports": {
                "Default": 8080
            },
            "tags": {
                "key": "value"
            }
        }
    ]
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| name | 实例池名称 | NotEmpty |
| instances | 包含一个实例 | Len=1 |
| instances[0].hostname | "127.0.0.1" | Equals |
| instances[0].ip | "127.0.0.1" | Equals |
| instances[0].weight | 1 | Equals |

---

### ALB-2-002：更新为空列表（异常参数）

#### 设计思路

验证更新为空列表的场景，API 要求实例列表至少包含 1 个元素，传入空列表时接口应返回参数校验错误。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 发送 PATCH 请求到 `/open-api/v1/alb-pool`，传入空实例列表
2. 验证返回错误码

#### 请求参数

```json
{
    "instances": []
}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "Instances must contain at least 1 item" 的错误信息  
**Data**：null

---

### ALB-2-003：缺少 instances（必填校验）

#### 设计思路

验证 `instances` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 构造请求 Body：缺少 instances 字段
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`
3. 验证返回错误码

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "instances" 的错误信息  
**Data**：null

---

### ALB-2-004：缺少实例必填字段（必填校验）

#### 设计思路

验证实例对象的必填字段校验，当实例缺少 hostname、ip、weight、ports 中的任一必填字段时，接口应返回参数校验错误。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 构造请求 Body：实例缺少 ip 字段
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`
3. 验证返回错误码

#### 请求参数

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

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "ip" 的错误信息  
**Data**：null

---

### ALB-2-005：实例权重超出范围（边界值）

#### 设计思路

验证实例权重的范围校验，当 weight > 100 或 weight < 0 时，接口应返回参数校验错误。

#### 前提数据准备

- 确保配置文件中 `DefaultAIInstancePoolName` 对应的实例池存在

#### 执行步骤

1. 构造请求 Body：实例权重为 101（超出范围）
2. 发送 PATCH 请求到 `/open-api/v1/alb-pool`
3. 验证返回错误码

#### 请求参数

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

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "weight" 的错误信息  
**Data**：null

---

