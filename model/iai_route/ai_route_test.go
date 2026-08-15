// Copyright(c) 2026 The Infinity AI Gateway Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iai_route

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
)

func TestBuildAIRouteCond(t *testing.T) {
	ctx := context.Background()

	t.Run("domain only", func(t *testing.T) {
		cond := BuildAIRouteCond(ctx, &BasicInfo{
			Domain: lib.PString("example.com"),
		})
		assert.Equal(t, `req_host_in("example.com")`, cond)
	})

	t.Run("full conditions", func(t *testing.T) {
		cond := BuildAIRouteCond(ctx, &BasicInfo{
			Domain: lib.PString("example.com"),
			PathFilter: &PathFilter{
				MatchMode:  lib.PString(MatchModePrefix),
				Path:       lib.PString("/api"),
				IgnoreCase: lib.PBool(true),
			},
			Method: lib.PString("POST"),
			HeaderFilters: []*BasicHeaderFilter{
				{
					Key:        lib.PString("X-Api-Key"),
					Value:      lib.PString("secret"),
					MatchMode:  lib.PString(MatchModeExact),
					IgnoreCase: lib.PBool(false),
				},
			},
			ModelFilter: &ModelFilter{
				Name:       lib.PString("gpt-4"),
				Pattern:    lib.PString("$.model"),
				IgnoreCase: lib.PBool(true),
			},
		})

		expected := `req_host_in("example.com")&&req_path_prefix_in("/api", true)&&req_method_in("POST")&&req_header_value_in("X-Api-Key", "secret", false)&&req_body_json_in("$.model", "gpt-4", true)`
		assert.Equal(t, expected, cond)
	})

	t.Run("empty basic info", func(t *testing.T) {
		cond := BuildAIRouteCond(ctx, &BasicInfo{})
		assert.Equal(t, "", cond)
	})
}

func TestBuildDomainCondition(t *testing.T) {
	t.Run("with domain", func(t *testing.T) {
		cond := buildDomainCondition(lib.PString("example.com"))
		assert.Equal(t, `req_host_in("example.com")`, cond)
	})

	t.Run("empty domain", func(t *testing.T) {
		cond := buildDomainCondition(lib.PString(""))
		assert.Equal(t, "", cond)
	})

	t.Run("nil domain", func(t *testing.T) {
		cond := buildDomainCondition(nil)
		assert.Equal(t, "", cond)
	})
}

func TestBuildPathCondition(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		cond := buildPathCondition("base", nil)
		assert.Equal(t, "base", cond)
	})

	t.Run("nil match mode", func(t *testing.T) {
		cond := buildPathCondition("base", &PathFilter{Path: lib.PString("/api")})
		assert.Equal(t, "base", cond)
	})

	t.Run("prefix match", func(t *testing.T) {
		cond := buildPathCondition("base", &PathFilter{
			MatchMode:  lib.PString(MatchModePrefix),
			Path:       lib.PString("/api"),
			IgnoreCase: lib.PBool(true),
		})
		assert.Equal(t, "base&&req_path_prefix_in(\"/api\", true)", cond)
	})

	t.Run("exact match", func(t *testing.T) {
		cond := buildPathCondition("", &PathFilter{
			MatchMode:  lib.PString(MatchModeExact),
			Path:       lib.PString("/api/v1"),
			IgnoreCase: lib.PBool(false),
		})
		assert.Equal(t, `req_path_in("/api/v1", false)`, cond)
	})

	t.Run("suffix match", func(t *testing.T) {
		cond := buildPathCondition("", &PathFilter{
			MatchMode:  lib.PString(MatchModeSuffix),
			Path:       lib.PString(".json"),
			IgnoreCase: lib.PBool(false),
		})
		assert.Equal(t, `req_path_suffix_in(".json", false)`, cond)
	})
}

func TestBuildMethodCondition(t *testing.T) {
	t.Run("with method", func(t *testing.T) {
		cond := buildMethodCondition("base", lib.PString("GET"))
		assert.Equal(t, "base&&req_method_in(\"GET\")", cond)
	})

	t.Run("empty method", func(t *testing.T) {
		cond := buildMethodCondition("base", lib.PString(""))
		assert.Equal(t, "base", cond)
	})

	t.Run("nil method", func(t *testing.T) {
		cond := buildMethodCondition("base", nil)
		assert.Equal(t, "base", cond)
	})
}

