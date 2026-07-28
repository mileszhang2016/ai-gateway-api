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

package ibasic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBFEClusterManager(t *testing.T) {
	store := &fakeBFEClusterStorager{}
	m := NewBFEClusterManager(&fakeTxn{}, store)
	require.NotNil(t, m)
	assert.Equal(t, store, m.storager)
}

func TestBFEClusterManager_FetchBFEClusters(t *testing.T) {
	ctx := context.Background()
	expected := []*BFECluster{{ID: 1, Name: "c1"}}
	store := &fakeBFEClusterStorager{
		fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
			assert.Equal(t, ptrString("c1"), param.Name)
			return expected, nil
		},
	}
	m := NewBFEClusterManager(&fakeTxn{}, store)

	got, err := m.FetchBFEClusters(ctx, &BFEClusterFilter{Name: ptrString("c1")})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestBFEClusterManager_CreateBFECluster(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				assert.Equal(t, ptrString("c1"), param.Name)
				return nil, nil
			},
			createBFEClusterFn: func(ctx context.Context, param *BFEClusterParam) error {
				assert.Equal(t, ptrString("c1"), param.Name)
				return nil
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		require.NoError(t, m.CreateBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")}))
	})

	t.Run("existed", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				return []*BFECluster{{Name: "c1"}}, nil
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		err := m.CreateBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BFE Cluster Record Existed")
	})

	t.Run("fetch error", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		err := m.CreateBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}

func TestBFEClusterManager_DeleteBFECluster(t *testing.T) {
	ctx := context.Background()

	t.Run("fetch error", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		err := m.DeleteBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				return nil, nil
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		err := m.DeleteBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BFE Cluster Record Not Exist")
	})

	t.Run("success", func(t *testing.T) {
		store := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
				return []*BFECluster{{ID: 2, Name: "c1"}}, nil
			},
			deleteBFEClusterFn: func(ctx context.Context, cluster *BFECluster) error {
				assert.Equal(t, int64(2), cluster.ID)
				return nil
			},
		}
		m := NewBFEClusterManager(&fakeTxn{}, store)
		require.NoError(t, m.DeleteBFECluster(ctx, &BFEClusterParam{Name: ptrString("c1")}))
	})
}

func TestBFEClusterID2NameMap(t *testing.T) {
	list := []*BFECluster{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	got := BFEClusterID2NameMap(list)
	assert.Equal(t, map[int64]string{1: "a", 2: "b"}, got)
}

func TestBFEClusterIDMap(t *testing.T) {
	c1 := &BFECluster{ID: 1, Name: "a"}
	c2 := &BFECluster{ID: 2, Name: "b"}
	got := BFEClusterIDMap([]*BFECluster{c1, c2})
	assert.Equal(t, map[int64]*BFECluster{1: c1, 2: c2}, got)
}

func TestBFEClusterNameMap(t *testing.T) {
	c1 := &BFECluster{ID: 1, Name: "a"}
	c2 := &BFECluster{ID: 2, Name: "b"}
	got := BFEClusterNameMap([]*BFECluster{c1, c2})
	assert.Equal(t, map[string]*BFECluster{"a": c1, "b": c2}, got)
}
