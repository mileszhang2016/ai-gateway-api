package shared

import "context"

type EntitySummary struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type EntityFilter struct {
	EntityID *string `form:"id"`
	Name     *string `form:"name"`
	Type     *string `form:"type"`
	ParentID *string `form:"parent_id"`
	Page     *int    `form:"page"`
	PageSize *int    `form:"page_size"`
}

type EntityStorager interface {
	FetchEntity(ctx context.Context, filter *EntityFilter) (*EntitySummary, error)
}

type BalanceSummary struct {
	Used      *int64 `json:"used"`
	Remaining *int64 `json:"remaining"`
}

type QuotaBalanceStorager interface {
	FetchQuotaBalance(ctx context.Context, quotaPlanID int64) (*BalanceSummary, error)
	CreateQuotaBalance(ctx context.Context, quotaPlanID int64, remaining *int64) error
	DeleteQuotaBalance(ctx context.Context, quotaPlanID int64) error
}

type QuotaPlanParam struct {
	Unlimited             *bool           `json:"unlimited"`
	PassWhenNoEnoughQuota *bool           `json:"pass_when_no_enough_quota"`
	Quota                 *int64          `json:"quota"`
	Unit                  *string         `json:"unit"`
	ResetPeriod           *string         `json:"reset_period"`
	Balance               *BalanceSummary `json:"balance,omitempty"`
}

type TPMConfig struct {
	Name          string `json:"name"`
	Model         string `json:"model"`
	WindowMinutes int    `json:"window_minutes"`
	MaxTokens     int    `json:"max_tokens"`
	StepMinutes   int    `json:"step_minutes"`
}

type RPMConfig struct {
	Name          string `json:"name"`
	Model         string `json:"model"`
	WindowMinutes int    `json:"window_minutes"`
	MaxRequests   int    `json:"max_requests"`
}

type RateLimitRules struct {
	TpmConfigs     []TPMConfig `json:"tpm"`
	RpmConfigs     []RPMConfig `json:"rpm"`
	MaxConcurrency *int        `json:"max_concurrency"`
}

type RateLimitPolicyParam struct {
	Enabled *bool           `json:"enabled"`
	Rules   *RateLimitRules `json:"rules"`
}

type QuotaPlanStorager interface {
	CreateQuotaPlan(ctx context.Context, param *QuotaPlanParam) (int64, error)
	UpdateQuotaPlan(ctx context.Context, id int64, param *QuotaPlanParam) (int64, error)
	DeleteQuotaPlan(ctx context.Context, id int64) error
	FetchQuotaPlan(ctx context.Context, id int64) (*QuotaPlanParam, error)
}

type RateLimitPolicyStorager interface {
	CreateRateLimitPolicy(ctx context.Context, param *RateLimitPolicyParam) (int64, error)
	UpdateRateLimitPolicy(ctx context.Context, id int64, param *RateLimitPolicyParam) (int64, error)
	DeleteRateLimitPolicy(ctx context.Context, id int64) error
	FetchRateLimitPolicy(ctx context.Context, id int64) (*RateLimitPolicyParam, error)
}
