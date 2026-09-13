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

// Package nested_audit_test 验证嵌套配额计划/限流策略写操作产生的操作日志
// 归属正确（resource_parent_id = 所属 Entity/API Key 业务 ID），
// 对应 issue #161 与 SC2101-TC047 的 API 层断言。
package nested_audit_test

import (
	"fmt"
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

func quotaPlanBody(quota float64) map[string]interface{} {
	return map[string]interface{}{
		"unlimited":    false,
		"quota":        quota,
		"unit":         "total_token",
		"reset_period": "monthly",
	}
}

// waitNestedAudit 轮询等待一条嵌套资源审计日志，并断言归属与成功状态。
func waitNestedAudit(t *testing.T, resourceType, action, parentID string) *testutil.OperationLogEntry {
	t.Helper()
	entry, err := testutil.WaitForOperationLog(map[string]string{
		"resource_type":      resourceType,
		"action":             action,
		"resource_parent_id": parentID,
	}, 15*time.Second)
	require.NoError(t, err, "nested %s/%s audit with parent=%s not found", resourceType, action, parentID)
	assert.Equal(t, resourceType, entry.ResourceType)
	assert.Equal(t, action, entry.Action)
	assert.Equal(t, parentID, entry.ResourceParentID, "resource_parent_id should be the owning resource ID")
	assert.Equal(t, float64(1), entry.Status, "nested audit should be success")
	assert.NotEmpty(t, entry.ResourceID, "nested audit should carry the quota plan ID")
	return entry
}

// TestOperationLog_NestedQuotaPlanAudit 覆盖 issue #161 的两层缺陷：
// 缺陷① 嵌套 CRUD 无审计；缺陷② reset 日志 resource_parent_id 为空。
// 场景：Entity 嵌套 create/update/reset/delete + API-Key 嵌套 create/reset/delete。
func TestOperationLog_NestedQuotaPlanAudit(t *testing.T) {
	client := testutil.GetClient()

	// 1. 创建 entity-type
	typeName := testutil.UniqueEntityTypeName()
	_, err := testutil.CreateEntityType(typeName, 1)
	require.NoError(t, err, "create entity type failed")
	defer testutil.DeleteEntityType(typeName)

	// 2. 创建带嵌套 quota_plan 的 entity → 期望 quota_plan/create，parent=entityID
	entityName := testutil.UniqueEntityName()
	resp, err := client.Post("/open-api/v1/entities", map[string]interface{}{
		"name":       entityName,
		"type":       typeName,
		"quota_plan": quotaPlanBody(1000000),
	})
	require.NoError(t, err, "create entity with nested quota plan failed")
	testutil.AssertSuccess(t, resp)
	idRaw, err := testutil.GetDataField(resp, "id")
	require.NoError(t, err)
	entityID := idRaw.(string)
	defer testutil.DeleteEntity(entityID)

	t.Run("OL-N-001 entity嵌套创建配额计划产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "create", entityID)
	})

	// 3. PATCH 更新嵌套 quota_plan → 期望 quota_plan/update，parent=entityID
	resp, err = client.Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"quota_plan": quotaPlanBody(2000000),
	})
	require.NoError(t, err, "patch entity quota plan failed")
	testutil.AssertSuccess(t, resp)

	t.Run("OL-N-002 entity嵌套更新配额计划产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "update", entityID)
	})

	// 4. 手动 reset entity 配额 → 期望 quota_plan/reset，parent=entityID（缺陷②）
	resp, err = client.Post("/open-api/v1/entities/"+entityID+"/quota-plan/reset", map[string]interface{}{})
	require.NoError(t, err, "reset entity quota failed")
	testutil.AssertSuccess(t, resp)

	t.Run("OL-N-003 entity手动重置配额产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "reset", entityID)
	})

	// 5. PATCH 嵌套 rate_limit_policy → 期望 rate_limit_policy/create，parent=entityID（同族把守）
	resp, err = client.Patch("/open-api/v1/entities/"+entityID, map[string]interface{}{
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []interface{}{
					map[string]interface{}{"name": "t1", "model": "*", "window_minutes": 1, "max_tokens": 100, "step_minutes": 1},
				},
			},
		},
	})
	require.NoError(t, err, "patch entity rate limit policy failed")
	testutil.AssertSuccess(t, resp)

	t.Run("OL-N-004 entity嵌套创建限流策略产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "rate_limit_policy", "create", entityID)
	})

	// 6. 创建带嵌套 quota_plan 的 api-key → 期望 quota_plan/create，parent=apiKeyID
	resp, err = client.Post("/open-api/v1/api-keys", map[string]interface{}{
		"description": fmt.Sprintf("nested-audit-key-%d", time.Now().UnixNano()),
		"entity_id":   entityID,
		"quota_plan":  quotaPlanBody(3000000),
	})
	require.NoError(t, err, "create api key with nested quota plan failed")
	testutil.AssertSuccess(t, resp)
	akRaw, err := testutil.GetDataField(resp, "id")
	require.NoError(t, err)
	apiKeyID := akRaw.(string)
	defer testutil.DeleteAPIKey(apiKeyID)

	t.Run("OL-N-005 api-key嵌套创建配额计划产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "create", apiKeyID)
	})

	// 7. 手动 reset api-key 配额 → 期望 quota_plan/reset，parent=apiKeyID（缺陷②）
	resp, err = client.Post("/open-api/v1/api-keys/"+apiKeyID+"/quota-plan/reset", map[string]interface{}{})
	require.NoError(t, err, "reset api key quota failed")
	testutil.AssertSuccess(t, resp)

	t.Run("OL-N-006 api-key手动重置配额产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "reset", apiKeyID)
	})

	// 8. 删除 api-key → 期望 quota_plan/delete，parent=apiKeyID
	require.NoError(t, testutil.DeleteAPIKey(apiKeyID), "delete api key failed")

	t.Run("OL-N-007 api-key删除级联删除配额计划产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "delete", apiKeyID)
	})

	// 9. 删除 entity → 期望 quota_plan/delete 与 rate_limit_policy/delete，parent=entityID
	require.NoError(t, testutil.DeleteEntity(entityID), "delete entity failed")

	t.Run("OL-N-008 entity删除级联删除配额计划产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "quota_plan", "delete", entityID)
	})

	t.Run("OL-N-009 entity删除级联删除限流策略产生归属日志", func(t *testing.T) {
		waitNestedAudit(t, "rate_limit_policy", "delete", entityID)
	})
}
