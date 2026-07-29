# AI Gateway API 单元测试规范

本文档定义 `ai-gateway-api` 模块内单元测试的组织方式、mock 模式及代码规范，供后续各模块参考。

## 1. 总体原则

- **测试文件与被测代码同包**：`xxx.go` 的测试写在 `xxx_test.go`，放在同一目录。
- **不依赖外部服务**：单元测试不应需要 MySQL、Redis、真实配置文件或启动二进制。
- **优先测业务逻辑**：Manager 层是单元测试重点；storage 层 SQL 执行、全局初始化流程用集成测试覆盖。
- **测试即文档**：测试函数名和 `t.Run` 描述要清晰表达场景，不另写模块级测试说明。

## 2. 目录与命名

```text
ai-gateway-api/model/quota/
├── entity_manager.go              # 被测代码
├── entity_manager_test.go         # 单元测试
├── mocks_test.go                  # 模块内共享 mock
└── ...
```

命名约定：

| 类型 | 命名 |
|---|---|
| 单元测试文件 | `<name>_test.go` |
| 测试函数 | `TestXxx` 或 `TestXxx_场景` |
| 子测试 | `t.Run("具体场景描述", ...)` |
| 共享 mock 文件 | `mocks_test.go` |

## 3. 依赖

主模块使用 `github.com/stretchr/testify` 做断言：

```bash
go get github.com/stretchr/testify
```

推荐用法：

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

require.NoError(t, err)          // 后续断言依赖该条件
assert.Equal(t, expected, actual) // 非致命断言
```

## 4. Mock 模式

### 4.1 事务 Mock

所有 Manager 通过 `itxn.TxnStorager` 控制事务。单元测试使用 `fakeTxn` 直接执行回调，不开启真实事务：

```go
// mocks_test.go
type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
    return do(ctx)
}
```

### 4.2 Storager Mock

采用**手写 callback mock**，按接口方法提供 `xxxFn` 字段，便于每个测试定制行为：

```go
type fakeEntityStorager struct {
    createFn func(ctx context.Context, param *EntityParam) (int64, error)
    fetchFn  func(ctx context.Context, filter *EntityFilter) (*EntityParam, error)
    // ...

    // 可选：记录调用入参，便于断言
    created []*EntityParam
}

func (s *fakeEntityStorager) CreateEntity(ctx context.Context, param *EntityParam) (int64, error) {
    s.created = append(s.created, param)
    if s.createFn != nil {
        return s.createFn(ctx, param)
    }
    return 0, nil
}
```

这样测试中可以：

```go
store := &fakeEntityStorager{
    fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
        return nil, nil // 模拟不存在
    },
}
```

### 4.3 全局依赖处理

- **Redis**：尽量不直接依赖。若必须（如 `BalanceSyncManager`），在测试中临时替换 `stateful.DefaultClientSet` 为 `stateful.NewMockRedisClient()`，并在测试结束后恢复：

  ```go
  orig := stateful.DefaultClientSet
  defer func() { stateful.DefaultClientSet = orig }()
  stateful.DefaultClientSet = &stateful.ClientSet{RedisClient: stateful.NewMockRedisClient()}
  ```

- **全局 Config**：若被测代码读取 `stateful.DefaultConfig`，优先通过重构注入；若无法重构，可在测试中做最小化初始化。

## 5. 测试写法示例

```go
func TestEntityManager_CreateEntity(t *testing.T) {
    ctx := context.Background()

    t.Run("success with all associated data", func(t *testing.T) {
        entityStore := &fakeEntityStorager{
            fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
                return nil, nil // no duplicate
            },
            createFn: func(ctx context.Context, param *EntityParam) (int64, error) {
                return 100, nil
            },
        }
        m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{},
            &fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{},
            &fakeRouteRulesStorager{}, &fakeQuotaBalanceStorager{})

        id, err := m.CreateEntity(ctx, &EntityParam{
            EntityID: lib.PString("ent-1"),
            Name:     lib.PString("entity-one"),
            Type:     lib.PString("tenant"),
        })

        require.NoError(t, err)
        assert.Equal(t, int64(100), id)
    })

    t.Run("duplicate name", func(t *testing.T) {
        entityStore := &fakeEntityStorager{
            fetchFn: func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) {
                return &EntityParam{Name: filter.Name}, nil
            },
        }
        m := NewEntityManager(&fakeTxn{}, entityStore, /* ... */)

        _, err := m.CreateEntity(ctx, &EntityParam{Name: lib.PString("entity-one")})
        require.Error(t, err)
        assert.Contains(t, err.Error(), "Entity Record Existed")
    })
}
```

## 6. 运行测试

```bash
# 全部单元测试
go test ./...

