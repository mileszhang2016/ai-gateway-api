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

package iversion_control

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTxn struct{}

func (f *fakeTxn) AtomExecute(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeVersionValuable struct {
	version   string
	updateErr error
}

func (f *fakeVersionValuable) UpdateVersion(version string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.version = version
	return nil
}

// badVersionValuable implements VersionValuable but cannot be JSON-marshaled,
// which forces Sign to return an error.
type badVersionValuable struct {
	Ch chan int
}

func (b *badVersionValuable) UpdateVersion(version string) error {
	return nil
}

type fakeVersionControlStorager struct {
	upsertFn func(ctx context.Context, css *ExportData) (string, error)
}

func (f *fakeVersionControlStorager) UpsertConfigLastExportedVersion(ctx context.Context, css *ExportData) (string, error) {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, css)
	}
	return "", nil
}

func TestVersion(t *testing.T) {
	got := Version(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))
	assert.Equal(t, "20240102030405", got)
}

func TestSign(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := Sign(map[string]interface{}{"a": 1})
		require.NoError(t, err)
		assert.Len(t, got, 32)
	})

	t.Run("marshal error", func(t *testing.T) {
		_, err := Sign(make(chan int))
		require.Error(t, err)
	})
}

func TestExportData_Version(t *testing.T) {
	ed := &ExportData{version: "v1"}
	assert.Equal(t, "v1", ed.Version())
}

func TestExportData_CalculateVersion(t *testing.T) {
	before := time.Now()
	ed := &ExportData{}
	got, err := ed.CalculateVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Equal(t, got, ed.version)
	// version should be current time formatted
	parsed, err := time.Parse("20060102150405", got)
	require.NoError(t, err)
	assert.True(t, parsed.After(before) || parsed.Equal(before))
}

func TestNewVersionControllerManager(t *testing.T) {
	storager := &fakeVersionControlStorager{}
	txn := &fakeTxn{}
	m := NewVersionControllerManager(txn, storager)
	require.NotNil(t, m)
	assert.Equal(t, storager, m.storager)
	assert.Equal(t, txn, m.txn)
}

func TestVersionControlManager_ExportConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		txn := &fakeTxn{}
		storager := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *ExportData) (string, error) {
				assert.Equal(t, "test-topic", css.Topic)
				assert.NotEmpty(t, css.DataSignWithoutVersion)
				// version zero before upsert, real version after
				assert.Equal(t, ZeroVersion, css.DataWithoutVersion.(*fakeVersionValuable).version)
				return "20240102000000", nil
			},
		}
		m := NewVersionControllerManager(txn, storager)

		data := &fakeVersionValuable{}
		generator := func(ctx context.Context) (*ExportData, error) {
			return &ExportData{
				Topic:              "test-topic",
				DataWithoutVersion: data,
			}, nil
		}

		result, err := m.ExportConfig(ctx, "test-topic", generator)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "20240102000000", result.version)
		assert.Equal(t, "20240102000000", data.version)
	})

	t.Run("generator error", func(t *testing.T) {
		m := NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{})
		generator := func(ctx context.Context) (*ExportData, error) {
			return nil, errors.New("gen failed")
		}

		_, err := m.ExportConfig(ctx, "test-topic", generator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gen failed")
	})

	t.Run("update zero version error", func(t *testing.T) {
		m := NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{})
		generator := func(ctx context.Context) (*ExportData, error) {
			return &ExportData{
				Topic:              "test-topic",
				DataWithoutVersion: &fakeVersionValuable{updateErr: errors.New("update zero failed")},
			}, nil
		}

		_, err := m.ExportConfig(ctx, "test-topic", generator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update zero failed")
	})

	t.Run("sign error", func(t *testing.T) {
		m := NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{})
		_, err := m.ExportConfig(ctx, "test-topic", func(ctx context.Context) (*ExportData, error) {
			return &ExportData{
				Topic:              "test-topic",
				DataWithoutVersion: &badVersionValuable{Ch: make(chan int)},
			}, nil
		})
		require.Error(t, err)
	})

	t.Run("storager error", func(t *testing.T) {
		storager := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *ExportData) (string, error) {
				return "", errors.New("upsert failed")
			},
		}
		m := NewVersionControllerManager(&fakeTxn{}, storager)
		generator := func(ctx context.Context) (*ExportData, error) {
			return &ExportData{
				Topic:              "test-topic",
				DataWithoutVersion: &fakeVersionValuable{},
			}, nil
		}

		_, err := m.ExportConfig(ctx, "test-topic", generator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upsert failed")
	})

	t.Run("update real version error", func(t *testing.T) {
		storager := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *ExportData) (string, error) {
				return "20240102000000", nil
			},
		}
		m := NewVersionControllerManager(&fakeTxn{}, storager)
		generator := func(ctx context.Context) (*ExportData, error) {
			return &ExportData{
				Topic:              "test-topic",
				DataWithoutVersion: &fakeVersionValuable{updateErr: errors.New("update real failed")},
			}, nil
		}

		_, err := m.ExportConfig(ctx, "test-topic", generator)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update real failed")
	})
}
