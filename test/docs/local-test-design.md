# AI Gateway API 本地接口测试设计方案

**版本号**：v2.1  
**创建日期**：2026-07-23  
**文档类型**：测试设计方案  
**目标读者**：后端开发工程师、测试工程师

---

## 一、概述

### 1.1 设计目标

为 AI Gateway API 提供一套完整的本地离线接口测试方案，实现：

- **环境隔离**：测试环境和测试用例完全独立于生产代码
- **离线运行**：使用 SQLite 本地数据库，无需外部依赖（MySQL/Redis）
- **白盒测试覆盖**：覆盖正常参数、异常参数、边界值、必填校验、返回数据验证
- **模块化组织**：测试环境与测试用例分层管理，每个用例独立目录

### 1.2 测试范围

| API 模块 | 接口路径前缀 | 说明 |
|----------|-------------|------|
| OpenAPI - Auth | `/open-api/v1/auth` | 用户认证与管理 |
| OpenAPI - API-Key | `/open-api/v1/api-keys` | API-Key 管理 |
| OpenAPI - Entity-Type | `/open-api/v1/entity-types` | 实体类型管理 |
| OpenAPI - Entity | `/open-api/v1/entities` | 实体管理 |
| OpenAPI - BFE Pools | `/open-api/v1/bfe-pools` | 实例池管理 |
| OpenAPI - Clusters | `/open-api/v1/clusters` | 集群管理 |
| OpenAPI - AI Route Rule | `/open-api/v1/ai-route-rules` | 路由规则管理 |
| OpenAPI - Certificate | `/open-api/v1/certificates` | 证书管理 |
| OpenAPI - Model Provider | `/open-api/v1/models` | 模型管理 |
| InnerAPI | `/inner-api/v1/configs` | 配置导出接口 |

### 1.3 设计原则

1. **不污染生产代码**：所有测试代码、配置、资源均放在 `ai-gateway-api/test/` 目录下，主项目仅需最小改动（新增 `sqlite-strip` 驱动 + 1 行兼容代码）
2. **真实二进制子进程**：使用 `make build` 编译的真实 `ai-gateway-api.exe`，通过 `exec.CommandContext` 启动为子进程，测试覆盖完整启动链路
3. **两层分离**：测试环境（`test-runner/`）与测试用例代码（`test-cases/`）和设计文档（`test-case-docs/`）物理隔离，test-cases 和 test-case-docs 位于 test-runner 内部
4. **独立模块**：测试环境作为独立 Go module，通过 `replace` 指令引用主项目和 bfe 依赖
5. **用例独立**：每个测试用例拥有独立目录，包含设计文档和用例代码
6. **数据库隔离**：每个测试套件使用独立的 SQLite 数据库文件（含进程ID）
7. **自动化清理**：测试完成后自动清理临时文件、复制的二进制和数据库文件（含 WAL/SHM）
8. **跳过认证**：测试环境配置 `SkipTokenValidate=true`，无需真实认证

---

## 二、测试环境设计

### 2.1 目录结构

```
ai-gateway-api/test/
├── README.md                                  # 测试使用说明
│
├── docs/                                      # 测试方案文档
│   └── local-test-design.md                   # 测试设计方案（本文档）
│
├── test-reports/                              # 测试结果报告（独立目录）
│   └── YYYY-MM-DD_HHMMSS/                     # 按时间戳命名的报告目录
│       ├── summary.md                         # 测试汇总报告（总览）
│       ├── auth.md                            # Auth 模块测试报告
│       ├── api_key.md                         # API-Key 模块测试报告
│       ├── entity_type.md                     # Entity-Type 模块测试报告
│       ├── entity.md                          # Entity 模块测试报告
│       ├── bfe_pools.md                       # 实例池模块测试报告
│       ├── clusters.md                        # 集群模块测试报告
│       ├── ai_route.md                        # AI 路由模块测试报告
│       ├── certificate.md                     # 证书模块测试报告
│       └── innerapi.md                        # InnerAPI 模块测试报告
│
├── test-runner/                               # 测试环境（测试框架基础设施）
│   ├── go.mod                                 # 测试专用 go.mod（独立 module，replace 指向主项目）
│   ├── go.sum
│   ├── conf/                                  # 测试专用配置
│   │   ├── ai_gateway_api.toml                # SQLite(sqlite-strip) + SkipTokenValidate=true
│   │   ├── nav_tree.toml                      # 导航树配置
│   │   └── i18n/                              # 国际化
│   │       └── zh.toml
│   ├── data/                                  # 测试运行时数据目录（gitignore）
│   │   └── .gitkeep
│   ├── testutil/                              # 测试工具包
│   │   ├── server.go                          # 子进程服务器管理（复制真实二进制、启动/关闭）
│   │   ├── client.go                          # HTTP 客户端封装
│   │   ├── assert.go                          # 自定义断言函数
│   │   ├── fixture.go                         # 测试数据工厂
│   │   ├── db.go                              # SQLite 数据库初始化/清理
│   │   └── mock_redis.go                      # 内存 Redis Mock（备用，当前不使用）
│   │
│   ├── test-case-docs/                        # 测试用例设计文档（按模块→接口组织）
│   │   ├── README.md                          # 设计文档总览 + 模板
│   │   ├── api_key/                           # API-Key 模块设计文档
│   │   │   ├── README.md
│   │   │   ├── create/design.md
│   │   │   ├── list/design.md
│   │   │   └── ...（共8个接口）
│   │   └── ai_route/                          # AI 路由规则模块设计文档
│   │       ├── README.md
│   │       ├── set_rules/design.md
│   │       └── get_rules/design.md
│   │
│   └── test-cases/                            # Go 测试代码（按模块→接口组织）
│   │
│   ├── auth/                                  # 认证模块
│   │   ├── create_user/                       # 创建用户
│   │   │   └── create_user_test.go            # 用例代码
│   │   ├── delete_user/                       # 删除用户
│   │   │   └── delete_user_test.go
│   │   ├── reset_password/                    # 重置密码
│   │   │   └── reset_password_test.go
│   │   ├── list_users/                        # 用户列表
│   │   │   └── list_users_test.go
│   │   ├── set_admin/                         # 设置管理员
│   │   │   └── set_admin_test.go
│   │   ├── bind_product/                      # 绑定产品线
│   │   │   └── bind_product_test.go
│   │   ├── unbind_product/                    # 解除产品线绑定
│   │   │   └── unbind_product_test.go
│   │   ├── search_by_product/                 # 按产品线查用户
│   │   │   └── search_by_product_test.go
│   │   ├── create_session_key/                # 创建 Session Key
│   │   │   └── create_session_key_test.go
│   │   ├── delete_session_key/                # 删除 Session Key
│   │   │   └── delete_session_key_test.go
│   │   ├── create_token/                      # 创建 Token
│   │   │   └── create_token_test.go
│   │   ├── delete_token/                      # 删除 Token
│   │   │   └── delete_token_test.go
│   │   ├── token_detail/                      # Token 详情
│   │   │   └── token_detail_test.go
│   │   ├── token_list/                        # Token 列表
│   │   │   └── token_list_test.go
│   │   └── search_token_by_product/           # 按产品线查 Token
│   │       └── search_token_by_product_test.go
│   │
│   ├── api_key/                               # API-Key 模块
│   │   ├── create/                            # 创建 API-Key
│   │   │   └── create_test.go
│   │   ├── list/                              # API-Key 列表
│   │   │   └── list_test.go
│   │   ├── detail/                            # API-Key 详情
│   │   │   └── detail_test.go
│   │   ├── full_update/                       # 全量更新
│   │   │   └── full_update_test.go
│   │   ├── partial_update/                    # 部分更新
│   │   │   └── partial_update_test.go
│   │   ├── delete/                            # 删除 API-Key
│   │   │   └── delete_test.go
│   │   ├── quota_query/                       # 查询配额计划
│   │   │   └── quota_query_test.go
│   │   └── quota_reset/                       # 重置配额
│   │       └── quota_reset_test.go
│   │
│   ├── entity_type/                           # Entity-Type 模块
│   │   ├── create/                            # 创建实体类型
│   │   │   └── create_test.go
│   │   ├── list/                              # 实体类型列表
│   │   │   └── list_test.go
│   │   ├── detail/                            # 实体类型详情
│   │   │   └── detail_test.go
│   │   ├── update/                            # 更新实体类型
│   │   │   └── update_test.go
│   │   └── delete/                            # 删除实体类型
│   │       └── delete_test.go
│   │
│   ├── entity/                                # Entity 模块
│   │   ├── create/                            # 创建实体
│   │   │   └── create_test.go
│   │   ├── list/                              # 实体列表
│   │   │   └── list_test.go
│   │   ├── detail/                            # 实体详情
│   │   │   └── detail_test.go
│   │   ├── full_update/                       # 全量更新
│   │   │   └── full_update_test.go
│   │   ├── partial_update/                    # 部分更新
│   │   │   └── partial_update_test.go
│   │   ├── delete/                            # 删除实体
│   │   │   └── delete_test.go
│   │   └── quota_reset/                       # 重置配额
│   │       └── quota_reset_test.go
│   │
│   ├── bfe_pools/                             # 实例池模块
│   │   ├── get_pool/                          # 获取实例池
│   │   │   └── get_pool_test.go
│   │   └── update_instance/                   # 更新实例
│   │       └── update_instance_test.go
│   │
│   ├── clusters/                              # 集群模块
│   │   ├── create/                            # 创建集群
│   │   │   └── create_test.go
│   │   ├── list/                              # 集群列表
│   │   │   └── list_test.go
│   │   ├── detail/                            # 集群详情
│   │   │   └── detail_test.go
│   │   ├── update/                            # 更新集群
│   │   │   └── update_test.go
│   │   ├── delete/                            # 删除集群
│   │   │   └── delete_test.go
│   │   ├── ready_check/                       # 就绪检查
│   │   │   └── ready_check_test.go
│   │   └── model_list/                        # 模型列表
│   │       └── model_list_test.go
│   │
│   ├── ai_route/                              # AI 路由规则模块
│   │   ├── set_rules/                         # 设置路由规则
│   │   │   └── set_rules_test.go
│   │   └── get_rules/                         # 获取路由规则
│   │       └── get_rules_test.go
│   │
│   ├── certificate/                           # 证书模块
│   │   ├── create/                            # 创建证书
│   │   │   └── create_test.go
│   │   ├── list/                              # 证书列表
│   │   │   └── list_test.go
│   │   ├── detail/                            # 证书详情
│   │   │   └── detail_test.go
│   │   ├── update/                            # 更新证书
│   │   │   └── update_test.go
│   │   └── delete/                            # 删除证书
│   │       └── delete_test.go
│   │
│   ├── model_provider/                        # 模型提供商模块
│   │   └── list_models/                       # 查询模型列表
│   │       └── list_models_test.go
│   │
│   └── innerapi/                              # InnerAPI 模块
│       ├── mod_api_key/                       # API-Key 配置导出
│       │   └── mod_api_key_test.go
│       ├── rate_limit_policy/                 # 限流策略导出
│       │   └── rate_limit_policy_test.go
│       ├── server_data/                       # 服务数据导出
│       │   └── server_data_test.go
│       ├── gslb_data/                         # GSLB 数据导出
│       │   └── gslb_data_test.go
│       ├── protocol/                          # 证书配置导出
│       │   └── protocol_test.go
│       ├── extra_file/                        # 额外文件导出
│       │   └── extra_file_test.go
│       └── mod_body_process/                  # Body 处理模块导出
│           └── mod_body_process_test.go
```

