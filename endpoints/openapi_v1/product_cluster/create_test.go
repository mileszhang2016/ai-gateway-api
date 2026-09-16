// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package product_cluster

import (
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeStickySessions(t *testing.T) {
	t.Run("nil defaults to disabled CLIENT_IP_ONLY", func(t *testing.T) {
		got := normalizeStickySessions(nil)
		require.NotNil(t, got)
		assert.Equal(t, false, *got.Enabled)
		assert.Equal(t, clusterHashStrategyClientIPOnly, *got.HashStrategy)
		assert.Equal(t, "", *got.HashHeader)
	})

	t.Run("partial values keep provided ones", func(t *testing.T) {
		got := normalizeStickySessions(&StickySessionsParam{
			Enabled: lib.PBool(true),
		})
		require.NotNil(t, got)
		assert.Equal(t, true, *got.Enabled)
		assert.Equal(t, clusterHashStrategyClientIPOnly, *got.HashStrategy)
		assert.Equal(t, "", *got.HashHeader)
	})
}

func TestValidateStickySessions(t *testing.T) {
	t.Run("nil sticky sessions", func(t *testing.T) {
		assert.NoError(t, validateStickySessions(nil))
	})

	t.Run("disabled", func(t *testing.T) {
		assert.NoError(t, validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(false),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString(""),
		}))
	})

	t.Run("enabled with CLIENT_ID_ONLY and hash_header", func(t *testing.T) {
		assert.NoError(t, validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString("Cookie:USERID"),
		}))
	})

	t.Run("enabled with CLIENT_ID_PREFERED and hash_header", func(t *testing.T) {
		assert.NoError(t, validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDPrefered),
			HashHeader:   lib.PString("Cookie:USERID"),
		}))
	})

	t.Run("enabled with CLIENT_IP_ONLY and empty hash_header", func(t *testing.T) {
		assert.NoError(t, validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIPOnly),
			HashHeader:   lib.PString(""),
		}))
	})

	t.Run("enabled with default hash_strategy and empty hash_header fails", func(t *testing.T) {
		err := validateStickySessions(&StickySessionsParam{
			Enabled:    lib.PBool(true),
			HashHeader: lib.PString(""),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash_header is required")
	})

	t.Run("enabled with CLIENT_ID_ONLY and empty hash_header fails", func(t *testing.T) {
		err := validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString(""),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash_header is required")
	})

	t.Run("enabled with CLIENT_ID_PREFERED and empty hash_header fails", func(t *testing.T) {
		err := validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDPrefered),
			HashHeader:   lib.PString(""),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash_header is required")
	})

	t.Run("invalid hash_strategy", func(t *testing.T) {
		err := validateStickySessions(&StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString("INVALID"),
			HashHeader:   lib.PString("Cookie:USERID"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash_strategy must be one of")
	})
}

func TestUpsertParamValidate_StickySessions(t *testing.T) {
	base := func() *UpsertParam {
		return &UpsertParam{
			Name: lib.PString("test-cluster"),
			LLMConfig: &icluster_conf.LLMConfig{
				Provider: lib.PString("openai"),
				Models:   []string{"gpt-4"},
			},
		}
	}

	t.Run("disabled sticky sessions with CLIENT_ID_ONLY and empty hash_header is allowed", func(t *testing.T) {
		p := base()
		p.StickySessions = &StickySessionsParam{
			Enabled:      lib.PBool(false),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString(""),
		}
		assert.NoError(t, p.Validate())
	})

	t.Run("nil sticky sessions is allowed", func(t *testing.T) {
		p := base()
		assert.NoError(t, p.Validate())
	})

	t.Run("enabled sticky sessions with CLIENT_ID_ONLY and empty hash_header fails", func(t *testing.T) {
		p := base()
		p.StickySessions = &StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString(""),
		}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "hash_header is required")
	})

	t.Run("enabled sticky sessions with CLIENT_ID_ONLY and hash_header passes", func(t *testing.T) {
		p := base()
		p.StickySessions = &StickySessionsParam{
			Enabled:      lib.PBool(true),
			HashStrategy: lib.PString(clusterHashStrategyClientIDOnly),
			HashHeader:   lib.PString("Cookie:USERID"),
		}
		assert.NoError(t, p.Validate())
	})
}

