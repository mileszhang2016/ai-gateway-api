// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package stateful

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsAreNotNil(t *testing.T) {
	assert.NotNil(t, MetricAPIAccessCounter)
	assert.NotNil(t, MetricAPICostHisCounter)
	assert.NotNil(t, MetricSQLAccessCounter)
	assert.NotNil(t, MetricSQLCostCounter)
	assert.NotNil(t, MetricPaincCounter)
}

func TestMetricAPIAccessCounterIncrement(t *testing.T) {
	MetricAPIAccessCounter.WithLabelValues("/test", "200", "GET").Inc()

	value, err := getCounterValue("api_access", map[string]string{
		"pattern":     "/test",
		"status_code": "200",
		"method":      "GET",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, float64(1))
}

func TestMetricAPICostHisCounterIncrement(t *testing.T) {
	MetricAPICostHisCounter.WithLabelValues("/test", "200", "GET").Inc()

	value, err := getCounterValue("api_cost", map[string]string{
		"pattern":     "/test",
		"status_code": "200",
		"method":      "GET",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, float64(1))
}

func TestMetricSQLAccessCounterIncrement(t *testing.T) {
	MetricSQLAccessCounter.WithLabelValues("SELECT * FROM test").Inc()

	value, err := getCounterValue("sql_access", map[string]string{"sql": "SELECT * FROM test"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, float64(1))
}

func TestMetricSQLCostCounterIncrement(t *testing.T) {
	MetricSQLCostCounter.WithLabelValues("SELECT * FROM test").Inc()

	value, err := getCounterValue("sql_cost", map[string]string{"sql": "SELECT * FROM test"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, float64(1))
}

func TestMetricPanicCounterIncrement(t *testing.T) {
	MetricPaincCounter.Inc()

	value, err := getCounterValue("panic", nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, float64(1))
}

func TestMetricAPIAccessCounterWrongLabelCount(t *testing.T) {
	_, err := MetricAPIAccessCounter.GetMetricWithLabelValues("/test", "200")
	assert.Error(t, err)
}

func getCounterValue(name string, labels map[string]string) (float64, error) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0, err
	}
	for _, f := range families {
		if f.GetName() != name {
			continue
		}
		for _, m := range f.GetMetric() {
			matched := len(labels)
			for _, lp := range m.GetLabel() {
				if labels[lp.GetName()] == lp.GetValue() {
					matched--
				}
			}
			if matched == 0 {
				return m.GetCounter().GetValue(), nil
			}
		}
	}
	return 0, assert.AnError
}