# 仅 model 层单元测试
go test ./model/...

# 指定包
go test ./model/quota/...

# 带覆盖率
go test ./model/quota/... -cover

# 生成覆盖率报告
go test ./model/quota/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# model 层测试 + 覆盖率门禁（阈值 70%，可在 Makefile 中调整）
make test-model-cover-gate
```

> `coverage.out` 是生成文件，不应提交到仓库。

## 6.1 CI 集成

项目已配置 GitHub Actions 工作流 `.github/workflows/model-ci.yml`：

- 在 `push` / `pull_request` 到 `main` / `master` 时触发。
- 执行 `go test ./model/... -count=1`。
- 执行 `make test-model-cover-gate`，当 `model/` 整体语句覆盖率低于 70% 时失败。

如需调整阈值，修改 `Makefile` 中的 `MODEL_COVERAGE_THRESHOLD`。

## 7. 覆盖率目标

- Manager 层核心模块：≥ 70%
- 工具/纯函数模块：≥ 80%
- 不强制要求 100%，但关键分支（错误处理、边界条件）必须覆盖。

## 8. 不适合写单元测试的情况

以下情况建议用集成测试覆盖，或先重构再补单元测试：

- 直接操作全局单例且无法注入（如 `stateful.DefaultConfig`、`stateful.DefaultClientSet` 无法替换的代码）。
- 依赖真实 DB 执行的 SQL（`storage/rdb/*` 的 DAO 实现）。
- 涉及 goroutine、`time.Ticker`、全局 logger 的调度器（如 `QuotaResetScheduler`）。
- 仅做参数绑定/HTTP 路由的 endpoint handler（建议用 `httptest` + mock Manager 测，但通常需要重构 handler 才能注入 mock）。

## 9. 参考样例

当前 `model/` 下全部可测模块均已建立单元测试，`go test ./model/...` 整体语句覆盖率为 **87.6%**。

各模块覆盖率（`go test -cover`）：

| 模块 | 覆盖率 | 说明 |
|---|---|---|
| `model/quota` | 80.7% | 多依赖 Manager、Redis 同步、复杂配置生成 |
| `model/iai_route` | 96.6% | AI 路由条件构建与校验 |
| `model/shared/route_rules.go` | 96.2% | 路由规则校验工具 |
| `model/iauth` | 89.1% | 认证、授权、Feature 权限 |
| `model/iversion_control` | 100.0% | 版本控制与导出签名 |
| `model/ibasic` | 100.0% | Product、BFECluster、ExtraFile |
| `model/iprotocol` | 92.9% | 证书管理与导出 |
| `model/iroute_conf` | 92.4% | 域名、路由规则、导出 |
| `model/icluster_conf` | 83.5% | Cluster、Pool、SubCluster、APIKey、导出 |
| `model/imods` | 92.2% | APIKey Rule、Body Process、AI Route 导出 |

> `model/itxn` 仅包含接口定义，无需单元测试。

推荐按模块查看对应测试文件：

- `model/quota/mocks_test.go`：共享 mock 定义
- `model/quota/entity_manager_test.go`：多依赖 Manager 测试
- `model/quota/rate_limit_policy_manager_test.go`：复杂配置生成逻辑测试
- `model/quota/balance_sync_test.go`：Redis 相关逻辑测试（含全局状态恢复）
- `model/iauth/mocks_test.go` / `authentication_test.go`：事务 + storager mock 典型用法
- `model/ibasic/mocks_test.go`：小模块全覆盖示例
- `model/iroute_conf/mocks_test.go`：跨包接口 mock（`icluster_conf.ClusterStorager`）
- `model/icluster_conf/cluster_test.go` / `exporter_test.go`：复杂结构体转换与导出测试
- `model/imods/exporter_test.go`：多层实体配额/模型继承与导出测试
