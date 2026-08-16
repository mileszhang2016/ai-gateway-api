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

package icluster_conf

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
)

func TestInstance_IPWithPort(t *testing.T) {
	t.Run("port set", func(t *testing.T) {
		i := &Instance{Addr: "127.0.0.1", Port: 8080}
		assert.Equal(t, "127.0.0.1:8080", i.IPWithPort())
	})
}

func TestInstance_UnmarshalJSON(t *testing.T) {
	t.Run("new format", func(t *testing.T) {
		var i Instance
		require.NoError(t, json.Unmarshal([]byte(`{"name":"rs1","addr":"127.0.0.1","port":8080,"weight":10}`), &i))
		assert.Equal(t, Instance{Name: "rs1", Addr: "127.0.0.1", Port: 8080, Weight: 10}, i)
	})

	t.Run("legacy format", func(t *testing.T) {
		var i Instance
		require.NoError(t, json.Unmarshal([]byte(`{"Name":"rs1","Addr":"127.0.0.1","Port":8080,"Weight":10}`), &i))
		assert.Equal(t, Instance{Name: "rs1", Addr: "127.0.0.1", Port: 8080, Weight: 10}, i)
	})

	t.Run("legacy format with ports map", func(t *testing.T) {
		var i Instance
		require.NoError(t, json.Unmarshal([]byte(`{"Name":"rs1","Addr":"127.0.0.1","Ports":{"Default":9090},"Weight":10}`), &i))
		assert.Equal(t, Instance{Name: "rs1", Addr: "127.0.0.1", Port: 9090, Weight: 10}, i)
	})
}

func TestNewPoolManager(t *testing.T) {
	m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	require.NotNil(t, m)
}

