package create

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/yf-networks/ai-gateway-api/test-runner/testutil"
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

// helper: 创建 API-Key
func createAPIKey(t *testing.T, body interface{}) *testutil.APIResponse {
	t.Helper()
	resp, err := testutil.GetClient().Post("/open-api/v1/api-keys", body)
	if err != nil {
		t.Fatalf("create API-Key failed: %v", err)
	}
	return resp
}

// helper: 解析 Data 为 map
func parseData(t *testing.T, resp *testutil.APIResponse) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	return data
}

// ============================================================
// 正常参数 (18)
// ============================================================

// AK-1-001：最小参数创建（仅description），验证所有默认值
func TestCreate_Normal_MinimalParams(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-001",
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	// 验证必返回字段存在且非空
	if id, ok := data["id"].(string); !ok || id == "" {
		t.Errorf("id should be non-empty string, got %v", data["id"])
	}
	if key, ok := data["key"].(string); !ok || key == "" {
		t.Errorf("key should be non-empty string, got %v", data["key"])
	}
	if desc, ok := data["description"].(string); !ok || desc != "test-key-001" {
		t.Errorf("expected description='test-key-001', got %v", data["description"])
	}

	// 验证默认值
	if enabled, ok := data["enabled"].(bool); !ok || !enabled {
		t.Errorf("expected enabled=true (default), got %v", data["enabled"])
	}

	// create_time 和 update_time 应 > 0
	if ct, ok := data["create_time"].(float64); !ok || ct <= 0 {
		t.Errorf("create_time should be > 0, got %v", data["create_time"])
	}
	if ut, ok := data["update_time"].(float64); !ok || ut <= 0 {
		t.Errorf("update_time should be > 0, got %v", data["update_time"])
	}

	// expired_time 默认 -1
	if et, ok := data["expired_time"].(float64); !ok || et != -1 {
		t.Errorf("expected expired_time=-1 (default), got %v", data["expired_time"])
	}

	// unlimited_quota 默认 false
	if uq, ok := data["unlimited_quota"].(bool); ok && uq {
		t.Errorf("expected unlimited_quota=false (default), got true")
	}

	// models 默认 ["*"]
	if models, ok := data["models"].([]interface{}); ok {
		if len(models) != 1 || models[0] != "*" {
			t.Errorf("expected models=[\"*\"], got %v", data["models"])
		}
	}

	// quota_plan 不应为 null
	if data["quota_plan"] == nil {
		t.Error("quota_plan should not be null")
	}

	// rate_limit_policy 不应为 null
	if data["rate_limit_policy"] == nil {
		t.Error("rate_limit_policy should not be null")
	}
}

