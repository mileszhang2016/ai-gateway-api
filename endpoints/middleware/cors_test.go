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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCors(t *testing.T) {
	c := NewCors()
	require.NotNil(t, c)

	handler := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))

	// Preflight request
	req := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, "*", rw.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, rw.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Equal(t, "true", rw.Header().Get("Access-Control-Allow-Credentials"))

	// Actual cross-origin request
	req = httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Origin", "https://example.com")
	rw = httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "hello", rw.Body.String())
	assert.Equal(t, "*", rw.Header().Get("Access-Control-Allow-Origin"))
}
