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
	"strings"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib/xerror"
	"github.com/rainway-ai-gateway/ai-gateway-api/storage/rdb/internal/dao/internal"
)

const tAPIKeyIDSeqTableName = "api_key_id_seq"

// TAPIKeyIDSeq maps to api_key_id_seq table.
type TAPIKeyIDSeq struct {
	ProductName string `db:"product_name"`
	NextSeq     int64  `db:"next_seq"`
}

// TAPIKeyIDSeqParam is used for query/update conditions.
type TAPIKeyIDSeqParam struct {
	ProductName *string `db:"product_name"`
	NextSeq     *int64  `db:"next_seq"`
}

// TAPIKeyIDSeqOne returns the sequence row for a product, or nil if not exists.
func TAPIKeyIDSeqOne(dbCtx lib.DBContexter, where *TAPIKeyIDSeqParam) (*TAPIKeyIDSeq, error) {
	t := &TAPIKeyIDSeq{}
	err := internal.QueryOne(dbCtx, tAPIKeyIDSeqTableName, where, t)
	if err == nil {
		return t, nil
	}
	if xerror.Cause(err) == internal.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

// TAPIKeyIDSeqCreate inserts a new sequence row.
func TAPIKeyIDSeqCreate(dbCtx lib.DBContexter, data *TAPIKeyIDSeqParam) (int64, error) {
	return internal.Create(dbCtx, tAPIKeyIDSeqTableName, data)
}

// TAPIKeyIDSeqUpdate updates a sequence row.
func TAPIKeyIDSeqUpdate(dbCtx lib.DBContexter, val, where *TAPIKeyIDSeqParam) (int64, error) {
	return internal.Update(dbCtx, tAPIKeyIDSeqTableName, where, val)
}

// TAPIKeyIDSeqAllocate atomically allocates and returns the next available
// API-Key sequence number for the given product. The returned number can be
// used directly to build an ID such as "api-key-{seq}".
//
// It uses a compare-and-set update inside a short DB transaction. If a
// concurrent caller wins the race, the function re-reads the current value and
// retries. Retryable DB errors (e.g. SQLite busy/locked) are retried from the
// outer loop.
func TAPIKeyIDSeqAllocate(dbCtx lib.DBContexter, productName string) (int64, error) {
	conn := dbCtx.Conn()
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		seq, err := allocateAPIKeyIDSeqOnce(conn, productName)
		if err == nil {
			return seq, nil
		}
		if !isRetryableSeqError(err) || i == maxRetries-1 {
			return 0, err
		}
		time.Sleep(time.Duration(i+1) * 5 * time.Millisecond)
	}
	return 0, fmt.Errorf("allocate api-key id seq for product %q exceeded max retries", productName)
}

func allocateAPIKeyIDSeqOnce(conn *sql.DB, productName string) (int64, error) {
	var allocated int64
	err := lib.Transaction(conn, func(tx *sql.Tx) error {
		for {
			row := tx.QueryRowContext(context.Background(),
				"SELECT next_seq FROM api_key_id_seq WHERE product_name = ?", productName)
			var current int64
			if err := row.Scan(&current); err != nil {
				if err == sql.ErrNoRows {
					// First allocation for this product: initialize with next=2 and return 1.
					_, insertErr := tx.ExecContext(context.Background(),
						"INSERT INTO api_key_id_seq (product_name, next_seq) VALUES (?, ?)",
						productName, 2)
					if insertErr != nil {
						if lib.DuplicateEntryError(insertErr) {
							// Another transaction created the row; retry within the same tx.
							continue
						}
						return insertErr
					}
					allocated = 1
					return nil
				}
				return err
			}

			// CAS update: only succeed if next_seq has not changed since we read it.
			res, updateErr := tx.ExecContext(context.Background(),
				"UPDATE api_key_id_seq SET next_seq = next_seq + 1 WHERE product_name = ? AND next_seq = ?",
				productName, current)
			if updateErr != nil {
				return updateErr
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				// Concurrent update happened; retry within the same tx.
				continue
			}
			// current is the sequence number allocated to this request;
			// next_seq now points to the next available number.
			allocated = current
			return nil
		}
	})
	return allocated, err
}

// isRetryableSeqError reports whether the allocation error is transient and
// worth retrying (e.g. SQLite busy/locked, MySQL deadlock).
func isRetryableSeqError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "busy") ||
		strings.Contains(msg, "error 1213") ||
		strings.Contains(msg, "deadlock")
}
