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
