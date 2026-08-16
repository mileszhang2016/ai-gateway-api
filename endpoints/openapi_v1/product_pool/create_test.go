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

package product_pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsertParamValidate(t *testing.T) {
	longName := string(make([]byte, 129))
	cases := []struct {
		name    string
		param   *UpsertParam
		wantErr bool
	}{
		{
			name:    "empty instances",
			param:   &UpsertParam{},
			wantErr: false,
		},
		{
			name: "valid single instance",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: "host1", IP: "192.0.2.1", Weight: 100, Ports: map[string]int{"Default": 8080}},
				},
			},
			wantErr: false,
		},
		{
			name: "name too long",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: longName, IP: "192.0.2.1", Weight: 100, Ports: map[string]int{"Default": 8080}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid ip/addr",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: "host1", IP: "-invalid", Weight: 100, Ports: map[string]int{"Default": 8080}},
				},
			},
			wantErr: true,
		},
		{
			name: "missing default port",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: "host1", IP: "192.0.2.1", Weight: 100, Ports: map[string]int{"HTTP": 8080}},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate hostname ip",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: "host1", IP: "192.0.2.1", Weight: 50, Ports: map[string]int{"Default": 8080}},
					{Hostname: "host1", IP: "192.0.2.1", Weight: 50, Ports: map[string]int{"Default": 8081}},
				},
			},
			wantErr: true,
		},
		{
			name: "zero weight defaulted to positive",
			param: &UpsertParam{
				Instances: []*Instance{
					{Hostname: "host1", IP: "192.0.2.1", Weight: 0, Ports: map[string]int{"Default": 8080}},
				},
			},
			wantErr: false,
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
