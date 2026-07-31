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

package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert_Success(t *testing.T) {
	called := false
	handler := func(req *http.Request) (*http.Request, error) {
		called = true
		return req, nil
	}

	mw := convert(handler)
	require.NotNil(t, mw)

	router := mux.NewRouter()
	router.Use(mw)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestConvert_Error(t *testing.T) {
	handler := func(req *http.Request) (*http.Request, error) {
		return nil, errors.New("convert failed")
	}

	mw := convert(handler)
	require.NotNil(t, mw)

	router := mux.NewRouter()
	router.Use(mw)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "convert failed")
}

func TestMiddlewareVariables(t *testing.T) {
	assert.NotNil(t, McProductProbe)
	assert.NotNil(t, McUserProbe)
}
