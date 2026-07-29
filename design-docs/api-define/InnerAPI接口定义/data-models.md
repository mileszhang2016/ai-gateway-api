# 数据模型定义

## 1. ModAPIKeyRuleConf（实际实现）

```go
// ModAPIKeyRuleConf 定义 API-Key 规则配置结构
type ModAPIKeyRuleConf struct {
    Version    *string                          `json:"version"`
    Config     map[string][]*TokenRuleFile      `json:"config"`
    QuotaPlans map[string][]*QuotaPlan           `json:"QuotaPlans"`
    Tokens     map[string]map[string]*TokenFile `json:"tokens"`
}
```

## 2. TokenFile（实际实现）

```go
// TokenFile 定义导出到 BFE 的 API-Key 信息结构
type TokenFile struct {
    Key            string      `json:"key"`
    Enabled        int         `json:"enabled"`
    Status         int         `json:"status"`
    Name           string      `json:"name"`
    UpdateTime     int64       `json:"update_time"`
    ExpiredTime    int64       `json:"expired_time"`
    UnlimitedQuota bool        `json:"unlimited_quota"`
    Models         *string     `json:"allow_models"`
    BlockModels    *string     `json:"block_models"`
    Subnet         *string     `json:"subnet"`
    Tags           []ApikeyTag
    QuotaPlans     []string    `json:"quota_plans"`
}

type ApikeyTag struct {
    TagName  string  // Entity 类型，如 department, team
    TagValue string  // Entity 名称，如 dept-engineering
}
```

## 3. QuotaPlan（实际实现）

```go
// QuotaPlan 定义配额计划结构
type QuotaPlan struct {
    Id          string  // 配额计划 ID
    Unlimited   bool    // 是否无限配额
    PassNoQuota bool    // 配额不足时是否放行
    RedisKey    string  // Redis Key，格式 QUOTA_{id}
    CreateTime  int64   // 创建时间戳
    ExpiredTime int64   // 过期时间戳，-1 表示永不过期
    Quota       int64   // 配额总量
    ResetMode   int     // 0: 非周期性; 1: 周期性
}
```

## 4. ExportRateLimitPolicyConfig

```go
// ExportRateLimitPolicyConfig 定义限流策略导出结构（与 BFE 动态配置格式一致）
type ExportRateLimitPolicyConfig struct {
    Config                        map[string][]*ExportRouteRule     `json:"Config"`
    RateLimitPolicies             map[string]*ExportRateLimitPolicy `json:"RateLimitPolicies"`
    ApikeyRateLimitPolicyBindings map[string][]string               `json:"ApikeyRateLimitPolicyBindings"`
    Version                       string                            `json:"Version"`
}

// ExportRouteRule 定义导出的路由规则
type ExportRouteRule struct {
    Cond      string           `json:"cond"`
    HitAction *ExportHitAction `json:"hit_action"`
}

// ExportHitAction 定义导出的命中动作
type ExportHitAction struct {
    Cmd    string   `json:"cmd"`
    Params []string `json:"params"`
}

// ExportRateLimitPolicy 定义导出的单个限流策略配置
type ExportRateLimitPolicy struct {
    Name    string               `json:"name"`
    Enabled bool                 `json:"enabled"`
    Rules   *ExportRateLimitRules `json:"rules"`
}

// ExportRateLimitRules 定义导出的限流规则集合
type ExportRateLimitRules struct {
    TPM            []ExportTPMConfig `json:"tpm"`
    RPM            []ExportRPMConfig `json:"rpm"`
    MaxConcurrency int               `json:"max_concurrency"`
}

// ExportTPMConfig 定义导出的 TPM 限制配置
type ExportTPMConfig struct {
    Name          string   `json:"name"`
    Models        []string `json:"models"`
    WindowMinutes int      `json:"window_minutes"`
    MaxTokens     int      `json:"max_tokens"`
    StepMinutes   int      `json:"step_minutes"`
}

// ExportRPMConfig 定义导出的 RPM 限制配置
type ExportRPMConfig struct {
    Name          string   `json:"name"`
    Models        []string `json:"models"`
    WindowMinutes int      `json:"window_minutes"`
    MaxRequests   int      `json:"max_requests"`
    Burst         int      `json:"burst"`
}
```

## 5. AiRouteDataExport

```go
// AiRouteDataExport 定义 AI 路由配置导出结构（与 BFE ai_route.data 格式一致）
type AiRouteDataExport struct {
    Version                  string                `json:"Version"`
    RouteRules               map[string]*RouteTableExport `json:"RouteRules"`
    ApikeyRouteTableBindings map[string][]string   `json:"ApikeyRouteTableBindings"`
}

// RouteTableExport 定义单张路由表
type RouteTableExport struct {
    Type   string            `json:"type"`
    Owner  string            `json:"owner"`
    Rules  []*RouteRuleExport `json:"rules"`
}

// RouteRuleExport 定义路由规则
type RouteRuleExport struct {
    Name      string                `json:"name"`
    Cond      string                `json:"Cond"`
    Targets   []*AiRouteTargetExport `json:"targets"`
    Fallbacks []*AiRouteFallbackExport `json:"fallbacks"`
}

// AiRouteTargetExport 定义转发目标
type AiRouteTargetExport struct {
    ClusterName string `json:"ClusterName"`
    Model       string `json:"Model"`
    Weight      int    `json:"Weight"`
}

// AiRouteFallbackExport 定义降级目标
type AiRouteFallbackExport struct {
    ClusterName string `json:"ClusterName"`
    Model       string `json:"Model"`
}
```

## 6. ModBodyProcessConf（实际实现）

```go
// ModBodyProcessConf 定义 mod_body_process 配置导出结构
type ModBodyProcessConf struct {
    Version *string             `json:"Version"`
    Config  map[string][]string `json:"Config"`
}
```

---

