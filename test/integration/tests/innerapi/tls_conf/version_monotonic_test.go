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

package innerapi_test

import (
	"encoding/json"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchServerDataConfVersion 导出 server_data_conf 并返回版本号。
// 版本串为 14 位定宽数字串（yyyyMMddHHmmss），字符串比较与时间比较等价。
func fetchServerDataConfVersion(t *testing.T, versionParam string) (string, json.RawMessage) {
	t.Helper()
	var resp *testutil.APIResponse
	var err error
	if versionParam == "" {
		resp, err = testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
	} else {
		resp, err = testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf", map[string]string{
			"version": versionParam,
		})
	}
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)

	var data struct {
		Version string          `json:"Version"`
		Config  json.RawMessage `json:"Config"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.NotEmpty(t, data.Version)
	return data.Version, data.Config
}

// TestInnerAPI_ExportVersionMonotonic（IN-TLS-1-009）验证导出配置版本号单调递增：
// 同一墙钟秒内两次内容变更（create → export → delete → export）必须产生严格
// 递增的版本号。预修复实现下两次导出版本串相同（同秒碰撞 + sign 短路），版本
// 通道被楔死，按版本轮询收敛的数据面消费者无限挂起（issue #142 / SC2101-TC018
// 六连失败的根因）。连续多轮使"同秒"窗口必然被覆盖。
func TestInnerAPI_ExportVersionMonotonic(t *testing.T) {
	var lastVersion string

	for i := 0; i < 6; i++ {
		clusterName := testutil.UniqueClusterName()
		_, err := testutil.CreateCluster(clusterName)
		require.NoError(t, err, "setup cluster failed")

		// 第一次变更（创建 cluster 改变了导出内容）后立即导出
		v1, _ := fetchServerDataConfVersion(t, "")
		require.NoError(t, testutil.DeleteCluster(clusterName))

		// 同秒内第二次变更（删除 cluster）后立即导出
		v2, _ := fetchServerDataConfVersion(t, "")

		assert.True(t, v2 > v1,
			"iteration %d: version must strictly increase after same-second double change (v1=%s, v2=%s)",
			i, v1, v2)
		lastVersion = v2
	}

	// 收敛语义：携带旧版本号拉取，内容已变更时必须返回新版本而非 null
	clusterName := testutil.UniqueClusterName()
	_, err := testutil.CreateCluster(clusterName)
	require.NoError(t, err, "setup cluster failed")
	defer testutil.DeleteCluster(clusterName)

	v3, config := fetchServerDataConfVersion(t, lastVersion)
	require.NotEqual(t, "null", string(config), "pull with stale version must return new content")
	assert.True(t, v3 > lastVersion, "converged version must exceed the stale one (v3=%s, stale=%s)", v3, lastVersion)
}
