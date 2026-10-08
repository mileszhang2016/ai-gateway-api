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
//limitations under the License. All rights reserved.

package ibatch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/iprovider"
)

// setupAdvanceFixture 组装状态推进夹具：provider 侧由 handler 模拟，
// DB 非终态任务列表由 tasks 给出。
func setupAdvanceFixture(t *testing.T, tasks []*BatchTask, handler http.HandlerFunc) (*Manager, *fakeBatchStorager, *httptest.Server) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	storager := &fakeBatchStorager{
		listTasksFn: func(ctx context.Context, filter *BatchTaskFilter) ([]*BatchTask, error) {
			assert.Equal(t, TerminalStatuses, filter.StatusNotIn)
			return tasks, nil
		},
		updateProgressFn: func(ctx context.Context, id int64, patch *BatchTaskProgressPatch) (int64, error) {
			return 1, nil
		},
		upsertFileFn: func(ctx context.Context, file *BatchFile) (int64, error) {
			return 1, nil
		},
	}
	addr, port := splitAddrPort(t, server.URL)
	providers := &fakeProviderQuerier{
		fetchFn: func(ctx context.Context, filter *iprovider.ProviderFilter) (*iprovider.Provider, error) {
			return &iprovider.Provider{
				Name:          "openai",
				Keys:          []iprovider.ProviderKey{{Name: "k1", Key: "sk"}},
				InstancePool:  []iprovider.ProviderInstance{{Addr: addr, Port: port}},
				ProtocolPaths: map[string]string{"openai": "/v1"},
			}, nil
		},
	}
	m := NewManager(storager, &fakeRedisClient{},
		&fakeAPIKeyQuerier{}, providers, &fakePriceQuerier{},
		func() *http.Client { return server.Client() },
		&fakeClock{t: time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)})
	return m, storager, server
}

func splitAddrPort(t *testing.T, rawURL string) (string, int) {
	t.Helper()
	addr := rawURL[len("http://"):]
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], atoiOrZero(addr[i+1:])
		}
	}
	return addr, 80
}

func TestAdvanceStatus_StatusForward(t *testing.T) {
	tasks := []*BatchTask{{
		ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai", KeyName: "k1",
		Status: TaskStatusInProgress, CreatedAt: time.Now().Add(-time.Hour),
	}}

	var gotPath, gotAuth string
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"b-1","status":"completed","output_file_id":"file-out","request_counts":{"completed":10,"failed":1,"total":11}}`)
	})

	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, advanced)
	assert.Equal(t, "/v1/batches/b-1", gotPath)
	assert.Equal(t, "Bearer sk", gotAuth)

	require.Len(t, storager.progressPatches, 1)
	patch := storager.progressPatches[0]
	require.NotNil(t, patch.Status)
	assert.Equal(t, TaskStatusCompleted, *patch.Status)
	require.NotNil(t, patch.OutputFileID)
	assert.Equal(t, "file-out", *patch.OutputFileID)
	require.NotNil(t, patch.RequestCounts)
	assert.Equal(t, int64(10), (*patch.RequestCounts)["completed"])
	require.NotNil(t, patch.TerminalAt, "进入终态记 terminal_at")

	// 输出文件补登记。
	require.Len(t, storager.upsertedFiles, 1)
	assert.Equal(t, "file-out", storager.upsertedFiles[0].FileID)
	assert.Equal(t, FileDirectionOutput, storager.upsertedFiles[0].Direction)
	assert.Equal(t, "p1", storager.upsertedFiles[0].ProductName)
}

func TestAdvanceStatus_NoChange(t *testing.T) {
	tasks := []*BatchTask{{
		ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai",
		Status:        TaskStatusInProgress,
		RequestCounts: map[string]int64{"completed": 1},
	}}
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"b-1","status":"in_progress","request_counts":{"completed":1}}`)
	})

	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, advanced)
	assert.Empty(t, storager.progressPatches)
}

func TestAdvanceStatus_ProviderAbsentExpired(t *testing.T) {
	old := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC) // > 24h 前
	tasks := []*BatchTask{{
		ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai",
		Status: TaskStatusQueued, CreatedAt: old,
	}}
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, advanced)
	require.Len(t, storager.progressPatches, 1)
	assert.Equal(t, TaskStatusExpired, *storager.progressPatches[0].Status)
	require.NotNil(t, storager.progressPatches[0].TerminalAt)
}

func TestAdvanceStatus_ProviderAbsentWithinGrace(t *testing.T) {
	recent := time.Date(2026, 10, 7, 11, 0, 0, 0, time.UTC) // 1h 前
	tasks := []*BatchTask{{
		ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai",
		Status: TaskStatusQueued, CreatedAt: recent,
	}}
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, advanced)
	assert.Empty(t, storager.progressPatches)
}

func TestAdvanceStatus_ProviderErrorSkipped(t *testing.T) {
	tasks := []*BatchTask{
		{ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai", Status: TaskStatusQueued, CreatedAt: time.Now()},
		{ID: 2, BatchID: "b-2", ProductName: "p1", Provider: "openai", Status: TaskStatusQueued, CreatedAt: time.Now()},
	}
	calls := 0
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		fmt.Fprint(w, `{"id":"b-2","status":"failed"}`)
	})

	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, advanced, "单任务失败不影响其他任务")
	require.Len(t, storager.progressPatches, 1)
	assert.Equal(t, TaskStatusFailed, *storager.progressPatches[0].Status)
}

func TestAdvanceStatus_BadJSONSkipped(t *testing.T) {
	tasks := []*BatchTask{{ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai", Status: TaskStatusQueued}}
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not-json`)
	})
	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, advanced)
	assert.Empty(t, storager.progressPatches)
}

func TestAdvanceStatus_ListError(t *testing.T) {
	storager := &fakeBatchStorager{
		listTasksFn: func(ctx context.Context, filter *BatchTaskFilter) ([]*BatchTask, error) {
			return nil, errors.New("db down")
		},
	}
	m := NewManager(storager, nil, &fakeAPIKeyQuerier{}, &fakeProviderQuerier{}, &fakePriceQuerier{}, nil, &fakeClock{})
	_, err := m.AdvanceStatus(context.Background())
	require.Error(t, err)
}

func TestAdvanceStatus_NoProviderQuerier(t *testing.T) {
	m := NewManager(&fakeBatchStorager{}, nil, &fakeAPIKeyQuerier{}, nil, &fakePriceQuerier{}, nil, &fakeClock{})
	n, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestAdvanceStatus_MismatchedID(t *testing.T) {
	tasks := []*BatchTask{{ID: 1, BatchID: "b-1", ProductName: "p1", Provider: "openai", Status: TaskStatusQueued}}
	m, storager, _ := setupAdvanceFixture(t, tasks, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":"other","status":"completed"}`)
	})
	advanced, err := m.AdvanceStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, advanced)
	assert.Empty(t, storager.progressPatches)
}
