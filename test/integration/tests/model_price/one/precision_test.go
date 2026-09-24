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

package model_price_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 8 位小数价格（issue #102 回归锚点）：Go 默认 %g 序列化会输出科学计数法
// （1.5e-06 / 2.5e-07），中间环节按文本截断将造成配额扣减按错误价格执行。
// 现行合同（MP-1-009 及 PriceMap 定义注释）：科学计数法与十进制均为合法
// 输出形态，因此本用例锁定"数值往返无损"，而非强制某一种文本形态。
const (
	precInputPrice  = 0.0000015
	precOutputPrice = 0.00000025
)

// priceToken 从原始报文中提取 "key": <token> 的 token 并解析为 float64，
// 无论该 token 是十进制（0.0000015）还是科学计数法（1.5e-06）形态。
// 用于锁定"序列化文本可被下游无损解析"，即无人对文本做中间截断。
func priceToken(t *testing.T, raw []byte, key string) float64 {
	t.Helper()
	needle := strconv.Quote(key) + ":"
	idx := strings.LastIndex(string(raw), needle)
	require.NotEqual(t, -1, idx, "key %s not found in raw body", key)
	rest := string(raw)[idx+len(needle):]
	end := strings.IndexAny(rest, ",}")
	require.NotEqual(t, -1, end, "token end not found for key %s", key)
	token := strings.TrimSpace(rest[:end])
	value, err := strconv.ParseFloat(token, 64)
	require.NoError(t, err, "token %q for key %s should be parseable", token, key)
	return value
}

// TestModelPrice_EightDecimalPrecision（MP-5-004）验证 8 位小数价格在
// OpenAPI 响应与 InnerAPI 导出两条链路上的数值无损往返（issue #102）。
func TestModelPrice_EightDecimalPrecision(t *testing.T) {
	provider := testutil.UniqueName("provider")
	model := "deepseek-v3-8dec"

	// 显式创建 provider 且 models 覆盖被测模型：cluster 导出要求 llm 模型
	// 为 provider models 子集（参照 IN-1-003 配方），自动创建的默认 provider
	// models 为 deepseek-chat，会导致 cluster 从导出中缺席。
	_, err := testutil.CreateProvider(provider, map[string]interface{}{
		"models": []string{model},
	})
	require.NoError(t, err, "setup provider failed")
	defer testutil.DeleteProvider(provider)

	id, err := testutil.CreateModelPrice(map[string]interface{}{
		"provider":   provider,
		"model":      model,
		"base_model": model,
		"mode":       "chat",
		"prices": map[string]interface{}{
			"input_cost_per_token":  precInputPrice,
			"output_cost_per_token": precOutputPrice,
		},
	})
	require.NoError(t, err, "setup model price failed")
	defer testutil.DeleteModelPrice(id)

	// 创建引用该 provider 与模型的 cluster，使 AIConf.ModelTable 导出该价格
	clusterName := testutil.UniqueClusterName()
	resp, err := testutil.GetClient().Post("/open-api/v1/clusters", map[string]interface{}{
		"name": clusterName,
		"llm_config": map[string]interface{}{
			"models":   []string{model},
			"provider": provider,
		},
	})
	require.NoError(t, err, "create cluster failed")
	testutil.AssertSuccess(t, resp)
	defer testutil.DeleteCluster(clusterName)

	t.Run("MP-5-004a OpenAPI 组合键查询数值无损", func(t *testing.T) {
		resp, err := testutil.GetClient().Get("/open-api/v1/model-prices", map[string]string{
			"provider": provider,
			"model":    model,
			"mode":     "chat",
		})
		require.NoError(t, err)
		testutil.AssertSuccess(t, resp)

		// 反序列化值精确（防中间截断：如被舍入为 0.000002 即 red）
		var one map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Data, &one))
		prices, ok := one["prices"].(map[string]interface{})
		require.True(t, ok, "prices should be an object")
		assert.InDelta(t, precInputPrice, prices["input_cost_per_token"], 1e-18)
		assert.InDelta(t, precOutputPrice, prices["output_cost_per_token"], 1e-18)

		// 原始报文 token 可无损解析（文本形态允许十进制或科学计数法）
		assert.InDelta(t, precInputPrice, priceToken(t, resp.RawBody, "input_cost_per_token"), 1e-18)
		assert.InDelta(t, precOutputPrice, priceToken(t, resp.RawBody, "output_cost_per_token"), 1e-18)
	})

	t.Run("MP-5-004b InnerAPI 导出 ModelTable 数值无损", func(t *testing.T) {
		export, err := testutil.GetClient().Get("/inner-api/v1/configs/tls_conf/server_data_conf")
		require.NoError(t, err)
		testutil.AssertSuccess(t, export)

		// 定位导出中该 cluster 的 AIConf.ModelTable 对应模型条目
		var data map[string]interface{}
		require.NoError(t, json.Unmarshal(export.Data, &data))
		config := data["ClusterConf"].(map[string]interface{})["Config"].(map[string]interface{})
		cluster, ok := config[clusterName].(map[string]interface{})
		require.True(t, ok, "cluster %s should exist in export", clusterName)
		aiConf, ok := cluster["AIConf"].(map[string]interface{})
		require.True(t, ok, "AIConf should exist")
		table, ok := aiConf["ModelTable"].(map[string]interface{})
		require.True(t, ok, "ModelTable should be exported for priced cluster")

		found := false
		for _, item := range table["Models"].([]interface{}) {
			entry := item.(map[string]interface{})
			if entry["Model"] != model {
				continue
			}
			found = true
			prices, ok := entry["Prices"].(map[string]interface{})
			require.True(t, ok, "Prices should exist for model %s", model)
			assert.InDelta(t, precInputPrice, prices["input_cost_per_token"], 1e-18)
			assert.InDelta(t, precOutputPrice, prices["output_cost_per_token"], 1e-18)
		}
		require.True(t, found, "model %s should exist in exported ModelTable", model)

		// 原始导出报文 token 可无损解析（BFE 配置文本不被中间截断）
		assert.InDelta(t, precInputPrice, priceToken(t, export.RawBody, "input_cost_per_token"), 1e-18)
		t.Logf("export raw snippet: %s", snippetAround(export.RawBody, "input_cost_per_token"))
	})
}

// snippetAround 返回 key 首次出现处前后各 80 字符的报文片段，仅供失败诊断日志。
func snippetAround(raw []byte, key string) string {
	idx := strings.Index(string(raw), strconv.Quote(key))
	if idx == -1 {
		return fmt.Sprintf("key %s not found", key)
	}
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + len(key) + 100
	if end > len(raw) {
		end = len(raw)
	}
	return string(raw[start:end])
}
