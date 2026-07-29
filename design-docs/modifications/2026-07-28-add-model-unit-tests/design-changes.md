# 新增 model 层单元测试与 CI 覆盖率门禁（2026-07-28）

## 1. 概述

### 1.1 变更背景

`ai-gateway-api` 此前仅在 `test/` 目录下存在集成测试，缺少针对 `model/` 各业务模块的单元测试。随着业务逻辑复杂度增加，手动回归成本上升，亟需建立单元测试体系以保障核心模块（认证、配额、路由、集群等）的可维护性。

### 1.2 目标版本

| 项目 | 说明 |
|------|------|
| 变更日期 | 2026-07-28 |
| 目标版本 | 在现有主分支基础上新增单元测试与 CI 门禁 |
| 涉及范围 | `model/*` 全部可测模块、CI 工作流、测试规范文档 |

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **不依赖外部服务** | 单元测试不连接 MySQL、Redis、真实配置文件或启动二进制。 |
| **同包测试** | 测试文件与被测代码同包，命名为 `xxx_test.go`。 |
| **手写 mock** | 对 `itxn.TxnStorager` 及各模块 `Storager` 接口使用 callback 风格手写 mock，不引入额外 mock 框架。 |
| **测试即文档** | 测试函数与 `t.Run` 子测试名称清晰描述场景。 |
| **最小侵入** | 不修改 OpenAPI 接口、不修改数据库存储结构、不修改生产代码行为。 |

---

## 2. 模块/子系统设计变更

### 2.1 新增单元测试模块

为 `model/` 下全部可测模块建立单元测试，`itxn` 仅含接口定义无需测试：

| 模块 | 覆盖率 | 说明 |
|------|--------|------|
| `model/quota` | 80.7% | Entity、EntityType、QuotaPlan、RateLimitPolicy、BalanceSync 等 |
| `model/iai_route` | 96.6% | AI 路由条件构建与校验 |
| `model/shared/route_rules.go` | 96.2% | 全局路由规则校验 |
| `model/iauth` | 89.1% | 认证、授权、Feature 权限、Token/Session/User 管理 |
| `model/iversion_control` | 100.0% | 版本计算、导出签名、版本控制管理器 |
| `model/ibasic` | 100.0% | Product、BFECluster、ExtraFile |
| `model/iprotocol` | 92.9% | 证书管理、证书导出 |
| `model/iroute_conf` | 92.4% | Domain、RouteRule、路由导出 |
| `model/icluster_conf` | 83.5% | Cluster、Pool、SubCluster、APIKey、集群导出 |
| `model/imods` | 92.2% | APIKey Rule、Body Process、AI Route 导出 |

整体 `model/` 语句覆盖率达到 **87.6%**。

### 2.2 Mock 设计

每个被测模块新增 `mocks_test.go`，统一提供：

- `fakeTxn`：直接执行回调，绕过真实事务：
  ```go
  type fakeTxn struct{}
  func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
      return do(ctx)
  }
  ```
- 各 `Storager` 接口的 callback mock，例如 `fakeClusterStorager`、`fakeAPIKeyStorager` 等。
- 跨包依赖接口的 mock，如 `iroute_conf` 中 mock `icluster_conf.ClusterStorager`。

### 2.3 CI 与覆盖率门禁

#### 2.3.1 GitHub Actions 工作流

新增 `.github/workflows/model-ci.yml`：

- 触发条件：`push` / `pull_request` 到 `main` / `master`
- 执行步骤：
  1. `go mod download`
  2. `go test ./model/... -count=1`
  3. `make test-model-cover-gate`

#### 2.3.2 Makefile 目标

在 `Makefile` 中新增：

```makefile
MODEL_COVERAGE_THRESHOLD := 70

test-model:
	$(GOTEST) -v ./model/...

test-model-cover:
	$(GOTEST) -cover ./model/...

test-model-cover-gate:
	@echo "Running model tests with coverage gate ($(MODEL_COVERAGE_THRESHOLD)%)..."
	@$(GOTEST) -coverprofile=$(HOMEDIR)/model_coverage.out ./model/... >/dev/null && \
	COV=$$(go tool cover -func=$(HOMEDIR)/model_coverage.out | grep '^total:' | awk '{print $$3}' | sed 's/%//') && \
	echo "Model coverage: $$COV%" && \
	if awk -v cov="$$COV" -v threshold="$(MODEL_COVERAGE_THRESHOLD)" 'BEGIN {exit (cov+0 >= threshold+0) ? 0 : 1}'; then \
		echo "Coverage gate passed"; \
	else \
		echo "Model coverage $$COV% is below threshold $(MODEL_COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi
```

默认阈值为 **70%**，可通过 `MODEL_COVERAGE_THRESHOLD` 调整。

---

## 3. 涉及文件

### 3.1 新增测试文件

- `model/quota/*_test.go`
- `model/iai_route/*_test.go`
- `model/shared/route_rules_test.go`
- `model/iauth/*_test.go`
- `model/iversion_control/version_control_test.go`
- `model/ibasic/*_test.go`
- `model/iprotocol/*_test.go`
- `model/iroute_conf/*_test.go`
- `model/icluster_conf/*_test.go`
- `model/imods/*_test.go`

### 3.2 新增/修改工程文件

- `.github/workflows/model-ci.yml`（新增）
- `Makefile`（新增 `test-model`、`test-model-cover`、`test-model-cover-gate` 目标）
- `.gitignore`（移除 `.github`，添加 `coverage.out`、`model_coverage.out` 等）
- `TESTING.md`（新增单元测试规范与覆盖率清单）

### 3.3 依赖变更

- `go.mod` / `go.sum`：新增 `github.com/stretchr/testify v1.11.1` 及其间接依赖，清理未使用的 `github.com/gofrs/uuid`。

---

## 4. 影响范围

| 维度 | 影响 |
|------|------|
| **OpenAPI 接口** | 无变化 |
| **数据库存储结构** | 无变化 |
| **数据面配置导出** | 无变化 |
| **生产代码行为** | 无变化 |
| **构建产物** | 无变化 |
| **CI 流程** | 新增 model 层单元测试与覆盖率门禁 |
| **开发规范** | 新增 `TESTING.md` 作为单元测试编写参考 |

---

## 5. 验证结果

```bash
cd ai-gateway-api && go test ./model/... -count=1
# 全部模块通过

make test-model-cover-gate
# Model coverage: 87.6%
# Coverage gate passed
```

---

## 6. 后续建议

1. 随着代码迭代，逐步提高 `MODEL_COVERAGE_THRESHOLD`（如从 70% 提升至 80%）。
2. 对 `endpoint/`、`lib/`、`stateful/` 中的纯逻辑部分继续补充单元测试。
3. 在 PR 模板中增加「是否补充/更新单元测试」的检查项。