### 2.2 目录设计说明

#### 2.2.1 test-runner/（测试环境）

`test-runner/` 是测试运行的基础设施，采用**真实二进制子进程**模式运行测试服务器：

| 子目录/文件 | 说明 |
|-------------|------|
| `go.mod` / `go.sum` | 独立 Go module，通过 `replace` 指令引用主项目和 bfe 依赖 |
| `conf/` | 测试专用配置文件（含 nav_tree.toml 和 i18n/zh.toml） |
| `data/` | 测试运行时数据目录，存放复制的二进制和日志文件，测试结束后自动清理 |
| `testutil/` | 测试工具包，提供子进程管理、HTTP 客户端、断言、数据库初始化 |
| `test-case-docs/` | 测试用例设计文档（按模块→接口组织，含 README.md 和 design.md） |
| `test-cases/` | Go 测试代码（按模块→接口组织，含 *_test.go） |

#### 2.2.2 test-case-docs/ 和 test-cases/

测试用例分为两个目录，物理隔离设计文档和测试代码：

- **`test-case-docs/`**：存放测试用例设计文档（`design.md`），按模块→接口组织。每个模块目录下含 `README.md` 说明该模块的接口列表和测试覆盖。
- **`test-cases/`**：存放 Go 测试代码（`*_test.go`），按模块→接口组织。每个接口目录下仅含 `*_test.go` 文件。

两目录保持相同的层级结构，便于交叉引用。

### 2.3 Go Module 设计

`test-runner/go.mod` 内容：

```go
module github.com/yf-networks/ai-gateway-api/test-runner

go 1.22

require (
    github.com/bfenetworks/bfe v1.8.1-0.20260427052401-8d3a8cd44d98
    github.com/glebarez/go-sqlite v1.21.2        // 纯 Go SQLite 驱动，无 CGO
    github.com/gorilla/mux v1.8.0                 // HTTP 路由
    github.com/stretchr/testify v1.10.0           // 测试断言库
    github.com/yf-networks/ai-gateway-api v0.0.0  // 主项目
    gopkg.in/tylerb/graceful.v1 v1.2.15           // 优雅关闭
)

replace github.com/yf-networks/ai-gateway-api => ../../
replace github.com/bfenetworks/bfe => ../../../bfe_ai
```

**说明**：
- `test-runner` 作为独立 Go module，通过 `replace` 指令指向主项目源码（`../../`）和 bfe 依赖（`../../../bfe_ai`）
- 必须引入 `github.com/yf-networks/ai-gateway-api` 以触发 `stateful` 包的 `init()` 注册 `sqlite-strip` 驱动
- 测试用例代码通过 `import "github.com/yf-networks/ai-gateway-api/test-runner/testutil"` 使用测试工具包

### 2.4 测试配置文件

`test-runner/conf/ai_gateway_api.toml`：

```toml
[Server]
ServerPort = 8199              # 测试端口（实际运行时会被替换为随机端口）
GracefulTimeOutInMs = 5000
MonitorPort = -1               # 禁用监控端口

[Loggers.access]
LogName = "access"
LogLevel = "DEBUG"
RotateWhen = "MIDNIGHT"
BackupCount = 1
Format = "[%D %T] [%L] [%S] %M"
StdOut = false                 # 关闭 stdout 输出，避免干扰子进程通信

[Loggers.sql]
LogName = "sql"
LogLevel = "DEBUG"
RotateWhen = "MIDNIGHT"
BackupCount = 1
Format = "[%D %T] %M"
StdOut = false

[Loggers.exception]
LogName = "exception"
LogLevel = "DEBUG"
RotateWhen = "MIDNIGHT"
BackupCount = 7
Format = "[%D %T] [%L] [%S] %M"
StdOut = false

[Databases.bfe_db]
DBName  = "./data/test_ai_gateway.db"  # 运行时会被替换为实际路径
Driver  = "sqlite-strip"               # 自定义驱动，剥离 FOR UPDATE 子句
MaxOpenConns = 5                       # 防止连接池死锁
MaxIdleConns = 1
ConnMaxIdleTimeInMs = 500000
ConnMaxLifetimeInMs = 5000000

[Depends]
NavTreeFile = "${conf_dir}/nav_tree.toml"
I18nDir     = "${conf_dir}/i18n"

[RunTime]
SkipTokenValidate = true               # 跳过 Token 验证，方便测试
RecordSQL = false                      # 关闭 SQL 日志
SessionExpireInDay = 10
StaticFilePath = "./data"
Debug = false
AIRouteInnerProductName = "AI_product"
DefaultAIInstancePoolName = "BFE.aipool"
DefaultAIClusterName = "BFE-AI_product.szyf"
```

**关键配置说明**：
- `Driver = "sqlite-strip"`：使用自定义 SQLite 驱动，在 Prepare 阶段自动剥离 `FOR UPDATE` 子句
- `StdOut = false`（所有日志器）：关闭 stdout 输出，避免干扰子进程管理
- `SkipTokenValidate = true`：跳过认证，所有请求直接放行
- `RecordSQL = false`：关闭 SQL 记录，减少日志噪音

### 2.5 测试服务器设计

#### 2.5.1 架构概述

测试服务器采用**真实二进制子进程**模式，而非嵌入式服务器。核心思路：

1. 使用 `make build`（或 `go build`）编译出真实的 `ai-gateway-api.exe`
2. 测试运行时，将二进制复制到 `test-runner/data/` 目录
3. 生成临时配置文件（覆盖端口和数据库路径）
4. 通过 `exec.CommandContext` 启动子进程
5. 通过 TCP 拨号检测确认服务就绪
6. 测试用例通过 HTTP 客户端连接子进程的 API 接口

**为什么采用子进程模式**：
- 使用真实的 `ai-gateway-api.exe`，测试覆盖完整的启动链路（配置加载、中间件注册、路由注册等）
- 避免嵌入式服务器中的依赖注入和初始化顺序问题
- 测试环境与生产环境行为一致

#### 2.5.2 服务器启动流程

`test-runner/testutil/server.go` 中的 `StartServer()` 函数：

```
StartServer()
    │
    ├─ 1. 定位 test-runner 目录和项目根目录
    ├─ 2. 生成唯一数据库路径（test_ai_gateway_{pid}.db）
    ├─ 3. 初始化 SQLite 数据库（执行 DDL）
    ├─ 4. 查找项目根目录下的 ai-gateway-api.exe（由 make build 编译）
    ├─ 5. 复制二进制到 data/ 目录（避免文件锁冲突）
    ├─ 6. 生成随机端口 + 创建临时配置文件（覆盖端口和 DB 路径）
    ├─ 7. 通过 exec.CommandContext 启动子进程
    ├─ 8. 消费 stdout 输出（防止子进程阻塞）
    ├─ 9. 等待服务器就绪（TCP 拨号检测，超时 10s）
    └─ 10. 返回 ServerManager（含 ServerURL）
```

#### 2.5.3 服务器停止流程

```
Shutdown()
    │
    ├─ 1. 取消 context，终止子进程
    ├─ 2. 等待子进程退出
    ├─ 3. 清理复制的二进制文件
    ├─ 4. 清理临时配置文件目录
    └─ 5. 清理 SQLite 数据库文件（.db, -wal, -shm）
```

#### 2.5.4 SQLite 自定义驱动（sqlite-strip）

SQLite 不支持 `SELECT ... FOR UPDATE` 语法，需要实现自定义驱动在 `Prepare` 阶段剥离该子句。

**重要**：`sqlite-strip` 驱动位于主项目 `stateful/sqlite_strip.go`（而非 testutil），通过 `init()` 自动注册。这样编译出的 `ai-gateway-api.exe` 天然包含该驱动。

```go
// stateful/sqlite_strip.go
package stateful

import (
    "context"
    "database/sql"
    "database/sql/driver"
    "strings"

    sqlite "github.com/glebarez/go-sqlite"
)

type sqliteStripDriver struct {
    inner driver.Driver
}

func init() {
    sql.Register("sqlite-strip", &sqliteStripDriver{inner: &sqlite.Driver{}})
}

func (d *sqliteStripDriver) Open(name string) (driver.Conn, error) {
    conn, err := d.inner.Open(name)
    if err != nil {
        return nil, err
    }
    return &stripConn{Conn: conn}, nil
}

type stripConn struct {
    driver.Conn
}

func (c *stripConn) Prepare(query string) (driver.Stmt, error) {
    return c.Conn.Prepare(stripForUpdate(query))
}

func (c *stripConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
    if cc, ok := c.Conn.(driver.ConnPrepareContext); ok {
        return cc.PrepareContext(ctx, stripForUpdate(query))
    }
    return c.Conn.Prepare(stripForUpdate(query))
}

func stripForUpdate(query string) string {
    upper := strings.ToUpper(query)
    idx := strings.Index(upper, "FOR UPDATE")
    if idx < 0 {
        return query
    }
    return strings.TrimSpace(query[:idx])
}
```

#### 2.5.5 主项目最小改动

