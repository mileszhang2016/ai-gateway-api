// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package icluster_conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterParamToMap_DropsOmittedFields guards the issue #201 family
// fix: ClusterParam pointer fields omitted from a partial update must not
// materialize as null entries in the audit snapshot (phantom diff_keys).
func TestClusterParamToMap_DropsOmittedFields(t *testing.T) {
	desc := "d1"
	m := clusterParamToMap(&ClusterParam{Description: &desc})
	require.NotNil(t, m)
	assert.Equal(t, map[string]interface{}{"Description": "d1"}, m)

	assert.Nil(t, clusterParamToMap(nil))
	assert.Nil(t, clusterToMap(nil))
}
