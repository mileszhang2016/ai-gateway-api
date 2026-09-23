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

package iprovider

import (
	"context"
	"encoding/json"
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

// issue #162：provider 上游凭证经整结构 marshal 进入审计 map，keys 数组中的
// key 必须脱敏（部分掩码契约），不得记录明文 bearer 凭证。
func TestProviderManager_CreateProvider_MasksKeysInOperationLog(t *testing.T) {
	ctx := context.Background()
	recorder := &fakeOperationLogRecorder{}

	name := "prov-one"
	rawKey := "sk-abcdefghijklmnop"

	manager := NewProviderManager(&fakeTxn{}, &fakeProviderStorager{
		fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, param *ProviderParam) (int64, error) {
			return 1, nil
		},
	})
	manager.SetOperationLogManager(recorder)

	_, err := manager.CreateProvider(ctx, &ProviderParam{
		Name:           &name,
		ModelProtocols: []string{"openai"},
		Models:         []string{"m1"},
		Keys:           []ProviderKey{{Name: "k1", Key: rawKey}},
		InstancePool:   []ProviderInstance{{Addr: "127.0.0.1", Port: 8080, Weight: 1}},
	})
	require.NoError(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionCreate), entry.Action)

	keys := maskedKeys(t, entry.ChangeSummary, "after")
	require.Len(t, keys, 1)
	k0, ok := keys[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "k1", k0["name"])
	assert.Equal(t, "sk-a****mnop", k0["key"])

	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), rawKey)
}

// issue #162：provider 更新审计的 before（库中快照）与 after（请求参数）
// 两侧 keys 都必须脱敏。
func TestProviderManager_UpdateProvider_MasksKeysInOperationLog(t *testing.T) {
	ctx := context.Background()
	recorder := &fakeOperationLogRecorder{}

	name := "prov-one"
	oldRawKey := "sk-oldkey1234567890"
	newRawKey := "sk-newkey1234567890"

	manager := NewProviderManager(&fakeTxn{}, &fakeProviderStorager{
		fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
			return &Provider{
				ID:             1,
				Name:           name,
				ModelProtocols: []string{"openai"},
				Keys:           []ProviderKey{{Name: "k1", Key: oldRawKey}},
				InstancePool:   []ProviderInstance{{Addr: "127.0.0.1", Port: 8080, Weight: 1}},
			}, nil
		},
		updateFn: func(ctx context.Context, name string, param *ProviderParam) error {
			return nil
		},
	})
	manager.SetOperationLogManager(recorder)

	err := manager.UpdateProvider(ctx, name, &ProviderParam{
		Name:           &name,
		ModelProtocols: []string{"openai"},
		Keys:           []ProviderKey{{Name: "k1", Key: newRawKey}},
		InstancePool:   []ProviderInstance{{Addr: "127.0.0.1", Port: 8080, Weight: 1}},
	})
	require.NoError(t, err)

	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, string(ioperlog.ActionUpdate), entry.Action)

	beforeKeys := maskedKeys(t, entry.ChangeSummary, "before")
	require.Len(t, beforeKeys, 1)
	b0, ok := beforeKeys[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "sk-o****7890", b0["key"])

	afterKeys := maskedKeys(t, entry.ChangeSummary, "after")
	require.Len(t, afterKeys, 1)
	a0, ok := afterKeys[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "sk-n****7890", a0["key"])

	serialized, err := json.Marshal(entry.ChangeSummary)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), oldRawKey)
	assert.NotContains(t, string(serialized), newRawKey)
}

func maskedKeys(t *testing.T, summary map[string]interface{}, side string) []interface{} {
	t.Helper()
	m, ok := summary[side].(map[string]interface{})
	require.True(t, ok)
	keys, ok := m["keys"].([]interface{})
	require.True(t, ok)
	return keys
}

// TestProviderParamToMap_DropsOmittedFields guards the issue #201 family
// fix: ProviderParam pointer fields without omitempty (e.g. name) must not
// materialize as null entries when omitted from a partial update, which
// would produce phantom diff_keys.
func TestProviderParamToMap_DropsOmittedFields(t *testing.T) {
	m := providerParamToMap(&ProviderParam{ModelProtocols: []string{"openai"}})
	require.NotNil(t, m)
	assert.NotContains(t, m, "name")
	assert.Equal(t, []interface{}{"openai"}, m["model_protocols"])

	name := "p1"
	m2 := providerParamToMap(&ProviderParam{Name: &name, ModelProtocols: []string{"openai"}})
	require.NotNil(t, m2)
	assert.Equal(t, "p1", m2["name"])

	// Provider (before snapshot) has the same nil-materialization shape.
	p := providerToMap(&Provider{Name: "p1"})
	require.NotNil(t, p)
	assert.NotContains(t, p, "model_endpoint")
	assert.Equal(t, "p1", p["name"])

	assert.Nil(t, providerParamToMap(nil))
	assert.Nil(t, providerToMap(nil))
}