为支持测试环境，主项目 `ai-gateway-api` 仅需两处改动：

1. **新增** `stateful/sqlite_strip.go`：注册 `sqlite-strip` 自定义驱动
2. **修改** `stateful/config_database.go` 第 48 行：`FormatDSN` 允许 `"sqlite-strip"` 驱动名

```go
// config_database.go 第 48 行
case DriverSQLite, "sqlite-strip":
```

#### 2.5.6 关键实现细节

**临时配置文件创建**：
- 复制 `conf/` 目录到系统临时目录
- 修改 `ai_gateway_api.toml` 中的 `ServerPort` 和 `DBName`
- 数据库路径使用正斜杠（`/`），避免 Windows 下 TOML 转义问题

**TCP 就绪检测**：
- 使用 `net.DialTimeout` 拨号 `127.0.0.1:{port}`，超时 10 秒
- 每 100ms 重试一次
- 不使用 HTTP GET 检测（避免 404 误判）

**子进程通信**：
- 通过 `context.WithCancel` 管理子进程生命周期
- 启动 goroutine 消费 stdout 输出（防止子进程写管道阻塞）
- stderr 输出缓存，用于启动失败时的调试

### 2.6 测试执行入口

使用 Go `TestMain` 模式管理服务器生命周期，每个测试包（接口目录）有独立的 `TestMain`：

```go
package set_rules

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
```

**说明**：
- `TestMain` 所在的包同时作为测试包名（如 `package set_rules`），该包下所有 `TestXxx` 函数共享同一个服务器实例
- 每个测试包（接口）启动一个独立的服务器实例，不同接口之间互相隔离
- `StartServer()` 返回 `*ServerManager`，包含 `ServerURL` 字段，供 HTTP 客户端使用

---

## 三、测试工具设计

### 3.1 HTTP 客户端封装 (`testutil/client.go`)

```go
// 封装 HTTP 请求，统一处理认证头、响应解析
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    Token      string  // 当前测试使用的 Token
}

// 通用方法
func (c *Client) Get(path string) (*APIResponse, error)
func (c *Client) Post(path string, body interface{}) (*APIResponse, error)
func (c *Client) Patch(path string, body interface{}) (*APIResponse, error)
func (c *Client) Put(path string, body interface{}) (*APIResponse, error)
func (c *Client) Delete(path string) (*APIResponse, error)
```

### 3.2 断言函数 (`testutil/assert.go`)

```go
// 验证 API 响应的通用断言
func AssertSuccess(t *testing.T, resp *APIResponse)
func AssertErrCode(t *testing.T, resp *APIResponse, expectedErrNum int)
func AssertDataNotEmpty(t *testing.T, resp *APIResponse)
func AssertDataFieldEquals(t *testing.T, resp *APIResponse, field string, expected interface{})
func AssertListLen(t *testing.T, resp *APIResponse, expectedLen int)
func AssertPagination(t *testing.T, resp *APIResponse, expectedPage, expectedPageSize, minTotal int)
```

### 3.3 测试数据工厂 (`testutil/fixture.go`)

提供常用测试数据的构造方法：

```go
// 各类测试数据的默认构造器
func NewDefaultClusterParam(name string) ClusterParam
func NewDefaultAPIKeyParam(desc string) APIKeyParam
func NewDefaultEntityTypeParam(name string, level int) EntityTypeParam
func NewDefaultEntityParam(name, etype string) EntityParam
func NewDefaultCertificateParam(name string) CertificateParam
func NewDefaultRouteRuleParam() RouteRuleParam
func NewDefaultUserParam(name string) UserParam
func UniqueName(prefix string) string  // 生成唯一名称，避免冲突
func GenerateCert() (string, string)   // 生成自签名证书（用于测试）
```

---

## 四、测试用例设计规范

### 4.1 测试用例编号规则

```
{模块缩写}-{接口编号}-{场景编号}
```

模块缩写：
- `AUTH` - 认证模块
- `AK` - API-Key 模块
- `ET` - Entity-Type 模块
- `E` - Entity 模块
- `BP` - 实例池模块
- `CL` - 集群模块
- `AR` - AI 路由模块
- `CERT` - 证书模块
- `MP` - 模型提供商模块
- InnerAPI 各子模块

### 4.2 白盒测试覆盖维度

| 测试类型 | 覆盖内容 | 示例 |
|----------|---------|------|
| **正常参数** | 所有必填参数合法值 | 正常创建 API-Key |
| **异常参数** | 参数缺失、类型错误、格式错误 | 缺少 description、无效 UUID |
| **边界值** | 参数取值范围边界 | quota=0、page_size=100（最大值） |
| **必填校验** | 必填字段为空或不传 | 缺少 name、缺少 id |
| **返回数据** | 返回字段完整性、类型正确性 | 验证返回 id、name、create_time |
| **业务规则** | 业务逻辑校验 | Entity 层级校验、Quota 余额更新 |

### 4.3 测试用例场景设计原则

#### 4.3.1 场景分组规范

每个接口的测试用例必须按以下场景类型进行分组，确保测试覆盖全面且结构清晰：

| 场景类型 | 说明 | 设计要点 |
|---------|------|---------|
| **正常参数** | 所有参数使用合法值，验证接口正常功能 | 覆盖每个参数的独立验证及组合场景 |
| **必填校验** | 验证必填字段缺失或为空时的行为 | 每个必填字段至少一个缺失用例和一个空值用例 |
| **边界值** | 验证参数取值范围边界 | 覆盖最小值、最大值、刚好合法、刚好非法 |
| **异常参数** | 验证非法参数的处理 | 类型错误、格式错误、超出范围、违反业务约束 |
| **返回数据** | 验证返回数据结构和字段完整性 | 验证所有字段存在、类型正确、值符合预期 |

#### 4.3.2 参数全覆盖原则

**每个请求参数（包括子结构内的每个嵌套参数）都必须有独立的测试用例场景覆盖**，确保不遗漏任何参数：

1. **顶层参数覆盖**：对于接口请求 Body 中的每个顶层参数，至少设计一个独立测试用例对该参数进行验证
2. **子结构参数覆盖**：对于 `object` 类型的参数（如 `quota_plan`、`rate_limit_policy`），其内部每个子字段都需要独立的测试用例
3. **嵌套子结构参数覆盖**：对于多层嵌套的参数（如 `rate_limit_policy.rules.tpm`），每个叶子字段都需要独立覆盖

示例：创建 API-Key 接口的参数覆盖要求：

| 参数层级 | 参数路径 | 覆盖方式 |
|---------|---------|---------|
| 顶层 | `description` | 正常参数 + 必填校验 + 边界值 |
| 顶层 | `expired_time` | 正常参数 + 边界值(-1) + 异常参数(过去时间、-2) |
| 顶层 | `enabled` | 正常参数(true/false) |
| 顶层 | `unlimited_quota` | 正常参数(true/false) |
| 顶层 | `models` | 正常参数 + 边界值(空数组) |
| 顶层 | `subnet` | 正常参数 + 边界值(空数组) + 异常参数(无效CIDR) |
| 顶层 | `entity_id` | 正常参数(有效Entity) + 异常参数(不存在的Entity) |
| 子结构 | `quota_plan.unlimited` | 独立正常参数用例 |
| 子结构 | `quota_plan.pass_when_no_enough_quota` | 独立正常参数用例 |
| 子结构 | `quota_plan.quota` | 独立正常参数 + 边界值(0) + 异常参数(负数) |
| 子结构 | `quota_plan.unit` | 独立正常参数用例 |
| 子结构 | `quota_plan.reset_period` | 独立正常参数用例 |
| 子结构 | `rate_limit_policy.enabled` | 正常参数 + 异常参数(enabled=true但无规则) |
| 嵌套子结构 | `rate_limit_policy.rules.tpm` | 独立正常参数用例(仅tpm) |
| 嵌套子结构 | `rate_limit_policy.rules.rpm` | 独立正常参数用例(仅rpm) |
| 嵌套子结构 | `rate_limit_policy.rules.max_concurrency` | 独立正常参数 + 边界值(0) |
| 嵌套叶子 | `rules.tpm.window_minutes` | 边界值(1,360) + 异常参数(0,361) |
| 嵌套叶子 | `rules.tpm.step_minutes` | 异常参数(step>window) |
| 嵌套叶子 | `rules.rpm.window_minutes` | 边界值(1) + 异常参数(0) |

#### 4.3.3 双文件同步更新原则

每个接口的测试用例设计完成后，必须同时更新以下两个文件：

| 文件 | 路径 | 内容 |
|------|------|------|
| 设计文档 | `test-case-docs/{module}/{interface}/design.md` | 接口信息、参数说明、场景总览、每个场景的详细设计 |
| 测试代码 | `test-cases/{module}/{interface}/{interface}_test.go` | 与设计文档中每个场景一一对应的 Go 测试函数 |

**同步要求**：
- 设计文档中的每个场景编号（如 `AK-1-001`）必须在测试代码中有对应的测试函数
- 测试函数命名规范：`Test{Interface}_{SceneType}_{ScenarioName}`，如 `TestCreate_Normal_MinimalParams`
- 设计文档中的预期结果（ErrNum、字段值）必须与测试代码中的断言一致
- 新增或修改测试场景时，两个文件必须同步更新

### 4.4 用例目录结构规范

测试用例分为两个目录，保持相同的层级结构：

- **设计文档**：`test-case-docs/{module}/{interface}/design.md`
- **测试代码**：`test-cases/{module}/{interface}/{interface}_test.go`

```
test-runner/
├── test-case-docs/{module}/{interface}/design.md   # 设计文档
└── test-cases/{module}/{interface}/{interface}_test.go  # Go 测试代码
```

**design.md 模板**：

