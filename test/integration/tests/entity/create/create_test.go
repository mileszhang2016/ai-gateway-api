package entity_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
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

func TestEntity_Create(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	typeName2 := testutil.UniqueEntityTypeName()
	if _, err := testutil.CreateEntityType(typeName, 1); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := testutil.CreateEntityType(typeName2, 2); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	parentName := testutil.UniqueEntityName()
	parentID, err := testutil.CreateEntity(parentName, typeName, "")
	if err != nil {
		t.Fatalf("setup parent failed: %v", err)
	}

	entRoot := testutil.UniqueEntityName()
	entQuota := testutil.UniqueEntityName()
	entDup := testutil.UniqueEntityName()
	entTeam := testutil.UniqueEntityName()
	entBadParent := testutil.UniqueEntityName()

	// 预先创建重复名称
	if _, err := testutil.CreateEntity(entDup, typeName, ""); err != nil {
		t.Fatalf("setup dup failed: %v", err)
	}

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
		check    func(t *testing.T, resp *testutil.APIResponse)
	}{
		{
			name:     "E-1-001 创建 Entity（仅必填）",
			body:     map[string]interface{}{"name": entRoot, "type": typeName},
			wantCode: 200,
		},
		{
			name: "E-1-002 创建 Entity（含 quota_plan）",
			body: map[string]interface{}{
				"name": entQuota,
				"type": typeName,
				"quota_plan": map[string]interface{}{
					"unlimited":    false,
					"quota":        1000000,
					"unit":         "total_token",
					"reset_period": "monthly",
				},
			},
			wantCode: 200,
		},
		{
			name:     "E-1-003 缺少 name",
			body:     map[string]interface{}{"type": typeName},
			wantCode: 422,
		},
		{
			name:     "E-1-004 缺少 type",
			body:     map[string]interface{}{"name": testutil.UniqueEntityName()},
			wantCode: 422,
		},
		{
			name:     "E-1-005 type 不存在",
			body:     map[string]interface{}{"name": testutil.UniqueEntityName(), "type": "not_exist"},
			wantCode: 422,
		},
		{
			name:     "E-1-006 重复 name",
			body:     map[string]interface{}{"name": entDup, "type": typeName},
			wantCode: 556,
		},
		{
			name: "E-1-007 创建层级 Entity（合法 parent）",
			body: map[string]interface{}{
				"name":      entTeam,
				"type":      typeName2,
				"parent_id": parentID,
			},
			wantCode: 200,
		},
		{
			name: "E-1-008 创建层级 Entity（非法 parent level）",
			body: map[string]interface{}{
				"name":      entBadParent,
				"type":      typeName,
				"parent_id": parentID,
			},
			wantCode: 422,
		},
		{
			name:     "E-1-009 type 格式非法（含大写）",
			body:     map[string]interface{}{"name": testutil.UniqueEntityName(), "type": "BadType"},
			wantCode: 422,
		},
		{
			name:     "E-1-010 Entity name 包含首尾空白",
			body:     map[string]interface{}{"name": " badname ", "type": typeName},
			wantCode: 422,
		},
		{
			name: "E-1-011 创建 Entity 并指定 RMB 配额",
			body: map[string]interface{}{
				"name": testutil.UniqueEntityName(),
				"type": typeName,
				"quota_plan": map[string]interface{}{
					"unlimited":    false,
					"quota":        5555.5555,
					"unit":         "RMB",
					"reset_period": "monthly",
				},
			},
			wantCode: 200,
			check: func(t *testing.T, resp *testutil.APIResponse) {
				var data map[string]interface{}
				if err := json.Unmarshal(resp.Data, &data); err != nil {
					t.Fatalf("unmarshal data: %v", err)
				}
				qp, ok := data["quota_plan"].(map[string]interface{})
				if !assert.True(t, ok, "quota_plan should be an object") {
					return
				}
				assert.Equal(t, "RMB", qp["unit"])
				assert.InDelta(t, float64(5555.5555), qp["quota"], 0.00001)

				id, _ := data["id"].(string)
				qpResp, err := testutil.GetClient().Get("/open-api/v1/entities/" + id + "/quota-plan")
				if err != nil {
					t.Fatalf("query quota-plan failed: %v", err)
				}
				testutil.AssertSuccess(t, qpResp)
				var qpData map[string]interface{}
				if err := json.Unmarshal(qpResp.Data, &qpData); err != nil {
					t.Fatalf("unmarshal quota-plan data: %v", err)
				}
				balance, ok := qpData["balance"].(map[string]interface{})
				if !assert.True(t, ok, "balance should be an object") {
					return
				}
				assert.InDelta(t, float64(5555.5555), balance["remaining"], 0.00001)
				assert.InDelta(t, float64(0), balance["used"], 0.00001)
			},
		},
		{
			name: "E-1-012 创建 Entity 时 RMB 配额超过 9000 万元上限",
			body: map[string]interface{}{
				"name": testutil.UniqueEntityName(),
				"type": typeName,
				"quota_plan": map[string]interface{}{
					"unlimited":    false,
					"quota":        90000000.01,
					"unit":         "RMB",
					"reset_period": "monthly",
				},
			},
			wantCode: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testutil.GetClient().Post("/open-api/v1/entities", tt.body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
			if tt.check != nil && resp.ErrNum == 200 {
				tt.check(t, resp)
			}
		})
	}

	t.Cleanup(func() {
		testutil.DeleteEntity(parentID)
		testutil.DeleteEntityType(typeName)
		testutil.DeleteEntityType(typeName2)
	})
}
