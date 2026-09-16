# Issue #172：Cluster passive_health_check 参数校验缺失修复方案

## 1. 问题来源

[rainway-ai-gateway/ai-gateway-api/issues/172](https://github.com/rainway-ai-gateway/ai-gateway-api/issues/172)

> POST /open-api/v1/clusters 与 PUT /open-api/v1/clusters/{name}/basic 对 passive_health_check 仅做 nil 默认化，不校验字段取值范围与格式。failnum=-1、interval=-1、statuscode=999、uri 非 `/` 开头四类非法值均以 200 被接受入库，进入 server_data_conf 权威下发通道。
>
> 已造成实际事故：非法 uri 导致 BFE reload 失败（`clusterTableLoad Error BfeClusterConf.Config:conf for passive_invalid:CheckConf:Uri should be start with '/'`），后继 cluster 配置无法下发。

危害分两层：

- **reload 层（已发生）**：uri / statuscode 非法值触发 BFE `BackendCheckCheck` 拒绝，整份 server_data_conf 热加载失败，阻塞全部后继配置下发；
- **运行时层（潜伏）**：failnum/interval 负值 BFE 不拦截，健康检查以非法参数运行（永不触发摘除或退避异常），削弱故障隔离。

E2E：SC2101-TC014（run_id 20260915-164011）FAILED_PRODUCT，A-01~A-04 断言失败、A-07 反向证实非法配置已泄漏至下发通道。

## 2. 根因（代码核验）

| 证据 | 位置 | 说明 |
|------|------|------|
| 合同合法性条件 | `design-docs/api-define/OpenAPI接口定义/clusters.md:149-157` | failnum `>=0`；interval `>=0`；uri 非空、以 `/` 开头；statuscode `0` 或 `100-599` |
| Validate() 漏校验 | `endpoints/openapi_v1/product_cluster/create.go:108-122` | 仅覆盖 Name / ClusterName / StickySessions / LLMConfig，无 PassiveHealthCheck 分支 |
| 归一化只填默认 | `create.go:307-328` `normalizePassiveHealthCheck` | 仅 nil → 默认值（3/1000/0/`/`）；非 nil 非法值原样穿过 |
| 更新路径同型缺口 | `update_basic.go:194-204` | `param.PassiveHealthCheck` 直接映射到模型层，无校验 |
| 模型层无校验 | `model/icluster_conf/cluster.go:108-115` | `ClusterPassiveHealthCheckParam` 纯结构定义，`toBackendCheck()` 直映射 BFE `BackendCheck` |
| BFE reload 侧校验 | `bfe/bfe_config/bfe_cluster_conf/cluster_conf/cluster_conf_load.go:742-805,625-645` | `BackendCheckCheck` 校验 uri `/` 前缀与 statuscode ∈ `[100,599]∪[0,31]`（0=任意）；**failnum/interval 无范围校验** |

**关键机制澄清（比 issue 原因分析更进一步）**：`lib/xreq/validate.go:27-44` 定义 `Validator` 接口，`xreq.Bind/BindJSON` 在绑定后**自动调用** `Validate()`。POST 走 `BindJSON`（`create.go:164`），PATCH 走 `Bind`（`update_basic.go:62`，`Bind` 先于 validate 将 mux URI vars 映射入 `Name`，`lib/xreq/param.go:98-112`）——**两条路径都已执行 `UpsertParam.Validate()`**，缺的不是调用点，而是 Validate() 内部的 PassiveHealthCheck 分支。因此修复收敛为单一挂点，无需在 update_basic.go 另设校验（issue 建议项 3 可简化）。

"failnum 传字符串被 4xx 拒绝"的假象：来自 JSON 反序列化层（`*int32` 拒收 string，`lib/xreq/param.go:68-80`），非业务校验。

## 3. 裁定：校验以合同为准，修正 issue 建议的两处偏差

issue 修复建议与合同（clusters.md:149-157）存在两处冲突，本方案**以合同为准**：

| 字段 | issue 建议 | 合同 | 本方案 | 理由 |
|------|-----------|------|--------|------|
| interval | `>= 1` | `>= 0` | **`>= 0`** | 合同是验收基准；BFE 侧同样不校验 interval 下限，控制面按合同收敛即可，不额外加码 |
| statuscode | `[100, 599]`（拒绝 0） | `0` 或 `100-599` | **允许 0** | 合同显式允许 0（"需要忽略返回码，此处可以填0"，默认即 0）；若按 issue 建议拒绝 0，会把合法的显式 0 挡在门外，破坏合同语义且与默认值 0 自相矛盾 |

校验在 Bind 阶段对**原始用户值**执行（先于归一化）：子字段 nil = 使用默认值（合法，不校验）；显式值必须满足合同条件。`host` 合同无约束，不校验。

## 4. 范围

| 范围 | 说明 |
|------|------|
| 涉及仓库 | `ai-gateway-api`（控制面入口校验）；BFE 侧零改动 |
| 主要文件 | `endpoints/openapi_v1/product_cluster/create.go`（`Validate()` + 新增 `validatePassiveHealthCheck`）；`create_test.go`（回归矩阵） |
| 接口契约 | 仅新增非法输入的 4xx 拒绝；合法请求（含全部缺省、显式 0 statuscode）行为不变 |
| 数据迁移 | 无（存量非法记录仍可能存在于 DB，由 E2E 环境清理；控制面只堵增量入口） |
| 明确不做 | 模型层/存储层校验（单一入口原则：校验在边界做一次）；BFE 侧加固（另案评估） |

## 5. 最终方案

### 5.1 新增校验函数（create.go，紧邻 `validateStickySessions`）

```go
func validatePassiveHealthCheck(phc *PassiveHealthCheckParam) error {
	if phc == nil {
		return nil
	}
	if phc.Failnum != nil && *phc.Failnum < 0 {
		return xerror.WrapParamErrorWithMsg("passive_health_check.failnum must be >= 0")
	}
	if phc.Interval != nil && *phc.Interval < 0 {
		return xerror.WrapParamErrorWithMsg("passive_health_check.interval must be >= 0")
	}
	if phc.Statuscode != nil && *phc.Statuscode != 0 &&
		(*phc.Statuscode < 100 || *phc.Statuscode > 599) {
		return xerror.WrapParamErrorWithMsg(
			"passive_health_check.statuscode must be 0 or in [100, 599]")
	}
	if phc.Uri != nil && (*phc.Uri == "" || !strings.HasPrefix(*phc.Uri, "/")) {
		return xerror.WrapParamErrorWithMsg("passive_health_check.uri must be non-empty and start with '/'")
	}
	return nil
}
```

### 5.2 挂点（仅一处，双路径覆盖）

`UpsertParam.Validate()` 内、LLMConfig 校验之前插入（与结构体字段顺序一致）：

```go
	if err := validatePassiveHealthCheck(p.PassiveHealthCheck); err != nil {
		return err
	}
```

- POST /clusters：`newCreateParam4Create` → `BindJSON` → `Validate()`（含 PHC 分支）✅
- PATCH /clusters/{name}/basic：`newUpdateParam4Update` → `Bind`（URI vars 先映射 Name）→ `Validate()`（含 PHC 分支）✅
- 归一化与校验的边界不变：`normalizePassiveHealthCheck` 仍在 `clusterParamControlModel` 内、校验通过之后执行，nil → 默认值（3/1000/0/`/`）逻辑不动。
- PATCH 部分更新语义不受影响：`passive_health_check` 未携带（nil）时跳过校验；携带对象但子字段全 nil 时同样全过（等价于不修改，DAO nil-skip 保留存量值）。

### 5.3 修复后行为对照（issue 四类非法值）

| 输入 | 修复前 | 修复后 |
|------|--------|--------|
| failnum=-1 | 200 入库 | 400 参数错误，`failnum must be >= 0` |
| interval=-1 | 200 入库 | 400 参数错误，`interval must be >= 0` |
| statuscode=999 | 200 入库 → BFE reload 失败 | 400 参数错误，`statuscode must be 0 or in [100, 599]` |
| uri="xxx"（非 `/` 开头） | 200 入库 → BFE reload 失败 | 400 参数错误，`uri must be non-empty and start with '/'` |
| statuscode=0（显式） | 200 | 200（合同允许的"忽略返回码"） |
| 全部缺省 / 子字段 nil | 200 + 默认值 | 200 + 默认值（不变） |

## 6. 合同文档变更

合同（clusters.md:149-157 表：被动健康检查）已写明各字段合法性条件，**本次无需修改合同**。方案按合同落地，实现与合同的对齐本身就是修复内容。

## 7. 回归防护

1. **handler 单测**（`endpoints/openapi_v1/product_cluster/create_test.go`，沿用 `TestUpsertParamValidate_StickySessions` 的 base() 构造风格）：
   - 四类非法值（failnum=-1、interval=-1、statuscode=999、uri="no-slash"）分别断言 `require.Error` + 错误文案包含字段归因；
   - 边界合法值：failnum=0、interval=0、statuscode=0、statuscode=100、statuscode=599、uri="/"、uri="/healthz" 全通过；
   - nil PHC、PHC 全 nil 子字段（`{}`）通过；
   - uri=""（显式空串）拒绝；
   - 与 StickySessions/LLMConfig 校验共存：PHC 非法时优先返回 PHC 错误（校验顺序确定）。
2. **集成测试**（`test/integration/tests/clusters/`，按模块既有 CL 编号惯例）：
   - 创建：`CL-1-032~036` 四类非法值 + uri 显式空串拒绝（断言 ErrNum=422 且 ErrMsg 字段可归因）、`CL-1-037` 边界合法值回显（failnum=0/interval=0/statuscode=0/uri="/probe"）、`CL-1-038` 空对象默认值填充（3/1000/0/`/`）；
   - 更新：`CL-4-014` PATCH 合法值回读 + InnerAPI `CheckConf` 导出一致性、`CL-4-015~017` PATCH 三类非法值拒绝、`CL-4-018` 非法 PATCH 被拒后 GET 回读存量值不变（防污染断言）；
   - 用例设计文档 `tests/clusters/design.md` 总览表同步新增上述编号（顺带补齐原缺失的 CL-1-029~031 行）。
3. **E2E**：SC2101-TC014 修复部署后走 requeue-real-verification 重跑，A-01~A-04 断言转 PASSED，并确认被拒集群名不再出现在 server_data_conf 导出（A-07 反向断言转 PASS）。

## 8. 风险与兼容性

| 项目 | 说明 |
|------|------|
| 兼容性 | 纯增量拒绝：此前 200 接受的请求若取值合法，行为完全不变；仅四类合同明令非法的取值从"静默入库"变为 4xx。web 前端（`ai-gateway-web` 集群表单）如已传合法值则不受影响 |
| 存量数据 | 已入库的非法记录不被本方案清理（无迁移）；它们在下一次该 cluster 被更新并以合法值重报前仍会触发 BFE reload 失败，需运维在目标环境手动修正或删除（E2E 环境随用例清理） |
| 校验边界 | 只在 OpenAPI 入口校验一次（单一入口原则）；InnerAPI 导出通道消费的都是已入库数据，不重复校验 |
| 与 BFE 的关系 | 控制面校验是 BFE reload 校验的超集收敛：uri/statuscode 规则与 BFE `BackendCheckCheck` 一致（杜绝 reload 失败），failnum/interval 按合同补 BFE 不拦的缺口 |
| 主要风险 | 低。单函数新增 + Validate() 单点挂接；模型层/存储层/导出层零改动 |

---

*文档生成日期：2026-09-16*
