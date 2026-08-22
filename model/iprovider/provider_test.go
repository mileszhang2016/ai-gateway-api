// Copyright(c) 2026 The Rainway AI Gateway (壬远AI网关) Authors.
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

package iprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/rainway-ai-gateway/ai-gateway-api/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validProviderParam() *ProviderParam {
	return &ProviderParam{
		Name:        lib.PString("deepseek"),
		Description: lib.PString("DeepSeek API"),
		ModelEndpoint: &ProviderEndpoint{
			Schema: "https",
			URI:    "/v1/models",
		},
		Models: []string{"deepseek-chat", "deepseek-coder"},
		Keys: []ProviderKey{
			{Name: "key-primary", Key: "sk-aaaaaaaaaaaa"},
		},
		InstancePool: []ProviderInstance{
			{Name: "backend-1", Addr: "api.deepseek.com", Port: 443, Weight: 100},
		},
		ModelProtocols: []string{"openai"},
	}
}

func TestNewProviderManager(t *testing.T) {
	m := NewProviderManager(&fakeTxn{}, &fakeProviderStorager{})
	require.NotNil(t, m)
}

func TestProviderManager_CreateProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		store := &fakeProviderStorager{}
		m := NewProviderManager(&fakeTxn{}, store)
		id, err := m.CreateProvider(ctx, validProviderParam())
		require.NoError(t, err)
		assert.Equal(t, int64(1), id)
	})

	t.Run("duplicate", func(t *testing.T) {
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{Name: "deepseek"}, nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		_, err := m.CreateProvider(ctx, validProviderParam())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider Record Existed")
	})

	t.Run("invalid name", func(t *testing.T) {
		m := NewProviderManager(&fakeTxn{}, &fakeProviderStorager{})
		param := validProviderParam()
		param.Name = lib.PString("")
		_, err := m.CreateProvider(ctx, param)
		require.Error(t, err)
	})
}

func TestProviderManager_UpdateProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		updated := false
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{Name: "deepseek"}, nil
			},
			updateFn: func(ctx context.Context, name string, param *ProviderParam) error {
				updated = true
				return nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		err := m.UpdateProvider(ctx, "deepseek", validProviderParam())
		require.NoError(t, err)
		assert.True(t, updated)
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeProviderStorager{}
		m := NewProviderManager(&fakeTxn{}, store)
		err := m.UpdateProvider(ctx, "deepseek", validProviderParam())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider Record Not Exist")
	})
}

func TestProviderManager_DeleteProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deleted := false
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{Name: "deepseek"}, nil
			},
			deleteFn: func(ctx context.Context, name string) error {
				deleted = true
				return nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		err := m.DeleteProvider(ctx, "deepseek")
		require.NoError(t, err)
		assert.True(t, deleted)
	})

	t.Run("referenced", func(t *testing.T) {
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{Name: "deepseek"}, nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		err := m.DeleteProvider(ctx, "deepseek", func(ctx context.Context, name string) error {
			return errors.New("referenced by cluster")
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "referenced by cluster")
	})
}

