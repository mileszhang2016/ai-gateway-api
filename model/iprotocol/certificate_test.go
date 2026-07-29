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
	"strings"
	"testing"

	"github.com/bfenetworks/bfe/bfe_config/bfe_tls_conf/server_cert_conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib"
	"github.com/yf-networks/ai-gateway-api/model/ibasic"
)

const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIC/zCCAeegAwIBAgIUUXnaztO0ae70JAnMcjhjhDgIbSswDQYJKoZIhvcNAQEL
BQAwDzENMAsGA1UEAwwEdGVzdDAeFw0yNjA3MjgxMjU0MTVaFw0yNjA3MjkxMjU0
MTVaMA8xDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDh8jdus7eatzuBA9cev+sGKZEuq0j4OFkITI1O4X3viJJg1I9F1crjkT+w
AP2htCnpOWdkwsYG8Kv+zM2pab38yXDZwum7azW0LG1OdxYajnmXRdPq86tET+Lx
siVsoqcPtoncmNWB2Wj8AAjYuQR3QcN+8h9zGZn4wh3FY1lwyQLY+Wlkg1b6T2Lo
v9RoRb2P4N/v54RKT/BwXNJmFY7Gd9Lvx81PplbAfsJelMVATJBYcKbdeELScgAB
CNI2i1xLSx07qFh+f1CW8l6ZUUtSEPgIAlTHtDw4I+tQz0r39owq6phheSVkWmbG
SBKDd7rxpAAvzBv6zlzyOKv+SHRLAgMBAAGjUzBRMB0GA1UdDgQWBBRh3KXxJRXo
iUrTuxkrcL0fQcYLWDAfBgNVHSMEGDAWgBRh3KXxJRXoiUrTuxkrcL0fQcYLWDAP
BgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBwRDGfB2SEez6dtq6U
3aNXSEQE6vntYUSsF3tmKc87i2/z8q4C9r8RaglmUXfz8N0+lpw2ek57nzE32AvE
9ptctqs2vNYWcVJKReKYeXDXAgIzU9RkSq6o7xbhnv0BQ1rI0hewVK+EoDmUSniC
gtxoP75GMOgrWDUGUNXEHNGaYtHRKriXU6s8AWNXNkSgEi4q6VavgRk0PXj5OPYZ
fDWgmOSLNJ4Am83HpGRub6QaY6YXLw+V+44mWv8wVok0euLoeDdhQamfh0o0syF8
pwd7BdgoVuNdFNF9PJhFf17p8qAFnFob9qmzFQcSXGSZfsRnVH/NtsmDkfYWm2fu
/VuH
-----END CERTIFICATE-----`

const testKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDh8jdus7eatzuB
A9cev+sGKZEuq0j4OFkITI1O4X3viJJg1I9F1crjkT+wAP2htCnpOWdkwsYG8Kv+
zM2pab38yXDZwum7azW0LG1OdxYajnmXRdPq86tET+LxsiVsoqcPtoncmNWB2Wj8
AAjYuQR3QcN+8h9zGZn4wh3FY1lwyQLY+Wlkg1b6T2Lov9RoRb2P4N/v54RKT/Bw
XNJmFY7Gd9Lvx81PplbAfsJelMVATJBYcKbdeELScgABCNI2i1xLSx07qFh+f1CW
8l6ZUUtSEPgIAlTHtDw4I+tQz0r39owq6phheSVkWmbGSBKDd7rxpAAvzBv6zlzy
OKv+SHRLAgMBAAECggEAW7onwut3CHqGz7Ota7BiS5godpfXAd5uVq4tV+63X71E
H8dreuB2g7h98IgWb8VilmpVjVR9bGfci469l590H+HvzJgSp6G4pbK7lXVHJfTd
bApPJD1UNGFyMskt5FKMHBbxFPt/Aj4vHs8syD8kjv3Fzg2hsuqb1Z+I0o+oyd4t
yXILyz214yjNmGDpRMPlRNWT7DPvvpo4qH02wCpARWAFD+0X3I9Xay5e3bj+ZIE0
hAobDLg5s+DazhrhGt9lvUwY3HO5FAS1TMVOfHUT0LTueC/hdi9+NHYWcuypNWJ4
GuC3BSn+AfbWTJe6w/AFuxl2LnGGaU2ULMUkLT5UMQKBgQD39Wh6jmo2Q5nlFI02
jqqmabnXvz3DfD+dtoAsofeEK79kKp51gw+ndUN9nAAOUqbhSq96Q2wW5Puyrt+f
m2Yl7yrXaZ8COQIL330sn/tr+pvhOUR2yNh+dQWCde2F46yiNT3vC0tdrb2Bo+n+
nxuAFArHwk3VLEPL8CP4Mkd4hwKBgQDpRg6yyIk7L02WS61vDWEvgS/H10vc/exi
BJshOI1EKobCopQ0qdKhGy6MSc7tBwOuCoQBy68pj0Ip/9bNrL+jReG/iDYy4naA
TOYuvkfNd1s7vrBhcvxCbeKhMOkBFLcWuhtpEEaS92/hnj1K3+03PKupso/gXTvn
I9JhEG4LHQKBgQC/5/1+rO5jJRrcg3VvfidxOG6PHgINZAJQa8jzwj8w2jL8sUeG
p3LcJhOgCba5XxqtTwJU3A2yAnMTLekBPGJohZxgr+xS6hA9ZDEa8o7CWWl/fLUS
Qgvcg3FKMT8t2rHnsNFISzN/Q1JiHZyiZj4AeIKHbEiU7fdixW7xTuilzQKBgBza
R3Mhjqe9YBFY5ui3dO/VQL2tCXsaBSTSgQWI4yAtSmHEjiQ9ZQn8PLOpZWi312Kt
dkpqkQ3I5FwhgsYJueJOAHAaPunoTNPtrwLVEjh9rNEk8tf6yuzEfqWFUSyLDWJI
Pp+uHayL4lC7q8UZEVQlsu3YYidUINaj/Z930sSZAoGAEcWaHeMdJT13NPQiD7de
RS0h43AAKJfCWpEZbwRGmMc6/v/TVgSG4A8VPFGCGx6cM2wU2XruPpCpYOOvOaXv
kKGDPLVuP49nM+9takKdKzFYdQ56B7c2OUpE206G88xNjCXkgGQ+izHVJ1nF18I8
V0cpIOiw0jCXydiOVTtvq1o=
-----END PRIVATE KEY-----`

