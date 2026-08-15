package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAiRouteRuleParamUnmarshalJSON_BackwardCompatible(t *testing.T) {
	legacyJSON := `{
		"name": "rule1",
		"Cond": "default_t()",
		"targets": [
			{"ClusterName": "c1", "Model": "m1", "Weight": 70},
			{"ClusterName": "c2", "Model": "", "Weight": 30}
		],
		"fallbacks": [
			{"ClusterName": "fb", "Model": "m2"}
		]
	}`

	var rule AiRouteRuleParam
	require.NoError(t, json.Unmarshal([]byte(legacyJSON), &rule))

	assert.Equal(t, "rule1", *rule.Name)
	assert.Equal(t, "default_t()", *rule.Cond)
	require.Len(t, rule.Targets, 2)
	assert.Equal(t, "c1", *rule.Targets[0].ClusterName)
	assert.Equal(t, "m1", *rule.Targets[0].Model)
	assert.Equal(t, 70, *rule.Targets[0].Weight)
	assert.Equal(t, "c2", *rule.Targets[1].ClusterName)
	require.Len(t, rule.Fallbacks, 1)
	assert.Equal(t, "fb", *rule.Fallbacks[0].ClusterName)
	assert.Equal(t, "m2", *rule.Fallbacks[0].Model)
}

func TestAiRouteRuleParamUnmarshalJSON_NewKeys(t *testing.T) {
	newJSON := `{
		"name": "rule1",
		"cond": "default_t()",
		"targets": [
			{"cluster_name": "c1", "model": "m1", "weight": 100}
		],
		"fallbacks": []
	}`

	var rule AiRouteRuleParam
	require.NoError(t, json.Unmarshal([]byte(newJSON), &rule))

	assert.Equal(t, "default_t()", *rule.Cond)
	require.Len(t, rule.Targets, 1)
	assert.Equal(t, "c1", *rule.Targets[0].ClusterName)
	assert.Equal(t, "m1", *rule.Targets[0].Model)
	assert.Equal(t, 100, *rule.Targets[0].Weight)
}

func TestAiRouteRuleParamMarshalJSON_NewKeysOnly(t *testing.T) {
	cond := "default_t()"
	name := "rule1"
	cluster := "c1"
	model := "m1"
	weight := 100
	rule := AiRouteRuleParam{
		Name: &name,
		Cond: &cond,
		Targets: []*AiRouteTargetParam{
			{ClusterName: &cluster, Model: &model, Weight: &weight},
		},
	}

	data, err := json.Marshal(&rule)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Contains(t, m, "cond")
	assert.NotContains(t, m, "Cond")

	targets := m["targets"].([]interface{})
	target := targets[0].(map[string]interface{})
	assert.Contains(t, target, "cluster_name")
	assert.Contains(t, target, "model")
	assert.Contains(t, target, "weight")
	assert.NotContains(t, target, "ClusterName")
}
