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

package ibasic

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtraFilePath(t *testing.T) {
	p := &Product{Name: "MyProduct"}
	got := ExtraFilePath("tls_conf", p, filepath.FromSlash("a/b.pem"))
	assert.Equal(t, "tls_conf/myproduct/a_b.pem", got)
}

func TestNewExtraFileManager(t *testing.T) {
	store := &fakeExtraFileStorager{}
	m := NewExtraFileManager(store)
	require.NotNil(t, m)
	assert.Equal(t, store, m.storager)
}

func TestExtraFileManager_FetchExtraFile(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		expected := &ExtraFile{ID: 1, Name: "f1"}
		store := &fakeExtraFileStorager{
			fetchExtraFilesFn: func(ctx context.Context, filter *ExtraFileFilter) ([]*ExtraFile, error) {
				require.NotNil(t, filter.Name)
				assert.Equal(t, "f1", *filter.Name)
				return []*ExtraFile{expected}, nil
			},
		}
		m := NewExtraFileManager(store)
		got, err := m.FetchExtraFile(ctx, "f1")
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeExtraFileStorager{
			fetchExtraFilesFn: func(ctx context.Context, filter *ExtraFileFilter) ([]*ExtraFile, error) {
				return nil, nil
			},
		}
		m := NewExtraFileManager(store)
		got, err := m.FetchExtraFile(ctx, "f1")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("error", func(t *testing.T) {
		store := &fakeExtraFileStorager{
			fetchExtraFilesFn: func(ctx context.Context, filter *ExtraFileFilter) ([]*ExtraFile, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewExtraFileManager(store)
		_, err := m.FetchExtraFile(ctx, "f1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}
