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

package ai_route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductRouteRuleParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *ProductRouteRuleParam
		wantErr bool
	}{
		{
			name:    "empty",
			param:   &ProductRouteRuleParam{},
			wantErr: false,
		},
		{
			name: "valid advance rule",
			param: &ProductRouteRuleParam{
				AdvanceRouteRules: []*AdvanceRouteRule{
					{Name: "r1", ClusterName: "cluster_1"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid basic rule",
			param: &ProductRouteRuleParam{
				BasicRouteRules: []*BasicRouteRule{
					{ClusterName: "cluster_1"},
				},
			},
			wantErr: false,
		},
		{
			name: "nil advance rule element",
			param: &ProductRouteRuleParam{
				AdvanceRouteRules: []*AdvanceRouteRule{nil},
			},
			wantErr: true,
		},
		{
			name: "invalid advance cluster name",
			param: &ProductRouteRuleParam{
				AdvanceRouteRules: []*AdvanceRouteRule{
					{Name: "r1", ClusterName: "-cluster"},
				},
			},
			wantErr: true,
		},
		{
			name: "nil basic rule element",
			param: &ProductRouteRuleParam{
				BasicRouteRules: []*BasicRouteRule{nil},
			},
			wantErr: true,
		},
		{
			name: "invalid basic cluster name",
			param: &ProductRouteRuleParam{
				BasicRouteRules: []*BasicRouteRule{
					{ClusterName: "cluster."},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.param.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
