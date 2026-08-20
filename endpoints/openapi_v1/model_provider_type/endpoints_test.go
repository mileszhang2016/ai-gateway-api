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

package model_provider_type

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModelProviderTypesAction_FileNotExist(t *testing.T) {
	reader := func() ([]byte, error) {
		return nil, errors.New("no such file")
	}

	_, err := listModelProviderTypesProcess(context.Background(), reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "models.json")
}

func TestListModelProviderTypesAction_Success(t *testing.T) {
	content := `[{"name": "OpenAI", "id": "openai"}, {"name": "Empty", "id": ""}]`
	reader := func() ([]byte, error) {
		return []byte(content), nil
	}

	data, err := listModelProviderTypesProcess(context.Background(), reader)
	require.NoError(t, err)
	types, ok := data.([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"openai"}, types)
}

func TestListModelProviderTypesAction_InvalidJSON(t *testing.T) {
	reader := func() ([]byte, error) {
		return []byte("not-json"), nil
	}

	_, err := listModelProviderTypesProcess(context.Background(), reader)
	require.Error(t, err)
}

func TestEndpoints(t *testing.T) {
	reader := func() ([]byte, error) {
		return []byte(`[{"name": "OpenAI", "id": "openai"}, {"name": "Azure", "id": "azure"}]`), nil
	}

	endpoints := NewEndpoints(reader)
	require.Len(t, endpoints, 1)
	ep := endpoints[0]
	require.NotNil(t, ep)
	assert.Equal(t, "/model-provider-types", ep.Path)
	assert.Equal(t, http.MethodGet, ep.Method)
	assert.NotNil(t, ep.Handler)

	// Ensure GET handler returns the expected provider types.
	req := httptest.NewRequest(http.MethodGet, "/model-provider-types", nil)
	rst := ep.Handler(req)
	require.NotNil(t, rst)
	require.NoError(t, rst.OriginErr)
	types, ok := rst.Data.([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"openai", "azure"}, types)
}
