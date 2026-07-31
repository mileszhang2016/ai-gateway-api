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

package export_util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExportFromReq(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/export?version=v1", nil)
	param, err := NewExportFromReq(req)

	require.NoError(t, err)
	require.NotNil(t, param)
	assert.Equal(t, "v1", param.Version)
}

func TestNewExportFromReq_EmptyVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/export", nil)
	param, err := NewExportFromReq(req)

	require.NoError(t, err)
	require.NotNil(t, param)
	assert.Equal(t, "", param.Version)
}
