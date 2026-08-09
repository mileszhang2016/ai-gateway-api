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

package protocol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/endpoints/innerapi_v1/internal/testutil"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iprotocol"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful/container"
)

type fakeCertificateStorager struct {
	fetchCertificatesFn func(ctx context.Context, param *iprotocol.CertificateFilter) ([]*iprotocol.Certificate, error)
}

func (f *fakeCertificateStorager) FetchCertificates(ctx context.Context, param *iprotocol.CertificateFilter) ([]*iprotocol.Certificate, error) {
	if f.fetchCertificatesFn != nil {
		return f.fetchCertificatesFn(ctx, param)
	}
	return nil, nil
}

func (f *fakeCertificateStorager) DeleteCertificate(ctx context.Context, cert *iprotocol.Certificate) error {
	return nil
}

func (f *fakeCertificateStorager) CreateCertificate(ctx context.Context, param *iprotocol.CertificateParam) error {
	return nil
}

func (f *fakeCertificateStorager) UpdateCertificate(ctx context.Context, cert *iprotocol.Certificate, param *iprotocol.CertificateParam) error {
	return nil
}

type fakeExtraFileStoragerForCert struct{}

func (f *fakeExtraFileStoragerForCert) CreateExtraFile(ctx context.Context, product *ibasic.Product, files ...*ibasic.ExtraFileParam) error {
	return nil
}

func (f *fakeExtraFileStoragerForCert) DeleteExtraFile(ctx context.Context, filter *ibasic.ExtraFileFilter) error {
	return nil
}

func (f *fakeExtraFileStoragerForCert) FetchExtraFiles(ctx context.Context, filter *ibasic.ExtraFileFilter) ([]*ibasic.ExtraFile, error) {
	return nil, nil
}

func setupCertificateManager(storager iprotocol.CertificateStorager) func() {
	old := container.CertificateManager
	container.CertificateManager = iprotocol.NewCertificateManager(&testutil.FakeTxn{}, storager, testutil.NewVersionControlManager("v2"), &fakeExtraFileStoragerForCert{})
	return func() {
		container.CertificateManager = old
	}
}

func TestServerCertExportAction(t *testing.T) {
	defer setupCertificateManager(&fakeCertificateStorager{
		fetchCertificatesFn: func(ctx context.Context, param *iprotocol.CertificateFilter) ([]*iprotocol.Certificate, error) {
			return []*iprotocol.Certificate{
				{
					CertName:     "default",
					IsDefault:    true,
					CertFilePath: "conf/default.crt",
					KeyFilePath:  "conf/default.key",
				},
			}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/configs/protocol/server_cert_conf?version=", nil)
	data, err := ServerCertExportAction(req)

	require.NoError(t, err)
	require.NotNil(t, data)

	conf, ok := data.(*iprotocol.ServerCertConf)
	require.True(t, ok)
	assert.Equal(t, "default", conf.Config.Default)
	assert.Contains(t, conf.Config.CertConf, "default")
}

func TestServerCertExportAction_VersionNotChanged(t *testing.T) {
	defer setupCertificateManager(&fakeCertificateStorager{
		fetchCertificatesFn: func(ctx context.Context, param *iprotocol.CertificateFilter) ([]*iprotocol.Certificate, error) {
			return []*iprotocol.Certificate{
				{
					CertName:     "default",
					IsDefault:    true,
					CertFilePath: "conf/default.crt",
					KeyFilePath:  "conf/default.key",
				},
			}, nil
		},
	})()

	req := httptest.NewRequest(http.MethodGet, "/configs/protocol/server_cert_conf?version=v2", nil)
	data, err := ServerCertExportAction(req)

	require.NoError(t, err)
	assert.Nil(t, data)
}
