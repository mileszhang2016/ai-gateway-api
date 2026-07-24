# AI Gateway API 本地测试环境（test-runner）

## 简介

test-runner 是 ai-gateway-api 的本地离线测试环境。通过 `make build` 编译项目二进制，使用 SQLite 本地数据库运行完整的 API 测试，无需 MySQL/Redis 等外部依赖。

## 目录结构

```
test-runner/
├── go.mod / go.sum                # 独立 Go module（通过 replace 指向主项目）
├── conf/
│   ├── ai_gateway_api.toml        # 测试配置文件（SQLite + SkipTokenValidate）
│   ├── nav_tree.toml              # 导航树配置
│   └── i18n/zh.toml               # 中文国际化
├── data/                          # SQLite 数据库文件（自动创建，含 .gitkeep）
├── testutil/                      # 测试工具包
│   ├── server.go                  # 编译/复制二进制、子进程启动/关闭管理
│   ├── client.go                  # HTTP 客户端封装（Get/Post/Put/Patch/Delete）
│   ├── assert.go                  # 断言函数（AssertSuccess, AssertErrCode 等）
│   ├── fixture.go                 # 测试数据工厂（随机字符串、唯一名称、证书生成等）
│   └── db.go                      # SQLite 数据库初始化（执行 DDL）
├── test-case-docs/                # 测试用例设计文档（按模块→接口组织）
│   ├── README.md                  # 设计文档总览
│   ├── api_key/                   # API-Key 模块设计文档
│   │   ├── README.md
│   │   ├── create/design.md
│   │   ├── list/design.md
│   │   └── ...（共8个接口）
│   └── ai_route/                  # AI 路由规则模块设计文档
│       ├── README.md
│       ├── set_rules/design.md
│       └── get_rules/design.md
├── test-cases/                    # Go 测试代码（按模块→接口组织）
│   ├── api_key/
│   │   ├── create/create_test.go
│   │   ├── list/list_test.go
│   │   └── ...（共8个接口）
│   └── ai_route/
│       ├── set_rules/set_rules_test.go
│       └── get_rules/get_rules_test.go
└── README.md
```

## 环境要求

| 项目 | 要求 |
|------|------|
| Go | 1.22+ |
| 操作系统 | Windows / Linux / macOS |
| 磁盘空间 | 建议 2GB+（go-sqlite 纯 Go 编译需要较大临时空间） |
| 外部依赖 | 无（SQLite 纯 Go 实现，无需 CGO；Mock Redis 无需 Redis 服务） |

## 快速开始

### 1. 编译项目二进制

```bash
# 在 ai-gateway-api 项目根目录编译
make build
# 或者: go build -o ai-gateway-api.exe .
```

### 2. 下载测试依赖

```bash
cd ai-gateway-api/test/test-runner
go mod tidy
```

### 3. 运行测试

```bash
# 运行所有模块测试
go test -v -count=1 -timeout 120s ./test-cases/api_key/create/
go test -v -count=1 -timeout 120s ./test-cases/api_key/list/
# ... 逐个包运行（避免并行编译内存不足）

# 运行指定模块测试
go test -v -count=1 -timeout 120s ./test-cases/ai_route/set_rules/
go test -v -count=1 -timeout 120s ./test-cases/ai_route/get_rules/

# 运行单个用例
go test -v -run TestSetRules_OnlyBasicRules ./test-cases/ai_route/set_rules
```

### 4. 生成测试报告

```bash
# 运行测试并输出 JSON
go test -json ./test-cases/... > ../test-reports/output.json

# 查看报告（test-reports 目录下按时间戳生成 Markdown 报告）
```

## 配置说明

### 数据库配置

```toml
[Databases.bfe_db]
DBName  = "./data/test_ai_gateway.db"  # SQLite 文件路径（自动包含进程 ID 隔离）
Driver  = "sqlite"                      # 使用 glebarez/go-sqlite 纯 Go 驱动
MaxOpenConns = 5                        # 防止连接池死锁
```

- 每次测试启动时自动执行 `db_ddl_sqlite.sql` 初始化表结构
- 数据库文件名包含进程 PID，确保多进程并发测试不会冲突
- 测试结束后自动删除数据库文件（包括 WAL/SHM 文件）

### 认证配置

```toml
[RunTime]
SkipTokenValidate = true           # 跳过 Token 认证，所有请求直接放行
```

### 服务器配置

```toml
[Server]
ServerPort = 0                    # 0 = 随机端口，避免端口冲突
MonitorPort = -1                  # -1 = 禁用监控端口
```

## testutil 工具包 API

### 服务器管理

```go
// 启动测试服务器（自动初始化 SQLite、加载配置、注册路由）
sm, err := testutil.StartServer()

// 停止测试服务器并清理资源
defer sm.Shutdown()
```

