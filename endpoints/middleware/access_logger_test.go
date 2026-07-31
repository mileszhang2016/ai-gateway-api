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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yf-networks/ai-gateway-api/lib/xreq"
	"github.com/yf-networks/ai-gateway-api/stateful"
)

func TestGetClientIp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Clientip", "10.0.0.1")
	assert.Equal(t, "10.0.0.1", GetClientIp(req))

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", GetClientIp(req))
}

func TestUpdateMonitor(t *testing.T) {
	labels := prometheus.Labels{
		"pattern":     "/products",
		"method":      "get",
		"status_code": "200",
	}
	before := testutil.ToFloat64(stateful.MetricAPIAccessCounter.With(labels))

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	requestInfo := &xreq.RequestInfo{
		URLPattern: "/products",
		Method:     http.MethodGet,
		StatusCode: 200,
		Duration:   5 * time.Millisecond,
	}
	UpdateMonitor(req, requestInfo)

	assert.Equal(t, before+1, testutil.ToFloat64(stateful.MetricAPIAccessCounter.With(labels)))
	assert.GreaterOrEqual(t, testutil.ToFloat64(stateful.MetricAPICostHisCounter.With(labels)), before)
}

func TestRecord(t *testing.T) {
	setupTestLoggers(t)

	// Should not panic for either success or failure status codes.
	Record(&xreq.RequestInfo{StatusCode: 200})
	Record(&xreq.RequestInfo{StatusCode: 500})
	Record(&xreq.RequestInfo{StatusCode: 404})
}

func TestLoggerMiddleWare_ServeHTTP(t *testing.T) {
	setupTestLoggers(t)

	lm := NewLoggerMiddleWare()
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	ctx, requestInfo := xreq.InitRequestInfo(req.Context(), req)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	nrw := negroni.NewResponseWriter(rec)

	called := false
	lm.ServeHTTP(nrw, req, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, requestInfo.StatusCode)
	assert.GreaterOrEqual(t, requestInfo.Duration.Nanoseconds(), int64(0))
}

func TestNewLoggerMiddleWare(t *testing.T) {
	lm := NewLoggerMiddleWare()
	require.NotNil(t, lm)
}
