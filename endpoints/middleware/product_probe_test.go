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

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeProductStorager struct {
	fetchProductsFn func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error)
	deleteProductFn func(ctx context.Context, p *ibasic.Product) error
	createProductFn func(ctx context.Context, p *ibasic.ProductParam) error
	updateProductFn func(ctx context.Context, p *ibasic.Product, newVal *ibasic.ProductParam) error
}

func (f *fakeProductStorager) FetchProducts(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
	if f.fetchProductsFn != nil {
		return f.fetchProductsFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeProductStorager) DeleteProduct(ctx context.Context, p *ibasic.Product) error {
	if f.deleteProductFn != nil {
		return f.deleteProductFn(ctx, p)
	}
	return nil
}

func (f *fakeProductStorager) CreateProduct(ctx context.Context, p *ibasic.ProductParam) error {
	if f.createProductFn != nil {
		return f.createProductFn(ctx, p)
	}
	return nil
}

func (f *fakeProductStorager) UpdateProduct(ctx context.Context, p *ibasic.Product, newVal *ibasic.ProductParam) error {
	if f.updateProductFn != nil {
		return f.updateProductFn(ctx, p, newVal)
	}
	return nil
}

func setProductManager(storager ibasic.ProductStorager) func() {
	old := container.ProductManager
	container.ProductManager = ibasic.NewProductManager(&fakeTxn{}, storager)
	return func() {
		container.ProductManager = old
	}
}

func TestProductProbeAction_NoParam(t *testing.T) {
	defer setProductManager(&fakeProductStorager{})()

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	newReq, err := ProductProbeAction(req)

	require.NoError(t, err)
	assert.Equal(t, req, newReq)
}

func TestProductProbeAction_SuccessByID(t *testing.T) {
	expected := &ibasic.Product{ID: 2, Name: "demo"}
	defer setProductManager(&fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
			require.NotNil(t, param.ID)
			assert.Equal(t, int64(2), *param.ID)
			return []*ibasic.Product{expected}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products/2", nil)
	req = mux.SetURLVars(req, map[string]string{"product_id": "2"})
	newReq, err := ProductProbeAction(req)

	require.NoError(t, err)
	require.NotNil(t, newReq)

	product, err := ibasic.MustGetProduct(newReq.Context())
	require.NoError(t, err)
	assert.Equal(t, expected, product)
}

func TestProductProbeAction_SuccessByName(t *testing.T) {
	expected := &ibasic.Product{ID: 3, Name: "ai-product"}
	defer setProductManager(&fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
			require.NotNil(t, param.Name)
			assert.Equal(t, "ai-product", *param.Name)
			return []*ibasic.Product{expected}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products/ai-product", nil)
	req = mux.SetURLVars(req, map[string]string{"product_name": "ai-product"})
	newReq, err := ProductProbeAction(req)

	require.NoError(t, err)
	require.NotNil(t, newReq)

	product, err := ibasic.MustGetProduct(newReq.Context())
	require.NoError(t, err)
	assert.Equal(t, expected, product)
}

func TestProductProbeAction_MultipleProducts(t *testing.T) {
	defer setProductManager(&fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
			return []*ibasic.Product{
				{ID: 1, Name: "a"},
				{ID: 2, Name: "b"},
			}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"product_id": "1"})
	newReq, err := ProductProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
	assert.Contains(t, err.Error(), "Product Record Not Exist")
}

func TestProductProbeAction_NotFound(t *testing.T) {
	defer setProductManager(&fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
			return nil, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	req = mux.SetURLVars(req, map[string]string{"product_id": "999"})
	newReq, err := ProductProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
	assert.Contains(t, err.Error(), "Product Record Not Exist")
}

func TestProductProbeAction_FetchError(t *testing.T) {
	defer setProductManager(&fakeProductStorager{
		fetchProductsFn: func(ctx context.Context, param *ibasic.ProductFilter) ([]*ibasic.Product, error) {
			return nil, errors.New("database down")
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"product_id": "1"})
	newReq, err := ProductProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
	assert.Equal(t, "database down", err.Error())
}

func TestProductProbeAction_BindURIError(t *testing.T) {
	defer setProductManager(&fakeProductStorager{})()

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"product_id": "abc"})
	newReq, err := ProductProbeAction(req)

	require.Error(t, err)
	assert.Nil(t, newReq)
}

func TestNewMiddleWareFunc_Success(t *testing.T) {
	called := false
	handler := func(req *http.Request) (*http.Request, error) {
		called = true
		return req, nil
	}

	mw := NewMiddleWareFunc(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	nextCalled := false
	mw(rw, req, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		assert.Equal(t, req, r)
	})

	assert.True(t, called)
	assert.True(t, nextCalled)
}

func TestNewMiddleWareFunc_Error(t *testing.T) {
	handler := func(req *http.Request) (*http.Request, error) {
		return nil, errors.New("probe failed")
	}

	mw := NewMiddleWareFunc(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	nextCalled := false
	mw(rw, req, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "probe failed")
}
