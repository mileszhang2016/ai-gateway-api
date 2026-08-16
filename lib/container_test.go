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

package lib

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64BoolMap2Slice(t *testing.T) {
	m := map[int64]bool{1: true, 2: false, 3: true}
	got := Int64BoolMap2Slice(m)
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	assert.Equal(t, []int64{1, 3}, got)
}

func TestStringBoolMap2Slice(t *testing.T) {
	m := map[string]bool{"a": true, "b": false, "c": true}
	got := StringBoolMap2Slice(m)
	sort.Strings(got)
	assert.Equal(t, []string{"a", "c"}, got)
}

func TestStringMap2Slice(t *testing.T) {
	m := map[string]bool{"x": true, "y": false}
	got := StringMap2Slice(m)
	sort.Strings(got)
	assert.Equal(t, []string{"x", "y"}, got)
}

func TestStringSlice2Map(t *testing.T) {
	assert.Equal(t, map[string]bool{"a": true, "b": true}, StringSlice2Map([]string{"a", "b", "a"}))
}

func TestInt64Map2Slice(t *testing.T) {
	m := map[int64]bool{10: true, 5: false}
	got := Int64Map2Slice(m)
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	assert.Equal(t, []int64{5, 10}, got)
}

func TestSortMapInt642String(t *testing.T) {
	m := map[int64]string{3: "c", 1: "a", 2: "b"}
	assert.Equal(t, []int64{1, 2, 3}, SortMapInt642String(m))
}

func TestStringSliceHasElement(t *testing.T) {
	assert.True(t, StringSliceHasElement([]string{"a", "b"}, "b"))
	assert.False(t, StringSliceHasElement([]string{"a", "b"}, "c"))
}

func TestStringSliceSubtract(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, StringSliceSubtract([]string{"a", "b", "c"}, []string{"b", "d"}))
}

func TestStringSliceSub(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, StringSliceSub([]string{"a", "b", "c"}, []string{"b", "d"}))
}