func TestNormalizeLLMConfig(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, normalizeLLMConfig(nil))
	})

	t.Run("copies prefix fields", func(t *testing.T) {
		in := &icluster_conf.LLMConfig{
			Models:      []string{"gpt-4"},
			MatchPrefix: lib.PString("openrouter/"),
			StripPrefix: lib.PBool(true),
		}
		got := normalizeLLMConfig(in)
		require.NotNil(t, got)
		assert.Equal(t, "openrouter/", *got.MatchPrefix)
		assert.Equal(t, true, *got.StripPrefix)
	})

	t.Run("copies key_affinity", func(t *testing.T) {
		in := &icluster_conf.LLMConfig{
			Models: []string{"gpt-4"},
			KeyAffinity: &icluster_conf.KeyAffinity{
				Enabled:       lib.PBool(true),
				TTL:           lib.PInt(600),
				RedisPrefix:   lib.PString("bfe:ai:key_affinity"),
				PenaltyEnable: lib.PBool(true),
			},
		}
		got := normalizeLLMConfig(in)
		require.NotNil(t, got)
		require.NotNil(t, got.KeyAffinity)
		assert.Equal(t, true, *got.KeyAffinity.Enabled)
		assert.Equal(t, 600, *got.KeyAffinity.TTL)
		assert.Equal(t, "bfe:ai:key_affinity", *got.KeyAffinity.RedisPrefix)
		assert.Equal(t, true, *got.KeyAffinity.PenaltyEnable)
	})

	t.Run("key_affinity fills defaults when empty object", func(t *testing.T) {
		in := &icluster_conf.LLMConfig{
			Models:      []string{"gpt-4"},
			KeyAffinity: &icluster_conf.KeyAffinity{},
		}
		got := normalizeLLMConfig(in)
		require.NotNil(t, got)
		require.NotNil(t, got.KeyAffinity)
		assert.Equal(t, true, *got.KeyAffinity.Enabled)
		assert.Equal(t, 600, *got.KeyAffinity.TTL)
		assert.Equal(t, "bfe:ai:key_affinity", *got.KeyAffinity.RedisPrefix)
		assert.Equal(t, true, *got.KeyAffinity.PenaltyEnable)
	})

	t.Run("key_affinity fills defaults when nil", func(t *testing.T) {
		in := &icluster_conf.LLMConfig{
			Models: []string{"gpt-4"},
		}
		got := normalizeLLMConfig(in)
		require.NotNil(t, got)
		require.NotNil(t, got.KeyAffinity)
		assert.Equal(t, true, *got.KeyAffinity.Enabled)
		assert.Equal(t, 600, *got.KeyAffinity.TTL)
		assert.Equal(t, "bfe:ai:key_affinity", *got.KeyAffinity.RedisPrefix)
		assert.Equal(t, true, *got.KeyAffinity.PenaltyEnable)
	})
}

// TestValidatePassiveHealthCheck verifies the passive_health_check legality
// conditions from the clusters contract (issue #172).
func TestValidatePassiveHealthCheck(t *testing.T) {
	t.Run("nil is allowed", func(t *testing.T) {
		assert.NoError(t, validatePassiveHealthCheck(nil))
	})

	t.Run("empty object is allowed", func(t *testing.T) {
		assert.NoError(t, validatePassiveHealthCheck(&PassiveHealthCheckParam{}))
	})

	t.Run("negative failnum", func(t *testing.T) {
		err := validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Failnum: lib.PInt32(-1),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "passive_health_check.failnum must be >= 0")
	})

	t.Run("negative interval", func(t *testing.T) {
		err := validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Interval: lib.PInt32(-1),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "passive_health_check.interval must be >= 0")
	})

	t.Run("statuscode out of range", func(t *testing.T) {
		for _, code := range []int32{99, 600, 999} {
			err := validatePassiveHealthCheck(&PassiveHealthCheckParam{
				Statuscode: lib.PInt32(code),
			})
			require.Error(t, err, "statuscode=%d should be rejected", code)
			assert.Contains(t, err.Error(), "passive_health_check.statuscode must be 0 or in [100, 599]")
		}
	})

	t.Run("uri without leading slash", func(t *testing.T) {
		err := validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Uri: lib.PString("healthz"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "passive_health_check.uri must be non-empty and start with '/'")
	})

	t.Run("empty uri", func(t *testing.T) {
		err := validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Uri: lib.PString(""),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "passive_health_check.uri must be non-empty and start with '/'")
	})

	t.Run("boundary values are allowed", func(t *testing.T) {
		assert.NoError(t, validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Failnum:    lib.PInt32(0),
			Interval:   lib.PInt32(0),
			Statuscode: lib.PInt32(0),
			Uri:        lib.PString("/"),
		}))
		assert.NoError(t, validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Statuscode: lib.PInt32(100),
		}))
		assert.NoError(t, validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Statuscode: lib.PInt32(599),
		}))
		assert.NoError(t, validatePassiveHealthCheck(&PassiveHealthCheckParam{
			Uri: lib.PString("/healthz"),
		}))
	})
}

// TestUpsertParamValidate_PassiveHealthCheck verifies that UpsertParam.Validate
// (the single hook shared by POST /clusters and PATCH /clusters/{name}) rejects
// illegal passive_health_check values (issue #172).
func TestUpsertParamValidate_PassiveHealthCheck(t *testing.T) {
	base := func() *UpsertParam {
		return &UpsertParam{
			Name: lib.PString("test-cluster"),
			LLMConfig: &icluster_conf.LLMConfig{
				Provider: lib.PString("openai"),
				Models:   []string{"gpt-4"},
			},
		}
	}

	t.Run("nil passive health check is allowed", func(t *testing.T) {
		assert.NoError(t, base().Validate())
	})

	t.Run("failnum=-1 fails", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{Failnum: lib.PInt32(-1)}
		require.Error(t, p.Validate())
	})

	t.Run("interval=-1 fails", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{Interval: lib.PInt32(-1)}
		require.Error(t, p.Validate())
	})

	t.Run("statuscode=999 fails", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{Statuscode: lib.PInt32(999)}
		require.Error(t, p.Validate())
	})

	t.Run("statuscode=0 explicit is allowed", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{Statuscode: lib.PInt32(0)}
		assert.NoError(t, p.Validate())
	})

	t.Run("uri without leading slash fails", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{Uri: lib.PString("no-slash")}
		require.Error(t, p.Validate())
	})

	t.Run("legal full config passes", func(t *testing.T) {
		p := base()
		p.PassiveHealthCheck = &PassiveHealthCheckParam{
			Interval:   lib.PInt32(1000),
			Failnum:    lib.PInt32(3),
			Statuscode: lib.PInt32(200),
			Uri:        lib.PString("/"),
		}
		assert.NoError(t, p.Validate())
	})
}
