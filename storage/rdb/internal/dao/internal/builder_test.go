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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRecord struct {
	ID     int64   `db:"id"`
	Name   string  `db:"name"`
	Status *int    `db:"status"`
	Ignore string  `db:"-"`
	Age    int     `db:"age,>"`
	Tags   []string `db:"tags"`
}

func TestStruct2Where(t *testing.T) {
	status := 1
	r := &testRecord{
		ID:     10,
		Name:   "foo",
		Status: &status,
		Ignore: "ignored",
		Age:    18,
		Tags:   []string{"a", "b"},
	}

	m := Struct2Where(r)
	assert.Equal(t, int64(10), m["id"])
	assert.Equal(t, "foo", m["name"])
	assert.Equal(t, 1, m["status"])
	assert.Equal(t, 18, m["age >"])
	assert.Equal(t, []string{"a", "b"}, m["tags"])
	assert.NotContains(t, m, "ignore")
	assert.NotContains(t, m, "-")
}

func TestStruct2Where_ZeroValue(t *testing.T) {
	r := &testRecord{}
	m := Struct2Where(r)
	// Zero-value primitive fields are still included unless they are pointers.
	assert.Equal(t, int64(0), m["id"])
	assert.Equal(t, "", m["name"])
	assert.Equal(t, 0, m["age >"])
	// Nil pointer and empty slice are omitted.
	assert.NotContains(t, m, "status")
	assert.NotContains(t, m, "tags")
}

func TestStruct2Assign(t *testing.T) {
	status := 2
	r := &testRecord{
		ID:     20,
		Name:   "bar",
		Status: &status,
		Age:    21,
	}

	m := Struct2Assign(r)
	assert.Equal(t, int64(20), m["id"])
	assert.Equal(t, "bar", m["name"])
	assert.Equal(t, 2, m["status"])
	// Options are ignored for assignments.
	assert.Equal(t, 21, m["age"])
}

func TestStruct2AssignList(t *testing.T) {
	rs := []testRecord{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
	}
	ms := Struct2AssignList(rs[0], rs[1])
	require.Len(t, ms, 2)
	assert.Equal(t, int64(1), ms[0]["id"])
	assert.Equal(t, "a", ms[0]["name"])
	assert.Equal(t, int64(2), ms[1]["id"])
	assert.Equal(t, "b", ms[1]["name"])
}

func TestTagSplitter(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		wantOpt string
	}{
		{"id", "id", ""},
		{"age,>", "age", ">"},
		{"  name  ,  like  ", "name", "like"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, opt := tagSplitter(tt.input)
			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantOpt, opt)
		})
	}
}

func TestSelectBuilder_Compile(t *testing.T) {
	b := NewSelectBuilder("users", map[string]interface{}{"id": 1}, []string{"id", "name"})
	sql, args, err := b.Compile()
	require.NoError(t, err)
	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "FROM users")
	assert.Contains(t, sql, "WHERE")
	assert.Len(t, args, 1)
}

func TestInsertBuilder_Compile(t *testing.T) {
	b := NewInsertBuilder("users", []map[string]interface{}{
		{"id": 1, "name": "foo"},
	})
	sql, args, err := b.Compile()
	require.NoError(t, err)
	assert.Contains(t, sql, "INSERT INTO users")
	assert.Len(t, args, 2)
}

func TestUpdateBuilder_Compile(t *testing.T) {
	b := NewUpdateBuilder("users", map[string]interface{}{"id": 1}, map[string]interface{}{"name": "foo"})
	sql, args, err := b.Compile()
	require.NoError(t, err)
	assert.Contains(t, sql, "UPDATE users")
	assert.Contains(t, sql, "SET")
	assert.Contains(t, sql, "WHERE")
	assert.Len(t, args, 2)
}

func TestDeleteBuilder_Compile(t *testing.T) {
	b := NewDeleteBuilder("users", map[string]interface{}{"id": 1})
	sql, args, err := b.Compile()
	require.NoError(t, err)
	assert.Contains(t, sql, "DELETE FROM users")
	assert.Contains(t, sql, "WHERE")
	assert.Len(t, args, 1)
}
