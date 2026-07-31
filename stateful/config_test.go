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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Init(t *testing.T) {
	cfg := &Config{
		Vars: map[string]string{
			"conf_dir": "/etc/ai-gateway",
		},
		Depends: DependsConfig{
			NavTreeFile: "${conf_dir}/nav_tree.toml",
			I18nDir:     "${conf_dir}/i18n",
		},
	}

	require.NoError(t, cfg.Init())

	assert.Equal(t, "/etc/ai-gateway/nav_tree.toml", cfg.Depends.NavTreeFile)
	assert.Equal(t, "/etc/ai-gateway/i18n", cfg.Depends.I18nDir)
}

func TestConfig_Init_NoVariables(t *testing.T) {
	cfg := &Config{
		Vars: map[string]string{},
		Depends: DependsConfig{
			NavTreeFile: "/absolute/nav_tree.toml",
			I18nDir:     "/absolute/i18n",
		},
	}

	require.NoError(t, cfg.Init())

	assert.Equal(t, "/absolute/nav_tree.toml", cfg.Depends.NavTreeFile)
	assert.Equal(t, "/absolute/i18n", cfg.Depends.I18nDir)
}
