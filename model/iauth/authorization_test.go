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

package iauth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rainway-ai-gateway/ai-gateway-api/model/ibasic"
	"github.com/rainway-ai-gateway/ai-gateway-api/stateful"
)

func TestAuthorizeManager_Authorizate(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	t.Run("admin allowed all", func(t *testing.T) {
		m := NewAuthorizeManager(&fakeTxn{}, &fakeAuthorizeStorager{})
		v := &Visitor{User: &User{Name: "admin", Admin: true}}
		ctx := NewVisitorContext(ctx, v)

		auth := NewFeatureAuthorization(FeatureUser, ActionDelete)
		require.NoError(t, m.Authorizate(ctx, auth))
	})

	t.Run("product scope allowed read", func(t *testing.T) {
		m := NewAuthorizeManager(&fakeTxn{}, &fakeAuthorizeStorager{})
		v := &Visitor{Token: &Token{Name: "token", Scope: ScopeProduct}}
		ctx := NewVisitorContext(ctx, v)

		auth := NewFeatureAuthorization(FeatureUser, ActionReadAll)
		require.NoError(t, m.Authorizate(ctx, auth))
	})

	t.Run("product scope denied delete", func(t *testing.T) {
		m := NewAuthorizeManager(&fakeTxn{}, &fakeAuthorizeStorager{})
		v := &Visitor{Token: &Token{Name: "token", Scope: ScopeProduct}}
		ctx := NewVisitorContext(ctx, v)

		auth := NewFeatureAuthorization(FeatureUser, ActionDelete)
		err := m.Authorizate(ctx, auth)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Feature Access Deny")
	})

	t.Run("unknown scope denied", func(t *testing.T) {
		m := NewAuthorizeManager(&fakeTxn{}, &fakeAuthorizeStorager{})
		v := &Visitor{Token: &Token{Name: "token", Scope: "UnknownScope"}}
		ctx := NewVisitorContext(ctx, v)

		auth := NewFeatureAuthorization(FeatureUser, ActionRead)
		err := m.Authorizate(ctx, auth)
		require.Error(t, err)
	})

	t.Run("product validation success", func(t *testing.T) {
		product := &ibasic.Product{Name: "product-1"}
		authStore := &fakeAuthorizeStorager{
			isTokenProductGrantedFn: func(ctx context.Context, token *Token, product *ibasic.Product) (bool, error) {
				return true, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{Token: &Token{Name: "token", Scope: ScopeProduct}}
		ctx := NewVisitorContext(ctx, v)
		ctx = ibasic.NewProductContext(ctx, product)

		auth := NewFeatureAuthorizerWithFactoryWithProduct(FeatureAPIKey, ActionRead)
		require.NoError(t, m.Authorizate(ctx, auth))
	})

	t.Run("product validation fail", func(t *testing.T) {
		product := &ibasic.Product{Name: "product-1"}
		authStore := &fakeAuthorizeStorager{
			isTokenProductGrantedFn: func(ctx context.Context, token *Token, product *ibasic.Product) (bool, error) {
				return false, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{Token: &Token{Name: "token", Scope: ScopeProduct}}
		ctx := NewVisitorContext(ctx, v)
		ctx = ibasic.NewProductContext(ctx, product)

		auth := NewFeatureAuthorizerWithFactoryWithProduct(FeatureAPIKey, ActionRead)
		err := m.Authorizate(ctx, auth)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Product Access Deny")
	})

	t.Run("visitor not login", func(t *testing.T) {
		stateful.DefaultConfig.RunTime.SkipTokenValidate = false
		m := NewAuthorizeManager(&fakeTxn{}, &fakeAuthorizeStorager{})

		auth := NewFeatureAuthorization(FeatureUser, ActionRead)
		err := m.Authorizate(ctx, auth)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "User Not Login")
	})
}

func TestAuthorizeManager_IsVisitorProductGranted(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	t.Run("user granted", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			isUserProductGrantedFn: func(ctx context.Context, user *User, product *ibasic.Product) (bool, error) {
				return true, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{User: &User{Name: "alice"}}

		ok, err := m.IsVisitorProductGranted(ctx, v, &ibasic.Product{Name: "p1"})
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("token granted", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			isTokenProductGrantedFn: func(ctx context.Context, token *Token, product *ibasic.Product) (bool, error) {
				return true, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{Token: &Token{Name: "token"}}

		ok, err := m.IsVisitorProductGranted(ctx, v, &ibasic.Product{Name: "p1"})
		require.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestAuthorizeManager_FetchVisitorProductList(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	t.Run("user products", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			fetchUserProductsFn: func(ctx context.Context, user *User) ([]*ibasic.Product, error) {
				return []*ibasic.Product{{Name: "p1"}}, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{User: &User{Name: "alice"}}

		products, err := m.FetchVisitorProductList(ctx, v)
		require.NoError(t, err)
		assert.Len(t, products, 1)
	})

	t.Run("token product", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			fetchTokenProductFn: func(ctx context.Context, token *Token) (*ibasic.Product, error) {
				return &ibasic.Product{Name: "p1"}, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)
		v := &Visitor{Token: &Token{Name: "token"}}

		products, err := m.FetchVisitorProductList(ctx, v)
		require.NoError(t, err)
		assert.Len(t, products, 1)
	})
}

func TestAuthorizeManager_FetchProductUsers(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	authStore := &fakeAuthorizeStorager{
		fetchProductUsersFn: func(ctx context.Context, product *ibasic.Product) ([]*User, error) {
			return []*User{{Name: "alice"}}, nil
		},
	}
	m := NewAuthorizeManager(&fakeTxn{}, authStore)

	users, err := m.FetchProductUsers(ctx, &ibasic.Product{Name: "p1"})
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestAuthorizeManager_BindUserProduct(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	authStore := &fakeAuthorizeStorager{
		bindUserProductFn: func(ctx context.Context, user *User, product *ibasic.Product) error {
			return nil
		},
	}
	m := NewAuthorizeManager(&fakeTxn{}, authStore)

	require.NoError(t, m.BindUserProduct(ctx, &User{Name: "alice"}, &ibasic.Product{Name: "p1"}))
}

func TestAuthorizeManager_UnBindUserProduct(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	t.Run("success", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			isUserProductGrantedFn: func(ctx context.Context, user *User, product *ibasic.Product) (bool, error) {
				return true, nil
			},
			unbindUserProductFn: func(ctx context.Context, user *User, product *ibasic.Product) error {
				return nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)

		require.NoError(t, m.UnBindUserProduct(ctx, &User{Name: "alice"}, &ibasic.Product{Name: "p1"}))
	})

	t.Run("binding not exist", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			isUserProductGrantedFn: func(ctx context.Context, user *User, product *ibasic.Product) (bool, error) {
				return false, nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)

		err := m.UnBindUserProduct(ctx, &User{Name: "alice"}, &ibasic.Product{Name: "p1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "User Product Binding Record Not Exist")
	})
}

func TestAuthorizeManager_FetchProductTokens(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	authStore := &fakeAuthorizeStorager{
		fetchProductTokensFn: func(ctx context.Context, product *ibasic.Product) ([]*Token, error) {
			return []*Token{{Name: "token-1"}}, nil
		},
	}
	m := NewAuthorizeManager(&fakeTxn{}, authStore)

	tokens, err := m.FetchProductTokens(ctx, &ibasic.Product{Name: "p1"})
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
}

func TestAuthorizeManager_UpdateUserIsAdmin(t *testing.T) {
	ctx := context.Background()
	setupConfig(t)

	t.Run("set admin", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			updateUserScopesFn: func(ctx context.Context, user *User, scopes []string) error {
				assert.Equal(t, []string{ScopeSystem}, scopes)
				return nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)

		require.NoError(t, m.UpdateUserIsAdmin(ctx, &User{Name: "alice"}, true))
	})

	t.Run("unset admin", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			updateUserScopesFn: func(ctx context.Context, user *User, scopes []string) error {
				assert.Equal(t, []string{ScopeProduct}, scopes)
				return nil
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)

		require.NoError(t, m.UpdateUserIsAdmin(ctx, &User{Name: "alice"}, false))
	})

	t.Run("storage error", func(t *testing.T) {
		authStore := &fakeAuthorizeStorager{
			updateUserScopesFn: func(ctx context.Context, user *User, scopes []string) error {
				return errors.New("db error")
			},
		}
		m := NewAuthorizeManager(&fakeTxn{}, authStore)

		err := m.UpdateUserIsAdmin(ctx, &User{Name: "alice"}, true)
		require.Error(t, err)
	})
}
