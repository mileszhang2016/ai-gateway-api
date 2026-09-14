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
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOperationLogRecorder struct {
	entries []*ioperlog.OperationLogEntry
}

func (f *fakeOperationLogRecorder) Record(ctx context.Context, entry *ioperlog.OperationLogEntry) {
	f.entries = append(f.entries, entry)
}

// TestRateLimitPolicyManager_OwnerPropagation verifies the owner business ID is
// written into resource_parent_id for manager methods and the nested audit
// entry points (issue #161).
func TestRateLimitPolicyManager_OwnerPropagation(t *testing.T) {
	ctx := context.Background()
	owner := shared.ResourceOwner{Type: "entity", ID: "entity-940"}

	t.Run("CreateRateLimitPolicy success", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		m := NewRateLimitPolicyManager(&fakeTxn{}, &fakeRateLimitPolicyStorager{
			createFn: func(ctx context.Context, param *RateLimitPolicyParam) (int64, error) {
				return 11, nil
			},
		}, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		id, err := m.CreateRateLimitPolicy(ctx, &RateLimitPolicyParam{Enabled: lib.PBool(true)}, owner)
		require.NoError(t, err)
		assert.Equal(t, int64(11), id)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, string(ioperlog.ActionCreate), entry.Action)
		assert.Equal(t, string(ioperlog.ResourceTypeRateLimitPolicy), entry.ResourceType)
		assert.Equal(t, "11", entry.ResourceID)
		assert.Equal(t, "entity-940", entry.ResourceParentID)
	})

	t.Run("AuditRateLimitPolicyUpdate and Delete carry owner", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		m := NewRateLimitPolicyManager(&fakeTxn{}, &fakeRateLimitPolicyStorager{}, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		oldPolicy := &shared.RateLimitPolicyParam{Enabled: lib.PBool(true)}
		newPolicy := &shared.RateLimitPolicyParam{Enabled: lib.PBool(false)}
		m.AuditRateLimitPolicyUpdate(ctx, oldPolicy, newPolicy, 11, owner, nil)
		m.AuditRateLimitPolicyDelete(ctx, oldPolicy, 11, owner, nil)

		require.Len(t, recorder.entries, 2)
		assert.Equal(t, string(ioperlog.ActionUpdate), recorder.entries[0].Action)
		assert.Equal(t, "11", recorder.entries[0].ResourceID)
		assert.Equal(t, "entity-940", recorder.entries[0].ResourceParentID)
		assert.Equal(t, string(ioperlog.ActionDelete), recorder.entries[1].Action)
		assert.Equal(t, "entity-940", recorder.entries[1].ResourceParentID)
	})
}
