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

package imodel_price

import (
	"context"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeModelPriceStorager struct {
	createFn    func(ctx context.Context, param *ModelPrice) (int64, error)
	updateFn    func(ctx context.Context, filter *ModelPriceFilter, param *ModelPrice) (int64, error)
	deleteFn    func(ctx context.Context, filter *ModelPriceFilter) error
	deleteAllFn func(ctx context.Context) error
	fetchFn     func(ctx context.Context, filter *ModelPriceFilter) (*ModelPrice, error)
	fetchListFn func(ctx context.Context, filter *ModelPriceFilter) ([]*ModelPrice, int64, error)
	created     []*ModelPrice
	updated     []struct {
		filter *ModelPriceFilter
		param  *ModelPrice
	}
	deletedFilters []*ModelPriceFilter
	deletedAll     bool
}

func (s *fakeModelPriceStorager) CreateModelPrice(ctx context.Context, param *ModelPrice) (int64, error) {
	s.created = append(s.created, param)
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 1, nil
}

func (s *fakeModelPriceStorager) UpdateModelPrice(ctx context.Context, filter *ModelPriceFilter, param *ModelPrice) (int64, error) {
	s.updated = append(s.updated, struct {
		filter *ModelPriceFilter
		param  *ModelPrice
	}{filter: filter, param: param})
	if s.updateFn != nil {
		return s.updateFn(ctx, filter, param)
	}
	return 1, nil
}

func (s *fakeModelPriceStorager) DeleteModelPrice(ctx context.Context, filter *ModelPriceFilter) error {
	s.deletedFilters = append(s.deletedFilters, filter)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, filter)
	}
	return nil
}

func (s *fakeModelPriceStorager) DeleteAllModelPrices(ctx context.Context) error {
	s.deletedAll = true
	if s.deleteAllFn != nil {
		return s.deleteAllFn(ctx)
	}
	return nil
}

func (s *fakeModelPriceStorager) FetchModelPrice(ctx context.Context, filter *ModelPriceFilter) (*ModelPrice, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeModelPriceStorager) FetchModelPriceList(ctx context.Context, filter *ModelPriceFilter) ([]*ModelPrice, int64, error) {
	if s.fetchListFn != nil {
		return s.fetchListFn(ctx, filter)
	}
	return nil, 0, nil
}

type fakeProviderStorager struct {
	fetchFn     func(ctx context.Context, filter *iprovider.ProviderFilter) (*iprovider.Provider, error)
	fetchListFn func(ctx context.Context, filter *iprovider.ProviderFilter) ([]*iprovider.Provider, int64, error)
}

func (s *fakeProviderStorager) CreateProvider(ctx context.Context, param *iprovider.ProviderParam) (int64, error) {
	return 1, nil
}

func (s *fakeProviderStorager) UpdateProvider(ctx context.Context, name string, param *iprovider.ProviderParam) error {
	return nil
}

func (s *fakeProviderStorager) DeleteProvider(ctx context.Context, name string) error {
	return nil
}

func (s *fakeProviderStorager) FetchProvider(ctx context.Context, filter *iprovider.ProviderFilter) (*iprovider.Provider, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	if filter != nil && filter.Name != nil {
		return &iprovider.Provider{Name: *filter.Name}, nil
	}
	return &iprovider.Provider{Name: "openai"}, nil
}

func (s *fakeProviderStorager) FetchProviderList(ctx context.Context, filter *iprovider.ProviderFilter) ([]*iprovider.Provider, int64, error) {
	if s.fetchListFn != nil {
		return s.fetchListFn(ctx, filter)
	}
	return nil, 0, nil
}

func providerStoragerAlwaysExists() *fakeProviderStorager {
	return &fakeProviderStorager{}
}

func TestManagerCreateModelPrice(t *testing.T) {
	ctx := context.Background()
	store := &fakeModelPriceStorager{}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	id, err := m.CreateModelPrice(ctx, validModelPrice())
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.Len(t, store.created, 1)

	_, err = m.CreateModelPrice(ctx, &ModelPrice{})
	assert.Error(t, err)
}

func TestManagerUpdateModelPrice(t *testing.T) {
	ctx := context.Background()
	store := &fakeModelPriceStorager{}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	id := int64(7)
	affected, err := m.UpdateModelPrice(ctx, &ModelPriceFilter{ID: &id}, validModelPrice())
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	require.Len(t, store.updated, 1)
	assert.Equal(t, &id, store.updated[0].filter.ID)

	_, err = m.UpdateModelPrice(ctx, &ModelPriceFilter{}, &ModelPrice{})
	assert.Error(t, err)
}

