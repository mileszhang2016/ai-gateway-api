// Copyright(c) 2026 Beijing Yingfei Networks Technology Co.Ltd.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http: //www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package imods

import (
	"context"
	"fmt"

	"github.com/yf-networks/ai-gateway-api/model/iversion_control"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

// ConfigTopicProductBodyProcess is the configuration topic for mod_body_process
const ConfigTopicProductBodyProcess = "mod_body_process"

// ModBodyProcessConf defines the export configuration structure for mod_body_process
type ModBodyProcessConf struct {
	Version *string             `json:"Version"`
	Config  map[string][]string `json:"Config"`
}

// UpdateVersion updates the configuration version
func (conf *ModBodyProcessConf) UpdateVersion(version string) error {
	conf.Version = &version
	return nil
}

// ModBodyProcessManager manages the body process configuration export
type ModBodyProcessManager struct {
	versionControlManager *iversion_control.VersionControlManager
}

// NewModBodyProcessManager creates a new ModBodyProcessManager
func NewModBodyProcessManager(versionControlManager *iversion_control.VersionControlManager) *ModBodyProcessManager {
	return &ModBodyProcessManager{
		versionControlManager: versionControlManager,
	}
}

// bodyProcessGenerator generates the body process configuration for export
func (m *ModBodyProcessManager) bodyProcessGenerator(ctx context.Context) (*iversion_control.ExportData, error) {
	productName := stateful.DefaultConfig.RunTime.AIRouteInnerProductName

	conf := &ModBodyProcessConf{
		Config: map[string][]string{
			productName: {},
		},
	}
	conf.UpdateVersion(iversion_control.ZeroVersion)

	return &iversion_control.ExportData{
		Topic:              ConfigTopicProductBodyProcess,
		DataWithoutVersion: conf,
	}, nil
}

// ConfigExport exports body process configuration for BFE
func (m *ModBodyProcessManager) ConfigExport(ctx context.Context, lastVersion string) (*ModBodyProcessConf, error) {
	rst, err := m.versionControlManager.ExportConfig(ctx, ConfigTopicProductBodyProcess, m.bodyProcessGenerator)
	if err != nil {
		return nil, err
	}

	if rst.DataWithoutVersion == nil {
		return nil, fmt.Errorf("bodyProcessGenerator.DataWithoutVersion is nil")
	}

	conf, ok := rst.DataWithoutVersion.(*ModBodyProcessConf)
	if ok {
		if *conf.Version == lastVersion {
			return nil, nil
		}
		return conf, nil
	}

	return nil, fmt.Errorf("convert bodyProcessGenerator.DataWithoutVersion to ModBodyProcessConf is error")
}