func TestNewCertificateManager(t *testing.T) {
	m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
	require.NotNil(t, m)
}

func TestCertificateManager_FetchCertificates(t *testing.T) {
	ctx := context.Background()
	expected := []*Certificate{{CertName: "c1"}}
	store := &fakeCertificateStorager{
		fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
			assert.True(t, *filter.IsDefault)
			return expected, nil
		},
	}
	m := NewCertificateManager(&fakeTxn{}, store, nil, &fakeExtraFileStorager{})

	got, err := m.FetchCertificates(ctx, &CertificateFilter{IsDefault: lib.PBool(true)})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestCertificateManager_DeleteCertificate(t *testing.T) {
	ctx := context.Background()

	t.Run("default certificate", func(t *testing.T) {
		m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
		err := m.DeleteCertificate(ctx, &Certificate{IsDefault: true})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cant Delete Default Certificate")
	})

	t.Run("referenced by product", func(t *testing.T) {
		m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
		err := m.DeleteCertificate(ctx, &Certificate{Products: []*ibasic.Product{{}}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cant Delete Certificate Be Refer By Product")
	})

	t.Run("success", func(t *testing.T) {
		deleted := false
		extraDeleted := false
		store := &fakeCertificateStorager{
			deleteCertificateFn: func(ctx context.Context, cert *Certificate) error {
				deleted = true
				return nil
			},
		}
		extraStore := &fakeExtraFileStorager{
			deleteExtraFileFn: func(ctx context.Context, filter *ibasic.ExtraFileFilter) error {
				extraDeleted = true
				assert.ElementsMatch(t, []string{"cert.pem", "key.pem"}, filter.Names)
				return nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, extraStore)
		require.NoError(t, m.DeleteCertificate(ctx, &Certificate{CertFilePath: "cert.pem", KeyFilePath: "key.pem"}))
		assert.True(t, extraDeleted)
		assert.True(t, deleted)
	})
}

func Test_validateCertPair(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, validateCertPair("cert.pem", testCertPEM, "key.pem", testKeyPEM))
	})

	t.Run("bad cert format", func(t *testing.T) {
		badCert := "-----BEGIN FOO-----\nYmFy\n-----END FOO-----"
		err := validateCertPair("cert.pem", badCert, "key.pem", testKeyPEM)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Certificate File Format Must Be PEM")
	})

	t.Run("bad key format", func(t *testing.T) {
		err := validateCertPair("cert.pem", testCertPEM, "key.pem", "not-pem")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Certificate Private Key File Format Must Be PEM")
	})

	t.Run("mismatched pair", func(t *testing.T) {
		otherKey := strings.ReplaceAll(testKeyPEM, "QcN+8h9z", "AAAAAAAA")
		err := validateCertPair("cert.pem", testCertPEM, "key.pem", otherKey)
		require.Error(t, err)
	})
}

