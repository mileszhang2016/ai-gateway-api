package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildBFERateLimitRedisKey(t *testing.T) {
	got := BuildBFERateLimitRedisKey(101, "RL_TPM", "tpm-1")
	assert.Equal(t, "default_bfe_rlp-101_RL_TPM_rlp-101_tpm-1", got)
}

func TestBuildRateLimitRedisKeys(t *testing.T) {
	rules := &RateLimitRules{
		TpmConfigs: []TPMConfig{
			{Name: "tpm-1", Model: "gpt-4", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
			{Name: "tpm-2", Model: "gpt-3.5", WindowMinutes: 10, MaxTokens: 200, StepMinutes: 1},
		},
		RpmConfigs: []RPMConfig{
			{Name: "rpm-1", Model: "*", WindowMinutes: 1, MaxRequests: 10},
		},
	}

	got := BuildRateLimitRedisKeys(101, rules)
	assert.Equal(t, []string{
		"default_bfe_rlp-101_RL_TPM_rlp-101_tpm-1",
		"default_bfe_rlp-101_RL_TPM_rlp-101_tpm-2",
		"default_bfe_rlp-101_RL_RPM_rlp-101_rpm-1",
	}, got)
}

func TestBuildRateLimitRedisKeys_NilRules(t *testing.T) {
	assert.Nil(t, BuildRateLimitRedisKeys(101, nil))
}

func TestDiffRateLimitRedisKeys(t *testing.T) {
	oldRules := &RateLimitRules{
		TpmConfigs: []TPMConfig{
			{Name: "keep", Model: "gpt-4", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
			{Name: "remove", Model: "gpt-3.5", WindowMinutes: 1, MaxTokens: 200, StepMinutes: 1},
		},
		RpmConfigs: []RPMConfig{
			{Name: "rpm-keep", Model: "*", WindowMinutes: 1, MaxRequests: 10},
			{Name: "rpm-remove", Model: "*", WindowMinutes: 1, MaxRequests: 20},
		},
	}

	newRules := &RateLimitRules{
		TpmConfigs: []TPMConfig{
			{Name: "keep", Model: "gpt-4", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
		},
		RpmConfigs: []RPMConfig{
			{Name: "rpm-keep", Model: "*", WindowMinutes: 1, MaxRequests: 10},
		},
	}

	got := DiffRateLimitRedisKeys(101, oldRules, newRules)
	assert.ElementsMatch(t, []string{
		"default_bfe_rlp-101_RL_TPM_rlp-101_remove",
		"default_bfe_rlp-101_RL_RPM_rlp-101_rpm-remove",
	}, got)
}

func TestDiffRateLimitRedisKeys_AllRemoved(t *testing.T) {
	oldRules := &RateLimitRules{
		TpmConfigs: []TPMConfig{
			{Name: "tpm-1", Model: "gpt-4", WindowMinutes: 1, MaxTokens: 100, StepMinutes: 1},
		},
	}

	got := DiffRateLimitRedisKeys(101, oldRules, nil)
	assert.Equal(t, []string{"default_bfe_rlp-101_RL_TPM_rlp-101_tpm-1"}, got)
}

func TestDiffRateLimitRedisKeys_NilOld(t *testing.T) {
	assert.Empty(t, DiffRateLimitRedisKeys(101, nil, &RateLimitRules{}))
}
