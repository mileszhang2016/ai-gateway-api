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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string { return &s }
func ptrInt64(i int64) *int64    { return &i }

func TestNewProductContext_MustGetProduct(t *testing.T) {
	ctx := context.Background()
	p := &Product{ID: 1, Name: "test"}

	t.Run("success", func(t *testing.T) {
		ctxWithProduct := NewProductContext(ctx, p)
		got, err := MustGetProduct(ctxWithProduct)
		require.NoError(t, err)
		assert.Equal(t, p, got)
	})

	t.Run("missing", func(t *testing.T) {
		_, err := MustGetProduct(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Fail To Get Product")
	})
}

func TestNewProductManager(t *testing.T) {
	store := &fakeProductStorager{}
	m := NewProductManager(&fakeTxn{}, store)
	require.NotNil(t, m)
	assert.Equal(t, store, m.storager)
}

func TestProductManager_FetchProducts(t *testing.T) {
	ctx := context.Background()
	expected := []*Product{{ID: 1, Name: "p1"}}
	store := &fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ProductFilter) ([]*Product, error) {
			assert.Equal(t, ptrString("p1"), param.Name)
			return expected, nil
		},
	}
	m := NewProductManager(&fakeTxn{}, store)

	got, err := m.FetchProducts(ctx, &ProductFilter{Name: ptrString("p1")})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestProductManager_DeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("build-in product", func(t *testing.T) {
		m := NewProductManager(&fakeTxn{}, &fakeProductStorager{})
		err := m.DeleteProduct(ctx, &Product{ID: 1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cant Delete Build-in Product")
	})

	t.Run("success", func(t *testing.T) {
		called := false
		store := &fakeProductStorager{
			deleteProductFn: func(ctx context.Context, p *Product) error {
				called = true
				assert.Equal(t, int64(2), p.ID)
				return nil
			},
		}
		m := NewProductManager(&fakeTxn{}, store)
		require.NoError(t, m.DeleteProduct(ctx, &Product{ID: 2}))
		assert.True(t, called)
	})
}

func TestProductManager_CreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		store := &fakeProductStorager{
			fetchProductsFn: func(ctx context.Context, param *ProductFilter) ([]*Product, error) {
				assert.Equal(t, ptrString("p1"), param.Name)
				return nil, nil
			},
			createProductFn: func(ctx context.Context, p *ProductParam) error {
				assert.Equal(t, ptrString("p1"), p.Name)
				return nil
			},
		}
		m := NewProductManager(&fakeTxn{}, store)
		require.NoError(t, m.CreateProduct(ctx, &ProductParam{Name: ptrString("p1")}))
	})

	t.Run("existed", func(t *testing.T) {
		store := &fakeProductStorager{
			fetchProductsFn: func(ctx context.Context, param *ProductFilter) ([]*Product, error) {
				return []*Product{{Name: "p1"}}, nil
			},
		}
		m := NewProductManager(&fakeTxn{}, store)
		err := m.CreateProduct(ctx, &ProductParam{Name: ptrString("p1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Product Record Existed")
	})

	t.Run("fetch error", func(t *testing.T) {
		store := &fakeProductStorager{
			fetchProductsFn: func(ctx context.Context, param *ProductFilter) ([]*Product, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewProductManager(&fakeTxn{}, store)
		err := m.CreateProduct(ctx, &ProductParam{Name: ptrString("p1")})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}

func TestProductManager_UpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("build-in product", func(t *testing.T) {
		m := NewProductManager(&fakeTxn{}, &fakeProductStorager{})
		err := m.UpdateProduct(ctx, &Product{ID: 1}, &ProductParam{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cant Delete Build-in Product")
	})

	t.Run("success", func(t *testing.T) {
		called := false
		store := &fakeProductStorager{
			updateProductFn: func(ctx context.Context, p *Product, newVal *ProductParam) error {
				called = true
				assert.Equal(t, int64(2), p.ID)
				assert.Equal(t, ptrString("new"), newVal.Name)
				return nil
			},
		}
		m := NewProductManager(&fakeTxn{}, store)
		require.NoError(t, m.UpdateProduct(ctx, &Product{ID: 2}, &ProductParam{Name: ptrString("new")}))
		assert.True(t, called)
	})
}

func TestProductID2NameMap(t *testing.T) {
	list := []*Product{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	got := ProductID2NameMap(list)
	assert.Equal(t, map[int64]string{1: "a", 2: "b"}, got)
}

func TestProductIDMap(t *testing.T) {
	p1 := &Product{ID: 1, Name: "a"}
	p2 := &Product{ID: 2, Name: "b"}
	got := ProductIDMap([]*Product{p1, p2})
	assert.Equal(t, map[int64]*Product{1: p1, 2: p2}, got)
}
