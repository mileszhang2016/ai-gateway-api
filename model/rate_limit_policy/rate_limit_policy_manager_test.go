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

package rate_limit_policy

import (
	"context"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/api_key"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/entity"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/iversion_control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitPolicyManager_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateRateLimitPolicy", func(t *testing.T) {
		store := &fakeRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *RateLimitPolicyParam) (int64, error) {
				return 11, nil
			},
		}
		m := NewRateLimitPolicyManager(&fakeTxn{}, store, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, nil)

		id, err := m.CreateRateLimitPolicy(ctx, &RateLimitPolicyParam{Enabled: lib.PBool(true)})
		require.NoError(t, err)
		assert.Equal(t, int64(11), id)
	})

	t.Run("FetchRateLimitPolicy", func(t *testing.T) {
		store := &fakeRateLimitPolicyStorager{
			fetchFn: func(ctx context.Context, filter *RateLimitPolicyFilter) (*RateLimitPolicyParam, error) {
				return &RateLimitPolicyParam{Enabled: lib.PBool(true)}, nil
			},
		}
		m := NewRateLimitPolicyManager(&fakeTxn{}, store, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, nil)

		policy, err := m.FetchRateLimitPolicy(ctx, &RateLimitPolicyFilter{ID: lib.PInt64(11)})
		require.NoError(t, err)
		require.NotNil(t, policy)
		assert.True(t, *policy.Enabled)
	})

	t.Run("FetchRateLimitPolicyList", func(t *testing.T) {
		store := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{{Enabled: lib.PBool(true)}}, nil
			},
		}
		m := NewRateLimitPolicyManager(&fakeTxn{}, store, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, nil)

		list, err := m.FetchRateLimitPolicyList(ctx, &RateLimitPolicyFilter{})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("UpdateRateLimitPolicy", func(t *testing.T) {
		store := &fakeRateLimitPolicyStorager{
			updateFn: func(ctx context.Context, filter *RateLimitPolicyFilter, param *RateLimitPolicyParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewRateLimitPolicyManager(&fakeTxn{}, store, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, nil)

		affected, err := m.UpdateRateLimitPolicy(ctx, &RateLimitPolicyFilter{ID: lib.PInt64(11)}, &RateLimitPolicyParam{Enabled: lib.PBool(false)})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("DeleteRateLimitPolicy", func(t *testing.T) {
		store := &fakeRateLimitPolicyStorager{
			deleteFn: func(ctx context.Context, filter *RateLimitPolicyFilter) error {
				return nil
			},
		}
		m := NewRateLimitPolicyManager(&fakeTxn{}, store, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, nil)

		require.NoError(t, m.DeleteRateLimitPolicy(ctx, &RateLimitPolicyFilter{ID: lib.PInt64(11)}))
	})
}

func TestRateLimitPolicyManager_ConfigExport(t *testing.T) {
	ctx := context.Background()

	buildManager := func(store *fakeRateLimitPolicyStorager, apiKeyStore *fakeAPIKeyStorager, entityStore *fakeEntityStorager, versionStorager *fakeVersionControlStorager) *RateLimitPolicyManager {
		var vcm *iversion_control.VersionControlManager
		if versionStorager != nil {
			vcm = iversion_control.NewVersionControllerManager(&fakeTxn{}, versionStorager)
		}
		return NewRateLimitPolicyManager(&fakeTxn{}, store, apiKeyStore, entityStore, vcm)
	}

	t.Run("returns nil when version unchanged", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return iversion_control.ZeroVersion, nil
			},
		}
		m := buildManager(&fakeRateLimitPolicyStorager{}, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, versionStore)

		conf, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		assert.Nil(t, conf)
	})

	t.Run("returns config when version changed", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "20250101000000", nil
			},
		}
		m := buildManager(&fakeRateLimitPolicyStorager{}, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, versionStore)

		conf, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.NoError(t, err)
		require.NotNil(t, conf)
		assert.Equal(t, "20250101000000", conf.Version)
		assert.NotNil(t, conf.Config["AI_product"])
	})

	t.Run("version control error propagates", func(t *testing.T) {
		versionStore := &fakeVersionControlStorager{
			upsertFn: func(ctx context.Context, css *iversion_control.ExportData) (string, error) {
				return "", errors.New("version db error")
			},
		}
		m := buildManager(&fakeRateLimitPolicyStorager{}, &fakeAPIKeyStorager{}, &fakeEntityStorager{}, versionStore)

		_, err := m.ConfigExport(ctx, iversion_control.ZeroVersion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version db error")
	})
}

