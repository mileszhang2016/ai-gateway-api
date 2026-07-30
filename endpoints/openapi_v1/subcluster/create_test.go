// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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

package subcluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yf-networks/ai-gateway-api/lib"
)

func TestCreateParamValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *CreateParam
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &CreateParam{Name: lib.PString("cluster_1"), InstancePool: lib.PString("pool_1")},
			wantErr: false,
		},
		{
			name:    "invalid name",
			param:   &CreateParam{Name: lib.PString("-cluster"), InstancePool: lib.PString("pool_1")},
			wantErr: true,
		},
		{
			name:    "invalid instance_pool",
			param:   &CreateParam{Name: lib.PString("cluster_1"), InstancePool: lib.PString("pool.")},
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
