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

package api_key

import (
	"context"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type quotaPlanAuditCall struct {
	action  string
	ownerID string
	planID  int64
	err     error
}

type fakeQuotaPlanAuditor struct {
	calls []quotaPlanAuditCall
}

func (f *fakeQuotaPlanAuditor) AuditQuotaPlanCreate(ctx context.Context, param *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, quotaPlanAuditCall{action: "create", ownerID: owner.ID, planID: planID, err: err})
}

func (f *fakeQuotaPlanAuditor) AuditQuotaPlanUpdate(ctx context.Context, oldPlan, param *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, quotaPlanAuditCall{action: "update", ownerID: owner.ID, planID: planID, err: err})
}

func (f *fakeQuotaPlanAuditor) AuditQuotaPlanDelete(ctx context.Context, oldPlan *shared.QuotaPlanParam, planID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, quotaPlanAuditCall{action: "delete", ownerID: owner.ID, planID: planID, err: err})
}

var _ shared.QuotaPlanAuditor = (*fakeQuotaPlanAuditor)(nil)

type rateLimitPolicyAuditCall struct {
	action   string
	ownerID  string
	policyID int64
	err      error
}

type fakeRateLimitPolicyAuditor struct {
	calls []rateLimitPolicyAuditCall
}

func (f *fakeRateLimitPolicyAuditor) AuditRateLimitPolicyCreate(ctx context.Context, param *shared.RateLimitPolicyParam, policyID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, rateLimitPolicyAuditCall{action: "create", ownerID: owner.ID, policyID: policyID, err: err})
}

func (f *fakeRateLimitPolicyAuditor) AuditRateLimitPolicyUpdate(ctx context.Context, oldPolicy, param *shared.RateLimitPolicyParam, policyID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, rateLimitPolicyAuditCall{action: "update", ownerID: owner.ID, policyID: policyID, err: err})
}

func (f *fakeRateLimitPolicyAuditor) AuditRateLimitPolicyDelete(ctx context.Context, oldPolicy *shared.RateLimitPolicyParam, policyID int64, owner shared.ResourceOwner, err error) {
	f.calls = append(f.calls, rateLimitPolicyAuditCall{action: "delete", ownerID: owner.ID, policyID: policyID, err: err})
}

var _ shared.RateLimitPolicyAuditor = (*fakeRateLimitPolicyAuditor)(nil)