func TestCertificateManager_CreateCertificate(t *testing.T) {
	ctx := context.Background()
	validParam := func() *CertificateParam {
		return &CertificateParam{
			CertName:        lib.PString("c1"),
			CertFileName:    lib.PString("cert.pem"),
			CertFileContent: lib.PString(testCertPEM),
			KeyFileName:     lib.PString("key.pem"),
			KeyFileContent:  lib.PString(testKeyPEM),
			IsDefault:       lib.PBool(true),
		}
	}

	t.Run("success as default", func(t *testing.T) {
		created := false
		extraCreated := false
		defaultUpdated := false
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{{CertName: "old", IsDefault: true}}, nil
			},
			createCertificateFn: func(ctx context.Context, param *CertificateParam) error {
				created = true
				assert.True(t, strings.HasSuffix(*param.CertFilePath, "cert.pem"))
				assert.True(t, strings.HasSuffix(*param.KeyFilePath, "key.pem"))
				return nil
			},
			updateCertificateFn: func(ctx context.Context, cert *Certificate, param *CertificateParam) error {
				defaultUpdated = true
				assert.Equal(t, "old", cert.CertName)
				assert.NotNil(t, param.IsDefault)
				assert.False(t, *param.IsDefault)
				return nil
			},
		}
		extraStore := &fakeExtraFileStorager{
			createExtraFileFn: func(ctx context.Context, product *ibasic.Product, files ...*ibasic.ExtraFileParam) error {
				extraCreated = true
				require.Len(t, files, 2)
				assert.Equal(t, []byte(testCertPEM), files[0].Content)
				assert.Equal(t, []byte(testKeyPEM), files[1].Content)
				return nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, extraStore)
		require.NoError(t, m.CreateCertificate(ctx, validParam()))
		assert.True(t, extraCreated)
		assert.True(t, defaultUpdated)
		assert.True(t, created)
	})

	t.Run("success as non-default when no default exists", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return nil, nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, &fakeExtraFileStorager{})
		param := validParam()
		param.IsDefault = lib.PBool(false)
		err := m.CreateCertificate(ctx, param)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must Has Default Certificate")
	})

	t.Run("duplicate cert name", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{{CertName: "c1"}}, nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, &fakeExtraFileStorager{})
		err := m.CreateCertificate(ctx, validParam())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Certification Record Existed")
	})

	t.Run("duplicate file name", func(t *testing.T) {
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{{CertName: "other", CertFileName: "cert.pem"}}, nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, &fakeExtraFileStorager{})
		err := m.CreateCertificate(ctx, validParam())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Certificate File Name cert.pem By Used By other")
	})

	t.Run("invalid cert pair", func(t *testing.T) {
		param := validParam()
		param.CertFileContent = lib.PString("invalid")
		m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
		err := m.CreateCertificate(ctx, param)
		require.Error(t, err)
	})

	t.Run("same cert and key file name", func(t *testing.T) {
		param := validParam()
		param.KeyFileName = lib.PString("cert.pem")
		m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
		err := m.CreateCertificate(ctx, param)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Certificate File Name cert.pem Existed")
	})
}

func TestCertificateManager_UpdateAsDefaultCertificate(t *testing.T) {
	ctx := context.Background()

	t.Run("already default", func(t *testing.T) {
		m := NewCertificateManager(&fakeTxn{}, &fakeCertificateStorager{}, nil, &fakeExtraFileStorager{})
		require.NoError(t, m.UpdateAsDefaultCertificate(ctx, &Certificate{IsDefault: true}))
	})

	t.Run("success", func(t *testing.T) {
		old := &Certificate{CertName: "old", IsDefault: true}
		target := &Certificate{CertName: "new"}
		updated := []string{}
		store := &fakeCertificateStorager{
			fetchCertificatesFn: func(ctx context.Context, filter *CertificateFilter) ([]*Certificate, error) {
				return []*Certificate{old}, nil
			},
			updateCertificateFn: func(ctx context.Context, cert *Certificate, param *CertificateParam) error {
				updated = append(updated, cert.CertName)
				return nil
			},
		}
		m := NewCertificateManager(&fakeTxn{}, store, nil, &fakeExtraFileStorager{})
		require.NoError(t, m.UpdateAsDefaultCertificate(ctx, target))
		assert.Equal(t, []string{"old", "new"}, updated)
	})
}

func TestServerCertConf_UpdateVersion(t *testing.T) {
	scc := &ServerCertConf{}
	scc.Config.CertConf = map[string]server_cert_conf.ServerCertConf{
		"default": {
			ServerCertFile: "tls_conf/p1/cert.pem",
			ServerKeyFile:  "tls_conf/p1/key.pem",
		},
	}

	require.NoError(t, scc.UpdateVersion("v1"))
	assert.Equal(t, "v1", scc.Version)
	assert.Equal(t, "tls_conf_v1/p1/cert.pem", scc.Config.CertConf["default"].ServerCertFile)
	assert.Equal(t, "tls_conf_v1/p1/key.pem", scc.Config.CertConf["default"].ServerKeyFile)
}

func TestServerCertConf_UpdateVersion_badPath(t *testing.T) {
	scc := &ServerCertConf{}
	scc.Config.CertConf = map[string]server_cert_conf.ServerCertConf{
		"default": {
			ServerCertFile: "cert.pem",
		},
	}

	err := scc.UpdateVersion("v1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ServerCertFile must has /")
}
