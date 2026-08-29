// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imods

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iai_route"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/quota"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAPIKeyRuleManager() *APIKeyRuleManager {
	return NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
		&fakeAPIKeyStorager{},
		&fakeAIRouteRuleStorager{},
		&fakeQuotaPlanStorager{},
		&fakeEntityStorager{},
		&fakeEntityTypeStorager{}, nil)
}

func TestAPIKeyRuleManager_ConfigExport(t *testing.T) {
	setupState()
	ctx := context.Background()

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					KeyCreateAt: lib.PTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)),
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "v-abc", nil
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore),
		apiKeyStore,
		aiRouteStore,
		&fakeQuotaPlanStorager{},
		&fakeEntityStorager{},
		&fakeEntityTypeStorager{}, nil)

	conf, err := m.ConfigExport(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, conf)
	assert.Equal(t, "v-abc", *conf.Version)
	assert.Len(t, conf.Config["AI_product"], 1)
	assert.Len(t, conf.Tokens["AI_product"], 1)

	// Same version returns nil
	conf2, err := m.ConfigExport(ctx, "v-abc")
	require.NoError(t, err)
	assert.Nil(t, conf2)
}

func TestAPIKeyRuleManager_ConfigExport_GeneratorError(t *testing.T) {
	setupState()
	ctx := context.Background()

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return nil, errors.New("api key fetch error")
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore

	conf, err := m.ConfigExport(ctx, "")
	require.Error(t, err)
	assert.Nil(t, conf)
}

