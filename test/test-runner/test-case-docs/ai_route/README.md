# AI 路由规则模块 - 测试用例总览

> 本目录存放 AI 路由规则模块的测试用例设计文档。  
> 对应的 Go 测试文件位于 `test/test-runner/test-cases/ai_route/` 目录。

**模块**：AI 路由规则（AI Route Rules）  
**接口数**：2  
**测试用例数**：24  
**覆盖维度**：正常参数、必填校验、边界值、异常参数、nil 校验、返回数据校验、业务规则

---

## 接口清单

| 编号 | 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|------|
| AR-1 | 全量更新AI路由规则 | PATCH | `/open-api/v1/ai-route-rules` | 全量替换路由规则，支持基础和高级路由 |
| AR-2 | 获取AI路由规则列表 | GET | `/open-api/v1/ai-route-rules` | 获取当前产品线的路由规则配置 |

---

## 测试场景总览

### AR-1：全量更新AI路由规则（PATCH）— 21 个用例

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AR-1-001 | 仅设置基础路由规则 | 正常参数 | 传入 basic_forward_rules 数组，验证 ErNum=200 |
| AR-1-002 | 仅设置高级路由规则 | 正常参数 | 传入 forward_rules 数组，验证 ErNum=200 |
| AR-1-003 | 同时设置基础和高级路由 | 正常参数 | 同时传入两种规则，验证 ErNum=200 |
| AR-1-004 | 清空所有规则 | 正常参数 | 传入空数组，验证 ErNum=200 |
| AR-1-005 | 设置多条高级路由规则 | 正常参数 | 传入多条 forward_rules，验证 ErNum=200 |
| AR-1-006 | 设置多条基础路由规则 | 正常参数 | 传入多条 basic_forward_rules，验证 ErNum=200 |
| AR-1-007 | 缺少 forward_rules[].expression | 必填校验 | 不传 expression 字段，验证 ErNum=422 |
| AR-1-008 | 缺少 forward_rules[].cluster_name | 必填校验 | 不传 cluster_name 字段，验证 ErNum=422 |
| AR-1-009 | 缺少 basic_forward_rules[].cluster_name | 必填校验 | 不传 cluster_name 字段，验证 ErNum=422 |
| AR-1-010 | forward_rules[].expression 为空字符串 | 边界值 | expression="" ，验证 ErNum=422 |
| AR-1-011 | forward_rules[].cluster_name 为空字符串 | 边界值 | cluster_name="" ，验证 ErNum=422 |
| AR-1-012 | basic_forward_rules[].cluster_name 为空字符串 | 边界值 | cluster_name="" ，验证 ErNum=422 |
| AR-1-013 | forward_rules 数组元素为 null | nil 校验 | 传入 [null]，验证 ErNum=422 |
| AR-1-014 | basic_forward_rules 数组元素为 null | nil 校验 | 传入 [null]，验证 ErNum=422 |
| AR-1-015 | 空 Body | 边界值 | 不传任何字段，验证 ErNum=422 |
| AR-1-016 | 非法 JSON Body | 异常输入 | 传入非 JSON 字符串，验证 ErNum=422 |
| AR-1-017 | forward_rules[].description 可选 | 可选字段 | 不传 description 字段，验证 ErNum=200 |
| AR-1-018 | basic_forward_rules[].host_names 可选 | 可选字段 | 不传 host_names 字段，验证 ErNum=200 |
| AR-1-019 | basic_forward_rules[].paths 可选 | 可选字段 | 不传 paths 字段，验证 ErNum=200 |
| AR-1-020 | 返回数据镜像请求 | 返回数据校验 | 验证响应 Data 结构与请求一致 |
| AR-1-021 | 最后一条自动追加 default_t() | 业务规则 | 通过 GET 验证最后一条规则自动追加 default_t() |

### AR-2：获取AI路由规则列表（GET）— 3 个用例

| 编号 | 场景 | 测试类型 | 简要说明 |
|------|------|---------|---------|
| AR-2-001 | 获取已设置的规则 | 正常参数 | 先设置规则，再获取，验证返回数据一致 |
| AR-2-002 | 获取未设置时的列表 | 空数据 | 不设置规则，直接获取，验证返回空数组 |
| AR-2-003 | 返回数据结构校验 | 返回数据校验 | 验证返回字段完整（basic_forward_rules、forward_rules） |

---

## 详细设计文档

| 接口 | 设计文档 |
|------|---------|
| 全量更新AI路由规则 | [set_rules/design.md](./set_rules/design.md) |
| 获取AI路由规则列表 | [get_rules/design.md](./get_rules/design.md) |

---

## 目录结构

```
test-case-docs/ai_route/          ← 设计文档（本目录）
├── README.md                     ← 本文件
├── set_rules/design.md           ← 全量更新路由规则 设计文档
└── get_rules/design.md           ← 获取路由规则 设计文档

test-cases/ai_route/              ← Go 测试文件
├── set_rules/set_rules_test.go
└── get_rules/get_rules_test.go
```

## 测试执行

```bash
cd ai-gateway-api/test/test-runner
go test -v -count=1 -timeout 120s ./test-cases/ai_route/set_rules/
go test -v -count=1 -timeout 120s ./test-cases/ai_route/get_rules/
```