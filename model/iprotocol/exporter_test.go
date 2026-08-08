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

package iprotocol

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
)

func TestCertificateManager_ExportServerCert(t *testing.T) {
	ctx := context.Background()

	t.Run("unchanged", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{
					{CertName: "default", IsDefault: true, CertFilePath: "tls_conf/p1/cert.pem", KeyFilePath: "tls_conf/p1/key.pem"},
				}, nil
			},
		}
		vcs := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return iversion_control.ZeroVersion, nil
			},
		}
		vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
		m := NewCertificateManager(&fakeTxn{}, store, vcm, &fakeExtraFileStorager{})

		conf, err := m.ExportServerCert(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		assert.Nil(t, conf)
	})

	t.Run("changed", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{
					{CertName: "default", IsDefault: true, CertFilePath: "tls_conf/p1/cert.pem", KeyFilePath: "tls_conf/p1/key.pem"},
				}, nil
			},
		}
		vcs := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				assert.Equal(t, ConfigTopicServerCert, css.Topic)
				return "20240102000000", nil
			},
		}
		vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, vcs)
		m := NewCertificateManager(&fakeTxn{}, store, vcm, &fakeExtraFileStorager{})

		conf, err := m.ExportServerCert(ctx, "old-version")
		require.NoError(t, err)
		require.NotNil(t, conf)
		assert.Equal(t, "20240102000000", conf.Version)
		assert.Equal(t, "default", conf.Config.Default)
	})

	t.Run("fetch error", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return nil, errors.New("db down")
			},
		}
		vcm := iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{})
		m := NewCertificateManager(&fakeTxn{}, store, vcm, &fakeExtraFileStorager{})

		_, err := m.ExportServerCert(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db down")
	})
}