func TestBuildHeaderConditions(t *testing.T) {
	t.Run("single exact header", func(t *testing.T) {
		cond := buildHeaderConditions("base", []*BasicHeaderFilter{
			{
				Key:        lib.PString("X-Api-Key"),
				Value:      lib.PString("secret"),
				MatchMode:  lib.PString(MatchModeExact),
				IgnoreCase: lib.PBool(false),
			},
		})
		assert.Equal(t, "base&&req_header_value_in(\"X-Api-Key\", \"secret\", false)", cond)
	})

	t.Run("multiple headers", func(t *testing.T) {
		cond := buildHeaderConditions("", []*BasicHeaderFilter{
			{
				Key:        lib.PString("X-Api-Key"),
				Value:      lib.PString("secret"),
				MatchMode:  lib.PString(MatchModeExact),
				IgnoreCase: lib.PBool(false),
			},
			{
				Key:        lib.PString("X-Env"),
				Value:      lib.PString("prod"),
				MatchMode:  lib.PString(MatchModePrefix),
				IgnoreCase: lib.PBool(true),
			},
		})
		assert.Equal(t, `req_header_value_in("X-Api-Key", "secret", false)&&req_header_value_prefix_in("X-Env", "prod", true)`, cond)
	})

	t.Run("nil match mode skipped", func(t *testing.T) {
		cond := buildHeaderConditions("base", []*BasicHeaderFilter{
			{Key: lib.PString("X-Api-Key"), Value: lib.PString("secret")},
		})
		assert.Equal(t, "base", cond)
	})

	t.Run("suffix match", func(t *testing.T) {
		cond := buildHeaderConditions("", []*BasicHeaderFilter{
			{
				Key:        lib.PString("X-Api-Key"),
				Value:      lib.PString("suffix"),
				MatchMode:  lib.PString(MatchModeSuffix),
				IgnoreCase: lib.PBool(false),
			},
		})
		assert.Equal(t, `req_header_value_suffix_in("X-Api-Key", "suffix", false)`, cond)
	})
}

func TestBuildModelCondition(t *testing.T) {
	t.Run("exact model filter", func(t *testing.T) {
		cond := buildModelCondition("base", &ModelFilter{
			Name:       lib.PString("gpt-4"),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
			MatchMode:  lib.PString(MatchModeExact),
		})
		assert.Equal(t, `base&&req_body_json_in("$.model", "gpt-4", false)`, cond)
	})

	t.Run("exact model filter by default", func(t *testing.T) {
		cond := buildModelCondition("base", &ModelFilter{
			Name:       lib.PString("gpt-4"),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
		})
		assert.Equal(t, `base&&req_body_json_in("$.model", "gpt-4", false)`, cond)
	})

	t.Run("prefix model filter", func(t *testing.T) {
		cond := buildModelCondition("base", &ModelFilter{
			Name:       lib.PString("openrouter/"),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
			MatchMode:  lib.PString(MatchModePrefix),
		})
		assert.Equal(t, `base&&req_body_json_prefix_in("$.model", "openrouter/", false)`, cond)
	})

	t.Run("prefix model filter with nested namespace", func(t *testing.T) {
		cond := buildModelCondition("base", &ModelFilter{
			Name:       lib.PString("openrouter/anthropic/"),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
			MatchMode:  lib.PString(MatchModePrefix),
		})
		assert.Equal(t, `base&&req_body_json_prefix_in("$.model", "openrouter/anthropic/", false)`, cond)
	})

	t.Run("nil filter", func(t *testing.T) {
		cond := buildModelCondition("base", nil)
		assert.Equal(t, "base", cond)
	})

	t.Run("empty model name", func(t *testing.T) {
		cond := buildModelCondition("base", &ModelFilter{
			Name:    lib.PString(""),
			Pattern: lib.PString("$.model"),
		})
		assert.Equal(t, "base", cond)
	})
}

func TestCombineConditions(t *testing.T) {
	t.Run("new cond empty", func(t *testing.T) {
		assert.Equal(t, "base", combineConditions("base", ""))
	})

	t.Run("current cond empty", func(t *testing.T) {
		assert.Equal(t, "new", combineConditions("", "new"))
	})

	t.Run("both non-empty", func(t *testing.T) {
		assert.Equal(t, "base&&new", combineConditions("base", "new"))
	})
}
