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

package cancel_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	helper "github.com/rainway-ai-gateway/ai-gateway-api/integration/tests/batch/helper"
	"github.com/rainway-ai-gateway/ai-gateway-api/integration/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBatchCancel_ConcurrentPreemptMySQL（BT-3-007）在 MySQL 后端下并发
// cancel 同一 in_progress 任务，验证 DB 条件更新抢占
// （UPDATE ... WHERE batch_id=? AND product_name=? AND status NOT IN (终态)）
// 在真实 MySQL 行锁/原子语义下恰好一个受理、其余 409。
//
// 需要真实 MySQL：DSN 取自 AIAPI_MYSQL_DSN（形如
// root:pass@tcp(127.0.0.1:3306)/）。SQLite 单写者串行无法暴露该行锁
// 竞态，本用例以 //go:build mysql 落盘（家族10，仿 traffic_mirror）。
//
// 审计语义说明：实现按"每次 cancel 尝试"记录审计（成功 status=1，
// 冲突失败 status=2），故断言为"恰好一条成功审计 + 其余为失败审计"，
// 而非仅一条审计（design.md §4.3 记录）。
func TestBatchCancel_ConcurrentPreemptMySQL(t *testing.T) {
	adminDSN := os.Getenv("AIAPI_MYSQL_DSN")
	if adminDSN == "" {
		t.Skip("AIAPI_MYSQL_DSN not set, skipping MySQL concurrency test")
	}

	sm, err := testutil.StartServerWithMySQL(adminDSN)
	require.NoError(t, err, "start MySQL-backed server")
	defer sm.Shutdown()

	client := &testutil.Client{BaseURL: sm.ServerURL, HTTPClient: &http.Client{Timeout: 60 * time.Second}}
	db := helper.OpenDB(t, sm)

	tag := helper.Unique("bt-3-007")
	provider := tag + "-provider"

	// 本地假 provider：受理方出网打回本测试进程（唯一期望的一次）。
	var upstreamHits int
	var upstreamMu sync.Mutex
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamMu.Lock()
		upstreamHits++
		upstreamMu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	_, upstreamPort, err := net.SplitHostPort(upstream.URL[len("http://"):])
	require.NoError(t, err)

	helper.InsertProvider(t, db, provider,
		`[{"addr":"127.0.0.1","port":`+upstreamPort+`,"weight":100}]`,
		`[{"name":"key-primary","key":"sk-test"}]`)
	defer helper.DeleteProvider(t, db, provider)

	helper.InsertBatchTask(t, db, helper.BatchTaskRow{
		BatchID: tag, APIKeyID: tag + "-ak", ProductName: helper.TestProductName,
		Provider: provider, KeyName: "key-primary", Status: "in_progress", IdemKey: tag + "-idem",
	})
	defer helper.DeleteBatchTask(t, db, tag)

	const workers = 20
	var wg sync.WaitGroup
	results := make(chan int, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Post("/open-api/v1/batches/"+tag+"/cancel", map[string]interface{}{})
			if err != nil {
				results <- -1
				return
			}
			results <- resp.ErrNum
		}()
	}
	wg.Wait()
	close(results)

	okCount, conflictCount, other := 0, 0, map[int]int{}
	for code := range results {
		switch code {
		case 200:
			okCount++
		case 409:
			conflictCount++
		default:
			other[code]++
		}
	}
	assert.Equal(t, 1, okCount, "exactly one cancel must win the preempt")
	assert.Equal(t, workers-1, conflictCount, "all losers must get 409")
	assert.Empty(t, other, "no unexpected error codes")

	// 受理方恰好出网一次（幂等释放/重试不得重复出网）。
	upstreamMu.Lock()
	assert.Equal(t, 1, upstreamHits, "upstream must receive exactly one cancel POST")
	upstreamMu.Unlock()

	// 审计：恰好一条成功审计；其余尝试为失败审计（实现按尝试记账）。
	var successAudits, totalAudits int
	deadline := time.Now().Add(15 * time.Second)
	for {
		err := db.QueryRow(`
			SELECT
				COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0),
				COUNT(*)
			FROM operation_logs WHERE resource_type = 'batch_task' AND resource_id = ?
		`, tag).Scan(&successAudits, &totalAudits)
		require.NoError(t, err)
		if totalAudits >= workers || time.Now().After(deadline) {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	assert.Equal(t, 1, successAudits, "exactly one success audit expected")
	assert.Equal(t, workers, totalAudits,
		"implementation records one audit per cancel attempt (success + conflicts)")

	// 终态校验：任务保持 cancelling（受理中），由状态推进对账收敛。
	var status string
	require.NoError(t, db.QueryRow("SELECT status FROM batch_tasks WHERE batch_id = ?", tag).Scan(&status))
	assert.Equal(t, "cancelling", status)
}
