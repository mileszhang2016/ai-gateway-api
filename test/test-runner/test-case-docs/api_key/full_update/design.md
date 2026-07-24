# 全量更新API-Key - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | API-Key |
| 接口名称 | 全量更新API-Key |
| 方法 | PUT |
| 路径 | /open-api/v1/api-keys/{id} |
| 说明 | 全量替换指定 API-Key 的所有字段 |

---

## 二、接口参数说明

### URI 参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | API-Key 唯一标识 |

### Body 参数

同创建接口的 Body 参数。

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-4-001 | 更新 description | 正常参数 | 验证 description 更新成功 |
| AK-4-002 | 更新不存在的 API-Key | 异常参数 | 验证返回 404 |
| AK-4-003 | 缺少 description | 必填校验 | 验证 ErrNum=422 |
| AK-4-004 | 更新 expired_time | 正常参数 | 验证过期时间更新 |
| AK-4-005 | 更新 enabled 状态 | 正常参数 | 验证启用/禁用切换 |

---

## 四、测试场景详细设计

---

### AK-4-001：更新 description

#### 设计思路

验证 PUT 全量更新时 description 字段被正确更新。

#### 前提数据准备

- 先创建一个 API-Key

#### 执行步骤

1. 创建 API-Key，获取 ID
2. 发送 PUT 请求，修改 description
3. 验证返回的 description 已更新

#### 请求参数

```json
{
    "description": "updated-description"
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.description**："updated-description"

---

### AK-4-002：更新不存在的 API-Key

#### 设计思路

验证 PUT 不存在的 ID 时返回 404。

#### 请求参数

```json
{
    "description": "test"
}
```

#### 预期返回结果

**ErrNum**：404

---

### AK-4-003：缺少 description

#### 设计思路

验证 PUT 时 description 仍为必填。

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErrNum**：422

---

### AK-4-004：更新 expired_time

#### 设设计思路

验证 PUT 更新 expired_time 生效。

#### 请求参数

```json
{
    "description": "test-key",
    "expired_time": 1735689600
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.expired_time**：1735689600

---

### AK-4-005：更新 enabled 状态

#### 设计思路

验证 PUT 更新 enabled 状态。

#### 请求参数

```json
{
    "description": "test-key",
    "enabled": false
}
```

#### 预期返回结果

**ErrNum**：200  
**Data.enabled**：false