func TestRateLimitPolicyManager_RateLimitPolicyGenerator(t *testing.T) {
	ctx := context.Background()

	buildManager := func(store *fakeRateLimitPolicyStorager, apiKeyStore *fakeAPIKeyStorager, entityStore *fakeEntityStorager) *RateLimitPolicyManager {
		return NewRateLimitPolicyManager(&fakeTxn{}, store, apiKeyStore, entityStore, nil)
	}

	t.Run("empty api keys", func(t *testing.T) {
		m := buildManager(&fakeRateLimitPolicyStorager{}, &fakeAPIKeyStorager{}, &fakeEntityStorager{})

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		require.NotNil(t, data)
		conf, ok := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)
		require.True(t, ok)
		assert.Empty(t, conf.RateLimitPolicies)
		assert.Empty(t, conf.ApikeyRateLimitPolicyBindings)
	})

	t.Run("api key with direct policy", func(t *testing.T) {
		apiKey := "ak-1"
		policyID := int64(101)
		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return []*api_key.APIKeyParam{
					{
						Key:               &apiKey,
						RateLimitPolicyID: &policyID,
					},
				}, nil
			},
		}
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{
					{
						ID:      &policyID,
						Enabled: lib.PBool(true),
						TpmConfigs: []TPMConfig{
							{Name: "tpm-1", Model: "*", WindowMinutes: 1, MaxTokens: 100},
						},
						RpmConfigs: []RPMConfig{
							{Name: "rpm-1", Model: "*", WindowMinutes: 1, MaxRequests: 10},
						},
					},
				}, nil
			},
		}
		m := buildManager(policyStore, apiKeyStore, &fakeEntityStorager{})

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

		assert.Contains(t, conf.RateLimitPolicies, "rlp-101")
		assert.Contains(t, conf.ApikeyRateLimitPolicyBindings, apiKey)
		assert.Equal(t, []string{"rlp-101"}, conf.ApikeyRateLimitPolicyBindings[apiKey])

		policy := conf.RateLimitPolicies["rlp-101"]
		require.NotNil(t, policy.Rules)
		assert.Len(t, policy.Rules.TPM, 1)
		assert.Equal(t, []string{"*"}, policy.Rules.TPM[0].Models)
	})

	t.Run("disabled policy is skipped", func(t *testing.T) {
		apiKey := "ak-1"
		policyID := int64(101)
		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return []*api_key.APIKeyParam{
					{Key: &apiKey, RateLimitPolicyID: &policyID},
				}, nil
			},
		}
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{
					{ID: &policyID, Enabled: lib.PBool(false)},
				}, nil
			},
		}
		m := buildManager(policyStore, apiKeyStore, &fakeEntityStorager{})

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

		assert.Empty(t, conf.RateLimitPolicies)
		assert.Empty(t, conf.ApikeyRateLimitPolicyBindings)
	})

	t.Run("api key without key value is skipped", func(t *testing.T) {
		policyID := int64(101)
		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return []*api_key.APIKeyParam{
					{RateLimitPolicyID: &policyID, Key: nil},
				}, nil
			},
		}
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{{ID: &policyID, Enabled: lib.PBool(true)}}, nil
			},
		}
		m := buildManager(policyStore, apiKeyStore, &fakeEntityStorager{})

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

		assert.Empty(t, conf.RateLimitPolicies)
	})

	t.Run("entity hierarchy policies are collected", func(t *testing.T) {
		apiKey := "ak-1"
		entityID := "ent-1"
		parentID := "ent-parent"
		apiPolicyID := int64(101)
		entityPolicyID := int64(102)
		parentPolicyID := int64(103)

		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return []*api_key.APIKeyParam{
					{Key: &apiKey, EntityID: &entityID, RateLimitPolicyID: &apiPolicyID},
				}, nil
			},
		}
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{
					{ID: &apiPolicyID, Enabled: lib.PBool(true)},
					{ID: &entityPolicyID, Enabled: lib.PBool(true)},
					{ID: &parentPolicyID, Enabled: lib.PBool(true)},
				}, nil
			},
		}
		entityStore := &fakeEntityStorager{
			fetchFn: func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
				if filter.EntityID != nil && *filter.EntityID == entityID {
					return &entity.EntityParam{
						EntityID:          &entityID,
						ParentID:          &parentID,
						RateLimitPolicyID: &entityPolicyID,
					}, nil
				}
				if filter.EntityID != nil && *filter.EntityID == parentID {
					return &entity.EntityParam{
						EntityID:          &parentID,
						RateLimitPolicyID: &parentPolicyID,
					}, nil
				}
				return nil, nil
			},
		}
		m := buildManager(policyStore, apiKeyStore, entityStore)

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

		bindings := conf.ApikeyRateLimitPolicyBindings[apiKey]
		require.Len(t, bindings, 3)
		assert.Equal(t, []string{"rlp-101", "rlp-102", "rlp-103"}, bindings)
	})

	t.Run("model wildcard becomes explicit when specified", func(t *testing.T) {
		apiKey := "ak-1"
		policyID := int64(101)
		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return []*api_key.APIKeyParam{{Key: &apiKey, RateLimitPolicyID: &policyID}}, nil
			},
		}
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return []*RateLimitPolicyParam{{
					ID:      &policyID,
					Enabled: lib.PBool(true),
					TpmConfigs: []TPMConfig{
						{Name: "tpm-1", Model: "gpt-4", WindowMinutes: 1, MaxTokens: 100},
					},
				}}, nil
			},
		}
		m := buildManager(policyStore, apiKeyStore, &fakeEntityStorager{})

		data, err := m.RateLimitPolicyGenerator(ctx)
		require.NoError(t, err)
		conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

		policy := conf.RateLimitPolicies["rlp-101"]
		require.Len(t, policy.Rules.TPM, 1)
		assert.Equal(t, []string{"gpt-4"}, policy.Rules.TPM[0].Models)
	})

	t.Run("api key fetch error propagates", func(t *testing.T) {
		apiKeyStore := &fakeAPIKeyStorager{
			fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
				return nil, errors.New("api key db error")
			},
		}
		m := buildManager(&fakeRateLimitPolicyStorager{}, apiKeyStore, &fakeEntityStorager{})

		_, err := m.RateLimitPolicyGenerator(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fetch api keys error")
	})

	t.Run("policy fetch error propagates", func(t *testing.T) {
		policyStore := &fakeRateLimitPolicyStorager{
			listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
				return nil, errors.New("policy db error")
			},
		}
		m := buildManager(policyStore, &fakeAPIKeyStorager{}, &fakeEntityStorager{})

		_, err := m.RateLimitPolicyGenerator(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fetch rate limit policies error")
	})
}
func TestRateLimitPolicyManager_RateLimitPolicyGenerator_MaxConcurrency(t *testing.T) {
	ctx := context.Background()
	apiKey := "ak-1"
	policyID := int64(101)
	apiKeyStore := &fakeAPIKeyStorager{
		fetchListFn: func(ctx context.Context, filter *api_key.APIKeyFilter) ([]*api_key.APIKeyParam, error) {
			return []*api_key.APIKeyParam{{Key: &apiKey, RateLimitPolicyID: &policyID}}, nil
		},
	}
	policyStore := &fakeRateLimitPolicyStorager{
		listFn: func(ctx context.Context, filter *RateLimitPolicyFilter) ([]*RateLimitPolicyParam, error) {
			return []*RateLimitPolicyParam{{
				ID:             &policyID,
				Enabled:        lib.PBool(true),
				MaxConcurrency: lib.PInt(20),
			}}, nil
		},
	}
	m := NewRateLimitPolicyManager(&fakeTxn{}, policyStore, apiKeyStore, &fakeEntityStorager{}, nil)

	data, err := m.RateLimitPolicyGenerator(ctx)
	require.NoError(t, err)
	conf := data.DataWithoutVersion.(*ExportRateLimitPolicyConfig)

	policy := conf.RateLimitPolicies["rlp-101"]
	require.NotNil(t, policy.Rules)
	assert.Equal(t, 20, policy.Rules.MaxConcurrency)
}

func TestRateLimitPolicyManager_fetchEntityRateLimitPolicyIDs_ParentNotFound(t *testing.T) {
	ctx := context.Background()
	entityID := "ent-1"
	parentID := "parent-1"
	policyID := int64(101)

	entityStore := &fakeEntityStorager{
		fetchFn: func(ctx context.Context, filter *entity.EntityFilter) (*entity.EntityParam, error) {
			if filter.EntityID != nil && *filter.EntityID == entityID {
				return &entity.EntityParam{EntityID: &entityID, ParentID: &parentID, RateLimitPolicyID: &policyID}, nil
			}
			return nil, nil
		},
	}
	m := NewRateLimitPolicyManager(&fakeTxn{}, &fakeRateLimitPolicyStorager{}, &fakeAPIKeyStorager{}, entityStore, nil)

	ids, err := m.fetchEntityRateLimitPolicyIDs(ctx, &entity.EntityParam{EntityID: &entityID, ParentID: &parentID, RateLimitPolicyID: &policyID})
	require.NoError(t, err)
	assert.Equal(t, []int64{policyID}, ids)
}
