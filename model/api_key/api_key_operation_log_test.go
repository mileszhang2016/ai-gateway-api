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
	"encoding/json"
	"fmt"
	"testing"

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

var _ ioperlog.OperationLogRecorder = (*fakeOperationLogRecorder)(nil)

// TestAPIKeyManager_CreateFailureAuditMasksKey reproduces issue #185: a
// duplicate-key create failure must not persist the raw API-Key value in the
// failed audit entry's error_msg, nor fan it out to the nested quota-plan and
// rate-limit-policy audit entries.
func TestAPIKeyManager_CreateFailureAuditMasksKey(t *testing.T) {
	ctx := context.Background()
	rawKey := "testproduct-abcdef012345"

	store := &fakeAPIKeyStorager{
		fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
			if filter.Key != nil {
				return []*APIKeyParam{{Key: ptrString(rawKey)}}, nil
			}
			return nil, nil
		},
		fetchAPIKeyTokenListFn: func(ctx context.Context, filter *APIKeyTokenFilter) ([]*APIKeyTokenParam, error) {
			return nil, nil
		},
	}
	recorder := &fakeOperationLogRecorder{}
	quotaAuditor := &fakeQuotaPlanAuditor{}
	rateLimitAuditor := &fakeRateLimitPolicyAuditor{}

	m := newAPIKeyManager(store)
	m.SetOperationLogManager(recorder)
	m.SetQuotaPlanAuditor(quotaAuditor)
	m.SetRateLimitPolicyAuditor(rateLimitAuditor)

	err := m.CreateAPIKey(ctx, &APIKeyParam{
		ID:              ptrString("id1"),
		ProductName:     ptrString("test"),
		Key:             ptrString(rawKey),
		Description:     ptrString("dup create"),
		QuotaPlan:       &shared.QuotaPlanParam{},
		RateLimitPolicy: &shared.RateLimitPolicyParam{},
	})
	require.Error(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionCreate), entry.Action)
	assert.Equal(t, ioperlog.StatusFailed, entry.Status)
	assert.NotContains(t, entry.ErrorMsg, rawKey)
	assert.Contains(t, entry.ErrorMsg, ioperlog.MaskAPIKeyToken(rawKey))

	// The nested auditors receive the same already-masked error (issue #185
	// fan-out): their audit entries must not carry the raw key either.
	require.Len(t, quotaAuditor.calls, 1)
	require.Error(t, quotaAuditor.calls[0].err)
	assert.NotContains(t, quotaAuditor.calls[0].err.Error(), rawKey)
	assert.Contains(t, quotaAuditor.calls[0].err.Error(), ioperlog.MaskAPIKeyToken(rawKey))
	require.Len(t, rateLimitAuditor.calls, 1)
	require.Error(t, rateLimitAuditor.calls[0].err)
	assert.NotContains(t, rateLimitAuditor.calls[0].err.Error(), rawKey)

	// Whole-entry serialization (error_msg + masked change_summary) holds no
	// raw key, mirroring the SC2101-TC046 secret-absent assertion.
	serialized, marshalErr := json.Marshal(entry)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(serialized), rawKey)
}

// TestAPIKeyManager_UpdateFailureAuditMasksKey verifies the sink redaction in
// recordAPIKeyOperation: even an error message that echoes the stored key
// (e.g. a DAO driver error) is masked via oldAPIKey.Key before persistence.
func TestAPIKeyManager_UpdateFailureAuditMasksKey(t *testing.T) {
	ctx := context.Background()
	rawKey := "testproduct-abcdef012345"

	store := &fakeAPIKeyStorager{
		fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
			return []*APIKeyParam{{ID: ptrString("id1"), Key: ptrString(rawKey)}}, nil
		},
		updateAPIKeyFn: func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
			return 0, fmt.Errorf("Error 1062 (23000): Duplicate entry '%s' for key 'api_key.key'", rawKey)
		},
	}
	recorder := &fakeOperationLogRecorder{}

	m := newAPIKeyManager(store)
	m.SetOperationLogManager(recorder)

	err := m.UpdateAPIKey(ctx, &APIKeyFilter{ID: ptrString("id1")},
		&APIKeyParam{Description: ptrString("new desc")})
	require.Error(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionUpdate), entry.Action)
	assert.Equal(t, ioperlog.StatusFailed, entry.Status)
	assert.NotContains(t, entry.ErrorMsg, rawKey)
	assert.Contains(t, entry.ErrorMsg, ioperlog.MaskAPIKeyToken(rawKey))
}