```markdown
# {接口名称} - 测试用例设计

## 一、接口信息

| 项目 | 值 |
|------|-----|
| 模块 | {模块名称，如 API-Key} |
| 接口名称 | {接口中文名，如 创建 API-Key} |
| 方法 | POST |
| 路径 | /open-api/v1/api-keys |
| 说明 | 创建一个新的 API-Key，支持配置配额计划和限流策略 |

---

## 二、接口参数说明

### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| description | string | 是 | API-Key 描述 | - |
| expired_time | int64 | 否 | 过期时间（Unix时间戳秒，-1永不过期） | -1 |
| enabled | bool | 否 | 是否启用 | true |
| unlimited_quota | bool | 否 | 是否无限配额 | false |
| models | []string | 否 | 允许访问的模型白名单 | ["*"] |
| subnet | []string | 否 | 允许的客户端子网 | ["*"] |
| quota_plan | object | 否 | 配额计划 | 使用默认值 |
| rate_limit_policy | object | 否 | 限流策略 | enabled=false |
| entity_id | string | 否 | 挂载的 Entity ID | 空 |

### 返回数据字段

| 参数名 | 类型 | 说明 |
|--------|------|------|
| id | string | API-Key 唯一标识 |
| key | string | API-Key 值 |
| description | string | 描述 |
| enabled | bool | 是否启用 |
| create_time | int64 | 创建时间 |
| update_time | int64 | 更新时间 |
| expired_time | int64 | 过期时间 |
| unlimited_quota | bool | 是否无限配额 |
| models | []string | 允许的模型白名单 |
| subnet | []string | 允许的客户端子网 |
| quota_plan | object | 配额计划（不含 balance） |
| rate_limit_policy | object | 限流策略 |
| entity_id | string | 挂载的 Entity ID |
| entity | object | 挂载的 Entity 摘要（id、name、type） |

---

## 三、测试场景总览

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AK-1-001 | 最小参数创建 | 正常参数 | 仅传 description，验证返回完整字段 |
| AK-1-002 | 带配额计划创建 | 正常参数 | 验证 quota_plan 返回正确 |
| AK-1-003 | 缺少 description | 必填校验 | 验证 ErrNum=422 |
| ... | ... | ... | ... |

---

## 四、测试场景详细设计

---

### AK-1-001：最小参数创建（正常参数）

#### 设计思路

验证 API-Key 创建接口的最基本功能：仅传入必填字段 description，所有可选字段使用默认值，确认接口返回正确的 API-Key 信息。

#### 前提数据准备

- 无需预先创建任何数据
- 确保测试环境已启动，认证方式为 SkipTokenValidate

#### 执行步骤

1. 构造请求 Body：`{"description": "test-key-001"}`
2. 发送 POST 请求到 `/open-api/v1/api-keys`
3. 验证响应状态码和返回结构

#### 请求参数

```json
{
    "description": "test-key-001"
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  
**WorkMode**：ModeNormal  

**Data 字段校验**：

| 字段 | 预期值 | 校验方式 |
|------|--------|---------|
| id | 非空字符串 | NotEmpty |
| key | 非空字符串 | NotEmpty |
| description | "test-key-001" | Equals |
| enabled | true | Equals（默认值） |
| create_time | 大于 0 的 int64 | GreaterThan(0) |
| update_time | 等于 create_time | Equals(create_time) |
| expired_time | -1 | Equals（默认值） |
| unlimited_quota | false | Equals（默认值） |
| models | ["*"] | Equals（默认值） |
| subnet | ["*"] | Equals（默认值） |
| quota_plan | 不为 null | NotNull |
| rate_limit_policy | 不为 null | NotNull |
| entity_id | 空字符串 | Equals（默认值） |
| entity | null | Equals（默认值，未挂载时为 null） |

---

### AK-1-002：带配额计划创建（正常参数）

#### 设计思路

验证创建 API-Key 时传入自定义配额计划参数的场景，确认配额计划字段正确持久化并返回。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造请求 Body，包含 `description` 和 `quota_plan`
2. 发送 POST 请求到 `/open-api/v1/api-keys`
3. 验证返回的 `quota_plan` 字段

#### 请求参数

```json
{
    "description": "test-key-002",
    "quota_plan": {
        "unlimited": false,
        "pass_when_no_enough_quota": false,
        "quota": 100000000,
        "unit": "total_token",
        "reset_period": "monthly"
    }
}
```

#### 预期返回结果

**ErrNum**：200  
**ErrMsg**：success  

**Data.quota_plan 字段校验**：

| 字段 | 预期值 |
|------|--------|
| quota_plan.unlimited | false |
| quota_plan.pass_when_no_enough_quota | false |
| quota_plan.quota | 100000000 |
| quota_plan.unit | "total_token" |
| quota_plan.reset_period | "monthly" |

---

### AK-1-003：缺少 description（必填校验）

#### 设计思路

验证 `description` 为必填字段，当请求 Body 中不传该字段时，接口应返回参数校验错误。

#### 前提数据准备

- 无需预先创建任何数据

#### 执行步骤

1. 构造空的请求 Body：`{}`
2. 发送 POST 请求到 `/open-api/v1/api-keys`
3. 验证返回错误码

#### 请求参数

```json
{}
```

#### 预期返回结果

**ErrNum**：422  
**ErrMsg**：包含 "description" 的错误信息  
**Data**：null

---

...(其他用例按同样格式编写)
```

---

**模板说明**：

每个 `design.md` 文件包含以下四个部分：

| 章节 | 说明 |
|------|------|
| **一、接口信息** | 接口的基本元信息（模块、方法、路径、说明） |
| **二、接口参数说明** | 请求参数和返回数据字段的完整定义 |
| **三、测试场景总览** | 所有测试场景的汇总表，快速了解测试覆盖范围 |
| **四、测试场景详细设计** | 每个场景按编号详细展开，包含设计思路、前提数据准备、执行步骤、请求参数、预期返回结果 |

每个测试场景详细设计包含以下 5 个要素：

| 要素 | 说明 |
|------|------|
| **设计思路** | 说明该用例的测试目标和验证要点 |
| **前提数据准备** | 列出执行该用例前需要预先创建的数据（如需要先创建 Entity、Entity-Type 等） |
| **执行步骤** | 明确的步骤化操作指令 |
| **请求参数** | 完整的 JSON 请求体示例 |
| **预期返回结果** | ErrNum、ErrMsg、Data 各字段的预期值及校验方式 |

### 4.5 测试用例设计

基于 [open_api 接口文档](../../docs/zh_cn/open_api/) 和 [InnerAPI 接口设计文档](../../../docs/瑛菲AI网关配额控制与限流-InnerAPI接口设计.md) 重新设计。

---

#### 4.5.1 Auth 模块（认证/授权）

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| AUTH-1-001 | 创建用户 | POST | `/auth/users` | 正常创建普通用户 | 正常参数 | user_name+password+is_admin=false |
| AUTH-1-002 | 创建用户 | POST | `/auth/users` | 正常创建管理员 | 正常参数 | is_admin=true |
| AUTH-1-003 | 创建用户 | POST | `/auth/users` | 缺少 user_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-004 | 创建用户 | POST | `/auth/users` | 缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-005 | 创建用户 | POST | `/auth/users` | 缺少 is_admin | 必填校验 | 验证 ErrNum=422 |
| AUTH-1-006 | 创建用户 | POST | `/auth/users` | 重复创建同名用户 | 业务规则 | 验证 ErrNum=555 |
| AUTH-1-007 | 创建用户 | POST | `/auth/users` | user_name 为空字符串 | 边界值 | 验证 ErrNum=422 |
| AUTH-1-008 | 创建用户 | POST | `/auth/users` | password 为空字符串 | 边界值 | 验证 ErrNum=422 |
| AUTH-2-001 | 删除用户 | DELETE | `/auth/users/{user_name}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-2-002 | 删除用户 | DELETE | `/auth/users/{user_name}` | 删除不存在的用户 | 异常参数 | 验证 ErrNum=404 |
| AUTH-3-001 | 重置密码 | PATCH | `/auth/users/{user_name}/passwd` | 管理员重置他人密码 | 正常参数 | 无需 old_password |
| AUTH-3-002 | 重置密码 | PATCH | `/auth/users/{user_name}/passwd` | 用户自己重置密码 | 正常参数 | 需 old_password |
| AUTH-3-003 | 重置密码 | PATCH | `/auth/users/{user_name}/passwd` | 缺少 password | 必填校验 | 验证 ErrNum=422 |
| AUTH-3-004 | 重置密码 | PATCH | `/auth/users/{user_name}/passwd` | 修改不存在的用户 | 异常参数 | 验证 ErrNum=404 |
| AUTH-3-005 | 重置密码 | PATCH | `/auth/users/{user_name}/passwd` | old_password 错误 | 异常参数 | 验证 ErrNum=422 |
| AUTH-4-001 | 用户列表 | GET | `/auth/users` | 获取用户列表 | 正常参数 | 验证返回数组，包含 user_name、is_admin |
| AUTH-4-002 | 用户列表 | GET | `/auth/users` | 创建后验证列表包含新用户 | 返回数据 | 验证列表包含新用户 |
| AUTH-5-001 | 设置管理员 | PATCH | `/auth/users/{user_name}/is_admin` | 设置为管理员 | 正常参数 | 验证 ErrNum=200 |
| AUTH-5-002 | 设置管理员 | PATCH | `/auth/users/{user_name}/is_admin` | 取消管理员权限 | 正常参数 | 验证 ErrNum=200 |
| AUTH-5-003 | 设置管理员 | PATCH | `/auth/users/{user_name}/is_admin` | 缺少 is_admin | 必填校验 | 验证 ErrNum=422 |
| AUTH-5-004 | 设置管理员 | PATCH | `/auth/users/{user_name}/is_admin` | 修改不存在用户 | 异常参数 | 验证 ErrNum=404 |
| AUTH-6-001 | 绑定产品线 | POST | `/auth/users/{user_name}/products/{product_name}` | 正常绑定 | 正常参数 | 验证 ErrNum=200 |
| AUTH-6-002 | 绑定产品线 | POST | `/auth/users/{user_name}/products/{product_name}` | 绑定不存在的产品线 | 异常参数 | 验证 ErrNum=404 |
| AUTH-6-003 | 绑定产品线 | POST | `/auth/users/{user_name}/products/{product_name}` | 为不存在的用户绑定 | 异常参数 | 验证 ErrNum=404 |
| AUTH-7-001 | 解除产品线 | DELETE | `/auth/users/{user_name}/products/{product_name}` | 正常解除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-7-002 | 解除产品线 | DELETE | `/auth/users/{user_name}/products/{product_name}` | 解除未绑定的产品线 | 异常参数 | 验证 ErrNum=404 |
| AUTH-8-001 | 按产品线查用户 | GET | `/auth/users/actions/search-by-product/{product_name}` | 绑定后查询 | 正常参数 | 验证返回列表包含该用户 |
| AUTH-8-002 | 按产品线查用户 | GET | `/auth/users/actions/search-by-product/{product_name}` | 无绑定用户的产品线 | 正常参数 | 验证返回空列表 |
| AUTH-9-001 | 创建Session Key | POST | `/auth/session-keys` | 正确用户名密码登录 | 正常参数 | 验证返回 session_key、user_name、is_admin |
| AUTH-9-002 | 创建Session Key | POST | `/auth/session-keys` | 密码错误 | 异常参数 | 验证 ErrNum=401 |
| AUTH-9-003 | 创建Session Key | POST | `/auth/session-keys` | 用户不存在 | 异常参数 | 验证 ErrNum=401 |
| AUTH-9-004 | 创建Session Key | POST | `/auth/session-keys` | 缺少 user_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-10-001 | 删除Session Key | DELETE | `/auth/session-keys/{session_key}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-10-002 | 删除Session Key | DELETE | `/auth/session-keys/{session_key}` | 删除不存在的 key | 异常参数 | 验证 ErrNum=404 |
| AUTH-11-001 | 创建Token | POST | `/auth/tokens` | 创建 Product scope Token | 正常参数 | 验证返回 token |
| AUTH-11-002 | 创建Token | POST | `/auth/tokens` | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-003 | 创建Token | POST | `/auth/tokens` | 缺少 scope | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-004 | 创建Token | POST | `/auth/tokens` | scope=Product 缺少 product_name | 必填校验 | 验证 ErrNum=422 |
| AUTH-11-005 | 创建Token | POST | `/auth/tokens` | 重复创建同名 Token | 业务规则 | 验证 ErrNum=555 |
| AUTH-12-001 | 删除Token | DELETE | `/auth/tokens/{token_name}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AUTH-12-002 | 删除Token | DELETE | `/auth/tokens/{token_name}` | 删除不存在的 Token | 异常参数 | 验证 ErrNum=404 |
| AUTH-13-001 | Token详情 | GET | `/auth/tokens/{token_name}` | 查询已存在的 Token | 正常参数 | 验证返回 name、token、scope、product_name |
| AUTH-13-002 | Token详情 | GET | `/auth/tokens/{token_name}` | 查询不存在的 Token | 异常参数 | 验证 ErrNum=404 |
| AUTH-14-001 | Token列表 | GET | `/auth/tokens` | 获取 Token 列表 | 正常参数 | 验证返回数组 |
| AUTH-14-002 | Token列表 | GET | `/auth/tokens` | 验证返回字段完整性 | 返回数据 | 验证每个元素包含 name、token、scope |
| AUTH-15-001 | 按产品线查Token | GET | `/auth/tokens/actions/search-by-product/{product_name}` | 查询有 Token 绑定的产品线 | 正常参数 | 验证返回 Token 列表 |
| AUTH-15-002 | 按产品线查Token | GET | `/auth/tokens/actions/search-by-product/{product_name}` | 查询无 Token 绑定的产品线 | 正常参数 | 验证返回空列表 |

---

#### 4.5.2 API-Key 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| AK-1-001 | 创建 | POST | `/api-keys` | 最小参数创建（仅 description） | 正常参数 | 验证返回完整字段 |
| AK-1-002 | 创建 | POST | `/api-keys` | 带配额计划创建 | 正常参数 | 验证 quota_plan 返回正确 |
| AK-1-003 | 创建 | POST | `/api-keys` | 带限流策略创建 | 正常参数 | 验证 rate_limit_policy 返回正确 |
| AK-1-004 | 创建 | POST | `/api-keys` | 挂载到 Entity | 正常参数 | 验证 entity_id 和 entity 返回正确 |
| AK-1-005 | 创建 | POST | `/api-keys` | expired_time=-1 永不过期 | 正常参数 | 验证返回 expired_time=-1 |
| AK-1-006 | 创建 | POST | `/api-keys` | models=["*"] 不限制模型 | 正常参数 | 验证返回 models=["*"] |
| AK-1-007 | 创建 | POST | `/api-keys` | subnet=["*"] 不限制子网 | 正常参数 | 验证返回 subnet=["*"] |
| AK-1-008 | 创建 | POST | `/api-keys` | 缺少 description | 必填校验 | 验证 ErrNum=422 |
| AK-1-009 | 创建 | POST | `/api-keys` | rate_limit_policy.enabled=true 但无规则 | 异常参数 | 验证 ErrNum=422 |
| AK-1-010 | 创建 | POST | `/api-keys` | entity_id 指向不存在的 Entity | 异常参数 | 验证 ErrNum=422 |
| AK-1-011 | 创建 | POST | `/api-keys` | description 为空字符串 | 边界值 | 验证 ErrNum=422 |
| AK-1-012 | 创建 | POST | `/api-keys` | 验证返回字段完整性 | 返回数据 | 验证 id、key、description、enabled、create_time、update_time、expired_time、unlimited_quota、models、subnet、quota_plan、rate_limit_policy、entity_id、entity |
| AK-2-001 | 列表 | GET | `/api-keys` | 默认分页查询 | 正常参数 | 验证返回 list 和 pagination |
| AK-2-002 | 列表 | GET | `/api-keys` | 指定分页参数 | 正常参数 | page=1&page_size=10 |
| AK-2-003 | 列表 | GET | `/api-keys` | 按 enabled=true 过滤 | 正常参数 | 验证结果正确 |
| AK-2-004 | 列表 | GET | `/api-keys` | 按 entity_id 过滤 | 正常参数 | 验证结果正确 |
| AK-2-005 | 列表 | GET | `/api-keys` | 按 unlimited_quota=true 过滤 | 正常参数 | 验证结果正确 |
| AK-2-006 | 列表 | GET | `/api-keys` | page_size=100（最大值） | 边界值 | 验证正常返回 |
| AK-2-007 | 列表 | GET | `/api-keys` | page_size=101（超最大值） | 边界值 | 验证退回最大值或报错 |
| AK-2-008 | 列表 | GET | `/api-keys` | 验证分页字段完整性 | 返回数据 | 验证 pagination 含 page、page_size、total |
| AK-3-001 | 详情 | GET | `/api-keys/{id}` | 查询存在的 API-Key | 正常参数 | 验证返回完整字段 |
| AK-3-002 | 详情 | GET | `/api-keys/{id}` | 查询不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |
| AK-4-001 | 全量更新 | PUT | `/api-keys/{id}` | 更新 description | 正常参数 | 验证返回 ErrNum=200 |
| AK-4-002 | 全量更新 | PUT | `/api-keys/{id}` | 更换 entity_id | 正常参数 | 验证 entity 信息更新 |
| AK-4-003 | 全量更新 | PUT | `/api-keys/{id}` | 更新不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |
| AK-4-004 | 全量更新 | PUT | `/api-keys/{id}` | 缺少 description | 必填校验 | 验证 ErrNum=422 |
| AK-5-001 | 部分更新 | PATCH | `/api-keys/{id}` | 仅修改 description | 正常参数 | 验证返回 ErrNum=200 |
| AK-5-002 | 部分更新 | PATCH | `/api-keys/{id}` | 禁用 API-Key | 正常参数 | enabled=false |
| AK-5-003 | 部分更新 | PATCH | `/api-keys/{id}` | 修改 quota_plan.quota | 正常参数 | 验证配额重置 |
| AK-6-001 | 删除 | DELETE | `/api-keys/{id}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| AK-6-002 | 删除 | DELETE | `/api-keys/{id}` | 删除后查询返回 404 | 业务规则 | 验证级联删除 |
| AK-6-003 | 删除 | DELETE | `/api-keys/{id}` | 删除不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |
| AK-7-001 | 查询配额 | GET | `/api-keys/{id}/quota-plan` | 查询有配额的 API-Key | 正常参数 | 验证返回 quota、balance 字段 |
| AK-7-002 | 查询配额 | GET | `/api-keys/{id}/quota-plan` | 查询无限配额 API-Key | 正常参数 | unlimited=true |
| AK-8-001 | 重置配额 | POST | `/api-keys/{id}/quota-plan/reset` | 传入 quota 重置 | 正常参数 | 验证 balance 重置 |
| AK-8-002 | 重置配额 | POST | `/api-keys/{id}/quota-plan/reset` | 不传 quota 重置 | 正常参数 | 按当前 quota 重置 |
| AK-8-003 | 重置配额 | POST | `/api-keys/{id}/quota-plan/reset` | 重置不存在的 API-Key | 异常参数 | 验证 ErrNum=404 |

---

#### 4.5.3 Entity-Type 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| ET-1-001 | 创建 | POST | `/entity-types` | 正常创建 level=1 | 正常参数 | 验证返回 ErrNum=200 |
| ET-1-002 | 创建 | POST | `/entity-types` | 正常创建 level=3 | 正常参数 | 验证返回 level=3 |
| ET-1-003 | 创建 | POST | `/entity-types` | 缺少 type_name | 必填校验 | 验证 ErrNum=422 |
| ET-1-004 | 创建 | POST | `/entity-types` | 缺少 level | 必填校验 | 验证 ErrNum=422 |
| ET-1-005 | 创建 | POST | `/entity-types` | 重复 type_name | 业务规则 | 验证 ErrNum=555 |
| ET-1-006 | 创建 | POST | `/entity-types` | level 超出范围（0） | 边界值 | 验证 ErrNum=422 |
| ET-1-007 | 创建 | POST | `/entity-types` | level 超出范围（6） | 边界值 | 验证 ErrNum=422 |
| ET-2-001 | 列表 | GET | `/entity-types` | 默认分页查询 | 正常参数 | 验证返回 list 和 pagination |
| ET-2-002 | 列表 | GET | `/entity-types` | 指定分页参数 | 正常参数 | page=1&page_size=10 |
| ET-3-001 | 详情 | GET | `/entity-types/{type_name}` | 查询存在的类型 | 正常参数 | 验证返回完整字段 |
| ET-3-002 | 详情 | GET | `/entity-types/{type_name}` | 查询不存在的类型 | 异常参数 | 验证 ErrNum=404 |
| ET-4-001 | 更新 | PATCH | `/entity-types/{type_name}` | 更新 description | 正常参数 | 验证返回 ErrNum=200 |
| ET-5-001 | 删除 | DELETE | `/entity-types/{type_name}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| ET-5-002 | 删除 | DELETE | `/entity-types/{type_name}` | 删除有 Entity 使用的类型 | 业务规则 | 验证 ErrNum=409 |

---

#### 4.5.4 Entity 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| E-1-001 | 创建 | POST | `/entities` | 正常创建根节点 | 正常参数 | 验证返回完整字段 |
| E-1-002 | 创建 | POST | `/entities` | 创建子实体 | 正常参数 | 带 parent_id |
| E-1-003 | 创建 | POST | `/entities` | 带配额计划创建 | 正常参数 | 验证 quota_plan 返回 |
| E-1-004 | 创建 | POST | `/entities` | 带限流策略创建 | 正常参数 | 验证 rate_limit_policy 返回 |
| E-1-005 | 创建 | POST | `/entities` | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| E-1-006 | 创建 | POST | `/entities` | 缺少 type | 必填校验 | 验证 ErrNum=422 |
| E-1-007 | 创建 | POST | `/entities` | type 不存在 | 异常参数 | 验证 ErrNum=422 |
| E-1-008 | 创建 | POST | `/entities` | parent_id 层级不合法 | 业务规则 | 子级 level 必须 > 父级 level |
| E-2-001 | 列表 | GET | `/entities` | 默认分页查询 | 正常参数 | 验证返回 list 和 pagination |
| E-2-002 | 列表 | GET | `/entities` | 按 type 过滤 | 正常参数 | 验证过滤正确 |
| E-2-003 | 列表 | GET | `/entities` | 按 parent_id 过滤 | 正常参数 | 验证过滤正确 |
| E-3-001 | 详情 | GET | `/entities/{id}` | 查询存在的 Entity | 正常参数 | 验证返回完整字段 |
| E-3-002 | 详情 | GET | `/entities/{id}` | 查询不存在的 Entity | 异常参数 | 验证 ErrNum=404 |
| E-4-001 | 全量更新 | PUT | `/entities/{id}` | 更新 name | 正常参数 | 验证返回 ErrNum=200 |
| E-4-002 | 全量更新 | PUT | `/entities/{id}` | 更新 parent_id | 正常参数 | 验证层级约束 |
| E-4-003 | 全量更新 | PUT | `/entities/{id}` | 修改 type（不允许） | 业务规则 | 验证 type 不可修改 |
| E-5-001 | 部分更新 | PATCH | `/entities/{id}` | 仅修改 allow_models | 正常参数 | 验证返回 ErrNum=200 |
| E-5-002 | 部分更新 | PATCH | `/entities/{id}` | 修改 quota_plan.quota | 正常参数 | 验证配额重置 |
| E-6-001 | 删除 | DELETE | `/entities/{id}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| E-6-002 | 删除 | DELETE | `/entities/{id}` | 删除有子 Entity 的 | 业务规则 | 验证 ErrNum=409 |
| E-6-003 | 删除 | DELETE | `/entities/{id}` | 删除被 API-Key 挂载的 | 业务规则 | 验证 ErrNum=409 |
| E-7-001 | 查询配额 | GET | `/entities/{id}/quota-plan` | 查询配额计划 | 正常参数 | 验证返回 quota、balance |

---

#### 4.5.5 BFE Pools 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| BP-1-001 | 获取实例池 | GET | `/alb-pool` | 获取默认实例池 | 正常参数 | 验证返回 name、instances |
| BP-2-001 | 更新实例池 | PATCH | `/alb-pool` | 全量更新实例列表 | 正常参数 | 验证返回 ErrNum=200 |
| BP-2-002 | 更新实例池 | PATCH | `/alb-pool` | 缺少 instances | 必填校验 | 验证 ErrNum=422 |
| BP-2-003 | 更新实例池 | PATCH | `/alb-pool` | 实例缺少 ip | 必填校验 | 验证 ErrNum=422 |
| BP-2-004 | 更新实例池 | PATCH | `/alb-pool` | weight 超出范围 | 边界值 | 验证 ErrNum=422 |

---

#### 4.5.6 Clusters 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| CL-1-001 | 创建 | POST | `/clusters` | 最小参数创建 | 正常参数 | 基本字段 + instance_pool |
| CL-1-002 | 创建 | POST | `/clusters` | 完整配置创建 | 正常参数 | 含 llm_config |
| CL-1-003 | 创建 | POST | `/clusters` | 重复集群名 | 业务规则 | 验证 ErrNum=555 |
| CL-1-004 | 创建 | POST | `/clusters` | 缺少 name | 必填校验 | 验证 ErrNum=422 |
| CL-1-005 | 创建 | POST | `/clusters` | 缺少 instance_pool | 必填校验 | 验证 ErrNum=422 |
| CL-2-001 | 列表 | GET | `/clusters` | 查询集群列表 | 正常参数 | 验证返回数组 |
| CL-3-001 | 详情 | GET | `/clusters/{name}` | 查询存在的集群 | 正常参数 | 验证返回完整字段 |
| CL-3-002 | 详情 | GET | `/clusters/{name}` | 查询不存在的集群 | 异常参数 | 验证 ErrNum=404 |
| CL-4-001 | 更新 | PUT | `/clusters/{name}` | 更新集群配置 | 正常参数 | 验证返回 ErrNum=200 |
| CL-5-001 | 删除 | DELETE | `/clusters/{name}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| CL-5-002 | 删除 | DELETE | `/clusters/{name}` | 删除不存在的集群 | 异常参数 | 验证 ErrNum=404 |
| CL-6-001 | 就绪检查 | GET | `/clusters/{name}/ready` | 检查就绪状态 | 正常参数 | 验证返回 ready 状态 |
| CL-7-001 | 模型列表 | POST | `/models` | 查询模型列表 | 正常参数 | 验证返回模型列表 |

---

#### 4.5.7 AI Route Rule 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| AR-1-001 | 设置规则 | PATCH | `/ai-route-rules` | 仅设置基础路由规则 | 正常参数 | 验证 basic_forward_rules 返回正确 |
| AR-1-002 | 设置规则 | PATCH | `/ai-route-rules` | 仅设置高级路由规则 | 正常参数 | 验证 forward_rules 返回正确 |
| AR-1-003 | 设置规则 | PATCH | `/ai-route-rules` | 同时设置两种规则 | 正常参数 | 验证两种规则均返回 |
| AR-1-004 | 设置规则 | PATCH | `/ai-route-rules` | 清空所有规则 | 正常参数 | 传空数组，验证仅剩 default_t() |
| AR-1-005 | 设置规则 | PATCH | `/ai-route-rules` | 设置多条高级路由规则 | 正常参数 | 验证 forward_rules 长度为 3 |
| AR-1-006 | 设置规则 | PATCH | `/ai-route-rules` | 设置多条基础路由规则 | 正常参数 | 验证 basic_forward_rules 长度为 2 |
| AR-1-007 | 设置规则 | PATCH | `/ai-route-rules` | 缺少 expression | 必填校验 | 验证 ErrNum=422 |
| AR-1-008 | 设置规则 | PATCH | `/ai-route-rules` | 缺少 cluster_name | 必填校验 | 验证 ErrNum=422 |
| AR-1-009 | 设置规则 | PATCH | `/ai-route-rules` | basic 缺少 cluster_name | 必填校验 | 验证 ErrNum=422 |
| AR-1-010 | 设置规则 | PATCH | `/ai-route-rules` | expression 为空字符串 | 边界值 | 验证 ErrNum=422 |
| AR-1-011 | 设置规则 | PATCH | `/ai-route-rules` | cluster_name 为空字符串 | 边界值 | 验证 ErrNum=422 |
| AR-1-012 | 设置规则 | PATCH | `/ai-route-rules` | basic cluster_name 为空 | 边界值 | 验证 ErrNum=422 |
| AR-1-013 | 设置规则 | PATCH | `/ai-route-rules` | forward_rules 元素为 null | 异常参数 | 验证 ErrNum=422 |
| AR-1-014 | 设置规则 | PATCH | `/ai-route-rules` | basic_forward_rules 元素为 null | 异常参数 | 验证 ErrNum=422 |
| AR-1-015 | 设置规则 | PATCH | `/ai-route-rules` | 空 Body | 边界值 | 验证 ErrNum=422 |
| AR-1-016 | 设置规则 | PATCH | `/ai-route-rules` | 非法 JSON Body | 异常参数 | 发送非 JSON 文本 |
| AR-1-017 | 设置规则 | PATCH | `/ai-route-rules` | description 可选 | 可选字段 | 不传 description 验证正常 |
| AR-1-018 | 设置规则 | PATCH | `/ai-route-rules` | host_names 可选 | 可选字段 | 不传 host_names 验证正常 |
| AR-1-019 | 设置规则 | PATCH | `/ai-route-rules` | paths 可选 | 可选字段 | 不传 paths 验证正常 |
| AR-1-020 | 设置规则 | PATCH | `/ai-route-rules` | 返回数据镜像请求 | 返回数据校验 | 逐字段验证返回与请求一致 |
| AR-1-021 | 设置规则 | PATCH | `/ai-route-rules` | 最后一条自动追加 default_t() | 业务规则 | GET 验证返回包含 default_t() |
| AR-2-001 | 获取规则 | GET | `/ai-route-rules` | 设置后获取规则 | 正常参数 | 验证返回完整规则 |
| AR-2-002 | 获取规则 | GET | `/ai-route-rules` | 未设置时获取空规则 | 正常参数 | 验证返回空结果 |
| AR-2-003 | 获取规则 | GET | `/ai-route-rules` | 验证返回数据结构 | 返回数据校验 | 验证 forward_rules 和 basic_forward_rules 字段 |

---

#### 4.5.8 Certificate 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| CERT-1-001 | 创建 | POST | `/certificates` | 正常创建证书 | 正常参数 | 验证返回 ErrNum=200 |
| CERT-1-002 | 创建 | POST | `/certificates` | 重复 cert_name | 业务规则 | 验证 ErrNum=555 |
| CERT-1-003 | 创建 | POST | `/certificates` | 缺少 cert_name | 必填校验 | 验证 ErrNum=422 |
| CERT-1-004 | 创建 | POST | `/certificates` | 缺少 cert_file_content | 必填校验 | 验证 ErrNum=422 |
| CERT-1-005 | 创建 | POST | `/certificates` | 缺少 key_file_content | 必填校验 | 验证 ErrNum=422 |
| CERT-1-006 | 创建 | POST | `/certificates` | 无效 PEM 格式 | 异常参数 | 验证 ErrNum=422 |
| CERT-2-001 | 列表 | GET | `/certificates` | 获取证书列表 | 正常参数 | 验证返回数组 |
| CERT-3-001 | 详情 | GET | `/certificates/{cert_name}` | 查询存在的证书 | 正常参数 | 验证返回完整字段 |
| CERT-3-002 | 详情 | GET | `/certificates/{cert_name}` | 查询不存在的证书 | 异常参数 | 验证 ErrNum=404 |
| CERT-4-001 | 设置默认 | PATCH | `/certificates/{cert_name}/default` | 设为默认证书 | 正常参数 | 验证返回 ErrNum=200 |
| CERT-5-001 | 删除 | DELETE | `/certificates/{cert_name}` | 正常删除 | 正常参数 | 验证 ErrNum=200 |
| CERT-5-002 | 删除 | DELETE | `/certificates/{cert_name}` | 删除默认证书 | 业务规则 | 验证 ErrNum=409 |

---

#### 4.5.9 InnerAPI 模块

| 编号 | 接口 | 方法 | 路径 | 场景 | 测试类型 | 测试内容 |
|------|------|------|------|------|---------|---------|
| INNER-1-001 | API-Key导出 | GET | `/configs/mod-api-key` | 首次拉取（不传 version） | 正常参数 | 返回完整配置 |
| INNER-1-002 | API-Key导出 | GET | `/configs/mod-api-key` | 增量拉取（相同 version） | 正常参数 | 返回 Data=null |
| INNER-1-003 | API-Key导出 | GET | `/configs/mod-api-key` | 配置变化后重新拉取 | 业务规则 | 返回新配置 |
| INNER-2-001 | 限流策略导出 | GET | `/configs/rate-limit-policy` | 导出限流策略 | 正常参数 | 验证返回 RateLimitPolicies |
| INNER-3-001 | TLS配置导出 | GET | `/configs/tls_conf/server_data_conf` | 导出 TLS 配置 | 正常参数 | 验证返回配置 |
| INNER-4-001 | GSLB配置导出 | GET | `/configs/gslb_data/gslb` | 导出 GSLB 配置 | 正常参数 | 验证返回配置 |
| INNER-5-001 | 证书配置导出 | GET | `/configs/protocol/server_cert_conf` | 导出证书配置 | 正常参数 | 验证返回配置 |
| INNER-6-001 | 额外文件导出 | GET | `/configs/extra_file` | 导出额外文件 | 正常参数 | 验证返回配置 |
| INNER-7-001 | Body处理导出 | GET | `/configs/mod-body-process` | 导出 Body 处理模块 | 正常参数 | 验证返回配置 |

---

### 4.5 测试用例统计

| 模块 | 接口数 | 用例目录数 | 测试用例数 |
|------|--------|-----------|-----------|
| AUTH 认证 | 15 | 15 | 43 |
| AK API-Key | 8 | 8 | 33 |
| ET Entity-Type | 5 | 5 | 14 |
| E Entity | 7 | 7 | 22 |
| BP 实例池 | 2 | 2 | 5 |
| CL 集群 | 7 | 7 | 13 |
| AR AI 路由 | 2 | 2 | 24 |
| CERT 证书 | 5 | 5 | 12 |
| InnerAPI | 7 | 7 | 9 |
| **总计** | **58** | **58** | **175** |

---

## 五、测试执行流程

### 5.1 前置条件：编译项目二进制

```bash
# 在 ai-gateway-api 项目根目录编译（必须执行，生成 ai-gateway-api.exe）
cd ai-gateway-api
make build
# 或者: go build -o ai-gateway-api.exe .
```

### 5.2 运行全部测试

```bash
cd ai-gateway-api/test/test-runner
go test -v ./test-cases/...
```

### 5.3 运行指定模块测试

```bash
# 运行 AI 路由模块全部测试
go test -v ./test-cases/ai_route/...

# 运行 API-Key 模块全部测试
go test -v ./test-cases/api_key/...
```

### 5.4 运行单个用例

```bash
# 运行 AI 路由 - 设置规则测试
go test -v ./test-cases/ai_route/set_rules/

# 运行指定测试函数
go test -v -run TestSetRules_OnlyBasicRules ./test-cases/ai_route/set_rules/
```

### 5.5 生成测试报告

测试报告统一保存在 `ai-gateway-api/test/test-reports/` 目录下，每次运行生成一个以时间戳（`YYYY-MM-DD_HHMMSS`）命名的子目录，包含按模块拆分的 Markdown 报告文件。

#### 5.5.1 报告生成命令

```bash
# 在 test-runner 目录下运行测试，生成 JSON 原始结果
cd ai-gateway-api/test/test-runner
go test -json ./test-cases/... > ../test-reports/temp_output.json

# 执行报告生成脚本（将 JSON 转换为 Markdown 报告）
python ../generate_report.py

# 报告输出到 test-reports/{timestamp}/ 目录
```

#### 5.5.2 报告目录结构

```
test-reports/
└── 2026-07-23_143000/              # 时间戳目录
    ├── summary.md                  # 汇总报告
    ├── auth.md                     # Auth 模块
    ├── api_key.md                  # API-Key 模块
    ├── entity_type.md              # Entity-Type 模块
    ├── entity.md                   # Entity 模块
    ├── bfe_pools.md                # 实例池模块
    ├── clusters.md                 # 集群模块
    ├── ai_route.md                 # AI 路由模块
    ├── certificate.md              # 证书模块
    └── innerapi.md                 # InnerAPI 模块
```

#### 5.5.3 汇总报告格式（summary.md）

```markdown
# AI Gateway API 测试报告

**测试时间**：2026-07-23 14:30:00  
**总用例数**：158  
**通过数**：152  
**失败数**：6  
**跳过数**：0  
**通过率**：96.2%  

---

## 模块统计

| 模块 | 总用例 | 通过 | 失败 | 通过率 |
|------|--------|------|------|--------|
| AUTH 认证 | 43 | 43 | 0 | 100% |
| AK API-Key | 33 | 33 | 0 | 100% |
| ET Entity-Type | 14 | 14 | 0 | 100% |
| E Entity | 22 | 20 | 2 | 90.9% |
| BP 实例池 | 5 | 5 | 0 | 100% |
| CL 集群 | 13 | 11 | 2 | 84.6% |
| AR AI 路由 | 7 | 7 | 0 | 100% |
| CERT 证书 | 12 | 12 | 0 | 100% |
| InnerAPI | 9 | 7 | 2 | 77.8% |

---

## 失败用例汇总

### E-6-002：删除有子 Entity 的 — 应返回 409

| 项目 | 值 |
|------|-----|
| 模块 | Entity |
| 接口 | DELETE /entities/{id} |
| 用例编号 | E-6-002 |
| 测试类型 | 业务规则 |
| 场景 | 删除有子 Entity 的父节点 |
| 实际结果 | ErrNum=500, ErrMsg="Unknown Exception" |
| 预期结果 | ErrNum=409 |
| 可能原因 | 删除前未校验子 Entity 存在性，SQLite 外键约束触发 panic |

### CL-7-001：查询模型列表 — 返回 404

| 项目 | 值 |
|------|-----|
| 模块 | Clusters |
| 接口 | POST /models |
| 用例编号 | CL-7-001 |
| 测试类型 | 正常参数 |
| 场景 | 查询模型列表 |
| 实际结果 | ErrNum=404 |
| 预期结果 | ErrNum=200 |
| 可能原因 | models 接口可能需要特定的 provider 参数或依赖已有集群配置 |

---

## 详细报告

各模块详细报告见：
- [auth.md](./auth.md)
- [api_key.md](./api_key.md)
- [entity_type.md](./entity_type.md)
- [entity.md](./entity.md)
- [bfe_pools.md](./bfe_pools.md)
- [clusters.md](./clusters.md)
- [ai_route.md](./ai_route.md)
- [certificate.md](./certificate.md)
- [innerapi.md](./innerapi.md)
```

#### 5.5.4 模块报告格式（以 api_key.md 为例）

```markdown
# API-Key 模块测试报告

**模块**：API-Key  
**总用例数**：33  
**通过数**：33  
**失败数**：0  
**通过率**：100%  

---

## 用例列表

| 编号 | 接口 | 场景 | 状态 |
|------|------|------|------|
| AK-1-001 | POST /api-keys | 最小参数创建 | PASS |
| AK-1-002 | POST /api-keys | 带配额计划创建 | PASS |
| AK-1-003 | POST /api-keys | 带限流策略创建 | PASS |
| ... | ... | ... | ... |

---

## 用例详细结果

---

### AK-1-001：最小参数创建（仅 description） — ✅ PASS

**接口**：`POST /open-api/v1/api-keys`  
**测试类型**：正常参数  
**执行时间**：0.045s  

#### 请求参数

```json
{
    "description": "test-key-001"
}
```

#### 实际返回

**HTTP Status**：200  
**ErrNum**：200  
**ErrMsg**：success  
**WorkMode**：ModeNormal  

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "ak-abcd1234",
        "key": "ak-xxxxxxxxxxxx",
        "description": "test-key-001",
        "enabled": true,
        "create_time": 1751963400,
        "update_time": 1751963400,
        "expired_time": -1,
        "unlimited_quota": false,
        "models": ["*"],
        "subnet": ["*"],
        "quota_plan": {
            "unlimited": false,
            "pass_when_no_enough_quota": false,
            "quota": 100000000,
            "unit": "total_token",
            "reset_period": "monthly"
        },
        "rate_limit_policy": {
            "enabled": false
        },
        "entity_id": "",
        "entity": null
    },
    "WorkMode": "ModeNormal"
}
```

#### 预期 vs 实际对比

| 校验项 | 预期值 | 实际值 | 结果 |
|--------|--------|--------|------|
| ErrNum | 200 | 200 | ✅ |
| Data.id 非空 | 非空 | "ak-abcd1234" | ✅ |
| Data.key 非空 | 非空 | "ak-xxxxxxxxxxxx" | ✅ |
| Data.description | "test-key-001" | "test-key-001" | ✅ |
| Data.enabled | true | true | ✅ |
| Data.expired_time | -1 | -1 | ✅ |
| Data.unlimited_quota | false | false | ✅ |
| Data.models | ["*"] | ["*"] | ✅ |
| Data.subnet | ["*"] | ["*"] | ✅ |
| Data.quota_plan | 非 null | {...} | ✅ |
| Data.rate_limit_policy | 非 null | {...} | ✅ |
| Data.entity_id | "" | "" | ✅ |
| Data.entity | null | null | ✅ |

---

### AK-1-008：缺少 description — ❌ FAIL

**接口**：`POST /open-api/v1/api-keys`  
**测试类型**：必填校验  
**执行时间**：0.012s  

#### 请求参数

```json
{}
```

#### 实际返回

**HTTP Status**：200  
**ErrNum**：200  
**ErrMsg**：success  

```json
{
    "ErrNum": 200,
    "ErrMsg": "success",
    "Data": {
        "id": "ak-xxxxyyyy",
        "description": "",
        ...
    },
    "WorkMode": "ModeNormal"
}
```

#### 预期 vs 实际对比

| 校验项 | 预期值 | 实际值 | 结果 |
|--------|--------|--------|------|
| ErrNum | 422 | 200 | ❌ `description` 字段未做必填校验，创建了空描述的 API-Key |

#### 建议修复方案

**文件**：`ai-gateway-api/endpoints/openapi_v1/api_key/create.go`  
**建议**：在 `CheckParams` 中增加 `description` 非空校验：

```go
if param.Description == nil || *param.Description == "" {
    return nil, xerror.WrapParamErrorWithMsg("description is required")
}
```

---

...(其余用例按同样格式)
```

#### 5.5.5 失败用例的修复建议

对于每个失败的测试用例，报告中必须包含：

| 字段 | 说明 |
|------|------|
| **实际返回** | 完整的 HTTP 响应（Status、ErrNum、ErrMsg、Data） |
| **预期对比** | 逐字段对比预期值和实际值，标注不匹配项 |
| **建议修复方案** | 指向具体的代码文件和行号，给出修改建议或示例代码 |

这样大模型可以直接读取报告，定位到失败用例的具体差异，并据此修复 `ai-gateway-api` 接口代码中的问题。

#### 5.5.6 报告生成脚本设计

`run_test_and_report.py` 脚本流程：

```
1. 执行 go test -json 获取原始测试结果
2. 解析 JSON 输出，为每个测试用例提取：
   - TestCaseID（从测试函数名解析）
   - Action（pass/fail/skip）
   - Elapsed（执行时间）
   - RequestLog（请求参数和返回数据，需在测试代码中写入）
3. 按模块分组，生成 summary.md 和各模块 .md 报告
4. 在 summary.md 中汇总失败用例，生成修复建议
5. 输出到 test-reports/{timestamp}/ 目录
```

**测试代码中的日志记录**：

为让报告包含详细的请求参数和返回数据，测试代码中需使用 `RequestLog` 记录每次 API 调用的请求和响应：

```go
func TestAPIKey_CreateMinimal(t *testing.T) {
    reqBody := map[string]interface{}{
        "description": "test-key-001",
    }
    resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", reqBody)
    
    // 记录请求日志，供报告生成使用
    testutil.RecordRequestLog(t, "AK-1-001", reqBody, resp)
    
    assert.NoError(t, err)
    testutil.AssertSuccess(t, resp)
    // ... 字段校验
}
```

---

## 六、注意事项

### 6.1 编译前置条件

- 运行测试前必须先执行 `make build`（或 `go build`）编译 `ai-gateway-api.exe`
- 测试运行时会自动查找项目根目录下的 `ai-gateway-api.exe`，复制到 `data/` 目录后启动

### 6.2 数据库隔离

- 每个测试进程使用独立的数据库文件（`test_ai_gateway_{pid}.db`）
- 测试完成后自动清理数据库文件和 WAL/SHM 文件
- 避免测试间数据污染

### 6.3 认证跳过

- 测试环境配置 `SkipTokenValidate=true`
- 中间件 `UserProbeAction` 在认证头为空时直接放行
- `MustGetVisitor` 返回假 visitor（ScopeSystem）

### 6.4 SQLite 兼容性

- SQLite 不支持 `SELECT ... FOR UPDATE`，通过自定义驱动 `sqlite-strip` 解决
- `sqlite-strip` 驱动位于主项目 `stateful/sqlite_strip.go`，编译时自动包含
- SQLite 不支持外键级联删除，需在代码层面处理

### 6.5 子进程管理

- 通过 `exec.CommandContext` 管理子进程生命周期，context 取消时自动终止
- 启动 goroutine 消费 stdout 输出，防止子进程写管道阻塞
- stderr 输出缓存，启动失败时用于调试

### 6.6 用例目录独立性

- 设计文档（`design.md`）和测试代码（`*_test.go`）分别位于 `test-case-docs/` 和 `test-cases/` 目录
- 两目录保持相同的层级结构，便于交叉引用
- 用例目录之间相互独立，无交叉依赖
- 新增用例时，在 `test-case-docs/` 对应模块下创建接口目录和 `design.md`，在 `test-cases/` 对应模块下创建接口目录和 `*_test.go`

---

## 七、附录

### 7.1 目录结构汇总

```
ai-gateway-api/test/
├── README.md
├── docs/
│   └── local-test-design.md              # 测试设计方案（本文档）
│
├── test-runner/                          # 测试环境
│   ├── go.mod / go.sum
│   ├── conf/
│   │   ├── ai_gateway_api.toml           # SQLite(sqlite-strip) + SkipTokenValidate
│   │   ├── nav_tree.toml
│   │   └── i18n/
│   │       └── zh.toml
│   ├── data/                             # 运行时数据（二进制、日志、DB）
│   │   └── .gitkeep
│   ├── testutil/
│   │   ├── server.go                     # 子进程服务器管理
│   │   ├── client.go                     # HTTP 客户端
│   │   ├── assert.go                     # 断言函数
│   │   ├── fixture.go                    # 测试数据工厂
│   │   ├── db.go                         # 数据库初始化/清理
│   │   └── mock_redis.go                 # Mock Redis（备用）
│   │
│   ├── test-case-docs/                   # 测试用例设计文档
│   │   ├── README.md                     # 设计文档总览 + 模板
│   │   ├── api_key/                      # API-Key 模块（8个接口）
│   │   │   ├── README.md
│   │   │   ├── create/design.md
│   │   │   ├── list/design.md
│   │   │   ├── detail/design.md
│   │   │   ├── full_update/design.md
│   │   │   ├── partial_update/design.md
│   │   │   ├── delete/design.md
│   │   │   ├── quota_query/design.md
│   │   │   └── quota_reset/design.md
│   │   └── ai_route/                     # AI 路由模块（2个接口）
│   │       ├── README.md
│   │       ├── set_rules/design.md
│   │       └── get_rules/design.md
│   │
│   └── test-cases/                       # Go 测试代码
│       ├── auth/                         # 认证模块（15个用例目录）
│       │   ├── create_user/              # {create_user_test.go}
│       │   ├── delete_user/
│       │   ├── reset_password/
│       │   ├── list_users/
│       │   ├── set_admin/
│       │   ├── bind_product/
│       │   ├── unbind_product/
│       │   ├── search_by_product/
│       │   ├── create_session_key/
│       │   ├── delete_session_key/
│       │   ├── create_token/
│       │   ├── delete_token/
│       │   ├── token_detail/
│       │   ├── token_list/
│       │   └── search_token_by_product/
│       │
│       ├── api_key/                      # API-Key 模块（8个用例目录）
│       │   ├── create/
│       │   ├── list/
│       │   ├── detail/
│       │   ├── full_update/
│       │   ├── partial_update/
│       │   ├── delete/
│       │   ├── quota_query/
│       │   └── quota_reset/
│       │
│       ├── entity_type/                  # Entity-Type 模块（5个用例目录）
│       │   ├── create/
│       │   ├── list/
│       │   ├── detail/
│       │   ├── update/
│       │   └── delete/
│       │
│       ├── entity/                       # Entity 模块（7个用例目录）
│       │   ├── create/
│       │   ├── list/
│       │   ├── detail/
│       │   ├── full_update/
│       │   ├── partial_update/
│       │   ├── delete/
│       │   └── quota_reset/
│       │
│       ├── bfe_pools/                    # 实例池模块（2个用例目录）
│       │   ├── get_pool/
│       │   └── update_instance/
│       │
│       ├── clusters/                     # 集群模块（7个用例目录）
│       │   ├── create/
│       │   ├── list/
│       │   ├── detail/
│       │   ├── update/
│       │   ├── delete/
│       │   ├── ready_check/
│       │   └── model_list/
│       │
│       ├── ai_route/                     # AI 路由模块（2个用例目录）
│       │   ├── set_rules/
│       │   └── get_rules/
│       │
│       ├── certificate/                  # 证书模块（5个用例目录）
│       │   ├── create/
│       │   ├── list/
│       │   ├── detail/
│       │   ├── update/
│       │   └── delete/
│       │
│       ├── model_provider/               # 模型提供商模块（1个用例目录）
│       │   └── list_models/
│       │
│       └── innerapi/                     # InnerAPI 模块（7个用例目录）
│           ├── mod_api_key/
│           ├── rate_limit_policy/
│           ├── server_data/
│           ├── gslb_data/
│           ├── protocol/
│           ├── extra_file/
│           └── mod_body_process/
│
├── test-reports/                         # 测试报告
│   └── YYYY-MM-DD_HHMMSS/
│       ├── summary.md
│       ├── ai_route.md
│       └── ...
│
└── generate_report.py                    # 报告生成脚本
```

### 7.2 相关文档

| 文档 | 路径 | 说明 |
|------|------|------|
| 测试使用说明 | `test/test-runner/README.md` | test-runner 详细使用说明 |
| OpenAPI 接口文档 | `docs/zh_cn/open_api/` | 各模块接口定义 |
| InnerAPI 接口设计 | `docs/瑛菲AI网关配额控制与限流-InnerAPI接口设计.md` | InnerAPI 接口详细设计 |

---

**文档结束**