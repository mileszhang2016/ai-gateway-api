package entity_type_test

import (
	"os"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
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

func TestEntityType_Create(t *testing.T) {
	typeName := testutil.UniqueEntityTypeName()
	typeName2 := testutil.UniqueEntityTypeName()

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
		pre      func()
	}{
		{
			name: "ET-1-001 创建 Entity-Type（完整参数）",
			body: map[string]interface{}{
				"type_name":   typeName,
				"description": "一级部门",
				"level":       1,
			},
			wantCode: 200,
		},
		{
			name: "ET-1-002 创建 Entity-Type（仅必填）",
			body: map[string]interface{}{
				"type_name": typeName2,
				"level":     2,
			},
			wantCode: 200,
		},
		{
			name:     "ET-1-003 缺少 type_name",
			body:     map[string]interface{}{"level": 1},
			wantCode: 422,
		},
		{
			name:     "ET-1-004 缺少 level",
			body:     map[string]interface{}{"type_name": testutil.UniqueEntityTypeName()},
			wantCode: 422,
		},
		{
			name: "ET-1-005 重复创建同名 Entity-Type",
			body: map[string]interface{}{
				"type_name": typeName,
				"level":     1,
			},
			wantCode: 556,
		},
		{
			name: "ET-1-006 level 超出范围",
			body: map[string]interface{}{
				"type_name": testutil.UniqueEntityTypeName(),
				"level":     6,
			},
			wantCode: 422,
		},
		{
			name: "ET-1-007 type_name 包含大写字母",
			body: map[string]interface{}{
				"type_name": "BadType",
				"level":     1,
			},
			wantCode: 422,
		},
		{
			name: "ET-1-008 type_name 以 - 开头",
			body: map[string]interface{}{
				"type_name": "-badtype",
				"level":     1,
			},
			wantCode: 422,
		},
		{
			name: "ET-1-009 type_name 包含空白",
			body: map[string]interface{}{
				"type_name": "bad type",
				"level":     1,
			},
			wantCode: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pre != nil {
				tt.pre()
			}
			resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", tt.body)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.ErrNum != tt.wantCode {
				t.Errorf("expected ErrNum=%d, got ErrNum=%d, ErrMsg=%s", tt.wantCode, resp.ErrNum, resp.ErrMsg)
			}
		})
	}

	t.Cleanup(func() {
		testutil.DeleteEntityType(typeName)
		testutil.DeleteEntityType(typeName2)
	})
}
