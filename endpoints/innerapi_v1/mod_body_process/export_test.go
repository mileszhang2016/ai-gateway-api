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

package mod_body_process

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/imods"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

func setupTestEnv(version string) func() {
	origConfig := stateful.DefaultConfig
	stateful.DefaultConfig = &stateful.Config{
		RunTime: stateful.RunTimeConfig{
			AIRouteInnerProductName: "AI_product",
		},
	}

	old := container.ModBodyProcessManager
	container.ModBodyProcessManager = imods.NewModBodyProcessManager(testutil.NewVersionControlManager(version))

	return func() {
		container.ModBodyProcessManager = old
		stateful.DefaultConfig = origConfig
	}
}

func TestExportAction(t *testing.T) {
	defer setupTestEnv("v2")()

	req := httptest.NewRequest(http.MethodGet, "/configs/mod-body-process?version=", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*imods.ModBodyProcessConf)
	require.True(t, ok)
	assert.Equal(t, []string{}, conf.Config["AI_product"])
	assert.Equal(t, "v2", *conf.Version)
}

func TestExportAction_VersionNotChanged(t *testing.T) {
	defer setupTestEnv("v1")()

	req := httptest.NewRequest(http.MethodGet, "/configs/mod-body-process?version=v1", nil)
	data, err := ExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}

