// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yf-networks/ai-gateway-api/lib"
)

func TestValidateRule(t *testing.T) {
	t.Run("valid rule", func(t *testing.T) {
		rule := &Rule{
			Name: "rule_1",
			Basic: &BasicInfo{
				Domain: lib.PString("example.com"),
				ExpectAction: &RouteAction{
					Forward: &ActionForward{ClusterName: "cluster-1"},
				},
			},
		}
		assert.NoError(t, ValidateRule(rule, 0))
	})

	t.Run("nil rule", func(t *testing.T) {
		assert.Error(t, ValidateRule(nil, 0))
	})

	t.Run("empty rule name", func(t *testing.T) {
		rule := &Rule{Name: ""}
		assert.Error(t, ValidateRule(rule, 0))
	})

	t.Run("invalid rule name", func(t *testing.T) {
		rule := &Rule{Name: "-rule"}
		assert.Error(t, ValidateRule(rule, 0))
	})

	t.Run("nil basic info", func(t *testing.T) {
		rule := &Rule{Name: "rule_1"}
		assert.Error(t, ValidateRule(rule, 0))
	})
}

func TestValidateBasicInfo(t *testing.T) {
	buildValidBasic := func() *BasicInfo {
		return &BasicInfo{
			Domain: lib.PString("example.com"),
			ExpectAction: &RouteAction{
				Forward: &ActionForward{ClusterName: "cluster-1"},
			},
		}
	}

	t.Run("valid with domain", func(t *testing.T) {
		assert.NoError(t, validateBasicInfo(buildValidBasic(), "rule_1"))
	})

	t.Run("valid with method", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.Method = lib.PString("POST")
		assert.NoError(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("valid with path filter", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.PathFilter = &PathFilter{
			MatchMode:  lib.PString(MatchModePrefix),
			Path:       lib.PString("/api"),
			IgnoreCase: lib.PBool(false),
		}
		assert.NoError(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("valid with header filter", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.HeaderFilters = []*BasicHeaderFilter{
			{
				Key:        lib.PString("X-Api-Key"),
				Value:      lib.PString("secret"),
				MatchMode:  lib.PString(MatchModeExact),
				IgnoreCase: lib.PBool(false),
			},
		}
		assert.NoError(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("valid with model filter", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.ModelFilter = &ModelFilter{
			Name:       lib.PString("gpt-4"),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
		}
		assert.NoError(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("no condition set", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		assert.Error(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("missing expect_action", func(t *testing.T) {
		basic := buildValidBasic()
		basic.ExpectAction = nil
		assert.Error(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("invalid method", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.Method = lib.PString("INVALID")
		assert.Error(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("invalid domain", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = lib.PString("")
		assert.Error(t, validateBasicInfo(basic, "rule_1"))
	})

	t.Run("empty model filter name", func(t *testing.T) {
		basic := buildValidBasic()
		basic.Domain = nil
		basic.ModelFilter = &ModelFilter{
			Name:       lib.PString(""),
			Pattern:    lib.PString("$.model"),
			IgnoreCase: lib.PBool(false),
		}
		assert.Error(t, validateBasicInfo(basic, "rule_1"))
	})
}

func TestValidatePathFilter(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		assert.NoError(t, validatePathFilter(nil, "rule_1"))
	})

	t.Run("valid prefix filter", func(t *testing.T) {
		pf := &PathFilter{
			MatchMode:  lib.PString(MatchModePrefix),
			Path:       lib.PString("/api"),
			IgnoreCase: lib.PBool(false),
		}
		assert.NoError(t, validatePathFilter(pf, "rule_1"))
	})

	t.Run("missing ignore_case", func(t *testing.T) {
		pf := &PathFilter{MatchMode: lib.PString(MatchModePrefix)}
		assert.Error(t, validatePathFilter(pf, "rule_1"))
	})

	t.Run("missing match_mode", func(t *testing.T) {
		pf := &PathFilter{Path: lib.PString("/api"), IgnoreCase: lib.PBool(false)}
		assert.Error(t, validatePathFilter(pf, "rule_1"))
	})

	t.Run("invalid match_mode", func(t *testing.T) {
		pf := &PathFilter{
			MatchMode:  lib.PString("invalid"),
			Path:       lib.PString("/api"),
			IgnoreCase: lib.PBool(false),
		}
		assert.Error(t, validatePathFilter(pf, "rule_1"))
	})

	t.Run("path too long", func(t *testing.T) {
		longPath := make([]byte, 2049)
		for i := range longPath {
			longPath[i] = 'a'
		}
		pf := &PathFilter{
			MatchMode:  lib.PString(MatchModePrefix),
			Path:       lib.PString(string(longPath)),
			IgnoreCase: lib.PBool(false),
		}
		assert.Error(t, validatePathFilter(pf, "rule_1"))
	})
}

func TestValidateExpectAction(t *testing.T) {
	t.Run("valid forward", func(t *testing.T) {
		action := &RouteAction{
			Forward: &ActionForward{ClusterName: "cluster-1"},
		}
		assert.NoError(t, validateExpectAction(action, "rule_1"))
	})

	t.Run("no action", func(t *testing.T) {
		assert.Error(t, validateExpectAction(&RouteAction{}, "rule_1"))
	})

	t.Run("empty cluster name", func(t *testing.T) {
		action := &RouteAction{
			Forward: &ActionForward{ClusterName: ""},
		}
		assert.Error(t, validateExpectAction(action, "rule_1"))
	})

	t.Run("invalid forward url", func(t *testing.T) {
		action := &RouteAction{
			Forward: &ActionForward{ClusterName: "cluster-1", URL: "not-a-url"},
		}
		assert.Error(t, validateExpectAction(action, "rule_1"))
	})
}

func TestValidateBasicHeaderFilter(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		assert.NoError(t, validateBasicHeaderFilter(nil, 0, "rule_1"))
	})

	t.Run("valid filter", func(t *testing.T) {
		filter := &BasicHeaderFilter{
			Key:        lib.PString("X-Api-Key"),
			Value:      lib.PString("secret"),
			MatchMode:  lib.PString(MatchModeExact),
			IgnoreCase: lib.PBool(false),
		}
		assert.NoError(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("empty key", func(t *testing.T) {
		filter := &BasicHeaderFilter{
			Key:   lib.PString(""),
			Value: lib.PString("secret"),
		}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("empty value", func(t *testing.T) {
		filter := &BasicHeaderFilter{
			Key:   lib.PString("X-Api-Key"),
			Value: lib.PString(""),
		}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("key value both nil", func(t *testing.T) {
		filter := &BasicHeaderFilter{}
		assert.NoError(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("only key provided", func(t *testing.T) {
		filter := &BasicHeaderFilter{Key: lib.PString("X-Api-Key")}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("invalid key format", func(t *testing.T) {
		filter := &BasicHeaderFilter{
			Key:   lib.PString("X Api Key"),
			Value: lib.PString("secret"),
		}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("invalid match mode", func(t *testing.T) {
		filter := &BasicHeaderFilter{
			Key:       lib.PString("X-Api-Key"),
			Value:     lib.PString("secret"),
			MatchMode: lib.PString("invalid"),
		}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})

	t.Run("value too long", func(t *testing.T) {
		longValue := make([]byte, 8193)
		for i := range longValue {
			longValue[i] = 'a'
		}
		filter := &BasicHeaderFilter{
			Key:   lib.PString("X-Api-Key"),
			Value: lib.PString(string(longValue)),
		}
		assert.Error(t, validateBasicHeaderFilter(filter, 0, "rule_1"))
	})
}

func TestIsValidRuleName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"rule_1", true},
		{"rule-1", true},
		{"Rule123", true},
		{"", false},
		{"-rule", false},
		{"rule space", false},
		{"rule!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidRuleName(tt.name))
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	assert.True(t, isValidDomain("example.com"))
	assert.False(t, isValidDomain(""))
}

func TestIsValidURL(t *testing.T) {
	assert.True(t, isValidURL("http://example.com"))
	assert.False(t, isValidURL("example.com"))
}

func TestIsValidHeaderKey(t *testing.T) {
	assert.True(t, isValidHeaderKey("X-Api-Key"))
	assert.False(t, isValidHeaderKey("X Api Key"))
	assert.False(t, isValidHeaderKey("X:Api"))
	assert.False(t, isValidHeaderKey("X(Api)"))
}
