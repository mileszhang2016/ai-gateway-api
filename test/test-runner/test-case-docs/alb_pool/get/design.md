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