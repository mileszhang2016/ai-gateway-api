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
	"os"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	aiCacheOpenPath   = "/open-api/v1/ai-cache-rules"
	aiCacheExportPath = "/inner-api/v1/configs/ai-cache-rule"
	validAICacheCond  = `req_path_in("/v1/chat/completions", false)`

	// productKey 为测试环境 AIRouteInnerProductName 配置值（合同默认 AI_product）。
	productKey = "AI_product"
)

// exportRuleKeys 为 Inner 导出规则的一期 5 个 tag（大写驼峰，与 Open API 词汇不同）。
var exportRuleKeys = []string{"cond", "cacheKeyStrategy", "cacheTTL", "maxBodyBytes", "maxValueBytes"}

// phaseTwoFields 为一期不得导出的二期/预留字段（合同锁定）。
var phaseTwoFields = []string{
	"cacheKeyFrom", "cacheValueFrom", "cacheStreamValueFrom",
	"cacheToolCallsFrom", "responseTemplate", "streamResponseTemplate",
}

var sm *testutil.ServerManager

func TestMain(m *testing.M) {
	var err error
	sm, err = testutil.StartServer()
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
	code := m.Run()
	sm.Shutdown()
	os.Exit(code)
}

type aiCacheExportData struct {
	Config  map[string][]map[string]interface{} `json:"Config"`
	Version string                              `json:"Version"`
}

func putAICacheRules(t *testing.T, rules []interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Put(aiCacheOpenPath, map[string]interface{}{"rules": rules})
	require.NoError(t, err)
	return resp
}

func putAICacheRulesOK(t *testing.T, rules []interface{}) {
	t.Helper()
	resp := putAICacheRules(t, rules)
	testutil.AssertSuccess(t, resp)
}

// exportAICache 拉取导出；version 为空表示首拉。返回原始响应（Data 可能为 null）。
func exportAICache(t *testing.T, version string) *testutil.APIResponse {
	t.Helper()
	var resp *testutil.APIResponse
	var err error
	if version == "" {
		resp, err = testutil.GetClient().Get(aiCacheExportPath)
	} else {
		resp, err = testutil.GetClient().Get(aiCacheExportPath, map[string]string{"version": version})
	}
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	return resp
}

// exportAICacheData 首拉导出（Data 必须非 null）并解析。
func exportAICacheData(t *testing.T) *aiCacheExportData {
	t.Helper()
	resp := exportAICache(t, "")
	testutil.AssertDataNotEmpty(t, resp)
	var data aiCacheExportData
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.NotEmpty(t, data.Version)
	_, ok := data.Config[productKey]
	require.True(t, ok, "Config must contain product key %s", productKey)
	return &data
}

func productRules(t *testing.T, data *aiCacheExportData) []map[string]interface{} {
	t.Helper()
	rules, ok := data.Config[productKey]
	require.True(t, ok, "product key %s must be present", productKey)
	return rules
}

// assertExportRuleValues 断言导出规则 5 tag 精确集合与取值。
func assertExportRuleValues(t *testing.T, rule map[string]interface{}, want map[string]interface{}) {
	t.Helper()
	keys := make([]string, 0, len(rule))
	for k := range rule {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, exportRuleKeys, keys,
		"exported rule must carry exactly the 5 phase-1 tags")
	for k, v := range want {
		assert.Equal(t, v, rule[k], "export tag %s", k)
	}
}

// ---------- AICE-1-001 首拉（家族8：存在性 + 取值 + 原始 body 文本形态） ----------

func TestInnerAPI_AICache_ExportFirstPull(t *testing.T) {
	n1 := testutil.UniqueName("aice-1-001-defaults")
	n2 := testutil.UniqueName("aice-1-001-explicit")
	putAICacheRulesOK(t, []interface{}{
		// 规则 1：仅 name+cond，依赖默认值回填。
		map[string]interface{}{"name": n1, "cond": validAICacheCond},
		// 规则 2：显式非默认值。
		map[string]interface{}{
			"name": n2, "cond": validAICacheCond,
			"cache_key_strategy": "disabled", "cache_ttl": 86400,
			"max_body_bytes": 524288, "max_value_bytes": 262144,
		},
	})

	resp := exportAICache(t, "")
	testutil.AssertSuccess(t, resp)

	var data aiCacheExportData
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.NotEmpty(t, data.Version, "Version must be non-empty")

	rules, ok := data.Config[productKey]
	require.True(t, ok, "Config must contain product key %s", productKey)
	require.Len(t, rules, 2, "Config.%s length must equal submitted collection", productKey)

	// 顺序=提交顺序（first-match-wins 优先级）。
	assertExportRuleValues(t, rules[0], map[string]interface{}{
		"cond": validAICacheCond, "cacheKeyStrategy": "lastQuestion",
		"cacheTTL": float64(0), "maxBodyBytes": float64(1048576), "maxValueBytes": float64(1048576),
	})
	assertExportRuleValues(t, rules[1], map[string]interface{}{
		"cond": validAICacheCond, "cacheKeyStrategy": "disabled",
		"cacheTTL": float64(86400), "maxBodyBytes": float64(524288), "maxValueBytes": float64(262144),
	})

	// 原始响应 body 文本形态断言（家族8/#102：仅反序列化值比较会静默放过序列化错误）。
	body := string(resp.RawBody)
	assert.Contains(t, body, `"cacheKeyStrategy":"lastQuestion"`)
	assert.Contains(t, body, `"cacheTTL":0`)
	assert.Contains(t, body, `"maxBodyBytes":1048576`)
	assert.Contains(t, body, `"maxValueBytes":1048576`)
	assert.Contains(t, body, `"cacheKeyStrategy":"disabled"`)
	assert.Contains(t, body, `"cacheTTL":86400`)
}

