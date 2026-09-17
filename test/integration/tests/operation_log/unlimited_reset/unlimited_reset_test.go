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

// Package unlimited_reset_test 验证 unlimited 配额计划的 reset 失败语义（issue #183，
// SC2101-TC047 断言 3）：受控 422 / Param Illegal（不得兜底 500），失败审计
// quota_plan/reset status=2 + 非空 error_msg + owner 归属，且配额不变。
package unlimited_reset_test

import (
	"os"
	"testing"
	"time"

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

// fetchRemaining 读取资源详情中 quota_plan.balance.remaining。
func fetchRemaining(t *testing.T, path string) float64 {
	t.Helper()
	resp, err := testutil.GetClient().Get(path, nil)
	require.NoError(t, err, "get %s failed", path)
	testutil.AssertSuccess(t, resp)

	var data struct {
		QuotaPlan struct {
			Balance struct {
				Remaining float64 `json:"remaining"`
			} `json:"balance"`
		} `json:"quota_plan"`
	}
	require.NoError(t, testutil.UnmarshalData(resp, &data))
	return data.QuotaPlan.Balance.Remaining
}

// assertUnlimitedResetFailure 断言 unlimited reset 的完整失败契约：
// 422 + Param Illegal、配额不变、失败审计 status=2 且归属 owner。
func assertUnlimitedResetFailure(t *testing.T, resetPath, detailPath, ownerID string) {
	t.Helper()
	client := testutil.GetClient()

	remainingBefore := fetchRemaining(t, detailPath)

	resp, err := client.Post(resetPath, map[string]interface{}{})
	require.NoError(t, err, "reset %s failed", resetPath)
	testutil.AssertErrCode(t, resp, 422)
	assert.Contains(t, resp.ErrMsg, "Param Illegal:", "ErrMsg should be stable Param Illegal, got %s", resp.ErrMsg)
	assert.Contains(t, resp.ErrMsg, "cannot reset balance for unlimited quota")

	assert.Equal(t, remainingBefore, fetchRemaining(t, detailPath),
		"failed reset must not change balance (SC2101-TC047 assertion 3)")

	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type":      "quota_plan",
		"action":             "reset",
		"resource_parent_id": ownerID,
		"status":             "2",
	}, 15*time.Second)
	require.NoError(t, err, "failed quota_plan/reset audit with parent=%s not found", ownerID)
	assert.Equal(t, string("reset"), entry.Action)
	assert.Equal(t, "quota_plan", entry.ResourceType)
	assert.Equal(t, ownerID, entry.ResourceParentID, "resource_parent_id should be the owning resource ID")
	assert.Equal(t, float64(2), entry.Status, "unlimited reset audit should be failed")
	assert.NotEmpty(t, entry.ResourceID, "audit should carry the quota plan ID")
	assert.Contains(t, entry.ErrorMsg, "cannot reset balance for unlimited quota",
		"failed audit should carry a non-empty stable error reason")
}

// TestOperationLog_UnlimitedResetFailure 覆盖 issue #183：
// OL-UR-001 api-key 路径；OL-UR-002 entity 路径（同根因同修复）。
func TestOperationLog_UnlimitedResetFailure(t *testing.T) {
	client := testutil.GetClient()

	// OL-UR-001：unlimited API-Key reset → 422 Param Illegal + status=2 审计 + 配额不变
	t.Run("OL-UR-001 api-key unlimited reset 返回422并记录失败审计", func(t *testing.T) {
		resp, err := client.Post("/open-api/v1/api-keys", map[string]interface{}{
			"description": testutil.UniqueAPIKeyDesc(),
			"quota_plan":  map[string]interface{}{"unlimited": true},
		})
		require.NoError(t, err, "create unlimited api key failed")
		testutil.AssertSuccess(t, resp)
		idRaw, err := testutil.GetDataField(resp, "id")
		require.NoError(t, err)
		apiKeyID := idRaw.(string)
		defer testutil.DeleteAPIKey(apiKeyID)

		assertUnlimitedResetFailure(t,
			"/open-api/v1/api-keys/"+apiKeyID+"/quota-plan/reset",
			"/open-api/v1/api-keys/"+apiKeyID,
			apiKeyID)
	})

	// OL-UR-002：unlimited Entity reset → 422 Param Illegal + status=2 审计 + 配额不变
	t.Run("OL-UR-002 entity unlimited reset 返回422并记录失败审计", func(t *testing.T) {
		typeName := testutil.UniqueEntityTypeName()
		_, err := testutil.CreateEntityType(typeName, 1)
		require.NoError(t, err, "create entity type failed")
		defer testutil.DeleteEntityType(typeName)

		entityName := testutil.UniqueEntityName()
		resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
			"name":       entityName,
			"type":       typeName,
			"quota_plan": map[string]interface{}{"unlimited": true},
		})
		require.NoError(t, err, "create unlimited entity failed")
		testutil.AssertSuccess(t, resp)
		idRaw, err := testutil.GetDataField(resp, "id")
		require.NoError(t, err)
		entityID := idRaw.(string)
		defer testutil.DeleteEntity(entityID)

		assertUnlimitedResetFailure(t,
			"/open-api/v1/entities/"+entityID+"/quota-plan/reset",
			"/open-api/v1/entities/"+entityID,
			entityID)
	})
}
