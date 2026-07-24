# 部分更新API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 部分更新API-Key |
| 方法 | PATCH |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 部分更新指定 API-Key 的字段，仅传需修改的字段 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### Body 参数

同创建接口的 Body 参数，仅传需修改字段。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-5-001 | 仅修改 description | 正常参数 | 验证部分更新 description |
| AK-5-002 | 禁用 API-Key | 正常参数 | 验证 enabled=false |
| AK-5-003 | 启用 API-Key | 正常参数 | 验证 enabled=true |
| AK-5-004 | 更新不存在的 API-Key | 异常参数 | 验证返回 404 |

---

## 四、测试场景详细设计

---

### AK-5-001：仅修改 description

#### 设计思路

验证 PATCH 仅传 description 时，其他字段保持不变。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PATCH 请求，仅修改 description
3. 验证 description 更新，其他字段不变

#### 请求参数

```json
{
    "description": "patched-description"
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.description**："patched-description"  
**Data.id**：不变  
**Data.enabled**：不变

---

### AK-5-002：禁用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=false 禁用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=true 的 API-Key

#### 请求参数

```json
{
    "enabled": false
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.enabled**：false

---

### AK-5-003：启用 API-Key

#### 设计思路

验证 PATCH 设置 enabled=true 启用 API-Key。

#### 前提数据准备

- 先创建一个 enabled=false 的 API-Key

#### 请求参数

```json
{
    "enabled": true
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.enabled**：true

---

### AK-5-004：更新不存在的 API-Key

#### 设计思路

验证 PATCH 不存在的 ID 时返回 404。

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404