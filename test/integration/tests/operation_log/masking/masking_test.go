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

package masking_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// changeSummaryBytes 将 change_summary 序列化，用于"不含明文"整体断言。
func changeSummaryBytes(t *testing.T, entry *testutil.OperationLogEntry) string {
	t.Helper()
	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	return string(serialized)
}

// TestOperationLog_TokenCreateMasksToken 验证 Token 创建审计日志的
// change_summary.after.token 为全掩码，不出现与接口响应一致的明文（issue #162 /
// SC2101-TC049 回归锚点）。
func TestOperationLog_TokenCreateMasksToken(t *testing.T) {
	tokenName := testutil.UniqueTokenName()
	rawToken, err := testutil.CreateToken(tokenName, "Support")
	require.NoError(t, err, "create token failed")
	require.NotEmpty(t, rawToken)
	defer testutil.DeleteToken(tokenName)

	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "token",
		"action":        "create",
		"resource_name": tokenName,
	}, 0)
	require.NoError(t, err, "token create operation log not found")

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "change_summary.after should be an object")
	for _, key := range []string{"id", "name", "scope", "token"} {
		assert.Contains(t, after, key, "after should keep key %s", key)
	}
	assert.Equal(t, "******", after["token"], "token must be fully masked in audit log")

	serialized := changeSummaryBytes(t, entry)
	assert.NotContains(t, serialized, rawToken, "audit log must not contain raw token")
}

// TestOperationLog_TokenDeleteMasksToken 验证 Token 删除审计日志的
// change_summary.before.token 为全掩码——删除资源不能阻断已落库明文的提权面（issue #162）。
func TestOperationLog_TokenDeleteMasksToken(t *testing.T) {
	tokenName := testutil.UniqueTokenName()
	rawToken, err := testutil.CreateToken(tokenName, "System")
	require.NoError(t, err, "create token failed")
	require.NoError(t, testutil.DeleteToken(tokenName), "delete token failed")

	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "token",
		"action":        "delete",
		"resource_name": tokenName,
	}, 0)
	require.NoError(t, err, "token delete operation log not found")

	before, ok := entry.ChangeSummary["before"].(map[string]interface{})
	require.True(t, ok, "change_summary.before should be an object")
	assert.Equal(t, "******", before["token"], "token must be fully masked in audit log")

	serialized := changeSummaryBytes(t, entry)
	assert.NotContains(t, serialized, rawToken, "audit log must not contain raw token")
}

// TestOperationLog_ProviderKeysMasked 验证 Provider 创建/更新审计日志中
// keys[].key（上游 bearer 凭证）按既有 key 契约部分脱敏，before/after 两侧均不泄漏
// 明文（issue #162 同族收口）。
func TestOperationLog_ProviderKeysMasked(t *testing.T) {
	providerName := testutil.UniqueProviderName()
	_, err := testutil.CreateProvider(providerName)
	require.NoError(t, err, "create provider failed")
	defer testutil.DeleteProvider(providerName)

	// 创建日志：after 侧断言。
	createEntry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "provider",
		"action":        "create",
		"resource_id":   providerName,
	}, 0)
	require.NoError(t, err, "provider create operation log not found")

	after, ok := createEntry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "change_summary.after should be an object")
	assertProviderKeyMasked(t, after, "key-primary", "sk-a****aaaa")
	assertProviderKeyMasked(t, after, "key-secondary", "sk-b****bbbb")

	serialized := changeSummaryBytes(t, createEntry)
	assert.NotContains(t, serialized, "sk-aaaaaaaaaaaa", "audit log must not contain raw provider key")
	assert.NotContains(t, serialized, "sk-bbbbbbbbbbbb", "audit log must not contain raw provider key")

	// 更新 keys（全量替换），触发 update 日志：before（库快照）/after（请求参数）双侧断言。
	rawNewKey := "sk-maskprobe-abcdefgh"
	resp, err := testutil.GetClient().Patch("/open-api/v1/providers/"+providerName, map[string]interface{}{
		"keys": []interface{}{
			map[string]interface{}{"name": "key-primary", "key": rawNewKey},
			map[string]interface{}{"name": "key-secondary", "key": "sk-bbbbbbbbbbbb"},
		},
		"instance_pool":   []interface{}{map[string]interface{}{"addr": "10.0.0.1", "weight": 100, "port": 8080}},
		"model_protocols": []string{"openai"},
	})
	require.NoError(t, err, "update provider request failed")
	testutil.AssertSuccess(t, resp)

	updateEntry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "provider",
		"action":        "update",
		"resource_id":   providerName,
	}, 0)
	require.NoError(t, err, "provider update operation log not found")

	before, ok := updateEntry.ChangeSummary["before"].(map[string]interface{})
	require.True(t, ok, "change_summary.before should be an object")
	assertProviderKeyMasked(t, before, "key-primary", "sk-a****aaaa")

	after, ok = updateEntry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "change_summary.after should be an object")
	assertProviderKeyMasked(t, after, "key-primary", "sk-m****efgh")

	serialized = changeSummaryBytes(t, updateEntry)
	assert.NotContains(t, serialized, "sk-aaaaaaaaaaaa", "before side must not contain raw provider key")
	assert.NotContains(t, serialized, rawNewKey, "after side must not contain raw provider key")
}

// assertProviderKeyMasked 在 keys 数组中按 name 找到元素，断言 key 为预期的部分掩码值。
func assertProviderKeyMasked(t *testing.T, side map[string]interface{}, name, wantMasked string) {
	t.Helper()
	keys, ok := side["keys"].([]interface{})
	require.True(t, ok, "keys should be an array")
	for _, item := range keys {
		element, ok := item.(map[string]interface{})
		require.True(t, ok, "keys element should be an object")
		if element["name"] != name {
			continue
		}
		assert.Equal(t, wantMasked, element["key"], "key of %s should be partially masked", name)
		return
	}
	t.Errorf("keys element %s not found", name)
}

// TestOperationLog_UserPasswordMasked 验证用户创建审计日志的 password 路径
// 保持全掩码（issue #162 回归保护：修复不得影响已干净的密码路径）。
func TestOperationLog_UserPasswordMasked(t *testing.T) {
	userName := testutil.UniqueUserName()
	password := "MaskProbe@123"
	require.NoError(t, testutil.CreateUser(userName, password), "create user failed")
	defer testutil.DeleteUser(userName)

	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "user",
		"action":        "create",
		"resource_name": userName,
	}, 0)
	require.NoError(t, err, "user create operation log not found")

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "change_summary.after should be an object")
	assert.Equal(t, "******", after["password"], "password must be fully masked in audit log")

	serialized := changeSummaryBytes(t, entry)
	assert.NotContains(t, serialized, password, "audit log must not contain raw password")
}
