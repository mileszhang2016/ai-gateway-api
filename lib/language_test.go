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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccepLanguages(t *testing.T) {
	t.Run("no header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		assert.Equal(t, []string{""}, AccepLanguages(req))
	})

	t.Run("single language", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		req.Header.Set("Accept-Language", "zh-CN")
		assert.Equal(t, []string{"zh-CN"}, AccepLanguages(req))
	})

	t.Run("multiple languages with q values", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		req.Header.Set("Accept-Language", "zh-CN,en-US;q=0.9,en;q=0.8")
		assert.Equal(t, []string{"zh-CN", "en-US", "en"}, AccepLanguages(req))
	})
}
