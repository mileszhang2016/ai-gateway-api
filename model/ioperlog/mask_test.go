// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package ioperlog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskAPIKeyToken(t *testing.T) {
	assert.Equal(t, "******", MaskAPIKeyToken("short"))
	assert.Equal(t, "abcd****ijkl", MaskAPIKeyToken("abcdefghijkl"))
	assert.Equal(t, "abcd****fghi", MaskAPIKeyToken("abcdefghi"))
}

func TestMaskSensitiveFields(t *testing.T) {
	input := map[string]interface{}{
		"name":        "test",
		"password":    "secret123",
		"token":       "rawtoken1234567890",
		"api_key":     "abcdefghijkl",
		"certificate": "-----BEGIN CERTIFICATE-----",
		"nested": map[string]interface{}{
			"secret": "nested-secret",
			"token":  "nested-raw-token",
		},
	}

	result := MaskSensitiveFields(input)
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "******", result["password"])
	assert.Equal(t, "******", result["token"])
	assert.Equal(t, "abcd****ijkl", result["api_key"])
	assert.Equal(t, "[已更新]", result["certificate"])

	nested, ok := result["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "******", nested["secret"])
	assert.Equal(t, "******", nested["token"])

	serialized, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), "rawtoken1234567890")
	assert.NotContains(t, string(serialized), "nested-raw-token")
	assert.NotContains(t, string(serialized), "secret123")
}

func TestMaskSensitiveFields_Arrays(t *testing.T) {
	input := map[string]interface{}{
		"keys": []interface{}{
			map[string]interface{}{"name": "k1", "key": "abcdefghijkl"},
			map[string]interface{}{"name": "k2", "key": "mnopqrstuvwx"},
		},
		"nested": []interface{}{
			[]interface{}{
				map[string]interface{}{"token": "deep-array-token", "note": "keep"},
			},
		},
		"tags": []interface{}{"plain", "strings"},
	}

	result := MaskSensitiveFields(input)

	keys, ok := result["keys"].([]interface{})
	require.True(t, ok)
	require.Len(t, keys, 2)
	k0, ok := keys[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "k1", k0["name"])
	assert.Equal(t, "abcd****ijkl", k0["key"])
	k1, ok := keys[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "k2", k1["name"])
	assert.Equal(t, "mnop****uvwx", k1["key"])

	outer, ok := result["nested"].([]interface{})
	require.True(t, ok)
	inner, ok := outer[0].([]interface{})
	require.True(t, ok)
	m, ok := inner[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "******", m["token"])
	assert.Equal(t, "keep", m["note"])

	assert.Equal(t, []interface{}{"plain", "strings"}, result["tags"])

	serialized, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), "abcdefghijkl")
	assert.NotContains(t, string(serialized), "deep-array-token")
}
