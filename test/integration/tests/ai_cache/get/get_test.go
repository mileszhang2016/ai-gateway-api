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

package ai_cache_get_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	aiCacheRulesPath = "/open-api/v1/ai-cache-rules"
	validAICacheCond = `req_path_in("/v1/chat/completions", false)`
)

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

func putAICacheRules(t *testing.T, body map[string]interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Put(aiCacheRulesPath, body)
	require.NoError(t, err)
	return resp
}

func getAICacheRules(t *testing.T) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Get(aiCacheRulesPath)
	require.NoError(t, err)
	testutil.AssertSuccess(t, resp)
	return resp
}

// ---------- AC-2-001 空集合形状（家族11：顶层键含 rules 且为 [] 非 null） ----------

func TestAICacheRules_Get_EmptyShape(t *testing.T) {
	resp := putAICacheRules(t, map[string]interface{}{"rules": []interface{}{}})
	testutil.AssertSuccess(t, resp)

	getResp := getAICacheRules(t)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(getResp.Data, &data))

	// 顶层键必须包含 rules，且值为 []（非 null/缺失）——防客户端按零值解码。
	rulesRaw, ok := data["rules"]
	require.True(t, ok, "top-level key rules must be present (家族11)")
	rules, ok := rulesRaw.([]interface{})
	require.True(t, ok, "rules must be an array (not null)")
	assert.Len(t, rules, 0)
	assert.Contains(t, string(getResp.Data), `"rules":[]`)
}

// ---------- AC-2-002 顺序：GET 顺序=提交顺序 ----------

func TestAICacheRules_Get_Order(t *testing.T) {
	n1 := testutil.UniqueName("ac-2-002-first")
	n2 := testutil.UniqueName("ac-2-002-second")
	n3 := testutil.UniqueName("ac-2-002-third")

	resp := putAICacheRules(t, map[string]interface{}{"rules": []interface{}{
		map[string]interface{}{"name": n1, "cond": validAICacheCond, "cache_ttl": 100},
		map[string]interface{}{"name": n2, "cond": validAICacheCond, "cache_ttl": 200},
		map[string]interface{}{"name": n3, "cond": validAICacheCond, "cache_ttl": 300},
	}})
	testutil.AssertSuccess(t, resp)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(getAICacheRules(t).Data, &data))
	rules, ok := data["rules"].([]interface{})
	require.True(t, ok)
	require.Len(t, rules, 3)

	gotNames := make([]string, 0, 3)
	for i, item := range rules {
		rule, ok := item.(map[string]interface{})
		require.True(t, ok, "rules[%d] should be object", i)
		gotNames = append(gotNames, rule["name"].(string))
	}
	assert.Equal(t, []string{n1, n2, n3}, gotNames, "GET order must equal submission order (first-match-wins)")
}

// ---------- AC-2-003 幂等：连续两次 GET 逐字段一致 ----------

func TestAICacheRules_Get_Idempotent(t *testing.T) {
	resp := putAICacheRules(t, map[string]interface{}{"rules": []interface{}{
		map[string]interface{}{
			"name": testutil.UniqueName("ac-2-003"), "cond": validAICacheCond,
			"cache_key_strategy": "allQuestions", "cache_ttl": 3600,
		},
	}})
	testutil.AssertSuccess(t, resp)

	first := getAICacheRules(t)
	second := getAICacheRules(t)

	var firstData, secondData map[string]interface{}
	require.NoError(t, json.Unmarshal(first.Data, &firstData))
	require.NoError(t, json.Unmarshal(second.Data, &secondData))
	assert.Equal(t, firstData, secondData, "two consecutive GETs must be field-by-field identical")
}
