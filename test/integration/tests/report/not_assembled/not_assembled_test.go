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

package not_assembled_test

import (
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
	var err error
	// 默认测试 conf 未配置 [Report]，Report 模块不装配，/report/* 不注册。
	sm, err = testutil.StartServer()
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

// TestReportEndpoints_NotAssembled 验证 [Report] 缺省时五个报表端点整体不注册，
// 请求落到 mux 的 NotFoundHandler，返回 JSON 404 信封（ErrNum=404）。
func TestReportEndpoints_NotAssembled(t *testing.T) {
	client := testutil.GetClient()

	window := map[string]string{
		"start": "1782345600",
		"end":   "1782349200",
	}
	endpoints := []struct {
		path  string
		query map[string]string
	}{
		{"/open-api/v1/report/overview", window},
		{"/open-api/v1/report/timeseries", withMetric(window, "qps")},
		{"/open-api/v1/report/rankings", withDimension(window, "model")},
		{"/open-api/v1/report/distribution", withDimension(window, "status")},
		{"/open-api/v1/report/logs", window},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			resp, err := client.Get(ep.path, ep.query)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, 404, resp.ErrNum,
				"report module is not assembled, %s must be 404, got ErrMsg=%s", ep.path, resp.ErrMsg)
		})
	}
}

// TestReportEndpoints_NotAssembled_UnknownSubPath 验证 /report/ 下的未知子路径同样 404。
func TestReportEndpoints_NotAssembled_UnknownSubPath(t *testing.T) {
	client := testutil.GetClient()

	resp, err := client.Get("/open-api/v1/report/nonexistent", map[string]string{
		"start": "1782345600",
		"end":   "1782349200",
	})
	require.NoError(t, err)
	assert.Equal(t, 404, resp.ErrNum)
}

func withMetric(base map[string]string, metric string) map[string]string {
	out := copyQuery(base)
	out["metric"] = metric
	return out
}

func withDimension(base map[string]string, dimension string) map[string]string {
	out := copyQuery(base)
	out["dimension"] = dimension
	return out
}

func copyQuery(base map[string]string) map[string]string {
	out := make(map[string]string, len(base)+1)
	for k, v := range base {
		out[k] = v
	}
	return out
}
