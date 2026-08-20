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

	"github.com/stretchr/testify/assert"
)

func TestSQLRecord_format(t *testing.T) {
	t.Run("intercept insert values", func(t *testing.T) {
		sr := &SQLRecord{
			SQL: "INSERT INTO mod_header_rules (actions,cond) VALUES ('[{}]','cond')",
		}
		sr.format()
		assert.Equal(t, "INSERT INTO mod_header_rules (actions,cond)intercept", sr.SQL)
	})

	t.Run("replace in clause", func(t *testing.T) {
		sr := &SQLRecord{
			SQL: "SELECT * FROM users WHERE id IN (?, ?, ?)",
		}
		sr.format()
		assert.Equal(t, "SELECT * FROM users WHERE id IN(intercept)", sr.SQL)
	})

	t.Run("replace lowercase in clause", func(t *testing.T) {
		sr := &SQLRecord{
			SQL: "select * from users where id in (?, ?)",
		}
		sr.format()
		assert.Equal(t, "select * from users where id IN(intercept)", sr.SQL)
	})

	t.Run("no change for simple select", func(t *testing.T) {
		sr := &SQLRecord{
			SQL: "SELECT * FROM users WHERE id = ?",
		}
		sr.format()
		assert.Equal(t, "SELECT * FROM users WHERE id = ?", sr.SQL)
	})
}
