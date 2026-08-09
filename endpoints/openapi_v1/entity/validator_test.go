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

package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

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

func TestValidateEntityParam(t *testing.T) {
	validName := "entity_1"
	validType := "dept_1"
	invalidNameWithSpace := " entity_1"
	invalidType := "Dep"
	duplicateRuleName := "r1"
	cond := "default_t()"
	cluster := "cluster_1"
	weight := 100

	cases := []struct {
		name            string
		param           *quota.EntityParam
		requireNameType bool
		wantErr         bool
	}{
		{
			name: "create valid",
			param: &quota.EntityParam{
				Name:       &validName,
				Type:       &validType,
				RouteRules: validRouteRules(),
			},
			requireNameType: true,
			wantErr:         false,
		},
		{
			name:            "create missing name",
			param:           &quota.EntityParam{Type: &validType},
			requireNameType: true,
			wantErr:         true,
		},
		{
			name:            "create missing type",
			param:           &quota.EntityParam{Name: &validName},
			requireNameType: true,
			wantErr:         true,
		},
		{
			name: "create invalid name leading space",
			param: &quota.EntityParam{
				Name:       &invalidNameWithSpace,
				Type:       &validType,
				RouteRules: validRouteRules(),
			},
			requireNameType: true,
			wantErr:         true,
		},
		{
			name: "create invalid type",
			param: &quota.EntityParam{
				Name:       &validName,
				Type:       &invalidType,
				RouteRules: validRouteRules(),
			},
			requireNameType: true,
			wantErr:         true,
		},
		{
			name: "create duplicate route rule names",
			param: &quota.EntityParam{
				Name: &validName,
				Type: &validType,
				RouteRules: &shared.RouteRulesParam{
					Rules: []*shared.AiRouteRuleParam{
						{Name: &duplicateRuleName, Cond: &cond, Targets: []*shared.AiRouteTargetParam{{ClusterName: &cluster, Weight: &weight}}},
						{Name: &duplicateRuleName, Cond: &cond, Targets: []*shared.AiRouteTargetParam{{ClusterName: &cluster, Weight: &weight}}},
					},
				},
			},
			requireNameType: true,
			wantErr:         true,
		},
		{
			name:            "update empty body valid",
			param:           &quota.EntityParam{},
			requireNameType: false,
			wantErr:         false,
		},
		{
			name: "update invalid name",
			param: &quota.EntityParam{
				Name: &invalidNameWithSpace,
			},
			requireNameType: false,
			wantErr:         true,
		},
		{
			name: "update invalid type",
			param: &quota.EntityParam{
				Type: &invalidType,
			},
			requireNameType: false,
			wantErr:         true,
		},
		{
			name: "update invalid quota plan unit",
			param: &quota.EntityParam{
				QuotaPlan: &shared.QuotaPlanParam{
					Quota: lib.PInt64(100),
					Unit:  lib.PString("invalid"),
				},
			},
			requireNameType: false,
			wantErr:         true,
		},
		{
			name: "update invalid rate limit window",
			param: &quota.EntityParam{
				RateLimitPolicy: &shared.RateLimitPolicyParam{
					Enabled: lib.PBool(true),
					Rules: &shared.RateLimitRules{
						TpmConfigs: []shared.TPMConfig{
							{Name: "t1", Model: "*", WindowMinutes: 0, MaxTokens: 100, StepMinutes: 1},
						},
					},
				},
			},
			requireNameType: false,
			wantErr:         true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEntityParam(tc.param, tc.requireNameType)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
