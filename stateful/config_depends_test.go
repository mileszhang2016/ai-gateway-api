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

package stateful

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNavTree_matchRole(t *testing.T) {
	tests := []struct {
		name     string
		tree     *NavTree
		role     string
		expected bool
	}{
		{
			name:     "nil allow roles matches all",
			tree:     &NavTree{ID: "root"},
			role:     "admin",
			expected: true,
		},
		{
			name:     "role in allow list",
			tree:     &NavTree{ID: "root", AllowRoles: []string{"admin", "user"}},
			role:     "admin",
			expected: true,
		},
		{
			name:     "role not in allow list",
			tree:     &NavTree{ID: "root", AllowRoles: []string{"admin"}},
			role:     "user",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tree.matchRole(tt.role))
		})
	}
}

func TestNavTree_deriveByRole(t *testing.T) {
	root := &NavTree{
		ID:         "root",
		Text:       "Root",
		AllowRoles: []string{"admin", "user"},
		Children: []*NavTree{
			{
				ID:         "admin-only",
				Text:       "Admin Only",
				AllowRoles: []string{"admin"},
			},
			{
				ID:   "public",
				Text: "Public",
				Children: []*NavTree{
					{
						ID:         "user-child",
						Text:       "User Child",
						AllowRoles: []string{"user"},
					},
				},
			},
		},
	}

	t.Run("admin role", func(t *testing.T) {
		derived, err := root.deriveByRole("admin")
		require.NoError(t, err)
		require.NotNil(t, derived)
		assert.Equal(t, "root", derived.ID)
		// "public" node is pruned because its only child "user-child" is not
		// visible to the admin role and "public" itself has no role restriction.
		require.Len(t, derived.Children, 1)
		assert.Equal(t, "admin-only", derived.Children[0].ID)
	})

	t.Run("user role", func(t *testing.T) {
		derived, err := root.deriveByRole("user")
		require.NoError(t, err)
		require.NotNil(t, derived)
		assert.Equal(t, "root", derived.ID)
		require.Len(t, derived.Children, 1)
		assert.Equal(t, "public", derived.Children[0].ID)
		require.Len(t, derived.Children[0].Children, 1)
		assert.Equal(t, "user-child", derived.Children[0].Children[0].ID)
	})

	t.Run("unauthorized role", func(t *testing.T) {
		derived, err := root.deriveByRole("guest")
		require.NoError(t, err)
		assert.Nil(t, derived)
	})
}
