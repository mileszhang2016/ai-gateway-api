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

import "context"

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeProductStorager struct {
	fetchProductsFn func(ctx context.Context, param *ProductFilter) ([]*Product, error)
	deleteProductFn func(ctx context.Context, p *Product) error
	createProductFn func(ctx context.Context, p *ProductParam) error
	updateProductFn func(ctx context.Context, p *Product, newVal *ProductParam) error
}

func (f *fakeProductStorager) FetchProducts(ctx context.Context, param *ProductFilter) ([]*Product, error) {
	if f.fetchProductsFn != nil {
		return f.fetchProductsFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeProductStorager) DeleteProduct(ctx context.Context, p *Product) error {
	if f.deleteProductFn != nil {
		return f.deleteProductFn(ctx, p)
	}
	return nil
}

func (f *fakeProductStorager) CreateProduct(ctx context.Context, p *ProductParam) error {
	if f.createProductFn != nil {
		return f.createProductFn(ctx, p)
	}
	return nil
}

func (f *fakeProductStorager) UpdateProduct(ctx context.Context, p *Product, newVal *ProductParam) error {
	if f.updateProductFn != nil {
		return f.updateProductFn(ctx, p, newVal)
	}
	return nil
}

type fakeBFEClusterStorager struct {
	fetchBFEClustersFn func(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error)
	createBFEClusterFn func(ctx context.Context, param *BFEClusterParam) error
	deleteBFEClusterFn func(ctx context.Context, cluster *BFECluster) error
}

func (f *fakeBFEClusterStorager) FetchBFEClusters(ctx context.Context, param *BFEClusterFilter) ([]*BFECluster, error) {
	if f.fetchBFEClustersFn != nil {
		return f.fetchBFEClustersFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeBFEClusterStorager) CreateBFECluster(ctx context.Context, param *BFEClusterParam) error {
	if f.createBFEClusterFn != nil {
		return f.createBFEClusterFn(ctx, param)
	}
	return nil
}

func (f *fakeBFEClusterStorager) DeleteBFECluster(ctx context.Context, cluster *BFECluster) error {
	if f.deleteBFEClusterFn != nil {
		return f.deleteBFEClusterFn(ctx, cluster)
	}
	return nil
}

type fakeExtraFileStorager struct {
	createExtraFileFn func(ctx context.Context, product *Product, files ...*ExtraFileParam) error
	deleteExtraFileFn func(ctx context.Context, filter *ExtraFileFilter) error
	fetchExtraFilesFn func(ctx context.Context, filter *ExtraFileFilter) ([]*ExtraFile, error)
}

func (f *fakeExtraFileStorager) CreateExtraFile(ctx context.Context, product *Product, files ...*ExtraFileParam) error {
	if f.createExtraFileFn != nil {
		return f.createExtraFileFn(ctx, product, files...)
	}
	return nil
}

func (f *fakeExtraFileStorager) DeleteExtraFile(ctx context.Context, filter *ExtraFileFilter) error {
	if f.deleteExtraFileFn != nil {
		return f.deleteExtraFileFn(ctx, filter)
	}
	return nil
}

func (f *fakeExtraFileStorager) FetchExtraFiles(ctx context.Context, filter *ExtraFileFilter) ([]*ExtraFile, error) {
	if f.fetchExtraFilesFn != nil {
		return f.fetchExtraFilesFn(ctx, filter)
	}
	return nil, nil
}
