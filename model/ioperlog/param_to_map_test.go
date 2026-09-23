// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package ioperlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type paramToMapFixture struct {
	Name    *string `json:"name"`
	Enable  *bool   `json:"enabled"`
	Count   *int    `json:"count"`
	Label   *string `json:"label"`
	Ignored *string `json:"-"`
}

func TestParamToMap_NilReturnsNil(t *testing.T) {
	assert.Nil(t, ParamToMap(nil))
	var nilParam *paramToMapFixture
	assert.Nil(t, ParamToMap(nilParam))
}

func TestParamToMap_DropsNilValuedKeys(t *testing.T) {
	name := "n1"
	m := ParamToMap(&paramToMapFixture{Name: &name})

	require.NotNil(t, m)
	assert.Equal(t, map[string]interface{}{"name": "n1"}, m)
}

func TestParamToMap_KeepsExplicitZeroValues(t *testing.T) {
	name := "n1"
	enable := false
	count := 0
	label := ""
	m := ParamToMap(&paramToMapFixture{
		Name:   &name,
		Enable: &enable,
		Count:  &count,
		Label:  &label,
	})

	// Explicit zero values are submitted fields and must survive, so that
	// they are still recorded in diff_keys as real changes.
	require.NotNil(t, m)
	assert.Equal(t, map[string]interface{}{
		"name":    "n1",
		"enabled": false,
		"count":   float64(0),
		"label":   "",
	}, m)
}

func TestParamToMap_MarshalErrorReturnsNil(t *testing.T) {
	type badParam struct {
		Ch chan int `json:"ch"`
	}
	assert.Nil(t, ParamToMap(&badParam{Ch: make(chan int)}))
}

// TestParamToMap_PartialUpdateDiffKeys guards issue #201 end to end at the
// pipeline level: a param with only some fields set must produce an after
// map (and therefore diff_keys) containing exactly the submitted fields.
func TestParamToMap_PartialUpdateDiffKeys(t *testing.T) {
	name := "n1"
	before := map[string]interface{}{
		"count":   float64(3),
		"enabled": true,
		"name":    "n0",
	}
	after := ParamToMap(&paramToMapFixture{Name: &name})

	summary := BuildChangeSummary(string(ActionUpdate), before, after)
	require.NotNil(t, summary)
	assert.Equal(t, []string{"name"}, summary["diff_keys"])
	assert.NotContains(t, summary["after"].(map[string]interface{}), "enabled")
}
