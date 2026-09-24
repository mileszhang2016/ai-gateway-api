# Issue #173：Cluster basic 数值子字段范围校验缺失修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/173](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/173)

> POST /open-api/v1/clusters 与 PATCH /open-api/v1/clusters/{cluster_name} 接受 basic.connection/retries/buffers/timeouts 的非法数值（负值、零值），返回 200 而非 422，违反产品设计 clusters.md 第 264-267 行的范围条款。非法配置进入权威下发通道 server_data_conf（`MaxIdleConnsPerHost: -1`、`TimeoutConnSrv: 0`、`ReqWriteBufferSize: 0`、`RetryMax: -1` 等）。

四个手工复现检查点：

| # | 字段 | 非法值 | 期望 | 合同条款 |
|---|------|--------|------|---------|
| 1 | basic.connection.max_idle_conn_per_rs | -1 | 422 | 须为 >=0 整数 |
| 2 | basic.retries.max_retry_in_cluster | -1 | 422 | 须为 >=0 整数 |
| 3 | basic.buffers.req_write_buffer_size | 0 | 422 | 须为 >0 整数 |
| 4 | basic.timeouts.timeout_conn_serv | 0 | 422 | 各项均须为 >0 整数 |

与 #172（passive_health_check 校验缺失）完全同族：Validate() 校了部分子结构却漏数值子字段；normalize* 只填默认不校验范围。危害为运行时层（BFE 侧对这四个字段**无范围校验**，非法值直接生效：TimeoutConnSrv=0 即瞬间超时、MaxIdleConnsPerHost=-1 语义非法、ReqWriteBufferSize=0 写缓冲失效、RetryMax=-1 重试次数非法），削弱转发稳定性但不像 #172 的 uri 那样直接触发 reload 失败。

## 2. 根因（代码核验）

| 证据 | 位置 | 说明 |
|------|------|------|
| 合同范围条款 | `design-docs/api-define/OpenAPI接口定义/clusters.md:264-267` | max_idle_conn_per_rs / max_retry_in_cluster `>=0`；req_write_buffer_size `>0`；timeouts.*（5 项）`>0` |
| Validate() 漏校验 | `endpoints/openapi_v1/product_cluster/create.go` `UpsertParam.Validate()` | 覆盖 Name / StickySessions / PassiveHealthCheck（#172 已补）/ LLMConfig，无 Basic 数值范围分支 |
| 归一化只填默认 | `create.go:256-315` `normalizeBasic` | nil → 默认值（0 / 2 / 512 / 50000,50000,30000,30000,60000）；非 nil 非法值原样穿过；另注意 265-269 行对非法 protocol **静默改写为 https**（相邻发现，见 §8） |
| 创建映射无校验 | `create.go:207-236` `clusterParamControlModel` | basic 各字段直接赋值 rst.Basic.* |
| 更新路径同型缺口 | `update_basic.go:140-176` | `param.Basic.*` 直接透传 rst.Basic.* |
| BFE 侧无范围校验 | `bfe/.../cluster_conf_load.go:563-` `BackendBasicCheck`、`:1076-` `ClusterBasicConfCheck`、`:808-` `GslbBasicConfCheck` | 对 TimeoutConnSrv / MaxIdleConnsPerHost / ReqWriteBufferSize / RetryMax 仅填 nil 默认，**不做范围校验**（已逐函数核读） |

**机制澄清（同 #172 结论）**：`xreq.Bind/BindJSON` 自动调用 `UpsertParam.Validate()`（`lib/xreq/validate.go:38-42`），POST 与 PATCH 两路径**都已执行 Validate()**。修复仍收敛为 Validate() 单一挂点，无需在 update_basic.go 另设校验。

## 3. 裁定：以合同为准，issue 建议与合同无冲突

issue 修复建议的四条规则与合同（clusters.md:264-267）**完全一致**，无需像 #172 那样修正偏差，直接采纳：

- `basic.connection.max_idle_conn_per_rs` 须为 >=0 整数；
- `basic.retries.max_retry_in_cluster` 须为 >=0 整数；
- `basic.buffers.req_write_buffer_size` 须为 >0 整数；
- `basic.timeouts.*`（timeout_conn_serv / timeout_response_header / timeout_readbody_client / timeout_read_client_again / timeout_write_client，共 5 项）均须为 >0 整数。

