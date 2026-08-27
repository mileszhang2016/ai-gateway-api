//go:build mysql

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

package dao

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mysqlDSN builds a DSN for the local MySQL test instance described in
// environment/mysql-installation.md. The password is read from the local file
// created during installation.
func mysqlDSN(t *testing.T) string {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	passFile := filepath.Join(home, "mysql-data", "root-password.txt")
	passBytes, err := os.ReadFile(passFile)
	require.NoError(t, err)
	pass := strings.TrimSpace(string(passBytes))
	return fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local", pass)
}

func setupMySQLSeqTestDB(t *testing.T) (*sql.DB, lib.DBContexter) {
	dsn := mysqlDSN(t)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS api_key_id_seq (
		product_name VARCHAR(255) PRIMARY KEY,
		next_seq BIGINT NOT NULL DEFAULT 1
	)`)
	require.NoError(t, err)

	_, err = db.Exec("TRUNCATE TABLE api_key_id_seq")
	require.NoError(t, err)

	return db, lib.NewDBContext(context.Background(), db)
}

func TestTAPIKeyIDSeqAllocate_MySQL_Basic(t *testing.T) {
	_, dbCtx := setupMySQLSeqTestDB(t)

	seq, err := TAPIKeyIDSeqAllocate(dbCtx, "AI_product")
	require.NoError(t, err)
	assert.Equal(t, int64(1), seq)

	seq, err = TAPIKeyIDSeqAllocate(dbCtx, "AI_product")
	require.NoError(t, err)
	assert.Equal(t, int64(2), seq)

	seq, err = TAPIKeyIDSeqAllocate(dbCtx, "AI_product")
	require.NoError(t, err)
	assert.Equal(t, int64(3), seq)
}

func TestTAPIKeyIDSeqAllocate_MySQL_Concurrent(t *testing.T) {
	_, dbCtx := setupMySQLSeqTestDB(t)

	const workers = 100
	var wg sync.WaitGroup
	results := make(chan int64, workers)
	errs := make(chan error, workers)
	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seq, err := TAPIKeyIDSeqAllocate(dbCtx, "AI_product")
			if err != nil {
				errs <- err
				return
			}
			results <- seq
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	t.Logf("allocated %d sequences in %v", workers, time.Since(start))

	for err := range errs {
		t.Fatalf("allocation failed: %v", err)
	}

	seen := make(map[int64]struct{})
	for seq := range results {
		assert.Greater(t, seq, int64(0))
		assert.NotContains(t, seen, seq, "sequence %d was allocated more than once", seq)
		seen[seq] = struct{}{}
	}
	assert.Len(t, seen, workers)
}
