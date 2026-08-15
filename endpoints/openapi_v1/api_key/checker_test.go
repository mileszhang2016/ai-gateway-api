// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package api_key

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

func validAPIKeyParam() *icluster_conf.APIKeyParam {
	desc := "test api key"
	key := "ak-test_123"
	neverExpire := int64(-1)
	return &icluster_conf.APIKeyParam{
		Description: &desc,
		Key:         &key,
		ExpiredTime: &neverExpire,
	}
}

func validRouteRules() *shared.RouteRulesParam {
	name := "r1"
	cond := "default_t()"
	cluster := "cluster_1"
	weight := 100
	return &shared.RouteRulesParam{
		Rules: []*shared.AiRouteRuleParam{
			{
				Name: &name,
				Cond: &cond,
				Targets: []*shared.AiRouteTargetParam{
					{ClusterName: &cluster, Weight: &weight},
				},
			},
		},
	}
}

func TestCheckCreateAPIKey(t *testing.T) {
	longDesc := make([]byte, 513)
	invalidKey := "ak@123"
	invalidCIDR := "not-a-cidr"
	past := time.Now().Unix() - 1000
	invalidUnit := "invalid"

	cases := []struct {
		name    string
		param   *icluster_conf.APIKeyParam
		wantErr bool
	}{
		{"valid", validAPIKeyParam(), false},
		{"missing description", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Description = nil
			return p
		}(), true},
		{"empty description", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Description = lib.PString("")
			return p
		}(), true},
		{"description too long", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Description = lib.PString(string(longDesc))
			return p
		}(), true},
		{"missing key", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Key = nil
			return p
		}(), true},
		{"invalid key", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Key = &invalidKey
			return p
		}(), true},
		{"invalid subnet", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Subnet = []string{invalidCIDR}
			return p
		}(), true},
		{"expired time in past", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.ExpiredTime = &past
			return p
		}(), true},
		{"invalid quota plan unit", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.QuotaPlan = &shared.QuotaPlanParam{Quota: lib.PFloat64(100), Unit: &invalidUnit}
			return p
		}(), true},
		{"invalid route rules", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.RouteRules = &shared.RouteRulesParam{Rules: []*shared.AiRouteRuleParam{}}
			return p
		}(), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkCreateAPIKey(tc.param, "product_1")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckUpdateAPIKey(t *testing.T) {
	longDesc := make([]byte, 513)
	invalidKey := "ak@123"
	invalidCIDR := "not-a-cidr"
	invalidUnit := "invalid"
	zeroWindow := 0

	cases := []struct {
		name    string
		param   *icluster_conf.APIKeyParam
		wantErr bool
	}{
		{"empty update valid", &icluster_conf.APIKeyParam{}, false},
		{"valid description", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Key = nil
			return p
		}(), false},
		{"description too long", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.Description = lib.PString(string(longDesc))
			return p
		}(), true},
		{"invalid subnet", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.Subnet = []string{invalidCIDR}
			return p
		}(), true},
		{"invalid key", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.Key = &invalidKey
			return p
		}(), true},
		{"invalid quota plan unit", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.QuotaPlan = &shared.QuotaPlanParam{Quota: lib.PFloat64(100), Unit: &invalidUnit}
			return p
		}(), true},
		{"invalid rate limit window", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.RateLimitPolicy = &shared.RateLimitPolicyParam{
				Enabled: lib.PBool(true),
				Rules: &shared.RateLimitRules{
					TpmConfigs: []shared.TPMConfig{
						{Name: "t1", Model: "*", WindowMinutes: zeroWindow, MaxTokens: 100, StepMinutes: 1},
					},
				},
			}
			return p
		}(), true},
		{"duplicate route rule names", func() *icluster_conf.APIKeyParam {
			p := &icluster_conf.APIKeyParam{}
			p.RouteRules = validRouteRules()
			name := "r1"
			cond := "default_t()"
			cluster := "cluster_1"
			weight := 100
			p.RouteRules.Rules = append(p.RouteRules.Rules, &shared.AiRouteRuleParam{
				Name:    &name,
				Cond:    &cond,
				Targets: []*shared.AiRouteTargetParam{{ClusterName: &cluster, Weight: &weight}},
			})
			return p
		}(), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkUpdateAPIKey(tc.param, "product_1")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckFullUpdateAPIKey(t *testing.T) {
	cases := []struct {
		name    string
		param   *icluster_conf.APIKeyParam
		wantErr bool
	}{
		{"valid", validAPIKeyParam(), false},
		{"missing description", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Description = nil
			return p
		}(), true},
		{"empty description", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.Description = lib.PString("")
			return p
		}(), true},
		{"invalid quota plan", func() *icluster_conf.APIKeyParam {
			p := validAPIKeyParam()
			p.QuotaPlan = &shared.QuotaPlanParam{Quota: lib.PFloat64(100), Unit: lib.PString("invalid")}
			return p
		}(), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkFullUpdateAPIKey(tc.param, "product_1")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
