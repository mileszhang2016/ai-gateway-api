// Copyright(c) 2026 The Infinity AI Gateway Authors.
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

package txn

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
)

func setupTxnTest(t *testing.T) (*RDBTxnStorager, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	factory := func(ctx context.Context, ops ...*lib.Op) (*lib.DBContext, error) {
		return lib.NewDBContext(ctx, db), nil
	}

	return NewRDBTxnStorager(factory), mock, func() {
		db.Close()
	}
}

func TestRDBTxnStorager_AtomExecute_Commit(t *testing.T) {
	storager, mock, cleanup := setupTxnTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit()

	executed := false
	err := storager.AtomExecute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRDBTxnStorager_AtomExecute_Rollback(t *testing.T) {
	storager, mock, cleanup := setupTxnTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("business error")
	err := storager.AtomExecute(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRDBTxnStorager_AtomExecute_FactoryError(t *testing.T) {
	expectedErr := errors.New("factory error")
	factory := func(ctx context.Context, ops ...*lib.Op) (*lib.DBContext, error) {
		return nil, expectedErr
	}

	storager := NewRDBTxnStorager(factory)
	err := storager.AtomExecute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
}
