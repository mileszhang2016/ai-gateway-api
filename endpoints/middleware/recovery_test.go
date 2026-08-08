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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecovery(t *testing.T) {
	rec := NewRecovery()
	require.NotNil(t, rec)
	assert.False(t, rec.StackAll)
	assert.Equal(t, 1024*8, rec.StackSize)
}

func TestRecovery_ServeHTTP_Normal(t *testing.T) {
	setupTestLoggers(t)

	rec := NewRecovery()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rw := httptest.NewRecorder()

	called := false
	rec.ServeHTTP(rw, req, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "ok", rw.Body.String())
}

func TestRecovery_ServeHTTP_Panic(t *testing.T) {
	setupTestLoggers(t)

	rec := NewRecovery()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rw := httptest.NewRecorder()

	rec.ServeHTTP(rw, req, func(w http.ResponseWriter, r *http.Request) {
		panic("something terrible")
	})

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "system error")
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
}

func TestMcConvert(t *testing.T) {
	setupTestLoggers(t)

	handler := NewRecovery()
	middleware := McConvert(handler)
	require.NotNil(t, middleware)

	router := mux.NewRouter()
	router.Use(middleware)
	router.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("router panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "system error")
}
