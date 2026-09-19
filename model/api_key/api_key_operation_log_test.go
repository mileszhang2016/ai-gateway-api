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
