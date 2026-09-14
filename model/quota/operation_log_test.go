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

package quota

import (
	"context"
	"errors"
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

// TestQuotaPlanManager_OwnerPropagation verifies the owner business ID is
// written into resource_parent_id for both manager methods and the nested
// audit entry points (issue #161).
func TestQuotaPlanManager_OwnerPropagation(t *testing.T) {
	ctx := context.Background()
	owner := shared.ResourceOwner{Type: "entity", ID: "entity-940"}

	t.Run("CreateQuotaPlan success", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		planStore := &fakeQuotaPlanStorager{
			createFn: func(ctx context.Context, param *QuotaPlanParam) (int64, error) {
				return 7, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		id, err := m.CreateQuotaPlan(ctx, &QuotaPlanParam{Quota: lib.PFloat64(100)}, owner)
		require.NoError(t, err)
		assert.Equal(t, int64(7), id)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, string(ioperlog.ActionCreate), entry.Action)
		assert.Equal(t, string(ioperlog.ResourceTypeQuotaPlan), entry.ResourceType)
		assert.Equal(t, "7", entry.ResourceID)
		assert.Equal(t, "entity-940", entry.ResourceParentID)
		assert.Equal(t, ioperlog.StatusSuccess, entry.Status)
	})

	t.Run("AuditQuotaPlanCreate failure keeps owner", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		m := NewQuotaPlanManager(&fakeTxn{}, &fakeQuotaPlanStorager{}, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		writeErr := errors.New("create failed")
		m.AuditQuotaPlanCreate(ctx, &shared.QuotaPlanParam{Quota: lib.PFloat64(100)}, 0, owner, writeErr)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, ioperlog.StatusFailed, entry.Status)
		assert.Equal(t, "", entry.ResourceID)
		assert.Equal(t, "entity-940", entry.ResourceParentID)
	})

	t.Run("AuditQuotaPlanUpdate and Delete carry owner", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		m := NewQuotaPlanManager(&fakeTxn{}, &fakeQuotaPlanStorager{}, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		oldPlan := &shared.QuotaPlanParam{Quota: lib.PFloat64(100)}
		newPlan := &shared.QuotaPlanParam{Quota: lib.PFloat64(200)}
		m.AuditQuotaPlanUpdate(ctx, oldPlan, newPlan, 7, owner, nil)
		m.AuditQuotaPlanDelete(ctx, oldPlan, 7, owner, nil)

		require.Len(t, recorder.entries, 2)
		assert.Equal(t, string(ioperlog.ActionUpdate), recorder.entries[0].Action)
		assert.Equal(t, "7", recorder.entries[0].ResourceID)
		assert.Equal(t, "entity-940", recorder.entries[0].ResourceParentID)
		assert.Equal(t, string(ioperlog.ActionDelete), recorder.entries[1].Action)
		assert.Equal(t, "entity-940", recorder.entries[1].ResourceParentID)
	})

	t.Run("ResetBalance success carries owner", func(t *testing.T) {
		recorder := &fakeOperationLogRecorder{}
		planStore := &fakeQuotaPlanStorager{
			fetchFn: func(ctx context.Context, filter *QuotaPlanFilter) (*QuotaPlanParam, error) {
				return &QuotaPlanParam{ID: lib.PInt64(1), Quota: lib.PFloat64(1000)}, nil
			},
			updateFn: func(ctx context.Context, filter *QuotaPlanFilter, param *QuotaPlanParam) (int64, error) {
				return 1, nil
			},
		}
		m := NewQuotaPlanManager(&fakeTxn{}, planStore, nil, nil, nil)
		m.SetOperationLogManager(recorder)

		err := m.ResetBalance(ctx, 1, nil, true, owner)
		require.NoError(t, err)

		require.Len(t, recorder.entries, 1)
		entry := recorder.entries[0]
		assert.Equal(t, string(ioperlog.ActionReset), entry.Action)
		assert.Equal(t, "1", entry.ResourceID)
		assert.Equal(t, "entity-940", entry.ResourceParentID)
	})
}
