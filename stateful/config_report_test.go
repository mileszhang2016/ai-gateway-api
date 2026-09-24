// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package stateful

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportConfig_ApplyDefaults(t *testing.T) {
	cfg := &ReportConfig{}
	cfg.applyDefaults()

	assert.Equal(t, 60, cfg.AggregateIntervalSec)
	assert.Equal(t, 7, cfg.RetentionDays)
	assert.Equal(t, "", cfg.Backend)
	assert.Equal(t, "", cfg.Datasource)
	assert.False(t, cfg.EnableAggregateJob)
	assert.False(t, cfg.EnablePartitionMgmt)
}

func TestReportConfig_ApplyDefaults_KeepsCustomValues(t *testing.T) {
	cfg := &ReportConfig{
		Backend:              "mysql",
		Datasource:           "report_db",
		Database:             "bfe_report",
		EnableAggregateJob:   true,
		AggregateIntervalSec: 30,
		RetentionDays:        14,
		EnablePartitionMgmt:  true,
	}
	cfg.applyDefaults()

	assert.Equal(t, 30, cfg.AggregateIntervalSec)
	assert.Equal(t, 14, cfg.RetentionDays)
}