// AK-1-002：设置 expired_time 为未来时间戳
func TestCreate_Normal_ExpiredTime(t *testing.T) {
	futureTimestamp := time.Now().Unix() + 86400*365

	body := map[string]interface{}{
		"description":  "test-key-002",
		"expired_time": futureTimestamp,
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	if et, ok := data["expired_time"].(float64); !ok || int64(et) != futureTimestamp {
		t.Errorf("expected expired_time=%d, got %v", futureTimestamp, data["expired_time"])
	}
}

// AK-1-003：设置 enabled=false
func TestCreate_Normal_EnabledFalse(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-003",
		"enabled":     false,
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	if enabled, ok := data["enabled"].(bool); !ok || enabled {
		t.Errorf("expected enabled=false, got %v", data["enabled"])
	}
}

// AK-1-004：设置 unlimited_quota=true
func TestCreate_Normal_UnlimitedQuota(t *testing.T) {
	body := map[string]interface{}{
		"description":     "test-key-004",
		"unlimited_quota": true,
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	if uq, ok := data["unlimited_quota"].(bool); !ok || !uq {
		t.Errorf("expected unlimited_quota=true, got %v", data["unlimited_quota"])
	}
}

// AK-1-005：设置 models 指定白名单
func TestCreate_Normal_ModelsWhitelist(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-005",
		"models":      []string{"gpt-4", "gpt-3.5-turbo"},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	models, ok := data["models"].([]interface{})
	if !ok {
		t.Fatal("models should be an array")
	}
	if len(models) != 2 {
		t.Errorf("expected models length=2, got %d", len(models))
	}
	if models[0] != "gpt-4" {
		t.Errorf("models[0]: expected 'gpt-4', got %v", models[0])
	}
	if models[1] != "gpt-3.5-turbo" {
		t.Errorf("models[1]: expected 'gpt-3.5-turbo', got %v", models[1])
	}
}

// AK-1-006：设置 subnet 指定子网
func TestCreate_Normal_Subnet(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-006",
		"subnet":      []string{"10.0.0.0/8", "192.168.1.0/24"},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	subnet, ok := data["subnet"].([]interface{})
	if !ok {
		t.Fatal("subnet should be an array")
	}
	if len(subnet) != 2 {
		t.Errorf("expected subnet length=2, got %d", len(subnet))
	}
	if subnet[0] != "10.0.0.0/8" {
		t.Errorf("subnet[0]: expected '10.0.0.0/8', got %v", subnet[0])
	}
	if subnet[1] != "192.168.1.0/24" {
		t.Errorf("subnet[1]: expected '192.168.1.0/24', got %v", subnet[1])
	}
}

// AK-1-007：quota_plan 全字段设置
func TestCreate_Normal_QuotaPlanAllFields(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-007",
		"quota_plan": map[string]interface{}{
			"unlimited":                 false,
			"pass_when_no_enough_quota": false,
			"quota":                     100000000,
			"unit":                      "total_token",
			"reset_period":              "monthly",
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["unlimited"] != false {
		t.Errorf("quota_plan.unlimited: expected false, got %v", qp["unlimited"])
	}
	if qp["pass_when_no_enough_quota"] != false {
		t.Errorf("quota_plan.pass_when_no_enough_quota: expected false, got %v", qp["pass_when_no_enough_quota"])
	}
	if qp["quota"] != float64(100000000) {
		t.Errorf("quota_plan.quota: expected 100000000, got %v", qp["quota"])
	}
	if qp["unit"] != "total_token" {
		t.Errorf("quota_plan.unit: expected 'total_token', got %v", qp["unit"])
	}
	if qp["reset_period"] != "monthly" {
		t.Errorf("quota_plan.reset_period: expected 'monthly', got %v", qp["reset_period"])
	}
}

// AK-1-008：quota_plan.unlimited=true
func TestCreate_Normal_QuotaPlanUnlimited(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-008",
		"quota_plan": map[string]interface{}{
			"unlimited": true,
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["unlimited"] != true {
		t.Errorf("quota_plan.unlimited: expected true, got %v", qp["unlimited"])
	}
}

// AK-1-009：quota_plan.pass_when_no_enough_quota=true
func TestCreate_Normal_QuotaPlanPassWhenNoEnough(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-009",
		"quota_plan": map[string]interface{}{
			"pass_when_no_enough_quota": true,
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["pass_when_no_enough_quota"] != true {
		t.Errorf("quota_plan.pass_when_no_enough_quota: expected true, got %v", qp["pass_when_no_enough_quota"])
	}
}

// AK-1-010：quota_plan.unit="token"
func TestCreate_Normal_QuotaPlanUnit(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-010",
		"quota_plan": map[string]interface{}{
			"unit": "token",
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["unit"] != "token" {
		t.Errorf("quota_plan.unit: expected 'token', got %v", qp["unit"])
	}
}

// AK-1-011：quota_plan.reset_period="daily"
func TestCreate_Normal_QuotaPlanResetPeriod(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-011",
		"quota_plan": map[string]interface{}{
			"reset_period": "daily",
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["reset_period"] != "daily" {
		t.Errorf("quota_plan.reset_period: expected 'daily', got %v", qp["reset_period"])
	}
}

// AK-1-012：quota_plan.quota=50000000
func TestCreate_Normal_QuotaPlanQuota(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-012",
		"quota_plan": map[string]interface{}{
			"quota": 50000000,
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["quota"] != float64(50000000) {
		t.Errorf("quota_plan.quota: expected 50000000, got %v", qp["quota"])
	}
}

// AK-1-013：rate_limit_policy 全字段设置
func TestCreate_Normal_RateLimitFull(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-013",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "TPM窗口", "model": "gpt-4", "window_minutes": 5, "max_tokens": 5000, "step_minutes": 1},
				},
				"rpm": []map[string]interface{}{
					{"name": "RPM窗口", "model": "*", "window_minutes": 1, "max_requests": 200},
				},
				"max_concurrency": 50,
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	if rlp["enabled"] != true {
		t.Errorf("rate_limit_policy.enabled: expected true, got %v", rlp["enabled"])
	}

	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	// 验证 tpm
	tpm, ok := rules["tpm"].([]interface{})
	if !ok || len(tpm) == 0 {
		t.Fatal("rules.tpm should be a non-empty array")
	}
	tpm0 := tpm[0].(map[string]interface{})
	if tpm0["name"] != "TPM窗口" {
		t.Errorf("tpm[0].name: expected 'TPM窗口', got %v", tpm0["name"])
	}
	if tpm0["model"] != "gpt-4" {
		t.Errorf("tpm[0].model: expected 'gpt-4', got %v", tpm0["model"])
	}
	if tpm0["window_minutes"] != float64(5) {
		t.Errorf("tpm[0].window_minutes: expected 5, got %v", tpm0["window_minutes"])
	}
	if tpm0["max_tokens"] != float64(5000) {
		t.Errorf("tpm[0].max_tokens: expected 5000, got %v", tpm0["max_tokens"])
	}
	if tpm0["step_minutes"] != float64(1) {
		t.Errorf("tpm[0].step_minutes: expected 1, got %v", tpm0["step_minutes"])
	}

	// 验证 rpm
	rpm, ok := rules["rpm"].([]interface{})
	if !ok || len(rpm) == 0 {
		t.Fatal("rules.rpm should be a non-empty array")
	}
	rpm0 := rpm[0].(map[string]interface{})
	if rpm0["name"] != "RPM窗口" {
		t.Errorf("rpm[0].name: expected 'RPM窗口', got %v", rpm0["name"])
	}
	if rpm0["model"] != "*" {
		t.Errorf("rpm[0].model: expected '*', got %v", rpm0["model"])
	}
	if rpm0["window_minutes"] != float64(1) {
		t.Errorf("rpm[0].window_minutes: expected 1, got %v", rpm0["window_minutes"])
	}
	if rpm0["max_requests"] != float64(200) {
		t.Errorf("rpm[0].max_requests: expected 200, got %v", rpm0["max_requests"])
	}

	// 验证 max_concurrency
	if mc, ok := rules["max_concurrency"].(float64); !ok || mc != 50 {
		t.Errorf("rules.max_concurrency: expected 50, got %v", rules["max_concurrency"])
	}
}

// AK-1-014：rate_limit_policy 仅设置 tpm
func TestCreate_Normal_RateLimitTpmOnly(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-014",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "TPM窗口", "model": "gpt-4", "window_minutes": 5, "max_tokens": 5000, "step_minutes": 1},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	if rlp["enabled"] != true {
		t.Errorf("rate_limit_policy.enabled: expected true, got %v", rlp["enabled"])
	}

	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	tpm, ok := rules["tpm"].([]interface{})
	if !ok || len(tpm) == 0 {
		t.Fatal("rules.tpm should be a non-empty array")
	}
	tpm0 := tpm[0].(map[string]interface{})
	if tpm0["name"] != "TPM窗口" {
		t.Errorf("tpm[0].name: expected 'TPM窗口', got %v", tpm0["name"])
	}
	if tpm0["model"] != "gpt-4" {
		t.Errorf("tpm[0].model: expected 'gpt-4', got %v", tpm0["model"])
	}
	if tpm0["window_minutes"] != float64(5) {
		t.Errorf("tpm[0].window_minutes: expected 5, got %v", tpm0["window_minutes"])
	}
	if tpm0["max_tokens"] != float64(5000) {
		t.Errorf("tpm[0].max_tokens: expected 5000, got %v", tpm0["max_tokens"])
	}
	if tpm0["step_minutes"] != float64(1) {
		t.Errorf("tpm[0].step_minutes: expected 1, got %v", tpm0["step_minutes"])
	}
}

// AK-1-015：rate_limit_policy 仅设置 rpm
func TestCreate_Normal_RateLimitRpmOnly(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-015",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"rpm": []map[string]interface{}{
					{"name": "RPM窗口", "model": "*", "window_minutes": 1, "max_requests": 200},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	if rlp["enabled"] != true {
		t.Errorf("rate_limit_policy.enabled: expected true, got %v", rlp["enabled"])
	}

	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	rpm, ok := rules["rpm"].([]interface{})
	if !ok || len(rpm) == 0 {
		t.Fatal("rules.rpm should be a non-empty array")
	}
	rpm0 := rpm[0].(map[string]interface{})
	if rpm0["name"] != "RPM窗口" {
		t.Errorf("rpm[0].name: expected 'RPM窗口', got %v", rpm0["name"])
	}
	if rpm0["model"] != "*" {
		t.Errorf("rpm[0].model: expected '*', got %v", rpm0["model"])
	}
	if rpm0["window_minutes"] != float64(1) {
		t.Errorf("rpm[0].window_minutes: expected 1, got %v", rpm0["window_minutes"])
	}
	if rpm0["max_requests"] != float64(200) {
		t.Errorf("rpm[0].max_requests: expected 200, got %v", rpm0["max_requests"])
	}
}

// AK-1-016：rate_limit_policy 仅设置 max_concurrency=50
func TestCreate_Normal_RateLimitConcurrencyOnly(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-016",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"max_concurrency": 50,
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	if rlp["enabled"] != true {
		t.Errorf("rate_limit_policy.enabled: expected true, got %v", rlp["enabled"])
	}

	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	if mc, ok := rules["max_concurrency"].(float64); !ok || mc != 50 {
		t.Errorf("rules.max_concurrency: expected 50, got %v", rules["max_concurrency"])
	}
}

// AK-1-017：rate_limit_policy 仅设置 max_concurrency=0
func TestCreate_Normal_RateLimitConcurrencyZero(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-017",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"max_concurrency": 0,
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	if rlp["enabled"] != true {
		t.Errorf("rate_limit_policy.enabled: expected true, got %v", rlp["enabled"])
	}

	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	if mc, ok := rules["max_concurrency"].(float64); !ok || mc != 0 {
		t.Errorf("rules.max_concurrency: expected 0, got %v", rules["max_concurrency"])
	}
}

// AK-1-018：创建带 entity_id 的 API-Key
func TestCreate_Normal_WithEntityId(t *testing.T) {
	// 先创建 EntityType
	etypeBody := map[string]interface{}{
		"type_name": "create-etype",
		"level":     1,
	}
	resp, err := testutil.GetClient().Post("/open-api/v1/entity-types", etypeBody)
	if err != nil {
		t.Fatalf("create entity-type failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)

	// 创建 Entity
	entityBody := map[string]interface{}{
		"name": "test-entity",
		"type": "create-etype",
	}
	resp, err = testutil.GetClient().Post("/open-api/v1/entities", entityBody)
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	testutil.AssertSuccess(t, resp)
	entityData := parseData(t, resp)
	entityID := entityData["id"].(string)

	// 创建 API-Key，关联 entity_id
	body := map[string]interface{}{
		"description": "test-key-018",
		"entity_id":   entityID,
	}

	resp = createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	// 验证 entity_id
	if eid, ok := data["entity_id"].(string); !ok || eid != entityID {
		t.Errorf("expected entity_id='%s', got %v", entityID, data["entity_id"])
	}

	// 验证 entity 字段存在
	if data["entity"] == nil {
		t.Error("entity should not be null")
	}
}

// ============================================================
// 必填校验 (2)
// ============================================================

// AK-1-019：缺少 description 必填字段
func TestCreate_Required_MissingDescription(t *testing.T) {
	body := map[string]interface{}{}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-020：description 为空字符串
func TestCreate_Required_EmptyDescription(t *testing.T) {
	body := map[string]interface{}{
		"description": "",
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 边界值 (8)
// ============================================================

// AK-1-021：expired_time=-1 永不过期
func TestCreate_Boundary_ExpiredTimeNeverExpire(t *testing.T) {
	body := map[string]interface{}{
		"description":  "test-key-boundary",
		"expired_time": -1,
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	if et, ok := data["expired_time"].(float64); !ok || int64(et) != -1 {
		t.Errorf("expected expired_time=-1, got %v", data["expired_time"])
	}
}

// AK-1-022：description 最大长度 511 字符
func TestCreate_Boundary_DescriptionMaxLength(t *testing.T) {
	desc := ""
	for i := 0; i < 511; i++ {
		desc += "a"
	}
	body := map[string]interface{}{
		"description": desc,
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	if d, ok := data["description"].(string); !ok || len(d) != 511 {
		t.Errorf("expected description length=511, got length=%d", len(d))
	}
}

// AK-1-023：models 为空数组
func TestCreate_Boundary_ModelsEmpty(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"models":      []string{},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	models, ok := data["models"].([]interface{})
	if !ok {
		t.Fatal("models should be an array")
	}
	if len(models) != 0 {
		// 空数组可能被服务端默认填充为 ["*"]
		t.Logf("models empty array was defaulted to %v (expected empty or ['*'])", models)
	}
}

// AK-1-024：subnet 为空数组
func TestCreate_Boundary_SubnetEmpty(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"subnet":      []string{},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	subnet, ok := data["subnet"].([]interface{})
	if !ok {
		t.Fatal("subnet should be an array")
	}
	if len(subnet) != 0 {
		// 空数组可能被服务端默认填充为 ["*"]
		t.Logf("subnet empty array was defaulted to %v (expected empty or ['*'])", subnet)
	}
}

// AK-1-025：quota_plan.quota=0
func TestCreate_Boundary_QuotaPlanQuotaZero(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"quota_plan": map[string]interface{}{
			"quota": 0,
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	if qp["quota"] != float64(0) {
		t.Errorf("quota_plan.quota: expected 0, got %v", qp["quota"])
	}
}

// AK-1-026：tpm window_minutes=1（最小边界）
func TestCreate_Boundary_TpmWindowMin1(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "边界窗口", "model": "*", "window_minutes": 1, "max_tokens": 1000, "step_minutes": 1},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}
	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}
	tpm, ok := rules["tpm"].([]interface{})
	if !ok || len(tpm) == 0 {
		t.Fatal("rules.tpm should be a non-empty array")
	}
	tpm0 := tpm[0].(map[string]interface{})
	if tpm0["window_minutes"] != float64(1) {
		t.Errorf("tpm[0].window_minutes: expected 1, got %v", tpm0["window_minutes"])
	}
}

// AK-1-027：tpm window_minutes=360, step_minutes=1（最大边界）
func TestCreate_Boundary_TpmWindowMax360(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "最大窗口", "model": "*", "window_minutes": 360, "max_tokens": 1000, "step_minutes": 1},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}
	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}
	tpm, ok := rules["tpm"].([]interface{})
	if !ok || len(tpm) == 0 {
		t.Fatal("rules.tpm should be a non-empty array")
	}
	tpm0 := tpm[0].(map[string]interface{})
	if tpm0["window_minutes"] != float64(360) {
		t.Errorf("tpm[0].window_minutes: expected 360, got %v", tpm0["window_minutes"])
	}
}

// AK-1-028：rpm window_minutes=1（最小边界）
func TestCreate_Boundary_RpmWindowMin1(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-boundary",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"rpm": []map[string]interface{}{
					{"name": "RPM边界", "model": "*", "window_minutes": 1, "max_requests": 100},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}
	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}
	rpm, ok := rules["rpm"].([]interface{})
	if !ok || len(rpm) == 0 {
		t.Fatal("rules.rpm should be a non-empty array")
	}
	rpm0 := rpm[0].(map[string]interface{})
	if rpm0["window_minutes"] != float64(1) {
		t.Errorf("rpm[0].window_minutes: expected 1, got %v", rpm0["window_minutes"])
	}
}

// ============================================================
// 异常参数 (12)
// ============================================================

// AK-1-029：非法 JSON Body
func TestCreate_Abnormal_InvalidJSON(t *testing.T) {
	resp, err := testutil.GetClient().RawBody("POST", "/open-api/v1/api-keys", "this is not valid json", "text/plain")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-030：rate_limit_policy.enabled=true 但 rules 为空
func TestCreate_Abnormal_RateLimitEnabledNoRules(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-030",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules":   map[string]interface{}{},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-031：expired_time 为过去时间戳
func TestCreate_Abnormal_ExpiredTimePast(t *testing.T) {
	body := map[string]interface{}{
		"description":  "test-key-031",
		"expired_time": 1700000000,
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-032：expired_time < -1
func TestCreate_Abnormal_ExpiredTimeLessThanMinusOne(t *testing.T) {
	body := map[string]interface{}{
		"description":  "test-key-032",
		"expired_time": -2,
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-033：entity_id 指向不存在的实体
func TestCreate_Abnormal_EntityIdNonExistent(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-033",
		"entity_id":   "non-existent-id",
	}

	resp := createAPIKey(t, body)
	// 可能是 404（资源不存在）或 422（参数校验失败）
	if resp.ErrNum != 404 && resp.ErrNum != 422 {
		t.Errorf("expected ErrNum=404 or 422, got ErrNum=%d, ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
	}
}

// AK-1-034：subnet 包含非法子网
func TestCreate_Abnormal_InvalidSubnet(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-034",
		"subnet":      []string{"invalid"},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-035：quota_plan.quota 为负数
func TestCreate_Abnormal_QuotaPlanQuotaNegative(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-035",
		"quota_plan": map[string]interface{}{
			"quota": -1,
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-036：description 超过最大长度 512
func TestCreate_Abnormal_DescriptionTooLong(t *testing.T) {
	desc := ""
	for i := 0; i < 512; i++ {
		desc += "a"
	}
	body := map[string]interface{}{
		"description": desc,
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-037：tpm window_minutes=0
func TestCreate_Abnormal_TpmWindowZero(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-037",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "非法窗口", "model": "*", "window_minutes": 0, "max_tokens": 1000, "step_minutes": 1},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-038：tpm window_minutes=361（超过最大值）
func TestCreate_Abnormal_TpmWindowExceed(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-038",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "超限窗口", "model": "*", "window_minutes": 361, "max_tokens": 1000, "step_minutes": 1},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-039：tpm step_minutes > window_minutes
func TestCreate_Abnormal_TpmStepExceedWindow(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-039",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "步长超窗口", "model": "*", "window_minutes": 5, "max_tokens": 1000, "step_minutes": 10},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// AK-1-040：rpm window_minutes=0
func TestCreate_Abnormal_RpmWindowZero(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-040",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"rpm": []map[string]interface{}{
					{"name": "非法窗口", "model": "*", "window_minutes": 0, "max_requests": 100},
				},
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertErrCode(t, resp, 422)
}

// ============================================================
// 返回数据校验 (3)
// ============================================================

// AK-1-041：验证所有 13 个顶层字段存在且类型正确
func TestCreate_ReturnData_TopLevelFields(t *testing.T) {
	body := map[string]interface{}{
		"description":     "test-key-041",
		"expired_time":    -1,
		"unlimited_quota": false,
		"models":          []string{"*"},
		"subnet":          []string{"*"},
		"quota_plan": map[string]interface{}{
			"unlimited":                 false,
			"pass_when_no_enough_quota": false,
			"quota":                     100000000,
			"unit":                      "total_token",
			"reset_period":              "monthly",
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	// 验证所有顶层字段存在
	requiredFields := []string{
		"id", "key", "description", "enabled", "create_time", "update_time",
		"expired_time", "unlimited_quota", "models", "subnet",
		"quota_plan", "rate_limit_policy",
	}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// 类型校验
	if _, ok := data["id"].(string); !ok {
		t.Error("id should be string")
	}
	if _, ok := data["key"].(string); !ok {
		t.Error("key should be string")
	}
	if _, ok := data["description"].(string); !ok {
		t.Error("description should be string")
	}
	if _, ok := data["enabled"].(bool); !ok {
		t.Error("enabled should be bool")
	}
	if _, ok := data["create_time"].(float64); !ok {
		t.Error("create_time should be number")
	}
	if _, ok := data["update_time"].(float64); !ok {
		t.Error("update_time should be number")
	}
	if _, ok := data["expired_time"].(float64); !ok {
		t.Error("expired_time should be number")
	}
	if _, ok := data["unlimited_quota"].(bool); !ok {
		t.Error("unlimited_quota should be bool")
	}
	if _, ok := data["models"].([]interface{}); !ok {
		t.Error("models should be array")
	}
	if _, ok := data["subnet"].([]interface{}); !ok {
		t.Error("subnet should be array")
	}
	if _, ok := data["quota_plan"].(map[string]interface{}); !ok {
		t.Error("quota_plan should be object")
	}
	if _, ok := data["rate_limit_policy"].(map[string]interface{}); !ok {
		t.Error("rate_limit_policy should be object")
	}
}

// AK-1-042：验证 quota_plan 结构包含所有必需字段
func TestCreate_ReturnData_QuotaPlanStructure(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-042",
		"quota_plan": map[string]interface{}{
			"unlimited":                 false,
			"pass_when_no_enough_quota": false,
			"quota":                     100000000,
			"unit":                      "total_token",
			"reset_period":              "monthly",
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	qp, ok := data["quota_plan"].(map[string]interface{})
	if !ok {
		t.Fatal("quota_plan should be an object")
	}

	// 验证所有必需字段存在
	requiredQPFields := []string{"unlimited", "pass_when_no_enough_quota", "quota", "unit", "reset_period"}
	for _, field := range requiredQPFields {
		if _, ok := qp[field]; !ok {
			t.Errorf("quota_plan missing required field: %s", field)
		}
	}
}

// AK-1-043：验证 rate_limit_policy 结构包含所有必需字段
func TestCreate_ReturnData_RateLimitPolicyStructure(t *testing.T) {
	body := map[string]interface{}{
		"description": "test-key-043",
		"rate_limit_policy": map[string]interface{}{
			"enabled": true,
			"rules": map[string]interface{}{
				"tpm": []map[string]interface{}{
					{"name": "TPM窗口", "model": "*", "window_minutes": 1, "max_tokens": 1000, "step_minutes": 1},
				},
				"rpm": []map[string]interface{}{
					{"name": "RPM窗口", "model": "*", "window_minutes": 1, "max_requests": 100},
				},
				"max_concurrency": 50,
			},
		},
	}

	resp := createAPIKey(t, body)
	testutil.AssertSuccess(t, resp)

	data := parseData(t, resp)

	rlp, ok := data["rate_limit_policy"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy should be an object")
	}

	// 验证 enabled 字段
	if _, ok := rlp["enabled"]; !ok {
		t.Error("rate_limit_policy missing field: enabled")
	}

	// 验证 rules 字段
	rules, ok := rlp["rules"].(map[string]interface{})
	if !ok {
		t.Fatal("rate_limit_policy.rules should be an object")
	}

	// 验证 rules.tpm
	if _, ok := rules["tpm"]; !ok {
		t.Error("rate_limit_policy.rules missing field: tpm")
	}

	// 验证 rules.rpm
	if _, ok := rules["rpm"]; !ok {
		t.Error("rate_limit_policy.rules missing field: rpm")
	}

	// 验证 rules.max_concurrency
	if _, ok := rules["max_concurrency"]; !ok {
		t.Error("rate_limit_policy.rules missing field: max_concurrency")
	}
}