校验在 Bind 阶段对**原始用户值**执行（先于归一化）：子字段 nil = 使用默认值（合法，跳过）；显式值必须满足合同条件。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`；BFE 侧零改动 |
| 主要文件 | `endpoints/openapi_v1/product_cluster/create.go`（`Validate()` + 新增 `validateBasicRanges`）；`create_test.go`（回归矩阵）；`test/integration/tests/clusters/`（集成用例，执行阶段落地） |
| 接口契约 | 仅新增非法输入的 4xx 拒绝；合法请求（含全部缺省、边界 0 值连接/重试、边界 1 值缓冲/超时）行为不变 |
| 数据迁移 | 无 |
| 明确不做 | 模型层/存储层校验（单一入口原则）；`basic.protocol` 静默改写问题（相邻发现，另案处理，见 §8） |

## 5. 最终方案

### 5.1 新增校验函数（create.go，紧邻 `validatePassiveHealthCheck`）

```go
// validateBasicRanges enforces the numeric range conditions of the basic
// contract (clusters.md): connection/retries >= 0, buffers/timeouts > 0.
// Nil fields are left to normalizeBasic for defaults and are not validated.
func validateBasicRanges(basic *BasicParam) error {
	if basic == nil {
		return nil
	}
	if basic.Connection != nil && basic.Connection.MaxIdleConnPerRs != nil &&
		*basic.Connection.MaxIdleConnPerRs < 0 {
		return xerror.WrapParamErrorWithMsg("basic.connection.max_idle_conn_per_rs must be >= 0")
	}
	if basic.Retries != nil && basic.Retries.MaxRetryInCluster != nil &&
		*basic.Retries.MaxRetryInCluster < 0 {
		return xerror.WrapParamErrorWithMsg("basic.retries.max_retry_in_cluster must be >= 0")
	}
	if basic.Buffers != nil && basic.Buffers.ReqWriteBufferSize != nil &&
		*basic.Buffers.ReqWriteBufferSize <= 0 {
		return xerror.WrapParamErrorWithMsg("basic.buffers.req_write_buffer_size must be > 0")
	}
	if basic.Timeouts != nil {
		if basic.Timeouts.TimeoutConnServ != nil && *basic.Timeouts.TimeoutConnServ <= 0 {
			return xerror.WrapParamErrorWithMsg("basic.timeouts.timeout_conn_serv must be > 0")
		}
		if basic.Timeouts.TimeoutResponseHeader != nil && *basic.Timeouts.TimeoutResponseHeader <= 0 {
			return xerror.WrapParamErrorWithMsg("basic.timeouts.timeout_response_header must be > 0")
		}
		if basic.Timeouts.TimeoutReadbodyClient != nil && *basic.Timeouts.TimeoutReadbodyClient <= 0 {
			return xerror.WrapParamErrorWithMsg("basic.timeouts.timeout_readbody_client must be > 0")
		}
		if basic.Timeouts.TimeoutReadClientAgain != nil && *basic.Timeouts.TimeoutReadClientAgain <= 0 {
			return xerror.WrapParamErrorWithMsg("basic.timeouts.timeout_read_client_again must be > 0")
		}
		if basic.Timeouts.TimeoutWriteClient != nil && *basic.Timeouts.TimeoutWriteClient <= 0 {
			return xerror.WrapParamErrorWithMsg("basic.timeouts.timeout_write_client must be > 0")
		}
	}
	return nil
}
```

命名取 `validateBasicRanges`（而非 issue 建议的 `validateBasicRange`），语义为"数值范围"，避免与 `normalizeBasic` 的"结构归一化"混淆。

### 5.2 挂点（仅一处，双路径覆盖）

`UpsertParam.Validate()` 内、`validatePassiveHealthCheck` 之前插入（先 basic、后 sticky/PHC/llm 的顺序与结构体字段顺序一致）：

```go
	if err := validateBasicRanges(p.Basic); err != nil {
		return err
	}
