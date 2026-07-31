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

package stateful

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLangMapping(t *testing.T) {
	t.Run("valid regex without placeholder", func(t *testing.T) {
		mm, err := NewLangMapping(`^not found$`, "未找到")
		require.NoError(t, err)
		assert.NotNil(t, mm.fromRegex)
		assert.Equal(t, 0, mm.placeholderCount)
	})

	t.Run("valid regex with placeholder", func(t *testing.T) {
		mm, err := NewLangMapping(`invalid value: (.*)$`, "无效值: %s")
		require.NoError(t, err)
		assert.Equal(t, 1, mm.placeholderCount)
	})

	t.Run("invalid regex", func(t *testing.T) {
		mm, err := NewLangMapping(`[`, "x")
		require.Error(t, err)
		assert.Nil(t, mm)
	})
}

func TestLangMapping_TryTrans(t *testing.T) {
	t.Run("no regex initialized", func(t *testing.T) {
		mm := &LangMapping{From: "x", To: "y"}
		got, ok := mm.TryTrans("x")
		assert.False(t, ok)
		assert.Equal(t, "x", got)
	})

	t.Run("match without placeholder", func(t *testing.T) {
		mm, err := NewLangMapping(`^not found$`, "未找到")
		require.NoError(t, err)
		got, ok := mm.TryTrans("not found")
		assert.True(t, ok)
		assert.Equal(t, "未找到", got)
	})

	t.Run("match with placeholder", func(t *testing.T) {
		mm, err := NewLangMapping(`invalid value: (.*)$`, "无效值: %s")
		require.NoError(t, err)
		got, ok := mm.TryTrans("invalid value: foo")
		assert.True(t, ok)
		assert.Equal(t, "无效值: foo", got)
	})

	t.Run("no match", func(t *testing.T) {
		mm, err := NewLangMapping(`^not found$`, "未找到")
		require.NoError(t, err)
		got, ok := mm.TryTrans("error")
		assert.False(t, ok)
		assert.Equal(t, "error", got)
	})
}

func TestTryMappingErrMsg(t *testing.T) {
	// Backup and restore global language packs.
	oldLangs := langs2languagePacks
	defer func() { langs2languagePacks = oldLangs }()

	langs2languagePacks = map[string]map[string]*LangMapping{
		"zh": {
			"ParamError": {
				From: "ParamError",
				To:   "参数错误",
			},
			`^value (.*) is invalid$`: func() *LangMapping {
				mm, _ := NewLangMapping(`^value (.*) is invalid$`, "值 %s 无效")
				return mm
			}(),
		},
	}

	t.Run("empty message", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", TryMappingErrMsg(req, ""))
	})

	t.Run("no matching language pack", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", "fr")
		assert.Equal(t, "ParamError: value foo is invalid", TryMappingErrMsg(req, "ParamError: value foo is invalid"))
	})

	t.Run("translate type and message", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", "zh")
		assert.Equal(t, "参数错误: 值 foo 无效", TryMappingErrMsg(req, "ParamError: value foo is invalid"))
	})
}
