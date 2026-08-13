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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "conf.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"Name":"test","Value":42}`), 0644))

	var data struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	require.NoError(t, LoadConf(file, &data, ".json"))
	assert.Equal(t, "test", data.Name)
	assert.Equal(t, 42, data.Value)
}

func TestLoadConfAuto(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "conf.toml")
	require.NoError(t, os.WriteFile(file, []byte("name = \"test\"\nvalue = 42\n"), 0644))

	var data struct {
		Name  string `toml:"name"`
		Value int    `toml:"value"`
	}
	require.NoError(t, LoadConfAuto(file, &data))
	assert.Equal(t, "test", data.Name)
	assert.Equal(t, 42, data.Value)
}

func TestLoadConfUnsupportedSuffix(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "conf.xyz")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0644))

	var data struct{}
	err := LoadConf(file, &data, ".xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported fileSuffix")
}

func TestLoadConfFileNotFound(t *testing.T) {
	var data struct{}
	err := LoadConf("/not/exist/conf.json", &data, ".json")
	assert.Error(t, err)
}
