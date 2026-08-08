// Copyright(c) 2026 The Infinity AI Gateway Authors.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/infinity-ai-gateway/ai-gateway-api/lib"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/icluster_conf"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/quota"
	"github.com/infinity-ai-gateway/ai-gateway-api/model/shared"
)

func newTestAIRouteExporter() *AIRouteExporter {
	return NewAIRouteExporter(
		&fakeAPIKeyStorager{},
		&fakeEntityStorager{},
		&fakeRouteRulesStorager{},
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
	)
}

func TestNewAIRouteExporter(t *testing.T) {
	e := newTestAIRouteExporter()
	assert.NotNil(t, e)
}

func TestAIRouteExporter_ConfigExport(t *testing.T) {
	setupState()
	ctx := context.Background()

	versionStore := &fakeVersionControlStorager{
		upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
			return "v-ai-route-1", nil
		},
	}

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
			return []*icluster_conf.APIKeyParam{
				{
					ID:  lib.PString("key-1"),
					Key: lib.PString("ak-key-1"),
				},
			}, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
			return nil, nil
		},
	}

	routeRulesStore := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			if ruleType == shared.RouteRulesTypeGlobal {
				enabled := true
				return &shared.RouteRulesParam{
					Enabled: &enabled,
					Rules:   []*shared.AiRouteRuleParam{},
				}, nil
			}
			return nil, nil
		},
	}

	e := NewAIRouteExporter(
		apiKeyStore,
		entityStore,
		routeRulesStore,
		iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStore),
	)

	conf, err := e.ConfigExport(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, conf)
	assert.Equal(t, "v-ai-route-1", conf.Version)

	conf2, err := e.ConfigExport(ctx, "v-ai-route-1")
	require.NoError(t, err)
	assert.Nil(t, conf2)
}

func TestAIRouteExporter_ConfigExport_Error(t *testing.T) {
	setupState()
	ctx := context.Background()

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
			return nil, errors.New("fetch failed")
		},
	}

	e := newTestAIRouteExporter()
	e.apiKeyStorager = apiKeyStore

	conf, err := e.ConfigExport(ctx, "")
	require.Error(t, err)
	assert.Nil(t, conf)
}

func TestAIRouteExporter_AIRouteGenerator(t *testing.T) {
	setupState()
	ctx := context.Background()

	routeRulesID := int64(100)
	entityRouteRulesID := int64(200)
	apiKey := "ak-key-1"
	entityID := "entity-1"

	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
			return []*icluster_conf.APIKeyParam{
				{
					ID:           lib.PString("key-1"),
					Key:          &apiKey,
					EntityID:     &entityID,
					RouteRulesID: &routeRulesID,
				},
			}, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
			return []*quota.EntityParam{
				{
					EntityID:     lib.PString(entityID),
					Name:         lib.PString("team-a"),
					RouteRulesID: &entityRouteRulesID,
				},
			}, nil
		},
	}

	routeRulesStore := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			if ruleType == shared.RouteRulesTypeGlobal {
				enabled := true
				return &shared.RouteRulesParam{
					Enabled: &enabled,
					Rules:   []*shared.AiRouteRuleParam{},
				}, nil
			}
			return nil, nil
		},
		fetchByIDFn: func(ctx context.Context, id int64) (*shared.RouteRulesParam, error) {
			enabled := true
			name := "rule-1"
			cluster := "cluster-1"
			model := "gpt-4"
			weight := 100
			return &shared.RouteRulesParam{
				Enabled: &enabled,
				Rules: []*shared.AiRouteRuleParam{
					{
						Name: &name,
						Cond: lib.PString("default_t()"),
						Targets: []*shared.AiRouteTargetParam{
							{
								ClusterName: &cluster,
								Model:       &model,
								Weight:      &weight,
							},
						},
					},
				},
			}, nil
		},
	}

	e := NewAIRouteExporter(
		apiKeyStore,
		entityStore,
		routeRulesStore,
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
	)

	data, err := e.AIRouteGenerator(ctx)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, ConfigTopicProductAIRoute, data.Topic)

	conf := data.DataWithoutVersion.(*AiRouteDataExport)
	assert.Contains(t, conf.RouteRules, "apikey_ak-key-1")
	assert.Contains(t, conf.RouteRules, "entity_team-a")
	assert.Contains(t, conf.RouteRules, "global_default")
	assert.Contains(t, conf.ApikeyRouteTableBindings, apiKey)
	assert.Len(t, conf.ApikeyRouteTableBindings[apiKey], 3)
}

