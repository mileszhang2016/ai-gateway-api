// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package product_cluster

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
)

func TestClusterModel2ControlInstancePoolFields(t *testing.T) {
	cluster := &icluster_conf.Cluster{
		Name:        "my-cluster",
		Description: "test cluster",
		Basic: &icluster_conf.ClusterBasic{
			Protocol: lib.PString("https"),
			Connection: &icluster_conf.ClusterBasicConnection{
				MaxIdleConnPerRs:    0,
				CancelOnClientClose: false,
			},
			Retries: &icluster_conf.ClusterBasicRetries{
				MaxRetryInSubcluster: 2,
			},
			Buffers: &icluster_conf.ClusterBasicBuffers{
				ReqWriteBufferSize: 512,
				ReqFlushInterval:   icluster_conf.ClusterDefaultReqFlushInterval,
				ResFlushInterval:   icluster_conf.ClusterDefaultResFlushInterval,
			},
			Timeouts: &icluster_conf.ClusterBasicTimeouts{
				TimeoutConnServ:        50000,
				TimeoutResponseHeader:  50000,
				TimeoutReadbodyClient:  30000,
				TimeoutReadClientAgain: 30000,
				TimeoutWriteClient:     60000,
			},
		},
		StickySessions: &icluster_conf.ClusterStickySessions{
			SessionSticky: false,
			HashStrategy:  icluster_conf.ClusterHashStrategyClientIPOnlyI,
			HashHeader:    "",
		},
		PassiveHealthCheck: &icluster_conf.ClusterPassiveHealthCheck{
			Interval:   1000,
			Failnum:    3,
			Statuscode: 0,
			Host:       "",
			Uri:        "/",
		},
		LLMConfig: &icluster_conf.LLMConfig{
			Models: []string{"gpt-4"},
		},
		SubClusters: []*icluster_conf.SubCluster{
			{
				InstancePool: &icluster_conf.Pool{
					Instances: []icluster_conf.Instance{
						{
							Name:    "backend-1",
							Addr:    "10.0.0.1",
							Port:    8080,
							Weight:  50,
							Disable: false,
						},
						{
							Name:    "backend-2",
							Addr:    "10.0.0.2",
							Port:    8080,
							Weight:  50,
							Disable: true,
						},
					},
				},
			},
		},
	}

	rsp := clusterModel2Control(cluster)
	assert.Len(t, rsp.InstancePool, 2)

	first := rsp.InstancePool[0]
	assert.Equal(t, "backend-1", first.Name)
	assert.Equal(t, "10.0.0.1", first.Addr)
	assert.Equal(t, 8080, first.Port)
	assert.Equal(t, int64(50), first.Weight)

	second := rsp.InstancePool[1]
	assert.Equal(t, "backend-2", second.Name)
	assert.Equal(t, "10.0.0.2", second.Addr)
	assert.Equal(t, 8080, second.Port)
	assert.Equal(t, int64(50), second.Weight)

	// Serialize the response and verify it matches the public API doc:
	// only name/addr/port/weight should appear; disable must not leak out.
	data, err := json.Marshal(rsp)
	assert.NoError(t, err)

	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &raw))

	pool, ok := raw["instance_pool"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, pool, 2)

	for i, item := range pool {
		inst := item.(map[string]interface{})
		assert.Contains(t, inst, "name", "instance %d", i)
		assert.Contains(t, inst, "addr", "instance %d", i)
		assert.Contains(t, inst, "port", "instance %d", i)
		assert.Contains(t, inst, "weight", "instance %d", i)
		assert.NotContains(t, inst, "disable", "instance %d", i)
		assert.NotContains(t, inst, "Disable", "instance %d", i)
	}

	// Also ensure no old-issue-37 fields are present.
	body := string(data)
	assert.False(t, strings.Contains(body, `"hostname"`))
	assert.False(t, strings.Contains(body, `"ip"`))
	assert.False(t, strings.Contains(body, `"ports"`))
}
