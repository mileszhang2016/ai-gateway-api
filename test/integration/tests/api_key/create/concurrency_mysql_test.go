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

//go:build mysql

package api_key_test

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKey_CreateConcurrentMySQL 在 MySQL 后端下并发创建 API-Key（不带 id，
// 走自动序号分配路径），回归 issue #80（非原子 max+1 生成导致 422 Duplicate
// id / 500）与 issue #99（CAS 序号分配在 RR 隔离级下死循环抢锁直至
// Lock wait timeout 50s 后 500）。
//
// 需要真实 MySQL：SQLite 单连接串行无法暴露竞态，故以 //go:build mysql
// 隔离，DSN 取自 AIAPI_MYSQL_DSN（形如 root:pass@tcp(127.0.0.1:3306)/）。
func TestAPIKey_CreateConcurrentMySQL(t *testing.T) {
	adminDSN := os.Getenv("AIAPI_MYSQL_DSN")
	if adminDSN == "" {
		t.Skip("AIAPI_MYSQL_DSN not set, skipping MySQL concurrency test")
	}

	sm, err := testutil.StartServerWithMySQL(adminDSN)
	require.NoError(t, err, "start MySQL-backed server")
	defer sm.Shutdown()

	client := &testutil.Client{BaseURL: sm.ServerURL, HTTPClient: &http.Client{Timeout: 60 * time.Second}}

	const workers = 50
	const perWorker = 14 // 700 个，对齐 issue #80 批量场景
	const total = workers * perWorker

	start := time.Now()
	ids := make(chan string, total)
	errs := make(chan error, total)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				resp, err := client.Post("/open-api/v1/api-keys", map[string]interface{}{
					"description": fmt.Sprintf("conc-%d-%d", w, i),
				})
				if err != nil {
					errs <- fmt.Errorf("request error: %w", err)
					continue
				}
				if resp.ErrNum != 200 {
					// 预修复形态：422 Duplicate id（查重窗口竞态）或
					// 500（唯一索引冲突/锁等待超时未捕获）
					errs <- fmt.Errorf("ErrNum=%d ErrMsg=%s", resp.ErrNum, resp.ErrMsg)
					continue
				}
				id, err := testutil.GetDataField(resp, "id")
				if err != nil {
					errs <- fmt.Errorf("missing id: %w", err)
					continue
				}
				ids <- id.(string)
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	close(ids)

	elapsed := time.Since(start)
	t.Logf("created %d api-keys in %v", total, elapsed)

	for err := range errs {
		t.Errorf("concurrent create failed: %v", err)
	}

	seen := make(map[string]bool, total)
	for id := range ids {
		assert.False(t, seen[id], "duplicate id %s", id)
		seen[id] = true
	}
	assert.Len(t, seen, total, "all concurrent creates must succeed with unique ids")
	assert.Less(t, elapsed, 90*time.Second, "lock wait regression (#99) would take ~50s per conflicting request")
}