func TestManagerDeleteModelPrice(t *testing.T) {
	ctx := context.Background()
	store := &fakeModelPriceStorager{}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	id := int64(7)
	err := m.DeleteModelPrice(ctx, &ModelPriceFilter{ID: &id})
	require.NoError(t, err)
	require.Len(t, store.deletedFilters, 1)
	assert.Equal(t, &id, store.deletedFilters[0].ID)
}

func TestManagerFetchModelPrice(t *testing.T) {
	ctx := context.Background()
	expected := validModelPrice()
	store := &fakeModelPriceStorager{
		fetchFn: func(ctx context.Context, filter *ModelPriceFilter) (*ModelPrice, error) {
			return expected, nil
		},
	}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	id := int64(7)
	one, err := m.FetchModelPrice(ctx, &ModelPriceFilter{ID: &id})
	require.NoError(t, err)
	assert.Equal(t, expected, one)
}

func TestManagerFetchModelPriceList(t *testing.T) {
	ctx := context.Background()
	expected := []*ModelPrice{validModelPrice()}
	store := &fakeModelPriceStorager{
		fetchListFn: func(ctx context.Context, filter *ModelPriceFilter) ([]*ModelPrice, int64, error) {
			return expected, 1, nil
		},
	}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	list, total, err := m.FetchModelPriceList(ctx, &ModelPriceFilter{})
	require.NoError(t, err)
	assert.Equal(t, expected, list)
	assert.Equal(t, int64(1), total)
}

func TestManagerImportModelPricesInvalidMode(t *testing.T) {
	ctx := context.Background()
	m := NewManager(&fakeTxn{}, &fakeModelPriceStorager{}, providerStoragerAlwaysExists())

	_, _, err := m.ImportModelPrices(ctx, []*ModelPrice{validModelPrice()}, "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "import mode must be replace or merge")
}

func TestManagerImportModelPricesInvalidEntry(t *testing.T) {
	ctx := context.Background()
	m := NewManager(&fakeTxn{}, &fakeModelPriceStorager{}, providerStoragerAlwaysExists())

	_, _, err := m.ImportModelPrices(ctx, []*ModelPrice{validModelPrice(), {}}, string(ImportModeReplace))
	assert.Error(t, err)
}

func TestManagerImportModelPricesReplace(t *testing.T) {
	ctx := context.Background()
	store := &fakeModelPriceStorager{}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	imported, skipped, err := m.ImportModelPrices(ctx, []*ModelPrice{validModelPrice()}, string(ImportModeReplace))
	require.NoError(t, err)
	assert.Equal(t, 1, imported)
	assert.Equal(t, 0, skipped)
	assert.True(t, store.deletedAll)
	assert.Len(t, store.created, 1)
}

func TestManagerImportModelPricesMerge(t *testing.T) {
	ctx := context.Background()
	existing := validModelPrice()
	existing.ID = 1
	existing.Provider = "openai"
	existing.Model = "gpt-4"
	existing.Mode = "chat"

	newOne := validModelPrice()
	newOne.Provider = "openai"
	newOne.Model = "gpt-4"
	newOne.Mode = "embedding"

	store := &fakeModelPriceStorager{
		fetchListFn: func(ctx context.Context, filter *ModelPriceFilter) ([]*ModelPrice, int64, error) {
			return []*ModelPrice{existing}, 1, nil
		},
	}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	imported, skipped, err := m.ImportModelPrices(ctx, []*ModelPrice{validModelPrice(), newOne}, string(ImportModeMerge))
	require.NoError(t, err)
	assert.Equal(t, 2, imported)
	assert.Equal(t, 0, skipped)
	require.Len(t, store.updated, 1)
	assert.Equal(t, int64(1), *store.updated[0].filter.ID)
	require.Len(t, store.created, 1)
	assert.Equal(t, "embedding", store.created[0].Mode)
}

func TestManagerImportModelPricesStorageError(t *testing.T) {
	ctx := context.Background()
	store := &fakeModelPriceStorager{
		deleteAllFn: func(ctx context.Context) error {
			return errors.New("delete failed")
		},
	}
	m := NewManager(&fakeTxn{}, store, providerStoragerAlwaysExists())

	_, _, err := m.ImportModelPrices(ctx, []*ModelPrice{validModelPrice()}, string(ImportModeReplace))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}
