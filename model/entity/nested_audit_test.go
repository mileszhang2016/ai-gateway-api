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

package entity

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

// TestEntityManager_NestedAudit verifies nested quota-plan and rate-limit-policy
// writes on the entity lifecycle emit audit entries carrying the owning
// entity's business ID as resource_parent_id (issue #161).
func TestEntityManager_NestedAudit(t *testing.T) {
	ctx := context.Background()

	newManager := func(entityStore *fakeEntityStorager, quotaPlanStore *fakeSharedQuotaPlanStorager,
		rateLimitStore *fakeSharedRateLimitPolicyStorager) (*EntityManager, *fakeQuotaPlanAuditor, *fakeRateLimitPolicyAuditor) {
		auditor := &fakeQuotaPlanAuditor{}
		rlAuditor := &fakeRateLimitPolicyAuditor{}
		m := NewEntityManager(&fakeTxn{}, entityStore, &fakeEntityTypeStorager{}, quotaPlanStore, rateLimitStore, &fakeRouteRulesStorager{}, nil)
		m.SetQuotaPlanAuditor(auditor)
		m.SetRateLimitPolicyAuditor(rlAuditor)
		return m, auditor, rlAuditor
	}

	t.Run("create emits create audit with entity parent", func(t *testing.T) {
		entityID := "entity-940"
		entityStore := &fakeEntityStorager{
			fetchFn:  func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) { return nil, nil },
			createFn: func(ctx context.Context, param *EntityParam) (int64, error) { return 100, nil },
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			createFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) { return 200, nil },
		}
		rateLimitStore := &fakeSharedRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *shared.RateLimitPolicyParam) (int64, error) { return 300, nil },
		}
		m, auditor, rlAuditor := newManager(entityStore, quotaPlanStore, rateLimitStore)

		_, err := m.CreateEntity(ctx, &EntityParam{
			EntityID:        &entityID,
			QuotaPlan:       &shared.QuotaPlanParam{Quota: lib.PFloat64(222)},
			RateLimitPolicy: &shared.RateLimitPolicyParam{},
		})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "entity-940", auditor.calls[0].ownerID)
		assert.NoError(t, auditor.calls[0].err)

		require.Len(t, rlAuditor.calls, 1)
		assert.Equal(t, "create", rlAuditor.calls[0].action)
		assert.Equal(t, int64(300), rlAuditor.calls[0].policyID)
		assert.Equal(t, "entity-940", rlAuditor.calls[0].ownerID)
	})

	t.Run("create failure emits failed audit with entity parent", func(t *testing.T) {
		entityID := "entity-940"
		entityStore := &fakeEntityStorager{
			fetchFn:  func(ctx context.Context, filter *EntityFilter) (*EntityParam, error) { return nil, nil },
			createFn: func(ctx context.Context, param *EntityParam) (int64, error) { return 100, nil },
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			createFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) {
				return 0, errors.New("quota plan create failed")
			},
		}
		m, auditor, _ := newManager(entityStore, quotaPlanStore, &fakeSharedRateLimitPolicyStorager{})

		_, err := m.CreateEntity(ctx, &EntityParam{
			EntityID:  &entityID,
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PFloat64(222)},
		})
		require.Error(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(0), auditor.calls[0].planID)
		assert.Equal(t, "entity-940", auditor.calls[0].ownerID)
		assert.Error(t, auditor.calls[0].err)
	})

	t.Run("update existing plan emits update audit", func(t *testing.T) {
		entityID := "entity-940"
		planID := int64(200)
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{EntityID: &entityID, QuotaPlanID: &planID}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{}
		m, auditor, _ := newManager(entityStore, quotaPlanStore, &fakeSharedRateLimitPolicyStorager{})

		affected, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PFloat64(333)},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "update", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "entity-940", auditor.calls[0].ownerID)
	})

	t.Run("update without existing plan emits create audit", func(t *testing.T) {
		entityID := "entity-940"
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				return []*EntityParam{{EntityID: &entityID}}, nil
			},
			updateFn: func(ctx context.Context, filter *EntityFilter, param *EntityParam) (int64, error) {
				return 1, nil
			},
		}
		quotaPlanStore := &fakeSharedQuotaPlanStorager{
			createFn: func(ctx context.Context, param *shared.QuotaPlanParam) (int64, error) { return 201, nil },
		}
		m, auditor, _ := newManager(entityStore, quotaPlanStore, &fakeSharedRateLimitPolicyStorager{})

		_, err := m.UpdateEntity(ctx, &EntityFilter{EntityID: &entityID}, &EntityParam{
			QuotaPlan: &shared.QuotaPlanParam{Quota: lib.PFloat64(333)},
		})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "create", auditor.calls[0].action)
		assert.Equal(t, int64(201), auditor.calls[0].planID)
		assert.Equal(t, "entity-940", auditor.calls[0].ownerID)
	})

	t.Run("delete emits delete audit with entity parent", func(t *testing.T) {
		entityID := "entity-940"
		planID := int64(200)
		policyID := int64(300)
		entityStore := &fakeEntityStorager{
			listFn: func(ctx context.Context, filter *EntityFilter) ([]*EntityParam, error) {
				if filter.ParentID != nil {
					return nil, nil // children check: no children
				}
				return []*EntityParam{{EntityID: &entityID, QuotaPlanID: &planID, RateLimitPolicyID: &policyID}}, nil
			},
			deleteFn: func(ctx context.Context, filter *EntityFilter) error { return nil },
		}
		m, auditor, rlAuditor := newManager(entityStore, &fakeSharedQuotaPlanStorager{}, &fakeSharedRateLimitPolicyStorager{})

		err := m.DeleteEntity(ctx, &EntityFilter{EntityID: &entityID})
		require.NoError(t, err)

		require.Len(t, auditor.calls, 1)
		assert.Equal(t, "delete", auditor.calls[0].action)
		assert.Equal(t, int64(200), auditor.calls[0].planID)
		assert.Equal(t, "entity-940", auditor.calls[0].ownerID)

		require.Len(t, rlAuditor.calls, 1)
		assert.Equal(t, "delete", rlAuditor.calls[0].action)
		assert.Equal(t, int64(300), rlAuditor.calls[0].policyID)
		assert.Equal(t, "entity-940", rlAuditor.calls[0].ownerID)
	})
}