func TestAPIKeyRuleManager_ConfigExport_Concurrent(t *testing.T) {
	setupState()
	ctx := context.Background()

	const numKeys = 50
	const numGoroutines = 100
	const iterations = 10

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			keys := make([]*api_key.APIKeyParam, numKeys)
			for i := 0; i < numKeys; i++ {
				id := fmt.Sprintf("key-%d", i)
				quotaPlanID := int64(i%5 + 1)
				entityID := fmt.Sprintf("entity-%d", i%10)
				keys[i] = &api_key.APIKeyParam{
					ID:          lib.PString(id),
					Key:         lib.PString(fmt.Sprintf("ak-%s", id)),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					QuotaPlanID: &quotaPlanID,
					EntityID:    lib.PString(entityID),
					KeyCreateAt: lib.PTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)),
				}
			}
			return keys, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		listFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
			plans := make([]*quota.QuotaPlanParam, 5)
			for i := 0; i < 5; i++ {
				id := int64(i + 1)
				plans[i] = &quota.QuotaPlanParam{
					ID:        lib.PInt64(id),
					Unlimited: lib.PBool(false),
					Quota:     lib.PFloat64(100),
				}
			}
			return plans, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
			entities := make([]*entity.EntityParam, 10)
			for i := 0; i < 10; i++ {
				id := fmt.Sprintf("entity-%d", i)
				entities[i] = &entity.EntityParam{
					EntityID: lib.PString(id),
					Name:     lib.PString(id),
					Type:     lib.PString("team"),
				}
			}
			return entities, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "v-concurrent", nil
		},
	}

	entityTypeStore := &fakeEntityTypeStorager{
		listFn: func(ctx context.Context, filter *entity.EntityTypeFilter) ([]*entity.EntityTypeParam, error) {
			return []*entity.EntityTypeParam{
				{TypeName: lib.PString("team"), Level: lib.PInt(2)},
			}, nil
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore),
		apiKeyStore,
		aiRouteStore,
		quotaPlanStore,
		entityStore,
		entityTypeStore, nil)

	// Increase concurrency to maximize the chance of triggering the race.
	prevMaxProcs := runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(prevMaxProcs)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errCh := make(chan error, numGoroutines*iterations)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				conf, err := m.ConfigExport(ctx, "")
				if err != nil {
					errCh <- err
					return
				}
				if conf == nil {
					continue
				}
				if len(conf.Tokens["AI_product"]) != numKeys {
					errCh <- fmt.Errorf("expected %d tokens, got %d", numKeys, len(conf.Tokens["AI_product"]))
					return
				}
				if len(conf.QuotaPlans["AI_product"]) == 0 {
					errCh <- fmt.Errorf("expected non-empty quota plans")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestAPIKeyRuleManager_FormatAIRouteAPIKeyRules(t *testing.T) {
	setupState()
	ctx := context.Background()

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return []*iai_route.Rule{
				{
					Basic: &iai_route.BasicInfo{
						Domain: lib.PString("example.com"),
						PathFilter: &iai_route.PathFilter{
							MatchMode:  lib.PString(iai_route.MatchModePrefix),
							Path:       lib.PString("/api"),
							IgnoreCase: lib.PBool(true),
						},
					},
				},
			}, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.aiRouteStorager = aiRouteStore

	rules, err := m.FormatAIRouteAPIKeyRules(ctx, "AI_product")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	assert.Equal(t, `req_host_in("example.com")&&req_path_prefix_in("/api", true)`, rules[0].Cond)
	assert.Equal(t, APIKeyActionCMD, rules[0].Actions[0].Cmd)
	assert.Equal(t, "default_t()", rules[1].Cond)
	assert.Equal(t, "AI_product", rules[0].ProductName)
}

func TestAPIKeyRuleManager_FormatAIRouteAPIKeyRules_Error(t *testing.T) {
	setupState()
	ctx := context.Background()

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, errors.New("route fetch error")
		},
	}

	m := newTestAPIKeyRuleManager()
	m.aiRouteStorager = aiRouteStore

	rules, err := m.FormatAIRouteAPIKeyRules(ctx, "AI_product")
	require.Error(t, err)
	assert.Nil(t, rules)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	futureExpire := time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	quotaPlanID := int64(10)
	entityID := "entity-1"

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					ExpiredTime: &futureExpire,
					KeyCreateAt: lib.PTime(keyCreateAt),
					QuotaPlanID: &quotaPlanID,
					EntityID:    &entityID,
					Models:      []string{"gpt-4", "gpt-3.5"},
					Subnet:      []string{"10.0.0.0/8"},
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		listFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
			return []*quota.QuotaPlanParam{
				{
					ID:        lib.PInt64(quotaPlanID),
					Unlimited: lib.PBool(false),
					Quota:     lib.PFloat64(1000),
				},
			}, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
			return []*entity.EntityParam{
				{
					EntityID:    lib.PString(entityID),
					Name:        lib.PString("team-a"),
					Type:        lib.PString("team"),
					AllowModels: []string{"gpt-4", "claude"},
					BlockModels: []string{"gpt-2"},
				},
			}, nil
		},
	}

	entityTypeStore := &fakeEntityTypeStorager{
		listFn: func(ctx context.Context, filter *entity.EntityTypeFilter) ([]*entity.EntityTypeParam, error) {
			return []*entity.EntityTypeParam{
				{
					TypeName: lib.PString("team"),
					Level:    lib.PInt(2),
				},
			}, nil
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
		apiKeyStore,
		aiRouteStore,
		quotaPlanStore,
		entityStore,
		entityTypeStore, nil)

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, ConfigTopicProductAPIKeyRule, data.Topic)

	conf, ok := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	require.True(t, ok)
	assert.NotNil(t, conf.Tokens["AI_product"]["ak-key-1"])
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.Equal(t, "key-1", token.KeyID)
	assert.True(t, token.Enabled)
	assert.Equal(t, "gpt-4", *token.Models)
	assert.Equal(t, "gpt-2", *token.BlockModels)
	assert.Equal(t, "10.0.0.0/8", *token.Subnet)
	assert.NotEmpty(t, token.QuotaPlans)
	assert.NotEmpty(t, token.Tags)
	assert.Equal(t, 2, token.Tags[0].TagLevel)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_Unlimited(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:             lib.PString("key-1"),
					Key:            lib.PString("ak-key-1"),
					ProductName:    lib.PString("AI_product"),
					Enable:         lib.PBool(true),
					UnlimitedQuota: lib.PBool(true),
					KeyCreateAt:    lib.PTime(keyCreateAt),
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.True(t, token.UnlimitedQuota)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_FalseUnlimitedWithEmptyQuotaPlans(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	quotaPlanID := int64(1)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:             lib.PString("key-1"),
					Key:            lib.PString("ak-key-1"),
					ProductName:    lib.PString("AI_product"),
					Enable:         lib.PBool(true),
					UnlimitedQuota: lib.PBool(false),
					QuotaPlanID:    &quotaPlanID,
					KeyCreateAt:    lib.PTime(keyCreateAt),
				},
			}, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		fetchFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) (*quota.QuotaPlanParam, error) {
			if filter.ID != nil && *filter.ID == quotaPlanID {
				return &quota.QuotaPlanParam{
					ID:        filter.ID,
					Unlimited: lib.PBool(true),
				}, nil
			}
			return nil, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.quotaPlanStorager = quotaPlanStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.Empty(t, token.QuotaPlans)
	assert.True(t, token.UnlimitedQuota)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_MinimalParamsWithNoQuotaPlan(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					KeyCreateAt: lib.PTime(keyCreateAt),
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.Empty(t, token.QuotaPlans)
	assert.True(t, token.UnlimitedQuota)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_FalseUnlimitedWithNonUnlimitedQuotaPlan(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	quotaPlanID := int64(1)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:             lib.PString("key-1"),
					Key:            lib.PString("ak-key-1"),
					ProductName:    lib.PString("AI_product"),
					Enable:         lib.PBool(true),
					UnlimitedQuota: lib.PBool(false),
					QuotaPlanID:    &quotaPlanID,
					KeyCreateAt:    lib.PTime(keyCreateAt),
				},
			}, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		listFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
			return []*quota.QuotaPlanParam{
				{
					ID:        lib.PInt64(quotaPlanID),
					Unlimited: lib.PBool(false),
					Quota:     lib.PFloat64(100),
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.quotaPlanStorager = quotaPlanStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.False(t, token.UnlimitedQuota)
	assert.NotEmpty(t, token.QuotaPlans)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_Expired(t *testing.T) {
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	pastExpire := keyCreateAt.Add(-24 * time.Hour).Unix()
	quota := float64(100)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					ExpiredTime: &pastExpire,
					KeyCreateAt: lib.PTime(keyCreateAt),
					QuotaPlan: &shared.QuotaPlanParam{
						Quota: &quota,
					},
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.Equal(t, pastExpire, token.ExpiredTime)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_Exhausted(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	futureExpire := time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	quota := float64(0)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					ExpiredTime: &futureExpire,
					KeyCreateAt: lib.PTime(keyCreateAt),
					QuotaPlan: &shared.QuotaPlanParam{
						Quota: &quota,
					},
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.True(t, token.Enabled)
	assert.Equal(t, futureExpire, token.ExpiredTime)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_Disabled(t *testing.T) {
	setupState()
	ctx := context.Background()

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(false),
					KeyCreateAt: lib.PTime(time.Now()),
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.False(t, token.Enabled)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_ModelIntersectionEmpty(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	entityID := "entity-1"
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					KeyCreateAt: lib.PTime(keyCreateAt),
					EntityID:    &entityID,
					Models:      []string{"gpt-4"},
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
			return []*entity.EntityParam{
				{
					EntityID:    lib.PString(entityID),
					AllowModels: []string{"claude"},
				},
			}, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore
	m.entityStorager = entityStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.Equal(t, "", *token.Models)
	assert.False(t, token.Enabled)
}

func TestAPIKeyRuleManager_APIKeyRuleGenerator_EntityHierarchy(t *testing.T) {
	setupState()
	ctx := context.Background()

	keyCreateAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	parentID := "parent-1"
	entityID := "entity-1"
	quotaPlanID := int64(20)

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{
				{
					ID:          lib.PString("key-1"),
					Key:         lib.PString("ak-key-1"),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					KeyCreateAt: lib.PTime(keyCreateAt),
					EntityID:    &entityID,
				},
			}, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		listFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
			return []*quota.QuotaPlanParam{
				{
					ID:        lib.PInt64(quotaPlanID),
					Unlimited: lib.PBool(false),
					Quota:     lib.PFloat64(500),
				},
			}, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
			return []*entity.EntityParam{
				{
					EntityID:    lib.PString(entityID),
					Name:        lib.PString("team-a"),
					Type:        lib.PString("team"),
					ParentID:    &parentID,
					QuotaPlanID: &quotaPlanID,
					AllowModels: []string{"gpt-4"},
				},
				{
					EntityID:    lib.PString(parentID),
					Name:        lib.PString("org-a"),
					Type:        lib.PString("org"),
					AllowModels: []string{"gpt-4", "claude"},
				},
			}, nil
		},
	}

	entityTypeStore := &fakeEntityTypeStorager{
		listFn: func(ctx context.Context, filter *entity.EntityTypeFilter) ([]*entity.EntityTypeParam, error) {
			return []*entity.EntityTypeParam{
				{TypeName: lib.PString("team"), Level: lib.PInt(2)},
				{TypeName: lib.PString("org"), Level: lib.PInt(1)},
			}, nil
		},
	}

	m := newTestAPIKeyRuleManager()
	m.apiKeyStorager = apiKeyStore
	m.aiRouteStorager = aiRouteStore
	m.quotaPlanStorager = quotaPlanStore
	m.entityStorager = entityStore
	m.entityTypeStorager = entityTypeStore

	data, err := m.APIKeyRuleGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ModAPIKeyRuleConf)
	token := conf.Tokens["AI_product"]["ak-key-1"]
	assert.Equal(t, "gpt-4", *token.Models)
	assert.Len(t, token.Tags, 2)
	assert.Equal(t, 2, token.Tags[0].TagLevel)
	assert.Equal(t, 1, token.Tags[1].TagLevel)
	assert.Contains(t, token.QuotaPlans, "entity-1")
}

func TestAPIKeyRuleManager_FetchQuotaPlansWithEntityHierarchy(t *testing.T) {
	setupState()
	ctx := context.Background()

	key := "ak-key-1"
	quotaPlanID := int64(10)
	entityID := "entity-1"

	apiKey := &api_key.APIKeyParam{
		Key:         &key,
		QuotaPlanID: &quotaPlanID,
		EntityID:    &entityID,
	}

	quotaPlanMap := map[int64]*quota.QuotaPlanParam{
		quotaPlanID: {
			ID:        lib.PInt64(quotaPlanID),
			Unlimited: lib.PBool(false),
			Quota:     lib.PFloat64(100),
		},
	}

	entityMap := map[string]*entity.EntityParam{
		entityID: {
			EntityID: lib.PString(entityID),
			Name:     lib.PString("team-a"),
			Type:     lib.PString("team"),
		},
	}

	entityTypeMap := map[string]*entity.EntityTypeParam{
		"team": {TypeName: lib.PString("team"), Level: lib.PInt(2)},
	}

	m := newTestAPIKeyRuleManager()
	collectedQuotaPlans := make(map[string][]*QuotaPlan)

	ids, tags, err := m.fetchQuotaPlansWithEntityHierarchy(ctx, apiKey, "AI_product", collectedQuotaPlans, entityMap, quotaPlanMap, entityTypeMap)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Len(t, tags, 1)
	assert.Equal(t, "team-a", tags[0].TagValue)
	assert.Equal(t, 2, tags[0].TagLevel)
}

func TestAPIKeyRuleManager_FetchEntityModelHierarchy(t *testing.T) {
	setupState()
	ctx := context.Background()

	parentID := "parent-1"
	entityID := "entity-1"

	entityMap := map[string]*entity.EntityParam{
		entityID: {
			EntityID:    lib.PString(entityID),
			ParentID:    &parentID,
			AllowModels: []string{"gpt-4", "claude"},
			BlockModels: []string{"gpt-2"},
		},
		parentID: {
			EntityID:    lib.PString(parentID),
			AllowModels: []string{"gpt-4"},
			BlockModels: []string{"bloom"},
		},
	}

	m := newTestAPIKeyRuleManager()

	allow, block, err := m.fetchEntityModelHierarchy(ctx, entityMap, entityID)
	require.NoError(t, err)
	assert.Equal(t, []string{"gpt-4"}, allow)
	assert.Equal(t, []string{"gpt-2", "bloom"}, block)
}

func TestAPIKeyRuleManager_FetchEntityModelHierarchy_EntityNotFound(t *testing.T) {
	setupState()
	ctx := context.Background()

	m := newTestAPIKeyRuleManager()

	allow, block, err := m.fetchEntityModelHierarchy(ctx, map[string]*entity.EntityParam{}, "missing")
	require.NoError(t, err)
	assert.Nil(t, allow)
	assert.Empty(t, block)
}

func TestAPIKeyRuleManager_ContainsQuotaPlan(t *testing.T) {
	collectedQuotaPlans := map[string][]*QuotaPlan{
		"AI_product": {
			{Id: "qp-1"},
		},
	}

	assert.True(t, containsQuotaPlan(collectedQuotaPlans, "AI_product", "qp-1"))
	assert.False(t, containsQuotaPlan(collectedQuotaPlans, "AI_product", "qp-2"))
	assert.False(t, containsQuotaPlan(collectedQuotaPlans, "other", "qp-1"))
}


func TestAPIKeyRuleManager_ConfigExport_Performance(t *testing.T) {
	setupState()
	ctx := context.Background()

	const numKeys = 690
	const numEntities = 100
	const numQuotaPlans = 20

	// Build 3-level entity hierarchy: leaf -> middle -> root.
	entityList := make([]*entity.EntityParam, 0, numEntities)
	for i := 0; i < numEntities; i++ {
		entityID := fmt.Sprintf("entity-%d", i)
		entityType := "team"
		var parentID *string
		if i%3 == 0 {
			// root
			entityType = "org"
		} else {
			// middle -> root, or leaf -> middle
			parent := fmt.Sprintf("entity-%d", i-1)
			parentID = &parent
		}

		var quotaPlanID *int64
		if i%2 == 0 {
			qpID := int64(i%numQuotaPlans + 1)
			quotaPlanID = &qpID
		}

		entityList = append(entityList, &entity.EntityParam{
			EntityID:    lib.PString(entityID),
			Name:        lib.PString(entityID),
			Type:        lib.PString(entityType),
			ParentID:    parentID,
			QuotaPlanID: quotaPlanID,
			AllowModels: []string{"gpt-4", "claude", "gpt-3.5"},
			BlockModels: []string{"gpt-2"},
		})
	}

	quotaPlanList := make([]*quota.QuotaPlanParam, 0, numQuotaPlans)
	for i := 1; i <= numQuotaPlans; i++ {
		id := int64(i)
		quotaPlanList = append(quotaPlanList, &quota.QuotaPlanParam{
			ID:        lib.PInt64(id),
			Unlimited: lib.PBool(false),
			Quota:     lib.PFloat64(1000),
		})
	}

	entityTypeList := []*entity.EntityTypeParam{
		{TypeName: lib.PString("team"), Level: lib.PInt(2)},
		{TypeName: lib.PString("org"), Level: lib.PInt(1)},
	}

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			keys := make([]*api_key.APIKeyParam, numKeys)
			for i := 0; i < numKeys; i++ {
				id := fmt.Sprintf("key-%d", i)
				quotaPlanID := int64(i%numQuotaPlans + 1)
				entityID := fmt.Sprintf("entity-%d", i%numEntities)
				keys[i] = &api_key.APIKeyParam{
					ID:          lib.PString(id),
					Key:         lib.PString(fmt.Sprintf("ak-%s", id)),
					ProductName: lib.PString("AI_product"),
					Enable:      lib.PBool(true),
					QuotaPlanID: &quotaPlanID,
					EntityID:    lib.PString(entityID),
					Models:      []string{"gpt-4"},
					KeyCreateAt: lib.PTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)),
				}
			}
			return keys, nil
		},
	}

	quotaPlanStore := &fakeQuotaPlanStorager{
		listFn: func(ctx context.Context, filter *quota.QuotaPlanFilter) ([]*quota.QuotaPlanParam, error) {
			return quotaPlanList, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *entity.EntityFilter) ([]*entity.EntityParam, error) {
			return entityList, nil
		},
	}

	entityTypeStore := &fakeEntityTypeStorager{
		listFn: func(ctx context.Context, filter *entity.EntityTypeFilter) ([]*entity.EntityTypeParam, error) {
			return entityTypeList, nil
		},
	}

	aiRouteStore := &fakeAIRouteRuleStorager{
		fetchFn: func(ctx context.Context, filter *iai_route.AIRouteFilter) ([]*iai_route.Rule, error) {
			return nil, nil
		},
	}

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "v-performance", nil
		},
	}

	m := NewAPIKeyRuleManager(
		&fakeTxn{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore),
		apiKeyStore,
		aiRouteStore,
		quotaPlanStore,
		entityStore,
		entityTypeStore, nil)

	start := time.Now()
	conf, err := m.ConfigExport(ctx, "")
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, conf)
	assert.Len(t, conf.Tokens["AI_product"], numKeys)
	assert.NotEmpty(t, conf.QuotaPlans["AI_product"])

	t.Logf("ConfigExport with %d keys, %d entities, %d quota plans took %d ms",
		numKeys, numEntities, numQuotaPlans, elapsed.Milliseconds())

	assert.Less(t, elapsed.Milliseconds(), int64(500), "ConfigExport should complete within 500 ms")
}
