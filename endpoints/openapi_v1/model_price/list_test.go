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

package model_price

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/imodel_price"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful/container"
)

// fakeModelPriceStorager implements imodel_price.ModelPriceStorager with
// per-method callbacks; unconfigured methods return zero values.
type fakeModelPriceStorager struct {
	fetchFn     func(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error)
	fetchListFn func(ctx context.Context, filter *imodel_price.ModelPriceFilter) ([]*imodel_price.ModelPrice, int64, error)
}

func (f *fakeModelPriceStorager) CreateModelPrice(ctx context.Context, param *imodel_price.ModelPrice) (int64, error) {
	return 0, nil
}

func (f *fakeModelPriceStorager) UpdateModelPrice(ctx context.Context, filter *imodel_price.ModelPriceFilter, param *imodel_price.ModelPrice) (int64, error) {
	return 0, nil
}

func (f *fakeModelPriceStorager) DeleteModelPrice(ctx context.Context, filter *imodel_price.ModelPriceFilter) error {
	return nil
}

func (f *fakeModelPriceStorager) DeleteAllModelPrices(ctx context.Context) error {
	return nil
}

func (f *fakeModelPriceStorager) FetchModelPrice(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error) {
	if f.fetchFn != nil {
		return f.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (f *fakeModelPriceStorager) FetchModelPriceList(ctx context.Context, filter *imodel_price.ModelPriceFilter) ([]*imodel_price.ModelPrice, int64, error) {
	if f.fetchListFn != nil {
		return f.fetchListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (f *fakeModelPriceStorager) ListProviders(ctx context.Context) ([]string, error) {
	return nil, nil
}

var _ imodel_price.ModelPriceStorager = (*fakeModelPriceStorager)(nil)

// setupListManager swaps container.ModelPriceManager for the duration of a test.
func setupListManager(storager imodel_price.ModelPriceStorager) func() {
	old := container.ModelPriceManager
	container.ModelPriceManager = imodel_price.NewManager(nil, storager)
	return func() { container.ModelPriceManager = old }
}

func TestListActionSingleRecord(t *testing.T) {
	var gotFilter *imodel_price.ModelPriceFilter
	storager := &fakeModelPriceStorager{
		fetchFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error) {
			gotFilter = filter
			return &imodel_price.ModelPrice{ID: 195, Provider: "deepseek", Model: "deepseek-v4-pro", Mode: "chat"}, nil
		},
	}
	defer setupListManager(storager)()

	req, _ := http.NewRequest(http.MethodGet,
		"/model-prices?provider=deepseek&model=deepseek-v4-pro&mode=chat", nil)
	data, err := ListAction(req)
	assert.NoError(t, err)

	one, ok := data.(*imodel_price.ModelPrice)
	assert.True(t, ok, "three params must yield a single *ModelPrice, got %T", data)
	assert.Equal(t, int64(195), one.ID)
	assert.Equal(t, "deepseek", one.Provider)
	assert.NotNil(t, gotFilter)
	assert.Equal(t, "deepseek", *gotFilter.Provider)
	assert.Equal(t, "deepseek-v4-pro", *gotFilter.Model)
	assert.Equal(t, "chat", *gotFilter.Mode)
}

func TestListActionSingleRecordNotFound(t *testing.T) {
	storager := &fakeModelPriceStorager{
		fetchFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error) {
			return nil, nil
		},
	}
	defer setupListManager(storager)()

	req, _ := http.NewRequest(http.MethodGet,
		"/model-prices?provider=deepseek&model=deepseek-v4-pro&mode=chat", nil)
	data, err := ListAction(req)
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ModelPrice Record Not Exist")
}

func TestListActionInvalidMode(t *testing.T) {
	storager := &fakeModelPriceStorager{
		fetchFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error) {
			t.Fatal("storager must not be called for an invalid mode")
			return nil, nil
		},
	}
	defer setupListManager(storager)()

	req, _ := http.NewRequest(http.MethodGet,
		"/model-prices?provider=deepseek&model=deepseek-v4-pro&mode=foo", nil)
	data, err := ListAction(req)
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode: foo")
}

func TestListActionMissingParamsWithModel(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"model only", "/model-prices?model=gpt-4"},
		{"provider and model without mode", "/model-prices?provider=deepseek&model=gpt-4"},
		{"model and mode without provider", "/model-prices?model=gpt-4&mode=chat"},
		{"provider and mode without model stays list", "/model-prices?provider=deepseek&mode=chat&page=2&page_size=10"},
	}

	for _, c := range cases[:3] {
		t.Run(c.name, func(t *testing.T) {
			storager := &fakeModelPriceStorager{
				fetchFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) (*imodel_price.ModelPrice, error) {
					t.Fatal("storager must not be called when params are incomplete")
					return nil, nil
				},
			}
			defer setupListManager(storager)()

			req, _ := http.NewRequest(http.MethodGet, c.url, nil)
			data, err := ListAction(req)
			assert.Nil(t, data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "provider, model and mode are required")
		})
	}

	t.Run(cases[3].name, func(t *testing.T) {
		var gotFilter *imodel_price.ModelPriceFilter
		storager := &fakeModelPriceStorager{
			fetchListFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) ([]*imodel_price.ModelPrice, int64, error) {
				gotFilter = filter
				return []*imodel_price.ModelPrice{{ID: 1, Provider: "deepseek", Model: "m1", Mode: "chat"}}, 1, nil
			},
		}
		defer setupListManager(storager)()

		req, _ := http.NewRequest(http.MethodGet, cases[3].url, nil)
		data, err := ListAction(req)
		assert.NoError(t, err)
		resp, ok := data.(*listResponse)
		assert.True(t, ok, "provider+mode without model must stay list shape, got %T", data)
		assert.Len(t, resp.List, 1)
		assert.Equal(t, int64(1), resp.Pagination.Total)
		assert.Nil(t, gotFilter.Model)
		assert.Equal(t, "deepseek", *gotFilter.Provider)
		assert.Equal(t, "chat", *gotFilter.Mode)
		assert.Equal(t, 2, *gotFilter.Page)
		assert.Equal(t, 10, *gotFilter.PageSize)
	})
}

func TestListActionListShapeUnchanged(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"no params", "/model-prices"},
		{"provider only", "/model-prices?provider=deepseek"},
		{"mode only", "/model-prices?mode=chat"},
		{"provider and mode", "/model-prices?provider=deepseek&mode=chat"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			storager := &fakeModelPriceStorager{
				fetchListFn: func(ctx context.Context, filter *imodel_price.ModelPriceFilter) ([]*imodel_price.ModelPrice, int64, error) {
					return []*imodel_price.ModelPrice{{ID: 1, Provider: "deepseek", Model: "m1", Mode: "chat"}}, 1, nil
				},
			}
			defer setupListManager(storager)()

			req, _ := http.NewRequest(http.MethodGet, c.url, nil)
			data, err := ListAction(req)
			assert.NoError(t, err)
			resp, ok := data.(*listResponse)
			assert.True(t, ok, "must keep list shape, got %T", data)
			assert.Len(t, resp.List, 1)
			assert.Equal(t, 1, resp.Pagination.Page)
			assert.Equal(t, 50, resp.Pagination.PageSize)
			assert.Equal(t, int64(1), resp.Pagination.Total)
		})
	}
}
