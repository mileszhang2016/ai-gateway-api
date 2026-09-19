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
	"strings"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maskAPIKeyToken 与 model/ioperlog.MaskAPIKeyToken 同形（首 4 + **** + 尾 4，
// 短值全掩码），避免跨 module 引用主项目内部包。
func maskAPIKeyToken(token string) string {
	if len(token) <= 8 {
		return "******"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// TestOperationLog_DuplicateKeyCreateMasksErrorMsg 验证重复创建相同 Key 值的
// API-Key 时，失败审计日志的 error_msg 不回显裸 Key（issue #185 /
// SC2101-TC046 回归锚点）：error_msg 只含部分掩码形态，整条记录序列化后
// 不含裸 Key；同时校验 422 响应 msg 同样为掩码形态。
func TestOperationLog_DuplicateKeyCreateMasksErrorMsg(t *testing.T) {
	desc := testutil.UniqueAPIKeyDesc()
	keyID, rawKey, err := testutil.CreateAPIKeyWithKey(desc, "")
	require.NoError(t, err, "create api-key failed")
	require.NotEmpty(t, rawKey)
	defer testutil.DeleteAPIKey(keyID)

	// 触发：以相同 key 值再次创建，网关返回 422。
	dupDesc := desc + "-dup"
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": dupDesc,
		"key":         rawKey,
	})
	require.NoError(t, err, "duplicate create request failed")
	testutil.AssertErrCode(t, resp, 422)
	assert.NotContains(t, resp.ErrMsg, rawKey, "422 response msg must not echo the raw key")
	assert.Contains(t, resp.ErrMsg, maskAPIKeyToken(rawKey), "422 response msg should carry the masked key")

	// 回读失败审计记录，锚定本次重复创建（resource_name = 重复请求的 description）。
	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type": "api_key",
		"action":        "create",
		"resource_name": dupDesc,
		"status":        "2",
	}, 0)
	require.NoError(t, err, "duplicate-create failed audit log not found")

	require.NotEmpty(t, entry.ErrorMsg, "failed audit entry must carry error_msg")
	assert.NotContains(t, entry.ErrorMsg, rawKey, "error_msg must not contain the raw API key")
	assert.Contains(t, entry.ErrorMsg, maskAPIKeyToken(rawKey), "error_msg should carry the masked key")

	// 整条记录（error_msg + 已脱敏 change_summary）序列化后不含裸 Key，
	// 对齐 SC2101-TC046 的 sc2101AuditAssertSecretAbsent 断言。
	serialized, err := json.Marshal(entry)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), rawKey, "audit entry must not contain the raw API key")
}

// waitNestedFailureAuditByMaskedKey 轮询 resource_type+action=create+status=2 的
// 失败审计，以 error_msg 中的掩码 Key 作为唯一锚点。失败创建的资源 ID 由服务端
// 内部生成、不在 422 响应中返回，其嵌套条目的 resource_parent_id 无法从外部预知，
// 而每条失败请求的掩码 Key 唯一，可作确定性锚点。
func waitNestedFailureAuditByMaskedKey(t *testing.T, resourceType, maskedKey string) *testutil.OperationLogEntry {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		result, _, err := testutil.QueryOperationLogs(map[string]string{
			"resource_type": resourceType,
			"action":        "create",
			"status":        "2",
		})
		require.NoError(t, err)
		for i := range result.List {
			if strings.Contains(result.List[i].ErrorMsg, maskedKey) {
				return &result.List[i]
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("%s failed audit log with masked key not found", resourceType)
	return nil
}

// TestOperationLog_DuplicateKeyCreateNestedAuditMasksErrorMsg 验证泄漏扇出收口
//（issue #185）：重复创建请求携带 quota_plan / rate_limit_policy 时，同一错误
// 经嵌套审计写入 quota_plan 与 rate_limit_policy 失败记录，其 error_msg 同样
// 不得含裸 Key。
func TestOperationLog_DuplicateKeyCreateNestedAuditMasksErrorMsg(t *testing.T) {
	desc := testutil.UniqueAPIKeyDesc()
	keyID, rawKey, err := testutil.CreateAPIKeyWithKey(desc, "")
	require.NoError(t, err, "create api-key failed")
	defer testutil.DeleteAPIKey(keyID)

	dupDesc := desc + "-dup-nested"
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": dupDesc,
		"key":         rawKey,
		"quota_plan": map[string]interface{}{
			"unlimited":    false,
			"quota":        1000000,
			"unit":         "total_token",
			"reset_period": "monthly",
		},
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{
						"name": "1m", "model": "*", "window_minutes": 1, "max_tokens": 10000, "step_minutes": 1,
					},
				},
			},
		},
	})
	require.NoError(t, err, "duplicate create request failed")
	testutil.AssertErrCode(t, resp, 422)

	// 嵌套失败审计：以 error_msg 中的掩码 Key 为锚点（失败请求的资源 ID 不在响应中返回，
	// 其嵌套条目的 resource_parent_id 无法从外部预知）。
	maskedKey := maskAPIKeyToken(rawKey)
	for _, resourceType := range []string{"quota_plan", "rate_limit_policy"} {
		entry := waitNestedFailureAuditByMaskedKey(t, resourceType, maskedKey)

		require.NotEmpty(t, entry.ErrorMsg, "%s failed audit entry must carry error_msg", resourceType)
		assert.NotContains(t, entry.ErrorMsg, rawKey, "%s error_msg must not contain the raw API key", resourceType)
		assert.Contains(t, entry.ErrorMsg, maskedKey, "%s error_msg should carry the masked key", resourceType)

		serialized, err := json.Marshal(entry)
		require.NoError(t, err)
		assert.NotContains(t, string(serialized), rawKey, "%s audit entry must not contain the raw API key", resourceType)
	}
}
