// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
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
			Name:        lib.PString("test-cluster"),
			InstancePool: []*Instance{
				{Name: "rs1", Addr: "127.0.0.1", Port: 80, Weight: 10},
			},
			LLMConfig: &icluster_conf.LLMConfig{
				Models: []string{"gpt-4"},
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