```

- POST /clusters：`BindJSON` → `Validate()` ✅
- PATCH /clusters/{cluster_name}：`Bind`（URI vars 先映射 Name）→ `Validate()` ✅
- 归一化边界不变：`normalizeBasic` 仍在 `clusterParamControlModel` 内、校验通过后执行，nil → 默认值逻辑不动；
- PATCH 部分更新语义不受影响：basic 未携带（nil）跳过；携带对象但子字段 nil 同样跳过（DAO nil-skip 保留存量值）。

### 5.3 修复后行为对照（issue 四个检查点）

| 输入 | 修复前 | 修复后 |
|------|--------|--------|
| max_idle_conn_per_rs=-1 | 200 入库 → 导出 MaxIdleConnsPerHost:-1 | 400 参数错误，`must be >= 0` |
| max_retry_in_cluster=-1 | 200 入库 → 导出 RetryMax:-1 | 400 参数错误，`must be >= 0` |
| req_write_buffer_size=0 | 200 入库 → 导出 ReqWriteBufferSize:0 | 400 参数错误，`must be > 0` |
| timeout_conn_serv=0 | 200 入库 → 导出 TimeoutConnSrv:0 | 400 参数错误，`must be > 0` |
| max_idle_conn_per_rs=0 / max_retry_in_cluster=0（边界合法） | 200 | 200（合同允许 0） |
| req_write_buffer_size=1 / timeouts.*=1（边界合法） | 200 | 200（合同允许） |
| 全部缺省 / 子字段 nil | 200 + 默认值 | 200 + 默认值（不变） |

## 6. 合同文档变更

合同（clusters.md:264-267）已写明各字段范围条款，**本次无需修改合同**。实现与合同的对齐本身就是修复内容。

**sys-design 同步**：经逐文档核验，`接口层设计文档.md`（4.2.6 节及路由表）、`总体设计文档.md`（能力概述与目录清单）、`模型层设计文档.md`（仅含导出结构体定义，无取值范围表述）、`summary.md` 索引描述均未涉及 basic 数值字段的校验行为或默认值细节，与本修复无矛盾，**均无需改动**；校验遵循的"仅拦截显式非法值、nil 跳过留给归一化"与 `details/部分更新语义与DAO-nil-skip约定.md` 的 PATCH 语义一致，亦无需补充。本次为单端点校验修正（同 #172 先例），规模不达新增 `details/` 长期文档门槛。

## 7. 回归防护

1. **handler 单测**（`endpoints/openapi_v1/product_cluster/create_test.go`，沿用 #172 的测试组织风格）：
   - 四个非法类（conn=-1、retries=-1、buffer=0、timeout_conn_serv=0）+ 五个 timeout 字段逐字段 0 值拒绝，断言 `require.Error` + 错误文案字段可归因；
   - 边界合法值：conn=0、retries=0、buffer=1、timeouts=1 全通过；nil basic、各子结构 nil、空对象 `{}` 通过；
   - 与既有校验共存：basic 非法时优先返回 basic 错误（校验顺序确定）；
   - PATCH 覆盖论证：Validate 为 POST/PATCH 共用挂点（同 #172 机制），无需重复用例。
2. **集成测试**（`test/integration/tests/clusters/`，按既有 CL 编号续接）：
   - 创建：`CL-1-039~042` 四个非法类拒绝（ErrNum=422 + ErrMsg 归因）、`CL-1-043` 边界合法值回显（conn=0/retries=0/buffer=1/timeouts=1）、`CL-1-044` basic 空对象默认值回显；
   - 更新：`CL-4-019~022` PATCH 四类非法值拒绝、`CL-4-023` 非法 PATCH 被拒后存量 basic 配置不被污染（GET 回读断言）；
   - `design.md` 总览表同步新增编号。
3. **E2E**：同 SC2101 cluster 校验缺口族（#172 同族），部署 qa 后随 SC2101 相关用例 requeue-real-verification 复验（非法值 422 + 导出无泄漏）。

## 8. 风险与兼容性

| 项目 | 说明 |
|------|------|
| 兼容性 | 纯增量拒绝：合法请求（含边界值）行为完全不变；仅合同明令非法的负值/零值从"静默入库"变为 4xx |
| 存量数据 | 已入库非法记录不清理（无迁移）；如需修正由运维以合法值重报，或被拒后按合法值重建 |
| 校验边界 | 只在 OpenAPI 入口校验一次（单一入口原则）；InnerAPI 导出消费已入库数据，不重复校验 |
| BFE 关系 | 控制面校验补 BFE 不拦的缺口（已核读 BFE 三个 Check 函数确认无范围校验） |
| **相邻发现（另案）** | `normalizeBasic` 对非法 `basic.protocol` **静默改写为 https**（`create.go:265-269`），违反合同"有效值为 http、https"的语义（静默纠错而非拒绝），本次不动，建议另案改为显式 422 或至少文档声明 |
| 主要风险 | 低。单函数新增 + Validate() 单点挂接；模型层/存储层/导出层零改动 |

---

*文档生成日期：2026-09-16*
