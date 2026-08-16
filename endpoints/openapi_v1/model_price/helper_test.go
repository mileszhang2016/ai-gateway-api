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

package model_price

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestIdFromURI(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		req := mux.SetURLVars(newTestRequest("/model-prices/42"), map[string]string{"id": "42"})
		id := idFromURI(req)
		assert.NotNil(t, id)
		assert.Equal(t, int64(42), *id)
	})

	t.Run("missing id", func(t *testing.T) {
		req := newTestRequest("/model-prices/")
		assert.Nil(t, idFromURI(req))
	})

	t.Run("invalid id", func(t *testing.T) {
		req := mux.SetURLVars(newTestRequest("/model-prices/abc"), map[string]string{"id": "abc"})
		assert.Nil(t, idFromURI(req))
	})
}

func TestQueryFilter(t *testing.T) {
	t.Run("all filters", func(t *testing.T) {
		req := newTestRequest("/model-prices?provider=openai&model=gpt-4&mode=chat")
		filter := queryFilter(req)
		assert.NotNil(t, filter)
		assert.Equal(t, "openai", *filter.Provider)
		assert.Equal(t, "gpt-4", *filter.Model)
		assert.Equal(t, "chat", *filter.Mode)
	})

	t.Run("no filters", func(t *testing.T) {
		req := newTestRequest("/model-prices")
		filter := queryFilter(req)
		assert.NotNil(t, filter)
		assert.Nil(t, filter.Provider)
		assert.Nil(t, filter.Model)
		assert.Nil(t, filter.Mode)
	})

	t.Run("empty values ignored", func(t *testing.T) {
		req := newTestRequest("/model-prices?provider=&model=&mode=")
		filter := queryFilter(req)
		assert.NotNil(t, filter)
		assert.Nil(t, filter.Provider)
		assert.Nil(t, filter.Model)
		assert.Nil(t, filter.Mode)
	})
}

func TestPageFilter(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		req := newTestRequest("/model-prices")
		page, pageSize := pageFilter(req)
		assert.Equal(t, 1, page)
		assert.Equal(t, 50, pageSize)
	})

	t.Run("custom values", func(t *testing.T) {
		req := newTestRequest("/model-prices?page=3&page_size=25")
		page, pageSize := pageFilter(req)
		assert.Equal(t, 3, page)
		assert.Equal(t, 25, pageSize)
	})

	t.Run("page size capped at 1000", func(t *testing.T) {
		req := newTestRequest("/model-prices?page=1&page_size=5000")
		page, pageSize := pageFilter(req)
		assert.Equal(t, 1, page)
		assert.Equal(t, 1000, pageSize)
	})

	t.Run("invalid values fall back to defaults", func(t *testing.T) {
		req := newTestRequest("/model-prices?page=abc&page_size=def")
		page, pageSize := pageFilter(req)
		assert.Equal(t, 1, page)
		assert.Equal(t, 50, pageSize)
	})
}

func TestEmptyString(t *testing.T) {
	assert.Nil(t, emptyString(""))
	assert.Equal(t, "x", *emptyString("x"))
}

func newTestRequest(rawURL string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, rawURL, nil)
	return req
}
