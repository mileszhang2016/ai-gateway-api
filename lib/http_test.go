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

package lib

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "token", r.Header.Get("X-Token"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	data, err := Read(server.URL, 5, map[string]string{"X-Token": "token"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(data))
}

func TestReadNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := Read(server.URL, 5, nil, nil)
	assert.Error(t, err)
}

func TestReadWithRetrySuccessOnFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data"))
	}))
	defer server.Close()

	data, err := ReadWithRetry(server.URL, 5, nil, 3, 1, nil)
	require.NoError(t, err)
	assert.Equal(t, "data", string(data))
}

func TestReadWithRetryFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := ReadWithRetry(server.URL, 5, nil, 2, 1, nil)
	assert.Error(t, err)
}

func TestSendPatchRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("patched"))
	}))
	defer server.Close()

	status, body, err := SendPatchRequest(server.URL, map[string]string{"key": "value"}, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "patched", string(body))
}

func TestSendDeleteRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	status, _, err := SendDeleteRequest(server.URL, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, status)
}

func TestSendFiles(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "upload.txt")
	require.NoError(t, os.WriteFile(file, []byte("hello"), 0644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseMultipartForm(1024))
		f, _, err := r.FormFile("file")
		require.NoError(t, err)
		defer f.Close()

		buf := make([]byte, 5)
		_, err = f.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(buf))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("uploaded"))
	}))
	defer server.Close()

	status, body, err := SendFiles([]string{file}, server.URL, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "uploaded", string(body))
}

func TestRecover(t *testing.T) {
	func() {
		defer Recover("test")
		panic(errors.New("intentional panic"))
	}()
	// If we reach here, recover worked.
	assert.True(t, true)
}

// errors.New is used to make the panic an error value.
