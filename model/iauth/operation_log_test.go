// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package iauth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/model/ioperlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOperationLogRecorder struct {
	entries []*ioperlog.OperationLogEntry
}

func (r *fakeOperationLogRecorder) Record(ctx context.Context, entry *ioperlog.OperationLogEntry) {
	r.entries = append(r.entries, entry)
}

// issue #162：token 创建审计的 change_summary.after.token 必须为全掩码，
// 不得记录与接口响应一致的明文凭证。
func TestAuthenticateManager_CreateToken_MasksTokenInOperationLog(t *testing.T) {
	ctx := context.Background()
	recorder := &fakeOperationLogRecorder{}

	name := "tc-probe"
	scope := "Support"
	rawToken := "-HnqrSE-fCjFaRN3ys-9"

	created := false
	manager := NewAuthenticateManager(&fakeTxn{}, &fakeAuthenticateStorager{
		fetchTokensFn: func(ctx context.Context, filter *TokenFilter) ([]*Token, error) {
			if filter.Token != nil {
				return nil, nil // token uniqueness check
			}
			if created {
				return []*Token{{ID: 7, Name: name, Token: rawToken, Scope: scope}}, nil
			}
			return nil, nil // name existence check
		},
		createTokenFn: func(ctx context.Context, token *TokenParam) error {
			created = true
			return nil
		},
	}, &fakeAuthorizeStorager{})
	manager.SetOperationLogManager(recorder)

	token, err := manager.CreateToken(ctx, &TokenParam{Name: &name, Scope: &scope}, nil)
	require.NoError(t, err)
	require.NotNil(t, token)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionCreate), entry.Action)
	assert.Equal(t, string(ioperlog.ResourceTypeToken), entry.ResourceType)
	assert.Equal(t, ioperlog.StatusSuccess, entry.Status)

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok)
	for _, key := range []string{"id", "name", "scope", "token"} {
		assert.Contains(t, after, key)
	}
	assert.Equal(t, "******", after["token"])

	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), rawToken)
}

// issue #162：token 删除审计的 change_summary.before.token 必须为全掩码，
// 删除资源不能阻断已落库明文的提权面。
func TestAuthenticateManager_DeleteToken_MasksTokenInOperationLog(t *testing.T) {
	ctx := context.Background()
	recorder := &fakeOperationLogRecorder{}

	rawToken := "-HnqrSE-fCjFaRN3ys-9"
	token := &Token{ID: 7, Name: "tc-probe", Token: rawToken, Scope: "Support"}

	manager := NewAuthenticateManager(&fakeTxn{}, &fakeAuthenticateStorager{
		deleteTokenFn: func(ctx context.Context, param *Token) error {
			return nil
		},
	}, &fakeAuthorizeStorager{})
	manager.SetOperationLogManager(recorder)

	err := manager.DeleteToken(ctx, token)
	require.NoError(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionDelete), entry.Action)

	before, ok := entry.ChangeSummary["before"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "******", before["token"])

	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), rawToken)
}

// issue #162：创建失败路径经 tokenParamToMap 写入，param 携带 token 时同样不得泄漏。
func TestAuthenticateManager_CreateTokenFailure_MasksParamTokenInOperationLog(t *testing.T) {
	ctx := context.Background()
	recorder := &fakeOperationLogRecorder{}

	name := "tc-probe"
	scope := "Support"
	rawToken := "param-raw-token-1234"

	manager := NewAuthenticateManager(&fakeTxn{}, &fakeAuthenticateStorager{
		fetchTokensFn: func(ctx context.Context, filter *TokenFilter) ([]*Token, error) {
			return nil, nil
		},
		createTokenFn: func(ctx context.Context, token *TokenParam) error {
			return errors.New("db connection failed")
		},
	}, &fakeAuthorizeStorager{})
	manager.SetOperationLogManager(recorder)

	_, err := manager.CreateToken(ctx, &TokenParam{Name: &name, Token: &rawToken, Scope: &scope}, nil)
	require.Error(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, ioperlog.StatusFailed, entry.Status)

	after, ok := entry.ChangeSummary["after"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "******", after["token"])

	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), rawToken)
}
