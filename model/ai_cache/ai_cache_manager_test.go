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

package ai_cache

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func storageRule(id int64, name, cond string) *AICacheRuleParam {
	now := time.Date(2026, 9, 24, 10, 30, 0, 0, time.UTC)
	return &AICacheRuleParam{
		ID:               lib.PInt64(id),
		Name:             lib.PString(name),
		Cond:             lib.PString(cond),
		CacheKeyStrategy: lib.PString("lastQuestion"),
		CacheTTL:         lib.PInt(3600),
		MaxBodyBytes:     lib.PInt64(1048576),
		MaxValueBytes:    lib.PInt64(1048576),
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}
}

func sharedRule(name, cond string) *shared.AICacheRuleParam {
	return &shared.AICacheRuleParam{
		Name: lib.PString(name),
		Cond: lib.PString(cond),
	}
}

func TestAICacheManager_GetAICacheRules(t *testing.T) {
	ctx := context.Background()

	t.Run("returns rules in storage order with timestamps", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{
					storageRule(1, "rule-a", "default_t()"),
					storageRule(2, "rule-b", "req_path_in(\"/v1\", false)"),
				}, nil
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		rules, err := m.GetAICacheRules(ctx)
		require.NoError(t, err)
		require.Len(t, rules, 2)
		assert.Equal(t, "rule-a", *rules[0].Name)
		assert.Equal(t, "rule-b", *rules[1].Name)
		assert.NotNil(t, rules[0].CreatedAt)
	})

	t.Run("empty collection returns empty slice", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{}, nil
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		rules, err := m.GetAICacheRules(ctx)
		require.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 0)
	})

	t.Run("storage error propagates", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return nil, errors.New("db down")
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		_, err := m.GetAICacheRules(ctx)
		require.Error(t, err)
	})
}

