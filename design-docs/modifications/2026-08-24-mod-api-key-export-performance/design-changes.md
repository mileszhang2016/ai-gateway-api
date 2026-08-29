# mod-api-key 配置导出性能优化

## 1. 背景

线上用户通过 Open API 批量写入 690 个 API Key 后，BFE 数据面出现部分 Key 报 `INVALID_API_KEY` 的故障。排查发现 conf-agent 拉取 `/inner-api/v1/configs/mod-api-key` 时反复超时：

```text
context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```

根本原因是 `ai-gateway-api/model/imods/exporter.go` 中的 `APIKeyRuleGenerator` 对每个 API Key 都进行多次单条数据库查询（Entity 层级、QuotaPlan、EntityType），形成 N+1 查询。690 个 Key 会触发数千次 DB 查询，导致接口响应超过 conf-agent 的 1.5 秒超时，BFE 只能拿到不完整的 `token_rule.data`。

相邻的 `ai-route` 导出逻辑（`ai-gateway-api/model/imods/ai_route_exporter.go`）已经采用“全量 Entity 预加载 + 内存 Map 回溯”的方式，mod-api-key 导出逻辑需要同步改造。

---

## 2. 目标

1. 将 `/inner-api/v1/configs/mod-api-key` 接口的 DB 查询次数从 **O(N × 层级数)** 降到 **O(4)**。
2. 保证 690 个 API Key、3 层 Entity 层级的场景下，单线程 `ConfigExport` 耗时 < 500 ms。
3. 保持导出的 JSON 配置结构与现有 BFE 消费侧完全兼容。
4. 补充单元测试与集成测试，覆盖批量 Key 场景。

---

## 3. 方案概述

### 3.1 核心思路

在 `APIKeyRuleGenerator` 中，先一次性全量加载三类数据到内存索引，再遍历 API Key：

- 全量 API Key（已有 `FetchAPIKeyList`）
- 全量 Entity（已有 `FetchEntityList`）
- 全量 QuotaPlan（已有 `FetchQuotaPlanList`）
- 全量 EntityType（需确认/新增 `FetchEntityTypeList`）

后续所有 Entity 层级、QuotaPlan、EntityType 查询全部走内存 Map，不再递归调用单条 `FetchEntity`/`FetchQuotaPlan`/`FetchEntityType`。

### 3.2 与 ai_route_exporter 的差异

ai_route 场景只需要 Entity Map；mod-api-key 还需要 QuotaPlan Map 和 EntityType Map，用于计算每个 Token 的 `QuotaPlans` 和 `Tags`。

---

## 4. 具体改动

### 4.1 `APIKeyRuleGenerator` 预加载数据

**文件：** `ai-gateway-api/model/imods/exporter.go`

当前 `APIKeyRuleGenerator` 直接遍历 API Key 列表，对每个 Key 调用递归查询：

```go
apiKeyList, err := rlm.apiKeyStorager.FetchAPIKeyList(ctx, &api_key.APIKeyFilter{})
// ...
for _, one := range apiKeyList {
    if one.EntityID != nil && *one.EntityID != "" && rlm.entityStorager != nil {
        entityAllowModels, entityBlockModels, err = rlm.fetchEntityModelHierarchy(ctx, *one.EntityID)
        // ...
    }
    quotaPlanIDs, tags, err := rlm.fetchQuotaPlansWithEntityHierarchy(ctx, one, *one.ProductName, collectedQuotaPlans)
    // ...
}
```

改造后在循环前批量加载：

```go
func (rlm *APIKeyRuleManager) APIKeyRuleGenerator(ctx context.Context) (*iversion_control.ExportData, error) {
    collectedQuotaPlans := make(map[string][]*QuotaPlan)

    product2config, err := rlm.buildAIRouteAPIKeyRules(ctx)
    if err != nil {
        return nil, err
    }

    productName2Config := make(map[string][]*TokenRuleFile)
    for productName, productConfig := range product2config {
        if len(productConfig.Rules) > 0 {
            productName2Config[productName] = convertAPIKeyRulesToBfeRules(productConfig.Rules)
        }
    }

    // 1. 全量 API Key
    apiKeyList, err := rlm.apiKeyStorager.FetchAPIKeyList(ctx, &api_key.APIKeyFilter{})
    if err != nil {
        return nil, err
    }

    // 2. 全量 Entity
    entityMap, err := rlm.buildEntityMap(ctx)
    if err != nil {
        return nil, err
    }

    // 3. 全量 QuotaPlan
    quotaPlanMap, err := rlm.buildQuotaPlanMap(ctx)
    if err != nil {
        return nil, err
    }

    // 4. 全量 EntityType
    entityTypeMap, err := rlm.buildEntityTypeMap(ctx)
    if err != nil {
        return nil, err
    }

    apiKey2Config := make(map[string]map[string]*TokenFile)
    for _, one := range apiKeyList {
        // ... 使用 entityMap / quotaPlanMap / entityTypeMap 填充 TokenFile
    }

    // ...
}
```

