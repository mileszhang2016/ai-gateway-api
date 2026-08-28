# Provider instance_pool 移除 name 字段 — 设计变更说明

## 1. 背景与目标

### 1.1 背景

当前 Provider 数据模型中，`instance_pool` 的每个实例包含 `name`、`addr`、`weight`、`port` 四个字段。`name` 为选填字段，未传入时默认与 `addr` 相同。

该设计带来以下问题：

1. **语义冗余**：实例在 BFE 数据面最终通过 `addr` 和 `port` 定位，`name` 仅作为 backend 名称使用，可由 `(addr, port)` 唯一推导。
2. **唯一性约束复杂**：需要同时校验 `name` 不重复（当非空时）以及 `(name, addr, port)` 组合不重复。
3. **配置生成不一致**：在 `cluster_table` 导出时，backend name 可能为实例传入的 `name`，也可能回退为 `addr`，缺乏统一规则。

### 1.2 目标

1. 简化 Provider 的 `instance_pool` 数据模型，移除实例级别的 `name` 字段。
2. 唯一性约束简化为：同一 provider 内 `(addr, port)` 组合不能重复。
3. 统一 BFE `cluster_table` 中 backend name 的生成规则：使用 `{addr}_{port}` 格式，确保唯一且可读。

## 2. OpenAPI 接口定义变更

### 2.1 数据模型

`design-docs/api-define/OpenAPI接口定义/providers.md` §1 中的 `Instance` 结构由：

```json
{
    "name": "backend-1",
    "addr": "api.deepseek.com",
    "weight": 100,
    "port": 443
}
```

简化为：

```json
{
    "addr": "api.deepseek.com",
    "weight": 100,
    "port": 443
}
```

### 2.2 字段说明

| 字段 | 类型 | 说明 | 合法性条件 |
|------|------|------|------------|
| `instance_pool` | []Instance | Provider 对应的后端实例池 | 必填；至少 1 个元素；同一 provider 内 `(addr, port)` 组合不能重复；至少有一个实例 `weight > 0` |

`Instance` 结构字段表同步删除 `name` 行，仅保留 `addr`、`weight`、`port`。

## 3. 控制面代码变更

### 3.1 Provider 模型 (`model/iprovider/provider.go`)

#### 3.1.1 `ProviderInstance` 结构

移除 `Name` 字段：

```go
type ProviderInstance struct {
    Addr    string `json:"addr"`
    Port    int    `json:"port"`
    Weight  int64  `json:"weight"`
    Disable bool   `json:"disable"`
}
```

#### 3.1.2 校验逻辑 (`validateProviderInstancePool`)

- 删除 `name` 长度校验。
- 删除 `name` 重复性校验。
- 唯一性 key 由 `fmt.Sprintf("%s|%s|%d", name, addr, port)` 改为 `fmt.Sprintf("%s|%d", addr, port)`。

#### 3.1.3 默认值填充 (`FillDefaults`)

删除实例 `name` 默认填充为 `addr` 的逻辑。

### 3.2 Cluster 实例池转换 (`model/icluster_conf/cluster.go`)

`providerInstancesToClusterInstances` 负责将 provider 的实例列表转换为 cluster 子集群的实例列表。由于 cluster 层 `Instance` 结构仍需要 `Name` 字段（用于 BFE 通用 pool 管理），转换时应自动生成：

```go
func providerInstancesToClusterInstances(instances []iprovider.ProviderInstance) []Instance {
    rst := make([]Instance, 0, len(instances))
    for _, inst := range instances {
        rst = append(rst, Instance{
            Name:    fmt.Sprintf("%s_%d", inst.Addr, inst.Port),
            Addr:    inst.Addr,
            Port:    inst.Port,
            Weight:  inst.Weight,
            Disable: inst.Disable,
        })
    }
    return rst
}
```

生成规则：`{addr}_{port}`。

