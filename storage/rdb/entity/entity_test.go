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

package entity

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStoragerWithDB(t *testing.T) (*EntityStorager, *sql.DB) {
	// The DAO layer consults stateful.DefaultConfig when recording SQL.
	if stateful.DefaultConfig == nil {
		stateful.DefaultConfig = &stateful.Config{}
	}

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
CREATE TABLE entities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entity_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  type TEXT NOT NULL,
  parent_id TEXT DEFAULT NULL,
  allow_models TEXT,
  block_models TEXT,
  quota_plan_id INTEGER DEFAULT NULL,
  rate_limit_policy_id INTEGER DEFAULT NULL,
  route_rules_id INTEGER DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (entity_id),
  UNIQUE (name)
);`)
	require.NoError(t, err)

	factory := lib.DBContextFactory(func(ctx context.Context, ops ...*lib.Op) (*lib.DBContext, error) {
		return lib.NewDBContext(ctx, db), nil
	})

	return NewEntityStorager(factory), db
}

func setupTestStorager(t *testing.T) *EntityStorager {
	storager, _ := setupTestStoragerWithDB(t)
	return storager
}

func fetchOne(t *testing.T, storager *EntityStorager, entityID string) *entity.EntityParam {
	one, err := storager.FetchEntity(context.Background(), &entity.EntityFilter{EntityID: &entityID})
	require.NoError(t, err)
	require.NotNil(t, one)
	return one
}

func TestUpdateEntity_OmittedModelsPreserveValues(t *testing.T) {
	storager := setupTestStorager(t)
	ctx := context.Background()

	id := "entity-1"
	name := "entity-one"
	typ := "tenant"
	_, err := storager.CreateEntity(ctx, &entity.EntityParam{
		EntityID:    &id,
		Name:        &name,
		Type:        &typ,
		AllowModels: []string{"model-a"},
		BlockModels: []string{"model-b"},
	})
	require.NoError(t, err)

	// PATCH with only name: omitted allow_models/block_models must be preserved
	newName := "entity-one-renamed"
	affected, err := storager.UpdateEntity(ctx,
		&entity.EntityFilter{EntityID: &id},
		&entity.EntityParam{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	one := fetchOne(t, storager, id)
	assert.Equal(t, newName, *one.Name)
	assert.Equal(t, []string{"model-a"}, one.AllowModels)
	assert.Equal(t, []string{"model-b"}, one.BlockModels)
}

func TestUpdateEntity_ProvidedModelsAreWritten(t *testing.T) {
	storager := setupTestStorager(t)
	ctx := context.Background()

	id := "entity-1"
	name := "entity-one"
	typ := "tenant"
	_, err := storager.CreateEntity(ctx, &entity.EntityParam{
		EntityID:    &id,
		Name:        &name,
		Type:        &typ,
		AllowModels: []string{"model-a"},
	})
	require.NoError(t, err)

	affected, err := storager.UpdateEntity(ctx,
		&entity.EntityFilter{EntityID: &id},
		&entity.EntityParam{
			AllowModels: []string{"model-x"},
			BlockModels: []string{"model-y"},
		})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	one := fetchOne(t, storager, id)
	assert.Equal(t, []string{"model-x"}, one.AllowModels)
	assert.Equal(t, []string{"model-y"}, one.BlockModels)
}

func TestCreateEntity_OmittedModelsDefaultToStar(t *testing.T) {
	storager := setupTestStorager(t)
	ctx := context.Background()

	id := "entity-1"
	name := "entity-one"
	typ := "tenant"
	_, err := storager.CreateEntity(ctx, &entity.EntityParam{
		EntityID: &id,
		Name:     &name,
		Type:     &typ,
	})
	require.NoError(t, err)

	one := fetchOne(t, storager, id)
	// 省略 allow_models 默认 ["*]（api-define entities.md §2.1，issue #202）；
	// block_models 默认 [] 不变。
	assert.Equal(t, []string{"*"}, one.AllowModels)
	assert.Empty(t, one.BlockModels)

	// 显式非空 allow_models/block_models 原样往返，防默认回填误伤。
	id2, name2 := "entity-2", "entity-two"
	_, err = storager.CreateEntity(ctx, &entity.EntityParam{
		EntityID:    &id2,
		Name:        &name2,
		Type:        &typ,
		AllowModels: []string{"gpt-4"},
		BlockModels: []string{"gpt-3"},
	})
	require.NoError(t, err)

	two := fetchOne(t, storager, id2)
	assert.Equal(t, []string{"gpt-4"}, two.AllowModels)
	assert.Equal(t, []string{"gpt-3"}, two.BlockModels)
}

// TestFetchEntity_LegacyEmptyAllowModelsNormalizedToStar 验证 issue #202 读回归一化：
// 修复前创建的存量行（持久化 "[]" 或 NULL）回读时按契约归一化为 ["*]。
func TestFetchEntity_LegacyEmptyAllowModelsNormalizedToStar(t *testing.T) {
	storager, db := setupTestStoragerWithDB(t)
	ctx := context.Background()

	typ := "tenant"
	newLegacyEntity := func(id, name, allowModelsSQL string) {
		nameVal := name
		_, err := storager.CreateEntity(ctx, &entity.EntityParam{
			EntityID: &id,
			Name:     &nameVal,
			Type:     &typ,
		})
		require.NoError(t, err)
		// 模拟修复前存量数据：绕过 storager 创建转换，把列改回遗留值。
		_, err = db.Exec("UPDATE entities SET allow_models = "+allowModelsSQL+" WHERE entity_id = ?", id)
		require.NoError(t, err)
	}

	// 存量 "[]" 行 → 回读 ["*]。
	newLegacyEntity("entity-1", "entity-one", "'[]'")
	assert.Equal(t, []string{"*"}, fetchOne(t, storager, "entity-1").AllowModels)

	// 存量 NULL 行 → 回读 ["*]。
	newLegacyEntity("entity-2", "entity-two", "NULL")
	assert.Equal(t, []string{"*"}, fetchOne(t, storager, "entity-2").AllowModels)

	// 存量非空行不受影响。
	newLegacyEntity("entity-3", "entity-three", "'[\"model-a\"]'")
	assert.Equal(t, []string{"model-a"}, fetchOne(t, storager, "entity-3").AllowModels)
}

func TestEntityDescription_RoundTrip(t *testing.T) {
	storager := setupTestStorager(t)
	ctx := context.Background()

	newEntity := func(id, name string) {
		typ := "tenant"
		_, err := storager.CreateEntity(ctx, &entity.EntityParam{
			EntityID: &id,
			Name:     &name,
			Type:     &typ,
		})
		require.NoError(t, err)
	}

	// Create 省略 description → 读回为空字符串（DB 列默认值）。
	newEntity("entity-1", "entity-one")
	one := fetchOne(t, storager, "entity-1")
	require.NotNil(t, one.Description)
	assert.Equal(t, "", *one.Description)

	// Create 显式携带 description → 正确写入。
	id2, name2, desc2 := "entity-2", "entity-two", "运营部"
	typ := "tenant"
	_, err := storager.CreateEntity(ctx, &entity.EntityParam{
		EntityID:    &id2,
		Name:        &name2,
		Type:        &typ,
		Description: &desc2,
	})
	require.NoError(t, err)
	assert.Equal(t, desc2, *fetchOne(t, storager, id2).Description)

	// Update 省略 description → nil-skip 保留原值。
	newName := "entity-two-renamed"
	affected, err := storager.UpdateEntity(ctx,
		&entity.EntityFilter{EntityID: &id2},
		&entity.EntityParam{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	assert.Equal(t, desc2, *fetchOne(t, storager, id2).Description)

	// Update 显式置空 → 写入空字符串（清空）。
	empty := ""
	affected, err = storager.UpdateEntity(ctx,
		&entity.EntityFilter{EntityID: &id2},
		&entity.EntityParam{Description: &empty})
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	assert.Equal(t, "", *fetchOne(t, storager, id2).Description)
}