func TestAIRouteExporter_AIRouteGenerator_NoGlobal(t *testing.T) {
	setupState()
	ctx := context.Background()

	apiKey := "ak-key-1"
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *icluster_conf.APIKeyFilter) ([]*icluster_conf.APIKeyParam, error) {
			return []*icluster_conf.APIKeyParam{
				{
					ID:  lib.PString("key-1"),
					Key: &apiKey,
				},
			}, nil
		},
	}

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
			return nil, nil
		},
	}

	routeRulesStore := &fakeRouteRulesStorager{
		fetchFn: func(ctx context.Context, ruleType string, owner *string) (*shared.RouteRulesParam, error) {
			return nil, nil
		},
	}

	e := NewAIRouteExporter(
		apiKeyStore,
		entityStore,
		routeRulesStore,
		iversion_control.NewVersionControllerManager(&fakeTxn{}, &fakeVersionControlStorager{}),
	)

	data, err := e.AIRouteGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*AiRouteDataExport)
	assert.Empty(t, conf.RouteRules)
	assert.Empty(t, conf.ApikeyRouteTableBindings)
}

func TestAIRouteExporter_AIRouteGenerator_FetchEntitiesError(t *testing.T) {
	setupState()
	ctx := context.Background()

	entityStore := &fakeEntityStorager{
		listFn: func(ctx context.Context, filter *quota.EntityFilter) ([]*quota.EntityParam, error) {
			return nil, errors.New("entity fetch failed")
		},
	}

	e := newTestAIRouteExporter()
	e.entityStorager = entityStore

	data, err := e.AIRouteGenerator(ctx)
	require.Error(t, err)
	assert.Nil(t, data)
}

func TestConvertToRouteTableExport(t *testing.T) {
	name := "rule-1"
	cond := "default_t()"
	cluster := "cluster-1"
	model := "gpt-4"
	weight := 100
	fallbackCluster := "cluster-2"
	fallbackModel := "gpt-3.5"

	param := &shared.RouteRulesParam{
		Enabled: lib.PBool(true),
		Rules: []*shared.AiRouteRuleParam{
			{
				Name: &name,
				Cond: &cond,
				Targets: []*shared.AiRouteTargetParam{
					{
						ClusterName: &cluster,
						Model:       &model,
						Weight:      &weight,
					},
				},
				Fallbacks: []*shared.AiRouteFallbackParam{
					{
						ClusterName: &fallbackCluster,
						Model:       &fallbackModel,
					},
				},
			},
		},
	}

	export := convertToRouteTableExport(shared.RouteRulesTypeAPIKey, "owner-1", param.Rules)
	assert.Equal(t, shared.RouteRulesTypeAPIKey, export.Type)
	assert.Equal(t, "owner-1", export.Owner)
	require.Len(t, export.Rules, 1)
	assert.Equal(t, "rule-1", export.Rules[0].Name)
	assert.Equal(t, "default_t()", export.Rules[0].Cond)
	require.Len(t, export.Rules[0].Targets, 1)
	assert.Equal(t, "cluster-1", export.Rules[0].Targets[0].ClusterName)
	assert.Equal(t, "gpt-4", export.Rules[0].Targets[0].Model)
	assert.Equal(t, 100, export.Rules[0].Targets[0].Weight)
	require.Len(t, export.Rules[0].Fallbacks, 1)
	assert.Equal(t, "cluster-2", export.Rules[0].Fallbacks[0].ClusterName)
	assert.Equal(t, "gpt-3.5", export.Rules[0].Fallbacks[0].Model)
}

func TestConvertToRouteTableExport_NilPointers(t *testing.T) {
	param := &shared.RouteRulesParam{
		Rules: []*shared.AiRouteRuleParam{
			{},
		},
	}

	export := convertToRouteTableExport(shared.RouteRulesTypeGlobal, "global", param.Rules)
	require.Len(t, export.Rules, 1)
	assert.Empty(t, export.Rules[0].Name)
	assert.Empty(t, export.Rules[0].Cond)
	assert.Empty(t, export.Rules[0].Targets)
	assert.Empty(t, export.Rules[0].Fallbacks)
}