func TestPoolManager_FetchPoolByName(t *testing.T) {
	ctx := context.Background()
	expected := &Pool{ID: 1, Name: "p1"}
	store := &fakePoolStorager{
		fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
			assert.Equal(t, "p1", name)
			return expected, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.FetchPoolByName(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestPoolManager_FetchBFEPool(t *testing.T) {
	ctx := context.Background()
	expected := &Pool{ID: 1, Name: "BFE.p1"}
	store := &fakePoolStorager{
		fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
			assert.Equal(t, "BFE.p1", name)
			return expected, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.FetchBFEPool(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestPoolManager_FetchProductPool(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}
	expected := &Pool{ID: 1, Name: "test.p1"}
	store := &fakePoolStorager{
		fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
			assert.Equal(t, "test.p1", name)
			return expected, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.FetchProductPool(ctx, product, "p1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestPoolManager_FetchBFEPools(t *testing.T) {
	ctx := context.Background()
	expected := []*Pool{{ID: 1, Name: "BFE.p1"}}
	store := &fakePoolStorager{
		fetchPoolsFn: func(ctx context.Context, param *PoolFilter) ([]*Pool, error) {
			assert.Equal(t, int64(1), *param.ProductID)
			return expected, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.FetchBFEPools(ctx)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestPoolManager_FetchProductPools(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}
	expected := []*Pool{{ID: 1, Name: "test.p1"}}
	store := &fakePoolStorager{
		fetchPoolsFn: func(ctx context.Context, param *PoolFilter) ([]*Pool, error) {
			assert.Equal(t, int64(2), *param.ProductID)
			return expected, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.FetchProductPools(ctx, product)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func Test_poolNameJudger(t *testing.T) {
	t.Run("already prefixed", func(t *testing.T) {
		got, err := poolNameJudger("test", "test.p1")
		require.NoError(t, err)
		assert.Equal(t, "test.p1", got)
	})

	t.Run("wrong prefix", func(t *testing.T) {
		_, err := poolNameJudger("test", "other.p1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Name Must Use Product Name as Prefix")
	})

	t.Run("add prefix", func(t *testing.T) {
		got, err := poolNameJudger("test", "p1")
		require.NoError(t, err)
		assert.Equal(t, "test.p1", got)
	})
}

func TestPoolManager_CanDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("referred by bfe cluster", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return []*ibasic.BFECluster{{Name: "bfe1"}}, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, bfeClusterStore, &fakeSubClusterStorager{})
		err := m.CanDelete(ctx, &Pool{Name: "p1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "BFECluster bfe1 Refer To This Pool")
	})

	t.Run("referred by sub cluster", func(t *testing.T) {
		bfeClusterStore := &fakeBFEClusterStorager{
			fetchBFEClustersFn: func(ctx context.Context, param *ibasic.BFEClusterFilter) ([]*ibasic.BFECluster, error) {
				return nil, nil
			},
		}
		subClusterStore := &fakeSubClusterStorager{
			fetchSubClusterListFn: func(ctx context.Context, param *SubClusterFilter) ([]*SubCluster, error) {
				return []*SubCluster{{Name: "sc1"}}, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, bfeClusterStore, subClusterStore)
		err := m.CanDelete(ctx, &Pool{Name: "p1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SubCluster sc1 Refer To This Pool")
	})

	t.Run("can delete", func(t *testing.T) {
		m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		require.NoError(t, m.CanDelete(ctx, &Pool{Name: "p1"}))
	})
}

func TestPoolManager_DeleteBFEPool(t *testing.T) {
	ctx := context.Background()
	store := &fakePoolStorager{
		fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
			return &Pool{ID: 1, Name: "BFE.p1"}, nil
		},
		deletePoolFn: func(ctx context.Context, pool *Pool) error {
			return nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.DeleteBFEPool(ctx, "p1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "BFE.p1", got.Name)
}

func TestPoolManager_DeleteProductPool(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("not found", func(t *testing.T) {
		store := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return nil, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		_, err := m.DeleteProductPool(ctx, product, "p1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Record Not Exist")
	})

	t.Run("fetch error", func(t *testing.T) {
		store := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		_, err := m.DeleteProductPool(ctx, product, "p1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}

func TestPoolManager_CreateBFEPool(t *testing.T) {
	ctx := context.Background()
	store := &fakePoolStorager{
		fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
			return nil, nil
		},
		createPoolFn: func(ctx context.Context, product *ibasic.Product, data *PoolParam) (*Pool, error) {
			assert.Equal(t, int64(1), product.ID)
			assert.Equal(t, "BFE.p1", *data.Name)
			assert.Equal(t, PoolTagBFE, *data.Tag)
			return &Pool{ID: 1, Name: "BFE.p1"}, nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	got, err := m.CreateBFEPool(ctx, &PoolParam{Name: lib.PString("p1")})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "BFE.p1", got.Name)
}

func TestPoolManager_CreateProductPool(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}

	t.Run("success", func(t *testing.T) {
		store := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return nil, nil
			},
			createPoolFn: func(ctx context.Context, product *ibasic.Product, data *PoolParam) (*Pool, error) {
				assert.Equal(t, "test.p1", *data.Name)
				assert.Equal(t, PoolTagProduct, *data.Tag)
				return &Pool{ID: 1, Name: "test.p1"}, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		got, err := m.CreateProductPool(ctx, product, &PoolParam{Name: lib.PString("p1")})
		require.NoError(t, err)
		assert.Equal(t, "test.p1", got.Name)
	})

	t.Run("existed", func(t *testing.T) {
		store := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return &Pool{Name: "test.p1"}, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		_, err := m.CreateProductPool(ctx, product, &PoolParam{Name: lib.PString("p1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Record Existed")
	})
}

func TestPoolManager_UpdateProductPool(t *testing.T) {
	ctx := context.Background()
	product := &ibasic.Product{ID: 2, Name: "test"}
	updated := false
	store := &fakePoolStorager{
		updatePoolFn: func(ctx context.Context, oldData *Pool, diff *PoolParam) error {
			updated = true
			return nil
		},
	}
	m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
	err := m.UpdateProductPool(ctx, product, &Pool{ID: 1, Name: "test.p1"}, &PoolParam{})
	require.NoError(t, err)
	assert.True(t, updated)
}

func TestPoolList2Map(t *testing.T) {
	p1 := &Pool{ID: 1, Name: "a"}
	p2 := &Pool{ID: 2, Name: "b"}
	got := PoolList2Map([]*Pool{p1, p2})
	assert.Equal(t, map[int64]*Pool{1: p1, 2: p2}, got)
}

func TestPoolManager_GetPoolByName(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := &Pool{ID: 1, Name: "p1"}
		store := &fakePoolStorager{
			fetchPoolFn: func(ctx context.Context, name string) (*Pool, error) {
				return expected, nil
			},
		}
		m := NewPoolManager(&fakeTxn{}, store, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		got, err := m.GetPoolByName(ctx, lib.PString("p1"))
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("nil name", func(t *testing.T) {
		m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		_, err := m.GetPoolByName(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Name Illegal")
	})

	t.Run("empty name", func(t *testing.T) {
		m := NewPoolManager(&fakeTxn{}, &fakePoolStorager{}, &fakeBFEClusterStorager{}, &fakeSubClusterStorager{})
		_, err := m.GetPoolByName(ctx, lib.PString(""))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pool Name Illegal")
	})
}
