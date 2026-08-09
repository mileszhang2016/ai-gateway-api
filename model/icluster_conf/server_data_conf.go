// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package icluster_conf

import (
	"github.com/bfenetworks/bfe/bfe_config/bfe_cluster_conf/cluster_conf"
)

// ServerDataAIKey is the exported AI key item for BFE server_data_conf.
// It mirrors the target BFE AIKey structure until the bfe module is updated.
type ServerDataAIKey struct {
	Name   string `json:"Name"`
	Key    string `json:"Key"`
	Weight int    `json:"Weight"`
}

// ServerDataAIKeyPolicy is the exported key routing policy for BFE server_data_conf.
// It mirrors the target BFE AIKeyPolicy structure until the bfe module is updated.
type ServerDataAIKeyPolicy struct {
	Strategy            string `json:"Strategy"`
	MaxRetries          int    `json:"MaxRetries"`
	RetryBackoffInitial int    `json:"RetryBackoffInitial"`
	RetryBackoffMax     int    `json:"RetryBackoffMax"`
}

// ServerDataAIConf is the exported AI conf for BFE server_data_conf.
// It extends the BFE cluster_conf.AIConf with multi API-Key support (Keys + KeyPolicy).
type ServerDataAIConf struct {
	Type         int                `json:"Type"`
	ModelMapping *map[string]string `json:"ModelMapping"`
	Keys         []ServerDataAIKey  `json:"Keys"`
	KeyPolicy    *ServerDataAIKeyPolicy `json:"KeyPolicy"`
}

// ServerDataClusterConf is the exported cluster conf for BFE server_data_conf.
// It reuses BFE sub-configs and only overrides AIConf to support multi API-Key.
type ServerDataClusterConf struct {
	BackendConf  *cluster_conf.BackendBasic     `json:"BackendConf"`
	CheckConf    *cluster_conf.BackendCheck     `json:"CheckConf"`
	GslbBasic    *cluster_conf.GslbBasicConf    `json:"GslbBasic"`
	ClusterBasic *cluster_conf.ClusterBasicConf `json:"ClusterBasic"`
	HTTPSConf    *cluster_conf.BackendHTTPS     `json:"HTTPSConf"`
	AIConf       *ServerDataAIConf              `json:"AIConf"`
}

// ServerDataBfeClusterConf is the exported cluster conf collection for BFE server_data_conf.
// It reuses BFE sub-configs and only overrides AIConf to support multi API-Key.
type ServerDataBfeClusterConf struct {
	Version *string                          `json:"Version"`
	Config  *map[string]ServerDataClusterConf `json:"Config"`
}

// UpdateVersion implements model/iversion_control.VersionValuable.
func (c *ServerDataBfeClusterConf) UpdateVersion(version string) error {
	c.Version = &version
	return nil
}
