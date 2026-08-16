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

package product_cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
)

func TestBindSubClusterValidate(t *testing.T) {
	cases := []struct {
		name    string
		param   *BindSubCluster
		wantErr bool
	}{
		{
			name:    "valid",
			param:   &BindSubCluster{ClusterName: lib.PString("cluster_1"), SubClusters: []string{"sub_1"}},
			wantErr: false,
		},
		{
			name:    "invalid cluster name",
			param:   &BindSubCluster{ClusterName: lib.PString("-cluster"), SubClusters: []string{"sub_1"}},
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
