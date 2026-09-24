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

package operation_log_test

import (
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOperationLog_ClusterUpdateSnapshotAPIVocabulary 是 issue #205 的回归锚点：
// cluster 更新审计快照（change_summary.before/after）的键名必须与 cluster
// Open API 词汇一致（小写 JSON 字段名），diff_keys 精确等于实际提交的字段
// （SC2101-TC046 assert-update-logs cluster 目标）。
func TestOperationLog_ClusterUpdateSnapshotAPIVocabulary(t *testing.T) {
	client := testutil.GetClient()

	// 1. 创建 run-scoped provider + cluster。
	clusterName := testutil.UniqueClusterName()
	_, err := testutil.CreateCluster(clusterName)
	require.NoError(t, err, "create cluster failed")
	defer testutil.DeleteCluster(clusterName)

	// 2. PATCH 仅 description，触发 update 审计。
	resp, err := client.Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
		"description": "updated cluster",
	})
	require.NoError(t, err, "update cluster request failed")
	testutil.AssertSuccess(t, resp)

	// 3. 收敛该 cluster 的 update 成功日志。
	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "cluster",
		"action":        "update",
		"resource_name": clusterName,
		"status":        "1",
	}, 0)
	require.NoError(t, err, "expected cluster update operation log not found")

	// 4. diff_keys 精确匹配：幻影键或未提交字段必须导致失败（issue #201 的
	// Contains 盲区教训，禁止降级为 Contains）。
	require.NotNil(t, entry.ChangeSummary)
	diffKeys, ok := entry.ChangeSummary["diff_keys"].([]interface{})
	require.True(t, ok, "diff_keys should be an array")
	assert.ElementsMatch(t, []interface{}{"description"}, diffKeys)

	// 5. after 含 URI 绑定的 name 与提交的 description，键名均为 API 小写词汇。
	// name 值与 before 一致，按 diff 语义不计入 diff_keys（issue 证据形态：
	// after={"Description":...,"Name":...} 的小写化）。
	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "after should be an object")
	assert.ElementsMatch(t, []string{"description", "name"}, mapKeysOf(after))
	assert.Equal(t, "updated cluster", after["description"])
	assert.Equal(t, clusterName, after["name"])

	// 6. before 为 API 词汇快照：无大写 Go 字段名键、无内部记账字段。
	before, ok := entry.ChangeSummary["before"].(map[string]interface{})
	require.True(t, ok, "before should be an object")
	for k := range before {
		require.NotEmpty(t, k)
		assert.False(t, k[0] >= 'A' && k[0] <= 'Z',
			"before key %q must be API lowercase vocabulary, not a Go field name", k)
	}
	for _, internal := range []string{"id", "ready", "product_id", "scheduler", "sub_clusters"} {
		assert.NotContains(t, before, internal)
	}
	assert.Contains(t, before, "name")
	assert.Contains(t, before, "balance_mode")
	assert.Contains(t, before, "llm_config")
}

// TestOperationLog_ClusterUpdateSnapshotEnumValue 验证快照值表示与 API 对齐
// （issue #205 值表示层）：sticky_sessions.hash_strategy 落字符串枚举
// （"CLIENT_ID_ONLY"），而非内部枚举整数。
func TestOperationLog_ClusterUpdateSnapshotEnumValue(t *testing.T) {
	client := testutil.GetClient()

	clusterName := testutil.UniqueClusterName()
	_, err := testutil.CreateCluster(clusterName)
	require.NoError(t, err, "create cluster failed")
	defer testutil.DeleteCluster(clusterName)

	resp, err := client.Patch("/open-api/v1/clusters/"+clusterName, map[string]interface{}{
		"sticky_sessions": map[string]interface{}{
			"enabled":       true,
			"hash_strategy": "CLIENT_ID_ONLY",
			"hash_header":   "x-uid",
		},
	})
	require.NoError(t, err, "update cluster request failed")
	testutil.AssertSuccess(t, resp)

	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "cluster",
		"action":        "update",
		"resource_name": clusterName,
		"status":        "1",
	}, 0)
	require.NoError(t, err, "expected cluster update operation log not found")

	require.NotNil(t, entry.ChangeSummary)
	diffKeys, ok := entry.ChangeSummary["diff_keys"].([]interface{})
	require.True(t, ok, "diff_keys should be an array")
	assert.ElementsMatch(t, []interface{}{"sticky_sessions"}, diffKeys)

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "after should be an object")
	sticky, ok := after["sticky_sessions"].(map[string]interface{})
	require.True(t, ok, "sticky_sessions should be an object")
	assert.Equal(t, "CLIENT_ID_ONLY", sticky["hash_strategy"])
	assert.Equal(t, true, sticky["enabled"])
	assert.Equal(t, "x-uid", sticky["hash_header"])
}

func mapKeysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