新增辅助函数：

```go
func (rlm *APIKeyRuleManager) buildEntityMap(ctx context.Context) (map[string]*entpkg.EntityParam, error) {
    if rlm.entityStorager == nil {
        return nil, nil
    }
    allEntities, err := rlm.entityStorager.FetchEntityList(ctx, &entpkg.EntityFilter{})
    if err != nil {
        return nil, err
    }
    entityMap := make(map[string]*entpkg.EntityParam, len(allEntities))
    for _, e := range allEntities {
        if e.EntityID != nil {
            entityMap[*e.EntityID] = e
        }
    }
    return entityMap, nil
}

func (rlm *APIKeyRuleManager) buildQuotaPlanMap(ctx context.Context) (map[int64]*quota.QuotaPlanParam, error) {
    if rlm.quotaPlanStorager == nil {
        return nil, nil
    }
    allPlans, err := rlm.quotaPlanStorager.FetchQuotaPlanList(ctx, &quota.QuotaPlanFilter{})
    if err != nil {
        return nil, err
    }
    planMap := make(map[int64]*quota.QuotaPlanParam, len(allPlans))
    for _, qp := range allPlans {
        if qp.ID != nil {
            planMap[*qp.ID] = qp
        }
    }
    return planMap, nil
}

func (rlm *APIKeyRuleManager) buildEntityTypeMap(ctx context.Context) (map[string]*entpkg.EntityTypeParam, error) {
    if rlm.entityTypeStorager == nil {
        return nil, nil
    }
    allTypes, err := rlm.entityTypeStorager.FetchEntityTypeList(ctx, &entpkg.EntityTypeFilter{})
    if err != nil {
        return nil, err
    }
    typeMap := make(map[string]*entpkg.EntityTypeParam, len(allTypes))
    for _, et := range allTypes {
        if et.TypeName != nil {
            typeMap[*et.TypeName] = et
        }
    }
    return typeMap, nil
}
```

### 4.2 递归查询改为内存回溯

**文件：** `ai-gateway-api/model/imods/exporter.go`

在文件顶部新增层级深度保护常量，防止异常 `parent_id` 循环导致死循环：

```go
// maxEntityHierarchyDepth limits the number of ancestor levels walked when
// resolving Entity hierarchy, protecting against accidental parent_id cycles.
const maxEntityHierarchyDepth = 10
```

#### 4.2.1 `fetchEntityModelHierarchy`

保留原函数名，增加 `entityMap` 参数，移除对 `collectEntityModels` 的递归调用：

```go
func (rlm *APIKeyRuleManager) fetchEntityModelHierarchy(ctx context.Context, entityMap map[string]*entpkg.EntityParam, entityID string) ([]string, []string, error) {
    var allAllowModels [][]string
    var allBlockModels []string

    for depth := 0; depth < maxEntityHierarchyDepth; depth++ {
        entity, ok := entityMap[entityID]
        if !ok || entity == nil {
            break
        }
        if len(entity.AllowModels) > 0 && !containsStar(entity.AllowModels) {
            allAllowModels = append(allAllowModels, entity.AllowModels)
        }
        if len(entity.BlockModels) > 0 && !containsStar(entity.BlockModels) {
            allBlockModels = append(allBlockModels, entity.BlockModels...)
        }
        if entity.ParentID == nil || *entity.ParentID == "" {
            break
        }
        entityID = *entity.ParentID
    }

    return intersectAllowModels(allAllowModels), allBlockModels, nil
}
```

#### 4.2.2 `fetchEntityQuotaPlanHierarchy`

保留原函数名，增加 `entityMap`、`quotaPlanMap`、`entityTypeMap` 参数：

