# /certificates

## 1. 数据模型

```json
{
  "cert_name": "cert_demo",
  "description": "abc",
  "is_default": true,
  "cert_file_name": "demo_cert_file_name",
  "cert_file_content": "-----BEGIN ...-----END CERTIFICATE-----",
  "key_file_name": "demo_key_file_name",
  "key_file_content": "-----BEGIN RSA PRIVATE KEY-----...-----END RSA PRIVATE KEY-----",
  "expired_date": "2021-08-23 16:02:31"
}
```

**字段说明**

| 字段 | 类型 | 说明 | 可能取值 |
|------|------|------|----------|
| `cert_name` | string | 证书名 | 必须唯一 |
| `description` | string | 证书描述 | - |
| `is_default` | bool | 是否是默认证书 | 全局必须有且只有一个默认证书 |
| `cert_file_name` | string | 主证书文件名 | - |
| `cert_file_content` | string | 主证书文件内容 | 创建/更新时必填；返回时不包含 |
| `key_file_name` | string | 主证书密钥名 | - |
| `key_file_content` | string | 主证书密钥文件内容 | 创建/更新时必填；返回时不包含 |
| `expired_date` | string | 主证书过期时间 | 如 `2021-08-23 16:02:31` |

**约束**

- 全局必须有且只有一个默认证书。
- 默认证书不能被删除。
- 更新为默认证书时，旧的默认证书自动变为非默认证书。

## 2. 接口清单

### 2.1 创建证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 创建证书 | - |
| 端点 | /certificates | - |
| 版本 | v1 | - |
| method | POST | - |

**输入参数（Body）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名 | Y | 必须唯一 |
| description | string | 证书描述 | Y | - |
| is_default | bool | 是否是默认证书 | Y | 全局必须有且只有一个默认证书 |
| cert_file_name | string | 主证书文件名 | Y | - |
| cert_file_content | string | 主证书文件内容 | Y | - |
| key_file_name | string | 主证书密钥名 | Y | - |
| key_file_content | string | 主证书密钥文件内容 | Y | - |
| expired_date | string | 主证书过期时间 | Y | - |

**约束**

- 全局必须有且只有一个默认证书。
- 若系统中已存在默认证书，新证书设置为默认时，旧默认证书自动变为非默认。

**HTTP BODY参数示例**

```json
{
    "cert_name": "cert_demo",
    "description": "abc",
    "is_default": true,
    "cert_file_name": "demo_cert_file_name",
    "cert_file_content": "-----BEGIN ...-----END CERTIFICATE-----",
    "key_file_name": "demo_key_file_name",
    "key_file_content": "-----BEGIN RSA PRIVATE KEY-----...-----END RSA PRIVATE KEY-----",
    "expired_date": "2021-08-23 16:02:31"
}
```

**执行逻辑**

1. 校验参数合法性。
2. 校验证书文件与密钥文件匹配。
3. 若新证书设置为默认，将旧默认证书更新为非默认。
4. 保存证书元数据及文件内容。
5. 返回证书元数据（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

| 参数名 | 类型 | 说明 |
|------|------|------|
| `cert_name` | string | 证书名 |
| `description` | string | 证书描述 |
| `is_default` | bool | 是否是默认证书 |
| `cert_file_name` | string | 主证书文件名 |
| `key_file_name` | string | 主证书密钥名 |
| `expired_date` | string | 主证书过期时间 |

> `cert_file_content`、`key_file_content` 不返回。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "cert_name": "cert_demo",
        "description": "abc",
        "is_default": true,
        "cert_file_name": "demo_cert_file_name",
        "key_file_name": "demo_key_file_name",
        "expired_date": "2021-08-23 16:02:31"
    }
}
```

### 2.2 证书列表

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取全体证书信息列表 | - |
| 端点 | /certificates | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（Query）**

无。

**执行逻辑**

1. 查询所有证书元数据。
2. 返回证书列表（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

Data 为数组，元素字段同 9.2.1 创建证书返回数据。

**成功返回示例**

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": [
        {
            "cert_name": "cert_demo",
            "description": "abc",
            "is_default": true,
            "cert_file_name": "demo_cert_file_name",
            "key_file_name": "demo_key_file_name",
            "expired_date": "2021-08-23 16:02:31"
        }
    ]
}
```

### 2.3 证书详情

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 获取单个证书信息 | - |
| 端点 | /certificates/{cert_name} | - |
| 版本 | v1 | - |
| method | GET | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 返回证书元数据（不包含 `cert_file_content` 和 `key_file_content`）。

**返回数据（Data内容）**

同 9.2.1 创建证书返回数据。

**成功返回示例**

同 9.2.1 创建证书成功返回示例。

### 2.4 更新证书为默认证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 更新证书为默认证书 | - |
| 端点 | /certificates/{cert_name}/default | - |
| 版本 | v1 | - |
| method | PATCH | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**约束**

- 更新为默认证书时，旧的默认证书自动变为非默认证书。

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 将当前默认证书更新为非默认。
3. 将目标证书更新为默认证书。
4. 返回目标证书元数据。

**返回数据（Data内容）**

同 9.2.1 创建证书返回数据。

**成功返回示例**

同 9.2.1 创建证书成功返回示例。

### 2.5 删除证书

**基本信息**

| 项目 | 值 | 说明 |
| - | - | - |
| 含义 | 删除证书 | - |
| 端点 | /certificates/{cert_name} | - |
| 版本 | v1 | - |
| method | DELETE | - |

**输入参数（URI）**

| 参数名 | 类型 | 参数含义 | 必填 | 补充描述 |
| - | - | - | - | - |
| cert_name | string | 证书名称 | Y | - |

**约束**

- 默认证书不能被删除。
- 全局必须始终保留一个默认证书。

**执行逻辑**

1. 根据 `cert_name` 查找证书。
2. 校验该证书不是默认证书。
3. 删除证书记录及证书文件内容。
4. 返回成功状态（Data 为 null）。

**返回数据（Data内容）**

无。

---