func TestAICacheManager_SetAICacheRules(t *testing.T) {
	ctx := context.Background()

	t.Run("replaces collection in slice order with defaults filled", func(t *testing.T) {
		var replaced [][]*AICacheRuleParam
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{}, nil
			},
			replaceAllFn: func(ctx context.Context, rules []*AICacheRuleParam) error {
				replaced = append(replaced, rules)
				return nil
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		updated, err := m.SetAICacheRules(ctx, &shared.AICacheRulesParam{
			Rules: []*shared.AICacheRuleParam{
				sharedRule("first", "default_t()"),
				{
					Name:             lib.PString("second"),
					Cond:             lib.PString("req_path_in(\"/v1\", false)"),
					CacheKeyStrategy: lib.PString("allQuestions"),
					CacheTTL:         lib.PInt(0),
					MaxBodyBytes:     lib.PInt64(2048),
					MaxValueBytes:    lib.PInt64(4096),
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, updated)

		// ReplaceAll must be invoked exactly once, in array order.
		require.Len(t, replaced, 1)
		require.Len(t, replaced[0], 2)
		assert.Equal(t, "first", *replaced[0][0].Name)
		assert.Equal(t, "second", *replaced[0][1].Name)

		// Omitted optional fields are filled with documented defaults.
		assert.Equal(t, DefaultCacheKeyStrategy, *replaced[0][0].CacheKeyStrategy)
		assert.Equal(t, DefaultCacheTTL, *replaced[0][0].CacheTTL)
		assert.Equal(t, int64(DefaultMaxBodyBytes), *replaced[0][0].MaxBodyBytes)
		assert.Equal(t, int64(DefaultMaxValueBytes), *replaced[0][0].MaxValueBytes)

		// Explicit values are preserved.
		assert.Equal(t, "allQuestions", *replaced[0][1].CacheKeyStrategy)
		assert.Equal(t, int64(2048), *replaced[0][1].MaxBodyBytes)
	})

	t.Run("nil param and nil rules clear the collection", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		_, err := m.SetAICacheRules(ctx, nil)
		require.NoError(t, err)
		_, err = m.SetAICacheRules(ctx, &shared.AICacheRulesParam{})
		require.NoError(t, err)

		require.Len(t, store.replaceAllCalls, 2)
		assert.Len(t, store.replaceAllCalls[0], 0)
		assert.Len(t, store.replaceAllCalls[1], 0)
	})

	t.Run("replace failure rolls back and records failed audit", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		store := &fakeAICacheRuleStorager{
			replaceAllFn: func(ctx context.Context, rules []*AICacheRuleParam) error {
				return errors.New("replace error")
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")
		m.SetOperationLogManager(recorder)

		_, err := m.SetAICacheRules(ctx, &shared.AICacheRulesParam{
			Rules: []*shared.AICacheRuleParam{sharedRule("r1", "default_t()")},
		})
		require.Error(t, err)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, string(ioperlog.ActionUpdate), entry.Action)
		assert.Equal(t, string(ioperlog.ResourceTypeAICacheRule), entry.ResourceType)
		assert.Equal(t, ioperlog.StatusFailed, entry.Status)
		assert.NotEmpty(t, entry.ErrorMsg)
	})

	t.Run("records successful audit with collection snapshots", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		calls := 0
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				calls++
				if calls == 1 {
					return []*AICacheRuleParam{storageRule(1, "old-rule", "default_t()")}, nil
				}
				return []*AICacheRuleParam{storageRule(2, "new-rule", "default_t()")}, nil
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")
		m.SetOperationLogManager(recorder)

		_, err := m.SetAICacheRules(ctx, &shared.AICacheRulesParam{
			Rules: []*shared.AICacheRuleParam{sharedRule("new-rule", "default_t()")},
		})
		require.NoError(t, err)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, ioperlog.StatusSuccess, entry.Status)
		assert.Empty(t, entry.ErrorMsg)

		before, ok := entry.ChangeSummary["before"].(map[string]interface{})
		require.True(t, ok)
		beforeRules, ok := before["rules"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, beforeRules, 1)
		assert.Equal(t, "old-rule", beforeRules[0]["name"])

		after, ok := entry.ChangeSummary["after"].(map[string]interface{})
		require.True(t, ok)
		afterRules, ok := after["rules"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, afterRules, 1)
		assert.Equal(t, "new-rule", afterRules[0]["name"])
		// Snapshot keys use the API lowercase vocabulary; internal id is absent.
		_, hasID := afterRules[0]["id"]
		assert.False(t, hasID)
	})

	t.Run("records failed audit for validation rejection", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{storageRule(1, "old-rule", "default_t()")}, nil
			},
		}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")
		m.SetOperationLogManager(recorder)

		m.RecordSetAICacheRulesFailure(ctx, &shared.AICacheRulesParam{
			Rules: []*shared.AICacheRuleParam{sharedRule("bad-rule", "cond ! compile")},
		}, errors.New("cond.Build(): syntax error"))

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, string(ioperlog.ActionUpdate), entry.Action)
		assert.Equal(t, string(ioperlog.ResourceTypeAICacheRule), entry.ResourceType)
		assert.Equal(t, aiCacheRulesResourceID, entry.ResourceID)
		assert.Equal(t, ioperlog.StatusFailed, entry.Status)
		assert.NotEmpty(t, entry.ErrorMsg)

		// Identity and before snapshot come from storage, not the request.
		before, ok := entry.ChangeSummary["before"].(map[string]interface{})
		require.True(t, ok)
		beforeRules, ok := before["rules"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, beforeRules, 1)
		assert.Equal(t, "old-rule", beforeRules[0]["name"])

		after, ok := entry.ChangeSummary["after"].(map[string]interface{})
		require.True(t, ok)
		afterRules, ok := after["rules"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, afterRules, 1)
		assert.Equal(t, "bad-rule", afterRules[0]["name"])
	})

	t.Run("no audit when operation log manager is nil", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{}
		m := NewAICacheManager(&fakeTxn{}, store, nil, "AI_product")

		_, err := m.SetAICacheRules(ctx, &shared.AICacheRulesParam{
			Rules: []*shared.AICacheRuleParam{sharedRule("r1", "default_t()")},
		})
		require.NoError(t, err)
	})
}