```go
func (rlm *APIKeyRuleManager) fetchEntityQuotaPlanHierarchy(ctx context.Context, entityMap map[string]*entpkg.EntityParam, quotaPlanMap map[int64]*quota.QuotaPlanParam, entityTypeMap map[string]*entpkg.EntityTypeParam, entityID string, productName string, collectedQuotaPlans map[string][]*QuotaPlan) ([]string, []ApikeyTag, error) {
    var quotaPlanIDs []string
    var tags []ApikeyTag

    for depth := 0; depth < maxEntityHierarchyDepth; depth++ {
        entity, ok := entityMap[entityID]
        if !ok || entity == nil {
            break
        }

        if entity.EntityID != nil && entity.Type != nil && entity.Name != nil {
            tag := ApikeyTag{
                TagName:  *entity.Type,
                TagValue: *entity.Name,
            }
            if rlm.entityTypeStorager != nil {
                entityType, ok := entityTypeMap[*entity.Type]
                if !ok || entityType == nil || entityType.Level == nil {
                    return nil, nil, fmt.Errorf("entity type %s not found or level invalid", *entity.Type)
                }
                tag.TagLevel = *entityType.Level
            }
            tags = append(tags, tag)
        }

        if entity.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
            if quotaPlan, ok := quotaPlanMap[*entity.QuotaPlanID]; ok && quotaPlan != nil && entity.EntityID != nil {
                if quotaPlan.Unlimited == nil || !*quotaPlan.Unlimited {
                    qp := convertQuotaPlanToExport(quotaPlan, *entity.EntityID, *entity.EntityID)
                    if !containsQuotaPlan(collectedQuotaPlans, productName, qp.Id) {
                        if _, ok := collectedQuotaPlans[productName]; !ok {
                            collectedQuotaPlans[productName] = make([]*QuotaPlan, 0)
                        }
                        collectedQuotaPlans[productName] = append(collectedQuotaPlans[productName], qp)
                    }
                    quotaPlanIDs = append(quotaPlanIDs, qp.Id)
                }
            }
        }

        if entity.ParentID == nil || *entity.ParentID == "" {
            break
        }
        entityID = *entity.ParentID
    }

    return quotaPlanIDs, tags, nil
}
```

#### 4.2.3 `fetchQuotaPlansWithEntityHierarchy` 适配

保留原函数名，增加三个 Map 参数，API Key 自身 QuotaPlan 也走 `quotaPlanMap`：

```go
func (rlm *APIKeyRuleManager) fetchQuotaPlansWithEntityHierarchy(ctx context.Context, apiKey *api_key.APIKeyParam, productName string, collectedQuotaPlans map[string][]*QuotaPlan, entityMap map[string]*entpkg.EntityParam, quotaPlanMap map[int64]*quota.QuotaPlanParam, entityTypeMap map[string]*entpkg.EntityTypeParam) ([]string, []ApikeyTag, error) {
    quotaPlanIDs := make([]string, 0)
    tags := make([]ApikeyTag, 0)

    if apiKey.QuotaPlanID != nil && rlm.quotaPlanStorager != nil {
        if quotaPlan, ok := quotaPlanMap[*apiKey.QuotaPlanID]; ok && quotaPlan != nil && apiKey.Key != nil {
            if quotaPlan.Unlimited == nil || !*quotaPlan.Unlimited {
                qp := convertQuotaPlanToExport(quotaPlan, *apiKey.Key, *apiKey.Key)
                if !containsQuotaPlan(collectedQuotaPlans, productName, qp.Id) {
                    if _, ok := collectedQuotaPlans[productName]; !ok {
                        collectedQuotaPlans[productName] = make([]*QuotaPlan, 0)
                    }
                    collectedQuotaPlans[productName] = append(collectedQuotaPlans[productName], qp)
                }
                quotaPlanIDs = append(quotaPlanIDs, qp.Id)
            }
        }
    }

    if apiKey.EntityID != nil && *apiKey.EntityID != "" && rlm.entityStorager != nil {
        entityIDs, entityTags, err := rlm.fetchEntityQuotaPlanHierarchy(ctx, entityMap, quotaPlanMap, entityTypeMap, *apiKey.EntityID, productName, collectedQuotaPlans)
        if err != nil {
            return nil, nil, err
        }
        quotaPlanIDs = append(quotaPlanIDs, entityIDs...)
        tags = append(tags, entityTags...)
    }

    return quotaPlanIDs, tags, nil
}
```

### 4.3 EntityType 批量查询接口

**文件：** `ai-gateway-api/model/entity/entity_type.go`

确认 `EntityTypeStorager` 接口是否已有 `FetchEntityTypeList`：

