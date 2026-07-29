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

package iprotocol

import (
	"context"

	"github.com/yf-networks/ai-gateway-api/model/ibasic"
	"github.com/yf-networks/ai-gateway-api/model/iversion_control"
)

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeCertificateStorager struct {
	fetchCertificatesFn  func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error)
	deleteCertificateFn  func(ctx context.Context, cert *Certificate) error
	createCertificateFn  func(ctx context.Context, param *CertificateParam) error
	updateCertificateFn  func(ctx context.Context, cert *Certificate, param *CertificateParam) error
}

func (f *fakeCertificateStorager) FetchCertificates(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
	if f.fetchCertificatesFn != nil {
		return f.fetchCertificatesFn(ctx, filter)
	}
	return nil, nil
}

func (f *fakeCertificateStorager) DeleteCertificate(ctx context.Context, cert *Certificate) error {
	if f.deleteCertificateFn != nil {
		return f.deleteCertificateFn(ctx, cert)
	}
	return nil
}

func (f *fakeCertificateStorager) CreateCertificate(ctx context.Context, param *CertificateParam) error {
	if f.createCertificateFn != nil {
		return f.createCertificateFn(ctx, param)
	}
	return nil
}

func (f *fakeCertificateStorager) UpdateCertificate(ctx context.Context, cert *Certificate, param *CertificateParam) error {
	if f.updateCertificateFn != nil {
		return f.updateCertificateFn(ctx, cert, param)
	}
	return nil
}

type fakeExtraFileStorager struct {
	createExtraFileFn func(ctx context.Context, product *ibasic.Product, files ...*ibasic.ExtraFileParam) error
	deleteExtraFileFn func(ctx context.Context, filter *ibasic.ExtraFileFilter) error
	fetchExtraFilesFn func(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error)
}

func (f *fakeExtraFileStorager) CreateExtraFile(ctx context.Context, product *ibasic.Product, files ...*ibasic.ExtraFileParam) error {
	if f.createExtraFileFn != nil {
		return f.createExtraFileFn(ctx, product, files...)
	}
	return nil
}

func (f *fakeExtraFileStorager) DeleteExtraFile(ctx context.Context, filter *ibasic.ExtraFileFilter) error {
	if f.deleteExtraFileFn != nil {
		return f.deleteExtraFileFn(ctx, filter)
	}
	return nil
}

func (f *fakeExtraFileStorager) FetchExtraFiles(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error) {
	if f.fetchExtraFilesFn != nil {
		return f.fetchExtraFilesFn(ctx, filter)
	}
	return nil, nil
}

type fakeVersionControlStorager struct {
	upsertFn func(ctx context.Context, css *iversion_control.ExportData) (string, error)
}

func (f *fakeVersionControlStorager) UpsertConfigLastExportedVersion(ctx context.Context, css *iversion_control.ExportData) (string, error) {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, css)
	}
	return "", nil
}