### HTTP 客户端

```go
client := testutil.GetClient()

// GET 请求
resp, err := client.Get("/open-api/v1/ai-route-rules")

// POST 请求
resp, err := client.Post("/open-api/v1/api-keys", body)

// PUT 请求
resp, err := client.Put("/open-api/v1/entities/{id}", body)

// PATCH 请求
resp, err := client.Patch("/open-api/v1/ai-route-rules", body)

// DELETE 请求
resp, err := client.Delete("/open-api/v1/api-keys/{id}")
```

### 响应结构

```go
type APIResponse struct {
    ErrNum   int             `json:"ErrNum"`   // 错误码：200=成功
    ErrMsg   string          `json:"ErrMsg"`   // 错误信息
    Data     json.RawMessage `json:"Data"`      // 返回数据
    WorkMode string          `json:"WorkMode"`  // 工作模式
}
```

### 断言函数

| 函数 | 说明 |
|------|------|
| `AssertSuccess(t, resp)` | 验证 ErrNum=200 |
| `AssertErrCode(t, resp, code)` | 验证指定错误码 |
| `AssertDataNotEmpty(t, resp)` | 验证 Data 不为空 |
| `AssertDataFieldEquals(t, resp, field, expected)` | 验证 Data 中指定字段值 |
| `AssertDataFieldNotEmpty(t, resp, field)` | 验证 Data 中指定字段不为空 |
| `AssertListLen(t, resp, len)` | 验证 Data 列表长度 |
| `AssertListFieldLen(t, resp, field, len)` | 验证 Data 中列表字段长度 |
| `AssertPagination(t, resp, page, pageSize, minTotal)` | 验证分页信息 |

### 辅助函数

```go
// 生成唯一名称
name := testutil.UniqueName("prefix")    // prefix-a1b2c3

// 生成随机字符串
s := testutil.RandomString(8)            // 8位随机字符串

// 生成自签名证书
certPEM, keyPEM, err := testutil.GenerateTestCert("common-name")

// 指针工具
p := testutil.StringPtr("value")
p := testutil.Int64Ptr(123)
p := testutil.BoolPtr(true)
```

## 测试用例编写规范

### 文件结构

测试用例分为两个目录：

- **`test-case-docs/`**：存放测试用例设计文档（`design.md`），按模块→接口组织
- **`test-cases/`**：存放 Go 测试代码（`*_test.go`），按模块→接口组织

```
test-runner/
├── test-case-docs/{module}/{interface}/design.md   # 设计文档
└── test-cases/{module}/{interface}/{interface}_test.go  # Go 测试代码
```

### 测试代码模板

```go
package {interface}

import (
    "os"
    "testing"
    "github.com/yf-networks/ai-gateway-api/test-runner/testutil"
)

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
    var err error
    sm, err = testutil.StartServer()
    if err != nil {
        panic("failed to start server: " + err.Error())
    }
    code := m.Run()
    sm.Shutdown()
    os.Exit(code)
}

func TestXXX_NormalCase(t *testing.T) {
    body := map[string]interface{}{
        "field": "value",
    }
    resp, err := testutil.GetClient().Post("/open-api/v1/xxx", body)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    testutil.AssertSuccess(t, resp)
    testutil.AssertDataFieldEquals(t, resp, "field", "value")
}
```

## 注意事项

1. **磁盘空间**：`modernc.org/sqlite` 是纯 Go 实现的 SQLite，编译时需要较大临时空间（约 1GB）。如果磁盘空间不足，请先清理 Go 缓存：
   ```bash
   go clean -cache -testcache
   ```

2. **进程隔离**：每个测试套件使用独立的数据库文件（含 PID），多进程可并发运行。

3. **认证跳过**：配置 `SkipTokenValidate = true`，所有 API 请求无需 Token 认证。

4. **Redis Mock**：Mock Redis 客户端仅支持基本操作（`GetInt64`、`IncrBy`、`Setex` 等），脚本执行仅模拟配额检查逻辑。

5. **不污染源码**：test-runner 和 test-cases 均在 `test/` 目录下，通过独立 Go module 和 replace 指令引用主项目，不会修改任何程序代码。

## 相关文档

| 文档 | 路径 |
|------|------|
| 测试用例设计文档 | [test-case-docs/](./test-case-docs/) |
| 测试设计方案 | [test/docs/local-test-design.md](../docs/local-test-design.md) |
| OpenAPI 接口文档 | [ai-gateway-api/docs/zh_cn/open_api/](../../docs/zh_cn/open_api/) |
| InnerAPI 接口设计 | [docs/瑛菲AI网关配额控制与限流-InnerAPI接口设计.md](../../../docs/瑛菲AI网关配额控制与限流-InnerAPI接口设计.md) |