// ---------- AICE-1-002 二期字段缺席（家族8 合同锁定） ----------

func TestInnerAPI_AICache_PhaseTwoFieldsAbsent(t *testing.T) {
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{
			"name": testutil.UniqueName("aice-1-002"), "cond": validAICacheCond,
			"cache_ttl": 3600,
		},
	})

	resp := exportAICache(t, "")
	testutil.AssertDataNotEmpty(t, resp)

	// 断言作用于原始 body 文本，覆盖 JSON 任意嵌套位置。
	body := string(resp.RawBody)
	for _, field := range phaseTwoFields {
		assert.NotContains(t, body, field, "phase-2 field %s must not appear in export (家族8 合同锁定)", field)
	}
}

// ---------- AICE-1-003 增量同步 ----------

func TestInnerAPI_AICache_Incremental(t *testing.T) {
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{"name": testutil.UniqueName("aice-1-003"), "cond": validAICacheCond},
	})

	first := exportAICacheData(t)
	v1 := first.Version

	// 携带刚返回的 version 再拉 → Data 为 null（未变化）。
	incrResp := exportAICache(t, v1)
	testutil.AssertDataNull(t, incrResp)

	// 不传 version 再拉 → 有数据且 Version 相同（内容未变则版本不变）。
	again := exportAICacheData(t)
	assert.Equal(t, v1, again.Version, "unchanged content must keep the same version")
	assert.Equal(t, productRules(t, first), productRules(t, again), "export content must be stable")
}

// ---------- AICE-1-004 版本单调（家族8/#142：同秒连续变更版本严格递增且收敛） ----------

func TestInnerAPI_AICache_VersionMonotonic(t *testing.T) {
	nA := testutil.UniqueName("aice-1-004-a")
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{"name": nA, "cond": validAICacheCond, "cache_ttl": 111},
	})
	v1 := exportAICacheData(t).Version

	// 同一秒内第二次变更。
	nB1 := testutil.UniqueName("aice-1-004-b1")
	nB2 := testutil.UniqueName("aice-1-004-b2")
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{"name": nB1, "cond": validAICacheCond, "cache_ttl": 222},
		map[string]interface{}{"name": nB2, "cond": validAICacheCond, "cache_ttl": 333},
	})

	// 轮询导出（间隔 500ms，最多 10 次）直到版本严格大于 v1（版本串为 14 位定宽，字典序即时间序）。
	var converged *aiCacheExportData
	for i := 0; i < 10; i++ {
		data := exportAICacheData(t)
		if data.Version > v1 {
			converged = data
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NotNil(t, converged, "version must strictly increase within bounded polls (家族8/#142)")

	// 内容收敛为最新集合（导出无 name 字段，以 cond + 各不相同的 cacheTTL 锁定顺序）。
	rules := productRules(t, converged)
	require.Len(t, rules, 2)
	assertExportRuleValues(t, rules[0], map[string]interface{}{
		"cond": validAICacheCond, "cacheKeyStrategy": "lastQuestion",
		"cacheTTL": float64(222), "maxBodyBytes": float64(1048576), "maxValueBytes": float64(1048576),
	})
	assertExportRuleValues(t, rules[1], map[string]interface{}{
		"cond": validAICacheCond, "cacheKeyStrategy": "lastQuestion",
		"cacheTTL": float64(333), "maxBodyBytes": float64(1048576), "maxValueBytes": float64(1048576),
	})
}

// ---------- AICE-1-005 清空导出（家族8：空数组非 null，product 键 present） ----------

func TestInnerAPI_AICache_ExportEmpty(t *testing.T) {
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{"name": testutil.UniqueName("aice-1-005-a"), "cond": validAICacheCond},
		map[string]interface{}{"name": testutil.UniqueName("aice-1-005-b"), "cond": validAICacheCond},
	})
	putAICacheRulesOK(t, []interface{}{})

	resp := exportAICache(t, "")
	testutil.AssertDataNotEmpty(t, resp)

	var data aiCacheExportData
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.NotEmpty(t, data.Version)

	// product 键 present 且为 []（非 null）。
	rules, ok := data.Config[productKey]
	require.True(t, ok, "product key %s must stay present for empty collection", productKey)
	require.NotNil(t, rules, "empty collection must export [] not null")
	assert.Len(t, rules, 0)
	assert.Contains(t, string(resp.RawBody), `"`+productKey+`":[]`)
}

// ---------- AICE-1-006 422 防泄漏（家族4：拒绝后导出完全一致、Version 不变） ----------

func TestInnerAPI_AICache_RejectedPutNoLeak(t *testing.T) {
	keeper := testutil.UniqueName("aice-1-006-keeper")
	putAICacheRulesOK(t, []interface{}{
		map[string]interface{}{"name": keeper, "cond": validAICacheCond, "cache_ttl": 777},
	})
	before := exportAICacheData(t)
	v1 := before.Version
	beforeRules := productRules(t, before)

	// PUT 被拒（非法 cond）。
	resp := putAICacheRules(t, []interface{}{
		map[string]interface{}{"name": testutil.UniqueName("aice-1-006-bad"), "cond": "not_a_valid_expr("},
	})
	testutil.AssertErrCode(t, resp, 422)

	// 带 v1 增量拉取 → Data null（版本未推进）。
	incrResp := exportAICache(t, v1)
	testutil.AssertDataNull(t, incrResp)

	// 首拉 → Version 不变且内容与拒绝前完全一致。
	after := exportAICacheData(t)
	assert.Equal(t, v1, after.Version, "rejected PUT must not bump export version (家族4)")
	assert.Equal(t, beforeRules, productRules(t, after),
		"rejected PUT must not leak into export content (家族4)")
}