func TestAICacheManager_AICacheRuleGenerator(t *testing.T) {
	ctx := context.Background()

	newManager := func(store *fakeAICacheRuleStorager, product string) *AICacheManager {
		return NewAICacheManager(&fakeTxn{}, store, nil, product)
	}

	t.Run("exports rules in id ascending order under injected product name", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{
					storageRule(1, "first", "default_t()"),
					storageRule(2, "second", "req_path_in(\"/v1\", false)"),
				}, nil
			},
		}
		m := newManager(store, "CustomProduct")

		data, err := m.AICacheRuleGenerator(ctx)
		require.NoError(t, err)
		assert.Equal(t, ConfigTopicProductAICache, data.Topic)

		conf, ok := data.DataWithoutVersion.(*ExportAICacheRuleConfig)
		require.True(t, ok)
		require.Contains(t, conf.Config, "CustomProduct")
		require.Len(t, conf.Config["CustomProduct"], 2)
		assert.Equal(t, "default_t()", conf.Config["CustomProduct"][0].Cond)
		assert.Equal(t, "req_path_in(\"/v1\", false)", conf.Config["CustomProduct"][1].Cond)
		assert.Equal(t, iversion_control.ZeroVersion, conf.Version)
	})

	t.Run("empty table exports empty array with product key present", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{}, nil
			},
		}
		m := newManager(store, "AI_product")

		data, err := m.AICacheRuleGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportAICacheRuleConfig)

		rules, ok := conf.Config["AI_product"]
		require.True(t, ok)
		require.NotNil(t, rules)
		assert.Len(t, rules, 0)

		bs, err := json.Marshal(conf)
		require.NoError(t, err)
		assert.Contains(t, string(bs), `"AI_product":[]`)
	})

	t.Run("marshaled JSON keys match the frozen contract verbatim", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return []*AICacheRuleParam{storageRule(1, "r1", "default_t()")}, nil
			},
		}
		m := newManager(store, "AI_product")

		data, err := m.AICacheRuleGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportAICacheRuleConfig)
		conf.UpdateVersion("20260924103000")

		bs, err := json.Marshal(conf)
		require.NoError(t, err)

		var top map[string]interface{}
		require.NoError(t, json.Unmarshal(bs, &top))
		assert.ElementsMatch(t, []string{"Version", "Config"}, keysOf(top))

		config := top["Config"].(map[string]interface{})
		assert.ElementsMatch(t, []string{"AI_product"}, keysOf(config))

		rules := config["AI_product"].([]interface{})
		require.Len(t, rules, 1)
		rule := rules[0].(map[string]interface{})
		assert.ElementsMatch(t,
			[]string{"cond", "cacheKeyStrategy", "cacheTTL", "maxBodyBytes", "maxValueBytes"},
			keysOf(rule))
	})

	t.Run("storage error propagates", func(t *testing.T) {
		store := &fakeAICacheRuleStorager{
			fetchAllFn: func(ctx context.Context) ([]*AICacheRuleParam, error) {
				return nil, errors.New("db down")
			},
		}
		m := newManager(store, "AI_product")

		_, err := m.AICacheRuleGenerator(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fetch ai cache rules error")
	})
}

func TestAICacheManager_ConfigExport(t *testing.T) {
	ctx := context.Background()

	buildManager := func(versionStore *fakeVersionControlStorager, product string) *AICacheManager {
		var vcm *iversion_control.VersionControlManager
		if versionStore != nil {
			vcm = iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore)
		}
		return NewAICacheManager(&fakeTxn{}, &fakeAICacheRuleStorager{}, vcm, product)
	}

	t.Run("returns nil when version unchanged", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return iversion_control.ZeroVersion, nil
			},
		}
		m := buildManager(versionStore, "AI_product")

		conf, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		assert.Nil(t, conf)
	})

	t.Run("returns config with new version", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "20260924103000", nil
			},
		}
		m := buildManager(versionStore, "AI_product")

		conf, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		require.NotNil(t, conf)
		assert.Equal(t, "20260924103000", conf.Version)
		assert.Contains(t, conf.Config, "AI_product")
	})

	t.Run("version control error propagates", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "", errors.New("version db error")
			},
		}
		m := buildManager(versionStore, "AI_product")

		_, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version db error")
	})
}

func TestAICacheRuleParam_Conversions(t *testing.T) {
	t.Run("nil inputs stay nil", func(t *testing.T) {
		assert.Nil(t, aiCacheRuleParamFromShared(nil))
		assert.Nil(t, aiCacheRuleParamToShared(nil))
	})

	t.Run("nil optional fields are filled with defaults", func(t *testing.T) {
		rule := aiCacheRuleParamFromShared(sharedRule("r1", "default_t()"))
		require.NotNil(t, rule)
		assert.Equal(t, "r1", *rule.Name)
		assert.Equal(t, DefaultCacheKeyStrategy, *rule.CacheKeyStrategy)
		assert.Equal(t, DefaultCacheTTL, *rule.CacheTTL)
		assert.Equal(t, int64(DefaultMaxBodyBytes), *rule.MaxBodyBytes)
		assert.Equal(t, int64(DefaultMaxValueBytes), *rule.MaxValueBytes)
	})

	t.Run("round trip preserves values", func(t *testing.T) {
		now := time.Date(2026, 9, 24, 10, 30, 0, 0, time.UTC)
		original := &shared.AICacheRuleParam{
			Name:             lib.PString("r1"),
			Cond:             lib.PString("default_t()"),
			CacheKeyStrategy: lib.PString("disabled"),
			CacheTTL:         lib.PInt(60),
			MaxBodyBytes:     lib.PInt64(1024),
			MaxValueBytes:    lib.PInt64(2048),
			CreatedAt:        &now,
			UpdatedAt:        &now,
		}

		rule := aiCacheRuleParamFromShared(original)
		rst := aiCacheRuleParamToShared(rule)
		assert.Equal(t, original, rst)
	})
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
