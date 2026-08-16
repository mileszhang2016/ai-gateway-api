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

package imods

import (
	"context"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModBodyProcessManager(t *testing.T) {
	m := NewModBodyProcessManager(iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}))
	assert.NotNil(t, m)
	assert.NotNil(t, m.versionControlManager)
}

func TestModBodyProcessManager_bodyProcessGenerator(t *testing.T) {
	setupState()
	ctx := context.Background()

	m := NewModBodyProcessManager(iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}))
	data, err := m.bodyProcessGenerator(ctx)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, ConfigTopicProductBodyProcess, data.Topic)

	conf, ok := data.DataWithoutVersion.(*ModBodyProcessConf)
	require.True(t, ok)
	assert.NotNil(t, conf.Version)
	assert.Equal(t, iversion_control.ZeroVersion, *conf.Version)
	assert.Contains(t, conf.Config, "AI_product")
	assert.Empty(t, conf.Config["AI_product"])
}

func TestModBodyProcessManager_ConfigExport(t *testing.T) {
	setupState()
	ctx := context.Background()

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "v-body-1", nil
		},
	}

	m := NewModBodyProcessManager(iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore))
	conf, err := m.ConfigExport(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, conf)
	assert.Equal(t, "v-body-1", *conf.Version)

	// Same version returns nil
	conf2, err := m.ConfigExport(ctx, "v-body-1")
	require.NoError(t, err)
	assert.Nil(t, conf2)
}

func TestModBodyProcessManager_ConfigExport_ExportError(t *testing.T) {
	setupState()
	ctx := context.Background()

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "", errors.New("upsert failed")
		},
	}

	m := NewModBodyProcessManager(iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore))
	conf, err := m.ConfigExport(ctx, "")
	require.Error(t, err)
	assert.Nil(t, conf)
}

func TestModBodyProcessConf_UpdateVersion(t *testing.T) {
	conf := &ModBodyProcessConf{}
	err := conf.UpdateVersion("v1")
	require.NoError(t, err)
	assert.Equal(t, "v1", *conf.Version)
}
