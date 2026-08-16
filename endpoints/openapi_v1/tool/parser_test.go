// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModelsWithConfig(t *testing.T) {
	config := &ParserConfig{
		Providers: map[string]FieldMapping{
			"openai": {
				ListPath:  "data",
				IDField:   "id",
				NameField: "name",
			},
		},
		DefaultParser: FieldMapping{
			ListPath:  "data",
			IDField:   "id",
			NameField: "name",
		},
	}

	response := []byte(`{"data":[{"id":"gpt-4","name":"GPT-4"}]}`)
	models, err := ParseModelsWithConfig(response, "openai", config)

	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "gpt-4", models[0]["id"])
	assert.Equal(t, "GPT-4", models[0]["name"])
}

func TestParseModelsWithConfig_UnknownProvider(t *testing.T) {
	config := &ParserConfig{
		DefaultParser: FieldMapping{
			ListPath:  "models",
			IDField:   "id",
			NameField: "name",
		},
	}

	response := []byte(`{"models":[{"id":"model-1","name":"Model 1"}]}`)
	models, err := ParseModelsWithConfig(response, "unknown", config)

	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "model-1", models[0]["id"])
}

func TestParseModelsWithConfig_InvalidJSON(t *testing.T) {
	_, err := ParseModelsWithConfig([]byte("not-json"), "openai", &ParserConfig{})
	require.Error(t, err)
}

func TestExtractModelList_Wildcard(t *testing.T) {
	config := FieldMapping{
		ListPath: "items.*",
		IDField:  "id",
	}
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"id": "m1"},
		},
	}

	models, err := extractModelList(data, config)
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "m1", models[0]["id"])
}

func TestExtractModelList_PathNotFound(t *testing.T) {
	config := FieldMapping{ListPath: "missing.path"}
	_, err := extractModelList(map[string]interface{}{}, config)
	require.Error(t, err)
}

func TestExtractModelFields_CustomFields(t *testing.T) {
	config := FieldMapping{
		IDField:    "id",
		NameField:  "name",
		Created:    "created",
		OwnerField: "owner",
		Custom: map[string]string{
			"vendor": "meta.vendor",
		},
	}
	model := map[string]interface{}{
		"id":      "m1",
		"name":    "Model 1",
		"created": "2024-01-01",
		"owner":   "team-a",
		"meta": map[string]interface{}{
			"vendor": "v1",
		},
	}

	result := extractModelFields(model, config)
	require.NotNil(t, result)
	assert.Equal(t, "m1", result["id"])
	assert.Equal(t, "Model 1", result["name"])
	assert.Equal(t, "2024-01-01", result["created"])
	assert.Equal(t, "team-a", result["owner"])
	assert.Equal(t, "v1", result["vendor"])
}

func TestExtractModelFields_NoID(t *testing.T) {
	config := FieldMapping{IDField: "id"}
	model := map[string]interface{}{"name": "Model 1"}
	assert.Nil(t, extractModelFields(model, config))
}

func TestGetNestedField(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "value",
		},
	}

	value, ok := getNestedField(data, "a.b")
	require.True(t, ok)
	assert.Equal(t, "value", value)

	_, ok = getNestedField(data, "a.c")
	assert.False(t, ok)
}
