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

package icluster_conf

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
)

func TestSubClusterList2MapByName(t *testing.T) {
	s1 := &SubCluster{ID: 1, Name: "a"}
	s2 := &SubCluster{ID: 2, Name: "b"}
	got := SubClusterList2MapByName([]*SubCluster{s1, s2})
	assert.Equal(t, map[string]*SubCluster{"a": s1, "b": s2}, got)
}

func TestSubClusterList2MapByID(t *testing.T) {
	s1 := &SubCluster{ID: 1, Name: "a"}
	s2 := &SubCluster{ID: 2, Name: "b"}
	got := SubClusterList2MapByID([]*SubCluster{s1, s2})
	assert.Equal(t, map[int64]*SubCluster{1: s1, 2: s2}, got)
}

func TestSubClusterList2IDSlice(t *testing.T) {
	got := SubClusterList2IDSlice([]*SubCluster{{ID: 1}, {ID: 2}})
	assert.Equal(t, []int64{1, 2}, got)
}

func TestSubClusterList2NameSlice(t *testing.T) {
	got := SubClusterList2NameSlice([]*SubCluster{{Name: "a"}, {Name: "b"}})
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestNewSubClusterManager(t *testing.T) {
	m := NewSubClusterManager(&fakeTxn{}, &fakeSubClusterStorager{}, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
	require.NotNil(t, m)
}

func TestSubClusterManager_SubClusterList(t *testing.T) {
	ctx := context.Background()
	expected := []*SubCluster{{ID: 1, Name: "sc1"}}
	store := &fakeSubClusterStorager{
		fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
			assert.Equal(t, []string{"sc1"}, param.Names)
			return expected, nil
		},
	}
	m := NewSubClusterManager(&fakeTxn{}, store, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
	got, err := m.SubClusterList(ctx, &SubClusterFilter{Names: []string{"sc1"}})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestSubClusterManager_FetchSubCluster(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		store := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{ID: 1, Name: "sc1"}}, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, store, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		got, err := m.FetchSubCluster(ctx, &SubClusterFilter{Name: lib.PString("sc1")})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(1), got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return nil, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, store, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		got, err := m.FetchSubCluster(ctx, &SubClusterFilter{})
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestSubClusterManager_CreateSubCluster(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("success", func(t *testing.T) {
		created := false
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return nil, nil
			},
			createSubClusterFn: func(ctx context.Context, param *SubClusterParam) error {
				created = true
				assert.Equal(t, "sc1", *param.Name)
				assert.Equal(t, "test.p1", *param.PoolName)
				return nil
			},
		}
		poolStore := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return &Pool{ID: 1, Name: "test.p1", Product: product}, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, subClusterStore, &fakeProductStorager{}, poolStore, &fakeClusterStorager{})
		err := m.CreateSubCluster(ctx, product, &SubClusterParam{
			Name:     lib.PString("sc1"),
			PoolName: lib.PString("test.p1"),
		})
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("already existed", func(t *testing.T) {
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{Name: "sc1"}}, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, subClusterStore, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		err := m.CreateSubCluster(ctx, product, &SubClusterParam{
			Name:     lib.PString("sc1"),
			PoolName: lib.PString("test.p1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SubCluster Record Existed")
	})

	t.Run("pool not exist", func(t *testing.T) {
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return nil, nil
			},
		}
		poolStore := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return nil, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, subClusterStore, &fakeProductStorager{}, poolStore, &fakeClusterStorager{})
		err := m.CreateSubCluster(ctx, product, &SubClusterParam{
			Name:     lib.PString("sc1"),
			PoolName: lib.PString("test.p1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Not Exist")
	})

	t.Run("pool not valid product", func(t *testing.T) {
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return nil, nil
			},
		}
		poolStore := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return &Pool{ID: 1, Name: "test.p1", Product: &ibasic.Product{ID: 3, Name: "other"}}, nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, subClusterStore, &fakeProductStorager{}, poolStore, &fakeClusterStorager{})
		err := m.CreateSubCluster(ctx, product, &SubClusterParam{
			Name:     lib.PString("sc1"),
			PoolName: lib.PString("test.p1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Not Valid")
	})

	t.Run("fetch error", func(t *testing.T) {
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, subClusterStore, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		err := m.CreateSubCluster(ctx, product, &SubClusterParam{
			Name:     lib.PString("sc1"),
			PoolName: lib.PString("test.p1"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}

func TestSubClusterManager_DeleteSubCluster(t *testing.T) {
	ctx := context.Background()

	t.Run("mounted", func(t *testing.T) {
		m := NewSubClusterManager(&fakeTxn{}, &fakeSubClusterStorager{}, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		err := m.DeleteSubCluster(ctx, &SubCluster{ID: 1, Name: "sc1", ClusterID: 10})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "be Mounted With Cluster 10")
	})

	t.Run("success", func(t *testing.T) {
		deleted := false
		store := &fakeSubClusterStorager{
			deleteSubClusterFn: func(ctx context.Context, param *SubCluster) error {
				deleted = true
				return nil
			},
		}
		m := NewSubClusterManager(&fakeTxn{}, store, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
		err := m.DeleteSubCluster(ctx, &SubCluster{ID: 1, Name: "sc1"})
		require.NoError(t, err)
		assert.True(t, deleted)
	})
}

func TestSubClusterManager_UpdateSubCluster(t *testing.T) {
	ctx := context.Background()
	updated := false
	store := &fakeSubClusterStorager{
		updateSubClusterFn: func(ctx context.Context, one *SubCluster, param *SubClusterParam) error {
			updated = true
			assert.Equal(t, "sc1", one.Name)
			return nil
		},
	}
	m := NewSubClusterManager(&fakeTxn{}, store, &fakeProductStorager{}, &fakePoolStorager{}, &fakeClusterStorager{})
	err := m.UpdateSubCluster(ctx, &SubCluster{ID: 1, Name: "sc1"}, &SubClusterParam{})
	require.NoError(t, err)
	assert.True(t, updated)
}