// TestAPIKeyParamToMap_DropsOmittedFields guards issue #201 at map
// construction: fields omitted from a partial update (nil pointers without
// omitempty, e.g. id/enabled/key) must not materialize as null entries.
func TestAPIKeyParamToMap_DropsOmittedFields(t *testing.T) {
	m := apiKeyParamToMap(&APIKeyParam{Description: ptrString("only-desc")})
	require.NotNil(t, m)
	assert.Equal(t, map[string]interface{}{"description": "only-desc"}, m)

	assert.Nil(t, apiKeyParamToMap(nil))
}

// TestAPIKeyManager_UpdateSuccessAuditDiffKeys verifies the issue #201 oracle
// assertion path: a successful PATCH submitting only "description" must
// produce diff_keys == ["description"], and change_summary.after must not
// carry the unsubmitted id/enabled/key fields. An explicit enabled=false is
// a real change and must stay in diff_keys; the failure path applies the
// same after construction.
func TestAPIKeyManager_UpdateSuccessAuditDiffKeys(t *testing.T) {
	ctx := context.Background()

	newStorager := func(updateFn func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error)) *fakeAPIKeyStorager {
		return &fakeAPIKeyStorager{
			fetchAPIKeyListFn: func(ctx context.Context, filter *APIKeyFilter) ([]*APIKeyParam, error) {
				return []*APIKeyParam{{
					ID:          ptrString("id1"),
					Key:         ptrString("testproduct-abcdef012345"),
					Enable:      ptrBool(true),
					Description: ptrString("old desc"),
				}}, nil
			},
			updateAPIKeyFn: updateFn,
		}
	}
	okUpdate := func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
		return 1, nil
	}

	recorder := &fakeOperationLogRecorder{}
	m := newAPIKeyManager(newStorager(okUpdate))
	m.SetOperationLogManager(recorder)

	// 1. Partial update: only description submitted.
	err := m.UpdateAPIKey(ctx, &APIKeyFilter{ID: ptrString("id1")},
		&APIKeyParam{Description: ptrString("new desc")})
	require.NoError(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionUpdate), entry.Action)
	assert.Equal(t, ioperlog.StatusSuccess, entry.Status)
	assert.Equal(t, []string{"description"}, entry.ChangeSummary["diff_keys"])

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "after should be a map")
	assert.NotContains(t, after, "id")
	assert.NotContains(t, after, "enabled")
	assert.NotContains(t, after, "key")

	before, ok := entry.ChangeSummary["before"].(map[string]interface{})
	require.True(t, ok, "before should be a map")
	assert.Contains(t, before, "id")
	assert.Contains(t, before, "enabled")
	assert.Contains(t, before, "key")

	// 2. Explicit zero value: enabled=false is a submitted change.
	err = m.UpdateAPIKey(ctx, &APIKeyFilter{ID: ptrString("id1")},
		&APIKeyParam{Description: ptrString("new desc 2"), Enable: ptrBool(false)})
	require.NoError(t, err)

	require.Len(t, recorder.entries, 2)
	entry2 := recorder.entries[1]
	assert.Equal(t, []string{"description", "enabled"}, entry2.ChangeSummary["diff_keys"])
	after2, ok := entry2.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok, "after should be a map")
	assert.Equal(t, false, after2["enabled"])

	// 3. Failure path: same after construction, no phantom diff keys.
	failingRecorder := &fakeOperationLogRecorder{}
	failing := newAPIKeyManager(newStorager(
		func(ctx context.Context, filter *APIKeyFilter, param *APIKeyParam) (int64, error) {
			return 0, fmt.Errorf("dao unavailable")
		}))
	failing.SetOperationLogManager(failingRecorder)

	err = failing.UpdateAPIKey(ctx, &APIKeyFilter{ID: ptrString("id1")},
		&APIKeyParam{Description: ptrString("new desc")})
	require.Error(t, err)

	require.Len(t, failingRecorder.entries, 1)
	entry3 := failingRecorder.entries[0]
	assert.Equal(t, ioperlog.StatusFailed, entry3.Status)
	assert.Equal(t, []string{"description"}, entry3.ChangeSummary["diff_keys"])
}