> 注意：IPv6 地址在此处暂不额外处理 `[]` 包裹，因为 `Instance.Name` 仅作为 BFE backend 名称标识，不影响实际连接地址。实际连接地址仍由 `Addr` 字段决定。

### 3.3 cluster_table 导出 (`model/icluster_conf/exporter.go`)

**重点变更**：`clusterTableConfGenerator` 在构造 `cluster_table_conf.BackendConf` 时，必须统一使用 `{addr}_{port}` 作为 `Name`。

当前逻辑：

```go
name := instance.Name
if name == "" {
    name = addr
}
subClusterBackend = append(subClusterBackend, &cluster_table_conf.BackendConf{
    Name:   lib.PString(name),
    Addr:   lib.PString(addr),
    Port:   lib.PInt(instance.Port),
    Weight: lib.PInt(int(instance.Weight)),
})
```

变更后逻辑：

```go
backendName := fmt.Sprintf("%s_%d", addr, instance.Port)
subClusterBackend = append(subClusterBackend, &cluster_table_conf.BackendConf{
    Name:   lib.PString(backendName),
    Addr:   lib.PString(addr),
    Port:   lib.PInt(instance.Port),
    Weight: lib.PInt(int(instance.Weight)),
})
```

其中 `addr` 已经过 IPv6 包裹处理（若 `net.ParseIP(addr)` 为 IPv6 则加 `[]`）。因此：

- IPv4 + port：`api.deepseek.com_443`、`192.168.1.1_443`
- IPv6 + port：`[2001:db8::1]_443`

### 3.4 Inner API 接口

通过 inner-api `/configs/gslb_data/cluster_table` 导出的 `cluster_table` 配置中，每个 backend 的 `Name` 字段将统一为上述 `{addr}_{port}` 格式。

示例：

```json
{
    "Version": "0",
    "Config": {
        "deepseek": {
            "deepseek": [
                {
                    "Name": "api.deepseek.com_443",
                    "Addr": "api.deepseek.com",
                    "Port": 443,
                    "Weight": 100
                }
            ]
        }
    }
}
```

## 4. 数据兼容性

### 4.1 已有 Provider 数据

对于数据库中已存在的 provider 记录，其 `instance_pool` JSON 可能仍包含 `name` 字段。由于 `ProviderInstance` 的 JSON 反序列化会忽略未知字段，因此旧数据可以正常读取；但写回时应以新的无 `name` 结构存储。

### 4.2 Cluster 子集群实例池

Cluster 子集群的实例池数据来自 provider 同步或 cluster 创建时的快照。变更后：

- 新建/更新 provider 时，同步到 cluster 子集群的实例 `Name` 将统一为 `{addr}_{port}`。
- 已有 cluster 子集群中的旧 `Name`（如用户传入的 `backend-1` 或裸 `addr`）不会被自动重写，除非触发 provider 同步或 cluster 更新。

## 5. 测试影响

需要更新以下测试用例与断言：

1. `model/iprovider/provider_test.go`：移除 `name` 相关校验测试，新增 `(addr, port)` 重复校验测试。
2. `model/icluster_conf/cluster_test.go`：更新 provider instance pool 同步后的实例 `Name` 断言。
3. `model/icluster_conf/exporter_test.go`：更新 `cluster_table` 导出中 `BackendConf.Name` 的断言。
4. 集成测试：更新创建 provider、更新 provider、instance pool 同步相关用例中的预期 backend name。

## 6. 待办

- [ ] 修改 `model/iprovider/provider.go`：`ProviderInstance` 结构、校验、默认值。
- [ ] 修改 `model/icluster_conf/cluster.go`：`providerInstancesToClusterInstances` 生成 `{addr}_{port}` 名称。
- [ ] 修改 `model/icluster_conf/exporter.go`：`clusterTableConfGenerator` 使用 `{addr}_{port}` 作为 `BackendConf.Name`。
- [ ] 更新 `design-docs/api-define/OpenAPI接口定义/providers.md` 中 §3 校验规则（如需要）。
- [ ] 更新相关单元测试与集成测试。
