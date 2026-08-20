// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quota

import (
	"context"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/rate_limit_policy"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
)

// NewEntityStoragerAdapter wraps entity.NewEntityStoragerAdapter for backward compatibility.
func NewEntityStoragerAdapter(entityStorager entity.EntityStorager) shared.EntityStorager {
	return entity.NewEntityStoragerAdapter(entityStorager)
}

// NewRateLimitPolicyStoragerAdapter wraps rate_limit_policy.NewRateLimitPolicyStoragerAdapter for backward compatibility.
func NewRateLimitPolicyStoragerAdapter(storager rate_limit_policy.RateLimitPolicyStorager) shared.RateLimitPolicyStorager {
	return rate_limit_policy.NewRateLimitPolicyStoragerAdapter(storager)
}

// quotaBalanceStoragerAdapter adapts quota.QuotaBalanceStorager to shared.QuotaBalanceStorager.
type quotaBalanceStoragerAdapter struct {
	inner QuotaBalanceStorager
}

// NewQuotaBalanceStoragerAdapter wraps a quota.QuotaBalanceStorager as a shared.QuotaBalanceStorager.
func NewQuotaBalanceStoragerAdapter(inner QuotaBalanceStorager) shared.QuotaBalanceStorager {
	return &quotaBalanceStoragerAdapter{inner: inner}
}

func (a *quotaBalanceStoragerAdapter) FetchQuotaBalance(ctx context.Context, quotaPlanID int64) (*shared.BalanceSummary, error) {
	param, err := a.inner.FetchQuotaBalance(ctx, &QuotaBalanceFilter{QuotaPlanID: &quotaPlanID})
	if err != nil || param == nil {
		return nil, err
	}
	return &shared.BalanceSummary{
		Used:      param.Used,
		Remaining: param.Remaining,
	}, nil
}

func (a *quotaBalanceStoragerAdapter) CreateQuotaBalance(ctx context.Context, quotaPlanID int64, remaining *float64) error {
	now := time.Now()
	_, err := a.inner.CreateQuotaBalance(ctx, &QuotaBalanceParam{
		QuotaPlanID: &quotaPlanID,
		Used:        lib.PFloat64(0),
		Remaining:   remaining,
		LastResetAt: &now,
	})
	return err
}

func (a *quotaBalanceStoragerAdapter) DeleteQuotaBalance(ctx context.Context, quotaPlanID int64) error {
	return a.inner.DeleteQuotaBalance(ctx, &QuotaBalanceFilter{QuotaPlanID: &quotaPlanID})
}
