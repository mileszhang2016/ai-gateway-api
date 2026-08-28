# Provider instance_pool 移除 name 字段并改用 addr_port 生成 BFE backend name

## 变更概述

简化 Provider 数据模型中的 `instance_pool` 元素：

1. **移除 `instance_pool` 元素中的 `name` 字段**
   - 原 `Instance` 结构包含 `name`、`addr`、`weight`、`port` 四个字段。
   - 变更后仅保留 `addr`、`weight`、`port` 三个字段。

2. **唯一性约束调整**
   - 原约束：同一 provider 内，对于 `name` 不为空的实例，`name` 不能重复；同一 provider 内 `(name, addr, port)` 组合不能重复。
   - 变更后：同一 provider 内 `(addr, port)` 组合不能重复。

3. **BFE cluster_table backend name 生成规则调整**
   - 在通过 inner-api `/configs/gslb_data/cluster_table` 生成 BFE 配置时，后端实例的 `Name` 字段统一使用 `{addr}_{port}` 格式生成，例如 `api.deepseek.com_443`。
   - 不再依赖实例上可能存在的 `name` 字段，也不再用裸 `addr` 作为 name。

## 影响范围

- OpenAPI 接口定义：`design-docs/api-define/OpenAPI接口定义/providers.md` §1 数据模型及 §2.1 示例。
- 控制面代码：
  - `model/iprovider/provider.go`：`ProviderInstance` 结构、校验逻辑、默认值填充。
  - `model/icluster_conf/cluster.go`：provider instance pool 到 cluster instance pool 的转换。
  - `model/icluster_conf/exporter.go`：`cluster_table` 导出时 `BackendConf.Name` 的生成。
- 数据面消费：BFE 通过 `/configs/gslb_data/cluster_table` 获取的 `cluster_table` 配置中，backend name 将由 `addr_port` 组成。

## 涉及文件

- `ai-gateway-api/design-docs/api-define/OpenAPI接口定义/providers.md`
- `ai-gateway-api/model/iprovider/provider.go`
- `ai-gateway-api/model/icluster_conf/cluster.go`
- `ai-gateway-api/model/icluster_conf/exporter.go`
- 相关单元测试与集成测试

## 状态

- [x] 接口定义文档更新
- [x] 控制面代码实现
- [x] 单元测试更新
- [x] 集成测试更新
