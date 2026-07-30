package entity_test

import (
	"os"
	"testing"

	"github.com/yf-networks/ai-gateway-api/integration/testutil"
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
	}{
		{
			name: "E-1-001 创建 Entity（仅必填）",
			body: map[string]interface{}{"name": entRoot, "type": typeName},
			wantCode: 200,
		},
		{
			name: "E-1-002 创建 Entity（含 quota_plan）",
			body: map[string]interface{}{
				"name": entQuota,
				"type": typeName,
				"quota_plan": map[string]interface{}{
					"unlimited":   false,
					"quota":       1000000,
					"unit":        "total_token",
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
		})
	}

	t.Cleanup(func() {
		testutil.DeleteEntity(parentID)
		testutil.DeleteEntityType(typeName)
		testutil.DeleteEntityType(typeName2)
	})
}
