# API-Key 模块测试

> 本目录存放 API-Key 模块的测试用例设计文档。  
> 对应的 Go 测试文件位于 `test/test-runner/test-cases/api_key/` 目录。

## 接口列表

| 接口 | 方法 | 路径 | 设计文档 | 用例数 |
|------|------|------|---------|--------|
| 创建API-Key | POST | `/open-api/v1/api-keys` | [create/design.md](./create/design.md) | 43 |
| 查询API-Key列表 | GET | `/open-api/v1/api-keys` | [list/design.md](./list/design.md) | 7 |
| 查询单个API-Key | GET | `/open-api/v1/api-keys/{id}` | [detail/design.md](./detail/design.md) | 3 |
| 全量更新API-Key | PUT | `/open-api/v1/api-keys/{id}` | [full_update/design.md](./full_update/design.md) | 5 |
| 部分更新API-Key | PATCH | `/open-api/v1/api-keys/{id}` | [partial_update/design.md](./partial_update/design.md) | 4 |
| 删除API-Key | DELETE | `/open-api/v1/api-keys/{id}` | [delete/design.md](./delete/design.md) | 3 |
| 查询配额计划 | GET | `/open-api/v1/api-keys/{id}/quota-plan` | [quota_query/design.md](./quota_query/design.md) | 3 |
| 重置配额余额 | POST | `/open-api/v1/api-keys/{id}/quota-plan/reset` | [quota_reset/design.md](./quota_reset/design.md) | 4 |
| **合计** | | | | **72** |

## 测试覆盖维度

- 正常参数：验证各接口最基本功能
- 必填校验：缺少必填字段时返回 422
- 边界值：空字符串、最大值、特殊值
- 异常参数：非法JSON、不存在资源
- 返回数据校验：逐字段验证返回结构完整性
- 业务规则：级联删除、配额重置等

## 目录结构

```
test-case-docs/api_key/          ← 设计文档（本目录）
├── README.md                    ← 本文件
├── create/design.md             ← 创建API-Key 设计文档
├── list/design.md               ← 查询列表 设计文档
├── detail/design.md             ← 查询单个 设计文档
├── full_update/design.md        ← 全量更新 设计文档
├── partial_update/design.md     ← 部分更新 设计文档
├── delete/design.md             ← 删除 设计文档
├── quota_query/design.md        ← 查询配额 设计文档
└── quota_reset/design.md        ← 重置配额 设计文档

test-cases/api_key/              ← Go 测试文件
├── create/create_test.go
├── list/list_test.go
├── detail/detail_test.go
├── full_update/full_update_test.go
├── partial_update/partial_update_test.go
├── delete/delete_test.go
├── quota_query/quota_query_test.go
└── quota_reset/quota_reset_test.go
```

## 运行命令

```bash
# 编译
cd ai-gateway-api
$env:CGO_ENABLED="0"; go build -o ai-gateway-api.exe ./main.go

# 运行测试
cd test/test-runner
go test -v -count=1 -timeout 120s ./test-cases/api_key/create/
go test -v -count=1 -timeout 120s ./test-cases/api_key/list/
go test -v -count=1 -timeout 120s ./test-cases/api_key/detail/
go test -v -count=1 -timeout 120s ./test-cases/api_key/full_update/
go test -v -count=1 -timeout 120s ./test-cases/api_key/partial_update/
go test -v -count=1 -timeout 120s ./test-cases/api_key/delete/
go test -v -count=1 -timeout 120s ./test-cases/api_key/quota_query/
go test -v -count=1 -timeout 120s ./test-cases/api_key/quota_reset/
```