// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iprovider

import "context"

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeProviderStorager struct {
	createFn        func(ctx context.Context, param *ProviderParam) (int64, error)
	updateFn        func(ctx context.Context, name string, param *ProviderParam) error
	deleteFn        func(ctx context.Context, name string) error
	fetchFn         func(ctx context.Context, filter *ProviderFilter) (*Provider, error)
	fetchListFn     func(ctx context.Context, filter *ProviderFilter) ([]*Provider, int64, error)
	fetchNamesFn    func(ctx context.Context) ([]string, error)
}

func (s *fakeProviderStorager) CreateProvider(ctx context.Context, param *ProviderParam) (int64, error) {
	if s.createFn != nil {
		return s.createFn(ctx, param)
	}
	return 1, nil
}

func (s *fakeProviderStorager) UpdateProvider(ctx context.Context, name string, param *ProviderParam) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, name, param)
	}
	return nil
}

func (s *fakeProviderStorager) DeleteProvider(ctx context.Context, name string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, name)
	}
	return nil
}

func (s *fakeProviderStorager) FetchProvider(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, filter)
	}
	return nil, nil
}

func (s *fakeProviderStorager) FetchProviderList(ctx context.Context, filter *ProviderFilter) ([]*Provider, int64, error) {
	if s.fetchListFn != nil {
		return s.fetchListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (s *fakeProviderStorager) FetchProviderNames(ctx context.Context) ([]string, error) {
	if s.fetchNamesFn != nil {
		return s.fetchNamesFn(ctx)
	}
	return nil, nil
}

type fakeDiscoverCaller struct {
	body        []byte
	err         error
	lastMethod  string
	lastURL     string
	lastHeaders map[string]string
}

func (c *fakeDiscoverCaller) Call(ctx context.Context, method, url string, headers map[string]string) ([]byte, error) {
	c.lastMethod = method
	c.lastURL = url
	c.lastHeaders = headers
	return c.body, c.err
}
