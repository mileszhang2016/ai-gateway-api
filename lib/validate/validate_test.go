package validate

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

func TestHostname(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "backend-1.example.com", false},
		{"valid ip", "192.0.2.1", false},
		{"valid ipv6", "2001:0db8::1", false},
		{"too short", "a", true},
		{"empty", "", true},
		{"too long", string(make([]byte, 256)), true},
		{"label starts with hyphen", "-host.example.com", true},
		{"label ends with hyphen", "host-.example.com", true},
		{"valid label with hyphen", "host-1.example.com", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Hostname(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPAddress(t *testing.T) {
	assert.NoError(t, IPAddress("192.0.2.1"))
	assert.NoError(t, IPAddress("::1"))
	assert.Error(t, IPAddress("not-an-ip"))
}

func TestPort(t *testing.T) {
	assert.NoError(t, Port(1))
	assert.NoError(t, Port(65535))
	assert.Error(t, Port(0))
	assert.Error(t, Port(65536))
}

func TestCIDR(t *testing.T) {
	assert.NoError(t, CIDR("*"))
	assert.NoError(t, CIDR("192.0.2.0/24"))
	assert.NoError(t, CIDR("2001:0db8::/32"))
	assert.Error(t, CIDR("invalid"))
}

func TestUserName(t *testing.T) {
	assert.NoError(t, UserName("user_1"))
	assert.Error(t, UserName("admin"))
	assert.Error(t, UserName("-user"))
	assert.Error(t, UserName("user."))
	assert.Error(t, UserName("user name"))
	assert.Error(t, UserName(""))
}

func TestPassword(t *testing.T) {
	assert.NoError(t, Password("password123", "user1"))
	assert.Error(t, Password("short1", "user1"))
	assert.Error(t, Password("user1", "user1"))
	assert.Error(t, Password("1resu", "user1"))
	assert.Error(t, Password("pass word", "user1"))
}

func TestTokenName(t *testing.T) {
	assert.NoError(t, TokenName("token_1"))
	assert.Error(t, TokenName("default"))
	assert.Error(t, TokenName("-token"))
}

func TestClusterName(t *testing.T) {
	assert.NoError(t, ClusterName("cluster_1"))
	assert.Error(t, ClusterName("-cluster"))
	assert.Error(t, ClusterName("cluster."))
}

func TestEntityTypeName(t *testing.T) {
	assert.NoError(t, EntityTypeName("dep_1"))
	assert.Error(t, EntityTypeName("Dep"))
	assert.Error(t, EntityTypeName("-dep"))
}

func TestAPIKeyDescription(t *testing.T) {
	assert.NoError(t, APIKeyDescription("valid desc"))
	assert.Error(t, APIKeyDescription(""))
	long := make([]byte, 513)
	assert.Error(t, APIKeyDescription(string(long)))
}

func TestAPIKeyValue(t *testing.T) {
	assert.NoError(t, APIKeyValue("ak-123_test"))
	assert.Error(t, APIKeyValue(""))
	assert.Error(t, APIKeyValue("ak@123"))

	// 1-128 characters are allowed; 129 characters should be rejected to match
	// the api-keys.md definition and the MySQL DDL (varchar(128)).
	assert.NoError(t, APIKeyValue(strings.Repeat("a", 128)))
	assert.Error(t, APIKeyValue(strings.Repeat("a", 129)))
}

func TestQuotaPlan(t *testing.T) {
	assert.NoError(t, QuotaPlan(nil))
	q := int64(-1)
	assert.Error(t, QuotaPlan(&shared.QuotaPlanParam{Quota: &q}))
	q = 100
	unit := "invalid"
	assert.Error(t, QuotaPlan(&shared.QuotaPlanParam{Quota: &q, Unit: &unit}))
	unit = "total_token"
	assert.NoError(t, QuotaPlan(&shared.QuotaPlanParam{Quota: &q, Unit: &unit}))
}

func TestRateLimitPolicy(t *testing.T) {
	assert.NoError(t, RateLimitPolicy(nil))
	enabled := true
	policy := &shared.RateLimitPolicyParam{Enabled: &enabled}
	assert.Error(t, RateLimitPolicy(policy))

	policy.Rules = &shared.RateLimitRules{
		TpmConfigs: []shared.TPMConfig{
			{Name: "t1", Model: "*", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
		},
	}
	assert.NoError(t, RateLimitPolicy(policy))

	policy.Rules.TpmConfigs = append(policy.Rules.TpmConfigs, shared.TPMConfig{Name: "t1", Model: "*", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1})
	assert.Error(t, RateLimitPolicy(policy))
}

func TestRouteRules(t *testing.T) {
	name := "r1"
	cond := "default_t()"
	cluster := "cluster_1"
	weight := 100
	rules := &shared.RouteRulesParam{
		Rules: []*shared.AiRouteRuleParam{
			{
				Name:    &name,
				Cond:    &cond,
				Targets: []*shared.AiRouteTargetParam{{ClusterName: &cluster, Weight: &weight}},
			},
		},
	}
	assert.NoError(t, RouteRules(rules))

	weight = 50
	assert.Error(t, RouteRules(rules))
}

func TestLLMConfig(t *testing.T) {
	c := &icluster_conf.LLMConfig{
		Models: []string{"m1"},
		ModelMappings: []*icluster_conf.Mapping{
			{SourceModel: lib.PString("old"), TargetModel: lib.PString("new")},
		},
		Keys: []icluster_conf.APIKey{
			{Name: lib.PString("key-primary"), Key: lib.PString("sk-aaa"), Weight: lib.PInt(70)},
			{Name: lib.PString("key-secondary"), Key: lib.PString("sk-bbb"), Weight: lib.PInt(30)},
		},
		KeyPolicy: &icluster_conf.KeyPolicy{
			Strategy:            lib.PString("weighted_random"),
			MaxRetries:          lib.PInt(3),
			RetryBackoffInitial: lib.PInt(500),
			RetryBackoffMax:     lib.PInt(5000),
		},
	}
	assert.NoError(t, LLMConfig(c))

	// duplicate model
	c2 := *c
	c2.Models = []string{"m1", "m1"}
	assert.Error(t, LLMConfig(&c2))

	// total weight not 100
	c3 := *c
	c3.Keys = []icluster_conf.APIKey{
		{Name: lib.PString("k1"), Key: lib.PString("sk-aaa"), Weight: lib.PInt(50)},
		{Name: lib.PString("k2"), Key: lib.PString("sk-bbb"), Weight: lib.PInt(30)},
	}
	assert.Error(t, LLMConfig(&c3))

	// duplicate key name
	c4 := *c
	c4.Keys = []icluster_conf.APIKey{
		{Name: lib.PString("k1"), Key: lib.PString("sk-aaa"), Weight: lib.PInt(50)},
		{Name: lib.PString("k1"), Key: lib.PString("sk-bbb"), Weight: lib.PInt(50)},
	}
	assert.Error(t, LLMConfig(&c4))

	// API_KEY placeholder without keys
	c5 := &icluster_conf.LLMConfig{
		Models: []string{"m1"},
		ModelEndpoint: &icluster_conf.Endpoint{
			Schema: "https",
			URI:    "/v1/models",
			Headers: map[string]string{
				"Authorization": "Bearer ${API_KEY}",
			},
		},
	}
	assert.Error(t, LLMConfig(c5))

	// invalid key_policy retry_backoff_max < retry_backoff_initial
	c6 := *c
	c6.KeyPolicy = &icluster_conf.KeyPolicy{
		Strategy:            lib.PString("weighted_random"),
		MaxRetries:          lib.PInt(3),
		RetryBackoffInitial: lib.PInt(500),
		RetryBackoffMax:     lib.PInt(100),
	}
	assert.Error(t, LLMConfig(&c6))

	// invalid key_policy strategy
	c7 := *c
	c7.KeyPolicy = &icluster_conf.KeyPolicy{
		Strategy: lib.PString("invalid"),
	}
	assert.Error(t, LLMConfig(&c7))
}

func TestInstancePool(t *testing.T) {
	instances := []icluster_conf.Instance{
		{Name: "backend-1", Addr: "10.0.0.1", Port: 8080, Weight: 100},
	}
	assert.NoError(t, InstancePool(instances))

	instances[0].Weight = 0
	assert.Error(t, InstancePool(instances))

	// duplicate name
	instances[0].Weight = 50
	assert.Error(t, InstancePool([]icluster_conf.Instance{
		{Name: "backend-1", Addr: "10.0.0.1", Port: 8080, Weight: 50},
		{Name: "backend-1", Addr: "10.0.0.2", Port: 8080, Weight: 50},
	}))

	// duplicate (name, addr)
	assert.Error(t, InstancePool([]icluster_conf.Instance{
		{Name: "backend-1", Addr: "10.0.0.1", Port: 8080, Weight: 50},
		{Name: "backend-1", Addr: "10.0.0.1", Port: 8081, Weight: 50},
	}))

	// same addr with empty name is not allowed
	assert.Error(t, InstancePool([]icluster_conf.Instance{
		{Addr: "10.0.0.1", Port: 8080, Weight: 50},
		{Addr: "10.0.0.1", Port: 8081, Weight: 50},
	}))
}

func TestExpiredTime(t *testing.T) {
	minusOne := int64(-1)
	future := time.Now().Unix() + 1000
	past := time.Now().Unix() - 1000
	minusTwo := int64(-2)
	assert.NoError(t, ExpiredTime(&minusOne))
	assert.NoError(t, ExpiredTime(&future))
	assert.Error(t, ExpiredTime(&minusTwo))
	assert.Error(t, ExpiredTime(&past))
	assert.NoError(t, ExpiredTime(nil))
}