// TestAPIKeyManager_NestedAudit verifies nested quota-plan and
// rate-limit-policy writes on the API key lifecycle emit audit entries
// carrying the owning API key's business ID as resource_parent_id
// (issue #161).
func TestAPIKeyManager_NestedAudit(t *testing.T) {
	ctx := context.Background()
	productName := "product-1"

	newManager := func(keyStore *fakeAPIKeyStorager, quotaPlanStore *fakeQuotaPlanStorager,
		rateLimitStore *fakeRateLimitPolicyStorager) (*APIKeyManager, *fakeQuotaPlanAuditor, *fakeRateLimitPolicyAuditor) {
		auditor := &fakeQuotaPlanAuditor{}
		rlAuditor := &fakeRateLimitPolicyAuditor{}
		m := NewAPIKeyManager(&fakeTxn{}, keyStore, quotaPlanStore, rateLimitStore, &fakeRouteRulesStorager{}, &fakeEntityStorager{}, nil)
		m.SetQuotaPlanAuditor(auditor)
		m.SetRateLimitPolicyAuditor(rlAuditor)
		return m, auditor, rlAuditor
	}

	t.Run("create emits create audit with api key parent", func(t *testing.T) {
		keyID := "ak-940"
		keyStore := &fakeAPIKeyStorager{
			createAPIKeyFn: func(ctx context.Context, param *APIKeyParam) (int64, error) { return 100, nil },
		}
		quotaPlanStore := &fakeQuotaPlanStorager{
			createQuotaPlanFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) { return 200, nil },
		}
		rateLimitStore := &fakeRateLimitPolicyStorager{
			createRateLimitPolicyFn: func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) { return 300, nil },
		}
		m, auditor, rlAuditor := newManager(keyStore, quotaPlanStore, rateLimitStore)

		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:              &keyID,
			ProductName:     &productName,
			QuotaPlan:       &shared.QuotaPlanParam{Quota: lib.PFloat64(222)},
			RateLimitPolicy: &shared.RateLimitPolicyParam{},
		})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "ak-940", auditor.calls[0].ownerID)
		assert.NoError(t, auditor.calls[0].err)

		require.Len(t, rlAuditor.calls, 1)
		assert.Equal(t, "create", rlAuditor.calls[0].action)
		assert.Equal(t, int64(300), rlAuditor.calls[0].policyID)
		assert.Equal(t, "ak-940", rlAuditor.calls[0].ownerID)
	})

	t.Run("create failure emits failed audit with api key parent", func(t *testing.T) {
		keyID := "ak-940"
		keyStore := &fakeAPIKeyStorager{}
		quotaPlanStore := &fakeQuotaPlanStorager{
			createQuotaPlanFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				return 0, errors.New("quota plan create failed")
			},
		}
		m, auditor, _ := newManager(keyStore, quotaPlanStore, &fakeRateLimitPolicyStorager{})

		err := m.CreateAPIKey(ctx, &APIKeyParam{
			ID:          &keyID,
			ProductName: &productName,
			QuotaPlan:   &shared.QuotaPlanParam{Quota: lib.PFloat64(222)},
		})
		require.Error(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(0), auditor.calls[0].planID)
		assert.Equal(t, "ak-940", auditor.calls[0].ownerID)
		assert.Error(t, auditor.calls[0].err)
	})

	t.Run("update existing plan emits update audit", func(t *testing.T) {
		keyID := "ak-940"
		planID := int64(200)
		keyStore := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{ID: &keyID, QuotaPlanID: &planID}}, nil
			},
			updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
				return 1, nil
			},
		}
		m, auditor, _ := newManager(keyStore, &fakeQuotaPlanStorager{}, &fakeRateLimitPolicyStorager{})

		err := m.UpdateAPIKey(ctx, &APIKeyFilter{ID: &keyID}, &APIKeyParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PFloat64(333)},
		})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "update", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "ak-940", auditor.calls[0].ownerID)
	})

	t.Run("update without existing plan emits create audit", func(t *testing.T) {
		keyID := "ak-940"
		keyStore := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{ID: &keyID}}, nil
			},
			updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
				return 1, nil
			},
		}
		quotaPlanStore := &fakeQuotaPlanStorager{
			createQuotaPlanFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) { return 201, nil },
		}
		m, auditor, _ := newManager(keyStore, quotaPlanStore, &fakeRateLimitPolicyStorager{})

		err := m.UpdateAPIKey(ctx, &APIKeyFilter{ID: &keyID}, &APIKeyParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PFloat64(333)},
		})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(201), auditor.calls[0].planID)
		assert.Equal(t, "ak-940", auditor.calls[0].ownerID)
	})

	t.Run("delete emits delete audit with api key parent", func(t *testing.T) {
		keyID := "ak-940"
		planID := int64(200)
		policyID := int64(300)
		keyStore := &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{ID: &keyID, QuotaPlanID: &planID, RateLimitPolicyID: &policyID}}, nil
			},
			deleteAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter) error { return nil },
		}
		m, auditor, rlAuditor := newManager(keyStore, &fakeQuotaPlanStorager{}, &fakeRateLimitPolicyStorager{})

		err := m.DeleteAPIKey(ctx, &APIKeyFilter{ID: &keyID})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "delete", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "ak-940", auditor.calls[0].ownerID)

		require.Len(t, rlAuditor.calls, 1)
		assert.Equal(t, "delete", rlAuditor.calls[0].action)
		assert.Equal(t, int64(300), rlAuditor.calls[0].policyID)
		assert.Equal(t, "ak-940", rlAuditor.calls[0].ownerID)
	})
}
