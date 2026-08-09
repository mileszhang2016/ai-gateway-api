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

package internal

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/stateful"
)

type curdTestRecord struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type curdTestWhere struct {
	ID int64 `db:"id"`
}

type curdTestAssign struct {
	Name string `db:"name"`
}

func setupCurdTest(t *testing.T) (lib.DBContexter, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	oldConfig := stateful.DefaultConfig
	stateful.DefaultConfig = &stateful.Config{
		RunTime: stateful.RunTimeConfig{
			RecordSQL: false,
		},
	}

	dbCtx := lib.NewDBContext(context.Background(), db)
	return dbCtx, mock, func() {
		db.Close()
		stateful.DefaultConfig = oldConfig
	}
}

func TestQueryOne(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "foo")
	mock.ExpectQuery("SELECT \\* FROM users WHERE \\(id=\\?\\) LIMIT \\?,\\?").WithArgs(1, 0, 1).WillReturnRows(rows)

	var rst curdTestRecord
	err := QueryOne(dbCtx, "users", &curdTestWhere{ID: 1}, &rst)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rst.ID)
	assert.Equal(t, "foo", rst.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryOne_NotFound(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery("SELECT \\* FROM users WHERE \\(id=\\?\\) LIMIT \\?,\\?").WithArgs(99, 0, 1).WillReturnRows(rows)

	var rst curdTestRecord
	err := QueryOne(dbCtx, "users", &curdTestWhere{ID: 99}, &rst)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryList(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "foo").
		AddRow(2, "bar")
	mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(rows)

	var rst []*curdTestRecord
	err := QueryList(dbCtx, "users", &curdTestWhere{}, &rst)
	require.NoError(t, err)
	require.Len(t, rst, 2)
	assert.Equal(t, "foo", rst[0].Name)
	assert.Equal(t, "bar", rst[1].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	mock.ExpectExec("INSERT INTO users").WithArgs(int64(1), "foo").WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := Create(dbCtx, "users", &curdTestRecord{ID: 1, Name: "foo"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	mock.ExpectExec("UPDATE users").WithArgs("foo", 1).WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := Update(dbCtx, "users", &curdTestWhere{ID: 1}, &curdTestAssign{Name: "foo"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	dbCtx, mock, cleanup := setupCurdTest(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM users").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := Delete(dbCtx, "users", &curdTestWhere{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)
	require.NoError(t, mock.ExpectationsWereMet())
}