func TestProviderManager_FetchProvider(t *testing.T) {
	ctx := context.Background()
	expected := &Provider{Name: "deepseek"}
	store := &fakeProviderStorager{
		fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
			return expected, nil
		},
	}
	m := NewProviderManager(&fakeTxn{}, store)
	got, err := m.FetchProvider(ctx, &ProviderFilter{Name: lib.PString("deepseek")})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestProviderManager_FetchProviderList(t *testing.T) {
	ctx := context.Background()

	t.Run("no filter", func(t *testing.T) {
		expected := []*Provider{{Name: "deepseek"}}
		store := &fakeProviderStorager{
			fetchListFn: func(ctx context.Context, filter *ProviderFilter) ([]*Provider, int64, error) {
				return expected, 1, nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		list, total, err := m.FetchProviderList(ctx, &ProviderFilter{})
		require.NoError(t, err)
		assert.Equal(t, expected, list)
		assert.Equal(t, int64(1), total)
	})

	t.Run("filter by model_protocol", func(t *testing.T) {
		all := []*Provider{
			{Name: "openai-provider", ModelProtocols: []string{"openai"}},
			{Name: "anthropic-provider", ModelProtocols: []string{"anthropic"}},
			{Name: "multi-provider", ModelProtocols: []string{"openai", "anthropic"}},
		}
		store := &fakeProviderStorager{
			fetchListFn: func(ctx context.Context, filter *ProviderFilter) ([]*Provider, int64, error) {
				return all, int64(len(all)), nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		proto := "openai"
		list, total, err := m.FetchProviderList(ctx, &ProviderFilter{ModelProtocol: &proto})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, list, 2)
		assert.Equal(t, "openai-provider", list[0].Name)
		assert.Equal(t, "multi-provider", list[1].Name)
	})
}

func TestProviderManager_DiscoverModels(t *testing.T) {
	ctx := context.Background()

	t.Run("success openai", func(t *testing.T) {
		updated := false
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{
					Name:           "deepseek",
					Keys:           []ProviderKey{{Name: "k1", Key: "sk-xxx"}},
					InstancePool:   []ProviderInstance{{Addr: "api.deepseek.com", Port: 443}},
					ModelProtocols: []string{"openai"},
					ModelEndpoint:  &ProviderEndpoint{Schema: "https", URI: "/v1/models"},
				}, nil
			},
			updateFn: func(ctx context.Context, name string, param *ProviderParam) error {
				updated = true
				assert.Equal(t, []string{"m1", "m2"}, param.Models)
				return nil
			},
		}
		caller := &fakeDiscoverCaller{body: []byte(`{"data":[{"id":"m1"},{"id":"m2"}]}`)}
		m := NewProviderManager(&fakeTxn{}, store)
		models, err := m.DiscoverModelsWithCaller(ctx, "deepseek", caller)
		require.NoError(t, err)
		assert.Equal(t, []string{"m1", "m2"}, models)
		assert.True(t, updated)
	})

	t.Run("no keys", func(t *testing.T) {
		store := &fakeProviderStorager{
			fetchFn: func(ctx context.Context, filter *ProviderFilter) (*Provider, error) {
				return &Provider{
					Name:           "deepseek",
					InstancePool:   []ProviderInstance{{Addr: "api.deepseek.com", Port: 443}},
					ModelProtocols: []string{"openai"},
				}, nil
			},
		}
		m := NewProviderManager(&fakeTxn{}, store)
		_, err := m.DiscoverModelsWithCaller(ctx, "deepseek", &fakeDiscoverCaller{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no keys")
	})

	t.Run("not found", func(t *testing.T) {
		m := NewProviderManager(&fakeTxn{}, &fakeProviderStorager{})
		_, err := m.DiscoverModelsWithCaller(ctx, "deepseek", &fakeDiscoverCaller{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider Record Not Exist")
	})
}

func TestValidateProviderParam(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, ValidateProviderParam(validProviderParam()))
	})

	t.Run("nil", func(t *testing.T) {
		require.Error(t, ValidateProviderParam(nil))
	})

	t.Run("missing name", func(t *testing.T) {
		p := validProviderParam()
		p.Name = nil
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("empty instance pool", func(t *testing.T) {
		p := validProviderParam()
		p.InstancePool = nil
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("empty model protocols", func(t *testing.T) {
		p := validProviderParam()
		p.ModelProtocols = nil
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("invalid protocol", func(t *testing.T) {
		p := validProviderParam()
		p.ModelProtocols = []string{"unknown"}
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("duplicate protocol", func(t *testing.T) {
		p := validProviderParam()
		p.ModelProtocols = []string{"openai", "openai"}
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("invalid endpoint uri", func(t *testing.T) {
		p := validProviderParam()
		p.ModelEndpoint.URI = "models"
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("duplicate model", func(t *testing.T) {
		p := validProviderParam()
		p.Models = []string{"m", "m"}
		require.Error(t, ValidateProviderParam(p))
	})

	t.Run("duplicate key name", func(t *testing.T) {
		p := validProviderParam()
		p.Keys = []ProviderKey{{Name: "k", Key: "a"}, {Name: "k", Key: "b"}}
		require.Error(t, ValidateProviderParam(p))
	})
}

func TestProviderName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NoError(t, ProviderName("deepseek"))
	})

	t.Run("empty", func(t *testing.T) {
		require.Error(t, ProviderName(""))
	})

	t.Run("invalid chars", func(t *testing.T) {
		require.Error(t, ProviderName("deep seek"))
	})

	t.Run("edge chars", func(t *testing.T) {
		require.Error(t, ProviderName("-deepseek"))
	})
}

func TestFillDefaults(t *testing.T) {
	p := &ProviderParam{}
	FillDefaults(p)
	require.NotNil(t, p.ModelEndpoint)
	assert.Equal(t, "https", p.ModelEndpoint.Schema)
	assert.Equal(t, "/v1/models", p.ModelEndpoint.URI)
	assert.NotNil(t, p.Models)
	assert.NotNil(t, p.Keys)
}

func TestKeyMap(t *testing.T) {
	keys := []ProviderKey{{Name: "a", Key: "1"}, {Name: "b", Key: "2"}}
	m := KeyMap(keys)
	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, m)
}

func TestHasModel(t *testing.T) {
	p := &Provider{Models: []string{"m1"}}
	assert.True(t, HasModel(p, "m1"))
	assert.False(t, HasModel(p, "m2"))
	assert.False(t, HasModel(nil, "m1"))
}

func TestBuildDiscoverURL(t *testing.T) {
	inst := ProviderInstance{Addr: "api.example.com", Port: 443}
	url := BuildDiscoverURL(&ProviderEndpoint{Schema: "https", URI: "/v1/models"}, inst)
	assert.Equal(t, "https://api.example.com:443/v1/models", url)
}

func TestBuildAuthHeader(t *testing.T) {
	k, v := BuildAuthHeader("openai", "sk-xxx")
	assert.Equal(t, "Authorization", k)
	assert.Equal(t, "Bearer sk-xxx", v)

	k, v = BuildAuthHeader("anthropic", "sk-xxx")
	assert.Equal(t, "x-api-key", k)

	k, v = BuildAuthHeader("gemini", "xxx")
	assert.Equal(t, "x-goog-api-key", k)
}

func TestParseModelDiscoveryResponse(t *testing.T) {
	t.Run("openai data", func(t *testing.T) {
		models, err := ParseModelDiscoveryResponse([]byte(`{"data":[{"id":"m1"},{"id":"m2"}]}`), "openai")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1", "m2"}, models)
	})

	t.Run("models array with name", func(t *testing.T) {
		models, err := ParseModelDiscoveryResponse([]byte(`{"models":[{"name":"m1"},{"name":"m2"}]}`), "gemini")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1", "m2"}, models)
	})

	t.Run("models array of strings", func(t *testing.T) {
		models, err := ParseModelDiscoveryResponse([]byte(`{"models":["m1","m2"]}`), "openai")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1", "m2"}, models)
	})

	t.Run("single model object", func(t *testing.T) {
		models, err := ParseModelDiscoveryResponse([]byte(`{"models":{"name":"m1"}}`), "gemini")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1"}, models)
	})

	t.Run("empty", func(t *testing.T) {
		_, err := ParseModelDiscoveryResponse([]byte(`{}`), "openai")
		require.Error(t, err)
	})
}