```go
type EntityTypeStorager interface {
    // ...
    FetchEntityTypeList(ctx context.Context, filter *EntityTypeFilter) ([]*EntityTypeParam, error)
}
```

若缺失，需新增接口方法并在 `storage/rdb/entity/entity_type.go` 中实现。

---

## 5. 涉及文件清单

| 文件 | 修改内容 |
|---|---|
| `ai-gateway-api/model/imods/exporter.go` | `APIKeyRuleGenerator` 预加载 Entity/QuotaPlan/EntityType；`fetchEntityModelHierarchy`、`fetchQuotaPlansWithEntityHierarchy`、`fetchEntityQuotaPlanHierarchy` 改为基于内存 Map 的回溯。 |
| `ai-gateway-api/model/imods/exporter_test.go` | mock `FetchEntityList`、`FetchQuotaPlanList`、`FetchEntityTypeList`；补充 690 Key 量级的性能/正确性测试。 |
| `ai-gateway-api/model/imods/mocks_test.go` | 若缺少批量查询 mock，需补充。 |
| `ai-gateway-api/model/entity/entity_type.go` | 确认/新增 `EntityTypeStorager.FetchEntityTypeList` 接口。 |
| `ai-gateway-api/storage/rdb/entity/entity_type.go` | 若接口新增 `FetchEntityTypeList`，需实现 DAO。 |

---

## 6. 测试计划

### 6.1 单元测试

**文件：** `ai-gateway-api/model/imods/exporter_test.go`

新增测试 `TestAPIKeyRuleManager_ConfigExport_Performance`：

1. 构造 690 个 API Key，每个关联不同 Entity / QuotaPlan，Entity 层级 3 层。
2. mock `FetchAPIKeyList`、`FetchEntityList`、`FetchQuotaPlanList`、`FetchEntityTypeList` 一次性返回。
3. 断言：
   - 导出结果包含全部 690 个 Token。
   - QuotaPlans 收集正确，无重复。
   - Entity 层级 Tags 正确。
   - `ConfigExport` 耗时 < 500 ms。

运行命令：

```bash
cd ai-gateway-api
go test ./model/imods/... -run TestAPIKeyRuleManager_ConfigExport_Performance -v
```

### 6.2 集成测试

**文件：** `ai-gateway-api/test/integration/tests/innerapi/mod_api_key/mod_api_key_test.go`

1. 批量创建 700+ API Key。
2. 调用 `/inner-api/v1/configs/mod-api-key`，断言返回 200 且包含全部 Key。
3. 验证 BFE `token_rule.data` 文件大小与 Key 数量成正比。

### 6.3 回归测试

```bash
cd ai-gateway-api
make test-model
make test-model-cover-gate
```

---

## 7. 兼容性说明

- `/configs/mod-api-key` 返回的 JSON 结构不变，BFE 消费侧无需修改。
- 数据库 schema 不变，仅查询方式从单条递归改为批量 + 内存回溯。

---

## 8. 风险与缓解

| 风险 | 缓解措施 |
|---|---|
| 批量加载全量 Entity/QuotaPlan/EntityType 导致瞬时内存 spike | 690 条记录全量加载内存占用极小（< 10 MB），可接受；后续若达十万级再考虑分页或缓存。 |
| `FetchEntityTypeList` 接口/DAO 当前不存在 | 需确认 `EntityTypeStorager` 接口；如缺失，需新增 `FetchEntityTypeList` 并在 `storage/rdb/entity/entity_type.go` 实现。 |
| 并发多个 conf-agent 同时拉取 | 结合版本控制：未变更时 `ExportConfig` 返回 `nil`，不会重复生成大数据包。 |
| 内存 Map 中 Entity 父级循环引用导致死循环 | Entity 层级按业务约束为树形结构；回溯时增加最大深度保护（如 10 层）。 |

---

## 9. 建议实施顺序

1. **本周**：确认 `EntityTypeStorager.FetchEntityTypeList` 接口与 DAO 实现是否齐全；不齐则补齐。
2. **本周**：在 `ai-gateway-api/model/imods/exporter.go` 中实现 Entity/QuotaPlan/EntityType 批量预加载，重构递归查询为内存回溯。
3. **本周**：补充 690 Key 量级的单元/集成测试，确保无回归。
4. **观察一周**：确认 `/inner-api/v1/configs/mod-api-key` 接口耗时稳定 < 500 ms，conf-agent 不再出现超时。
